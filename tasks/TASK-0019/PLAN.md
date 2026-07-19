# PLAN — TASK-0019: minimal audit and source-free attested artifact, replanned

## Authority, route, and stop conditions

**PLAN evidence only.** This authorizes neither DEV nor any Git, GitHub,
workflow, host, deployment, approval, merge, commit, or operational-evidence
action. This is a product change: `PLAN -> independent TASK-first QA_PLAN ->
DEV -> same-candidate independent REVIEW and QA -> Main merge -> postmerge
live provenance verification`. Main alone owns Git and external operations;
children do not write `.git`, stage, commit, merge, or publish.

The baseline is 1407 canonical production SLOC. The readable limit is **+90**
and cumulative limit is **1497**; 1500's three-line remainder and the 1800
hard guard are not allowance. Stop and replan before candidate fixation if a
fourth production path, compression, a weakened required field/fail-closed
rule, sensitive audit input, an unprovable exact archive, or unavailable
postmerge attestation would be required. TASK-0018's uncommitted +102 draft is
superseded: this plan sheds only its random fixed-width identifier and
lease-internal invalidation/admission callback.

| Production path | Readable maximum | Planned responsibility |
| --- | ---: | --- |
| `internal/backend/runtime.go` | +57 | Runtime audit context/event encoder, process-local atomic ID sequence, one `auditMu`, bounded sink and permanent closure. |
| `internal/ipc/server_linux.go` | +14 | Pass authenticated numeric `SO_PEERCRED` UID privately to backend context only after request decoding. |
| `internal/lease/lease.go` | +19 | Add only a deadline accessor returning the active immutable deadline. |
| **Total** | **+90** | **1407 + 90 = 1497** |

Tests, workflow, and declarative inputs do not count under the canonical
definition. `main.go`, `protocol.go`, a lease `Invalidate`, and an
audit-admission callback are explicitly excluded.

## Audit boundary and implementation contract

`backend.New` continues to provide the broker's fixed operational audit sink
as stderr; its package-private test constructor receives a bounded test writer.
No environment or request selects a sink. `Server.handle`, after successful
peer authentication and exact request decoding, adds only the numeric UID to a
private context key before `Backend.Handle`; malformed, partial, credential
failure, and wrong-UID requests retain their existing no-backend/no-audit
behavior. The runtime assigns the ID, so no ID crosses IPC wire formats.

For each backend-admitted `ready`, `otp`, or `authorize` operation, Runtime
uses a monotonically incremented process-local atomic sequence and formats the
nonzero value as lowercase hexadecimal. It is fresh within the process, but is
neither random nor globally unique; journald process metadata supplies the
namespace. Overflow or absent/malformed audit context is deny and permanently
closed before an authority result is published.

The encoded JSON line has exactly these five keys and no optional fields:

```json
{"correlation_id":"1a","actor_uid":1000,"scope":"otp","result":"allow","lease_expiry":"2026-07-20T00:00:00.000000000Z"}
```

`scope` is only `ready`, `otp`, or `authorize`; `result` is only `allow` or
`deny`; `actor_uid` is numeric; ready and every deny have `lease_expiry:null`;
allowed OTP and authorize use the same immutable UTC RFC3339Nano deadline.
The fixed encoder has a compile-time bounded maximum and writes one newline.
It never accepts or serializes payload, OTP, seed, token, key, environment,
lease identifier, error, or error text.

Runtime owns exactly one `auditMu`. It holds it from authenticated audit-context
validation through handler decision, deadline capture, one bounded write, and
response publication. Thus accepted concurrent calls cannot cross IDs, actor,
scope, result, or expiry. OTP may create its lease before the write, but no
authorize call can observe it: all authority decisions are behind `auditMu`.
It obtains the candidate/active deadline through `State.Deadline()` only;
there is no lease invalidation and no callback-based admission API.

Immediately before writing, cancellation/shutdown/closed state changes the
final decision to deny. A short write, write error, malformed event, ID
overflow, or bound failure makes the current response deny and invokes
`Runtime.Close()` while `auditMu` remains held, before it is unlocked. The
closure cancels Runtime state and permanently denies current and future
authority calls; it is the sole way a prior/new lease becomes unusable. There
is no retry, recovery, cache success, compensating event, or direct mutation
of lease internals. A failed sink has one bounded attempted write, not a
durable audit event. A successful write is the decision linearization point.

`State.Deadline()` locks, expires stale state, and returns `(UTC deadline,
true)` only for an active lease; it cannot extend, revive, invalidate, or
admit a lease. Existing `VerifyAndActivate`, expiry, replay, rate, lifecycle,
and cancellation semantics otherwise remain unchanged.

## Source-free artifact contract

Add `.github/workflows/release.yml`, `deploy/pam/codex-authority`, and
`deploy/systemd/codex-authority-broker.service`; reuse existing declarative
`deploy/sudo/codex-authority`. The workflow runs on `main` push and manual
dispatch, builds exactly Linux/amd64 (`CGO_ENABLED=0`) broker, CLI, and sudo
helper into clean staging, then packages only those binaries and the three
declarative files. It adds sorted relative `SHA256SUMS` for exactly the six
payloads and creates deterministic `codex-authority-linux-amd64.tar.gz`
(sorted entries, numeric uid/gid zero, commit-derived `SOURCE_DATE_EPOCH`,
`gzip -n`). No installer, deploy, release, cache, token custody, or source
archive is introduced.

The exact seven-member archive manifest is:

```text
SHA256SUMS
bin/codex-authority
bin/codex-authority-broker
bin/codex-authority-sudo
deploy/pam/codex-authority
deploy/sudo/codex-authority
deploy/systemd/codex-authority-broker.service
```

Before attestation/upload, workflow logic must reject a directory, symlink,
absolute/traversal/hidden member and every extra member; explicitly reject
source/module files, `.git`, tasks/backlog/evidence, and seed/credential/token/
key-named paths. Extract into an empty directory and run `sha256sum -c`.
Attest the archive digest and upload only archive plus checksum artifact with
`if-no-files-found: error`.

Permissions are exactly `contents: read`, `attestations: write`, `id-token:
write`, and `artifact-metadata: write`. Every executable action is official
and full-SHA pinned (retain version annotation):

```text
actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
actions/attest@f7c74d28b9d84cb8768d0b8ca14a4bac6ef463e6 # v4.2.0
actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7.0.1
```

A deterministic local workflow/package test must reject tags/branches,
non-`actions/` owners, missing/mismatched pins, excess permissions, wrong
target/file list, and each manifest/checksum negative mutation. Candidate
evidence records official release provenance for these exact pins.

After Main merges, it alone observes/dispatches the main-bound workflow for
the merge SHA, downloads the named artifact into an empty directory, repeats
the exact manifest and checksum checks, then runs:

```sh
gh attestation verify codex-authority-linux-amd64.tar.gz --repo autotaker/codex-authority-broker
```

It records redacted repository, merge SHA, workflow/run identity, subject
digest, and successful binding. Local package evidence, another ref/digest,
or unavailable Actions/`gh` access cannot substitute; this `live-e2e` remains
blocked until performed.

## Required tests and candidate gates

| Case | Named test mapping | Mode / pass criterion |
| --- | --- | --- |
| Exact audit schema and redaction | New audit-focused cases in `internal/backend/runtime_test.go`, plus IPC integration in `internal/ipc/server_linux_test.go` | `focused-rerun`: strictly decode allowed/denied ready/OTP/authorize events; assert exactly five keys, UID, scope/result, null/immutable expiry, bounded line, one event, and no sensitive sentinels. |
| Atomic correlation and OTP visibility | Runtime barrier/concurrency cases and Unix-server cases in the same backend/IPC tests | `focused-rerun`: concurrent calls have distinct nonzero lowercase-hex IDs and correct tuples; no authorize observes an OTP lease until its allow write succeeds; race run passes. |
| Deadline accessor semantics | `internal/lease/lease_test.go` | `focused-rerun`: active deadline is immutable and UTC; expiry returns absent; no callback/invalidation API exists or is used. |
| Sink failure closure | `internal/backend/runtime_test.go` with error and short writers | `focused-rerun`: ready/OTP/authorize get no allow on failure; OTP has no usable authority; pre-existing and future authority are unusable; exactly one attempt, no retry/compensation. |
| Existing negative/lifecycle behavior | Existing `internal/ipc/server_linux_test.go`, `internal/backend/runtime_test.go`, `cmd/codex-authority-broker/main_test.go`, lease/TOTP tests | `evidence-review` plus `make check`: malformed/unauthorized/no-context paths retain generic denial; cancellation, shutdown, redaction, expiry, replay, rate, seed wiping, and no-cache behavior are not weakened. |
| Archive and workflow | New repository-native workflow/package test associated with `.github/workflows/release.yml` | `focused-rerun`: twice-build determinism, exact manifest/sums, forbidden-member mutations, official pin/provenance and permissions assertions. |
| Scope and cap | Candidate diff plus canonical counter | `evidence-review`: only declared paths, readable deltas within 57/14/19 and +90/1497, no secret-bearing output. |
| Main provenance | Main postmerge artifact inspection and `gh attestation verify` | `live-e2e`: exact artifact digest/repository/workflow/run/merged-main SHA bind together. |

DEV fixes `candidate_commit` and `candidate_tree` and records case ID, command,
fixture/cache, exit, result/artifact digest, negative detection, and any
unexecuted reason. REVIEW and QA begin independently and concurrently from
that exact tree; both run required checks, including REVIEW's `make check`.
Run `go test -count=1 -race ./internal/backend ./internal/ipc ./internal/lease
./cmd/codex-authority-broker`, `go test -count=1 ./...`, `go vet ./...`,
`make check`, `make task-check TASK=TASK-0019`, and `git diff --check` as the
repository provides them. A candidate/tree mismatch, audit/IPC/lease/workflow/
test change after QA, or any affected QA case prohibits carry-forward.

## Planner evidence and verdict

Read: root `AGENTS.md`, delivery contract, TASK-0019, TASK-0018 `FAILURE.md`,
TASK-0018 PLAN/QA plan, `backlog.json`, and current runtime/IPC/lease/main and
test surfaces. Current source has no audit context/sink/transaction, IPC calls
backend without peer UID context, and lease has no deadline accessor; it does
already provide the state lock/expiry needed by the narrow accessor. No
workflow/PAM/systemd artifact exists. The +102 TASK-0018 measurement (62/27/13)
is the direct predecessor; this 57/14/19 allocation is a forecast only and
must be remeasured readably before candidate fixation.

**Verdict: PLAN-ready for independent QA planning; DEV is not approved.**
