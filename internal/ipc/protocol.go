// Package ipc provides a bounded, versioned Unix-socket transport.
package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

const (
	ProtocolVersion = uint16(1)
	MaxFrameBytes   = 4096
	OperationReady  = "ready"
	OperationOTP    = "otp"
)

var (
	ErrProtocol  = errors.New("ipc: invalid request")
	ErrTransport = errors.New("ipc: transport failure")
)

type Request struct {
	Version   uint16          `json:"version"`
	Operation string          `json:"operation"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type Response struct {
	Version uint16          `json:"version"`
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func readRequest(reader io.Reader) (Request, error) {
	body, err := readFrame(reader)
	if err != nil {
		return Request{}, err
	}
	var request Request
	if err := decodeStrict(body, &request); err != nil {
		return Request{}, ErrProtocol
	}
	if request.Version != ProtocolVersion || !validOperation(request.Operation) {
		return Request{}, ErrProtocol
	}
	return request, nil
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

func writeRequest(writer io.Writer, request Request) error {
	if request.Version != ProtocolVersion || !validOperation(request.Operation) {
		return ErrProtocol
	}
	return writeJSONFrame(writer, request)
}

func writeResponse(writer io.Writer, response Response) error {
	response.Version = ProtocolVersion
	return writeJSONFrame(writer, response)
}

func readFrame(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, ErrTransport
	}
	length := binary.BigEndian.Uint32(header[:])
	if length == 0 || length > MaxFrameBytes {
		return nil, ErrProtocol
	}
	body := make([]byte, int(length))
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, ErrTransport
	}
	return body, nil
}

func writeJSONFrame(writer io.Writer, value any) error {
	body, err := json.Marshal(value)
	if err != nil || len(body) == 0 || len(body) > MaxFrameBytes {
		return ErrProtocol
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(body)))
	if err := writeAll(writer, header[:]); err != nil {
		return ErrTransport
	}
	if err := writeAll(writer, body); err != nil {
		return ErrTransport
	}
	return nil
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)
		if err != nil || written == 0 {
			return ErrTransport
		}
		data = data[written:]
	}
	return nil
}

func decodeStrict(body []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return ErrProtocol
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrProtocol
	}
	return nil
}

func validOperation(operation string) bool {
	return operation == OperationReady || operation == OperationOTP
}
