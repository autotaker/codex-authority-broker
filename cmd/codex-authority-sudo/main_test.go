package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
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

type fixtureFileInfo struct{ stat syscall.Stat_t }

func (f fixtureFileInfo) Name() string       { return "fixture" }
func (f fixtureFileInfo) Size() int64        { return 0 }
func (f fixtureFileInfo) Mode() os.FileMode  { return os.FileMode(f.stat.Mode) }
func (f fixtureFileInfo) ModTime() time.Time { return time.Time{} }
func (f fixtureFileInfo) IsDir() bool        { return f.stat.Mode&syscall.S_IFMT == syscall.S_IFDIR }
func (f fixtureFileInfo) Sys() any           { value := f.stat; return &value }

type identityFixture struct {
	events       []string
	paths        []string
	socketStats  []fixedSocketStat
	groups       []int
	gid, egid    int
	uid, euid    int
	setgroupsErr error
	getgroupsErr error
	setgidErr    error
	setuidErr    error
}

func newIdentityFixture() *identityFixture {
	return &identityFixture{
		socketStats: []fixedSocketStat{{dev: 1, ino: 2, mode: syscall.S_IFSOCK | 0o600, uid: 1000, gid: 1000}},
	}
}

func (f *identityFixture) hooks() identityHooks {
	return identityHooks{
		lstat: func(path string) (os.FileInfo, error) {
			f.paths = append(f.paths, path)
			if path == defaultSocketDir {
				return fixtureFileInfo{stat: syscall.Stat_t{Mode: syscall.S_IFDIR | 0o755, Uid: 0, Gid: 0}}, nil
			}
			index := len(f.paths) - 2
			if index < 0 || index >= len(f.socketStats) {
				index = len(f.socketStats) - 1
			}
			if index < 0 {
				return nil, errors.New("missing fixture socket")
			}
			stat := f.socketStats[index]
			return fixtureFileInfo{stat: syscall.Stat_t{Dev: stat.dev, Ino: stat.ino, Mode: stat.mode, Uid: stat.uid, Gid: stat.gid}}, nil
		},
		setgroups: func(groups []int) error {
			f.events = append(f.events, "groups")
			if f.setgroupsErr != nil {
				return f.setgroupsErr
			}
			f.groups = append([]int(nil), groups...)
			return nil
		},
		getgroups: func() ([]int, error) {
			f.events = append(f.events, "groups-empty")
			if f.getgroupsErr != nil {
				return nil, f.getgroupsErr
			}
			return append([]int(nil), f.groups...), nil
		},
		setgid: func(gid int) error {
			f.events = append(f.events, "gid")
			if f.setgidErr != nil {
				return f.setgidErr
			}
			f.gid, f.egid = gid, gid
			return nil
		},
		getgid: func() int {
			f.events = append(f.events, "gid-real")
			return f.gid
		},
		getegid: func() int {
			f.events = append(f.events, "gid-effective")
			return f.egid
		},
		setuid: func(uid int) error {
			f.events = append(f.events, "uid")
			if f.setuidErr != nil {
				return f.setuidErr
			}
			f.uid, f.euid = uid, uid
			return nil
		},
		getuid: func() int {
			f.events = append(f.events, "uid-real")
			return f.uid
		},
		geteuid: func() int {
			f.events = append(f.events, "uid-effective")
			return f.euid
		},
	}
}

func metadataInfo(stat fixedSocketStat) os.FileInfo {
	return fixtureFileInfo{stat: syscall.Stat_t{
		Dev: stat.dev, Ino: stat.ino, Mode: stat.mode, Uid: stat.uid, Gid: stat.gid,
	}}
}

func metadataLstat(parent fixedSocketStat, sockets []fixedSocketStat, failures map[int]error) func(string) (os.FileInfo, error) {
	call := 0
	return func(path string) (os.FileInfo, error) {
		index := call
		call++
		if err := failures[index]; err != nil {
			return nil, err
		}
		if index == 0 {
			if path != defaultSocketDir {
				return nil, errors.New("unexpected parent path")
			}
			return metadataInfo(parent), nil
		}
		if path != defaultSocketPath || index-1 >= len(sockets) {
			return nil, errors.New("unexpected socket path")
		}
		return metadataInfo(sockets[index-1]), nil
	}
}

func validParentStat() fixedSocketStat {
	return fixedSocketStat{dev: 1, ino: 1, mode: syscall.S_IFDIR | 0o755, uid: 0, gid: 0}
}

func validAuthoritySocketStat() fixedSocketStat {
	return fixedSocketStat{dev: 1, ino: 2, mode: syscall.S_IFSOCK | 0o660, uid: 1000, gid: 1000}
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
	status := runWithHooks(args, strings.NewReader("authority-from-stdin\n"), &stdout, &stderr, recorder.call, newIdentityFixture().hooks())
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

func TestFixedSocketIdentityParentAndPathAdmission(t *testing.T) {
	parent := validParentStat()
	socket := validAuthoritySocketStat()
	replacement := socket
	replacement.ino++
	tests := []struct {
		name     string
		parent   fixedSocketStat
		sockets  []fixedSocketStat
		failures map[int]error
		wantID   uint32
		wantOK   bool
	}{
		{name: "valid fixed metadata", parent: parent, sockets: []fixedSocketStat{socket, socket}, wantID: 1000, wantOK: true},
		{name: "missing parent", parent: parent, sockets: []fixedSocketStat{socket, socket}, failures: map[int]error{0: os.ErrNotExist}},
		{name: "parent regular", parent: fixedSocketStat{mode: syscall.S_IFREG | 0o755}, sockets: []fixedSocketStat{socket, socket}},
		{name: "parent symlink", parent: fixedSocketStat{mode: syscall.S_IFLNK | 0o755}, sockets: []fixedSocketStat{socket, socket}},
		{name: "parent non-root", parent: fixedSocketStat{mode: syscall.S_IFDIR | 0o755, uid: 1}, sockets: []fixedSocketStat{socket, socket}},
		{name: "parent group writable", parent: fixedSocketStat{mode: syscall.S_IFDIR | 0o775}, sockets: []fixedSocketStat{socket, socket}},
		{name: "parent world writable", parent: fixedSocketStat{mode: syscall.S_IFDIR | 0o757}, sockets: []fixedSocketStat{socket, socket}},
		{name: "missing socket", parent: parent, sockets: []fixedSocketStat{socket, socket}, failures: map[int]error{1: os.ErrNotExist}},
		{name: "socket symlink", parent: parent, sockets: []fixedSocketStat{{mode: syscall.S_IFLNK, uid: 1000, gid: 1000}, socket}},
		{name: "socket regular", parent: parent, sockets: []fixedSocketStat{{mode: syscall.S_IFREG | 0o660, uid: 1000, gid: 1000}, socket}},
		{name: "socket directory", parent: parent, sockets: []fixedSocketStat{{mode: syscall.S_IFDIR | 0o660, uid: 1000, gid: 1000}, socket}},
		{name: "missing final metadata", parent: parent, sockets: []fixedSocketStat{socket, socket}, failures: map[int]error{2: os.ErrNotExist}},
		{name: "replaced socket", parent: parent, sockets: []fixedSocketStat{socket, replacement}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, ok := fixedSocketIdentity(metadataLstat(test.parent, test.sockets, test.failures))
			if id != test.wantID || ok != test.wantOK {
				t.Fatalf("fixedSocketIdentity() = %d, %v; want %d, %v", id, ok, test.wantID, test.wantOK)
			}
		})
	}
}

func TestFixedSocketIdentityRejectsZeroAndMismatchedUIDGID(t *testing.T) {
	parent := validParentStat()
	for _, test := range []struct {
		name string
		uid  uint32
		gid  uint32
	}{
		{name: "both zero"},
		{name: "uid zero", uid: 0, gid: 1000},
		{name: "gid zero", uid: 1000, gid: 0},
		{name: "distinct nonzero", uid: 1000, gid: 1001},
	} {
		t.Run(test.name, func(t *testing.T) {
			socket := validAuthoritySocketStat()
			socket.uid, socket.gid = test.uid, test.gid
			if id, ok := fixedSocketIdentity(metadataLstat(parent, []fixedSocketStat{socket, socket}, nil)); ok || id != 0 {
				t.Fatalf("identity %d:%d accepted as %d", test.uid, test.gid, id)
			}
			fixture := newIdentityFixture()
			hooks := fixture.hooks()
			hooks.lstat = metadataLstat(parent, []fixedSocketStat{socket, socket}, nil)
			calls := 0
			var stdout, stderr bytes.Buffer
			status := runWithHooks(nil, nil, &stdout, &stderr, func(context.Context, ipc.Request) (ipc.Response, error) {
				calls++
				return ipc.Response{Version: ipc.ProtocolVersion, OK: true}, nil
			}, hooks)
			requireDenied(t, status, stdout.String(), stderr.String())
			if calls != 0 || len(fixture.events) != 0 {
				t.Fatalf("identity %d:%d performed actions=%v calls=%d", test.uid, test.gid, fixture.events, calls)
			}
		})
	}
}

func TestRunDropsIdentityBeforeExactlyOneAuthorize(t *testing.T) {
	fixture := newIdentityFixture()
	var request ipc.Request
	callCount := 0
	call := func(_ context.Context, got ipc.Request) (ipc.Response, error) {
		fixture.events = append(fixture.events, "call")
		callCount++
		request = got
		return ipc.Response{Version: ipc.ProtocolVersion, OK: true}, nil
	}
	var stdout, stderr bytes.Buffer
	status := runWithHooks([]string{"ignored"}, errorReader{err: errors.New("stdin read")}, &stdout, &stderr, call, fixture.hooks())
	wantOrder := []string{"groups", "groups-empty", "gid", "gid-real", "gid-effective", "uid", "uid-real", "uid-effective", "call"}
	if status != 0 || stdout.Len() != 0 || stderr.Len() != 0 || callCount != 1 {
		t.Fatalf("valid handoff status=%d stdout=%q stderr=%q calls=%d", status, stdout.String(), stderr.String(), callCount)
	}
	if !reflect.DeepEqual(fixture.events, wantOrder) {
		t.Fatalf("handoff order = %v, want %v", fixture.events, wantOrder)
	}
	if !reflect.DeepEqual(fixture.paths, []string{defaultSocketDir, defaultSocketPath, defaultSocketPath}) {
		t.Fatalf("metadata paths = %v", fixture.paths)
	}
	requireAuthorizeRequest(t, request)
}

func TestRunIdentityDropFailuresStopBeforeTransport(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*identityFixture, *identityHooks)
		want   []string
	}{
		{name: "setgroups", mutate: func(f *identityFixture, _ *identityHooks) { f.setgroupsErr = errors.New("setgroups") }, want: []string{"groups"}},
		{name: "getgroups", mutate: func(f *identityFixture, _ *identityHooks) { f.getgroupsErr = errors.New("getgroups") }, want: []string{"groups", "groups-empty"}},
		{name: "residual group", mutate: func(f *identityFixture, hooks *identityHooks) {
			hooks.getgroups = func() ([]int, error) { f.events = append(f.events, "groups-empty"); return []int{7}, nil }
		}, want: []string{"groups", "groups-empty"}},
		{name: "setgid", mutate: func(f *identityFixture, _ *identityHooks) { f.setgidErr = errors.New("setgid") }, want: []string{"groups", "groups-empty", "gid"}},
		{name: "real gid verification", mutate: func(f *identityFixture, hooks *identityHooks) {
			hooks.getgid = func() int { f.events = append(f.events, "gid-real"); return 999 }
		}, want: []string{"groups", "groups-empty", "gid", "gid-real"}},
		{name: "effective gid verification", mutate: func(f *identityFixture, hooks *identityHooks) {
			hooks.getegid = func() int { f.events = append(f.events, "gid-effective"); return 999 }
		}, want: []string{"groups", "groups-empty", "gid", "gid-real", "gid-effective"}},
		{name: "setuid", mutate: func(f *identityFixture, _ *identityHooks) { f.setuidErr = errors.New("setuid") }, want: []string{"groups", "groups-empty", "gid", "gid-real", "gid-effective", "uid"}},
		{name: "real uid verification", mutate: func(f *identityFixture, hooks *identityHooks) {
			hooks.getuid = func() int { f.events = append(f.events, "uid-real"); return 999 }
		}, want: []string{"groups", "groups-empty", "gid", "gid-real", "gid-effective", "uid", "uid-real"}},
		{name: "effective uid verification", mutate: func(f *identityFixture, hooks *identityHooks) {
			hooks.geteuid = func() int { f.events = append(f.events, "uid-effective"); return 999 }
		}, want: []string{"groups", "groups-empty", "gid", "gid-real", "gid-effective", "uid", "uid-real", "uid-effective"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityFixture()
			hooks := fixture.hooks()
			test.mutate(fixture, &hooks)
			calls := 0
			var stdout, stderr bytes.Buffer
			status := runWithHooks(nil, nil, &stdout, &stderr, func(context.Context, ipc.Request) (ipc.Response, error) {
				calls++
				return ipc.Response{Version: ipc.ProtocolVersion, OK: true}, nil
			}, hooks)
			requireDenied(t, status, stdout.String(), stderr.String())
			if calls != 0 || !reflect.DeepEqual(fixture.events, test.want) {
				t.Fatalf("failure actions=%v calls=%d, want actions=%v and zero calls", fixture.events, calls, test.want)
			}
		})
	}
}

func TestRunMetadataFailuresStopBeforeDropAndTransport(t *testing.T) {
	parent := validParentStat()
	socket := validAuthoritySocketStat()
	replacement := socket
	replacement.ino++
	for _, test := range []struct {
		name     string
		parent   fixedSocketStat
		sockets  []fixedSocketStat
		failures map[int]error
	}{
		{name: "unsafe parent", parent: fixedSocketStat{mode: syscall.S_IFDIR | 0o777}, sockets: []fixedSocketStat{socket, socket}},
		{name: "missing socket", parent: parent, sockets: []fixedSocketStat{socket, socket}, failures: map[int]error{1: os.ErrNotExist}},
		{name: "non-socket", parent: parent, sockets: []fixedSocketStat{{mode: syscall.S_IFREG | 0o600, uid: 1000, gid: 1000}, socket}},
		{name: "replacement", parent: parent, sockets: []fixedSocketStat{socket, replacement}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityFixture()
			hooks := fixture.hooks()
			hooks.lstat = metadataLstat(test.parent, test.sockets, test.failures)
			calls := 0
			var stdout, stderr bytes.Buffer
			status := runWithHooks(nil, nil, &stdout, &stderr, func(context.Context, ipc.Request) (ipc.Response, error) {
				calls++
				return ipc.Response{Version: ipc.ProtocolVersion, OK: true}, nil
			}, hooks)
			requireDenied(t, status, stdout.String(), stderr.String())
			if calls != 0 || len(fixture.events) != 0 {
				t.Fatalf("metadata denial performed actions=%v calls=%d", fixture.events, calls)
			}
		})
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
	const sentinel = "untrusted-pam-identity-sentinel"
	t.Setenv("PAM_USER", sentinel)
	t.Setenv("PAM_RUSER", sentinel)
	t.Setenv("CODEX_AUTHORITY_UID", "4242")
	recorder := &authorizeRecorder{responses: []authorizeResult{allowResult()}}
	stdin := errorReader{err: errors.New("stdin must not be read")}
	fixture := newIdentityFixture()
	var stdout, stderr bytes.Buffer
	status := runWithHooks([]string{sentinel, "4242"}, stdin, &stdout, &stderr, recorder.call, fixture.hooks())
	if status != 0 || recorder.count() != 1 || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("untrusted-input-independent invocation failed: status=%d calls=%d stdout=%q stderr=%q", status, recorder.count(), stdout.String(), stderr.String())
	}
	if fixture.uid != 1000 || fixture.euid != 1000 || fixture.gid != 1000 || fixture.egid != 1000 {
		t.Fatalf("selected identity = uid %d/%d gid %d/%d", fixture.uid, fixture.euid, fixture.gid, fixture.egid)
	}
	request := recorder.requestsCopy()[0]
	requireAuthorizeRequest(t, request)
	if strings.Contains(string(request.Payload), sentinel) {
		t.Fatal("untrusted identity entered request")
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
	status := runWithHooks(nil, nil, io.Discard, &stderr, recorder, newIdentityFixture().hooks())
	if status == 0 || stderr.String() != deniedLine || time.Since(start) > authorityTimeout+time.Second {
		t.Fatalf("bounded denial failed: status=%d stderr=%q elapsed=%s", status, stderr.String(), time.Since(start))
	}
	select {
	case <-started:
	default:
		t.Fatal("transport was not attempted")
	}
}
