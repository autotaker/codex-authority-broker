package lease

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fc struct{ n time.Duration }

func (c *fc) Now() time.Duration { return c.n }

type fv struct{}

func (fv) Validate(b []byte) (uint64, bool) { return 42, string(b) == "123456" }

type mr struct {
	sync.Mutex
	m map[uint64]bool
}

func (r *mr) Consumed(s uint64) (bool, error) { r.Lock(); defer r.Unlock(); return r.m[s], nil }
func (r *mr) Consume(s uint64) error          { r.Lock(); defer r.Unlock(); r.m[s] = true; return nil }
func setup(t *testing.T) (*Manager, *fc) {
	c := &fc{}
	var n atomic.Int32
	m, e := New(Config{Clock: c, Validator: fv{}, Replay: &mr{m: map[uint64]bool{}}, IDs: func() (string, error) { return string(rune('a' + n.Add(1))), nil }, PerChallengeFailures: 2})
	if e != nil {
		t.Fatal(e)
	}
	return m, c
}
func TestAbsoluteLease(t *testing.T) {
	m, c := setup(t)
	a, _ := m.ConfirmReady()
	c.n = 100 * time.Second
	b, _ := m.ConfirmReady()
	if a.Challenge != b.Challenge || a.Deadline != b.Deadline {
		t.Fatal("extended")
	}
	if _, e := m.SubmitOTP(a.Challenge, []byte("123456")); e != nil {
		t.Fatal(e)
	}
	c.n = 399*time.Second + 999*time.Millisecond
	if !m.Active() {
		t.Fatal("early")
	}
	c.n = 400 * time.Second
	if m.Active() {
		t.Fatal("valid at deadline")
	}
}
func TestBeforeReadyLimitAndConcurrency(t *testing.T) {
	m, _ := setup(t)
	if _, e := m.SubmitOTP("x", []byte("123456")); e != ErrDenied {
		t.Fatal(e)
	}
	s, _ := m.ConfirmReady()
	var ok atomic.Int32
	var w sync.WaitGroup
	for range 2 {
		w.Add(1)
		go func() {
			defer w.Done()
			if _, e := m.SubmitOTP(s.Challenge, []byte("123456")); e == nil {
				ok.Add(1)
			}
		}()
	}
	w.Wait()
	if ok.Load() != 1 {
		t.Fatal(ok.Load())
	}
	m, _ = setup(t)
	s, _ = m.ConfirmReady()
	for range 2 {
		_, _ = m.SubmitOTP(s.Challenge, []byte("bad"))
	}
	if m.Status().State != "idle" {
		t.Fatal("limit")
	}
}
