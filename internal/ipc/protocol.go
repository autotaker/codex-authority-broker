package ipc

import (
	"encoding/binary"
	"errors"
)

const (
	Magic             uint32 = 0x43415831
	Version           byte   = 1
	HeaderSize               = 12
	MaxPayload               = 256
	TypeCheck         byte   = 1
	TypeConfirmReady  byte   = 2
	TypeSubmitOTP     byte   = 3
	TypeAuthorizePush byte   = 4
	StatusOK          byte   = 0
	StatusDenied      byte   = 1
)

var ErrProtocol = errors.New("invalid authority protocol")

type Message struct {
	Type, Status byte
	Payload      []byte
}

func Encode(m Message) ([]byte, error) {
	if len(m.Payload) > MaxPayload {
		return nil, ErrProtocol
	}
	b := make([]byte, HeaderSize+len(m.Payload))
	binary.BigEndian.PutUint32(b, Magic)
	b[4] = Version
	b[5] = m.Type
	b[6] = m.Status
	binary.BigEndian.PutUint32(b[8:], uint32(len(m.Payload)))
	copy(b[12:], m.Payload)
	return b, nil
}
func Decode(b []byte) (Message, error) {
	if len(b) < HeaderSize || len(b) > HeaderSize+MaxPayload || binary.BigEndian.Uint32(b) != Magic || b[4] != Version || int(binary.BigEndian.Uint32(b[8:])) != len(b)-HeaderSize {
		return Message{}, ErrProtocol
	}
	return Message{b[5], b[6], append([]byte(nil), b[12:]...)}, nil
}
