package lease

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"time"
)

const (
	totpStepSeconds = int64(30)
	totpDigits      = 6
	totpRateLimit   = 5
	totpRateWindow  = 60 * time.Second
)

// TOTPVerifier retains a private copy of a caller-supplied secret.
type TOTPVerifier struct {
	secret []byte
}

// NewTOTPVerifier constructs a verifier from a non-empty raw secret.
func NewTOTPVerifier(secret []byte) (*TOTPVerifier, error) {
	if len(secret) == 0 {
		return nil, ErrInvalidTOTP
	}
	copyOfSecret := append([]byte(nil), secret...)
	return &TOTPVerifier{secret: copyOfSecret}, nil
}

func totpCounter(now time.Time) (int64, bool) {
	unix := now.Unix()
	if unix < 0 {
		return 0, false
	}
	return unix / totpStepSeconds, true
}

func validTOTPInput(code string) bool {
	if len(code) != totpDigits {
		return false
	}
	for i := range code {
		if code[i] < '0' || code[i] > '9' {
			return false
		}
	}
	return true
}

func (v *TOTPVerifier) matches(counter int64, code string) bool {
	if counter < 0 {
		return false
	}
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], uint64(counter))
	hash := hmac.New(sha1.New, v.secret)
	_, _ = hash.Write(message[:])
	digest := hash.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff
	value %= 1000000
	var expected [totpDigits]byte
	for index := totpDigits - 1; index >= 0; index-- {
		expected[index] = byte(value%10) + '0'
		value /= 10
	}
	return hmac.Equal(expected[:], []byte(code))
}
