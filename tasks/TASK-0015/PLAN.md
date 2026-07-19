# PLAN — TASK-0015: secure daemon review-gap replacement

## Decision and boundary

**PASS — implementation plan approved pending the independent TASK-first
`QA_PLAN.md` gate.** This is one counted 30-minute Lap, not a continuation of
TASK-0014.  The replacement owns only:

| Path | Ownership |
| --- | --- |
| `cmd/codex-authority-broker/main.go` | Linux seed admission, runtime/listener ordering, daemon lifecycle. |
| `cmd/codex-authority-broker/main_test.go` | Deterministic descriptor, schema, lifecycle, restart, and client evidence. |

Do not change backend, IPC, protocol, lease, CLI, APIs, sudo, installation,
persistence, audit, credentials, Git, push, or release behavior.  A need for
one is a security-boundary/scope stop and split.  DEV is `dev-luna`/
`luna-xhigh`; PLAN, independent REVIEW, and QA are separate `Terra/medium`
roles.  Children do not stage, commit, merge, or write `.git`; Main alone owns
Git closure.

## Fixed implementation contract

Implement private dependency seams in `main.go` for descriptor operations,
final-reader conversion, runtime construction, and listen/serve/close.  They
exist solely for deterministic tests and must not alter `backend` or `ipc`
exports.

1. `main` uses `signal.NotifyContext` for SIGINT/SIGTERM and calls `run`.
   `run` loads and validates the seed; calls `backend.New(decodedSecret)`;
   **wipes the caller-owned decoded buffer immediately when that factory call
   returns, before inspecting its result or calling listen**; then listens.
   A deferred wipe still covers every pre-construction/error return.  The
   factory's successful private copy is the only retained secret ownership;
   no broader Go-memory-erasure claim is made.
2. Walk the fixed seed path from `/` descriptor-relatively: root and every
   literal parent use `O_DIRECTORY|O_RDONLY|O_NOFOLLOW|O_CLOEXEC`, are `fstat`ed
   as directories, and are closed before advancing.  The final `openat` uses
   `O_RDONLY|O_NOFOLLOW|O_CLOEXEC`; admit only a regular UID-0 file whose
   `mode&07777 == 0600` and size is `1..maxSeedBytes`.  Every root/openat/
   stat/close/reader failure, symlink, type/owner/mode/size denial maps to the
   same private `errSeed`.  Once `os.NewFile` conversion succeeds, the reader
   owns the final descriptor and closes it exactly once.
3. Read exactly the declared bounded length (reject short, read-error, and
   extra input) and decode one bounded JSON object: exactly unique
   `totp_secret_b64` and `allowed_uid`, no trailing value; positive uint32 UID;
   canonical standard base64 decoding to a nonempty, bounded secret.  Reject
   malformed, duplicate, unknown, missing, empty, oversized, wrong-type, and
   noncanonical inputs.  Wipe owned parser/secret buffers where applicable;
   errors and any captured diagnostics must expose no seed, marker, path, UID,
   descriptor, token, or secret detail.
4. Construction precedes listen.  A runtime-factory failure makes zero listen
   calls.  A listener error closes a non-nil returned server first, then the
   newly built runtime.  On successful listen, `Serve(ctx)` runs; one idempotent
   shutdown path closes server and runtime exactly once.  A nil/normal Serve
   result is clean only after cancellation; unexpected Serve or close errors
   fail.  Existing server identity-checked unlink is relied on unchanged.
   Each run rereads the seed and creates fresh runtime/server state; a denied
   restart never listens.

## Deterministic acceptance map

All fixtures use temporary sockets and injected descriptor-relative operations;
they never read a production seed, mutate root-owned metadata, install a
service, or elevate privileges.  Barriers/channels and bounded contexts replace
sleep-based proofs.

| Test name(s) | Implementation evidence required |
| --- | --- |
| `TestRunWipesCallerSecretBeforeListen` | Listener-order barrier observes the exact factory input buffer zeroed before its first listen call; factory success and failure returns retain error-path wiping. |
| `TestLoadSeedDescriptorWalk`, `TestLoadSeedDescriptorErrors`, `TestLoadSeedFinalReaderOwnership` | Root open/close and parent/final openat/fstat/close terminals; parent/final symlink no-follow denial; final regular/type, UID, every mode including special bits, all size boundaries (`0`, min, max, max+1), short/read-error, and reader-close ownership without a second final-fd close. |
| `TestLoadSeedSchema` | Valid minimum and maximum secret; malformed, duplicate, unknown, missing, empty, oversized, wrong-type/trailing JSON, invalid UID forms, and invalid/noncanonical/oversized base64; each denial is `errSeed` and marker-redacted. |
| `TestRunConstructionAndListenFailures`, `TestRunServeAndCloseFailures` | Runtime-factory error; construction-before-listen; listener error with a non-nil server closed before runtime; no Serve after listen failure; unexpected Serve return and server-close error fail closed. |
| `TestRunSignalsAndShutdown`, `TestRunActiveConcurrentRepeatedShutdown` | SIGINT and SIGTERM cancellation; active and concurrent clients; idempotent/repeated shutdown, exactly-once close/unlink, and no false clean result before cancellation. |
| `TestRunSocketReplacementAndRestart`, `TestRunRestartWithoutSeed` | Socket identity replacement is preserved (replacement not unlinked); a successful fresh restart uses new seed/runtime/server; missing/invalid subsequent seed makes zero listen calls. |
| `TestBrokerClientOTPAndMalformedRequest` | Existing IPC client obtains a valid ready/OTP outcome and malformed input denies without changing TASK-0013 runtime or IPC behavior. |

DEV must add explicit test cases rather than collapse these rows into an
untraceable broad test.  QA_PLAN must retain this exact mapping (or stricter
named cases) before DEV starts.

## Counted Lap and verification

Preflight before the counted Lap: approved PLAN and independent QA_PLAN,
Linux/Go 1.23, merged TASK-0013 base, only owned paths eligible, and a
socket-capable fixture environment.  Socket creation denial is classified once
as `environment` with its exact null reason; it does not waive injected tests
or permit a product PASS.

| By minute | Gate and evidence |
| ---: | --- |
| 5 | Main starts DEV after both planning gates; DEV implements fixed seams/order first and runs focused broker tests. |
| 20 | DEV completes candidate and every map row; formats, counts SLOC, and records focused/race/full/static evidence. |
| 25 | Independent REVIEW inspects complete candidate, reruns required checks, and specifically rejects late wiping, descriptor leaks/following, disclosure, ordering, lifecycle, scope, or compression. |
| 30 | Independent QA reruns the mapped matrix and regressions; Main alone performs scope/hook/Git closure after both PASS results. |

Required commands at the appropriate independent gates are:

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease
go test -count=1 -race ./cmd/codex-authority-broker
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./...
go vet ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
```

Record UTC start/end, `active_ms`, `wait_ms`, retries, classification
(`pass`, `implementation_defect`, `planning_defect`, `environment`, or
`regression`), redacted null reason, mapped test results, and independent SLOC
counts. Planning plus pure wait targets at most 20% of the counted interval.

## SLOC and stop/retry rules

The local forecast is +280 against merged 922 (cumulative 1202).  It is a
warning for Main judgment, not automatic reapproval: preserve ordinary readable
Go, named bounds, separate cleanup/error paths, comments, and full tests.
The mandatory-v1 global target is 1500 and the unconditional global hard limit
is 1800.  Stop immediately for security-boundary change, scope expansion,
unreadable compression, target overflow requiring the ordered shedding audit,
or hard-limit risk; never shed seed/lifecycle controls or test evidence.

Lap 2 is permitted only if **all four** facts are recorded: concrete estimated
residue, no redesign/research, exactly one or two classified failure causes,
and a demonstrable fix within its first 20 minutes.  Otherwise split.  The same
cause permits at most one replan, and there is no Lap 3.

## Planner evidence

| Item | Evidence |
| --- | --- |
| Changed path | `tasks/TASK-0015/PLAN.md` only |
| Source lineage | TASK-0014 final REVIEW FAIL: P-001 immediate caller-buffer wipe and P-002 missing deterministic coverage; TASK-0015 contract preserves both as mandatory replacement acceptance. |
| State | PLAN ready for independent TASK-first QA planning; DEV has not started. |
| Accounting | `active_ms=unavailable`, `wait_ms=0`, `retries=0`, classification `pass`; planner runtime did not expose a reliable turn-start timestamp, so active duration is not inferred. |
