# TASK-0014: secure bounded seed and daemon lifecycle

**Depends on:** TASK-0013 (merged).

**Status:** unfinished and non-executable; superseded by TASK-0015.

## Contract metadata

```json
{
  "id": "TASK-0014",
  "title": "secure bounded seed and daemon lifecycle",
  "status": "unfinished",
  "executable": false,
  "depends_on": ["TASK-0013"],
  "expected_production_sloc": 0,
  "production_sloc_added": 0,
  "expected_cumulative_production_sloc": 922,
  "target_cumulative_cap": null,
  "projected_cap_trigger_sloc": null,
  "hard_cumulative_guard": null,
  "production_paths": [],
  "test_paths": [],
  "entrypoint": null,
  "completion": false,
  "superseded_by": ["TASK-0015"],
  "fixture_elevation_needs": "None after closure; the reviewed candidate was never merged or installed.",
  "lap_1": "Stopped on local planning-boundary churn before a reviewable product candidate was completed.",
  "lap_2": "Produced an unmerged candidate, but independent REVIEW failed on immediate secret wipe and deterministic test coverage; no Lap 3 is permitted.",
  "exclusions": ["runtime API changes", "new IPC operation", "sudo", "push", "GitHub credentials", "persistence", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Terminal. Preserve both Lap records and the independent REVIEW FAIL; move the unmerged boundary to TASK-0015 without resetting counters or adopting the failed candidate as merged product.",
  "measurement_lineage": "TASK-0014 observed a 278-SLOC candidate but merged production delta is zero. Preserve its final REVIEW FAIL as the source for TASK-0015 acceptance and test coverage.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0014/TASK.md"
}
```

## Purpose and ownership

This Task is closed unfinished. It attempted to build the privileged broker
entrypoint against the merged TASK-0013 runtime.
Read the fixed seed once through a descriptor-relative Linux `openat` walk
from `/` with no-follow and close-on-exec semantics, validate final-descriptor
type, root ownership, exact mode 0600 and bounded strict schema, then listen
only after construction succeeds. Handle signals, close and unlink cleanly,
and restart with fresh in-memory state.

Its reviewed 278-SLOC candidate and tests were never merged. The final
independent REVIEW failed because the caller-owned decoded seed was wiped only
when `Serve` returned, and because deterministic tests did not cover every
specified ownership, bound, lifecycle, signal, restart, race, and malformed
client branch. TASK-0015 owns that replacement boundary.

## Closed two-Lap record

Lap 1 was consumed by repeated local-boundary reconciliation. Lap 2 produced a
candidate, but independent REVIEW recorded the two failure classifications
above. No Lap 3 is allowed and this Task cannot resume.

## Stop and SLOC controls

The actual merged baseline remains 922 and this Task added zero merged
production SLOC. Historical candidate measurements remain evidence, not
product. TASK-0015 may use that evidence but must satisfy fresh PLAN, QA_PLAN,
DEV, independent REVIEW, and independent QA gates.
