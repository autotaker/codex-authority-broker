# TASK-0019 QA plan — TASK-first audit, artifact, and provenance gate

## Independence, authority, and preconditions

This QA plan was derived from `TASK-0019/TASK.md` and
`TASK-0018/FAILURE.md` before reading `PLAN.md`.  QA owns only the
candidate-bound verification evidence and failure classification.  It must
not change product, tests, workflow, deployment inputs, Task/plan contracts,
Git state, GitHub state, or a host.

Main must record TASK-0012's merged/PASS dependency, approve this plan and
`PLAN.md`, and have DEV fix one `candidate_commit` and `candidate_tree`.
QA and REVIEW then begin independently and concurrently from that exact tree;
neither role's PASS is a precondition for the other.  Every execution record
must include case ID, candidate commit/tree, command, fixture/environment and
cache condition, exit, bounded artifact/result digest, negative-detection
claim, and an unexecuted reason where applicable.  Candidate/tree mismatch,
unknown impact, omitted negative control, or weakened test is FAIL, never an
evidence-review PASS.

No test fixture, terminal output, manifest, attestation record, or QA evidence
may contain request payload, OTP, seed, token, key, credential, environment
value, lease identifier, or internal-error text.  The allowed candidate
surface is the three declared production paths; named tests; release workflow;
and declarative PAM/systemd inputs, while reusing the existing sudoers input.
No installer/deployment, release publication, host PAM/sudo mutation, manual
canary, generic/remote logging, GitHub credential custody, or v2 work is in
scope.

Stop before candidate acceptance and classify `requirement_gap` if readable
production source exceeds +90/cumulative 1497, needs a fourth production path
or compression, weakens a required field/redaction/fail-closed rule, permits
sensitive audit input, cannot prove the exact archive, or cannot support
postmerge provenance verification.  The three lines below 1500 and the 1800
guard are not implementation allowance.  TASK-0018's +102 uncommitted draft
is not candidate evidence: this Task may shed only its random fixed-width ID
format and separate lease-internal invalidation.

## Fixed acceptance oracle

For each *admitted backend operation*, exactly one successful audit write is a
single JSON line with exactly these five keys and no others:
`correlation_id`, numeric `actor_uid`, `scope`, `result`, and `lease_expiry`.
`correlation_id` is a fresh, nonzero, lowercase hexadecimal representation of
a process-local sequence: it is neither random, fixed-width, nor globally
unique.  `scope` is `ready`, `otp`, or `authorize`; `result` is `allow` or
`deny`.  Ready and every denial have JSON `null` expiry.  Allowed OTP and
authorize have the same immutable UTC RFC3339Nano lease deadline.  Payload,
OTP, seed, token, key, environment, lease identifier, error, and error text
must never enter the event.

One Runtime mutex must linearize audit-context validation, handler decision,
deadline capture, one bounded sink write, and response publication.  A
successful write is the final decision linearization point.  Cancellation,
shutdown, closed state, malformed context/encoder/bound, sequence overflow,
writer error, or short write denies.  For an attempted failing write there is
no durable allow event; it must have no retry or compensating event.  Sink
failure permanently closes Runtime while this linearization is held, so both a
just-created OTP lease and all pre-existing leases are unusable thereafter;
there is no separate lease invalidation.

## Acceptance matrix

| ID | Acceptance, mutations, and independent procedure | `qa_execution_mode` and rationale | PASS / fail-closed result |
| --- | --- | --- | --- |
| Q19-01 | Use deterministic clock, peer UID, sink, and sequence fixtures for allowed and denied ready/OTP/authorize. Strictly decode each line; mutate every key name/count/type, scope/result enum, `null`/deadline value, identifier zero/case/non-hex/reuse, and inject unique sensitive sentinels through all request paths. | `focused-rerun` — bounded backend and Unix-socket fixtures in `internal/backend/runtime_test.go` and `internal/ipc/server_linux_test.go` can deterministically reproduce the audit boundary. | Each admitted operation has one successful bounded five-field event with the authenticated numeric peer UID and correct tuple. IDs are fresh, nonzero lowercase hex process-local sequence values; they need not be fixed-width/random. Any missing/duplicate/extra field, bad tuple, deadline mutation, leak, or fabricated success fails as `implementation_defect`. |
| Q19-02 | Barrier-interleave accepted/denied calls, OTP activation, authorize, cancellation, shutdown, and close. Force sink error and short write before/after an existing lease; observe response, write count, subsequent authority calls, and race detector. Verify no authorize observes the newly created OTP lease until its allow write completes. | `focused-rerun` — deterministic barriers and bounded failing writers fully exercise concurrency/fail-closed truth locally; run `go test -count=1 -race ./internal/backend ./internal/ipc ./internal/lease ./cmd/codex-authority-broker`. | No tuple crossing or ID reuse; cancellation before write is denied/audited; sink error/short write yields no allow, exactly one attempted write, no retry/compensation, and permanent Runtime closure for current/future authority, including new/prior OTP leases. Any later usable authority is `implementation_defect`; a non-hermetic race fixture is `qa_plan_defect` until corrected. |
| Q19-03 | Exercise active/expired/replayed/rate-limited OTP, immutable deadline accessor, malformed/partial/oversized/wrong-UID/no-context requests, no-cache client, seed wiping/redaction, lifecycle, and existing cancellation behavior. Inspect changed tests for deletion or assertion weakening. | `focused-rerun` — these established package tests are deterministic and bounded: targeted backend/IPC/lease/main suites, followed by `go test -count=1 ./...` and `go vet ./...`. | Expiry cannot extend/revive; ready/deny expiry is null; no lease callback/invalidation API is introduced; pre-admission requests create neither authority nor audit context; old security/lifecycle behavior remains. Candidate-caused prior behavior failure is `regression`; missing required assertion is `implementation_defect` or `qa_plan_defect` according to evidence. |
| Q19-04 | In clean temporary copies at a fixed checked-out commit epoch, execute the exact checked-in package path twice; compare archive digests/metadata, extract into empty directory, verify `sha256sum -c`, and mutate checksum plus every forbidden member class. Parse workflow and mutate tags/branches, owners, SHA/version annotations, permissions, target, payload list, and attestation subject. Independently record official release provenance for each exact full SHA. | `focused-rerun` — a local deterministic package/workflow test with fixed epoch and hermetic staging reproduces archive and static-policy truth without GitHub side effects. | Archive has exactly seven members: `SHA256SUMS`, three specified binaries, and PAM/sudo/systemd declarative files; sums cover exactly six payloads. It rejects source/module/.git/tasks/evidence/secret-named/extra, directory, symlink, absolute, traversal, and hidden members. Exactly Linux amd64 builds, least permissions, official `actions/*` full-SHA pins with recorded official provenance, attestation of archive digest, and archive/checksum upload are required. Any discrepancy is `implementation_defect`; unavailable local tooling is `environment_issue`, not a substituted PASS. |
| Q19-05 | Independently inspect complete candidate diff, ownership, executable nonblank/noncomment production count, test changes, secret-pattern scan without printing matches, `git diff --check`, and repository gates. Attempt `make check` and `make task-check TASK=TASK-0019` in addition to the focused/race/full/vet checks; record absent targets verbatim as tooling evidence. | `evidence-review` — scope/cap and candidate-bound command/test-integrity evidence are independently auditable; product high-risk cases are covered by focused reruns above. | Only `runtime.go` (+57 max), `server_linux.go` (+14 max), and `lease.go` (+19 max) are production changes; baseline 1407 + at most 90 is <=1497 and no sensitive/log/excluded/v2 scope exists. A missing required repository target is `environment_issue` and blocks the gate until Main resolves its contract; a cap/scope/security breach is `requirement_gap`, not presumed DEV fault. |
| Q19-06 | After Main merges, observe/dispatch the main-bound workflow for the merge SHA; download its named artifact into an empty directory; independently recheck seven-member manifest/six checksums and bind archive digest, repository, workflow/run, and merge SHA; run `gh attestation verify codex-authority-linux-amd64.tar.gz --repo autotaker/codex-authority-broker`. Reject wrong repository/ref/digest/run. | `live-e2e` — GitHub Actions provenance and `gh` attestation depend on live repository/network state and cannot be truthfully reproduced locally. | Main records redacted binding evidence for `autotaker/codex-authority-broker`, merged `main` SHA, workflow/run, subject digest, manifest/checksum, and successful verification. No local archive, another ref/repository/digest, or unit result substitutes. Missing safe access/cleanup is `environment_issue` and remains blocked; it cannot PASS via another mode. |

## Candidate execution, evidence, and classification

QA independently verifies candidate identity/diff/test integrity first, then
runs Q19-01 through Q19-04 and Q19-03's full/vet commands. Q19-02's race run,
Q19-04 package/workflow run, and Q19-05 scope evidence are retained as
separate case records. `git diff --check` is mandatory. `make check` and
`make task-check TASK=TASK-0019` are required attempts under the repository
contract; if this repository still lacks those targets, QA records the exact
tooling absence and Main must classify/resolve it before a final gate verdict,
rather than claiming a product failure or silently omitting them.

Failure classification is evidence-led and Main owns the final label:

- `implementation_defect`: a candidate violates the fixed audit, authority,
  archive, workflow, or checksum contract.
- `regression`: an established security/lifecycle/IPC/lease behavior fails due
  to the candidate.
- `qa_plan_defect`: a planned procedure/fixture cannot prove the stated
  condition or has a wrong execution mode.
- `requirement_gap`: Task/PLAN conflict, impossible safety condition, SLOC or
  permitted-path breach, or unprovable official/provenance policy.
- `environment_issue`: unavailable socket/toolchain/package tooling, GitHub
  Actions, or `gh` access prevents execution without candidate-failure proof.

Any candidate change after QA begins affects audit/auth/IPC/lease/workflow/test
surface and therefore prohibits carry-forward; Main must choose affected or
full reruns under CF-1 through CF-7.  After merge, Q19-06 always remains a
case-level live confirmation even if `merge_tree == candidate_tree`.

## PLAN reconciliation and verdict

After fixing the TASK-first matrix above, QA read `PLAN.md`.  It aligns on the
three production paths and +57/+14/+19 cap, private post-admission
SO_PEERCRED UID, fixed stderr sink, fresh nonzero lowercase-hex process-local
sequence, exactly five audit fields, UTC immutable deadline, single Runtime
linearization, permanent Runtime closure without lease-internal invalidation,
exact seven-member/six-checksum archive, official pinned actions, and
main-bound `gh` verification.

QA adds two binding evidence clarifications rather than changing acceptance:
the local workflow test must demonstrate archive-digest attestation semantics,
not merely recognize an action name; and candidate evidence must map every
pinned action SHA/version annotation to an official release source.  The
plan's local package test must be an actual repository-native deterministic
test/command, not a workflow-only assertion.  These requirements are
consistent with the Task's source-free and provenance acceptance conditions.

**Verdict: APPROVED for the DEV gate, conditional on Main recording the
dependency PASS, both plan approvals, a fixed candidate/tree, and eventual
GitHub Actions/`gh` availability for Q19-06.**  This is not a candidate QA
PASS.  Final QA PASS requires Q19-01 through Q19-05 on one candidate and
postmerge Q19-06; otherwise the Task and manual-canary milestone remain
blocked.
