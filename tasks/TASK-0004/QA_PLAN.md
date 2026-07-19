# QA PLAN — TASK-0004: Versioned SO_PEERCRED IPC

## Independent TASK-only criteria

This section is derived from `TASK.md` and the QA assignment before reading
`PLAN.md`.  It is the acceptance baseline; later reconciliation with PLAN may
clarify execution details but may not silently replace or weaken it.

### Scope and gate prerequisites

QA owns only independent validation and evidence.  It does not implement
product code, edit PLAN, stage, commit, merge, or publish.  TASK-0004 owns the
versioned Unix-domain IPC boundary, socket lifecycle, and authorization from
kernel-provided `SO_PEERCRED` identity.  It excludes Codex access/redaction,
sudo/PAM, push, audit, packaging, deployment, and release behavior.  TOTP and
lease policy from earlier tasks may be represented only by a controlled
backend test double; this task must not reimplement them.

QA starts only after the candidate is available and independent REVIEW has
recorded PASS with parser/peer-credential, SLOC, and no-compression evidence.
A missing REVIEW PASS blocks QA and merge.  QA PASS and REVIEW PASS are both
mandatory before the main Agent may merge.

### Acceptance and failure conditions

| ID | Independent criterion | Required observation |
| --- | --- | --- |
| Q4-01 | Send a well-formed request using the supported protocol version through a real Unix socket from an authorized peer. | The server accepts the envelope, dispatches exactly once to the controlled backend, and returns only the defined versioned response. This is a fixture sanity check, not approval of later access policy. |
| Q4-02 | Send truncated, invalidly framed/encoded, structurally malformed, missing-required-field, trailing-data, or otherwise malformed requests. | Every request fails closed with no backend call, no panic, and no state/authority change. |
| Q4-03 | Send a request with an unsupported/wrong protocol version. | It fails closed with no backend call and no downgrade or alternate-version interpretation. |
| Q4-04 | Send a request at the maximum permitted size and one byte over it; also exercise a declared-size/body mismatch where the framing permits. | The in-limit request is parsed according to the protocol; every oversize or inconsistent request is rejected before backend dispatch and without unbounded allocation/read. |
| Q4-05 | Connect to an absent socket, stop the server while clients exist, and attempt a request before readiness/after shutdown. | Unavailable IPC never yields an allow/success result. Shutdown closes the listener and connections as defined, leaves no server accepting requests, and permits a clean subsequent start. |
| Q4-06 | Connect through a real Unix socket from authorized and unauthorized OS peer identities and attempt to forge identity in request data. | Authorization uses `SO_PEERCRED`, not caller-supplied fields. Unauthorized peers fail closed before backend dispatch; forged identity has no effect. If the QA environment cannot create the required UID peers, classify the environment limitation and use the strongest real-kernel credential fixture available; do not claim full PASS without approved disposition. |
| Q4-07 | Exercise socket creation, normal close, stale-socket recovery, repeated start/stop, and startup failures at the configured path. | Lifecycle is deterministic and fail closed: only an actual socket owned by the intended lifecycle may be removed/replaced; regular files, symlinks, wrong-type nodes, unsafe parent/path conditions, or conflicting live listeners are not clobbered. Permissions/ownership are restrictive before serving. Failure leaves no falsely ready or partially usable endpoint. |
| Q4-08 | For every rejection in Q4-02 through Q4-07, instrument a backend spy and observe response/error behavior. | Backend call count remains zero, no lease/authority transition occurs, and diagnostics do not turn malformed input or claimed identity into authority. |
| Q4-09 | Inspect the candidate diff and exported behavior. | No Codex access/redaction, sudo/PAM, push, audit, package/delivery, TOTP policy, or unrelated later-task work is added. Absence of those later features is not a QA failure. |
| Q4-10 | Count cumulative production source and inspect readability. | Exact cumulative production SLOC is `<=650`; source is idiomatic and gofmt-formatted with no semicolon/one-line packing, collapsed error handling, cryptic names, removed security comments, or functions combined solely to evade the ceiling. |

Q4-01 through Q4-10 are P0.  A skipped P0 is FAIL unless a QA-plan defect,
requirement gap, or environment issue is explicitly classified and the main
Agent approves the disposition before merge.  In particular, a generic error
alone is insufficient evidence: fail-closed cases must also prove backend
noncall and absence of authority/state change.

### Independent execution and evidence

Use real Linux Unix-domain sockets for parser, framing, lifecycle, and
`SO_PEERCRED` checks; an in-memory transport alone cannot satisfy this plan.
Use bounded deadlines so unavailable or partial-message tests cannot hang.
Record OS/kernel, Go version, candidate revision, peer UID/GID/PID fixture,
socket path/type/mode/ownership, focused test names, exit codes, backend call
counts, and concise redacted output.

After DEV and REVIEW PASS, run from the candidate repository root:

```sh
go test ./...
go test -race ./...
test -z "$(gofmt -l .)"
go vet ./...
make check
```

`go test -race ./...` is required because listener shutdown, connection
handling, and backend dispatch can overlap.  Run `go vet ./...` when the
candidate is a valid Go module and the installed toolchain supports it.
`make check` is required when present in the repository contract.  Any
unsupported or unavailable command must be recorded and classified rather
than silently skipped.

QA independently runs the following exact cumulative production-SLOC command,
records every counted file and subtotal, and compares `TOTAL` to `<=650`.
Count nonblank/non-comment executable-source lines only; exclude tests,
generated/vendor files, task documents, workflows, and declarative
configuration.  A line containing executable code plus an inline comment
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

Preserve the command alongside the result so the count is reproducible.
Over-cap or compressed candidates FAIL; QA never fixes or compresses them.

### Failure classification

Classify each failed observation before attributing it.  Preserve its matrix
ID, command, peer/socket fixture, candidate revision, expected versus actual
result, backend call count, and minimal reproduction.

| Classification | Meaning | Disposition |
| --- | --- | --- |
| `implementation_defect` | Candidate violates an approved TASK-only criterion, boundary, SLOC ceiling, or readability guardrail. | FAIL; return to DEV with reproduction evidence. |
| `regression` | Previously passing behavior outside the new contract is broken by the candidate. | FAIL; return to DEV with baseline comparison. |
| `qa_plan_defect` | QA procedure, fixture, or expectation is invalid or contradicts the approved task while product behavior satisfies it. | Pause attribution; amend and re-approve QA plan. |
| `requirement_gap` | TASK lacks a decision required to judge an observed IPC or lifecycle behavior. | Return to PLAN/task authority; do not infer DEV fault. |
| `environment_issue` | OS permissions, UID fixture, kernel/toolchain, or test infrastructure prevents valid assessment and is not candidate behavior. | Record blocker and rerun in a corrected environment; do not assign DEV fault. |

QA PASS requires all P0 observations, real-Unix focused evidence, full Go and
repository checks, race evidence, exact `TOTAL <=650`, scope review, and
no-compression review.  Any QA FAIL returns to its responsible gate and never
authorizes merge.

## PLAN reconciliation

This table was added after the TASK-only criteria above were fixed and
`PLAN.md` was read.  PLAN details are accepted only where they match or safely
refine that baseline.

| TASK-only criterion | PLAN comparison | Reconciliation decision |
| --- | --- | --- |
| Versioned, malformed, wrong-version, and oversize requests fail closed with backend noncall (Q4-02–Q4-04, Q4-08). | Matches. PLAN defines a 4-byte length, strict JSON/version 1, 4096-byte bound, one request per connection, and enumerated malformed cases before dispatch. | Adopt PLAN framing/version/size values as focused fixtures; retain zero backend calls and no state change as QA evidence. |
| Unavailable, partial, stalled, and shutdown-time IPC fails closed (Q4-05, Q4-08). | Matches and refines. PLAN adds finite deadlines, handler bound 16, context/Close shutdown, active-connection close/wait, and no post-shutdown dispatch. | Test these bounded lifecycle cases through real Unix sockets and under `-race`. |
| Kernel identity is authoritative; unauthorized peers and credential lookup failure deny before backend (Q4-06, Q4-08). | Partial match. Implementation uses real `SO_PEERCRED`; PLAN requires real current-UID success but proposes injected wrong-UID/syscall-error tests and explicitly plans no distinct-user fixture. | **PLAN evidence defect/gap:** injection proves branch behavior but does not independently prove rejection of a real kernel-reported unauthorized UID. QA requires a real distinct-UID Unix peer when safely available. If unavailable, record `environment_issue` and obtain main/task-authority disposition or a PLAN amendment; do not silently claim full Q4-06 PASS from injection alone. |
| Socket lifecycle and path safety fail closed (Q4-07). | Matches and refines. PLAN requires absolute clean paths, `Lstat` parent traversal, safe ownership/mode, refusal of every pre-existing final node, `0600`, idempotent close, and identity-safe cleanup. | Treat “stale-socket recovery” as a refusal test for pre-existing sockets; recovery is only clean restart after this server removes its own unchanged socket. Never require deletion of an unowned stale path. |
| Scope excludes later access/redaction, sudo/PAM, push, audit, and delivery work (Q4-09). | Matches. PLAN also excludes CLI/daemon/config/TOTP and keeps lease semantics behind an injected backend. | Diff outside the approved `internal/ipc`/module-test boundary or added later-task behavior fails scope review. Generic transport errors are part of IPC fail-closed behavior, not a new redaction subsystem. |
| Independent REVIEW prerequisite, Go checks, real Unix tests, race evidence, and exact cumulative `<=650` SLOC/no-compression gate (Q4-10). | Matches. PLAN provides the exact count, merged baseline 233, required Go commands, real socket tests, and independent REVIEW/QA evidence. | Use the exact command above. Require REVIEW PASS before QA, then independently rerun focused checks and count. `make check` remains required only when it exists and applies; do not add it as product scope. |

No other PLAN mismatch changes the TASK-only acceptance baseline.  The
unauthorized-real-peer evidence gap above must be resolved or explicitly
classified before QA PASS; it is not inherited as an acceptable omission.

### Final reconciliation after PLAN amendment

The amended PLAN resolves the recorded unauthorized-peer evidence gap without
changing the TASK-first acceptance criterion or erasing the prior finding and
disposition.  The real peer need not run as a separately created OS user: it
is unauthorized when its kernel-reported UID does not equal the server's
deliberately nonmatching `AllowedUID`.

| Prior gap | Amended PLAN evidence | Final decision |
| --- | --- | --- |
| Wrong-UID rejection was covered only through an injected credential, which did not prove rejection of a real kernel-reported peer identity. | Start the real Unix server with `AllowedUID` deliberately different from the current process UID; connect from that process through the real socket; obtain its actual UID through `SO_PEERCRED`; prove rejection before parsing/dispatch and backend call count zero. The private injection seam is limited to credential-extraction/syscall failure. | **Resolved / reconciliation PASS.** This is a real unauthorized-peer test because authorization compares the actual kernel credential to configured policy. QA must independently repeat it and record actual UID, configured `AllowedUID`, denial, and zero backend calls. |

Final PLAN reconciliation status: **PASS with no remaining PLAN gaps**.  This
does not pre-approve implementation evidence: Q4-01 through Q4-10, REVIEW
PASS, real-Unix focused tests, full Go/race checks, exact cumulative
`TOTAL <=650`, scope, and no-compression checks remain mandatory for QA PASS.

## Lap02 independent TASK/candidate re-estimation (before PLAN revision)

This re-estimation was recorded from TASK, the observed candidate, and the
existing QA baseline before reading the Lap02 PLAN revision.  It preserves all
earlier QA history and supersedes neither Q4-01–Q4-10 nor the resolved prior
PLAN gap.

### Candidate assessment and decision

Observed candidate evidence is cumulative production **617/650 SLOC**
(384 added production SLOC), 256 test SLOC, with basic Go checks passing.
That is 94.9% of the cap and leaves exactly **33 production SLOC**.  Passing
basic checks does not establish Q4-05, Q4-07, or the concurrency portion of
Q4-08 because replacement-safe cleanup, pre-existing-socket refusal,
saturation, and active-client shutdown evidence is missing.

There is also an observed `implementation_defect`: `handle` holds the server
mutex while calling `Backend.Handle`, whereas `Close` must acquire that mutex
before it can cancel the backend context and close active connections.  A
backend waiting for cancellation can therefore deadlock with `Close`.  This
violates deterministic fail-closed shutdown and cannot be accepted merely
because the current tests pass.

The `<=650` ceiling remains **credible but conditional**.  The missing
evidence is test-only and therefore does not increase production SLOC.  The
deadlock fix should be a small ordinary lock-scope correction; QA does not
pre-authorize a particular implementation.  The 33-line headroom may not be
used as a reason to pack statements, collapse error handling, weaken path or
shutdown controls, rename cryptically, remove security comments, or combine
functions solely for the count.

DEV must not resume until both the Lap02 PLAN amendment and this QA
re-estimation are approved.  On resume, stop DEV and return to PLAN/QA before
continuing if any of the following occurs:

- projected or measured cumulative production SLOC exceeds 650;
- the deadlock cannot be corrected in the remaining 33 production SLOC with
  normal gofmt-formatted, idiomatic lock ownership and explicit errors;
- a mandatory fail-closed lifecycle/control would need to be removed,
  weakened, or deferred;
- any compression guardrail is approached or violated; or
- remediation reveals a new lifecycle/authority decision not resolved by
  TASK and the approved PLAN.

### Exact mandatory remediation evidence

| ID | Required focused test/remediation evidence | Pass condition |
| --- | --- | --- |
| L2-01 | Backend-under-mutex/Close regression: use a backend that begins handling and waits for its context cancellation; concurrently call `Close` with bounded test time. | `Close` can acquire shutdown state, cancel the backend context, close the active connection, wait for the handler, and return without deadlock; backend executes at most once; run under `go test -race`. The test must fail against the observed lock scope or otherwise demonstrate the defect before accepting the fix. |
| L2-02 | Active-client shutdown: hold a real authorized Unix client with a partial or stalled frame, then call `Close`. | Shutdown returns within a bounded fixture deadline, the client is closed, backend count remains zero, no request dispatches after shutdown starts, and the owned unchanged socket is cleaned up. |
| L2-03 | Replacement-safe cleanup: after `Listen`, unlink the owned socket path and put a distinct regular file (and, where practical, a distinct socket) at the same path before `Close`. | `Close` reports the lifecycle error and leaves the replacement inode/content untouched; it never removes a path merely because its name matches. |
| L2-04 | Pre-existing socket refusal: bind a separate live Unix listener at the configured final path, then call `Listen`. | `Listen` fails closed without removing, closing, replacing, or disrupting the pre-existing listener; backend count remains zero. Retain the existing regular-file/symlink/unsafe-parent cases. |
| L2-05 | Handler saturation: hold all 16 handler slots with bounded real Unix connections, then attempt at least one additional connection/request and finally release the held clients. | Excess work is closed/denied with no backend dispatch and no hang; after controlled release, shutdown completes cleanly. Evidence must be deterministic and race-clean, not timing-only. |
| L2-06 | Re-run the full acceptance and size gates after remediation. | Focused tests plus `go test ./...`, `go test -race ./...`, gofmt check, supported `go vet ./...`, applicable `make check`, and `git diff --check` pass; exact cumulative command reports `TOTAL <=650`; manual no-compression and later-scope exclusion checks pass. |

Lap02 cannot receive QA PASS from the current candidate.  The correct current
classification is `implementation_defect` for the Close deadlock and missing
required acceptance evidence, not an environment or QA-plan failure.

## Lap02 PLAN reconciliation

This reconciliation was appended only after the independent Lap02 assessment
above was fixed and the revised PLAN was read.

| Independent Lap02 finding | Revised PLAN comparison | Reconciliation decision |
| --- | --- | --- |
| Candidate is 617/650 production SLOC with 33 lines of headroom; tests do not count, and basic passes do not make it gate-ready. | Exact match: PLAN records 233 baseline + 384 IPC = 617, 256 test lines, the >90% trigger, and the same missing evidence. | PASS. Retain independent exact recount after remediation; do not treat prior check passes as acceptance. |
| Backend-under-mutex creates a Close/cancellation deadlock and must be fixed without weakening shutdown admission. | Matches and supplies a coherent lock contract: classify an already tracked handler under the mutex, unlock before `Backend.Handle`; `Close` sets closing/cancels/closes under the mutex, then unlocks and waits. | PASS. L2-01 remains the controlling regression test. QA judges externally observable bounded cancellation, no post-shutdown admission, and race cleanliness rather than prescribing source layout beyond the approved invariant. |
| Replacement-safe cleanup, pre-existing live socket, saturation, and active-client shutdown evidence is mandatory. | Matches. PLAN requires all four, including a context-blocking backend, 16 held partial clients plus excess denial, intact pre-existing listener, and replacement identity preservation. | PASS. L2-02–L2-05 remain mandatory; PLAN's blocking-backend and saturation-close fixtures jointly cover active-client shutdown. |
| `<=650` is credible only if the fix remains idiomatic and no mandatory control is shed or compressed. | Matches and strengthens the stop rule. PLAN identifies production-only client helpers `writeRequest`/`readResponse` as test-only scope that may legitimately move to `_test.go`, targets 600–625, and stops at projected/measured >640 to retain at least 10 lines of review margin. | PASS, conditional. Moving genuinely test-only client fixtures out of production is scope correction, not compression. QA adopts the stricter 640 resume-stop threshold; 650 remains the unconditional acceptance ceiling. |
| New scope/requirement decisions require replanning, and DEV may not resume on a unilateral estimate. | Matches. Remaining DEV is limited to the lock/scope correction and named tests/checks; new production feature, protocol operation, client API, dependency, or approved-path expansion triggers stop. | PASS. Both revised PLAN and this Lap02 QA re-estimation require main approval before DEV resumes. |

Lap02 PLAN reconciliation result: **PASS with no remaining planning gaps**.
The production ceiling is credible without compression because the forecast
uses removal of out-of-scope production client helpers plus a small ordinary
lock-scope correction, while test growth is excluded from production SLOC.
This is a planning decision only, not candidate QA PASS.

DEV resume conditions are exact: main approval of both Lap02 PLAN and QA_PLAN,
projected cumulative production `<=640`, unchanged approved file/scope
boundary, and no compression/control shedding.  DEV must stop immediately on
any breach.  After remediation, REVIEW PASS must precede independent QA, which
must execute L2-01–L2-06 and Q4-01–Q4-10 and confirm exact cumulative
`TOTAL <=650` before merge can be considered.
