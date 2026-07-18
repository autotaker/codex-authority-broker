# QA PLAN — TASK-0003: TOTP replay, rate limits, and concurrency

## Scope, independence, and prerequisites

This is the independent pre-merge QA gate for TASK-0003 and its approved
PLAN. It assesses only the in-process TOTP verifier and its atomic use with
the merged TASK-0001 lease state. QA owns no implementation, review, staging,
commit, merge, release, or product-interface work.

QA starts only when DEV has supplied its candidate and redacted evidence and
an independent reviewer has recorded `PASS` in
`tasks/TASK-0003/REVIEW_RESULT.md`. QA reads the TASK, approved PLAN,
candidate diff, and that review result before execution. Missing, non-PASS,
or scope-incomplete review evidence is a gate failure: QA must not substitute
for REVIEW and must not authorize a merge.

Approved implementation/test paths are `internal/lease/lease.go`,
`internal/lease/totp.go`, `internal/lease/lease_test.go`, and
`internal/lease/totp_test.go`. QA may reject an out-of-bound candidate but
must not add code or fixtures. IPC, sockets, services, commands/CLI arguments,
configuration or environment secret input, files, persistence, databases,
keyrings, PAM/sudo, audit/logging, packaging, and release work are excluded.
No external TOTP module is permitted.

All time evidence uses the existing injectable fake clock; real-time waits
are not evidence. Test secrets are documented synthetic byte fixtures only.
Do not put a secret, submitted OTP, calculated OTP, counter, or time-derived
candidate in a failing assertion, command line, log, or captured test output.
A fresh `State` plus reinjection of the same documented synthetic secret is
the restart fixture. It must use only the new state's private boot replay
floor; persisted or restored lease, accepted-watermark, or secret state would
expand scope.

## Independent acceptance matrix

| ID | Setup and operation | Required observation and evidence |
| --- | --- | --- |
| QA-01 | Use RFC 6238 SHA-1 vectors at their specified timestamps to validate the calculation; use the task's six-digit profile for activation fixtures. | The implementation uses HMAC-SHA-1, `T0=0`, 30-second Unix steps, dynamic truncation, and six decimal digits. The focused deterministic vector test passes without exposing a real or synthetic secret/OTP in output. |
| QA-02 | At a fake-clock current step, open readiness and call the sole OTP activation API with a valid, unused six-digit current-step code. | It succeeds exactly once, returns one lease, consumes the challenge, and makes the lease active. No alternate OTP activation path exists. |
| QA-03 | Repeat QA-02 with codes for the immediately preceding and following 30-second counters. Then use codes beyond each side of that `-1, 0, +1` window and a negative Unix counter case. | Each adjacent valid counter is accepted when otherwise unused; older/newer and negative-counter cases deny and create no authority. |
| QA-04 | Submit malformed input (including wrong width and non-digits), an invalid six-digit value, stale values, and a valid value previously accepted after a new readiness challenge. | Every case denies, remains inactive, and does not create a lease. Input is rejected as six digits before a successful match; replay records deny a matching valid counter at or below the highest accepted counter, including after a fresh challenge. |
| QA-05 | For one live challenge, make five submissions within its fixed 60-second window, including denied inputs where practical; attempt a sixth before the boundary. Advance the fake clock to exactly the first-submission time plus 60 seconds, then submit again. Also open a new challenge after expiry/consumption. | Every live-challenge attempt counts, the sixth and later pre-boundary attempts return the rate-limit denial without a successful verification, exactly `+60s` begins a new window, and new readiness resets its rate state. Consumption and expiry clear attempt state. |
| QA-06 | Start multiple goroutines behind a barrier, all submitting the same valid code to the same current challenge. | Exactly one call succeeds, exactly one lease becomes active, and every other call denies. Race-detector output is clean. |
| QA-07 | Activate through a valid OTP at fake time `a`; inspect activity just before `a+300s`, exactly at it, and later. Attempt readiness/activation while live. | TASK-0001's separate, absolute 300-second lease deadline is unchanged: active strictly before its deadline and inactive at/after it; no readiness or activation renews, replaces, or extends it. |
| QA-08 | Create a used-counter/replay condition and a live lease, then construct a fresh `State` in the same fake-clock 30-second step, reinject the same synthetic secret, and open new readiness. Try the same-step and older counters, then advance exactly to the next 30-second counter and try its code. | The fresh state is inactive and restores neither lease, accepted watermark, nor secret. Its boot floor records the current non-negative counter and denies the same and all older counters after reinjection; the strictly next counter may activate when otherwise valid and unused. A normal-clock restart can therefore impose up to about 30 seconds of startup wait. This fail-closed delay is explicit and no replay, lease, or secret state is persisted. |
| QA-09 | Review returned errors, package source, focused test output, and candidate diff. | Errors have fixed sentinel text/identity and disclose neither secret, submitted/expected OTP, counter, nor time. There is no production log/format/String path for the verifier and tests do not print secret/OTP fixtures. |
| QA-10 | Review the candidate diff and public surface against approved paths and exclusions. | No IPC or CLI/command, transport, configuration/environment, persistence/recovery, external TOTP dependency, or other later-boundary work exists. Existing `Activate` remains only the TASK-0001 primitive; `VerifyAndActivate` is the only OTP activation path and the verify/consume transition is mutex-atomic. |

All rows are P0. A skipped row is `FAIL` unless the main Agent records an
approved amendment to both PLAN and this QA plan before a merge decision.

## Required execution and evidence

After REVIEW PASS, run from the candidate repository root and record the
candidate revision, Go version, complete command/exit status, and concise
redacted result for:

```sh
go test ./...
test -z "$(gofmt -l .)"
go vet ./...
go test -race ./...
git diff --check
```

Independently run the focused fake-clock tests covering QA-01 through QA-08,
including the barrier-based duplicate-submission test under `-race`. Use
test names discovered in the candidate rather than inventing an API or
printing fixture values. A missing, unsupported, or failing command is
evidence to classify, never a reason to silently skip an acceptance row.

Perform a source and diff review for QA-09 and QA-10. Check that every state
transition and boot-floor/replay/rate access is protected by the same
`State.mu`; there is no timer, goroutine, or persistence mechanism. Confirm
that `New` records the current non-negative counter as the private boot floor
(or conservatively zero for negative Unix time), and that only a strictly
newer counter can pass that floor. Confirm the only
cryptographic primitives are Go standard-library primitives specified by the
PLAN, and that the verifier copies a nonempty caller-supplied secret into an
unexported field. Do not attempt to prove runtime zeroization beyond the
PLAN's stated non-promise.

## Cumulative production-SLOC and readability gate

QA independently runs this exact command at the candidate root and preserves
its per-file output and `TOTAL`. It counts all candidate production Go files,
including `internal/lease/lease.go` and `internal/lease/totp.go`, excluding
tests, generated/vendor/configuration files, and task documents:

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

`TOTAL <= 430` is mandatory. QA also manually checks for semicolon/one-line
packing, collapsed error handling, cryptic names, removed security comments,
or functions combined solely to meet the ceiling. Normal gofmt-formatted,
idiomatic structure is required. A candidate over the ceiling, or one that
conceals complexity to appear under it, is `FAIL`; QA must not compress it.

## Failure classification and merge decision

For every failure or blocker, preserve the matrix ID (when applicable),
candidate revision, command, fixture time/reproduction conditions, and a
redacted excerpt. Classify before assigning responsibility:

| Classification | Use when | Disposition |
| --- | --- | --- |
| `implementation_defect` | The candidate violates an approved matrix row, atomicity, fixed rate/replay boundary, boot-floor restart behavior, absolute lease behavior, secret boundary, scope, SLOC ceiling, formatting, or readability requirement. | `FAIL`; return to DEV with a minimal redacted reproduction. |
| `regression` | An approved TASK-0001 behavior or another relevant baseline behavior passes before the candidate but fails with it. | `FAIL`; return to DEV with comparison evidence. |
| `qa_plan_defect` | This procedure/expectation contradicts the approved TASK or PLAN while the candidate satisfies them. | Pause; amend and re-approve the QA plan. |
| `requirement_gap` | The TASK/PLAN omits a decision needed to judge an observed behavior. | Escalate to the main Agent/PLAN gate; do not infer a DEV defect. |
| `environment_issue` | Toolchain, module/cache, race runtime, fixture, or test infrastructure blocks assessment and evidence shows it is not candidate behavior. | Record the blocker and rerun after correction; do not assign DEV fault. |

QA `PASS` requires REVIEW `PASS`, all applicable P0 rows, all required
commands, focused and race evidence, `git diff --check`, scope/secret review,
and the exact SLOC/readability gate to pass. Independent REVIEW PASS and QA
PASS are both mandatory for the main Agent's merge decision. Any `FAIL`
returns to its responsible gate and never authorizes a merge.
