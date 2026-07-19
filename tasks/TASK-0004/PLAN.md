# PLAN — TASK-0004: Versioned SO_PEERCRED IPC

## Approval boundary and preflight

Implement only a bounded, versioned Linux Unix-socket server in
`internal/ipc`.  It authenticates the kernel-reported peer UID before parsing
or dispatching one request to an injected backend.  It does not create a CLI,
daemon, sudo/PAM integration, push path, audit system, release artifact, or
secret/redaction policy.

**DEV profile:** `luna-xhigh` (`dev-luna`).  This is a narrow Go package using
the Linux standard-library socket/syscall surface and deterministic local
tests; no external dependency or privileged fixture is planned.

The 30-minute execution clock starts only after preflight confirms TASK-0003
is merged, the task worktree is usable, Linux Go is available,
`syscall.SO_PEERCRED` and `syscall.GetsockoptUcred` compile on the active
toolchain, and the focused Unix-socket fixture is ready.  The measured merged
cumulative production baseline is **233 SLOC**, leaving **417 SLOC** under the
TASK ceiling.  Tests use Go's `net` package directly; `socat` is neither
required nor to be added.  A failed prerequisite is not-started evidence.
DEV also requires a separately approved TASK-0004-only `QA_PLAN.md`.

## Owned paths and exclusions

| Path | Responsibility |
| --- | --- |
| `internal/ipc/protocol.go` | Fixed version, frame bound, strict request/response envelopes, and parser/encoder. |
| `internal/ipc/server_linux.go` | Linux Unix listener, kernel credential check, safe lifecycle, bounded dispatch, and backend interface. |
| `internal/ipc/protocol_test.go` | Version, framing, malformed, trailing-data, and size-bound tests. |
| `internal/ipc/server_linux_test.go` | Real Unix-socket peer checks, injected credential-extraction failure, path/lifecycle, backend, and unavailable-server tests. |

Do not change `internal/lease`, add commands or services, define Codex access,
accept configuration from argv/environment/files, implement TOTP behavior,
or add CLI, redaction, sudo, push, audit, packaging, delivery, or operator
work.  The package emits no logs and tests use only non-secret payloads.

## Protocol and backend contract

Use a four-byte big-endian payload length followed by one JSON envelope.  Set
`ProtocolVersion = 1` and a fixed `MaxFrameBytes = 4096`.  A zero-length frame,
declared length above the bound, short header/body, invalid JSON, trailing JSON,
unknown field, missing/unknown operation, or version other than 1 is denied
without backend invocation.  The parser reads the four-byte header first and
does not allocate the declared body until the bound passes.  One connection
carries exactly one request and one response, then closes; pipelining and
streaming are out of scope.

Keep the public surface small and transport-focused:

```go
type Request struct {
    Version   uint16
    Operation string
    Payload   json.RawMessage
}

type Response struct {
    Version uint16
    OK      bool
    Payload json.RawMessage
}

type Backend interface {
    Handle(context.Context, Request) (Response, error)
}

type Config struct {
    Path       string
    AllowedUID uint32
}

func Listen(Config, Backend) (*Server, error)
func (s *Server) Serve(context.Context) error
func (s *Server) Close() error
```

Exact names may change for idiomatic implementation, but the boundary may not:
only a validated, authorized request reaches `Backend`; transport errors and
backend errors produce a fixed generic failure with no input or error text.
The protocol package does not interpret lease or TOTP semantics.  Tests inject
a recording backend; production wiring belongs to later tasks.

## Peer authorization, bounds, and lifecycle invariants

`Listen` supports only an absolute, cleaned Unix-socket path.  It walks the
existing parent path with `Lstat`, rejects symlink or non-directory components,
and requires the final parent to be owned by the effective server UID and not
group/world writable.  It refuses every pre-existing final path (regular file,
directory, symlink, or socket) and never removes one to make progress.  After
binding it applies restrictive `0600` permissions.  TASK-0005 may add its own
approved access provisioning; this task does not broaden filesystem access.

On accept, obtain credentials from the accepted `*net.UnixConn` through
`SyscallConn` and `syscall.GetsockoptUcred(fd, SOL_SOCKET, SO_PEERCRED)`.  Trust
only the returned `Ucred.Uid`, never a UID in request data.  Credential lookup
error or UID mismatch closes/denies before reading or dispatching.  The public
constructor always installs this kernel credential reader; an unexported test
constructor may inject a credential function solely to deterministically test
credential-extraction/syscall failure.  It is not evidence for unauthorized
peer rejection.

Use fixed conservative bounds: at most 16 accepted handlers at once, one frame
per connection, the 4096-byte frame maximum, and finite read/write deadlines.
When capacity is full, close the new connection without dispatch.  A partial or
stalled request times out and fails closed.  `Serve` stops accepting when its
context is cancelled or `Close` is called, closes active connections, waits for
handlers, and is safe under repeated `Close` calls.  Cleanup removes the path
only if `Lstat` still identifies the same socket created by this server; a
missing or replaced path is left untouched and reported as a lifecycle error.
No request can be dispatched after shutdown begins.

## DEV sequence and focused evidence

1. Implement the bounded frame encoder/strict decoder and sentinel transport
   errors in `protocol.go`.  Do not include request content in errors.
2. Implement Linux path validation, listener creation, real SO_PEERCRED reader,
   bounded accept/handler lifecycle, generic denial, and idempotent shutdown in
   `server_linux.go`, using only the Go standard library.
3. Add table-driven parser tests for every malformed/version/size boundary,
   including exactly 4096 bytes versus 4097, short reads, unknown fields, and
   trailing data.  Assert the backend call count remains zero on every denial.
4. Add Unix-socket tests proving current-UID success through the real syscall.
   For the unauthorized-peer case, configure `AllowedUID` to a value different
   from the current test process UID, connect from that process over the real
   Unix socket, and prove the kernel-reported current UID is rejected and the
   backend call count stays zero.  This requires no distinct OS user and must
   not substitute an injected credential.  Use the private extraction seam
   only for deterministic syscall-failure denial.  Also cover backend-error
   denial, one-request-only behavior, partial-frame failure, and unavailable
   behavior before listen and after close.  Test refusal of relative,
   symlinked, unsafe-parent, and pre-existing final paths, plus
   replacement-safe cleanup.  Use temporary directories and standard-library
   clients only; no root/user creation, network, `socat`, shell fixture, or
   real secret is needed.
5. Run `gofmt` on owned Go files, then record concise evidence for `go test
   ./...`, `go test -race ./...`, `go vet ./...`, `test -z "$(gofmt -l .)"`,
   the exact cumulative SLOC count below, and `git diff --check`.  Run another
   repository check only if it already exists and applies; do not add a
   Makefile or planning-only script.

The initial target was **240–320 added production SLOC**, for cumulative
**473–553**, within the 650 ceiling and 417-SLOC available headroom.  The
Lap02 re-estimation below supersedes that estimate.  If the candidate exceeds
650, stop, follow `backlog.json`'s exact ordered shedding list, and obtain
revised PLAN and QA approval before resuming.  Versioned, bounded, fail-closed
SO_PEERCRED IPC is mandatory and cannot be shed.  A mandatory-v1 total over
1500 is a requirement gap, never a reason to compress.

## Lap02 re-estimation and remaining DEV

The >90% trigger has fired.  The observed candidate is **233 baseline + 384
IPC = 617 cumulative production SLOC**, leaving only **33 SLOC** below 650;
tests add 256 lines but are excluded from production SLOC.  Basic and full
checks have passed outside the sandbox.  The candidate is not yet gate-ready:
tests are missing for replacement-safe cleanup, a pre-existing socket,
handler saturation, and shutdown with an active client.

Read-only inspection found one implementation defect that those tests must
not merely document: `handle` currently holds `Server.mu` while calling
`Backend.Handle`, but `Close` needs the same mutex before it can set closing,
cancel the backend context, and close active clients.  A blocking backend can
therefore deadlock shutdown.  DEV must, while holding `Server.mu`, reject a
handler if closing has begun and otherwise classify the already waitgroup-
tracked handler as in flight; it must then unlock before `Backend.Handle`.
`Close` must acquire the mutex, set `closing`, cancel the shared backend
context, close the listener and active connections, unlock, and wait for the
in-flight handler.  Thus no newly parsed request is admitted for dispatch
after shutdown begins, while an already admitted backend observes context
cancellation without blocking `Close` on the mutex.

The SLOC ceiling remains credible, conditionally, without compression or
control shedding.  Two client-side helpers in production—`writeRequest` and
`readResponse`—are unexported and referenced only by tests.  This task owns a
server, not the TASK-0005 client; move equivalent fixture helpers to `_test.go`
or defer them to TASK-0005.  Also remove the redundant assignment of response
version immediately before `writeResponse`, which already assigns it.  These
are scope corrections, not one-line packing or combined error handling, and
should recover roughly 18 production SLOC before the shutdown correction.
Do not simplify the strict decoder, frame bound, real SO_PEERCRED check,
handler semaphore, deadlines, safe path walk, socket identity check, generic
failure behavior, waitgroup, or idempotent close.

Remaining DEV is limited to:

1. Apply only the client-helper scope correction and the mutex/shutdown fix
   above, keeping ordinary gofmt-formatted control flow and named errors.
2. Add a blocking backend that signals entry and returns on `ctx.Done()`;
   prove `Close` cancels it and completes within a bounded test deadline,
   without post-shutdown backend admission.
3. Add saturation evidence: occupy all 16 handler slots with bounded partial
   clients, prove an additional client is closed and never dispatched, then
   close the server without leaked handlers.
4. Add lifecycle evidence that a pre-existing live socket is refused and left
   intact, and that replacing the created socket path before `Close` yields
   `ErrLifecycle` without deleting the replacement.
5. Re-run all focused and full commands already listed, the race detector,
   exact SLOC count, and independent REVIEW/QA gates.

The revised target is **600–625 cumulative production SLOC** after the
scope correction and shutdown fix; test additions do not affect it.  Stop DEV
immediately if any remaining acceptance test requires a new production
feature, protocol operation, client API, dependency, or path outside the
approved files.  Stop and return to PLAN/QA if projected or measured
cumulative production SLOC exceeds **640**, preserving at least 10 SLOC of
review-fix margin; exceeding 650 is an unconditional FAIL.  If the mandatory
IPC behavior cannot fit through legitimate deletion of out-of-scope helpers
and clear idiomatic code, classify a requirement gap—do not shed the control,
compress functions, collapse errors, shorten names, or weaken tests.

## Cumulative SLOC and independent gate evidence

Count all candidate non-test production Go files, including the merged 233
SLOC and new `internal/ipc` files.  Exclude tests, generated/vendor files,
configuration, and task documents.  REVIEW and QA independently run this exact
command at candidate root and record file subtotals and `TOTAL <= 650`:

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

Independent REVIEW records: approved-path/scope inspection; real
SO_PEERCRED/current-UID success; real-socket rejection with a deliberately
nonmatching configured UID and zero backend calls; injected syscall-failure
evidence; parser version, exact-bound, malformed, backend-not-called,
unavailable, lifecycle, and path-safety evidence; test/race/vet/gofmt/diff
results; and the cumulative SLOC count.  It rejects semicolon/one-line packing,
collapsed error handling, cryptic names, removed security comments, or
functions combined solely for the cap.  QA independently repeats focused
acceptance and SLOC/scope checks after REVIEW PASS.  Any FAIL returns to its
responsible gate; this PLAN authorizes no product use, stage, commit, or merge.
