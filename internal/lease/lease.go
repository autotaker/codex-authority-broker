// Package lease contains the in-process, time-bounded authority state.
package lease

import (
	"errors"
	"sync"
	"time"
)

const (
	challengeDuration = 300 * time.Second
	leaseDuration     = 300 * time.Second
)

var (
	// ErrLeaseActive means a lease is already active and cannot be replaced.
	ErrLeaseActive = errors.New("lease: lease is active")
	// ErrNoChallenge means readiness has not opened a challenge.
	ErrNoChallenge = errors.New("lease: readiness challenge is not open")
	// ErrInvalidChallenge means the supplied challenge is not current.
	ErrInvalidChallenge = errors.New("lease: challenge is not current")
	// ErrInvalidTOTP means the submitted value did not verify.
	ErrInvalidTOTP = errors.New("lease: invalid totp")
	// ErrTOTPReplay means a verified counter was already consumed or is below
	// the process boot replay floor.
	ErrTOTPReplay = errors.New("lease: totp replay")
	// ErrTOTPRateLimit means this challenge has exceeded its attempt budget.
	ErrTOTPRateLimit = errors.New("lease: totp rate limit")
)

// Clock supplies the time used for all state transitions.
type Clock interface {
	Now() time.Time
}

// Challenge is an opaque readiness value accepted by Activate.
type Challenge struct {
	owner    *State
	sequence uint64
}

// Lease is an opaque value returned after a successful activation.
type Lease struct {
	owner    *State
	sequence uint64
}

// State is an in-memory lease state machine. Its zero value is not ready for
// use; construct it with New so that a restart always starts idle.
type State struct {
	mu sync.Mutex

	clock Clock

	nextSequence uint64
	challenge    Challenge
	challengeEnd time.Time
	lease        Lease
	leaseEnd     time.Time

	bootReplayFloor   int64
	acceptedWatermark int64
	attemptStart      time.Time
	attemptCount      int
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

// New constructs an idle, process-local state. A nil clock selects time.Now.
func New(clock Clock) *State {
	if clock == nil {
		clock = systemClock{}
	}
	now := clock.Now()
	floor := int64(0)
	if counter, ok := totpCounter(now); ok {
		floor = counter
	}
	return &State{clock: clock, bootReplayFloor: floor}
}

// BeginReadiness opens an absolute 300-second challenge without creating a
// lease. An already-open challenge is returned unchanged.
func (s *State) BeginReadiness() (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	s.expire(now)
	if s.lease.sequence != 0 {
		return Challenge{}, ErrLeaseActive
	}
	if s.challenge.sequence != 0 {
		return s.challenge, nil
	}

	s.nextSequence++
	s.challenge = Challenge{owner: s, sequence: s.nextSequence}
	s.challengeEnd = now.Add(challengeDuration)
	s.attemptStart = time.Time{}
	s.attemptCount = 0
	return s.challenge, nil
}

// Activate applies a transition for the current, unexpired challenge and
// creates a separate absolute 300-second lease.
func (s *State) Activate(challenge Challenge) (Lease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	s.expire(now)
	if s.lease.sequence != 0 {
		return Lease{}, ErrLeaseActive
	}
	if s.challenge.sequence == 0 {
		return Lease{}, ErrNoChallenge
	}
	if challenge.owner != s || challenge.sequence != s.challenge.sequence {
		return Lease{}, ErrInvalidChallenge
	}

	return s.activateLocked(now)
}

// VerifyAndActivate atomically verifies one OTP and consumes the challenge.
func (s *State) VerifyAndActivate(challenge Challenge, code string, verifier *TOTPVerifier) (Lease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	s.expire(now)
	if s.lease.sequence != 0 {
		return Lease{}, ErrLeaseActive
	}
	if s.challenge.sequence == 0 {
		return Lease{}, ErrNoChallenge
	}
	if challenge.owner != s || challenge.sequence != s.challenge.sequence {
		return Lease{}, ErrInvalidChallenge
	}
	if !s.recordAttempt(now) {
		return Lease{}, ErrTOTPRateLimit
	}
	if verifier == nil || !validTOTPInput(code) {
		return Lease{}, ErrInvalidTOTP
	}
	counter, ok := totpCounter(now)
	if !ok {
		return Lease{}, ErrInvalidTOTP
	}
	if counter <= s.bootReplayFloor {
		return Lease{}, ErrTOTPReplay
	}
	for offset := int64(-1); offset <= 1; offset++ {
		candidate := counter + offset
		if offset < 0 && counter == 0 {
			continue
		}
		if verifier.matches(candidate, code) {
			floor := s.bootReplayFloor
			if s.acceptedWatermark > floor {
				floor = s.acceptedWatermark
			}
			if candidate <= floor {
				return Lease{}, ErrTOTPReplay
			}
			s.acceptedWatermark = candidate
			return s.activateLocked(now)
		}
	}
	return Lease{}, ErrInvalidTOTP
}

// Active reports whether the lease is strictly before its immutable deadline.
func (s *State) Active() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expire(s.clock.Now())
	return s.lease.sequence != 0
}

func (s *State) expire(now time.Time) {
	if s.challenge.sequence != 0 && !now.Before(s.challengeEnd) {
		s.challenge = Challenge{}
		s.challengeEnd = time.Time{}
		s.attemptStart = time.Time{}
		s.attemptCount = 0
	}
	if s.lease.sequence != 0 && !now.Before(s.leaseEnd) {
		s.lease = Lease{}
		s.leaseEnd = time.Time{}
	}
}

func (s *State) activateLocked(now time.Time) (Lease, error) {
	s.nextSequence++
	s.lease = Lease{owner: s, sequence: s.nextSequence}
	s.leaseEnd = now.Add(leaseDuration)
	s.challenge = Challenge{}
	s.challengeEnd = time.Time{}
	s.attemptStart = time.Time{}
	s.attemptCount = 0
	return s.lease, nil
}

func (s *State) recordAttempt(now time.Time) bool {
	if s.attemptStart.IsZero() || !now.Before(s.attemptStart.Add(totpRateWindow)) {
		s.attemptStart = now
		s.attemptCount = 0
	}
	s.attemptCount++
	return s.attemptCount <= totpRateLimit
}
