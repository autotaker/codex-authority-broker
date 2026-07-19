# REVIEW_RESULT — TASK-0019

## Decision: FAIL

Independent REVIEW fails candidate
`ef7bba6085bd0f63f62cf470951b9749c9803084` (tree
`0339747d7b9217090c1f85d15a8a93ca6665f6c6`).

## Blocking findings

1. `Runtime.Close` is not serialized by `auditMu`. `Handle` releases `mu`
   after computing `final`, then calls the potentially blocking audit writer.
   A concurrent shutdown can therefore complete while an allow write is
   blocked; after the writer succeeds, that request still publishes allow.
   This violates the required single linearization boundary for closure,
   audit success, and response publication. Existing tests cover closure in
   `beforePublish`, before the final-state check, but do not barrier-interleave
   `Close` with an in-progress writer. Serialize closure with the audit
   transaction without deadlocking the sink-failure close path, and add the
   missing deterministic barrier case.

2. `TestReleaseArchiveIsDeterministicExactAndSourceFree` does not execute the
   checked-in workflow/package path or build its real payloads. It packages
   synthetic `fixture:` bytes using a separately implemented tar pipeline,
   so it can pass if the workflow's build, staging, manifest, rejection, or
   checksum commands regress independently. The workflow test also checks
   fixed strings rather than the QA-plan's required negative mutations for
   ref/owner/pin/permission/target/payload/attestation policy. Add one
   repository-native packaging command shared with, or directly extracted
   from, the workflow and run that exact path twice with the required
   mutations.

## Bounded check evidence

- Exact candidate commit/tree and ten-path scope matched; `git diff --check`,
  `go vet ./...`, gofmt, and backlog JSON validation passed.
- Socket-capable focused race suites passed outside the restricted sandbox.
  Main independently reported full, focused-race, and vet PASS on the same
  candidate/tree. The reviewer's full run hit an environment-specific nested
  CLI-build failure and is not classified as a candidate failure.
- Official tag refs match all four pinned action SHAs and version annotations.
- Canonical production deltas are runtime `+57`, IPC `+8`, lease `+6`: total
  `+71`, cumulative `1478`, within the approved per-file and cumulative caps.
- No Makefile exists; unavailable `make check`/`make task-check` targets are a
  repository tooling limitation, not an invented product PASS.

---

## Re-review decision: PASS

Independent re-review passes fixed candidate
`ef8b346bded352ee0e4714cb72563544b9c392e9` (tree
`687eb8125d51875b74ad3667fe43e1d3c34c1cfc`). BLOCKING: none.

Both prior blockers are resolved. Public `Close` now serializes on `auditMu`,
while sink failure, overflow, and missing-context paths use the private
`closeUnderAudit` path and cannot recursively deadlock. The deterministic
barrier test proves that close cannot complete during a blocked audit write
and completes after successful publication. The release test now extracts
and executes the exact checked-in YAML `run` block twice against real broker,
CLI, and sudo-helper binaries; it validates deterministic real archives and
rejects workflow policy, forbidden-member, and checksum mutations.

Reviewer reruns passed the backend race suite, the focused real release
workflow/archive suite, `go vet ./...`, gofmt, backlog JSON, scope inspection,
and the complete product/test/workflow/declarative diff. Main independently
reported socket-capable focused race and full-suite PASS on the same fixed
commit/tree. Canonical production deltas remain runtime `+57`, IPC `+8`, and
lease `+6`: total `+71`, cumulative `1478`, within every approved cap. The
action pins, exact permissions, PAM `seteuid`, systemd restrictions,
five-field audit/redaction boundary, actor context, immutable deadline, and
source-free seven-member/six-checksum archive remain intact. The absent Make
targets remain a classified repository tooling limitation.
