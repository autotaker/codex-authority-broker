package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

type authorizeRecorder struct {
	mu         sync.Mutex
	responses  []authorizeResult
	requests   []ipc.Request
	invocation int
}

type authorizeResult struct {
	response ipc.Response
	err      error
}

func (r *authorizeRecorder) call(_ context.Context, request ipc.Request) (ipc.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, request)
	r.invocation++
	if len(r.responses) == 0 {
		return ipc.Response{}, errors.New("fixture daemon unavailable")
	}
	result := r.responses[0]
	r.responses = r.responses[1:]
	return result.response, result.err
}

func (r *authorizeRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.requests)
}

func (r *authorizeRecorder) requestsCopy() []ipc.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]ipc.Request(nil), r.requests...)
}

func allowResult() authorizeResult {
	return authorizeResult{response: ipc.Response{Version: ipc.ProtocolVersion, OK: true}}
}

func denyResult() authorizeResult {
	return authorizeResult{response: ipc.Response{Version: ipc.ProtocolVersion, OK: false}}
}

func invoke(t *testing.T, recorder *authorizeRecorder, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	status := run(args, strings.NewReader("authority-from-stdin\n"), &stdout, &stderr, recorder.call)
	return status, stdout.String(), stderr.String()
}

func requireAuthorizeRequest(t *testing.T, request ipc.Request) {
	t.Helper()
	if request.Version != ipc.ProtocolVersion || request.Operation != ipc.OperationAuthorize || len(request.Payload) != 0 {
		t.Fatalf("request was not fixed payload-free authorize: %+v", request)
	}
}

func requireDenied(t *testing.T, status int, stdout, stderr string) {
	t.Helper()
	if status == 0 || stdout != "" || stderr != deniedLine {
		t.Fatalf("expected bounded deny, status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestLiveLeasePermitsPerInvocation(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult()}}
	status, stdout, stderr := invoke(t, recorder)
	if status != 0 || stdout != "" || stderr != "" {
		t.Fatalf("live allow was not silent success: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	if recorder.count() != 1 {
		t.Fatalf("authorize calls = %d, want exactly one", recorder.count())
	}
	requireAuthorizeRequest(t, recorder.requestsCopy()[0])
}

func TestExpiryDeniesWithoutCachedReuse(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult(), denyResult()}}
	if status, stdout, stderr := invoke(t, recorder); status != 0 || stdout != "" || stderr != "" {
		t.Fatal("initial unexpired lease did not permit")
	}
	status, stdout, stderr := invoke(t, recorder)
	requireDenied(t, status, stdout, stderr)
	if recorder.count() != 2 {
		t.Fatalf("expiry path calls = %d, want two live requests", recorder.count())
	}
}

func TestDaemonUnavailableDeniesWithoutCachedReuse(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult(), {err: errors.New("socket unavailable")}}}
	_, _, _ = invoke(t, recorder)
	status, stdout, stderr := invoke(t, recorder)
	requireDenied(t, status, stdout, stderr)
	if recorder.count() != 2 {
		t.Fatalf("unavailable path calls = %d, want two live attempts", recorder.count())
	}
}

func TestDaemonRestartDeniesUntilFreshLiveAllow(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult(), denyResult(), allowResult()}}
	_, _, _ = invoke(t, recorder)
	status, stdout, stderr := invoke(t, recorder)
	requireDenied(t, status, stdout, stderr)
	status, stdout, stderr = invoke(t, recorder)
	if status != 0 || stdout != "" || stderr != "" {
		t.Fatal("fresh post-restart allow was not accepted")
	}
	if recorder.count() != 3 {
		t.Fatalf("restart path calls = %d, want three independent requests", recorder.count())
	}
}

func TestMalformedReplyDeniesWithoutCachedReuse(t *testing.T) {
	malformed := []authorizeResult{
		{response: ipc.Response{Version: ipc.ProtocolVersion, OK: true, Payload: []byte(`{"raw":"reply"}`)}},
		{response: ipc.Response{Version: 99, OK: true}},
		{err: errors.New("truncated frame")},
	}
	for index, result := range malformed {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			recorder := &authorizeRecorder{responses: []authorizeResult{allowResult(), result}}
			_, _, _ = invoke(t, recorder)
			status, stdout, stderr := invoke(t, recorder)
			requireDenied(t, status, stdout, stderr)
			if recorder.count() != 2 {
				t.Fatalf("malformed path calls = %d, want two", recorder.count())
			}
		})
	}
}

func TestUnauthorizedReplyDeniesWithoutCachedReuse(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{
		allowResult(),
		{response: ipc.Response{Version: ipc.ProtocolVersion, OK: true, Payload: []byte(`{"identity":"unauthorized"}`)}},
	}}
	_, _, _ = invoke(t, recorder)
	status, stdout, stderr := invoke(t, recorder)
	requireDenied(t, status, stdout, stderr)
	if recorder.count() != 2 {
		t.Fatalf("unauthorized path calls = %d, want two", recorder.count())
	}
}

func TestNoTimestampCacheTwoConsecutiveInvocations(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult(), denyResult()}}
	status, stdout, stderr := invoke(t, recorder)
	if status != 0 || stdout != "" || stderr != "" {
		t.Fatal("first invocation did not permit")
	}
	status, stdout, stderr = invoke(t, recorder)
	requireDenied(t, status, stdout, stderr)
	requests := recorder.requestsCopy()
	if len(requests) != 2 {
		t.Fatalf("consecutive invocation requests = %d, want two", len(requests))
	}
	for _, request := range requests {
		requireAuthorizeRequest(t, request)
	}
}

func TestArgvAndLogRedaction(t *testing.T) {
	const sentinel = "lease-secret-sentinel"
	recorder := &authorizeRecorder{responses: []authorizeResult{{response: ipc.Response{Version: ipc.ProtocolVersion, OK: true, Payload: []byte(`{"sentinel":"` + sentinel + `"}`)}}}}
	status, stdout, stderr := invoke(t, recorder, sentinel)
	requireDenied(t, status, stdout, stderr)
	for _, output := range []string{stdout, stderr} {
		if strings.Contains(output, sentinel) {
			t.Fatalf("sentinel leaked in output %q", output)
		}
	}
	if recorder.count() != 1 {
		t.Fatalf("argv-bearing invocation made %d requests, want exactly one", recorder.count())
	}
	requireAuthorizeRequest(t, recorder.requestsCopy()[0])
}

func TestFixtureScaffoldingIsIsolated(t *testing.T) {
	root := os.Getenv("CODEX_AUTHORITY_SUDO_FIXTURE_ROOT")
	if root == "" {
		t.Skip("actual sudo fixture is owned and launched by Main")
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == "/" {
		t.Fatal("fixture root must be an explicit non-root absolute path")
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		t.Fatalf("fixture root is not an existing directory: %v", err)
	}
	// The test intentionally performs no mkdir, chmod, identity, policy, PAM,
	// socket, or timestamp mutation. Main's isolated namespace owns those.
}

func TestFixtureRollbackAndNoWorkstationMutation(t *testing.T) {
	root := os.Getenv("CODEX_AUTHORITY_SUDO_FIXTURE_ROOT")
	if root == "" {
		t.Skip("actual sudo fixture rollback is owned and launched by Main")
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == "/" {
		t.Fatal("fixture root must be an explicit non-root absolute path")
	}
	for _, relative := range []string{"etc", "run", "var/log"} {
		path := filepath.Join(root, relative)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("fixture path %q is unavailable: %v", relative, err)
		}
	}
	// Rollback and host hash comparison are performed by Main outside this
	// process. This test remains read-only and never touches workstation state.
}

func TestRunNeverReadsStdinOrEnvironmentAuthority(t *testing.T) {
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult()}}
	stdin := errorReader{err: errors.New("stdin must not be read")}
	var stdout, stderr bytes.Buffer
	status := run(nil, stdin, &stdout, &stderr, recorder.call)
	if status != 0 || recorder.count() != 1 {
		t.Fatalf("stdin-independent invocation failed: status=%d calls=%d", status, recorder.count())
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestRunContextIsBounded(t *testing.T) {
	started := make(chan struct{})
	recorder := func(ctx context.Context, _ ipc.Request) (ipc.Response, error) {
		close(started)
		<-ctx.Done()
		return ipc.Response{}, ctx.Err()
	}
	var stderr bytes.Buffer
	start := time.Now()
	status := run(nil, nil, io.Discard, &stderr, recorder)
	if status == 0 || stderr.String() != deniedLine || time.Since(start) > authorityTimeout+time.Second {
		t.Fatalf("bounded denial failed: status=%d stderr=%q elapsed=%s", status, stderr.String(), time.Since(start))
	}
	select {
	case <-started:
	default:
		t.Fatal("transport was not attempted")
	}
}
