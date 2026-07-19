---
task_id: "TASK-0013"
status: complete
reviewer_agent: "/root/task_0013_review (Terra/medium)"
reviewed_commit: "df140ef884279b04a027bd8f6c23344ecfba3650 (candidate files untracked)"
decision: pass
make_check: environment_unavailable
reviewed_at: "2026-07-19T07:55:17Z"
---

# TASK-0013 REVIEW RESULT

## Attempt 2 — PASS

**PASS (`pass`, retries 0).** R-001 is resolved. The revised candidate adds a
minimal unexported `beforeGate` test hook and three deterministic ordering
tests. The hook is nil/no-op in production construction, runs only after the
fixed version/operation allowlist and immediately before the close gate, and
does not read or mutate authority, select a handler, or make a dispatch
decision. Tests configure it before concurrent use; independent race runs are
clean.

### R-001 resolution

| Required evidence | Attempt 2 result |
|---|---|
| Caller cancellation during an admitted handler | PASS: `TestCallerCancellationReachesAdmittedHandler` blocks inside the admitted handler, cancels the caller, observes handler context cancellation, and asserts denial. |
| Caller cancellation after handler return but before publication | PASS: `TestCallerCancellationWinsBeforeSuccessPublication` holds the final runtime gate, observes handler return, cancels the caller, then releases the publication check and asserts denial. |
| Close wins against a waiting/new Handle | PASS: `TestCloseWinsWaitingHandleGate` blocks before gate acquisition, lets `Close` linearize, releases Handle, asserts denial, and proves the handler did not start. |

### Attempt 2 independent evidence

| Command | Result | Evidence / classification |
|---|---|---|
| `go test -count=100 ./internal/backend` | PASS | `ok .../internal/backend 0.043s`. |
| `go test -race -count=10 ./internal/backend` | PASS | `ok .../internal/backend 1.035s`. |
| `go test ./internal/backend ./internal/ipc ./internal/lease` | PASS | backend 0.004s; IPC and lease cached. |
| `GOFLAGS=-buildvcs=false GOCACHE=$(mktemp -d) go test ./...` | PASS | cmd 0.307s; backend 0.003s; IPC 0.043s; lease 0.003s. Socket-capable execution. |
| `go vet ./...` | PASS | no output. |
| `gofmt -l $(find cmd internal -type f -name '*.go' -print)` | PASS | no output. |
| `git diff --check` | PASS | no output; candidate files remain untracked and were inspected directly. |
| `make check` | ENVIRONMENT | Re-run result: `make: *** No rule to make target 'check'.  Stop.` No Makefile/target exists in the supplied worktree; full native Go/static checks above pass. |

- Canonical production SLOC: runtime **171**, cumulative **922**, within the
  inclusive Revision 2 maxima 174/925. Test file: **457 physical LOC**.
- Scope remains the two candidate backend files plus Task evidence/docs
  already present as untracked files. REVIEW modified only this result.
- Attempt 2 accounting: `active_ms=unavailable` (runtime exposed no attempt
  start timestamp; not inferred), `wait_ms=0`, `retries=0`,
  `classification=pass`. The `make check` sub-result remains `environment`.

## Attempt 1 history — FAIL (resolved by Attempt 2)

### Decision

**FAIL — implementation_defect.** The implementation behavior and the executed
checks pass, but the candidate test suite does not provide the deterministic
caller-cancellation evidence explicitly required by approved PLAN Revision 2
and QA_PLAN Revision 2. No product, test, PLAN/QA_PLAN, operational-log, or
Git file was changed by REVIEW; this file is the sole review output.

## Reviewed scope and measurement

- Worktree / branch / base: `/tmp/codex-authority-broker-task0013`,
  `task/TASK-0013-runtime-assembly`, `df140ef`.
- Candidate status before evidence: untracked `internal/backend/runtime.go`,
  `internal/backend/runtime_test.go`, plus local untracked PLAN/QA_PLAN
  copies. `git diff --name-only` was empty because the candidate production
  files are untracked; review inspected both files directly. No tracked
  production path was modified.
- Canonical executable SLOC: `internal/backend/runtime.go` **167**;
  all candidate production Go files **918**. Both satisfy the Revision 2
  maxima of 174 and 925. `runtime_test.go`: **346 physical LOC**.
- Readability: ordinary named constants/bounds and explicit error paths; no
  semicolon packing or compressed active-call registry was found.

## Independent checks

| Command | Result | Evidence / classification |
|---|---|---|
| `go test ./internal/backend` | PASS | `ok .../internal/backend 0.004s`; initial sandbox attempt could not write Go build cache, then safe approved execution passed. |
| `go test -count=100 ./internal/backend` | PASS | `ok .../internal/backend 0.030s`. |
| `go test -race ./internal/backend` | PASS | `ok .../internal/backend 1.014s`. |
| `go test ./internal/backend ./internal/ipc ./internal/lease` | PASS | backend cached; IPC 0.042s; lease 0.002s. Ran in approved socket-capable environment. |
| `GOFLAGS=-buildvcs=false GOCACHE=$(mktemp -d) go test ./...` | PASS | cmd 0.339s; backend 0.003s; IPC 0.043s; lease 0.002s. Socket-capable full suite executed. |
| `go vet ./...` | PASS | no output. |
| `gofmt -l $(find cmd internal -type f -name '*.go' -print)` | PASS | no output. |
| `git diff --check` | PASS | no output; note untracked candidate files are not represented by Git diff. |
| `make check` | ENVIRONMENT | FAIL: `make: *** No rule to make target 'check'.  Stop.` The worktree has no `Makefile`; this is repository-native command unavailability, not attributed to runtime code. |

## Contract review

| Condition | Result | Evidence |
|---|---|---|
| Exact ready/OTP routing and payload-free decisions | PASS | Fixed version and operation allowlist precede lookup; ready accepts only zero bytes; `decodeOTP` admits only the 17-byte ASCII form. Rejection returns `ipc.Response{OK:false}`, nil error, and no payload. |
| Bounded instance-local registry / third nondispatch | PASS | Exactly ready and otp installed; Register is mutex-guarded, max 3, validates name/handler/duplicate/closed; Handle only dispatches fixed allowlist first. |
| Lease/challenge state and replay denial | PASS | Dedicated challenge lock protects ready/OTP challenge operations; existing lease verifier owns transition/replay/rate semantics. Focused IPC/lease regressions pass. |
| Close linearizability and redaction implementation | PASS by code inspection and existing close test | One gate marks closed and cancels shutdown before release; admitted calls use `AfterFunc`, then final gate-protected close/caller check discards success. No logging or exposed authority fields; String is constant. |
| Required deterministic caller-cancellation test evidence | **FAIL** | `TestCallerCancellation` covers only an already-cancelled context before admission. It does not prove cancellation during an admitted blocked handler or cancellation after handler return but before success publication, both expressly required by PLAN R2 and QA_PLAN R2. |
| Required waiting-Handle-vs-Close ordering evidence | **FAIL** | Existing close test proves an already-admitted blocked handler is cancelled and Close returns promptly, but does not deterministically hold a Handle at the gate, let Close linearize first, and prove that the handler never starts. This is separately required by QA_PLAN R2. |

## Finding

| ID | Severity | Classification | Finding |
|---|---|---|---|
| R-001 | blocking | implementation_defect | Mandatory deterministic cancellation/order test coverage is absent: no caller cancellation during handler, no caller cancellation before publication, and no waiting/new Handle that loses the close-gate race. Passing race and repetition runs do not substitute for those specified assertions. |

## Timing and evidence accounting

| Stage | active_ms | wait_ms | retries | classification | null reason |
|---|---:|---:|---:|---|---|
| Independent review, inspection, and checks | unavailable | 0 | 0 | implementation_defect | Runtime did not expose a review-stage start timestamp; duration is not inferred. |
| `make check` | unavailable | 0 | 0 | environment | Runtime did not expose a command-stage start timestamp; Makefile/`check` target unavailable in supplied worktree. |

Tool/approval waits were not separately observable in this reviewer runtime;
recorded `wait_ms` is therefore 0 rather than inferred. The source of truth
for command output is this review run at the timestamp above.
