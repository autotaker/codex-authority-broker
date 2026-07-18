# PLAN — TASK-0002 (revision 5)

## Rolling-wave decision

Later 30-minute sizing is not knowable yet. Only the following detailed items are executable/mergeable:

| Task | Boundary | Depends on | Cumulative SLOC |
| --- | --- | --- | --- |
| 0001 | lease/readiness/absolute expiry | — | <=250 |
| 0003 | TOTP replay/rate/concurrency | 0001 | <=430 |
| 0004 | versioned SO_PEERCRED IPC | 0003 | <=650 |
| 0005 | Codex CLI/socket/redaction | 0004 | <=820 |
| 0006 | measurement/replanning gate | 0005 | <=820 (adds 0) |

**Preflight prerequisites:** prior dependency PRs are merged; required tools and permissions are available; the focused fixture is prepared; any required remote or clean Ubuntu VM is ready. The 30-minute execution clock starts only after these conditions are confirmed. A preflight failure means the task has not started.

## Non-executable reserves

The index records only these reserves: sudo <=950; push <=1400; audit plus attested release <=1500; clean canary adds 0. Each carries `"executable": false` and is ineligible for PLAN/DEV/branch/PR. TASK-0006, after independent REVIEW PASS, QA PASS, and PR merge, may convert only the next 2-3 evidence-supported milestones into detailed contracts, update reserves, and insert another measurement gate after those tasks.

## Measurement contract

For every completed task, TASK-0006 records planned_production_sloc, actual_production_sloc, test_loc, plan_minutes, dev_minutes, review_minutes, qa_minutes, ci_push_merge_minutes, retries, and failure_classifications. Its approved next-wave PLAN derives size from observed completed cycles; excludes preflight time (a preflight failure is not-started); adds 20% time contingency; does not assume fixed SLOC throughput when evidence is sparse; converts only the next 2-3; and inserts the next measurement gate after those tasks.

## Human gate evidence

REVIEW records a checklist confirming task-acceptance evidence, the production-SLOC count against its ceiling, and no-compression/guardrail compliance. QA independently repeats the focused validation and SLOC count, compares the result with the acceptance and dependency boundary, and records PASS or FAIL evidence.

## Guardrails

Global production SLOC remains <=1500. No compression, >90% re-estimation, mandatory controls, and the exact JSON shedding order remain unchanged. The order is:

1. automated canary executable; retain manual runbook/evidence
2. rich status/JSON UX; retain activate and immediate revoke
3. rich audit schema/correlation; retain minimal external-trace correlation ID, actor, scope, result, expiry
4. precomputed pack-size/history diagnostics; retain exact repo/ref/clean tree and normal non-force rejection
5. explicit remote-OID prefetch/race diagnostics; rely on standard non-force git rejection and generic failure
6. automated installer/rollback executable; retain declarative units and manual verified install/rollback
7. GitHub push capability moves to v2, leaving TOTP full-sudo authority as v1

Never shed readiness, TOTP replay/rate/absolute lease, SO_PEERCRED fail-closed IPC, per-sudo pam_exec live check/no cache, argv/log secret non-disclosure, source-free attested artifact, or minimal external-trace-compatible audit. Mandatory v1 above 1500 is a requirement gap, never compression.
