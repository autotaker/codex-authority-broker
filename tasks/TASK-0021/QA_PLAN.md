# TASK-0021 QA PLAN — production installer, recovery, uninstall, and manuals

## Status and authority

This QA plan is derived only from `tasks/TASK-0021/TASK.md` and its matching `backlog.json` entry. It was written before and independently of the implementation PLAN.

TASK-0021 remains `planned` and `executable:false`. This document does not authorize DEV, privileged installation, or host mutation. Execution requires approved TASK, PLAN, this TASK-first QA_PLAN, and independent plan review.

Target platform:

- Ubuntu 24.04 LTS amd64
- real systemd, PAM, sudo, journald, passwd/group databases
- canonical production user and group: `coding-agent`
- disposable VM with snapshot/restore, scripted reboot/power-loss injection, and out-of-band root console
- the user owns and operates this VM from Main's digest-bound manual E2E runbook; no VM credential, seed, QR, or TOTP is returned

All cases are P0. Any mandatory omission prevents PASS.

## Fixed security contract

The accepted production configuration must establish all of the following:

- `coding-agent` is a nonzero dedicated identity with equal numeric UID and GID.
- Product- or vendor-named identities such as `codex-fixture` are fixture-only.
- Only `coding-agent` receives full sudo.
- Every `coding-agent` sudo invocation uses the dedicated `codex-authority` PAM service.
- `timestamp_timeout=0` disables sudo timestamp reuse.
- `NOPASSWD` is absent.
- A fresh payload-free broker authorization occurs for every sudo invocation.
- The lease controls only entry into root authority. It does not contain or reverse effects created after root acquisition.
- Other users’ sudo/PAM behavior and the independent root recovery route remain usable.
- TASK-0010/TASK-0011 GitHub push authority remains excluded.

The expected sudoers semantics are equivalent to:

```text
Defaults:coding-agent pam_service=codex-authority
Defaults:coding-agent pam_login_service=codex-authority
Defaults:coding-agent pam_askpass_service=codex-authority
Defaults:coding-agent timestamp_timeout=0
coding-agent ALL=(root:root) PASSWD: ALL
```

The candidate may use a syntactically equivalent, independently justified representation. It must not use global defaults, affect another user, use `NOPASSWD`, or provide a route around the dedicated PAM service.

## Fixture and snapshot requirements

Before testing, independently record:

- exact Ubuntu release, architecture, kernel, systemd, PAM, sudo, and journald versions;
- VM image digest and snapshot identifier;
- boot identifier and UTC start time;
- passwd, group, shadow, and gshadow metadata and cryptographic digests;
- `/etc/pam.d`, `/etc/sudoers`, `/etc/sudoers.d`, relevant systemd directories, installation parents, seed parent, `/run`, and installer recovery-state paths;
- enabled/running unit state;
- existing `coding-agent` or colliding identity state;
- root recovery procedure and a successful out-of-band recovery probe;
- shared journal configuration, retention policy, and a cursor proving pre-existing entries remain readable.

Raw shadow data, seed material, QR content, TOTP values, environment dumps, and unrestricted journal output are never evidence.

Use separate restored snapshots for at least:

1. clean installation;
2. compatible identity reuse;
3. identity/name/UID/GID collision cases;
4. each controlled failure family;
5. each abrupt interruption or reboot point;
6. rotation;
7. uninstall and reinstall.

A workstation, shared server, or irreplaceable host is not an acceptable fixture.

## Evidence format

Each case retains only a bounded record containing case ID, candidate/artifact/installer digests, VM snapshot and boot labels, UTC start/end, bounded command/probe name, exit status, expected and observed count, PASS/FAIL, and a digest of secret-free normalized evidence where needed.

Evidence must not retain seed or QR content, raw TOTP, request/response payloads, raw audit JSON, credentials, tokens, keys, environment dumps, shell history, PAM conversation, unrestricted stdout/stderr or journal output, or secret-derived hashes that could become a reusable verifier.

## Acceptance matrix

| ID | Case | Required procedure and oracle | Failure classification |
| --- | --- | --- | --- |
| Q21-01 | Contract and scope | Compare TASK fenced JSON with backlog exactly; require planned/non-executable baseline1478, installer exception, dependency, and unchanged deferred-v2 scope. | `requirement_gap` |
| Q21-02 | Supported platform | Prove exact Ubuntu 24.04 LTS amd64 and genuine PAM/sudo syntax before mutation. | `environment_issue` / `requirement_gap` |
| Q21-03 | Installer trust chain | Pre-merge, bind deterministic functional archive to candidate commit/tree/checksum and reject malformed inputs without claiming provenance. Post-merge, bind the main artifact to repository, signer workflow, main ref, merge source, subject digest and identical payload checksums; reject every wrong binding, extra/link/traversal member, and source substitute. | `implementation_defect` |
| Q21-04 | Privileged entry boundary | Pre-merge fixture phase uses copy-only then root literal candidate-digest/manifest/hash/exec without provenance claim. Post-merge production phase uses copy-only then root re-attest/extract/hash/exec on a fresh snapshot. Both use fixed staging, trusted tools, empty environment allowlist and bounded argv, and reject TOCTOU/changed bytes/path/options/environment/shell evaluation. | `implementation_defect` |
| Q21-05 | Clean enrollment and install | Complete preflight, secret-safe QR enrollment, transaction, verifier, service start; require exact files/modes/state and no undeclared mutation. | `implementation_defect` |
| Q21-06 | Secret-safe enrollment | Keep secret/URI out of argv, environment, history, capture, logs, journal, evidence, temp files, process listings, repository; QR only on controlling terminal. | `security_defect` |
| Q21-07 | Seed boundary | Require root regular 0600 strict bounded schema/safe traversal; reject owner/mode/schema/base64/UID/path replacement defects. | `security_defect` |
| Q21-08 | New identity lifecycle | Create exactly `coding-agent` equal nonzero UID/GID; uninstall removes only installer-created identity. | `implementation_defect` |
| Q21-09 | Compatible identity reuse | Preserve exact identity/home state; require equal IDs, matching primary group, no supplementary group/capability, locked password, exact home/shell, and no pre-existing grant/timestamp, otherwise reject before mutation. | `implementation_defect` |
| Q21-10 | Identity collision rejection | Reject user-only, group-only, unequal/colliding/root IDs, wrong primary group, and production vendor/tool names before mutation. | `security_defect` |
| Q21-11 | PAM and sudo isolation | Genuine validation; coding-agent-only root full sudo, per-user pam_service/pam_login_service/pam_askpass_service/no-cache, no NOPASSWD/!authenticate/exempt_group/command-Defaults/global/other grant, unchanged recovery/control user. | `security_defect` |
| Q21-12 | Real full-sudo allow | Fresh activation then distinct harmless actual `sudo`, `sudo -i`, `sudo -A`, `sudo -s`, `sudo -u root`, and `sudo -g root` calls; euid0 and one fresh authorize each, no cache/material reuse. | `implementation_defect` |
| Q21-13 | Full-sudo negative controls | For every normal/login/askpass/shell/runas form, deny before activation and after expiry; also deny stopped/fresh broker, wrong socket, malformed/timeout/audit failure with correct event admission counts. | `regression` / `security_defect` |
| Q21-14 | Identity isolation | Root, another user, `codex-fixture`, and claimed identity inputs cannot obtain production authority; kernel peer identity controls. | `security_defect` |
| Q21-15 | Root-effect non-containment | Create harmless persistent root marker during lease; after expiry new sudo denies while marker remains; manuals deny sandbox/rollback claims. | `requirement_gap` |
| Q21-16 | Rotation | No disclosure; old secret/code and old lease fail, fresh broker starts idle, only new readiness/TOTP restores; interruption is unambiguous/fail-closed. | `security_defect` |
| Q21-17 | Idempotent reinstall | No duplicate/change for identical input; changed artifact or incompatible config stops. | `implementation_defect` |
| Q21-18 | Controlled-failure rollback | Publish finite mutation manifest; inject before/after every mutation and catchable signal; require automatic safe rollback. | `implementation_defect` |
| Q21-19 | Abrupt interruption recovery | At every install, rotation, and uninstall durable crash point inject SIGKILL/reboot/power-loss; use the lifetime `/var/lib/codex-authority-recovery` fallback when installed recovery is absent/self-removed; recover without mixed generation and repeat idempotently. | `implementation_defect` |
| Q21-20 | Recovery journal ordering | Intent/pre-state durable before mutation, completion after verify; corrupt/symlink/wrong metadata state fails closed. | `security_defect` |
| Q21-21 | Uninstall | Independently remove only owned resources/created identity, restore shared pre-state, leave no process/socket/policy/secret/temp/state. | `implementation_defect` |
| Q21-22 | Shared journal and audit retention | Do not delete/truncate/vacuum/rewrite/reconfigure shared journal/security audit; preserve cursor/config; no secret logs. | `security_defect` |
| Q21-23 | Rollback equality | Exact declared shared metadata/tree/identity/unit/root-recovery equality except site-managed append-only journal activity. | `environment_issue` / `implementation_defect` |
| Q21-24 | Documentation | Complete administrator and user install/enroll/rotate/recover/uninstall/incident/non-containment guidance. | `requirement_gap` |
| Q21-25 | Installer-family SLOC exception | Measure installer family separately without charging caps or compressing; existing runtime remains budgeted. | `scope_violation` |
| Q21-26 | Product regression | Focused/full Go, race, vet, format, JSON, shell/static, PAM/sudo syntax, secret-safe scan, diff; existing behavior unchanged. | `regression` |
| Q21-27 | Final evidence review | Independently reconcile same-candidate records, hashes, trust, mutation/crash manifests, separation, equality, SLOC, and secrecy. | `evidence_gap` |

## Required static and unit coverage

Deterministic tests cover strict options, fixed paths/modes, checksum/provenance binding, transaction transitions, atomic journal/corruption, ownership/path traversal, pre-state/rollback, crash recovery, identity decisions, coding-agent-only sudoers, PAM isolation, secret channels, uninstall ownership, SLOC classification, and bounded redaction. Mutation controls must prove critical assertions detect removed protections.

## Destructive E2E sequence

Restore a clean supported snapshot; verify artifact before elevation; capture pre-state and recovery console; enroll/install; verify identity/files/PAM/sudo/unit/socket/seed; activate; run two uncached sudo calls; prove root-effect non-containment; wait natural expiry and deny; rotate and reject old authority; prove stop/restart denial and new activation; exercise controlled rollback and every crash/reboot point from restored snapshots; reinstall; independently uninstall; prove equality/journal retention/no residue; restore the original snapshot.

No destructive case may reuse a host whose prior cleanup result is unknown. The user returns only the
runbook's bounded `Q21-NN result=PASS|FAIL count=N digest=...` records and outer cleanup result. QA binds
the runbook and artifact digests to the reviewed candidate before accepting them; narrative confirmation
or output from a modified runbook is not evidence.

## PASS rule

Final PASS requires independent REVIEW on the same candidate, Q21-01 through Q21-27 PASS including
post-merge main artifact provenance, complete mutation/crash coverage, working out-of-band recovery,
secret-free evidence, unchanged budgeted runtime cumulative SLOC, measured readable installer-family
SLOC, and Main-owned Git/publication. Pre-merge QA may mark only Q21-03 post-merge provenance and Q21-04
post-merge production-bootstrap pending after all pre-merge functional trust negatives pass; this authorizes
implementation merge but not TASK completion. Any other
missing case, secret-bearing record, bypassing sudo rule, ambiguous recovery, unexplained host difference,
or unsafe uninstall is FAIL.
