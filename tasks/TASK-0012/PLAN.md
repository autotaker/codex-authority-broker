# PLAN — TASK-0012: final zero-SLOC measurement and later-reserve gate

## Decision and authority

This is approval-ready PLAN evidence only.  It authorizes no product or test
implementation, canonical/operational-log write, fixture, network/elevation,
approval, staging, commit, merge, or push.  TASK-0012 may start only after
Main confirms the merged TASK-0017 result and its independent REVIEW and QA
PASS.  Main alone owns locks, the frozen-source copy, candidate selection,
Git, and closure.

The known merged status provenance is commit `212fd03` (PR #18), with
TASK-0017 independent REVIEW and QA PASS.  The canonical stream has no
TASK-0017 terminal; the run-lap30 contract forbids appending a synthetic one.
Therefore TASK-0017 completion status comes from that merged Git/PR/REVIEW/QA
provenance, while its timing remains `null` with reason `no canonical
TASK-0017 terminal`.  It is neither zero nor usable in timing arithmetic.

## Frozen inputs and fail-closed provenance

Main's preflight identified the read-only post-TASK-0017 v1 JSONL snapshot as
404 records with SHA-256 prefix `00482bd…`.  Before derivation, Main records
the complete digest, immutable source identity, absolute frozen-copy path,
capture UTC time, and merged TASK-0017 commit/tree in `MEASUREMENT.md`.  It
also records the prior source identity: `/tmp/task0009-events-frozen.jsonl`,
SHA-256
`abe31c9da1fbcfd32daef6013a9ce58063ccbb086265b4760ca20a22962af09b`, 389
records.  No mutable canonical or operational log is consulted after capture.

For the post-TASK-0017 snapshot, require parseable JSONL, unique `event_id`,
and every correction to name a nonempty target that is earlier in file order,
has the same `task_id` and `lap_id`, and has a smaller `sequence`:

```sh
jq -s -e '
  to_entries as $rows | all($rows[];
    if .value.event == "correction" then . as $c
    | (.value.annotations.corrects_event_id // "") as $id
    | ($id != "") and any($rows[];
        .key < $c.key and .value.event_id == $id
        and .value.task_id == $c.value.task_id
        and .value.lap_id == $c.value.lap_id
        and .value.sequence < $c.value.sequence)
    else true end)' "$EVENTS" >/dev/null
```

Retain raw event ID, order, original value, and nullable `superseded_by`.
Derive effective fields only from validated, unsuperseded values.  A correction
is neither a task nor a terminal; an invalid edge, raw/effective mixture,
hash/count mismatch, missing terminal, or unexplainable discrepancy stops as
canonical/measurement evidence defect.  Null remains null with its reason.

## Exact measurement and cap rule

Select terminals exactly as TASK-0009 does: historical cohort
TASK-0001/0003/0004/0005 remains separate; partial
`counts_as_product_lap=false` records, corrections, governance-only cycles,
terminated and unfinished work are not product terminals.  In particular,
there is no TASK-0017 terminal to select or synthesize.

Use the canonical full-production counter on the merged tree, not a
TASK-0017 terminal annotation: broker `282`, sudo helper `188`, CLI `83`,
runtime `183`, client `35`, protocol `120`, server `283`, lease `173`, and
TOTP `60`, for actual `1407` executable nonblank/non-comment production Go
lines.  Independently reconcile the Git delta: TASK-0017 parent `1262` to
merged `1407` is `+145`.  TASK-0012 adds `0`, so the actual gate arithmetic is
`1407 + 0 = 1407`; target reserve is `1500 - 1407 = 93` and hard reserve is
`1800 - 1407 = 393`.

The previously recorded `1253 + 145 = 1398` differs by nine lines because
TASK-0008 recorded the new sudo helper as 38 lines while its canonical count
is 47.  The complete historical reconciliation is therefore
`1253 + (47 - 38) + 145 = 1407`.  This documented undercount resolves the
drift; it is not a new TASK-0012 increment, a forecast, or reserve borrowing.
Acceptance requires actual `1407 < 1500` and `1407 < 1800`; the hard limit
never permits exceeding the target.  There is no velocity, throughput,
contingency-to-SLOC, forecast substitution, or deferred-v2 reserve borrowing.
TASK-0010 and TASK-0011 are inspection-only proof points: both must remain
`deferred-v2` and `executable:false`; no later milestone receives DEV detail.

## Timing, regeneration, and narrow checks

For any timing reported, pair only same-task/lap/stage/attempt
`stage_started` with its terminal; exclude `preflight`; retain unpaired,
non-authoritative, active, and wait values as reasoned nulls.  Keep active
and wait separate, propagate retries by maximum, and apply
`ceil(observed_ms * 1.20)` only to observable non-preflight time—not SLOC,
test LOC, retries, paths, or scope.

Lap 1 regenerates the source/correction, cohort/terminal, SLOC/test-LOC,
stage/null, active/wait/retry, Git-status provenance, full-tree count/delta,
historical reconciliation, and cap evidence.  In Lap 2, independent REVIEW
and QA each repeat that regeneration rather than accepting Lap 1's arithmetic.
Their narrow evidence-only checks are the correction predicate, production
SLOC and test-LOC counts, changed-path scope, and:

```sh
git diff --check
jq -e . backlog.json >/dev/null
```

Record command, environment/cache condition, exit, and a bounded artifact
digest.  Missing capability or a failing check is classified before retry;
it cannot be represented as a zero, a PASS, or a substitute measurement.

## Process-only evidence boundary and stop rule

The only future writable evidence paths are `backlog.json`,
`tasks/TASK-0012/TASK.md`, `tasks/TASK-0012/PLAN.md`,
`tasks/TASK-0012/QA_PLAN.md`, `tasks/TASK-0012/MEASUREMENT.md`,
`tasks/TASK-0012/REVIEW_RESULT.md`, and `tasks/TASK-0012/QA_RESULT.md`.
TASK-0010/TASK-0011 metadata and all product, test, configuration, schema,
dependency, generated, canonical-log, and later-milestone paths are read-only
validation inputs for this gate.  No product DEV/REVIEW/QA result is implied
by this process evidence.

Stop without enabling later reserve on a source/provenance defect,
nonreproducible arithmetic, actual result at or above 1500, any hard-limit
risk, changed deferred-v2 status, missing independent regeneration, or a
REVIEW/QA failure.  Only independent REVIEW PASS, independent QA PASS, and a
Main-owned merge may make the two named later reserves eligible for their own
PLAN; this PLAN neither enables nor details them.

## Planner evidence

Sources read: `AGENTS.md`, TASK-0012 contract and backlog metadata,
TASK-0009 PLAN/measurement/REVIEW/QA evidence, merged TASK-0017 PLAN, QA
plan, REVIEW, QA, and merge lineage, plus Main's 404-record/full-counter
preflight.  This Planner changed only this file; active/wait/retry evidence is
not an execution measurement.  The absence of a TASK-0017 canonical terminal
is resolved by the existing no-synthetic-terminal contract and merged status
provenance; the nine-line drift is fully reconciled above.
