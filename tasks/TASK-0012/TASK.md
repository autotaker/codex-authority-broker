# TASK-0012: final zero-SLOC measurement and later-reserve gate

**Depends on:** TASK-0011 (merged and PASS).

**Status:** planned and executable.

## Contract metadata

```json
{
  "id": "TASK-0012",
  "title": "final zero-SLOC measurement and later-reserve gate",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0011"],
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1311,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1350,
  "hard_cumulative_guard": 1650,
  "production_paths": [],
  "test_paths": ["tasks/TASK-0012/MEASUREMENT.md"],
  "entrypoint": null,
  "fixture_elevation_needs": "Read-only frozen completed TASK-0010/TASK-0011 event snapshot; no elevation, network, product fixture, or operational-log write.",
  "lap_1": "After TASK-0011 merge, freeze the wave snapshot and regenerate full historical plus wave SLOC/test/time/active/wait/retry/classification evidence, reconcile actual capacity to target 1500 and hard 1800, and record eligibility evidence for later audit/attestation/manual-canary lineage.",
  "lap_2": "Independent REVIEW and QA each regenerate evidence, run canonical correction and repository-native full Go/format/diff checks, validate actual target/hard reserve, and prove no later DEV is enabled; main owns Git.",
  "exclusions": ["all product/test implementation", "audit", "attestation", "release", "installer", "canary", "editing canonical log", "enabling later DEV"],
  "split_stop_rule": "Stop on canonical defects, unexplained cap drift, actual above 1500 without approved contingency disposition, any hard-limit risk, or inability to close REVIEW/QA in Lap 2; later reserve PLAN/DEV remains blocked until this gate passes and merges.",
  "measurement_lineage": "Apply the exact TASK-0009 correction predicate and raw/effective provenance to the frozen stream; compare measured baseline and wave delta, preserve null reasons, and reconcile actual 1500/1800 capacities without fixed throughput.",
  "later_reserve_eligibility": "Only after independent REVIEW PASS, independent QA PASS, and main merge may MILESTONE-audit-attestation then MILESTONE-manual-canary-rollback receive PLAN; both remain non-executable beforehand.",
  "contract_path": "tasks/TASK-0012/TASK.md"
}
```

## Purpose and evidence boundary

This is the final zero-production-SLOC measurement and later-reserve gate for
the six-task wave. It measures the immutable historical baseline plus
TASK-0010/TASK-0011, reconciles actual capacity against the 1500 planning
target and unconditional 1800 hard limit, and records eligibility evidence for
the still-non-executable audit/attestation/manual-canary lineage. It does not
implement or enable later product behavior.

The source is the read-only canonical JSONL snapshot. Completed tasks are
identified only by `lap_completed` with `status == "completed"`; correction
events are not tasks. Preserve all raw source event IDs, classifications,
`superseded_by`, null reasons, and effective correction-validated values.

## Preflight and two-Lap delivery

Preflight requires merged TASK-0011 and a readable frozen completed-wave
snapshot. A preflight failure is `not_started`, excluded from stage/cycle
timing, and never imputed as zero.

Lap 1 regenerates full historical and wave SLOC/test/time/active/wait/retry/
classification evidence, with same-task/lap/stage/attempt start-terminal
pairs, separate active/wait observations, maximum propagated retries, and
`ceil(observed_non_preflight_time * 1.20)` only for observable time. It
reconciles actual reserve and records that later DEV remains blocked.

Apply this deterministic correction check to the frozen stream before deriving
effective values:

```sh
EVENTS=/home/ubuntu/git/agent-harness-work/lap30/events.jsonl
jq -e . "$EVENTS" >/dev/null
test -z "$(jq -r '.event_id' "$EVENTS" | sort | uniq -d)"
jq -s -e '
  to_entries as $rows
  | all($rows[];
      if .value.event == "correction" then
        . as $correction
        | any($rows[];
            .key < $correction.key
            and .value.event_id == $correction.value.annotations.corrects_event_id
            and .value.task_id == $correction.value.task_id
            and .value.lap_id == $correction.value.lap_id
            and .value.sequence < $correction.value.sequence)
      else true end)
' "$EVENTS" >/dev/null
jq -s -e 'map(select(.task_id == "TASK-0001" or .task_id == "TASK-0003" or .task_id == "TASK-0004" or .task_id == "TASK-0005") | select(.event == "lap_completed" and .status == "completed") | .task_id) | unique == ["TASK-0001", "TASK-0003", "TASK-0004", "TASK-0005"]' "$EVENTS" >/dev/null
```

Every correction must have a nonempty target, earlier file position, matching
task and lap, and smaller sequence. Invalid edges stop measurement; they never
rewrite or discard raw history. Only a passing edge may establish
`superseded_by` and omit a superseded effective value.

Lap 2 has independent REVIEW and QA each regenerate all evidence, validate the
actual 1500/1800 reserve, prove no later DEV is enabled, and run:

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

Main owns final checks and Git. No operational log, product source, or result
file is written by TASK-0006 DEV.

## Acceptance, exclusions, and capacity

- Measured baseline remains exactly 751 and TASK-0006 adds exactly 0
  production SLOC; wave forecast is cumulative 1311 unless independent
  evidence changes it.
- Actual target reserve and hard reserve are independently recomputed; the
  forecast amounts 189 to 1500 and 489 to 1800 are not implemented SLOC.
- The later owner lineage is exactly
  `TASK-0012 measurement PASS+merge -> MILESTONE-audit-attestation ->
  MILESTONE-manual-canary-rollback`.
- No later milestone receives PLAN, branch, DEV, or PR-ready detail before
  this gate passes and merges.

This Task excludes all product/test implementation, audit, attestation,
release, installer, canary, canonical-log editing, and enabling later DEV.

The planned increment is +0 production SLOC and cumulative 1311; target cap
1500, 90%-trigger 1350, hard guard 1650. Stop on canonical defects,
unexplained cap drift, actual above 1500 without approved contingency
disposition, any hard-limit risk, or inability to close REVIEW/QA in Lap 2.
An actual or forecast value above the unconditional 1800 system limit is an
immediate safe stop. Preserve nulls and classify environment or planning
failures before retry; never use fixed throughput or compression.

## Gate and later reserve

Independent REVIEW PASS and QA PASS are required, followed by main-owned
merge. Only then may the audit/attestation milestone be planned; the manual
canary/rollback milestone remains ordered behind it and both remain
non-executable until their own future gates.
