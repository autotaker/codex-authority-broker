# QA PLAN — TASK-0007: production daemon/backend assembly and bounded seed

## Independent TASK-first acceptance baseline

This baseline was fixed from `AGENTS.md`, `TASK.md`, the user-specified later
TASK-0011 integration constraint, and the existing IPC/lease/client source
before reading `PLAN.md`.  QA owns this file only.  It does not authorize or
edit PLAN, product/test source, operational evidence, the Lap log, or Git.

TASK-0007 may add only:

1. `cmd/codex-authority-broker/main.go`;
2. `cmd/codex-authority-broker/main_test.go`;
3. `internal/backend/runtime.go`; and
4. `internal/backend/runtime_test.go`.

Existing `cmd/codex-authority`, `internal/ipc`, and `internal/lease` behavior
must remain unchanged.  Sudo, push, credentials, audit, release, installer,
packaging, canary, persistence, and broad configuration frameworks are out of
scope.

### P0 acceptance and boundary matrix

| ID | Criterion | Required PASS evidence / failure condition |
| --- | --- | --- |
| Q7-01 | Dependency and exact scope | TASK-0006 is merged before DEV; the candidate changes only the four owned paths and adds no generated, operational, Git, or secret-bearing file.  Any other product/test path is FAIL. |
| Q7-02 | Secure seed file acquisition | The seed path is fixed/bounded configuration, never argv/environment/IPC.  Open rejects symlinks and non-regular files, then owner, exact `0600` permissions, type, and bounded size are verified from the opened descriptor rather than a path-level check followed by a racy open.  Missing, empty, oversized, wrong-owner/mode/type, symlink, replacement/TOCTOU, short read, and read error fail before readiness. |
| Q7-03 | Strict seed schema and lifetime | One bounded document is accepted; unknown/duplicate fields, trailing data, malformed encoding, invalid/empty seed, and values outside the explicit size bounds deny.  The verifier receives a private copy.  Seed/config bytes are not returned, retained in avoidable copies, or exposed through logs/errors/status. |
| Q7-04 | Exact fixed routing and payloads | The runtime admits exactly existing protocol v1 `ready` and `otp`. `ready` requires no payload; `otp` requires exactly one six-digit `code` field and rejects absent/null/unknown/duplicate/trailing/malformed/oversized payloads. Unknown operations and payload/operation confusion never invoke lease/TOTP state. Success responses are bounded and contain no seed, OTP, internal challenge, deadline, or error detail. |
| Q7-05 | State construction and semantics | One in-memory `lease.State` and seed-derived `TOTPVerifier` are constructed for the process. `ready` maps only to readiness/challenge creation and `otp` atomically verifies/activates the current challenge. Existing absolute expiry, replay floor, rate limit, and lease behavior are preserved; malformed or failed requests do not create authority. |
| Q7-06 | Readiness and fail-closed startup | The socket/backend does not report ready before secure seed validation and complete state/server construction. Missing/invalid seed, construction/listen failure, unavailable backend, and restart without a freshly valid seed fail closed; no partially initialized listener or state is usable. |
| Q7-07 | Lifecycle and cleanup | The privileged entrypoint has one bounded startup path, deterministic signal handling, cancellation propagation, clean server close/socket cleanup, and nonzero failure exit. Shutdown is idempotent and does not hang active/partial clients; restart creates fresh in-memory state and never inherits authority. |
| Q7-08 | Redaction and observable failures | Seed, OTP, config contents, and sensitive paths/values never appear in stdout, stderr, logs, errors, panic text, response payloads, or readiness output. External failures are generic and bounded while tests prove distinct internal failure branches without leaking data. |
| Q7-09 | Safe later registration seam | TASK-0011 must be able to attach `internal/backend/push_registration.go` through explicit construction-time dependency injection, including its root-configured caller UID, without editing TASK-0007-owned `runtime.go`/broker `main.go`, package `init`, mutable global registries, hidden side effects, generic command dispatch, or enabling push now. A new same-package file that no existing construction path calls is not a seam and is FAIL. |
| Q7-10 | Regression and checks | Focused `go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease` passes, including old-client `ready`/`otp` behavior. Independent REVIEW and QA also run `GOCACHE="$(mktemp -d)" go test ./...`, gofmt cleanliness, `git diff --check`, and applicable JSON validation. Socket-restricted runners are `environment_issue`, not an excuse to waive execution in the approved fixture. |
| Q7-11 | SLOC/caps and no compression | TASK-0007 adds at most the approved candidate amount and remains below cumulative target 950 and hard guard 1050; forecast above cumulative 855 stops before DEV for approved replan. Production/test SLOC are independently counted. Packing, collapsed errors, cryptic names, removed security comments, or weakened/deleted tests is FAIL. |
| Q7-12 | Two-Lap and role gates | Preflight is `not_started` and excluded from DEV timing. Lap 1 must yield a focused-test-passing, gate-ready candidate; Lap 2 retains independent REVIEW, independent QA, and main-owned Git. A third likely Lap, combined role, partial merge, or gate bypass stops for split/replan. |
| Q7-13 | Measurement evidence | Evidence records production/test SLOC; same-task/lap/stage/attempt paired timings; separate `active_ms`/`wait_ms`; maximum propagated retries and source IDs; raw/effective correction-aware classifications; null reasons; preflight exclusion; and `ceil(observed_non_preflight_time * 1.20)`. Null is not zero and no fixed SLOC throughput is inferred. |

All Q7-01 through Q7-13 are P0.  QA PASS also requires independent REVIEW
PASS.  No P0 may be skipped because of the two-Lap target or SLOC forecast.

### Required adversarial evidence

QA independently exercises valid startup/ready/OTP/shutdown/restart and at
least these mutations: absent seed; symlink; directory/device/non-regular
input; wrong UID; group/other or extra permission bits; empty, oversized,
truncated, replaced, malformed, duplicate/unknown/trailing schema; invalid
encoded secret; listen/construction failure; ready with payload; OTP with
missing/null/non-string/non-six-digit/duplicate/unknown/trailing payload;
unknown operation; replay, rate limit, expiry, shutdown during a partial or
active request; repeated shutdown; and restart without seed.  Every deny path
must withhold readiness/authority and preserve redaction.

The later-registration test must compile and execute a test-only external
registrar using only the stable TASK-0007 API, prove it is called exactly once
at construction (not import/init time), receives the explicit dependency/config
surface it needs, can add one named handler without replacing `ready`/`otp`,
and remains absent when not injected.  This proves architectural reachability;
it does not implement or admit `push` in TASK-0007.

### Failure classification

| Classification | Disposition |
| --- | --- |
| `implementation_defect` | Candidate violates an approved P0, source boundary, redaction, lifecycle, payload, seed, regression, or cap rule. Return to DEV with the smallest reproducible mismatch. |
| `planning_defect` | PLAN cannot satisfy the TASK/user boundary, including an unreachable later registration/configuration path that requires changing forbidden ownership. Revise and reapprove PLAN/QA before DEV. |
| `qa_plan_defect` | A QA expectation contradicts TASK/user authority or existing fixed behavior. Amend/reapprove QA planning; do not blame DEV. |
| `requirement_gap` | A necessary user/security decision is genuinely absent and no safe interpretation is authorized. Return to task authority; do not guess. |
| `environment_issue` | The approved isolated root/socket fixture or required tooling is unavailable independent of candidate behavior. Record `not_started` where preflight applies and rerun in a capable fixture. |
| `regression` | Previously passing IPC/lease/client/security behavior changes outside the owned boundary. FAIL with before/after evidence. |

Independent pre-PLAN decision: **CONDITIONAL / DEV BLOCKED pending PLAN
reconciliation**.  In particular, PLAN must provide a concrete construction-
time registration and configuration path satisfying Q7-09.  If it instead
forbids later edits to both assembly files while relying only on a future
`push_registration.go`, classification is `planning_defect`, not
`requirement_gap`: the desired later owner is known, but the file-ownership
contract must be amended.

## PLAN reconciliation

`PLAN.md` was read only after the independent matrix above was fixed.  It
correctly keeps DEV stopped and preserves most of the TASK boundary, but it
does not yet provide an executable secure contract.

| Area | PLAN evidence | QA decision |
| --- | --- | --- |
| Four-path scope and role gates | PLAN limits TASK-0007 to the two broker/backend production paths and their tests, preserves the existing client/IPC/lease sources, keeps PLAN -> DEV -> independent REVIEW -> independent QA -> main Git, and stops before DEV. | **MATCH / PASS.** Q7-01 and Q7-12 are preserved. |
| Fixed seed schema and redaction | Fixed production socket/seed paths, 1024-byte disk cap, strict two-field schema, base64 secret bounds, non-root allowed UID, exact `0600`, generic failure, buffer zeroing limits, and the capture matrix are explicit. | **PARTIAL / FAIL on acquisition below.** Schema, bounded-copy, and redaction intent otherwise satisfy Q7-03/Q7-08. |
| Symlink and TOCTOU-safe acquisition | PLAN requires a non-symlink regular file but specifies production `os.Open` followed by descriptor `Stat`. Descriptor metadata closes path-stat/file-swap races after open, but `os.Open` follows a final-component symlink; a symlink to a qualifying root-owned `0600` regular file passes the proposed checks. Parent-component resolution is also not bounded by a descriptor walk. | **MISMATCH / FAIL — `planning_defect`.** Q7-02 is not implementable from the stated production seam while preserving the promised symlink rejection. |
| Exact ready/OTP routing | PLAN admits only payload-free `ready` and a strict single-field six-digit OTP object, maps them to one process-local state/verifier, returns payload-free generic responses, preserves outer IPC/SO_PEERCRED admission, and specifies fail-closed construction/restart. | **MATCH / PASS.** Q7-04 through Q7-08 are sufficiently concrete, subject to implementation evidence. |
| Later TASK-0011 reachability | PLAN correctly observes that an exported `Register` method plus a future `push_registration.go` has no invocation source. TASK-0011 forbids changes to broker `main.go`; import-time `init` or a mutable global registry would be the only implicit attachment and is disallowed. The later push-specific root configuration also has no explicit construction-time injector. | **MISMATCH / FAIL — `planning_defect`.** Q7-09 fails. A capacity slot is not a reachable assembly path, and combined `planning_defect/requirement_gap` is not an acceptable fixed classification. |
| SLOC/cap evidence | PLAN retains +90/841 forecast, 855 trigger, 950 target, 1050 hard guard, and stop/replan rules, but calls `git diff --numstat` an actual production-SLOC count. `numstat` reports added/deleted physical diff lines and churn; it does not count canonical nonblank/non-comment executable production SLOC or establish cumulative SLOC. | **MISMATCH / FAIL — `planning_defect`.** Q7-11 requires the canonical baseline-plus-candidate count; churn may be supplemental only. |
| Focused/full checks and measurement | Focused package tests, full Go/gofmt/diff/JSON checks, adversarial secret capture, separate timings, retries/classifications, null handling, contingency, and no-throughput sizing remain required. | **MATCH / PASS.** Q7-10/Q7-13 are preserved; an incapable socket fixture remains `environment_issue`, never a waiver. |

### Smallest safe contract amendments

1. Revise TASK-0011's production/test ownership and exclusions to permit a
   narrowly bounded change to `cmd/codex-authority-broker/main.go` and
   `main_test.go`.  That change must explicitly construct the push-specific
   root-configured dependencies and call the fixed registration function from
   `push_registration.go` exactly once.  TASK-0007 may provide the bounded
   instance-level registration API now, but no global registry, `init`, file
   discovery, or push admission.  If push configuration must alter runtime's
   seed schema rather than remain owned by `push_registration.go`, TASK-0011
   must also explicitly own the corresponding bounded runtime/test change.
2. Replace production `os.Open` with a Linux descriptor-relative no-follow
   acquisition: walk the fixed root-owned directory components without
   symlinks, open the final file with no-follow/close-on-exec semantics, and
   validate regular type, UID, exact mode, and size from that same descriptor
   before the bounded read.  Tests must include final and parent symlink and
   replacement attempts.
3. Replace the cap gate's claimed `numstat` SLOC measurement with the canonical
   nonblank/non-comment executable-source count over the merged baseline and
   candidate owned production files.  Keep `numstat` only as optional churn
   evidence and reconcile the resulting cumulative count to 855/950/1050.

PLAN reconciliation decision: **FAIL — DEV remains blocked**.
Classification: **`planning_defect`**.  The user decision is present: a later
bounded push registrar must attach without global side effects.  What is
missing is compatible file ownership and a construction call site, so this is
not a `requirement_gap`.  The same fixed classification covers the insecure
open design and noncanonical cap measurement.  After the three narrow
amendments, PLAN and QA_PLAN must be reapproved before preflight or DEV.

## PLAN reconciliation Revision 2 — Revision 3 contract re-estimate

Main-authorized consistency edits to `backlog.json` and TASK-0007 through
TASK-0012 contracts were independently checked against PLAN Revision 3 and the
three prior QA findings.  The authorization supersedes the earlier Q7-09
prohibition on any later broker-main edit only for TASK-0011's single explicit
fixed registration call; it does not weaken the no-global/no-`init` boundary.
It also supersedes Q7-11's original +90/855/950 planning values through the
documented trigger and approved re-estimation below.

| Area | Revision 3 evidence | QA decision |
| --- | --- | --- |
| Later registrar reachability | TASK-0011 now owns `cmd/codex-authority-broker/main.go` and `main_test.go` solely to construct push-specific dependencies and call the TASK-0011-owned fixed registration function exactly once with the runtime instance.  TASK-0007 supplies only an instance-level bounded third-slot API; TASK-0011 forbids `init`, mutable globals, discovery, generic routing, and every unrelated ready/OTP/seed/signal/socket edit. | **MATCH / PASS.** The former unreachable-file defect is closed with an explicit construction call and narrow sequential ownership. |
| Descriptor-safe seed acquisition | PLAN/TASK-0007 now walk from `/` through fixed components using descriptor-relative `openat` with `O_DIRECTORY`, `O_NOFOLLOW`, and `O_CLOEXEC`, open the final file `O_RDONLY|O_NOFOLLOW|O_CLOEXEC`, `fstat` that same descriptor, validate directory/file type, UID/mode/size, close every descriptor, and test final/parent symlinks and replacement attempts. `os.Open` is explicitly excluded. | **MATCH / PASS.** The final-symlink and path-stat/open TOCTOU defect is closed without elevation or a new dependency. |
| Canonical SLOC evidence | PLAN supplies repository-wide and owned-path commands for nonblank/non-comment non-test Go SLOC, includes tracked and untracked candidate files, reports file subtotals, and requires all-source total to equal `751 + OWNED_ADDED`; `numstat` is optional churn only.  The all-source command independently reproduced **TOTAL 751** on the merged source. | **MATCH / PASS.** Canonical baseline and candidate reconciliation replace the invalid churn-as-SLOC claim. |
| Exact index/contract equivalence | The first JSON metadata block in each TASK-0007 through TASK-0012 contract compared equal as a complete JSON object to its matching `backlog.json` task entry. IDs are unique and dependencies remain the linear `0006 -> 0007 -> 0008 -> 0009 -> 0010 -> 0011 -> 0012` chain. | **MATCH / PASS.** No field-level identity, ownership, path, fixture, Lap, exclusion, stop, lineage, or eligibility drift was found. |
| Re-estimate and cap arithmetic | Starting at 751, additions `170 + 120 + 0 + 130 + 220 + 0 = 640` yield forecasts `921, 1041, 1041, 1171, 1391, 1391`. Stops are `975, 1100, 1100, 1250, 1450, 1450`; targets `1000, 1150, 1150, 1300, 1500, 1500`; hard guards `1050, 1250, 1250, 1400, 1650, 1650`. Every forecast is strictly below its stop, every stop below target, and every target below hard. | **MATCH / PASS.** The original 855 trigger fired and caused an explicit re-estimate; no forecast self-blocks. Global target/hard remain 1500/1800. `1500 - 1391 = 109` and `1800 - 1391 = 409` remain gated later reserves. |
| Two-Lap, checks, profile, and gates | Every contract retains only Lap 1 and Lap 2, with no Lap 3; TASK-0007 Lap 1 remains one complete bounded candidate and Lap 2 independent REVIEW/QA. Focused and repository-native full checks remain additive. PLAN selects **sol-high** for the interacting privileged descriptor, secret, lifecycle, state, and future-registration risks. | **MATCH / PASS.** No role merge, gate bypass, compression, or extra-Lap assumption was introduced. |

Revision 2 final reconciliation decision: **PASS**. Classification:
**`none`**. The prior `planning_defect` is resolved by the authorized contract
and re-estimation amendments. TASK-0007 may proceed to preflight and DEV only
under the approved **sol-high** profile, within its exact four paths, stop 975,
target 1000, hard guard 1050, max-two-Lap contract, and normal independent
REVIEW/QA plus main-owned Git gates. This reconciliation changes only
`tasks/TASK-0007/QA_PLAN.md`; it does not edit PLAN, contracts, backlog,
product/test source, operational evidence, Git state, or the Lap log.

## Independent split remeasurement — Revision 3

This assessment was derived first from the then-current TASK-0007 authority,
the two candidate source files, and the supplied canonical measurements only;
it did not use a third TASK-0007 Lap or rely on TASK-0007's former 170-SLOC
forecast. Observed facts are `main.go=186`, `runtime.go=150`, no owned tests,
and baseline `751`, therefore `751 + 186 + 150 = 1087`. Both packages compile
independently, but that is compile evidence only: there are still no owned
tests, REVIEW, or QA candidate evidence.

The combined candidate exceeds TASK-0007's current stop `975` and hard guard
`1050`. This is a **`planning_defect`**, not a reason to pack code, remove seed
admission/redaction/lifecycle controls, weaken test coverage, or create a
third Lap. It cannot receive QA PASS as the current TASK-0007 candidate.

### Derived safe split acceptance boundary

| Resulting task | Compile-valid owned boundary | P0 acceptance retained | Lap limit / failure condition |
| --- | --- | --- | --- |
| TASK-0007 (runtime-only) | `internal/backend/runtime.go` plus its test; no seed open, listener, signals, or daemon startup. | Process-local lease/TOTP assembly from injected secret; exact current `ready`/`otp` denial/routing; close/fail-closed; one bounded instance registration seam with no `init`, global registry, discovery, or enabled push. | Lap 1 DEV and Lap 2 independent REVIEW+QA only. Need for seed, socket, lifecycle, persistence, or broker-main work stops/replans; it never moves to TASK-0008. |
| TASK-0013 (secure seed/daemon) | `cmd/codex-authority-broker/main.go` plus its test; constructs only merged TASK-0007 runtime and compiles against unchanged IPC. | Descriptor-relative `openat`/`O_NOFOLLOW` acquisition; final-descriptor metadata/schema/size checks; redaction; fixed listen/serve/signal-close/restart and fail-closed readiness. Tests retain valid and every invalid seed/lifecycle mutation in Q7-02, Q7-03, Q7-06--Q7-08. | Lap 1 DEV and Lap 2 independent REVIEW+QA only under a fresh approved TASK/PLAN/QA set. Need for runtime/API change or added operation after TASK-0007 merge stops/replans. |

The split preserves, rather than compresses, every Q7-02--Q7-08 security
property. TASK-0008 remains the sudo live-check/no-cache task and receives
neither seed, lifecycle, runtime, nor broker-main work.

### PLAN Revision 4 reconciliation

| Check | Independent QA result | Required minimum correction |
| --- | --- | --- |
| Arithmetic and candidate disposition | **PASS.** 150 + 186 reproduces 336; `751 + 336 = 1087`, correctly blocking the combined candidate. Proposed cumulative 901 after runtime-only and 1087 after seed/daemon are sound. | Keep both candidates as unaccepted draft evidence; do not delete, compact, stage, merge, or treat either as a completed Lap. |
| Compile-valid boundary | **PASS, conditional on contracts.** The two observed packages compile independently, and TASK-0013 consumes a merged runtime. | Re-authorize TASK-0007 to runtime/test paths only; give TASK-0013 its own contract, PLAN, QA plan, tests, and independent gates before either DEV start. |
| Dependency and current-draft handling | **FAIL — planning_defect.** Proposed `0006 -> 0007 -> 0013 -> 0008 -> 0009 -> 0010 -> 0011 -> 0012` is correct, but current index/contracts still declare `0007 -> 0008`, old cumulative values, and no TASK-0013. The PLAN labels reconciliation later, so it is not executable authority. | Make one separately approved atomic contract/index amendment: add TASK-0013; change TASK-0008 dependency to TASK-0013; propagate dependencies/cumulative caps/stops/targets/hards through TASK-0012; reconcile every TASK metadata block to the index. Do not start DEV from this draft. |
| Two-Lap / role preservation | **PASS in PLAN; FAIL as executable authority.** Revision 4 describes two Laps per task, but the current TASK-0007 still assigns combined work to its original two-Lap contract. | Atomic amendment must state exactly two Laps for TASK-0007 and TASK-0013, retain PLAN -> DEV -> independent REVIEW -> independent QA -> main Git, and prohibit third-Lap implementation. |
| TASK-0008 scope isolation | **PASS in PLAN; FAIL until recorded.** PLAN keeps TASK-0008 unchanged while its contract names old predecessor/1041 forecast. | Update only dependency/cap lineage needed to consume TASK-0013; preserve TASK-0008 production/test ownership and sudo/no-cache acceptance. |
| 1500 target and 1800 hard | **FAIL — planning_defect.** `751+150+186+120+0+130+220+0 = 1557`: 57 over target and below hard 1800. More importantly TASK-0011's proposed start stop is 1490, so saving only 57 still self-blocks; at least **67** must be removed from 220 to reach `1337 + 153 = 1490`. PLAN gives priorities but no bounded revised TASK-0011 scope/allocation. | Before TASK-0011 DEV, amend its contract with explicit retained-core/removed-optional inventory and per-path canonical forecast totaling **<=153** (or defer named optional diagnostics to a new non-executable post-TASK-0012 reserve). Do not count already-excluded features as savings; retain authorization, schema admission, UID/live-lease, custody/redaction, system-Git capture, and non-force. Recompute TASK-0011/0012 values from that approved number. |

**Revision 3 decision: FAIL — DEV remains blocked.** Classification is
`planning_defect`, not implementation, environment, or requirement gap. The
smallest safe remedy is the atomic reconciliation plus a measurable <=153-SLOC
TASK-0011 core (or an explicitly deferred later reserve), followed by fresh
approvals. No product code, Git state, operational repository, or Lap evidence
was changed by this QA assessment.

QA planning observation: `active_ms=null` (no authoritative same-attempt
elapsed pair supplied), `wait_ms=null`, `retries=0`,
classification=`planning_defect`; null means unobserved, never zero.

## PLAN reconciliation Revision 4 — PLAN Revision 5

The TASK-first Revision 3 split criteria above remain authoritative for this
comparison. QA independently re-added every allocation and cumulative value;
no current contract/index draft was treated as approved authority.

| Check | Revision 5 QA decision |
| --- | --- |
| Proposed eight-document atomic amendment | **PASS only for Revision 5 disposition 1; otherwise FAIL.** The set is exactly `backlog.json`, TASK-0007, new TASK-0013, and TASK-0008 through TASK-0012: eight contract/index documents. It correctly excludes the separately required TASK-0013 PLAN/QA plan, requires whole-object index/contract equivalence, inserts `0007 -> 0013 -> 0008`, and preserves TASK-0008's sudo scope. No subset is executable. If runtime receives a new Task ID under disposition 2, however, the set and dependency chain necessarily gain another new contract and must be recalculated; the eight-document claim then ceases to be exact. |
| TASK-0011 220 -> 153 allocation | **PASS.** `16 + 4 + 26 + 36 + 33 + 38 = 153`; reductions `19 + 19 + 0 + 22 + 7 = 67`, and the unchanged registration/gate allocation is correctly split `4 + 36 = 40`. The removed items are convenience, extensibility, caching/retry/telemetry, and diagnostics, while fixed caller/ref identity, strict schema, UID/live-lease/TASK-0010 policy gates, memory-only custody/redaction, deterministic system-Git capture, and single-ref non-force behavior remain P0. Already excluded features are explicitly not credited. The readability/no-packing stop makes these ceilings forecasts rather than authority to compress. |
| 1490/1495 and global caps | **PASS.** `751 + 150 + 186 + 120 + 0 + 130 + 153 + 0 = 1490`; TASK-0011/0012 stop 1495 leaves 5 SLOC forecast margin, target 1500 leaves 10, and unconditional hard 1800 leaves 310. `1490 < 1495 < 1500 < 1800`. TASK-0011 must still stop/replan if readable retained-core implementation is forecast or measured above 1495; TASK-0012 must reconcile actuals to both global gates. |
| Current TASK-0007 Lap history | **FAIL — planning_defect.** Revision 5 correctly says the current attempt reached its authorized Lap 2, cannot use Lap 2 for missing-test DEV, cannot PASS, and must preserve its evidence. But disposition 1 then proposes a fresh two-Lap cycle under the same TASK-0007 ID and declares it is not Lap 3. With the existing history retained, that is a counter reset/relabel in effect and violates Q7-12 and the user-required maximum two Laps per resulting Task. Task authority can supersede scope, but QA cannot infer that changing the boundary makes additional Laps under the same ID cease to be third-and-later Laps. |

**Revision 4 decision: FAIL — authority decision required; all DEV remains
blocked.** Classification remains `planning_defect`. The smallest compliant
next action is to reject disposition 1 and record disposition 2: terminate the
combined TASK-0007 attempt without completion, assign runtime-only and
seed/daemon to distinct new Task IDs, then regenerate the atomic document set,
dependency chain, and otherwise unchanged 150/186/153 arithmetic. Each new
Task may then receive its own maximum-two-Lap PLAN/DEV/REVIEW/QA cycle without
altering or relabeling TASK-0007 history. If task authority instead intends an
explicit exception to the maximum-two-Lap invariant, that is a new requirement
decision and must return to QA planning; it is not approved here.

No product/test source, contract/index document, Git state, operational
repository, or Lap evidence was changed. QA observation:
`active_ms=null` (no authoritative same-attempt elapsed pair supplied),
`wait_ms=null`, `retries=0`, classification=`planning_defect`; null means
unobserved, never zero.

## PLAN reconciliation Revision 5 — PLAN Revision 6

Revision 6 was checked read-only against the independent TASK-first split and
Lap-history criteria above. This decision evaluates internal consistency of a
proposal only; it grants no authority to terminate a Task, amend contracts,
reuse draft source, or begin DEV.

| Check | Revision 6 QA decision |
| --- | --- |
| Disposition 2 and Lap history | **MATCH / PASS as proposal.** TASK-0007 terminates incomplete with zero merged production SLOC, no PASS/merge/completion, intact Lap 1/Lap 2 evidence, and no reset, rename, Lap 3, or further gate. Runtime and seed/daemon become distinct TASK-0013 and TASK-0014 cycles, each with only Lap 1 DEV and Lap 2 independent REVIEW+QA. This closes Revision 4's same-ID restart defect. |
| Ownership and dependency chain | **MATCH / PASS as proposal.** TASK-0013 depends on TASK-0006 and exclusively owns runtime/test; TASK-0014 depends on TASK-0013 and exclusively owns broker main/test. The executable chain `0006 -> 0013 -> 0014 -> 0008 -> 0009 -> 0010 -> 0011 -> 0012` is acyclic and compile-valid. Terminated TASK-0007 remains a non-completing sibling after TASK-0006 and is not falsely represented as a prerequisite. TASK-0008 retains its original sudo/no-cache scope. |
| Nine-document atomic set | **MATCH / PASS as proposal.** The exact set is backlog plus TASK-0007, new TASK-0013, new TASK-0014, and TASK-0008--0012: nine documents. TASK-0013/0014 are unused in the current index. Whole-metadata equivalence, explicit termination fields, exclusive ownership, downstream dependency/cap propagation, and no-subset execution are all required. Replacement PLAN/QA plans correctly remain separate gate documents rather than being miscounted as contract/index reconciliation. |
| TASK-0011 safety core | **MATCH / PASS as proposal.** The unchanged allocation totals `16 + 4 + 26 + 36 + 33 + 38 = 153`; the 67-SLOC reduction removes only optional convenience/framework/cache/retry/diagnostic scope, credits no already-excluded feature, retains every fixed authorization/custody/redaction/non-force control, and stops rather than compresses if readable implementation cannot fit. |
| Cumulative target and hard guard | **MATCH / PASS as proposal.** Terminated TASK-0007 adds zero merged SLOC. `751 + 150 + 186 + 120 + 0 + 130 + 153 + 0 = 1490`; final stop margin is `1495-1490=5`, target margin `1500-1490=10`, and hard margin `1800-1490=310`. Every forecast is below its stated stop, while TASK-0012 still must reconcile actuals to both 1500 and 1800. |
| Authority and executability | **BLOCKED, by design.** Terminating TASK-0007 and replacing its objective changes the user-authorized outcome. PLAN explicitly withholds that authority and forbids the nine-document amendment, candidate adoption, and all DEV until user/task-authority approval is recorded. Main or any Agent treating this QA match as authorization is FAIL. |

**Revision 5 decision: PASS as an internally consistent, non-executable
proposal; DEV AUTHORIZATION: NO.** The earlier `planning_defect` remains the
classification of the terminated combined attempt. The only next action is an
explicit user/task-authority decision approving or rejecting Disposition 2.
If approved, Main may perform the separately controlled atomic contract/index
reconciliation and obtain fresh TASK/PLAN/QA approvals for TASK-0013 and
TASK-0014; no DEV begins merely from this PLAN/QA reconciliation. If rejected,
all affected work remains blocked and no document subset may be applied.

No product/test source, contract/index document, Git state, operational
repository, or Lap evidence was changed. QA observation:
`active_ms=null` (no authoritative same-attempt elapsed pair supplied),
`wait_ms=null`, `retries=0`, classification=`planning_defect`; null means
unobserved, never zero.
