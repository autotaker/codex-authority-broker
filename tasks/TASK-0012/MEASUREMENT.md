# TASK-0012 final zero-SLOC measurement

## Frozen sources and integrity

This evidence-only gate changes no product, test, configuration, dependency,
schema, generated artifact, or canonical log.

| Source | Identity | Result |
| --- | --- | --- |
| Post-TASK-0017 canonical snapshot | `/tmp/task0012-events-frozen.jsonl`; SHA-256 `00482bd0db254b700848999d79e1cd67fe5dd3a0e5064b0847d92bbe7f932e05`; 404 lines; captured 2026-07-19T22:16:03Z | JSON parse, unique event IDs, and correction predicate PASS |
| Prior TASK-0009 snapshot | `/tmp/task0009-events-frozen.jsonl`; SHA-256 `abe31c9da1fbcfd32daef6013a9ce58063ccbb086265b4760ca20a22962af09b`; 389 lines | Historical cohort and timing source retained by reference |
| Prior measurement report | `tasks/TASK-0009/MEASUREMENT.md`; SHA-256 `f72f53d6f8d1f27f0d6e74accb91f7794f0143b764b23fc11f7dfd7bb53e689c` | Historical rows unchanged |
| TASK-0017 merged status | PR #18; merge `212fd03492d72a38970e54b6a303806a04c9c7c5`; tree `eaf0d3710205bf5cf577c5cd785eb85a4932b773` | Independent REVIEW and QA PASS |

Every correction in the 404-line snapshot has a nonempty target that occurs
earlier in file order, has the same task and lap, and has a smaller sequence.
Raw rows remain immutable; only predicate-valid corrections affect effective
values. Corrections are not terminals or additional tasks.

The stream contains no TASK-0017 event. The Lap30 contract forbids a
synthetic terminal after completion, so Git/PR/REVIEW/QA prove status while
all TASK-0017 timing and classification fields below remain reasoned null.

## Actual production and test LOC

The counter includes every nonblank, non-comment line in non-test production
Go files under `cmd` and `internal`. There are no shipped production Python,
shell, or C files; declarative sudoers text is excluded by the project
definition.

| Production file | merged SLOC |
| --- | ---: |
| `cmd/codex-authority-broker/main.go` | 282 |
| `cmd/codex-authority-sudo/main.go` | 188 |
| `cmd/codex-authority/main.go` | 83 |
| `internal/backend/runtime.go` | 183 |
| `internal/ipc/client_linux.go` | 35 |
| `internal/ipc/protocol.go` | 120 |
| `internal/ipc/server_linux.go` | 283 |
| `internal/lease/lease.go` | 173 |
| `internal/lease/totp.go` | 60 |
| **Total** | **1407** |

The identical counter gives TASK-0017 parent `1d903d7` (tree
`8b1339bd7213f7a984c89d142ff1e608b5378a97`) **1262** and merge
`212fd03` **1407**, so the TASK-0017 Git delta is **+145**. Test LOC across
all `*_test.go` files under `cmd`, `internal`, and `deploy` is **3399**
(parent **3090**, TASK-0017 delta **+309**).

The prior ledger's `1253 + 145 = 1398` forecast is nine lines low. TASK-0008
recorded its new sudo helper as 38 canonical lines, while its merged pre-
TASK-0017 file contains 47. The discrepancy is exactly reconciled:

```text
1253 + (47 - 38 historical undercount) + 145 TASK-0017 = 1407
1407 + 0 TASK-0012 = 1407
target reserve: 1500 - 1407 = 93
hard reserve:   1800 - 1407 = 393
```

This is correction of an earlier ledger, not TASK-0012 product SLOC, reserve
borrowing, compression, throughput inference, or a new feature decision.

## Required measurement row

| Field | Value and provenance |
| --- | --- |
| planned / actual production SLOC | `0 / 0` for TASK-0012 |
| cumulative production SLOC | `1407` from merged full-tree count |
| test LOC | `3399` from merged full-tree count |
| plan / dev / review / QA / CI-push-merge ms and minutes | `null`; no canonical TASK-0017 terminal or valid same-task/lap/stage/attempt pairs |
| task cycle ms and minutes | `null`; no canonical TASK-0017 terminal and no synthetic Lap permitted |
| active_ms / wait_ms | `null / null`; no authoritative nonoverlapping observations |
| retries | `null`; no canonical TASK-0017 event counter; not represented as zero |
| raw / effective failure classifications | `null / null`; fixture classifications exist only in bounded REVIEW/QA prose, not canonical TASK-0017 events |
| source event IDs | none for TASK-0017; status source is PR #18, merge/tree, and hashed REVIEW/QA files |
| notes | TASK-0017 status proven externally; timing remains null; historical cohort remains the TASK-0009 report |

No timing contingency is calculable for these nulls. The 20% rule applies
only to observed non-preflight time and never to SLOC, test LOC, retry,
active/wait, paths, or scope.

## Cohort, status, and reserve guards

The TASK-0009 historical cohort (TASK-0001/0003/0004/0005), product-terminal
selection, TASK-0013 null timing, correction edges, stage pairs, retry maxima,
and raw/effective classifications are unchanged and incorporated by the
hashed prior report. The new frozen stream adds no valid TASK-0017 terminal;
no partial, correction, governance, terminated, unfinished, or `pass` pseudo-
terminal is promoted.

TASK-0017 is completed; TASK-0010 and TASK-0011 remain `deferred-v2`,
`executable:false`, and zero v1 SLOC. Both mandatory later milestones remain
reserved and non-executable until this candidate independently passes REVIEW
and QA and Main merges it. Actual `1407 < 1500 < 1800`; no mandatory control
is removed or compressed.

## Affected checks

The frozen hash/count, JSON parse, unique IDs, correction predicate, full-tree
production/test counters, parent/merge delta, `git diff --check`, backlog JSON,
and exact changed-path scope are the complete affected check set. Unrelated
product tests and formatting cannot prove this evidence-only result and are
not represented as product PASS.
