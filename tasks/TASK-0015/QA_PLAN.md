---
task_id: "TASK-0015"
qa_role: "independent QA (Terra/medium)"
status: "reconciled with PLAN; QA gate PASS for DEV readiness"
baseline_status: "TASK-first baseline preserved; PLAN comparison complete"
---

# TASK-0015 QA plan — task-first baseline

## Gate disposition

This is an independent baseline derived from `TASK.md` and the final
TASK-0014 review result.  TASK-0015 `PLAN.md` was intentionally not read when
this version was prepared.  Therefore this document does **not** approve DEV:
the main Agent must first compare the submitted PLAN against this acceptance
matrix and approve both gates.  Any conflict in security ordering, coverage,
scope, or feasibility is a `requirement_gap`/`qa_plan_defect` return to the
appropriate gate, not permission to weaken this matrix.

TASK-0014's final product review is baseline evidence, not an approved product
candidate: it found `implementation_defect`s for delayed caller-secret wiping
and missing mandatory deterministic cases.  TASK-0015 must independently
close both gaps.

## Scope and invariants

- Owned candidate paths: `cmd/codex-authority-broker/main.go` and
  `cmd/codex-authority-broker/main_test.go` only.  Runtime and IPC APIs must
  remain unchanged; no new IPC operation, sudo, installation, persistence,
  audit, credential, push, release, or canary work is in scope.
- Production source must remain readable and at or below the Task's local
  forecast/boundary: merged baseline 922 plus at most 280, cumulative at most
  1202.  A boundary/security change, unreadable compression, or hard-limit
  risk stops for Main judgment; it is not solved by removing required tests or
  controls.
- No command output, assertion message, fixture name, golden data, or QA
  artifact may include a seed, decoded secret, OTP/TOTP value, token, or raw
  credential.  Negative cases assert generic/redacted errors only.
- The code must use a descriptor-relative `/` walk with no-follow and
  close-on-exec semantics; fixture injection must simulate root-owned 0600
  metadata without host ownership/mode mutation or service installation.

## Required deterministic test mapping

Every test below is required by name before DEV is approved.  Test names name
the contract; compatible subtests may be used, but an aggregate test may not
hide an omitted row.  `go test -run '^TestName$'` must be deterministic for
each named top-level test.

| TASK acceptance row | Exact deterministic test name | Operation and required observation |
| --- | --- | --- |
| caller-owned decoded secret is wiped before first listen | `TestRunWipesCallerSecretBeforeListen` | A factory retains the caller-owned input while a listener barrier checks every byte is zero before it permits/records the first listen call.  Include factory-error wiping in the same test's named subcase. |
| root descriptor open, stat, and close errors | `TestLoadSeedRootOpenError`, `TestLoadSeedRootStatError`, and `TestLoadSeedRootCloseError` | Inject each root terminal failure; return generic seed failure and leak no descriptor/secret diagnostic. |
| parent/final descriptor-walk open, stat, and close errors | `TestLoadSeedParentOpenError`, `TestLoadSeedFinalOpenError`, `TestLoadSeedParentStatError`, `TestLoadSeedFinalStatError`, `TestLoadSeedParentCloseError`, and `TestLoadSeedFinalDescriptorCloseError` | Inject each terminal failure independently; reject safely and prove final wrapped reader does not cause a second raw-final-fd close. |
| final reader close error | `TestLoadSeedReaderCloseError` | Reader close error rejects after bounded read and is redacted. |
| parent and final symlinks | `TestLoadSeedRejectsParentSymlink` and `TestLoadSeedRejectsFinalSymlink` | Descriptor-relative no-follow walk rejects each link location. |
| exact owner, mode, and regular-file type | `TestLoadSeedRejectsNonRootOwner`, `TestLoadSeedRejectsNon0600Mode`, and `TestLoadSeedRejectsNonRegularFile` | Reject UID other than 0, every mode whose `mode&07777 != 0600`, and non-regular metadata. |
| size lower/upper boundaries and read error | `TestLoadSeedSizeBounds`, `TestLoadSeedShortRead`, and `TestLoadSeedReadError` | Cover file sizes `0`, minimum, maximum, and maximum+1; valid maximum secret/encoded payload; exactly-limit accepted input, limit+1 rejection, short read, and injected read error. |
| strict schema valid secret and valid maximum secret | `TestLoadSeedAcceptsValidSchema` and `TestLoadSeedAcceptsMaximumSecret` | Accept only canonical schema/base64 and verify decoded bytes/length without printing them. |
| malformed, duplicate, unknown, missing, empty, wrong-type, trailing, invalid UID/base64, and oversized schema inputs | `TestLoadSeedRejectsMalformedSchema`, `TestLoadSeedRejectsDuplicateSchemaField`, `TestLoadSeedRejectsUnknownSchemaField`, `TestLoadSeedRejectsMissingSchemaField`, `TestLoadSeedRejectsEmptySecret`, `TestLoadSeedRejectsWrongSchemaType`, `TestLoadSeedRejectsTrailingJSON`, `TestLoadSeedRejectsInvalidAllowedUID`, `TestLoadSeedRejectsInvalidBase64`, `TestLoadSeedRejectsNonCanonicalBase64`, and `TestLoadSeedRejectsOversizedSchemaInput` | Each is isolated and asserts only generic/redacted error content.  Oversized includes encoded payload that would decode above the maximum. |
| runtime factory failure | `TestRunRuntimeFactoryError` | Factory error prevents listen and wipes decoded caller buffer. |
| construction before listen and configuration order | `TestRunConstructsRuntimeBeforeListen` and `TestRunConfiguresServerBeforeListen` | Ordered fakes/barriers prove factory then server/configuration construction complete before first listen. |
| listener failure returns non-nil server | `TestRunClosesServerOnListenError` | Listener returns `(server, err)`; server close occurs before runtime close, with no serve. |
| unexpected serve return | `TestRunRejectsUnexpectedServeReturn` | A non-cancellation `Serve` return is a failure and all resources are closed/unlinked. |
| server close failure | `TestRunReportsServerCloseError` | Shutdown close error is surfaced/redacted and still attempts remaining cleanup. |
| SIGINT and SIGTERM | `TestRunShutsDownOnSIGINT` and `TestRunShutsDownOnSIGTERM` | Inject signal source; each causes orderly cancellation, server close, listener close, unlink, and runtime close. |
| active client during shutdown | `TestRunWaitsForActiveClientOnShutdown` | A real/injected blocking client is active at shutdown; no premature teardown or hang, then orderly completion. |
| concurrent and repeated shutdown | `TestRunHandlesConcurrentShutdown` and `TestRunShutdownIsIdempotent` | Concurrent signal/close triggers race-safe one-time cleanup; repeated triggers do not double-close, panic, or change final result. |
| socket identity replacement protection | `TestRunDoesNotUnlinkReplacementSocket` | Replace the socket-path identity after listen; cleanup must preserve replacement rather than unlinking by path alone. |
| successful restart | `TestRunRestartsWithFreshSeed` | Two complete runs use fresh seed/runtime/listener state and both validly serve. |
| restart with missing seed | `TestRunFailsClosedOnRestartMissingSeed` | After a successful run, remove/make seed unavailable; restart must not listen or reuse prior secret/runtime. |
| valid existing-client OTP request | `TestRunServesValidOTPRequest` | Exercise the TASK-0013 client/protocol through the broker with a valid request; response remains compatible without exposing seed/OTP in test logs. |
| malformed existing-client request | `TestRunRejectsMalformedClientRequest` | Malformed request is rejected while process remains safe and no TASK-0013 runtime/IPC behavior changes. |
| no output leakage across success/error paths | `TestRunRedactsSecretFromErrorsAndLogs` | Capture controlled diagnostics from representative seed, schema, listener, and client failures; assert fixture secret and derived OTP do not occur. |

## Execution evidence

1. Before implementation, Main compares the eventual PLAN line-by-line with
   this mapping.  QA records a revision if an expectation/procedure changes;
   only Main may approve a scope/expectation change.
2. DEV supplies the named focused command results, including `go test
   ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend
   ./internal/ipc ./internal/lease` and race evidence for shutdown/client
   tests: `go test -count=1 -race ./cmd/codex-authority-broker`.  If VCS
   stamping blocks the focused command, record its exact failure once, then
   rerun the same scope with `GOFLAGS=-buildvcs=false`; classify the first
   result as environment evidence rather than product failure.
3. QA independently runs every named test (or the focused package command plus
   a name/result transcript), then `GOFLAGS=-buildvcs=false go test -count=1
   ./...`, `go vet ./...`, `gofmt -l` on modified Go files, and `git diff
   --check`.  `make check` is also attempted when present; if absent, record
   the exact missing-target evidence as `environment_issue`, not a product
   PASS/FAIL substitute.
4. QA compares the candidate and reviewer result to TASK scope, checks no
   runtime/IPC API changes, measures production SLOC with the project canonical
   method, and inspects captured output for prohibited data.  QA must not
   manufacture host-root fixtures, install a service, or use sudo.
5. For Unix socket restrictions, preserve the exact failing command and errno.
   Classify as `environment_issue` only when the same test passes in a known
   socket-capable runner or the denial is independently attributable to the
   sandbox; otherwise retain product failure evidence.

## Failure classification and release criteria

| Condition | Default classification | Return |
| --- | --- | --- |
| Any mapped behavior absent, including wipe-before-listen, descriptor validation, lifecycle, or client compatibility | `implementation_defect` | DEV |
| Existing TASK-0013 runtime/IPC behavior changes or full regression fails | `regression` | DEV (Main decides revert versus correction) |
| Matrix/procedure proves inconsistent with TASK contract or an unresolvable requirement is absent | `requirement_gap` | PLAN/Main |
| QA expectation/fixture/procedure is wrong while TASK is clear | `qa_plan_defect` | QA |
| Reproducible sandbox/tooling/Unix-socket denial blocks an otherwise valid check | `environment_issue` | QA/infrastructure; do not call it DEV fault |

QA PASS requires a passing result for every named row, redaction inspection,
scope/SLOC review, independent review disposition, race test, full regression,
format/vet/diff checks, and any environmental null reason explicitly accepted
by Main.  A partial suite, an aggregate test that omits a named row, or a
secret-bearing transcript is FAIL, not a waiver.

## PLAN reconciliation — PASS for DEV readiness

Compared after the independent baseline was frozen.  The PLAN preserves the
same two owned paths, immediate factory-return wipe before listen, descriptor
walk/no-follow/CLOEXEC/metadata controls, strict bounded schema/redaction,
ordering/lifecycle/restart/client acceptance, and no runtime or IPC API change.
Its aggregate test labels are compatible because this QA plan retains stricter,
individually runnable names for every aggregate subcase; the extra root/parent/
final stat/open, short-read, wrong-type, trailing JSON, UID, and canonical
base64 cases above close all PLAN-only detail.

The commands now include the PLAN focused package command, race test, full
`GOFLAGS=-buildvcs=false` regression, vet, gofmt, and diff check.  A direct
focused-command VCS-stamping failure is an `environment_issue` only when its
retry with `-buildvcs=false` passes; it never waives a mapped test.

The counted one-Lap sequence is agreed: preflight requires Linux/Go 1.23,
merged TASK-0013, only owned paths, and a socket-capable fixture runner; DEV
starts by minute 5, complete candidate/evidence by 20, independent REVIEW by
25, and independent QA plus Main-only Git closure by 30.  Planning plus pure
wait targets at most 20% of that interval.  Roles remain separated (DEV
`dev-luna/luna-xhigh`; PLAN/REVIEW/QA `Terra/medium`) and children do not write
Git.  The local +280/cumulative 1202 estimate is a Main warning, while global
1500 target and unconditional 1800 hard limit remain controls; stops cover
security/scope change, unreadable compression, target overflow/shedding audit,
and hard-limit risk without shedding seed/lifecycle/test evidence.

Lap 2 remains exceptional only with all four recorded facts: concrete residue,
no redesign/research, exactly one or two classified causes, and a demonstrable
fix within its first 20 minutes.  One replan per cause is allowed; Lap 3 is
prohibited.  No discrepancy remains that blocks the QA planning gate.  **QA
gate decision: PASS for DEV readiness**, subject to Main approval of both
planning artifacts; this is not a product QA PASS.

## Accounting

- Baseline status: TASK-first matrix preserved; deferred TASK-0015 PLAN
  comparison completed with no blocking discrepancy; QA gate PASS for DEV
  readiness, pending Main approval.
- `active_ms=unavailable`, because this QA runtime exposes no reliable
  stage-start timestamp and duration is not inferred.
- `wait_ms=0`; `retries=0`; classification `pass` (planning gate reconciled;
  no product test or product approval performed).
- Changed path: `tasks/TASK-0015/QA_PLAN.md` only.
