# TASK-0016 QA plan — independent baseline

## Baseline and QA decision rule

This baseline is derived from TASK-0016 and the merged code/tests only, before
reading the implementation plan.  QA PASS requires every acceptance case below,
the stated focused and repository-wide checks, and an implementation inspection
that confirms the result; a passing test name alone is not evidence.

The candidate may change only:

- `internal/ipc/protocol.go`
- `internal/backend/runtime.go`
- `internal/ipc/protocol_test.go`
- `internal/backend/runtime_test.go`

The one new public operation is exactly `authorize`.  It is fixed and has no
payload.  Its response is the existing bounded `Response{OK: bool}` shape:
no lease ID, deadline, identity, token, reason, or other response payload.
It keeps protocol version `1`; a version change is a stop/replan condition.

## Acceptance-to-evidence matrix

| Contract outcome | Required named test evidence | QA source inspection |
| --- | --- | --- |
| The protocol admits exactly `ready`, `otp`, and payload-free `authorize`. | `TestReadRequestAuthorizeAdmission`, `TestWriteRequestAuthorizeAdmission`, and the existing `TestReadRequestRejectsMalformedFrames` extended for unknown/payload-bearing/wrong-version authorize frames. | Confirm both reader and writer use one exact fixed-operation allowlist, and that authorize payload is rejected at the protocol boundary. |
| A pre-activation authorization denies; a valid ready/OTP activation allows; the immutable deadline is exclusive. | `TestAuthorizeLeaseDecisionAtActivationAndDeadline`. Use the in-process fake clock: deny before activation, allow immediately after, allow at deadline-minus-one-nanosecond, and deny at the exact deadline and afterwards. | Confirm the decision queries the process-local lease at the handler decision point, rather than inferring readiness or OTP success, caching a result, or changing the deadline. |
| Restart, close, cancellation, and their races cannot publish allow. | `TestAuthorizeFailsClosedAfterFreshRuntimeAndClose`, `TestAuthorizeCallerCancellationFailsClosed`, and `TestAuthorizeCloseAndExpiryRacesFailClosed`. The race test must synchronize the authorize decision/publication boundary with Close and fake-clock expiry, then require `OK=false` for every winning close/expiry/cancel interleaving. | Confirm admission, handler context, and success publication retain the existing fail-closed gates; inspect for lock ordering/deadlock risks and for any allow after shutdown/caller cancellation. |
| `ready` and `otp` remain semantically unchanged. | `TestAuthorizeDoesNotMutateReadyOTPState`, plus the existing `TestReadyOTPExactAdmissionAndState`, `TestExpiredAndRateLimitedChallengesDeny`, and `TestReadyOTPBarrierDoesNotUseStaleChallenge`. The new test must show authorize neither creates/replaces a challenge, consumes OTP, extends/replaces a lease, nor changes subsequent ready/OTP outcomes. | Confirm authorize does not call readiness, verification, challenge mutation, or activation paths. |
| The existing custom-registration seam retains one usable slot. | `TestRegisterBoundsAndFixedAllowlist` updated so a single valid custom registration still succeeds after all three built-ins exist, while a second custom registration and dispatch still fail. | Confirm built-ins consume exactly three of the fixed operation capacity and custom handlers remain non-dispatchable through `Handle`. |
| Denials and allows disclose no state beyond the boolean. | Assertions in every new runtime test require an empty response payload; protocol tests cover no request payload. Existing `TestVersionContextAndRedaction` remains green. | Inspect response construction and error paths for no deadline/identity/reason/lease serialization or logging. |

## Required test shape and commands

Run these after DEV and again independently in QA:

```sh
go test ./internal/ipc ./internal/backend ./internal/lease
go test -race ./internal/ipc ./internal/backend ./internal/lease
make check
make task-check TASK=TASK-0016
git diff --check
```

The focused tests are additive to, not replacements for, the existing IPC,
runtime, and lease tests.  The fake-clock expiry test must not sleep.  Race
tests must use deterministic barriers/channels and timeouts only to detect a
deadlock; they must run under `-race`.  If this checkout does not provide a
listed wrapper command, record the command/environment failure verbatim and
classify it before retrying; do not silently omit the check.

## Boundary and failure criteria

Fail QA and require replan before accepting any candidate that introduces a
payload, protocol-version change, state-bearing response, persistence, sudo or
push coupling, lifecycle/seed/audit/release/canary work, or a changed path
outside the four-path allowlist.  Fail if unknown, malformed, wrong-version,
or payload-bearing authorize is accepted; if authorize affects ready/OTP state;
if it can allow at or after expiry; or if any close/cancel race can publish
allow.  Fail unreadable compression even when the line count passes.

## SLOC and Lap-1 gate

Baseline cumulative production SLOC is **1200**.  The Task forecast is +20,
therefore **1220** cumulative.  Count nonblank, non-comment executable
production-source lines in the two permitted production files only; tests and
this plan do not count.  Record the actual production delta and cumulative
total with the QA evidence.

- **1230** is the local trigger: stop/replan if the forecast or candidate rises
  above it; it is not permission to compress or weaken checks.
- **1250** is the Task target ceiling: a result above it requires approved
  replan and shedding review.
- **1350** is the absolute local guard: do not accept or continue with a
  candidate at or above/over that guard; stop and replan.

Lap 1 is one pass only: approved TASK-first PLAN and this QA plan, DEV,
independent REVIEW, independent QA, all required checks, and Main-owned Git
closure.  Lap 2 is exceptional only for one or two concrete REVIEW/QA findings
that require no redesign, research, Task redefinition, or fixture change and
are demonstrably reviewable inside the first 20 minutes.  There is no Lap 3.

## QA failure classification

Classify each failed check before retrying: **implementation defect** (code
violates this baseline), **QA-plan defect** (a necessary assertion/check was
missing or wrong), **requirement/scope conflict** (the Task requires a
forbidden boundary change), **environment/tooling** (reproducible independent
of the candidate), or **regression** (unrelated previously passing behavior).
Do not attribute a post-merge failure to DEV without this evidence.

## Execution accounting

Record active time separately from waiting time; record every retry with its
prior failure classification and the changed precondition.  Initial QA status:
`active=baseline authored`, `wait=0`, `retries=0`,
`classification=not-applicable (pre-DEV baseline)`.

## PLAN reconciliation — PASS

Reconciliation against `PLAN.md` is **PASS**.  The independent baseline and
PLAN have no acceptance, scope, security-boundary, or stop-rule conflict.  The
following minimal clarifications are appended without replacing the TASK-first
criteria above:

- Protocol admission is the same requirement under different proposed test
  names.  QA accepts the PLAN names `TestAuthorizeProtocolAdmission` plus the
  extended `TestRequestRoundTripAndGenericErrors` as the baseline's
  `TestReadRequestAuthorizeAdmission` / `TestWriteRequestAuthorizeAdmission`
  evidence only if inspection and assertions prove payload-free authorize is
  accepted by both read/write paths and every payload-bearing authorize is
  rejected before backend dispatch.  Wrong-version, malformed, and unknown
  operations remain covered by the extended
  `TestReadRequestRejectsMalformedFrames`.
- Runtime capacity must be implemented explicitly as `maxOperations=4`:
  exactly three fixed built-ins (`ready`, `otp`, `authorize`) plus exactly one
  valid custom registration.  Duplicate, second-custom/fifth-total, and
  post-close registration fail; the custom operation remains non-dispatchable.
- The PLAN's test names are accepted as direct equivalents when their asserted
  behavior matches the baseline: `TestAuthorizeActiveLeaseBoundary` maps to
  the activation/deadline test; `TestAuthorizeFreshRuntimeDenies` maps to the
  fresh-runtime case; `TestAuthorizePayloadAndReadinessOTPNonInterference`
  maps to the ready/OTP and payload invariants; and
  `TestAuthorizeExpiryAndCloseRaceFailsClosed` plus authorize-specific or
  extended cancellation/publication tests map to the close/cancel/expiry race
  cases.  Exact names do not excuse any omitted assertion.
- The race boundary is aligned: fake-clock advancement and deterministic
  channel barriers must show that expiry, Close, or caller cancellation that
  wins before publication yields an empty-payload denial.  `authorize` may
  only call the lease `Active()` decision and must not open/replace readiness,
  consume OTP, extend the lease, or mutate a challenge.
- SLOC semantics are aligned and fixed as: baseline 1200, forecast 1220;
  cumulative **>1230** stops for approved replan, 1250 is the Task target and
  not ordinary allocation, and cumulative **>=1350** is a hard stop.  QA also
  rejects unreadable packing or compression below those numbers.
- The PLAN's counted Lap-1 schedule is added to QA evidence: Main verifies both
  plans and starts DEV at minutes 0–5; DEV completes mapped checks by minute
  20 or stops for classification; independent REVIEW starts at minute 20;
  after REVIEW PASS, independent QA runs during minutes 20–30 before
  Main-owned closure.  Planning and pure wait are recorded separately and may
  consume at most 20% of the counted interval.  Lap 2 and no-Lap-3 rules remain
  those stated above.
- QA will also run the PLAN's uncached full regression, `go vet ./...`, and
  repository-wide `gofmt` check in addition to the baseline's broader focused
  race command, `make check`, `make task-check`, and `git diff --check`.

Reconciliation accounting:
`active=PLAN read and baseline reconciled`, `wait=0`, `retries=0`,
`classification=PASS (clarifications only; no QA-plan defect)`.

## Contract / PLAN / candidate reconciliation — re-review ready PASS

**Re-review readiness: PASS.**  This is a scope and proof reconciliation, not
the final independent QA execution or permission to publish.

- The candidate allowlist remains exactly four product/test paths:
  `internal/ipc/protocol.go`, `internal/backend/runtime.go`,
  `internal/ipc/protocol_test.go`, and `internal/backend/runtime_test.go`.
  The contract and index now separately allow exactly six process-evidence
  paths: `backlog.json`, `tasks/TASK-0016/TASK.md`,
  `tasks/TASK-0016/PLAN.md`, `tasks/TASK-0016/QA_PLAN.md`,
  `tasks/TASK-0016/REVIEW_RESULT.md`, and
  `tasks/TASK-0016/QA_RESULT.md`.  Those six are role/Main-owned evidence,
  not product scope or production SLOC; every other path remains forbidden.
- The candidate's private, nil-by-default `beforePublish func(bool)` seam is
  after the authorize handler decision and before final response publication.
  The cancellation, exact-expiry, and Close tests first observe a positive
  decision at that barrier, make the adverse event win, release publication,
  and require an empty-payload denial.  Source inspection confirms the final
  gate rechecks caller cancellation, closed/shutdown state, and—for authorize
  only—`state.Active()`, so a stale positive decision cannot publish.  The
  seam adds no authority state or goroutine and does not alter ready/OTP paths.
- Candidate production delta **+15** gives cumulative **1215** from baseline
  1200.  This is below the +20/1220 forecast and below the **>1230** replan
  trigger.  The 1250 target remains non-ordinary allocation and **>=1350**
  remains a hard stop; no cap or no-compression rule is relaxed.

No new contract, QA-plan, requirement, environment, or regression finding was
identified in this reconciliation.  The corrected candidate is ready for
independent re-review and its required checks.

Reconciliation accounting:
`active=candidate/contract inspection and QA append`, `wait=0`, `retries=0`,
`classification=re-review_ready_pass (prior process-scope/race-proof finding corrected)`.
