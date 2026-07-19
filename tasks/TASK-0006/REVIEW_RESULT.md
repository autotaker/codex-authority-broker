# REVIEW RESULT — TASK-0006

## Decision: PASS (revision 2)

The first independent review failed with a `measurement_defect`: TASK-0001
and TASK-0003 cycle values inferred preflight exclusion from the absence of a
preflight event. DEV changed those cycle fields to `null` and retained their
Lap elapsed observations only as provenance. Independent re-review confirms
the finding is resolved with no regression or open finding.

## Evidence

- The selection-time 83-line snapshot and its hash reproduce; filtering the
  append-only log yields the stable 72 events for four completed Tasks.
- Schema, event IDs, per-Lap sequences, completion counts, stage arithmetic,
  retries, SLOC/test fields, and correction targets reproduce.
- TASK-0005 raw history retains the superseded `requirement_gap` edge while
  its effective classifications are `planning_defect` and
  `environment_issue`.
- TASK-0001 and TASK-0003 cycle fields are null because preflight exclusion is
  not provable. TASK-0004/0005 cycles remain 3,779,000/2,575,000 ms.
- The TASK-0004 analogue gives
  `ceil((3779000 / 60000) * 1.20) = 76` minutes. No fixed SLOC throughput is
  assumed.
- Only TASK-0007, TASK-0008, and immediate TASK-0009 are executable/detailed.
  Dependencies, 950/1400/1400 ceilings, retained reserves, global 1500 cap,
  mandatory controls, and exact shedding order are consistent.
- Production remains 751 SLOC; TASK-0006 adds zero and has no `.go` diff.
  JSON parsing and `git diff --check` pass.

Initial review timing was approximately `active_ms=239000`, `wait_ms=0`.
Re-review timing was approximately `active_ms=69000`, `wait_ms=0`. The prior
measurement defect remains historical resolved evidence. The reviewer changed
no file, Git state, or Lap30 log.

## Revision 8 independent review — FAIL

### Finding

- **P1 `qa_plan_defect` — stale canonical-row assertion.**
  `tasks/TASK-0006/QA_PLAN.md:599` states that the Revision 8 predicate ran on
  "all 136 rows."  The read-only canonical event stream did not match that
  assertion during this review: it contained 142 rows for the correction check
  and 144 rows on the subsequent formatting/evidence check.  The executable
  correction predicate itself passed at 142 rows (10 correction events), so
  this is not a DEV/index/contract/product defect; however the approved QA
  evidence cannot claim an exact row count that is not reproducible from the
  referenced source.  Freeze and identify the intended snapshot, or correct
  the assertion and reapprove the QA plan before QA relies on it.

### Passed independent evidence

- DEV scope is exactly `backlog.json` and `tasks/TASK-0007` through
  `TASK-0012` `TASK.md`; the separate PLAN/QA_PLAN edits are gate artifacts.
  No production-source path is changed.
- `jq -e . backlog.json`, `git diff --check`, canonical JSON parsing,
  unique-event-ID validation, completed-set validation, and the executable
  same-task/same-lap/earlier-file-order/smaller-sequence correction predicate
  all passed (10 corrections at the checked snapshot).
- Each of the six embedded contract metadata objects exactly equals its
  `backlog.json` entry. IDs/dependencies form the expected linear chain;
  90% triggers, target/hard guards, and both reserved milestones validate.
- Arithmetic validates: `751 + 90 + 120 + 0 + 130 + 220 + 0 = 1311`, with
  reserves 189 to 1500 and 489 to 1800. Production source remains exactly 751
  nonblank/non-comment non-test Go lines; source diff is `0/0` lines.
- No fixed-throughput sizing, compressed-source allowance, or product
  implementation was introduced. `gofmt -l` is empty and `git diff --check`
  passes.
- `make check` fails because this repository has no `check` target. This is a
  repository-command mismatch, not a product failure; the documented native
  full suite was used instead. The first sandboxed `GOCACHE=$(mktemp -d) go
  test ./...` was blocked by Unix-socket creation (`operation not permitted`),
  classified `environment_issue`; one permitted socket-capable rerun passed
  `cmd/codex-authority`, `internal/ipc`, and `internal/lease`.

`active_ms≈540000`; `wait_ms≈12000`; `retries=1` (environment rerun only);
classification: `qa_plan_defect` for the finding and `environment_issue` for
the sandbox-only Unix-socket restriction. No candidate contract, product,
operational log, or Git state was modified by this reviewer.

## Revision 8.1 independent re-review — PASS

### Findings

No open findings. The sole Revision 8 P1 `qa_plan_defect` is resolved.

QA_PLAN Revision 8.1 now attributes the earlier predicate result to an exact
frozen prefix rather than treating 136 as a live-stream invariant. Independent
reproduction confirmed all three identifiers:

- prefix rows: `136`;
- final prefix event ID: `task0006-r3-lap02-replan8-start`;
- SHA-256: `9b01bb157c666aba691a87665bbdef5fb06ae6ba7359198e2e0fda7ceb3dbfc1`.

That prefix passes JSON parsing and the same-task/same-lap/earlier-file-order/
smaller-sequence correction predicate. The append-only stream contained 146
rows during this re-review; Revision 8.1 correctly describes live counts as
observations rather than acceptance values and requires each future gate to
freeze and identify its complete available snapshot.

Regression checks confirm the seven DEV outputs are unchanged in substance:
all six embedded contract metadata objects still exactly match `backlog.json`;
the 751 -> 1311 arithmetic and 189/489 reserves still pass; production source
remains exactly 751 SLOC with no source diff; and `git diff --check` passes.
The full Go suite was not rerun because the candidate did not change and the
Revision 8 permitted run already passed.

`active_ms≈55000`; `wait_ms=0`; `retries=0`; classification: `none`. This
reviewer changed only this review evidence and did not modify candidate
contracts, product source, the operational repository/log, or Git state.
