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

## Revision 3 TASK-first acceptance baseline (before reading PLAN Revision 3)

This section was fixed before reading PLAN Revision 3.  Its inputs were only
the explicit user decision for this revision, the existing TASK/QA
requirements, current `backlog.json`, and the existing TASK-0007 through
TASK-0009 contracts.  The explicit Revision 3 decision supersedes the older
wave-size and cap assumptions where they conflict; it does not waive role
separation, gates, evidence, mandatory controls, or no-compression rules.

### Revision 3 decision and arithmetic boundary

- Plan exactly **six detailed executable Task contracts** for the approved
  wave.  A measurement contract counts as one of the six when it is detailed
  and executable; reserves outside the wave remain non-executable.
- The cumulative production planning **target is 1500 SLOC**.  The absolute
  **hard ceiling is 1800 SLOC**.  From the measured 751 baseline, these leave
  749 target lines and 1049 hard-cap lines.  Contract ceilings must be
  monotonic and their planned increments must reconcile without overlap or
  double counting.
- The 300-line interval from target to hard cap is review/remediation
  contingency, not a pre-approved feature budget.  Projected or measured use
  above 1500 stops product work and returns to PLAN/QA with explicit cause,
  shedding applicability, and revised evidence.  A candidate above 1800 is
  unconditional FAIL.  The >90% per-Task re-estimation trigger remains
  binding before DEV.
- TASK-0006 itself still adds **0 production SLOC**.  Revision 3 planning may
  change contracts/index/evidence only; it must not implement any wave
  feature.

The explicit 1800 hard ceiling replaces the older statement that mandatory-v1
over 1500 is immediately a terminal requirement gap.  The new effective rule
is: target overflow above 1500 requires replan and the exact shedding audit;
mandatory controls remain unsheddable; inability to fit all mandatory scope
idiomatically within 1800 is `requirement_gap`.  Existing documents/backlog
must be revised consistently before execution; leaving contradictory active
820/1500 terminal rules is FAIL.

### Six-contract completeness and ownership

Each of the six contracts must contain: unique ID/title; dependency; exact
owned paths/responsibility; exclusions; preflight; acceptance and denial
cases; planned increment and cumulative target/hard ceiling; 2-Lap delivery
split; focused DEV evidence; independent REVIEW and QA evidence; measurement
fields; and stop/replan conditions.  IDs, titles, dependencies, executable
status, ceilings, increments, and contract paths must agree exactly with
`backlog.json`.

One contract must explicitly own the missing **production seed / daemon /
backend assembly boundary** before sudo or push can depend on live authority.
It must define production process startup/shutdown, bounded configuration and
secret injection, construction of lease/TOTP state, mapping of fixed
`ready`/`otp` IPC operations to the backend, fail-closed unavailable/restart
behavior, and the narrow interface consumed by later sudo/push tasks.  This
responsibility may not be hidden in preflight, split without an integration
owner, or duplicated inside sudo/push/audit work.  If no contract owns it, the
wave is non-executable and FAIL.

Existing TASK-0007 sudo/no-cache and TASK-0008 restricted-push/token-custody
requirements remain mandatory, although Revision 3 may renumber or resequence
them to insert the production seed.  Later contracts may cover the remaining
mandatory audit/release/canary/measurement boundaries only with exclusive
ownership and dependency on the production chain.  No contract may silently
drop prior readiness, TOTP replay/rate/absolute expiry, strict IPC,
SO_PEERCRED, redaction, sudo live/no-cache, non-force push, minimal audit, or
source-free attestation controls.

### Two-Lap fit and safe pipeline concurrency

Every contract must show a credible two-Lap path, not merely label itself
“two laps”:

1. Lap A: prerequisites/preflight, approved PLAN and independent QA_PLAN,
   bounded DEV, focused checks, and a gate-ready candidate; and
2. Lap B: independent REVIEW, independent QA, main-owned final checks/Git/PR/
   merge, plus complete measurement evidence.

The estimate must identify fixtures, permission/elevation needs, dependency
risks, expected production/test scope, and a stop condition.  Two-Lap fit is a
planning hypothesis, never permission to skip PLAN/DEV/REVIEW/QA, combine
roles, compress source, truncate evidence, or merge on elapsed-time pressure.
If a Task cannot finish safely in two laps, it stops, classifies the cause,
and replans; downstream DEV remains blocked.

Pipeline parallelism is allowed only where safe and explicitly evidenced:

- PLAN/QA_PLAN or read-only exploration for a later Task may overlap an
  earlier independent gate only when it does not assume an unmerged interface;
- DEV for a dependent Task cannot start until its dependency is merged and
  preflight revalidates the interface;
- one Task's PLAN -> DEV -> independent REVIEW -> independent QA -> main merge
  order remains sequential, with DEV/REVIEW/QA roles separated;
- parallel workers have disjoint file ownership and never stage, commit,
  merge, write `.git`, or concurrently edit shared backlog/contracts/lap logs;
- main alone owns shared locks, scope checks, Git, merge, and external writes;
  a failed upstream gate cancels/invalidates dependent speculative planning.

“Parallel” that overlaps dependent DEV, shares mutable files without a lock,
combines DEV with reviewer/QA roles, bypasses dependency merge, or makes an
operational log multi-writer is FAIL.

### Measurement gates across the six-Task wave

Expanding from 2–3 to six detailed contracts does not remove rolling-wave
measurement.  The dependency chain must contain zero-production measurement
gates so that **no more than 2–3 feature Tasks** execute between a prior
measurement and the next gate.  At least the final contract must prevent any
post-wave PLAN/DEV from starting until it passes REVIEW, QA, and merge.  Each
measurement gate repeats canonical completed-cycle completeness, null and
correction semantics, preflight exclusion, retries/classifications, actual
SLOC/test/time arithmetic, 20% time contingency, cap reconciliation, and
fixed-throughput prohibition.  It may stop or resize remaining provisional
contracts when evidence changes.

### Revision 3 independent P0 matrix

| ID | Criterion | Required PASS evidence / failure condition |
| --- | --- | --- |
| R3-01 | Six-contract/index completeness | Exactly six detailed executable wave Tasks exist, no seventh detailed reserve exists, and every contract/index field and dependency agrees. Missing, duplicate, dangling, cyclic, or simultaneously reserved/executable items FAIL. |
| R3-02 | Target/hard-cap arithmetic | Baseline 751, per-Task increments, cumulative ceilings, target 1500, and hard 1800 reconcile monotonically with no overlap. Above-target work stops/replans; above-hard fails. Treating the 300 interval as ordinary scope or retaining contradictory effective caps FAILS. |
| R3-03 | Production seed/backend owner | One predecessor contract exclusively and completely owns daemon/backend assembly and proves fixed ready/OTP routing, secret/config boundary, lifecycle, and fail-closed behavior before sudo/push. Missing, preflight-only, duplicated, or implicit ownership FAILS. |
| R3-04 | Feature boundary and dependencies | Sudo live/no-cache, restricted non-force push/token custody, minimal audit, source-free release/attestation, canary/rollback evidence, and measurement responsibilities have explicit nonoverlapping owners and dependency order. Mandatory omissions or circular/unsafe ordering FAIL. |
| R3-05 | Credible two-Lap fit | Each Task has Lap A/Lap B work, fixtures, expected production/test scope, elevation/dependency risks, gate evidence, and stop/replan criteria. A label without bounded work or any proposed gate/role shortcut FAILS. |
| R3-06 | Safe pipeline parallelism | Only disjoint planning/read-only work overlaps; dependent DEV waits for merge/preflight; shared writes and Git are main/lock owned; DEV/REVIEW/QA remain separate and sequential per Task. Unsafe shared or dependent concurrency FAILS. |
| R3-07 | Measurement cadence | Zero-SLOC measurement gates leave no span longer than 2–3 feature Tasks and the final gate blocks any next wave until PASS/merge. Missing or bypassable measurement, or a six-feature run without interim evidence, FAILS. |
| R3-08 | Role/gate contract | Native role/model contracts, approved PLAN+QA_PLAN before DEV, independent REVIEW+QA, reviewer checks, parent Git ownership, and failure classification remain mandatory in every Task. Any compressed two-Lap process or combined role FAILS. |
| R3-09 | Mandatory/no-compression/shedding | Mandatory controls are unsheddable; exact ordered shedding is retained and applied only after stop/replan; ordinary idiomatic structure, tests, comments, errors, and names cannot be compressed for caps/laps. Any weakened control, reordered shedding, or semantic compression FAILS. |
| R3-10 | TASK-0006 zero-product scope | Revision 3 changes only QA/PLAN/backlog/contracts/measurement evidence and adds no product/test implementation or operational-log write. Exact production SLOC remains 751 for TASK-0006. Any product delta or Git/lap-log mutation FAILS. |

All R3-01 through R3-10 are P0.  Revision 3 QA PASS requires independent
REVIEW PASS plus all prior measurement/correction integrity checks that remain
applicable.  User-authorized wave/cap changes supersede only the conflicting
old limits; they do not erase historical measurement evidence.

Independent pre-PLAN Revision 3 decision: **CONDITIONAL / GATE CLOSED UNTIL
PLAN RECONCILIATION**.  Six Tasks and 1500/1800 can be credible only if the
updated PLAN supplies complete cap arithmetic, an explicit production-seed
owner, safe dependencies, real two-Lap work breakdowns, safe concurrency, and
measurement cadence as defined above.

## PLAN reconciliation Revision 3

PLAN Revision 3 was read only after the TASK-first section above was fixed.
Its pipeline restrictions and raw cap sums are coherent, but the proposed wave
is not executable because mandatory ownership, seed-boundary detail,
two-Lap evidence, and cap-trigger semantics remain incomplete.

### Reconciliation matrix

| Revision 3 P0 | PLAN evidence | QA decision |
| --- | --- | --- |
| R3-01 Six contracts/index completeness | PLAN names exactly TASK-0007 through TASK-0012: four production Tasks and two zero-SLOC measurement gates. Dependencies are linear. | **PARTIAL / FAIL.** The intended six IDs are clear, but PLAN does not enumerate the TASK-0006 allowed output paths for all six contract files/backlog nor require each future contract's exact product-path ownership. Detailed index/contract consistency cannot yet be guaranteed. |
| R3-02 Target/hard arithmetic | Target increments are `199 + 200 + 0 + 150 + 200 + 0 = 749`; `751 + 749 = 1500`. Named contingency is `100 + 100 + 0 + 50 + 50 + 0 = 300`; hard cumulative ends at 1800. Borrowing requires revised PLAN/QA. | **ARITHMETIC PASS, ENFORCEMENT FAIL.** PLAN also says a forecast above 90% of each Task's *target cumulative cap* stops before DEV, while every planned Task forecast equals 100% of that target cap (950/950, 1150/1150, 1300/1300, 1500/1500). The contracts therefore trigger their own mandatory stop before every DEV. The trigger basis/headroom must be made coherent without weakening re-estimation. |
| R3-03 Production seed/backend owner | TASK-0007 is labeled daemon/backend assembly and deterministic seed; TASK-0008 depends on it. | **MISMATCH / FAIL.** The contract summary does not define owned production paths, process entrypoint/startup/shutdown, bounded configuration and secret injection, construction of lease/TOTP state, or exact fixed `ready`/`otp` IPC-to-backend routing. “Seed only the fixture state” is not a complete production-seed boundary. These cannot remain preflight decisions or implicit later work. |
| R3-04 Mandatory feature ownership | TASK-0008 owns sudo; TASK-0010 owns local push policy; TASK-0011 owns custody/system-Git push. PLAN repeats minimal audit and source-free attestation in its mandatory list. | **MISMATCH / FAIL.** No one of the six contracts owns minimal external-trace audit or source-free attested release. The 1500 target is fully allocated before those controls, and TASK-0012 defers later reserves. Automated canary may be shed only with retained manual runbook/evidence, but no contract owns that retained evidence either. Mandatory text without a Task owner/cap/dependency is not an executable v1 plan. |
| R3-05 Credible two-Lap fit | A generic Lap 1/Lap 2 protocol and “Max DEV laps 2” column exist; broad completion evidence is named. | **MISMATCH / FAIL.** Per-Task contracts lack concrete Lap A/B work breakdown, allowed product/test paths, fixture/elevation/permission needs, expected test scope, size basis, and Task-specific split/stop evidence. A common label and cap alone do not establish that daemon/secret, sudo host, or push credential boundaries fit two laps. |
| R3-06 Safe pipeline | Successor planning/read-only work may overlap REVIEW/QA; successor DEV waits for dependency merge, approved PLAN+QA_PLAN, and preflight. Roles and main-owned Git remain separate. | **MATCH / PASS.** This is safe speculative planning, not dependent DEV overlap. Shared writes/Git remain prohibited. |
| R3-07 Measurement cadence | TASK-0009 follows two feature Tasks; TASK-0012 follows two more and blocks later reserves until PASS/merge. Both add zero SLOC. | **MATCH / PASS.** No span exceeds two feature Tasks; final remeasurement is not bypassable. |
| R3-08 Role/gate contract | PLAN explicitly retains PLAN -> DEV -> independent REVIEW -> independent QA -> main Git and requires both plans before DEV. | **MATCH / PASS.** The two-Lap envelope does not textually waive gates, though R3-05 feasibility remains unproved. |
| R3-09 Mandatory/no-compression/shedding | PLAN retains mandatory controls, no-compression, >90% re-estimation, and the seven-step shedding order. | **PARTIAL / FAIL through R3-04.** Text and order are retained, but unowned audit/attestation/manual-canary evidence makes the mandatory-control implementation plan incomplete. Contingency cannot silently absorb those omitted responsibilities. |
| R3-10 TASK-0006 zero-product scope | PLAN states TASK-0006 adds zero product SLOC and authorizes no product/Git/log work. | **MATCH / PASS.** Candidate execution must later prove exact 751 and document-only scope. |

### Required remediation before approval

PLAN Revision 3 must be revised and independently reconciled again to:

1. assign every mandatory v1 result to one of exactly six contracts, including
   minimal audit, source-free attestation, and manual canary/rollback evidence
   if the automated canary is shed;
2. make one production-seed contract explicitly own daemon entrypoint and
   lifecycle, configuration/secret injection, state construction, fixed
   ready/OTP backend routing, fail-closed restart/unavailable behavior, and
   nonoverlapping product/test paths;
3. rebalance or repartition target/hard increments so all mandatory owners fit
   within target 1500 or an explicitly approved replan path below hard 1800;
4. resolve the >90% trigger contradiction by defining a ceiling/forecast basis
   that preserves real pre-DEV headroom and still stops risky projections;
5. provide a Task-specific two-Lap fixture, path, test, permission/elevation,
   size, gate, and stop breakdown for every one of the six contracts; and
6. list the exact contract/backlog files TASK-0006 may create/update and require
   index/dependency/ownership consistency before any contract is executable.

Revision 3 final reconciliation decision: **FAIL — planning revision
required**.  Classification: **`planning_defect`**.  The raw target/hard
addition, pipeline safety, role separation, and measurement cadence pass, so
this is not yet evidence of an unavoidable `requirement_gap`; a six-contract
repartition may resolve it.  DEV remains blocked.  This QA_PLAN does not edit
PLAN, backlog, product source, Git state, or the operational log.

## PLAN reconciliation Revision 4

PLAN Revision 4 was independently re-read against the TASK-first Revision 3
P0 matrix and all five Revision 3 remediation findings.  The revised plan now
defines a coherent six-Task executable wave while explicitly preserving later
mandatory work outside that wave.

### Revision 4 reconciliation matrix

| Revision 3 finding / regression area | Revision 4 evidence | QA decision |
| --- | --- | --- |
| Mandatory audit, attestation, and manual canary ownership | The six-Task wave is explicitly not represented as v1 completion.  It establishes the mandatory later lineage `TASK-0012 measurement PASS+merge -> MILESTONE-audit-attestation -> MILESTONE-manual-canary-rollback`; later work remains non-executable until that gate.  Minimal audit, source-free attestation, and retained manual canary/rollback evidence are named mandatory results and cannot be shed or silently absorbed into this wave. | **MATCH / PASS.** The earlier false implication that all mandatory v1 work was allocated inside the six Tasks is removed.  Later ownership, order, and entry gate are explicit. |
| Forecast, target cap, hard cap, and 90% trigger | Starting at 751, planned additions are `90 + 120 + 0 + 130 + 220 + 0 = 560`, giving cumulative forecasts `841, 961, 961, 1091, 1311, 1311`.  The corresponding target caps are `950, 1100, 1100, 1250, 1500, 1500`; exact 90% triggers are `855, 990, 990, 1125, 1350, 1350`, and every forecast is below its trigger.  Task hard caps are `1050, 1250, 1250, 1400, 1650, 1650`; the system hard limit remains 1800. | **MATCH / PASS.** `751 + 560 = 1311`; forecast headroom is 189 to target 1500 and 489 to system hard 1800.  The earlier self-blocking 100%-of-target forecasts are gone.  Task contingency remains stop/replan space, not ordinary allocation or silent borrowing from later work. |
| TASK-0007 production daemon, seed, and backend boundary | TASK-0007 owns `cmd/codex-authority/main.go` and `internal/backend/runtime.go` plus their focused tests.  It defines the single daemon entrypoint, startup/readiness, signal shutdown/restart, bounded root-owned mode-0600 seed/config read once, validation and in-process copying, lease/TOTP state construction, exact fixed `ready`/`otp` routing, non-disclosure, and fail-closed missing/malformed/oversize/owner/mode/backend/restart cases. | **MATCH / PASS.** Production lifecycle, root-owned bounded injection, state assembly, routing, failure behavior, paths, and predecessor ownership are no longer implicit. |
| Per-Task two-Lap executability | TASK-0007 through TASK-0012 each name owned production/test or evidence paths, concrete Lap 1 candidate work, Lap 2 independent REVIEW/QA evidence, required fixture/environment constraints, cumulative-size basis, and a Task-specific split/stop condition.  TASK-0008 includes isolated sudo/PAM elevation fixtures; TASK-0010 a local bare push fixture; TASK-0011 fake token/system-Git/local-bare/live-lease and capture sinks; TASK-0009/0012 are bounded zero-SLOC measurement gates. | **MATCH / PASS.** The generic two-Lap label has been replaced by bounded path, fixture, test, permission, gate, and split criteria for every contract. |
| Exact TASK-0006 outputs and index consistency | Revision 4 enumerates exactly seven DEV outputs: `backlog.json` and `tasks/TASK-0007/TASK.md` through `tasks/TASK-0012/TASK.md`.  It forbids other product, test, measurement, review, QA, log, or Git outputs and requires backlog/contract agreement across IDs, dependencies, ownership, paths, caps, risks, fixtures, gates, and later lineage, with no seventh executable contract. | **MATCH / PASS.** Output scope and bidirectional index/contract consistency are explicit and auditable. |
| Safe pipeline and role separation regression | Each Task retains PLAN -> DEV -> independent REVIEW -> independent QA -> main-owned Git.  Successor overlap is limited to disjoint PLAN/TASK-first QA/read-only work; dependent DEV waits for predecessor merge, preflight, and approved PLAN/QA_PLAN.  Shared writes and Git remain parent/main owned. | **MATCH / PASS.** No dependent DEV overlap, combined role, gate bypass, or shared mutable writer was introduced. |
| Measurement cadence regression | Zero-SLOC TASK-0009 follows TASK-0007/0008; zero-SLOC TASK-0012 follows TASK-0010/0011 and blocks the later mandatory lineage until its REVIEW, QA, and merge pass.  Both repeat measurement completeness, correction/null semantics, actual SLOC/test/time arithmetic, contingency, cap reconciliation, retry/classification, and fixed-throughput prohibition. | **MATCH / PASS.** No run exceeds two production Tasks between measurement gates, and the final gate remains non-bypassable. |
| Mandatory controls, shedding, and no-compression regression | The revised plan retains readiness, TOTP, strict IPC/SO_PEERCRED, daemon lifecycle and bounded seed, sudo live/no-cache, non-disclosure, restricted non-force push, minimal audit, source-free attestation, and manual canary/rollback evidence.  It retains the exact ordered shedding policy and prohibits semantic compression or weakened tests/errors/names. | **MATCH / PASS.** The six-Task scope boundary does not waive later mandatory controls. |

### Reserve and lineage interpretation

The 189-SLOC target reserve and 489-SLOC system-hard reserve are measured from
the Revision 4 forecast of 1311.  The current wave's per-Task hard guards may
only be reached through an explicit stop/replan and do not pre-authorize use of
that reserve.  TASK-0012 must reconcile actual SLOC before either later
milestone becomes executable.  Thus the plan neither promises that all 489
SLOC will remain after a future approved contingency replan nor silently
allocates it now; it preserves the required measurement and approval boundary
below the unconditional 1800 stop.

### Revision 4 decision

All R3-01 through R3-10 requirements applicable to the revised six-Task wave
now reconcile.  In particular, all five Revision 3 findings are closed, and
the previously passing pipeline, role-separation, measurement-cadence, zero-
product-scope, mandatory-control, shedding, and no-compression properties have
not regressed.

Revision 4 final reconciliation decision: **PASS**.  Classification:
**`none`**.  TASK-0006 DEV may proceed only within the exact seven-output
scope and the normal approved PLAN/QA_PLAN, independent REVIEW, independent
QA, and parent/main Git gates.  This reconciliation changes only this
`QA_PLAN.md`; it does not edit PLAN, backlog, contracts, product/test source,
Git state, or the operational lap log.

## PLAN reconciliation Revision 5

PLAN Revision 5 was independently reconciled against the TASK-first Revision
3 P0 matrix, the Revision 4 entrypoint-ownership defect, and the existing
source boundary.  Revision 5 correctly preserves the current non-privileged
client, but its later push registration is not reachable through the fixed
existing IPC protocol, so the six-Task wave is not yet executable as written.

### Revision 5 reconciliation matrix

| Area | Revision 5 and repository evidence | QA decision |
| --- | --- | --- |
| Existing client / daemon ownership | Existing `cmd/codex-authority/main.go` constructs `ipc.Client`, admits only `ready`/`otp`, reads OTP from stdin, and sends a request to the Unix socket.  Revision 5 leaves it and `main_test.go` unchanged, gives TASK-0007 the new privileged `cmd/codex-authority-broker/main.go` daemon entrypoint plus `internal/backend/runtime.go`, and includes both the old client and new broker in focused regression commands. | **MATCH / PASS.** The Revision 4 conflict that would have converted the established client into a daemon is resolved. |
| TASK-0007 daemon/backend boundary | TASK-0007 still exclusively owns lifecycle, readiness, signal shutdown/restart, root-owned mode-0600 bounded seed/config input, lease/TOTP state construction, fixed `ready`/`otp` routing, fail-closed cases, and bounded handler-registration seam through the new broker/runtime paths and focused tests. | **MATCH / PASS.** R3-03 remains satisfied without changing the existing client. |
| TASK-0011 registration ownership and reachability | TASK-0011 now exclusively owns `internal/backend/push_registration.go` and its test, while neither client nor broker entrypoint changes.  However, current `internal/ipc/protocol.go` defines and admits only `OperationReady` and `OperationOTP`; `validOperation` rejects every other operation.  Revision 5 gives no Task ownership of that protocol path and defines no other caller/trigger that can reach the registered push handler.  TASK-0007 expressly preserves fixed `ready`/`otp` routing, and the existing client cannot initiate push. | **MISMATCH / FAIL.** Moving registration out of the client fixes file ownership but does not create an executable IPC-to-push path.  A contract must exclusively own the bounded push operation/request schema and admission tests, or define another concrete authorized trigger, without weakening the established fixed-action client or TASK-0007 routing. |
| Exact TASK-0006 DEV outputs and index consistency | Revision 5 retains exactly `backlog.json` plus `tasks/TASK-0007/TASK.md` through `tasks/TASK-0012/TASK.md`, forbids product/test/result/log/Git outputs, and requires exact index/contract agreement across identity, dependencies, caps, paths, entrypoint, fixtures, laps, split rules, lineage, and eligibility. | **MATCH / PASS.** Seven-output and six-contract scope did not regress. |
| Forecast/cap arithmetic and reserve | The baseline remains 751; increments remain `90 + 120 + 0 + 130 + 220 + 0 = 560`; forecasts remain `841, 961, 961, 1091, 1311, 1311`, all below exact 90% triggers `855, 990, 990, 1125, 1350, 1350`.  Target/hard guards and the 189-to-1500/489-to-1800 later reserve are unchanged. | **MATCH / PASS.** No arithmetic, self-blocking-trigger, or reserve regression. |
| Six Tasks, dependencies, and mandatory lineage | TASK-0007 through TASK-0012 remain a linear six-Task wave with zero-SLOC measurement gates after each two production Tasks.  Later mandatory lineage remains `TASK-0012 PASS+merge -> MILESTONE-audit-attestation -> MILESTONE-manual-canary-rollback`, with no seventh executable contract or silent reserve borrowing. | **MATCH / PASS.** The wave still does not misrepresent v1 completion. |
| Task-specific two-Lap evidence | All six contracts retain exclusive production/test or measurement paths, concrete Lap 1/Lap 2 work, focused and full checks, fixtures/elevation constraints, expected size, and split/stop criteria.  TASK-0007 adds existing-client regression evidence and TASK-0011 adds both-entrypoint regression evidence. | **STRUCTURE PASS, EXECUTABILITY FAIL through registration reachability.** The two-Lap form is complete, but TASK-0011 cannot produce its claimed push-path evidence until the trigger/protocol owner is specified. |
| Pipeline, roles, measurement, and scope regressions | PLAN -> DEV -> independent REVIEW -> independent QA -> main Git remains sequential per Task; overlap is limited to disjoint successor planning/read-only work; dependent DEV waits for merge/preflight/approved plans.  Measurement null/correction/provenance, active/wait/retry/classification, mandatory controls, shedding order, no-compression, 0 TASK-0006 production SLOC, and Git/Lap-log prohibitions remain intact. | **MATCH / PASS.** No regression in the previously passing gates. |

### Required Revision 5 remediation

Before TASK-0006 DEV can be approved, TASK-0011's contract and backlog entry
must identify a complete, exclusive, bounded invocation path for restricted
push.  If push is a new IPC operation, TASK-0011 must own the necessary
`internal/ipc` production/test paths, exact operation and payload admission,
authorization and denial behavior, and focused regression coverage while
preserving the existing client as `ready`/`otp` only.  If push is not an IPC
operation, the plan must instead name the concrete authorized caller,
entrypoint, owned paths, fixture, tests, and lifecycle that invoke
`push_registration.go`.  Any resulting path or SLOC/cap change requires the
same index/contract arithmetic and two-Lap reconciliation before DEV.

Revision 5 final reconciliation decision: **FAIL — planning revision
required**.  Classification: **`planning_defect`**.  The Revision 4
client/daemon ownership defect is closed and every requested scope, arithmetic,
six-Task, measurement, role, and pipeline regression check passes; the sole
blocking defect is that TASK-0011's registered push handler has no admitted or
owned invocation route.  DEV remains blocked.  This reconciliation changes
only this `QA_PLAN.md`; it does not edit PLAN, backlog, contracts,
product/test source, Git state, or the operational Lap log.

## PLAN reconciliation Revision 6

PLAN Revision 6 was independently reconciled against the TASK-first Revision
3 P0 matrix and the sole Revision 5 unreachable-push-path finding.  It now
defines a complete bounded invocation route while preserving the established
client/daemon split and all previously passing wave invariants.

### Revision 6 reconciliation matrix

| Area | Revision 6 evidence | QA decision |
| --- | --- | --- |
| Dedicated caller and existing-client preservation | TASK-0011 exclusively owns new `cmd/codex-authority-push/main.go` and its test as the only supported restricted-push caller.  It accepts only configured repository identity and one permitted local source/destination ref intent and exposes no token, force, tag, delete, arbitrary refspec, remote-command, credential-environment, or generic-operation option.  Existing `cmd/codex-authority` and its tests remain unchanged and `ready`/`otp` only; TASK-0007's `cmd/codex-authority-broker` daemon entrypoint also remains unchanged. | **MATCH / PASS.** The caller is concrete and bounded without reopening the established client or conflating client and daemon ownership. |
| Push protocol schema and admission | TASK-0011 now owns existing `internal/ipc/protocol.go` and `protocol_test.go`.  It adds exactly `OperationPush` and a bounded `PushRequest` containing repository identity and one source/destination ref pair.  Unknown, missing, equivalent-duplicate, oversized, malformed, force/tag/delete/multiple-ref, noncanonical fields, and unknown operations deny before backend dispatch; the admitted operation set is exactly `ready`, `otp`, and `push`. | **MATCH / PASS.** The Revision 5 missing protocol owner and request route are explicit, strict, and testable. |
| Backend reachability and authorization gates | TASK-0011 owns `internal/backend/push_registration.go` and its test through TASK-0007's already-merged bounded registration seam.  Focused evidence must prove the authorized caller reaches exactly one handler.  Existing fail-closed SO_PEERCRED admission verifies the root-configured dedicated caller UID before dispatch; registration additionally requires a live lease and TASK-0010 local-policy PASS before token retrieval or Git.  Wrong UID, missing/expired lease, invalid policy/schema, unavailable registration, or unknown operation denies before custody/Git as applicable. | **MATCH / PASS.** The registered handler is reachable only through the named caller and all three required UID, live-lease, and policy gates; no generic transport escape hatch is introduced. |
| TASK-0011 ownership and +220 basis | Exclusive production ownership is the caller, IPC protocol, custody, system-Git, and backend-registration paths, with corresponding focused tests.  The production forecast decomposes to 35 caller + 45 protocol/schema + 40 registration/gates + 55 custody + 45 system-Git/redaction = 220; tests are separate.  Expansion beyond these five units or forecast above cumulative 1350 stops for split/replan. | **MATCH / PASS.** The unchanged +220 forecast has an auditable boundary allocation rather than a throughput assumption, and growth cannot silently consume cap/reserve. |
| Forecast, caps, and reserves | Baseline 751 plus `90 + 120 + 0 + 130 + 220 + 0` remains 1311.  Forecasts `841, 961, 961, 1091, 1311, 1311` remain below 90% triggers `855, 990, 990, 1125, 1350, 1350`; target/hard guards are unchanged.  Forecast reserve remains 189 to target 1500 and 489 to unconditional hard 1800, gated by TASK-0012 actual reconciliation. | **MATCH / PASS.** Caller/protocol ownership is included inside the existing TASK-0011 allocation with no arithmetic, trigger, or later-reserve regression. |
| Six Tasks, seven TASK-0006 DEV outputs, and index consistency | The wave remains exactly TASK-0007 through TASK-0012, linearly dependent, with no seventh detailed Task.  TASK-0006 DEV remains limited to exactly `backlog.json` plus the six `TASK.md` files and no product/test/result/log/Git output.  Index/contracts must agree on the expanded TASK-0011 paths, entrypoint, gates, fixtures, caps, laps, split criteria, and lineage before execution. | **MATCH / PASS.** The Revision 6 future product-path additions change contract content, not TASK-0006's exact seven-document output boundary. |
| Two-Lap feasibility and checks | TASK-0011 Lap 1 covers dedicated UID/handler-seam preflight, local bare remote, fake token, system-Git capture, live lease, strict caller/schema/registration/gates/custody/execution, and focused regressions across both existing entrypoints plus the new caller.  Lap 2 independently mutates schema, operation, UID, lease, policy, leakage, force, and ambiguity cases before `make check`.  Its split/stop rules reject generic schema/refspec, indistinguishable UID, unsafe token injection, uncapturable Git, cap growth, or optional remote diagnostics.  Other five contracts retain their concrete path/fixture/Lap/split evidence. | **MATCH / PASS.** The complete route and denial matrix fit the stated two-Lap hypothesis without weakening gates or security boundaries. |
| Pipeline, measurement, later lineage, and controls | PLAN -> DEV -> independent REVIEW -> independent QA -> main Git remains sequential per Task; only disjoint successor planning/read-only work may overlap.  Zero-SLOC TASK-0009 and TASK-0012 retain the two-feature measurement cadence and final later-work block.  Audit/attestation/manual-canary lineage, active/wait/retry/classification provenance, mandatory controls, exact shedding, no-compression, TASK-0006 zero-product scope, and Git/Lap-log prohibitions remain unchanged. | **MATCH / PASS.** No regression in R3-04 or R3-06 through R3-10. |

### Revision 6 decision

The Revision 5 sole blocker is closed: TASK-0011 has an exclusive bounded
caller, protocol/schema owner, admitted route, backend registration owner,
and explicit pre-custody UID/live-lease/policy gates.  Existing `ready`/`otp`
client behavior and the separate privileged daemon remain protected by owned
paths and focused regressions.

Revision 6 final reconciliation decision: **PASS**.  Classification:
**`none`**.  TASK-0006 DEV may proceed only within the exact seven-document
scope and the normal approved PLAN/QA_PLAN, independent REVIEW, independent
QA, and parent/main Git gates.  Future TASK-0011 DEV remains subject to its
90% and two-Lap stop/replan rules and must prove the complete denial matrix.
This reconciliation changes only this `QA_PLAN.md`; it does not edit PLAN,
backlog, contracts, product/test source, Git state, or the operational Lap log.

## PLAN reconciliation Revision 7

Revision 7 correctly removes the nonexistent `make check` / `make task-check`
assumption.  The repository has no `Makefile`; its future focused Go commands
are package-addressable, and its required full check is executable as
`GOCACHE="$(mktemp -d)" go test ./...` when needed, plus gofmt, diff, and
applicable JSON validation.  The exact seven TASK-0006 DEV documents, six
linear contracts, Revision 6 dedicated caller/protocol/registration route,
forecast arithmetic (`751 + 90 + 120 + 0 + 130 + 220 + 0 = 1311`), 1500/1800
reserves (189/489), at-most-two-Lap gate, pipeline/role separation, and later
mandatory lineage all remain intact.

### Revision 7 reconciliation matrix

| Area | Revision 7 evidence | QA decision |
| --- | --- | --- |
| Repository-native focused and full checks | TASK-0007/0008/0010/0011 name package-scoped `go test` commands and their required fixtures; TASK-0009/0012 name snapshot regeneration and the common full Go/format/diff checks.  The explicit no-Make rule prevents the former nonexistent targets from being treated as executable. | **MATCH / PASS.** Subject to each future Task's prerequisite fixture and merged owned paths, the package/full-check route is repository-native. |
| Zero-SLOC JSONL integrity | The explicit TASK-0009/0012 command parses the frozen JSONL, rejects duplicate `event_id`s, and verifies that every correction's `corrects_event_id` exists. | **PARTIAL / FAIL.** Existence alone is not the correction validation required by this QA plan.  It does not prove that a correction targets an earlier event in the same Task and Lap, so it can accept a cross-Task, cross-Lap, or forward correction edge. |
| Zero-SLOC correction and arithmetic regeneration | Revision 7 requires independently regenerating `MEASUREMENT.md` and checking null/provenance, SLOC/test/time/active/wait/retry/classification, contingency, and cap arithmetic. | **PARTIAL / FAIL through the incomplete correction predicate.** The described regeneration is necessary, but a future reviewer/QA command set cannot establish canonical effective classifications until the required correction-edge predicate is made executable. |
| Revision 6 dedicated push route | TASK-0011 retains the dedicated restricted caller, `OperationPush`/bounded request admission, separate backend registration through the TASK-0007 seam, and UID/live-lease/TASK-0010-policy checks before custody or Git; the old client remains `ready`/`otp` only. | **MATCH / PASS.** No reachability or ownership regression. |
| Scope, caps, cadence, and reserve lineage | Seven-document TASK-0006 scope, six contracts, cap/trigger arithmetic, two zero-SLOC gates, no more than two feature Tasks between gates, and post-TASK-0012 audit/attestation/manual-canary lineage remain explicit. | **MATCH / PASS.** No scope, arithmetic, pipeline, or later-reserve regression. |

### Required Revision 7 remediation

For both TASK-0009 and TASK-0012, replace or extend the explicit JSONL
correction check with a repository-native executable predicate that verifies
every correction has a nonempty target event ID, the target occurs earlier in
the frozen snapshot, and target and correction have the same `task_id` and
`lap`.  The predicate must remain additive to JSONL parsing, unique-ID,
measurement regeneration, null/provenance/effective-classification checks,
contingency/cap arithmetic, and the required full Go/format/diff suite.  This
is a planning-contract correction only; it requires no script, source change,
gate weakening, or new DEV output.

Revision 7 final reconciliation decision: **FAIL — planning revision
required**.  Classification: **`planning_defect`**.  The missing executable
same-Task/lap/precedence correction validation leaves both zero-SLOC measurement
gates unable to prove a TASK-first P0 invariant.  This is not attributed to
DEV or the current environment.  All other Revision 6 regression checks listed
above pass.  DEV remains blocked pending an approved PLAN/QA_PLAN correction.
This reconciliation changes only this `QA_PLAN.md`; it does not edit PLAN,
backlog, contracts, product/test source, Git state, or the operational Lap log.

## PLAN reconciliation Revision 8

Revision 8 was independently reconciled against the sole Revision 7
`planning_defect` and all retained Revision 6/7 invariants.  Its repository-
native zero-SLOC predicate was executed against the current canonical JSONL
stream: all 136 rows parsed, event IDs were unique, and all 10 correction edges
passed the required earlier-file-order, same-`task_id`, same-`lap_id`, and
smaller-within-lap-`sequence` checks.

### Revision 8 reconciliation matrix

| Area | Revision 8 evidence | QA decision |
| --- | --- | --- |
| Revision 7 correction-edge blocker | The `jq -s` predicate retains zero-based file order with `to_entries` and requires each correction target to precede the correction in file order, match `task_id` and `lap_id`, and have a smaller `sequence`.  It ran successfully on the canonical stream. | **MATCH / PASS.** Cross-Task, cross-lap, forward-file, and non-earlier-sequence correction edges now fail executable validation.  The sole Revision 7 blocker is closed. |
| Raw and effective correction provenance | Target and correction records remain raw with source event IDs.  `superseded_by` and effective omission are derived only after the edge passes; invalid edges stop measurement rather than rewriting history. | **MATCH / PASS.** TASK-first correction and provenance semantics are preserved. |
| Zero-SLOC focused/full checks | TASK-0009 and TASK-0012 retain JSONL parse, unique-ID, corrected-edge validation, independent `MEASUREMENT.md` regeneration, null/correction/provenance and SLOC/test/time/active/wait/retry/classification/cap arithmetic checks, plus the repository-native Go/format/diff full suite. | **MATCH / PASS.** The correction predicate is additive and requires no script or weakened REVIEW/QA gate. |
| Revision 6 dedicated push route | TASK-0011 retains its dedicated caller, bounded `OperationPush` schema/admission, backend registration through the TASK-0007 seam, and SO_PEERCRED UID/live-lease/TASK-0010-policy gating before custody/Git; the established client remains `ready`/`otp` only. | **MATCH / PASS.** No reachability, admission, authorization, or ownership regression. |
| Scope, arithmetic, cadence, and lineage | TASK-0006 remains limited to exactly seven documents and zero production SLOC; the wave remains six linear contracts with forecasts totaling 1311, reserves of 189 to target 1500 and 489 to hard 1800, no more than two production Tasks between measurement gates, at most two Laps per Task, separate PLAN/DEV/REVIEW/QA roles, main-owned Git, and later audit/attestation/manual-canary lineage gated by TASK-0012 PASS+merge. | **MATCH / PASS.** No Revision 6/7 scope, cap, pipeline, measurement, role, no-compression, shedding, or later-lineage regression. |

The previously classified Unix-socket restriction remains reusable
`environment_issue` evidence for this sandbox; the unchanged full Go suite was
not purposelessly repeated during PLAN reconciliation.  It is not DEV-fault
evidence and does not alter the future REVIEW/QA obligation to run the full
suite in a capable environment.

Revision 8 final reconciliation decision: **PASS**.  Classification:
**`none`**.  The Revision 7 `planning_defect` is resolved, and all retained
Revision 6/7 invariants pass.  TASK-0006 DEV may proceed only within the exact
seven-document scope and the normal approved PLAN/QA_PLAN, independent REVIEW,
independent QA, and parent/main Git gates.  This reconciliation changes only
this `QA_PLAN.md`; it does not edit PLAN, backlog, contracts, product/test
source, Git state, the operational repository, or the Lap log.

## PLAN reconciliation Revision 8.1 — QA evidence correction

Independent REVIEW correctly found that Revision 8's phrase "all 136 rows"
could be read as a current canonical-stream invariant even though the
operational JSONL is append-only.  Classification: **`qa_plan_defect`**.  This
finding concerns QA evidence attribution, not PLAN, DEV, product behavior, or
the validity of the Revision 8 correction predicate.

The predicate evidence actually checked by Revision 8 is the frozen prefix
formed by `head -n 136 /home/ubuntu/git/agent-harness-work/lap30/events.jsonl`.
Its reproducible identity at this reconciliation is:

- row count: **136** (prefix length only);
- final prefix `event_id`: `task0006-r3-lap02-replan8-start`;
- SHA-256 over the exact 136 newline-terminated JSONL records:
  `9b01bb157c666aba691a87665bbdef5fb06ae6ba7359198e2e0fda7ceb3dbfc1`.

That exact prefix independently passed JSON parsing, unique-event-ID, and all
10 same-Task/lap, earlier-file-order, smaller-sequence correction-edge checks.
The live append-only log was separately observed at 144 rows and then 146 rows
during Revision 8.1; the full 146-row observation also passed the correction
predicate.  These live counts are observations only, not acceptance values.

TASK-0009, TASK-0012, and every current or future REVIEW/QA gate must freeze
and identify the complete snapshot available to that gate, then parse and
validate that entire frozen snapshot.  They must not truncate to, compare
against, or require 136 rows.  A later snapshot may use a different row count,
last event ID, and digest without contradicting Revision 8 evidence.

Revision 8.1 final reconciliation decision: **PASS**.  Classification:
**`none`** after correction; the REVIEW `qa_plan_defect` is resolved by the
precise prefix provenance and non-invariant clarification above.  Revision 8's
predicate closure and all retained Revision 6/7 scope, arithmetic, push-route,
pipeline, role, measurement-cadence, no-compression, shedding, and later-
lineage invariants remain unchanged.  This amendment changes only
`QA_PLAN.md`; it does not edit PLAN, candidate documents, REVIEW_RESULT,
product source, Git state, the operational repository, or the Lap log.
