# REVIEW RESULT — TASK-0004 (attempt 2)

## Decision: FAIL

The remediated candidate passes its checked-in elevated real-socket, full,
vet, format, race, size, and readability checks, including the Lap02 shutdown
and saturation tests.  It nevertheless violates the approved fail-closed
socket lifecycle invariant when the owned pathname disappears, and its
constructor failure path can remove an unverified replacement pathname.  This
is an `implementation_defect`; TASK-0004 must return to DEV and is not ready
for QA or merge.

## Finding

### Major — lifecycle cleanup is not identity-safe for missing/startup-raced paths

The approved PLAN requires cleanup to remove the path only when `Lstat` still
identifies the exact socket created by this server; a missing or replaced path
must be left untouched **and reported as a lifecycle error**.

`Server.cleanup` instead treats `os.IsNotExist` as success.  If the server's
owned socket pathname is removed before `Close`, shutdown reports `nil`, hiding
the lifecycle loss.  A reviewer-only real-socket test reproduced this:

```text
GOCACHE=/tmp/codex-authority-broker-task4-review-cache \
  go test -vet=off -overlay=/tmp/task4-reviewer-overlay.json \
  ./internal/ipc -run TestReviewerCloseReportsMissingOwnedSocket -count=1
--- FAIL: TestReviewerCloseReportsMissingOwnedSocket
    Close error = <nil>, want lifecycle failure
FAIL
```

There is a related unsafe constructor path.  After `ListenUnix`, the local
`fail` closure closes the listener and calls `os.Remove(config.Path)` without
checking device/inode identity.  If the pathname is removed/replaced between
bind and a failing `Chmod`/`identifySocket`, startup cleanup can delete the
replacement merely because its name matches.  This contradicts the same
identity-safe lifecycle rule and Q4-07/L2-03.

Action: make missing-path cleanup return `ErrLifecycle`, and make every
post-bind failure cleanup remove only a path proven to have the server's
socket identity.  If identity cannot be established or the path is missing or
different, leave it untouched and return `ErrLifecycle`.  Add focused tests
for missing-path `Close` and for post-bind failure/replacement cleanup; preserve
idempotent repeated `Close` behavior and the existing replacement-file test.

## Positive acceptance evidence

- Protocol framing is a four-byte big-endian length plus one strict JSON
  envelope, with version 1 and `MaxFrameBytes=4096`.  Zero, 4097, short
  header/body, invalid JSON, trailing JSON, unknown fields, missing/unknown
  operation, and wrong version deny.  The 4096-byte boundary accepts without
  allocating an over-bound body.
- Only a request that passes credential, frame, strict schema, version, and
  operation validation reaches `Backend.Handle`.  Static control flow and
  recording-backend tests show malformed/partial, credential failure,
  unauthorized, saturation, and path/lifecycle denials do not dispatch.
- The public listener installs `SyscallConn` plus
  `GetsockoptUcred(SOL_SOCKET, SO_PEERCRED)`.  Elevated real Unix-socket tests
  passed for kernel-reported UID 1000 with `AllowedUID=1000`, and rejected the
  same real peer with deliberately mismatched `AllowedUID=1001` and zero
  backend calls.  The private credential seam separately proved syscall
  extraction failure denial.
- Relative/unclean paths, symlink components, unsafe parent modes, regular
  pre-existing paths, and a pre-existing live socket are refused without
  replacement.  Bound socket mode is `0600`.  Normal close removes the
  unchanged owned socket; replacing it with a regular file produces
  `ErrLifecycle` and preserves replacement content.
- The 16-handler saturation test holds all slots with real partial clients;
  the seventeenth is closed without response or backend dispatch.  Partial
  active-client shutdown closes the client, dispatches nothing, and completes
  within its bound.
- `Close` sets closing, cancels the shared backend context, closes listener and
  active clients, and waits outside the mutex.  The blocking backend observes
  cancellation; `Close` completes without deadlock; repeated close is safe;
  and no additional backend dispatch is admitted after shutdown begins.  Race
  execution is clean.
- Transport/backend failures return only fixed generic errors or a generic
  `OK:false` response.  No request, backend-error, secret, credential, or later
  authority detail is logged or reflected.
- Changes are confined to the four approved `internal/ipc` files plus task
  evidence.  There is no Codex access/redaction, TOTP change, CLI/daemon,
  configuration input, sudo/PAM, push, audit, package, release, external
  dependency, or other later-task scope.
- Manual review found ordinary gofmt-formatted, idiomatic code.  There is no
  semicolon/one-line packing, collapsed error handling, cryptic naming,
  removed security comment, weakened control, or function combination solely
  for the SLOC cap.

## Execution evidence and classification

Candidate baseline: `50590e3657471868c0f4deae8b0e3a2d903564f6` on
`task/TASK-0004-versioned-peercred-ipc`.  Environment: Linux
`6.8.0-36-generic` x86_64, Go `go1.23.12 linux/amd64`, effective UID 1000.

The first sandbox execution could not bind any Unix socket and failed the
real-socket tests with `socket: operation not permitted`.  This was classified
as `environment_issue` (`bind(AF_UNIX) -> EPERM`), not candidate behavior.
One elevated suite retry was performed; there were no test-flake retries.

```text
GOCACHE=/tmp/codex-authority-broker-task4-review-cache go test -count=1 -v ./internal/ipc  PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-review-cache go test ./...                         PASS (elevated)
test -z "$(gofmt -l .)"                                                                      PASS
GOCACHE=/tmp/codex-authority-broker-task4-review-cache go vet ./...                          PASS
GOCACHE=/tmp/codex-authority-broker-task4-review-cache go test -race ./...                   PASS (elevated)
git diff --check                                                                              PASS
```

Lap02 focused tests passing in the elevated run include
`TestCloseCancelsBlockingBackend`, `TestCloseStopsActivePartialClient`,
`TestListenRefusesLiveSocketWithoutRemovingIt`,
`TestCloseLeavesReplacementUntouched`, and
`TestHandlerLimitRejectsSeventeenthClient`.

Measured command timing for this review was `active_ms=19543` (sandbox suite,
elevated retry, and focused reproduction) and `wait_ms=10777` (approval/tool
wait attributable to the two required sandbox escalations).  Retry count:
one elevated suite retry after the classified sandbox failure, plus one
separate reviewer-only focused reproduction; no candidate-test retry.

IPC test size is 451 physical lines total (`protocol_test.go` 92,
`server_linux_test.go` 359), or 420 nonblank/non-line-comment lines (84 and
336 respectively).  The exact approved cumulative production-SLOC command
reported:

```text
96 internal/ipc/protocol.go
270 internal/ipc/server_linux.go
173 internal/lease/lease.go
60 internal/lease/totp.go
TOTAL 599
```

`599 <= 650` passes the task ceiling and is below the Lap02 `>640` stop rule,
with 51 lines of ceiling headroom and 41 lines below the stop threshold.
`git status --short` before this result showed only untracked `internal/ipc`
and TASK-0004 PLAN/QA evidence.  No Makefile exists, so no repository check is
applicable under the approved PLAN.

No product, plan, test, lap log, or Git state was changed by this reviewer.
This review created only this evidence file; reviewer reproductions used
temporary `/tmp` overlay files.

---

## Revision 2 independent re-review

## Decision: PASS

The lifecycle remediation resolves the prior FAIL.  The current candidate
satisfies TASK-0004, the amended PLAN, and Lap02 L2 review criteria for strict
versioned framing, real `SO_PEERCRED` authorization, bounded fail-closed
dispatch, identity-safe socket lifecycle, deterministic shutdown, scope,
readability, and cumulative production SLOC.  Independent QA is still
required before any merge decision.

### Prior finding disposition: resolved

- `cleanup` now calls `removeOwnedSocket`, which returns `ErrLifecycle` for a
  missing path, any `Lstat`/identity failure, or a device/inode mismatch.  It
  calls `os.Remove` only after the current path identifies as the server's
  exact socket.
- After bind, `listen` identifies the created socket before any path removal.
  If initial identity cannot be established, it closes the listener, leaves
  the unverified path untouched, and fails.  The later `Chmod` failure path
  also uses `removeOwnedSocket`; normal `Close` uses the same helper.  Search
  confirms this helper contains the only production post-bind `os.Remove`.
- `TestCloseReportsMissingOwnedSocket` now proves missing-path `Close` returns
  `ErrLifecycle` without recreating or otherwise changing the path.
  `TestCloseLeavesReplacementUntouched` and
  `TestIdentityCheckedRemovalLeavesReplacement` prove a replacement returns
  `ErrLifecycle` and retains its inode content.  Pre-existing regular,
  symlink/unsafe-parent, and live-socket paths remain refused without removal.
- The revision-1 reviewer-only reproduction now passes:

```text
GOCACHE=/tmp/codex-authority-broker-task4-r2-review-cache \
  go test -vet=off -overlay=/tmp/task4-reviewer-overlay.json \
  ./internal/ipc -run TestReviewerCloseReportsMissingOwnedSocket -count=1
ok github.com/autotaker/codex-authority-broker/internal/ipc
```

### Full acceptance evidence

- The four-byte big-endian protocol remains strict at version 1 and
  `MaxFrameBytes=4096`.  Exact-bound input accepts; zero/4097, short
  header/body, invalid/trailing JSON, unknown fields, missing/unknown
  operation, and wrong version deny before backend dispatch.
- The public listener uses real `SyscallConn` and
  `GetsockoptUcred(SOL_SOCKET, SO_PEERCRED)`.  Elevated tests pass for actual
  UID 1000 with `AllowedUID=1000`; the same real peer is rejected with
  deliberate `AllowedUID=1001` and zero backend calls.  Injected credential
  failure independently denies without dispatch.
- The handler bound remains 16.  Sixteen held partial clients saturate the
  server; the seventeenth receives no response and causes no backend call.
  Partial and stalled active clients are closed during shutdown with no
  dispatch or handler leak.
- `Close` marks shutdown, cancels the shared backend context, closes listener
  and clients, unlocks, and waits.  The blocking backend cancellation test
  completes without mutex deadlock, repeated `Close` is safe, and no request
  is newly admitted after shutdown starts.  Race execution is clean.
- Backend and transport errors remain fixed/generic and do not reflect input
  or private backend details.  The package emits no logs.
- Product scope remains exactly the four approved `internal/ipc` files.  No
  lease/TOTP change, Codex access/redaction, secret input, CLI/daemon/service,
  sudo/PAM, push, audit, package/release, dependency, or later-task behavior
  was added.
- Manual review found normal gofmt-formatted, idiomatic control flow and named
  errors.  No semicolon/one-line packing, collapsed error handling, cryptic
  names, removed security comments, weakened controls, or functions combined
  solely for the cap were found.

### Revision-2 execution evidence

Candidate baseline remains `50590e3657471868c0f4deae8b0e3a2d903564f6` on
`task/TASK-0004-versioned-peercred-ipc`.  Environment: Linux
`6.8.0-36-generic` x86_64, Go `go1.23.12 linux/amd64`, effective UID 1000.

Because attempt 1 already classified sandbox Unix-socket bind as an
`environment_issue` (`EPERM`), revision 2 ran the real-socket commands directly
with elevation rather than repeating a known-invalid sandbox fixture.  Every
command ran once; candidate-test retries: zero, flaky retries: zero,
environment retries: zero.  One elevation approval was required.

```text
reviewer missing-path overlay test                                           PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-r2-review-cache go test -count=1 -v ./internal/ipc  PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-r2-review-cache go test ./...                         PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-r2-review-cache go vet ./...                          PASS (elevated)
GOCACHE=/tmp/codex-authority-broker-task4-r2-review-cache go test -race ./...                   PASS (elevated)
test -z "$(gofmt -l .)"                                                                      PASS
git diff --check                                                                              PASS
```

The elevated focused output explicitly passed real-peer success/mismatch,
credential failure, malformed/partial denial, unsafe/pre-existing paths,
blocking-backend cancellation, active partial shutdown, live-socket refusal,
replacement and missing-path cleanup, identity-checked removal, and
seventeenth-client saturation.

Measured revision-2 command timing was `active_ms=18779` (elevated checks plus
local format/count/diff checks) and `wait_ms=5147` (elevation approval/tool
wait).  IPC tests now total 496 physical lines (`protocol_test.go` 92,
`server_linux_test.go` 404), or 463 nonblank/non-line-comment lines (84 and
379 respectively).

The exact approved cumulative production-SLOC command reported:

```text
96 internal/ipc/protocol.go
271 internal/ipc/server_linux.go
173 internal/lease/lease.go
60 internal/lease/totp.go
TOTAL 600
```

`600 <= 650` passes and remains below the Lap02 `>640` stop threshold, leaving
50 lines to the task ceiling and 40 lines to that stop threshold.  Final
`git status --short` showed only the approved untracked `internal/ipc`
candidate, TASK-0004 PLAN/QA documents, and this review evidence.  No Makefile
exists, so no additional repository check applies under the PLAN.

No product, plan, test, lap log, or Git state was changed.  This reviewer
appended only this revision-2 evidence.
