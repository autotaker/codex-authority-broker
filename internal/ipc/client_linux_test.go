//go:build linux

package ipc

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClientRoundTripsFixedOperations(t *testing.T) {
	backend := &recordingBackend{response: Response{OK: true}}
	server, path, cancel := startTestServer(t, uint32(os.Geteuid()), backend, nil)
	client := Client{Path: path}
	for _, operation := range []string{OperationReady, OperationOTP} {
		response, err := client.Call(context.Background(), Request{Version: ProtocolVersion, Operation: operation})
		if err != nil || !response.OK {
			t.Fatalf("fixed operation failed: %v", err)
		}
	}
	if backend.callCount() != 2 {
		t.Fatalf("backend calls = %d", backend.callCount())
	}
	cancel()
	_ = server.Close()
}

func TestClientFailsClosedWithGenericErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.sock")
	_, err := (Client{Path: missing}).Call(context.Background(), Request{Version: ProtocolVersion, Operation: OperationReady})
	if !errors.Is(err, ErrTransport) || strings.Contains(err.Error(), missing) {
		t.Fatal("unavailable error was not generic")
	}
	if _, err := (Client{Path: "relative.sock"}).Call(context.Background(), Request{}); !errors.Is(err, ErrTransport) {
		t.Fatal("relative path was accepted")
	}
}

func TestClientRejectsMalformedResponse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "malformed.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	listener.SetUnlinkOnClose(false)
	done := make(chan struct{})
	go func() {
		defer close(done)
		connection, acceptErr := listener.AcceptUnix()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		_, _ = readRequest(connection)
		_, _ = connection.Write(frame([]byte(`{"version":1,"ok":true,"extra":true}`)))
	}()
	_, err = (Client{Path: path}).Call(context.Background(), Request{Version: ProtocolVersion, Operation: OperationReady})
	if !errors.Is(err, ErrProtocol) {
		t.Fatalf("malformed response error = %v", err)
	}
	_ = listener.Close()
	_ = os.Remove(path)
	<-done
}

func TestClientHonorsContextDeadline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stalled.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	listener.SetUnlinkOnClose(false)
	accepted := make(chan *net.UnixConn, 1)
	go func() {
		connection, _ := listener.AcceptUnix()
		accepted <- connection
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err = (Client{Path: path}).Call(ctx, Request{Version: ProtocolVersion, Operation: OperationReady})
	if !errors.Is(err, ErrTransport) {
		t.Fatalf("deadline error = %v", err)
	}
	connection := <-accepted
	if connection != nil {
		_ = connection.Close()
	}
	_ = listener.Close()
	_ = os.Remove(path)
}
