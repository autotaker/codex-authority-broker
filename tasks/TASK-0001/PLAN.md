# PLAN — TASK-0001: Core lease state, readiness, and absolute expiry

## Approval boundary

This is the first executable rolling-wave task.  It owns only the in-process
Go lease state and the readiness/challenge and absolute-expiry transitions.
The implementation is a library, not a daemon or privilege boundary.

**DEV profile:** `luna-xhigh` (`dev-luna`).  The work is a small, isolated Go
state machine with deterministic tests; it does not cross an OS, network, or
secret boundary.

The execution clock is **30 minutes after preflight**, not before it.
Preflight must confirm that the dependency set is merged, Go and the required
test command are available, the focused clock/restart fixture is prepared,
and the worktree is usable.  A failed preflight is recorded as not started.

Before DEV, the independent QA role must replace the historical broad
`QA_PLAN.md` with a TASK-0001-only QA plan and it must be approved alongside
this plan.  Neither document authorizes DEV by itself.

## Scope and explicit exclusions

Allowed implementation and test paths:

| Path | Responsibility |
| --- | --- |
| `go.mod`, `go.sum` | Minimal module metadata only if absent. |
| `internal/lease/lease.go` | Clock abstraction and mutex-protected in-memory lease state. |
| `internal/lease/lease_test.go` | Deterministic transition, boundary, and restart tests. |

No command, service, configuration, socket, storage, secret, TOTP,
rate-limit, replay, concurrency-consume, PAM/sudo, process, GitHub/push,
packaging, workflow, operator-documentation, or IPC work belongs here.
In particular, this task must not parse, validate, retain, log, or transport
an OTP.  TASK-0003 owns TOTP verification/replay/rate/concurrent consume;
TASK-0004 owns IPC; later tasks own their stated boundaries.

## Go design and state invariants

`internal/lease` provides a deliberately small API, with no exported mutable
state:

```go
type Clock interface { Now() time.Time }

func New(clock Clock) *State
func (s *State) BeginReadiness() (Challenge, error)
func (s *State) Activate(challenge Challenge) (Lease, error)
func (s *State) Active() bool
```

`Challenge` and `Lease` are opaque values sufficient for the next task to
request the transition; neither contains OTP/TOTP data.  `Activate` is not an
OTP verifier or an authority endpoint: it only applies the state transition
after a future verifier has accepted a challenge.  It must reject an absent,
stale, expired, or non-current challenge.  There is deliberately no API that
accepts an OTP-like value or grants a lease without a current challenge.

`State` is protected by one mutex.  Its only states are `idle`,
`challenge-open`, and `lease-active`; expired challenge/lease state is
observationally `idle`.  The real clock uses `time.Now()` and retains Go's
monotonic component.  Tests inject a fake `Clock` with deterministic time.

Both durations are fixed at 300 seconds: readiness creates
`challengeDeadline = now + 300s`, and successful activation creates the
separate immutable `leaseDeadline = activationNow + 300s`.  `BeginReadiness`
never creates a lease.  When a challenge is already open it returns the same
challenge (or a defined `ErrChallengeOpen`) without changing its deadline;
when a lease is active it refuses.  A lease is active only when
`clock.Now().Before(leaseDeadline)` is true, so it denies exactly at the
deadline and after it.  No readiness or activation call can renew, replace,
or extend either deadline.

State is intentionally process-memory only.  `New` always initializes
`idle`; therefore a process restart or any loss of monotonic-clock continuity
has no restoration path and fails closed.  No persistence/recovery mechanism
may be added in this task.

## DEV sequence and focused evidence

1. Add the module only if it is required, then implement `Clock`, `State`,
   opaque challenge/lease values, and the three transitions with ordinary
   named helpers and explicit errors.  Run `gofmt`.
2. Add table-driven fake-clock tests for: readiness from idle creates only an
   absolute challenge; activation before readiness denies and leaves idle;
   challenge expiry denies activation; activation produces one lease; an
   active lease refuses readiness/another activation; and `Active` is true at
   `deadline-1ns` and false at `deadline` and later.
3. Add a restart test that creates an active lease in one `State`, constructs
   a fresh `State` using the same fake clock, and proves the fresh instance is
   inactive.  Include a backward/forward wall-time fixture only to show the
   injected deadline comparison is deterministic; production's monotonic
   `time.Time` remains the authority.
4. Run `go test ./...`, verify `gofmt` produces no diff, and run
   `go vet ./...`.  Run any additional repository check only when that check
   already exists and applies to this task.  Record commands, Go version, and
   redacted pass/fail output for the reviewer.  Do not create a Makefile,
   check script, or other planning-only check scaffold.

The planned implementation is approximately 90–130 production SLOC, leaving
headroom under the cumulative 250-SLOC ceiling.  If the projection exceeds
225 SLOC (>90% of the cap), stop and re-estimate before further DEV.  A
candidate over 250 SLOC stops DEV: follow the exact ordered feature-shedding
list in `backlog.json`, re-run PLAN and QA approval, and never compress source.
No mandatory-v1 control may be shed; a mandatory-v1 total over 1500 SLOC is a
requirement gap.

## Exact production-SLOC evidence

For this task, count only non-test executable source files added/changed for
the implementation (`*.go`, excluding `*_test.go`; there should be no shipped
`*.py`, `*.sh`, or `*.c`).  Count every nonblank line that is not solely a
line comment or block-comment content; a line containing executable code and
an inline comment counts once.  Tests, task documents, workflows, generated
files, vendor, and declarative configuration do not count.

REVIEW and QA independently record the file list and a reproducible count
from the candidate worktree.  Run this exact count from the repository root
(it enumerates tracked and untracked candidate source and prints file
subtotals followed by `TOTAL`):

```sh
git ls-files --cached --others --exclude-standard -z -- '*.go' |
  grep -zv '_test\.go$' |
  xargs -0r awk '
    FNR == 1 { if (seen++) { print count " " previous; total += count }; previous = FILENAME; count = 0; in_comment = 0 }
    { line = $0; sub(/^[[:space:]]+/, "", line) }
    in_comment && line !~ /\*\// { next }
    in_comment { sub(/^.*\*\//, "", line); in_comment = 0; sub(/^[[:space:]]+/, "", line) }
    line ~ /^\/\*/ && line !~ /\*\// { in_comment = 1; next }
    line ~ /^\/\*/ { sub(/^\/\*.*\*\//, "", line); sub(/^[[:space:]]+/, "", line) }
    line != "" && line !~ /^\/\// { count++ }
    END { if (seen) { print count " " previous; total += count }; print "TOTAL " total }'
```

The `TOTAL` is compared with **<=250**, not with the estimate.  The reviewer
also checks the diff manually for
semicolon/one-line packing, collapsed error handling, cryptic names, removed
security comments, and functions combined only to meet LOC.  `gofmt` and
normal idiomatic Go structure are mandatory.

## Review handoff and completion criteria

Independent REVIEW must inspect only the approved paths and confirm:

- readiness creates a challenge but no lease;
- an attempted activation before readiness denies without changing state;
- fixed deadlines are absolute, including the exact expiry boundary;
- a fresh `State` after restart is inactive;
- fake-clock tests cover those observations; `go test ./...`, gofmt-clean
  verification, and `go vet ./...` pass; any existing applicable repository
  check passes; and the production-SLOC evidence is <=250 with
  no-compression controls satisfied.

QA independently repeats the focused transition/restart checks and the SLOC
count against the TASK boundary, then records PASS or FAIL.  A REVIEW or QA
FAIL returns to its responsible gate; no merge is authorized without both
independent PASS results.  This plan authorizes no staging, commit, merge, or
product work.
