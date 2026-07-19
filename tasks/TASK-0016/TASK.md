# TASK-0016: bounded active-lease authorization operation

**Depends on:** TASK-0015 (merged as `fa70d2fc5b8001a00b8cff476292b626cfe61740`).

**Status:** planned and executable after independent PLAN and QA_PLAN approval.

## Contract metadata

```json
{
  "id": "TASK-0016",
  "title": "bounded active-lease authorization operation",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0015"],
  "expected_production_sloc": 20,
  "expected_cumulative_production_sloc": 1220,
  "target_cumulative_cap": 1250,
  "projected_cap_trigger_sloc": 1230,
  "hard_cumulative_guard": 1350,
  "production_paths": ["internal/ipc/protocol.go", "internal/backend/runtime.go"],
  "test_paths": ["internal/ipc/protocol_test.go", "internal/backend/runtime_test.go"],
  "process_evidence_paths": ["backlog.json", "tasks/TASK-0016/TASK.md", "tasks/TASK-0016/PLAN.md", "tasks/TASK-0016/QA_PLAN.md", "tasks/TASK-0016/REVIEW_RESULT.md", "tasks/TASK-0016/QA_RESULT.md"],
  "entrypoint": null,
  "fixture_elevation_needs": "None; fake-clock in-process runtime tests and existing temporary Unix-socket IPC tests only.",
  "lap_1": "With PLAN and TASK-first QA_PLAN approved in the partial interval, add one fixed payload-free authorize operation whose only positive result means the process-local lease is active at the decision point; preserve ready/otp semantics, one custom registration slot, bounded boolean responses, cancellation, and close fail-closed behavior; complete DEV, independent REVIEW/QA, and Main Git closure in one counted Lap.",
  "lap_2": "Exceptional only for one or two concrete review/QA findings requiring no redesign, research, Task redefinition, or fixture change and demonstrably reviewable within the first 20 minutes; otherwise split. No Lap 3.",
  "exclusions": ["sudo client or policy", "daemon lifecycle", "seed handling", "push operation or credentials", "lease persistence", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Stop before DEV if authorization requires payload, identity material in the response, lease deadline disclosure, persistence, sudo/push coupling, or a protocol version change. Stop/replan above the local 1230 trigger; 1250 is the Task target and 1350 the absolute local guard. Never alter ready/otp meaning, weaken expiry, or consume the existing custom registration capacity.",
  "measurement_lineage": "Requirement gap discovered during TASK-0008 partial preflight: ready calls BeginReadiness and therefore denies an active lease, while TASK-0008 needs a live allow decision and forbids runtime/IPC changes. Baseline is merged TASK-0015 cumulative 1200; forecast +20 = 1220.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0016/TASK.md"
}
```

## Purpose and boundary

Add exactly one fixed IPC operation, `authorize`, that accepts no payload and
returns only the existing bounded `OK` decision. `OK=true` is permitted only
when the runtime's process-local lease is active at the handler's decision
point. Expired, absent, restarted, cancelled, or closed authority denies.

The operation does not return a lease identifier, deadline, token, identity,
reason, or other payload. Existing `ready` and `otp` meanings and wire encoding
remain unchanged. Runtime capacity must still retain one slot for the existing
bounded custom registration seam.

Only the four production/test paths named in metadata may change the product
candidate. The explicitly listed Main/role-owned process-evidence paths may be
created or updated in the same Task PR; they are not product scope and do not
count toward production SLOC. Every other path remains forbidden. Sudo/PAM,
broker lifecycle, seed admission, push, credentials, persistence, audit,
installation, release, and canary work are excluded.

## Acceptance

- Protocol read/write accepts exactly `ready`, `otp`, and payload-free
  `authorize`; malformed, unknown, payload-bearing, and wrong-version requests
  remain fail-closed.
- Before activation, `authorize` denies. Immediately after valid activation it
  allows; at the immutable lease deadline and after it denies.
- A fresh runtime after restart denies; `Close`, caller cancellation, or a
  cancellation race cannot produce an allow.
- `authorize` does not open or replace a readiness challenge, consume OTP,
  extend the lease, disclose state, or alter subsequent `ready`/`otp` results.
- Concurrent authorization and expiry/close remain race-clean and fail closed.
- Adding the built-in operation preserves one usable custom registration slot.
- Focused and repository-wide checks pass with readable production delta at or
  below the 1230 local trigger and without compression.

## Required evidence and stops

Lap 1 runs focused backend, lease, and IPC tests plus full regression, race,
format, static, diff, and metadata checks. Independent REVIEW and QA must map
each acceptance condition to named tests and inspect the implementation rather
than relying on names alone. Main alone performs Git publication and merge.

Stop before DEV if the design needs a payload, protocol-version bump, returned
identity/deadline, persistence, sudo or push coupling, or loss of custom
registration capacity. Forecast or candidate above cumulative 1230 requires
approved replan; 1250 is the target and 1350 is absolute. Lap 2 is exceptional
only under the four facts in metadata; no Lap 3.
