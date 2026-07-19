# QA PLAN — TASK-0005: Codex CLI, socket access, and redaction

## Independent TASK-only acceptance baseline

This baseline is derived from `TASK.md`, the assigned boundary, the merged
TASK-0004 IPC/readiness/TOTP candidate, and `backlog.json` before reading
`PLAN.md`.  A later PLAN may make the fixture more precise, but cannot weaken
or replace these requirements.

QA owns independent validation and evidence only.  It does not implement
product code, edit `PLAN.md`, stage, commit, merge, publish, or create a host
account.  QA begins only after independent REVIEW PASS records the requested
CLI/socket fixture, secret-capture scan, exact SLOC count, and no-compression
review.  Missing REVIEW PASS blocks QA PASS and merge.

TASK-0005 owns a deliberately narrow, **non-MCP** Codex CLI and its local
socket-access provisioning.  The only supported user actions are fixed
`readiness` and `otp` actions carried through the existing transport client.
The CLI must work when there is no current host user literally named `codex`:
provisioning must select/configure the authorized socket UID/GID without
depending on that account's presence.  Other identities must deny before
authority dispatch.  This task does not add a daemon/server protocol,
arbitrary operation selection, general RPC/MCP surface, PAM/sudo, push,
audit, release, package/installer, or canary work.

| ID | P0 criterion and fixture | Required PASS evidence / failure condition |
| --- | --- | --- |
| Q5-01 | Invoke each of the two documented fixed actions from an authorized socket identity against the focused real Unix-socket fixture. | `ready` performs only the readiness request and reports the defined nonsecret result; `otp` performs only OTP activation and reports only its defined nonsecret result. Each uses the existing bounded transport client and produces one intended backend transition/call. Any unsupported, missing, duplicated, or arbitrary action/flag fails closed without dispatch. |
| Q5-02 | Provide a valid synthetic OTP **only through stdin** to `otp`; capture the child argv and environment at execution. Repeat with OTP supplied as an argument and environment value. | The accepted path reads OTP from stdin only. The fixture's argv contains no OTP, `--otp`, positional secret, or equivalent; its environment contains no OTP value/name used to carry it. Argument/environment attempts deny before dispatch and cannot become compatibility aliases. |
| Q5-03 | Deliberately cause invalid OTP, unavailable socket, malformed/failed transport response, and client-side usage errors while using a unique synthetic OTP sentinel. Capture stdout, stderr, logs/diagnostics, returned errors, and any fixture recorder. | The sentinel is absent from every captured sink, including error formatting; diagnostics are generic/nonsecret. Failure yields no authority, no backend activation, and no retry that resends or records the OTP. A scan hit is an `implementation_defect`, except a test fixture's intentionally declared input source that is outside captured sinks. |
| Q5-04 | Run authorized success and a real denial fixture using actual kernel-reported peer credentials; record peer UID/GID/PID, configured/provisioned socket UID/GID/mode/ownership, and backend count. | The authorized identity alone can connect and dispatch. A nonmatching UID and a nonmatching GID (where the provisioning policy uses that dimension) deny before request/authority dispatch with backend count zero. Do not accept claimed JSON/request identity. Use the existing `SO_PEERCRED` server, not an in-memory authorization substitute. |
| Q5-05 | Exercise provisioning with no host `codex` user (or an explicit fixture proving no lookup of that name), plus repeated provision/launch and unsafe/unauthorized socket setup. | Provisioning has a deterministic configured UID/GID outcome without a `codex`-name lookup or a requirement to create that account. Socket permissions/ownership admit only the intended principal/group; failures leave no permissive or falsely usable endpoint. If the target UID/GID cannot be created or inspected in the QA environment, record the exact limitation as `environment_issue`; it is not proof that a name lookup is acceptable. |
| Q5-06 | Inspect CLI command grammar, client calls, candidate diff, and exported surface. | It is non-MCP and bounded to the two fixed actions. No generic method/payload passthrough, daemon/server, socket protocol change, OTP/TOTP policy rewrite, PAM/sudo, push, audit, release, packaging, installer, or canary behavior is added. Earlier readiness/TOTP/IPC controls remain fail closed. |
| Q5-07 | Count all candidate production source and review formatting/readability. | Exact cumulative production SLOC is `<=820`; tests are excluded. No semicolon/one-line packing, collapsed error handling, cryptic identifiers, removed security comments, or functions combined solely to meet the cap. `gofmt` is clean and ordinary idiomatic structure remains. |

All Q5-01 through Q5-07 are P0.  A skipped P0 is FAIL unless it is first
classified as a QA-plan defect, requirement gap, or environment issue and the
main Agent approves the disposition.  A generic denial alone never proves a
security control: record no dispatch/activation and the relevant socket or
secret-capture evidence.

## Independent execution and capture evidence

Use a real Linux Unix-domain socket for Q5-01/Q5-04/Q5-05; in-memory clients
cannot establish OS credential or filesystem-access policy.  Use finite
fixture deadlines.  Record candidate revision, OS/kernel and Go versions,
authorized and denied peer UID/GID/PID, configured target UID/GID, socket
path/type/mode/owner/group, action, exit status, stdout/stderr, backend call
count, and a redacted fixture transcript.  Never paste the synthetic OTP or
secret sentinel into a QA result.

The secret-capture fixture must feed a generated unique sentinel via stdin,
retain its value only in the harness's private comparison variable, and scan
the captured argv, environment, stdout, stderr, errors, and logs for it.  The
evidence records only `argv=clean`, `env=clean`, and each sink as `clean` or
`hit`; it never records the sentinel itself.  Add a negative-control test that
would detect an intentionally injected sink leak, so a clean scan is not
merely a vacuous capture.  Do not use shell tracing or an invocation that
places the OTP in a command line while constructing the positive fixture.

After REVIEW PASS, QA independently runs the focused CLI/provisioning/redaction
tests plus:

```sh
go test ./...
go test -race ./...
test -z "$(gofmt -l .)"
go vet ./...
git diff --check
```

Run `make check` if a repository Makefile/contract supplies it; do not add a
Makefile as TASK scope.  Record an unsupported or unavailable command and
classify it rather than silently skipping it.

QA independently runs this exact cumulative production-SLOC command and
records every file/subtotal.  It counts nonblank, non-comment executable Go
source only; excludes `*_test.go`, generated/vendor files, documents,
workflows, and configuration.  An executable line with an inline comment
counts once.

```sh
git ls-files --cached --others --exclude-standard -z -- '*.go' |
  grep -zv '_test\.go$' |
  xargs -0r awk '
    FNR == 1 { if (seen++) { print count " " previous; total += count }; previous = FILENAME; count = 0; in_comment = 0 }
    { line = $0; sub(/^[[:space:]]+/, "", line) }
    in_comment && line !~ /\*\// { next }
    in_comment { sub(/^.*\*\//, "", line); in_comment = 0; sub(/^[[:space:]]+/, "", line) }
    line ~ /^\/\*/ && line !~ /\*\// { in_comment = 1; next }
    line ~ /^\/\*/ { sub(/^\/\*.*\*\//, "", line); sub(/^[[:space:]]+/, "", line) }
    line != "" && line !~ /^\/\// { count++ }
    END { if (seen) { print count " " previous; total += count }; print "TOTAL " total }'
```

The merged baseline is **600** production SLOC.  Before DEV starts, the PLAN
must show a credible projected cumulative total `<=738` (90% of 820), leaving
ordinary review-fix margin.  At a projection or measured total `>738`, stop
and re-estimate/reconcile PLAN and QA before DEV continues; at `>820`, FAIL
the candidate.  Tests do not count toward this production gate.  If an
over-cap candidate is considered, apply only the exact `backlog.json`
ordered-feature-shedding order, reapprove PLAN and QA, never compress source,
and never shed readiness, TOTP replay/rate/absolute-lease, SO_PEERCRED
fail-closed IPC, or secret non-disclosure.  A mandatory-v1 total above 1500
is a `requirement_gap`, not permission to weaken controls.

## Failure classification and QA decision

| Classification | Meaning and disposition |
| --- | --- |
| `implementation_defect` | Candidate violates a P0 criterion, leaks the sentinel, accepts an unauthorized socket identity, requires a `codex` account, exceeds the cap, or expands scope. FAIL and return to DEV with redacted reproduction evidence. |
| `regression` | Candidate breaks prior readiness/TOTP/IPC behavior. FAIL and return to DEV with baseline comparison. |
| `qa_plan_defect` | A fixture or expectation contradicts the approved TASK while the candidate satisfies it. Pause and amend/reapprove QA planning; do not blame DEV. |
| `requirement_gap` | The task/approved PLAN lacks a decision needed to determine target UID/GID provisioning, fixed action semantics, or secret-safe observable behavior. Return to task authority/PLAN. |
| `environment_issue` | OS user/group controls, real-peer fixture, Unix-socket capability, or tooling prevents a valid test and is not candidate behavior. Preserve the limitation and rerun in a suitable environment; do not claim full PASS. |

QA PASS requires REVIEW PASS; all P0 observations; real socket UID/GID evidence;
stdin-only and clean capture scans; full Go/race/format/vet/repository checks;
exact `TOTAL <=820`; and scope/no-compression review.  Any QA FAIL returns to
the responsible gate and never authorizes merge.

## PLAN reconciliation

`PLAN.md` was read after the independent baseline above.  Its concrete design
matches the TASK boundary and supplies testable details without weakening the
P0 criteria.

| TASK-only criterion | PLAN evidence | QA reconciliation decision |
| --- | --- | --- |
| Q5-01 fixed non-MCP actions and transport | `cmd/codex-authority` accepts exactly `ready` and `otp`; `ready` sends a fixed `ready` request with no payload and `otp` sends a fixed `otp` request. `internal/ipc.Client.Call(context.Context, Request)` is the bounded one-request/one-response Unix client. Protocol admits only those two operations. | **Matches.** QA uses these exact names, validates strict action/argument rejection before transport, and confirms the CLI ignores response payload for display. The renamed CLI spelling `ready` is a safe concrete refinement of the TASK's readiness action, not a third action. |
| Q5-02 stdin-only OTP and redaction | `otp` accepts exactly six ASCII digits after a line ending is removed; it rejects missing, extra, oversize, argv, and OTP-looking flag input locally. It has no env/file/child-process OTP source; success and every denial have fixed text. | **Matches.** QA captures argv, environment, stdout, stderr, returned errors, and logs while passing a generated sentinel only through stdin. It also verifies readiness has no OTP payload and a positional OTP rejects without transport. No runtime-zeroization claim is required. |
| Q5-03 clean failure paths | Client validates path, uses `DialContext`, finite deadlines, strict bounded framing, and fixed sentinel errors; CLI never formats underlying errors or response payload. PLAN requires unavailable/malformed/deadline/backend-denial fixtures. | **Matches.** QA treats any sentinel occurrence in a captured sink, any echo/underlying transport text, or any dispatch on a local rejection as an implementation defect. |
| Q5-04 real UID/GID access control | Optional explicit numeric owner UID/group GID changes an otherwise `0600` socket to `0660`; authorization remains exact `SO_PEERCRED` `AllowedUID`. PLAN's real permitted fixture uses `AllowedUID=geteuid()` and real denied fixture uses a deliberately different numeric `AllowedUID`, with zero backend calls. | **Matches with policy clarification.** The group controls filesystem connection permission, but it is not an authorization alternative: a group member with a different kernel UID must still be denied before parsing/dispatch. QA records current peer UID/GID/PID, configured owner/group, mode, and the real mismatched-UID denial. The same-peer/different-`AllowedUID` fixture is valid real kernel-credential denial evidence and requires no privileged second user. |
| Q5-05 numeric provisioning without a host `codex` user | The package receives numeric IDs, does not call `os/user`, and neither creates nor assumes an account/group. Tests use current numeric eUID/eGID, preserve default `0600`, test provisioned `0660`, and exercise identity-safe cleanup on ownership/mode failure. | **Matches.** QA examines the absence of account-name lookup, records numeric policy/socket metadata, and tests success/failure cleanup. No current host user named `codex` is a prerequisite. |
| Q5-06 scope and retained controls | Owned paths are the CLI, IPC client, protocol operation admission, and optional provisioned socket access only. PLAN expressly excludes backend wiring/daemon/MCP, `internal/lease` changes, user/group creation, PAM/sudo, push, audit, release, canary, packaging/service/installer, logs, telemetry, config, env OTP, and dependencies. | **Matches.** QA fails a candidate that adds any excluded surface or weakens existing IPC lifecycle, strict framing, peer-credential, readiness, or TOTP controls. It does not require absent later assembly/backend behavior. |
| Q5-07 gates and projected size | PLAN reproduces the 600-SLOC merged baseline, estimates 105–135 added (705–735 cumulative), binds the `>738` stop/replan threshold and `<=820` cap, supplies the exact count command, and excludes tests. | **Matches.** The 705–735 estimate is below the trigger, so the plan is eligible for DEV only while that projection holds. QA independently recounts all candidate production Go files, requires `TOTAL <=820`, and treats `>738` during DEV as a stop/replan event before further work. |

The PLAN gives specific real-socket/subprocess and capture-scan evidence,
including a documented synthetic OTP and a recording backend.  QA strengthens
the evidence only by requiring a non-vacuous capture negative control and by
recording sink-by-sink redacted results; neither adds product scope.

Final reconciliation status: **PASS — no PLAN compatibility gap.**  This is
not implementation approval.  REVIEW PASS, Q5-01 through Q5-07, the exact
size gate, and all independent QA evidence remain mandatory before merge.

## Revision 2 independent TASK/candidate re-estimation (before PLAN Revision 2)

This section was fixed from `TASK.md`, the existing independent QA baseline,
and the following observed facts only, before reading PLAN Revision 2:

- merged cumulative production baseline: **600 SLOC**;
- incomplete production skeleton: **151 SLOC**;
- observed cumulative production total: **751/820 SLOC**;
- TASK-0005 focused tests: **not yet added**; and
- unconditional TASK ceiling: **820 cumulative production SLOC**.

It supersedes the earlier pre-DEV `<=738` projection/stop rule for Revision 2
because the observed 751 total has already crossed TASK's >90% re-estimation
trigger.  It does not weaken the 820 ceiling or any Q5-01 through Q5-07 P0.

### Independent acceptance, boundary, and failure decision

The accepted product boundary remains the narrow non-MCP `ready`/`otp` CLI,
bounded existing IPC client use, numeric socket provisioning, real
kernel-credential authorization, stdin-only OTP ingestion, and complete OTP
non-disclosure.  Readiness, replay/rate/absolute-lease, strict framing,
`SO_PEERCRED`, lifecycle identity, shutdown, and backend-noncall controls from
prior tasks are mandatory regressions gates.  No MCP/generic RPC, daemon or
backend assembly, account creation/name lookup, config/env/file OTP source,
lease/TOTP policy rewrite, PAM/sudo, push, audit, packaging/release, canary,
dependency, log, or telemetry surface may be added to make the estimate fit.

The incomplete 751-SLOC skeleton is **not gate-ready**: absence of focused
tests leaves Q5-01 through Q5-06, the non-vacuous secret-capture scan, real
socket/provisioning evidence, and prior-control regressions unproven.  This is
a planning/DEV-completion observation, not by itself attribution of a product
defect.  If the skeleton were submitted to REVIEW or QA as complete, the
missing P0 evidence would be an `implementation_defect` and FAIL.

All of the following remain mandatory before REVIEW or QA PASS:

1. exact fixed-action and local-rejection/backend-noncall evidence;
2. stdin-only OTP with argv/environment/stdout/stderr/error/log capture and a
   negative-control proving the scan detects an injected leak;
3. real Unix-socket authorized and deliberately mismatched kernel-UID denial,
   numeric owner/group/mode evidence, unsafe provisioning cleanup, and no
   dependency on a host user named `codex`;
4. unavailable/malformed/timeout/backend-denial generic-output evidence and
   no OTP retry, echo, payload display, or underlying-error disclosure;
5. full prior-task regression, race, vet, gofmt, diff, scope, and readability
   gates; and
6. exact independent cumulative production-SLOC evidence at or below 820.

### Revision 2 SLOC and no-compression controls

A revised target of **745–775 cumulative production SLOC** is conditionally
credible.  Relative to the observed 751 skeleton, it permits legitimate
deletion of incomplete/out-of-scope scaffolding and at most 24 net production
lines of ordinary completion while leaving 45–75 lines below the TASK cap.
Tests and capture fixtures do not count toward production SLOC, so their
absence is an evidence/completeness blocker rather than a reason to consume
the production budget.

The target is safe only if PLAN Revision 2 inventories every unfinished
mandatory behavior, identifies only approved-path deletions/additions, and
shows that all controls fit without semantic compression or scope shedding.
Semicolon/one-line packing, collapsed error handling, cryptic shortening,
combined functions solely for the count, removed security comments, weakened
validation/cleanup, or moving production behavior into tests is forbidden.
Mandatory controls may not be deferred or deleted.

Adopt **`>790` projected or measured cumulative production SLOC** as the
Revision 2 stop/replan threshold.  This preserves at least 30 production lines
below 820 for independent review corrections.  At 791 or above, DEV stops
before further product work and PLAN/QA must be re-estimated and re-approved;
QA does not permit compression or feature shedding to continue.  `>820` is an
unconditional FAIL.  Exactly 790 is not pre-approved for merge: it still must
pass every P0, readability, scope, and independent exact-count gate.

Independent pre-PLAN Revision 2 decision: **CONDITIONAL CREDIBLE / DEV STOP
UNTIL RECONCILIATION APPROVAL**.  The 745–775 target and `>790` stop rule are
safe planning bounds, but credibility depends on PLAN Revision 2 accounting
for the incomplete skeleton and preserving every mandatory control above.

## Revision 2 PLAN reconciliation

PLAN Revision 2 was read only after the independent section above was fixed.
Read-only recount then reproduced the observed cumulative total exactly:

```text
83 cmd/codex-authority/main.go
35 internal/ipc/client_linux.go
117 internal/ipc/protocol.go
283 internal/ipc/server_linux.go
173 internal/lease/lease.go
60 internal/lease/totp.go
TOTAL 751
```

The focused TASK-0005 test files `cmd/codex-authority/main_test.go` and
`internal/ipc/client_linux_test.go` are absent.  Existing IPC tests still
contain the duplicate client framing helpers and legacy `OperationRequest`
fixtures identified by PLAN.  These observations confirm the replan's stated
starting point; they are not implementation changes or QA execution evidence.

| Independent Revision 2 requirement | PLAN Revision 2 evidence | Reconciliation decision |
| --- | --- | --- |
| Preserve the narrow non-MCP fixed-action, stdin-only, redaction, numeric provisioning, real `SO_PEERCRED`, and prior-task controls; add no later boundary. | PLAN retains exactly `ready`/`otp`, one bounded client call, fixed output/errors, numeric UID/GID policy, exact-UID kernel authorization, and all TASK-0004 lifecycle/framing/shutdown controls. It expressly excludes backend/daemon/MCP, account creation, env/file OTP, lease changes, PAM/sudo, push, audit, packaging/release, canary, logs, telemetry, config, and dependencies. | **MATCH / PASS.** No mandatory control is shed and no new product responsibility is used to make the estimate fit. Q5-01 through Q5-07 remain unchanged P0 gates. |
| Account for why the 751 skeleton is incomplete and keep missing acceptance evidence in tests rather than production padding. | PLAN identifies exactly the absent CLI/client/provisioning/redaction tests, duplicate `_test.go` framing helpers, and legacy operation fixtures. Its remaining sequence is test-source work unless a new focused failure proves a production defect. | **MATCH / PASS.** The observed file/test scan confirms this inventory. Tests remain mandatory even though excluded from production SLOC. |
| Provide a non-vacuous secret capture fixture across argv, environment, stdout, stderr, errors, and logs without recording the sentinel. | PLAN requires stdin-only injection, sink-by-sink scan, rejected positional/environment attempts, no echo/retry/payload display, and an intentional negative-control leak. | **MATCH / PASS.** Existing Q5-02/Q5-03 capture requirements are fully retained; no zeroization promise is inferred. |
| Prove real authorized/mismatched peer behavior and provisioning without requiring a host `codex` account. | PLAN uses numeric current eUID/eGID for ownership fixtures, real kernel UID success, a deliberately different `AllowedUID` for real denial/backend noncall, default `0600`, provisioned `0660`, and identity-safe failure cleanup. It forbids `os/user` lookup and account creation. | **MATCH / PASS.** This satisfies Q5-04/Q5-05 while preserving exact UID authorization; group access is not an authorization bypass. |
| Show a credible idiomatic path from 751 to completion with all controls and no compression. | PLAN states the production skeleton already contains all planned responsibilities, limits further production edits to focused acceptance failures, allows only clarity-improving deletion of duplicated glue, and places the missing work in tests. | **MATCH / CONDITIONAL PASS.** The path is credible because mandatory evidence growth is test-only and up to 24 net production lines remain inside the 775 target. Any newly discovered production responsibility or dependency triggers replanning. |
| Use 745–775 as the revised target, stop/replan above 790, retain the unconditional 820 ceiling, and forbid compression/control shedding. | PLAN adopts the same target and stop values, preserves 30 lines below the cap at the stop boundary, requires new PLAN/QA approval above 790, and treats inability to fit mandatory controls idiomatically as `requirement_gap`. | **MATCH / PASS.** The threshold is safe for controlled remaining DEV. Exactly 790 still receives no acceptance presumption; 791 stops, and 821 fails unconditionally. |

### Revision 2 gate decision

**RECONCILIATION PASS — credible path, conditional on main approval.**  The
PLAN's **745–775** target and **`>790` stop/replan** rule safely supersede the
historical `>738` rule for Revision 2.  At the observed 751 total, every known
remaining mandatory activity is test-source work, while the target retains
45–75 lines below 820 and the stop threshold preserves at least 30 lines for
review fixes.  No mandatory control, capture sink, real-socket fixture, or
prior-task regression gate is removed.

This is planning approval evidence only.  DEV remains stopped until the main
Agent approves both Revision 2 documents.  After DEV, independent REVIEW PASS
must precede QA.  QA PASS still requires all Q5-01 through Q5-07 evidence,
non-vacuous sink scans, real UID/provisioning checks, full test/race/vet/gofmt/
diff checks, exact cumulative `TOTAL <=820`, scope review, and no-compression
review.  A projection or measurement above 790 returns immediately to
PLAN/QA; QA must not permit mandatory-control deletion or compressed source.
