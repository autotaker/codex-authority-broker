# TASK-0012: final zero-SLOC measurement and later-reserve gate

**Depends on:** TASK-0017 (merged as PR #18 after independent REVIEW and QA
PASS), with the post-TASK-0009 measurement/replan included in that
predecessor's evidence.

**Status:** in progress as a zero-product-SLOC evidence gate; no push
implementation or later milestone is enabled.

## Replanned contract metadata

```json
{
  "id": "TASK-0012",
  "title": "final zero-SLOC measurement and later-reserve gate",
  "status": "planned",
  "executable": true,
  "depends_on": ["TASK-0017"],
  "baseline_production_sloc": 1407,
  "expected_production_sloc": 0,
  "expected_cumulative_production_sloc": 1407,
  "target_cumulative_cap": 1500,
  "projected_cap_trigger_sloc": 1495,
  "hard_cumulative_guard": 1800,
  "production_paths": [],
  "test_paths": ["tasks/TASK-0012/MEASUREMENT.md"],
  "entrypoint": null,
  "fixture_elevation_needs": "Read-only frozen 404-line canonical JSONL snapshot plus merged Git trees and TASK-0017 REVIEW/QA/PR evidence; no elevation, network, product fixture, or operational-log write.",
  "lap_1": "Freeze the current post-TASK-0017 canonical stream without adding a synthetic terminal; regenerate correction/cohort evidence, full-tree production/test LOC, Git delta, null reasons, and cap arithmetic; reconcile the nine-line historical TASK-0008 undercount and record TASK-0017 timing as null because no canonical terminal exists.",
  "lap_2": "Independent REVIEW and QA each regenerate the frozen correction predicate, full-tree/Git SLOC arithmetic, null provenance, status and reserve guards, then run only affected evidence checks (diff, JSON, scope); main owns Git.",
  "exclusions": ["all product/test implementation", "TASK-0010/TASK-0011 push implementation", "audit", "attestation", "release", "installer", "canary", "editing canonical log", "enabling later DEV"],
  "split_stop_rule": "Stop on canonical defects, unexplained cap drift, actual above 1500, any hard-limit risk, missing independent regeneration, or inability to close REVIEW/QA in Lap 2; later reserve remains blocked.",
  "measurement_lineage": "The frozen canonical stream has 404 records and no TASK-0017 terminal, which must not be synthesized. Merged Git trees prove TASK-0017 net +145 (1262 to 1407 full-tree SLOC). The earlier 1253 ledger plus +145 forecast gave 1398; its nine-line drift is exactly the TASK-0008 sudo helper undercount (47 canonical lines recorded as 38). The reconciled actual is 1407 plus zero TASK-0012 delta, with TASK-0017 timing null and reasoned.",
  "later_reserve_eligibility": "Only after independent REVIEW PASS, independent QA PASS, and main merge may MILESTONE-audit-attestation then MILESTONE-manual-canary-rollback receive PLAN; both remain non-executable beforehand.",
  "replan_reason": "Post-merge full-tree measurement is 1407 rather than forecast 1398 because TASK-0008 historically undercounted its new helper by nine lines. The explained actual remains below target 1500 and hard 1800. User-confirmed TASK-0010/TASK-0011 deferral remains unchanged, so this is still the direct zero-SLOC gate before later mandatory work.",
  "contract_path": "tasks/TASK-0012/TASK.md"
}
```

## Purpose and explicit downstream decision

TASK-0009 recorded **1253 production SLOC** before TASK-0017. TASK-0017's
approved delta is exactly +145, but the post-merge full-tree recount is 1407,
not 1398: the prior ledger undercounted TASK-0008's new sudo helper by nine
lines (`47` canonical lines recorded as `38`). This gate records the explained
full-tree actual rather than preserving that historical arithmetic defect. Former
TASK-0010 (+130) and TASK-0011 (+153) forecasts would reach 1536, which is 36
above the mandatory-v1 target 1500. The unconditional hard limit 1800 is not
permission to exceed the target. The user-confirmed v1 scope decision drops
GitHub push: TASK-0010 and TASK-0011 remain explicitly deferred-v2,
non-executable reserves with zero v1 production SLOC; they cannot provide
branch, DEV, PR-ready, or production detail in this v1 wave.

This Task is consequently the direct final zero-SLOC measurement gate after
TASK-0017. It preserves `1253 + 145 = 1398` only as forecast lineage, then
records the reconciled full-tree actual `1407 + 0 = 1407` and preserves the
mandatory TOTP/full-sudo, readiness, fail-closed IPC, live/no-cache sudo,
secret non-disclosure, readability, and independent-gate boundaries, and
keeps all later audit/attestation/manual-canary work blocked until this gate
passes and main merges.

## Preflight and two-Lap evidence

Preflight freezes the 404-line post-TASK-0017 canonical stream, verifies its
SHA-256, line count, JSON parse,
unique IDs, and the exact same-task/lap/earlier-file/smaller-sequence
correction predicate. Preserve TASK-0009's 389-line source as the prior
measurement evidence. Because the stream has no TASK-0017 terminal and the
Lap30 contract forbids a synthetic completion event, use merged Git/PR/REVIEW/
QA evidence for status and SLOC and retain TASK-0017 timing as null with that
reason.
A failed preflight is `not_started`, excluded from timing, and never
represented as zero.

Lap 1 regenerates historical and v1 SLOC/test/time/active/wait/retry/null and
raw/effective classification evidence. Stage timing pairs only same-task,
same-lap, same-stage, same-attempt starts and terminals; `preflight` is
excluded; active and wait remain separate; retries are maximum propagated
values; and only observable non-preflight time receives
`ceil(observed_ms * 1.20)`.

Lap 2 is independent REVIEW and QA. Each role regenerates the source-ID,
correction, cohort, full-tree/Git SLOC, null/status, and cap arithmetic and
runs only the affected evidence checks:

```sh
git diff --check
jq -e . backlog.json >/dev/null
```

QA must prove that TASK-0010 and TASK-0011 remain `deferred-v2` and
`executable:false`, that no later milestone receives PLAN/DEV detail, and that
the frozen post-TASK-0017 actual baseline is below target 1500 and hard 1800.
The prior measured 1253 baseline and +145/+1398 values remain lineage
evidence, not a substitute for the actual. Main owns final checks, Git, and
merge.

## Acceptance, exclusions, and stop rule

- Exactly zero production SLOC is added by TASK-0012: after the frozen
  post-TASK-0017 actual baseline is measured, the gate records `actual + 0`.
  The `1253 + 145 = 1398` value is forecast lineage only; full-tree actual is
  `1407`, including the explained historical nine-line correction.
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
