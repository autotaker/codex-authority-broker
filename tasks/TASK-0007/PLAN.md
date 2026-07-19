# PLAN — TASK-0007: termination and replacement Tasks (Revision 6)

## Decision and present gate state

This is PLAN evidence only. It authorizes no product/test implementation,
approval, contract/index edit, Git operation, operational/Lap-log write,
deletion, or Lap-counter reset.

The merged canonical production baseline is **751**. The unfinished candidate
measures `internal/backend/runtime.go=150` and
`cmd/codex-authority-broker/main.go=186` canonical nonblank/non-comment
production SLOC, for **336** added and **1087 = 751 + 336** cumulative. Neither
owned test file exists. The combined candidate therefore exceeds the current
TASK-0007 stop **975** and hard guard **1050**.

**TASK-0007 cannot complete in the current second Lap.** Runtime tests,
independent REVIEW, and independent QA are all absent. Lap 2 is the independent
REVIEW/QA gate, not an implementation Lap in which missing tests may be added.
The current combined attempt is `planning_defect` and cannot receive PASS,
merge, or completion evidence. A third TASK-0007 Lap, silent retry, code
compression, security deletion, or test weakening is prohibited.

## Proposed Disposition 2 boundary — authority required

Revision 6 rejects the same-ID restart. The only proposed disposition is:

* terminate TASK-0007 incomplete after its consumed two Laps, record
  `planning_defect`, zero merged production SLOC, no PASS, and
  `superseded_by=[TASK-0013,TASK-0014]`;
* assign runtime-only work to new TASK-0013; and
* assign secure seed/daemon work to new TASK-0014.

This changes the original objective from “complete TASK-0007” to “terminate
TASK-0007 incomplete and complete replacement Tasks.” **Only the user/task
authority may approve that objective change.** This PLAN does not supply that
authority. Until explicit approval is recorded, TASK-0007, TASK-0013,
TASK-0014, and all dependent DEV are non-executable.

The proposed ownership is:

| Task | Depends on | Production path | Test path | Canonical forecast | Complete boundary |
| --- | --- | --- | --- | ---: | --- |
| terminated TASK-0007 | TASK-0006 | none; former draft ownership released only by authority | none | **0 merged** | Two Laps consumed; incomplete termination, no PASS/merge/completion, and no further DEV/REVIEW/QA Lap. |
| new TASK-0013 | TASK-0006 | `internal/backend/runtime.go` | `internal/backend/runtime_test.go` | **150** | Process-local lease/TOTP assembly from an injected secret; exact current `ready`/`otp` admission and denial; close/fail-closed behavior; one bounded instance registration seam. No seed open, listener, signal, daemon startup, push enablement, `init`, mutable global registry, or discovery. |
| new TASK-0014 | TASK-0013 | `cmd/codex-authority-broker/main.go` | `cmd/codex-authority-broker/main_test.go` | **186** | Descriptor-relative no-follow seed admission, fixed schema/bounds/redaction, construction of the merged runtime, and listen/serve/signal-close/restart lifecycle. No runtime API change or added IPC operation. |

`TASK-0013` and `TASK-0014` are absent from the current backlog/contracts and
are the proposed new unused identifiers. Each is a replacement Task with its
own approved TASK/PLAN/QA set and maximum-two-Lap cycle, not a continuation or
relabeling of TASK-0007's Laps. TASK-0008 keeps
its original sudo live-check/no-cache production, test, and acceptance scope;
it receives no runtime, seed, broker, or lifecycle work.

The proposed dependency chain is exactly:

```text
TASK-0006 -> TASK-0013 -> TASK-0014 -> TASK-0008 -> TASK-0009
          -> TASK-0010 -> TASK-0011 -> TASK-0012
```

TASK-0007 terminates as a non-completing sibling outcome after TASK-0006; it
is not a prerequisite in the executable replacement chain and contributes no
merged production SLOC.

## Required atomic contract/index reconciliation

Before any DEV resumes, Main must obtain authority and amend this exact set of
**nine** contract/index documents atomically:

1. `backlog.json`;
2. `tasks/TASK-0007/TASK.md`;
3. new `tasks/TASK-0013/TASK.md`;
4. new `tasks/TASK-0014/TASK.md`;
5. `tasks/TASK-0008/TASK.md`;
6. `tasks/TASK-0009/TASK.md`;
7. `tasks/TASK-0010/TASK.md`;
8. `tasks/TASK-0011/TASK.md`; and
9. `tasks/TASK-0012/TASK.md`.

The amendment must make every backlog entry and the first JSON metadata block
of each contract identical for ID, dependency, production/test ownership,
entrypoint, forecast/cumulative values, stop/target/hard gates, exclusions,
fixtures, two-Lap text, measurement lineage, and later-reserve eligibility.
It must mark TASK-0007 terminated/incomplete/non-executable with two Laps
consumed, zero merged SLOC and explicit supersession. Its exact terminal
metadata is `status=terminated`, `executable=false`, `depends_on=[TASK-0006]`,
`expected_production_sloc=0`, `production_sloc_added=0`,
`expected_cumulative_production_sloc=751`, empty production/test paths,
`entrypoint=null`, `completion=false`, `termination_classification=planning_defect`,
and `superseded_by=[TASK-0013,TASK-0014]`; it has no executable stop/target/hard
gate after termination. The amendment must add TASK-0013 and
TASK-0014 with exclusive ownership; change only TASK-0008's predecessor/cap
lineage while preserving its sudo scope; propagate the chain through
TASK-0012; and replace
TASK-0011's 220 forecast with the bounded 153-SLOC core below. TASK-0013's and
TASK-0014's PLAN and QA_PLAN and revised termination QA approval remain
separately required gate documents; they are not silently folded into this
nine-document
contract/index amendment.

No subset is executable authority. A partially reconciled index, old
dependency, old ownership, or old cap leaves all affected DEV blocked.

## TASK-0011 retained core: maximum 153 production SLOC

TASK-0011 is replanned from **220 to at most 153 canonical production SLOC**.
Tests remain separate from this production count and may not be weakened to
meet it. The production-path ceilings are:

| Production path | Max SLOC | Retained core |
| --- | ---: | --- |
| `cmd/codex-authority-push/main.go` | 16 | One fixed caller path that conveys only configured repository identity and one permitted source/destination ref intent; generic bounded failure exit. |
| `cmd/codex-authority-broker/main.go` | 4 | One explicit construction-time call to the fixed push registrar with already-constructed dependencies; no lifecycle/seed/ready/OTP change. |
| `internal/ipc/protocol.go` | 26 | Exactly `OperationPush` plus one strict bounded request schema; malformed, unknown, force/tag/delete/multiple-ref/ambiguous input denies before dispatch. |
| `internal/backend/push_registration.go` | 36 | Exactly one handler registration and the fixed UID, live-lease, and TASK-0010 policy gates before custody or Git. |
| `internal/push/custody.go` | 33 | One short-lived token acquisition/use boundary, memory-only custody, generic errors, and redaction; no token in argv, environment, output, or persistent cache. |
| `internal/push/system_git.go` | 38 | One captured system-Git, single-ref, non-force push path with bounded credential injection and output suppression. |
| **Total** | **153** | Fixed authorization/custody/non-force core only. |

This retains the non-negotiable fixed identity and ref intent, caller UID,
live lease, TASK-0010 policy, strict schema admission, token custody,
redaction, deterministic system-Git capture, and no-force behavior.

The 67-SLOC reduction from the previous 220 allocation is contractual scope
removal, not compression:

* caller allocation **35 -> 16 (-19)**: remove convenience CLI modes,
  user-selectable formatting, rich exit/error taxonomy, and duplicate local
  presentation validation; retain the one fixed request and generic result;
* protocol/schema **45 -> 26 (-19)**: remove reusable generic command/schema
  frameworks, extensible field machinery, and per-field diagnostic responses;
  retain direct strict push admission and generic denial;
* registration/gates remain **40 = 4 broker + 36 registrar**: no security gate
  is removed and no saving is claimed here;
* custody **55 -> 33 (-22)**: remove proactive refresh, token caching,
  retry/backoff, telemetry, and multi-provider abstraction; retain one
  short-lived acquisition, memory-only use, cleanup, and redaction;
* system Git **45 -> 38 (-7)**: remove retry orchestration, progress/report
  parsing, and post-push diagnostic classification; retain one captured
  non-force invocation and denial-before-transport evidence.

Already-excluded arbitrary refspec, remote-OID prefetch, force/tag/delete,
generic IPC commands, network race diagnostics, sudo, audit, release,
installer, and canary are not counted as savings. If the retained core cannot
be forecast and implemented readably within 153 without weakening a fixed
requirement, TASK-0011 stops for authority replan; source packing is forbidden.

## Revised cumulative gates

With the bounded TASK-0011 core, the wave forecast is:

**1490 = 751 + 150 + 186 + 120 + 0 + 130 + 153 + 0**.

It leaves 10 SLOC to the global target 1500 and 310 to the unconditional hard
1800. Every forecast is strictly below its stop, so no row self-stops:

| Task | Add | Cumulative | Stop | Target | Hard | Forecast-to-stop |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| TASK-0007 terminated | 0 merged | 751 | n/a | n/a | n/a | n/a |
| TASK-0013 runtime | 150 | 901 | 925 | 950 | 1000 | 24 |
| TASK-0014 seed/daemon | 186 | 1087 | 1125 | 1150 | 1200 | 38 |
| TASK-0008 | 120 | 1207 | 1250 | 1300 | 1400 | 43 |
| TASK-0009 | 0 | 1207 | 1250 | 1300 | 1400 | 43 |
| TASK-0010 | 130 | 1337 | 1380 | 1425 | 1500 | 43 |
| TASK-0011 | 153 | 1490 | 1495 | 1500 | 1650 | 5 |
| TASK-0012 | 0 | 1490 | 1495 | 1500 | 1800 | 5 |

Crossing a stop requires explicit re-estimation before further DEV; it does
not raise a target or hard guard. TASK-0012 must reconcile actual values to
both 1500 and 1800. Later audit/attestation/manual-canary work remains
non-executable until TASK-0012 independent REVIEW and QA PASS and Main merge.

## Fixed Lap disposition

TASK-0007 has consumed Lap 1 and Lap 2. It terminates incomplete if and only if
the user/task authority approves Disposition 2. Its history, measurements, and
failure classification remain intact; there is no counter reset, renamed
Lap, Lap 3, further implementation, PASS, or merge under TASK-0007.

After authority approval and atomic reconciliation, each replacement Task has
its own strict maximum of two Laps:

| Task | Lap 1 | Lap 2 | Stop condition |
| --- | --- | --- | --- |
| TASK-0013 | DEV owns only `runtime.go` and `runtime_test.go`; produces a compile-valid, focused-test-passing runtime candidate. | Independent REVIEW runs focused/full checks, then independent QA executes routing, denial, close, registration, and regression evidence; Main alone owns Git. | Need for seed/listener/lifecycle paths, missing gate-ready tests, cap crossing, or further DEV after Lap 1 stops/replans the Task; no Lap 3. |
| TASK-0014 | DEV owns only broker `main.go` and `main_test.go`; produces secure seed/lifecycle tests and a compile-valid candidate against merged TASK-0013. | Independent REVIEW runs focused/full checks, then independent QA executes the seed mutation, redaction, startup/shutdown/restart, and client regression matrix; Main alone owns Git. | Need for runtime/API/IPC change, missing gate-ready tests, cap crossing, or further DEV after Lap 1 stops/replans the Task; no Lap 3. |

Preflight failure is `not_started`, excluded from DEV timing, and not retried
unchanged. PLAN -> DEV -> independent REVIEW -> independent QA -> Main Git
remains sequential for each replacement Task.

## Candidate and measurement disposition

Preserve both unfinished source candidates and all unrelated working-tree
changes as unaccepted evidence. Do not delete, revert, compact, stage, merge,
or claim either candidate as a completed Lap. `runtime.go` may inform
TASK-0013 and `main.go` may inform TASK-0014 only after the authority,
contract, PLAN, and QA gates above.

This Revision 6 changes only `tasks/TASK-0007/PLAN.md`. Planner observation:
`active_ms=null`, `wait_ms=null`, `retries=0`,
`classification=planning_defect`; null means unobserved, never zero. QA can
reproduce the counts, the 153 per-path total, the 67 reduction, the 1490 final
cumulative, the nine-document atomic set, and the explicit Disposition 2 Lap
history. Without user/task-authority approval of the objective change, this
proposal is non-executable and no atomic amendment or DEV may begin.
