# REVIEW RESULT — TASK-0008: sudo live check and no cache

## Verdict

**PASS.** Independent source review finds no release-blocking defect in the candidate. The only local check limitations are repository/sandbox environment conditions recorded below; they are not substituted for product evidence.

## Boundary and contract review

| Review item | Result | Evidence |
| --- | --- | --- |
| Exact candidate scope | PASS | Worktree candidate is confined to the four permitted product/test paths: `cmd/codex-authority-sudo/main.go`, `cmd/codex-authority-sudo/main_test.go`, `deploy/sudo/codex-authority`, and `deploy/sudo/codex-authority_test.go`. Accompanying changed/untracked process files are within the seven explicitly permitted evidence paths. No backend, IPC/protocol, daemon, PAM installation, Git, or operational-policy candidate change was found. |
| One fixed live decision | PASS | `run` constructs exactly `Request{Version: ProtocolVersion, Operation: OperationAuthorize}` with no payload, executes one `callWithContext`, and has no loop/retry/fallback/ready/otp path. Success additionally requires matching protocol version, `OK`, and an empty response payload. |
| Non-authority inputs and redaction | PASS | argv is ignored; stdin is not read; the implementation does not read environment/file/cache/process-global authority inputs. Allow is silent; every deny is the bounded `request denied\\n` line. Tests cover stdin independence and sentinel redaction. |
| Fail-closed/no cache | PASS | Transport/context/validation error, version mismatch, `OK=false`, or any response payload returns nonzero. No prior response or timestamp state is retained. Named tests independently exercise expiry, unavailable daemon, restart, malformed variants, unauthorized payload, and two invocations. |
| Production sudo policy | PASS | `deploy/sudo/codex-authority` is exactly `Defaults:codex-fixture timestamp_timeout=0`; policy tests reject grants, PAM/client invocation, global/broad selectors, and imperative cache clearing. The fixture-only PAM/command scaffolding is not introduced as production policy. |
| Size and local controls | PASS | Production SLOC definition excludes declarative sudoers configuration. Independent executable-source count for `main.go` is 38; actual delta is **+38** from merged baseline **1215**, giving cumulative **1253**. This remains below the 1325 reapproval trigger, 1350 target stop, and 1450 hard stop. |

## Check evidence

| Command / evidence | Result | Classification |
| --- | --- | --- |
| `GOCACHE=/tmp/task0008-review-gocache go test -race ./cmd/codex-authority-sudo` | PASS | Candidate-focused race check passed. |
| `GOCACHE=/tmp/task0008-review-gocache go test -race ./internal/ipc` | ENVIRONMENT | The restricted review sandbox rejects Unix-socket creation (`socket: operation not permitted`), so the package's real socket tests cannot run here. This is not attributed to the candidate. |
| `GOCACHE=/tmp/task0008-review-gocache go vet ./cmd/codex-authority-sudo`; `gofmt -l` check; `git diff --check`; `jq -e . backlog.json` | PASS | Vet passed for the candidate package; gofmt output count was zero; diff and JSON checks passed. Main separately supplied socket-capable full/race/vet/format/diff PASS evidence. |
| `make check` | ENVIRONMENT | `make: *** No rule to make target 'check'. Stop.` The worktree supplies no such target. |
| Main elevated fixture evidence | PASS (reviewed handoff) | Private mount namespace with tmpfs `/etc`; disposable dedicated and distinct identities; actual PAM → client → Unix server allow; second invocation denial/no timestamp reuse with exactly two calls; unavailable, malformed, unexpected-payload, and distinct-identity denial; fixture `visudo`; and exact host hash/list rollback comparison all reported PASS. |

The supplied socket-capable Main evidence also reports full suite, race, vet, format, and diff PASS. It closes the sandbox-only Unix-socket execution gap; it does not broaden this reviewer's source-review scope.

## Required named behavior review

`TestLiveLeasePermitsPerInvocation`, expiry, unavailable, restart, malformed, unauthorized, and two-consecutive tests each observe a distinct fixed request and the correct allow/deny outcome. `TestArgvAndLogRedaction` verifies that an argv sentinel and an unexpected response payload do not leak. Policy tests assert dedicated-only `timestamp_timeout=0` and prohibit production PAM/client coupling. The Main fixture evidence supplies the required actual-sudo, dedicated/unauthorized-identity, and rollback observations not safely runnable inside this sandbox.

## Accounting

Independent REVIEW source/contract inspection and candidate-focused race test completed. `active_ms=null`, `wait_ms=null`, `retries=0`: the runtime does not provide an authoritative same-attempt paired elapsed record, so null is not recorded as zero. Local IPC socket and `make check` limitations are classified as environment evidence only. No product, Git, operational, stage, commit, or merge action was performed by this reviewer.
