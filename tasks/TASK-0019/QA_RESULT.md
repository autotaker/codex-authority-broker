# TASK-0019 independent QA result — fixed candidate

## Candidate and verdict

- `candidate_commit`: `ef8b346bded352ee0e4714cb72563544b9c392e9`
- `candidate_tree`: `687eb8125d51875b74ad3667fe43e1d3c34c1cfc`
- QA plan: approved TASK-first `tasks/TASK-0019/QA_PLAN.md`, unchanged
- REVIEW was not a QA precondition and no REVIEW PASS was used as evidence.

**Premerge verdict: FAIL (candidate hygiene), not attributed to DEV product
behavior.** Q19-01 through Q19-04 product/security evidence passes and Q19-05
scope/SLOC passes, but mandatory `git diff --check` fails on the fixed
candidate: `tasks/TASK-0019/REVIEW_RESULT.md:46: new blank line at EOF.` Main
must classify and fix or replace the candidate before merge. Q19-06 live
GitHub provenance is **PENDING POSTMERGE by design** and was not run or
substituted locally.

## Case results

| Case | Mode | Result and candidate-bound evidence |
| --- | --- | --- |
| Q19-01 exact audit schema/redaction/expiry | `focused-rerun` | **PASS.** `go test -count=1 ./internal/backend ./internal/ipc ./internal/lease ./cmd/codex-authority-broker` passed in the socket-capable context (exit 0). Candidate tests strictly decode five fields, numeric actor, fresh nonzero lowercase-hex sequence IDs, scope/result/null rules, common immutable UTC OTP/authorize expiry, 256-byte bound, and sensitive sentinels. |
| Q19-02 concurrency/linearization/fail-closed | `focused-rerun` | **PASS.** `go test -count=1 -race ./internal/backend ./internal/ipc ./internal/lease ./cmd/codex-authority-broker` passed (exit 0). Focused barrier cases cover OTP invisibility before audit publication, `Close` waiting for publication, distinct tuples/IDs, error and short writers, no retry, permanent closure for new/prior authority, missing actor, cancellation/closure deny, and sequence overflow without a write. |
| Q19-03 actor/deadline/regressions | `focused-rerun` | **PASS with recorded QA-environment limitation.** The same focused socket-capable command passed backend/IPC/lease/broker; `go vet ./...` passed (exit 0). Main supplied fixed-candidate socket-capable full-suite PASS. QA's broad `go test -count=1 ./...` attempt reached all packages but its nested CLI self-build failed under this execution environment; this is `environment_issue`, not candidate-failure evidence, and did not replace the successful socket-capable full result. Actor context is added only after peer/request admission; `State.Deadline` locks, expires, returns UTC state, and introduces no invalidation/admission API. |
| Q19-04 exact source-free artifact/workflow | `focused-rerun` | **PASS.** `go test -count=1 ./cmd/codex-authority-broker -run 'TestReleaseWorkflowPinsPermissionsAndPayload\|TestReleaseArchiveIsDeterministicExactAndSourceFree' -v` passed (exit 0, 16.84 s). The checked-in workflow block performed two real clean builds, produced byte-identical archives, checked the exact seven-member/six-checksum manifest, and rejected source-member and corrupted-payload archives. Static mutations rejected wrong ref/owner/pin/permission/build/payload/attestation subject. PAM requires `pam_exec.so quiet seteuid`; systemd assertions cover fixed broker path, root user/group, `NoNewPrivileges`, strict protection, and `/run`. Official tag refs independently resolved exactly to all four pinned SHAs. |
| Q19-05 scope/cap/hygiene | `evidence-review` | **FAIL only on mandatory diff hygiene.** Canonical candidate production count is 1478: baseline 1407 + runtime 57 + IPC 8 + lease 6 = +71, within each per-file maximum and cumulative <=1497. Only the three declared production paths changed; no extra product path was used. `make check` and `make task-check TASK=TASK-0019` were attempted and both returned exit 2 because the repository has no such targets; classify as inherited `environment_issue`, not a product PASS or DEV defect. `git diff --check` returned exit 2 for the trailing blank line noted above, which blocks this fixed candidate. |
| Q19-06 merged-main provenance | `live-e2e` | **PENDING POSTMERGE.** Main must run/observe the merged-main workflow, download into an empty directory, repeat exact manifest/checksum verification, and run `gh attestation verify codex-authority-linux-amd64.tar.gz --repo autotaker/codex-authority-broker`, recording redacted repository, merge SHA, workflow/run, and subject digest binding. No local result substitutes. |

## Exact commands and failure classification

- Initial restricted-sandbox focused/race attempts failed only where Unix
  socket creation returned `operation not permitted`; the required commands
  were rerun in a socket-capable context and passed. Classification:
  `environment_issue`, prerequisite changed, retry count 1.
- Focused package/workflow real-build test: exit 0; no cache result was treated
  as acceptance without executing the checked-in block.
- `go vet ./...`: exit 0.
- `git ls-remote` against the official `actions/checkout`, `actions/setup-go`,
  `actions/attest`, and `actions/upload-artifact` repositories: exit 0 and
  exact tag/SHA matches for v7.0.0, v7.0.0, v4.2.0, and v7.0.1.
- `git diff --check 0dbeec7..ef8b346bded352ee0e4714cb72563544b9c392e9`:
  exit 2, candidate evidence formatting defect. This is the sole candidate
  blocker found; it is not evidence of an audit/runtime implementation defect.

No product, test, workflow, deployment, Git, GitHub, or live-host state was
changed by QA. Only this QA result was written.

## Re-QA — final evidence candidate

- Final evidence candidate: `dde3501ecebad314a47a0e6cc692bcfd9bc00b12`
  (`fc2f22bb86f687d2393d5d01ef3f47a4e81319ec`).
- `git diff --check 0dbeec721ffffbb757265efc69d1e791b33f7611..dde3501`:
  **PASS**, exit 0.
- Product, workflow, deployment, and test paths are byte-identical to the
  already-tested `ef8b346bded352ee0e4714cb72563544b9c392e9`; only
  `QA_RESULT.md` was added and `REVIEW_RESULT.md` evidence was appended.
- Canonical production remains **1478**: baseline 1407 + runtime 57 + IPC 8 +
  lease 6 = +71. No security suite was rerun because the product tree did not
  change and the affected QA case set was only Q19-05 diff hygiene.

**Re-QA verdict: PASS for the premerge evidence candidate.** The prior FAIL is
preserved above as history of the intermediate candidate. Q19-06 remains
**PENDING POSTMERGE** and cannot be carried forward or replaced by this
evidence-only re-QA.
