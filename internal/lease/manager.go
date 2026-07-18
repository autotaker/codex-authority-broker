package lease

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const Duration = 300 * time.Second

var ErrDenied = errors.New("authority denied")

type Clock interface{ Now() time.Duration }
type Validator interface{ Validate([]byte) (uint64, bool) }
type ReplayStore interface {
	Consumed(uint64) (bool, error)
	Consume(uint64) error
}
type Config struct {
	Clock                                Clock
	Validator                            Validator
	Replay                               ReplayStore
	IDs                                  func() (string, error)
	PerChallengeFailures, GlobalFailures int
	GlobalWindow                         time.Duration
}
type Status struct {
	State     string        `json:"state"`
	Challenge string        `json:"challenge,omitempty"`
	Deadline  time.Duration `json:"deadline,omitempty"`
	LeaseID   string        `json:"lease_id,omitempty"`
}
type Manager struct {
	mu                        sync.Mutex
	cfg                       Config
	state, challenge, leaseID string
	deadline                  time.Duration
	failures                  int
	global                    []time.Duration
}

func New(c Config) (*Manager, error) {
	if c.Clock == nil || c.Validator == nil || c.Replay == nil {
		return nil, errors.New("missing dependency")
	}
	if c.IDs == nil {
		c.IDs = randomID
	}
	if c.PerChallengeFailures <= 0 {
		c.PerChallengeFailures = 5
	}
	if c.GlobalFailures <= 0 {
		c.GlobalFailures = 20
	}
	if c.GlobalWindow <= 0 {
		c.GlobalWindow = 5 * time.Minute
	}
	return &Manager{cfg: c, state: "idle"}, nil
}
func randomID() (string, error) {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return hex.EncodeToString(b), nil
}
func (m *Manager) expire(n time.Duration) {
	if (m.state == "challenge" || m.state == "active") && n >= m.deadline {
		m.state = "idle"
		m.challenge = ""
		m.leaseID = ""
		m.deadline = 0
		m.failures = 0
	}
}
func (m *Manager) ConfirmReady() (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := m.cfg.Clock.Now()
	m.expire(n)
	if m.state == "active" {
		return Status{}, ErrDenied
	}
	if m.state == "challenge" {
		return m.status(), nil
	}
	id, e := m.cfg.IDs()
	if e != nil {
		return Status{}, ErrDenied
	}
	m.state = "challenge"
	m.challenge = id
	m.deadline = n + Duration
	return m.status(), nil
}
func (m *Manager) SubmitOTP(h string, otp []byte) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := m.cfg.Clock.Now()
	m.expire(n)
	m.trim(n)
	if m.state != "challenge" || h == "" || h != m.challenge {
		return Status{}, ErrDenied
	}
	if m.failures >= m.cfg.PerChallengeFailures || len(m.global) >= m.cfg.GlobalFailures {
		m.close()
		return Status{}, ErrDenied
	}
	step, ok := m.cfg.Validator.Validate(otp)
	if !ok {
		m.fail(n)
		return Status{}, ErrDenied
	}
	used, e := m.cfg.Replay.Consumed(step)
	if e != nil || used {
		m.fail(n)
		return Status{}, ErrDenied
	}
	if m.cfg.Replay.Consume(step) != nil {
		m.close()
		return Status{}, ErrDenied
	}
	id, e := m.cfg.IDs()
	if e != nil {
		m.close()
		return Status{}, ErrDenied
	}
	m.state = "active"
	m.leaseID = id
	m.challenge = ""
	m.deadline = n + Duration
	m.failures = 0
	return m.status(), nil
}
func (m *Manager) fail(n time.Duration) {
	m.failures++
	m.global = append(m.global, n)
	if m.failures >= m.cfg.PerChallengeFailures || len(m.global) >= m.cfg.GlobalFailures {
		m.close()
	}
}
func (m *Manager) close() { m.state = "idle"; m.challenge = ""; m.deadline = 0; m.failures = 0 }
func (m *Manager) trim(n time.Duration) {
	cut := n - m.cfg.GlobalWindow
	i := 0
	for i < len(m.global) && m.global[i] <= cut {
		i++
	}
	m.global = append([]time.Duration(nil), m.global[i:]...)
}
func (m *Manager) Active() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expire(m.cfg.Clock.Now())
	return m.state == "active"
}
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expire(m.cfg.Clock.Now())
	return m.status()
}
func (m *Manager) status() Status {
	return Status{State: m.state, Challenge: m.challenge, Deadline: m.deadline, LeaseID: m.leaseID}
}

type MonotonicClock struct{ start time.Time }

func NewMonotonicClock() *MonotonicClock     { return &MonotonicClock{start: time.Now()} }
func (c *MonotonicClock) Now() time.Duration { return time.Since(c.start) }
