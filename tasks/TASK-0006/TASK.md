# TASK-0006: Measurement and rolling-wave replanning gate

**Depends on:** TASK-0005.

**Preflight prerequisites:** prior dependency PRs are merged; required tools and permissions are available; the focused fixture is prepared; any required remote or clean Ubuntu VM is ready. The 30-minute execution clock starts only after these conditions are confirmed. A preflight failure means the task has not started.

**Owns:** measurement evidence and the approved next-wave planning update; it adds no authority-broker production code.

**Acceptance:** for every completed task, evidence records planned and actual production SLOC; test LOC; PLAN, DEV, REVIEW, QA, and CI/push/merge time; retries; and failure classifications. Its approved PLAN derives the next-wave size from observed completed cycles, excludes preflight time, applies a 20% time contingency, and does not assume a fixed SLOC throughput when evidence is sparse. It converts only the next 2–3 future milestones into detailed executable contracts, updates backlog/contracts and cap reserves, and inserts the next measurement gate after those 2–3 tasks.

**Excludes:** implementation of sudo, push, audit, release, or canary behavior; creation of detailed contracts beyond the next 2–3 evidence-supported items.

**Production LOC ceiling:** cumulative production SLOC after merge remains **<=820**; this task adds **0** production SLOC.

**Human gate evidence:** REVIEW records evidence-table completeness, arithmetic/replanning, index/contract consistency, and the 0-added-SLOC ceiling check. QA independently repeats those checks, verifies the acceptance and boundary, and records PASS/FAIL evidence.

**No-compression and scope control:** count nonblank/non-comment executable-source SLOC only. REVIEW rejects semicolon/one-line packing, collapsed error handling, cryptic names, removal of security comments, and functions combined solely to meet LOC; gofmt and normal idiomatic structure are mandatory. At projected >90% re-estimate before DEV. An over-cap candidate follows the exact backlog.json delivery_guardrails ordered_feature_shedding order and reruns PLAN/QA before DEV continues. Mandatory-v1 controls cannot be shed; if they exceed 1500, stop as a requirement gap, never compress.

**Merge rule:** independent REVIEW PASS and QA PASS are both required; a FAIL returns to its responsible gate and never merges.
