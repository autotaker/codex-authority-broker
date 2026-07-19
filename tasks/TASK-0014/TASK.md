# TASK-0014: secure bounded seed and daemon lifecycle

**Depends on:** TASK-0013 (merged).

**Status:** planned and executable after TASK-0013 merge.

## Contract metadata

```json
{
  "id": "TASK-0014",
  "title": "secure bounded seed and daemon lifecycle",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0013"],
  "expected_production_sloc": 186,
  "expected_cumulative_production_sloc": 1087,
  "target_cumulative_cap": 1150,
  "projected_cap_trigger_sloc": 1125,
  "hard_cumulative_guard": 1200,
  "production_paths": ["cmd/codex-authority-broker/main.go"],
  "test_paths": ["cmd/codex-authority-broker/main_test.go"],
  "entrypoint": "cmd/codex-authority-broker/main.go",
  "fixture_elevation_needs": "Temporary Unix socket and injected descriptor-relative openat/O_NOFOLLOW seed fixture simulating root-owned mode-0600 metadata without host mutation; no real service installation.",
  "lap_1": "After TASK-0013 merge and approved PLAN and QA_PLAN, implement bounded seed admission and daemon lifecycle only; run go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease.",
  "lap_2": "Independent REVIEW runs focused and full checks; QA independently exercises valid and invalid seed metadata/schema, redaction, startup, shutdown, restart, and existing-client regression; main owns Git.",
  "exclusions": ["runtime API changes", "new IPC operation", "sudo", "push", "GitHub credentials", "persistence", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Stop before or during DEV if runtime/IPC changes are required, descriptor-safe metadata cannot be isolated, Lap 1 lacks gate-ready tests, or cumulative forecast or measurement exceeds 1125; never use Lap 3.",
  "measurement_lineage": "Replacement for the secure seed/daemon half of terminated TASK-0007; record production/test SLOC, paired stage timing, active/wait, retries, classifications, null reasons, and time-only contingency.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0014/TASK.md"
}
```

## Purpose and ownership

Build the privileged broker entrypoint against the merged TASK-0013 runtime.
Read the fixed seed once through a descriptor-relative Linux `openat` walk
from `/` with no-follow and close-on-exec semantics, validate final-descriptor
type, root ownership, exact mode 0600 and bounded strict schema, then listen
only after construction succeeds. Handle signals, close and unlink cleanly,
and restart with fresh in-memory state.

Production owns only `cmd/codex-authority-broker/main.go`; tests own only its
test. Runtime and IPC APIs, sudo, push, credentials, persistence, installation,
and audit are outside this Task.

## Two-Lap delivery

Lap 1 starts only after TASK-0013 merge and approved PLAN and TASK-first
QA_PLAN. DEV runs:

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease
```

Evidence covers valid startup, final and parent symlink denial, missing,
malformed, duplicate, unknown, oversized, wrong-owner/mode/type and read-error
seed denial, redaction, listen failure, signal shutdown, socket cleanup, valid
restart, and restart-without-seed denial.

Lap 2 contains no planned DEV. Independent REVIEW runs focused/full checks;
independent QA repeats the mutation, redaction, lifecycle, and existing-client
matrix. Main alone owns Git.

## Stop and SLOC controls

Forecast is +186 and cumulative 1087. Stop above 1125, target 1150, or hard
1200. Stop if a runtime or IPC production change is required, descriptor-safe
metadata cannot be isolated, or Lap 1 is not gate-ready. Do not compress code,
weaken seed or lifecycle controls, delete tests, or use a third Lap.
