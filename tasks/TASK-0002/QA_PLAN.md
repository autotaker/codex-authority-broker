# QA PLAN — TASK-0002 (revision 7)

## Decision and boundary

This is an independent, human-inspection QA plan for the rolling-wave planning
change. QA starts only after independent planning REVIEW is recorded PASS.
It inspects `backlog.json`, this task's approved PLAN, the five executable
task contracts (`TASK-0001`, `TASK-0003` through `TASK-0006`), the repository
filesystem, relevant historical PR #1 evidence, and the candidate diff.

QA performs the checklist by direct human inspection. It does not create
product behaviour, create or validate detailed TASK-0007-or-later contracts,
or write Git state. TASK-0002 is mergeable only when both REVIEW PASS and this
QA PASS are recorded. Any FAIL returns to its classified responsible gate and
never merges.

PR #1 / TASK-0001 revision-1 REVIEW remains historical **FAIL** evidence.
This planning change must preserve that decision and its findings; it neither
removes, downgrades, nor reclassifies them.

## Preconditions and execution clock

For each executable task, manually confirm and record that dependency PRs are
merged, required tools and permissions are available, the focused fixture is
ready, and any required remote or clean Ubuntu VM is ready. The 30-minute
delivery clock begins only after all four are confirmed. A failed preflight is
recorded as `not_started`, excluded from cycle-time evidence, and is not a
throughput observation.

## Independent human checklist

| ID | Human inspection | PASS evidence to record |
| --- | --- | --- |
| Q2-01 | Read `backlog.json`, TASK-0002 PLAN, and TASK-0001/0003/0004/0005/0006 contracts side by side. | Exactly five executable, mergeable tasks: 0001, 0003, 0004, 0005, 0006. Their sole dependency chain is `0001 -> 0003 -> 0004 -> 0005 -> 0006`; 0001 explicitly depends on none. Each estimate is 30 minutes and cumulative caps are exactly 250, 430, 650, 820, 820. |
| Q2-02 | Check each executable contract's ownership, acceptance, exclusions, preflight, cap, human-gate evidence, and merge rule against the index. | Boundaries are: 0001 readiness/absolute lease; 0003 TOTP replay/rate/concurrency; 0004 versioned fail-closed SO_PEERCRED IPC; 0005 Codex CLI/socket/redaction; 0006 measurement/replanning. Each has a bounded, non-overlapping scope, the stated dependency, the common preflight rule, its cap, and independent REVIEW PASS plus QA PASS / never-merge-on-FAIL. |
| Q2-03 | Manually inspect the filesystem under `tasks/` for directories and detailed task documents. | No detailed task contract or task directory with ID `TASK-0007` or later exists. TASK-0002 is recognized only as the planning contract, not an executable delivery task. Record the inspected paths and any absence result; do not manufacture a forbidden contract to test the rule. |
| Q2-04 | Compare `future_milestones` with PLAN and all task directories. | Exactly four reserves, and only four, exist: sudo <=950; push <=1400; audit plus attested release <=1500; clean canary adds 0 at <=1500. Every reserve is explicitly `executable: false` and ineligible for PLAN/DEV/branch/PR until conversion; no reserve has a detailed task contract. |
| Q2-05 | Inspect TASK-0006, PLAN, and `measurement_contract`. | Conversion occurs only after TASK-0006 independent REVIEW PASS, QA PASS, and PR merge; it converts only the next 2–3 evidence-supported milestones, updates contracts/backlog/cap reserves, and inserts another zero-SLOC measurement/replanning gate after them. No larger wave or speculative detailed contract is allowed. |
| Q2-06 | Inspect the measurement fields and rules, then review one complete representative evidence record when TASK-0006 is run. | Every completed executable task has `planned_production_sloc`, `actual_production_sloc`, `test_loc`, `plan_minutes`, `dev_minutes`, `review_minutes`, `qa_minutes`, `ci_push_merge_minutes`, `retries`, and `failure_classifications`. Preflight is separate and excluded; next-wave sizing uses observed completed cycles, applies exactly 20% time contingency, and makes no fixed-SLOC-throughput claim when evidence is sparse. |
| Q2-07 | Establish and record the production-SLOC baseline and counting method before reviewing caps. Inspect the production-source file inventory manually. | Production SLOC means nonblank, non-comment executable-source lines in shipped, installed, or runtime/installation non-test `.go`, shipped `.py`, `.sh`, or `.c`; exclude tests, planning/review/task/operator docs, workflows, declarative systemd/PAM/sudoers configuration, generated, and vendor. Current filesystem-inventory evidence is 0 qualifying production SLOC because the current tree contains no qualifying production-source file. Record the reviewed file inventory and resulting count. TASK-0006 adds exactly 0 and remains <=820; the global ceiling is <=1500. |
| Q2-08 | Compare no-compression and cap controls across `backlog.json`, PLAN, and all five contracts. | REVIEW rejects semicolon/one-line packing, collapsed error handling, cryptic names, removed security comments, and functions combined solely for LOC; `gofmt` and idiomatic structure are mandatory. Projected use above 90% of a cumulative cap requires re-estimation before DEV. An over-cap candidate follows the exact seven-item shedding order and obtains renewed PLAN and QA approval before DEV continues. |
| Q2-09 | Compare the ordered shedding list and mandatory-v1 controls verbatim across their planning sources. | The seven-item order is unchanged: canary; status/JSON UX; audit schema/correlation; pack-size/history diagnostics; remote-OID diagnostics; installer/rollback executable; GitHub push to v2. Readiness; TOTP replay/rate/absolute lease; fail-closed SO_PEERCRED; per-sudo `pam_exec` live check/no cache; argv/log secret non-disclosure; source-free attested artifact; and minimal external-trace-compatible audit cannot be shed. Mandatory v1 above 1500 is a `requirement_gap`, never compression or an implementation defect. |
| Q2-10 | Check the nine `acceptance_lineage` entries against the five task boundaries. | Each listed acceptance has exactly one owner: the two readiness/lease entries in 0001, two TOTP entries in 0003, two IPC entries in 0004, two CLI/redaction entries in 0005, and rolling-wave measurement/replanning in 0006. No lineage item points to a reserve or forbidden later task. |
| Q2-11 | Read the retained PR #1 `REVIEW_RESULT.md` and handover evidence, separately from the current TASK-0002 review record, and compare them with the historical statement in the index and plans. | The historical decision remains FAIL, its incompleteness/findings remain available, and no current artifact claims it passed or reclassifies it. |
| Q2-12 | Inspect `git diff` and `git status` manually for TASK-0002 scope. | The candidate change is limited to planning/index/contract evidence required by TASK-0002 and contains no planning-automation artifact. Record every changed/untracked path reviewed. No authority-broker production source, secret, generated release output, staging, commit, merge, or `.git` mutation is attributed to QA or this task. |

## Exact feature-shedding order

The canonical list must be exactly this, in order:

1. `automated canary executable; retain manual runbook/evidence`
2. `rich status/JSON UX; retain activate and immediate revoke`
3. `rich audit schema/correlation; retain minimal external-trace correlation ID, actor, scope, result, expiry`
4. `precomputed pack-size/history diagnostics; retain exact repo/ref/clean tree and normal non-force rejection`
5. `explicit remote-OID prefetch/race diagnostics; rely on standard non-force git rejection and generic failure`
6. `automated installer/rollback executable; retain declarative units and manual verified install/rollback`
7. `GitHub push capability moves to v2, leaving TOTP full-sudo authority as v1`

## Failure classification and evidence

An incorrect task set, dependency, cap, reserve, conversion boundary,
measurement field/rule, SLOC count/definition, contract/index disagreement,
or prohibited later detailed contract is a `planning_defect` and returns to the
TASK-0002 planning gate. An ambiguous conversion authority, missing mandatory
control, or mandatory v1 above 1500 is a `requirement_gap` and stops for
task-authority direction. Missing tooling, permissions, fixture, remote, or
clean-VM readiness is an `environment_issue`. Changed PR #1 FAIL evidence,
scope mutation, or a broken previously passing planning surface is a
`regression`. Product behaviour violating an approved executable contract is
an `implementation_defect`.

QA records the REVIEW PASS reference; checklist result for every ID; inspected
paths; manual SLOC inventory/count; focused-validation, acceptance, cap, and
dependency-boundary evidence; historical PR #1 evidence; `git diff`/status
scope evidence; and the classification/routing of every failure in
`QA_RESULT.md`. Main makes the final classification and merge decision.
