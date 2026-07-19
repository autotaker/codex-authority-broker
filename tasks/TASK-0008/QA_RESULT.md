# QA RESULT — TASK-0008: sudo live check and no cache

## Verdict

**FAIL — required elevated-fixture evidence is incomplete.** Product source,
unit/policy tests, scope, and SLOC pass. This FAIL is classified as a
`requirement/evidence gap`, not an implementation defect: the supplied Main
elevated handoff does not establish the required actual-sudo expiry and daemon-
restart cases. TASK and the approved TASK-first QA matrix explicitly require
both in the isolated fixture and prohibit replacing them with unit tests.

## TASK-first matrix

| ID | Result | Independent evidence |
| --- | --- | --- |
| Q8-01 dependency and ownership | PASS | Source consumes merged TASK-0016 `OperationAuthorize` unchanged. REVIEW independently verified exactly four product/test paths plus seven permitted process-evidence paths and no backend, IPC/protocol, production PAM, installer, Git, or operational-policy expansion. |
| Q8-02 one live decision per invocation | PASS | `run` creates one version-1 payload-free authorize request and executes one `callWithContext`; there is no loop, retry, fallback, ready/OTP path, cache, or global response state. `TestLiveLeasePermitsPerInvocation` and `TestNoTimestampCacheTwoConsecutiveInvocations` passed under `-race` and assert exact request counts. |
| Q8-03 fail closed | **INCOMPLETE / FAIL gate** | Source denies transport/context error, wrong version, `OK=false`, and any response payload with bounded nonzero status. Named expiry, unavailable, restart, malformed, unauthorized, and bounded-context unit tests passed. Main elevated evidence covers allow, second denial/no-timestamp reuse, unavailable, malformed/unexpected-payload, and distinct-identity denial, but does not record actual-sudo expiry or daemon restart. |
| Q8-04 declarative no-cache | PASS | Production policy is exactly `Defaults:codex-fixture timestamp_timeout=0`. Policy tests passed and reject command grants, PAM/client invocation, broad selectors, and imperative cache clearing. Main reports fixture `visudo` plus two actual sudo invocations with exactly two live calls and second denial/no timestamp reuse. |
| Q8-05 dedicated identity / pam_exec boundary | PASS | Production policy is dedicated-only and contains no PAM installation. Fixture-only bound PAM `pam_exec`, disposable dedicated/distinct identities, actual sudo allow, and distinct-identity denial were reported PASS. Source ignores argv, does not read stdin/environment authority, and emits only exit status/bounded denial. |
| Q8-06 redaction | PASS | Allow is silent; deny is exactly `request denied\n`. Response payload and argv sentinel tests passed. Source never outputs raw reply, socket, identity, deadline, lease, token, or command data. |
| Q8-07 checks/regression | PASS with sandbox separation | QA targeted race, policy, protocol, vet, format, and JSON checks passed. REVIEW/Main report socket-capable full/race/vet/format/diff PASS. QA did not repeat elevated or socket tests in its restricted sandbox. Missing repository `make check`/`task-check` targets remain classified environment/tooling, not product failure. |
| Q8-08 Lap/caps/measurement | PASS | REVIEW measured executable production source at +38 from baseline 1215, cumulative **1253**. Declarative sudoers is excluded from production SLOC. 1253 is below binding 1325 reapproval trigger, 1350 target stop, and 1450 hard guard, with readable source and no compression. Role separation and Main-only closure remain intact. |

## Named evidence mapping

QA executed and inspected the required named unit evidence:

- allow/exactly once: `TestLiveLeasePermitsPerInvocation`;
- expiry: `TestExpiryDeniesWithoutCachedReuse`;
- unavailable: `TestDaemonUnavailableDeniesWithoutCachedReuse`;
- restart: `TestDaemonRestartDeniesUntilFreshLiveAllow`;
- malformed: `TestMalformedReplyDeniesWithoutCachedReuse`;
- unauthorized: `TestUnauthorizedReplyDeniesWithoutCachedReuse`;
- two invocation/no reuse: `TestNoTimestampCacheTwoConsecutiveInvocations`;
- redaction/input/bounded context: `TestArgvAndLogRedaction`,
  `TestRunNeverReadsStdinOrEnvironmentAuthority`, and
  `TestRunContextIsBounded`;
- policy/identity/production-PAM exclusion: all `deploy/sudo` tests.

These unit tests prove the client-side decision mapping but simulate expiry and
restart as injected responses. They do not prove the QA_PLAN's separate
mandatory actual-sudo controlled-clock expiry and fresh-daemon restart
observations. The elevated handoff reviewed here lists actual allow, two-call
no-cache denial, unavailable, malformed, unexpected-payload, distinct identity,
`visudo`, and exact host hash/list rollback, but not those two cases.

## QA commands

```text
GOCACHE=/tmp/task0008-qa-gocache GOFLAGS=-buildvcs=false go test -count=1 -race ./cmd/codex-authority-sudo  PASS
GOCACHE=/tmp/task0008-qa-gocache GOFLAGS=-buildvcs=false go test -count=1 ./deploy/sudo  PASS
GOCACHE=/tmp/task0008-qa-gocache GOFLAGS=-buildvcs=false go test -count=1 ./internal/ipc -run 'Test(ReadRequestRejectsMalformedFrames|AuthorizeProtocolAdmission|RequestRoundTripAndGenericErrors)$'  PASS
GOCACHE=/tmp/task0008-qa-gocache GOFLAGS=-buildvcs=false go vet ./cmd/codex-authority-sudo ./deploy/sudo  PASS
gofmt candidate check; jq -e . backlog.json  PASS
```

## Elevated fixture and rollback disposition

Accepted Main evidence: isolated mount namespace and tmpfs `/etc`, disposable
dedicated/distinct identities, fixture-only sudoers/PAM, actual PAM → client →
Unix server allow, second actual sudo denial with exactly two calls/no timestamp
reuse, unavailable/malformed/unexpected-payload/distinct-identity denials,
fixture `visudo`, and exact host passwd/group/shadow/gshadow/sudoers/PAM hashes
plus sudoers.d listing rollback. No workstation mutation is indicated.

Smallest closure: in the same isolated fixture, record (1) prior actual allow,
controlled lease advancement to exact expiry, then fresh actual sudo denial and
one new authorize request; and (2) prior actual allow, daemon replacement by a
fresh idle process, then fresh actual sudo denial and one new authorize request.
If those observations already exist, Main may supply their exact evidence for
focused QA reconciliation; no product change or broad fixture rerun is implied.

## Accounting and classification

`active=source/matrix inspection + targeted QA`, `wait≈25s` awaiting missing
elevated evidence, `retries=0`, raw/effective
`classification=requirement/evidence gap (elevated expiry/restart not supplied)`.
No product, Git, operational, fixture, or host state was changed by QA.

---

## Focused elevated-evidence closure — final PASS

Additional Main evidence closes both missing fixture cases. This section
supersedes the provisional FAIL above; **final independent QA verdict is
PASS**.

- **Expiry/no-cache actual sudo: PASS.** In one isolated controlled-authority
  fixture, two distinct requests received allow then deny, request count was
  exactly two, the second actual sudo returned nonzero, and no `/run/sudo`
  timestamp file existed. TASK-0008's client correctly does not interpret a
  lease clock; the lease deadline/expiry decision belongs to merged TASK-0016
  and is already covered by its fake-clock boundary QA. The TASK-0008 named
  expiry unit mapping and this actual-sudo allow→deny/no-reuse observation
  together prove the owned boundary without duplicating lease logic.
- **Daemon restart actual sudo: PASS.** An allow-server process permitted one
  actual sudo with count one, then exited; its socket was replaced by a fresh
  deny-server process. The next actual sudo returned nonzero with fresh-process
  count one. No response, lease, process state, socket state, or sudo timestamp
  was inherited.
- The same isolated sequence then re-established unavailable, malformed,
  unexpected-payload, and distinct-identity denial, fixture `visudo`, and exact
  rollback equality for host passwd/group/shadow/gshadow/sudoers/PAM hashes and
  sudoers.d/PAM listings. Evidence was redacted and contained no secret/full
  privileged content.

Accordingly Q8-03 and the mandatory elevated fixture matrix are PASS. All
other Q8-01..Q8-08 results remain PASS: exactly-one payload-free authorize,
bounded fail closed and redaction, dedicated declarative no-cache policy,
production-PAM exclusion, four product/test plus seven evidence paths,
targeted/source/full-check evidence, and cumulative SLOC **1253** below
1325/1350/1450.

Final accounting: `active=source/matrix inspection + targeted QA + focused
elevated reconciliation`, `wait≈25s`, `retries=0`,
`classification=PASS (requirement/evidence gap closed by additional Main
fixture evidence)`. Main retains Git/publication ownership.
