package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

type commandRecorder struct {
	mu         sync.Mutex
	calls      int
	operations []string
	otpShapeOK bool
}

func (r *commandRecorder) Handle(_ context.Context, request ipc.Request) (ipc.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	r.operations = append(r.operations, request.Operation)
	if request.Operation == ipc.OperationOTP {
		var payload struct {
			Code string `json:"code"`
		}
		err := json.Unmarshal(request.Payload, &payload)
		r.otpShapeOK = err == nil && sixDigits(payload.Code)
	}
	return ipc.Response{OK: true}, nil
}

func (r *commandRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *commandRecorder) otpValid() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.otpShapeOK
}

func TestRunAcceptsOnlyFixedActions(t *testing.T) {
	secret := syntheticOTP(t)
	calls := 0
	caller := func(_ context.Context, request ipc.Request) (ipc.Response, error) {
		calls++
		if request.Operation == ipc.OperationOTP {
			var payload struct {
				Code string `json:"code"`
			}
			if json.Unmarshal(request.Payload, &payload) != nil || payload.Code != secret {
				t.Fatal("OTP payload mismatch")
			}
		} else if len(request.Payload) != 0 {
			t.Fatal("readiness carried a payload")
		}
		responsePayload, _ := json.Marshal(struct {
			Value string `json:"value"`
		}{Value: secret})
		return ipc.Response{OK: true, Payload: responsePayload}, nil
	}
	for _, fixture := range []struct {
		args  []string
		stdin string
	}{
		{args: []string{ipc.OperationReady}},
		{args: []string{ipc.OperationOTP}, stdin: secret + "\n"},
	} {
		var stdout, stderr bytes.Buffer
		if status := run(fixture.args, strings.NewReader(fixture.stdin), &stdout, &stderr, caller); status != 0 {
			t.Fatal("fixed action failed")
		}
		if stderr.Len() != 0 || !capturesClean(secret, stdout.String()) {
			t.Fatal("success output was not clean")
		}
	}
	if calls != 2 {
		t.Fatalf("transport calls = %d", calls)
	}
}

func TestRunRejectsInvalidInputBeforeTransport(t *testing.T) {
	secret := syntheticOTP(t)
	t.Setenv("CODEX_AUTHORITY_OTP", secret)
	fixtures := []struct {
		args  []string
		stdin string
	}{
		{args: nil},
		{args: []string{"unknown"}},
		{args: []string{ipc.OperationReady, "extra"}},
		{args: []string{ipc.OperationOTP, secret}},
		{args: []string{"--otp", secret}},
		{args: []string{ipc.OperationOTP}},
		{args: []string{ipc.OperationOTP}, stdin: "12345\n"},
		{args: []string{ipc.OperationOTP}, stdin: "1234567\n"},
		{args: []string{ipc.OperationOTP}, stdin: "12x456\n"},
		{args: []string{ipc.OperationOTP}, stdin: "123456"},
		{args: []string{ipc.OperationOTP}, stdin: "123456\nextra"},
	}
	for _, fixture := range fixtures {
		calls := 0
		caller := func(context.Context, ipc.Request) (ipc.Response, error) {
			calls++
			return ipc.Response{OK: true}, nil
		}
		var stdout, stderr bytes.Buffer
		if status := run(fixture.args, strings.NewReader(fixture.stdin), &stdout, &stderr, caller); status == 0 {
			t.Fatal("invalid input succeeded")
		}
		if calls != 0 || stdout.Len() != 0 || stderr.String() != deniedLine {
			t.Fatal("invalid input did not fail locally and generically")
		}
		if !capturesClean(secret, stdout.String(), stderr.String()) {
			t.Fatal("invalid input leaked to output")
		}
	}
}

func TestRunRedactsTransportAndBackendDenials(t *testing.T) {
	secret := syntheticOTP(t)
	for _, result := range []struct {
		response ipc.Response
		err      error
	}{
		{err: errors.New("private transport detail")},
		{response: ipc.Response{OK: false, Payload: json.RawMessage(`{"private":true}`)}},
	} {
		var stdout, stderr bytes.Buffer
		caller := func(context.Context, ipc.Request) (ipc.Response, error) {
			return result.response, result.err
		}
		if status := run([]string{ipc.OperationOTP}, strings.NewReader(secret+"\n"), &stdout, &stderr, caller); status == 0 {
			t.Fatal("denial succeeded")
		}
		if stdout.Len() != 0 || stderr.String() != deniedLine || !capturesClean(secret, stdout.String(), stderr.String()) {
			t.Fatal("denial output was not generic and clean")
		}
	}
}

func TestCLIRealSocketCaptureScan(t *testing.T) {
	secret := syntheticOTP(t)
	binaryPath := filepath.Join(t.TempDir(), "codex-authority")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := build.Run(); err != nil {
		t.Fatal("CLI build failed")
	}
	recorder := &commandRecorder{}
	socketPath := filepath.Join(t.TempDir(), "server.sock")
	server, err := ipc.Listen(ipc.Config{Path: socketPath, AllowedUID: uint32(os.Geteuid())}, recorder)
	if err != nil {
		t.Fatal("server listen failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = server.Serve(ctx) }()
	defer func() {
		cancel()
		_ = server.Close()
	}()

	for _, fixture := range []struct {
		action string
		stdin  string
	}{
		{action: ipc.OperationReady},
		{action: ipc.OperationOTP, stdin: secret + "\n"},
	} {
		command := exec.Command(binaryPath, "--socket", socketPath, fixture.action)
		command.Env = os.Environ()
		command.Stdin = strings.NewReader(fixture.stdin)
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		err := command.Run()
		if err != nil || !capturesClean(secret, strings.Join(command.Args, "\x00"), strings.Join(command.Env, "\x00"), stdout.String(), stderr.String(), errorText(err), "") {
			t.Fatal("positive capture scan failed")
		}
	}
	if recorder.count() != 2 || !recorder.otpValid() {
		t.Fatal("real socket dispatch mismatch")
	}

	before := recorder.count()
	positional := exec.Command(binaryPath, "--socket", socketPath, ipc.OperationOTP, secret)
	positional.Env = os.Environ()
	var positionalOut, positionalErr bytes.Buffer
	positional.Stdout, positional.Stderr = &positionalOut, &positionalErr
	if err := positional.Run(); err == nil || !capturesClean(secret, positionalOut.String(), positionalErr.String(), errorText(err)) {
		t.Fatal("positional rejection capture failed")
	}
	environment := exec.Command(binaryPath, "--socket", socketPath, ipc.OperationOTP)
	environment.Env = append(os.Environ(), "CODEX_AUTHORITY_OTP="+secret)
	var environmentOut, environmentErr bytes.Buffer
	environment.Stdout, environment.Stderr = &environmentOut, &environmentErr
	if err := environment.Run(); err == nil || !capturesClean(secret, environmentOut.String(), environmentErr.String(), errorText(err)) {
		t.Fatal("environment rejection capture failed")
	}
	if recorder.count() != before {
		t.Fatal("rejected secret source dispatched")
	}
	if capturesClean(secret, "negative-control:"+secret) {
		t.Fatal("capture scanner missed negative control")
	}
}

func capturesClean(secret string, captures ...string) bool {
	for _, capture := range captures {
		if strings.Contains(capture, secret) {
			return false
		}
	}
	return true
}

func syntheticOTP(t *testing.T) string {
	t.Helper()
	value := make([]byte, 6)
	if _, err := rand.Read(value); err != nil {
		t.Fatal("synthetic input generation failed")
	}
	for index := range value {
		value[index] = '0' + value[index]%10
	}
	return string(value)
}

func sixDigits(value string) bool {
	if len(value) != 6 {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
