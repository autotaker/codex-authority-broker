# QA PLAN — TASK-0006: Measurement and rolling-wave replanning gate

## Independent TASK-first acceptance baseline

This section was fixed before reading `PLAN.md`.  Its only planning inputs
were `TASK.md`, the current `backlog.json`, and the append-only source of truth
`/home/ubuntu/git/agent-harness-work/lap30/events.jsonl` (74 valid schema-v1
events at assessment time).  PLAN may supply a reproducible implementation of
these checks, but may not weaken or replace them.

QA owns independent evidence and this `QA_PLAN.md` only.  It does not edit
PLAN, backlog, product/test source, Git state, the operational repository, or
the Lap30 log.  TASK-0006 adds no authority-broker production code and does
not implement sudo, push, audit, release, or canary behavior.

### Canonical completion and correction semantics

A completed task is identified only by an event with
`event == "lap_completed" && status == "completed"`.  A correction event whose
own status is `completed` is not another completed task.  The canonical log
therefore contains exactly four completed tasks for this gate: TASK-0001,
TASK-0003, TASK-0004, and TASK-0005.  Evidence must contain one and only one
task row for each of those IDs and no row for an incomplete task.

Corrections are applied by `annotations.corrects_event_id`.  The original
event remains raw history, while corrected fields/classifications are the
effective values.  Evidence must distinguish at least `raw_historical` from
`effective`; it must never place a superseded and replacement value together
in one unlabeled current-value list.

In particular, TASK-0005's raw stop classification `requirement_gap` is
explicitly corrected to effective `planning_defect` by
`task0005-lap01-correct-dev-class`.  A raw-history appendix may preserve both
with event IDs and `superseded_by`, but current
`failure_classifications` must not present `requirement_gap` as still
effective.  Keeping `requirement_gap` and `planning_defect` together without
that distinction destroys the correction meaning and is FAIL.  TASK-0004's
two initial planned-SLOC values of 417 are likewise corrected to 320; its
later approved final plan is 392.  TASK-0003's correction preserves an unknown
failure occurrence time and must not be used to manufacture a timestamp or
duration.

### Required completed-task evidence

Every completed-task row must have all of these named fields, even when the
canonical value is unavailable:

- `planned_production_sloc`, `actual_production_sloc`, and `test_loc`;
- `plan_minutes`, `dev_minutes`, `review_minutes`, `qa_minutes`, and
  `ci_push_merge_minutes`;
- `retries`;
- effective `failure_classifications`; and
- source event IDs, derivation rule, units/rounding, and a missing-value reason
  for each derived or null field.

Canonical null is evidence, not a blank to fill.  An unavailable timing stays
JSON `null` (or an equivalently explicit unavailable marker) with its source
reason.  Zero is permitted only when the canonical record proves zero.  Do
not infer stage time from retrospective `ts`, split a lap total among stages,
copy one task's values to another, estimate from SLOC, or turn absent timing
into zero.  Checkpoints and repeated stop/failure/finish snapshots are not
independent durations and must not be summed twice.

The SLOC/test facts from the effective completed records are independently
reproducible as follows:

| Task | Planned production SLOC | Actual production SLOC | Test LOC |
| --- | ---: | ---: | ---: |
| TASK-0001 | 130 | 98 | 116 |
| TASK-0003 | 220 | 135 | 216 |
| TASK-0004 | 392 | 367 | 496 |
| TASK-0005 | 175 | 151 | 1257 |

Task-level retry evidence must cover the entire completed cycle across laps,
not merely the final Git lap's local retry field and not a sum of repeated
cumulative snapshots.  If a single aggregate is used, its derivation must be
declared and match the propagated canonical counter (independently expected
maxima: TASK-0001 0, TASK-0003 1, TASK-0004 4, TASK-0005 1).  Per-stage retry
detail may supplement but not contradict that total.

### Independent P0 matrix

| ID | Criterion | Required PASS evidence / failure condition |
| --- | --- | --- |
| Q6-01 | Canonical source integrity and completed-task set | Parse every JSONL line, require schema version 1, unique nonempty event IDs, valid correction targets, and exactly the four `lap_completed/completed` task IDs above. Missing/duplicate rows, treating corrections as tasks, or using another log as truth is FAIL. |
| Q6-02 | Required-field completeness | Each completed task has every named metric field, provenance, units, and null reason. Missing keys, silent omission, placeholder text, copied estimates, or invented zero/non-null values is FAIL. |
| Q6-03 | Correction-aware values and classifications | Apply planned-SLOC and classification corrections without deleting raw history. Effective TASK-0005 classification is `planning_defect`, while the old `requirement_gap` is labeled superseded. An undifferentiated list containing both is FAIL. Unknown correction occurrence time remains unknown. |
| Q6-04 | Timing arithmetic and preflight exclusion | Recompute each observable stage duration from nonduplicated canonical evidence, preserve unobservable fields as null, and exclude every `stage == "preflight"` event from cycle/stage totals. Summing snapshots twice, using recording timestamps as occurrence times, or counting preflight is FAIL. |
| Q6-05 | Retry and failure completeness | Reproduce task-wide retries across all laps and all effective failure classifications with source event IDs. Final-lap-only retry counts, null-as-zero, lost historical failures, or superseded classifications presented as current are FAIL. |
| Q6-06 | Replanning arithmetic and contingency | Show formulas in one unit, state rounding, independently recompute every total/range, and apply a multiplicative 20% time contingency (`observed_time * 1.20`, conservatively rounded). Applying 20 percentage points to SLOC, omitting contingency, or hiding arithmetic is FAIL. |
| Q6-07 | Sparse-evidence sizing discipline | Derive the next wave from observed completed cycles and qualitative scope/risk; explicitly disclose sparse/null timing. No constant SLOC/minute, LOC velocity, linear throughput, or imputed fixed delivery rate may size a task. A fixed SLOC-throughput assumption is FAIL. |
| Q6-08 | Rolling-wave boundary | Convert only the next 2–3 ordered future milestones into detailed executable task contracts. Milestones beyond that wave remain non-executable reserves without PLAN/DEV/branch/PR-ready detail. Implementing future product behavior or creating more than three detailed contracts is FAIL. |
| Q6-09 | Next measurement gate | Insert one explicit measurement/replanning task immediately after the selected 2–3 contracts, depending on the last selected task. It must repeat this evidence contract before another wave becomes executable and add zero product SLOC. Missing, misplaced, or bypassable remeasurement is FAIL. |
| Q6-10 | Backlog/index/contract consistency | JSON parses; schema/strategy remains coherent; task and milestone IDs are unique; dependency references exist and are acyclic; converted IDs, titles, status, executable flags, dependencies, ceilings/reserves, added SLOC, and contract paths agree exactly between backlog index and each `tasks/<ID>/TASK.md`. A converted milestone must not remain simultaneously executable and reserved. |
| Q6-11 | Zero-product-SLOC and scope | Candidate diff adds exactly 0 production SLOC and cumulative production remains 751 (and therefore `<=820`). Only planning evidence, backlog/index, and the next 2–3 plus measurement contracts may change. Any `.go`/shipped source or sudo/push/audit/release/canary implementation is FAIL. |
| Q6-12 | Cap, reserves, no-compression, and shedding order | Preserve the 1500 mandatory-v1 ceiling, >90% re-estimation trigger, no-compression text, exact mandatory-control list, and exact ordered feature-shedding sequence from current backlog. No mandatory control may be shed; inability to fit mandatory v1 within 1500 is `requirement_gap`. Reordering/omitting shedding steps, using compression, or presenting reserve as implemented SLOC is FAIL. |

All Q6-01 through Q6-12 are P0.  A skipped row is FAIL unless it is first
classified as `qa_plan_defect`, `requirement_gap`, or `environment_issue` and
the main Agent approves a revised disposition before merge.

### Independent arithmetic and consistency procedure

QA and REVIEW independently run read-only checks equivalent to:

```sh
jq -e . backlog.json
jq -s 'length > 0 and all(.[]; .schema_version == 1 and (.event_id | type == "string" and length > 0))' \
  /home/ubuntu/git/agent-harness-work/lap30/events.jsonl
jq -s 'group_by(.event_id) | all(.[]; length == 1)' \
  /home/ubuntu/git/agent-harness-work/lap30/events.jsonl
jq -s 'map(select(.event == "lap_completed" and .status == "completed") | .task_id) | unique' \
  /home/ubuntu/git/agent-harness-work/lap30/events.jsonl
git diff --check
git status --short
```

Also validate every correction target exists and precedes its correction in
the same task/lap, independently recompute evidence-table fields and all 20%
contingency figures, compare converted contracts with backlog by ID, and walk
dependencies to prove ordering/no cycles.  Search both backlog and contracts
for forbidden detailed future scope and for fixed throughput expressions such
as SLOC/minute or LOC velocity.

Run the TASK production-SLOC count from the merged candidate and compare it
with the baseline; the candidate must remain exactly 751 cumulative and add
zero.  Documents, JSON, and task contracts do not count as product SLOC.  A
source diff or a changed exact count fails Q6-11 even if still below 820.

### Failure classification and QA decision

| Classification | Meaning and disposition |
| --- | --- |
| `implementation_defect` | Candidate evidence/backlog/contracts violate a P0, arithmetic, correction semantics, exact index/contract mapping, 0-SLOC boundary, cap, no-compression rule, or rolling-wave limit. FAIL and return to DEV/planning implementation with the smallest reproducible mismatch. |
| `regression` | Candidate changes a completed-task fact, prior mandatory control, schema/index invariant, or production baseline that previously passed. FAIL with source comparison. |
| `qa_plan_defect` | This fixture or expectation contradicts TASK/backlog/canonical events while the candidate satisfies them. Pause and amend/reapprove QA planning. |
| `requirement_gap` | Evidence cannot support a required next-wave decision, or mandatory-v1 controls cannot fit within 1500 without weakening. Return to task authority; do not invent metrics or compress. |
| `environment_issue` | Required read-only tooling or source access prevents assessment and is demonstrably not candidate behavior. Record the blocker and rerun; do not synthesize results. |

QA PASS requires independent REVIEW PASS; Q6-01 through Q6-12; exact source,
arithmetic, correction, timing/null, retry/classification, contingency,
rolling-wave, contract/index, scope, and 0-SLOC evidence.  This plan authorizes
no product work, operational-log write, stage, commit, or merge.

## PLAN reconciliation (read after independent baseline)

`PLAN.md` was read only after the independent section above was fixed.  Most
of its source selection, arithmetic, scope, and rolling-wave rules safely
implement the independent criteria, but one correction-semantics defect is a
blocking mismatch.

| Independent criterion | PLAN evidence | QA reconciliation |
| --- | --- | --- |
| Canonical snapshot/completed set | PLAN filters the current 74-line log to the 72 events belonging to TASK-0001/0003/0004/0005, requires unique IDs/schema validation, and measures one `lap_completed` record per Task. | **MATCH / PASS.** TASK-0006 start/preflight events are excluded from the completed-cycle snapshot without editing the source log. |
| Complete fields, null preservation, and preflight exclusion | PLAN names all required SLOC/test/timing/retry/classification/provenance fields, preserves non-derivable values as null, forbids elapsed splitting and timestamp guesses, and subtracts only explicitly measured preflight. | **MATCH / PASS.** QA will additionally require a reason beside every null and reject any inferred zero. |
| Stage and cycle arithmetic | PLAN pairs same-task/lap/stage/attempt start and terminal events and sums attempts without using a cumulative elapsed snapshot as a duration. Its TASK-0004 totals (DEV 878000 ms, REVIEW 615000 ms, QA 244000 ms, Git 177000 ms) and TASK-0005 totals (DEV 625000 ms, REVIEW 392000 ms, QA 347000 ms, Git 178000 ms) reproduce the canonical pairs. | **MATCH / PASS.** Earlier Task stage splits remain null. Exact formulas, source IDs, raw milliseconds, and two-decimal minute display remain required. |
| SLOC/test/retry arithmetic | PLAN uses final completed incremental values, derives cumulative `98 + 135 + 367 + 151 = 751`, and uses maximum propagated retries 0/1/4/1 instead of summing snapshots. | **MATCH / PASS.** Corrected/replanned earlier estimates must remain provenance notes, not replace final values. |
| Correction-aware effective classifications | Independent rule requires raw historical and effective classification to be separate. TASK-0005 `task0005-lap01-correct-dev-class` explicitly says `old_classification=requirement_gap`, `new_classification=planning_defect`, and corrects the earlier stop event. | **MISMATCH / FAIL (`planning_defect`).** PLAN aggregation rule 7 makes one ordered unique set of every non-null classification and says not to erase `requirement_gap`; its reproduced TASK-0005 row consequently lists `requirement_gap`, `planning_defect`, and `environment_issue` together without a raw/effective distinction. That represents the superseded value as current and damages canonical correction meaning. |
| Sparse-evidence sizing and contingency | PLAN rejects SLOC/minute, LOC velocity, averages, and regression; it uses TASK-0004's observed non-preflight complex-boundary cycle and conservatively rounds `62.98 * 1.20` to a 76-minute minimum, while push time remains unknown. | **MATCH / PASS, subject to explicit rounding provenance.** Null remains null, contingency is time-only, and reserve SLOC is not a schedule forecast. |
| Only next 2–3 contracts and next gate | PLAN converts the first two reserves to TASK-0007 sudo and TASK-0008 push, then inserts zero-SLOC TASK-0009 measurement depending on TASK-0008. Audit-release and clean-canary remain non-executable reserves. | **MATCH / PASS.** This is two feature contracts plus the required immediate measurement gate; no later contract becomes executable. |
| Zero SLOC, cap, compression, and shedding | PLAN limits DEV to measurement/backlog/three contract documents, keeps cumulative 751, sets next ceilings 950/1400 and global 1500, retains >90% re-estimation, exact ordered shedding, mandatory controls, and no compression. | **MATCH / PASS.** Candidate must still prove an unchanged exact production count and exact backlog/contract/index consistency. |

### Required PLAN correction before approval

PLAN must replace rule 7 and its reproduced TASK-0005 classification cell with
an explicit correction-aware representation, for example:

```text
effective_failure_classifications: [planning_defect, environment_issue]
raw_historical_classifications:
  - value: requirement_gap
    source_event_id: task0005-lap01-dev-stop-sloc
    superseded_by: task0005-lap01-correct-dev-class
```

Exact structure may vary, but the effective list must not contain the
superseded `requirement_gap`; raw history must remain visible with correction
provenance.  The same general rule must apply to every correction without
inventing occurrence time or erasing source history.  `MEASUREMENT.md` must
follow the corrected rule.

Final reconciliation decision: **FAIL — PLAN correction required**.
Classification: **`planning_defect`**, not an implementation defect,
requirement gap, QA-plan defect, or environment issue.  The next-wave choice
and arithmetic remain credible, but DEV must not start until PLAN is corrected,
independently re-read, and both PLAN/QA gates are approved.  This QA_PLAN does
not edit or silently compensate for the defective PLAN rule.

## PLAN reconciliation Revision 2

Revision 2 was independently re-read after the prior FAIL.  The previous
finding is resolved without weakening the independent baseline.

### Finding disposition: resolved

PLAN's measurement schema now has separate
`raw_failure_classifications` and `effective_failure_classifications` fields.
Rule 7 requires every raw classification record to retain its
`classification`, `source_event_id`, and nullable `superseded_by`, validates a
classification correction against an earlier same-Task target and matching
old/new values, and derives the effective ordered unique set only from raw
records that are not superseded.

For TASK-0005 it now requires exactly the correction edge:

```text
task0005-lap01-dev-stop-sloc
  -> task0005-lap01-correct-dev-class
```

The raw historical `requirement_gap` remains visible with the correction event
as `superseded_by`; it is absent from the effective set.  The reproduced table
shows effective TASK-0005 classifications as only `planning_defect` and
`environment_issue`.  `MEASUREMENT.md`, REVIEW, and QA are all explicitly
required to prove that separation.  This satisfies Q6-03 and the exact prior
remediation request; no source history is deleted and no free-text
reclassification is permitted.

### Regression check of previously passing areas

| Area | Revision 2 result |
| --- | --- |
| Canonical snapshot/completed set | **PASS unchanged:** 72 selected events for exactly four `lap_completed/completed` Tasks, with schema/ID/lap-sequence checks and the operational log immutable. |
| Completeness/null/preflight | **PASS unchanged:** all required metrics/provenance remain named; non-derivable values remain null; zero requires evidence; preflight is excluded and cannot be assigned to a stage. |
| Timing/SLOC/test/retry arithmetic | **PASS unchanged:** matched stage-pair totals, final incremental SLOC values, cumulative 751, test LOC, and propagated retry maxima 0/1/4/1 are unchanged. |
| 20% contingency/sparse evidence | **PASS unchanged:** time-only `*1.20` contingency, stated rounding, unknown push time, and prohibition on fixed SLOC throughput remain intact. |
| Rolling-wave boundary | **PASS unchanged:** only TASK-0007 and TASK-0008 are converted feature contracts, followed immediately by zero-SLOC TASK-0009 measurement; later reserves remain non-executable. |
| Scope/index/contracts | **PASS unchanged:** only measurement, backlog, and TASK-0007/0008/0009 contracts may change; exact ID/dependency/ceiling/reserve consistency remains an independent gate. |
| Zero SLOC/caps/guardrails | **PASS unchanged:** cumulative production remains 751, TASK-0006 adds zero, next ceilings remain 950/1400, global cap remains 1500, and >90% re-estimation, mandatory controls, exact shedding order, and no-compression rules remain binding. |

Revision 2 final reconciliation decision: **PASS — prior finding resolved and
no regression found**.  Classification: **none/open findings 0**.  The prior
`planning_defect` remains historical disposition evidence, not a current
failure.  This is PLAN/QA compatibility approval only; DEV, independent
REVIEW, independent QA, exact candidate arithmetic/index/SLOC checks, and
main-owned Git boundaries remain mandatory.
