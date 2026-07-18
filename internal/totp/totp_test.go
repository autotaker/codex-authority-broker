package totp

import (
	"testing"
	"time"
)

func TestRFC6238(t *testing.T) {
	v := New([]byte("12345678901234567890"), func() time.Time { return time.Unix(59, 0) }, 0)
	s, ok := v.Validate([]byte("287082"))
	if !ok || s != 1 {
		t.Fatal(ok, s)
	}
}
