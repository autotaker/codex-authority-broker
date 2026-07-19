# TASK-0008: sudo live check and no cache

**Depends on:** TASK-0015 (merged).

**Status:** planned and executable.

## Contract metadata

```json
{
  "id": "TASK-0008",
  "title": "sudo live check and no cache",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0015"],
  "expected_production_sloc": 120,
  "expected_cumulative_production_sloc": 1322,
  "target_cumulative_cap": 1350,
  "projected_cap_trigger_sloc": 1325,
  "hard_cumulative_guard": 1450,
  "production_paths": ["cmd/codex-authority-sudo/main.go", "deploy/sudo/codex-authority"],
  "test_paths": ["cmd/codex-authority-sudo/main_test.go", "deploy/sudo/codex-authority_test.go"],
  "entrypoint": "cmd/codex-authority-sudo/main.go",
  "fixture_elevation_needs": "Isolated Ubuntu sudo/PAM fixture, disposable dedicated identity, controlled clock/socket, and approved narrow elevation/rollback procedure; never mutate workstation sudo policy.",
  "lap_1": "After TASK-0015 merge and approved plans, implement per-invocation live request and declarative timestamp-cache disablement; run go test ./cmd/codex-authority-sudo ./internal/ipc plus the isolated sudo fixture covering allow, expiry, daemon unavailable/restart, malformed/unauthorized reply, and two consecutive invocations.",
  "lap_2": "Independent REVIEW runs focused tests and repository-native full check; QA uses the isolated elevated fixture to prove a live unexpired lease permits and every deny case fails closed with no cached reuse; main owns Git closure.",
  "exclusions": ["daemon/backend assembly", "push", "GitHub credentials", "rich audit", "release", "installer", "packaging", "canary", "real workstation policy mutation"],
  "split_stop_rule": "Classify not_started/environment_issue if the isolated elevated fixture or rollback proof is unavailable. Split before DEV if more than one client entrypoint and declarative policy is required, forecast exceeds the post-reestimate trigger 1325, or platform/PAM differences cannot fit two laps; never weaken live-per-call or no-cache behavior. A forecast or candidate above target 1350 stops for explicit replan and exact ordered shedding review; hard guard 1450 is absolute.",
  "measurement_lineage": "Record fixture/elevation waits separately, paired stage timing, active/wait, retries, raw/effective classifications, source IDs, null reasons, and time-only 20% contingency; no SLOC throughput sizing.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0008/TASK.md"
}
```

## Purpose and owned boundary

Provide the minimal `pam_exec`-compatible live-check client and a dedicated
identity's declarative no-cache sudo policy. Each invocation requests current
authority through TASK-0015's fixed IPC daemon; it never relies on sudo timestamp
caching. TASK-0013/TASK-0015 runtime and daemon assembly are consumed, not changed.

The production paths are exactly `cmd/codex-authority-sudo/main.go` and
`deploy/sudo/codex-authority`; tests are exactly the two paths in the metadata.
No real workstation sudo policy is installed or modified.

## Preflight and two-Lap delivery

Preflight requires merged TASK-0015 and approved PLAN and QA_PLAN; an isolated
Ubuntu sudo/PAM fixture; a disposable dedicated identity; controlled clock and
socket; required tools; and a narrow elevation/rollback procedure. A missing
fixture or rollback proof is `not_started/environment_issue`; preflight does
not consume DEV timing.

Lap 1 implements one live request per invocation and declarative timestamp-cache
disablement, then runs:

```sh
go test ./cmd/codex-authority-sudo ./internal/ipc
```

The isolated fixture covers an unexpired allow, expiry, daemon unavailable or
restart, malformed and unauthorized replies, and two consecutive invocations.
Lap 2 is independent REVIEW with focused tests and the repository-native full
check, followed by QA in the isolated elevated fixture. QA proves every deny
case fails closed and no cached authority is reused; main owns Git closure.

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

## Acceptance and exclusions

- A live unexpired lease permits the dedicated sudo check.
- Expired, unavailable, restarted, malformed, or unauthorized authority
  denies, without a stale timestamp grant.
- Two consecutive invocations each perform an independent live check.
- The policy declaratively disables sudo timestamp caching for this identity.
- No secret or authority decision is placed in argv, logs, or output beyond a
  bounded decision result.

This Task excludes daemon/backend assembly, push, GitHub credentials, rich
audit, release, installer, packaging, and canary work.

## Measurement, caps, and split/stop rule

The forecast is +120 production SLOC and cumulative 1322; post-reestimate trigger
1325, target cap 1350, hard guard 1450. Forecast above 1325 stops before DEV for
split/re-estimation and approved PLAN/QA_PLAN revision. Record elevation and
fixture waiting separately from active work; record paired stage timing,
active/wait, propagated retries, raw/effective classifications, source IDs,
null reasons, preflight exclusion, and time-only
`ceil(observed_non_preflight_time * 1.20)` contingency. Never size from
SLOC/minute or another fixed throughput assumption.

If the isolated elevated fixture or rollback proof is unavailable, stop as
`not_started/environment_issue`. Split before DEV if more than this single
client entrypoint and declarative policy are required, forecast exceeds 1325,
or platform/PAM differences cannot be covered in two laps. Do not weaken
live-per-call or no-cache behavior. A candidate above target or hard limits
stops safely.

## Gate and later reserve

Independent REVIEW PASS and QA PASS are required; a FAIL returns to its
responsible gate and never merges. Audit/attestation/manual-canary reserve is
non-executable until TASK-0012 independently passes REVIEW and QA and main
merges it.
