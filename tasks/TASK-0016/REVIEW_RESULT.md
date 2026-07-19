# REVIEW_RESULT — TASK-0016

## Decision: FAIL — candidate unavailable for independent review

This is an independent REVIEW-gate result, not an implementation finding. At
review time the supplied repository was `/home/ubuntu/git/agent-harness` at
`2a80256510901acbb063f12396fde152f5348de7` (`main`), with one worktree,
no `task/TASK-0016-*` branch, a clean index/worktree, and no diff. The four
contract paths are absent from that checkout:

- `internal/ipc/protocol.go`
- `internal/backend/runtime.go`
- `internal/ipc/protocol_test.go`
- `internal/backend/runtime_test.go`

Consequently, there is no candidate implementation or test diff to inspect,
and the named semantic evidence cannot be established. This is classified as
**environment/candidate-handoff failure**, not as a product implementation
defect. REVIEW cannot pass until Main supplies the completed candidate
worktree/commit in the review environment and requests a fresh independent
review.

## Required acceptance review

All of the following remain **NOT VERIFIED** because the candidate is absent:

| Required property | Result |
| --- | --- |
| Exact `ready` / `otp` / payload-free `authorize` protocol admission; malformed, unknown, payload-bearing, and wrong-version denial | Not verified |
| `Active()` decision at the handler decision point; no identity/deadline/reason payload | Not verified |
| `ready`/`otp` semantic and wire invariance; authorize has no readiness/OTP/lease mutation | Not verified |
| `maxOperations=4`: three fixed built-ins plus exactly one non-dispatchable custom registration slot | Not verified |
| Pre-activation, expiry boundary, fresh-runtime, Close, cancellation, and publication-race fail-closed behavior | Not verified |
| Deterministic expiry/close/cancellation race tests and `-race` coverage | Not verified |
| Named-test semantic coverage required by TASK/PLAN/QA_PLAN | Not verified |
| Four-path scope and production SLOC delta/cumulative cap | Not verified (empty diff; no candidate SLOC) |

## Evidence

The task contract, approved PLAN, and independent QA plan in this Task
directory were read. They consistently require the four paths above and the
listed acceptance evidence. `git worktree list --porcelain` showed only the
repository main worktree; `git status --short`, `git diff --name-only`, and
`git diff --numstat -- internal/ipc/protocol.go internal/backend/runtime.go`
were empty. `find internal` found no candidate source files.

Commands run in this review sandbox:

| Command | Result | Classification |
| --- | --- | --- |
| `make check` | PASS for the baseline Kakesu repository | Baseline-only; does not evidence TASK-0016 |
| `git diff --check` | PASS (no diff) | Baseline-only |
| `go test ./internal/backend ./internal/ipc ./internal/lease` | FAIL: no Go module at repository root | Environment/candidate handoff |
| `go test -count=1 -race ./internal/backend ./internal/ipc ./internal/lease` | FAIL: no Go module at repository root | Environment/candidate handoff |
| `go vet ./...` | FAIL: directory prefix does not contain a Go module | Environment/candidate handoff |
| targeted format/file-existence check | FAIL: both production target files absent | Environment/candidate handoff |

Main separately reported socket-capable evidence (`GOFLAGS=-buildvcs=false`
full `./...` PASS, backend/IPC/lease race PASS, and vet PASS). It is recorded
as external evidence only: it cannot replace this review sandbox's candidate
inspection, `make check`, or required review-gate proof.

## Accounting

`active_ms≈420000`; `wait_ms≈5000`; `retries=0`; classification
`environment/candidate-handoff failure`; null reason: the completed TASK-0016
candidate worktree/commit was not present or resolvable in this reviewer
runtime. No product files, Git state, staging, commit, merge, or operational
records were changed by this reviewer.

---

## Attempt 2 — supplied candidate worktree

## Decision: FAIL — scope violation and incomplete required race evidence

Candidate reviewed at `/tmp/codex-authority-broker-task0016`, branch
`task/TASK-0016-active-lease-authorization`. The four permitted implementation
and test files contain a small, readable authorization change, but this is not
gate-ready.

1. **Scope violation:** the candidate also modifies `backlog.json`. TASK-0016
   says that only the four metadata-named paths may change. Its 29-line
   backlog edit registers TASK-0016 and retargets TASK-0008, so it is not a
   mechanical consequence inside an allowed file. This must be removed from
   the candidate or explicitly approved as a Task scope change before REVIEW
   can pass.
2. **Required cancellation/publication-race evidence is missing.**
   `TestAuthorizeCallerCancellationFailsClosed` and both subtests of
   `TestAuthorizeExpiryAndCloseRaceFailsClosed` block at `Runtime.beforeGate`.
   That hook executes before `Handle` locks, selects the handler, or calls
   `handleAuthorize`; therefore these tests prove only a pre-admission race.
   They do not synchronize the required authorization decision-to-success
   publication boundary. The TASK/PLAN/QA_PLAN require a cancellation, Close,
   or expiry winner at that boundary to yield an empty-payload denial. Existing
   `handleAuthorize` cannot be deterministically paused by these tests, so the
   claimed post-handler fail-closed behavior is not independently established.

### Implementation inspection

The directly reviewed behavior otherwise aligns with the intended narrow
design: `authorize` is a fixed version-1 operation; IPC reader and writer
reject nonempty authorize payloads; the runtime independently repeats that
guard; its handler only calls `state.Active()` and returns a boolean; and final
publication rechecks active state, caller cancellation, and close/shutdown.
`ready` and `otp` code and wire shapes are unchanged. `maxOperations=4` gives
three fixed handlers plus one custom registration, and fixed-only Handle
admission keeps custom names non-dispatchable. Test helpers assert an empty
response payload. These observations do not cure the two FAIL findings.

### Attempt-2 checks

| Command | Result | Classification |
| --- | --- | --- |
| `go test -count=1 ./internal/backend -run 'TestAuthorize|TestRegisterBoundsAndFixedAllowlist'` | PASS | Candidate evidence |
| `go test -count=1 ./internal/ipc -run 'Test(ReadRequestRejectsMalformedFrames|AuthorizeProtocolAdmission|RequestRoundTripAndGenericErrors)$'` | PASS | Candidate evidence |
| `go test -count=1 -race ./internal/backend -run 'TestAuthorize|TestRegisterBoundsAndFixedAllowlist'` | PASS | Candidate evidence |
| `go vet ./...` | PASS | Candidate evidence |
| repository-wide `gofmt` check | PASS | Candidate evidence |
| `git diff --check` | PASS | Candidate evidence |
| full focused `go test ./internal/backend ./internal/ipc ./internal/lease` | IPC socket tests fail: Unix sockets are not permitted in this sandbox | Environment limitation; backend and lease passed |
| `make task-check TASK=TASK-0016` | FAIL: no such target | Repository tooling mismatch/environment |
| `make check` | FAIL: no such target | Repository tooling mismatch/environment |

The first focused run also failed before testing because the sandbox's default
Go build cache is read-only. It was retried once with a fresh `/tmp` GOCACHE;
the socket/tooling results above are from that retry. Main's separately
reported socket-capable full/race/vet PASS remains external corroboration and
does not replace this review's scope inspection or missing deterministic test
proof.

Production diff is `+22/-10` lines across the two allowed production files
(`+12/-3` runtime, `+10/-7` protocol), for a net `+12` textual line delta;
this is below the 1230 local trigger from the 1200 baseline and is readable.
The unapproved fifth changed path nevertheless prevents a scope PASS.

Attempt-2 accounting: `active_ms≈540000`; `wait_ms≈5000`; `retries=1`
(read-only Go cache, classified environment; changed precondition: isolated
`/tmp` cache); final classification `FAIL (scope violation; required test
evidence gap)`. No product file, Git state, staging, commit, merge, or
operational record was changed by this reviewer.

---

## Attempt 3 — focused corrected-candidate re-review

## Decision: PASS

The two Attempt-2 findings are corrected and no new blocking finding was
identified.

- **P1 scope reconciliation: PASS.** TASK, backlog index, PLAN, and QA_PLAN
  now consistently distinguish exactly four product/test paths from exactly
  six role/Main-owned process-evidence paths. The observed candidate changes
  are confined to that combined allowlist; `backlog.json` is explicitly
  approved process evidence and is excluded from production SLOC.
- **P2 publication-race proof: PASS.** The private, nil-by-default
  `beforePublish func(bool)` hook runs after the handler's authorization
  decision and before the final publication gate. The cancellation, exact
  expiry, and Close tests first assert that the authorize decision was
  positive, make the adverse event win while publication is blocked, release
  the call, and require an empty-payload denial. Source inspection confirms
  final rechecks of caller cancellation, close/shutdown, and authorize-only
  `state.Active()`. The hook adds no authority state or production goroutine.
- **Protocol/runtime semantics: PASS.** `authorize` remains version 1 and
  payload-free at both protocol and defensive runtime boundaries. It returns
  only the existing boolean response, queries the process-local lease, and
  does not call or alter ready/OTP challenge or activation paths.
  `maxOperations=4` retains exactly one custom registration slot while fixed
  Handle admission keeps that custom operation non-dispatchable.
- **SLOC/readability: PASS.** Independent nonblank/noncomment counts are
  runtime `171 -> 183` (+12) and protocol `117 -> 120` (+3), total **+15**;
  cumulative is **1215** from the 1200 baseline. This is below forecast and
  the 1230 trigger, with no compression.

### Attempt-3 checks

| Check | Result |
| --- | --- |
| `GOFLAGS=-buildvcs=false go test -count=1 -race ./internal/backend ./internal/ipc ./internal/lease` in socket-capable execution | PASS |
| `GOFLAGS=-buildvcs=false go test -count=1 ./...` in socket-capable execution | PASS |
| `go vet ./...` | PASS |
| repository-wide `gofmt` check | PASS |
| `git diff --check` | PASS |
| `jq -e . backlog.json` | PASS |
| exact product/process scope allowlist check | PASS |

Main's separately reported socket-capable full/race/vet PASS is consistent
with, but kept distinct from, this independent Attempt-3 execution. The absent
`make check` and `make task-check` targets were already recorded verbatim in
Attempt 2 as a repository-tooling condition; the underlying required focused,
race, full, static, format, JSON, diff, and scope checks passed here.

Attempt-3 accounting: `active_ms≈240000`; `wait_ms≈20000`; `retries=0`;
classification `PASS (prior scope and publication-race findings corrected)`.
No product file, Git state, staging, commit, merge, or operational record was
changed by this reviewer.
