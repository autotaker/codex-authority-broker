package ipc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
)

func TestReadRequestRejectsMalformedFrames(t *testing.T) {
	tests := map[string][]byte{
		"short header":      {0, 0},
		"zero length":       frameHeader(0),
		"oversize":          frameHeader(MaxFrameBytes + 1),
		"short body":        append(frameHeader(10), []byte("{}")...),
		"invalid json":      frame([]byte("{")),
		"unknown field":     frame([]byte(`{"version":1,"operation":"ready","extra":true}`)),
		"missing operation": frame([]byte(`{"version":1}`)),
		"unknown operation": frame([]byte(`{"version":1,"operation":"other"}`)),
		"wrong version":     frame([]byte(`{"version":2,"operation":"ready"}`)),
		"authorize payload": frame([]byte(`{"version":1,"operation":"authorize","payload":{}}`)),
		"authorize null":    frame([]byte(`{"version":1,"operation":"authorize","payload":null}`)),
		"authorize version": frame([]byte(`{"version":2,"operation":"authorize"}`)),
		"trailing json":     frame([]byte(`{"version":1,"operation":"ready"}{}`)),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := readRequest(bytes.NewReader(input)); err == nil {
				t.Fatal("malformed frame was accepted")
			}
		})
	}
}

func TestAuthorizeProtocolAdmission(t *testing.T) {
	request := []byte(`{"version":1,"operation":"authorize"}`)
	decoded, err := readRequest(bytes.NewReader(frame(request)))
	if err != nil || decoded.Operation != OperationAuthorize || len(decoded.Payload) != 0 {
		t.Fatalf("payload-free authorize rejected: request=%+v err=%v", decoded, err)
	}
	for _, payload := range []string{`{}`, `null`, `[]`, `"authorize"`, `1`} {
		body := []byte(`{"version":1,"operation":"authorize","payload":` + payload + `}`)
		if _, err := readRequest(bytes.NewReader(frame(body))); !errors.Is(err, ErrProtocol) {
			t.Fatalf("payload %s accepted with err=%v", payload, err)
		}
	}
}

func TestReadRequestAcceptsExactMaximum(t *testing.T) {
	prefix := `{"version":1,"operation":"ready","payload":"`
	suffix := `"}`
	body := []byte(prefix + strings.Repeat("x", MaxFrameBytes-len(prefix)-len(suffix)) + suffix)
	if len(body) != MaxFrameBytes {
		t.Fatal("maximum fixture has wrong size")
	}
	request, err := readRequest(bytes.NewReader(frame(body)))
	if err != nil || request.Operation != OperationReady {
		t.Fatalf("maximum frame rejected: %v", err)
	}
}

func TestRequestRoundTripAndGenericErrors(t *testing.T) {
	request := Request{Version: ProtocolVersion, Operation: OperationReady, Payload: []byte(`{"value":true}`)}
	var transport bytes.Buffer
	if err := writeRequest(&transport, request); err != nil {
		t.Fatal(err)
	}
	decoded, err := readRequest(&transport)
	if err != nil || decoded.Operation != request.Operation {
		t.Fatalf("round trip failed: %v", err)
	}
	if strings.Contains(ErrProtocol.Error(), "value") || strings.Contains(ErrTransport.Error(), "value") {
		t.Fatal("generic error disclosed input")
	}
	if err := writeRequest(&bytes.Buffer{}, Request{}); !errors.Is(err, ErrProtocol) {
		t.Fatalf("invalid request error = %v", err)
	}
	var authorize bytes.Buffer
	if err := writeRequest(&authorize, Request{Version: ProtocolVersion, Operation: OperationAuthorize}); err != nil {
		t.Fatalf("payload-free authorize write failed: %v", err)
	}
	decoded, err = readRequest(&authorize)
	if err != nil || decoded.Operation != OperationAuthorize || len(decoded.Payload) != 0 {
		t.Fatalf("authorize round trip failed: request=%+v err=%v", decoded, err)
	}
	for _, payload := range [][]byte{[]byte(`{}`), []byte(`null`), []byte(`"x"`)} {
		if err := writeRequest(&bytes.Buffer{}, Request{Version: ProtocolVersion, Operation: OperationAuthorize, Payload: payload}); !errors.Is(err, ErrProtocol) {
			t.Fatalf("authorize payload write accepted %q: %v", payload, err)
		}
	}
}

func frame(body []byte) []byte {
	return append(frameHeader(len(body)), body...)
}

func frameHeader(length int) []byte {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(length))
	return header
}
