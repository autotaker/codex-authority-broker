# TASK-0007: Per-sudo live lease check and no cache

**Depends on:** TASK-0006.

**Status:** planned and executable.

**Preflight prerequisites:** TASK-0006 is merged; the exact daemon/backend
assembly boundary and an isolated Linux sudo fixture are approved and usable;
required tools and permissions are available.  The execution clock starts
only after those checks.  A preflight failure is `not_started` and excluded
from measured execution time.

**Owns:** the smallest Linux sudo integration that checks live authority on
every invocation through the existing fail-closed local boundary, with sudo
timestamp caching disabled for the dedicated Codex identity.

**Acceptance:** a live unexpired lease permits the focused sudo fixture;
expired, restarted, unavailable, malformed, or unauthorized checks deny.
Consecutive invocations independently check current lease state and cannot
reuse cached authority.

**Excludes:** push, GitHub credentials, audit, release, installer, packaging,
and canary work.

**Production LOC ceiling:** cumulative production SLOC **<=950**.  The
available reserve from the measured 751 baseline is 199 SLOC.  If mandatory
assembly and per-invocation/no-cache controls cannot fit idiomatically, stop
as `requirement_gap`; never weaken them or compress source.

**Initial execution-time evidence:** post-preflight minimum **76 minutes**,
upper bound `unknown`.  Source analogue TASK-0004 used 3,779,000 ms = 62.9833
minutes; `ceil(62.9833 * 1.20) = 76`.  This is not a fixed SLOC throughput.

**Human gate evidence:** REVIEW records the isolated host fixture, live/deny,
restart/unavailable/malformed/unauthorized, consecutive no-cache checks,
production-SLOC count, and readability.  QA independently repeats acceptance
and SLOC checks and records PASS/FAIL.

**No-compression and scope control:** projected use above 90% of the cumulative
ceiling triggers re-estimation before DEV.  An over-cap candidate follows
`backlog.json`'s exact ordered shedding process and reruns PLAN/QA.  Mandatory
per-sudo live check/no cache cannot be shed; mandatory v1 above 1500 is a
requirement gap, never permission to compress code, errors, names, comments,
or tests.

**Merge rule:** independent REVIEW PASS and QA PASS are required.  Any FAIL
returns to its responsible gate and never merges.
