# PLAN — TASK-0005: Codex CLI, socket access, and redaction

## Approval boundary and fit decision

Add one narrow, non-MCP Codex CLI, the missing IPC client transport, and
numeric UID/GID socket provisioning.  The CLI exposes exactly readiness and
OTP submission; OTP input is stdin-only and all user-visible results are fixed
redacted text.  This task does not add a daemon or wire IPC operations to the
lease/TOTP backend.  A later approved assembly boundary will implement that
backend; TASK-0005 proves the CLI can issue the two fixed requests to an
injected/fixture backend over the real TASK-0004 transport.

**DEV profile:** `luna-xhigh` (`dev-luna`).  The work is bounded Linux Go CLI
and socket plumbing using the standard library, with no privileged operation,
external service, or new dependency.

The merged cumulative production baseline is **600 SLOC**, leaving 220 under
the TASK cap of 820.  The original PLAN estimated **105–135 added production
SLOC** (cumulative **705–735**) and required a stop above 738.  That estimate
was exceeded and is retained only as historical planning evidence.  The
Revision 2 re-estimation below supersedes its target and stop decision.

The 30-minute execution clock starts only after preflight confirms TASK-0004
is merged, Linux Go and the focused real-socket/subprocess fixtures are ready,
the worktree is usable, and the 600-SLOC baseline is reproduced.  No current
host `codex` account is required: fixtures use explicit numeric UID/GID values.
A failed prerequisite is not-started evidence.  DEV requires this PLAN and the
independently reconciled `QA_PLAN.md` to be approved first.

## Revision 2 re-estimation at the >90% gate

The observed unfinished candidate is **600 baseline + 151 TASK-0005 skeleton
= 751 cumulative production SLOC**, leaving **69 SLOC** below the unchanged
820 cap.  The initial estimate missed by 16 SLOC at its upper bound because the
combined CLI, bounded client, fixed operation migration, and optional socket
ownership provisioning required more ordinary Go structure than projected.
This is an `estimate_miss/replan`, not evidence authorizing compression or
feature removal.

Read-only inspection shows that the production skeleton already contains all
planned TASK-0005 responsibilities: the fixed `ready`/`otp` CLI, stdin-only
six-digit input, fixed output/errors, the bounded Unix client, strict fixed
operations, and optional numeric UID/GID provisioning with `0600` default and
`0660` provisioned mode.  No daemon/backend wiring is required by this task.
The observed unfinished work is:

1. `protocol_test.go` still declares test-local `writeRequest` and
   `readResponse` helpers that now duplicate the production client framing.
2. TASK-0004 tests still use the removed `OperationRequest` instead of one of
   the two fixed operations.
3. TASK-0005 client, CLI, provisioning, and redaction/capture tests have not
   been added.

All three are test-source work and are excluded from production SLOC.  Remove
only the duplicated `_test.go` helpers, migrate existing fixtures to fixed
operations, and add the missing tests.  Do not discard or rewrite the current
production skeleton merely to chase the original estimate.  Production edits
after this replan are allowed only when a new focused test demonstrates a
TASK-0005 acceptance defect; record the failing test and SLOC delta before the
edit.  The small duplicated `socketPath`/client-selection flow in `main.go` may
be simplified only if it improves clarity while preserving the injectable
test seam—never through packed conditionals or collapsed error handling.

The revised credible target is **745–775 cumulative production SLOC** (145–175
added to the 600 baseline).  The lower end allows ordinary deletion of truly
duplicated production glue; deletion is not required.  The upper end reserves
24 SLOC for clear fixes discovered by the mandatory tests.  Stop and return to
PLAN/QA if projected or measured production SLOC exceeds **790**, if a test
requires a new production responsibility/path/dependency, or if backend/daemon
wiring appears necessary.  Exceeding **820** is an unconditional FAIL.

No feature shedding is approved at 751: all current responsibilities are
within TASK scope and all acceptance controls remain mandatory.  If a later
candidate cannot fit below 820, apply only `backlog.json`'s exact ordered
feature-shedding process with new PLAN/QA approval.  TASK-0005 has no rich
status/JSON UX to shed, and SO_PEERCRED fail-closed authorization plus OTP
non-disclosure from argv/log/output/error cannot be shed.  If those mandatory
controls still cannot fit idiomatically, conclude `requirement_gap`; do not
shorten names, pack lines, collapse errors, weaken capture tests, or move
production behavior into tests.

## Owned paths and exclusions

| Path | Responsibility |
| --- | --- |
| `cmd/codex-authority/main.go` | Fixed `ready` and `otp` commands, stdin handling, fixed output/error, and exit status. |
| `cmd/codex-authority/main_test.go` | Argument/input/output/error capture tests and real-socket subprocess fixture. |
| `internal/ipc/client_linux.go` | Bounded request/response Unix client with generic errors and context deadline. |
| `internal/ipc/client_linux_test.go` | Round-trip, unavailable/malformed response, deadline, and generic-error tests. |
| `internal/ipc/protocol.go`, `protocol_test.go` | Admit only the two named actions and share bounded client framing. |
| `internal/ipc/server_linux.go`, `server_linux_test.go` | Optional numeric socket ownership/group provisioning while preserving SO_PEERCRED authorization and TASK-0004 safety. |

Do not modify `internal/lease`, add a backend/daemon/MCP server, create users or
groups, invoke sudo, or add PAM, push, audit, release, canary, packaging,
service, installer, general shell, or operator workflow.  Do not add logs,
telemetry, configuration files, environment-based OTP input, or dependencies.

## Fixed CLI and client contract

The command accepts only `ready` or `otp`, with an optional non-secret socket
path option for the test/operator-selected endpoint.  The default path is one
fixed package constant.  Unknown actions, extra positional arguments, an OTP
argument, and OTP-looking flags are rejected before transport.  This is a
normal CLI, not MCP, JSON-RPC, or an extensible command dispatcher.

`ready` sends an IPC request with the fixed readiness operation and no payload.
`otp` accepts no code argument and reads one bounded line from stdin.  It
requires exactly six ASCII digits after removing the line ending and rejects
missing, extra, or oversized input locally with the same fixed denial.  It
does not read an OTP from any environment variable, write it to a file, log
it, or pass it to a child process.  The in-memory value exists only long enough
to encode the bounded request payload; do not claim guaranteed Go-runtime
zeroization.

Restore the TASK-0004 client-side framing only as a small exported transport:

```go
type Client struct { Path string }
func (c Client) Call(context.Context, Request) (Response, error)
```

Exact names may vary, but `Call` validates the absolute cleaned socket path,
uses `net.Dialer.DialContext` with network `unix`, sets a finite deadline,
writes one bounded version-1 request, reads one strict bounded response, and
closes.  It returns fixed sentinel errors only; raw JSON, socket error text,
request payload, and backend response payload never appear in an error.
Protocol operations are exactly `ready` and `otp`; unknown operations remain
fail-closed before backend dispatch.

The CLI ignores response payload for display.  Success prints one fixed status
line appropriate to the action; every local, transport, server, or backend
denial prints the same fixed error line to stderr and exits nonzero.  It never
echoes stdin or formats an underlying error.  No production logger is added.

## Socket provisioning and authorization policy

TASK-0004 currently creates a server-owned `0600` socket and authorizes one
numeric `AllowedUID` via real `SO_PEERCRED`.  Extend its `Config` with an
optional explicit access policy containing numeric owner UID and group GID.
When absent, preserve TASK-0004's `0600` behavior.  When present, after bind
and socket identity capture, set the requested ownership and fixed `0660`
mode; any ownership/mode failure closes the listener and removes only the
same inode using the existing replacement-safe cleanup.  Keep the parent-path,
symlink, identity, deadline, saturation, shutdown, and parser controls intact.

The intended future policy is server owner plus the dedicated `codex` group
on the socket, while `AllowedUID` is the dedicated `codex` UID.  Group mode
permits connection, but kernel `SO_PEERCRED` still admits only that exact UID;
another member or UID is denied before parsing/backend dispatch.  This package
accepts numeric IDs from its caller and never calls `os/user`, creates an
account/group, or assumes `/etc/passwd` contains `codex`.

Tests provision with the current numeric effective UID/GID so they need no
privilege or host account.  Authorized evidence uses `AllowedUID = geteuid()`.
Unauthorized evidence uses the same accessible real socket but deliberately
sets `AllowedUID` to a different numeric UID, proving the kernel-reported peer
is rejected and the backend is not called.  Injected credentials are not a
substitute for that evidence.

## Revised remaining DEV and redaction evidence

1. First make the existing test suite compile without product changes: remove
   the duplicate test helpers, replace every legacy `OperationRequest` fixture
   with the appropriate fixed operation, and prove unknown operations still
   fail closed.
2. Test the existing client for ready/OTP round trips, unavailable/malformed
   response, exact frame bound, context/deadline, and fixed generic errors.
3. Test optional numeric socket ownership/group provisioning without weakening
   exact-UID SO_PEERCRED authorization or safe cleanup.  Cover default `0600`,
   provisioned `0660`, same-inode cleanup on provisioning failure, real
   authorized UID, and real mismatched-UID denial with zero backend calls.
4. Unit-test the existing fixed CLI through injected stdin/stdout/stderr and
   client-call seams: action grammar, positional/environment OTP rejection,
   exact-six-digit input, oversize input, fixed output, and generic errors.
   Do not create backend wiring.
5. Build/run the CLI against a real temporary Unix server and recording
   backend.  Feed a documented synthetic OTP only through stdin.  Capture the
   child argv, environment, stdout, stderr, returned errors, and any test log;
   scan every capture for the fixture and require zero occurrences.  Prove
   readiness has no OTP payload, OTP dispatch reaches the backend once, and
   rejected transport/backend paths do not echo the fixture.  Also prove an
   attempted OTP positional/environment input is rejected without transport.
   Include a negative-control leak so the scanner is demonstrably effective.
6. Run `gofmt` on owned files and record concise evidence for `go test ./...`,
   `go test -race ./...`, `go vet ./...`, `test -z "$(gofmt -l .)"`, the exact
   SLOC count, and `git diff --check`.  Run an additional repository check only
   if already present/applicable; do not add a Makefile or planning script.

## Cumulative SLOC and independent gates

Count all candidate non-test production Go files, including the merged 600
SLOC.  Tests are excluded.  REVIEW and QA independently run this exact command
and record file subtotals and `TOTAL <= 820`.  Revision 2 has satisfied and
superseded the historical >738 stop; the new >790 stop/replan threshold above
is binding for remaining DEV:

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

Independent REVIEW records CLI real-socket readiness/OTP evidence, real
authorized and unauthorized peer results, provisioning modes/identity-safe
failure, full secret capture-scan results, fixed errors/output, scope, all
checks, the observed-versus-final SLOC delta, and cumulative SLOC.  It rejects semicolon/one-line packing, collapsed
errors, cryptic names, removed security comments, or functions combined only
for LOC.  QA independently repeats acceptance, capture scans, SLOC, and scope
after REVIEW PASS.  Any FAIL returns to its responsible gate.  This PLAN
authorizes no DEV before both plans are approved, and no stage, commit, merge,
backend, or later-feature work.

The prior QA plan still contains the historical `<=738` pre-DEV target and
must be independently reconciled to this Revision 2 evidence before approval.
Until that occurs, the gate remains closed even though a credible <=820 path
exists.
