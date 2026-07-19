---
task_id: "TASK-0015"
qa_role: "independent QA (Terra/medium)"
decision: pass
classification: pass
completed_at: "2026-07-19T10:33:03Z"
---

# TASK-0015 QA result

## Decision — PASS

The corrected product satisfies the approved TASK-first QA matrix.  QA read
the complete broker candidate, final REVIEW append, and QA_PLAN; independently
executed every exact named broker test; and checked focused/race/full/static,
scope, metadata, and SLOC evidence.  The local runner denies Unix socket
creation, so its three socket-only exact tests are classified separately as
`environment_issue`.  They are not waived: Main supplied a post-correction
socket-capable PASS for broker/backend/IPC/lease, and the final independent
REVIEW records the same corrected socket-capable PASS.

No seed, decoded secret, OTP/TOTP value, token, credential, or production seed
content appeared in QA command output or this evidence.

## Acceptance evidence

| Area | Result | Evidence |
| --- | --- | --- |
| Immediate caller-buffer wipe | PASS | `TestRunWipesCallerSecretBeforeListen` passed. Source order is load -> factory -> wipe exact caller slice -> inspect result/listen; the listener barrier observes zeroed bytes, and factory-error input is also zeroed. |
| Descriptor admission and ownership | PASS | Every exact root/parent/final open, stat, close, reader-close, symlink, UID, type, bounds, short/read-error, and final-reader ownership test passed. Walk uses descriptor-relative no-follow/CLOEXEC; final wrapped reader owns one close. |
| Exact mode | PASS; P-003 closed | `TestLoadSeedRejectsNon0600Mode` passed while enumerating all 4096 `07777` combinations and accepting no value except `0600` for the permission/special-bit field. |
| Strict bounded schema and redaction | PASS | Valid/minimum/maximum, malformed, duplicate, unknown, missing, empty, wrong-type, trailing, UID, invalid/noncanonical base64, and oversized exact tests passed. `TestRunRedactsSecretFromErrorsAndLogs` passed; source exposes the generic seed sentinel only. |
| Runtime/listener/serve/close ordering | PASS | Factory failure, construction/config before listen, listener error with returned-server-before-runtime cleanup, unexpected Serve, and close-error exact tests passed. |
| Signals and shutdown | PASS | SIGINT, SIGTERM, active client barrier, concurrent cancellation, and repeated/idempotent close exact tests passed; race package test passed. |
| Restart | PASS; P-004 closed | Fresh restart uses distinct seed documents, runtimes, servers, and retained private copies. Missing-seed test performs a successful first run, then proves zero factory/listen calls. Aggregate sibling restart runs even when its socket sibling skips. |
| Socket replacement and existing client | PASS with local environment null | Local replacement, valid-client, and malformed-client exact tests skip only because `ipc.Listen` returns `ipc: server unavailable`; focused/full IPC failures expose `socket: operation not permitted`. Main's post-correction socket-capable broker/backend/IPC/lease run passes all three branches. |
| P-001 client correction | CLOSED | Corrected client integration uses real `ipc.Client`, `ipc.Listen`, framing, peer UID, Serve, and Close with a deterministic syntax backend. Backend and lease regressions independently pass for real OTP/replay semantics. |
| P-002 aggregate proof | CLOSED | PLAN aggregates execute named sibling subtests for every descriptor/schema/construction/lifecycle/restart/client row; local output proves socket skips do not suppress restart. |

## Independent commands

| Check | Result / classification |
| --- | --- |
| Anchored selection of every QA and PLAN exact broker test | PASS; only three socket-only tests SKIP with `ipc: server unavailable` (`environment_issue`). All non-socket names and aggregate non-socket siblings PASS. |
| `GOFLAGS=-buildvcs=false GOCACHE=/tmp/task0015-qa-gocache go test -count=1 -race ./cmd/codex-authority-broker` | PASS, no race. |
| Focused CLI/broker/backend/IPC/lease with fresh cache | Broker/backend/lease PASS; CLI and IPC FAIL only at Unix socket creation/server availability (`environment_issue`). Main's socket-capable corrected focused evidence PASSes broker/backend/IPC/lease. |
| Full `GOFLAGS=-buildvcs=false ... go test -count=1 ./...` | Same local CLI/IPC socket denial; all other packages PASS (`environment_issue`, not regression). |
| Focused command with default cache | `environment_issue`: default Go cache is read-only. The fresh-cache execution above is controlling. |
| `go vet ./...` with fresh cache/VCS disabled | PASS. |
| `gofmt -l` for broker production and test files | PASS, no output. |
| `git diff --check` | PASS, no output. |
| `make check` | `environment_issue`: repository has no `check` target; not used as product evidence. |
| `jq -e . backlog.json` and canonical TASK-0015 metadata comparison | PASS; backlog entry equals the TASK JSON contract. |

## Scope, source, and limits

- Product candidate is confined to
  `cmd/codex-authority-broker/main.go` and `main_test.go`; no runtime, IPC,
  backend, lease, CLI, or exported API product diff exists.  Unrelated
  pre-existing backlog/TASK changes were not attributed to DEV or modified by
  QA.
- Canonical nonblank/noncomment broker production SLOC is **278**; cumulative
  is **1200** from merged baseline 922.  This is within +280/cumulative 1202,
  target 1250, hard guard 1350, global target 1500, and global hard 1800.
  Production is 307 physical lines; tests are 1057 physical lines, with no
  compression or generated disguise observed.
- P-001 through P-004 are closed by the bounded test-only correction and final
  REVIEW PASS.  `main.go` remained unchanged during that correction.

## Classification and accounting

- Overall: `pass`.  No `implementation_defect`, `regression`, requirement gap,
  or QA-plan defect remains.
- Separate environment nulls: read-only default Go cache; absent `make check`
  target; local Unix socket `EPERM`/server-unavailable.  Socket acceptance is
  proven by the supplied post-correction socket-capable PASS.
- `active_ms=unavailable`: the QA runtime exposes no reliable turn-start
  timestamp, so elapsed work is not inferred.  `wait_ms=0`; `retries=0`.
- QA-owned changed path: `tasks/TASK-0015/QA_RESULT.md` only.  QA performed no
  product, Git, staging, commit, merge, or operations write.
