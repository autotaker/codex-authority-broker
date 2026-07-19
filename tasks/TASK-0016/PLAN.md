# PLAN — TASK-0016: bounded active-lease authorization operation

## Decision and boundary

**PLAN evidence prepared — Main approval of this PLAN and an independent,
TASK-first `QA_PLAN.md` remain prerequisites for any implementation.** This
specifies a single counted 1-Lap delivery; it does not approve DEV, Git
operations, operational-log writes, or a merge. DEV uses `dev-luna` /
`luna-xhigh`; Planner, REVIEW, and QA use separate `Terra/medium` roles.
Children must neither stage, commit, merge, nor write `.git`; Main alone owns
the shared lock, publication, and merge.

Only these paths are eligible for the candidate:

| Path | Owned change |
| --- | --- |
| `internal/ipc/protocol.go` | Declare and strictly admit fixed `authorize`. |
| `internal/ipc/protocol_test.go` | Prove exact protocol admission and payload rejection. |
| `internal/backend/runtime.go` | Install and route built-in payload-free authorization. |
| `internal/backend/runtime_test.go` | Prove lease, non-interference, capacity, cancellation, close, and races. |

No client/daemon/server/lease API, protocol version, persistence, identity or
deadline disclosure, sudo/policy, seed, push, credential, audit, installer,
release, or canary change is permitted. Do not modify the meaning or wire
encoding of `ready` or `otp`. A need for excluded behavior is a pre-DEV stop
and split/replan, not an implementation shortcut.

## Fixed implementation contract

1. Add exactly `ipc.OperationAuthorize = "authorize"` to fixed operation
   admission. `readRequest` and `writeRequest` continue to reject unknown and
   wrong-version requests and must reject every payload-bearing `authorize`
   request before backend dispatch. `authorize` has no request schema; its
   runtime handler repeats the zero-length guard defensively. Every response
   remains the existing payload-free `ipc.Response{OK: bool}` and returns no
   identifier, deadline, token, identity, reason, or state payload.
2. Increase the built-in handler set from two to three and increase
   `maxOperations` from 3 to **4**. The capacity is exactly three fixed
   built-ins plus one custom registration slot. That one custom operation
   remains non-dispatchable because `Handle` admits only the fixed built-in
   allowlist; a second custom registration (fifth total operation) must fail.
   Do not turn registration into dynamic dispatch or replacement.
3. `Handle` admits `authorize` under the existing version, caller-context,
   shutdown, and post-handler publication gate. Its handler accepts only
   zero-length payload and returns true only when its context is live and
   `r.state.Active()` reports an active process-local lease at that decision
   point. It must not call `BeginReadiness` or `VerifyAndActivate`, mutate the
   stored challenge, consume OTP, or extend a lease. Nil state, absent/expired
   lease, cancellation, closed runtime, and failed final publication deny with
   an empty payload.
4. Preserve existing mutex/shutdown ordering. A call cancelled or losing to
   `Close` after handler admission must discard a positive result. Tests use
   fake-clock advancement and channel barriers—not sleeps—to prove expiry,
   close, and caller-cancellation races. A newly constructed runtime owns a
   fresh idle `lease.State` and cannot authorize an old process lease.

## Acceptance-to-test map

DEV must add explicit named cases (or retain these names if a table is split).
QA_PLAN, REVIEW, and QA must map evidence to these conditions and inspect
behavior, not merely test names.

| Acceptance condition | Required named test evidence |
| --- | --- |
| Exact operation admission; malformed, unknown, payload-bearing, and wrong-version requests fail closed. | Extend `TestReadRequestRejectsMalformedFrames` with authorize-payload/version cases; add `TestAuthorizeProtocolAdmission`; extend `TestRequestRoundTripAndGenericErrors` for payload-free authorize round-trip and invalid write rejection. |
| Before activation deny; after valid activation allow; deadline and after deny. | Add `TestAuthorizeActiveLeaseBoundary` with `testClock`: before activation false, after OTP true, deadline-minus-1ns true, deadline/after false. |
| Fresh runtime denies. | Add `TestAuthorizeFreshRuntimeDenies` after activation of a distinct runtime. |
| No payload/state side effect; ready/otp unchanged. | Add `TestAuthorizePayloadAndReadinessOTPNonInterference`: payload variants deny; authorize neither opens/replaces challenge nor consumes OTP/extends expiry; subsequent ready/OTP behavior remains correct; every response is payload-free. |
| Close, pre-cancelled caller, and cancellation/publication race cannot allow. | Extend `TestCallerCancellation`, `TestCallerCancellationWinsBeforeSuccessPublication`, and `TestCloseCancelsAdmittedCallAndFailsClosed`, or add authorize-specific counterparts, with barriers and `OperationAuthorize`. |
| Concurrent authorization with expiry/close is race-clean and fail closed. | Add `TestAuthorizeExpiryAndCloseRaceFailsClosed` using fake clock/barriers; run `go test -race ./internal/backend`; no positive result publishes after expiry/close wins. |
| Three built-ins retain exactly one custom registration slot. | Amend `TestRegisterBoundsAndFixedAllowlist`: with `maxOperations=4`, exactly one valid custom registration succeeds after the three built-ins; duplicate, second-custom/fifth-total, and post-close registrations deny; the custom name remains non-dispatchable. |
| Focused/full regressions, scope, and cap pass. | Commands below; REVIEW and QA independently inspect both production files, payload absence, and SLOC. |

## Counted 1-Lap execution and evidence

Preflight records merged TASK-0015 base, approved PLAN plus independent
QA_PLAN, four eligible paths, Go availability, and fake-clock/temporary
Unix-socket fixture availability. No elevated fixture is required. A predictable
missing tool/socket condition is classified once as `environment` with a
redacted null reason; it is not blindly retried or treated as product PASS.

| By minute | Gate and completion evidence |
| ---: | --- |
| 0–5 | Main verifies both plans and starts DEV **within five minutes**. DEV adds constant/allowlists and minimal handler, then runs focused IPC/backend/lease tests. |
| 5–20 | DEV completes mapped tests, barrier races, formatting, SLOC, diff, race, static, and full-regression evidence. Required unresolved work at minute 20 stops for classification, not compression. |
| 20 | Main starts independent REVIEW **at minute 20** on the complete candidate. REVIEW inspects payload-free fixed contract, `Active()` decision point, ready/otp invariance, registration capacity, close/cancel/expiry races, SLOC, and scope. |
| 20–30 | After REVIEW PASS, independent QA executes mapped acceptance and regressions. Main alone performs scope/hook/Git closure after both PASS results. |

DEV first runs, and REVIEW/QA independently rerun their focused/full checks once:

```sh
go test ./internal/backend ./internal/ipc ./internal/lease
go test -count=1 -race ./internal/backend
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./...
go vet ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
make check
make task-check TASK=TASK-0016
```

If a repository target is absent, exact output is evidence and classified; it
is not silently omitted. REVIEW must complete `make check`. QA treats a
post-merge failure under the QA guideline (`implementation_defect`,
`qa_plan_defect`, `requirement_gap`, `environment_issue`, or `regression`);
Main makes final attribution.

## SLOC, stops, and retry discipline

Immutable baseline is merged TASK-0015 cumulative **1200**. Forecast **+20**
ordinary readable production SLOC gives **1220**. Local forecast/replan trigger
is **1230**; **1250** is Task target; **1350** is absolute hard guard. DEV,
REVIEW, and QA independently count nonblank, non-comment executable production
Go lines (tests/documents excluded), report per-file/cumulative totals, and
reject semicolon packing, merged error paths, cryptic naming, deleted comments,
or artificial function compression.

At forecast or candidate **>1230 cumulative**, stop before further DEV and
obtain approved replan; 1250 target and 1350 hard guard are not ordinary
allocation. At or above 1350, hard-stop; never borrow later reserve. Also stop
before DEV for an authorization payload, returned identity/deadline or other
payload, protocol-version change, persistence, sudo/push coupling, or loss of
custom registration capacity. `ready`/`otp` drift, weakened expiry, or a
non-gate-ready required test is likewise a stop/split condition.

Lap 2 is exceptional only when all four facts are recorded: exactly one or two
concrete REVIEW/QA findings; no redesign, research, Task redefinition, or
fixture change; a bounded residual estimate; and a demonstrably reviewable
correction within Lap 2's first 20 minutes. Otherwise split. No Lap 3 exists.

For every PLAN/DEV/REVIEW/QA attempt, Main records UTC start/end, `active_ms`,
`wait_ms`, retry count (including zero), classification, source/command
evidence, and redacted null reason where applicable. Planning and pure wait
consume at most 20% of the counted interval; contingency applies to observed
active time only, never SLOC. This Planner attempt has no reliable turn-start
timestamp: `active_ms=unavailable`, `wait_ms=0`, `retries=1`, classification
`planning_defect_corrected`; null reason: planner runtime start time unavailable.

## Planner evidence

| Item | Evidence |
| --- | --- |
| Source inspected | TASK-0016 contract; applicable `AGENTS.md`; delivery skill; development, role, and QA guidance; current protocol/runtime/lease code/tests; TASK-0015 plan. |
| Current design fact | IPC and runtime dispatch allow only `ready`/`otp`; runtime currently installs two handlers with `maxOperations=3`; `lease.State.Active()` expires at its immutable deadline under mutex. |
| Corrected capacity invariant | The candidate changes `maxOperations` to 4: three fixed built-ins including payload-free `authorize`, plus exactly one custom registration slot; custom names remain non-dispatchable and a second custom registration fails. |
| Correction classification | Main preflight found the original `three built-ins + maxOperations=3 + one custom slot` contradiction; corrected as `planning_defect`, retry 1, before DEV. |
| State | PLAN evidence only; DEV, approval, merge, Git, and operational-log actions have not occurred. |

## REVIEW P1 reconciliation — process evidence and race seam

**PASS (`planning_defect_corrected`).** Main has reconciled TASK-0016's index
and contract with an explicit `process_evidence_paths` allowlist. The product
and test candidate remains restricted to exactly the four metadata paths:

- `internal/ipc/protocol.go`
- `internal/backend/runtime.go`
- `internal/ipc/protocol_test.go`
- `internal/backend/runtime_test.go`

Main/role-owned process evidence may be created or updated in the same Task PR
only at these exact six paths:

- `backlog.json`
- `tasks/TASK-0016/TASK.md`
- `tasks/TASK-0016/PLAN.md`
- `tasks/TASK-0016/QA_PLAN.md`
- `tasks/TASK-0016/REVIEW_RESULT.md`
- `tasks/TASK-0016/QA_RESULT.md`

Those six files are process evidence, not product scope, and are excluded from
production SLOC. Every other product, test, or process path remains forbidden.
This reconciliation does not transfer Git, approval, merge, or shared-lock
ownership from Main, and does not permit a role to write another role's
evidence.

The candidate's private `Runtime.beforePublish func(bool)` test seam is within
the owned `internal/backend/runtime.go` path and is compatible with the PLAN's
deterministic publication-race proof. It gives barrier tests a precise point
after the handler decision and before final publication checks; expiry,
`Close`, or caller cancellation can then win, and the existing final closed,
shutdown, caller, and authorize-only `state.Active()` recheck must discard the
earlier positive result. The seam remains unexported, nil in production, adds
no goroutine or authority state, and must not change `ready`/`otp` results.

The observed readable production delta is **+15**, giving cumulative
**1215 = 1200 + 15**. It is below both the +20/cumulative-1220 forecast and the
**1230** replan trigger, so it fits the approved SLOC and race-proof envelope
without compression. The 1250 target and 1350 hard guard remain unchanged and
provide no additional ordinary allocation.

Reconciliation accounting: `active_ms=unavailable`, `wait_ms=0`, cumulative
`retries=2`, classification `planning_defect_corrected`; null reason for active
time: planner runtime start time unavailable. This is PLAN evidence only; no
product, Git, operational-log, approval, or merge action is authorized here.
