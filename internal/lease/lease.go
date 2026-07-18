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
	return &State{clock: clock}
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

	s.nextSequence++
	s.lease = Lease{owner: s, sequence: s.nextSequence}
	s.leaseEnd = now.Add(leaseDuration)
	s.challenge = Challenge{}
	s.challengeEnd = time.Time{}
	return s.lease, nil
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
	}
	if s.lease.sequence != 0 && !now.Before(s.leaseEnd) {
		s.lease = Lease{}
		s.leaseEnd = time.Time{}
	}
}
