# TASK-0011: token custody and system-Git non-force push

**Depends on:** TASK-0010 (merged and PASS).

**Status:** planned and executable.

## Contract metadata

```json
{
  "id": "TASK-0011",
  "title": "token custody and system-Git non-force push",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0010"],
  "expected_production_sloc": 220,
  "expected_cumulative_production_sloc": 1311,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1350,
  "hard_cumulative_guard": 1650,
  "production_paths": ["cmd/codex-authority-push/main.go", "internal/ipc/protocol.go", "internal/push/custody.go", "internal/push/system_git.go", "internal/backend/push_registration.go"],
  "test_paths": ["cmd/codex-authority-push/main_test.go", "internal/ipc/protocol_test.go", "internal/push/custody_test.go", "internal/push/system_git_test.go", "internal/backend/push_registration_test.go"],
  "entrypoint": "cmd/codex-authority-push/main.go",
  "fixture_elevation_needs": "TASK-0007 handler seam and dedicated caller UID, local bare remote, fake short-lived GitHub App token provider, system-Git binary, credential-capture sentinel, live-lease fixture, no network and no elevation.",
  "lap_1": "After TASK-0010 PASS+merge and approved plans, implement the bounded caller, OperationPush/PushRequest admission, backend registration and UID/live-lease/policy gates, in-memory token custody, and one system-Git single-ref non-force path; run go test ./cmd/codex-authority ./cmd/codex-authority-broker ./cmd/codex-authority-push ./internal/ipc ./internal/push ./internal/backend.",
  "lap_2": "Independent REVIEW runs the focused suite, malformed/unknown-operation and credential-capture mutations, then repository-native full check; QA proves exactly one authorized push handler with correct UID, live lease, and TASK-0010 policy, and denies old client, wrong UID, malformed/expired authority, leak, force, and ambiguity cases before custody/Git; main owns Git.",
  "exclusions": ["changes to cmd/codex-authority/main.go", "changes to cmd/codex-authority-broker/main.go", "generic IPC commands", "arbitrary refspec", "remote-OID prefetch", "force/tag/delete push", "sudo", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Stop before DEV if caller schema cannot stay bounded, SO_PEERCRED cannot distinguish configured UID, token injection leaks through a named channel, system Git cannot be captured deterministically, the five-unit forecast exceeds 1350, or remote OID/race diagnostics are required; shed optional diagnostics in order or split without weakening authorization, custody, schema, or non-force behavior.",
  "measurement_lineage": "Forecast allocation is 35 caller + 45 protocol/schema + 40 registration/gates + 55 custody + 45 system-Git/redaction = 220, not throughput. Record stage pairs, active/wait, retries, raw/effective classifications, source IDs, null reasons, preflight exclusion, and time-only contingency.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0011/TASK.md"
}
```

## Purpose and owned boundary

This Task owns the only supported local restricted-push caller,
`cmd/codex-authority-push/main.go`; one bounded `push` operation and
`PushRequest` admission in `internal/ipc/protocol.go`; bounded in-memory token
custody; one system-Git single-ref non-force path; and
`internal/backend/push_registration.go` attaching the route through the
already-merged TASK-0007 handler seam. Tests are exactly the five named test
paths. Neither `cmd/codex-authority/main.go` nor
`cmd/codex-authority-broker/main.go` changes.

The caller accepts only configured repository identity and one permitted local
source/destination ref intent. It exposes no token, force, tag, delete,
arbitrary refspec, remote-command, credential-environment, or generic IPC
operation option. It strictly constructs `ipc.OperationPush` with bounded
repository and single-ref fields required by TASK-0010. Unknown, missing,
duplicate-equivalent, oversized, malformed, force/tag/delete/multiple-ref, or
noncanonical fields deny before backend dispatch.

The protocol admits exactly `ready`, `otp`, and `push`. Backend admission
requires the dedicated UID from TASK-0007 root-owned configuration, verified by
the existing fail-closed `SO_PEERCRED` boundary; a live lease; and TASK-0010
policy PASS. Wrong UID, absent/expired lease, invalid policy/schema,
unavailable registration, unknown operation, or malformed payload denies before
token retrieval or Git. Token material is absent from argv, environment,
logs, output, errors, and credential-helper storage.

## Preflight and two-Lap delivery

Preflight requires merged TASK-0010, stable TASK-0007 handler seam and caller
UID, local bare remote, fake short-lived GitHub App token provider, system-Git
binary, credential-capture sentinel, live-lease fixture, no network, and no
elevation. A missing prerequisite is `not_started` and excluded from timing.

Lap 1 implements the bounded caller, schema/admission, registration/gates,
custody, and one system-Git single-ref non-force path, then runs:

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./cmd/codex-authority-push ./internal/ipc ./internal/push ./internal/backend
```

Evidence covers old-client rejection, schema bounds, authorized reachability,
wrong-UID/pre-dispatch denial, token-channel absence, and ambiguity/no-force
retry. Lap 2 is independent REVIEW of focused tests, malformed/unknown-
operation and capture-sentinel mutations, and the full check. QA independently
proves exactly one push handler only with the correct UID, live lease, and
TASK-0010 policy; old CLI, wrong UID, malformed/expired authority, leakage,
force, and ambiguity deny before custody/Git. Main owns Git.

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

## Measurement, allocation, and stop rule

The forecast is +220 production SLOC and cumulative 1311; target cap 1500,
90%-trigger 1350, hard guard 1650. The boundary allocation is 35 caller + 45
protocol/schema + 40 registration/gates + 55 custody + 45 system-Git/redaction
= 220; it is not a throughput estimate. Record paired stage timing,
active/wait, retries, raw/effective classifications/source IDs, null reasons,
preflight exclusion, and time-only 20% contingency.

Stop before DEV if the caller cannot stay within one bounded schema,
`SO_PEERCRED` cannot distinguish the configured UID, token injection cannot
keep secrets out of every named channel, system Git cannot be captured
deterministically, the five-unit forecast exceeds 1350, or remote OID/race
diagnostics are required. Shed optional diagnostics only in the approved order
or split; never weaken caller authorization, schema admission, custody, or
non-force-only behavior. Candidate target/hard overflow stops safely.

## Exclusions and gate

This Task excludes changes to the existing ready/OTP client or broker
entrypoint, generic IPC commands, arbitrary refspecs, remote-OID prefetch,
force/tag/delete pushes, sudo, audit, release, installer, and canary work.

Independent REVIEW PASS and QA PASS are required; a FAIL returns to its
responsible gate and never merges. Later audit/attestation/manual-canary work
remains non-executable until TASK-0012 independently passes REVIEW and QA and
main merges it.
