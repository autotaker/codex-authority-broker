# TASK-0009 QA plan — independent measurement gate

## QA authority and frozen baseline

This is the TASK-first QA plan for the zero-production-SLOC measurement gate.
It was derived from `tasks/TASK-0009/TASK.md`, `backlog.json`, and the
read-only canonical source
`/home/ubuntu/git/agent-harness-work/lap30/events.jsonl`; it deliberately does
not depend on, or compare itself with, `PLAN.md` before the baseline is fixed.
QA owns only independent validation and the resulting `QA_RESULT.md`. It must
not edit the canonical log, implement product or test code, stage, commit,
merge, or make a push decision.

The frozen-baseline report is:

| Item | Fixed QA fact |
| --- | --- |
| Canonical snapshot | 389 JSONL records; every line parses and `event_id` is unique. |
| Historical performance cohort | Exactly TASK-0001, TASK-0003, TASK-0004, TASK-0005; one historical evidence row per member. It is not the product-terminal population. |
| Product terminals | Only `task0015-counted-lap02-complete`, `task0016-counted-lap02-complete`, and `task0008-counted-lap01-complete`. |
| Excluded records | TASK-0006 governance cycles; any `annotations.counts_as_product_lap == "false"` partial; TASK-0007 terminated and TASK-0014 unfinished evidence. Corrections are history, never extra tasks. |
| TASK-0013 | Its merged SLOC lineage remains included, but completion timing is **null** with reason `no canonical TASK-0013 terminal`: its two `lap_completed` records have status `pass`, not `completed`. |
| Measured cumulative SLOC | **1253**: canonical TASK-0015 terminal says 1200, TASK-0016 terminal says 1215, and TASK-0008 terminal says 1253. TASK-0009 contributes 0. |

`backlog.json` currently carries a 1270 expected cumulative value for TASK-0009
while the updated Task contract and canonical terminal evidence specify 1253.
QA must report that discrepancy as a requirement/process-evidence conflict; it
must neither choose 1270 nor silently rewrite the baseline. The acceptance
calculation for this gate is the reproducible canonical 1253 value.

## Preconditions and independent checks

QA starts only after independent REVIEW PASS and after the candidate
`MEASUREMENT.md` identifies its frozen source and source IDs. QA independently
runs, from the candidate worktree:

```sh
EVENTS=/home/ubuntu/git/agent-harness-work/lap30/events.jsonl
jq -e . "$EVENTS" >/dev/null
test -z "$(jq -r '.event_id' "$EVENTS" | sort | uniq -d)"
jq -s -e '
  to_entries as $rows
  | all($rows[];
      if .value.event == "correction" then
        . as $correction
        | (.value.annotations.corrects_event_id // "") as $target
        | ($target != "") and any($rows[];
            .key < $correction.key
            and .value.event_id == $target
            and .value.task_id == $correction.value.task_id
            and .value.lap_id == $correction.value.lap_id
            and .value.sequence < $correction.value.sequence)
      else true end)
' "$EVENTS" >/dev/null
jq -s -e '
  [ .[] | select(.task_id == "TASK-0001" or .task_id == "TASK-0003" or
                  .task_id == "TASK-0004" or .task_id == "TASK-0005")
        | select(.event == "lap_completed" and .status == "completed")
        | .task_id ] | unique == ["TASK-0001","TASK-0003","TASK-0004","TASK-0005"]
' "$EVENTS" >/dev/null
```

Every correction target must be nonempty, earlier in frozen file order, and
match task, lap, and a smaller sequence. Only after this predicate passes may
QA accept raw rows annotated with `superseded_by` and effective rows that omit
the superseded value. For example, the raw TASK-0005 `requirement_gap` stays
visible with `superseded_by: task0005-lap01-correct-dev-class`; its correction
provenances effective `planning_defect`. An unlabeled blend of raw and
effective classifications fails.

## Acceptance matrix

| ID | Independent QA procedure | PASS evidence | FAIL / classification |
| --- | --- | --- | --- |
| Q9-01 | Parse frozen JSONL, check unique IDs, and run the correction-edge predicate. | All commands pass; each raw correction target is retained and valid before effective derivation. | Missing, duplicate, forward, cross-task/lap, or non-monotonic target: `requirement_or_canonical_evidence_defect`; stop. |
| Q9-02 | Regenerate cohort and terminal selection from source IDs, not a blanket `lap_completed` filter. | Four historical rows (0001/0003/0004/0005), separate terminal rows only for 0015/0016/0008; no governance, partial, terminated, unfinished, or correction row is a terminal. | Any population mixing: `measurement_implementation_defect`; FAIL. |
| Q9-03 | Reproduce lineage and SLOC arithmetic from the terminal annotations. | `1200 -> 1215 -> 1253`, TASK-0009 increment 0, and canonical cumulative 1253. | Non-reproducible sum, nonzero TASK-0009 SLOC, or result over 1350: `measurement_implementation_defect`; stop. |
| Q9-04 | Inspect all TASK-0013 timing claims and terminal predicate. | SLOC lineage is present; completion/cycle terminal timing is null with the exact missing-terminal reason, never zero or imputed. | Invented terminal or reasonless null: `measurement_implementation_defect`. |
| Q9-05 | Rebuild stage timing only from same task/lap/stage/attempt start-terminal pairs; exclude preflight. | Each timing is paired or null with reason; preflight failures are `not_started`, excluded from stage/cycle totals, and snapshots are not double-counted. | Cross-attempt/stage pairing, synthetic zero, or preflight in elapsed totals: `measurement_implementation_defect`. |
| Q9-06 | Recalculate contingency field-by-field. | Each observable non-preflight time is `ceil(observed_ms * 1.20)`, units/rounding stated; null remains null. No SLOC, test LOC, active/wait, or retry count is multiplied. | Multiplier on unobservable time or non-time metric: `measurement_implementation_defect`. |
| Q9-07 | Trace `active_ms`, `wait_ms`, and retries to source events. | Active and wait remain separate, never summed to invent elapsed time; retries are the maximum propagated task-level value across a complete cycle, with source IDs and null reasons. | Double count, fabricated elapsed, or retry reset/loss: `measurement_implementation_defect`. |
| Q9-08 | Inspect `MEASUREMENT.md`, Task/backlog metadata, and diff scope. | Provenance covers SLOC, test LOC, timing/null, stage pair, active/wait, retry, raw/effective classification, source ID, and units. Only process-evidence outputs change; canonical JSONL is untouched. | Missing provenance: `measurement_implementation_defect`; canonical/product/test change: `scope_violation`. Backlog 1270 vs Task/canonical 1253: `requirement_or_process_evidence_conflict`, not presumptive DEV fault. |
| Q9-09 | Verify downstream disposition is explicit. | TASK-0010–0012 arithmetic is invalid pending TASK-0009 PASS+merge **and explicit replan**. The record explicitly selects or defers push-to-v2; silence is not selection. Later reserve remains non-executable. | Implicit enablement, silent push-v2 choice, or downstream execution detail: `requirement_or_process_evidence_defect`; stop. |

## Repository and process evidence

QA independently runs the required repository checks once after measurement
regeneration:

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

A toolchain/fixture failure is recorded with its exact command and classified as
`environment_issue` unless candidate-specific reproduction demonstrates an
implementation cause. A measurement mismatch is not presumed DEV fault:
`QA_RESULT.md` must classify it as `measurement_implementation_defect`,
`requirement_or_process_evidence_conflict`, `environment_issue`, or
`regression`, preserving minimal evidence for the responsible gate.

## Gate decision

PASS requires Q9-01 through Q9-09, reproducible 1253 arithmetic, full
repository evidence, REVIEW PASS, and a written explicit downstream replan and
push-to-v2 disposition. Missing or contradictory canonical evidence,
unreproducible arithmetic, actual cumulative above 1350, inability to
independently regenerate, or absent explicit downstream decision is a stop:
classify before retry. No retry can alter the frozen canonical log or bypass
later replanning. This QA plan now waits for the baseline-fixed PLAN comparison
before it may be reconciled or approved.

## Baseline-fixed PLAN reconciliation

Reconciled against the baseline-fixed `PLAN.md` without changing the frozen
QA baseline. The earlier statement that `backlog.json` held TASK-0009 at 1270
is retained above as an observation made before Main's metadata correction.
Current evidence resolves it: `backlog.json` wave baseline/top wave and
TASK-0009 `expected_cumulative_production_sloc` are both 1253, matching the
TASK-0009 contract and canonical terminal chain. **Metadata equality: PASS.**
TASK-0008's historical merged-contract metadata remains 1270; it is a prior
forecast/contract value, not the TASK-0009 current-baseline field, and is not
rewritten or treated as a terminal selection input.

The frozen source named by PLAN exists, is byte-identical to the canonical
source, has 389 lines, and hashes to
`abe31c9da1fbcfd32daef6013a9ce58063ccbb086265b4760ca20a22962af09b`.
PLAN and this QA plan agree on the exact historical/product cohort separation,
TASK-0013 null-terminal timing, partial/governance/terminated/unfinished
exclusions, correction raw/effective rule, paired timing/null/active/wait/
retry fields, and time-only 20% contingency. They also agree that the process
candidate may touch exactly the ten listed paths, must produce the stated
measurement fields, and requires an explicit TASK-0010–0012 replan plus an
explicit push-to-v2 decision or deferral.

**DEV readiness: PASS, conditional on Main approval of this reconciled
TASK-first QA plan.** The authorized DEV work remains process-only and zero
production SLOC; any candidate expanding beyond the ten paths, mismatching the
hash, changing canonical evidence, or omitting the downstream decision fails
the relevant Q9 gate and stops rather than beginning product work.
