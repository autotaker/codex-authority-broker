# TASK-0012: final zero-SLOC measurement and later-reserve gate

**Depends on:** TASK-0017 (planned v1 blocker; must pass independent REVIEW
and QA and merge), with the post-TASK-0009 measurement/replan included in that
predecessor's evidence.

**Status:** planned and executable only after TASK-0017 passes independent
REVIEW and QA and merges; no push implementation is enabled.

## Replanned contract metadata

```json
{
  "id": "TASK-0012",
  "title": "final zero-SLOC measurement and later-reserve gate",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0017"],
  "baseline_production_sloc": 1308,
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1308,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1495,
  "hard_cumulative_guard": 1800,
  "production_paths": [],
  "test_paths": ["tasks/TASK-0012/MEASUREMENT.md"],
  "entrypoint": null,
  "fixture_elevation_needs": "Read-only frozen post-TASK-0017 v1 event snapshot; no elevation, network, product fixture, or operational-log write.",
  "lap_1": "After TASK-0017 PASS+merge, freeze the post-TASK-0017 v1 snapshot and regenerate historical/product SLOC, test LOC, stage, cycle, active/wait, retry, null, and raw/effective correction evidence; record TASK-0017's actual production delta rather than substituting its +55 forecast, then reconcile the actual cumulative value against target 1500 and hard 1800.",
  "lap_2": "Independent REVIEW and QA each regenerate the frozen post-TASK-0017 actual evidence and run the canonical correction and repository-native full checks; prove TASK-0010/TASK-0011 remain non-executable v2 reserve and no later milestone receives DEV detail; main owns Git.",
  "exclusions": ["all product/test implementation", "TASK-0010/TASK-0011 push implementation", "audit", "attestation", "release", "installer", "canary", "editing canonical log", "enabling later DEV"],
  "split_stop_rule": "Stop on canonical defects, unexplained cap drift, actual above 1500, any hard-limit risk, missing independent regeneration, or inability to close REVIEW/QA in Lap 2; later reserve remains blocked.",
  "measurement_lineage": "The pre-TASK-0017 merged baseline is 1253 and the planned TASK-0017 forecast is +55 (projected 1308), but both are planning evidence until that Task passes and merges. Then apply TASK-0009's exact correction predicate and raw/effective provenance to a frozen post-TASK-0017 v1 stream and measure the actual cumulative baseline plus zero Task-0012 delta without fixed throughput, forecast substitution, or reserve borrowing.",
  "later_reserve_eligibility": "Only after independent REVIEW PASS, independent QA PASS, and main merge may MILESTONE-audit-attestation then MILESTONE-manual-canary-rollback receive PLAN; both remain non-executable beforehand.",
  "replan_reason": "TASK-0009 measured 1253 and TASK-0017 is the planned v1 blocker with a readable +55 forecast (projected 1308); after its PASS+merge this gate must freeze and measure actual post-TASK-0017 production SLOC. The user-confirmed v1 push drop keeps former TASK-0010/TASK-0011 forecasts (+130/+153) deferred-v2 and zero v1 SLOC, so this remains the direct zero-SLOC v1 gate before later reserve work.",
  "contract_path": "tasks/TASK-0012/TASK.md"
}
```

## Purpose and explicit downstream decision

TASK-0009 measured **1253 actual merged production SLOC** before the planned
TASK-0017 blocker. TASK-0017's readable +55 forecast projects cumulative
production SLOC at 1308, but that forecast is not an implementation or E2E
result and must be replaced by the actual post-merge measurement. Former
TASK-0010 (+130) and TASK-0011 (+153) forecasts would reach 1536, which is 36
above the mandatory-v1 target 1500. The unconditional hard limit 1800 is not
permission to exceed the target. The user-confirmed v1 scope decision drops
GitHub push: TASK-0010 and TASK-0011 remain explicitly deferred-v2,
non-executable reserves with zero v1 production SLOC; they cannot provide
branch, DEV, PR-ready, or production detail in this v1 wave.

This Task is consequently the direct final zero-SLOC measurement gate after
TASK-0017. It records the planned baseline as `1253 + 55 = 1308` only for
forecasting, then freezes and measures the actual post-TASK-0017 baseline
before recording `actual baseline + 0`, and preserves the
mandatory TOTP/full-sudo, readiness, fail-closed IPC, live/no-cache sudo,
secret non-disclosure, readability, and independent-gate boundaries, and
keeps all later audit/attestation/manual-canary work blocked until this gate
passes and main merges.

## Preflight and two-Lap evidence

Preflight is read-only after TASK-0017 PASS+merge: freeze the resulting
post-TASK-0017 v1 event snapshot, verify its SHA-256, line count, JSON parse,
unique IDs, and the exact same-task/lap/earlier-file/smaller-sequence
correction predicate. Preserve TASK-0009's 389-line source as the prior
measurement evidence, but do not treat the +55 TASK-0017 forecast as actual.
A failed preflight is `not_started`, excluded from timing, and never
represented as zero.

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
the frozen post-TASK-0017 actual baseline is below target 1500 and hard 1800.
The prior measured 1253 baseline and +55/+1308 values remain lineage
evidence, not a substitute for the actual. Main owns final checks, Git, and
merge.

## Acceptance, exclusions, and stop rule

- Exactly zero production SLOC is added by TASK-0012: after the frozen
  post-TASK-0017 actual baseline is measured, the gate records `actual + 0`.
  The `1253 + 55 = 1308` value is a forecast only and is never substituted for
  the actual TASK-0017 delta.
- Historical performance and product-terminal cohorts remain separate; no
  partial, correction, governance, terminated, unfinished, or `pass`-status
  TASK-0013 event becomes a product terminal.
- Null timing retains an explicit reason; no stage, active/wait, retry,
  classification, contingency, or cap value is inferred from throughput.
- TASK-0010/TASK-0011 v2 reserve status and the user-confirmed v1 push drop are
  explicit, not silently selected; dedicated socket/PAM identity handoff is
  the v1 blocker, and mandatory safety and readable structure are preserved.
- Stop on canonical contradiction, non-reproducible arithmetic, actual above
  1500, any hard-limit risk, missing independent regeneration, or failed
  REVIEW/QA. Later reserve remains non-executable until this gate passes and
  merges.
