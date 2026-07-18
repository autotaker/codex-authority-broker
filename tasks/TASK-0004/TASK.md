# TASK-0004: Versioned SO_PEERCRED IPC

**Depends on:** TASK-0003.

**Preflight prerequisites:** prior dependency PRs are merged; required tools and permissions are available; the focused fixture is prepared; any required remote or clean Ubuntu VM is ready. The 30-minute execution clock starts only after these conditions are confirmed. A preflight failure means the task has not started.

**Owns:** versioned Unix IPC, socket lifecycle, and SO_PEERCRED authorization.

**Acceptance:** malformed, wrong-version, oversize, unavailable, or unauthorized-peer requests fail closed.

**Excludes:** Codex access/redaction, sudo, push, audit, and delivery.

**Production LOC ceiling:** cumulative production SLOC after merge is **<=650**.

**Human gate evidence:** REVIEW records IPC parser and peer-credential evidence, the production-SLOC count against <=650, and no-compression/guardrail compliance. QA independently repeats the focused validation and SLOC count, checks the acceptance and boundary, and records PASS/FAIL evidence.

**No-compression and scope control:** count nonblank/non-comment executable-source SLOC only. REVIEW rejects semicolon/one-line packing, collapsed error handling, cryptic names, removal of security comments, and functions combined solely to meet LOC; gofmt and normal idiomatic structure are mandatory. At projected >90% re-estimate before DEV. An over-cap candidate follows the exact backlog.json delivery_guardrails ordered_feature_shedding order and reruns PLAN/QA before DEV continues. Mandatory-v1 controls cannot be shed; if they exceed 1500, stop as a requirement gap, never compress.

**Merge rule:** independent REVIEW PASS and QA PASS are both required; a FAIL returns to its responsible gate and never merges.
