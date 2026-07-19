# TASK-0017 QA plan — TASK-first independent matrix

## Independence and scope

This plan is derived from `TASK.md` before reading any implementation plan. QA
owns evidence and failure classification only; it does not change production,
deployment, fixtures, Git state, or external services. The allowed evidence
paths are the Task production/test paths named in the contract and the
private-namespace fixture outputs. GitHub push, credentials, installer work,
live-host PAM mutation, audit, attestation, release, and canary are excluded.

QA must keep all secrets out of command output and artifacts: no seed, TOTP
value, lease token/state, GitHub credential, or installation token. It must
also reject evidence that runs against the workstation rather than the
specified disposable private mount namespace.

## Preconditions / stop conditions

Before execution, record the approved Task branch/worktree and verify the
merged PASS dependencies TASK-0008, TASK-0015, TASK-0016, plus TASK-0009's
measurement/replan. Require the fixed socket path to be beneath a root-owned,
non-user-writable directory; a dedicated numeric identity with equal,
nonzero UID/GID; one socket; a mode-0600 root-owned disposable seed; copied
tmpfs `/etc`; tmpfs `/run`; fixture-only PAM and sudoers; real Authenticator
TOTP; and exact host-state rollback instrumentation.

Stop and classify as **requirement / implementation-blocking** before an E2E
run if any Task split-stop condition is true: the identity cannot be obtained
only from fixed socket metadata, equal nonzero UID/GID cannot be enforced,
groups/GID/UID cannot be irreversibly dropped before connect, peer authority
rules conflict, a second socket or seed disclosure is needed, or the real
isolated E2E/rollback fixture is unavailable. The 1350 trigger fired and is
closed only by the approved remeasurement: immutable baseline 1253, net
production +145, cumulative 1398. From that measured four-file draft the
additional production allowance is zero. Any added production line, net delta
above +145, cumulative above 1398, use of nominal headroom below target 1400,
compression, unowned path, or inconsistent TASK/PLAN/QA arithmetic stops
before further DEV and requires split/replan. Reaching 1450 is always
forbidden. Root must never be accepted as authority client and SO_PEERCRED
must not be bypassed.

## Acceptance matrix

| Area | Required isolated evidence / named method | Pass condition | Initial failure classification |
| --- | --- | --- | --- |
| Dedicated socket | Focused Go tests for broker access configuration; fixture starts real broker at the one fixed path and runs real CLI as dedicated identity. Inspect `stat`/`lstat` metadata. | Dedicated nonzero equal UID/GID performs `ready` and valid TOTP `otp` without external `chown`; only the fixed socket exists and its metadata matches invariant. | Implementation unless fixture setup is demonstrably unavailable. |
| Peer enforcement | Real socket probes as root, allowed identity, and distinct UID; broker observation of `SO_PEERCRED`. | Allowed dedicated peer succeeds where protocol permits; root and other UID are denied; no caller-selected UID/GID route exists. | Implementation/security regression. |
| PAM identity source | Focused unit/race tests plus isolated actual PAM/sudo run; mutate `PAM_RUSER`, `PAM_USER`, environment, stdin, and caller arguments. | Helper derives identity exclusively by inspecting fixed-path socket metadata, with no trusted PAM/caller input. Mutations cannot change authority or permit access. | Implementation/security regression. |
| Socket validation and TOCTOU | Negative table: missing path, regular file, non-socket, symlink, root/mismatched/zero ownership, and replacement between metadata inspection and connect. Run race/stress test repeatedly. | Every malformed/replaced/raced path fails closed; no connection/authorize occurs under attacker-controlled replacement. | Implementation/security regression; flaky reproduction first recorded as QA evidence. |
| Irreversible privilege drop | Instrumented focused test/inspection of syscall order and isolated E2E. Assert supplementary groups cleared, then GID dropped, then UID dropped before the sole transport/authorize call; failure injection for each drop. | Exactly one payload-free authorize call only after successful permanent drops. Drop failure denies without a connection/allow. No regain path. | Implementation/security regression. |
| Real lease behavior | Real Authenticator TOTP → broker → actual PAM → `sudo /usr/bin/true` twice in private namespace, with sudo timestamp disabled/cleared and authorize-call counter/log-safe probe. | One valid OTP makes one 300-second process-local lease. Both sudo invocations succeed during it, each making one fresh authorize; no sudo timestamp reuse or PAM/helper cache. | Implementation or fixture configuration; distinguish with direct counter and sudo policy evidence. |
| Fail-closed lifecycle | In separate runs: precise lease-expiry boundary, broker stop, fresh broker restart, socket/helper state inspection; next actual sudo invocation each time. | The next sudo fails in all three cases. Allow state never survives PAM, sudo, socket, helper, or new daemon process. | Implementation unless lifecycle action was not actually applied. |
| Output/redaction | Capture bounded stdout/stderr and fixture logs for allow and each denial; secret-scanner/pattern inspection using known values only without printing them. | Allow output is empty; denials are bounded/redacted. Seed, OTP, lease, UID metadata, internal errors, credentials do not occur in results/logs. | Implementation/security regression. |
| Host rollback | Before fixture, hash host passwd/group/shadow/gshadow/sudoers/PAM targets and capture required directory listings; compare exactly after namespace exit. Validate fixture PAM and sudoers syntax inside namespace. | All hashes/listings exactly match after exit; no live workstation PAM/sudoers/identity mutation; syntax is valid in fixture. | Environment/fixture if no host mutation but isolation cannot be established; implementation/regression if host state changes. |
| Code-size and v1 scope | `git diff --check`; production SLOC measurement on only the two contract production paths, baseline/cumulative calculation; inspect Task metadata and changed paths. | Measure only the two contract production paths against baseline 1253. Net production delta is at most +145 and cumulative at most 1398; 1400 is a ceiling, not a +2 allowance; additional production allowance is zero; hard guard 1450 is not approached or crossed. Exactly the same four candidate paths remain owned, with no push/v2/excluded additions. | Requirement/process boundary (or scope regression), not an E2E implementation verdict. |
| Full quality gate | `gofmt`/format check, `go vet`, race-enabled relevant tests, full test suite, `make check`, and scoped diff inspection. | All pass; no unrelated edits, generated drift, secret logging, or Git operation by child agents. | Classify tool/dependency/namespace failures as environment first; reproducible behavioral failures by affected acceptance row. |

## Execution sequence and evidence record

1. Run static/focused validation for socket access, metadata checks, drop order,
   negative paths, peer credentials, and redaction before privileged fixture
   work. Preserve command, exit status, concise failure excerpt, and whether
   authorize was reached.
2. Create the disposable private namespace fixture and record pre/post exact
   hashes/listings. Validate PAM/sudoers syntax before authenticating.
3. Run real TOTP → real broker/socket → actual PAM → actual sudo scenarios,
   including the two in-lease allows, expiry, stop, restart, and peer/identity
   denial matrix. Never record the OTP or seed.
4. Run race stress, full quality checks, diff/SLOC/scope review, then attach
   concise command/exit/observation evidence to the QA result.

For active/wait/retry reporting: mark **active** while a named check is
running; **wait** only for a bounded fixture/process readiness condition with
the observed condition and deadline; **retry** only after classifying a
transient environmental condition and changing the relevant precondition. Do
not blindly rerun functional failures. Record each retry's original failure,
classification, changed condition, and final result.

## Failure classification rules

- **Implementation:** a controlled isolated test violates a Task acceptance
  condition (including fail-open, SO_PEERCRED, metadata/race, identity drop,
  cache/lifecycle, redaction, or rollback behavior).
- **QA-plan:** this plan omits or mis-specifies a Task requirement; amend the
  plan and rerun the affected evidence rather than attributing fault to DEV.
- **Requirement:** Task constraints conflict, are untestable as specified, or
  a split-stop/SLOC boundary requires a main-agent decision.
- **Environment:** namespace privilege, dependency, toolchain, fixture
  provisioning, or host policy prevents a valid test, with no evidence of
  product behavior. Preserve the exact prerequisite failure.
- **Regression:** a previously passing independent baseline/check now fails
  because of out-of-scope drift or changed behavior. Establish baseline and
  changed-path evidence before attribution.

## PLAN reconciliation (performed only after this TASK-first matrix)

Read attempted after completing the TASK-first matrix: `PLAN.md` was not
present in this worktree. No reconciliation is possible yet. When an approved
plan becomes available, append only deviations, additional test hooks, or
missing implementation coverage below; the acceptance matrix above remains
the independent QA baseline.

**Active/wait/retry evidence at handoff:** active—QA planning completed;
wait—approved `PLAN.md` is absent for reconciliation; retry—not applicable,
because no executable QA check has been attempted or repeated.

### Reconciliation completed after PLAN availability

`PLAN.md` was read only after the TASK-first matrix above was complete. It is
consistent with the Task boundary and supplies deterministic test seams and
fixture details needed to make every QA row independently observable. The
prior handoff's “PLAN absent” wait is superseded by this reconciliation.

| TASK-first QA row | PLAN mapping and reconciliation | Result |
| --- | --- | --- |
| Dedicated socket | Broker `run` wires the parsed seed UID into existing `ipc.Config.Access` as equal `OwnerUID`/`GroupGID`, retaining fixed path and `AllowedUID`; injected-listener test verifies all fields and no listen on failure. | Aligned. No external chown or second listener is permitted. |
| Peer enforcement | Existing server `AllowedUID` remains the SO_PEERCRED check; fixture explicitly probes dedicated UID, UID 0, and a distinct nonroot UID. | Aligned. Root remains denied. |
| PAM identity source | `fixedSocketIdentity` reads only fixed `/run` and fixed socket metadata; argv, stdin, PAM/process environment, and caller UID/GID are explicitly non-authoritative and tested with sentinels. | Aligned. |
| Socket validation / TOCTOU | PLAN requires non-symlink root-owned `/run`, `mode & 022 == 0`, then `Lstat` of fixed socket; it covers malformed inputs and replacement before final metadata read. The plan correctly narrows race resistance to an unprivileged dedicated user: such a user cannot replace an entry in a non-writable root-owned parent after validation. Concurrent root filesystem replacement is an explicit split-stop requiring descriptor-pinned design, not a property to claim from current pathname IPC. | Aligned with clarified threat boundary. QA must fail any evidence that claims root-adversary TOCTOU protection without a redesign. |
| UID/GID invariant | Socket metadata requires numeric UID == GID and both nonzero; tests include mismatch, UID 0, GID 0, and distinct values. | Aligned. |
| Irreversible identity drop | Plan fixes order as `Setgroups([]) → observe empty → Setgid → observe real/effective GID → Setuid → observe real/effective UID`, then exactly one call; it requires individual failure injection and forbids restoration/seteuid/setegid. | Aligned. QA must confirm transport count stays zero for every pre-call failure. |
| Real lease behavior | Main-owned namespace fixture uses real Authenticator, real broker/socket, real PAM `pam_exec`, actual sudo, dedicated identity, `timestamp_timeout=0`, and fresh-authorize counting for two lease-period sudo calls. | Aligned. Deterministic tests do not substitute for this E2E evidence. |
| Lifecycle/no cache | Plan retains existing live/no-cache, expiry, unavailable, restart tests and mandates actual sudo failure at exact expiry, broker stop, and fresh restart; new ready/TOTP is required for another allow. | Aligned. |
| Output/redaction | Exact silent allow and bounded `request denied\n` denial are specified; fixture must capture redacted command/status/count evidence and no secret/identity/internal metadata. | Aligned. |
| Host rollback | Plan requires tmpfs `/etc` and `/run`, `visudo`, and exact before/after hashes/listings of the specified host identity/sudo/PAM state, plus cleanup evidence. | Aligned. Live-host mutation remains prohibited. |
| SLOC/v1 scope and gates | PLAN confines candidate implementation to four Go paths, counts executable nonblank/non-comment production lines only in the two contract production files, preserves 1253/+145/1398, records that the 1350 trigger was exercised and closed by the approved remeasurement, permits zero further production, retains target 1400 and hard 1450, and explicitly excludes push/v2 and all Task exclusions. | Aligned after numeric amendment. QA must treat any production increase, excluded-path change, compression, or metadata inconsistency as a requirement/process block that E2E PASS cannot waive. |

No acceptance mismatch was found. Two safety qualifications must remain visible
at execution: (1) the fixed `/run` parent ownership/non-writability is a
mandatory prerequisite to the pathname TOCTOU conclusion; and (2) test seams
may prove forced drop failures but cannot replace the real valid irreversible
handoff and actual TOTP/PAM/sudo fixture.

### Remeasurement revision

An independent TASK-first QA recheck approved only the numeric resource
amendment from `+55 / 1308 / trigger 1350` to measured `+145 / 1398` with
zero further production allowance. Acceptance, authority, threat model, four
owned paths, execution modes, and the live isolated E2E are unchanged. A
missing reproducible count, baseline mismatch, net above +145, cumulative
above 1398, any additional production line/path, compression, or inconsistent
contract arithmetic is fail-closed requirement/process evidence.

### DEV-readiness verdict

**QA-plan ready, DEV not yet authorized by QA evidence alone.** The PLAN and
this independent QA plan are mutually compatible. DEV may begin only when Main
records all Task dependencies as merged PASS, approves both plans, and
preflights a capable private namespace fixture (including `sudo -n unshare`,
tmpfs `/etc`/`/run`, disposable equal nonzero UID/GID identities, real
Authenticator, PAM/sudo/`visudo`, Unix socket capability, controlled expiry,
and rollback comparison). Any missing preflight item is
`not_started/environment_issue`, with one redacted observation and no
live-host or fake-E2E fallback.

**Active/wait/retry evidence after reconciliation:** active—QA reconciliation
complete; wait—Main approval/dependency verification and fixture preflight;
retry—0 (no QA execution attempted, so no failure has been retried).
