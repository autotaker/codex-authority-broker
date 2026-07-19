# TASK-0013: process-local ready/otp runtime assembly

**Depends on:** TASK-0006 (merged).

**Status:** planned and executable after the Disposition 2 contract merge.

## Contract metadata

```json
{
  "id": "TASK-0013",
  "title": "process-local ready/otp runtime assembly",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0006"],
  "expected_production_sloc": 150,
  "expected_cumulative_production_sloc": 901,
  "target_cumulative_cap": 950,
  "projected_cap_trigger_sloc": 925,
  "hard_cumulative_guard": 1000,
  "production_paths": ["internal/backend/runtime.go"],
  "test_paths": ["internal/backend/runtime_test.go"],
  "entrypoint": "internal/backend/runtime.go",
  "fixture_elevation_needs": "Injected lease clock, verifier secret, and in-process IPC requests; no socket, seed file, network, credentials, or elevation.",
  "lap_1": "After approved PLAN and QA_PLAN, implement the process-local runtime only; run go test ./internal/backend ./internal/ipc ./internal/lease and capture exact ready/otp admission, denial, close, registration, and regression evidence.",
  "lap_2": "Independent REVIEW runs focused and repository-native full checks; QA independently exercises routing, payload mutations, state denial, close, bounded registration, and existing IPC/lease regressions; main owns Git.",
  "exclusions": ["seed file acquisition", "listener", "signals", "daemon entrypoint", "sudo", "push enablement", "credentials", "persistence", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Stop before or during DEV if a seed/listener/lifecycle path is required, runtime tests are not gate-ready in Lap 1, cumulative forecast or measurement exceeds 925, or another production path is needed; never use Lap 3.",
  "measurement_lineage": "Replacement for the runtime half of terminated TASK-0007; record production/test SLOC, paired stage timing, active/wait, retries, classifications, null reasons, and time-only contingency.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0013/TASK.md"
}
```

## Purpose and ownership

Construct the process-local backend from an injected TOTP secret and one
in-memory lease state. Admit only the existing `ready` and `otp` operations,
return bounded payload-free decisions, close fail-closed, and expose one
instance-level third-slot registration seam for TASK-0011 without enabling a
third operation now.

Production owns only `internal/backend/runtime.go`; tests own only
`internal/backend/runtime_test.go`. Seed files, sockets, signals, daemon
startup, sudo, push behavior, credentials, persistence, and installation are
outside this Task.

## Two-Lap delivery

Lap 1 starts only after approved PLAN and TASK-first QA_PLAN. DEV produces a
focused-test-passing candidate and runs:

```sh
go test ./internal/backend ./internal/ipc ./internal/lease
```

Evidence covers exact ready/otp payload admission, challenge/activation state
mapping, malformed and unknown denial, close behavior, duplicate/invalid/full
registration denial, absence of push, and existing IPC/lease regression.

Lap 2 contains no planned DEV. Independent REVIEW runs the focused command and
repository-native full checks; independent QA repeats the mutation matrix and
regressions. Main alone owns staging, commit, PR, and merge.

## Stop and SLOC controls

Forecast is +150 and cumulative 901. Stop at forecast or measurement above
925, target 950, or hard guard 1000. Stop if another production path, seed,
listener, signal, or lifecycle behavior is required, or if Lap 1 is not
gate-ready. Do not compress code, remove tests, weaken denial behavior, or use
a third Lap.
