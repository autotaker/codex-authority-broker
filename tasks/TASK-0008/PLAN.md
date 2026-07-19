# PLAN Revision — TASK-0008: sudo live check and no cache

## Status, decision, and sole ownership

**PLAN evidence only.  DEV is not authorized.**  This revision corrects the
QA reconciliation `planning_defect`; it remains pending independent QA
reconciliation and Main approval of both plans.  It authorizes no product,
test, fixture, host, operational-log, approval, stage, commit, merge, or
`.git` action.

This Planner owns and changed only `tasks/TASK-0008/PLAN.md`.  The intended
execution worktree is explicitly `/tmp/codex-authority-broker-task0008`; Main
must verify that its checked-out base contains merged TASK-0016 commit
`a0d72ff482efc00b81e551df7e0c652aba820f2c` before minute 0.  That merge adds
the consumed fixed IPC operation; it does not authorize TASK-0008.

TASK-0016 defines exactly the version-1, payload-free
`ipc.OperationAuthorize` operation.  Its sole positive result means the
process-local lease was active at the authorization decision point.  It
preserves `ready`/`otp`, retains one custom registration slot, and fails
closed for cancellation, close, expiry, malformed, unknown, and
payload-bearing requests.  TASK-0008 consumes that contract unchanged:
`ready` is expressly not an acceptable substitute because it begins
readiness and denies an already active lease.

Roles remain separated: Planner/REVIEW/QA are Terra/medium, DEV is
`dev-luna`/luna-xhigh, and Main alone owns approvals, locks, Git, and final
classification.  Children do not stage, commit, merge, or write `.git`.

## Immutable candidate boundary

Only these **four** product/test paths are eligible in the candidate:

| Path | Responsibility |
| --- | --- |
| `cmd/codex-authority-sudo/main.go` | Fixed `pam_exec`-compatible live client. |
| `deploy/sudo/codex-authority` | Dedicated-identity `timestamp_timeout=0`-equivalent setting only. |
| `cmd/codex-authority-sudo/main_test.go` | Fixed-client request, denial, and invocation tests. |
| `deploy/sudo/codex-authority_test.go` | Policy narrowness/no-cache tests. |

Only these **seven** process-evidence paths may accompany that candidate;
they are not product SLOC, and each role writes only its own evidence:

1. `backlog.json`
2. `tasks/TASK-0008/TASK.md`
3. `tasks/TASK-0008/PLAN.md`
4. `tasks/TASK-0008/QA_PLAN.md`
5. `tasks/TASK-0008/REVIEW_RESULT.md`
6. `tasks/TASK-0008/QA_RESULT.md`
7. `tasks/TASK-0009/TASK.md`

Everything else is forbidden: broker/backend/runtime/IPC/protocol changes,
another client entrypoint, daemon assembly, seed/persistence, credentials,
push/GitHub, audit/release/installer/packaging/canary, PAM installation or
production PAM hook, command grant, and any real workstation policy or
identity mutation.  A need for one is a
pre-DEV security/scope stop and split, never an implementation shortcut.

## Fixed implementation contract

1. `cmd/codex-authority-sudo` accepts no action, OTP, authority material, or
   privileged command data from argv, stdin, environment, file, cache, or
   process-global state.  Its fixed request is exactly
   `ipc.Request{Version: ipc.ProtocolVersion, Operation: ipc.OperationAuthorize}`
   with no payload.
2. Each client process performs **exactly one** `ipc.Client.Call` before it
   returns.  One valid `OK=true` response is the only success condition.  It
   must make neither a retry nor fallback call; `OK=false`, any transport,
   timeout, framing/schema/version error, malformed/unauthorized response,
   cancellation, or local validation error returns bounded nonzero denial.
   A successful prior process, sudo timestamp, daemon restart, or stale local
   state cannot grant a later invocation.
3. The client is `pam_exec` compatible: bounded I/O and exit status only.
   It emits neither raw reply nor lease/secret/identity/deadline/socket or
   command detail in argv, stdout, stderr, log, or fixture artifact.
4. The production `deploy/sudo/codex-authority` policy is declarative and
   contains only the dedicated identity's `timestamp_timeout=0`-equivalent
   setting.  It contains no command grant, PAM hook, client invocation, or
   installer behavior; production PAM installation is excluded from this
   Task and belongs to a later installer Task.  It must not change global
   defaults, broaden users/commands, or simulate the result with `sudo -K`.
   Client invocation is proved only by the isolated fixture-local
   `/etc/pam.d/sudo` `pam_exec` hook described below.

## Required test and fixture mapping

DEV must create independently runnable tests (equivalent names only when the
individual mutation and assertion remain separately observable):

| Required evidence | Concrete structure and assertion |
| --- | --- |
| `TestLiveLeasePermitsPerInvocation` | Inject a recording `callFunc`; require one payload-free `OperationAuthorize` request and zero exit only for `Response{OK:true}`. |
| `TestExpiryDeniesWithoutCachedReuse` | First process permits; controlled lease clock then expires; second process makes its own request and denies. |
| `TestDaemonUnavailableDeniesWithoutCachedReuse` | After a permit, remove/listen-fail the fixture socket; next process attempts once and bounded-denies. |
| `TestDaemonRestartDeniesUntilFreshLiveAllow` | After a permit, replace the daemon with idle/no-lease state; deny until a new live `authorize` allow. |
| `TestMalformedReplyDeniesWithoutCachedReuse` | After a permit, use truncated, invalid-frame/schema, and oversized reply fixtures; each denies without raw-reply output. |
| `TestUnauthorizedReplyDeniesWithoutCachedReuse` | A syntactically valid but contract-invalid/unauthorized reply denies after a prior permit. |
| `TestNoTimestampCacheTwoConsecutiveInvocations` | Drive two separate sudo invocations, record sequence numbers, and require exactly two distinct client calls plus no inherited permit. |
| `TestUnauthorizedIdentityCannotUseDedicatedPolicy` | Run the identical sudo/PAM path under the distinct fixture user; policy denies before authority grant. |
| `TestPolicyDisablesTimestampCachingDeclaratively` | Assert dedicated-only policy content; fixture `visudo -cf` accepts it; effective consecutive invocation behavior proves no cache without imperative clearing. |
| `TestArgvAndLogRedaction` | Capture controllable argv/stdout/stderr/PAM/sudo logs and fixture artifacts with sentinels; assert bounded output and no sentinel/raw decision. |

The isolated fixture is also mandatory acceptance evidence, not a unit-test
replacement.  It runs `allow`, `expiry`, `daemon-unavailable`,
`daemon-restart`, `malformed`, `unauthorized`, and `two-consecutive`; every
negative case follows a prior permit when feasible.  It uses an isolated
Ubuntu mount namespace with a disposable dedicated non-root identity and a
distinct unauthorized identity, controlled clock/socket, and fixture-local
sudo timestamp/log locations.

## Main-owned preflight and rollback evidence

Before DEV, Main performs and records a **read-only-to-host, disposable**
rehearsal through `sudo -n unshare` in a new mount namespace.  The procedure
must mount tmpfs `/etc`, copy the fixture `/etc` content into that tmpfs,
create the disposable identities there, generate a tmpfs-only narrow command
grant in `sudoers.d`, and bind the fixture-local `/etc/pam.d/sudo` `pam_exec`
hook, validate with `visudo -cf`, then execute actual `sudo` as the dedicated
identity.  The narrow command grant and PAM hook are fixture scaffolding, not
contents of `deploy/sudo/codex-authority` and not production installation
artifacts.  The rehearsal must never point an operation at the workstation's
live policy files.

The preflight record names every fixture path and records before/after hashes
and listings for host `/etc/passwd`, `/etc/group`, `/etc/shadow`, sudoers and
`sudoers.d`, and PAM configuration.  It records the namespace commands and
their exit status, the actual privileged command result, and proves rollback
PASS by unchanged host hashes/listings after namespace teardown.  It also
records removal of fixture socket, timestamps, logs, test policy/PAM files,
processes, and identities from the disposable namespace.  `visudo` success
alone is insufficient; the dedicated identity must execute actual sudo
through the bound PAM hook.

If `sudo -n`, `unshare`, tmpfs, identity creation, bind mount, `visudo`,
actual sudo, controlled socket/clock, or the rollback comparison is not
available, stop once as `not_started/environment_issue`, with a redacted null
reason.  Do not retry unchanged conditions, run a host partial test, or charge
fixture/elevation wait to DEV time.

## Binding size, stop, and delivery rules

The measured merged baseline is **1215** production SLOC.  The source-based
forecast is **+55** (ordinary range **+45..65**), therefore cumulative
**1270** (range **1260..1280**).  This is based on the existing 83-SLOC
general ready/OTP CLI: the dedicated client deletes argument parsing, OTP
input/JSON, and multi-operation output, while adding the fixed client and
two-line policy.  It is not a throughput estimate.

The following local gates are binding and cannot be replaced by global caps:

| Condition | Required disposition |
| --- | --- |
| Forecast or re-estimate exceeds **1325** | Stop before DEV/further DEV; split or re-estimate, reconcile QA, and obtain approved revised PLAN and QA_PLAN before resuming. |
| Candidate/forecast exceeds **1350** | Stop for explicit replan and exact ordered shedding review; never compress or weaken a mandatory control. |
| Candidate reaches/exceeds **1450** | Absolute hard stop; no borrowing later reserve. |

Also stop/split for a second entrypoint, non-declarative/broad policy,
interface change, platform/PAM variance that defeats the fixture, missing
fixture/rollback proof, or inability to complete the named matrix.  Never
shed live-per-invocation, exactly-once invocation, fail-closed behavior,
dedicated identity, declarative no-cache, redaction, or rollback proof.

The normal path is one counted Lap: Main preflight is outside DEV timing;
Lap 1 is DEV followed by independent REVIEW and QA.  Lap 2 is exceptional
only with all four recorded facts: bounded residue estimate, no redesign or
research, exactly one or two classified causes, and a demonstrable fix in its
first 20 minutes.  The same cause allows at most one replan; otherwise split.
No Lap 3.

At each attempt, record UTC start/end, paired same-task/lap/stage/attempt
`active_ms` and `wait_ms`, propagated retry count, raw and effective
classification, source ID/command, redacted null reason, and fixture/elevation
wait separately.  Time contingency is only
`ceil(observed_non_preflight_time * 1.20)`; no SLOC/minute sizing is allowed.

## Gate commands and approval prerequisite

After successful preflight and only after both plans are approved, DEV runs:

```sh
go test ./cmd/codex-authority-sudo ./internal/ipc
```

Independent REVIEW and QA rerun focused evidence and repository checks once:

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
make check
make task-check TASK=TASK-0008
```

Absent targets/tools are recorded with exact output and classified rather than
silently omitted.  REVIEW must complete `make check`; QA applies the QA
guideline to post-merge failures.  Main alone performs scope/hook/Git closure
after independent REVIEW PASS and QA PASS.

**DEV approval remains withheld until QA reconciliation marks this revision
consistent with the frozen QA baseline and Main explicitly approves both
plans.**

## Planner evidence

| Item | Evidence |
| --- | --- |
| Read sources | Current TASK-0008/TASK metadata and backlog entry; merged TASK-0016 contract/code; current 83-SLOC general client and IPC protocol; existing TASK-0008 QA reconciliation FAIL. |
| Corrected defects | Replaces stale `ready` invocation with fixed payload-free `OperationAuthorize`; restores mandatory 1325/1350/1450 local controls and reapproval gate. |
| Scope proof | This revision changes only this PLAN; candidate scope is exactly four product/test and seven process-evidence paths. |
| Current state | `active_ms=null`, `wait_ms=null`, `retries=1`, raw/effective classification `planning_defect_corrected`; no authoritative same-attempt elapsed pair exists, so null is not zero. |
