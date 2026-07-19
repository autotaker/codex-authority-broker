# TASK-0014 QA plan — TASK-first baseline, Revision 2 reconciliation

**Role:** independent QA planning (Terra/medium). The TASK-first baseline below remains the source of the descriptor, schema, redaction, lifecycle, mutation, fixture, command, and no-compression requirements. This Revision 2 independently reconciles that baseline with `AGENTS.md`, the revised `tasks/TASK-0014/TASK.md` and `backlog.json` index, historical TASK-0014 PLAN Revision 1 FAIL, TASK-0014 REVIEW_RESULT attempt 1, TASK-0008/TASK-0009 revisions, and TASK-0006 global controls. It changes no product, contract, review/result, Git, or operational-log content.

## Revision 2 reconciliation — PASS for the planning gate

**PASS (`none`, retries 0):** the revised contract/index resolves the cap contradiction without weakening the baseline. PLAN Revision 1 remains historically correct as a `planning_defect`: under its then-current 1125 stop, the ordinary readable `+225..240` broker estimate could not fit. The revised contract now supplies an explicit, coherent post-reestimate trigger and does not reclassify that historical FAIL as a product or test defect.

The reconciled arithmetic is **922** actual merged baseline + **232** forecast (readable range **+225..240**) = **1154** cumulative (range **1147..1162**). Thus `1154 < 1200 < 1250 < 1350 < 1500 < 1800`. The per-Task trigger is **1200**, target is **1250**, and absolute hard guard is **1350**; the global mandatory-v1 target and unconditional hard limit remain **1500** and **1800** respectively. The current-wave forecast remains `922 + 232 (TASK-0014) + 120 (TASK-0008) + 0 (TASK-0009) = 1274`, below 1500 and 1800, with no silent current-wave saving.

DEV may open only when the independently counted forecast is **<=1200** and every other pre-DEV gate below passes. A forecast or measurement **>1200** stops for explicit replan; it is not usable headroom. A forecast or candidate **>1250** stops for explicit replan and the exact shedding audit below. **1350 is absolute**: crossing it, or forecasting above it, is FAIL/stop and cannot be approved, borrowed, raised, or repaired by compression. `push-to-v2` is **not selected** here. TASK-0009 must, after its own PASS+merge, explicitly replan invalidated TASK-0010--TASK-0012 arithmetic from its frozen evidence; it alone may make the later `push-to-v2` decision, never silently.

If a forecast or candidate is above 1250, the replan must audit this exact ordered list, in order, and retain each stated minimum; mandatory-v1 controls are unsheddable:

1. automated canary executable; retain manual runbook/evidence;
2. rich status/JSON UX; retain activate and immediate revoke;
3. rich audit schema/correlation; retain correlation ID, actor, scope, result, and expiry;
4. precomputed pack-size/history diagnostics; retain exact repo/ref/clean-tree validation and normal non-force rejection;
5. remote-OID prefetch/race diagnostics; retain standard non-force rejection and generic failure;
6. automated installer/rollback executable; retain declarative units and manual verified install/rollback;
7. move GitHub push to v2, leaving TOTP full-sudo authority as v1.

This reconciliation attempt is planning evidence only, not candidate QA or a product PASS. `active_ms=unavailable` because the planner runtime did not expose a start timestamp; `wait_ms=0`; `retries=0`; null reason: duration is not inferred. It was completed for the stated **2026-07-19T08:39:41Z** deadline. Execution must continue to record command start/end UTC, duration, active/wait, retries, classification, and null reason without secrets.

## Baseline gate and timing

| Item | Baseline | Gate |
| --- | --- | --- |
| TASK-0013 dependency/runtime | `backend.New(secret)` constructs isolated in-memory runtime; only IPC `ready` and `otp` are admitted; `Close` fails calls closed | PASS: usable contract present |
| Existing transport/lifecycle API | bounded 4096-byte framed strict JSON IPC, Unix listener with ownership-checked cleanup, and client are present | PASS: broker can be tested without API changes |
| TASK-0014 entrypoint/candidate | `cmd/codex-authority-broker/main.go` and its test are not yet present | NOT EXECUTED / no product PASS asserted |
| QA-plan planning gate | all required TASK-first acceptance, mutation, lifecycle, SLOC, and command gates are specified below | **PASS** |

Timing evidence: planning baseline observed at **2026-07-19T08:12:18Z**; deadline **2026-07-19T08:39:41Z**; 27m23s remained. This is planning-time evidence only: no candidate was run, no retries occurred, and active/wait split is not applicable. At execution, record command start/end UTC, duration, active versus waiting time, retry count, result classification, and null reason when a check cannot run. Do not log seed values, raw metadata content, credentials, TOTP values, GitHub App material, installation tokens, or other secrets.

## Scope, ownership, and non-negotiable stops

QA may assess only the new broker entrypoint and its test: `cmd/codex-authority-broker/main.go` and `cmd/codex-authority-broker/main_test.go`. It must consume existing runtime, IPC, and lease APIs unchanged. No runtime/IPC operation or API, sudo/elevation, persistence, audit, installer, release, push, credentials, or operational logging is in scope.

Stop and return **SPLIT/FAIL** before accepting a candidate if:

- it needs a production path other than the two owned paths, or changes the runtime/IPC API;
- descriptor-safe seed metadata cannot be isolated behind fixture-safe seams;
- focused Lap-1 tests are absent or fail;
- independently counted cumulative forecast or measurement exceeds the **1200** trigger; or
- a control is removed, weakened, or compressed merely to meet a line limit.

Reconciled SLOC ledger: the actual merged cumulative baseline is **922**. The ordinary readable forecast is **+232** (range **+225..240**) and cumulative forecast is **1154** (range **1147..1162**). DEV may open only at **<=1200**. A forecast or measurement above **1200** stops for explicit replan; a forecast or candidate above **1250** stops for explicit replan and the exact ordered shedding audit in Revision 2; **1350** is absolute. The global mandatory-v1 target/hard limits remain **1500/1800**, neither of which supplies local Task headroom. Count production code by the repository's ordinary readable SLOC convention and report baseline, changed-path forecast, candidate total, and test SLOC separately. No code compression, control deletion, generated-code disguise, or test deletion is an acceptable remedy.

## Acceptance criteria and deterministic evidence

1. **Secure one-shot seed admission.** Startup reads the fixed seed exactly once by a descriptor-relative Linux walk beginning at `/`. The test seam must observe the root descriptor and each component operation; it must not implement the security decision with `filepath.EvalSymlinks`, a path pre-check, or a direct final pathname open. Every parent and final component must be opened/checked with no-follow semantics; each acquired descriptor is close-on-exec. The final descriptor must be a regular file, owned by UID 0, with *exact* permission mode 0600. Check metadata on the descriptor actually opened, not a later path lookup.
2. **Bounded names and walk.** Reject absolute/empty/`.`/`..` components, embedded separators, overlong component names, a component-count/path-byte bound breach, and root-walk escape. A valid root-relative fixed seed path succeeds only when every intermediate component is a real directory and the final component is the checked regular file. Bounds must be explicit, tested at boundary and boundary+1, and execute deterministically without a host-owned seed path.
3. **Strict bounded schema.** Accept only the documented seed schema needed to create `backend.New`; require exactly one value for every required field, reject duplicate fields, unknown fields, malformed JSON, trailing JSON, empty/incomplete input, invalid value form, and input/schema size beyond a stated bound. The parser may not accept a partially decoded prefix. Its output must have a bounded non-empty secret appropriate for the existing runtime; malformed material never reaches listener construction.
4. **Read failures and secret handling.** Short reads, partial reads followed by an error, immediate read errors, close errors where surfaced, and oversized streams deny startup without a listener. Returned errors and test diagnostics are redacted: no raw seed, secret substring, or decoded secret is present. Wipe mutable read/parse buffers as soon as practical after use and on denial. QA must state the Go limit precisely: copies made by immutable strings, JSON/runtime internals, stack/register/compiler behavior, and kernel/page-cache data cannot be guaranteed zeroed. The test must verify only owned mutable-buffer zeroing and redaction, not claim whole-process or disk erasure.
5. **Construction before listen.** For every seed/open/metadata/schema/runtime-construction denial, the injected listener factory must have zero calls and no socket node may exist. On success, construct the runtime from the admitted secret before calling `ipc.Listen`; then zero owned seed buffers. A listen failure returns a classified failure, exposes no secret, and leaves no broker-owned socket.
6. **Lifecycle and concurrency.** A valid fixture starts a Unix socket, existing IPC clients can issue unchanged `ready` and `otp` requests, and no third operation/payload-bearing decision is introduced. SIGINT and SIGTERM cause deterministic cancellation: stop accepting, close the runtime fail-closed, wait for serving/accepted connections as existing server contract does, and close/unlink only the socket identity owned by this server. Repeated shutdown must be safe. Replacing the path with a different node before cleanup must not unlink that replacement. A fresh restart has new process-local state: it starts idle, reads its seed once, has no persisted lease/challenge/replay state, and must deny restart-without-seed before listening.

## Mutation matrix and classifications

| Class | Required deterministic mutation | Expected classification/evidence |
| --- | --- | --- |
| valid control | root-relative fixture walk, real parents, regular UID-0/mode-0600 final descriptor, minimal strict schema | allow construction, exactly one seed read, then one listen call |
| traversal/name/path | `..`, `.`, empty, slash-bearing/absolute component, overlong name, depth/path bound and +1 | **SECURITY_DENY**; no listen |
| symlink | symlink at each parent position and separately final symlink | **SECURITY_DENY**; prove no-follow rejection at that exact component |
| final descriptor | directory, FIFO/socket/device, non-root owner, 0644/0660/0000 mode, mode bits beyond 0600 | **SECURITY_DENY**; descriptor metadata was checked |
| schema | duplicate key, unknown key, malformed/trailing JSON, missing/empty/invalid field, over-bound frame/schema | **INPUT_DENY**; bounded strict parser, redacted error, no listen |
| reader | zero/short, partial then error, immediate error, over-bound stream | **I/O_DENY**; no partial acceptance and owned buffer zeroing check |
| construction/listen | invalid runtime secret or injected listener failure | **CONSTRUCTION_FAIL** / **LISTEN_FAIL**; no secret and no residual owned socket |
| signals/close | SIGINT and SIGTERM while serving; concurrent/repeated shutdown; active client | **LIFECYCLE_PASS** only after Serve returns, runtime is closed, and owned socket is removed |
| unlink race | replace socket path with another file/socket identity before cleanup | **OWNERSHIP_DENY**; replacement remains |
| restart | valid shutdown then new instance; missing seed new instance | **RESTART_PASS** fresh idle state / **INPUT_DENY** with zero listen calls |
| existing client | `ipc.Client` sends valid ready/OTP and malformed/unknown requests against broker | unchanged valid protocol behavior; malformed/unknown denial; no API change |

An unexpected allow for any deny mutation is **SECURITY FAIL**. A denial that opens/listens first is **ORDERING FAIL**. Nondeterministic timing, races, or unbounded blocking is **HARNESS FAIL** until replaced by a controllable seam; it is never evidence of product PASS. A changed runtime/IPC interface is **SCOPE/SPLIT FAIL**. Record OS/platform unavailability as **ENVIRONMENT NULL** with a reason, not as a pass; Linux socket-capable evidence remains required.

## Fixture-safe execution and required commands

Tests must run on Linux with a temporary Unix socket directory and injected descriptor-relative `openat`/stat/read/close and listener seams. Simulate UID 0 and mode 0600 metadata in the seam; do not `sudo`, chown real host files, install a service, read a real privileged seed, or bind outside a test temporary directory. Use barriers/channels and bounded test contexts for signal, active-client, and unlink-race ordering; avoid sleep-only proofs.

Required Lap-1 command (exact Task contract):

```sh
go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease
```

Required focused QA command after candidate tests exist:

```sh
go test -count=1 -race ./cmd/codex-authority-broker
```

Required final regression command:

```sh
go test -count=1 ./...
```

Capture each command's exit status and timing. The socket/lifecycle test must also assert pre-listen zero calls for each failure class, socket existence only after successful construction, and absence after owned shutdown. Run the focused lifecycle/race test repeatedly (document count and all outcomes) only as supplemental confidence; the channel-controlled single run is the primary deterministic evidence.

## QA disposition template

Candidate QA may be marked **PASS** only when all acceptance rows and required commands pass, the independently counted forecast/candidate is **<=1200**, no scope stop applies, and timing/classification evidence is complete. Above **1200**, stop/replan; above **1250**, forecast/candidate stop plus the exact ordered shedding audit; at or above a forecasted/crossed **1350**, absolute FAIL/stop. Otherwise mark **FAIL**, **SPLIT**, or **ENVIRONMENT NULL** with the exact row, command, observed ordering, and redacted evidence. This document is a plan baseline and reconciliation, not a QA result, and does not modify product, tests, review/QA results, Git, or operational logs.
