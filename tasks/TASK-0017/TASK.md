# TASK-0017: dedicated socket ownership and PAM peer identity handoff

**Depends on:** TASK-0008, TASK-0015, and TASK-0016 merged and PASS; the
TASK-0009 measurement/replan merged.

**Status:** planned v1 blocker. TASK-0012 cannot start until this Task passes
independent REVIEW and QA and merges.

## Contract metadata

```json
{
  "id": "TASK-0017",
  "title": "dedicated socket ownership and PAM peer identity handoff",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0008", "TASK-0015", "TASK-0016"],
  "baseline_production_sloc": 1253,
  "expected_production_sloc": 55,
  "expected_cumulative_production_sloc": 1308,
  "target_cumulative_cap": 1400,
  "projected_cap_trigger_sloc": 1350,
  "hard_cumulative_guard": 1450,
  "production_paths": ["cmd/codex-authority-broker/main.go", "cmd/codex-authority-sudo/main.go"],
  "test_paths": ["cmd/codex-authority-broker/main_test.go", "cmd/codex-authority-sudo/main_test.go", "deploy/sudo/codex-authority_test.go"],
  "entrypoint": "cmd/codex-authority-sudo/main.go",
  "fixture_elevation_needs": "Private mount namespace with copied tmpfs /etc and tmpfs /run, a disposable root-owned mode-0600 seed for a dedicated nonzero equal UID/GID identity, fixture-only PAM and sudoers, real Authenticator TOTP, real broker/socket/PAM/sudo, exact host hash/list rollback, and no live workstation mutation.",
  "lap_1": "With approved PLAN and TASK-first QA_PLAN, make the broker create the fixed socket for the configured dedicated nonzero equal UID/GID and make the PAM helper derive that identity only from the fixed root-directory socket metadata, permanently drop supplementary groups/GID/UID before its one authorize call, then run focused and isolated real TOTP/PAM/sudo tests.",
  "lap_2": "Exceptional only for one or two bounded findings requiring no redesign. Independent REVIEW and QA each verify socket ownership and replacement resistance, irreversible identity drop before transport, SO_PEERCRED denial, live/no-cache sudo, broker-stop/restart denial, redaction, rollback, full/race/vet/format/diff checks, and Main-only Git.",
  "exclusions": ["trusting PAM_RUSER, PAM_USER, environment, stdin, or caller-selected UID/GID", "allowing UID 0", "multiple authority sockets", "push or GitHub credentials", "installer or live-host PAM mutation", "audit", "attestation", "release", "canary"],
  "split_stop_rule": "Stop before DEV if identity cannot be derived from a fixed-path socket beneath a non-writable root-owned directory, the dedicated identity cannot require equal nonzero UID/GID, groups/GID/UID cannot be irreversibly dropped before connect, broker and PAM peers need different authority rules, a second socket or seed disclosure is required, forecast exceeds 1350 cumulative, or the real isolated E2E/rollback fixture is unavailable; never bypass SO_PEERCRED or accept root as the authority client.",
  "measurement_lineage": "The post-TASK-0009 real E2E accepted ready/OTP/authorize for UID 1000 only after a fixture chown; production broker omitted ipc.Config.Access. Real PAM then launched the helper as UID 0, which the single nonzero allowed_uid correctly rejected. A root-peer retry could not start because seed parsing correctly rejects UID 0. Forecast +55 covers broker Access wiring and bounded socket-metadata identity-drop logic, not test throughput.",
  "later_reserve_eligibility": "TASK-0012 and later audit/attestation/manual-canary work remain blocked until this Task passes, merges, and its actual SLOC/E2E evidence is available.",
  "contract_path": "tasks/TASK-0017/TASK.md"
}
```

## Purpose and discovered requirement gap

The real isolated E2E proved that the v1 components are not yet deployably
connected. The root broker creates a root-only socket unless its existing
`ipc.Config.Access` is populated, while the dedicated UID must issue `ready`
and `otp`. After a valid TOTP creates a lease, `sudo` invokes PAM and the
current helper reaches the broker as UID 0; the server correctly rejects that
peer and the seed schema correctly forbids configuring root as `allowed_uid`.

This Task closes both sides atomically. The broker must provision the fixed
socket for the configured dedicated identity. The PAM helper must use no PAM
or caller-controlled identity input: it inspects the fixed socket beneath the
root-owned, non-user-writable `/run` directory, requires a socket with equal
nonzero numeric owner/group matching the deployment invariant, clears all
supplementary groups, drops GID and UID irreversibly, and only then performs
its existing single payload-free `authorize` call. Any metadata, drop, race,
transport, response, expiry, cancellation, or broker lifecycle failure denies.

## Required scenarios

- A dedicated UID can use the real CLI and socket for `ready` and OTP without
  an external `chown` bridge; root and a distinct UID are rejected by
  SO_PEERCRED.
- Real Authenticator OTP activates one 300-second process-local lease. Actual
  PAM and actual `sudo /usr/bin/true` succeed twice during the lease, with one
  fresh authorize call per invocation and no sudo timestamp reuse.
- Exact lease expiry, broker stop, and fresh broker restart each make the next
  actual sudo fail closed. A prior allow never survives in PAM, sudo, socket,
  helper, or daemon state.
- The helper rejects missing/non-socket/symlink/replaced socket paths,
  UID/GID zero or mismatch, failed group/GID/UID drop, and any attempt to use
  environment, stdin, PAM identity variables, or caller-selected identity.
- Allowed output remains empty and denials remain bounded and redacted. Seed,
  TOTP, lease, UID metadata, and internal errors never enter logs or results.
- The private namespace fixture validates PAM/sudoers syntax and proves exact
  host passwd/group/shadow/gshadow/sudoers/PAM hashes and directory listings
  are unchanged after exit.

## v1 boundary

GitHub push is not part of this Task or v1. TASK-0010 and TASK-0011 remain
`deferred-v2`, `executable:false`, and contribute zero v1 production SLOC.
The remaining v1 sequence is TASK-0017 followed by the zero-SLOC TASK-0012
measurement gate; later milestones remain blocked.
