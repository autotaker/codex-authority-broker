# PLAN — TASK-0017: dedicated socket ownership and PAM peer identity handoff

## Status and authority

**PLAN evidence only.** DEV may start only after Main verifies merged
TASK-0008/TASK-0015/TASK-0016 plus TASK-0009 measurement, approves this plan
and an independent TASK-first `QA_PLAN.md`, and records a capable isolated
fixture preflight. This plan authorizes no implementation, fixture execution,
host mutation, approval, Git action, stage, commit, merge, or operational-log
write.

The candidate is deliberately limited to these four paths:

| Path | Owned responsibility |
| --- | --- |
| `cmd/codex-authority-broker/main.go` | Pass the seed's allowed UID to existing `ipc.Config.Access`. |
| `cmd/codex-authority-broker/main_test.go` | Prove the exact broker configuration wiring. |
| `cmd/codex-authority-sudo/main.go` | Fixed-path socket metadata admission and irreversible PAM-helper identity drop before its one authorization call. |
| `cmd/codex-authority-sudo/main_test.go` | Deterministic metadata/drop/order/denial/redaction tests. |

No IPC/server/protocol/runtime/seed-schema/policy/PAM installation changes are
eligible. `deploy/sudo/codex-authority` remains the merged dedicated-only
`timestamp_timeout=0` policy. Fixture-local PAM, sudoers, passwd/group, and
mount namespace files are acceptance scaffolding, never candidate production
files. A need for another socket, an alternate identity source, a changed
broker authority rule, a seed disclosure, a root client, or any excluded path
stops and splits before DEV.

Roles remain separate: PLAN, REVIEW, and QA use Terra/medium; approved DEV is
`dev-luna`/luna-xhigh; Main alone owns approvals, locks, Git, and closure.
Children must not stage, commit, merge, or write `.git`.

## Current evidence and fixed boundary

`cmd/codex-authority-broker/main.go:run` currently calls
`listen(ipc.Config{Path: socketPath, AllowedUID: uid}, runtime)`. The existing
Linux `ipc.Listen` already honors a non-nil `Config.Access` by `chown`ing the
new socket, setting mode `0660`, and retaining `AllowedUID` SO_PEERCRED
admission. Because `Access` is omitted, the socket stays root-owned/mode 0600:
the dedicated nonzero seed UID cannot issue real `ready`/`otp`; the PAM helper
then connects as root and is correctly rejected because `parseSeed` forbids
`allowed_uid:0`.

The fixed socket is `/run/codex-authority.sock`; the fixed parent is `/run`.
The parent must be an existing, non-symlink directory, numerically UID 0, and
not group- or world-writable (`mode & 022 == 0`). This root-owned,
non-writable-parent invariant is what prevents the dedicated identity from
unlinking/replacing the pathname after metadata validation. It is not replaced
by a caller-selected directory, `PATH`, PAM item, environment value, stdin,
argv, or account lookup.

The helper accepts only an `Lstat` result for the fixed pathname that is a
socket (not absent, file, directory, or symlink), has numeric UID equal to
numeric GID, and has a nonzero value. It obtains no identity from
`PAM_RUSER`, `PAM_USER`, any PAM environment, process environment, stdin, or
caller-provided UID/GID. Its fixed output contract remains silent allow and
exact bounded `request denied\n` denial; no seed, TOTP, lease, UID/GID,
metadata, or internal error enters output/log/fixture evidence.

## Implementation sequence and invariants

1. In broker `run`, construct the existing configuration as
   `ipc.Config{Path: socketPath, AllowedUID: uid, Access: &ipc.SocketAccess{OwnerUID: uid, GroupGID: uid}}`.
   This makes the dedicated identity own both fields of the one authority
   socket, while the existing server still requires `SO_PEERCRED.Uid == uid`.
   No server-side alternate/root allowlist, `chown` bridge, or second listener
   is permitted. Add an injected-listener assertion that `AllowedUID`,
   `Access.OwnerUID`, and `Access.GroupGID` are the parsed nonzero seed UID,
   with the unchanged fixed path and no listen after seed/runtime failure.

2. In sudo `run`, retain ignored argv and unread stdin, but perform a private
   `fixedSocketIdentity` step before building/transporting the existing fixed,
   payload-free `OperationAuthorize` request. It must inspect `/run` and then
   `Lstat(defaultSocketPath)` without following symlinks; return only a
   generic internal failure. Test-only dependency seams may wrap `Lstat`,
   `Setgroups`, `Setgid`, `Setuid`, and ID/group observations, but must be
   unexported, nil-free production defaults and introduce neither a new
   identity input nor a second call path.

3. On valid metadata let `id` be the one numeric socket owner/group value.
   Before dialing, in this exact order: call `Setgroups([]int{})`; verify the
   supplementary group list is empty; call `Setgid(int(id))`; verify real and
   effective GID equal `id`; call `Setuid(int(id))`; verify real and effective
   UID equal `id`. Any syscall or verification failure denies before transport.
   `id != 0` and `UID == GID` are mandatory: mismatched owner/group, either
   zero field, residual group, or any failed observation denies. Do not use
   `seteuid`/`setegid`, retain saved root credentials, defer restoration, or
   attempt a retry. Linux privileged `setgid`/`setuid` after all groups are
   cleared is the required irreversible transition; the code must not expose a
   regain-privilege branch.

4. Only after all drop checks succeed, execute exactly one existing
   `ipc.Client.Call` with the unchanged two-second context and request
   `{Version: ipc.ProtocolVersion, Operation: ipc.OperationAuthorize}` with
   empty payload. The existing response test remains the sole allow path.
   Metadata error, metadata replacement observed before drop, drop failure,
   dial/transport/timeout/cancellation, malformed/versioned/payload-bearing
   response, `OK:false`, expiry, stopped broker, and fresh broker all fail
   closed with no retry or fallback.

The validation-to-connect pathname check is intentionally backed by the
fixed root-owned non-writable `/run` parent: an unprivileged dedicated user
cannot replace the entry in that interval. A process with root filesystem
authority can replace any root-owned deployment object and is the trusted
deployment/broker authority, not a PAM-controlled peer. If the required threat
model instead demands resistance to concurrent root replacement, a pathname
client cannot establish that property with the current IPC API; stop and split
for descriptor-pinned connection design rather than claim this plan supplies
it.

## Required deterministic tests

The sudo test seam must record event order and number of transport calls, so a
unit test cannot accidentally prove a post-connect drop. Each named row must
be independently observable (subtests are acceptable).

| Evidence | Required assertion |
| --- | --- |
| Broker access wiring | A parsed allowed UID produces exactly one `Config.Access` with matching nonzero `OwnerUID`/`GroupGID`, unchanged fixed path and `AllowedUID`; errors make zero or cleaned-up listen calls as existing lifecycle tests require. |
| Valid metadata/drop/order | Root-owned fixed parent and socket UID=GID=dedicated nonzero yields `groups-empty → gid → uid → one-call`; request is payload-free authorize and only valid empty allow is silent success. |
| Parent/path admission | Missing/non-directory/symlink/non-root/writable `/run`; missing, symlink, regular, directory, or replaced socket all deny with zero calls. Mutation fixtures must cover replacement before the final metadata read. |
| UID/GID invariant | Socket owner/group mismatch, UID 0, GID 0, and distinct values each deny before every drop/transport action. |
| Irreversible-drop failures | `Setgroups`, residual-group observation, `Setgid`, GID verification, `Setuid`, and UID verification failures each deny; later actions and call count remain zero. Assert no restoration syscall is available/invoked. |
| Untrusted inputs | Sentinels in argv, stdin, PAM-like/process environment variables, and caller-selected UID/GID test inputs cannot affect selected identity or request. The code does not read environment authority and stdin's error reader is never read. |
| One live decision/redaction | A prior allow cannot affect next helper process; transport failure, false/malformed/wrong-version/payload response each emit only bounded denial and make at most one call, with no sentinel/metadata/identity leak. |
| Existing regressions | Preserve the fixed request, timeout, live/no-cache, expiry, unavailable, restart, malformed, unauthorized, and silent-success tests; adjust their seams so they exercise the successful metadata/drop prerequisite rather than bypass it. |

Run focused deterministic checks before fixture work:

```sh
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./cmd/codex-authority-broker ./cmd/codex-authority-sudo ./internal/ipc
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 -race ./cmd/codex-authority-broker ./cmd/codex-authority-sudo ./internal/ipc
```

Unix-socket denial by this sandbox is `environment_issue` with the exact
redacted error, not a product waiver. Do not repeat unchanged incapable runs;
the capable fixture below supplies the real transport evidence.

## Main-owned isolated E2E and rollback acceptance

Before DEV, Main must demonstrate a disposable, private mount namespace with
tmpfs `/etc` and `/run`, fixture-only root-owned seed mode 0600, and a newly
created dedicated identity whose numeric UID equals GID and is nonzero, plus a
distinct nonroot identity. It must use real broker, actual Unix socket,
Authenticator-generated TOTP, fixture-local `/etc/pam.d/sudo` `pam_exec`
binding to the helper, fixture-local narrow sudoers command grant, and
`timestamp_timeout=0`; validate fixture policy with `visudo`. No workstation
PAM/sudo/identity/socket mutation is allowed.

The fixture must prove, with redacted command/status/count evidence:

1. The dedicated identity performs real `ready` then real TOTP `otp` through
   the broker-created socket without an external `chown`; SO_PEERCRED rejects
   UID 0 and the distinct UID.
2. Two actual `sudo /usr/bin/true` invocations via PAM succeed inside the
   one 300-second lease, each creates exactly one fresh authorize call, and no
   sudo timestamp reuse exists.
3. At exact expiry the next actual sudo fails; after broker stop it fails; a
   fresh broker restart with no lease also fails. A new real ready/TOTP is
   required before a subsequent allow. Fixture evidence must show that neither
   socket/PAM/helper/sudo/daemon state carried an old allow.
4. The PAM helper's fixed-path failures (missing, non-socket, symlink,
   replacement, UID/GID zero/mismatch, and forced group/GID/UID drop failure)
   deny with zero authorization transport. Where a real kernel failure is
   impractical, the deterministic seam proves that terminal; E2E proves the
   actual valid irreversible peer handoff.
5. Before namespace entry and after teardown, exact hashes and directory
   listings for host passwd/group/shadow/gshadow, sudoers/sudoers.d, and PAM
   configuration compare equal. Record fixture socket/process/timestamp/log
   cleanup and never print seed/TOTP/lease/identity/internal-error values.

If `sudo -n unshare`, tmpfs, disposable users, `visudo`, PAM/sudo execution,
real Authenticator TOTP, socket capability, controlled expiry, or rollback
comparison is unavailable, stop once as `not_started/environment_issue` with
a redacted null reason. It is not permission to use a live host or fake the
E2E. Fixture/elevation waiting is recorded separately from DEV active time.

## Size, gates, and stop rules

The immutable merged baseline is **1253** production SLOC. The contract's
readable forecast is **+55**, so projected cumulative is **1308**; **1350**
is the replan trigger and **1450** the absolute hard guard. Independently
count nonblank, non-comment executable lines in the two production Go files;
tests and fixture/process evidence do not count. Stop before/further DEV if
the forecast or candidate exceeds 1350, reaches 1450, needs compression, or
needs an unowned path. Never shed fixed-path validation, equal nonzero UID/GID,
group/GID/UID drop, SO_PEERCRED, exactly-one live authorize, no-cache, real
E2E, redaction, or rollback.

Lap 1 is the complete bounded implementation plus focused and isolated E2E
evidence. Lap 2 is allowed only for one or two classified, bounded findings
with no redesign/research/fixture change and a demonstrable correction within
its first 20 minutes; otherwise split. No third Lap exists.

DEV then REVIEW and QA must independently run their focused checks once and
REVIEW must complete `make check`:

```sh
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./...
go vet ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
make check
make task-check TASK=TASK-0017
```

Absent tools/targets and restricted socket execution are recorded exactly and
classified, never silently omitted. QA classifies post-merge failures under
the QA guideline rather than presuming implementation fault.

## Planner evidence

| Item | Evidence |
| --- | --- |
| Sources inspected | Root `AGENTS.md`, task contract, delivery skill, current broker/sudo/IPC source and tests, and prior TASK-0008/0015/0016 acceptance evidence. |
| Active work | Repository/task/source inspection and PLAN composition. |
| Wait | None observed. |
| Retry | 0. |
| Classification | `plan_ready`; no approval or DEV started. |
| Residual risk | TOCTOU is contained for unprivileged replacement by the fixed root-owned, non-writable parent. A requirement to distrust concurrent root filesystem replacement exceeds current pathname IPC capability and is an explicit split-stop. |
