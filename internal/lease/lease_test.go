package lease

import (
	"errors"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

func TestReadinessCreatesOnlyAnAbsoluteChallenge(t *testing.T) {
	base := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: base}
	state := New(clock)

	challenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatalf("BeginReadiness() error = %v", err)
	}
	if state.Active() {
		t.Fatal("readiness created an active lease")
	}

	clock.now = base.Add(100 * time.Second)
	repeated, err := state.BeginReadiness()
	if err != nil {
		t.Fatalf("repeated BeginReadiness() error = %v", err)
	}
	if repeated != challenge {
		t.Fatal("repeated readiness replaced the current challenge")
	}

	clock.now = base.Add(300 * time.Second)
	if _, err := state.Activate(challenge); !errors.Is(err, ErrNoChallenge) {
		t.Fatalf("expired challenge error = %v, want %v", err, ErrNoChallenge)
	}
}

func TestActivationRequiresCurrentUnexpiredChallenge(t *testing.T) {
	base := time.Date(2026, time.July, 19, 13, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: base}
	state := New(clock)

	if _, err := state.Activate(Challenge{}); !errors.Is(err, ErrNoChallenge) {
		t.Fatalf("activation before readiness error = %v, want %v", err, ErrNoChallenge)
	}
	if state.Active() {
		t.Fatal("activation before readiness created a lease")
	}

	challenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatalf("BeginReadiness() error = %v", err)
	}
	clock.now = base.Add(300*time.Second - time.Nanosecond)
	if _, err := state.Activate(challenge); err != nil {
		t.Fatalf("activation before challenge deadline error = %v", err)
	}
	if !state.Active() {
		t.Fatal("successful activation did not create a lease")
	}

	other := New(clock)
	otherChallenge, err := other.BeginReadiness()
	if err != nil {
		t.Fatalf("other BeginReadiness() error = %v", err)
	}
	if _, err := state.Activate(otherChallenge); !errors.Is(err, ErrLeaseActive) {
		t.Fatalf("activation while lease is active error = %v, want %v", err, ErrLeaseActive)
	}
}

func TestLeaseHasSeparateAbsoluteDeadline(t *testing.T) {
	base := time.Date(2026, time.July, 19, 14, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: base}
	state := New(clock)
	challenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatalf("BeginReadiness() error = %v", err)
	}

	activation := base.Add(2 * time.Minute)
	clock.now = activation
	if _, err := state.Activate(challenge); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	deadline, active := state.Deadline()
	if !active || !deadline.Equal(activation.Add(300*time.Second)) || deadline.Location() != time.UTC {
		t.Fatal("deadline accessor did not return the immutable UTC deadline")
	}
	clock.now = activation.Add(time.Minute)
	repeated, active := state.Deadline()
	if !active || !repeated.Equal(deadline) {
		t.Fatal("deadline accessor extended or replaced the lease")
	}

	clock.now = activation.Add(300*time.Second - time.Nanosecond)
	if !state.Active() {
		t.Fatal("lease expired early")
	}
	clock.now = activation.Add(300 * time.Second)
	if state.Active() {
		t.Fatal("lease remained active at its deadline")
	}
	if expired, active := state.Deadline(); active || !expired.IsZero() {
		t.Fatal("expired lease retained a deadline")
	}
	clock.now = activation.Add(301 * time.Second)
	if state.Active() {
		t.Fatal("lease remained active after its deadline")
	}
}

func TestNewStateFailsClosedAfterRestart(t *testing.T) {
	base := time.Date(2026, time.July, 19, 15, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: base}
	state := New(clock)
	challenge, err := state.BeginReadiness()
	if err != nil {
		t.Fatalf("BeginReadiness() error = %v", err)
	}
	if _, err := state.Activate(challenge); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	clock.now = base.Add(-time.Hour)
	if !state.Active() {
		t.Fatal("backward fake-clock movement incorrectly expired the lease")
	}
	fresh := New(clock)
	if fresh.Active() {
		t.Fatal("fresh state recovered the prior lease")
	}

	clock.now = base.Add(300 * time.Second)
	if state.Active() {
		t.Fatal("original state remained active at the deadline")
	}
}
