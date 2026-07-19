# Independent product review — TASK-0015

**FAIL — `implementation_defect`.** This independent REVIEW inspected the
complete broker candidate and ran the required checks.  The production
caller-buffer wipe now occurs immediately after `makeRuntime` returns and
before `listen`, but the candidate is not acceptance-complete: the real
socket-capable valid-OTP test fails, and several required named acceptance
tests do not prove their stated rows.  No product or non-review evidence file
was changed by REVIEW.

## Blocking findings

| ID | Classification | Finding and required resolution |
| --- | --- | --- |
| P-001 | implementation_defect | In a socket-capable execution supplied to Main, `TestRunServesValidOTPRequest` fails because the OTP response has `OK=false`; its PLAN alias `TestBrokerClientOTPAndMalformedRequest` consequently fails too.  This is the mandatory existing-client valid OTP acceptance, not a sandbox waiver.  Make the broker/client integration deterministic and demonstrate a successful valid OTP through the broker.  The test currently uses a real clock and creates a code for `time.Now()/30+1`, while the runtime's lease state records a boot replay floor at construction; the real integration result is therefore the controlling evidence. |
| P-002 | implementation_defect | The PLAN-level aliases do not prove their full named rows.  `TestLoadSeedDescriptorErrors` executes only root/parent/final **open** errors, omitting the required stat and close terminals; `TestLoadSeedSchema` forwards to an aggregate that omits the valid maximum and oversized-input cases; `TestRunConstructionAndListenFailures` omits its construction/configuration observations.  `TestRunSocketReplacementAndRestart` calls a socket test first, so `t.Skip` prevents its restart proof from running; the client alias similarly skips the malformed branch.  Keep independently runnable exact names, but make every PLAN aggregate execute/prove every listed subcase (or replace it with explicit subtests). |
| P-003 | implementation_defect | `TestLoadSeedRejectsNon0600Mode` exercises only mode `0644`; its row requires every `mode&07777 != 0600`, including special bits.  The older broad test has one special-bit case, but that does not make the required exact-name test prove its row.  Parameterize the exact test over permission and special-bit mismatches. |
| P-004 | implementation_defect | Restart acceptance is incomplete. `TestRunRestartsWithFreshSeed` uses the same seed document for both runs, so it does not demonstrate rereading a changed seed; `TestRunFailsClosedOnRestartMissingSeed` has no successful preceding run and therefore does not prove that a later missing seed cannot reuse prior state.  Run a successful first invocation, change/remove the next seed, and prove zero construction/listen on restart; separately prove changed seed/runtime/server state on a successful restart. |

## Independent source review

- `run` loads the seed, calls `makeRuntime(secret)`, wipes that exact caller
  slice before checking the result or invoking `listen` (main.go:106-115).
  Its factory-error path retains the wipe, and the listener-order test observes
  the factory-owned slice zeroed before its listener runs.
- The descriptor-relative root/parent/final walk uses no-follow and CLOEXEC;
  it stats and closes parents, validates regular/root-owned
  `mode&07777 == 0600` final metadata and bounded size, and transfers the
  final descriptor to the successfully created reader without a second raw
  close.  Generic `errSeed` preserves redaction.  Strict JSON/base64/UID
  admission and construction-before-listen are present by source inspection.
- The lifecycle paths handle nil factory/listener dependencies, close a
  non-nil server returned alongside a listen error before runtime close, and
  treat a normal/nil Serve result as clean only after cancellation. SIGINT and
  SIGTERM use `signal.NotifyContext`; production server identity protection is
  unchanged in `internal/ipc`.
- Production source is readable ordinary Go; no compression or out-of-scope
  production/API change was found. Canonical nonblank/noncomment broker SLOC
  is 278, for cumulative 1200 against the documented 1202 boundary. Physical
  file length is 307; tests are 1014 physical LOC.

## Command evidence

| Command | Result | Classification / observation |
| --- | --- | --- |
| `make check` | FAIL | Environment: this worktree has no `check` target (`No rule to make target 'check'`). It is not a product substitute. |
| focused packages without cache override | FAIL | Environment: Go build cache under `/home/ubuntu/.cache/go-build` is read-only. |
| focused packages with fresh cache and `GOFLAGS=-buildvcs=false` | FAIL | Environment socket denial: CLI and IPC tests fail with `socket: operation not permitted`/`ipc: server unavailable`; broker/backend/lease pass. This does not waive P-001. |
| `go test -count=1 -race ./cmd/codex-authority-broker` (fresh cache, VCS disabled) | PASS | No race reported. |
| every declared broker top-level test selected by exact anchored name | PASS with 5 SKIPs | Descriptor/lifecycle unit rows pass; socket identity, valid OTP, malformed client, and their aggregate aliases skip here due Unix socket EPERM. The skips are environment evidence only and leave those acceptances unproven locally. |
| socket-capable focused evidence obtained by Main | FAIL | `TestRunServesValidOTPRequest`: response `OK=false`; its PLAN client alias fails. Product acceptance failure P-001. |
| full `GOFLAGS=-buildvcs=false GOCACHE=$(mktemp -d) go test -count=1 ./...` | FAIL | Same reproducible Unix socket EPERM/server-unavailable environment failures in CLI/IPC. |
| `go vet ./...` (fresh cache) | PASS | No findings. |
| `gofmt -l $(find cmd internal -type f -name '*.go' -print)` | PASS | No output. |
| `git diff --check` | PASS | No output; broker candidate is untracked, so it was inspected directly. |

## Scope and accounting

- REVIEW-owned path changed: `tasks/TASK-0015/REVIEW_RESULT.md` only.
- Candidate ownership inspected: `cmd/codex-authority-broker/main.go` and
  `cmd/codex-authority-broker/main_test.go`; unrelated pre-existing dirty
  backlog/TASK files were not modified or attributed to this candidate.
- `active_ms=unavailable`, `wait_ms=0`, `retries=0`; no reliable REVIEW
  start timestamp is exposed, so duration is not inferred.
- Overall classification: `implementation_defect`. Separate environmental
  evidence: absent Make target, read-only default Go cache, and Unix socket
  `EPERM`; VCS stamping was proactively disabled for reproducible Go checks.
- Completed: 2026-07-19T10:23:05Z.

---

## Bounded P-001–P-004 correction re-review — PASS

**PASS (`pass`) for the bounded test-only correction.** `main.go` is unchanged;
the re-review inspected only the P-001–P-004 changes in `main_test.go`. The
four blocking findings above are closed. Earlier FAIL evidence remains as the
record of the rejected candidate and is superseded for this corrected scope.

### Finding closure

| Finding | Result | Independent evidence |
| --- | --- | --- |
| P-001 deterministic valid/malformed client integration | CLOSED | The test now uses real `ipc.Client`, `ipc.Listen`, framing, peer-UID admission, Serve, and Close with a deterministic backend that accepts only the production 17-byte `{"code":"NNNNNN"}` OTP payload shape and denies malformed payloads. This intentionally separates transport/syntax evidence from clock-sensitive TOTP generation. Independent `internal/backend` and `internal/lease` regressions pass and retain real OTP decoding, verifier/window/replay, readiness, and activation semantics. Main's socket-capable rerun of broker/backend/ipc/lease passes, so the real transport branches are executed rather than waived. |
| P-002 aggregate coverage and skip isolation | CLOSED | PLAN aggregates now use named `t.Run` subtests. Descriptor errors include root, parent, and final open/stat/close plus reader, short-read, and read-error cases; schema includes maximum and oversized cases; construction/listen includes runtime, ordering, configuration, and listener error. Socket and restart plus valid and malformed client branches are siblings, so a socket-only skip cannot suppress a non-socket sibling. Local output confirms `socket-replacement` SKIP while `restart` PASS. |
| P-003 exact `07777` mode rejection | CLOSED | `TestLoadSeedRejectsNon0600Mode` enumerates all 4096 values from `0000` through `07777`, excludes only `0600`, and asserts `validSeedFile` rejects each remaining permission/special-bit combination. |
| P-004 restart behavior | CLOSED | The successful restart uses distinct seed documents, runtimes, and servers and verifies different retained private secret copies. The missing-seed case now completes one successful run, then injects root-open `os.ErrNotExist` and proves zero runtime-factory and listen calls on the second run. |

### Re-review commands and metrics

| Check | Result |
| --- | --- |
| Corrected exact broker tests with fresh cache and VCS disabled | PASS; local Unix-socket-only cases SKIP with `ipc: server unavailable`; aggregate sibling restart still PASSes, proving skip isolation. |
| `go test -count=1 ./internal/backend ./internal/lease` with fresh cache and VCS disabled | PASS. |
| Main socket-capable broker/backend/ipc/lease rerun | PASS (supplied independent execution evidence). |
| `gofmt -l cmd/codex-authority-broker/main.go cmd/codex-authority-broker/main_test.go` | PASS, no output. |
| `git diff --check` | PASS, no output. |

- Canonical broker production SLOC remains **278**, cumulative **1200**;
  corrected test file is **1057 physical LOC**.
- Changed REVIEW path: `tasks/TASK-0015/REVIEW_RESULT.md` only. No product,
  Git, or operations write was performed by REVIEW.
- `active_ms=unavailable`, `wait_ms=0`, `retries=0`; classification `pass`.
- Completed: 2026-07-19T10:30:13Z.
