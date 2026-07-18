# REVIEW RESULT — TASK-0003

## Decision: FAIL

The candidate passes its checked-in tests and static checks, but it does not
enforce the approved fail-closed boot replay floor for the entire startup
time step.  This is an `implementation_defect` affecting TASK-0003 acceptance
and QA-08.  Return the candidate to DEV; it is not ready for QA or merge.

## Finding

### Major — the adjacent future code bypasses the startup-step boot floor

`New` records the current TOTP counter in `bootReplayFloor`, but
`VerifyAndActivate` compares that floor only with the counter whose code
matched.  During the same clock step, the verifier considers `current + 1`
because the normal drift window is `-1, 0, +1`.  That future candidate is
greater than the boot floor and is therefore accepted immediately.

This contradicts the approved PLAN requirement that the boot floor reject the
entire step in which the process starts and that a normal clock reach a
strictly newer 30-second step before activation is possible.  The same logic
also weakens the negative-time rule: after a negative-time boot records the
conservative floor zero, a future-adjacent counter can pass while the current
clock counter is still zero, although verification must remain denied until
the current non-negative counter is strictly above zero.

Reviewer-only deterministic reproduction (supplied through `go test
-overlay`; no repository product/test file was added):

1. Start a fresh `State` at non-negative counter `C` and open readiness.
2. Without advancing the fake clock, submit the valid synthetic-fixture code
   for adjacent counter `C+1`.
3. Required: deny and remain inactive until the clock itself advances beyond
   the boot step. Observed: activation succeeds.

```text
GOCACHE=/tmp/codex-authority-broker-task3-review-cache \
  go test -vet=off -overlay=/tmp/task3-reviewer-overlay.json \
  ./internal/lease \
  -run TestReviewerBootFloorRejectsFutureAdjacentCodeDuringBootStep -count=1
--- FAIL: TestReviewerBootFloorRejectsFutureAdjacentCodeDuringBootStep
    future adjacent code activated before the clock advanced beyond the boot step
FAIL
```

Action: before accepting any drift-window candidate, require the current clock
counter itself to be strictly greater than `bootReplayFloor` (including
strictly greater than zero after a negative-time boot).  Add focused tests for
the `+1` code during the startup step and for counter zero after negative-time
startup; both must deny and leave no active lease.  Preserve normal `-1, 0,
+1` drift after the boot floor has been crossed.

## Positive evidence

- RFC 6238 calculation uses standard-library HMAC-SHA-1, `T0=0`, 30-second
  Unix steps, dynamic truncation, and six decimal digits; comparison uses
  `hmac.Equal` after exact six-ASCII-digit input validation.
- The verifier copies a nonempty caller secret into an unexported field.  The
  package has no log, persistence, transport, environment/configuration, IPC,
  CLI, file, database, PAM/sudo, or external-module path.  Returned errors are
  fixed sentinels and contain no secret, submitted/expected code, counter, or
  time.
- Accepted-watermark replay checks, the five-attempt limit, exact 60-second
  window reset, challenge reset, and activation are all under the existing
  `State.mu`.  The barrier-based concurrent test and race detector show one
  winner and no race.  No timer or goroutine is introduced in production.
- The checked-in TASK-0001 tests are unchanged and pass.  The existing
  readiness/challenge and separate absolute 300-second lease behavior remains
  covered, including denial exactly at expiry and restart inactivity.
- Product changes are confined to approved `internal/lease` paths.  Manual
  review found ordinary gofmt-formatted, idiomatic structure with no
  semicolon/one-line packing, collapsed error handling, cryptic naming,
  removed security comments, or functions combined solely to reduce SLOC.

## Verification evidence

Candidate baseline: `d39eabd99e3df8830e6b9d571695d1b8778feabb` on
`task/TASK-0003-totp-controls`; Go version: `go1.23.12 linux/amd64`.

```text
GOCACHE=/tmp/codex-authority-broker-task3-review-cache go test ./...       PASS
test -z "$(gofmt -l .)"                                                   PASS
GOCACHE=/tmp/codex-authority-broker-task3-review-cache go vet ./...        PASS
GOCACHE=/tmp/codex-authority-broker-task3-review-cache go test -race ./... PASS
git diff --check                                                           PASS
```

The exact approved cumulative production-SLOC command reported:

```text
60 internal/lease/totp.go
170 internal/lease/lease.go
TOTAL 230
```

`230 <= 430` passes the cumulative ceiling.  `git status --short` showed only
the approved `internal/lease` candidate plus TASK-0003 PLAN/QA evidence before
this result was created.  No Makefile exists and the approved PLAN expressly
says not to add or invoke one.

No product, plan, test, or Git state was changed by this review.  The reviewer
created only this evidence file in the repository; the failing reproduction
used temporary `/tmp` overlay files.

---

## Revision 2 independent review

## Decision: PASS

The boot-floor remediation resolves the revision-1 finding.  The current
candidate satisfies the approved TASK-0003 replay, rate, RFC 6238,
concurrency, secret-boundary, scope, and cumulative-SLOC requirements.  This
PASS advances only the independent REVIEW gate; independent QA is still
required before any merge decision.

### Revision-1 finding disposition

**Resolved — adjacent future code no longer bypasses the startup-step boot
floor.** `VerifyAndActivate` now denies whenever the current clock counter is
less than or equal to `bootReplayFloor`, before examining `-1, 0, +1` drift
candidates.  Therefore even a valid `current+1` code denies during the boot
step, and a negative-time startup remains denied through current counter zero.
Once the fake clock advances to the exact next 30-second counter, the current
code may activate when otherwise valid and unused.

The checked-in `TestReplayWatermarkAndBootFloor` now explicitly submits the
future-adjacent code during the fresh state's boot step and receives
`ErrTOTPReplay`, then verifies same/older denial and activation at the exact
next step.  The revision-1 reviewer-only overlay reproduction now passes:

```text
GOCACHE=/tmp/codex-authority-broker-task3-r2-review-cache \
  go test -vet=off -overlay=/tmp/task3-reviewer-overlay.json \
  ./internal/lease \
  -run 'TestReviewerBootFloorRejectsFutureAdjacentCodeDuringBootStep|TestReplayWatermarkAndBootFloor|TestRateWindowAndChallengeExpiryReset' \
  -count=1
ok github.com/autotaker/codex-authority-broker/internal/lease
```

### Acceptance evidence

- Boot-floor, watermark, attempt-count, rate-window, challenge, and lease
  state remain guarded by the single `State.mu`.  Boot-floor denials occur
  after `recordAttempt`, so valid-adjacent, malformed, stale, and replay
  attempts reaching a live current challenge all consume the same budget.
- The first five attempts within the fixed 60-second window are admitted for
  verification, the sixth and later pre-boundary attempts return
  `ErrTOTPRateLimit`, and exact `firstAttempt+60s` starts a new window.
  Challenge expiry or consumption clears attempt state; a new readiness
  challenge receives a fresh window.
- RFC 6238 remains standard-library HMAC-SHA-1 with `T0=0`, 30-second Unix
  steps, dynamic truncation, six decimal digits, exact-width ASCII validation,
  and `hmac.Equal`.  Drift remains exactly `-1, 0, +1` after the boot floor is
  crossed.
- The barrier-based duplicate test still establishes exactly one successful
  activation; every other caller denies, and the race detector is clean.
- The verifier still copies a nonempty caller-supplied secret into an
  unexported field.  Errors remain fixed sentinels.  No production logging,
  formatting method, persistence, secret transport/input mechanism, external
  dependency, IPC/CLI/service, PAM/sudo, release, timer, or goroutine was
  introduced.
- Existing TASK-0001 tests are unchanged and pass.  Readiness, direct
  activation compatibility, restart inactivity, and the separate immutable
  300-second challenge/lease boundaries remain intact.
- Product changes remain confined to approved `internal/lease` paths.  Manual
  review found idiomatic gofmt structure with no semantic compression,
  semicolon packing, collapsed error handling, cryptic naming, removed
  security comments, or functions combined solely to meet the cap.

### Revision-2 verification evidence

Candidate baseline remains `d39eabd99e3df8830e6b9d571695d1b8778feabb`
on `task/TASK-0003-totp-controls`; Go version is `go1.23.12 linux/amd64`.

```text
GOCACHE=/tmp/codex-authority-broker-task3-r2-review-cache go test ./...       PASS
test -z "$(gofmt -l .)"                                                      PASS
GOCACHE=/tmp/codex-authority-broker-task3-r2-review-cache go vet ./...        PASS
GOCACHE=/tmp/codex-authority-broker-task3-r2-review-cache go test -race ./... PASS
git diff --check                                                              PASS
```

The exact approved cumulative production-SLOC command reported:

```text
60 internal/lease/totp.go
173 internal/lease/lease.go
TOTAL 233
```

`233 <= 430` passes.  `git status --short` showed only the approved
`internal/lease` candidate, TASK-0003 PLAN/QA documents, and this review
evidence.  No Makefile exists, and the approved PLAN says not to add or invoke
one.  This reviewer changed no product, plan, test, or Git state and appended
only this revision-2 evidence.
