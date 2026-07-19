# TASK-0012 QA plan — frozen actual-measurement gate

## Independent baseline, authority, and preconditions

This TASK-first plan is derived from `TASK.md`, `backlog.json`, and the
TASK-0009 measurement contract before any TASK-0012 `PLAN.md` exists. QA
validates process evidence only: it must not change product/test code, the
canonical JSONL, TASK-0010/0011 reserve scope, Git state, or later-milestone
detail.

The forecast lineage is **1253 + 145 = 1398**, but neither value substitutes
for the frozen post-TASK-0017 actual result. TASK-0012 must report `actual
post-TASK-0017 cumulative + 0`, with actual cumulative below both 1500 and
1800. No throughput, reserve borrowing, or time contingency may create SLOC.

The read-only frozen post-TASK-0017 source is
`/tmp/task0012-events-frozen.jsonl`: 404 lines, SHA-256
`00482bd0db254b700848999d79e1cd67fe5dd3a0e5064b0847d92bbe7f932e05`.
Parse, unique-ID, and exact correction-predicate checks pass. It has no
TASK-0017 event, so no terminal may be synthesized and TASK-0017 timing is
null with that explicit reason. Status provenance is instead merge
`212fd03492d72a38970e54b6a303806a04c9c7c5` / PR #18 plus its independent
REVIEW and QA PASS records. The merged full-tree count is 1407; parent
`1d903d7` is 1262, yielding the independently reproducible +145 delta.

## Acceptance matrix

| ID | Mode | Independent procedure and PASS evidence | Fail-closed classification |
| --- | --- | --- | --- |
| Q12-01 | evidence-review | Verify frozen snapshot hash, line count, JSON parse, unique `event_id`s, and every correction: nonempty target; earlier file order; same task/lap; smaller sequence. Retain raw target/source IDs and derive effective values only after this predicate. | Missing/mutable/mismatched source or invalid correction: `requirement_or_canonical_evidence_defect`; stop. |
| Q12-02 | evidence-review | Independently select historical TASK-0001/0003/0004/0005 and product terminals. Exclude corrections, `counts_as_product_lap:false` partials, governance, terminated, unfinished, TASK-0013 `pass` pseudo-terminals, and any synthetic TASK-0017 terminal. Preserve TASK-0013 and TASK-0017 timing as null with their reasons. | Cohort mixing, invented terminal, or reasonless/zero null: `measurement_implementation_defect`. |
| Q12-03 | evidence-review | Rebuild full-tree executable nonblank/non-comment production SLOC: merged 1407, parent 1262, delta +145. Reconcile `1253 + (47 - 38) + 145 = 1407`; the nine lines are the documented TASK-0008 sudo-helper undercount, not TASK-0012 scope. Confirm TASK-0012 adds zero and records `1407 + 0 = 1407`, with target/hard reserves 93/393; 1398 remains forecast lineage only. | Unexplained drift, non-reproducible count/delta, forecast substitution, nonzero TASK-0012 production SLOC, actual >= 1500, or hard-limit risk: `requirement_or_process_evidence_conflict`; stop. |
| Q12-04 | evidence-review | Trace every required SLOC/test LOC/stage/cycle/active/wait/retry/classification field to source IDs. Stage duration uses only same-task/lap/stage/attempt start-terminal pairs; preflight is `not_started`; active and wait stay separate; retries are propagated maxima; nulls retain reasons; only observable non-preflight time uses `ceil(ms * 1.20)`. | Imputation, cross-pairing, double count, raw/effective blend, or multiplier on non-time/null data: `measurement_implementation_defect`. |
| Q12-05 | evidence-review | Inspect Task/backlog/future-milestone metadata and changed paths. Candidate may update only TASK-0012 process evidence and the matching arithmetic/status metadata. TASK-0010/0011 remain `deferred-v2`, `executable:false`, and zero v1 SLOC; milestones remain reserved/non-executable with no PLAN/DEV/branch/PR-ready detail. Canonical JSONL and product/test paths are untouched. | Reserve enablement, later-work detail, canonical/product/test edit, or scope expansion: `scope_or_requirement_defect`; stop. |
| Q12-06 | focused-rerun | From the candidate worktree, independently rerun affected deterministic evidence checks: frozen-source correction predicate, full-tree/Git-count reconciliation, changed-path scope, `git diff --check`, and `jq -e . backlog.json >/dev/null`. Record command, candidate commit/tree, source hash, exit, and output digest. | Candidate/tree mismatch or check failure: classify `measurement_implementation_defect`, `requirement_or_process_evidence_conflict`, or `environment_issue` from direct evidence; do not presume DEV fault. |

Repository-wide Go tests and formatting checks are intentionally excluded:
this candidate changes no product/build input and they cannot prove the
measurement result. They are not substituted with a product PASS.

## Gate and reconciliation

Candidate PASS requires Q12-01 through Q12-06, actual cumulative `<1500` and `<1800`,
independent REVIEW and QA regeneration of the same frozen candidate, and
Main's merge. A canonical defect, unexplained cap drift, absent independent
regeneration, any QA/REVIEW failure, or unresolved provenance/state conflict
keeps later reserve non-executable.

## PLAN reconciliation (after matrix fixation)

The amended PLAN and TASK align with this matrix: frozen-source identity and
correction validation; no synthetic TASK-0017 terminal; Git/REVIEW/QA status
provenance; full-tree 1407 count; +145 tree delta; explained TASK-0008
nine-line correction; zero TASK-0012 delta; caps; reserve boundary; and only
affected evidence checks. The candidate must synchronize `backlog.json` with
the amended Task arithmetic/status; Q12-05 independently verifies it.

**QA-plan approval: APPROVED for the evidence-only candidate.** This is not a
candidate QA PASS and does not enable later reserve work.
