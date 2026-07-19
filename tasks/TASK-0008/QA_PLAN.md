# QA PLAN — TASK-0008: sudo live check and no cache

## Independent TASK-first baseline

This QA plan was derived from `AGENTS.md` and `tasks/TASK-0008/TASK.md`
without reading TASK-0008 `PLAN.md`.  QA owns this file only.  It does not
authorize DEV, modify the client, daemon, sudo policy, workstation state,
operational evidence, or Git state.

TASK-0014 must already be merged.  TASK-0008 may change only:

1. `cmd/codex-authority-sudo/main.go`;
2. `deploy/sudo/codex-authority`;
3. `cmd/codex-authority-sudo/main_test.go`; and
4. `deploy/sudo/codex-authority_test.go`.

The daemon/backend assembly, IPC protocol implementation, push, credentials,
audit, release, installer, packaging, canary, and all real workstation sudo
policy are excluded.  Any product/test change outside those four paths is a
P0 scope FAIL.

## P0 acceptance and safety matrix

| ID | Acceptance | Required QA evidence / fail condition |
| --- | --- | --- |
| Q8-01 | Exact dependency and ownership | TASK-0014 merge is verified before DEV.  Candidate changes are confined to the four owned paths; the IPC daemon is consumed unchanged.  Unapproved path changes, generated artifacts, secrets, operational/Git writes, or host-policy edits are FAIL. |
| Q8-02 | One live decision for every invocation | Each `pam_exec`-compatible client process makes exactly one current IPC request before its result.  A fresh valid unexpired lease permits; no local decision, socket response, timestamp, environment value, file, process-global state, or prior invocation can grant later authority.  Zero/multiple request behavior, a stale reuse path, or permit before a live result is FAIL. |
| Q8-03 | Fail closed | Expired lease, unavailable daemon, daemon restart, malformed reply, unauthorized reply, connection/protocol error, timeout, and any non-allow bounded result deny the sudo check.  The client emits a nonzero result and never grants via a previous sudo timestamp.  Any fail-open or unbounded hang is FAIL. |
| Q8-04 | Declarative no-cache policy | The dedicated fixture identity has an explicit, narrowly scoped sudo policy disabling timestamp caching (including no timestamp reuse across commands/invocations).  Policy is parsed by `visudo -cf` inside the fixture and effective behavior is demonstrated by consecutive sudo invocations.  A global/broad policy, an imperative cache-clearing workaround, or absence of effective cache disablement is FAIL. |
| Q8-05 | Dedicated identity and pam_exec compatibility | The policy applies only to the disposable dedicated identity and invokes the fixed client as a `pam_exec`-compatible check with a bounded stdin/stdout/stderr contract.  It neither accepts authority material from argv/environment nor broadens commands/users.  Root, another fixture identity, and an unauthorized peer cannot obtain a grant. |
| Q8-06 | Redaction | Lease/token/authority details, secret-like input, raw IPC reply, socket path if sensitive, and privileged command arguments never appear in argv captures, stdout, stderr, logs, sudo diagnostics, or persisted fixture artifacts.  Output is a bounded decision result only.  Any sensitive or decision-bearing payload leak is FAIL. |
| Q8-07 | Regression and repository checks | Lap 1 passes `go test ./cmd/codex-authority-sudo ./internal/ipc`.  Independent QA also runs `GOCACHE="$(mktemp -d)" go test ./...`, gofmt cleanliness, `git diff --check`, and `jq -e . backlog.json >/dev/null`.  Existing IPC behavior remains compatible. |
| Q8-08 | Two-Lap, caps, and measurement | Preflight is excluded from DEV time.  Only Lap 1 DEV and Lap 2 independent REVIEW/QA are permitted.  Forecast above 1325 stops for approved replan; target 1350 and hard guard 1450 are stops, never compression incentives.  Evidence records fixture/elevation wait separately; paired stage timing, active/wait, retries, raw/effective classifications, source IDs, null reasons, and `ceil(observed_non_preflight_time * 1.20)`. |

All rows are P0.  QA PASS requires independent REVIEW PASS, complete fixture
evidence, and no skipped deny case.  This document is an acceptance baseline,
not DEV approval; PLAN reconciliation is required later.

## Isolated elevated-fixture preflight

The fixture must be an Ubuntu VM, container, or equivalent isolated root
environment whose `/etc/sudoers`, `/etc/sudoers.d`, PAM configuration, sudo
timestamp directory, socket namespace, logs, and test identity are all
private to that fixture.  The host/workstation must not be used as a fixture
or modified in any way.

Before an elevated command, evidence must show:

1. a disposable dedicated non-root identity and a distinct unauthorized
   identity, with fixture-local home, sudo timestamp location, socket, clock,
   and log capture paths;
2. a controlled clock (or bounded injectable time source) able to create an
   unexpired and expired lease deterministically, and a fixture-local socket
   controlled by the test daemon;
3. a narrow, preapproved elevation procedure which writes only identified
   fixture-local policy/PAM files and only for the dedicated identity;
4. pre-change content/hash, owner, mode, and existence evidence for every
   fixture file to be changed; `visudo -cf` validation before activation; and
5. a rollback trap/procedure proved after the run: remove test policy/PAM
   additions, stop test processes, remove fixture-local timestamps/socket/log
   artifacts, restore each captured file exactly, and demonstrate no matching
   dedicated identity/policy/timestamp remains.

If the isolated fixture, controlled socket/clock, narrow elevation, or proven
rollback is unavailable, classify `not_started/environment_issue`.  Do not
run a partial test on a workstation, install a policy there, weaken the test,
or charge the preflight wait to DEV timing.

## Named test inventory

The unit/integration names below are required QA evidence (equivalent naming
is acceptable only when it proves the same observable condition).

| Test | Setup and assertion |
| --- | --- |
| `TestLiveLeasePermitsPerInvocation` | Start the fixture daemon with a current unexpired lease.  Invoke the client once through the dedicated sudo/PAM path; capture exactly one IPC request and a permit. |
| `TestNoTimestampCacheTwoConsecutiveInvocations` | Perform two distinct sudo invocations for the dedicated identity.  Make the daemon record request sequence numbers; require two live requests (one per invocation), validate effective no-cache policy, and prove the second cannot inherit the first result. |
| `TestExpiryDeniesWithoutCachedReuse` | Permit once, advance controlled time beyond expiry, then invoke again.  The second invocation must make a fresh request and deny despite the first permit/timestamp history. |
| `TestDaemonUnavailableDeniesWithoutCachedReuse` | Permit once, make the socket unavailable, then invoke again.  Require bounded nonzero deny and a connection-attempt record; no prior permit is reusable. |
| `TestDaemonRestartDeniesUntilFreshLiveAllow` | Permit once, restart the daemon with fresh/no authority, then invoke again.  Require denial against the new daemon state; only a new live unexpired allow may subsequently permit. |
| `TestMalformedReplyDeniesWithoutCachedReuse` | Have the socket peer return truncated, invalid framing/schema, and oversized/bounded-malformed replies.  Every variant denies with no raw reply leak. |
| `TestUnauthorizedReplyDeniesWithoutCachedReuse` | Return a syntactically valid reply that identifies an unauthorized peer/caller or otherwise fails the fixed authorization contract.  Require deny; it must never be treated as an allow or fall back to a prior result. |
| `TestUnauthorizedIdentityCannotUseDedicatedPolicy` | Run the same sudo/PAM command as the distinct fixture identity.  Require policy denial before/without an authority grant and confirm no broad policy applies. |
| `TestPolicyDisablesTimestampCachingDeclaratively` | Inspect the rendered dedicated-identity policy and validate it with fixture-local `visudo -cf`; use sudo timestamp inspection/effective two-invocation behavior to prove the configured no-cache semantics, rather than relying on `sudo -K` or another cleanup command. |
| `TestArgvAndLogRedaction` | Feed distinctive sentinel authority/lease-like values through all controllable inputs and capture client argv, daemon argv, stdout, stderr, sudo/PAM logs, and fixture artifacts.  Assert no sentinel or raw authority decision appears; output remains bounded. |
| `TestFixtureRollbackAndNoWorkstationMutation` | After all cases, execute rollback and compare fixture file metadata/content to the preflight manifest, verify policy/timestamps/processes/socket are gone, and verify the host path set is untouched (the test setup must contain no host write capability). |

Each negative test is run after a prior valid permit where possible, so it
proves the absence of cached reuse rather than merely an initial denial.  Test
timeouts must be bounded and their expiry must deny, not permit.

## Failure classification and handoff

| Classification | Disposition |
| --- | --- |
| `implementation_defect` | An approved acceptance item fails in the candidate: no live request, fail-open/timeout, cache reuse, weak policy scope, redaction leak, bad exit result, or test/check regression.  Return the smallest reproducible evidence to DEV. |
| `planning_defect` | PLAN cannot implement the fixed one-request-per-invocation, declarative no-cache, dedicated-identity, or isolated-fixture/rollback contract within the four paths/two laps.  Revise and reapprove PLAN/QA before DEV. |
| `qa_plan_defect` | This QA plan conflicts with TASK authority or fixed TASK-0014 protocol behavior.  Amend/reapprove QA planning; do not attribute it to DEV. |
| `requirement_gap` | A necessary security decision (for example the exact dedicated-identity policy surface) is absent and cannot be safely inferred.  Return to task authority; do not broaden sudo access. |
| `environment_issue` | Isolated Ubuntu root fixture, controlled clock/socket, elevation, rollback proof, or required tooling is unavailable independent of the candidate.  Record `not_started` for preflight and rerun in a capable fixture; never waive it. |
| `regression` | Existing IPC or sudo/client behavior outside the owned boundary changes.  FAIL with before/after evidence. |

Initial status: **independent baseline complete; DEV authorization withheld
pending later PLAN reconciliation.**  Current QA metrics are
`active_ms=null`, `wait_ms=null`, `retries=0`,
`classification=null`; null means unobserved, not zero.

## PLAN reconciliation — 2026-07-19

This comparison was made after the independent baseline above was frozen.
The supplied dependency fact is that TASK-0015 is merged at
`fa70d2fc5b8001a00b8cff476292b626cfe61740`; the current TASK-0008 worktree
has not yet been synchronized by Main.  That fact supersedes PLAN's stale
statement that TASK-0015 is only reported REVIEW+QA PASS and unmerged.  It
does not itself prove that the local checkout contains the required merged
TASK-0014 IPC base, nor does it waive any fixture preflight.

| Reconciliation area | PLAN evidence compared with frozen QA baseline | QA result |
| --- | --- | --- |
| Exact path/role boundary | PLAN owns the same four paths, forbids daemon/IPC/PAM-global/workstation changes, assigns `dev-luna`/`luna-xhigh` to DEV, and keeps independent Terra PLAN/REVIEW/QA plus Main-only Git. | **PASS.** The roles remain separated even though their work is timeboxed sequentially. |
| Per-invocation live check and fail-closed response handling | PLAN requires exactly one fresh `ready` IPC call per process, success only for valid/current allow, and deny for expiry, absence/restart, timeout, malformed/unknown/unauthorized response, or local validation error.  It prohibits fallback, retry reuse, cache, and authority-bearing argv/log/output. | **PASS.** This meets Q8-02/Q8-03/Q8-06, subject to candidate and fixture evidence. |
| Named-test mapping | PLAN's `TestRunLiveAllow` maps to `TestLiveLeasePermitsPerInvocation`; `TestRunFailsClosed` maps to expiry, unavailable, restart, malformed, and unauthorized QA tests; `TestRunTwoInvocationsAreIndependent` maps to `TestNoTimestampCacheTwoConsecutiveInvocations`; its policy pair maps to `TestPolicyDisablesTimestampCachingDeclaratively` and `TestUnauthorizedIdentityCannotUseDedicatedPolicy`; its fixture matrix and rollback proof map to all remaining named QA cases. | **PASS, conditional.** Equivalent implementation names are acceptable only if the individual mutations, one-request observations, bounded deny exits, redaction capture, and cleanup assertions remain separately executable; grouping them into a single broad test is not enough. |
| Dedicated no-cache policy | PLAN limits a declarative `timestamp_timeout=0`-style policy to the disposable dedicated identity, rejects global/default policy and broad grants, and requires Ubuntu syntax/PAM-hook validation. | **PASS, conditional on fixture proof.** Q8-04/Q8-05 still require fixture-local `visudo -cf`, effective consecutive invocation evidence, and no `sudo -K` substitute. |
| Fixture, elevation, rollback, and no workstation mutation | PLAN requires isolated Ubuntu sudo/PAM, disposable identity, clock/socket, narrow reversible elevation, artifact removal, and restoration proof; it classifies missing isolation/elevation/rollback as `not_started/environment_issue` outside counted DEV time. | **PASS as a contract; DEV preflight BLOCKED.** No actual fixture/elevation/rollback evidence is supplied.  Its absence remains a hard blocker and cannot be replaced by host testing. |
| Dependency and local synchronization | PLAN requires a merged dependency exposing the fixed IPC contract, not TASK-0015 artifacts.  TASK-0015 is now known merged at the supplied hash, but PLAN's unmerged wording is stale and the local worktree still needs Main synchronization. | **PARTIAL / preflight BLOCKED.** Main must synchronize and verify the actual TASK-0014 dependency commit/protocol before minute 0.  TASK-0015's merge is governance evidence, not a TASK-0014 substitute. |
| One-Lap timing and role split | PLAN sets one counted 30-minute Lap with separate DEV, independent REVIEW, and independent QA gates at minutes 0/25/30.  The merged TASK-0015 fact is accepted as the stated source of this one-Lap governance update. | **PASS, conditional on Main recording the governing TASK-0015 merge in the synchronized evidence.** The earlier TASK-first two-Lap wording is superseded only for process timing; it does not permit a combined role or skipped REVIEW/QA.  Planning/pure wait is capped at 20% of the counted interval and fixture/elevation wait stays outside it. |
| Exceptional Lap 2 | PLAN allows it only for a concrete estimated residue, no redesign/research, exactly one or two classified causes, and a demonstrable first-20-minute fix; the same cause may replan once, then must split; no Lap 3. | **PASS.** Any missing fact, a third cause, redesign/research, no first-20-minute demonstration, repeated same-cause retry, or incomplete fixture matrix instead requires split/stop, never an automatic second Lap. |
| TASK-local SLOC controls | Frozen TASK authority requires forecast above **1325** to stop before DEV for split/re-estimation and approved PLAN/QA revision; **1350** is the target cap and **1450** absolute hard guard.  PLAN instead says local warnings do not require PLAN reapproval and replaces the mandatory limits with global 1500/1800. | **FAIL — `planning_defect`.** A global capacity number cannot relax this TASK's explicit trigger, target, hard guard, or approval requirement.  PLAN must retain 1325/1350/1450 as controlling local gates and require reapproved PLAN/QA before DEV when the 1325 forecast is crossed. |

### Readiness decision

**FAIL — DEV is not ready.**  The controlling failure is the SLOC/stop-rule
`planning_defect` above.  Separately, preflight remains
`not_started/environment_issue` until Main synchronizes the local base and
verifies the isolated Ubuntu fixture, narrow elevation, controlled
clock/socket, and a successful rollback rehearsal.  Neither condition may be
waived by TASK-0015's merge or by a 30-minute one-Lap target.

Smallest safe correction: amend PLAN to restore the TASK-0008 local
1325/1350/1450 gates and its approved-replan requirement, record the supplied
TASK-0015 merge rather than its stale unmerged status, have Main synchronize
the local dependency base, then obtain an actual fixture preflight/rollback
record.  Reconcile and reapprove PLAN/QA before DEV; do not modify code or
policy while this status is FAIL.

Reconciliation metrics: `active_ms=null`, `wait_ms=null`, `retries=0`,
`classification=planning_defect`; no authoritative same-attempt elapsed pair
or fixture execution evidence was supplied, so null is not zero.

## TASK-first Revision baseline — merged TASK-0016

This revision was fixed before reading the new TASK-0008 PLAN. It preserves
the original matrix while superseding stale TASK-0014/`ready` assumptions.
The controlling dependency is TASK-0016 merged as
`a0d72ff482efc00b81e551df7e0c652aba820f2c`; its QA evidence establishes the
fixed version-1, payload-free `authorize` operation and cumulative production
baseline **1215**.

### Revised scope and capacity gates

- Candidate product/test scope is exactly four paths:
  `cmd/codex-authority-sudo/main.go`, `deploy/sudo/codex-authority`,
  `cmd/codex-authority-sudo/main_test.go`, and
  `deploy/sudo/codex-authority_test.go`.
- Process evidence is separate and limited to exactly seven paths:
  `backlog.json`, `tasks/TASK-0008/TASK.md`,
  `tasks/TASK-0008/PLAN.md`, `tasks/TASK-0008/QA_PLAN.md`,
  `tasks/TASK-0008/REVIEW_RESULT.md`, `tasks/TASK-0008/QA_RESULT.md`, and
  `tasks/TASK-0009/TASK.md`. Evidence paths do not expand product scope or
  production SLOC; role ownership and Main-only Git remain unchanged.
- Forecast is **+55** ordinary production SLOC, with acceptable planning range
  **+45..65**, from baseline 1215: expected cumulative **1270**, range
  **1260..1280**. Forecast **>1325** requires split/re-estimation and approved
  PLAN/QA revision before DEV; candidate **>1350** stops for explicit replan
  and ordered shedding review; **1450** is absolute. No threshold permits
  packing, deleted checks, broadened policy, or weakened fail-closed behavior.

### Revised live-check and fixture acceptance

Every client process must issue exactly one request with
`Version=ProtocolVersion`, `Operation=OperationAuthorize`, and zero payload.
It must make no `ready`, `otp`, probe, retry, or second authorize call. Permit
requires exactly one current `OK=true` response; errors, timeout, EOF,
payload-bearing/malformed/wrong-version response, `OK=false`, unavailable or
restarted daemon, expired lease, and unauthorized identity/peer all terminate
nonzero without fallback. No password prompt/input, password-based success,
sudo timestamp reuse, local cache, prior response, or prior invocation may
grant authority.

The elevated acceptance fixture must use an isolated mount namespace with
private fixture-local `/etc`, PAM, sudoers, timestamp, socket, log, passwd,
group, and shadow views; disposable dedicated and unauthorized identities;
and actual `sudo` through the configured PAM/`pam_exec` path. It must validate
policy with fixture-local `visudo -cf`, demonstrate noninteractive/no-password
execution, and prove timestamp reuse is disabled without `sudo -K` as the
mechanism. Before/after host evidence must include hashes for passwd, group,
shadow, sudoers and relevant PAM files plus listings of sudoers/PAM/timestamp
locations; rollback restores/removes fixture artifacts and exact comparison
must show the host hash/list snapshot unchanged. Missing isolation, actual
sudo, disposable identity, narrow elevation, or rollback proof is
`not_started/environment_issue`, not a reason to test on the workstation.

### Required named mapping

| Revised observable | Required named evidence |
| --- | --- |
| One live allow and exactly one authorize call | `TestLiveLeasePermitsPerInvocation`; request recorder asserts one payload-free `authorize`, successful actual sudo, no password prompt/input. |
| Two independent invocations and no timestamp reuse | `TestNoTimestampCacheTwoConsecutiveInvocations`; two actual sudo processes produce exactly two authorize requests and the second is not bypassed by a timestamp. |
| Expiry after prior allow | `TestExpiryDeniesWithoutCachedReuse`; exact fresh authorize occurs and actual sudo denies nonzero after controlled expiry. |
| Daemon unavailable | `TestDaemonUnavailableDeniesWithoutCachedReuse`; one bounded connection attempt, nonzero deny, no timestamp fallback. |
| Daemon restart | `TestDaemonRestartDeniesUntilFreshLiveAllow`; fresh idle process denies and cannot inherit the former process lease. |
| Malformed authority reply | `TestMalformedReplyDeniesWithoutCachedReuse`; truncated framing, invalid JSON/schema, wrong version, unexpected payload, and oversized response all deny without disclosure. |
| Unauthorized identity/peer/reply | `TestUnauthorizedReplyDeniesWithoutCachedReuse` and `TestUnauthorizedIdentityCannotUseDedicatedPolicy`; syntactically valid but unauthorized conditions deny, and the distinct identity cannot use the dedicated policy. |
| Declarative policy/no password/rollback | `TestPolicyDisablesTimestampCachingDeclaratively`, `TestArgvAndLogRedaction`, and `TestFixtureRollbackAndNoWorkstationMutation`; prove `visudo`, actual PAM/sudo, bounded output, no password/cache path, and exact host hash/list restoration. |

Equivalent test names are acceptable only with separately visible assertions
for every mutation and exact request count. A grouped success/failure table
cannot omit restart, unauthorized, malformed variants, actual-sudo behavior,
or rollback evidence.

### Revised Lap rule and initial readiness

One counted Lap contains DEV, independent REVIEW, independent elevated QA, and
Main closure after preflight; fixture/elevation wait is recorded separately
and does not consume DEV time. Lap 2 is exceptional only for exactly one or
two classified findings, no redesign/research/Task or fixture change, a
bounded residual estimate, and a correction demonstrably reviewable in its
first 20 minutes. Otherwise split; no Lap 3.

Revision-baseline readiness is **conditional**: the contract is testable, but
DEV may start only after the new PLAN reconciles this baseline and Main records
the mount-namespace actual-sudo rehearsal plus complete host hash/list rollback
proof. Accounting: `active=TASK/index/TASK-0016/code revision analysis`,
`wait=0`, `retries=0`, `classification=revision_baseline_fixed`.

## New PLAN reconciliation and DEV readiness

### Reconciliation result

The revised PLAN is **PASS on dependency, scope, behavior, SLOC, and delivery
structure**:

- it consumes merged TASK-0016's payload-free `OperationAuthorize`, explicitly
  rejects `ready` as a substitute, and requires exactly one `Client.Call` per
  process with no retry/fallback;
- its allow, expiry, unavailable, restart, malformed, unauthorized, and two-
  invocation named mapping covers the Revision baseline, including distinct
  actual-sudo fixture mutations rather than a unit-only substitute;
- it fixes exactly four product/test paths and seven separate process-evidence
  paths, with evidence excluded from production SLOC;
- it uses baseline 1215, forecast +55/range +45..65, cumulative 1270/range
  1260..1280, and preserves binding >1325 reapproval, >1350 replan/shedding,
  and >=1450 hard-stop behavior without compression;
- it requires an isolated Ubuntu mount namespace, disposable dedicated and
  unauthorized identities, private PAM/sudoers/timestamp/socket/log views,
  fixture `visudo -cf`, actual sudo through the PAM hook, and host passwd,
  group, shadow, sudoers, PAM hash/list rollback comparison; and
- it defines one counted Lap after preflight, with Lap 2 limited to one or two
  classified causes, bounded residue, no redesign/research/fixture change, and
  a first-20-minute demonstrable correction; no Lap 3.

### Remaining readiness failures

**DEV readiness: FAIL (`not_started/environment_issue` plus one minimal
QA-contract clarification).** The PLAN describes Main's preflight as work that
must occur before DEV, but the materials inspected here do not contain the
required concrete record: namespace command/status, disposable identities,
actual dedicated-identity sudo result, every before/after host hash and list,
and cleanup/rollback comparison. The TASK/index statement that a rehearsal
proved the mechanism is useful lineage, but it does not expose the exact
acceptance evidence required by this baseline. QA cannot infer it.

The smallest PLAN/test clarification is also required: run the dedicated-
identity actual-sudo cases noninteractively and assert **no password prompt and
no password bytes/input**, separately from `sudo -n unshare` used to create the
namespace. Capture stdin/stdout/stderr/PAM/sudo logs, fail on any password
request, and pair that observation with the two-invocation timestamp proof.
The current PLAN prohibits client stdin and requires actual sudo, but does not
name this dedicated-sudo no-password assertion explicitly.

No product change is authorized to resolve readiness. Main can unblock DEV by
recording the already-rehearsed exact preflight/rollback evidence and approving
a minimal PLAN clarification (or equivalent named fixture assertion), then
reconciling both plans. If any preflight element is unavailable, retain
`not_started/environment_issue`; do not retry unchanged conditions or use host
policy.

Final accounting: `active=revision baseline + PLAN reconciliation`, `wait=0`,
`retries=0`, `classification=DEV_NOT_READY (environment evidence absent;
minimal QA-contract clarification required)`.

## Main preflight update and final readiness supersession

Main's final preflight evidence resolves the environment condition above:
tmpfs `/etc` copy, disposable `codex-fixture` identity, tmpfs `sudoers.d`, a
bound PAM auth stack containing only the `pam_exec` check, fixture `visudo`,
and `runuser` to actual `sudo true` all passed. Host passwd/group/shadow/
gshadow/sudoers/PAM hashes and the `sudoers.d` listing matched exactly before
and after teardown. This is accepted as PASS for isolation, disposable
identity, actual sudo without a password-auth fallback, and host rollback.

One blocking PLAN ambiguity remains. The fixed implementation contract says
the declarative policy “invokes the fixed client,” while the stated and proven
boundary is:

- production `deploy/sudo/codex-authority` is only the dedicated-identity,
  no-timestamp sudoers policy; it does **not** launch the client and must not
  contain or imply a PAM hook; and
- only the isolated fixture's bound PAM configuration invokes the
  `pam_exec`-compatible client to exercise actual sudo. No production/global
  PAM file is in TASK-0008 scope.

Without that clarification DEV could implement an invalid sudoers/client
coupling or expand into forbidden production PAM scope. Therefore final
**DEV readiness remains FAIL (`planning_defect`)**. Smallest correction: amend
PLAN fixed-contract item 4 and the policy test mapping with the two bullets
above; retain the accepted Main preflight evidence, then re-reconcile and
approve both plans. No fixture rerun or product change is required for this
wording correction.

Superseding accounting: `active=preflight evidence reconciliation + boundary
inspection`, `wait=0`, `retries=0`, `classification=planning_defect (sudoers
policy vs fixture-only PAM invocation ambiguity)`.

## Focused PLAN correction reconciliation — PASS

The blocking boundary ambiguity is corrected. PLAN now states that production
`deploy/sudo/codex-authority` contains only the dedicated identity's
`timestamp_timeout=0`-equivalent setting, with no command grant, PAM hook,
client invocation, or installer behavior. Production/global PAM installation
remains excluded. The narrow command grant and `/etc/pam.d/sudo` `pam_exec`
hook are explicitly fixture-local tmpfs scaffolding used only to prove the
client through actual sudo.

This matches the Revision baseline and the accepted Main preflight: isolated
tmpfs `/etc`, disposable identity, fixture-local sudoers/PAM, `visudo`, actual
sudo without a password-auth fallback, and exact host hash/list rollback all
PASS. Exactly-once payload-free authorize, deny/no-cache matrix, four product/
test plus seven evidence paths, 1215/+55/1270 range arithmetic, binding
1325/1350/1450 gates, one-Lap path, and exceptional Lap-2/no-Lap-3 rules remain
unchanged and aligned.

**DEV readiness: PASS**, subject to Main's explicit approval of both plans and
normal minute-0 base/scope verification. No further PLAN, fixture, product,
Git, or operational correction is required by QA before DEV.

Focused reconciliation accounting: `active=focused PLAN boundary recheck`,
`wait=0`, `retries=0`, `classification=PASS (planning_defect_corrected)`.
