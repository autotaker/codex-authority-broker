package backend

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
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
	runtime, err := New(secret)
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
	runtime, err := New([]byte("registration-secret"))
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
	runtime, err := New([]byte("close-secret"))
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
	runtime, err := New([]byte("caller-secret"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err := runtime.Handle(ctx, testRequest(ipc.OperationReady, nil))
	assertDenied(t, response, err)
}

func TestCallerCancellationReachesAdmittedHandler(t *testing.T) {
	runtime, err := New([]byte("admitted-caller-secret"))
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
	runtime, err := New([]byte("publication-caller-secret"))
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
	runtime, err := New([]byte("waiting-close-secret"))
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
