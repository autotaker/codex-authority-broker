# QA PLAN — TASK-0001: Core lease state, readiness, and absolute expiry

## Scope, independence, and prerequisites

This is the independent, pre-merge QA gate for the approved TASK-0001 and
PLAN.  It tests only the in-process Go lease state API in the approved paths:
`go.mod`, `go.sum`, `internal/lease/lease.go`, and
`internal/lease/lease_test.go`.  It neither broadens the public API nor
authorizes implementation, staging, committing, merging, packaging, or
release work.

QA starts only after DEV has supplied its candidate and evidence, and
independent REVIEW has recorded PASS.  QA reads the TASK, approved PLAN,
candidate diff, and REVIEW evidence before testing.  A missing REVIEW PASS is
a gate failure: do not run this as a substitute for review and do not permit a
merge.

The following are explicitly out of scope and must not be required or added
as test fixtures: TOTP/OTP verification or replay handling; IPC, sockets,
daemon/service behavior; PAM or sudo; push, GitHub, packaging, deployment, or
release behavior.  QA must reject a candidate that adds such scope, but must
not treat its absence as a failure.

Use a deterministic injectable fake `Clock` for time-dependent checks.  No
real-time sleep or wall-clock waiting is acceptable evidence.  A fresh
`State` instance is the restart fixture; no persistence or recovery fixture is
permitted.

## Acceptance matrix

| ID | Setup and operation | Required observation and evidence |
| --- | --- | --- |
| QA-01 | From a new `State`, call `BeginReadiness` at fake time `t`. | An opaque current challenge is created with deadline exactly `t + 300s`; no lease is active. Capture the focused test name/output. |
| QA-02 | From a new `State`, attempt `Activate` with absent, fabricated, or otherwise non-current challenge before readiness. | It returns a defined error/denial, remains inactive, and exposes no authority. In particular, the core API has no OTP-like input or alternative activation path that can create a lease before readiness. |
| QA-03 | Open a challenge at `t`; attempt activation just before deadline and exactly at/after `t + 300s`. | Activation before expiry may transition once; activation at and after the absolute deadline denies and cannot extend or replace the challenge deadline. |
| QA-04 | Activate a current, unexpired challenge at fake time `a`; observe `Active` at `a + 300s - 1ns`, `a + 300s`, and later. | Lease is active strictly before the immutable deadline, and inactive exactly at and after it. Readiness or activation during a live lease cannot renew, replace, or extend it. |
| QA-05 | Create a live lease, then construct a fresh `State` with the same fake clock. | The fresh state is idle/inactive and has no recovery path. Expired challenge and lease observations also fail closed. |
| QA-06 | Inspect transition tests and implementation claims for mutex/state behavior. | Exercise concurrency or additional state-transition races only when they are exposed by the approved core API/contract. If exercised, evidence must establish no more than one valid activation and no deadline extension. Do not invent later TASK-0003 consume/replay semantics. |

All six rows are P0 for this task.  A skipped row is FAIL unless the main
Agent records an approved QA-plan amendment before a merge decision.

## Required execution and evidence

After DEV, from the candidate repository root, QA records the command,
exit status, Go version, and a redacted concise result for:

```sh
go test ./...
test -z "$(gofmt -l .)"
go vet ./...
make check
```

Run `go vet ./...` when the candidate is a valid Go module and the installed
toolchain supports it; if the module/toolchain does not support that command,
record the exact reason and classify it before deciding PASS/FAIL.  The
`gofmt` check covers Go source only; generated and third-party material must
not be introduced by this task.  `make check` is required when available in
the candidate repository.  A missing or inapplicable command is evidence to
classify, never a reason to silently skip an acceptance row.

Independently repeat the focused fake-clock tests that cover QA-01 through
QA-05.  QA-06 is conditional as described in the matrix.  Review the diff
against the approved path boundary and confirm it has no OTP/TOTP parsing,
storage, validation, logging, or transport, and no IPC/sudo/push/release
surface.

## Exact cumulative production-SLOC and readability gate

QA independently counts every nonblank, non-comment executable-source line
in candidate Go production files, excluding `*_test.go`, generated files,
vendor, configuration, and task documents.  A line with executable code and
an inline comment counts once.  Run this exact command at the repository root
and preserve its per-file output and `TOTAL`:

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

`TOTAL <= 250` is mandatory.  QA also manually checks for prohibited source
compression: semicolon or one-line packing, collapsed error handling, cryptic
names, removed security comments, or functions combined only to meet the
ceiling.  Normal `gofmt`-formatted, idiomatic structure is required.  A
candidate over the ceiling, or one that conceals complexity to appear under
it, is FAIL; do not compress it during QA.

## Failure classification and merge decision

For each failure, preserve the matrix ID, command, candidate revision,
redacted excerpt, fixture time, and reproduction conditions.  Classify before
attribution:

| Classification | Use when | Disposition |
| --- | --- | --- |
| `implementation_defect` | Candidate violates an approved acceptance row, scope boundary, deadline, restart-deny rule, SLOC ceiling, or formatting/readability guardrail. | FAIL; return to DEV with a minimal reproduction. |
| `regression` | An approved behavior that passed on the relevant baseline now fails. | FAIL; return to DEV with comparison evidence. |
| `qa_plan_defect` | QA procedure/expectation is invalid or contradicts the approved TASK/PLAN while the candidate satisfies them. | Pause; amend and re-approve this QA plan. |
| `requirement_gap` | TASK/PLAN lacks a decision needed to judge a core observation. | Escalate to the main Agent/PLAN gate; do not infer a DEV defect. |
| `environment_issue` | Toolchain, module support, fixture, or test infrastructure prevents assessment and is demonstrably not candidate behavior. | Record blocker and rerun after correction; do not assign DEV fault. |

QA PASS requires all applicable matrix rows, required checks, focused evidence,
scope review, exact SLOC evidence at or below 250, and no-compression review
to pass.  Independent REVIEW PASS and independent QA PASS are both mandatory
preconditions for the main Agent's merge decision.  Any FAIL returns to its
responsible gate and never authorizes a merge.
