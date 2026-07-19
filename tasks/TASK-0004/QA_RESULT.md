# QA RESULT — TASK-0004

## Decision: PASS

Independent pre-merge QA is PASS for TASK-0004.  `REVIEW_RESULT.md` revision
2 records PASS and resolves the prior identity-safe lifecycle finding.  This
result authorizes no Git operation or merge; those remain main-Agent duties.

## Candidate, environment, and scope

- Candidate baseline: `50590e3657471868c0f4deae8b0e3a2d903564f6`
  on `task/TASK-0004-versioned-peercred-ipc`.
- Environment: Linux `6.8.0-36-generic` x86_64, Go `go1.23.12 linux/amd64`,
  effective UID/GID `1000/1000`.
- Real-peer fixtures used kernel credentials over actual temporary Unix
  sockets.  Authorized `AllowedUID=1000` dispatched once; deliberately
  mismatched `AllowedUID=1001` denied with zero backend calls.  Peer PID is
  ephemeral and is neither used nor logged; authorization trusts only the UID
  returned by `SO_PEERCRED`.
- Created endpoints were absolute temporary Unix-socket paths, owned by the
  effective UID and mode `0600` before serving.
- Candidate changes are limited to the four approved `internal/ipc` files and
  TASK-0004 evidence.  No lease/TOTP change, Codex access/redaction, secret
  input, CLI/daemon/service, sudo/PAM, push, audit, packaging/release,
  dependency, or other later-task behavior was added.  Production emits no
  logs.

## TASK-first acceptance matrix

| ID | Result | Independent evidence |
| --- | --- | --- |
| Q4-01 | PASS | `TestServerUsesRealPeerCredentials` passed through a real Unix socket: the supported version dispatched exactly once and returned the defined versioned success response. |
| Q4-02 | PASS | Strict malformed-frame tests passed for truncated header/body, zero length, invalid JSON, unknown/missing operation/field, and trailing JSON. Real partial-frame denial returned failure with zero backend calls and no panic. |
| Q4-03 | PASS | Wrong protocol version denied with no alternate interpretation; static dispatch order confirms rejection precedes backend invocation. |
| Q4-04 | PASS | Exact 4096-byte payload parsed successfully; 4097-byte declaration and declared/body mismatch denied before backend dispatch. Header bounds are checked before body allocation. |
| Q4-05 | PASS | Absent/closed endpoint dialing denied. Active partial connections closed during bounded shutdown, the listener ceased accepting, unchanged owned paths were cleaned, repeated `Close` passed, and no post-close dispatch occurred. |
| Q4-06 | PASS | Real kernel peer UID 1000 succeeded only with matching policy; the same real peer denied against configured UID 1001 with zero backend calls. Injected credential-read failure separately denied. Strict unknown-field rejection prevents forged request identity from affecting authorization. |
| Q4-07 | PASS | Unsafe relative/symlink/world-writable-parent and pre-existing regular/live-socket paths were refused without clobbering. Normal close removed only the owned unchanged socket. Replacement and missing paths returned `ErrLifecycle`; replacement inode/content remained intact. Mode was `0600`. |
| Q4-08 | PASS | Recording-backend assertions remained zero for credential, malformed/partial, path, saturation, and shutdown denials. Only the authorized request and deliberate backend-error fixture invoked the backend, each exactly once. |
| Q4-09 | PASS | Diff and exported-surface inspection found only bounded versioned IPC, backend injection, and lifecycle behavior within the approved scope. |
| Q4-10 | PASS | Exact cumulative production SLOC is 600, below both the 650 ceiling and Lap02's `>640` stop threshold. Formatting and manual readability/no-compression review passed. |

## Lap02 matrix

| ID | Result | Independent evidence |
| --- | --- | --- |
| L2-01 | PASS | `TestCloseCancelsBlockingBackend` passed under `-race`: shutdown cancelled the admitted blocking backend, closed the client, waited without mutex deadlock, invoked the backend once, and admitted no post-close request. |
| L2-02 | PASS | `TestCloseStopsActivePartialClient` passed: bounded shutdown closed a real stalled partial client, dispatched zero backend calls, removed the unchanged owned socket, and did not hang. |
| L2-03 | PASS | Replacement, missing-path, and identity-checked-removal tests passed. `Close`/cleanup returned `ErrLifecycle` and preserved replacement content/inode rather than deleting by pathname. |
| L2-04 | PASS | A separate live Unix listener at the final path was refused and retained with unchanged socket identity; backend calls remained zero. |
| L2-05 | PASS | Sixteen real partial clients occupied all handler slots; the seventeenth was closed without response or backend dispatch. Controlled shutdown completed cleanly under race execution. |
| L2-06 | PASS | Focused/full test, race, vet, gofmt, diff, cumulative size, scope, and no-compression gates all passed. No Makefile exists, so `make check` is inapplicable under the approved PLAN. |

## Command and timing evidence

The known sandbox `bind(AF_UNIX) -> EPERM` limitation was already classified
by REVIEW as `environment_issue`; QA therefore ran the approved real-socket
commands directly with elevation.  Each command ran once.  Candidate retries:
0; flaky retries: 0; environment retries: 0.

```text
GOCACHE=/tmp/codex-authority-broker-task4-qa-cache go test -count=1 -v ./internal/ipc -> PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-qa-cache go test -count=1 ./... -> PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-qa-cache go vet ./... -> PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-qa-cache go test -count=1 -race ./... -> PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-qa-cache go test -count=1 -race -v ./internal/ipc -run '<Lap02 focused tests>' -> PASS (elevated; 7 tests)
test -z "$(gofmt -l .)" -> PASS
git diff --check -> PASS
```

Measured QA command time was `active_ms=18606` (18,577 elevated plus 29
local).  Approval/tool wait observed outside command execution was
`wait_ms=0`; one elevation approval was required.

## Size and readability evidence

The exact approved cumulative production-SLOC command reported:

```text
96 internal/ipc/protocol.go
271 internal/ipc/server_linux.go
173 internal/lease/lease.go
60 internal/lease/totp.go
TOTAL 600
```

`600 <= 650` and `600 < 640`; 50 task-ceiling lines and 40 Lap02 stop-rule
lines remain.  IPC tests total 496 physical lines (`protocol_test.go` 92,
`server_linux_test.go` 404), or 463 nonblank/non-line-comment lines (84 and
379 respectively).

Manual review found normal gofmt-formatted, idiomatic control flow and named
errors.  No semicolon/one-line packing, collapsed error handling, cryptic
naming, removed security comments, weakened lifecycle/shutdown control, or
function combination solely to reduce the count was found.

## Classification

No candidate failure occurred.  The prior lifecycle `implementation_defect`
is resolved and independently covered by Q4-07/L2-03.  The sandbox Unix-bind
restriction remains a known `environment_issue` with an approved elevated
fixture; it did not block assessment.  No regression, QA-plan defect, or
requirement gap is classified.
