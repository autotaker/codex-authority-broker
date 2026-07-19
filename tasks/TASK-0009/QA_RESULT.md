# TASK-0009 independent QA result

**Verdict: FAIL**

Independent post-DEV QA executed Q9-01 through Q9-09 against
`/tmp/task0009-events-frozen.jsonl` without reading or depending on the
REVIEW verdict. One blocking measurement-provenance defect remains. All other
acceptance rows pass, and the native-suite failure is separately classified as
an environment issue rather than attributed to DEV.

## Frozen source and execution evidence

- Source: `/tmp/task0009-events-frozen.jsonl`
- SHA-256: `abe31c9da1fbcfd32daef6013a9ce58063ccbb086265b4760ca20a22962af09b`
- Records: 389
- JSON parse: PASS
- Unique `event_id`: PASS
- Correction predicate: PASS for all 14 correction records
- QA active_ms: `null`
- QA wait_ms: `null`
- QA active/wait null reason: the QA runtime did not expose authoritative
  active-versus-wait boundaries; elapsed tool time is not relabeled as active
  and zero is not synthesized.
- QA retries: `1` (one QA-local malformed `jq --argjson` verification
  command was corrected before evidence was accepted; no candidate retry or
  canonical mutation occurred)

## Q9 acceptance results

| ID | Result | Independent evidence |
| --- | --- | --- |
| Q9-01 | PASS | Frozen hash/389-line check, JSON parse, unique-ID check, and same-task/lap, earlier-file, smaller-sequence predicate pass for all 14 corrections. |
| Q9-02 | PASS | Historical cohort regenerates exactly TASK-0001/0003/0004/0005. Product terminals regenerate only `task0015-counted-lap02-complete`, `task0016-counted-lap02-complete`, and `task0008-counted-lap01-complete`. Three completed partial terminals are excluded; governance, corrections, TASK-0007, and TASK-0014 are not product terminals. |
| Q9-03 | PASS | Terminal snapshots independently reproduce `1200 -> 1215 -> 1253`; normalized increments are +278, +15, and +38. TASK-0009 adds 0: `1253 + 0 = 1253`. |
| Q9-04 | PASS | No TASK-0013 `lap_completed/status=completed` record exists. Its two lap-completed records have status `pass`; the measurement preserves completion timing as null with reason `no canonical TASK-0013 terminal`. |
| Q9-05 | PASS | All 31 source-ID stage pairs listed by the measurement independently reproduce their stated duration and match task/lap/stage/attempt. Historical/product cycle arithmetic and null pairing rules are reproducible; preflight is not used as a stage duration. |
| Q9-06 | PASS | Each of the 31 observed non-preflight stage durations reproduces its stated `ceil(ms * 1.20)` contingency. Nulls and SLOC/test/active/wait/retry values receive no multiplier. |
| Q9-07 | PASS | Maximum propagated retries independently reproduce 0/1/4/1/2/1/2/2 for TASK-0001/0003/0004/0005/0013/0015/0016/0008 and 4/3/8 for governance/terminated/unfinished TASK-0006/0007/0014. Active/wait observations remain separate and task-level totals remain null rather than fabricated. |
| Q9-08 | **FAIL** | Exact ten-path process-only scope is satisfied and no `cmd`, `internal`, or `deploy` file differs. However, correction provenance in `MEASUREMENT.md` is contradictory: the row for `task0006-r3-lap01-correct-stop-event -> task0006-r3-lap01-stop` labels the same-lap proof as `TASK-0006/lap03`. Both frozen records actually have `lap_id=task0006-r3-lap01` (sequences 24 and 23). Classification: `measurement_implementation_defect`. |
| Q9-09 | PASS | Arithmetic `1253 + 130 + 153 = 1536` is explicit, 36 above target 1500. The canonical ordered-shedding list has item 7 “move GitHub push to v2”; backlog and downstream Tasks explicitly select it. TASK-0010 and TASK-0011 are `deferred-v2`, `executable:false`, zero current increment, while TASK-0012 depends directly on TASK-0009 and is the executable zero-SLOC `1253 + 0 = 1253` gate. No later reserve is enabled. |

## Blocking reproduction

Frozen records:

```text
task0006-r3-lap01-stop:
  task_id=TASK-0006 lap_id=task0006-r3-lap01 sequence=23
task0006-r3-lap01-correct-stop-event:
  task_id=TASK-0006 lap_id=task0006-r3-lap01 sequence=24
  corrects_event_id=task0006-r3-lap01-stop
```

Candidate evidence at `tasks/TASK-0009/MEASUREMENT.md:178` instead says
`TASK-0006/lap03; 23 < 24`. The edge validates, but its displayed lap
provenance does not. The minimal responsible-gate correction is to replace
that displayed lap label with the actual frozen `task0006-r3-lap01` value
and rerun independent scope/diff validation. QA did not edit the measurement.

## Repository-native checks and failure classification

The exact QA-plan commands produced:

| Check | Result | Classification |
| --- | --- | --- |
| `GOCACHE="$(mktemp -d)" go test ./...` | FAIL in the ordinary sandbox: CLI real-socket capture cannot build, and `internal/ipc` real Unix sockets fail with `socket: operation not permitted`; all environment-permitted broker, sudo, deploy, backend, and lease packages pass. | `environment_issue`, non-blocking and not candidate-specific. Main separately reports that elevated IPC execution passes, the nested CLI build initially exposes `error obtaining VCS status`, and elevated `GOFLAGS=-buildvcs=false go test ./...` passes all packages. |
| `gofmt -l $(find cmd internal -type f -name '*.go' -print)` | PASS, empty output. | none |
| `git diff --check` | PASS. | none |
| `jq -e . backlog.json` | PASS. | none |
| Exact changed-path enumeration after this result | PASS: exactly the authorized ten process paths; no product/test/canonical path. | none |

The full-suite sandbox failure is environment-only and does not determine the
FAIL verdict. The verdict is solely the reproducible Q9-08 measurement
provenance mismatch. After that candidate evidence is corrected, QA must
independently rerun the affected correction-provenance and exact-scope checks;
PASS is not granted by this report.

## QA retry 1 — post-correction verdict

**Current verdict: PASS**

The initial FAIL above is preserved as history. DEV changed only the blocked
displayed provenance label in `MEASUREMENT.md`. Independent retry evidence:

- Q9-08 correction provenance: PASS. The candidate now states
  `TASK-0006/task0006-r3-lap01; 23 < 24`, matching frozen target
  `task0006-r3-lap01-stop` (sequence 23) and correction
  `task0006-r3-lap01-correct-stop-event` (sequence 24), both with
  `task_id=TASK-0006` and `lap_id=task0006-r3-lap01`.
- Exact scope: PASS. Changed/untracked enumeration contains exactly the ten
  authorized process-evidence paths, including this `QA_RESULT.md`; no
  `cmd`, `internal`, or `deploy` path differs.
- Regression checks: `git diff --check` PASS and `jq -e . backlog.json` PASS.
  The environment-classified full-suite evidence from the initial run is
  unchanged and was not redundantly rerun.
- QA retry active_ms: `null`; wait_ms: `null`; reason: this runtime exposes no
  authoritative active-versus-wait boundary. Retry counter: `1` post-candidate
  correction (the earlier QA-local command-construction retry remains recorded
  separately above).

The sole blocking `measurement_implementation_defect` is corrected. Q9-01
through Q9-09 now pass with exact ten-path, zero-product-SLOC scope. This QA
retry grants PASS; Main retains Git and merge ownership.
