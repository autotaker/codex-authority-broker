# TASK-0008: Restricted non-force push and token custody

**Depends on:** TASK-0007.

**Status:** planned and executable.

**Preflight prerequisites:** TASK-0007 is merged; an isolated local bare-repo
fixture, fake-token fixture, and credential-injection design are approved;
required tools and permissions are available.  The execution clock starts
only after preflight.  A failed prerequisite is `not_started` and excluded
from measured execution time.

**Owns:** exact local repository/ref/clean-tree validation, GitHub App token
custody, and one system-Git single-ref non-force push path.

**Acceptance:** only the configured repository and `main` or
`task/TASK-*` branch single-ref non-force update may proceed with a live lease.
Wrong repository/ref, dirty tree, force, tag, delete, multiple ref, expired
authority, token leakage, or ambiguous Git/transport state denies without a
force retry.  Tokens are absent from argv, environment, logs, output, errors,
and credential-helper storage.

**Excludes:** sudo changes, rich audit, release/attestation, installer,
packaging, and canary work.

**Production LOC ceiling:** cumulative production SLOC **<=1400**, a 450-SLOC
reserve above TASK-0007's ceiling.  This reserve is a security/scope cap, not
a schedule forecast.

**Initial execution-time evidence:** minimum `unknown`, upper bound `unknown`.
No completed transport boundary is comparable enough to support a duration.
Preflight and PLAN must decompose the fixture before the execution clock; time
must not be inferred from SLOC throughput.

**Human gate evidence:** REVIEW records exact repo/ref/tree validation,
single-ref non-force behavior, all denial cases, live-lease use, token custody
capture scans, no force retry, SLOC, and readability.  QA independently
repeats acceptance, redaction, and SLOC checks and records PASS/FAIL.

**No-compression and scope control:** projected use above 90% of the cumulative
ceiling triggers re-estimation before DEV.  Over-cap work follows the exact
ordered shedding process in `backlog.json` and reruns PLAN/QA.  Secret
non-disclosure and normal non-force rejection cannot be shed; mandatory v1
above 1500 is a requirement gap, never permission to compress.

**Merge rule:** independent REVIEW PASS and QA PASS are required.  Any FAIL
returns to its responsible gate and never merges.
