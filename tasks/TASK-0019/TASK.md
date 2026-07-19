# TASK-0019: minimal audit and source-free attested artifact, replanned

**Depends on:** TASK-0012 merged/PASS and TASK-0018 superseded with no product
candidate.

**Status:** planned v1 blocker.

## Contract metadata

```json
{
  "id": "TASK-0019",
  "title": "minimal audit and source-free attested artifact, replanned",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0012", "TASK-0018"],
  "baseline_production_sloc": 1407,
  "expected_production_sloc": 90,
  "expected_cumulative_production_sloc": 1497,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1497,
  "hard_cumulative_guard": 1800,
  "production_paths": ["internal/backend/runtime.go", "internal/ipc/server_linux.go", "internal/lease/lease.go"],
  "test_paths": ["cmd/codex-authority-broker/main_test.go", "internal/backend/runtime_test.go", "internal/ipc/server_linux_test.go", "internal/lease/lease_test.go", ".github/workflows/release.yml", "deploy/pam/codex-authority", "deploy/systemd/codex-authority-broker.service"],
  "entrypoint": "cmd/codex-authority-broker/main.go",
  "fixture_elevation_needs": "Local audit/package tests need a socket-capable environment but no elevation. Postmerge GitHub Actions and gh attestation verification require network and repository Actions access. The later isolated manual canary owns PAM/sudo elevation.",
  "lap_1": "Implement the measured single-mutex audit boundary with a fresh process-local correlation ID and permanent Runtime closure on sink failure; add the exact source-free attested workflow/declarative files and deterministic tests; pass independent REVIEW/QA and Main Git.",
  "lap_2": "Exceptional only for one or two bounded findings needing no redesign; otherwise stop. Postmerge live provenance verification remains mandatory before completion.",
  "exclusions": ["random or globally unique correlation format", "GitHub push/token custody", "generic or remote logging", "installer or automatic deployment", "live-host mutation", "manual canary", "GitHub Release publication"],
  "split_stop_rule": "Stop if readable production exceeds +90/cumulative1497, any required field or fail-closed decision weakens, sensitive input can enter audit, archive allowlisting/attestation cannot be proven, or another production path is needed. Do not compress; 1800 is not implementation allowance.",
  "measurement_lineage": "TASK-0018 measured a +102 readable draft: backend62, IPC27, lease13. Ordered shedding item3 removes only the rich random fixed-width correlation format while retaining a fresh process-local correlation ID; permanent Runtime closure removes separate lease-internal invalidation. Forecast <=90/cumulative<=1497.",
  "later_reserve_eligibility": "Independent REVIEW/QA PASS, merge, successful main-bound workflow, exact source-free/checksum verification, and gh provenance verification enable the manual canary/rollback milestone.",
  "contract_path": "tasks/TASK-0019/TASK.md"
}
```

## Required audit boundary

For each admitted backend operation the fixed JSON event has exactly:
`correlation_id`, numeric SO_PEERCRED `actor_uid`, fixed `scope`, final
`result`, and `lease_expiry`. The correlation ID is a fresh, nonzero,
lowercase hexadecimal process-local sequence; journald process metadata gives
the boot/process namespace. It need not be random or globally unique.

Ready and every deny use null expiry. Allowed OTP and authorize use the same
immutable UTC RFC3339Nano lease deadline. Audit contains no request payload,
OTP, seed, token, key, environment, lease identifier, or internal error.

One Runtime mutex linearizes handler decision, deadline capture, one bounded
audit write, and response publication. Sink error/short write denies the
current request and permanently closes Runtime before unlocking; therefore a
just-created OTP lease and every prior lease become unusable to all later
authority calls. Separate mutation of inaccessible lease internals is not
required. Cancellation observed before the write is audited as deny; the
successful write is the final decision linearization point.

## Source-free artifact

The official full-SHA-pinned GitHub workflow builds exactly the broker, CLI,
and sudo helper for Linux amd64, packages only those binaries plus declarative
PAM/sudoers/systemd files and exact checksums, rejects every unexpected/source/
secret-bearing member, creates GitHub build provenance for the archive digest,
and uploads the archive/checksum artifact. No installer or release is added.

Main must after merge verify the exact manifest and checksums and run `gh
attestation verify` for `autotaker/codex-authority-broker`, bound to the merged
main workflow/commit. Local tests cannot substitute.

## Acceptance

- Allowed/denied ready, OTP, and authorize produce one exact five-field event
  with correct actor/scope/result/expiry and no sensitive sentinel.
- Concurrent accepted requests have distinct IDs and no tuple crossing; OTP
  authority is unavailable until its allow event succeeds.
- Sink failure yields no allow and permanently makes current/future authority
  unusable without retry or compensating event.
- Existing IPC, lease, lifecycle, cancellation, no-cache, and redaction tests,
  focused audit/package tests, race/full/vet/diff checks all pass.
- Production remains readable and cumulative <=1497; the archive is exact,
  source-free, checksummed, and live-provenance verified after merge.
