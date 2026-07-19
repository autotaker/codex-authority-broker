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

