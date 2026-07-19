# TASK-0009: Measurement and rolling-wave replanning gate

**Depends on:** TASK-0008.

**Status:** planned and executable.

**Preflight prerequisites:** TASK-0007 and TASK-0008 are merged; their
canonical event/evidence records are complete and readable; the worktree and
read-only measurement source are usable.  The clock starts only after those
checks.  A preflight failure is `not_started` and excluded from cycle timing.

**Owns:** the next zero-production-SLOC measurement and replanning gate.  It
retains TASK-0001 through TASK-0006 history, measures TASK-0007/TASK-0008,
classifies raw/effective failures with correction provenance, applies 20% time
contingency, and converts only the next 2–3 evidence-supported reserves.

**Acceptance:** every completed new Task records planned/actual/cumulative
production SLOC, test LOC, PLAN/DEV/REVIEW/QA/CI-push-merge and cycle time,
active/wait observations, retries, raw/effective classifications, source IDs,
null reasons, and preflight exclusion.  Replanning assumes no fixed SLOC
throughput and inserts the following measurement gate after the converted
wave.

**Excludes:** implementation of audit, release, installer, canary, or any
other product behavior; detailed executable contracts beyond the next 2–3
evidence-supported items.

**Production LOC ceiling:** adds **0 production SLOC**; cumulative ceiling
remains **<=1400** before any next-wave conversion.

**Reserve boundary:** `MILESTONE-audit-release` (cumulative reserve <=1500)
and `MILESTONE-clean-canary` (+0) remain non-executable until TASK-0009 REVIEW
PASS, QA PASS, and merge.  No branch, PLAN, DEV, or PR-ready contract may begin
for them earlier.

**Human gate evidence:** REVIEW records source integrity, complete/null-aware
arithmetic, corrections, contingency, sparse-evidence discipline, index/task
consistency, reserves, exact zero-SLOC delta, and scope.  QA independently
repeats every check and records PASS/FAIL.

**No-compression and scope control:** preserve the global <=1500 cap, >90%
re-estimation trigger, exact ordered shedding sequence, and mandatory-control
exclusions.  Mandatory v1 above 1500 is `requirement_gap`, never permission to
compress or bypass measurement.

**Merge rule:** independent REVIEW PASS and QA PASS are required.  Any FAIL
returns to its responsible gate and never merges.
