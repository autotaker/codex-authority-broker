# QA_PLAN — TASK-0013: process-local ready/otp runtime assembly

## QA gate and independent baseline

This QA plan is derived first from `TASK.md` and the applicable `AGENTS.md`.
It is independent of the implementation plan and is the QA acceptance contract
for the process-local runtime.  QA begins only after DEV and independent REVIEW
PASS; QA changes no files, stages nothing, and does not commit, merge, write
`.git`, or update the operational repository.  Main owns Git and records the
paired gate evidence.

The acceptance boundary is exactly these candidate-owned paths:

| Path | QA scope |
| --- | --- |
| `internal/backend/runtime.go` | Process-local assembly, routing, close, and bounded registration only. |
| `internal/backend/runtime_test.go` | Deterministic in-process evidence for that assembly only. |

QA fails closed for any candidate change to another production path or any
need for seed acquisition, a socket/listener, signals, daemon/lifecycle
startup, sudo, network, credentials, persistence, audit, installer, release,
canary, or push enablement.  Those are requirements/planning-scope failures,
not work to repair during this task.  DEV, REVIEW, and QA must remain separate
roles; QA does not attribute a failure to DEV until classification evidence
supports it.

## P0 acceptance assertions

Each P0 rejection must yield only a bounded decision: `ipc.Response{OK:false}`
with no payload and no secret, OTP, verifier, challenge, lease, counter,
timestamp, state error, or diagnostic detail exposed through the response or
returned error.  A success is `OK:true` with no payload.  No runtime logging,
formatting, callback, global registry, discovery, or push path may disclose or
emit this information.

1. **Exact routing and admission.**  Only protocol version 1 with the exact
   IPC names `ready` and `otp` is dispatchable. `ready` accepts only absent,
   nil, or zero-length payload. `otp` accepts only the exact 17-byte JSON
   representation `{"code":"NNNNNN"}` where each `N` is an ASCII digit.  All
   other bytes deny before a lease transition or registered-handler invocation.
2. **Lease/TOTP semantics.**  A successful `ready` opens/reuses the one
   process-local readiness challenge and does not create a lease. A valid OTP
   for that current challenge activates one lease. OTP before readiness;
   expired, foreign, stale, replayed, invalid, or rate-limited OTP; and ready
   while a lease is active all deny identically.  The existing injected-clock
   lease contract remains authoritative: challenge and lease are separate,
   absolute 300-second intervals, expiration is fail-closed, and restart/new
   state does not recover authority.
3. **Payload strictness mutation matrix.**  For `ready`, test non-empty JSON,
   whitespace, `null`, `{}`, array, string, number, and oversized payloads.
   For `otp`, test empty, whitespace, missing/unknown/duplicate fields,
   trailing JSON, wrong type, 5/7 digits, non-digit, non-ASCII digit, escaped
   form, extra whitespace, and oversized payload.  Test wrong version and an
   unknown operation as well. Every case asserts false/no-payload and no
   unintended handler call, challenge mutation, or lease activation.
4. **Close and concurrency.**  `Close` is idempotent and fail-closed. After it
   returns, every Handle and Register denies. Barrier tests must cover two
   simultaneous valid OTPs (no more than one activation), ready/OTP
   interleaving, and Handle racing Close. If Close wins, no handler may start
   afterward and an in-flight handler result cannot become a success; Close
   must not deadlock waiting for a handler. Run under the race detector.
5. **Bounded Register seam.**  The runtime has at most three instance-local
   registration slots. It installs exactly ready/otp itself, permits at most
   one construction-time-safe third registration, and rejects nil, empty,
   noncanonical/oversize, duplicate/replacement, fourth, and post-close
   registrations. A registered third name is still non-dispatchable: the
   fixed IPC allowlist precedes lookup. Verify no global state and no push.
6. **Redaction.**  Use unique synthetic secret and OTP markers in exercised
   paths. Verify neither appears in response payloads, response/returned
   errors, captured diagnostics, nor formatted runtime values. Fixtures may
   retain their inputs outside captured output.
7. **Regression and scope.**  Existing IPC framing/allowlist and existing
   lease/TOTP behavior must retain their tests unchanged and pass. Confirm the
   diff is limited to the two owned paths, has no out-of-scope functionality,
   and introduces no socket/seed/listener/signal/push behavior.

## Independent QA execution

First inspect `git diff --name-only` and the candidate diff. Then run the
runtime mutation/lifecycle/registration/redaction tests by their exact names
or a focused package command, followed by the mandatory focused regression
command:

```sh
go test ./internal/backend ./internal/ipc ./internal/lease
go test -race ./internal/backend
```

QA independently runs the repository-native full checks once after focused
PASS:

```sh
GOCACHE="$(mktemp -d)" go test ./...
go vet ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
```

Timeouts in deterministic barrier tests are FAIL evidence; do not relax the
test or substitute timing-sensitive success. A missing `internal/backend`
package, unavailable toolchain/cache, unavailable repository-native command,
or sandbox restriction is recorded with its exact excerpt and classified
before retry. It does not silently waive focused evidence.

## SLOC and stop controls

Independently count nonblank, non-comment executable lines in candidate
production Go files (exclude `*_test.go`, generated, vendor, config, and task
documents), record the `runtime.go` subtotal and cumulative total, and retain
the command/output as QA evidence. Forecast is +150 production SLOC and
cumulative 901. A forecast or measured cumulative total above 925 stops and
returns the task for a revised PLAN and QA approval; 950 target and 1000 hard
guard also stop/split. Never accept compressed code, removed tests, or weaker
strict admission, close, concurrency, redaction, or denial behavior to fit a
cap. No Lap 3 is permitted.

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

## Failure classification and required return evidence

Record every QA attempt with start/end, `active_ms`, `wait_ms`, `retries`
(zero is explicit), source/command, exact concise evidence, and one of:

| Classification | Use when |
| --- | --- |
| `pass` | All P0 assertions, scope/SLOC controls, focused checks, and full checks pass. |
| `implementation_defect` | Candidate behavior or its tests violate a P0 assertion or regression. |
| `planning_defect` | Acceptance requires an excluded path, unapproved scope, or a correction to this PLAN/QA_PLAN. |
| `environment` | Toolchain, cache, sandbox, or repository-native check cannot execute independently of the candidate. |
| `regression` | A pre-existing IPC/lease or unrelated repository test fails and evidence separates it from the candidate. |
| `not_started` | A required sequential gate/preflight condition was absent before QA began. |

QA returns `PASS` only with all evidence present. On `FAIL`, return the failed
P0 row/check, classification (without presuming DEV fault), `active_ms`,
`wait_ms`, and retries; DEV remains closed until Main resolves the classified
cause. PLAN/QA contradictions found during execution are `FAIL` as a planning
defect, not an implementation instruction.

## PLAN reconciliation

After this independent baseline was fixed, `PLAN.md` was checked. Its fixed
ready/OTP routing, exact payload rules, lease/TOTP mapping, close and
concurrency requirements, bounded third-slot seam, redaction, regression
checks, role separation, focused/full checks, and 150/901/925/950/1000 SLOC
controls are consistent with this QA contract. No contradiction requiring a
PLAN FAIL or DEV closure was found.

## Revision 2 reconciliation

Revision 2 was assessed against the unchanged TASK-first baseline, the
existing `ipc` and `lease` APIs, and the stopped read-only draft. The
reconciliation result is **PASS for PLAN/QA approval**, not acceptance of the
draft: that draft remains non-gate-ready at 191 canonical production SLOC
with no `runtime_test.go` and must not be adopted as evidence.

The proposed simplifications are compatible with the P0 contract under these
fixed interpretations and checks:

- One exported production constructor, `New(secret)`, plus one unexported
  same-package clock-injected constructor/helper is sufficient. It must still
  build exactly one state and verifier, reject an empty secret generically,
  copy the secret through the existing verifier API, and expose no production
  clock or alternate constructor surface.
- Removing the active-call map, call IDs, and Close-time walk is acceptable
  when every admitted call derives from its caller context and registers
  exactly one `context.AfterFunc(shutdownCtx, cancelCall)` while holding the
  same close gate used by `Close`. The callback is a bounded per-call
  cancellation bridge, not a persistent runtime goroutine, timer loop, wait
  lifecycle, or authority registry.
- `context.AfterFunc` scheduling is asynchronous and is not, by itself,
  evidence of close ordering. Admission must linearize under the gate: a call
  that loses that gate race to `Close` cannot invoke a handler. A call admitted
  before `Close` is in flight; its caller-derived context must be cancelled,
  and a final gate-protected check of `closed` plus the original caller and
  shutdown cancellation state must discard its result if shutdown wins before
  publication. Cleanup cancellation must not be mistaken for caller
  cancellation and must not make every completed call fail.
- Removing duplicate pre-dispatch payload validation is acceptable because
  the selected fixed handler remains the single validation point. `ready`
  still accepts only zero-length payload, and the named-boundary OTP parser
  still accepts only the exact 17 bytes `{"code":"NNNNNN"}` with six ASCII
  digits. The fixed ready/OTP allowlist must be checked before registry lookup,
  so neither malformed input nor a registered third name reaches an
  unintended handler or lease transition.
- The unexported clock seam and same-package deterministic barriers are enough
  to exercise the built-in handlers without adding a production callback
  replacement, dispatching the third slot, or exporting test-only authority.
  Any additional seam must remain unexported, test-scoped, and within the two
  owned files.
- A readable 155–165 production-SLOC implementation is a supported forecast;
  +174 and cumulative 925 are absolute inclusive maxima. Runtime above 174 or
  cumulative above 925 stops and replans. The 950 target and 1000 hard guard
  remain later outer limits, not permission to cross 925. Semicolon packing,
  cryptic names, collapsed checks, removed comments/tests, or weakening any
  denial/race condition is a planning FAIL rather than an optimization.

DEV may resume the same Lap 1 only after Main approves PLAN Revision 2 and
this reconciliation, confirms the original branch/worktree/base and role
separation, and confirms the candidate has been reworked within the two owned
files. DEV must remove the stopped draft's active map and duplicate exported
constructor, create the required tests, measure `runtime.go` at no more than
174 canonical production SLOC/cumulative 925, and produce focused PASS plus
race PASS before REVIEW. No stopped-draft work is accepted retroactively.

Revision 2 adds no waiver to the mutation matrix. Required evidence includes
all ready and OTP byte mutations listed in P0 item 3, wrong version/unknown
operation, nil and pre-cancelled caller contexts, every Register bound, third
slot non-dispatchability/no push, state denial, payload-free decisions, and
redaction. Deterministic race evidence must separately prove:

1. Close linearizes before a waiting/new Handle admission: no handler starts
   and the response denies.
2. Handle is admitted before Close but blocked in a test seam: Close returns
   promptly, caller/shutdown cancellation reaches the call, no success is
   published, and repeated Close is safe.
3. Caller cancellation wins before admission, during an admitted handler, and
   before success publication: each denies without leaked detail or authority
   result.
4. Simultaneous valid OTP submissions accept at most one activation, and
   ready/OTP interleaving never consumes or clears a stale challenge.
5. `go test -race ./internal/backend` passes, followed by the unchanged
   focused regressions and full checks in this QA plan.

Failure of any ordering assertion is classified from its evidence; it is not
presumed to be a DEV defect. If the architecture cannot meet an assertion
readably within 174 SLOC, the minimum contradiction is the Revision 2
architecture/SLOC premise, classified `planning_defect`, and DEV closes for
replanning without weakening the TASK-first baseline.
