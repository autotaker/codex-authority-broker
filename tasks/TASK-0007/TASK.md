# TASK-0007: production daemon/backend assembly and bounded seed

**Depends on:** TASK-0006 (merged).

**Status:** planned and executable.

## Contract metadata

```json
{
  "id": "TASK-0007",
  "title": "production daemon/backend assembly and bounded seed",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0006"],
  "expected_production_sloc": 90,
  "expected_cumulative_production_sloc": 841,
  "target_cumulative_cap": 950,
  "projected_cap_trigger_sloc": 855,
  "hard_cumulative_guard": 1050,
  "production_paths": ["cmd/codex-authority-broker/main.go", "internal/backend/runtime.go"],
  "test_paths": ["cmd/codex-authority-broker/main_test.go", "internal/backend/runtime_test.go"],
  "entrypoint": "cmd/codex-authority-broker/main.go",
  "fixture_elevation_needs": "Temporary Unix socket and root-owned mode-0600 seed fixture; isolated runner must simulate ownership/mode without broad host mutation; no real service installation.",
  "lap_1": "After approved PLAN and QA_PLAN, confirm TASK-0006 merge, existing ready/otp protocol and client/daemon separation, then implement only the entrypoint and runtime assembly; run go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease and capture startup, ready/OTP, shutdown, restart, and existing-client regression evidence.",
  "lap_2": "Independent REVIEW runs the focused suite and repository-native full check; QA independently exercises valid and missing/malformed/oversized/wrong-owner/wrong-mode seed, exact routing, redaction, shutdown, and fail-closed restart; main owns final checks and Git.",
  "exclusions": ["persistent store/config framework", "sudo", "push", "GitHub credentials", "audit", "release", "installer", "canary", "changes to existing client or daemon path"],
  "split_stop_rule": "Stop before DEV if the entrypoint plus one assembly boundary is incomplete, persistence is required, root ownership cannot be isolated, expected cumulative exceeds 855, or Lap 1 cannot produce a gate-ready candidate; split lifecycle from persistence and never defer seed/lifecycle into TASK-0008.",
  "measurement_lineage": "Record production/test SLOC, paired stage timing, active/wait, propagated retries, raw/effective classifications, source event IDs, null reasons, preflight exclusion, and ceil(observed non-preflight time * 1.20) time contingency.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0007/TASK.md"
}
```

## Purpose and owned boundary

Create the one privileged daemon entrypoint and one backend assembly/lifecycle
boundary. Wire the existing IPC and lease packages without changing their
security semantics. The existing `cmd/codex-authority` client remains the
non-privileged `ready`/`otp` client and is not converted into a daemon.

Production owns startup, readiness, signal-driven shutdown, and restart; a
root-owned bounded seed/config injection; in-memory lease/TOTP state
construction; and exact fixed routing of IPC `ready` and `otp` requests. The
seed is read once at startup from a root-owned mode-0600 bounded config input,
size-limited, schema-validated, copied into process state, and never accepted
from argv, environment, peer IPC, or logs. Missing, malformed, oversized,
wrong-owner, wrong-mode, construction, and restart-without-seed cases fail
closed and never report ready. Fixture-only seed helpers are test-only.

## Preflight and two-Lap delivery

Preflight confirms the TASK-0006 merge; exact existing `ready`/`otp` protocol;
client/daemon path separation; a temporary Unix-socket fixture; and whether
the isolated runner can simulate root ownership and mode without broad host
mutation. A preflight failure is `not_started`, spends no DEV lap, and is
excluded from timing.

Lap 1, after approved PLAN and QA_PLAN, changes only the two owned production
paths and their two owned test paths. Run:

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease
```

Capture existing-client regression plus startup, ready/OTP routing, shutdown,
and fail-closed restart candidate evidence. Lap 2 is independent REVIEW of the
focused suite plus the repository-native checks, followed by independent QA of
valid seed, missing/malformed/oversized/owner/mode failures, exact routing,
secret redaction, shutdown, and restart. Main owns final checks and Git; no
real system service is installed.

The repository-native full checks are additive to focused tests:

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

## Acceptance and exclusions

- A valid bounded seed starts the backend and permits only fixed `ready` and
  `otp` routing.
- Readiness is withheld for every invalid seed, failed construction, or
  unavailable/restarted backend.
- Signals shut down cleanly; restart without a valid seed remains fail closed.
- The seed never appears in argv, environment, peer input, logs, errors, or
  readiness output.
- Existing client behavior and existing daemon path remain unchanged.

This Task excludes persistence, a new config framework, sudo, push, GitHub
credentials, audit, release, installer, packaging, and canary work.

## Measurement, caps, and stop rule

The forecast is +90 production SLOC and cumulative 841; target cap 950,
90%-trigger 855, hard guard 1050. Forecast above 855 stops before DEV for
split/re-estimation and approved PLAN/QA_PLAN revision. A candidate above the
target cap or any hard-limit risk stops safely. Record production/test SLOC,
same-task/lap/stage/attempt start-terminal timing pairs, separate `active_ms`
and `wait_ms`, maximum propagated retries, raw classifications with source IDs
and nullable `superseded_by`, correction-validated effective classifications,
null reasons, and `ceil(observed_non_preflight_time * 1.20)` only for observable
time. Do not use SLOC/minute or another fixed throughput assumption.

Stop before DEV if entrypoint plus one assembly boundary is incomplete,
persistence is required, root ownership cannot be isolated, expected cumulative
exceeds 855, or Lap 1 cannot produce a gate-ready candidate. Split lifecycle
from any persistence requirement; never defer seed/lifecycle ownership into
TASK-0008. Classify predictable unchanged environment failures before retry.

## Gate and later reserve

Independent REVIEW PASS and QA PASS are required; a FAIL returns to its
responsible gate and never merges. Later audit/attestation/manual-canary work
remains non-executable until TASK-0012 independently passes REVIEW and QA and
main merges it.
