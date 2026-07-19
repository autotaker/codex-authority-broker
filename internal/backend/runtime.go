package backend

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
	"github.com/autotaker/codex-authority-broker/internal/lease"
)

const (
	maxOperations = 4
	maxNameBytes  = 16
	otpBytes      = 17
	otpPrefix     = `{"code":"`
	maxAuditBytes = 256
)

var ErrRegistration = errors.New("backend: invalid registration")

// Handler returns only a bounded allow or deny decision.
type Handler func(context.Context, ipc.Request) bool

// Runtime owns one process-local lease state and fixed IPC routing.
type Runtime struct {
	mu            sync.Mutex
	auditMu       sync.Mutex
	closed        bool
	audit         io.Writer
	auditID       atomic.Uint64
	handlers      map[string]Handler
	beforeGate    func()
	beforePublish func(bool)

	shutdown       context.Context
	shutdownCancel context.CancelFunc
	challengeMu    sync.Mutex
	state          *lease.State
	verifier       *lease.TOTPVerifier
	challenge      lease.Challenge
}

type auditEvent struct {
	CorrelationID string     `json:"correlation_id"`
	ActorUID      uint32     `json:"actor_uid"`
	Scope         string     `json:"scope"`
	Result        string     `json:"result"`
	LeaseExpiry   *time.Time `json:"lease_expiry"`
}

// New constructs an idle runtime using the ordinary system clock.
func New(secret []byte) (*Runtime, error) { return newRuntimeWithAudit(secret, nil, os.Stderr) }

func (*Runtime) String() string { return "backend.Runtime" }

func newRuntime(secret []byte, clock lease.Clock) (*Runtime, error) {
	return newRuntimeWithAudit(secret, clock, nil)
}

func newRuntimeWithAudit(secret []byte, clock lease.Clock, sink io.Writer) (*Runtime, error) {
	verifier, err := lease.NewTOTPVerifier(secret)
	if err != nil {
		return nil, err
	}
	shutdown, shutdownCancel := context.WithCancel(context.Background())
	runtime := &Runtime{
		handlers:       make(map[string]Handler, maxOperations),
		shutdown:       shutdown,
		shutdownCancel: shutdownCancel,
		state:          lease.New(clock),
		verifier:       verifier,
		audit:          sink,
	}
	runtime.handlers[ipc.OperationReady] = runtime.handleReady
	runtime.handlers[ipc.OperationOTP] = runtime.handleOTP
	runtime.handlers[ipc.OperationAuthorize] = runtime.handleAuthorize
	return runtime, nil
}

func (r *Runtime) Register(operation string, handler Handler) error {
	if r == nil || handler == nil || !validOperationName(operation) {
		return ErrRegistration
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || len(r.handlers) >= maxOperations || r.handlers[operation] != nil {
		return ErrRegistration
	}
	r.handlers[operation] = handler
	return nil
}

func (r *Runtime) Handle(ctx context.Context, request ipc.Request) (ipc.Response, error) {
	denied := ipc.Response{OK: false}
	if r == nil || ctx == nil || ctx.Err() != nil || request.Version != ipc.ProtocolVersion {
		return denied, nil
	}
	if request.Operation != ipc.OperationReady && request.Operation != ipc.OperationOTP && request.Operation != ipc.OperationAuthorize {
		return denied, nil
	}
	uid, audited := ipc.ActorUID(ctx)
	if r.audit != nil {
		r.auditMu.Lock()
		defer r.auditMu.Unlock()
		if !audited {
			r.Close()
			return denied, nil
		}
	}
	id := r.auditID.Add(1)
	if audited && id == 0 {
		r.Close()
		return denied, nil
	}
	if r.beforeGate != nil {
		r.beforeGate()
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return denied, nil
	}
	handler := r.handlers[request.Operation]
	if handler == nil {
		r.mu.Unlock()
		return denied, nil
	}
	callCtx, cancelCall := context.WithCancel(ctx)
	stopCancel := context.AfterFunc(r.shutdown, cancelCall)
	r.mu.Unlock()
	if callCtx.Err() != nil {
		stopCancel()
		cancelCall()
		return denied, nil
	}
	ok := handler(callCtx, request)
	stopCancel()
	cancelCall()
	if r.beforePublish != nil {
		r.beforePublish(ok)
	}
	r.mu.Lock()
	deadline, leaseActive := time.Time{}, true
	if request.Operation == ipc.OperationOTP || request.Operation == ipc.OperationAuthorize {
		deadline, leaseActive = r.state.Deadline()
	}
	final := !r.closed && r.shutdown.Err() == nil && ctx.Err() == nil && ok && leaseActive
	r.mu.Unlock()
	if audited && !r.writeAudit(id, uid, request.Operation, final, deadline) {
		r.Close()
		final = false
	}
	if !final {
		return denied, nil
	}
	return ipc.Response{OK: true}, nil
}

func (r *Runtime) writeAudit(id uint64, uid uint32, scope string, allow bool, deadline time.Time) bool {
	result := "deny"
	if allow {
		result = "allow"
	}
	var expiry *time.Time
	if allow && scope != ipc.OperationReady {
		expiry = &deadline
	}
	data, err := json.Marshal(auditEvent{strconv.FormatUint(id, 16), uid, scope, result, expiry})
	if err != nil || len(data)+1 > maxAuditBytes {
		return false
	}
	data = append(data, '\n')
	n, err := r.audit.Write(data)
	return err == nil && n == len(data)
}

// Close makes all new and in-flight calls fail closed; it is idempotent.
func (r *Runtime) Close() {
	if r == nil {
		return
	}
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		r.shutdownCancel()
	}
	r.mu.Unlock()
}

func (r *Runtime) handleReady(ctx context.Context, request ipc.Request) bool {
	if ctx == nil || ctx.Err() != nil || len(request.Payload) != 0 {
		return false
	}
	r.challengeMu.Lock()
	defer r.challengeMu.Unlock()
	if ctx.Err() != nil || r.state == nil {
		return false
	}
	challenge, err := r.state.BeginReadiness()
	if err != nil || ctx.Err() != nil {
		return false
	}
	r.challenge = challenge
	return true
}

func (r *Runtime) handleOTP(ctx context.Context, request ipc.Request) bool {
	code, ok := decodeOTP(request.Payload)
	if ctx == nil || ctx.Err() != nil || !ok {
		return false
	}
	r.challengeMu.Lock()
	defer r.challengeMu.Unlock()
	if ctx.Err() != nil || r.state == nil || r.verifier == nil {
		return false
	}
	_, err := r.state.VerifyAndActivate(r.challenge, code, r.verifier)
	if err != nil || ctx.Err() != nil {
		return false
	}
	r.challenge = lease.Challenge{}
	return true
}

func (r *Runtime) handleAuthorize(ctx context.Context, request ipc.Request) bool {
	if ctx == nil || ctx.Err() != nil || len(request.Payload) != 0 || r.state == nil {
		return false
	}
	return r.state.Active() && ctx.Err() == nil
}

func validOperationName(operation string) bool {
	if len(operation) == 0 || len(operation) > maxNameBytes {
		return false
	}
	for index := range operation {
		character := operation[index]
		if character >= 'a' && character <= 'z' {
			continue
		}
		if index == 0 || !((character >= '0' && character <= '9') || character == '-' || character == '_') {
			return false
		}
	}
	return true
}

func decodeOTP(payload []byte) (string, bool) {
	if len(payload) != otpBytes || string(payload[:len(otpPrefix)]) != otpPrefix || payload[15] != '"' || payload[16] != '}' {
		return "", false
	}
	for index := len(otpPrefix); index < 15; index++ {
		if payload[index] < '0' || payload[index] > '9' {
			return "", false
		}
	}
	return string(payload[len(otpPrefix):15]), true
}
