package backend

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
	"github.com/autotaker/codex-authority-broker/internal/lease"
)

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

type auditRecord struct {
	CorrelationID string  `json:"correlation_id"`
	ActorUID      uint32  `json:"actor_uid"`
	Scope         string  `json:"scope"`
	Result        string  `json:"result"`
	LeaseExpiry   *string `json:"lease_expiry"`
}

func startAuditServer(t *testing.T, runtime *Runtime) ipc.Client {
	t.Helper()
	path := filepath.Join(t.TempDir(), "audit.sock")
	server, err := ipc.Listen(ipc.Config{Path: path, AllowedUID: uint32(os.Geteuid())}, runtime)
	if err != nil {
		t.Skipf("Unix socket fixture unavailable: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		_ = server.Close()
		runtime.Close()
	})
	return ipc.Client{Path: path}
}

func auditCall(t *testing.T, client ipc.Client, request ipc.Request) ipc.Response {
	t.Helper()
	response, err := client.Call(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func parseAuditRecords(t *testing.T, data []byte) []auditRecord {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	records := make([]auditRecord, len(lines))
	for index, line := range lines {
		if len(line)+1 > maxAuditBytes {
			t.Fatal("audit line exceeded bound")
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil || len(fields) != 5 {
			t.Fatal("audit event was not an exact five-field object")
		}
		for _, name := range []string{"correlation_id", "actor_uid", "scope", "result", "lease_expiry"} {
			if _, ok := fields[name]; !ok {
				t.Fatalf("audit field %q missing", name)
			}
		}
		if err := json.Unmarshal(line, &records[index]); err != nil {
			t.Fatal(err)
		}
	}
	return records
}

func TestAuditExactSchemaSequenceExpiryAndRedaction(t *testing.T) {
	secret := []byte("AUDIT-SECRET-SENTINEL")
	clock := &testClock{now: time.Unix(1_701_000_000, 0)}
	var sink bytes.Buffer
	runtime, err := newRuntimeWithAudit(secret, clock, &sink)
	if err != nil {
		t.Fatal(err)
	}
	client := startAuditServer(t, runtime)
	if response := auditCall(t, client, testRequest(ipc.OperationReady, []byte(`"PAYLOAD-SENTINEL"`))); response.OK {
		t.Fatal("ready payload was allowed")
	}
	if response := auditCall(t, client, testRequest(ipc.OperationReady, nil)); !response.OK {
		t.Fatal("ready was denied")
	}
	clock.advance(30 * time.Second)
	expiry := clock.Now().Add(300 * time.Second).UTC().Format(time.RFC3339Nano)
	if response := auditCall(t, client, testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now()))); !response.OK {
		t.Fatal("otp was denied")
	}
	if response := auditCall(t, client, testRequest(ipc.OperationAuthorize, nil)); !response.OK {
		t.Fatal("authorize was denied")
	}
	if response := auditCall(t, client, testRequest(ipc.OperationReady, nil)); response.OK {
		t.Fatal("ready replaced an active lease")
	}

	records := parseAuditRecords(t, sink.Bytes())
	if len(records) != 5 {
		t.Fatalf("audit records = %d", len(records))
	}
	wantScope := []string{"ready", "ready", "otp", "authorize", "ready"}
	wantResult := []string{"deny", "allow", "allow", "allow", "deny"}
	for index, record := range records {
		value, parseErr := strconv.ParseUint(record.CorrelationID, 16, 64)
		if parseErr != nil || value != uint64(index+1) || record.CorrelationID != strconv.FormatUint(value, 16) {
			t.Fatal("correlation sequence was not fresh nonzero lowercase hex")
		}
		if record.ActorUID != uint32(os.Geteuid()) || record.Scope != wantScope[index] || record.Result != wantResult[index] {
			t.Fatal("audit tuple did not match its request")
		}
		if index == 2 || index == 3 {
			if record.LeaseExpiry == nil || *record.LeaseExpiry != expiry {
				t.Fatal("allowed authority expiry was not immutable")
			}
		} else if record.LeaseExpiry != nil {
			t.Fatal("ready or denial carried an expiry")
		}
	}
	for _, sentinel := range []string{"AUDIT-SECRET-SENTINEL", "PAYLOAD-SENTINEL", `"code"`} {
		if bytes.Contains(sink.Bytes(), []byte(sentinel)) {
			t.Fatal("sensitive input entered audit output")
		}
	}
}

type failingAuditWriter struct {
	calls  int
	failAt int
	short  bool
}

func (w *failingAuditWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.calls != w.failAt {
		return len(data), nil
	}
	if w.short {
		return len(data) - 1, nil
	}
	return 0, io.ErrClosedPipe
}

func TestAuditFailureClosesRuntimeWithoutRetry(t *testing.T) {
	for _, test := range []struct {
		name   string
		failAt int
		short  bool
		scope  string
	}{
		{name: "ready-error", failAt: 1, scope: ipc.OperationReady},
		{name: "otp-short", failAt: 2, short: true, scope: ipc.OperationOTP},
		{name: "authorize-error", failAt: 3, scope: ipc.OperationAuthorize},
	} {
		t.Run(test.name, func(t *testing.T) {
			secret := []byte("audit-failure-fixture")
			clock := &testClock{now: time.Unix(1_701_100_000, 0)}
			writer := &failingAuditWriter{failAt: test.failAt, short: test.short}
			runtime, err := newRuntimeWithAudit(secret, clock, writer)
			if err != nil {
				t.Fatal(err)
			}
			client := startAuditServer(t, runtime)
			if test.scope != ipc.OperationReady {
				if response := auditCall(t, client, testRequest(ipc.OperationReady, nil)); !response.OK {
					t.Fatal("readiness setup failed")
				}
				clock.advance(30 * time.Second)
			}
			if test.scope == ipc.OperationAuthorize {
				if response := auditCall(t, client, testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now()))); !response.OK {
					t.Fatal("otp setup failed")
				}
			}
			request := testRequest(test.scope, nil)
			if test.scope == ipc.OperationOTP {
				request = testRequest(test.scope, otpPayload(secret, clock.Now()))
			}
			if response := auditCall(t, client, request); response.OK {
				t.Fatal("sink failure published an allow")
			}
			if response := auditCall(t, client, testRequest(ipc.OperationAuthorize, nil)); response.OK {
				t.Fatal("authority remained usable after sink failure")
			}
			if writer.calls != test.failAt {
				t.Fatalf("audit writes = %d, want %d", writer.calls, test.failAt)
			}
		})
	}
}

type blockingAuditWriter struct {
	mu      sync.Mutex
	calls   int
	lines   [][]byte
	entered chan struct{}
	release chan struct{}
	blockAt int
}

func (w *blockingAuditWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	w.calls++
	w.lines = append(w.lines, append([]byte(nil), data...))
	call := w.calls
	w.mu.Unlock()
	if call == w.blockAt {
		close(w.entered)
		<-w.release
	}
	return len(data), nil
}

func TestOTPAuthorityWaitsForSuccessfulAuditWrite(t *testing.T) {
	secret := []byte("audit-publication-barrier")
	clock := &testClock{now: time.Unix(1_701_200_000, 0)}
	writer := &blockingAuditWriter{entered: make(chan struct{}), release: make(chan struct{}), blockAt: 2}
	runtime, err := newRuntimeWithAudit(secret, clock, writer)
	if err != nil {
		t.Fatal(err)
	}
	client := startAuditServer(t, runtime)
	if response := auditCall(t, client, testRequest(ipc.OperationReady, nil)); !response.OK {
		t.Fatal("readiness setup failed")
	}
	clock.advance(30 * time.Second)
	type callResult struct {
		response ipc.Response
		err      error
	}
	otpDone := make(chan callResult, 1)
	go func() {
		response, callErr := client.Call(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
		otpDone <- callResult{response: response, err: callErr}
	}()
	select {
	case <-writer.entered:
	case <-time.After(time.Second):
		t.Fatal("otp audit write did not reach barrier")
	}
	authorizeDone := make(chan callResult, 1)
	go func() {
		response, callErr := client.Call(context.Background(), testRequest(ipc.OperationAuthorize, nil))
		authorizeDone <- callResult{response: response, err: callErr}
	}()
	select {
	case <-authorizeDone:
		t.Fatal("authorize observed lease before otp audit completed")
	case <-time.After(50 * time.Millisecond):
	}
	close(writer.release)
	if result := <-otpDone; result.err != nil || !result.response.OK {
		t.Fatal("otp was denied after successful audit")
	}
	if result := <-authorizeDone; result.err != nil || !result.response.OK {
		t.Fatal("authorize was denied after otp audit completed")
	}
	writer.mu.Lock()
	lines := append([][]byte(nil), writer.lines...)
	writer.mu.Unlock()
	records := parseAuditRecords(t, bytes.Join(lines, nil))
	if len(records) != 3 || records[1].CorrelationID == records[2].CorrelationID || records[1].Scope != "otp" || records[2].Scope != "authorize" || records[1].Result != "allow" || records[2].Result != "allow" {
		t.Fatal("concurrent audit identifiers or tuples crossed")
	}
}

func TestCloseWaitsForAuditPublication(t *testing.T) {
	writer := &blockingAuditWriter{entered: make(chan struct{}), release: make(chan struct{}), blockAt: 1}
	runtime, err := newRuntimeWithAudit([]byte("audit-close-barrier"), nil, writer)
	if err != nil {
		t.Fatal(err)
	}
	client := startAuditServer(t, runtime)
	done := make(chan ipc.Response, 1)
	go func() {
		response, _ := client.Call(context.Background(), testRequest(ipc.OperationReady, nil))
		done <- response
	}()
	select {
	case <-writer.entered:
	case <-time.After(time.Second):
		t.Fatal("audit write did not reach publication barrier")
	}
	closed := make(chan struct{})
	go func() { runtime.Close(); close(closed) }()
	select {
	case <-closed:
		t.Fatal("Close completed before audit publication")
	case <-time.After(50 * time.Millisecond):
	}
	close(writer.release)
	if response := <-done; !response.OK {
		t.Fatal("successfully audited decision was denied")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not complete after audit publication")
	}
}

func TestAuditSequenceOverflowClosesWithoutWrite(t *testing.T) {
	var sink bytes.Buffer
	runtime, err := newRuntimeWithAudit([]byte("audit-overflow-fixture"), nil, &sink)
	if err != nil {
		t.Fatal(err)
	}
	runtime.auditID.Store(^uint64(0))
	client := startAuditServer(t, runtime)
	if response := auditCall(t, client, testRequest(ipc.OperationReady, nil)); response.OK {
		t.Fatal("sequence overflow published an allow")
	}
	if response := auditCall(t, client, testRequest(ipc.OperationAuthorize, nil)); response.OK {
		t.Fatal("runtime remained usable after sequence overflow")
	}
	if sink.Len() != 0 {
		t.Fatal("sequence overflow attempted an audit write")
	}
}

func TestAuditContextAbsenceClosesWithoutWrite(t *testing.T) {
	var sink bytes.Buffer
	runtime, err := newRuntimeWithAudit([]byte("audit-context-fixture"), nil, &sink)
	if err != nil {
		t.Fatal(err)
	}
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
	runtime.mu.Lock()
	closed := runtime.closed
	runtime.mu.Unlock()
	if !closed || sink.Len() != 0 {
		t.Fatal("missing actor context did not close before audit")
	}
}

func TestClosureBeforeAuditWritesDeny(t *testing.T) {
	var sink bytes.Buffer
	runtime, err := newRuntimeWithAudit([]byte("audit-cancel-fixture"), nil, &sink)
	if err != nil {
		t.Fatal(err)
	}
	runtime.beforePublish = func(bool) { runtime.closeUnderAudit() }
	client := startAuditServer(t, runtime)
	if response := auditCall(t, client, testRequest(ipc.OperationReady, nil)); response.OK {
		t.Fatal("closed decision was published as allow")
	}
	records := parseAuditRecords(t, sink.Bytes())
	if len(records) != 1 || records[0].Result != "deny" || records[0].LeaseExpiry != nil {
		t.Fatal("pre-audit closure was not a deny with null expiry")
	}
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) advance(delta time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(delta)
	c.mu.Unlock()
}

func testRequest(operation string, payload []byte) ipc.Request {
	return ipc.Request{Version: ipc.ProtocolVersion, Operation: operation, Payload: payload}
}

func otpPayload(secret []byte, now time.Time) []byte {
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], uint64(now.Unix()/30))
	hash := hmac.New(sha1.New, secret)
	_, _ = hash.Write(message[:])
	digest := hash.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff
	code := fmt.Sprintf("%06d", value%1000000)
	return []byte(`{"code":"` + code + `"}`)
}

func assertDenied(t *testing.T, response ipc.Response, err error) {
	t.Helper()
	if err != nil || response.OK || len(response.Payload) != 0 {
		t.Fatalf("got response=%+v err=%v", response, err)
	}
}

func assertAccepted(t *testing.T, response ipc.Response, err error) {
	t.Helper()
	if err != nil || !response.OK || len(response.Payload) != 0 {
		t.Fatalf("got response=%+v err=%v", response, err)
	}
}

func TestNewRejectsEmptySecret(t *testing.T) {
	runtime, err := New(nil)
	if runtime != nil || !errors.Is(err, lease.ErrInvalidTOTP) {
		t.Fatalf("runtime=%v err=%v", runtime, err)
	}
	if strings.Contains(fmt.Sprint(err), "secret") {
		t.Fatalf("secret marker in construction error: %v", err)
	}
}

func TestReadyOTPExactAdmissionAndState(t *testing.T) {
	secret := []byte("synthetic-runtime-secret")
	clock := &testClock{now: time.Unix(1_700_000_000, 0)}
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	validReady := testRequest(ipc.OperationReady, nil)
	validOTP := func() ipc.Request { return testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())) }

	response, err := runtime.Handle(context.Background(), validOTP())
	assertDenied(t, response, err)
	readyPayloads := [][]byte{[]byte(`{}`), []byte(`null`), []byte(`[]`), []byte(`"ready"`), []byte(`1`), []byte(" "), []byte(strings.Repeat("x", 128))}
	for _, payload := range readyPayloads {
		response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, payload))
		assertDenied(t, response, err)
	}
	response, err = runtime.Handle(context.Background(), validReady)
	assertAccepted(t, response, err)
	response, err = runtime.Handle(context.Background(), validReady)
	assertAccepted(t, response, err)

	badOTP := [][]byte{
		[]byte{}, []byte(" "), []byte(`{"code":"12345"}`), []byte(`{"code":"1234567"}`),
		[]byte(`{"code":"12345x"}`), []byte(`{"code":"１２３４５６"}`), []byte(`{"code":123456}`),
		[]byte(`{"code":"123456","x":1}`), []byte(`{"code":"123456","code":"123456"}`),
		[]byte(`{"code":"123456"} `), []byte(`{"cod\\u0065":"123456"}`), []byte(`{"code":"123456"}{}`),
		[]byte(strings.Repeat("x", 128)),
	}
	for _, payload := range badOTP {
		response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, payload))
		assertDenied(t, response, err)
	}
	clock.advance(30 * time.Second)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertAccepted(t, response, err)
	response, err = runtime.Handle(context.Background(), validReady)
	assertDenied(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertDenied(t, response, err)
}

func TestAuthorizeActiveLeaseBoundary(t *testing.T) {
	secret := []byte("authorize-boundary-secret")
	clock := &testClock{now: time.Unix(1_700_400_000, 0)}
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	request := testRequest(ipc.OperationAuthorize, nil)
	response, err := runtime.Handle(context.Background(), request)
	assertDenied(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertAccepted(t, response, err)
	response, err = runtime.Handle(context.Background(), request)
	assertAccepted(t, response, err)
	clock.advance(300*time.Second - time.Nanosecond)
	response, err = runtime.Handle(context.Background(), request)
	assertAccepted(t, response, err)
	clock.advance(time.Nanosecond)
	response, err = runtime.Handle(context.Background(), request)
	assertDenied(t, response, err)
	clock.advance(30 * time.Second)
	response, err = runtime.Handle(context.Background(), request)
	assertDenied(t, response, err)
}

func TestAuthorizeFreshRuntimeDenies(t *testing.T) {
	secret := []byte("authorize-fresh-runtime-secret")
	clock := &testClock{now: time.Unix(1_700_500_000, 0)}
	active, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	response, err := active.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	response, err = active.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertAccepted(t, response, err)
	response, err = active.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
	assertAccepted(t, response, err)

	fresh, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	response, err = fresh.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
	assertDenied(t, response, err)
}

func TestAuthorizePayloadAndReadinessOTPNonInterference(t *testing.T) {
	secret := []byte("authorize-noninterference-secret")
	clock := &testClock{now: time.Unix(1_700_600_000, 0)}
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range [][]byte{[]byte(`{}`), []byte(`null`), []byte(`[]`), []byte(`"authorize"`), []byte(" ")} {
		response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationAuthorize, payload))
		assertDenied(t, response, err)
	}
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertDenied(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
	assertDenied(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertAccepted(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
	assertAccepted(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
}

func authorizeActiveRuntime(t *testing.T, secret []byte, clock *testClock) *Runtime {
	t.Helper()
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertAccepted(t, response, err)
	return runtime
}

func TestAuthorizeCallerCancellationFailsClosed(t *testing.T) {
	clock := &testClock{now: time.Unix(1_700_700_000, 0)}
	runtime := authorizeActiveRuntime(t, []byte("authorize-cancel-secret"), clock)
	decision := make(chan bool, 1)
	release := make(chan struct{})
	runtime.beforePublish = func(ok bool) {
		decision <- ok
		<-release
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(ctx, testRequest(ipc.OperationAuthorize, nil))
		result <- response
	}()
	select {
	case ok := <-decision:
		if !ok {
			t.Fatal("authorize decision was not positive")
		}
	case <-time.After(time.Second):
		t.Fatal("authorize did not reach publication barrier")
	}
	cancel()
	close(release)
	assertDenied(t, <-result, nil)
}

func TestAuthorizeExpiryAndCloseRaceFailsClosed(t *testing.T) {
	t.Run("expiry", func(t *testing.T) {
		clock := &testClock{now: time.Unix(1_700_800_000, 0)}
		runtime := authorizeActiveRuntime(t, []byte("authorize-expiry-race-secret"), clock)
		decision := make(chan bool, 1)
		release := make(chan struct{})
		runtime.beforePublish = func(ok bool) {
			decision <- ok
			<-release
		}
		result := make(chan ipc.Response, 1)
		go func() {
			response, _ := runtime.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
			result <- response
		}()
		select {
		case ok := <-decision:
			if !ok {
				t.Fatal("authorize decision was not positive")
			}
		case <-time.After(time.Second):
			t.Fatal("authorize did not reach publication barrier")
		}
		clock.advance(300 * time.Second)
		close(release)
		assertDenied(t, <-result, nil)
	})

	t.Run("close", func(t *testing.T) {
		clock := &testClock{now: time.Unix(1_700_900_000, 0)}
		runtime := authorizeActiveRuntime(t, []byte("authorize-close-race-secret"), clock)
		decision := make(chan bool, 1)
		release := make(chan struct{})
		runtime.beforePublish = func(ok bool) {
			decision <- ok
			<-release
		}
		result := make(chan ipc.Response, 1)
		go func() {
			response, _ := runtime.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
			result <- response
		}()
		select {
		case ok := <-decision:
			if !ok {
				t.Fatal("authorize decision was not positive")
			}
		case <-time.After(time.Second):
			t.Fatal("authorize did not reach publication barrier")
		}
		runtime.Close()
		close(release)
		assertDenied(t, <-result, nil)
	})
}

func TestVersionContextAndRedaction(t *testing.T) {
	secret := []byte("UNIQUE-SECRET-MARKER")
	runtime, err := newRuntime(secret, nil)
	if err != nil {
		t.Fatal(err)
	}
	requests := []ipc.Request{
		{Version: 0, Operation: ipc.OperationReady},
		{Version: ipc.ProtocolVersion, Operation: "audit"},
		{Version: ipc.ProtocolVersion, Operation: ipc.OperationReady, Payload: []byte("marker")},
		{Version: ipc.ProtocolVersion, Operation: ipc.OperationOTP, Payload: []byte(`{"code":"UNIQUE"}`)},
	}
	for _, request := range requests {
		response, err := runtime.Handle(context.Background(), request)
		assertDenied(t, response, err)
		if strings.Contains(fmt.Sprintf("%+v%v", response, err), "UNIQUE") {
			t.Fatal("marker leaked in denial")
		}
	}
	response, err := runtime.Handle(nil, testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
	if strings.Contains(fmt.Sprintf("%v", runtime), string(secret)) {
		t.Fatal("secret leaked through runtime formatting")
	}
}

func TestExpiredAndRateLimitedChallengesDeny(t *testing.T) {
	secret := []byte("state-denial-secret")
	clock := &testClock{now: time.Unix(1_700_200_000, 0)}
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(301 * time.Second)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertDenied(t, response, err)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	payload := otpPayload(secret, clock.Now())
	invalidPayload := append([]byte(nil), payload...)
	if invalidPayload[len(otpPrefix)] == '0' {
		invalidPayload[len(otpPrefix)] = '1'
	} else {
		invalidPayload[len(otpPrefix)] = '0'
	}
	for index := 0; index < 6; index++ {
		response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, invalidPayload))
		assertDenied(t, response, err)
	}
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, payload))
	assertDenied(t, response, err)
}

func TestRegisterBoundsAndFixedAllowlist(t *testing.T) {
	runtime, err := newRuntime([]byte("registration-secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	called := 0
	handler := func(context.Context, ipc.Request) bool { called++; return true }
	if err := runtime.Register("audit", handler); err != nil {
		t.Fatalf("custom registration: %v", err)
	}
	for _, operation := range []string{"", "Audit", "a", "a!", "a.", strings.Repeat("a", 17)} {
		if err := runtime.Register(operation, handler); err == nil {
			t.Fatalf("accepted invalid operation %q", operation)
		}
	}
	if err := runtime.Register("audit", handler); err == nil {
		t.Fatal("accepted duplicate")
	}
	if err := runtime.Register("fourth", handler); err == nil {
		t.Fatal("accepted second custom slot")
	}
	response, err := runtime.Handle(context.Background(), testRequest("audit", nil))
	assertDenied(t, response, err)
	if called != 0 {
		t.Fatal("registered custom operation was dispatched")
	}
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationAuthorize, nil))
	assertDenied(t, response, err)
	runtime.Close()
	if err := runtime.Register("later", handler); err == nil {
		t.Fatal("accepted post-close registration")
	}
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
	runtime.Close()
}

func TestConcurrentOTPAndReadyInterleaving(t *testing.T) {
	secret := []byte("concurrency-secret")
	clock := &testClock{now: time.Unix(1_700_100_000, 0)}
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	payload := otpPayload(secret, clock.Now())
	var wait sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for index := 0; index < 24; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			operation := ipc.OperationOTP
			requestPayload := payload
			if index%6 == 0 {
				operation = ipc.OperationReady
				requestPayload = nil
			}
			response, err := runtime.Handle(context.Background(), testRequest(operation, requestPayload))
			if operation == ipc.OperationOTP && err == nil && response.OK {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}(index)
	}
	wait.Wait()
	if successes != 1 {
		t.Fatalf("concurrent OTP successes=%d", successes)
	}
}

func TestReadyOTPBarrierDoesNotUseStaleChallenge(t *testing.T) {
	secret := []byte("barrier-secret")
	clock := &testClock{now: time.Unix(1_700_300_000, 0)}
	runtime, err := newRuntime(secret, clock)
	if err != nil {
		t.Fatal(err)
	}
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertAccepted(t, response, err)
	clock.advance(30 * time.Second)
	started := make(chan struct{})
	release := make(chan struct{})
	runtime.mu.Lock()
	originalReady := runtime.handlers[ipc.OperationReady]
	runtime.handlers[ipc.OperationReady] = func(ctx context.Context, request ipc.Request) bool {
		close(started)
		<-release
		return originalReady(ctx, request)
	}
	runtime.mu.Unlock()
	readyResult := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
		readyResult <- response
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("ready did not reach barrier")
	}
	otpResult := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
		otpResult <- response
	}()
	assertAccepted(t, <-otpResult, nil)
	close(release)
	assertDenied(t, <-readyResult, nil)
	response, err = runtime.Handle(context.Background(), testRequest(ipc.OperationOTP, otpPayload(secret, clock.Now())))
	assertDenied(t, response, err)
}

func TestCloseCancelsAdmittedCallAndFailsClosed(t *testing.T) {
	runtime, err := newRuntime([]byte("close-secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	canceled := make(chan struct{})
	runtime.mu.Lock()
	runtime.handlers[ipc.OperationReady] = func(ctx context.Context, _ ipc.Request) bool {
		close(started)
		select {
		case <-ctx.Done():
			close(canceled)
		case <-release:
		}
		return true
	}
	runtime.mu.Unlock()
	result := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
		result <- response
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	closed := make(chan struct{})
	go func() { runtime.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close blocked behind handler")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel handler context")
	}
	close(release)
	assertDenied(t, <-result, nil)
	runtime.Close()
	response, err := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
}

func TestCallerCancellation(t *testing.T) {
	runtime, err := newRuntime([]byte("caller-secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err := runtime.Handle(ctx, testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
}

func TestCallerCancellationReachesAdmittedHandler(t *testing.T) {
	runtime, err := newRuntime([]byte("admitted-caller-secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	canceled := make(chan struct{})
	runtime.mu.Lock()
	runtime.handlers[ipc.OperationReady] = func(ctx context.Context, _ ipc.Request) bool {
		close(started)
		<-ctx.Done()
		close(canceled)
		return true
	}
	runtime.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(ctx, testRequest(ipc.OperationReady, nil))
		result <- response
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler was not admitted")
	}
	cancel()
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("caller cancellation did not reach handler")
	}
	assertDenied(t, <-result, nil)
}

func TestCallerCancellationWinsBeforeSuccessPublication(t *testing.T) {
	runtime, err := newRuntime([]byte("publication-caller-secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	returned := make(chan struct{})
	runtime.mu.Lock()
	runtime.handlers[ipc.OperationReady] = func(context.Context, ipc.Request) bool {
		defer close(returned)
		close(started)
		<-release
		return true
	}
	runtime.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(ctx, testRequest(ipc.OperationReady, nil))
		result <- response
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler was not admitted")
	}
	runtime.mu.Lock()
	close(release)
	select {
	case <-returned:
	case <-time.After(time.Second):
		runtime.mu.Unlock()
		t.Fatal("handler did not return before publication barrier")
	}
	cancel()
	runtime.mu.Unlock()
	assertDenied(t, <-result, nil)
}

func TestCloseWinsWaitingHandleGate(t *testing.T) {
	runtime, err := newRuntime([]byte("waiting-close-secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{}, 1)
	runtime.handlers[ipc.OperationReady] = func(context.Context, ipc.Request) bool {
		started <- struct{}{}
		return true
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	runtime.beforeGate = func() {
		close(entered)
		<-release
	}
	result := make(chan ipc.Response, 1)
	go func() {
		response, _ := runtime.Handle(context.Background(), testRequest(ipc.OperationReady, nil))
		result <- response
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("Handle did not reach the gate barrier")
	}
	runtime.Close()
	close(release)
	assertDenied(t, <-result, nil)
	select {
	case <-started:
		t.Fatal("handler started after Close won the gate")
	default:
	}
}
