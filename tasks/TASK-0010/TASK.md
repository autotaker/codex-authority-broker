# TASK-0010: local push policy and validation (v2 deferred)

**Depends on:** TASK-0009 measurement PASS+merge and the explicit v2 replan.

**Status:** deferred-v2; non-executable.

## Replan metadata

```json
{
  "id": "TASK-0010",
  "title": "local push policy and validation",
  "status": "deferred-v2",
  "executable": false,
  "depends_on": ["TASK-0009"],
  "baseline_production_sloc": 1253,
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1253,
  "v2_reserved_production_sloc": 130,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1495,
  "hard_cumulative_guard": 1800,
  "production_paths": ["internal/push/policy.go", "internal/push/validate.go"],
  "test_paths": ["internal/push/policy_test.go", "internal/push/validate_test.go"],
  "entrypoint": "internal/push/policy.go",
  "fixture_elevation_needs": "Future v2 replan only: temporary worktree and local bare-repository matrix; no network, credentials, child Git process, or elevation.",
  "lap_1": "Not authorized while deferred. A future approved v2 PLAN must remeasure the retained readable boundary and prove exact repository, clean tree, permitted branch, one-ref non-force shape, and zero transport on denials.",
  "lap_2": "Not authorized while deferred. A future approved v2 QA_PLAN must independently review and QA the focused matrix and repository-native checks.",
  "exclusions": ["token custody", "credential helper", "network transport", "Git child process", "backend registration", "sudo", "audit", "release", "installer", "canary"],
  "split_stop_rule": "No v2 DEV starts without a fresh PLAN/QA_PLAN, fresh frozen evidence, explicit ordered-shedding review, readable estimate, and cap proof; never compress or weaken authorization and denial-before-transport safety.",
  "measurement_lineage": "TASK-0009 measured baseline is 1253. User-confirmed v1 scope drops GitHub push; the former +130 forecast is retained only as deferred-v2 reserve evidence, contributes zero v1 production SLOC, and carries no throughput or SLOC velocity forward.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains blocked until TASK-0012 independently passes REVIEW and QA and main merges it.",
  "replan_reason": "User-confirmed v1 scope drops the coupled GitHub push capability. The former TASK-0010 (+130) and TASK-0011 (+153) forecasts total 283 and would reach 1536, 36 over the mandatory-v1 1500 target; TASK-0010 remains deferred-v2, executable:false, and contributes zero v1 production SLOC.",
  "contract_path": "tasks/TASK-0010/TASK.md"
}
```

## Explicit v2 disposition

The user-confirmed v1 scope decision drops GitHub push. TASK-0010 therefore
remains `deferred-v2`, `executable:false`, and contributes zero v1 production
SLOC; its former +130 estimate is retained only as a v2 reserve for a future
approved replan.

The frozen TASK-0009 measurement fixes the current merged baseline at **1253
production SLOC**. The former TASK-0010/TASK-0011 forecasts totalled 283 and
would reach 1536, which exceeds the mandatory-v1 target of 1500 by 36. The
unconditional 1800 hard limit is not permission to exceed that target.

The ordered shedding list was reviewed in order. Items 1–6 have no remaining
applicable optional scope in the coupled push work, so item 7 is selected:

**GitHub push capability moves to v2.** TASK-0010 is retained as a named v2
reserve with its paths and security boundary, but it is not executable. No
branch, DEV, PR-ready detail, production allocation, or implementation may be
started from this document. The future policy must retain exact repository and
ref identity, clean-tree and permitted-branch checks, one-ref non-force
admission, and denial before any transport boundary; readable idiomatic code
and mandatory security tests cannot be compressed or removed.

TASK-0011 is the coupled v2 reserve and remains blocked behind this deferred
Task. The planned TASK-0017 blocker must pass and merge before TASK-0012 can
freeze the post-TASK-0017 actual zero-SLOC measurement gate, so no speculative
TASK-0010 arithmetic is enabled in v1.

## Future re-entry and gate

Re-entry requires a new approved PLAN and independent QA_PLAN based on fresh
evidence, explicit cap arithmetic against baseline 1253/target 1500/hard
1800, and a renewed ordered-shedding review. Until then this Task contributes
zero production SLOC, cannot be started, and does not unlock later reserve
work. Main owns any future branch, Git, review, QA, and merge actions.
