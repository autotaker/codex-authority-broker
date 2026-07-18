package lease

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRFC6238SHA1Vector(t *testing.T) {
	verifier, err := NewTOTPVerifier([]byte("12345678901234567890"))
	if err != nil {
		t.Fatal("synthetic verifier was rejected")
	}
	if !verifier.matches(1, "287082") {
		t.Fatal("RFC 6238 SHA-1 vector did not verify")
	}
}

func TestVerifyAndActivateConsumesOnce(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0).UTC()}
	state := New(clock)
	challenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewTOTPVerifier([]byte("fixture-secret"))
	if err != nil {
		t.Fatal(err)
	}
	clock.now = time.Unix(30, 0).UTC()
	code := testCode(verifier, 1)
	if _, err := state.VerifyAndActivate(challenge, code, verifier); err != nil {
		t.Fatalf("verification failed: %v", err)
	}
	if !state.Active() {
		t.Fatal("successful verification did not activate")
	}
}

func TestConcurrentDuplicateVerificationHasOneWinner(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0).UTC()}
	state := New(clock)
	challenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewTOTPVerifier([]byte("fixture-secret"))
	if err != nil {
		t.Fatal(err)
	}
	clock.now = time.Unix(30, 0).UTC()
	code := testCode(verifier, 1)
	const callers = 8
	start := make(chan struct{})
	results := make(chan error, callers)
	var group sync.WaitGroup
	for i := 0; i < callers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_, callErr := state.VerifyAndActivate(challenge, code, verifier)
			results <- callErr
		}()
	}
	close(start)
	group.Wait()
	close(results)
	winners := 0
	for callErr := range results {
		if callErr == nil {
			winners++
		}
	}
	if winners != 1 || !state.Active() {
		t.Fatalf("duplicate verification winners = %d, active = %v", winners, state.Active())
	}
	if _, err := state.VerifyAndActivate(challenge, code, verifier); !errors.Is(err, ErrLeaseActive) {
		t.Fatal("post-activation duplicate did not deny")
	}
}

func TestVerificationWindowAndNegativeTime(t *testing.T) {
	const targetCounter int64 = 10
	for _, acceptedCounter := range []int64{targetCounter - 1, targetCounter, targetCounter + 1} {
		clock := &fakeClock{now: time.Unix((targetCounter-2)*30, 0).UTC()}
		state := New(clock)
		challenge, err := state.BeginReadiness()
		if err != nil {
			t.Fatal(err)
		}
		verifier, err := NewTOTPVerifier([]byte("window-fixture"))
		if err != nil {
			t.Fatal(err)
		}
		clock.now = time.Unix(targetCounter*30, 0).UTC()
		if _, err := state.VerifyAndActivate(challenge, testCode(verifier, acceptedCounter), verifier); err != nil {
			t.Fatalf("counter offset %d denied: %v", acceptedCounter-targetCounter, err)
		}
	}

	clock := &fakeClock{now: time.Unix((targetCounter-1)*30, 0).UTC()}
	state := New(clock)
	challenge, _ := state.BeginReadiness()
	verifier, _ := NewTOTPVerifier([]byte("window-fixture"))
	clock.now = time.Unix(targetCounter*30, 0).UTC()
	if _, err := state.VerifyAndActivate(challenge, testCode(verifier, targetCounter-1), verifier); !errors.Is(err, ErrTOTPReplay) {
		t.Fatalf("floor counter error = %v", err)
	}
	if _, err := state.VerifyAndActivate(challenge, testCode(verifier, targetCounter+2), verifier); !errors.Is(err, ErrInvalidTOTP) {
		t.Fatalf("beyond-window error = %v", err)
	}

	negativeClock := &fakeClock{now: time.Unix(-1, 0).UTC()}
	negativeState := New(negativeClock)
	negativeChallenge, _ := negativeState.BeginReadiness()
	if _, err := negativeState.VerifyAndActivate(negativeChallenge, "000000", verifier); !errors.Is(err, ErrInvalidTOTP) {
		t.Fatalf("negative-time error = %v", err)
	}
}

func TestRateWindowAndChallengeExpiryReset(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0).UTC()}
	state := New(clock)
	challenge, _ := state.BeginReadiness()
	verifier, _ := NewTOTPVerifier([]byte("rate-fixture"))
	clock.now = time.Unix(30, 0).UTC()
	for attempt := 1; attempt <= 5; attempt++ {
		if _, err := state.VerifyAndActivate(challenge, "bad", verifier); !errors.Is(err, ErrInvalidTOTP) {
			t.Fatalf("attempt %d error = %v", attempt, err)
		}
	}
	if _, err := state.VerifyAndActivate(challenge, "bad", verifier); !errors.Is(err, ErrTOTPRateLimit) {
		t.Fatalf("sixth attempt error = %v", err)
	}
	clock.now = time.Unix(90, 0).UTC()
	if _, err := state.VerifyAndActivate(challenge, testCode(verifier, 3), verifier); err != nil {
		t.Fatalf("exact window boundary did not reset: %v", err)
	}

	clock.now = time.Unix(390, 0).UTC()
	if state.Active() {
		t.Fatal("lease did not expire")
	}
	newChallenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := state.VerifyAndActivate(newChallenge, "bad", verifier); !errors.Is(err, ErrInvalidTOTP) {
		t.Fatalf("fresh challenge did not reset rate state: %v", err)
	}
}

func TestReplayWatermarkAndBootFloor(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0).UTC()}
	verifier, _ := NewTOTPVerifier([]byte("replay-fixture"))
	state := New(clock)
	challenge, _ := state.BeginReadiness()
	clock.now = time.Unix(30, 0).UTC()
	if _, err := state.VerifyAndActivate(challenge, testCode(verifier, 1), verifier); err != nil {
		t.Fatal(err)
	}
	clock.now = time.Unix(330, 0).UTC()
	if state.Active() {
		t.Fatal("lease did not expire")
	}
	clock.now = time.Unix(60, 0).UTC()
	challenge, _ = state.BeginReadiness()
	if _, err := state.VerifyAndActivate(challenge, testCode(verifier, 1), verifier); !errors.Is(err, ErrTOTPReplay) {
		t.Fatalf("replayed counter error = %v", err)
	}

	clock.now = time.Unix(30, 0).UTC()
	fresh := New(clock)
	freshChallenge, _ := fresh.BeginReadiness()
	if _, err := fresh.VerifyAndActivate(freshChallenge, testCode(verifier, 2), verifier); !errors.Is(err, ErrTOTPReplay) {
		t.Fatalf("future adjacent code at boot step error = %v", err)
	}
	if _, err := fresh.VerifyAndActivate(freshChallenge, testCode(verifier, 1), verifier); !errors.Is(err, ErrTOTPReplay) {
		t.Fatalf("same-step boot replay error = %v", err)
	}
	if _, err := fresh.VerifyAndActivate(freshChallenge, testCode(verifier, 0), verifier); !errors.Is(err, ErrTOTPReplay) {
		t.Fatalf("older boot replay error = %v", err)
	}
	clock.now = time.Unix(60, 0).UTC()
	if _, err := fresh.VerifyAndActivate(freshChallenge, testCode(verifier, 2), verifier); err != nil {
		t.Fatalf("strictly newer boot counter denied: %v", err)
	}
}

func TestTOTPErrorDoesNotDiscloseInputs(t *testing.T) {
	secret := "private-fixture-secret"
	code := "123456"
	verifier, _ := NewTOTPVerifier([]byte(secret))
	if _, err := NewTOTPVerifier(nil); !errors.Is(err, ErrInvalidTOTP) {
		t.Fatal("empty secret was accepted")
	}
	clock := &fakeClock{now: time.Unix(0, 0).UTC()}
	state := New(clock)
	challenge, _ := state.BeginReadiness()
	_, err := state.VerifyAndActivate(challenge, code, verifier)
	if err == nil || strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), code) {
		t.Fatalf("error disclosed verifier input: %v", err)
	}
}

func testCode(verifier *TOTPVerifier, counter int64) string {
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], uint64(counter))
	hash := hmac.New(sha1.New, verifier.secret)
	_, _ = hash.Write(message[:])
	digest := hash.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff
	return formatTestCode(value % 1000000)
}

func formatTestCode(value uint32) string {
	code := strconv.FormatUint(uint64(value), 10)
	for len(code) < totpDigits {
		code = "0" + code
	}
	return code
}
