package ipc

import "testing"

func TestMalformed(t *testing.T) {
	b, _ := Encode(Message{Type: TypeCheck})
	for _, x := range [][]byte{nil, b[:5], append(b, 0), make([]byte, HeaderSize+MaxPayload+1)} {
		if _, e := Decode(x); e == nil {
			t.Fatal(len(x))
		}
	}
}
