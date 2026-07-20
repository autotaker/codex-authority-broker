# TASK-0020: artifact-only manual canary and exact rollback evidence

**Depends on:** TASK-0019 completed with live artifact and provenance PASS.

**Status:** planned safety-contract and live operational-evidence Task; no
product DEV is authorized.

## Contract metadata

```json
{
  "id": "TASK-0020",
  "title": "artifact-only manual canary and exact rollback evidence",
  "status": "planned",
  "executable": true,
  "work_classification": "safety_contract_and_live_operational_evidence",
  "depends_on": ["TASK-0019"],
  "baseline_production_sloc": 1478,
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1478,
  "target_cumulative_cap": 1500,
  "hard_cumulative_guard": 1800,
  "production_paths": [],
  "test_paths": [],
  "evidence_paths": ["tasks/TASK-0020/TASK.md", "tasks/TASK-0020/PLAN.md", "tasks/TASK-0020/QA_PLAN.md", "tasks/TASK-0020/PLAN_REVIEW.md", "tasks/TASK-0020/STAGE_RUNBOOK.sh", "tasks/TASK-0020/CANARY_RUNBOOK.sh", "tasks/TASK-0020/CANARY_RESULT.md", "tasks/TASK-0020/EVIDENCE_REVIEW.md", "backlog.json"],
  "artifact_run": 29720021660,
  "artifact_source_commit": "09487b104f32cad23a695ec3f1a0c7e7a68e6163",
  "artifact_sha256": "5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd",
  "fixture_elevation_needs": "Existing noninteractive administrator authority with no new privilege widening; execution is limited by Main approval to the independently reviewed, digest-pinned unshare launcher and fixed staging path. Fixture policy, identities, seed, socket, timestamps, logs and installation paths remain tmpfs-only; exact outer staging rollback is mandatory.",
  "exclusions": ["all product/test/workflow/dependency changes", "source or local-build substitution", "installer implementation", "persistent installation", "live workstation PAM/sudoers/identity mutation", "GitHub push capability", "secret-bearing evidence", "counted product Lap"],
  "split_stop_rule": "Do not start the live fixture unless artifact identity, noninteractive namespace elevation, exact host pre-state capture, cleanup commands, and rollback comparison are ready. Any product behavior defect stops this Task and requires a separately approved product Task; never patch the artifact or weaken a negative case in place.",
  "contract_path": "tasks/TASK-0020/TASK.md"
}
```

## Fixed input and route

The canary consumes only the artifact named `codex-authority-linux-amd64`
from successful main run `29720021660`, bound to commit `09487b1`, release
workflow `.github/workflows/release.yml`, and `refs/heads/main`. Its archive
SHA-256 is
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`.
Before host work, repeat the exact outer-file, seven-member, six-checksum, and
GitHub attestation verification. A local build, source checkout binary, other
run, mutable tag, or unverified copy cannot substitute.

This conversion defines a safety contract and live operational evidence; it
does not change product artifacts. It requires an approved PLAN, independent
TASK-first QA_PLAN, and independent plan review before Main operates the
fixture. It creates no product DEV, counted Lap, product REVIEW_RESULT, or
product QA_RESULT. Main performs the authorized host operation; an independent
evidence Reviewer evaluates the resulting `CANARY_RESULT.md` against the
approved matrix without editing product or Git.

## Isolation and rollback boundary

The host already grants the operator broad noninteractive administrator
authority; TASK-0020 does not widen or present that pre-existing policy as a
capability boundary. Main authorizes only the independently reviewed,
digest-pinned fixed launcher invocation, records the existing policy as
pre-state, and proves that the executed program touches only the declared
staging tree before entering its private boundary. The fixture uses a private
mount/PID namespace with tmpfs-backed `/etc`,
`/run`, fixture installation locations, logs, and sudo timestamps. It creates
only disposable dedicated and distinct nonroot identities inside the fixture;
the dedicated numeric UID and GID are equal and nonzero. The seed is
fixture-only, root-owned mode 0600, never printed, and removed at teardown.
PAM and sudoers changes exist only inside the namespace and use the verified
artifact's declarative files plus the narrow fixture command grant required
to execute actual sudo. The broker, CLI, and PAM helper must all come from the
verified archive.

Before namespace entry, record hashes, metadata, and directory listings for
host passwd/group/shadow/gshadow, PAM, sudoers/sudoers.d, relevant `/run` and
fixture installation parents. After teardown, require byte-identical hashes,
metadata and listings; no fixture process, mount, socket, timestamp, log,
identity, seed, binary, or policy file may remain. Unknown cleanup or any host
pre-state difference is a blocking `environment_issue`, not a partial PASS.

## Live acceptance

1. Verify the exact artifact/provenance binding again, extract only into the
   private fixture, validate PAM/sudoers/systemd declarative syntax, and start
   the artifact broker against its fixed socket and root-only seed.
2. Use a real independently generated TOTP for the fixture seed. The dedicated
   identity completes real readiness and OTP activation; root and the distinct
   identity are rejected by the real SO_PEERCRED boundary. No OTP or seed is
   recorded.
3. During the active lease, two separate actual sudo invocations through real
   PAM and the artifact helper succeed. Each must cause one fresh authorize
   request, with no sudo timestamp reuse or locally cached permit.
4. After the real lease deadline passes, a fresh actual sudo invocation fails
   closed and makes one fresh authorize request. No timestamp or prior helper
   process may convert the denial to success.
5. After a new real TOTP activation, actual sudo succeeds again. Stopping the
   broker then makes a fresh sudo fail. Starting a fresh broker with the same
   fixture installation but no active lease also fails; only another real
   readiness/TOTP sequence restores an allow.
6. Audit evidence for the exercised allows and denials has the exact five-field
   schema and expected actor/scope/result/expiry relationships. Evidence records
   only bounded counts, types and digests; it contains no request payload, OTP,
   seed, token, key, environment, lease identifier or internal error text.
7. Teardown proves the exact rollback boundary above. Canary PASS requires all
   positive, negative, redaction, cleanup and host-equality checks; there is no
   partial, unit-only, source-built, or prior-TASK evidence substitution.

## Failure classification and completion

- `permission_issue`: required narrow noninteractive namespace elevation is
  not available.
- `environment_issue`: namespace, tmpfs, PAM/sudo, TOTP generator, natural
  expiry, cleanup, or exact rollback cannot execute safely and completely.
- `implementation_defect` or `regression`: the verified artifact violates a
  fixed product or previously accepted behavior. Stop and open a normal
  product Task before any fix.
- `requirement_gap` or `qa_plan_defect`: this contract or its fixture cannot
  prove a mandatory v1 property without changing the safety boundary.

Completion requires `CANARY_RESULT.md`, independent `EVIDENCE_REVIEW.md` PASS,
clean host rollback, unchanged product SLOC 1478, Main-owned evidence Git, and
a final v1 requirement-by-requirement audit. Secret values and full raw command
output are never evidence.
