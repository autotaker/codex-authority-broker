package ipc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
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
		"unknown field":     frame([]byte(`{"version":1,"operation":"request","extra":true}`)),
		"missing operation": frame([]byte(`{"version":1}`)),
		"unknown operation": frame([]byte(`{"version":1,"operation":"other"}`)),
		"wrong version":     frame([]byte(`{"version":2,"operation":"request"}`)),
		"trailing json":     frame([]byte(`{"version":1,"operation":"request"}{}`)),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := readRequest(bytes.NewReader(input)); err == nil {
				t.Fatal("malformed frame was accepted")
			}
		})
	}
}

func TestReadRequestAcceptsExactMaximum(t *testing.T) {
	prefix := `{"version":1,"operation":"request","payload":"`
	suffix := `"}`
	body := []byte(prefix + strings.Repeat("x", MaxFrameBytes-len(prefix)-len(suffix)) + suffix)
	if len(body) != MaxFrameBytes {
		t.Fatal("maximum fixture has wrong size")
	}
	request, err := readRequest(bytes.NewReader(frame(body)))
	if err != nil || request.Operation != OperationRequest {
		t.Fatalf("maximum frame rejected: %v", err)
	}
}

func TestRequestRoundTripAndGenericErrors(t *testing.T) {
	request := Request{Version: ProtocolVersion, Operation: OperationRequest, Payload: []byte(`{"value":true}`)}
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
}

func frame(body []byte) []byte {
	return append(frameHeader(len(body)), body...)
}

func frameHeader(length int) []byte {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(length))
	return header
}

func writeRequest(writer io.Writer, request Request) error {
	if request.Version != ProtocolVersion || request.Operation != OperationRequest {
		return ErrProtocol
	}
	return writeJSONFrame(writer, request)
}

func readResponse(reader io.Reader) (Response, error) {
	body, err := readFrame(reader)
	if err != nil {
		return Response{}, err
	}
	var response Response
	if err := decodeStrict(body, &response); err != nil || response.Version != ProtocolVersion {
		return Response{}, ErrProtocol
	}
	return response, nil
}
