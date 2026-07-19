# TASK-0009 measurement — frozen 389-event gate

## Source, integrity, and method

This report is derived only from the immutable frozen snapshot
`/tmp/task0009-events-frozen.jsonl`; no mutable operational log, product
fixture, network, elevation, or canonical-log write was used.

| item | evidence |
| --- | --- |
| source path | `/tmp/task0009-events-frozen.jsonl` |
| SHA-256 | `abe31c9da1fbcfd32daef6013a9ce58063ccbb086265b4760ca20a22962af09b` |
| records/lines | 389 JSONL records |
| derivation time | 2026-07-19T13:04:31Z |
| commands | `sha256sum "$EVENTS"`; `test "$(wc -l < "$EVENTS")" -eq 389`; `jq -e . "$EVENTS"`; unique `event_id` check; correction-edge check; historical-cohort check |
| integrity result | PASS: every line parses, every `event_id` is unique, and every correction target is nonempty, earlier in file order, same task/lap, and smaller sequence |
| units/rounding | durations are integer milliseconds; displayed minutes are `ms / 60000` to two decimals; time contingency is `ceil(observed_ms * 1.20)`; SLOC/test LOC are not multiplied |

Stage timing uses only same-task, same-lap, same-stage, same-attempt
`stage_started` to terminal pairs. A `stage_finished` or a matching
`lap_stopped` is a terminal; checkpoints and snapshots are never durations.
`preflight` is excluded. Unmatched or non-authoritative values remain `null`
with a reason. Active and wait observations remain separate and are never
added to make elapsed time. Retries are the maximum propagated task-level
counter, not a sum.

## Cohorts and terminal selection

The historical performance cohort is exactly the four completed tasks below;
it is kept separate from product terminals and cannot be used for a fixed
SLOC/minute or average forecast. Product-terminal selection excludes
`annotations.counts_as_product_lap == "false"`, governance-only cycles,
corrections, the terminated TASK-0007, and unfinished TASK-0014.

| population | task | terminal/source event | planned SLOC | actual SLOC | test LOC | cumulative SLOC | result |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| historical performance | TASK-0001 | `task0001-lap01-complete` | 130 | 98 | 116 | 98 | completed |
| historical performance | TASK-0003 | `task0003-lap02-complete` | 220 | 135 | 216 | 233 | completed; lap01 stop is not a terminal |
| historical performance | TASK-0004 | `task0004-lap03-complete` | 392 | 367 | 496 | 600 | completed |
| historical performance | TASK-0005 | `task0005-lap02-complete` | 175 | 151 | 1257 | 751 | completed; raw requirement gap is corrected below |
| product terminal | TASK-0015 | `task0015-counted-lap02-complete` | 280 | 278 | 1057 | 1200 | counted terminal; source snapshot 1200 minus prior 922 = +278; partial interval excluded |
| product terminal | TASK-0016 | `task0016-counted-lap02-complete` | 20 | 15 | 698 | 1215 | counted terminal; source snapshot 1215 minus prior 1200 = +15; planning partial excluded |
| product terminal | TASK-0008 | `task0008-counted-lap01-complete` | 55 | 38 | 283 | 1253 | counted terminal; source snapshot 1253 minus prior 1215 = +38; partial intervals excluded |

Product terminal `actual_production_sloc` values are normalized to the
increment for this report: the canonical terminal annotations are cumulative
snapshots (1200, 1215, 1253), so increments are `1200 - 922 = 278`,
`1215 - 1200 = 15`, and `1253 - 1215 = 38`. The cumulative chain is a
snapshot chain, not an additive sum: `1200 -> 1215 -> 1253`; therefore
TASK-0009 contributes exactly zero and
`1253 + 0 = 1253`. TASK-0013 has merged SLOC lineage of 922 in
`task0013-counted-lap02-complete`, but both of its lap terminals have
`status: pass`, not `status: completed`; its completion/cycle timing is
`null`, reason `no canonical TASK-0013 terminal`. TASK-0007 remains
terminated with zero merged SLOC despite an unmerged 1087-SLOC candidate, and
TASK-0014 remains unfinished with no product terminal. TASK-0006 is a separate
governance cycle and contributes zero product SLOC.

## Measurement rows

`null` is intentional and has a reason; it is never a synthetic zero. The
cycle value is the sum of authoritative non-preflight lap intervals where the
source establishes the interval. For the completed historical rows this is
`TASK-0004: 1,801,000 + 1,801,000 + 177,000 = 3,779,000 ms` and
`TASK-0005: 1,830,000 + 745,000 = 2,575,000 ms`; TASK-0001 and TASK-0003 do
not prove preflight exclusion and therefore retain a null cycle.

| row | planned / actual production SLOC | cumulative SLOC | test LOC | plan ms / min | dev ms / min | review ms / min | QA ms / min | CI/push/merge ms / min | task cycle ms / min | active_ms | wait_ms | retries |
| --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- | --- | ---: |
| TASK-0001 historical | 130 / 98 | 98 | 116 | null / null | null / null | null / null | null / null | null / null | null / null (no authoritative preflight exclusion) | null | null | 0 |
| TASK-0003 historical | 220 / 135 | 233 | 216 | null / null | null / null | null / null | null / null | null / null | null / null (no authoritative preflight exclusion) | null | null | 1 |
| TASK-0004 historical | 392 / 367 | 600 | 496 | null / null | 878000 / 14.63 | 615000 / 10.25 | 244000 / 4.07 | 177000 / 2.95 | 3779000 / 62.98 | null | null | 4 |
| TASK-0005 historical | 175 / 151 | 751 | 1257 | null / null | 625000 / 10.42 | 392000 / 6.53 | 347000 / 5.78 | 178000 / 2.97 | 2575000 / 42.92 | null | null | 1 |
| TASK-0013 lineage-only | 165 / 171 | 922 | 457 | null (no canonical terminal) | null (no canonical terminal) | null (no canonical terminal) | null (no canonical terminal) | null (no canonical terminal) | null (no canonical TASK-0013 terminal) | null | null | 2 |
| TASK-0015 product terminal | 280 / 278 | 1200 | 1057 | 682000 / 11.37 | 1211000 / 20.18 | 218000 / 3.63 | 161000 / 2.68 | 306000 / 5.10 | 2788000 / 46.47 | null | null | 1 |
| TASK-0016 product terminal | 20 / 15 | 1215 | 698 | null (attempt mismatch) | 852000 / 14.20 | 404000 / 6.73 | 109000 / 1.82 | 205000 / 3.42 | 2424000 / 40.40 | null | null | 2 |
| TASK-0008 product terminal | 55 / 38 | 1253 | 283 | null (partials have no pair) | 626000 / 10.43 | 130000 / 2.17 | 218000 / 3.63 | 300000 / 5.00 | 1568000 / 26.13 | null | null | 2 |
| TASK-0006 governance-only | 0 / 0 | 751 (governance baseline) | null | null | null | null | null | null | null (governance laps kept separate) | null | null | 4 |
| TASK-0007 terminated | 170 / 0 merged | 751 | null | null | null | null | null | null | null (terminated; no completed terminal) | null | null | 3 |
| TASK-0014 unfinished | 280 / 0 merged | 922 lineage | 535 (unfinished evidence) | null | null | null | null | null | null (unfinished; no completed terminal) | null | null | 8 |

TASK-0015's 682000 ms plan is the two paired partial planning intervals
`302000->760000` and `760000->984000`; its product terminal still excludes
the `counts_as_product_lap=false` partial from SLOC selection. TASK-0016's
review attempt 3 and its plan/QA partial have no same-attempt numeric start,
so those portions are not imputed. TASK-0008 timing in the row is only its
counted product lap; its partial preparation records remain provenance and
are not product terminals.

TASK-0013 has raw stage pairs in the frozen stream (DEV 269000 + 488000 +
497000, REVIEW 281000 + 44000 + 324000 + 168000, QA 292000, and Git
119000 + 235000), but its row deliberately keeps every completion-stage and
cycle field null with reason `no canonical TASK-0013 terminal`. These pairs
are source provenance, not an invented completed cycle.

## Stage-pair provenance and time-only contingency

The following are the reproducible paired intervals used above. Each
contingency is independent and applies only to that observable non-preflight
time; null pairs receive no multiplier.

| task/stage | paired source IDs (start -> terminal) | observed ms | contingency `ceil(ms*1.20)` ms |
| --- | --- | ---: | ---: |
| TASK-0004 DEV | `task0004-lap01-dev-start -> task0004-lap01-dev-stopped`; `task0004-lap02-dev-start -> task0004-lap02-dev-finish`; `task0004-lap02-dev-retry-start -> task0004-lap02-dev-retry-finish` | 426000 + 267000 + 185000 = 878000 | 1053600 |
| TASK-0004 REVIEW | `task0004-lap02-review-start -> task0004-lap02-review-finish-fail`; `task0004-lap02-review-retry-start -> task0004-lap02-review-retry-finish` | 371000 + 244000 = 615000 | 738000 |
| TASK-0004 QA | `task0004-lap02-qa-start -> task0004-lap02-qa-finish` | 244000 | 292800 |
| TASK-0004 Git | `task0004-lap03-git-start -> task0004-lap03-git-finish` | 177000 | 212400 |
| TASK-0005 DEV | `task0005-lap01-dev-start -> task0005-lap01-dev-stop-sloc`; `task0005-lap01-dev2-start -> task0005-lap01-dev2-finish` | 191000 + 434000 = 625000 | 750000 |
| TASK-0005 REVIEW | `task0005-lap01-review-start -> task0005-lap01-stop`; `task0005-lap02-review-start -> task0005-lap02-review-finish` | 172000 + 220000 = 392000 | 470400 |
| TASK-0005 QA | `task0005-lap02-qa-start -> task0005-lap02-qa-finish` | 347000 | 416400 |
| TASK-0005 Git | `task0005-lap02-git-start -> task0005-lap02-git-finish` | 178000 | 213600 |
| TASK-0015 PLAN | `task0015-partial01-contract-start -> task0015-partial01-contract-finish`; `task0015-partial01-plan-start -> task0015-partial01-plan-finish` | 458000 + 224000 = 682000 | 818400 |
| TASK-0015 DEV | `task0015-counted-lap01-dev-start -> task0015-counted-lap01-dev-finish`; `task0015-counted-lap01-dev-correction-start -> task0015-counted-lap01-dev-correction-finish` | 937000 + 274000 = 1211000 | 1453200 |
| TASK-0015 REVIEW | `task0015-counted-lap01-review-start -> task0015-counted-lap01-review-finish`; `task0015-counted-lap01-review-retry-start -> task0015-counted-lap01-review-retry-finish` | 163000 + 55000 = 218000 | 261600 |
| TASK-0015 QA | `task0015-counted-lap01-qa-start -> task0015-counted-lap01-qa-finish` | 161000 | 193200 |
| TASK-0015 Git | `task0015-counted-lap02-git-start -> task0015-counted-lap02-git-finish` | 306000 | 367200 |
| TASK-0016 DEV | `task0016-counted-lap01-dev-start -> task0016-counted-lap01-dev-finish`; `task0016-counted-lap01-dev-correction-start -> task0016-counted-lap01-dev-correction-finish` | 611000 + 241000 = 852000 | 1022400 |
| TASK-0016 REVIEW | `task0016-counted-lap01-review-start -> task0016-counted-lap01-review-finish` (185000); `task0016-counted-lap01-review-retry-start -> task0016-counted-lap01-review-retry-finish` (219000); attempt 3 has no paired start | 185000 + 219000 = 404000 | 484800 |
| TASK-0016 QA | `task0016-counted-lap01-qa-start -> task0016-counted-lap01-qa-finish` | 109000 | 130800 |
| TASK-0016 Git | `task0016-counted-lap02-git-start -> task0016-counted-lap02-git-finish` | 205000 | 246000 |
| TASK-0008 DEV | `task0008-counted-lap01-dev-start -> task0008-counted-lap01-dev-finish` | 626000 | 751200 |
| TASK-0008 REVIEW | `task0008-counted-lap01-review-start -> task0008-counted-lap01-review-finish` | 130000 | 156000 |
| TASK-0008 QA | `task0008-counted-lap01-qa-start -> task0008-counted-lap01-qa-finish` | 218000 | 261600 |
| TASK-0008 Git | `task0008-counted-lap01-git-start -> task0008-counted-lap01-git-finish` | 300000 | 360000 |

The historical lap cycles and product-terminal cycle values above use their
explicit terminal intervals; the contingency is shown for each observable
stage and is not added to SLOC, test LOC, active/wait, retries, paths, or
scope. `plan_ms` and unmatched attempts remain null even when a checkpoint
contains a timestamp.

## Active, wait, and retry provenance

Task-level `active_ms` and `wait_ms` remain null because stage observations are
partial/overlapping and cannot be safely summed. The raw observations are
retained separately; examples covering every non-null completed/product row
are:

| task | source observations `(event_id: active_ms, wait_ms)` |
| --- | --- |
| TASK-0004 | `task0004-lap01-reestimate-trigger: 430000,27000`; `task0004-lap02-dev-finish: 410000,18000`; `task0004-lap02-dev-retry-finish: 190000,11000`; `task0004-lap02-review-failure: 19543,10777`; `task0004-lap02-review-retry-finish: 18779,5147`; `task0004-lap02-qa-finish: 18606,0` |
| TASK-0005 | `task0005-lap01-dev-stop-sloc: 190000,2000`; `task0005-lap01-dev2-finish: 620000,28000`; `task0005-lap02-review-finish: 19983,0`; `task0005-lap02-qa-finish: 20168,0` |
| TASK-0015 | `task0015-counted-lap02-git-finish: 306000,0` (the counted terminal repeats this observation) |
| TASK-0016 | `task0016-partial01-complete: 419000,0`; `task0016-counted-lap01-dev-finish: 30000,0`; `task0016-counted-lap01-review-finish: 420000,5000`; `task0016-counted-lap01-dev-correction-finish: 12000,0`; `task0016-counted-lap01-review-attempt3-finish: 240000,20000`; `task0016-counted-lap02-git-finish: 205000,0` |
| TASK-0008 | `task0008-counted-lap01-qa-finish: null,25000`; `task0008-counted-lap01-git-finish: 300000,0` |

Retries are maximum propagated counters with source IDs, not sums:

| task/population | max retries | source evidence |
| --- | ---: | --- |
| TASK-0001 | 0 | `task0001-lap01-complete` |
| TASK-0003 | 1 | `task0003-lap02-start`, `task0003-lap02-review-fail`, `task0003-lap02-complete` |
| TASK-0004 | 4 | `task0004-lap02-checkpoint20`, `task0004-lap02-stop`, `task0004-lap03-start` |
| TASK-0005 | 1 | `task0005-lap01-checkpoint20`, `task0005-lap01-dev2-start`, `task0005-lap02-complete` |
| TASK-0013 | 2 | `task0013-counted-lap02-devfix-start`, `task0013-counted-lap02-complete` |
| TASK-0015 | 1 | `task0015-partial01-complete`, `task0015-counted-lap01-dev-finish` |
| TASK-0016 | 2 | `task0016-counted-lap01-checkpoint20`, `task0016-counted-lap01-stop` |
| TASK-0008 | 2 | `task0008-partial02-preflight-finish`, `task0008-partial02-complete` |
| TASK-0006 / TASK-0007 / TASK-0014 | 4 / 3 / 8 | respective latest governance, terminated, and unfinished records |

## Correction edges and raw/effective classification

The validated correction predicate passed before any effective derivation.
Raw records remain immutable; `superseded_by` is only attached to a target
after this check. Corrections are never additional tasks or terminals.

| correction event | raw target | same task/lap/sequence proof | corrected field/class | effective consequence |
| --- | --- | --- | --- | --- |
| `task0003-lap02-failure-time-correction` | `task0003-lap02-review-fail` | TASK-0003/lap02; target earlier; 2 < 4 | occurrence timestamp unknown | classification remains `implementation_defect`; no time is invented |
| `task0004-lap01-plan-sloc-correction` | `task0004-lap01-plan-finished` | TASK-0004/lap01; 3 < 5 | planned SLOC 417 -> 320 | effective planned value 320 |
| `task0004-lap01-qa-plan-sloc-correction` | `task0004-lap01-qa-plan-finished` | TASK-0004/lap01; 4 < 6 | planned SLOC 417 -> 320 | effective planned value 320 |
| `task0005-lap01-correct-dev-class` | `task0005-lap01-dev-stop-sloc` | TASK-0005/lap01; 7 < 8 | classification `requirement_gap` -> `planning_defect` | raw requirement gap is superseded; effective set excludes it |
| `task0006-lap01-correct-dev-class-schema` | `task0006-lap01-dev-finish` | TASK-0006/lap01; 12 < 16 | classification schema correction | effective schema classification `implementation_defect` |
| `task0006-lap02-correct-review-class-schema` | `task0006-lap02-review-fail` | TASK-0006/lap02; 3 < 13 | classification schema correction | effective schema classification `implementation_defect` |
| `task0006-lap02-correct-devstart-class-schema` | `task0006-lap02-dev-retry-start` | TASK-0006/lap02; 4 < 14 | classification schema correction | effective schema classification `implementation_defect` |
| `task0006-lap02-correct-devfinish-class-schema` | `task0006-lap02-dev-retry-finish` | TASK-0006/lap02; 5 < 15 | classification schema correction | effective schema classification `implementation_defect` |
| `task0006-lap02-correct-complete-class-schema` | `task0006-lap02-complete` | TASK-0006/lap02; 12 < 16 | classification schema correction | effective completion classification `implementation_defect` |
| `task0006-r3-lap01-correct-stop-event` | `task0006-r3-lap01-stop` | TASK-0006/task0006-r3-lap01; 23 < 24 | terminal record correction | raw stop retained; effective stop is corrected record |
| `task0007-lap01-correct-qaplan-start` | `task0007-lap01-qaplan-start` | TASK-0007/lap01; 4 < 5 | not-started/environment record | no product terminal created |
| `task0013-lap01-correct-start-preflight` | `task0013-lap01-start` | TASK-0013/lap01; 1 < 18 | preflight status | no synthetic TASK-0013 terminal |
| `task0013-lap01-correct-preflight-pass` | `task0013-lap01-preflight-finish` | TASK-0013/lap01; 2 < 19 | preflight status | no synthetic TASK-0013 terminal |
| `task0013-counted-lap02-correct-git-count` | `task0013-counted-lap02-git-start` | TASK-0013/count lap02; 12 < 13 | Git count correction | timing remains null at task completion level |

Effective classification sets, with raw source IDs retained, are:

| task | raw classification records (source IDs) | effective unsuperseded classifications |
| --- | --- | --- |
| TASK-0001 | none | `[]` |
| TASK-0003 | `implementation_defect`: `task0003-lap02-review-fail`, `task0003-lap02-failure-time-correction` | `[implementation_defect]` |
| TASK-0004 | `environment_issue`: `task0004-lap01-checkpoint10`, `task0004-lap01-dev-environment-failure`, `task0004-lap02-qa-plan-cwd-failure`; `planning_defect`: `task0004-lap01-reestimate-trigger`, `task0004-lap01-dev-stopped`, `task0004-lap01-checkpoint30`, `task0004-lap01-stop`, `task0004-lap02-checkpoint20`; `implementation_defect`: `task0004-lap02-review-failure`, `task0004-lap02-review-finish-fail` | `[environment_issue, planning_defect, implementation_defect]` |
| TASK-0005 | raw `requirement_gap`: `task0005-lap01-dev-stop-sloc` (superseded by `task0005-lap01-correct-dev-class`); raw `planning_defect`: `task0005-lap01-correct-dev-class`, `task0005-lap01-replan-finish`, `task0005-lap01-reqaplan-finish`, `task0005-lap01-checkpoint20`, `task0005-lap01-checkpoint30`, `task0005-lap01-stop`, `task0005-lap02-complete`; raw `environment_issue`: `task0005-lap01-dev2-finish`, `task0005-lap02-review-finish`, `task0005-lap02-qa-finish` | `[planning_defect, environment_issue]` |
| TASK-0013 | raw planning/environment/implementation IDs retained | timing/effective product terminal `null` (no canonical completed terminal) |
| TASK-0015 | raw `environment_issue`: partial preflight and lap01 checkpoint; raw `implementation_defect`: lap01 review failure/correction | `[environment_issue, implementation_defect]` |
| TASK-0016 | raw `planning_defect`: partial completion/planning failure/checkpoint; raw `environment_issue`: review failure; raw `implementation_defect`: review retry failure | `[planning_defect, environment_issue, implementation_defect]` |
| TASK-0008 | raw `planning_defect`: partial QA failure/partial replan; raw `requirement_gap`: failed preflight/stop; raw `environment_issue`: counted QA start | `[planning_defect, requirement_gap, environment_issue]` |

In particular, TASK-0005's raw `requirement_gap` remains visible and points
to `superseded_by: task0005-lap01-correct-dev-class`; only the effective
`planning_defect` value is used for classification arithmetic. No raw/effective
unlabeled mixture is used.

## Cap reconciliation and explicit downstream replan

The measured baseline is **1253 actual merged production SLOC** and this gate
adds **0**, so:

```
1253 + 0 TASK-0009 = 1253
target capacity: 1500 - 1253 = 247
hard capacity:   1800 - 1253 = 547
```

The prior downstream arithmetic was `130 + 153 = 283`, which would produce
`1253 + 283 = 1536`: 36 over the mandatory-v1 target, even though below the
unconditional 1800 hard limit. The hard limit is not permission to exceed the
target. The canonical ordered shedding list was reviewed without compression
or weakening safety controls. Items 1–6 have no remaining applicable scope in
these two coupled Tasks; therefore item 7 is explicitly selected:

**push-to-v2: selected.** Defer the coupled TASK-0010 local-push policy and
TASK-0011 token/system-Git push implementations as a non-executable v2
reserve. Preserve all mandatory TOTP/full-sudo, readiness, fail-closed IPC,
live/no-cache sudo, secret non-disclosure, readability, and independent-gate
requirements. No SLOC is silently shrunk, packed, or borrowed from the hard
reserve. TASK-0012 is replanned as the direct zero-SLOC final v1 measurement
and later-reserve gate after TASK-0009; its actual cumulative baseline is
1253, target 1500, hard 1800, and later work remains blocked until its own
independent REVIEW/QA and main merge.

This is an explicit replan, not a silent v2 selection. TASK-0010 and TASK-0011
must not receive PLAN/branch/DEV/PR-ready detail while deferred. If a future
v2 replan restores push, it must provide fresh evidence, an approved PLAN and
QA_PLAN, and re-run the ordered shedding and cap checks. There is no
throughput, SLOC velocity, average, or contingency-based capacity assumption.

## Scope and gate result

The candidate changes only the authorized process-evidence paths and adds
zero production SLOC. Canonical JSONL and product/test code are untouched.
Independent REVIEW and QA must independently regenerate the parse, unique-ID,
correction, cohort/terminal, stage-pair, null, active/wait, retry,
classification, and cap arithmetic and run the repository-native checks. The
measurement is ready for REVIEW; it does not itself grant merge or later
reserve execution.
