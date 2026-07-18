# TASK-0002: Rolling-wave planning and measurement contract

## Objective

Replace fixed long-range task detail with one executable first wave: TASK-0001, TASK-0003, TASK-0004, TASK-0005, and zero-SLOC TASK-0006 measurement/replanning. Preserve PR #1 revision-1 FAIL evidence, global <=1500 cap, no-compression rules, exact shedding order, and mandatory controls.

## Acceptance criteria

- backlog.json separates exactly five executable mergeable tasks from four future_milestones, each explicitly `"executable": false`.
- Future milestones are ineligible for PLAN, DEV, branch, and PR until TASK-0006 converts only the next 2-3 into detailed contracts.
- TASK-0006 records planned_production_sloc, actual_production_sloc, test_loc, plan_minutes, dev_minutes, review_minutes, qa_minutes, ci_push_merge_minutes, retries, and failure_classifications; its approved PLAN uses observed cycles with preflight excluded/not-started and 20% contingency, without fixed-throughput assumption when sparse.
- TASK-0006 updates backlog/contracts, cap reserves, and inserts the next remeasurement gate after 2-3 tasks.
- REVIEW records its manual checklist for acceptance evidence, SLOC/cap, and no-compression/guardrail compliance; QA independently records focused-validation, SLOC/cap, acceptance, and dependency-boundary PASS/FAIL evidence.

## Constraints

Planning docs only: do not edit QA_PLAN, REVIEW_RESULT, implementation, or Git state.

**Preflight prerequisites:** prior dependency PRs are merged; required tools and permissions are available; the focused fixture is prepared; any required remote or clean Ubuntu VM is ready. The 30-minute execution clock starts only after these conditions are confirmed. A preflight failure means the task has not started.

**Merge rule:** independent planning REVIEW PASS and QA PASS are both required; a FAIL returns to its responsible gate and never merges.
