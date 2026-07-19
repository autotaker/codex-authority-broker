# QA RESULT — TASK-0006

## Decision: PASS

Independent QA began after REVIEW revision 2 PASS and independently passed
Q6-01 through Q6-12.

- The current append-only log still filters to the same 72 selected events,
  counts 3/7/38/24, four completed Tasks, and stable filtered hash.
- Required metrics, null reasons, provenance, stage/cycle arithmetic, retries
  0/1/4/1, raw/effective correction semantics, SLOC, and test values reproduce.
- TASK-0001/0003 cycle fields are null; TASK-0004/0005 cycle and stage values
  reproduce without double counting.
- The 20% contingency recomputes to 76 minutes; no fixed SLOC throughput or
  numeric value inferred from null is present.
- Only TASK-0007/0008 plus immediate zero-SLOC TASK-0009 are newly executable.
  Task/index paths, dependencies, caps, later reserves, global 1500 ceiling,
  mandatory controls, no-compression rule, and exact shedding order match.
- Production SLOC is 751 and TASK-0006 adds zero. JSON and diff checks pass.

QA timing was `active_ms=182582`, `wait_ms=0`, retries 0. There is no open
classification or finding. QA changed no file, Git state, or Lap30 log.

## Revision 8.1 independent post-REVIEW QA — PASS

### Decision and classification

**PASS.**  No open candidate, planning, QA-plan, requirement, or regression
finding remains.  The sole retry was a sandbox-only Unix-domain-socket
restriction and is classified **`environment_issue`**, not a DEV or candidate
failure: the identical focused suite passed when rerun in a socket-capable
environment, as did the required full suite.

### Frozen canonical snapshot

QA froze and validated the complete available append-only stream, rather than
treating the historical 136-row Revision 8 prefix as an invariant:

- rows: `153`;
- last `event_id`: `task0006-r3-lap03-qa-start`;
- SHA-256: `224aa37d12e8891c7139277eafcf2489386071801e2a6ff5a169862bbc0fb689`.

`jq -e .` passed, event IDs were unique, and all 10 correction edges passed
the required target-exists, earlier-file-order, same-Task, same-Lap, and
smaller-sequence predicate.  The Revision 8.1 historical-prefix provenance
also reproduced exactly: 136 rows ending at
`task0006-r3-lap02-replan8-start`, SHA-256
`9b01bb157c666aba691a87665bbdef5fb06ae6ba7359198e2e0fda7ceb3dbfc1`.

### Independent evidence

- DEV candidate scope is exactly the seven approved outputs: `backlog.json`
  and `TASK.md` for TASK-0007 through TASK-0012.  The changed PLAN,
  QA_PLAN, and REVIEW_RESULT are gate artifacts; no product source or Lap log
  changed.
- Each of the six JSON metadata blocks in its Task contract is exactly equal
  to its `backlog.json` entry.  IDs are unique; dependencies are the linear
  TASK-0006 -> 0007 -> 0008 -> 0009 -> 0010 -> 0011 -> 0012 chain; every
  contract contains its exclusive paths, fixture/elevation condition, two Lap
  work, focused/repository-native checks, and split/stop rule.
- Arithmetic passes: `751 + 90 + 120 + 0 + 130 + 220 + 0 = 1311`; target and
  hard reserves are respectively `189` (to 1500) and `489` (to 1800).  The
  two measurement contracts have zero product paths and zero expected SLOC;
  later milestones are non-executable.
- The current nonblank/non-comment tracked non-test Go count is `751`.
  Production Go diff is empty, `gofmt -l` is empty, `git diff --check` passes,
  and `jq -e . backlog.json` passes.  The PLAN retains the no-compression,
  mandatory-control, exact shedding-order, independent-role, and
  main-owned-Git constraints.
- Focused check: `GOCACHE=$(mktemp -d) go test ./cmd/codex-authority
  ./internal/ipc ./internal/lease` initially failed only because Unix sockets
  were blocked by the sandbox, then passed unchanged in the capable rerun.
  Required full check: `GOCACHE=$(mktemp -d) go test ./...` passed once in the
  same capable environment (all repository packages).

QA timing: `active_ms≈350000`, `wait_ms≈19000`, `retries=1` (the classified
environment rerun only).  QA changed only this evidence file; it did not edit
the candidate documents, PLAN, QA_PLAN, REVIEW_RESULT, product source,
operational repository/Lap log, or Git state.
