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
  "expected_production_sloc": 280,
  "expected_cumulative_production_sloc": 1202,
  "target_cumulative_cap": 1250,
  "projected_cap_trigger_sloc": 1202,
  "hard_cumulative_guard": 1350,
  "production_paths": ["cmd/codex-authority-broker/main.go"],
  "test_paths": ["cmd/codex-authority-broker/main_test.go"],
  "entrypoint": "cmd/codex-authority-broker/main.go",
  "fixture_elevation_needs": "Temporary Unix socket and injected descriptor-relative openat/O_NOFOLLOW seed fixture simulating root-owned mode-0600 metadata without host mutation; no real service installation.",
  "lap_1": "After TASK-0013 merge and approved PLAN and QA_PLAN, implement bounded seed admission and daemon lifecycle only; run go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease.",
  "lap_2": "Because Lap 1 stopped at the prior trigger, Lap 2 may contain the single conditionally approved DEV correction and deterministic broker-test completion within the 1202 trigger before independent REVIEW and QA; no Lap 3. Independent REVIEW then runs focused and full checks; QA independently exercises valid and invalid seed metadata/schema, redaction, startup, shutdown, restart, and existing-client regression; main owns Git.",
  "exclusions": ["runtime API changes", "new IPC operation", "sudo", "push", "GitHub credentials", "persistence", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Stop before or during DEV if runtime/IPC changes are required, descriptor-safe metadata cannot be isolated, the single conditionally approved Lap 2 DEV correction/test completion is not gate-ready, or cumulative forecast or measurement exceeds the 1202 trigger; never use Lap 3. A forecast or candidate above target 1250 stops for explicit replan and exact ordered shedding review; hard guard 1350 is absolute.",
  "measurement_lineage": "Replacement for the secure seed/daemon half of terminated TASK-0007; baseline 922 + independently measured readable TASK-0014 floor 280 = cumulative 1202. Record production/test SLOC, paired stage timing, active/wait, retries, classifications, null reasons, and time-only contingency.",
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

Lap 1 stopped at the prior 1200 trigger. Subject to the amended 1202 trigger,
Lap 2 may contain the single conditionally approved DEV correction and
deterministic broker-test completion before independent REVIEW and QA; no Lap
3. That DEV work remains within the same two-path scope and runs:

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease
```

Evidence covers valid startup, final and parent symlink denial, missing,
malformed, duplicate, unknown, oversized, wrong-owner/mode/type and read-error
seed denial, redaction, listen failure, signal shutdown, socket cleanup, valid
restart, and restart-without-seed denial.

After that single conditional DEV completion, independent REVIEW runs
focused/full checks; independent QA repeats the mutation, redaction,
lifecycle, and existing-client matrix. Main alone owns Git.

## Stop and SLOC controls

The actual merged baseline is 922. The independently measured readable floor
is +280 and cumulative 1202. The projected trigger is 1202; stop on a forecast
or measurement above that trigger. Target 1250 requires explicit replan and
the exact global ordered-shedding review; hard guard 1350 is absolute. Stop if
a runtime or IPC production change is required, descriptor-safe metadata cannot
be isolated, or the single conditionally approved Lap 2 DEV correction/test
completion is not gate-ready. Do not compress code, weaken seed or lifecycle
controls, delete tests, or use a third Lap.
