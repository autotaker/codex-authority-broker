# TASK-0009: first zero-SLOC measurement gate

**Depends on:** TASK-0008 (merged).

**Status:** planned and executable.

## Contract metadata

```json
{
  "id": "TASK-0009",
  "title": "first zero-SLOC measurement gate",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0008"],
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1207,
  "target_cumulative_cap": 1300,
  "projected_cap_trigger_sloc": 1250,
  "hard_cumulative_guard": 1400,
  "production_paths": [],
  "test_paths": ["tasks/TASK-0009/MEASUREMENT.md"],
  "entrypoint": null,
  "fixture_elevation_needs": "Read-only frozen canonical JSONL snapshot; no elevation, network, product fixture, or operational-log write.",
  "lap_1": "After TASK-0013, TASK-0014, and TASK-0008 merge, freeze the completed-event snapshot and regenerate provenance-complete historical plus new-wave SLOC/test/stage/active/wait/retry/raw/effective classification evidence, applying ceil(observed non-preflight time * 1.20) only to observable time.",
  "lap_2": "Independent REVIEW and QA each run canonical parse/unique-ID/correction-edge checks, independently regenerate measurement and cap arithmetic, and run the repository-native full Go/format/diff checks; main owns Git. TASK-0010 speculative planning is invalidated if evidence changes its boundary.",
  "exclusions": ["all product/test implementation", "audit", "attestation", "release", "installer", "canary", "editing canonical log"],
  "split_stop_rule": "Stop on missing or contradictory canonical evidence, non-reproducible arithmetic, actual cumulative above target 1300, or inability to independently regenerate in Lap 2; classify before retry and do not bypass TASK-0010.",
  "measurement_lineage": "Include the terminated TASK-0007 raw evidence and completed replacement TASK-0013/TASK-0014 lineage, preserve null with reasons, validate every correction target earlier in file order and same task/lap with smaller sequence, retain raw source IDs and superseded_by, and derive effective values only after validation.",
  "later_reserve_eligibility": "Audit/attestation/manual-canary reserve remains non-executable until TASK-0012 PASS+merge; no converted milestone remains simultaneously reserved and executable.",
  "contract_path": "tasks/TASK-0009/TASK.md"
}
```

## Purpose and evidence boundary

This is a zero-production-SLOC measurement and replanning gate after exactly
the two replacement production Tasks TASK-0013/TASK-0014 and TASK-0008. It measures the immutable
historical baseline and the new completed records, then allows only the next
bounded contract (TASK-0010) to proceed after independent REVIEW and QA. It
does not implement product behavior and does not edit the canonical log.

The canonical source is the read-only
`/home/ubuntu/git/agent-harness-work/lap30/events.jsonl`. A completed task is
only `event == "lap_completed" && status == "completed"`; the completed set
for this gate remains exactly TASK-0001, TASK-0003, TASK-0004, and TASK-0005.
Corrections are not additional tasks.

## Preflight and two-Lap delivery

Preflight verifies merged TASK-0013/TASK-0014/TASK-0008, freezes the completed-event
snapshot, and confirms that the source and worktree are readable. A preflight
failure is `not_started`, excluded from cycle/stage timing, and never replaced
with a synthetic zero.

Lap 1 produces provenance-complete SLOC/test/stage/active/wait/retry and
raw/effective-classification evidence. Required fields include planned and
actual production SLOC, cumulative SLOC, test LOC, PLAN/DEV/REVIEW/QA and
CI-push-merge milliseconds and minutes, cycle time, `active_ms`, `wait_ms`,
maximum propagated retries, source event IDs, null reasons, and units/rounding.
Stage timing uses only same-task/lap/stage/attempt start-terminal pairs;
`stage == "preflight"` is excluded. Apply
`ceil(observed_non_preflight_time * 1.20)` only to observable time; null stays
null and SLOC gets no multiplier.

Use this executable validation before deriving effective values:

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

The predicate requires a nonempty correction target, earlier frozen-file
position, same `task_id`, same `lap_id`, and smaller within-lap `sequence`.
Only after it passes may raw history gain `superseded_by` and effective values
omit the superseded field/classification. Invalid edges stop measurement;
history is never deleted or rewritten. In particular, TASK-0005's raw
`requirement_gap` remains visible as superseded by
`task0005-lap01-correct-dev-class`, while effective classifications contain
`planning_defect` (and any other unsuperseded value), not an unlabeled mixture.

Lap 2 has independent REVIEW and QA each regenerate the measurement and cap
arithmetic from the frozen snapshot and run the full repository checks below.
TASK-0010's boundary is invalidated and replanned if measured evidence changes
it. Main owns final checks and Git.

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

## Acceptance and exclusions

- Exactly one evidence row exists for each of the four completed canonical
  task IDs, with no row created from a correction or incomplete task.
- SLOC/test, timing/null, preflight, active/wait, retry, correction, source-ID,
  contingency, and cap arithmetic is independently reproducible.
- Raw historical classifications and effective unsuperseded classifications
  are separate and correction-provenanced.
- No fixed SLOC throughput, LOC velocity, average, or imputed timing sizes the
  next contract.
- The gate adds exactly 0 production SLOC; forecast cumulative production is
  1207 before independent measurement reconciliation.

This Task excludes all product/test implementation, audit, attestation,
release, installer, canary, and detailed contracts beyond the next bounded
wave. It owns no `MEASUREMENT.md` at DEV time; that result is a later REVIEW/
QA output and is not one of TASK-0006's seven DEV outputs.

## Measurement, caps, and stop rule

The forecast is +0 production SLOC and cumulative 1207; post-reestimate stop
1250, target cap 1300, hard guard 1400. Stop on missing or contradictory canonical
evidence, non-reproducible arithmetic, actual cumulative above 1300, or
inability to independently regenerate in Lap 2. Classify before retry and do
not bypass TASK-0010. Record active/wait and retries without double-counting
snapshots, and preserve null with an explicit reason.

## Gate and later reserve

Independent REVIEW PASS and QA PASS are required; a FAIL returns to its
responsible gate and never merges. No later audit/attestation/manual-canary
milestone may receive PLAN, branch, DEV, or PR-ready detail until TASK-0012
passes independent REVIEW and QA and main merges it.
