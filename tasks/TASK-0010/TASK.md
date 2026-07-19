# TASK-0010: local push policy and validation

**Depends on:** TASK-0009 (merged and PASS).

**Status:** planned and executable.

## Contract metadata

```json
{
  "id": "TASK-0010",
  "title": "local push policy and validation",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0009"],
  "expected_production_sloc": 130,
  "expected_cumulative_production_sloc": 1337,
  "target_cumulative_cap": 1425,
  "projected_cap_trigger_sloc": 1380,
  "hard_cumulative_guard": 1500,
  "production_paths": ["internal/push/policy.go", "internal/push/validate.go"],
  "test_paths": ["internal/push/policy_test.go", "internal/push/validate_test.go"],
  "entrypoint": "internal/push/policy.go",
  "fixture_elevation_needs": "Temporary worktree and local bare repository matrix; no network, credentials, child Git process, or elevation; denied cases must prove no fake transport boundary is crossed.",
  "lap_1": "After TASK-0009 PASS+merge and approved plans, validate exact configured repository, clean tree, main or task/TASK-* branch, and one-ref update; reject wrong repo/ref, dirty tree, force, tag, delete, multiple-ref, and ambiguous local Git state; run go test ./internal/push.",
  "lap_2": "Independent REVIEW runs focused tests and repository-native full check; QA independently mutates every local condition and proves each denial cannot cross the fake transport boundary; main owns Git.",
  "exclusions": ["token custody", "credential helper", "network transport", "Git child process", "backend registration", "sudo", "audit", "release", "installer", "canary"],
  "split_stop_rule": "Stop if validation requires remote state, credentials, or a second policy language; stop if forecast exceeds the post-reestimate stop 1380, the fixture cannot prove zero transport on denial, or Lap 1 is not review-ready.",
  "measurement_lineage": "Record exact local fixture state, stage pairs, active/wait, retries, raw/effective classifications, source IDs, null reasons, and time-only contingency; do not infer duration or size from throughput.",
  "later_reserve_eligibility": "Later audit/attestation/manual-canary reserve remains ineligible until TASK-0012 PASS+merge.",
  "contract_path": "tasks/TASK-0010/TASK.md"
}
```

## Purpose and owned boundary

Implement only local policy and validation in `internal/push/policy.go` and
`internal/push/validate.go`, with tests in the two named test paths. Validate
the exact configured repository, clean tree, `main` or `task/TASK-*` branch,
and one source/destination ref update. Reject wrong repository/ref, dirty
tree, force, tag, delete, multiple-ref, and ambiguous local Git state. This
Task owns no token, credential helper, network transport, or Git child process.

## Preflight and two-Lap delivery

Preflight requires merged TASK-0009 and approved plans, then prepares a
temporary worktree/local bare-repository matrix with no network or elevation.
The execution clock begins only after the fixture and zero-transport denial
observation are usable; a failed prerequisite is `not_started`.

Lap 1 implements the policy and validator and runs:

```sh
go test ./internal/push
```

The focused matrix mutates repository identity, ref, clean state, force/tag/
delete/multiple-ref shape, and ambiguous local Git state. Lap 2 is independent
REVIEW of the focused suite plus the repository-native full check, then QA
independently mutates each condition and proves denied requests cannot cross a
fake transport boundary. Main owns Git.

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

## Acceptance and exclusions

- Exactly the configured local repository, clean tree, permitted branch, and
  one-ref non-force update pass.
- Every wrong/dirty/force/tag/delete/multiple/ambiguous condition denies
  before any transport boundary.
- Local validation does not read remote state, credentials, or network data.
- The implementation remains idiomatic and does not compress source to fit a
  cap.

This Task excludes token custody, credential helpers, network transport, Git
child processes, backend registration, sudo, audit, release, installer, and
canary work.

## Measurement, caps, and stop rule

The forecast is +130 production SLOC and cumulative 1337; post-reestimate stop
1380, target cap 1425, hard guard 1500. Record exact fixture state, paired stage
timing, separate active/wait, propagated retries, raw/effective
classifications/source IDs, null reasons, preflight exclusion, and time-only
`ceil(observed_non_preflight_time * 1.20)` contingency. No SLOC/minute or
fixed throughput sizing is allowed.

Stop if validation requires remote state, credentials, or a second policy
language; if forecast exceeds 1380; if the fixture cannot prove zero transport
on denial; or if Lap 1 is not review-ready. Split discovery rather than
expanding this local boundary. Candidate target/hard overflow stops safely and
requires the approved shedding/replan order.

## Gate and later reserve

Independent REVIEW PASS and QA PASS are required; a FAIL returns to its
responsible gate and never merges. Later audit/attestation/manual-canary work
remains non-executable until TASK-0012 independently passes REVIEW and QA and
main merges it.
