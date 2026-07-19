# TASK-0006 completed-cycle measurement

## Canonical snapshot and method

The source is the append-only, read-only file
`/home/ubuntu/git/agent-harness-work/lap30/events.jsonl`.  At measurement time
it contained 83 newline-delimited JSON objects.  Eleven TASK-0006 events are
outside the selection, nine of them appended after the PLAN's earlier 74-line
observation.  The input
is the 72 file-order-preserving events whose `task_id` is TASK-0001,
TASK-0003, TASK-0004, or TASK-0005.

- Full-file SHA-256 at selection: `bb0689371c15bc6c4ad59f221dba4b8b6eea6795bda86237d03d47ea759a4a20`
- Filtered compact-JSON SHA-256: `a77eb360a022eb6962fb1f279e1b9e9ec6294552c76f0a005725966165cc5f3a`
- Review-remediation reread: 91/91 full-file schema-v1 events with unique IDs,
  full-file SHA-256 `2642e270b947573c877c0831936f15590bbaf29f488e7efb8423796d3645d8fc`;
  the selected count remains 72 and its filtered SHA-256 is unchanged.
- Selected counts: TASK-0001 3, TASK-0003 7, TASK-0004 38, TASK-0005 24
- Integrity: 83/83 full-file and 72/72 selected event IDs are unique; every
  object has schema version 1; `(lap_id, sequence)` is unique within each lap.
- Completion set: exactly one `lap_completed/completed` event for each of the
  four selected Tasks.  Sequence resets between laps and is not a global time.
- Every correction target exists earlier in file order in the same Task/lap.
  Source nulls remain null.  Recording timestamps and late corrections are not
  used to invent occurrence times or durations.

Milliseconds are preserved below.  Display minutes are `ms / 60000`, rounded
to two decimals.  A stage duration is the terminal `elapsed_ms` minus the
matched same-Task/lap/stage/attempt start.  Attempts are summed; checkpoint
snapshots are not durations.  Terminal lap elapsed values become cycle time
only when selected evidence establishes preflight exclusion.  TASK-0001 and
TASK-0003 have observable lap elapsed values but no selected event proving
that exclusion, so their cycle fields remain null.  TASK-0005 explicitly
records preflight as zero; TASK-0004/TASK-0005 retain their established cycle
evidence.  Preflight is assigned to no stage or contingency.

## Completed-Task evidence

`null` means not derivable under the rules, never zero.  `source_event_ids`
means every selected event for that Task (counts above), with the exact
metric-bearing IDs in the provenance sections.

| Field | TASK-0001 | TASK-0003 | TASK-0004 | TASK-0005 |
| --- | ---: | ---: | ---: | ---: |
| planned_production_sloc | 130 | 220 | 392 | 175 |
| actual_production_sloc | 98 | 135 | 367 | 151 |
| cumulative_production_sloc | 98 | 233 | 600 | 751 |
| test_loc | 116 | 216 | 496 | 1257 |
| plan_ms / plan_minutes | null / null | null / null | null / null | null / null |
| dev_ms / dev_minutes | null / null | null / null | 878000 / 14.63 | 625000 / 10.42 |
| review_ms / review_minutes | null / null | null / null | 615000 / 10.25 | 392000 / 6.53 |
| qa_ms / qa_minutes | null / null | null / null | 244000 / 4.07 | 347000 / 5.78 |
| ci_push_merge_ms / ci_push_merge_minutes | null / null | null / null | 177000 / 2.95 | 178000 / 2.97 |
| task_cycle_ms / task_cycle_minutes | null / null | null / null | 3779000 / 62.98 | 2575000 / 42.92 |
| active_ms | null | null | null | null |
| wait_ms | null | null | null | null |
| retries | 0 | 1 | 4 | 1 |
| raw_failure_classifications | `[]` | `implementation_defect` records | `environment_issue`, `planning_defect`, `implementation_defect` records | superseded `requirement_gap`, plus `planning_defect`, `environment_issue` records |
| effective_failure_classifications | `[]` | `[implementation_defect]` | `[environment_issue, planning_defect, implementation_defect]` | `[planning_defect, environment_issue]` |
| source_event_ids | all 3 selected Task events | all 7 selected Task events | all 38 selected Task events | all 24 selected Task events |
| notes | named-stage timing absent; lap elapsed observed but preflight exclusion unproved, so cycle null | named-stage timing absent; lap elapsed observed but preflight exclusion unproved, so cycle null; exact failure occurrence remains unknown | PLAN timing null; initial 417 estimates corrected to 320 before final approved 392 | PLAN timing null; initial SLOC stop reclassified by validated correction |

Production increments are from the final completion records and reconcile in
dependency order: `98 + 135 + 367 + 151 = 751`.  The completion records are
`task0001-lap01-complete`, `task0003-lap02-complete`,
`task0004-lap03-complete`, and `task0005-lap02-complete`.

TASK-0004's earlier `planned_production_sloc=417` entries at
`task0004-lap01-plan-finished` and `task0004-lap01-qa-plan-finished` are
corrected to 320 by `task0004-lap01-plan-sloc-correction` and
`task0004-lap01-qa-plan-sloc-correction`; the later final approved completion
value is 392.  Corrections remain provenance and do not overwrite the final
completion field.

### Timing provenance and arithmetic

- TASK-0001 lap elapsed is observed at
  `task0001-lap01-complete=925000`, but no selected event proves preflight
  exclusion.  It is retained as provenance and not adopted as cycle time.
- TASK-0003 lap elapsed is observed at `task0003-lap01-stop=884000` plus
  `task0003-lap02-complete=858000` (1742000 total), but no selected event
  proves preflight exclusion.  It is retained as provenance and not adopted
  as cycle time.
- TASK-0004 DEV: (`task0004-lap01-dev-start=1242000` to
  `task0004-lap01-dev-stopped=1668000`) +
  (`task0004-lap02-dev-start=490000` to
  `task0004-lap02-dev-finish=757000`) +
  (`task0004-lap02-dev-retry-start=1128000` to
  `task0004-lap02-dev-retry-finish=1313000`) =
  `426000 + 267000 + 185000 = 878000`.
- TASK-0004 REVIEW: (`task0004-lap02-review-start=757000` to
  `task0004-lap02-review-finish-fail=1128000`) +
  (`task0004-lap02-review-retry-start=1313000` to
  `task0004-lap02-review-retry-finish=1557000`) =
  `371000 + 244000 = 615000`.
- TASK-0004 QA: `task0004-lap02-qa-start=1557000` to
  `task0004-lap02-qa-finish=1801000`, or 244000.  Git:
  `task0004-lap03-git-start=0` to
  `task0004-lap03-git-finish=177000`, or 177000.  The cycle is
  `1801000 + 1801000 + 177000 = 3779000`.
- TASK-0005 DEV: (`task0005-lap01-dev-start=557000` to
  `task0005-lap01-dev-stop-sloc=748000`) +
  (`task0005-lap01-dev2-start=1224000` to
  `task0005-lap01-dev2-finish=1658000`) =
  `191000 + 434000 = 625000`.
- TASK-0005 REVIEW: (`task0005-lap01-review-start=1658000` to matching
  `task0005-lap01-stop=1830000`) +
  (`task0005-lap02-review-start=0` to
  `task0005-lap02-review-finish=220000`) =
  `172000 + 220000 = 392000`.
- TASK-0005 QA: `task0005-lap02-qa-start=220000` to
  `task0005-lap02-qa-finish=567000`, or 347000.  Git:
  `task0005-lap02-git-start=567000` to
  `task0005-lap02-git-finish=745000`, or 178000.  The cycle is
  `1830000 + 745000 = 2575000`.

The two Git-stage terminals explicitly report that no remote CI checks were
configured.  Their 177000/178000 ms values still record main-owned final
checks, commit/push/PR/merge work; absence of configured checks is not treated
as zero CI/push/merge effort.

TASK-0001/TASK-0003 have no matched numeric named-stage pairs.  TASK-0004 and
TASK-0005 PLAN finishes lack numeric matched starts, so those fields remain
null.  The active/wait observations are also null at Task level because some
stages lack them and the values are independently estimated and may overlap.
Available terminal event pairs are retained rather than summed: TASK-0004
DEV `(430000,27000)`, `(410000,18000)`, `(190000,11000)`, REVIEW
`(19543,10777)`, `(18779,5147)`, QA `(18606,0)`; TASK-0005 DEV
`(190000,2000)`, `(620000,28000)`, REVIEW `(19983,0)`, QA `(20168,0)`.

### Retry provenance

Retries are the maximum propagated Task counter, not a sum.  TASK-0001 has
no positive retry event.  TASK-0003's positive IDs are
`task0003-lap02-start`, `task0003-lap02-review-fail`,
`task0003-lap02-complete`, and
`task0003-lap02-failure-time-correction` (maximum 1).

TASK-0004's positive IDs are `task0004-lap01-checkpoint10`,
`task0004-lap01-plan-finished`, `task0004-lap01-plan-sloc-correction`,
`task0004-lap01-checkpoint20`, `task0004-lap01-dev-environment-failure`,
`task0004-lap01-reestimate-trigger`, `task0004-lap01-dev-stopped`,
`task0004-lap01-checkpoint30`, `task0004-lap01-stop`,
`task0004-lap02-start`, `task0004-lap02-plan-start`,
`task0004-lap02-plan-finish`, `task0004-lap02-qa-plan-cwd-failure`,
`task0004-lap02-qa-plan-finish`, `task0004-lap02-review-failure`,
`task0004-lap02-review-finish-fail`, `task0004-lap02-dev-retry-start`,
`task0004-lap02-review-retry-start`, `task0004-lap02-checkpoint10`,
`task0004-lap02-checkpoint20`, `task0004-lap02-checkpoint30`,
`task0004-lap02-stop`, and `task0004-lap03-start` (maximum 4).

TASK-0005's positive IDs are `task0005-lap01-checkpoint20`,
`task0005-lap01-dev2-start`, `task0005-lap01-review-start`,
`task0005-lap01-checkpoint30`, `task0005-lap01-stop`,
`task0005-lap02-start`, `task0005-lap02-review-start`,
`task0005-lap02-qa-finish`, and `task0005-lap02-complete` (maximum 1).

### Raw and effective classifications

Each raw record is immutable.  Only a validated classification correction
sets `superseded_by`; effective values are ordered unique classifications from
unsuperseded records.

| Task | classification | source_event_id | superseded_by |
| --- | --- | --- | --- |
| TASK-0003 | implementation_defect | task0003-lap02-review-fail | null |
| TASK-0003 | implementation_defect | task0003-lap02-failure-time-correction | null |
| TASK-0004 | environment_issue | task0004-lap01-dev-environment-failure | null |
| TASK-0004 | planning_defect | task0004-lap01-reestimate-trigger | null |
| TASK-0004 | planning_defect | task0004-lap01-dev-stopped | null |
| TASK-0004 | planning_defect | task0004-lap01-stop | null |
| TASK-0004 | environment_issue | task0004-lap02-qa-plan-cwd-failure | null |
| TASK-0004 | implementation_defect | task0004-lap02-review-failure | null |
| TASK-0004 | implementation_defect | task0004-lap02-review-finish-fail | null |
| TASK-0005 | requirement_gap | task0005-lap01-dev-stop-sloc | task0005-lap01-correct-dev-class |
| TASK-0005 | planning_defect | task0005-lap01-correct-dev-class | null |
| TASK-0005 | planning_defect | task0005-lap01-replan-finish | null |
| TASK-0005 | planning_defect | task0005-lap01-reqaplan-finish | null |
| TASK-0005 | environment_issue | task0005-lap01-dev2-finish | null |
| TASK-0005 | planning_defect | task0005-lap01-stop | null |
| TASK-0005 | environment_issue | task0005-lap02-review-finish | null |
| TASK-0005 | environment_issue | task0005-lap02-qa-finish | null |
| TASK-0005 | planning_defect | task0005-lap02-complete | null |

The validated edge is
`task0005-lap01-dev-stop-sloc -> task0005-lap01-correct-dev-class`:
old `requirement_gap`, new `planning_defect`, same Task/lap, earlier target.
Consequently `requirement_gap` remains raw history but is absent from the
effective TASK-0005 set.  TASK-0003's correction changes only the occurrence
description to unknown; it creates no timestamp and does not supersede the
classification record.

## Replanning decision

Incremental production varied 98, 135, 367, and 151 SLOC, while two Tasks lack
stage splits and the security boundaries differ.  This evidence is too sparse
for a fixed SLOC/minute, LOC velocity, average throughput, or regression.  SLOC
values below are ceilings/reserves, not delivery forecasts.

For a comparable completed boundary, time contingency is
`ceil(observed_non_preflight_minutes * 1.20)`.  TASK-0004 supplies the sudo
analogue: `3779000 ms / 60000 = 62.9833... minutes`; conservatively displayed
as 62.98 and calculated from raw time,
`ceil(62.9833... * 1.20) = ceil(75.58) = 76 minutes`.  This is a post-preflight
minimum; its upper bound is unknown.  Restricted push/token custody has no
completed transport analogue, so both delivery bounds remain unknown.  A 20%
factor is not applied to SLOC, and null does not become a number.

Only TASK-0007 (sudo live check) and TASK-0008 (restricted push/token custody)
are converted, followed immediately by zero-SLOC TASK-0009 measurement.
Audit/release (`<=1500`) and clean-canary (`+0`) remain non-executable reserves
until TASK-0009 PASS and merge.  Next-wave cumulative ceilings are 950 and
1400; global mandatory-v1 remains <=1500 with the exact backlog shedding order
and mandatory-control exclusions preserved.
