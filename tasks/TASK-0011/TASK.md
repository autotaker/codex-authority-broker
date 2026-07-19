# TASK-0011: token custody and system-Git non-force push (v2 deferred)

**Depends on:** deferred TASK-0010 and a later explicit v2 replan.

**Status:** deferred-v2; non-executable.

## Replan metadata

```json
{
  "id": "TASK-0011",
  "title": "token custody and system-Git non-force push",
  "status": "deferred-v2",
  "executable": false,
  "depends_on": ["TASK-0010"],
  "baseline_production_sloc": 1253,
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1253,
  "v2_reserved_production_sloc": 153,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1495,
  "hard_cumulative_guard": 1800,
  "production_paths": ["cmd/codex-authority-push/main.go", "cmd/codex-authority-broker/main.go", "internal/ipc/protocol.go", "internal/push/custody.go", "internal/push/system_git.go", "internal/backend/push_registration.go"],
  "test_paths": ["cmd/codex-authority-push/main_test.go", "cmd/codex-authority-broker/main_test.go", "internal/ipc/protocol_test.go", "internal/push/custody_test.go", "internal/push/system_git_test.go", "internal/backend/push_registration_test.go"],
  "entrypoint": "cmd/codex-authority-push/main.go",
  "fixture_elevation_needs": "Future v2 replan only: stable handler seam, configured caller UID, local bare remote, fake short-lived token provider, system-Git binary, credential-capture sentinel, and live-lease fixture; no network or elevation.",
  "lap_1": "Not authorized while deferred. A future approved v2 PLAN must remeasure the bounded caller/schema, UID/live-lease/policy gates, memory-only custody, redaction, and single-ref non-force path.",
  "lap_2": "Not authorized while deferred. A future approved v2 QA_PLAN must independently mutate malformed, wrong-identity, expired-authority, leakage, force, and ambiguity cases before custody/Git and run repository-native checks.",
  "exclusions": ["changes to cmd/codex-authority/main.go", "broker lifecycle changes beyond one fixed registration call", "generic IPC commands", "arbitrary refspec", "remote-OID prefetch", "force/tag/delete push", "sudo", "audit", "release", "installer", "canary"],
  "split_stop_rule": "No v2 DEV starts without fresh approved plans, fresh evidence, explicit cap/shedding proof, and a readable retained-core estimate; never weaken UID, lease, schema, custody, redaction, or non-force safety.",
  "measurement_lineage": "TASK-0009 measured baseline is 1253. The former +153 allocation is retained only as v2 reserve evidence; no throughput, reserve borrowing, or contingency-to-SLOC conversion is allowed.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains blocked until TASK-0012 independently passes REVIEW and QA and main merges it.",
  "replan_reason": "The coupled former TASK-0010 (+130) and TASK-0011 (+153) forecasts total 283; 1253 + 283 = 1536, 36 above the mandatory-v1 target. Ordered shedding item 7 explicitly defers GitHub push to v2.",
  "contract_path": "tasks/TASK-0011/TASK.md"
}
```

## Explicit v2 disposition

TASK-0009 measured 1253 actual merged production SLOC. Restoring the former
TASK-0010/TASK-0011 forecasts would produce 1536, over the mandatory-v1 1500
target by 36. Although 1800 is the unconditional hard limit, it does not
authorize target overflow. The ordered feature-shedding review reached item 7
after confirming that items 1–6 contain no remaining optional scope in this
coupled push implementation.

**GitHub push capability moves to v2.** TASK-0011 therefore remains a named,
non-executable reserve. Its future boundary retains the dedicated caller,
bounded `OperationPush`/`PushRequest` admission, fail-closed configured UID,
live lease and TASK-0010 policy gates before token/Git, memory-only token
custody, secret non-disclosure, deterministic system-Git capture, and a single
ref non-force path. Readability, normal error handling, and mandatory safety
tests are preserved; no code is compressed or silently removed.

No branch, DEV, PR-ready detail, production allocation, fixture execution, or
implementation is authorized while this Task is deferred. TASK-0010 must first
be explicitly replanned and approved in a future v2 wave. TASK-0012 now
measures the direct zero-SLOC v1 wave after TASK-0009.

## Future re-entry and gate

Re-entry requires a fresh frozen evidence snapshot, new approved PLAN and
independent QA_PLAN, exact cap arithmetic from baseline 1253 against target
1500/hard 1800, and renewed ordered-shedding review. Until then this Task
contributes zero production SLOC and cannot unlock later reserve work. Main
owns future branch, Git, review, QA, and merge actions.
