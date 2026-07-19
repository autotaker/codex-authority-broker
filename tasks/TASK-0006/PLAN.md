# PLAN — TASK-0006: Measurement and six-Task rolling-wave revision (Revision 8)

## Decision, scope, and immutable evidence

Revision 8 preserves the Revision 6 reachable-push-path resolution and all
Revision 7 repository-native checks, and closes the Revision 7 zero-SLOC
correction-order `planning_defect`.  It is PLAN evidence only and authorizes
no product implementation, approval, Git operation, or Lap-log write.
TASK-0006 adds **0 production SLOC**.  The immutable measured
baseline remains `98 + 135 + 367 + 151 = 751` cumulative production SLOC from
the 72 canonical events for TASK-0001/0003/0004/0005 in read-only
`/home/ubuntu/git/agent-harness-work/lap30/events.jsonl`.

The existing measurement rules remain mandatory: unmatched timing is `null`,
not zero; stage durations require same-Task/lap/stage/attempt start-terminal
pairs; active and wait are separate observations; retries are the maximum
propagated task counter; raw classifications retain source event IDs and
`superseded_by`, while only correction-validated unsuperseded values are
effective.  No SLOC/minute, average velocity, regression, or other fixed
throughput assumption sizes this wave.

Revision 8 DEV is limited to these **exact seven outputs**:

1. `backlog.json`
2. `tasks/TASK-0007/TASK.md`
3. `tasks/TASK-0008/TASK.md`
4. `tasks/TASK-0009/TASK.md`
5. `tasks/TASK-0010/TASK.md`
6. `tasks/TASK-0011/TASK.md`
7. `tasks/TASK-0012/TASK.md`

No `.go`, product/test source, measurement result, REVIEW/QA result, other
Task document, operational repository, `.git`, or Lap log is an allowed
Revision 8 DEV output.  The contract documents describe future ownership;
they do not implement it.

Before any of the six contracts becomes executable, the index and its contract
must agree exactly on ID, title, status/executable flag, dependency, expected
increment, target cumulative cap, hard cumulative guard, production and test
paths, entrypoint, fixture/elevation needs, Lap 1/Lap 2 work, exclusions,
split/stop rule, measurement lineage, and later-reserve eligibility.  IDs are
unique, dependency references exist and are acyclic, no converted Task remains
simultaneously reserved, and no seventh detailed contract is created.

## Forecast, target caps, hard guards, and later reserve

Forecast and cap are deliberately different values:

- **Expected production SLOC** is the planning forecast for the bounded
  contract.  It is not a ceiling and is independently remeasured.
- **Target cumulative cap** is the per-Task v1 planning ceiling.  Forecast
  above 90% of this cap stops before DEV for split/re-estimation and approved
  PLAN/QA_PLAN revision.  A candidate above the cap stops even if below hard.
- **Hard cumulative guard** is an absolute per-Task guard.  It is neither
  forecast nor ordinary scope.  Crossing or forecasting above it safely stops
  the Task.  The system-wide **1800** production-SLOC hard limit is
  unconditional and cannot be raised by approval, contingency, or shedding.

| Task | Expected add | Expected cumulative | Target cumulative cap | 90% trigger | Hard cumulative guard | Trigger check |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| TASK-0007 | 90 | **841** | 950 | 855 | 1050 | 841 <= 855 |
| TASK-0008 | 120 | **961** | 1100 | 990 | 1250 | 961 <= 990 |
| TASK-0009 | 0 | **961** | 1100 | 990 | 1250 | 961 <= 990 |
| TASK-0010 | 130 | **1091** | 1250 | 1125 | 1400 | 1091 <= 1125 |
| TASK-0011 | 220 | **1311** | 1500 | 1350 | 1650 | 1311 <= 1350 |
| TASK-0012 | 0 | **1311** | 1500 | 1350 | 1650 | 1311 <= 1350 |

Arithmetic is `751 + 90 + 120 + 0 + 130 + 220 + 0 = 1311`.  Therefore
**189 SLOC remains to the 1500 target** and **489 SLOC remains to the 1800
unconditional hard limit**.  Those amounts stay a non-executable later reserve
for mandatory minimal audit, source-free release/attestation, and manual
canary/rollback runbook and evidence.  Owner lineage is
`TASK-0012 measurement PASS+merge -> MILESTONE-audit-attestation ->
MILESTONE-manual-canary-rollback`; neither milestone receives PLAN, branch,
DEV, PR-ready detail, or production allocation in this six-Task wave.

The six Tasks do **not** complete v1.  TASK-0012 must reconcile the actual
remaining target/hard capacity and may only propose the later contracts; the
later mandatory owners remain ineligible for DEV until TASK-0012 independent
REVIEW and QA both PASS and it merges.  The reserve is not available to the
six wave Tasks by silent borrowing.

## Six-Task ownership and two-Lap execution contracts

All future paths below are exclusive ownership proposed for the named Task.
If an approved predecessor establishes a different exact path/interface, the
affected contract must be revised and reapproved before DEV; paths are not
left as an implementation-time choice.

The existing `cmd/codex-authority/main.go` and
`cmd/codex-authority/main_test.go` remain the non-privileged `ready`/`otp`
client and its tests.  TASK-0007 does not convert them into a daemon, and
TASK-0011 does not add a push hook to them.  They are unchanged by this wave.

| Task | Exclusive future production paths | Exclusive future test paths | Focused command |
| --- | --- | --- | --- |
| TASK-0007 | `cmd/codex-authority-broker/main.go`; `internal/backend/runtime.go` | `cmd/codex-authority-broker/main_test.go`; `internal/backend/runtime_test.go` | `go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease` (existing CLI included as regression) |
| TASK-0008 | `cmd/codex-authority-sudo/main.go`; `deploy/sudo/codex-authority` | `cmd/codex-authority-sudo/main_test.go`; `deploy/sudo/codex-authority_test.go` | `go test ./cmd/codex-authority-sudo ./internal/ipc` plus the isolated sudo fixture |
| TASK-0009 | no product path | `tasks/TASK-0009/MEASUREMENT.md` only; REVIEW/QA own their result files | canonical JSONL parse/unique-ID/correction-reference check plus measurement regeneration and cap arithmetic |
| TASK-0010 | `internal/push/policy.go`; `internal/push/validate.go` | `internal/push/policy_test.go`; `internal/push/validate_test.go` | `go test ./internal/push` |
| TASK-0011 | `cmd/codex-authority-push/main.go`; `internal/ipc/protocol.go`; `internal/push/custody.go`; `internal/push/system_git.go`; `internal/backend/push_registration.go` | `cmd/codex-authority-push/main_test.go`; `internal/ipc/protocol_test.go`; `internal/push/custody_test.go`; `internal/push/system_git_test.go`; `internal/backend/push_registration_test.go` | `go test ./cmd/codex-authority ./cmd/codex-authority-broker ./cmd/codex-authority-push ./internal/ipc ./internal/push ./internal/backend` |
| TASK-0012 | no product path | `tasks/TASK-0012/MEASUREMENT.md` only; REVIEW/QA own their result files | canonical JSONL parse/unique-ID/correction-reference check plus measurement regeneration and actual-cap arithmetic |

TASK-0007's runtime exposes the bounded handler-registration seam while
installing only fixed `ready`/`otp` routing.  TASK-0011 later owns the bounded
`push` protocol operation/schema and the separate
`internal/backend/push_registration.go` module that attaches it through that
already-merged seam.  The existing client remains `ready`/`otp` only, and the
daemon entrypoint remains unchanged, so ownership remains disjoint.

### TASK-0007 — production daemon/backend assembly and bounded seed

**Dependency:** TASK-0006.  **Expected/caps:** +90, cumulative 841; target 950;
hard guard 1050.

**Future production ownership:** new `cmd/codex-authority-broker/main.go` as
the one privileged daemon entrypoint, and new `internal/backend/runtime.go` as the assembly and
lifecycle boundary.  **Future test ownership:**
`cmd/codex-authority-broker/main_test.go` and
`internal/backend/runtime_test.go`.
The Task wires the existing IPC/lease packages without changing their
security semantics.  The existing `cmd/codex-authority` client remains intact.

The production contract owns startup, readiness, signal-driven shutdown, and
restart behavior; root-owned bounded configuration/seed injection; in-memory
lease/TOTP state construction; and exact fixed routing of IPC `ready` and
`otp` requests to that state.  The seed source is a root-owned mode-0600
bounded config input, read once at startup, size-limited, schema-validated,
copied into process state, and never accepted from argv, environment, peer IPC,
or logs.  Missing/malformed/oversized/wrong-owner/wrong-mode input, backend
construction failure, and restart without valid seed fail closed and never
report ready.  Fixture-only seed helpers remain test-only and cannot replace
this production injection boundary.

| Lap | Concrete work and evidence |
| --- | --- |
| **Lap 1** | Preflight confirms TASK-0006 merge, exact existing `ready`/`otp` protocol, the existing client/daemon path separation, temp Unix-socket and root-owned-config fixture design, and whether the isolated runner can simulate root ownership without broad host mutation.  After approved PLAN+QA_PLAN, DEV changes only the owned paths, runs focused `go test ./cmd/codex-authority ./cmd/codex-authority-broker ./internal/backend ./internal/ipc ./internal/lease`, and produces existing-client regression plus startup/ready/OTP/shutdown/restart candidate evidence. |
| **Lap 2** | Independent REVIEW runs focused tests plus the repository-native full check below; QA independently tests valid seed, missing/malformed/oversized/owner/mode failures, exact ready/OTP routing, secret redaction, shutdown, and fail-closed restart; main owns final checks/Git.  No real system service installation occurs. |

**Split/stop:** stop before DEV if entrypoint plus one assembly boundary is not
the complete path, if a new persistent store/config framework is required, if
root ownership cannot be safely simulated or isolated, if expected cumulative
exceeds 855, or if Lap 1 cannot leave a gate-ready candidate.  Split lifecycle
from any newly discovered persistence requirement; never defer seed/lifecycle
ownership into TASK-0008.

### TASK-0008 — sudo live check and no cache

**Dependency:** merged TASK-0007.  **Expected/caps:** +120, cumulative 961;
target 1100; hard guard 1250.

**Future production ownership:** new `cmd/codex-authority-sudo/main.go` for the
minimal `pam_exec`-compatible live-check client and `deploy/sudo/codex-authority`
for the dedicated identity's declarative no-cache policy.  **Future test
ownership:** `cmd/codex-authority-sudo/main_test.go` and
`deploy/sudo/codex-authority_test.go`.  It consumes TASK-0007's fixed IPC and
does not change daemon/backend assembly.

| Lap | Concrete work and evidence |
| --- | --- |
| **Lap 1** | Preflight requires merged TASK-0007, an isolated Ubuntu sudo/PAM fixture, disposable dedicated identity, controlled clock/socket, and approved narrow elevation procedure.  DEV implements per-invocation live request and declarative timestamp-cache disablement; focused tests cover allow, expiry, daemon unavailable/restart, malformed/unauthorized reply, and two consecutive invocations. |
| **Lap 2** | REVIEW runs focused tests and the repository-native full check below; QA uses the isolated elevated fixture to prove a live unexpired lease permits and every deny case fails closed with no cached reuse, then main performs Git closure.  Real workstation sudo policy is never mutated. |

**Split/stop:** `not_started/environment_issue` if the isolated elevated fixture
or rollback proof is unavailable.  Split before DEV if more than the single
client entrypoint and declarative policy are required, forecast exceeds 990,
or platform/PAM differences cannot be covered in two laps.  Do not weaken
live-per-call or no-cache behavior.

### TASK-0009 — first zero-SLOC measurement gate

**Dependency:** merged TASK-0008.  **Expected/caps:** +0, cumulative 961;
target 1100; hard guard 1250.  **Future owned paths:**
`tasks/TASK-0009/MEASUREMENT.md`, `tasks/TASK-0009/REVIEW_RESULT.md`, and
`tasks/TASK-0009/QA_RESULT.md`; no product/test path.

| Lap | Concrete work and evidence |
| --- | --- |
| **Lap 1** | Preflight freezes the read-only completed-event snapshot and verifies TASK-0007/0008 merges.  DEV produces provenance-complete SLOC/test/stage/active/wait/retry/raw+effective-classification arithmetic and applies `ceil(observed_non_preflight_time * 1.20)` only to observable time.  Focused checks validate JSONL uniqueness/corrections and cap arithmetic. |
| **Lap 2** | Independent REVIEW regenerates the table and runs the explicit zero-SLOC measurement and repository-native full checks below; QA independently repeats canonical/null/correction/preflight/cap checks; main owns Git.  TASK-0010's speculative plans are invalidated if measured evidence changes its boundary. |

**Split/stop:** stop on missing/contradictory canonical evidence,
non-reproducible arithmetic, actual cumulative above 1100, or inability to
complete independent regeneration in Lap 2.  Classify before retry; no imputed
metric and no bypass to TASK-0010.

### TASK-0010 — local push policy and validation

**Dependency:** merged TASK-0009 PASS.  **Expected/caps:** +130, cumulative
1091; target 1250; hard guard 1400.

**Future production ownership:** new `internal/push/policy.go` and
`internal/push/validate.go`.  **Future test ownership:**
`internal/push/policy_test.go` and `internal/push/validate_test.go`.  No token,
credential helper, network transport, or Git child process is owned here.

| Lap | Concrete work and evidence |
| --- | --- |
| **Lap 1** | Preflight requires merged TASK-0009 and a temp worktree/local bare-repository matrix with no network or elevation.  DEV validates the exact configured repository, clean tree, `main` or `task/TASK-*`, one-ref update, and rejects wrong repo/ref, dirty tree, force, tag, delete, multiple-ref, and ambiguous local Git state.  Focused package tests cover the full matrix. |
| **Lap 2** | REVIEW runs focused tests and the repository-native full check below; QA independently mutates each local condition and proves denied cases cannot cross a fake transport boundary; main owns Git. |

**Split/stop:** stop if validation requires remote state, credentials, or a
second policy language; split that discovery from this local boundary.  Stop
if forecast exceeds 1125, fixture cannot prove zero transport on deny, or the
candidate is not review-ready in Lap 1.

### TASK-0011 — token custody and system-Git non-force push

**Dependency:** merged TASK-0010.  **Expected/caps:** +220, cumulative 1311;
target 1500; hard guard 1650.

**Future production ownership:** new `cmd/codex-authority-push/main.go` as the
only supported local restricted-push caller; `internal/ipc/protocol.go` for one
bounded `push` operation and payload admission; new
`internal/push/custody.go` and `internal/push/system_git.go`; and new
`internal/backend/push_registration.go` for the fixed backend route through
TASK-0007's merged handler seam.  **Future test ownership:**
`cmd/codex-authority-push/main_test.go`, `internal/ipc/protocol_test.go`,
`internal/push/custody_test.go`, `internal/push/system_git_test.go`, and
`internal/backend/push_registration_test.go`.  Neither
`cmd/codex-authority/main.go` nor `cmd/codex-authority-broker/main.go` changes.
TASK-0010 policy remains unchanged and must PASS before custody is invoked.

The caller accepts only the configured repository plus one permitted local
branch/ref intent; it has no token, force, tag, delete, arbitrary refspec,
remote-command, environment-credential, or generic IPC-operation option.  It
strictly constructs `ipc.OperationPush` with a bounded `PushRequest` containing
only the exact repository identity and single source/destination ref fields
needed by TASK-0010 validation.  Unknown, missing, duplicate-equivalent,
oversized, malformed, force/tag/delete/multiple-ref, or noncanonical fields are
rejected before backend dispatch.  The existing `cmd/codex-authority` parser
continues to reject `push`, so sudo/ready/OTP client behavior is not reopened.

The executable path is a supported caller boundary, not an authentication
claim.  Broker admission additionally requires the dedicated local caller UID
from TASK-0007's root-owned configuration, verified by the existing
SO_PEERCRED fail-closed server boundary; a live lease; and TASK-0010 local
policy PASS.  Wrong UID, absent/expired lease, invalid policy, malformed schema,
unknown operation, or unavailable registration denies before token retrieval
or Git execution.  The protocol change admits exactly `ready`, `otp`, and
`push`; it adds no generic command or transport escape hatch.

The +220 expected production SLOC remains credible as a boundary allocation,
not a throughput estimate: approximately 35 caller parsing/construction, 45
strict IPC operation/schema admission, 40 backend registration plus
UID/lease/policy gating, 55 bounded token custody, and 45 system-Git execution
and redaction (`35 + 45 + 40 + 55 + 45 = 220`).  Tests are counted separately.
If path-level design expands beyond these five bounded units, the 90%/two-Lap
stop applies before DEV; the expected/cap/reserve arithmetic does not silently
change.

| Lap | Concrete work and evidence |
| --- | --- |
| **Lap 1** | Preflight requires merged TASK-0010, TASK-0007's stable handler seam and configured dedicated caller UID, local bare remote, fake short-lived GitHub App token provider, system-Git binary, credential-capture sentinel, live-lease fixture, no network, and no elevation.  DEV implements the exact caller, strict push operation/payload admission, backend registration/gates, bounded in-memory custody, and one system-Git single-ref non-force path.  Focused `go test ./cmd/codex-authority ./cmd/codex-authority-broker ./cmd/codex-authority-push ./internal/ipc ./internal/push ./internal/backend` proves old-client rejection, protocol schema bounds, reachability from the authorized caller, wrong-UID/pre-dispatch denial, token-channel absence, and ambiguity/no-force-retry. |
| **Lap 2** | REVIEW runs the focused suite, malformed/unknown-operation and capture-sentinel mutations, then the repository-native full check below; QA independently proves the authorized caller reaches exactly one push handler only with correct UID, live lease, and TASK-0010 policy, while old CLI, wrong UID, malformed schema, wrong/expired authority, leak, force, and ambiguity cases deny before custody/Git as applicable; main owns Git. |

**Split/stop:** stop before DEV if the caller cannot be bounded to one schema,
SO_PEERCRED cannot distinguish the configured caller identity, protocol
admission would require a generic command/refspec, the approved injection
design cannot keep the token out of every named channel, system Git cannot be
deterministically captured locally, the five-unit forecast exceeds 1350, or
the boundary requires remote OID prefetch/race diagnostics.  Shed those
optional diagnostics in order or split; never weaken caller authorization,
schema admission, token custody, or non-force-only behavior.

### TASK-0012 — final zero-SLOC measurement and later-reserve gate

**Dependency:** merged TASK-0011.  **Expected/caps:** +0, cumulative 1311;
target 1500; hard guard 1650.  **Future owned paths:**
`tasks/TASK-0012/MEASUREMENT.md`, `tasks/TASK-0012/REVIEW_RESULT.md`, and
`tasks/TASK-0012/QA_RESULT.md`; no product/test path.

| Lap | Concrete work and evidence |
| --- | --- |
| **Lap 1** | Preflight freezes the completed TASK-0010/0011 event snapshot.  DEV regenerates full historical plus wave SLOC/test/time/active/wait/retry/classification evidence, reconciles actual capacity to 1500/1800, and records eligibility evidence for the still-non-executable audit/attestation/manual-canary lineage. |
| **Lap 2** | REVIEW independently regenerates evidence and runs the explicit zero-SLOC measurement and repository-native full checks below; QA repeats all measurements, validates the 189/489 reserve using actual rather than forecast values, and proves no later DEV is enabled; main owns Git. |

**Split/stop:** stop on canonical defects, unexplained cap drift, actual above
1500 without prior approved contingency disposition, any hard-limit risk, or
inability to close REVIEW/QA in Lap 2.  Later reserve PLAN/DEV remains blocked
until this gate passes and merges.

## Pipeline, role, and two-Lap invariants

Each Task preserves `PLAN -> DEV -> independent REVIEW -> independent QA ->
main-owned Git`.  DEV, reviewer, and QA are separate agents under their
canonical role/model contracts.  Reviewer and QA each run their planned
focused checks and required full check; only main owns locks, staging, commits,
pushes, PRs, and merges.

The repository has no Makefile; no future Task may substitute `make check` or
`make task-check`.  In each independent REVIEW and QA, the required full check
is `GOCACHE="$(mktemp -d)" go test ./...` when the runtime's default Go cache
is not writable (otherwise `go test ./...` is sufficient), followed by
`test -z "$(gofmt -l $(find cmd internal -type f -name '*.go' -print))"`,
`git diff --check`, and `jq -e . backlog.json >/dev/null` whenever that Task
changes the backlog or a JSON document.  Focused tests and fixture checks in
the contracts remain additive; this full suite never replaces them.

For TASK-0009 and TASK-0012, REVIEW and QA each use the frozen canonical
snapshot rather than an invented task-check target and run this JSONL-stream
validation (the `jq -s` step retains zero-based file order):

```sh
EVENTS=/home/ubuntu/git/agent-harness-work/lap30/events.jsonl
jq -e . "$EVENTS" >/dev/null
test -z "$(jq -r '.event_id' "$EVENTS" | sort | uniq -d)"
jq -s -e '
  to_entries as $rows
  | all($rows[];
      if .value.event == "correction" then
        . as $correction
        | any($rows[];
            .key < $correction.key
            and .value.event_id == $correction.value.annotations.corrects_event_id
            and .value.task_id == $correction.value.task_id
            and .value.lap_id == $correction.value.lap_id
            and .value.sequence < $correction.value.sequence)
      else true end)
' "$EVENTS" >/dev/null
```

Thus every correction target must exist, be in the same Task and lap, and
precede the correction in both canonical file order and within-lap sequence;
recording timestamps are not used to invent occurrence order.  Each target and
correction remains in raw provenance with its source event ID.  Only after the
edge passes these predicates may the target raw record name the correction in
`superseded_by` and effective classification omit the superseded value; an
invalid edge stops measurement rather than rewriting or discarding history.
REVIEW and QA then independently regenerate `MEASUREMENT.md` from that snapshot
and check its null/correction/provenance, SLOC/test/time/active/wait/retry,
classification, and cap arithmetic; the full Go/format/diff checks above also
remain required.  If a later frozen snapshot has a different approved path,
the same parse, unique-ID, same-Task/lap, earlier-file-order, and earlier-sequence
predicates are applied to that exact frozen file.

While the current Task is in REVIEW/QA, only the immediate successor's PLAN,
TASK-first QA_PLAN, and read-only exploration may proceed in parallel, with
disjoint file ownership and no assumption that an unmerged interface exists.
Successor DEV is prohibited until the dependency actually merges, preflight
revalidates it, and both successor PLAN and QA_PLAN are approved.  An upstream
FAIL invalidates affected speculative planning.  No shared backlog/contract/
Lap-log write runs concurrently.

Lap 1 targets preflight, approved plans, bounded DEV, focused checks, and a
gate-ready candidate.  Lap 2 targets remaining DEV, independent REVIEW,
focused remediation, independent QA, and main-owned Git closure.  The tables
above are Task-specific feasibility evidence, not authority to skip a gate.
If two-Lap fit is doubtful before DEV, split the contract first.  If Lap 2 is
exceeded or likely to be exceeded, safely stop, classify, and contract-split/
replan; partial merge, gate bypass, scope absorption, and code compression are
forbidden.

## Active, wait, retries, classifications, and preflight

Every contract and both measurement gates record raw milliseconds and source
event IDs where available:

- `active_ms` and `wait_ms` separately; they may overlap and are never summed
  to invent elapsed time.  Approval, dependency merge, fixture, lock,
  permission, or elevation waiting is wait, not hidden DEV.
- `retries` as the maximum propagated task-level counter, with every
  retry-bearing event ID; checkpoint snapshots are not double-counted.
- raw classifications with source event ID and nullable `superseded_by`, and
  correction-validated effective classifications.  Categories are
  `implementation_defect`, `planning_defect`, `qa_plan_defect`,
  `requirement_gap`, `environment_issue`, and `regression`.
- preflight failure as `not_started` with classification.  It spends no DEV
  lap and creates no synthetic zero duration.  A predictable unchanged
  environment failure is classified before any retry.

## Mandatory controls, later reserve, shedding, and no compression

Unsheddable controls remain readiness; TOTP replay/rate/absolute expiry;
SO_PEERCRED fail-closed IPC; production daemon lifecycle and bounded seed;
per-sudo live check/no cache; secret non-disclosure; restricted non-force push;
minimal external-trace audit; source-free attested artifact; and retained
manual canary/rollback evidence.  This wave owns the first seven through
TASK-0011.  The final three named later results retain explicit post-TASK-0012
owner lineage and the 189 target/489 hard reserve; thus the six-wave is not
misrepresented as v1 completion.

If optional scope threatens a target cap, stop, re-estimate, reapprove PLAN and
QA_PLAN, and shed in this exact order:

1. automated canary executable; retain manual runbook/evidence;
2. rich status/JSON UX; retain activate and immediate revoke;
3. rich audit schema/correlation; retain correlation ID, actor, scope, result,
   and expiry;
4. precomputed pack-size/history diagnostics; retain exact repo/ref/clean-tree
   validation and normal non-force rejection;
5. remote-OID prefetch/race diagnostics; retain standard non-force rejection
   and generic failure;
6. automated installer/rollback executable; retain declarative units and
   manual verified install/rollback;
7. move GitHub push to v2, leaving TOTP full-sudo authority as v1.

Mandatory controls are never shed.  No semicolon/one-line packing, collapsed
errors, cryptic names, removed security comments, test deletion, or functions
combined solely to meet SLOC/Lap limits is accepted.  Idiomatic formatting,
errors, comments, and full security tests remain required.  A mandatory target
shortfall is `requirement_gap` requiring an explicit below-1800 replan; a
forecast or candidate above 1800 is an unconditional safe stop.

## Revision 8 acceptance

PLAN/QA reconciliation must independently verify the seven-file output scope;
index/contract exact consistency; forecast/cap/90% arithmetic; six dependency
and ownership boundaries; TASK-0007 production lifecycle/seed/routing detail;
TASK-0011 caller reachability, strict push schema/admission, SO_PEERCRED UID,
live-lease/policy gating, and existing-client `ready`/`otp` preservation;
each Task's paths, fixture/elevation, concrete focused/full checks, Lap work, and
split/stop rule; intermediate/final measurement gates; later mandatory owner
lineage and 189/489 reserve; pipeline restrictions; mandatory controls;
shedding; no-compression; and active/wait/retry/classification provenance.

Any mismatch returns to its responsible gate and is classified before retry.
This PLAN authorizes no DEV, approval, product change, operational-log write,
stage, commit, push, or merge.
