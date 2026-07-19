# TASK-0018 QA plan — TASK-first independent matrix

## Independence, authority, and stop conditions

This plan was drafted from `TASK.md` before reading `PLAN.md`. QA owns
candidate-bound evidence and failure classification only. It must not change
product, deployment, workflow, Task, plan, Git state, GitHub, or a live host.
QA starts only after Main has recorded TASK-0012's merged independent PASS,
approved both plans, and DEV has fixed one `candidate_commit` and
`candidate_tree`. REVIEW and QA start independently from that same tree; a
REVIEW PASS is not a QA precondition.

The permitted implementation surface is the four Task production paths, their
named tests, the release workflow, and declarative deployment inputs. GitHub
push/token custody, generic/remote logging, retention, installer/deployment,
host PAM/sudo mutation, manual canary, release publication, and all v2 work
are prohibited. No command output, fixture, manifest, or QA evidence may
contain a request payload, OTP, seed, token, key, credential, environment
content, lease identifier, internal error text, or GitHub authentication data.

Stop and return **requirement/process block** before candidate fixation if any
of these occur: production delta exceeds +90 or cumulative 1497; a fifth
production path (including `protocol.go`) is needed; a required audit field,
source-free proof, immutable expiry, or fail-closed condition is weakened; an
archive member cannot be exactly allowlisted; an action is not an official
immutable pin; or postmerge GitHub provenance cannot be independently verified.
The 1500 target's three-line reserve and the 1800 hard guard are not usable
implementation allowance; code compression is not a remedy.

## Candidate evidence requirements

For every case below DEV and QA evidence must name the case ID,
`candidate_commit`, `candidate_tree`, command, test/fixture and cache
condition, exit status, bounded artifact digest or result digest, negative
detection capability, and unexecuted reason where relevant. QA independently
checks that the command/tests actually detect the stated regression and were
not weakened. Candidate/tree mismatch, missing high-risk evidence, or an
unknown impact is a FAIL, not evidence-review PASS.

An audit JSON event is valid only if it has exactly five keys:
`correlation_id`, numeric `actor_uid`, `scope`, `result`, and `lease_expiry`.
The ID must be fresh 32-lowercase-hex, scope exactly `ready`/`otp`/`authorize`,
result exactly `allow`/`deny`, and expiry either JSON null (every deny and
every ready) or the immutable UTC RFC3339Nano expiry of an allowed OTP or
authorize decision. The encoded single line must be bounded and contain no
other field or sensitive sentinel. A valid operation that reaches the backend
has one decision/event; with an unavailable sink, it instead has exactly one
bounded write attempt and must fail closed without claiming a durable event.

## Acceptance matrix

| ID | Acceptance and required mutation/negative control | QA execution mode and named check | Pass / fail-closed condition |
| --- | --- | --- | --- |
| Q18-01 | For allowed and denied ready, OTP, and authorize requests, inject deterministic ID, clock, SO_PEERCRED UID, and bounded sink; strictly decode every line. Mutate scope/result/field count/type/expiry and use unique payload, OTP, seed, token, key, environment, and internal-error sentinels. | `focused-rerun`: `go test -count=1` for IPC/backend/lease/main audit tests, plus strict schema/sentinel test. Hermetic, deterministic fixtures reproduce the decision boundary. | Exactly one bounded five-field event for every valid backend operation; actor is peer UID, fields/enums and null/non-null expiry match exactly, and no sentinel or extra data leaks. Any leak, missing/duplicate event, fabricated/mutable expiry, or malformed event fails. |
| Q18-02 | Issue concurrently interleaved allowed/denied operations with distinct test UIDs, scopes, clocks, and barriers. Verify no ID reuse or cross-association of actor, result, scope, or expiry; race OTP activation against authorize. | `focused-rerun`: concurrent Unix-socket audit test and `go test -count=1 -race ./internal/ipc ./internal/backend ./internal/lease ./cmd/codex-authority-broker`. Bounded local concurrency is hermetic/deterministic enough when barrier assertions are present. | Each accepted request has one unique correlation ID and its own correct tuple. An authorize cannot observe a newly created OTP lease before that OTP allow write succeeds. Race detector and barriers pass. |
| Q18-03 | Exercise allowed OTP and authorize at/around expiry with a fake clock; attempt ready and denies; mutate time after activation and invoke repeated authorize. | `focused-rerun`: lease/runtime expiry tests with injected clock. | Ready and every deny serialize null; OTP and authorize use the same immutable deadline; authorize neither extends nor revives a lease; stale deadline denies. |
| Q18-04 | Make sink return error/short write for ready, OTP, and authorize, including after an existing lease. Observe response bytes, write count, lease state, and subsequent requests. | `focused-rerun`: failing-writer tests in backend/lease, with deterministic response capture. | No successful response or usable new lease is published without its successful matching audit write. OTP failure creates no lease; ready/authorize failure invalidates active authority; every subsequent authority call denies; no retry/cache/compensating event occurs. A one-attempt undurable write is permitted only as sink-failure evidence, never as an allow. |
| Q18-05 | Send malformed, partial, oversized, unauthorized, wrong-UID, cancelled, and shutdown-raced requests; force ID-source failure. Check that malformed/unauthorized paths have no invented audit context. | `evidence-review`: independently inspect candidate-bound existing/adjusted IPC tests and DEV command evidence; run the focused negative test selection only if the evidence lacks a direct assertion. | Existing SO_PEERCRED, protocol-size, cancellation, shutdown, and generic/silent-denial behavior remains intact; no backend authority or audit context is created before valid admission; no fail-open path. Evidence must demonstrate negative detection, not merely line coverage. |
| Q18-06 | Re-run existing OTP replay/rate/expiry, lease close/expiry race, no-cache sudo-client, seed wiping/redaction, and full behavioral tests; inspect changed tests for deletion/weakening. | `evidence-review`: candidate-bound DEV evidence for named existing suites, followed by QA `make check`, `make task-check TASK=TASK-0018`, `go test -count=1 ./...`, and `go vet ./...`. | All prior regression behavior remains intact and test strength is not reduced. A reproducible behavioral failure is a regression or implementation failure by affected row; no omission can PASS by review. |
| Q18-07 | Build/package twice in clean temporary copies with a fixed checked-out commit epoch. Compare bytes/digests, inspect tar metadata, corrupt a checksum, and inject each forbidden source/metadata/secret/unexpected path. | `focused-rerun`: repository-native workflow/package test (or exact checked-in shell command included by `make check`) twice, then `sha256sum -c`. Hermetic staging and fixed epoch make this deterministic. | The sole archive has exactly: `SHA256SUMS`, three named binaries, and PAM/sudo/systemd declarative files; no directory/symlink/absolute/traversal/hidden member. Checksum covers exactly six payload paths. Go/source/mod/sum, `.git`, tasks/backlog/evidence, seeds/credentials/tokens/keys, and all extra paths are rejected before upload. |
| Q18-08 | Statically parse workflow and validate actual resolved pins and annotations against official GitHub Actions release provenance. Mutate a tag, branch, non-`actions/` owner, SHA, version comment, permission, build target, and artifact file list. | `focused-rerun`: repository-native workflow test plus independent QA inspection. | All executable actions are official `actions/*` immutable reviewed SHAs whose release/version provenance is recorded; build-provenance/attestation semantics are official and produce provenance for the archive digest. Permissions are exactly least-required `contents: read`, `attestations: write`, `id-token: write`, `artifact-metadata: write`; exactly Linux amd64 three binaries and archive/checksum artifact upload occur. Any unpinned/nonofficial/overprivileged action fails. |
| Q18-09 | Recount only canonical production SLOC in the four declared paths and inspect complete candidate diff/path ownership and secret scan. | `evidence-review`: independent count with the Task canonical definition, `git diff --check`, `git diff --name-only`, and bounded secret-pattern inspection that does not print sentinels. | Baseline 1407 + at most 90 readable executable production lines = at most 1497. No excluded scope, unowned path, source/log secret, deployment action, or v2 change. This is a requirement/process result, not waived by passing tests. |
| Q18-10 | On merged `main` only, dispatch/observe the intended workflow for the merge SHA, download the named artifact into an empty directory, recheck manifest/checksums, and verify provenance. Intentionally reject wrong repository/ref/digest/run evidence. | `live-e2e`: Main runs `gh attestation verify codex-authority-linux-amd64.tar.gz --repo autotaker/codex-authority-broker`, records redacted output plus subject digest, repository, workflow identity/run, and merge-SHA binding. | Artifact digest, exact manifest, checksums, repository, workflow/run, and `main` merge commit all bind together. A local artifact, unit test, different ref/digest/repository, unavailable Actions/auth, or unredacted evidence is blocked/environment evidence and cannot be replaced by another mode. |

## Execution and classification

QA first verifies candidate identity, diff scope, SLOC arithmetic, test
integrity, and Q18-01--Q18-05 deterministic evidence. It then independently
runs Q18-02 race checks, package/workflow checks (Q18-07--Q18-08), full
quality gates (Q18-06), and task/scope checks. `make check` is mandatory in
this gate. The postmerge Q18-10 remains pending/blocked until Main merges;
even `merge_tree == candidate_tree` cannot omit it.

Record active only while a named test is running, wait only for a bounded
observable prerequisite, and retry only after recording the original error,
classification, changed prerequisite, and resulting evidence. Do not rerun a
functional failure blindly. Each live command must redact credentials and
must not log GitHub auth diagnostics wholesale.

- **Implementation:** a controlled candidate test violates audit atomicity,
  expiry, sensitivity, packaging, workflow, or authority behavior.
- **Regression:** an established behavior or full-suite check fails because
  the candidate changed it; retain before/after/path evidence.
- **QA-plan:** this plan lacks, misassigns, or cannot detect a Task condition;
  amend the QA plan before an acceptance verdict and rerun affected cases.
- **Requirement/process:** impossible/conflicting contract, SLOC/scope stop,
  invalid pin/provenance policy, or required security property needing a
  replan/split.
- **Environment:** unavailable toolchain, Unix socket fixture, GitHub Actions,
  or `gh` access prevents valid execution without evidence of product failure.
  Live provenance remains blocked; it is never locally substituted.

## PLAN reconciliation (after TASK-first matrix)

`PLAN.md` was read after the matrix above. It aligns with the Task-first
baseline in these material respects: it fixes stderr as the non-configurable
operational sink; restricts audit fields to the five required values; creates
correlation only after valid peer/request admission; serializes the decision,
write, and success publication; makes OTP publication conditional on an audit
callback under lease locking; latches audit failure and invalidates authority;
uses the exact seven-member archive with six payload checksums; requires
immutable official action pins/least permissions; and reserves `main`-bound
GitHub attestation verification as live postmerge work.

The plan's explicit sink-failure clarification is compatible with the Task:
an available sink yields one durable event, while a failed sink yields one
attempt but no allow/new usable lease. QA must reject any implementation that
calls a failed write an emitted audit event, emits a compensating duplicate,
or lets a concurrent authorize use the OTP lease before its write succeeds.

Two execution qualifications are added by this independent plan. First, the
workflow test must prove the configured official attestation action creates
provenance for the archive digest—not merely accept a pinned action name.
Second, Q18-08's claimed action-release provenance must be independently
recorded from an official GitHub source during candidate evidence review;
static SHA syntax alone is insufficient. These are evidence-strength
requirements, not a semantic conflict with PLAN.

## DEV-readiness verdict

**QA plan approved for DEV gate, conditional on Main recording dependency
PASS, approval of this plan and PLAN, clean scope ownership, and GitHub
Actions/`gh` availability for the required postmerge check.** The candidate
cannot receive final QA PASS until Q18-01 through Q18-09 pass on its fixed
tree. The Task cannot close and the manual canary remains blocked until Main
also completes Q18-10 on merged `main` with redacted evidence.

Active: QA planning/reconciliation complete. Wait: Main approval, candidate
fixation, and later main-bound live verification. Retry: 0; no QA execution
has occurred.
