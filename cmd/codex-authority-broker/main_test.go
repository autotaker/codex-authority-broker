//go:build linux

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/autotaker/codex-authority-broker/internal/backend"
	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

const testSeed = `{"totp_secret_b64":"AQIDBAUGBwgJCgsMDQ4PEA==","allowed_uid":1000}`

type trackingReader struct {
	data       []byte
	err        error
	closeErr   error
	closeCount int
}

func (r *trackingReader) Read(p []byte) (int, error) {
	if len(r.data) != 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	if r.err != nil {
		err := r.err
		r.err = nil
		return 0, err
	}
	return 0, io.EOF
}

func (r *trackingReader) Close() error {
	r.closeCount++
	return r.closeErr
}

type descriptorFixture struct {
	mu       sync.Mutex
	opens    []string
	flags    []int
	stats    map[int]descriptorStat
	statErr  map[int]error
	openErr  map[string]error
	closeErr map[int]error
	closed   []int
	reader   io.ReadCloser
	fileErr  error
}

func newDescriptorFixture(document string) *descriptorFixture {
	return &descriptorFixture{
		stats: map[int]descriptorStat{
			10: {mode: syscall.S_IFDIR | 0o755},
			11: {mode: syscall.S_IFDIR | 0o755},
			12: {mode: syscall.S_IFDIR | 0o755},
			13: {mode: syscall.S_IFREG | 0o600, uid: 0, size: int64(len(document))},
		},
		statErr:  map[int]error{},
		openErr:  map[string]error{},
		closeErr: map[int]error{},
		reader:   &trackingReader{data: []byte(document)},
	}
}

func (f *descriptorFixture) openRoot(flags int) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.opens = append(f.opens, "/")
	f.flags = append(f.flags, flags)
	if err := f.openErr["/"]; err != nil {
		return -1, err
	}
	return 10, nil
}

func (f *descriptorFixture) openAt(dirfd int, name string, flags int) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fmt.Sprintf("%d/%s", dirfd, name)
	f.opens = append(f.opens, key)
	f.flags = append(f.flags, flags)
	if err := f.openErr[key]; err != nil {
		return -1, err
	}
	switch name {
	case "etc":
		return 11, nil
	case "codex-authority":
		return 12, nil
	case "seed.json":
		return 13, nil
	default:
		return -1, errors.New("unexpected component")
	}
}

func (f *descriptorFixture) stat(fd int) (descriptorStat, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.statErr[fd]; err != nil {
		return descriptorStat{}, err
	}
	value, ok := f.stats[fd]
	if !ok {
		return descriptorStat{}, errors.New("unknown descriptor")
	}
	return value, nil
}

func (f *descriptorFixture) close(fd int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = append(f.closed, fd)
	return f.closeErr[fd]
}

func (f *descriptorFixture) file(int) (io.ReadCloser, error) {
	if f.fileErr != nil {
		return nil, f.fileErr
	}
	return f.reader, nil
}

func (f *descriptorFixture) closeSet() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int(nil), f.closed...)
}

func (f *descriptorFixture) asDescriptors() seedDescriptors { return f }

func TestLoadSeedClosesParentsAndReaderOwnsFinalDescriptor(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	secret, uid, err := loadSeed(fixture)
	if err != nil || uid != 1000 || !bytes.Equal(secret, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}) {
		t.Fatalf("loadSeed() rejected valid fixture: uid %d err %v", uid, err)
	}
	reader := fixture.reader.(*trackingReader)
	if reader.closeCount != 1 {
		t.Fatalf("reader close count = %d", reader.closeCount)
	}
	if got := fixture.closeSet(); !reflect.DeepEqual(got, []int{10, 11, 12}) {
		t.Fatalf("descriptor closes = %v", got)
	}
}

func TestLoadSeedFileWrapperFailureClosesFinalDescriptor(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	fixture.fileErr = errors.New("wrapper failure")
	if _, _, err := loadSeed(fixture); err != errSeed {
		t.Fatalf("error = %v, want errSeed", err)
	}
	if got := fixture.closeSet(); !reflect.DeepEqual(got, []int{10, 11, 12, 13}) {
		t.Fatalf("descriptor closes = %v", got)
	}
}

func TestOpenSeedUsesNoFollowCloexecAndClosesOnFailures(t *testing.T) {
	baseFlags := syscallFlags()
	tests := []struct {
		name    string
		mutate  func(*descriptorFixture)
		wantErr bool
		closes  []int
	}{
		{name: "valid", closes: []int{10, 11, 12}},
		{name: "root stat", mutate: func(f *descriptorFixture) { f.statErr[10] = errors.New("stat") }, wantErr: true, closes: []int{10}},
		{name: "parent open", mutate: func(f *descriptorFixture) { f.openErr["10/etc"] = errors.New("open") }, wantErr: true, closes: []int{10}},
		{name: "parent stat", mutate: func(f *descriptorFixture) { f.statErr[11] = errors.New("stat") }, wantErr: true, closes: []int{11, 10}},
		{name: "parent symlink", mutate: func(f *descriptorFixture) { f.stats[11] = descriptorStat{mode: syscall.S_IFLNK} }, wantErr: true, closes: []int{10, 11}},
		{name: "final open", mutate: func(f *descriptorFixture) { f.openErr["12/seed.json"] = errors.New("open") }, wantErr: true, closes: []int{10, 11, 12}},
		{name: "final stat", mutate: func(f *descriptorFixture) { f.statErr[13] = errors.New("stat") }, wantErr: true, closes: []int{10, 11, 13, 12}},
		{name: "final symlink", mutate: func(f *descriptorFixture) { f.stats[13] = descriptorStat{mode: syscall.S_IFLNK, uid: 0, size: 10} }, wantErr: true, closes: []int{10, 11, 12, 13}},
		{name: "final directory", mutate: func(f *descriptorFixture) { f.stats[13] = descriptorStat{mode: syscall.S_IFDIR, uid: 0, size: 10} }, wantErr: true, closes: []int{10, 11, 12, 13}},
		{name: "wrong owner", mutate: func(f *descriptorFixture) {
			f.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o600, uid: 1000, size: 10}
		}, wantErr: true, closes: []int{10, 11, 12, 13}},
		{name: "special mode", mutate: func(f *descriptorFixture) {
			f.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o4600, uid: 0, size: 10}
		}, wantErr: true, closes: []int{10, 11, 12, 13}},
		{name: "wrong permissions", mutate: func(f *descriptorFixture) {
			f.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o644, uid: 0, size: 10}
		}, wantErr: true, closes: []int{10, 11, 12, 13}},
		{name: "empty", mutate: func(f *descriptorFixture) {
			f.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o600, uid: 0, size: 0}
		}, wantErr: true, closes: []int{10, 11, 12, 13}},
		{name: "oversized", mutate: func(f *descriptorFixture) {
			f.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o600, uid: 0, size: maxSeedBytes + 1}
		}, wantErr: true, closes: []int{10, 11, 12, 13}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newDescriptorFixture(testSeed)
			if test.mutate != nil {
				test.mutate(fixture)
			}
			fd, size, err := openSeed(fixture)
			if test.wantErr {
				if err != errSeed || fd != -1 || size != 0 {
					t.Fatalf("openSeed() = %d, %d, %v", fd, size, err)
				}
			} else if err != nil || fd != 13 || size != int64(len(testSeed)) {
				t.Fatalf("openSeed() = %d, %d, %v", fd, size, err)
			}
			if got := fixture.closeSet(); !reflect.DeepEqual(got, test.closes) {
				t.Fatalf("descriptor closes = %v, want %v", got, test.closes)
			}
			for _, flags := range fixture.flags {
				if flags&baseFlags != baseFlags {
					t.Fatalf("flags %#x omit required %#x", flags, baseFlags)
				}
			}
		})
	}
}

func syscallFlags() int { return syscall.O_NOFOLLOW | syscall.O_CLOEXEC | syscall.O_RDONLY }

func TestSeedComponentsBoundsAndNames(t *testing.T) {
	longName := strings.Repeat("x", maxComponentBytes+1)
	deep := "/" + strings.TrimSuffix(strings.Repeat("x/", maxSeedComponents+1), "/")
	longPath := "/" + strings.Repeat("x", maxSeedPathBytes)
	for _, test := range []struct {
		path string
		ok   bool
	}{
		{path: seedPath, ok: true},
		{path: "", ok: false}, {path: "/", ok: false}, {path: "/etc//seed", ok: false},
		{path: "/etc/./seed", ok: false}, {path: "/etc/../seed", ok: false}, {path: "/etc/a\\seed", ok: false},
		{path: "/" + longName, ok: false}, {path: deep, ok: false}, {path: longPath, ok: false},
	} {
		components, ok := seedComponents(test.path)
		if ok != test.ok || (ok && len(components) == 0) {
			t.Errorf("seedComponents(%q) = %#v, %v; want ok %v", test.path, components, ok, test.ok)
		}
	}
}

func TestReadSeedRejectsBoundsShortAndErrors(t *testing.T) {
	readErr := errors.New("read failure")
	for _, test := range []struct {
		name string
		data string
		size int64
		err  error
	}{
		{name: "zero", size: 0}, {name: "oversized", size: maxSeedBytes + 1},
		{name: "short", data: "abc", size: 4}, {name: "extra", data: strings.Repeat("x", maxSeedBytes+1), size: maxSeedBytes + 1},
		{name: "reader error", data: "abc", size: 3, err: readErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader := &trackingReader{data: []byte(test.data), err: test.err}
			data, err := readSeed(reader, test.size)
			if err != errSeed || data == nil && test.name == "reader error" {
				t.Fatalf("readSeed() rejected expected error: bytes %d err %v", len(data), err)
			}
		})
	}
}

func TestParseSeedStrictSchema(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
	valid := fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1}`, secret)
	for _, test := range []struct {
		name string
		doc  string
		ok   bool
	}{
		{name: "valid", doc: valid, ok: true},
		{name: "duplicate secret", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"totp_secret_b64":%q,"allowed_uid":1}`, secret, secret)},
		{name: "duplicate uid", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1,"allowed_uid":2}`, secret)},
		{name: "unknown", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1,"x":1}`, secret)},
		{name: "missing", doc: fmt.Sprintf(`{"totp_secret_b64":%q}`, secret)},
		{name: "wrong secret type", doc: `{"totp_secret_b64":1,"allowed_uid":1}`},
		{name: "wrong uid type", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":"1"}`, secret)},
		{name: "fractional uid", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1.0}`, secret)},
		{name: "negative uid", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":-1}`, secret)},
		{name: "zero uid", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":0}`, secret)},
		{name: "overflow uid", doc: fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":4294967296}`, secret)},
		{name: "noncanonical base64", doc: `{"totp_secret_b64":"AQI= ","allowed_uid":1}`},
		{name: "invalid base64", doc: `{"totp_secret_b64":"!!!","allowed_uid":1}`},
		{name: "empty base64", doc: `{"totp_secret_b64":"","allowed_uid":1}`},
		{name: "trailing", doc: valid + ` {}`},
		{name: "malformed", doc: valid[:len(valid)-1]},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, uid, err := parseSeed([]byte(test.doc))
			if test.ok {
				if err != nil || uid != 1 || !bytes.Equal(got, []byte{1, 2, 3}) {
					t.Fatalf("parseSeed() rejected valid schema: uid %d err %v", uid, err)
				}
				return
			}
			if err != errSeed || got != nil || uid != 0 || strings.Contains(err.Error(), secret) {
				t.Fatalf("parseSeed() accepted rejected schema: uid %d err %v", uid, err)
			}
		})
	}
	if _, _, err := parseSeed([]byte(strings.Repeat("x", maxSeedBytes+1))); err != errSeed {
		t.Fatalf("oversized parse error = %v", err)
	}
}

type fakeRuntime struct {
	mu       sync.Mutex
	closed   int
	secret   []byte
	closeLog *[]string
}

func (r *fakeRuntime) Handle(context.Context, ipc.Request) (ipc.Response, error) {
	return ipc.Response{OK: true}, nil
}

func (r *fakeRuntime) Close() {
	r.mu.Lock()
	r.closed++
	if r.closeLog != nil {
		*r.closeLog = append(*r.closeLog, "runtime")
	}
	r.mu.Unlock()
}

type fakeServer struct {
	mu         sync.Mutex
	serve      func(context.Context) error
	serveCalls int
	closed     int
	closeErr   error
	closeLog   *[]string
}

func (s *fakeServer) Serve(ctx context.Context) error {
	s.mu.Lock()
	s.serveCalls++
	serve := s.serve
	s.mu.Unlock()
	if serve != nil {
		return serve(ctx)
	}
	return nil
}

func (s *fakeServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed++
	if s.closeLog != nil {
		*s.closeLog = append(*s.closeLog, "server")
	}
	return s.closeErr
}

func runFixture(t *testing.T, server *fakeServer, runtime *fakeRuntime, listenErr error, serve func(context.Context) error) (int, *descriptorFixture) {
	t.Helper()
	fixture := newDescriptorFixture(testSeed)
	server.serve = serve
	status := run(context.Background(), fixture, func(secret []byte) (brokerRuntime, error) {
		runtime.secret = append([]byte(nil), secret...)
		return runtime, nil
	}, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		return server, listenErr
	})
	return status, fixture
}

func TestRunOrderingFailureAndCancellation(t *testing.T) {
	var order []string
	runtime := &fakeRuntime{closeLog: &order}
	server := &fakeServer{closeLog: &order}
	status, _ := runFixture(t, server, runtime, nil, func(ctx context.Context) error {
		order = append(order, "serve")
		return nil
	})
	if status != 1 || server.serveCalls != 1 || runtime.closed != 1 || server.closed != 1 {
		t.Fatalf("uncancelled Serve status=%d runtime=%d server=%d calls=%d", status, runtime.closed, server.closed, server.serveCalls)
	}
	if !reflect.DeepEqual(order, []string{"serve", "runtime", "server"}) {
		t.Fatalf("close order = %v", order)
	}

	order = nil
	runtime = &fakeRuntime{closeLog: &order}
	server = &fakeServer{closeLog: &order}
	ctx, cancel := context.WithCancel(context.Background())
	fixture := newDescriptorFixture(testSeed)
	status = run(ctx, fixture, func([]byte) (brokerRuntime, error) { return runtime, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		return server, nil
	})
	_ = cancel
	if status != 1 {
		t.Fatal("Serve returned without cancellation but unexpectedly passed")
	}

	order = nil
	runtime = &fakeRuntime{closeLog: &order}
	ctx, cancel = context.WithCancel(context.Background())
	server = &fakeServer{closeLog: &order, serve: func(context.Context) error {
		cancel()
		return nil
	}}
	status = run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return runtime, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		return server, nil
	})
	if status != 0 || runtime.closed != 1 || server.closed != 1 {
		t.Fatalf("cancelled Serve status=%d runtime=%d server=%d", status, runtime.closed, server.closed)
	}
}

func TestRunNilDependenciesAndListenReturnedServer(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	if status := run(context.Background(), fixture, nil, nil); status != 1 {
		t.Fatal("nil dependencies passed")
	}
	if status := run(context.Background(), fixture, func([]byte) (brokerRuntime, error) { return nil, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { t.Fatal("listen called"); return nil, nil }); status != 1 {
		t.Fatal("nil runtime passed")
	}
	order := []string{}
	runtime := &fakeRuntime{closeLog: &order}
	server := &fakeServer{closeLog: &order}
	status := run(context.Background(), newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return runtime, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		return server, errors.New("listen")
	})
	if status != 1 || server.closed != 1 || runtime.closed != 1 || !reflect.DeepEqual(order, []string{"server", "runtime"}) {
		t.Fatalf("listen failure status=%d server=%d runtime=%d order=%v", status, server.closed, runtime.closed, order)
	}
}

func TestRunDeniesSeedBeforeConstructionAndRestartReadsFreshState(t *testing.T) {
	bad := newDescriptorFixture(testSeed)
	bad.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o644, uid: 0, size: int64(len(testSeed))}
	listenCalls := 0
	if status := run(context.Background(), bad, func([]byte) (brokerRuntime, error) { t.Fatal("runtime called"); return nil, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		listenCalls++
		return nil, nil
	}); status != 1 || listenCalls != 0 {
		t.Fatalf("bad seed status=%d listenCalls=%d", status, listenCalls)
	}
	first := newDescriptorFixture(testSeed)
	second := newDescriptorFixture(testSeed)
	var runtimes []*fakeRuntime
	makeRuntime := func(secret []byte) (brokerRuntime, error) {
		runtime := &fakeRuntime{secret: append([]byte(nil), secret...)}
		runtimes = append(runtimes, runtime)
		return runtime, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	server := &fakeServer{serve: func(context.Context) error { cancel(); return nil }}
	if status := run(ctx, first, makeRuntime, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 {
		t.Fatal("first restart fixture failed")
	}
	ctx, cancel = context.WithCancel(context.Background())
	server = &fakeServer{serve: func(context.Context) error { cancel(); return nil }}
	if status := run(ctx, second, makeRuntime, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 {
		t.Fatal("second restart fixture failed")
	}
	if len(runtimes) != 2 || runtimes[0] == runtimes[1] || !bytes.Equal(runtimes[0].secret, runtimes[1].secret) {
		t.Fatalf("restart runtimes = %d", len(runtimes))
	}
}

func TestExistingClientProtocolAgainstRuntimeServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "authority.sock")
	runtime, err := backend.New([]byte("01234567890123456789"))
	if err != nil {
		t.Fatal(err)
	}
	server, err := ipc.Listen(ipc.Config{Path: path, AllowedUID: uint32(os.Geteuid())}, runtime)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "server unavailable") {
			t.Skipf("Unix sockets unavailable: %v", err)
		}
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer runtime.Close()
	go func() { _ = server.Serve(ctx) }()
	client := ipc.Client{Path: path}
	response, err := client.Call(context.Background(), ipc.Request{Version: ipc.ProtocolVersion, Operation: ipc.OperationReady})
	if err != nil || !response.OK {
		t.Fatalf("ready = %#v, %v", response, err)
	}
	response, err = client.Call(context.Background(), ipc.Request{Version: ipc.ProtocolVersion, Operation: "unknown"})
	if err == nil && response.OK {
		t.Fatal("unknown operation accepted")
	}
	if err := server.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		t.Fatal(err)
	}
}

type integrationBackend struct{}

func (integrationBackend) Handle(_ context.Context, request ipc.Request) (ipc.Response, error) {
	if request.Operation == ipc.OperationReady && len(request.Payload) == 0 {
		return ipc.Response{OK: true}, nil
	}
	valid := request.Operation == ipc.OperationOTP && len(request.Payload) == 17 &&
		string(request.Payload[:9]) == `{"code":"` && string(request.Payload[15:]) == `"}`
	if valid {
		for _, digit := range request.Payload[9:15] {
			if digit < '0' || digit > '9' {
				valid = false
			}
		}
	}
	return ipc.Response{OK: valid}, nil
}

func exerciseClientProtocol(t *testing.T, validOTP bool) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "authority.sock")
	server, err := ipc.Listen(ipc.Config{Path: path, AllowedUID: uint32(os.Geteuid())}, integrationBackend{})
	if err != nil {
		if errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "server unavailable") {
			t.Skipf("Unix sockets unavailable: %v", err)
		}
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx) }()
	client := ipc.Client{Path: path}
	if response, err := client.Call(context.Background(), ipc.Request{Version: ipc.ProtocolVersion, Operation: ipc.OperationReady}); err != nil || !response.OK {
		t.Fatalf("ready request failed: ok=%v err=%v", response.OK, err)
	}
	if validOTP {
		payload := []byte(`{"code":"123456"}`)
		response, err := client.Call(context.Background(), ipc.Request{Version: ipc.ProtocolVersion, Operation: ipc.OperationOTP, Payload: payload})
		if err != nil || !response.OK {
			t.Fatalf("valid OTP request failed: err=%v ok=%v", err, response.OK)
		}
	} else {
		response, err := client.Call(context.Background(), ipc.Request{Version: ipc.ProtocolVersion, Operation: ipc.OperationOTP, Payload: []byte(`{"code":"bad"}`)})
		if err != nil || response.OK {
			t.Fatalf("malformed OTP request result: err=%v ok=%v", err, response.OK)
		}
	}
	if err := server.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		t.Fatal(err)
	}
}

func TestRunRedactsSeedErrors(t *testing.T) {
	marker := "private-secret-marker"
	fixture := newDescriptorFixture(fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1}`, base64.StdEncoding.EncodeToString([]byte(marker))))
	fixture.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o644, uid: 0, size: int64(len(testSeed))}
	if status := run(context.Background(), fixture, func([]byte) (brokerRuntime, error) { return nil, errors.New(marker) }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return nil, errors.New(marker) }); status != 1 {
		t.Fatal("redaction fixture passed")
	}
	if strings.Contains(errSeed.Error(), marker) {
		t.Fatal("sentinel contains secret marker")
	}
}

func TestConcurrentFakeCloseIsSafeToObserve(t *testing.T) {
	server := &fakeServer{}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = server.Close()
		}()
	}
	wg.Wait()
	if server.closed != 8 {
		t.Fatalf("fake close observations = %d", server.closed)
	}
}

func allZero(data []byte) bool {
	for _, value := range data {
		if value != 0 {
			return false
		}
	}
	return true
}

func TestRunWipesCallerSecretBeforeListen(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	var input []byte
	ctx, cancel := context.WithCancel(context.Background())
	runtime := &fakeRuntime{}
	server := &fakeServer{serve: func(context.Context) error { cancel(); return nil }}
	listenSeen := make(chan struct{})
	status := run(ctx, fixture, func(secret []byte) (brokerRuntime, error) {
		input = secret
		return runtime, nil
	}, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		if !allZero(input) {
			t.Fatal("caller-owned secret remained live at listen")
		}
		close(listenSeen)
		return server, nil
	})
	if status != 0 {
		t.Fatalf("wiping run status = %d", status)
	}
	select {
	case <-listenSeen:
	default:
		t.Fatal("listen barrier was not observed")
	}
	var factoryInput []byte
	if status := run(context.Background(), newDescriptorFixture(testSeed), func(secret []byte) (brokerRuntime, error) {
		factoryInput = secret
		return nil, errors.New("factory failure")
	}, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		t.Fatal("listen called after factory failure")
		return nil, nil
	}); status != 1 || !allZero(factoryInput) {
		t.Fatal("factory-error path did not wipe caller buffer")
	}
}

func assertDescriptorError(t *testing.T, mutate func(*descriptorFixture)) {
	t.Helper()
	fixture := newDescriptorFixture(testSeed)
	mutate(fixture)
	if _, _, err := loadSeed(fixture); err != errSeed {
		t.Fatalf("loadSeed error = %v, want generic seed error", err)
	}
}

func TestLoadSeedRootOpenError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.openErr["/"] = errors.New("open") })
}
func TestLoadSeedRootStatError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.statErr[10] = errors.New("stat") })
}
func TestLoadSeedRootCloseError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.closeErr[10] = errors.New("close") })
}
func TestLoadSeedParentOpenError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.openErr["10/etc"] = errors.New("open") })
}
func TestLoadSeedFinalOpenError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.openErr["12/seed.json"] = errors.New("open") })
}
func TestLoadSeedParentStatError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.statErr[11] = errors.New("stat") })
}
func TestLoadSeedFinalStatError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.statErr[13] = errors.New("stat") })
}
func TestLoadSeedParentCloseError(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.closeErr[11] = errors.New("close") })
}
func TestLoadSeedFinalDescriptorCloseError(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	fixture.fileErr = errors.New("wrapper")
	fixture.closeErr[13] = errors.New("close")
	if _, _, err := loadSeed(fixture); err != errSeed {
		t.Fatalf("loadSeed error = %v, want generic seed error", err)
	}
}
func TestLoadSeedReaderCloseError(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	fixture.reader.(*trackingReader).closeErr = errors.New("close")
	if _, _, err := loadSeed(fixture); err != errSeed {
		t.Fatalf("loadSeed error = %v, want generic seed error", err)
	}
}
func TestLoadSeedRejectsParentSymlink(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.stats[11] = descriptorStat{mode: syscall.S_IFLNK} })
}
func TestLoadSeedRejectsFinalSymlink(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.stats[13] = descriptorStat{mode: syscall.S_IFLNK, uid: 0, size: 10} })
}
func TestLoadSeedRejectsNonRootOwner(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) {
		f.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o600, uid: 1000, size: 10}
	})
}
func TestLoadSeedRejectsNon0600Mode(t *testing.T) {
	for mode := uint32(0); mode <= 0o7777; mode++ {
		if mode == 0o600 {
			continue
		}
		if validSeedFile(descriptorStat{mode: syscall.S_IFREG | mode, uid: 0, size: 1}) {
			t.Fatalf("mode %#o accepted", mode)
		}
	}
}
func TestLoadSeedRejectsNonRegularFile(t *testing.T) {
	assertDescriptorError(t, func(f *descriptorFixture) { f.stats[13] = descriptorStat{mode: syscall.S_IFDIR, uid: 0, size: 10} })
}
func TestLoadSeedShortRead(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	fixture.reader.(*trackingReader).data = []byte("short")
	fixture.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o600, uid: 0, size: int64(len(testSeed))}
	if _, _, err := loadSeed(fixture); err != errSeed {
		t.Fatalf("short read error = %v", err)
	}
}
func TestLoadSeedReadError(t *testing.T) {
	fixture := newDescriptorFixture(testSeed)
	fixture.reader.(*trackingReader).err = errors.New("read")
	if _, _, err := loadSeed(fixture); err != errSeed {
		t.Fatalf("read error = %v", err)
	}
}
func TestLoadSeedSizeBounds(t *testing.T) {
	for _, size := range []int64{0, 1, maxSeedBytes, maxSeedBytes + 1} {
		fixture := newDescriptorFixture(testSeed)
		fixture.stats[13] = descriptorStat{mode: syscall.S_IFREG | 0o600, uid: 0, size: size}
		fd, got, err := openSeed(fixture)
		if size == 0 || size > maxSeedBytes {
			if err != errSeed || fd != -1 || got != 0 {
				t.Fatalf("size %d accepted: fd=%d size=%d err=%v", size, fd, got, err)
			}
		} else if err != nil || fd != 13 || got != size {
			t.Fatalf("size %d rejected: fd=%d got=%d err=%v", size, fd, got, err)
		}
	}
}

func maximumSeedDocument() []byte {
	secret := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, maxSecretBytes))
	return []byte(fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1}`, secret))
}
func TestLoadSeedAcceptsValidSchema(t *testing.T) {
	secret, uid, err := parseSeed([]byte(testSeed))
	if err != nil || uid != 1000 || len(secret) == 0 {
		t.Fatalf("valid schema rejected: uid=%d err=%v", uid, err)
	}
}
func TestLoadSeedAcceptsMaximumSecret(t *testing.T) {
	secret, uid, err := parseSeed(maximumSeedDocument())
	if err != nil || uid != 1 || len(secret) != maxSecretBytes {
		t.Fatalf("maximum schema rejected: uid=%d length=%d err=%v", uid, len(secret), err)
	}
}
func assertSchemaError(t *testing.T, document []byte) {
	t.Helper()
	if got, uid, err := parseSeed(document); err != errSeed || got != nil || uid != 0 {
		t.Fatalf("schema error = uid=%d err=%v", uid, err)
	}
}
func schemaBase64() string { return base64.StdEncoding.EncodeToString([]byte{1, 2, 3}) }
func TestLoadSeedRejectsMalformedSchema(t *testing.T) {
	assertSchemaError(t, []byte(`{"totp_secret_b64":`))
}
func TestLoadSeedRejectsDuplicateSchemaField(t *testing.T) {
	s := schemaBase64()
	assertSchemaError(t, []byte(fmt.Sprintf(`{"totp_secret_b64":%q,"totp_secret_b64":%q,"allowed_uid":1}`, s, s)))
}
func TestLoadSeedRejectsUnknownSchemaField(t *testing.T) {
	s := schemaBase64()
	assertSchemaError(t, []byte(fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1,"unknown":true}`, s)))
}
func TestLoadSeedRejectsMissingSchemaField(t *testing.T) {
	assertSchemaError(t, []byte(`{"allowed_uid":1}`))
}
func TestLoadSeedRejectsEmptySecret(t *testing.T) {
	assertSchemaError(t, []byte(`{"totp_secret_b64":"","allowed_uid":1}`))
}
func TestLoadSeedRejectsWrongSchemaType(t *testing.T) {
	assertSchemaError(t, []byte(`{"totp_secret_b64":1,"allowed_uid":1}`))
}
func TestLoadSeedRejectsTrailingJSON(t *testing.T) {
	assertSchemaError(t, append([]byte(testSeed), []byte(` {}`)...))
}
func TestLoadSeedRejectsInvalidAllowedUID(t *testing.T) {
	assertSchemaError(t, []byte(`{"totp_secret_b64":"AQI=","allowed_uid":0}`))
}
func TestLoadSeedRejectsInvalidBase64(t *testing.T) {
	assertSchemaError(t, []byte(`{"totp_secret_b64":"!!!","allowed_uid":1}`))
}
func TestLoadSeedRejectsNonCanonicalBase64(t *testing.T) {
	assertSchemaError(t, []byte(`{"totp_secret_b64":"AQI= ","allowed_uid":1}`))
}
func TestLoadSeedRejectsOversizedSchemaInput(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{8}, maxSecretBytes+1))
	assertSchemaError(t, []byte(fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1}`, secret)))
}

func TestRunRuntimeFactoryError(t *testing.T) {
	called := false
	if status := run(context.Background(), newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { called = true; return nil, errors.New("factory") }, func(ipc.Config, ipc.Backend) (brokerServer, error) { t.Fatal("listen called"); return nil, nil }); status != 1 || !called {
		t.Fatal("runtime factory error did not fail closed")
	}
}
func TestRunConstructsRuntimeBeforeListen(t *testing.T) {
	order := []string{}
	ctx, cancel := context.WithCancel(context.Background())
	status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { order = append(order, "runtime"); return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		order = append(order, "listen")
		return &fakeServer{serve: func(context.Context) error { cancel(); return nil }}, nil
	})
	if status != 0 || !reflect.DeepEqual(order, []string{"runtime", "listen"}) {
		t.Fatalf("construction order=%v status=%d", order, status)
	}
}
func TestRunConfiguresServerBeforeListen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(config ipc.Config, _ ipc.Backend) (brokerServer, error) {
		if config.Path != socketPath || config.AllowedUID != 1000 {
			t.Fatalf("unexpected listener config")
		}
		return &fakeServer{serve: func(context.Context) error { cancel(); return nil }}, nil
	})
	if status != 0 {
		t.Fatalf("configured listener status=%d", status)
	}
}
func TestRunClosesServerOnListenError(t *testing.T) {
	order := []string{}
	runtime := &fakeRuntime{closeLog: &order}
	server := &fakeServer{closeLog: &order}
	status := run(context.Background(), newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return runtime, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) {
		return server, errors.New("listen")
	})
	if status != 1 || server.closed != 1 || runtime.closed != 1 || !reflect.DeepEqual(order, []string{"server", "runtime"}) {
		t.Fatalf("listen failure status=%d server=%d runtime=%d order=%v", status, server.closed, runtime.closed, order)
	}
}
func TestRunRejectsUnexpectedServeReturn(t *testing.T) {
	server := &fakeServer{serve: func(context.Context) error { return nil }}
	if status := run(context.Background(), newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 1 {
		t.Fatal("unexpected Serve return was accepted")
	}
}
func TestRunReportsServerCloseError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := &fakeServer{closeErr: errors.New("close"), serve: func(context.Context) error { cancel(); return nil }}
	if status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 1 {
		t.Fatal("server close error was reported clean")
	}
}

func runCancelled(t *testing.T, sig os.Signal) {
	t.Helper()
	ctx, stop := signal.NotifyContext(context.Background(), sig)
	defer stop()
	server := &fakeServer{serve: func(ctx context.Context) error {
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatalf("find process: %v", err)
		}
		if err := process.Signal(sig); err != nil {
			t.Fatalf("send signal: %v", err)
		}
		<-ctx.Done()
		return nil
	}}
	if status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 {
		t.Fatalf("signal %v status=%d", sig, status)
	}
}
func TestRunShutsDownOnSIGINT(t *testing.T)  { runCancelled(t, os.Interrupt) }
func TestRunShutsDownOnSIGTERM(t *testing.T) { runCancelled(t, syscall.SIGTERM) }
func TestRunShutdownIsIdempotent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := &fakeServer{serve: func(context.Context) error { cancel(); cancel(); return nil }}
	runtime := &fakeRuntime{}
	if status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return runtime, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 || runtime.closed != 1 || server.closed != 1 {
		t.Fatalf("shutdown counts runtime=%d server=%d status=%d", runtime.closed, server.closed, status)
	}
}
func TestRunHandlesConcurrentShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := &fakeServer{serve: func(context.Context) error {
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); cancel() }()
		}
		wg.Wait()
		return nil
	}}
	if status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 {
		t.Fatalf("concurrent shutdown status=%d", status)
	}
}
func TestRunWaitsForActiveClientOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started, release := make(chan struct{}), make(chan struct{})
	server := &fakeServer{serve: func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		<-release
		return nil
	}}
	done := make(chan int, 1)
	go func() {
		done <- run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil })
	}()
	<-started
	cancel()
	select {
	case <-done:
		t.Fatal("run returned while active client barrier held")
	default:
	}
	close(release)
	if status := <-done; status != 0 {
		t.Fatalf("active shutdown status=%d", status)
	}
}
func TestRunRestartsWithFreshSeed(t *testing.T) {
	secondDocument := fmt.Sprintf(`{"totp_secret_b64":%q,"allowed_uid":1000}`, base64.StdEncoding.EncodeToString([]byte{16, 15, 14, 13, 12, 11, 10, 9}))
	first, second := newDescriptorFixture(testSeed), newDescriptorFixture(secondDocument)
	var runtimes []*fakeRuntime
	var servers []*fakeServer
	makeRuntime := func(secret []byte) (brokerRuntime, error) {
		runtime := &fakeRuntime{secret: append([]byte(nil), secret...)}
		runtimes = append(runtimes, runtime)
		return runtime, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	server := &fakeServer{serve: func(context.Context) error { cancel(); return nil }}
	servers = append(servers, server)
	if status := run(ctx, first, makeRuntime, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 {
		t.Fatal("first restart failed")
	}
	ctx, cancel = context.WithCancel(context.Background())
	server = &fakeServer{serve: func(context.Context) error { cancel(); return nil }}
	servers = append(servers, server)
	if status := run(ctx, second, makeRuntime, func(ipc.Config, ipc.Backend) (brokerServer, error) { return server, nil }); status != 0 {
		t.Fatal("second restart failed")
	}
	if len(runtimes) != 2 || runtimes[0] == runtimes[1] || len(servers) != 2 || servers[0] == servers[1] || bytes.Equal(runtimes[0].secret, runtimes[1].secret) {
		t.Fatalf("restart state counts runtimes=%d servers=%d", len(runtimes), len(servers))
	}
}
func TestRunFailsClosedOnRestartMissingSeed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	firstServer := &fakeServer{serve: func(context.Context) error { cancel(); return nil }}
	if status := run(ctx, newDescriptorFixture(testSeed), func([]byte) (brokerRuntime, error) { return &fakeRuntime{}, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { return firstServer, nil }); status != 0 {
		t.Fatal("successful first run failed")
	}
	missing := newDescriptorFixture(testSeed)
	missing.openErr["/"] = os.ErrNotExist
	runtimeCalls, listenCalls := 0, 0
	status := run(context.Background(), missing, func([]byte) (brokerRuntime, error) { runtimeCalls++; return nil, nil }, func(ipc.Config, ipc.Backend) (brokerServer, error) { listenCalls++; return nil, nil })
	if status != 1 || runtimeCalls != 0 || listenCalls != 0 {
		t.Fatalf("missing restart status=%d runtimeCalls=%d listenCalls=%d", status, runtimeCalls, listenCalls)
	}
}
func TestRunDoesNotUnlinkReplacementSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "identity.sock")
	runtime, err := backend.New([]byte("01234567890123456789"))
	if err != nil {
		t.Fatal(err)
	}
	server, err := ipc.Listen(ipc.Config{Path: path, AllowedUID: uint32(os.Geteuid())}, runtime)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "server unavailable") {
			t.Skipf("Unix sockets unavailable: %v", err)
		}
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = server.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("replacement socket path removed: %v", err)
	}
}
func TestRunServesValidOTPRequest(t *testing.T)          { exerciseClientProtocol(t, true) }
func TestRunRejectsMalformedClientRequest(t *testing.T)  { exerciseClientProtocol(t, false) }
func TestRunRedactsSecretFromErrorsAndLogs(t *testing.T) { TestRunRedactsSeedErrors(t) }

// PLAN-level names retain the finer-grained QA mapping as executable aliases.
func TestLoadSeedDescriptorWalk(t *testing.T) {
	t.Run("walk-and-flags", TestOpenSeedUsesNoFollowCloexecAndClosesOnFailures)
	t.Run("parent-and-reader-ownership", TestLoadSeedClosesParentsAndReaderOwnsFinalDescriptor)
}
func TestLoadSeedDescriptorErrors(t *testing.T) {
	t.Run("root-open", TestLoadSeedRootOpenError)
	t.Run("root-stat", TestLoadSeedRootStatError)
	t.Run("root-close", TestLoadSeedRootCloseError)
	t.Run("parent-open", TestLoadSeedParentOpenError)
	t.Run("parent-stat", TestLoadSeedParentStatError)
	t.Run("parent-close", TestLoadSeedParentCloseError)
	t.Run("final-open", TestLoadSeedFinalOpenError)
	t.Run("final-stat", TestLoadSeedFinalStatError)
	t.Run("final-descriptor-close", TestLoadSeedFinalDescriptorCloseError)
	t.Run("reader-close", TestLoadSeedReaderCloseError)
	t.Run("short-read", TestLoadSeedShortRead)
	t.Run("read-error", TestLoadSeedReadError)
}
func TestLoadSeedFinalReaderOwnership(t *testing.T) {
	t.Run("reader-owns-final", TestLoadSeedClosesParentsAndReaderOwnsFinalDescriptor)
	t.Run("wrapper-failure-closes-final", TestLoadSeedFileWrapperFailureClosesFinalDescriptor)
}
func TestLoadSeedSchema(t *testing.T) {
	t.Run("strict-schema", TestParseSeedStrictSchema)
	t.Run("valid", TestLoadSeedAcceptsValidSchema)
	t.Run("maximum", TestLoadSeedAcceptsMaximumSecret)
	t.Run("malformed", TestLoadSeedRejectsMalformedSchema)
	t.Run("duplicate", TestLoadSeedRejectsDuplicateSchemaField)
	t.Run("unknown", TestLoadSeedRejectsUnknownSchemaField)
	t.Run("missing", TestLoadSeedRejectsMissingSchemaField)
	t.Run("empty", TestLoadSeedRejectsEmptySecret)
	t.Run("wrong-type", TestLoadSeedRejectsWrongSchemaType)
	t.Run("trailing", TestLoadSeedRejectsTrailingJSON)
	t.Run("invalid-uid", TestLoadSeedRejectsInvalidAllowedUID)
	t.Run("invalid-base64", TestLoadSeedRejectsInvalidBase64)
	t.Run("noncanonical-base64", TestLoadSeedRejectsNonCanonicalBase64)
	t.Run("oversized", TestLoadSeedRejectsOversizedSchemaInput)
}
func TestRunConstructionAndListenFailures(t *testing.T) {
	t.Run("runtime-factory", TestRunRuntimeFactoryError)
	t.Run("construction-order", TestRunConstructsRuntimeBeforeListen)
	t.Run("configuration-order", TestRunConfiguresServerBeforeListen)
	t.Run("listener-error", TestRunClosesServerOnListenError)
}
func TestRunServeAndCloseFailures(t *testing.T) {
	t.Run("unexpected-serve", TestRunRejectsUnexpectedServeReturn)
	t.Run("server-close", TestRunReportsServerCloseError)
}
func TestRunSignalsAndShutdown(t *testing.T) {
	t.Run("sigint", TestRunShutsDownOnSIGINT)
	t.Run("sigterm", TestRunShutsDownOnSIGTERM)
}
func TestRunActiveConcurrentRepeatedShutdown(t *testing.T) {
	t.Run("active-client", TestRunWaitsForActiveClientOnShutdown)
	t.Run("concurrent", TestRunHandlesConcurrentShutdown)
	t.Run("repeated", TestRunShutdownIsIdempotent)
}
func TestRunSocketReplacementAndRestart(t *testing.T) {
	t.Run("socket-replacement", TestRunDoesNotUnlinkReplacementSocket)
	t.Run("restart", TestRunRestartsWithFreshSeed)
}
func TestRunRestartWithoutSeed(t *testing.T) {
	t.Run("successful-then-missing", TestRunFailsClosedOnRestartMissingSeed)
}
func TestBrokerClientOTPAndMalformedRequest(t *testing.T) {
	t.Run("valid-otp", TestRunServesValidOTPRequest)
	t.Run("malformed", TestRunRejectsMalformedClientRequest)
}
