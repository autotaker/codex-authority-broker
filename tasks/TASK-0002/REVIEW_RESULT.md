# REVIEW RESULT — TASK-0002

## Decision: FAIL

The proposed index is parseable and its dependency graph is acyclic, but the
contract does not meet its required exact-one-owner lineage or per-task
contract completeness.  It must return to the planning gate; it is not ready
for QA or merge.

## Findings

### Critical — historical requirements and PR findings do not have one downstream owner

`backlog.yaml` assigns `secret_non_disclosure` only to TASK-0004.  The original
TASK-0001 P0-19 requires non-disclosure also from HTTP diagnostics, packages,
SBOMs, checksums, and release artifacts.  TASK-0004 expressly excludes GitHub
HTTP token transport and release-artifact scans; TASK-0007 owns HTTP transport
redaction, TASK-0009 owns integration capture scans, and TASK-0010 owns package
and artifact secret scans.  Thus that one original acceptance is split among
four tasks while the lineage map claims a single owner.

The same defect affects PR #1 finding 1: TASK-0006 owns local push policy and
explicitly excludes transport/race behavior, while TASK-0007 owns the GitHub
App conditional transport and race behavior.  PR #1 reported those as one
finding (“Restricted GitHub push is not implemented”), so it too has multiple
downstream owners.  This contradicts TASK-0002’s acceptance criterion and
Q2-05/Q2-06.  The task contracts need a non-overlapping ownership model that
assigns each original requirement/finding exactly once, with other tasks named
only as prerequisites or evidence consumers.

### Major — TASK-0001 lacks the required dependency declaration

Q2-03 requires every implementation task to state its dependency/dependencies
and match `backlog.yaml`.  TASK-0001 has no `Depends on: none` (or equivalent)
statement.  The empty dependency is present only in the index, so the task
contract is incomplete.

### Major — several slices are not credibly bounded to a complete 30-minute gate cycle

TASK-0009 combines integration harnesses for socket peers, PAM/sudo, GitHub
conditional-push races, process groups, broad capture/secret scanning, full Go
tests, vet, and `make check`.  TASK-0010 combines Debian packaging, installed
runtime files, a pinned release workflow, checksums, provenance, artifact
secret scans, and three operational documents.  Those are multi-boundary
deliverables rather than narrow follow-up slices, especially against the stated
14-minute / 932-line foundation baseline.  Their exclusions prevent overlap
with adjacent tasks but do not make the owned work realistically reviewable,
QAable, committed, pushed, and merged in one 30-minute cycle.

### Environment/tooling blocker — required `make check` could not run

`make check` returned `make: *** No rule to make target 'check'.  Stop.`  The
candidate contains no `Makefile`, `makefile`, or `GNUmakefile`.  This is not
product-code evidence and should be classified as an environment/repository
contract issue before attribution, but it prevents recording the required
check as PASS.

## Positive evidence

- Parsed `backlog.yaml` with Python 3 / PyYAML 6.0.1: required top-level keys,
  ten expected IDs, unique IDs, 30-minute estimates, resolvable dependencies,
  and unique lineage-map values passed.
- Directed dependency check passed: the graph is acyclic and TASK-0011 is its
  only terminal node.
- Text inspection confirms TASK-0003 through TASK-0011 each declare a
  dependency, owned boundary, acceptance, exclusions, focused checks, and a
  REVIEW PASS + QA PASS / never-merge-on-FAIL rule.  TASK-0001 has the same
  merge rule but lacks its explicit empty dependency declaration.
- `git diff --check` passed.  `git status --short`, the tracked diff, and
  untracked-path inventory show planning artifacts only: amended
  `tasks/TASK-0001/TASK.md`, `backlog.yaml`, and task contracts/plans.  No
  product source, packaging, workflow, secret, or Git-state change was found.

## Historical evidence reviewed

Reviewed TASK-0001 revision 1 TASK/PLAN/QA_PLAN, PR #1 `REVIEW_RESULT.md`, and
`HANDOVER.md` from `task/TASK-0001-totp-authority-lease`, plus the amended
TASK-0001 and TASK-0003 through TASK-0011 contracts.  PR #1 remains correctly
described as FAIL historical evidence; this review does not reclassify it.

No product files or Git state were changed.  This review wrote only this
evidence file.

---

## Revision 2 independent review (revision 3 candidate)

## Decision: FAIL

The revised eight-task index is valid JSON, has the required eight IDs, a
resolvable acyclic dependency graph, one terminal (`TASK-0009`), fourteen
valid atomic lineage keys, the required cumulative-cap sequence, and the
canonical seven-item shedding order.  `make check`, `python3 -m py_compile
scripts/check-task-planning.py`, `python3 -m json.tool backlog.json`, and
`git diff --check` all passed.

It still cannot pass the independent planning-review gate.  The checker does
not enforce several controls which TASK-0002 and its QA plan expressly assign
to it, and the proposed slices are not credible as full PLAN -> DEV -> REVIEW
-> QA -> PR-merge cycles in thirty minutes under normal, non-compressed
implementation and verification.

## Disposition of revision-1 findings

| Revision-1 finding | Disposition | Evidence |
| --- | --- | --- |
| Exact-one-owner lineage split historic secret/push requirements across multiple old tasks | Resolved for the new planning model | `backlog.json` has 14 atomic lower-snake-case keys, each mapping to one of the eight current contracts.  Local push validation and observed-OID/transport behavior are separate atomic keys owned by TASK-0005 and TASK-0006, respectively; the former TASK-0007/TASK-0009/TASK-0010 ownership arrangement is retired.  This does not change PR #1's historical FAIL. |
| TASK-0001 omitted its empty dependency declaration | Resolved | `tasks/TASK-0001/TASK.md` now declares exactly `**Depends on:** none.`, and the checker verifies this declaration. |
| Old TASK-0009/TASK-0010 were not credible thirty-minute slices | Persists under the replacement plan | Those task IDs are retired, but equivalent multi-boundary work remains concentrated in TASK-0006, TASK-0008, and TASK-0009; see the major finding below. |
| `make check` had no target | Resolved | `Makefile` supplies `check`; it ran successfully and invokes the planning checker plus `git diff --check`. |

## New findings

### Major — checker acceptance is materially incomplete

The 133-physical-line, standard-library `scripts/check-task-planning.py` is
within its line limit and parses/validates the current happy path, but it does
not implement the checker contract stated in TASK-0002/PLAN and QA_PLAN:

- Lines 92–98 accept any 10–15 lineage entries.  QA_PLAN Q2-04 requires the
  candidate's exactly 14 entries; a 10-entry candidate would pass.
- Lines 115–124 require canonical shedding strings only in PLAN and require
  only five labels plus a weak merge-rule substring check in each contract.
  They do not require each contract's production-SLOC definition,
  no-compression review criteria, `>90%` re-estimation trigger, PLAN/QA
  reapproval, mandatory controls, or requirement-gap stop, despite the
  published checker contract.
- The SLOC counter (lines 27–44) does not exclude a `tests/` directory, so a
  shipped-extension test such as `tests/integration/fixture.sh` is counted
  contrary to the documented exclusion for tests.  It also cannot establish
  semantic non-compression; that is appropriately a human REVIEW duty, but
  the checker must at least validate the required textual controls.

Consequently `make check` PASS is not sufficient evidence that the candidate
meets TASK-0002's promised planning enforcement.  Return this to TASK-0002
DEV/PLAN to complete the checker contract and re-run independent REVIEW/QA.

### Major — thirty-minute full-cycle and <=1500-SLOC claim is not credible

The cap arithmetic is correct (700, 880, 1000, 1150, 1400, 1430, 1500,
1500; final <=1500), but a cap is not an implementation estimate.  The
following scope remains too large for a complete implementation, independent
review, QA, PR merge, and retained evidence in a thirty-minute cycle without
either compressed code or deferred required behavior:

| Task | Assessment |
| --- | --- |
| TASK-0001 | Not credible: new fail-closed readiness/TOTP/replay/rate/concurrency/monotonic-restart state and versioned SO_PEERCRED IPC have a 700-SLOC cumulative budget and multiple security boundary tests. |
| TASK-0003 | Borderline but dependent on a complete TASK-0001; CLI/socket group-mode and redaction boundary can only be credible after a narrower interface/test fixture is specified. |
| TASK-0004 | Not credible as written: root helper, PAM/sudoers policy, live IPC failure modes, and an isolated sudo fixture need privileged-environment verification. |
| TASK-0005 | Not credible as written: canonical worktree, ref grammar, clean-tree/history/size policy, and allow/deny repository fixtures exceed a narrow follow-up gate. |
| TASK-0006 | Not credible: GitHub App JWT/installation token custody, root pipe/askpass, fixed system-git invocation, observed-OID race and redaction tests are multiple security-sensitive boundaries in one 250-SLOC increment. |
| TASK-0007 | Not credible: a three-boundary integration matrix and external-trace-compatible audit verification are allocated only 30 cumulative SLOC. |
| TASK-0008 | Not credible: systemd/PAM/sudoers bundle, hosted release workflow, source-free tarball, checksum/provenance, and install/rollback documentation require release-system evidence, even if most files are excluded from production SLOC. |
| TASK-0009 | Not credible: a clean Ubuntu artifact-only install, service/CLI/sudo/push smoke test, rollback, redaction, and failure classification is an external canary operation, not a 30-minute zero-production-LOC merge slice. |

The final <=1500 number therefore remains a numerical ceiling only, not a
credible non-compressed production estimate.  The mandatory-v1 requirement
must be re-estimated and decomposed into independently reviewable gates; if it
cannot fit, the documented requirement-gap stop applies rather than code
compression or schedule-driven shedding.

## Positive revision-2 evidence

- `backlog.json` parsed independently with `python3 -m json.tool`.
- Independent topological sort produced TASK-0001, TASK-0003, TASK-0004,
  TASK-0005, TASK-0006, TASK-0007, TASK-0008, TASK-0009; TASK-0009 is the
  sole terminal.
- The 14 lineage keys are lower-snake-case and map as follows by owner:
  TASK-0001=4, TASK-0003=2, TASK-0004=1, TASK-0005=1, TASK-0006=2,
  TASK-0007=2, TASK-0008=1, TASK-0009=1.  Every contract has exact matching
  dependency text, required labels, cap, focused checks, and merge rule.
- Title/boundary simplifications are present: no MCP server (0003), no custom
  C PAM module (0004), no in-process smart HTTP (0006), no process tracking/
  signalling/killing (0007), no Debian package (0008), and zero added
  production LOC (0009).
- The JSON and PLAN contain the exact ordered feature-shedding strings and
  identify all non-sheddable controls.  Contracts contain the no-compression,
  >90% re-estimation, PLAN/QA reapproval, and requirement-gap language, but
  the checker does not verify that fact.
- Read via `git show task/TASK-0001-totp-authority-lease:...`: original
  TASK-0001 PLAN/QA_PLAN, PR #1 REVIEW_RESULT, and HANDOVER.  The historical
  decision remains FAIL and the current TASK-0001 calls it historical only;
  it is neither removed nor reclassified.
- Scope inspection found only the amended TASK-0001 contract and untracked
  planning/checker/Makefile artifacts.  No authority-broker product source,
  secret, generated release output, staging, commit, merge, or `.git` change
  was found.  `git diff --check` passed.

No product files or Git state were changed.  This revision appended only this
review evidence file.

---

## Revision 3 independent review (rolling-wave revision 4)

## Decision: FAIL

The current planning data itself has the required rolling-wave shape: exactly
five executable tasks (`TASK-0001`, `TASK-0003`–`TASK-0006`) form the sole
30-minute dependency chain; only the first four are implementation slices;
`TASK-0006` is the zero-production-SLOC remeasurement/replanning gate; and
there are exactly four explicitly ineligible reserves. The structured
measurement fields, preflight exclusion/not-started rule, 20% contingency,
sparse-evidence rule, <=1500 cumulative cap, exact seven-item shedding order,
and mandatory-v1/no-compression controls are present. The historical PR #1
FAIL remains unchanged.

The candidate nevertheless fails this independent gate because its required
small planning checker does not actually enforce the prohibition on a detailed
future task contract. That is a material contract-enforcement omission, so
TASK-0002 must return to DEV/PLAN before QA or merge.

### Major — checker accepts a forbidden TASK-0007+ detailed contract

Q2-04 requires a negative fixture for a `TASK-0007+` contract and says that
no such contract is present or accepted. Although the current worktree has
only `tasks/TASK-0001` through `tasks/TASK-0006`,
`scripts/check-task-planning.py` only verifies the five required contracts
(lines 130–140) and reserve-named directories (line 100). It never enumerates
`tasks/TASK-*` to reject an additional detailed contract.

Independent negative-fixture evidence: I copied the candidate to a temporary
directory, added an otherwise empty `tasks/TASK-0007/TASK.md`, and ran
`python3 <temporary-copy>/scripts/check-task-planning.py`. It returned:

```
planning check: PASS (production SLOC 0, active cap 1500)
semantic compression checks are owned by REVIEW
```

An empty later contract is sufficient to violate the first-wave-only planning
boundary. The checker must fail any `tasks/TASK-0007` or later detailed
contract (and a corresponding forbidden index reference) before its PASS can
serve as the requested enforcement evidence.

## Positive evidence

- `make check` passed: byte compilation, the 146-physical-line
  standard-library checker, and `git diff --check` all passed. Independent
  `python3 -m json.tool backlog.json` also passed.
- The checker enforces the exact five executable IDs, exact linear dependency
  chain, 30-minute estimates, caps `250, 430, 650, 820, 820`, sole terminal
  TASK-0006, its zero added production SLOC, four exact reserves, the ten
  measurement fields, 20%/preflight/sparse-evidence rules, guardrails, and the
  contract labels/dependencies/caps/merge rule.
- Direct inspection confirmed the current tree contains only TASK-0001 through
  TASK-0006 directories. The four reserves are non-executable and ineligible
  for PLAN/DEV/branch/PR; TASK-0006 alone may convert the next 2–3 after its
  independent REVIEW PASS, QA PASS, and PR merge, then insert another
  measurement gate.
- All five contracts specify preflight, bounded ownership/exclusions, a
  cumulative production-SLOC ceiling, focused checks, no-compression controls,
  and independent REVIEW PASS plus QA PASS before merge. The measurement
  contract separately records PLAN, DEV, REVIEW, QA, and CI/push/merge times.
- Historical inspection of `task/TASK-0001-totp-authority-lease` confirms PR
  #1's `REVIEW_RESULT.md` remains `FAIL`; the current artifacts preserve it as
  historical evidence rather than reclassifying it.
- Scope/history inspection found the amended TASK-0001 contract plus
  untracked planning/index/checker/Makefile artifacts only; no authority-broker
  production source, secret, release artifact, staging, commit, merge, or
  `.git` mutation was found. No product files or Git state were changed by
  this review.

This revision appended only this review evidence file.

---

## Revision 3 addendum — checker readability and non-compression

### Major — the checker is line-compressed to meet its self-imposed limit

The <=150-line requirement is expressly for a **readable**, standard-library
checker; it is not part of the <=1500 production-SLOC cap. The 146-line script
does use only the standard library, but it is not reasonably readable in its
current compressed form. Examples include semicolon-packed independent
statements and compound control flow on lines 19, 45, 53, 59, 65, 69, 71–72,
79, 94, 110, 133, 136, and 140–144. Line 53, for example, combines diagnostic
output and a return; lines 65 and 94 fold type discrimination, iteration, and
data extraction into one expression; and line 140 combines three distinct
verification operations.

This is precisely the sort of line packing the planning guardrails reject for
production source, and the separate checker requirement adds readability as a
direct acceptance criterion. The numerical 150-line ceiling must not cause
non-idiomatic compression. Split the logic into normal statements/functions
even if that requires revisiting the arbitrary ceiling; retain the checker’s
standard-library and contract-validation scope.

---

## Revision 4 independent review (checklist-only revision 5)

## Decision: PASS

Independent manual review of `backlog.json`, TASK-0002 TASK/PLAN/QA_PLAN, the
five executable task contracts, filesystem inventory, historical PR #1 review,
and candidate status/diff passes. This revision deliberately replaces the
previous planning checker with human REVIEW and QA checklists; no checker,
`Makefile`, or current planning-automation reference remains outside this
immutable historical review evidence. No `make check` was required or run for
this planning-only, checklist-based revision.

### Checklist evidence

- `backlog.json` parsed successfully with Node JSON parsing. Its only
  executable tasks are TASK-0001, TASK-0003, TASK-0004, TASK-0005, and
  TASK-0006. Independent dependency traversal yields exactly
  `TASK-0001 -> TASK-0003 -> TASK-0004 -> TASK-0005 -> TASK-0006`; the
  30-minute estimates and cumulative ceilings are exactly 250, 430, 650, 820,
  and 820. TASK-0006 explicitly adds zero production SLOC.
- The five contracts have bounded, non-overlapping ownership, common
  preflight/not-started rule, focused acceptance, exclusions, cap, manual gate
  evidence, no-compression controls, and independent REVIEW PASS plus QA PASS
  merge rule. TASK-0006 alone permits conversion of only the next 2–3
  evidence-supported reserves after its REVIEW PASS, QA PASS, and PR merge,
  with a new zero-SLOC remeasurement gate after them.
- There are exactly four reserves (`sudo` <=950, `push` <=1400,
  `audit-release` <=1500, and zero-added-SLOC `clean-canary` <=1500). Each is
  explicitly non-executable and ineligible for PLAN/DEV/branch/PR. Filesystem
  inspection found task directories only through TASK-0006 and no TASK-0007+
  detailed contract.
- The measurement contract contains all ten required fields: planned/actual
  production SLOC, test LOC, PLAN/DEV/REVIEW/QA/CI-push-merge minutes,
  retries, and failure classifications. It excludes failed preflight as
  not-started, applies 20% contingency, uses observed cycles, makes no fixed
  throughput claim under sparse evidence, and restricts conversion to 2–3
  items.
- Manual production-source inventory found no non-test shipped/runtime `.go`,
  `.py`, `.sh`, or `.c` files, so the current qualifying production SLOC
  baseline is 0. The documented definition excludes tests, planning/evidence,
  workflows, declarative configuration, generated, and vendor; the global
  ceiling remains <=1500. Index and PLAN retain the exact seven-item shedding
  order, >90% re-estimation/reapproved PLAN+QA rule, no-compression rule, and
  mandatory-v1 requirement-gap stop.
- The nine lineage entries map only to the five first-wave owners. Historical
  `task/TASK-0001-totp-authority-lease:tasks/TASK-0001/REVIEW_RESULT.md`
  remains FAIL; current planning artifacts identify it only as historical and
  do not downgrade or reclassify it.
- Candidate inspection: `git diff --check` exited 0. `git status --short`
  lists the amended TASK-0001 contract and untracked planning/index/contract
  files only; there are no production source, secret, generated release,
  staging, commit, merge, or `.git` mutations attributable to this review.

### Disposition of revision-3 findings

- **Forbidden TASK-0007+ contract accepted by the checker:** resolved by the
  approved human-checklist design. QA_PLAN Q2-03 and Q2-04 require direct
  filesystem and reserve inspection; this review found no such contract. The
  checker and its enforcement claim are absent from the current planning
  surface.
- **Checker readability/non-compression finding:** superseded by removal of
  the checker under the approved checklist-only revision. The applicable
  production no-compression controls remain explicit in the index, PLAN, and
  all five executable contracts and are owned by human REVIEW.

No product files or Git state were changed. This revision appended only this
review evidence file.

---

## Revision 5 independent re-review (QA_PLAN revision 7)

## Decision: PASS

The QA planning defect recorded in `QA_RESULT.md` is resolved exactly. Q2-10
now says **two TOTP entries**, so its owner distribution is
`2 + 2 + 2 + 2 + 1 = 9`, matching all nine `backlog.json`
`acceptance_lineage` entries: TASK-0001=2, TASK-0003=2, TASK-0004=2,
TASK-0005=2, and TASK-0006=1. The two TASK-0003 entries are valid-unused
activation and replay/rate-limit/concurrency denial, consistent with the task
contract. No lineage entry points to a reserve or later task.

Independent current-state reinspection also confirmed the prior revision-4
PASS evidence remains valid: JSON parses; the sole executable DAG is
TASK-0001 -> TASK-0003 -> TASK-0004 -> TASK-0005 -> TASK-0006; all estimates
are 30 minutes; cumulative ceilings are 250, 430, 650, 820, 820; the four
reserves remain non-executable at 950, 1400, 1500, and 1500; there is no
TASK-0007+ contract, production-source file, planning checker, or Makefile
artifact; and the historical PR #1 FAIL and the checklist-only gate boundary
remain preserved.

`git diff --check` exited 0. `git status --short` still shows only the amended
TASK-0001 planning contract and untracked TASK-0002 planning/evidence plus
TASK-0003 through TASK-0006 contracts and `backlog.json`; no product or Git
state mutation was made by this review. The prior QA FAIL remains immutable
historical evidence and correctly routes to renewed independent QA after this
PASS.

No `make check` was run under the approved planning-only human-checklist
exception. This revision appended only this review evidence file.
