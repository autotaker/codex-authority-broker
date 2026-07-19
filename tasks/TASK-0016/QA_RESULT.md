# QA_RESULT — TASK-0016

## Decision: PASS

Independent product QA passes the TASK-first matrix. Source inspection and
targeted tests establish every acceptance condition; REVIEW attempt 3 supplies
independent socket-capable full/race/static/scope evidence. Sandbox-only socket
failures below are classified separately and are not product failures.

## TASK-first acceptance matrix

| Requirement | QA result and evidence |
| --- | --- |
| Exact protocol admission and payload rejection | PASS. `protocol.go` admits only version-1 `ready`, `otp`, and `authorize`; read and write reject every nonempty authorize payload. `TestReadRequestRejectsMalformedFrames`, `TestAuthorizeProtocolAdmission`, and `TestRequestRoundTripAndGenericErrors` passed, covering malformed/unknown/wrong-version, `{}`, `null`, array/string/number payloads, and payload-free round trip. |
| Active lease boundary | PASS. `handleAuthorize` defensively rejects payload/context/nil state and calls only `state.Active()`. `TestAuthorizeActiveLeaseBoundary` passed: pre-activation deny, post-OTP allow, deadline-minus-1ns allow, exact deadline and after deny. |
| Fresh runtime, cancellation, Close, and expiry races | PASS. `beforePublish` is a private nil-by-default post-decision/pre-publication barrier. `TestAuthorizeFreshRuntimeDenies`, `TestAuthorizeCallerCancellationFailsClosed`, and `TestAuthorizeExpiryAndCloseRaceFailsClosed` passed. The barrier tests observe an initially positive decision, make cancel/exact-expiry/Close win, then require empty-payload denial; final source gates recheck caller, closed/shutdown, and authorize-only `state.Active()`. |
| ready/OTP invariance and no mutation | PASS. Authorize calls neither `BeginReadiness` nor `VerifyAndActivate` and does not touch `challenge`. `TestAuthorizePayloadAndReadinessOTPNonInterference` and existing ready/OTP/challenge tests passed; authorize neither creates/replaces readiness, consumes OTP, nor extends the lease. |
| Exactly one custom slot | PASS. `maxOperations=4` installs three fixed handlers. `TestRegisterBoundsAndFixedAllowlist` passed: one custom registration succeeds; duplicate, second custom, and post-close registration fail; custom dispatch remains denied. |
| Bounded response/redaction | PASS. Success returns only `ipc.Response{OK:true}`; all denial helpers require `OK=false` and empty payload. No identity, deadline, token, lease, or reason serialization was added; `TestVersionContextAndRedaction` passed. |
| Scope and SLOC | PASS. Contract/index/PLAN/QA distinguish exactly four product/test paths from exactly six process-evidence paths. REVIEW attempt 3 verified the combined allowlist. Independent reviewed counts are runtime +12 and protocol +3: production delta **+15**, cumulative **1215** from 1200, below forecast 1220 and the >1230 replan trigger; 1250 target and >=1350 hard stop remain intact. Source is readable and uncompressed. |

## P1/P2 closure

- P1 PASS: the four product/test paths and six exact role/Main-owned process
  evidence paths are explicitly separate; process evidence is excluded from
  production SLOC.
- P2 PASS: the former pre-admission-only race proof is replaced by the precise
  `beforePublish` decision/publication barrier with deterministic cancel,
  exact-expiry, and Close winners. No new authority state or goroutine exists.

## QA execution and classification

QA targeted commands:

```text
GOCACHE=/tmp/task0016-qa-gocache GOFLAGS=-buildvcs=false go test -count=1 ./internal/ipc -run 'Test(ReadRequestRejectsMalformedFrames|AuthorizeProtocolAdmission|RequestRoundTripAndGenericErrors)$'  PASS
GOCACHE=/tmp/task0016-qa-gocache GOFLAGS=-buildvcs=false go test -count=1 ./internal/backend ./internal/lease  PASS
```

The first focused attempt could not read the default Go cache; classification:
`environment_issue`. One retry changed the precondition to a writable `/tmp`
cache. The combined IPC/backend/lease run then passed backend and lease but IPC
Unix-socket cases failed with `socket: operation not permitted`; classification:
`environment_issue`, not implementation defect. QA narrowed only the IPC rerun
to the protocol acceptance matrix, which passed. REVIEW attempt 3 independently
records socket-capable backend/IPC/lease race PASS, full `./...` PASS, vet,
gofmt, diff, JSON, and exact-scope PASS. Missing `make check`/`task-check`
targets remain the previously classified repository-tooling condition.

No unresolved implementation, QA-plan, requirement, environment-blocking, or
regression finding remains. Main retains Git/publication ownership.

Accounting: `active=semantic inspection + targeted QA`, `wait=0`, `retries=1`
(read-only cache; changed to writable cache),
`classification=PASS with non-blocking sandbox environment limitations`.
