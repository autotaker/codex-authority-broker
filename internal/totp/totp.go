package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/binary"
	"time"
)

type Validator struct {
	seed []byte
	now  func() time.Time
	skew int
}

func New(seed []byte, now func() time.Time, skew int) *Validator {
	if now == nil {
		now = time.Now
	}
	return &Validator{append([]byte(nil), seed...), now, skew}
}
func (v *Validator) Validate(in []byte) (uint64, bool) {
	if len(in) != 6 {
		return 0, false
	}
	base := v.now().Unix() / 30
	for d := -v.skew; d <= v.skew; d++ {
		s := base + int64(d)
		if s >= 0 {
			c := code(v.seed, uint64(s))
			if subtle.ConstantTimeCompare(in, c[:]) == 1 {
				return uint64(s), true
			}
		}
	}
	return 0, false
}
func code(k []byte, s uint64) [6]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], s)
	h := hmac.New(sha1.New, k)
	h.Write(b[:])
	x := h.Sum(nil)
	o := x[19] & 15
	n := ((uint32(x[o])&127)<<24 | uint32(x[o+1])<<16 | uint32(x[o+2])<<8 | uint32(x[o+3])) % 1e6
	var r [6]byte
	for i := 5; i >= 0; i-- {
		r[i] = byte('0' + n%10)
		n /= 10
	}
	return r
}
func (v *Validator) Close() {
	for i := range v.seed {
		v.seed[i] = 0
	}
}
