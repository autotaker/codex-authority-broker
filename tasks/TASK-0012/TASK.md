# TASK-0012: final zero-SLOC measurement and later-reserve gate

**Depends on:** TASK-0009 (merged and PASS).

**Status:** planned and executable after TASK-0009; no push implementation is
enabled.

## Replanned contract metadata

```json
{
  "id": "TASK-0012",
  "title": "final zero-SLOC measurement and later-reserve gate",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0009"],
  "baseline_production_sloc": 1253,
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1253,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1495,
  "hard_cumulative_guard": 1800,
  "production_paths": [],
  "test_paths": ["tasks/TASK-0012/MEASUREMENT.md"],
  "entrypoint": null,
  "fixture_elevation_needs": "Read-only frozen TASK-0009/v1 event snapshot; no elevation, network, product fixture, or operational-log write.",
  "lap_1": "After TASK-0009 PASS+merge, freeze the v1 snapshot and regenerate historical/product SLOC, test LOC, stage, cycle, active/wait, retry, null, and raw/effective correction evidence; reconcile 1253 actual against target 1500 and hard 1800.",
  "lap_2": "Independent REVIEW and QA each regenerate the evidence and run the canonical correction and repository-native full checks; prove TASK-0010/TASK-0011 remain non-executable v2 reserve and no later milestone receives DEV detail; main owns Git.",
  "exclusions": ["all product/test implementation", "TASK-0010/TASK-0011 push implementation", "audit", "attestation", "release", "installer", "canary", "editing canonical log", "enabling later DEV"],
  "split_stop_rule": "Stop on canonical defects, unexplained cap drift, actual above 1500, any hard-limit risk, missing independent regeneration, or inability to close REVIEW/QA in Lap 2; later reserve remains blocked.",
  "measurement_lineage": "Apply TASK-0009's exact correction predicate and raw/effective provenance to the frozen v1 stream; compare measured 1253 baseline and zero delta without fixed throughput or reserve borrowing.",
  "later_reserve_eligibility": "Only after independent REVIEW PASS, independent QA PASS, and main merge may MILESTONE-audit-attestation then MILESTONE-manual-canary-rollback receive PLAN; both remain non-executable beforehand.",
  "replan_reason": "TASK-0009 measured 1253. Former TASK-0010/TASK-0011 forecasts totalled 283 and would reach 1536, 36 over target; ordered shedding item 7 moves both coupled push Tasks to v2, so this gate measures the direct zero-SLOC v1 wave.",
  "contract_path": "tasks/TASK-0012/TASK.md"
}
```

## Purpose and explicit downstream decision

TASK-0009 measured **1253 actual merged production SLOC** and adds zero.
Former TASK-0010 (+130) and TASK-0011 (+153) forecasts would reach 1536,
which is 36 above the mandatory-v1 target 1500. The unconditional hard limit
1800 is not permission to exceed the target. Ordered shedding items 1–6 have
no remaining applicable optional scope in the coupled push work, so item 7 is
selected: **GitHub push moves to v2**. TASK-0010 and TASK-0011 are explicitly
deferred, non-executable reserves; they cannot provide branch, DEV, PR-ready,
or production detail in this v1 wave.

This Task is consequently the direct final zero-SLOC measurement gate after
TASK-0009. It records the v1 baseline as `1253 + 0 = 1253`, preserves the
mandatory TOTP/full-sudo, readiness, fail-closed IPC, live/no-cache sudo,
secret non-disclosure, readability, and independent-gate boundaries, and
keeps all later audit/attestation/manual-canary work blocked until this gate
passes and main merges.

## Preflight and two-Lap evidence

Preflight is read-only: freeze the completed TASK-0009/v1 event snapshot,
verify its SHA-256, 389-line count, JSON parse, unique IDs, and the exact
same-task/lap/earlier-file/smaller-sequence correction predicate. A failed
preflight is `not_started`, excluded from timing, and never represented as
zero.

Lap 1 regenerates historical and v1 SLOC/test/time/active/wait/retry/null and
raw/effective classification evidence. Stage timing pairs only same-task,
same-lap, same-stage, same-attempt starts and terminals; `preflight` is
excluded; active and wait remain separate; retries are maximum propagated
values; and only observable non-preflight time receives
`ceil(observed_ms * 1.20)`.

Lap 2 is independent REVIEW and QA. Each role regenerates the source-ID,
correction, cohort/terminal, SLOC/cap, stage/null, active/wait, retry, and
contingency arithmetic and runs:

```sh
GOCACHE="$(mktemp -d)" go test ./...
test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"
git diff --check
jq -e . backlog.json >/dev/null
```

QA must prove that TASK-0010 and TASK-0011 remain `deferred-v2` and
`executable:false`, that no later milestone receives PLAN/DEV detail, and that
the measured 1253 baseline is below target and hard limits. Main owns final
checks, Git, and merge.

## Acceptance, exclusions, and stop rule

- Exactly zero production SLOC is added: `1253 + 0 = 1253`.
- Historical performance and product-terminal cohorts remain separate; no
  partial, correction, governance, terminated, unfinished, or `pass`-status
  TASK-0013 event becomes a product terminal.
- Null timing retains an explicit reason; no stage, active/wait, retry,
  classification, contingency, or cap value is inferred from throughput.
- TASK-0010/TASK-0011 v2 reserve status and the push-to-v2 decision are
  explicit, not silently selected; mandatory safety and readable structure are
  preserved.
- Stop on canonical contradiction, non-reproducible arithmetic, actual above
  1500, any hard-limit risk, missing independent regeneration, or failed
  REVIEW/QA. Later reserve remains non-executable until this gate passes and
  merges.
