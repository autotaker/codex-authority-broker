//go:build linux

package ipc

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

type recordingBackend struct {
	mu       sync.Mutex
	calls    int
	response Response
	err      error
}

func (b *recordingBackend) Handle(context.Context, Request) (Response, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	return b.response, b.err
}

func (b *recordingBackend) callCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls
}

func TestServerUsesRealPeerCredentials(t *testing.T) {
	backend := &recordingBackend{response: Response{OK: true, Payload: []byte(`{"accepted":true}`)}}
	server, path, cancel := startTestServer(t, uint32(os.Geteuid()), backend, nil)
	response, err := exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady})
	if err != nil || !response.OK || backend.callCount() != 1 {
		t.Fatalf("authorized exchange response=%+v err=%v calls=%d", response, err, backend.callCount())
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket mode = %v, err = %v", info.Mode().Perm(), err)
	}
	cancel()
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := net.DialTimeout("unix", path, 100*time.Millisecond); err == nil {
		t.Fatal("closed server remained available")
	}
	if err := server.Close(); err != nil {
		t.Fatal("repeated close failed")
	}
}

func TestServerRejectsRealMismatchedUID(t *testing.T) {
	backend := &recordingBackend{}
	allowed := uint32(os.Geteuid()) + 1
	if allowed == uint32(os.Geteuid()) {
		allowed--
	}
	server, path, cancel := startTestServer(t, allowed, backend, nil)
	_, err := exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady})
	if err == nil || backend.callCount() != 0 {
		t.Fatalf("unauthorized exchange err=%v calls=%d", err, backend.callCount())
	}
	cancel()
	_ = server.Close()
}

func TestServerRejectsCredentialFailureAndBackendError(t *testing.T) {
	credentialFailure := func(*net.UnixConn) (uint32, error) { return 0, ErrServer }
	backend := &recordingBackend{}
	server, path, cancel := startTestServer(t, uint32(os.Geteuid()), backend, credentialFailure)
	if _, err := exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady}); err == nil {
		t.Fatal("credential failure returned a response")
	}
	if backend.callCount() != 0 {
		t.Fatal("credential failure called backend")
	}
	cancel()
	_ = server.Close()

	backend = &recordingBackend{err: errors.New("private backend detail")}
	server, path, cancel = startTestServer(t, uint32(os.Geteuid()), backend, nil)
	response, err := exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady})
	if err != nil || response.OK || backend.callCount() != 1 {
		t.Fatalf("backend failure response=%+v err=%v calls=%d", response, err, backend.callCount())
	}
	cancel()
	_ = server.Close()
}

func TestServerRejectsMalformedAndPartialWithoutBackend(t *testing.T) {
	backend := &recordingBackend{}
	server, path, cancel := startTestServer(t, uint32(os.Geteuid()), backend, nil)
	connection, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], 10)
	_, _ = connection.Write(append(header[:], []byte("{}")...))
	_ = connection.CloseWrite()
	_ = connection.SetReadDeadline(time.Now().Add(time.Second))
	response, err := readResponse(connection)
	_ = connection.Close()
	if err != nil || response.OK || backend.callCount() != 0 {
		t.Fatalf("partial response=%+v err=%v calls=%d", response, err, backend.callCount())
	}
	cancel()
	_ = server.Close()
}

func TestListenRejectsUnsafeAndExistingPaths(t *testing.T) {
	backend := &recordingBackend{}
	if _, err := Listen(Config{Path: "relative.sock"}, backend); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("relative path error = %v", err)
	}
	root := t.TempDir()
	unsafe := filepath.Join(root, "unsafe")
	if err := os.Mkdir(unsafe, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unsafe, 0o777); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(Config{Path: filepath.Join(unsafe, "server.sock")}, backend); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("unsafe parent error = %v", err)
	}
	realParent := filepath.Join(root, "real")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	linkParent := filepath.Join(root, "link")
	if err := os.Symlink(realParent, linkParent); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(Config{Path: filepath.Join(linkParent, "server.sock")}, backend); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("symlink parent error = %v", err)
	}
	existing := filepath.Join(realParent, "existing")
	if err := os.WriteFile(existing, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(Config{Path: existing}, backend); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("existing path error = %v", err)
	}
}

func TestProvisionedSocketUsesNumericOwnershipAndMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "provisioned.sock")
	access := &SocketAccess{OwnerUID: uint32(os.Geteuid()), GroupGID: uint32(os.Getegid())}
	backend := &recordingBackend{response: Response{OK: true}}
	server, err := Listen(Config{Path: path, AllowedUID: uint32(os.Geteuid()), Access: access}, backend)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode().Perm() != 0o660 {
		t.Fatal("provisioned socket mode mismatch")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != access.OwnerUID || stat.Gid != access.GroupGID {
		t.Fatal("provisioned numeric ownership mismatch")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = server.Serve(ctx) }()
	response, err := exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady})
	if err != nil || !response.OK || backend.callCount() != 1 {
		t.Fatal("provisioned authorized peer failed")
	}
	cancel()
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestProvisionedSocketStillRejectsMismatchedUID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "provisioned-denial.sock")
	access := &SocketAccess{OwnerUID: uint32(os.Geteuid()), GroupGID: uint32(os.Getegid())}
	backend := &recordingBackend{}
	allowedUID := uint32(os.Geteuid()) + 1
	if allowedUID == uint32(os.Geteuid()) {
		allowedUID--
	}
	server, err := Listen(Config{Path: path, AllowedUID: allowedUID, Access: access}, backend)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = server.Serve(ctx) }()
	if _, err := exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady}); err == nil {
		t.Fatal("mismatched UID received a response")
	}
	if backend.callCount() != 0 {
		t.Fatal("mismatched UID dispatched backend")
	}
	cancel()
	_ = server.Close()
}

func TestProvisioningFailureCleansOwnedSocket(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can provision arbitrary numeric ownership")
	}
	path := filepath.Join(t.TempDir(), "denied.sock")
	access := &SocketAccess{OwnerUID: uint32(os.Geteuid()) + 1, GroupGID: uint32(os.Getegid())}
	if _, err := Listen(Config{Path: path, AllowedUID: uint32(os.Geteuid()), Access: access}, &recordingBackend{}); !errors.Is(err, ErrLifecycle) {
		t.Fatalf("provisioning failure error = %v", err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("failed provisioning left socket path")
	}
}

type blockingBackend struct {
	mu      sync.Mutex
	calls   int
	entered chan struct{}
}

func (b *blockingBackend) Handle(ctx context.Context, _ Request) (Response, error) {
	b.mu.Lock()
	b.calls++
	if b.calls == 1 {
		close(b.entered)
	}
	b.mu.Unlock()
	<-ctx.Done()
	return Response{}, ctx.Err()
}

func (b *blockingBackend) callCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls
}

func TestCloseCancelsBlockingBackend(t *testing.T) {
	backend := &blockingBackend{entered: make(chan struct{})}
	server, path, _ := startTestServer(t, uint32(os.Geteuid()), backend, nil)
	clientDone := make(chan struct{})
	go func() {
		_, _ = exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady})
		close(clientDone)
	}()
	select {
	case <-backend.entered:
	case <-time.After(time.Second):
		t.Fatal("backend was not admitted")
	}
	closeDone := make(chan error, 1)
	go func() { closeDone <- server.Close() }()
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close blocked behind backend")
	}
	<-clientDone
	if backend.callCount() != 1 {
		t.Fatalf("backend calls = %d", backend.callCount())
	}
	_, _ = exchange(path, Request{Version: ProtocolVersion, Operation: OperationReady})
	if backend.callCount() != 1 {
		t.Fatal("request dispatched after shutdown")
	}
}

func TestCloseStopsActivePartialClient(t *testing.T) {
	backend := &recordingBackend{}
	server, path, _ := startTestServer(t, uint32(os.Geteuid()), backend, nil)
	connection, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Write(frameHeader(10)); err != nil {
		t.Fatal(err)
	}
	waitForHandlers(t, server, 1)
	closeDone := make(chan error, 1)
	go func() { closeDone <- server.Close() }()
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not stop active client")
	}
	_ = connection.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := connection.Read(make([]byte, 1)); err == nil {
		t.Fatal("active client remained open")
	}
	_ = connection.Close()
	if backend.callCount() != 0 {
		t.Fatal("partial client dispatched")
	}
}

func TestListenRefusesLiveSocketWithoutRemovingIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "live.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	listener.SetUnlinkOnClose(false)
	before, err := identifySocket(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(Config{Path: path, AllowedUID: uint32(os.Geteuid())}, &recordingBackend{}); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("live socket error = %v", err)
	}
	after, err := identifySocket(path)
	if err != nil || before != after {
		t.Fatal("live socket was replaced")
	}
	_ = listener.Close()
	_ = os.Remove(path)
}

func TestCloseLeavesReplacementUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.sock")
	server, err := Listen(Config{Path: path, AllowedUID: uint32(os.Geteuid())}, &recordingBackend{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	replacement := []byte("replacement")
	if err := os.WriteFile(path, replacement, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := server.Close(); !errors.Is(err, ErrLifecycle) {
		t.Fatalf("replacement close error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != string(replacement) {
		t.Fatal("replacement was altered")
	}
}

func TestCloseReportsMissingOwnedSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.sock")
	server, err := Listen(Config{Path: path, AllowedUID: uint32(os.Geteuid())}, &recordingBackend{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := server.Close(); !errors.Is(err, ErrLifecycle) {
		t.Fatalf("missing socket close error = %v", err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("missing socket path was recreated")
	}
}

func TestIdentityCheckedRemovalLeavesReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "created.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	listener.SetUnlinkOnClose(false)
	identity, err := identifySocket(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	replacement := []byte("post-bind replacement")
	if err := os.WriteFile(path, replacement, 0o600); err != nil {
		t.Fatal(err)
	}
	_ = listener.Close()
	if err := removeOwnedSocket(path, identity); !errors.Is(err, ErrLifecycle) {
		t.Fatalf("identity cleanup error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != string(replacement) {
		t.Fatal("identity cleanup altered replacement")
	}
}

func TestHandlerLimitRejectsSeventeenthClient(t *testing.T) {
	backend := &recordingBackend{}
	server, path, _ := startTestServer(t, uint32(os.Geteuid()), backend, nil)
	clients := make([]*net.UnixConn, 0, maxHandlers)
	for i := 0; i < maxHandlers; i++ {
		connection, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
		if err != nil {
			t.Fatal(err)
		}
		clients = append(clients, connection)
	}
	waitForHandlers(t, server, maxHandlers)
	excess, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	_ = excess.SetDeadline(time.Now().Add(time.Second))
	_ = writeRequest(excess, Request{Version: ProtocolVersion, Operation: OperationReady})
	if _, err := readResponse(excess); err == nil {
		t.Fatal("seventeenth client received a response")
	}
	_ = excess.Close()
	if backend.callCount() != 0 {
		t.Fatal("saturated server dispatched backend")
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	for _, connection := range clients {
		_ = connection.Close()
	}
}

func waitForHandlers(t *testing.T, server *Server, expected int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for len(server.handlers) != expected {
		if time.Now().After(deadline) {
			t.Fatalf("active handlers = %d, want %d", len(server.handlers), expected)
		}
		time.Sleep(time.Millisecond)
	}
}

func startTestServer(t *testing.T, uid uint32, backend Backend, credential credentialReader) (*Server, string, context.CancelFunc) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "server.sock")
	var server *Server
	var err error
	if credential == nil {
		server, err = Listen(Config{Path: path, AllowedUID: uid}, backend)
	} else {
		server, err = listen(Config{Path: path, AllowedUID: uid}, backend, credential)
	}
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = server.Serve(ctx) }()
	return server, path, cancel
}

func exchange(path string, request Request) (Response, error) {
	connection, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return Response{}, err
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(time.Second))
	if err := writeRequest(connection, request); err != nil {
		return Response{}, err
	}
	return readResponse(connection)
}
