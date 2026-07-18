# PLAN — TASK-0003: TOTP replay, rate limits, and concurrency

## Approval boundary

This PLAN owns only in-process TOTP verification and its replay, rate-limit,
and one-time activation controls on the merged TASK-0001 lease state.  It
does not create an authority endpoint or a secret-delivery mechanism.

**DEV profile:** `luna-xhigh` (`dev-luna`).  The change is a small extension
to the existing deterministic Go state machine; it has no OS, network, IPC,
or persistence boundary.

The execution clock is 30 minutes after preflight.  Before it starts, confirm
TASK-0001 is merged, the `task/TASK-0003-totp-controls` worktree is usable,
Go and the commands below are available, an injectable fake-clock fixture is
ready, and the secret test fixture is synthetic.  A failed prerequisite is
recorded as not-started.  DEV also requires a separately approved
TASK-0003-only `QA_PLAN.md`; this document alone authorizes neither DEV nor
any Git operation.

## Paths and exclusions

| Path | Responsibility |
| --- | --- |
| `internal/lease/lease.go` | Add the locked verification-and-activation transition and its bounded rate/replay state; preserve the TASK-0001 public transitions. |
| `internal/lease/totp.go` | Implement the in-memory verifier and RFC 6238 calculation with Go standard-library primitives. |
| `internal/lease/lease_test.go` | Extend state/atomicity tests using the existing fake clock. |
| `internal/lease/totp_test.go` | Add deterministic RFC-vector and verifier/rate/replay tests with synthetic secrets. |

No IPC, socket, command, service, configuration, environment-variable,
argument, file, database, keyring, persistence, PAM/sudo, audit, release, or
operator-documentation work is allowed.  No external TOTP package or module
is added.  In particular, this task does not define how a production secret
arrives at the process, accept one over a future IPC/CLI boundary, persist it,
or log it.  TASK-0004 and TASK-0005 retain those delivery/redaction
boundaries.

## API, secret, and verification design

Add a package-level `TOTPVerifier`, constructed only from caller-supplied raw
secret bytes, for example:

```go
func NewTOTPVerifier(secret []byte) (*TOTPVerifier, error)
func (s *State) VerifyAndActivate(challenge Challenge, code string, verifier *TOTPVerifier) (Lease, error)
```

The constructor rejects an empty secret and copies non-empty bytes into an
unexported field.  `State` exposes neither the secret nor a TOTP value, and
the verifier has no formatting/String method.  Production code emits no logs
from this package and all returned errors are fixed sentinel text: they must
not contain the submitted code, secret, expected code, counter, or time.
Tests use only a documented synthetic byte fixture.  The creating component
is responsible for its injection lifetime; this task creates no durable
secret store and makes no promise that Go can zero every caller/runtime copy.

`New` records the current non-negative 30-second counter as a private boot
replay floor.  `VerifyAndActivate` is the only new activation path.  It first
holds the existing `State.mu`, expires current state, requires the supplied
current challenge and an inactive lease, applies the rate rule, verifies the
code, checks/records replay, and performs the existing activation transition
before unlocking.  Refactor the current transition into a small private
locked helper as needed; do not lock recursively or expose a separately
callable verify-then-activate sequence.  Therefore two concurrent submissions
of the same valid code can produce at most one lease: the winner consumes the
challenge and its time-step record while the loser observes a denial.  The
new boot field must not change `BeginReadiness`, `Activate`, `Active`, their
deadlines, or the existing TASK-0001 tests.  `Activate` remains a TASK-0001
primitive for compatibility, but it is not an OTP API and no IPC is expanded.

Use RFC 6238 defaults implemented with the standard library: HMAC-SHA-1,
30-second Unix steps from `T0 = 0`, dynamic truncation, and six decimal
digits.  Compare the fixed-width decimal candidate with `hmac.Equal`; reject
non-six-digit input before calculating a successful match.  Use
`State.clock.Now()` for every counter/rate decision, keeping tests fully
deterministic.  A valid code is current-step or one adjacent 30-second step
(`-1, 0, +1`) to allow bounded drift; an older/newer counter is stale and
denies.  Negative Unix timestamps and counters deny without conversion to an
unsigned value.  If `New` observes a negative Unix timestamp, it records zero
as the conservative boot floor and verification remains denied until a
non-negative counter strictly above zero is available.  RFC test vectors may
validate the calculation at their specified timestamps, while activation
tests use the six-digit profile.

Replay is process-local and fail-closed without persistence: retain the boot
replay floor and a bounded watermark for the highest accepted TOTP counter.
Deny every otherwise-valid counter less than or equal to the greater of those
two values; record a successful counter as the new accepted watermark before
unlocking.  This intentionally prefers replay safety over accepting an
out-of-order adjacent-window code after a newer one was accepted.  A fresh
`State` has no lease, secret, or restored accepted watermark, but its boot
floor rejects the entire step in which it starts.  Thus restart cannot reuse
the same-step or any older OTP after secret reinjection; a normal non-negative
clock must reach a strictly newer 30-second step before activation (a maximum
wait of about 30 seconds for a same-clock authenticator).  This startup delay
is an explicit fail-closed UX tradeoff, not a reason to persist replay state
or relax the floor.

For each live challenge, allow at most **five** verification submissions in a
fixed **60-second** window beginning with its first submission.  Count every
attempt that reaches a current live challenge, including malformed, stale,
replay, and successful submissions; the sixth and later submission before the
window boundary returns the rate-limit denial without TOTP computation.  At
exactly 60 seconds a new window starts.  A new readiness challenge starts a
fresh rate window; consuming or expiring a challenge clears its attempt state.
Rate data, boot floor, and accepted watermark are private `State` fields
guarded by the same mutex—no timer, goroutine, or persistence is introduced.

## DEV sequence and focused evidence

1. Implement the verifier in `totp.go` using `crypto/hmac`, `crypto/sha1`,
   `encoding/binary`, and ordinary named helpers.  Keep RFC parameters as
   named constants and comments; do not compress cryptographic or error
   handling logic.
2. Add the mutex-held `VerifyAndActivate` path and minimal private state to
   `lease.go`.  Preserve `BeginReadiness`, `Activate`, `Active`, fixed
   300-second challenge/lease expiry, and process-memory restart behavior.
3. Add fake-clock tests for a known RFC calculation vector, valid current and
   adjacent-step activation, malformed/stale/negative-time denial, reused
   accepted counter denial after a new challenge, five-attempt boundary and
   reset exactly at 60 seconds, and no authority after every denial.  Create a
   fresh `State` in the same TOTP step and prove the same and older counters
   deny after secret reinjection; then advance exactly to the next step and
   prove its strictly newer counter may activate.  This test records only
   synthetic fixture labels, never an OTP or secret.  Assert errors only by
   sentinel identity.
4. Add a barrier-based concurrent test that submits the same valid code to a
   single challenge from multiple goroutines and proves exactly one success,
   one active lease, and denial for every other caller.  Run it under the race
   detector.  Keep the existing TASK-0001 restart/deadline tests unchanged and
   passing; the boot floor affects only the new TOTP transition.
5. Run `gofmt -w` only on the allowed Go files, then record redacted concise
   output and Go version for `go test ./...`, `go vet ./...`, `go test -race
   ./...`, `test -z "$(gofmt -l .)"`, the exact SLOC count below, and `git
   diff --check`.  Do not add a Makefile or invoke one.

The target is approximately 170–220 added production SLOC (cumulative
roughly 268–318 from the measured TASK-0001 baseline of 98), well below the
430 cumulative ceiling.  If projected cumulative use exceeds 387 (>90% of
430), stop and re-estimate before continuing.  If the candidate exceeds 430,
stop DEV, follow `backlog.json`'s exact ordered feature-shedding list, and
obtain new PLAN and QA approval before resuming.  TOTP replay, rate, and
absolute lease controls are mandatory and cannot be shed; a mandatory-v1
total above 1500 is a requirement gap, never a reason to compress source.

## Cumulative production-SLOC and review handoff

Count all candidate production Go files (including the existing
`internal/lease/lease.go` and the new `internal/lease/totp.go`), excluding
`*_test.go`, generated/vendor/configuration files, and task documents.  Count
each nonblank line that is not solely comment content; inline code plus a
comment counts once.  REVIEW and QA independently run this exact command at
the candidate root and record file subtotals and `TOTAL` against **<=430**:

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

Independent REVIEW must confirm the approved paths only; standard-library
RFC-6238 parameters; no secret/code disclosure, persistence, transport, or
log path; atomic mutex coverage; exact drift, boot-floor/restart replay, and
rate-window boundaries; unchanged TASK-0001 behavior; and focused test, vet,
race, format, diff, SLOC, and readability evidence.  REVIEW rejects
semicolon/one-line packing, collapsed error
handling, cryptic names, removed security comments, or functions combined
only to meet LOC.  Independent QA repeats the acceptance tests and SLOC/scope
review after REVIEW PASS.  Any FAIL returns to its responsible gate; no merge,
stage, commit, or product work is authorized by this PLAN.
