# TASK-0007: terminated combined daemon/backend attempt

**Depends on:** TASK-0006 (merged).

**Status:** terminated incomplete under user-approved Disposition 2.

## Contract metadata

```json
{
  "id": "TASK-0007",
  "title": "terminated combined daemon/backend attempt",
  "status": "terminated",
  "executable": false,
  "depends_on": ["TASK-0006"],
  "expected_production_sloc": 0,
  "production_sloc_added": 0,
  "expected_cumulative_production_sloc": 751,
  "target_cumulative_cap": null,
  "projected_cap_trigger_sloc": null,
  "hard_cumulative_guard": null,
  "production_paths": [],
  "test_paths": [],
  "entrypoint": null,
  "completion": false,
  "termination_classification": "planning_defect",
  "superseded_by": ["TASK-0013", "TASK-0014"],
  "fixture_elevation_needs": "None after termination; the unfinished draft was never merged or installed.",
  "lap_1": "Consumed by planning and contract repair; no gate-ready product candidate was produced.",
  "lap_2": "Consumed by interrupted DEV and remeasurement; the combined draft exceeded its task hard guard before tests and was not reviewed, QA-approved, committed, or merged.",
  "exclusions": ["further DEV", "Lap 3", "counter reset", "candidate adoption", "product merge"],
  "split_stop_rule": "Terminal. User-approved Disposition 2 preserves both Lap records and moves all unmerged functionality to TASK-0013 and TASK-0014; TASK-0007 cannot resume.",
  "measurement_lineage": "Canonical TASK-0007 Lap 1 and Lap 2 events remain the source for the planning_defect termination; merged production delta is zero.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0007/TASK.md"
}
```

## Terminal disposition

TASK-0007 consumed exactly two Laps. Lap 1 ended after planning repair and
Lap 2 stopped during DEV when the readable two-file draft measured 336
canonical production SLOC and cumulative 1087, above its task hard guard
1050, before either owned test file existed. It received no independent
REVIEW, QA PASS, commit, PR, or merge. Its merged production delta is zero.

The user explicitly approved Disposition 2. The original combined boundary is
therefore terminated incomplete with classification `planning_defect`, not
reset or relabeled. There is no TASK-0007 Lap 3 and no further executable gate.

## Replacement ownership

- TASK-0013 exclusively owns `internal/backend/runtime.go` and
  `internal/backend/runtime_test.go`.
- TASK-0014 depends on TASK-0013 and exclusively owns
  `cmd/codex-authority-broker/main.go` and its test.

The unfinished local drafts remain unaccepted evidence until their respective
replacement TASK/PLAN/QA gates authorize DEV. Termination does not authorize
their staging, adoption, or merge.

## Evidence and reserve

The append-only Lap30 records remain the canonical evidence for both consumed
Laps and the stop classification. TASK-0013 and TASK-0014 each receive their
own maximum-two-Lap cycle. Later audit, attestation, and manual canary work
remain ineligible until TASK-0012 independently passes REVIEW and QA and Main
merges it.
