# PLAN — TASK-0009: first zero-SLOC measurement gate

## Decision, ownership, and gate

This is PLAN evidence for the first zero-production-SLOC measurement gate. It
authorizes neither DEV nor Git, canonical-log, operational-log, product, test,
approval, merge, or push work. The only file this Planner has changed is this
`tasks/TASK-0009/PLAN.md`; Main owns any later process-only candidate, shared
lock, staging, publication, and merge.

The observed handoff is `candidate_handoff_environment`, retry **1**.
It has no authoritative stage-start/end boundary: `active_ms=null` and
`wait_ms=null`, reason `handoff runtime did not expose authoritative
active/wait boundaries`. Neither value is zero or included in a time total.

DEV remains closed until Main approves this PLAN and a reconciled independent,
TASK-first QA_PLAN, then independent REVIEW and QA pass the exact process-only
candidate. Expected product SLOC is **0**; current baseline/top wave is
**1,253**. Target cap is 1,350, re-estimate trigger 1,325, hard guard 1,450.
A contradiction, non-reproducible calculation, missing independent
regeneration, or actual cumulative above 1,350 stops, classifies before retry,
and cannot be hidden as wait time.

## Frozen source and correction validation

The sole input is read-only `/tmp/task0009-events-frozen.jsonl`, SHA-256
`abe31c9da1fbcfd32daef6013a9ce58063ccbb086265b4760ca20a22962af09b`.
Before every derivation, `MEASUREMENT.md` records path, hash, 389-line count,
command, and UTC time. Hash/line-count mismatch is
`requirement_or_canonical_evidence_defect`; stop before retry. Do not read,
write, or reconcile against a mutable operational log.

```sh
EVENTS=/tmp/task0009-events-frozen.jsonl
sha256sum "$EVENTS"
test "$(wc -l < "$EVENTS")" -eq 389
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
```

Raw reporting retains source event ID/order/original field and nullable
`superseded_by`. Only a correction passing that predicate can supersede the
named raw field; effective tables contain only validated unsuperseded values.
A correction is never an extra task or terminal. TASK-0005 retains raw
`requirement_gap` with
`superseded_by: task0005-lap01-correct-dev-class`; its effective value is
`planning_defect`. An unlabeled raw/effective mixture fails.

## Cohorts, terminals, and lineage

Historical performance is exactly TASK-0001, TASK-0003, TASK-0004, and
TASK-0005: exactly one historical completed row per task. It stays separate
from product terminals and cannot yield LOC velocity, averages, or forecasts.

| Source event | Task | Cumulative product SLOC | Timing treatment |
| --- | --- | ---: | --- |
| `task0015-counted-lap02-complete` | TASK-0015 | 1,200 | Observed; exceptional predeclared Git-residue Lap. |
| `task0016-counted-lap02-complete` | TASK-0016 | 1,215 | Observed; exceptional predeclared Git-residue Lap. |
| `task0008-counted-lap01-complete` | TASK-0008 | **1,253** | Latest terminal and effective baseline. |

Snapshots are not additive: `1200 -> 1215 -> 1253`. TASK-0008 adds 38 and
TASK-0009 adds zero. Exclude all
`annotations.counts_as_product_lap == "false"` partials, TASK-0006 governance
cycles, corrections, terminated TASK-0007, and unfinished TASK-0014.

TASK-0013's merged SLOC lineage (922) remains in the later baseline/product
tree. Its completion timing is `null`, reason `no canonical TASK-0013
terminal`: its `lap_completed` values have `status: pass`, not
`status: completed`. Do not invent a terminal or impute time.

## Measurement and timing rules

`tasks/TASK-0009/MEASUREMENT.md` must reproduce historical/product rows,
planned/actual product SLOC, cumulative SLOC, test LOC, source IDs,
raw/effective classification, stage/cycle values, active/wait, retries, units,
rounding, and null reasons. It states baseline arithmetic `1253 + 0 = 1253`.

Timing is valid only for a same-task/lap/stage/attempt
`stage_started`-terminal pair. Preflight is excluded from totals; failed
preflight is `not_started`, not zero. Do not pair snapshots, cross attempts,
or double count. Keep `active_ms` and `wait_ms` independent. Absent source
or authoritative pair means null plus reason: never derive active from elapsed
minus wait or sum active/wait into elapsed. Retries are the maximum propagated
task-level counter across a complete cycle, sourced by event ID.

For each observable non-preflight time only, record
`ceil(observed_ms * 1.20)` with unit/rounding. Null remains null. The 20%
contingency is time-only: it cannot multiply/create SLOC, test LOC,
active/wait, retries, paths, scope, or authority.

## One-Lap standard

One counted Lap: minute 0 preflight; minute 5 frozen-source/selection proof;
minute 10 correction and paired-stage proof; minute 20 complete candidate;
minute 25 independent REVIEW start; minute 30 independent QA/Git-ready stop.
At each boundary Main records UTC boundary where available, active/wait/null
reason, retries including zero, classification, source IDs, and completeness.
Planning/pure wait consumes at most 20% of observable counted time.

There is no ordinary Lap 2. Residue is allowed only for an already-complete
candidate with one/two concrete REVIEW/QA findings, no redesign/research/
fixture/Task change, bounded correction within its first 20 minutes, and no
new product/test scope. Otherwise stop, classify, and return to approved
PLAN/QA_PLAN or split. Roles remain separate; children never stage, commit,
merge, or write `.git`.

## Exact process-only candidate

After PLAN/QA approval, Main may create exactly these ten paths:

1. `backlog.json`
2. `tasks/TASK-0009/TASK.md`
3. `tasks/TASK-0009/PLAN.md`
4. `tasks/TASK-0009/QA_PLAN.md`
5. `tasks/TASK-0009/MEASUREMENT.md`
6. `tasks/TASK-0009/REVIEW_RESULT.md`
7. `tasks/TASK-0009/QA_RESULT.md`
8. `tasks/TASK-0010/TASK.md`
9. `tasks/TASK-0011/TASK.md`
10. `tasks/TASK-0012/TASK.md`

No product, implementation-test, canonical-log, audit, release, installer,
canary, operational-log, or `.git` path is eligible. `backlog.json` and
TASK-0009 agree on cumulative/top wave 1,253. QA_PLAN's claim of backlog
1,270 is stale process evidence; reconcile it before approval as
`requirement_or_process_evidence_conflict`, never silently choose a value or
blame DEV.

## Downstream replan and push-to-v2

TASK-0010–0012 arithmetic is invalid until this gate independently passes,
merges, and records an explicit replan. Each downstream TASK must then state
baseline 1,253, dependency terminal(s), expected increment, cumulative
target/trigger/hard guards, exact future paths, one-Lap checkpoints, time-only
contingency, recording rules, and split stops. No old arithmetic enables DEV.

Only after recomputation may Main record TASK-0009's explicit `not pushed`,
`split`, or `push-to-v2` decision. A push identifies deferred work,
preserved security/acceptance boundary, dependency effect, target/hard
arithmetic, and explicit approval. Silence is not selection. Do not use
contingency or reserve as SLOC capacity. Above 1,325 re-estimates; above 1,350
stops; 1,450 is never borrowed. Project target 1,500 may be exceeded only by a
later explicit replan.

## Independent validation and state

After QA_PLAN reconciliation, independent REVIEW and QA each rerun frozen
parse/unique-ID/correction, cohort/terminal, SLOC/cap, paired-stage,
active/wait/retry, null, and contingency calculations. They verify exact
ten-path scope, zero product SLOC, and downstream disposition, then run once:

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

Tool/fixture absence records exact output as `environment_issue` unless
candidate-specific reproduction proves otherwise. Measurement mismatch is
classified as measurement implementation, requirement/process-evidence,
environment, or regression before retry. This is draft PLAN evidence only:
approval, QA_PLAN reconciliation, DEV, Git, and merge are pending.
