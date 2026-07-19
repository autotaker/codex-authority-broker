# TASK-0018: minimal audit and source-free attested artifact

**Depends on:** TASK-0012 merged with independent REVIEW and QA PASS.

**Status:** planned v1 blocker. The manual canary/rollback milestone cannot start
until this Task passes independent REVIEW and QA, merges, and its live
attestation verification succeeds.

## Contract metadata

```json
{
  "id": "TASK-0018",
  "title": "minimal audit and source-free attested artifact",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0012"],
  "baseline_production_sloc": 1407,
  "expected_production_sloc": 90,
  "expected_cumulative_production_sloc": 1497,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1497,
  "hard_cumulative_guard": 1800,
  "production_paths": ["cmd/codex-authority-broker/main.go", "internal/backend/runtime.go", "internal/ipc/server_linux.go", "internal/lease/lease.go"],
  "test_paths": ["cmd/codex-authority-broker/main_test.go", "internal/backend/runtime_test.go", "internal/ipc/server_linux_test.go", "internal/lease/lease_test.go", ".github/workflows/release.yml", "deploy/pam/codex-authority", "deploy/systemd/codex-authority-broker.service"],
  "entrypoint": "cmd/codex-authority-broker/main.go",
  "fixture_elevation_needs": "Deterministic local audit and packaging tests require no elevation. Live postmerge GitHub Actions attestation and gh attestation verification require repository Actions access and network, but no secret material. The later private-namespace canary owns PAM/sudo elevation.",
  "lap_1": "With approved PLAN and TASK-first QA_PLAN, add bounded fail-closed authority audit events, package only the three binaries plus declarative deployment files and checksums, attest the archive in GitHub Actions, and run focused/full local checks plus independent REVIEW and QA.",
  "lap_2": "Exceptional only for one or two bounded findings requiring no redesign. Merge, run the workflow on main, download the artifact, verify its GitHub provenance and source-free manifest, and record postmerge evidence before enabling the manual canary/rollback Task.",
  "exclusions": ["GitHub push capability or token custody", "generic logging framework", "remote log transport or retention service", "installer", "automatic deployment", "live-host PAM or sudo mutation", "manual canary execution", "release publishing beyond the attested workflow artifact"],
  "split_stop_rule": "Stop before or during DEV if a readable implementation exceeds +90 production SLOC or cumulative 1497, any mandatory audit field or fail-closed behavior must be weakened, payload/OTP/seed/token/private-key data could enter audit, the archive cannot be proven source-free, or attestation cannot be independently verified. The 1800 hard limit is not implementation allowance; never compress code to fit.",
  "measurement_lineage": "TASK-0012 measured the merged v1 core at 1407 production SLOC. This Task may add at most 90 readable production lines for cumulative 1497, leaving three lines to the 1500 target and 303 to the 1800 hard limit. Workflows, operator documents, tests, and declarative PAM/systemd/sudoers files are excluded by the canonical production-SLOC definition.",
  "later_reserve_eligibility": "Only this Task's independent REVIEW/QA PASS, main merge, successful live workflow, source-free manifest check, and gh attestation verification enable the manual canary/rollback milestone.",
  "contract_path": "tasks/TASK-0018/TASK.md"
}
```

## Required audit boundary

Each accepted broker connection receives a fresh correlation ID. Every valid
authority operation reaching the backend emits exactly one bounded JSON event
to the broker's operational output with:

- correlation ID;
- numeric SO_PEERCRED actor UID;
- fixed scope (`ready`, `otp`, or `authorize`);
- final result (`allow` or `deny`);
- the immutable lease expiry for an allowed OTP/authorize decision, otherwise
  an explicit null expiry.

The event must never contain request payloads, OTP values, the seed, tokens,
keys, environment contents, or internal error strings. Malformed and
unauthorized connections may be silently denied before a valid audit context
exists. An audit write failure must not publish an allow or leave a usable
new lease; subsequent authority must fail closed. Audit is synchronous at the
decision boundary and does not add a cache or extend a lease.

## Source-free artifact and attestation

The GitHub Actions workflow builds the three Linux binaries and creates one
archive containing only those binaries, the declarative PAM/systemd/sudoers
files, and checksums. A deterministic manifest check rejects Go source,
repository metadata, task/evidence files, seeds, credentials, and unexpected
paths before upload. The workflow uses GitHub's official build-provenance
attestation action with least-required permissions and uploads the archive and
checksums as the workflow artifact.

After merge, Main runs the workflow on `main`, downloads the artifact, verifies
its provenance for `autotaker/codex-authority-broker` with `gh attestation
verify`, rechecks checksums and the exact archive allowlist, and records only
redacted evidence. A locally built archive or a successful unit test cannot
substitute for this live verification.

## Required scenarios

- Allowed and denied ready/OTP/authorize decisions produce one schema-valid,
  bounded event with the correct actor, scope, result, correlation separation,
  and expiry semantics.
- Concurrent requests do not reuse correlation IDs or associate one request's
  actor/result/expiry with another request.
- Sink failure denies and renders any just-created lease unusable; it never
  exposes a successful response without the matching audit event.
- Existing expiry, cancellation, shutdown, SO_PEERCRED, protocol size,
  no-cache sudo, seed redaction, race, and full-suite behavior remain intact.
- The workflow's archive is reproducibly allowlisted and source-free, uses
  pinned official actions, carries checksums, and produces independently
  verifiable GitHub provenance on the merged commit.

## v1 boundary

TASK-0010 and TASK-0011 remain `deferred-v2`, `executable:false`, and zero v1
SLOC. This Task does not install or deploy anything. The final v1 gate remains
a separate manual canary and exact rollback using this verified artifact.
