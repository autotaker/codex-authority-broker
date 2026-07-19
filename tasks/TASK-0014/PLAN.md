# PLAN — TASK-0014: secure bounded seed and daemon lifecycle (Revision 1)

## Decision — FAIL: planning contradiction, DEV blocked

**FAIL (`planning_defect`).**  TASK-0013 is merged at `ca3f303`; its independent
REVIEW and QA evidence records `internal/backend/runtime.go` at 171 canonical
production SLOC.  The current canonical production total is therefore **922**,
not the 901 assumed by the immutable TASK-0014 contract metadata.  The contract's
`+186` forecast consequently yields **1108**, leaving only **17** ordinary lines
before the 1125 stop.  The stop is a gate, not extra budget: its absolute local
maximum is **+203 / cumulative 1125**.

That maximum is not feasible without code compression.  The retained, unmerged
TASK-0007 daemon draft is exactly 186 SLOC, but it had no test file or gate
evidence and is not implementation authority.  Read-only comparison shows that
it also lacks the explicit production affordances needed to make all required
failure paths and deterministic descriptor/lifecycle seams testable without
host mutation (in particular strict canonical seed-value admission and explicit
post-listen cancellation/close ordering).  Re-deriving those controls from the
merged runtime needs an ordinary readable **+225--240** SLOC, with **+232** as
the working estimate (cumulative **1154**).  Even the low estimate is 22 lines
above the 1125 stop.  Packing branches, merging cleanup/error paths, deleting
security checks, or moving work into runtime/IPC would be an impermissible way
to claim feasibility.

Accordingly this PLAN specifies the intended bounded implementation and test
gate, but authorizes **no DEV** until Main obtains a revised TASK-0014 contract
and an independently revised TASK-0014-only `QA_PLAN.md` that reconcile the
ordinary estimate and downstream cumulative gates.  Only this PLAN supersedes
the stale TASK cumulative metadata; `TASK.md`, QA plan, product, tests, Git,
and operational logs remain untouched by this planning result.  No Lap is
consumed by this pre-DEV contradiction.

## Evidence and preflight

Read-only evidence inspected before this decision:

| Source | Result |
| --- | --- |
| `AGENTS.md` | PLAN → DEV → independent REVIEW → QA; roles stay separate; children do not write Git; Main alone owns Git. |
| `tasks/TASK-0014/TASK.md` | Exclusive intended paths are broker `main.go` and `main_test.go`; runtime/IPC edits are excluded. |
| `ca3f303` and TASK-0013 REVIEW/QA results | TASK-0013 is merged and PASS; runtime is 171 SLOC and its socket-free full checks provide the current dependency evidence. |
| Current canonical counter | `TOTAL 922`: CLI 83, runtime 171, IPC client 35, protocol 117, server 283, lease 173, TOTP 60. |
| Unaccepted TASK-0007 draft | 186-SLOC broker-only structural reference, no broker tests; never copy it as authority. |
| Current `backend`, `ipc`, `lease`, client/server code and repository tests | `backend.New` privately copies its secret; `Runtime.Close`, `ipc.Listen`, `Server.Serve`, and `Server.Close` supply the existing backend/listener boundary and socket-unlink semantics. |

If authority resolves the cap contradiction, preflight must confirm Linux and
Go 1.23, the merged `ca3f303` base, only the two owned paths dirty, a temporary
Unix-socket-capable test environment, and injected descriptor fixtures.  The
fixture must simulate `open`, `openat`, `fstat`, `close`, and `os.NewFile`
ownership; it must never change real `/etc`, create a service, invoke sudo, or
need real root ownership.  A missing socket capability is `environment`, not
a code retry; descriptor and unit coverage remains mandatory.

## Ownership and implementation contract after re-approval

**DEV profile:** `luna-xhigh` (`dev-luna`).  This is a small Linux entrypoint
integration with adversarial descriptor and lifecycle tests.  Planner,
independent reviewer, and independent QA remain distinct `Terra/medium`
roles.  DEV may change only:

| Path | Responsibility |
| --- | --- |
| `cmd/codex-authority-broker/main.go` | Linux-only fixed seed walk/schema, backend-before-listen construction, bounded daemon lifecycle. |
| `cmd/codex-authority-broker/main_test.go` | Deterministic descriptor, schema, redaction, startup, signal/close/unlink/restart tests. |

No runtime, IPC, lease, client, protocol, socket-server, CLI, installer,
sudo, persistence, audit, credentials, push, Git, or operational-log edit is
permitted.  A need for any of those is an immediate split/replan.

The production design is fixed as follows.

1. `main` creates a `signal.NotifyContext` for `SIGINT` and `SIGTERM`, calls a
   dependency-injected `run`, then exits only with its bounded status.  Normal
   construction first opens/validates and reads the seed, then calls
   `backend.New(secret)`, and only then calls `ipc.Listen`; a denied seed or
   failed backend construction must never create a listener.  Zero the decoded
   temporary secret on every return after `backend.New` has made its private
   copy. `openSeed`/`loadSeed` return exactly the private sentinel
   `errSeed = errors.New("seed unavailable")` for every seed denial; `run`
   returns status 1 for that sentinel and every backend/listen/serve/close
   failure, emits no diagnostic, and exposes no pathname, seed bytes, secret,
   UID, descriptor, parser, or server detail.
2. The fixed seed path is walked from an `O_DIRECTORY|O_RDONLY|O_NOFOLLOW|
   O_CLOEXEC` descriptor for `/`, using `openat` for every literal component.
   Each opened directory is verified by `fstat` as a directory and closed
   before advancing; every success/failure branch owns and closes precisely one
   descriptor.  The final `openat` uses `O_RDONLY|O_NOFOLLOW|O_CLOEXEC`; its
   final descriptor alone is admitted only when `fstat` reports regular file,
   UID 0, exact permission bits `0600`, and a size in `1..1024`.  Any syscall,
   type, ownership, mode, size, symlink, or file-wrapper failure maps to the
   same `errSeed` failure.
3. Read exactly the validated length through a `maxSeedBytes+1` limit and
   reject short/read-error/extra input.  Decode one JSON object with
   `UseNumber`, exactly two unique fields (`totp_secret_b64` and `allowed_uid`),
   no unknown fields or trailing value, a positive uint32 UID, and canonical
   standard base64 that decodes to 1..128 secret bytes.  Reject duplicate,
   missing, wrong-type, fractional/negative/zero/out-of-range UID, whitespace
   or non-canonical base64, and oversized values.  Parser data and partial
   secrets are zeroed; all parser denials return `errSeed` and no data.
4. After successful `Listen`, call `Serve(ctx)`, then in one deferred,
   deterministic ownership path close the runtime and server exactly once.
   Treat a signal/context cancellation as clean only when `Serve` and close
   report the expected clean result; otherwise return status 1.
   Rely on the existing server's identity-checked cleanup to unlink only its
   owned socket.  A second fresh `run` must reread the seed and construct a
   new runtime/server; it must not retain an old secret, backend, listener, or
   socket.  A restart with missing/invalid seed must deny before listen.

## Deterministic Lap-1 test matrix

Tests inject a complete descriptor façade (recorded open/openat/fstat/close and
reader conversion), an in-memory bounded reader, a fake listener/server, and
a cancellable context.  They do not call the production `/etc` path or require
root.  Test every descriptor terminal condition and assert the exact close
set/order: root, each parent, final regular root-owned 0600 success; final and
parent symlink/no-follow denial; missing component; directory/nonregular final;
wrong owner; each non-0600 mode; zero/oversized/short/read-error seed; and
every open/openat/fstat/file-wrapper failure.  Ensure no descriptor survives
each denial or success path.

Use table-driven byte fixtures for the strict schema: valid minimum/maximum
secret; duplicate, unknown, missing, trailing, malformed, wrong-type,
fractional/negative/zero/overflow UID; invalid/empty/oversized/noncanonical
base64; and oversized JSON. Every denial asserts exactly `errSeed` and checks
a unique synthetic secret/path marker appears in neither returned error nor
captured diagnostic output.

Lifecycle tests prove in this order: valid seed constructs `backend.New`
before fake `listen`; listen failure closes the constructed runtime and leaves
no serve call; signal cancellation reaches `Serve`, triggers close/unlink once,
and yields status 0; seed/backend/listen/serve/close failure yields status 1;
normal close is idempotent; a second valid run has distinct backend/server state and socket
cleanup; and second run with a denied seed makes no listen call.  An optional
real-socket integration uses only `t.TempDir` and current UID, exercises the
existing IPC client ready request, shutdown cleanup, and valid restart.  If
the sandbox denies socket creation, record it once as `environment`; retain
the injected lifecycle evidence and repeat the integration in a socket-capable
environment during REVIEW/QA.

DEV formats only its two files, then records the exact outputs of:

```sh
go test ./cmd/codex-authority-broker ./cmd/codex-authority ./internal/backend ./internal/ipc ./internal/lease
go test -race ./cmd/codex-authority-broker
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test ./...
go vet ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
```

`make check` is not a required gate because the merged evidence establishes no
such target.  REVIEW independently reruns the focused/race/full/static suite
and rejects scope expansion, descriptor leaks, a symlink-following path,
noncanonical schema, secret disclosure, listener-before-backend ordering,
or non-idempotent/unowned socket cleanup.  QA independently repeats the full
fixture mutation and lifecycle matrix and the existing CLI/IPC/lease
regressions.  No DEV occurs in Lap 2 and no Lap 3 exists.

## SLOC, stop/split, and stage accounting

After re-approval, DEV, REVIEW, and QA independently count nonblank,
non-comment executable lines in every candidate tracked production Go file;
tests, generated files, vendor, configuration, and task documents do not
count.  Inline code with a comment counts once.  Do not use semicolon packing,
collapsed error handling, cryptic names, or deleted security commentary to fit
a cap.

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

The superseded contract arithmetic is `922 + 186 = 1108`; its unchanged stop
permits at most `1125 - 922 = 203` added SLOC.  The revised ordinary range is
`922 + 225..240 = 1147..1162` (working estimate 1154), therefore crossing the
present stop before implementation begins.  Main must reconcile this planning
contradiction in the contract/QA gate and downstream cap plan before DEV.  If
the approved replacement forecast or a measured candidate exceeds its newly
approved stop, or any required test is not gate-ready in Lap 1, stop and split;
do not compress or borrow runtime/IPC scope.

For every PLAN/DEV/REVIEW/QA attempt, Main records UTC start/end,
`active_ms`, `wait_ms`, retries (including zero), classification (`pass`,
`not_started`, `planning_defect`, `implementation_defect`, `environment`, or
`regression`), raw command/source evidence, and a null reason.  Apply any
20-percent contingency only to observed active time, never to SLOC.  This
planner attempt began before the available terminal timestamp and completed at
`2026-07-19T08:14:23Z`; exact start and active duration were not exposed, so
`active_ms=unavailable`, `wait_ms=0`, `retries=0`, classification
`planning_defect`, null reason `planner runtime did not expose the turn start`.
The stated deadline is `2026-07-19T08:39:41Z`; the observed completion evidence
is before it.  Main owns any operational-log entry and all Git work.

---

# PLAN — TASK-0014 Revision 2: merged-gate execution decision

## Decision — PASS: DEV may open under the reconciled gate

**PASS (`none`, retries 0) against merge
`d0736bbc6838e78ff5acc3f397f2b338a009b035`.**  Revision 1 remains intact as
historical **FAIL (`planning_defect`)** evidence: its then-effective 1125 stop
could not admit the ordinary readable estimate.  The merged contract, independent
REVIEW attempt 2, and independent QA result now reconcile the operative local
ledger without changing product scope: baseline **922**, forecast **+232**
(readable **+225..240**), cumulative **1154** (range **1147..1162**), trigger
**1200**, target **1250**, absolute **1350**, and global **1500/1800**.  Thus
`1154 < 1200 < 1250 < 1350 < 1500 < 1800`; the global limits supply no local
headroom.

`git diff --check` passed, the canonical non-test production counter is 922,
and no broker candidate exists.  The merged APIs are sufficient without a
runtime or IPC edit: `backend.New([]byte)` builds an isolated runtime and
`Runtime.Close()` is idempotent/fail-closed; `ipc.Listen(Config, Backend)`
accepts the admitted UID, while `Server.Serve(context)` and `Server.Close()`
already provide cancellation, connection draining, and identity-checked unlink.
`dev-luna` is therefore approved for **Lap 1 only**, provided it uses ordinary
readable code within this forecast.  No compression, API change, or control
shedding is approved.

## Locked ownership, construction, and seams

DEV may modify only `cmd/codex-authority-broker/main.go` and
`cmd/codex-authority-broker/main_test.go`.  It must retain every Revision 1
and TASK-first QA-plan control: fixed root-relative descriptor walk; an
`openat`/`O_NOFOLLOW`/`O_CLOEXEC` open and `fstat` for root, every parent, and
the final descriptor; directory parents; final regular UID-0 exact-0600 file;
explicit name/depth/path/size bounds; and exact close ownership/order on every
success and denial.  No pathname pre-check, symlink resolution, host mutation,
sudo, or real privileged seed fixture is allowed.

The seed remains one bounded read and one strict JSON object: exactly
`totp_secret_b64` and `allowed_uid`, unique/no unknown or trailing values,
positive uint32 UID, canonical standard base64, and bounded non-empty decoded
secret.  All metadata/read/schema denials map to a private generic error;
errors and diagnostics must reveal no path, descriptor, UID, seed, or secret.
Only owned mutable buffers are wiped (with the QA-plan Go-memory limitation
stated in tests); no broader erasure claim is permitted.

Use private interfaces/adapters in `main.go` for descriptor operations,
reader/file conversion, `backend.New`, and listen/serve/close so tests can
record calls without changing the merged runtime/IPC surface.  Success order
is mandatory: validate/read seed, construct runtime, zero owned seed material,
then `ipc.Listen` with the admitted UID.  Every preconstruction denial makes
zero listen calls; listen failure closes the newly built runtime and leaves no
owned socket.  Signal-driven serve shutdown closes runtime and server exactly
once, treats unexpected serve/close errors as failure, relies on server-owned
identity cleanup, and creates fresh runtime/server/seed state on restart.

## Lap-1 acceptance and stop rule

The deterministic test seam must cover the full descriptor walk (root, each
parent, final), each open/openat/fstat/file-wrapper/read/close terminal branch,
parent and final symlinks, nonregular/wrong-owner/all wrong-mode finals, and
path/name/depth bounds including boundary+1.  It must cover valid minimum and
maximum schema plus duplicate/unknown/missing/wrong-type/trailing/malformed
JSON, invalid UID forms, invalid/noncanonical/empty/oversized base64, and
oversized/short/error reads.  Each denial asserts redaction, owned-descriptor
closure, buffer handling where applicable, and no listen.

Lifecycle/mutation acceptance remains: construction-before-listen; listen,
serve, and close failure; SIGINT and SIGTERM; repeated/concurrent close with
an active client; owned-unlink replacement race; valid fresh restart; denied
restart-without-seed; and unchanged existing-client ready/OTP plus malformed
and unknown-request denial.  Use channel/barrier ordering and bounded contexts,
not sleep proofs.  A Unix-socket capability denial is `environment` with a
null reason, never product PASS; fixture coverage remains mandatory and a
socket-capable REVIEW/QA rerun is required.

Required DEV evidence (record exit status, UTC start/end, active/wait, retry,
classification, and redacted null reason) is:

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease
go test -race ./cmd/codex-authority-broker
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test ./...
go vet ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
```

Recount ordinary non-test production SLOC before implementation and after every
meaningful addition.  At **>1200**, stop for explicit replan; at **>1250**,
stop and perform the exact merged ordered shedding audit; at/above **1350**,
FAIL/stop absolutely.  Stop/split immediately if a runtime/IPC or other-path
change, an unisolatable descriptor seam, or a non-gate-ready Lap-1 test is
needed.  Never compress branches/names/errors, delete tests/comments/controls,
disguise generated code, borrow global capacity, or use Lap 3.

## Planner evidence

| Item | Evidence |
| --- | --- |
| Changed path | `tasks/TASK-0014/PLAN.md` only |
| Merge / precondition | `d0736bbc6838e78ff5acc3f397f2b338a009b035` is `HEAD`; merged REVIEW attempt 2 and QA contract gate are PASS |
| Current candidate state | `TOTAL 922`; broker `main.go` and `main_test.go` absent; no product decision asserted |
| Active / wait / retries | `active_ms=unavailable`; `wait_ms=0`; `retries=0` |
| Classification / null reason | `none`; planner runtime did not expose a reliable turn-start timestamp, so active duration is not inferred |
| Completion / next state | `2026-07-19T08:44:03Z`; **active: dev-luna Lap 1**, **wait: independent REVIEW then QA**, **retry: none** |

Main alone records any external/Lap timing log and Git operation.  Planner,
DEV, REVIEW, and QA remain separate roles.

---

# PLAN — TASK-0014 Revision 3: Revision 2 command-gate correction

## Decision — PASS: DEV remains approved under the same reconciled gate

**PASS (`none`, retries 0).** This corrects only Revision 2's two omitted
cache-bypass flags, as identified by independent QA_PLAN Revision 3. Revision
1 remains historical **FAIL (`planning_defect`)** evidence, and Revision 2
remains otherwise unchanged. The reconciled ledger, gates, mandatory controls,
no-compression rule, Lap-1-only authorization, and all descriptor, schema,
redaction, lifecycle, mutation, fixture, and existing-client controls remain
exactly as approved: `1154 < 1200 < 1250 < 1350 < 1500 < 1800`.

`dev-luna` remains approved only for Lap 1 and only for the same two owned
paths: `cmd/codex-authority-broker/main.go` and
`cmd/codex-authority-broker/main_test.go`. No runtime/IPC edit, scope change,
or other control change is authorized.

The required command gate in Revision 2 is corrected to restore these exact
commands:

```sh
go test -count=1 -race ./cmd/codex-authority-broker
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./...
```

All other Revision 2 required commands and their timing, exit-status,
active/wait/retry, classification, and redacted-null-reason recording remain
required without restatement or weakening.

## Planner correction evidence

| Item | Evidence |
| --- | --- |
| Changed path | `tasks/TASK-0014/PLAN.md` only |
| Timing | completed `2026-07-19T08:47:45Z`; `active_ms=unavailable`; `wait_ms=0` |
| Retry / classification | `retries=0`; `none` |
| Null reason / next state | planner runtime did not expose a reliable turn-start timestamp, so active duration is not inferred; **active: dev-luna Lap 1**, **wait: independent REVIEW then QA**, **retry: none** |

---

# PLAN — TASK-0014 Revision 4: explicit post-trigger replan

## Decision — FAIL/SPLIT: no DEV retry under the 1200 trigger

**FAIL/SPLIT (`planning_defect`, planner retries 0).** Revisions 1--3,
including their historical failures and their prior Lap-1 authorization, remain
unchanged evidence. This is a new, read-only assessment of the stopped,
unformatted broker candidate; it does not reclassify a product or test result.
The canonical counter is now **1205**: baseline **922** plus broker
`main.go` **283**, with no broker test file. The approved trigger is **1200**,
so it permits at most **+278**. The candidate is already **+5** over that
trigger and is not gate-ready.

This is not a five-line arithmetic replan. The candidate has controls that
must be corrected or made explicit before it may be tested:

1. The final-file predicate uses `mode&0777 == 0600`, which admits setuid,
   setgid, and sticky bits. It must require exact permission plus special bits
   with `mode&07777 == 0600`; this is an in-place replacement, not a line
   saving.
2. `makeRuntime` and `listen` are invoked without nil guards. Fold both guards
   into the existing context admission condition so a fixture or dependency
   failure returns status 1 rather than panicking; this is line-neutral.
3. A listener factory can return both a server and an error. Its failure path
   currently closes only the runtime, leaving the returned server/socket
   ownership ambiguous. Close a non-nil returned server before returning
   status 1, then close the runtime. This adds one ordinary readable line and
   preserves server-owned identity-checked unlinking.
4. A nil `Serve` result before cancellation is currently accepted as status 0.
   Extend the existing deferred failure condition to require `ctx.Err() != nil`
   for a clean zero-status result; an expected cancellation remains clean and
   unexpected serve termination is status 1. This is line-neutral. The final
   descriptor's `os.NewFile` wrapper owns that final fd after successful
   conversion, so the retry's fixture must prove reader-close ownership and
   must not add a second descriptor close on that path.

The only verified non-compressive reductions are also insufficient: remove the
three-line `openSeed` forwarding wrapper by retaining its named fixed-path
implementation directly, and replace the duplicate identical root/parent
directory-flag constants with one named `directoryFlags` constant (**-4**).
Together with the required non-nil-server failure close (**+1**), the credible
readable floor is **280** production SLOC (cumulative **1202**, net **+280**).
The mode, nil-dependency, and cancellation repairs do not change that count.
Further reduction to 278 would require removing useful named seams/bounds or
combining control/error paths; neither is authorized. No schema, descriptor,
redaction, construction-before-listen, wipe, cancellation, close/unlink,
restart, API, scope, test, or gate control may be deleted or weakened to fit.

Accordingly, do **not** authorize the requested DEV retry at the present
trigger and do not treat this as a gate raise. The smallest split boundary for
Main's contract/QA reconciliation is a TASK-0014-only local candidate limit
of **1202 / +280**, solely to admit the named corrections and simplifications
above; all later 1250/1350 and global gates remain unchanged. If that boundary
is not independently approved, stop with this FAIL and leave the candidate
untouched. If it is approved, authorize exactly **one** DEV retry, with a
measured target of **280** (acceptable readable range **279--280**, never
above 1202), followed by independent REVIEW and QA. DEV must add the required
deterministic broker tests before any PASS claim; their absence is
`not_started`, not an excuse to waive the gate.

## Revision 4 evidence and handoff

| Item | Evidence |
| --- | --- |
| Changed path | `tasks/TASK-0014/PLAN.md` only |
| Candidate inspected read-only | `cmd/codex-authority-broker/main.go`, 283 counted production SLOC; no `main_test.go` exists |
| Measured ledger / forecast | `922 + 283 = 1205` observed; corrected readable floor `922 + 280 = 1202` |
| Timing | completed `2026-07-19T08:59:17Z`; `active_ms=unavailable`; `wait_ms=0` |
| Retry / classification | planner `retries=0`, `planning_defect`; conditional DEV retry budget `1` only after independent 1202 split approval |
| Null reason / next state | turn-start timestamp was unavailable; **active:** Main contract/QA split decision; **wait:** no DEV, then independent REVIEW and QA only if approved |

Main alone owns the requested contract/QA reconciliation, timing/log entries,
and Git work. Planner must not edit product, tests, QA plans/results, backlog,
or Lap logs.
