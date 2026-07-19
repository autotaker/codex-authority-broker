# PLAN — TASK-0013: process-local ready/otp runtime assembly (Revision 2)

## Revision 2 disposition

Lap-1 DEV stopped correctly before tests: the read-only candidate measured
**191** canonical production SLOC, making cumulative **942**, above the
approved **925** stop.  `runtime_test.go` was absent, so the candidate was not
gate-ready.  Classify this stop as `planning_defect`: Revision 1 specified an
unnecessarily elaborate active-call registry and ambiguous dual constructor
surface while forecasting +150.  This is not evidence that strict payload,
close, concurrency, registration, or redaction controls should be removed.

Read-only structural remeasurement supports an idiomatic retained-core
implementation at **155–165 ordinary production SLOC**, with **174** as the
absolute local candidate maximum (cumulative **925**).  The reduction is
architectural, not line packing:

- replace the `active map[uint64]context.CancelFunc`, call IDs, cleanup helper,
  and Close-time map walk with one runtime shutdown context and
  `context.AfterFunc(shutdownCtx, cancelCall)` registered while holding the
  close gate;
- expose one clear production `New(secret)` constructor and use one unexported
  clock-injected constructor/helper for tests, instead of variadic plus
  duplicate exported constructor spellings;
- validate ready/OTP payloads once in their fixed handlers, not once before
  lookup and again inside each handler; and express the exact 17-byte OTP
  layout with named prefix/length boundaries rather than seventeen chained
  byte comparisons.

These changes preserve the fixed allowlist, maximum-three registry, prompt
linearizable cancellation, post-close denial, exact payload spelling, state
mapping, and payload-free/redacted result.  `context.AfterFunc` may schedule
only its bounded cancellation callback; the runtime owns no persistent
goroutine, timer loop, active-call collection, or wait lifecycle.  If a
readable implementation cannot satisfy all tests at **<=174**, this finding is
falsified: stop and replan with an ordinary higher SLOC estimate and reconcile
the downstream/global cap before any further DEV.

## Approval boundary and preflight

This PLAN authorizes no DEV by itself.  The stopped DEV Agent may resume in
the **same counted Lap 1** only after Main approves Revision 2 and an
independent QA Agent revises/reapproves the TASK-0013-only `QA_PLAN.md` against
this architecture and 174-SLOC maximum.  No work performed during the stop is
retroactively accepted.  Resume also requires TASK-0006 to be merged and Main to have
confirmed the dedicated worktree `/tmp/codex-authority-broker-task0013` on
`task/TASK-0013-runtime-assembly` at base `df140ef`.  Preflight also confirms
Go, the focused packages, a synthetic non-empty verifier secret, an injected
lease clock, and in-process IPC requests.  No socket, seed file, network,
credential, elevation, or external service is a preflight need.  A failed
preflight is `not_started`, with its wait and null reason recorded, not
repeated as a DEV failure.

**DEV profile:** `luna-xhigh` (`dev-luna`).  The owned change is one small,
deterministic in-process Go assembler plus table/barrier tests; existing IPC
and lease already own framing, transport, mutex-protected authority state,
and TOTP verification.  This is bounded integration rather than an
OS/network/secret-custody redesign, so the prescribed Luna profile is enough.

Only these paths may change:

| Path | Responsibility |
| --- | --- |
| `internal/backend/runtime.go` | One process-local runtime, fixed ready/OTP routing, bounded registration seam, shutdown gate, and payload-free decisions. |
| `internal/backend/runtime_test.go` | Deterministic injected-clock/secret routing, mutation, lifecycle, registration, race, and redaction tests. |

Everything else is excluded: IPC/protocol/lease changes, sockets/listeners,
daemon entrypoint, signals, seed acquisition, files, persistence, credentials,
push enablement, audit, release, installer, and canary work.  The old
TASK-0007 `runtime.go` is read-only, unapproved draft material; DEV must
rederive behavior from this PLAN and current contracts, not copy it as an
authority.

## Fixed runtime contract

The exported production constructor `New(secret)` must reject an empty secret
through the existing generic lease sentinel and delegate, if useful, to one
unexported clock-injected constructor/helper used only by same-package tests.
Construction builds exactly one
`lease.State` from the injected clock, and one `lease.TOTPVerifier` from a
private copied secret.  Tests may inject the clock; production supplies the
ordinary clock through `lease.New(nil)`.  Runtime exposes neither secret,
verifier, challenge, lease, OTP, counter, timestamp, nor state error; it logs
nothing and has no `String`/formatting path for them.

`Handle(context.Context, ipc.Request)` returns `ipc.Response{OK:false}` and a
nil, payload-free response for every rejection.  It must never return an
authority-bearing or diagnostic payload.  Nil/cancelled context, wrong
version, unknown operation, malformed payload, missing state, registration
failure, TOTP/lease error, and closed/concurrent-close state all deny
identically.  Backend errors are not needed for normal denial; if an internal
error is retained, it is fixed/non-secret and the server's existing generic
failure path must still disclose no detail.

The only currently dispatchable mapping is exact and strict:

| Request | Strict payload admission | Lease transition | Result |
| --- | --- | --- | --- |
| version 1, `ready` | absent/nil or zero-byte payload only | `BeginReadiness`; retain the returned opaque challenge for this runtime | `OK:true` only on nil error; no payload |
| version 1, `otp` | exact 17-byte JSON spelling `{"code":"NNNNNN"}` with six ASCII digits; reject every other byte sequence, including whitespace, unknown/missing/duplicate fields, wrong type, and trailing JSON | `VerifyAndActivate(currentChallenge, code, verifier)`; clear the stored challenge only after success | `OK:true` only on nil error; no payload |
| anything else | deny before a handler or lease transition | none | `OK:false`, no payload |

The `ready` handler must not create a lease; repeated accepted readiness uses
the lease package's current challenge.  OTP before readiness, expired or
foreign/currently invalid challenge, active lease, malformed/invalid/replayed/
rate-limited OTP, and post-success OTP all remain a generic false result.
The response mapping intentionally does not distinguish these state cases.

Construction installs exactly the fixed `ready` and `otp` handlers.  The
instance `Register` seam is construction-time-safe and bounded to **three**
total slots: reject empty/oversize/noncanonical operation names, nil handlers,
duplicates/replacement, registration after `Close`, and a fourth handler.  It
may reserve one valid third slot for TASK-0011, but no operation beyond the
current IPC allowlist is dispatched now: `Handle` checks the current fixed
ready/OTP allowlist before registry lookup.  There is no global registry,
`init`, discovery, callback replacement, or push behavior.

Use a small runtime mutex/gate and a private shutdown context so shutdown is
linearizable.  While holding the gate, an admitted call checks `closed`,
selects its fixed handler, creates its caller-derived context, and registers a
bounded `context.AfterFunc` cancellation callback on the shutdown context.
`Close` marks closed and cancels the shutdown context before it releases the
same gate; new and waiting calls then deny without invoking a handler.  A
handler runs outside the gate so Close cannot deadlock behind it.  Handle
stops the callback, cancels its per-call context, and rechecks close/context
state before publishing success, so its result is discarded if shutdown won
the race.  No active-call map, call ID, cleanup helper, or Close-time call walk
is permitted unless this PLAN is re-estimated again.
Guard challenge read/update with the same dedicated state/challenge lock (or a
single documented equivalent) so concurrent ready/OTP submissions cannot use
or clear a stale challenge.  `Close` is idempotent, does not create authority,
and all calls after it returns fail closed.  Do not add runtime-owned
persistent goroutines, timers, or a wait-for-handler lifecycle.

## DEV sequence, tests, and checks

1. Restructure only `runtime.go` and create `runtime_test.go` in the new backend package;
   use existing exported `ipc` and `lease` interfaces and a bounded
   fixed-layout OTP parser.  Keep named bounds and readable error paths; do
   not modify the fixed IPC allowlist or existing packages.  Remove the
   stopped candidate's active-call registry and duplicate constructor surface
   before adding tests; do not minify statements or comments.
2. Use a synthetic secret and fake clock to prove exact ready admission, no
   lease after ready, known valid OTP activation, response payload absence,
   and the complete false mapping for OTP-before-ready, invalid/malformed/
   replayed/rate-limited OTP, expired challenge, and active-lease readiness.
3. Run the adversarial matrix: wrong version; unknown operation; ready with
   every payload form; OTP with empty, missing, extra, duplicate, trailing,
   wrong-type, non-six-digit, non-ASCII, and oversized payloads; context nil
   and cancelled; empty secret; duplicate/invalid/nil/full/after-close
   registration; and a registered third name which remains non-dispatchable.
   Each denial asserts `OK=false`, empty payload, and no unintended lease or
   handler invocation.
4. Add barrier-based races for simultaneous valid OTP submissions (at most one
   accepted activation), ready/OTP interleaving, and `Handle` racing `Close`.
   Assert Close returns promptly, no handler starts after its linearization,
   no success survives a winning close, repeat Close is safe, and `go test
   -race` is clean.  Use only deterministic channels/timeouts; a timeout is a
   failure, not a reason to weaken the gate.
5. Add a redaction sentinel test that injects a unique synthetic OTP/secret
   marker into denied paths and inspects returned errors/responses and any
   test-captured diagnostic sink.  It must appear in neither; production adds
   no logging.  Fixtures may retain the declared input outside captured sinks.
6. DEV runs `gofmt -w` only on its two owned files, then first runs:

   ```sh
   go test ./internal/backend ./internal/ipc ./internal/lease
   go test -race ./internal/backend
   ```

   It records concise exact results.  REVIEW independently repeats the
   focused command and runs the repository-native full checks once; QA
   independently repeats the matrix and regressions after REVIEW PASS:

   ```sh
   GOCACHE="$(mktemp -d)" go test ./...
   go vet ./...
   test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
   git diff --check
   ```

   An unavailable `make check`, VCS-stamping failure, or sandbox socket denial
   is recorded with exact excerpt and classified `environment`, not retried
   blindly or attributed to runtime code; focused package evidence remains
   mandatory.

## SLOC, stop/split, and evidence

Revision 2 re-forecast is **+155–165** ordinary readable production SLOC,
cumulative **906–916**, with a candidate maximum of **+174 / cumulative 925**.
The original +150/cumulative-901 forecast is superseded by this measured
range, not silently retained.  Before and after
implementation, DEV counts nonblank, non-comment executable lines in all
candidate tracked production `.go` files (exclude `*_test.go`, generated,
vendor, config, and task documents); inline code with a comment counts once.
REVIEW and QA independently record per-file subtotals and total with this
canonical command.  No semicolon packing, collapsed error handling, cryptic
names, removed security comments, or artificial function merging is allowed
to fit the budget.

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

At a forecast or measured runtime above **174**, or cumulative total above
**925**, stop DEV before any new scope and obtain a revised PLAN/QA approval.
The local **950** target and **1000** hard guard remain unchanged and are not
permission to cross the earlier 925 stop; either candidate threshold stops and splits the
work; it is never solved by deleting mandatory tests or weakening strict
admission, fail-closed close, redaction, or concurrency behavior.  Also stop
and split immediately if any additional production path, seed/listener/signal/
lifecycle behavior, IPC change, or a non-gate-ready Lap-1 failure is required.
There is no Lap 3.

The wave's downstream global forecast was 1490 against target 1500.  An
eventual TASK-0013 measurement above +160 consumes more than the remaining
10-SLOC forecast margin (for example, +165 implies 1505 and +174 implies
1514).  This does not authorize edits to downstream contracts in this PLAN or
block the bounded same-Lap recovery at <=174; Main must carry the accepted
actual into TASK-0009's measurement/replanning gate, and affected downstream
contracts must be amended and independently approved before their DEV if the
1500 target would otherwise be exceeded.  The 1500 target is not relaxed.

For every attempted stage, record start/end and `active_ms` (work time),
`wait_ms` (approval/tool/agent wait), retries (including zero), classification
(`pass`, `not_started`, `implementation_defect`, `planning_defect`,
`environment`, or `regression`), raw evidence/source ID, and a null reason
where absent.  Preflight wait is excluded from delivery timing; apply any
20-percent contingency to observed active time only, never invent a SLOC
throughput claim.  Parent/Main owns the paired PLAN/DEV/REVIEW/QA/Git evidence
and external operational-log writes.

## Gate and Git ownership

PLAN → DEV → independent REVIEW → independent QA is strictly sequential.
DEV is `dev-luna`/`luna-xhigh`; Planner, Reviewer, and QA are distinct
`Terra/medium` roles.  If native agent spawning lacks the requested role or
observed model/effort differs, Main records requested/observed runtime
evidence and stops before considering the narrowly permitted fallback.
Children may neither stage, commit, merge, nor write `.git`; they preserve
other-agent changes.  Main alone takes the shared operational-repository lock,
checks scope/hooks, stages, commits, pushes/opens PRs, and performs any
`--no-ff` merge.  REVIEW must reject out-of-scope files, third-operation
reachability, secret disclosure, non-linearizable close, unbounded registry,
and SLOC compression; QA FAIL is classified before any rework and no merge
occurs without both independent PASS results.
