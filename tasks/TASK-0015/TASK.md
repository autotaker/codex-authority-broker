# TASK-0015: secure daemon review-gap replacement

**Depends on:** TASK-0013 (merged).

**Status:** planned and executable after approved PLAN and QA_PLAN.

## Contract metadata

```json
{
  "id": "TASK-0015",
  "title": "secure daemon review-gap replacement",
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
  "lap_1": "With PLAN and TASK-first QA_PLAN approved during the preceding partial interval, start DEV within 5 minutes, complete implementation and deterministic tests by minute 20, independent REVIEW by minute 25, then QA and main-owned Git closure by minute 30.",
  "lap_2": "Exceptional only when Lap 1 leaves concrete estimated work, no redesign or research, one or two classified failure causes, and a fix demonstrably possible in the first 20 minutes; otherwise split the Task. No Lap 3.",
  "exclusions": ["runtime API changes", "new IPC operation", "sudo", "push", "GitHub credentials", "persistence", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Local SLOC projections are warnings for Main judgment, not automatic reapproval gates. Stop for a security-boundary change, scope expansion, unreadable compression, target-overflow requiring ordered shedding, or the absolute 1800 global hard limit; never weaken the seed/lifecycle controls or use Lap 3.",
  "measurement_lineage": "Replacement for unfinished TASK-0014. Baseline 922 + readable forecast 280 = cumulative 1202; the TASK-0014 REVIEW FAIL supplies the immediate-wipe and complete deterministic-test acceptance gaps. Record planning/wait separately and target one counted Lap.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0015/TASK.md"
}
```

## Purpose and owned boundary

Build the privileged broker entrypoint against the merged TASK-0013 runtime.
Read the seed once through a descriptor-relative Linux `openat` walk from `/`
with no-follow and close-on-exec semantics; validate type, root ownership,
exact mode 0600, bounded strict schema, and maximum secret size. Construct the
runtime, immediately wipe the caller-owned decoded secret before listening,
then handle signals, active clients, close, unlink, and restart fail-closed.

Only `cmd/codex-authority-broker/main.go` and its test are owned. Runtime and
IPC APIs, sudo, push, credentials, persistence, installation, and audit are
outside this Task.

## Acceptance and named deterministic evidence

- `TestRunWipesCallerSecretBeforeListen` uses a listener-order barrier and
  proves the caller-owned decoded secret is zero before the first listen call.
- Descriptor tests cover root open/close errors, parent and final symlinks,
  final reader close, exact uid/mode/type, every size boundary, and read errors.
- Schema tests cover valid and maximum secrets plus malformed, duplicate,
  unknown, missing, empty, and oversized input without secret diagnostics.
- Lifecycle tests cover runtime-factory error, listener error with non-nil
  server cleanup, construction-before-listen, unexpected `Serve` return,
  server-close error, SIGINT, SIGTERM, active/concurrent clients, repeated
  shutdown, socket identity replacement, successful restart, and restart with
  missing seed.
- Client tests cover a valid OTP request and malformed input without changing
  existing TASK-0013 runtime or IPC behavior.

QA_PLAN must map every row above to an exact test name before DEV. Early review
may prepare and inspect partial read-only diffs, but its final verdict remains
independent and follows the complete candidate.

## One-Lap standard and exception

PLAN and independent TASK-first QA_PLAN are prepared before the counted Lap.
During the Lap: DEV starts by minute 5, implementation and tests target minute
20, independent REVIEW targets minute 25, and QA plus Main-owned Git closure
target minute 30. Planning plus pure wait should remain at or below 20%.

Lap 2 is exceptional and requires all four facts: concrete estimated residue,
no redesign or research, only one or two classified failure causes, and a fix
possible in its first 20 minutes. Otherwise split. The same cause may trigger
at most one replan; no Lap 3.

## SLOC and stop controls

The merged baseline is 922; readable forecast +280 gives cumulative 1202.
Local projections and the 1250 Task target are warnings for Main judgment and
do not by themselves trigger plan reapproval. Preserve readable Go and full
tests. The mandatory-v1 target remains 1500 and the unconditional global hard
limit remains 1800. Stop for scope or security-boundary change, unreadable
compression, target overflow requiring the ordered shedding audit, or hard
limit risk.
