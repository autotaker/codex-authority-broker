# PLAN — TASK-0018: minimal audit and source-free attested artifact

## Status, authority, and fixed scope

**PLAN evidence only.** This plan authorizes no DEV, approval, workflow run,
GitHub operation, host mutation, stage, commit, merge, or operational-evidence
write. DEV may start only after Main confirms the merged TASK-0012 REVIEW/QA
PASS, approves this PLAN and the independently authored TASK-first QA plan,
and records that the expected GitHub repository/Actions access is available.

This is a product change and follows the full path: `PLAN -> QA_PLAN -> DEV ->`
`same-candidate independent REVIEW + QA -> Main merge -> live postmerge
attestation verification`. PLAN/QA/REVIEW use Terra/medium; DEV uses
`dev-luna`/luna-xhigh; Main alone owns Git and external operations. Children do
not write `.git`, stage, commit, merge, or publish.

| Owned path | Responsibility |
| --- | --- |
| `cmd/codex-authority-broker/main.go` | Wire one broker-operational audit writer to stderr; preserve seed wiping and fixed daemon construction. |
| `internal/ipc/server_linux.go` | Obtain the authenticated SO_PEERCRED UID, create one fresh correlation ID only after a valid request is decoded, and pass only this audit context into the backend. |
| `internal/backend/runtime.go` | Perform the synchronous bounded audit decision, latch sink failure fail-closed, and make the response conditional on a matching successful audit write. |
| `internal/lease/lease.go` | Expose the immutable active deadline and make OTP activation conditional on an audit admission callback while the lease lock prevents publication. |
| `.github/workflows/release.yml` | Build, exact-allowlist, checksum, attest, and upload the one Linux artifact. |
| `deploy/pam/codex-authority` | Declarative PAM deployment input included in the artifact; no live installation. |
| `deploy/systemd/codex-authority-broker.service` | Declarative service deployment input included in the artifact; no live installation. |
| Named `*_test.go` files and workflow/deploy tests | Deterministic audit, lease, concurrency, packaging, and regression evidence. |

`internal/ipc/protocol.go` is deliberately not changed: correlation, actor, and
lease expiry never enter the wire request/response. No generic logging,
remote sink, retention, token custody, push, installer, automatic deployment,
PAM/sudo host mutation, canary, or release publishing is in scope.

## Current evidence and size contract

The merged tree at `101c75a` has 1407 canonical production SLOC (nonblank,
non-comment executable lines in non-test shipped source). The allowable net
delta is at most **+90**, producing cumulative **1497**. The target 1500 has
only three lines of reserve and is not a working allowance; 1800 is an
absolute guard, not an implementation budget. Workflows, tests, and
declarative PAM/systemd/sudoers files are excluded from that count.

| Production file | Forecast maximum | Why |
| --- | ---: | --- |
| `cmd/codex-authority-broker/main.go` | 8 | Inject stderr only; no configurable sink or secret-bearing configuration. |
| `internal/ipc/server_linux.go` | 25 | Fresh opaque ID generation and immutable context handoff after valid SO_PEERCRED/request admission. |
| `internal/backend/runtime.go` | 42 | Fixed event schema, serialized sink/write decision, failure latch, and expiry-aware result handling. |
| `internal/lease/lease.go` | 15 | Deadline accessor plus pre-publication OTP admission callback. |
| **Total** | **90** | Exact Task ceiling; Main must remeasure before candidate fixation. |

Stop before or during DEV if the readable delta would exceed any file forecast
or +90/1497 total, needs `protocol.go` or another production path, requires
payload/error/secret logging, weakens an audit field or fail-closed behavior,
cannot prove the exact archive allowlist, or cannot complete live provenance
verification. Do not compress code to fit. Any of those conditions is a
requirement/scope block requiring a split or replan.

## Audit design and concurrency contract

The operational output is the broker process's existing stderr; it is passed
as an `io.Writer` at construction, not selected from environment, request, or
seed. The writer seam is test-only and defaults to `os.Stderr`. The audit JSON
object has exactly these fields and no `omitempty` behavior that can omit a
required value:

```json
{"correlation_id":"32-lowercase-hex","actor_uid":1000,"scope":"otp","result":"allow","lease_expiry":"RFC3339Nano UTC"}
```

`correlation_id` is exactly 32 lowercase hexadecimal characters from 16 bytes
of `crypto/rand`; `actor_uid` is the numeric `SO_PEERCRED.Uid`; `scope` is
exactly `ready`, `otp`, or `authorize`; `result` is exactly `allow` or `deny`;
and `lease_expiry` is a JSON `null` for every deny and every ready result, or
the immutable UTC RFC3339Nano deadline for an allowed OTP/authorize decision.
The encoded line is bounded by a compile-time maximum and is written once with
one newline. It has no payload, OTP, seed, token, key, environment, lease
identifier, error, timestamp other than the required expiry, or error text.

1. `Server.handle` retains its silent denial for credential failure or wrong
   UID. It decodes a valid fixed operation before allocating the correlation
   ID. Malformed/oversize/partial requests retain the existing generic deny
   and do not receive audit context. For an allowed UID and decoded request,
   it cryptographically creates exactly one ID; random-source failure denies
   before backend invocation. It stores only `{id, uid}` in a private IPC
   context value. The ID factory is injected only by the server test helper.
2. `Runtime.Handle` accepts a valid operation only when that context value is
   present and structurally valid. It holds a dedicated audit-decision mutex
   across the state decision, audit write, and success publication. This
   serializes audit lines and prevents two concurrent requests from sharing or
   crossing actor, scope, result, or expiry. It is not a cache and does not
   change the existing handler limit, cancellation, shutdown, or fixed
   operation allowlist.
3. Every backend-reaching request makes one event decision. Denied ready/OTP/
   authorize writes one `deny/null` event before the generic false response.
   Allowed ready writes `allow/null`; allowed authorize reads the current
   immutable lease deadline and writes it as `allow`; neither operation can
   extend a lease. OTP receives its candidate deadline from lease state and
   writes its `allow` event while that state lock is held, before a new lease
   becomes visible. Thus an OTP allow cannot be observed by concurrent
   authorize before its corresponding audit write succeeds.
4. The lease method computes `now + leaseDuration` once, invokes the audit
   admission callback under its mutex, and creates the lease only if that
   callback succeeds. Its deadline accessor expires stale state before
   returning `(deadline, true)`, so no expiry is fabricated or mutable.
5. A short write, writer error, malformed context/event, or audit-size breach
   never publishes an allow. Runtime latches `auditBroken`, invalidates any
   active lease, and all later backend authority calls deny without attempting
   a recovery/retry/cache path. For a failed OTP write the activation callback
   returns false before lease creation. For a failed ready/authorize write the
   current response is false and the latch prevents subsequent authority.
   The failed write is still one attempted event; no second compensating event
   is emitted, because that could duplicate a decision and the sink is known
   unavailable. Shutdown/cancellation winning before publication remains a
   deny; no audit line may turn it into an allow.

This resolves the apparent sink-failure tension: an available sink records
exactly one final event for every valid backend operation; an unavailable sink
cannot durably record an event, so the broker performs exactly one bounded
write attempt and permanently fails closed rather than claiming a nonexistent
audit record or allowing authority. Tests must distinguish those two cases.

## Artifact and workflow contract

Create a single `release.yml`, triggered manually and on a `main` push. The
job has only `contents: read`, `attestations: write`, `id-token: write`, and
`artifact-metadata: write`; it uses no PAT, deployment credential, secret,
release publication, or cache. Each executable action is official and pinned
to this reviewed immutable release SHA (with the version comment retained):

```yaml
actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
actions/attest@f7c74d28b9d84cb8768d0b8ca14a4bac6ef463e6 # v4.2.0
actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7.0.1
```

`actions/attest` is the current official provenance action; the former
`attest-build-provenance` is not selected for a new workflow. A static
workflow test rejects tag/branch references, non-`actions/` actions, missing
pins, pin/annotation mismatch, or any permission beyond the four listed. The
DEV evidence must record each SHA and its official release provenance; an
unpinned action is a stop, not a later cleanup.

The workflow sets `GOOS=linux`, `GOARCH=amd64`, and `CGO_ENABLED=0`; builds
exactly `codex-authority-broker`, `codex-authority`, and
`codex-authority-sudo` into a clean staging directory. It copies only
`deploy/pam/codex-authority`, `deploy/systemd/codex-authority-broker.service`,
and the existing `deploy/sudo/codex-authority`. It creates `SHA256SUMS` over
those six payload paths using sorted, relative names. The canonical artifact
is `codex-authority-linux-amd64.tar.gz`, generated with sorted name order,
numeric owner/group zero, a fixed `SOURCE_DATE_EPOCH` derived from the checked
out commit, and `gzip -n`.

Before upload, a shell check lists the archive, sorts it, and requires this
exact seven-path manifest (no leading `./`, directory, symlink, or extra
member):

```text
SHA256SUMS
bin/codex-authority
bin/codex-authority-broker
bin/codex-authority-sudo
deploy/pam/codex-authority
deploy/sudo/codex-authority
deploy/systemd/codex-authority-broker.service
```

It separately rejects source suffixes (`.go`, `.mod`, `.sum`), `.git` paths,
`tasks/`, `backlog.json`, hidden metadata, seed/credential/token/key names,
absolute or traversal paths, and anything outside the exact manifest. It
extracts into an empty directory and runs `sha256sum -c SHA256SUMS`; failure
blocks attestation/upload. The workflow attests the archive's digest using
the pinned official action and uploads exactly the archive and its extracted
checksum file as one named workflow artifact with `if-no-files-found: error`.
The workflow artifact is not a GitHub Release.

After Main merges the approved tree, Main—not a child—dispatches or observes
the `main` workflow for the merge SHA, downloads its named artifact, extracts
it to an empty directory, reruns `sha256sum -c`, and repeats the byte-for-byte
sorted manifest comparison above. Main then runs:

```sh
gh attestation verify codex-authority-linux-amd64.tar.gz \
  --repo autotaker/codex-authority-broker
```

and records the successful subject digest, repository, workflow identity/run,
and merged commit binding only. It redacts/localizes any credential-bearing
environment or CLI diagnostics. A local archive, a unit-test result, an
attestation for another digest/ref/repository, unavailable `gh` auth, or a
workflow without `main` commit binding is not a substitute and leaves the
manual canary milestone blocked.

## Required deterministic evidence and QA mapping

QA must derive its independent matrix from `TASK.md` before reconciling this
PLAN. The planned DEV tests below map every Task condition to a concrete
method; each QA row must receive exactly one execution mode.

| QA condition | Deterministic DEV/QA evidence | Mode |
| --- | --- | --- |
| Allowed and denied `ready`/`otp`/`authorize` audit schema | Inject fixed ID and line sink; parse each line strictly, assert five exact fields/enums, bounded length, numeric peer UID, null/non-null expiry, and one event per backend call. Use sentinel payload/OTP/seed/token/key/error strings and assert none occur. | `focused-rerun` |
| Correlation/actor/result/expiry isolation under concurrency | Real temporary Unix server plus many allowed concurrent client calls; collect JSON lines, assert unique IDs, correct per-request UID/scope/result, exactly one line per accepted request, and no cross-association. Run race-enabled IPC/backend/lease suite. | `focused-rerun` |
| Lease expiry and OTP publication | Fake clock covers exact expiry, ready null expiry, OTP's immutable deadline, authorize deadline, and no extension. Barrier tests prove no authorize observes a new OTP lease before its allow line has completed. | `focused-rerun` |
| Sink failure fail-closed | Writer seams fail/short-write on ready, OTP, and authorize. Assert false response/no successful response bytes, one attempted write, no newly active lease after OTP failure, existing lease invalidated, all following authority denied, and no retry/second event. | `focused-rerun` |
| Existing security/lifecycle/protocol behavior | Existing plus adjusted tests cover malformed/unauthorized silent or generic denial, SO_PEERCRED, size bounds, OTP replay/rate/expiry, cancellation/shutdown, close/expiry race, no-cache sudo client, seed wipe/redaction, and full suite. | `evidence-review` |
| Exact source-free reproducible archive | Hermetic test invokes the workflow packaging script or its exact checked-in command twice from a clean temporary copy with fixed epoch; compare manifests/digests, assert allowlist/checksums and all negative-path/source/metadata/credential cases. Static parse confirms three builds and pinned official actions/least permissions. | `focused-rerun` |
| GitHub provenance on merged artifact | After merge, Main runs main-bound workflow, downloads the produced artifact, verifies the exact manifest/checksums, then runs the stated `gh attestation verify` against `autotaker/codex-authority-broker` and confirms digest/ref/workflow binding. | `live-e2e` |

Focused commands, once implementation exists, are:

```sh
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./internal/ipc ./internal/backend ./internal/lease ./cmd/codex-authority-broker
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 -race ./internal/ipc ./internal/backend ./internal/lease ./cmd/codex-authority-broker
GOFLAGS=-buildvcs=false GOCACHE="$(mktemp -d)" go test -count=1 ./...
go vet ./...
make check
make task-check TASK=TASK-0018
git diff --check
```

The workflow/package focused test must be added to the repository-native test
surface or documented as an exact shell test invoked by `make check`; QA must
reject a workflow-only assertion that cannot run deterministically locally.
Any unavailable Unix socket or GitHub capability is recorded once with a
redacted error and classified as environment; it cannot convert a
`focused-rerun` or `live-e2e` row to PASS by evidence review.

## Candidate gates, evidence, and stop verdict

Before DEV Main records the approved branch/worktree, baseline 1407, maximum
+90/1497, all owned paths, one QA mode per row, and no unapproved dirty-path
overlap. DEV fixes `candidate_commit` and `candidate_tree`; it records command,
fixture/cache condition, exit, artifact digest, test result, and unexecuted
reason for each case. REVIEW and QA start independently and concurrently from
that exact candidate; REVIEW runs `make check` and checks no payload/secret or
unowned-path leak. A changed candidate requires Main to choose a valid rerun
or the narrowly allowed carry-forward process; audit, lease, IPC, workflow,
test, acceptance, or QA-plan changes are categorically not carry-forward.

Main compares `merge_tree` with the reviewed candidate tree. Even if equal,
the provenance row remains a mandatory postmerge `live-e2e`; it cannot be
omitted. The manual canary/rollback reserve becomes eligible only after both
independent PASSes, merge, source-free artifact verification, and successful
main-bound `gh attestation verify` evidence.

## Planner evidence and verdict

| Item | Evidence |
| --- | --- |
| Sources read | Root `AGENTS.md`, delivery skill, TASK-0018, `backlog.json`, current main/broker IPC/runtime/lease source and tests, TASK-0017 PLAN/QA conventions, and official GitHub action/provenance documentation. |
| Current implementation gap | No audit context/sink/expiry transaction exists; no `.github` workflow or PAM/systemd deployment files exist. IPC currently authenticates UID before invoking `Backend.Handle`; lease activation currently publishes before any audit. |
| Arithmetic | Canonical current count 1407; forecast 8 + 25 + 42 + 15 = 90; projected 1497, with 3 lines to target 1500. |
| Secrets handling | Audit event is fixed-field only; no seed/OTP/payload/error/environment reads; commands/evidence require redaction. |
| Residual risks | The audit failure latch is intentionally availability-reducing; a sink outage denies all later authority. GitHub attestation depends on live Actions and `gh` access and remains blocked rather than locally substituted. |
| Active/wait/retry | Active: repository/task analysis and PLAN composition. Wait: Main approval and independently authored QA plan. Retry: 0. |

**Verdict: PLAN-ready for independent QA planning; DEV is not approved.** Stop
instead of implementing if the independently authored QA plan finds a semantic
conflict, the +90 forecast cannot preserve the listed invariants readably, or
live GitHub provenance cannot be verified after merge.
