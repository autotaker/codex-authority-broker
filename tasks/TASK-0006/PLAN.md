# PLAN — TASK-0006: Measurement and rolling-wave replanning gate

## Decision and execution boundary

TASK-0006 records the four completed delivery cycles from the canonical Lap30
event stream, converts only the next **two** reserves (`sudo`, then `push`) into
detailed executable Task contracts, and inserts a new zero-production-SLOC
measurement gate immediately after them.  `audit-release` and `clean-canary`
remain non-executable reserves.

This task adds **0 authority-broker production SLOC**.  It may change only:

| Path | Responsibility |
| --- | --- |
| `tasks/TASK-0006/MEASUREMENT.md` | Reproducible 72-event/four-Task measurement table, provenance, nulls, classifications, and arithmetic. |
| `tasks/TASK-0007/TASK.md` | Detailed per-sudo live-check/no-cache contract. |
| `tasks/TASK-0008/TASK.md` | Detailed restricted non-force push/token-custody contract. |
| `tasks/TASK-0009/TASK.md` | Next measurement/replanning gate after TASK-0007 and TASK-0008. |
| `backlog.json` | Mark only those three Tasks executable, retain later reserves, update dependencies/caps and measurement lineage. |

No `.go`, test, command, PAM, sudo, push, audit, package, release, canary, or
other product behavior is implemented here.  Detailed contracts beyond the
next two feature Tasks and their measurement gate are prohibited.

**DEV profile:** `luna-xhigh` (`dev-luna`).  Work is bounded schema/document
transcription and arithmetic, with independent REVIEW and QA; no product or
external system mutation is involved.

The 30-minute execution clock starts only after preflight confirms TASK-0005
is merged, this dedicated worktree is usable, the canonical JSONL is readable,
all 72 completed-cycle source events parse, the four completion records and
evidence files are present, and the zero-production-SLOC boundary is clean.
Preflight time is excluded; a failed prerequisite is `not_started`, not Task
execution.  DEV requires approved PLAN and independent QA_PLAN first.

## Canonical source and immutable snapshot

The source of truth is:

```text
/home/ubuntu/git/agent-harness-work/lap30/events.jsonl
```

The file currently has 74 lines because two TASK-0006 start/preflight events
were appended after the completed-cycle snapshot.  Measurement input is the
**72 events whose `task_id` is one of TASK-0001, TASK-0003, TASK-0004, or
TASK-0005**.  Never edit, copy back to, reorder, or “repair” the operations
JSONL.  Record the source path, total current line count, filtered count,
event-ID uniqueness, schema versions, selected task IDs, and the hash of the
read snapshot in `MEASUREMENT.md` so REVIEW and QA can reproduce the input.

Validate before aggregation:

- every selected line parses as one JSON object and has a unique `event_id`;
- the filtered count is exactly 72 and there is exactly one `lap_completed`
  event for each of the four measured Tasks;
- `lap_id` plus `sequence` is unique within each lap; sequence resets between
  laps and is not treated as a global clock;
- timestamps and file order are evidence, not silently normalized—late
  correction/failure events remain attached to their Task;
- all source nulls remain null.  Contradictions or missing keys are disclosed,
  not filled from intuition or overwritten by a result document.

Task REVIEW/QA results are corroborating evidence for acceptance, SLOC command
output, environment classifications, and readable explanations.  They may
clarify provenance but never replace a conflicting canonical event value.

## Measurement schema and aggregation rules

Create one row per completed Task with these required fields:

```text
task_id
planned_production_sloc
actual_production_sloc
cumulative_production_sloc
test_loc
plan_ms / plan_minutes
dev_ms / dev_minutes
review_ms / review_minutes
qa_ms / qa_minutes
ci_push_merge_ms / ci_push_merge_minutes
task_cycle_ms / task_cycle_minutes
active_ms
wait_ms
retries
raw_failure_classifications
effective_failure_classifications
source_event_ids
notes
```

Use the following deterministic rules:

1. `planned_production_sloc`, `actual_production_sloc`, and `test_loc` come
   from the Task's final `lap_completed` event.  These production values are
   incremental.  Sum actual increments in dependency order to derive and
   cross-check cumulative SLOC; retain every earlier estimate/replan in notes.
2. A stage duration is calculated only from a same-Task, same-lap, same-stage,
   same-attempt start and terminal event with numeric `elapsed_ms`:
   `finish.elapsed_ms - start.elapsed_ms`.  A `lap_stopped` may close the
   matching active stage.  Sum all attempts, including failed/stopped attempts.
   Never treat a stage's cumulative `elapsed_ms` as the stage duration by
   itself, and never infer overlapping PLAN/QA_PLAN work from wall timestamps.
3. `plan`, `dev`, `review`, and `qa` use their named stages.  The `git` stage is
   `ci_push_merge`; it includes main-owned final checks, commit, push, PR/check,
   and merge when the event says so.  “No configured remote checks” is recorded
   as observed, not as zero CI effort.
4. `task_cycle_ms` is the sum of terminal lap elapsed values for the completed
   Task, after subtracting only explicitly measured nonzero preflight duration.
   Preflight is never assigned to PLAN/DEV or used for contingency.  If its
   exclusion cannot be established, cycle time is null with an explanation.
5. Convert milliseconds to minutes by division by 60,000 and display at two
   decimals while preserving raw milliseconds.  Do not add `active_ms` and
   `wait_ms` to reconstruct elapsed time: they are independently reported
   observations and may overlap or be absent.  Aggregate them only where the
   event contract and non-overlapping attempts make the sum valid; otherwise
   use null and list the available event-level values in notes.
6. `retries` is the maximum observed task-level cumulative retry counter across
   the Task's events, avoiding double-counting checkpoint snapshots.  Also list
   retry-bearing event IDs so REVIEW can distinguish candidate, environment,
   planning, and approval retries.
7. Build `raw_failure_classifications` first from every non-null classification
   in failure, correction, stopped, stage-finished, and completion evidence.
   Each raw record contains at least `classification`, `source_event_id`, and
   `superseded_by` (initially null); do not collapse raw records to a unique
   string set.  Apply a correction only when its `annotations.corrects_event_id`
   resolves to an earlier event in the same Task, its `old_classification`
   matches that event, and its `new_classification` matches the correction
   event.  Set only the corrected raw record's `superseded_by` to the correction
   event ID.  A malformed, cross-Task, missing-target, or inconsistent
   correction is a `measurement_defect`, not a best-effort rewrite.
   `effective_failure_classifications` is the ordered unique set of
   classifications from raw records whose `superseded_by` remains null.  Thus
   raw history is immutable while superseded old values do not drive current
   replanning.  For TASK-0005, retain raw `requirement_gap` from
   `task0005-lap01-dev-stop-sloc` with
   `superseded_by=task0005-lap01-correct-dev-class`; the effective set is
   `planning_defect`, `environment_issue`.
8. A missing or non-derivable numeric value is JSON/Markdown `null`; a missing
   categorical explanation is `unknown`.  Zero is used only when the source
   explicitly proves zero.  No interpolation, equal stage split, or timestamp
   guess is allowed.

## Reproduced four-Task measurement

The DEV evidence table must reproduce the following values from the 72-event
snapshot.  Minutes shown here are derived only where the rule above permits:

| Task | Planned prod. | Actual prod. | Cumulative | Test LOC | PLAN min | DEV min | REVIEW min | QA min | Git min | Retries | Raw classifications | Effective classifications |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| TASK-0001 | 130 | 98 | 98 | 116 | null | null | null | null | null | 0 | none observed | none observed |
| TASK-0003 | 220 | 135 | 233 | 216 | null | null | null | null | null | 1 | `implementation_defect` | `implementation_defect` |
| TASK-0004 | 392 | 367 | 600 | 496 | null | 14.63 | 10.25 | 4.07 | 2.95 | 4 | `environment_issue`, `planning_defect`, `implementation_defect` | `environment_issue`, `planning_defect`, `implementation_defect` |
| TASK-0005 | 175 | 151 | 751 | 1257 | null | 10.42 | 6.53 | 5.78 | 2.97 | 1 | `requirement_gap` (superseded), `planning_defect`, `environment_issue` | `planning_defect`, `environment_issue` |

For TASK-0004 the raw derived stage totals are DEV 878,000 ms, REVIEW
615,000 ms, QA 244,000 ms, and git 177,000 ms.  For TASK-0005 they are DEV
625,000 ms, REVIEW 392,000 ms, QA 347,000 ms, and git 178,000 ms.  The earlier
Tasks lack sufficient matched stage events, so their stage fields remain null
even though whole-lap elapsed evidence exists.  Actual production increments
sum to `98 + 135 + 367 + 151 = 751`; this must equal the merged cumulative
SLOC count independently rerun at the TASK-0006 candidate.

## Evidence-based sizing and 20% contingency

The sample is only four Tasks, two lack stage splits, scopes differ sharply,
and two complex boundary Tasks needed replanning or remediation.  Incremental
production was 98, 135, 367, and 151 SLOC; this variation is evidence against
a fixed “SLOC per minute” or fixed-SLOC-per-Task throughput.  Do not calculate
or use such a rate, an average velocity, or a regression to size future work.

Where a comparable completed boundary exists, apply time contingency as:

```text
contingent_time = ceil(observed_non_preflight_time * 1.20)
```

Apply the formula independently to range endpoints and retain the source Task
and raw time.  An unknown base remains unknown; 20% does not convert null into
a number.  SLOC reserves are security/scope ceilings, not delivery forecasts,
and do not receive a 20% SLOC multiplier.

- The sudo live-check boundary is closest to TASK-0004's measured complex IPC
  cycle (62.98 minutes across its terminal laps).  Its initial execution-time
  envelope is **76 minutes minimum after preflight** (`ceil(62.98 * 1.20)`),
  with no evidence-supported upper bound until its host fixture is preflighted.
- Restricted push/token custody has no completed transport analogue.  Its time
  remains **unknown** at this gate; TASK-0008 preflight and PLAN must decompose
  the fixture before starting its clock.  Do not derive time from its 450-SLOC
  reserve or from TASK-0004 SLOC throughput.

## Next wave: only two feature Tasks plus measurement gate

DEV converts these contracts and no others:

### TASK-0007 — per-sudo live lease check and no cache

- Depends on TASK-0006; cumulative production SLOC ceiling **<=950**.
- Owns the smallest Linux sudo integration that checks live authority on every
  invocation through the existing fail-closed local boundary, with timestamp
  caching disabled for the dedicated Codex identity.
- Acceptance: live unexpired lease permits the focused sudo fixture; expired,
  restarted, unavailable, malformed, or unauthorized checks deny; consecutive
  invocations independently check live state and cannot reuse cached authority.
- Excludes push, GitHub credentials, audit, release, installer, and canary.
- Preflight must resolve the exact daemon/backend assembly and isolated host
  fixture before the clock.  If that mandatory assembly cannot fit the 199-SLOC
  reserve idiomatically, stop as a requirement gap and replan; do not weaken
  per-sudo/no-cache behavior.
- Initial time: 76-minute post-preflight minimum, upper bound unknown, with the
  stated TASK-0004 analogue and 20% arithmetic recorded.

### TASK-0008 — restricted non-force push and token custody

- Depends on TASK-0007; cumulative production SLOC ceiling **<=1400**.
- Owns exact local repository/ref/clean-tree validation, GitHub App token
  custody, and a single system-Git non-force push path.
- Acceptance: only the configured repository and `main`/`task/TASK-*` branch
  single-ref non-force update may proceed with a live lease; wrong repo/ref,
  dirty tree, force/tag/delete/multiple ref, expired authority, token leakage,
  or Git/transport ambiguity denies without force retry.  Tokens remain absent
  from argv, environment, logs, output, errors, and credential-helper storage.
- Excludes sudo changes, rich audit, release/attestation, installer, and canary.
- Preflight must establish a local bare-repository/fake-token fixture and
  credential injection design.  Delivery time is unknown until that bounded
  design is approved; it must not be inferred from SLOC throughput.

### TASK-0009 — next measurement and replanning gate

- Depends on TASK-0008; adds **0 production SLOC**, cumulative ceiling remains
  **<=1400** before any next-wave conversion.
- Repeats this measurement contract for TASK-0007/0008 plus retained history,
  classifies failures, applies 20% time contingency, and converts only the next
  2–3 evidence-supported reserves.
- `MILESTONE-audit-release` (reserve <=1500) and
  `MILESTONE-clean-canary` (adds 0) remain non-executable until TASK-0009 PASS
  and merge.  No branch/PLAN/DEV may begin for them earlier.

## Global cap, shedding, and mandatory controls

The global v1 production ceiling remains **<=1500**.  TASK-0006 adds zero;
the next-wave cumulative ceilings are 950 and 1400, leaving the existing 1500
reserve for later audit/release planning.  At projected use above 90% of a
Task's cumulative ceiling, stop and re-estimate before DEV.  An over-cap
candidate stops and follows this exact ordered shedding sequence, marking a
step not-present rather than silently skipping it:

1. automated canary executable; retain manual runbook/evidence
2. rich status/JSON UX; retain activate and immediate revoke
3. rich audit schema/correlation; retain minimal external-trace correlation
   ID, actor, scope, result, expiry
4. precomputed pack-size/history diagnostics; retain exact repo/ref/clean tree
   and normal non-force rejection
5. explicit remote-OID prefetch/race diagnostics; rely on standard non-force
   Git rejection and generic failure
6. automated installer/rollback executable; retain declarative units and
   manual verified install/rollback
7. GitHub push moves to v2, leaving TOTP full-sudo authority as v1

After any applicable shedding, PLAN and QA must be rerun before DEV continues.
Never shed readiness; TOTP replay/rate/absolute lease controls; SO_PEERCRED
fail-closed IPC; per-sudo live check/no cache; OTP/secret non-disclosure from
argv/log; source-free attested artifact; or minimal external-trace-compatible
audit.  If mandatory v1 still exceeds 1500, stop as `requirement_gap`; never
compress code, errors, names, comments, or tests to appear compliant.

## DEV, REVIEW, and QA evidence

DEV performs only the allowed planning/backlog edits, runs JSON parsing/schema
checks, regenerates the measurement table from the read-only snapshot, verifies
all arithmetic and dependencies, and runs `git diff --check`.  It also reruns
the exact production-SLOC command used by prior Tasks and proves the cumulative
total remains 751 and TASK-0006 adds zero.  No Makefile/script is added solely
for this gate.

`MEASUREMENT.md` must expose both `raw_failure_classifications` and
`effective_failure_classifications`.  Raw entries list their source event ID
and nullable `superseded_by`; effective entries are generated only after the
deterministic correction validation in rule 7.  DEV records the TASK-0005
correction edge explicitly as
`task0005-lap01-dev-stop-sloc -> task0005-lap01-correct-dev-class`, proves the
raw `requirement_gap` remains visible, and proves it is absent from the
effective set.  No manual deletion or free-text reclassification is allowed.

Independent REVIEW must record:

- 72 selected/valid/unique events and four unique completed Tasks;
- every required metric, its event provenance, null/unknown preservation,
  stage-pair arithmetic, retry aggregation, raw/effective classification
  columns, deterministic correction edges, and preflight exclusion;
- `98 + 135 + 367 + 151 = 751`, test LOC values, 20% calculation, and the
  absence of fixed SLOC throughput assumptions;
- only TASK-0007, TASK-0008, and TASK-0009 are executable/detailed; dependencies,
  ceilings, exclusions, retained reserves, and next gate are consistent across
  documents and `backlog.json`;
- exact shedding order, mandatory controls, global <=1500 constraint, zero
  added production SLOC, scope, and `git diff --check` PASS.

QA independently repeats those checks after REVIEW PASS and records PASS/FAIL.
Any discrepancy is classified as `measurement_defect`, `planning_defect`,
`qa_plan_defect`, `requirement_gap`, or `environment_issue` before attribution.
Missing timing remains null and is not itself a failure when provenance proves
it is unavailable.  A REVIEW or QA FAIL returns to its responsible gate.  This
PLAN authorizes no stage, commit, merge, product work, or operations-log write.
