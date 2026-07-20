# TASK-0020 QA plan — TASK-first live canary and rollback gate

## Independence, classification, and authority

This QA plan was derived from `TASK-0020/TASK.md` and the fixed planning
packet before reading any `TASK-0020/PLAN.md`.  The packet is bound to
`TASK.md` SHA-256
`ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
and `backlog.json` SHA-256
`3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`.

This is a safety-contract plus live operational-evidence Task.  It authorizes
no product, test, workflow, dependency, or generated-product change; no
product DEV, product `REVIEW_RESULT.md`, product `QA_RESULT.md`, or counted
product Lap exists.  Main alone performs the approved host operation and owns
Git.  A separate evidence Reviewer, who did not operate the fixture, reviews
the resulting `CANARY_RESULT.md` and writes `EVIDENCE_REVIEW.md`.  The
Reviewer must not edit product, the canary record, or Git, and must derive the
verdict from the approved matrix rather than Main's conclusions.

All live work is blocked until TASK-0019 has a successful main-run artifact
and provenance PASS, PLAN and this TASK-first QA plan have Main approval, and
independent plan review authorizes the exact next phase.  Planning approval
may authorize only the digest-pinned Q20-02L live preflight.  The separate
canary invocation remains blocked until documentary Q20-02 independently
reviews Q20-02L PASS and Q20-01 is complete.  Unknown cleanup, incomplete
pre-state, unavailable elevation, unsafe host scope, or an inability to run a
mandatory negative control is not waivable.
No failure may be worked around by patching/rebuilding the artifact, using a
source binary, widening host privileges, weakening an assertion, or retaining
fixture state.

## Fixed evidence and secrecy rules

The sole executable input is artifact `codex-authority-linux-amd64` from
successful run `29720021660` of `.github/workflows/release.yml` on
`refs/heads/main`, source commit
`09487b104f32cad23a695ec3f1a0c7e7a68e6163`, archive SHA-256
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`.
The outer file, exact seven members, exactly six payload checksum entries,
and GitHub attestation must all verify again before extraction into the
fixture.  No mutable tag, other run/ref/commit, local build, source checkout
binary, or unverified copied archive is equivalent evidence.

The reviewed operational program is
`tasks/TASK-0020/CANARY_RUNBOOK.sh`, SHA-256
`4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`.
The exact outer staging/rollback program is
`tasks/TASK-0020/STAGE_RUNBOOK.sh`, SHA-256
`79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`.
The host already has broader noninteractive administrator authority.  This
Task neither widens it nor calls it a capability boundary.  The safety
control is Main approval of one of the two exact PLAN-listed `unshare`
argument vectors, the root-owned fixed staging path, byte equality and digest
binding to this independently reviewed program, fail-fast behavior, and exact
outer cleanup.  No claim that sudo itself admits only these vectors is part of
the PASS oracle.

Evidence records, per case: case ID; UTC start/end; operator and fixture
identity labels; artifact name/run/ref/full source commit/archive digest;
bounded command or probe name (not full output); exit status; before/after or
count deltas; sanitized result type; and a digest of retained sanitized
evidence.  It may include no raw request or response payload, OTP, seed,
token, key, credential, environment, lease identifier, internal error text,
or unbounded command output.  Secret scans report only scanner identity,
bounded category counts, exit status, and evidence digest—never matching
content.  A leak invalidates the evidence and triggers immediate protected
cleanup and classification; redaction after collection cannot turn it into
PASS.

The exact audit object has only `correlation_id`, numeric `actor_uid`,
`scope`, `result`, and `lease_expiry`.  The evidence may retain only bounded
counts, types, relationship assertions, and digests of those events; it must
not reproduce complete events where doing so violates the secrecy boundary.

## TASK-first acceptance and mutation matrix

| ID | Acceptance procedure and negative controls | `qa_execution_mode` and rationale | PASS oracle / stop condition |
| --- | --- | --- | --- |
| Q20-01 | Outside the privileged runbook, in an empty non-root staging area, re-bind the live GitHub API/run and attestation records to the exact repository, successful workflow/run/attempt, `refs/heads/main`, full source commit, artifact name, archive digest, signer workflow, and subject digest. Reverify the two-file outer artifact, stream-only seven-member/six-checksum checks, and no host payload extraction. Review and bind the already recorded strict verifier evidence that wrong repository/workflow/ref/source/run selections are rejected for this same fixed artifact; no new privileged or host-payload mutation is required. | `live-e2e` — current GitHub run/attestation truth is live and deliberately remains outside `CANARY_RUNBOOK.sh`; prior strict negative controls are evidence-reviewed only when bound to this exact artifact/verifier. | The exact immutable binding and archive pass, and the existing wrong-identity controls are complete and applicable. No local/source artifact, mutable identity, other run, or unverified copy substitutes. Unavailable live verification is `environment_issue`; an incapable or unbound verifier record is `qa_plan_defect`/`requirement_gap`; a mismatch in the named genuine artifact is `implementation_defect`/`regression`. |
| Q20-02 | Independently review authorization and exact artifacts without entering a namespace: current TASK/backlog evidence paths; source/installed runbook byte equality and fixed digest; root ownership/non-writability; fixed staging path and empty-rootfs setup; the two exact PLAN-listed argument vectors; fail-fast trap, full mount/copy/install/cleanup/comparator text; host pre-state capture; and administrator-owned outer staging removal/parent comparison. Record the pre-existing broader `sudo -n -l` policy and confirm TASK-0020 adds or changes no sudo rule. A one-byte disposable runbook copy must miss the digest, and wrong mode/path/argument values must fail the runbook's own entry checks; do not invoke an unapproved privileged vector merely to test broad sudo. | `evidence-review` — this case decides whether the exact Q20-02L program and invocation are reviewable and authorized. It does not claim sudo is the capability boundary and performs no privileged fixture action. | Exact text/digest, vectors, staging, capture, cleanup, and no-policy-widening evidence are complete. Independent plan approval may authorize Q20-02L only. Missing or changed text/digest/capture/cleanup is `qa_plan_defect`/`environment_issue`; unavailable authority is `permission_issue`. The canary is still forbidden. |
| Q20-02L | After Q20-02 planning approval, invoke only the digest-pinned outer `preflight` vector with administrator `sudo -n --`. Inside the fixture, prove EUID 0, `NoNewPrivs: 0`, PID 1, distinct mount/PID namespaces, private propagation, enumerated tmpfs/read-only bindings, and real setuid sudo plus real prompt-free `pam_permit` for the harmless fixed command. Fixture sudo deliberately omits `-n` and reads stdin only from `/dev/null`; the stack has no password, conversation, terminal, or interactive input path. Require removal of preflight PAM policy, unconditional inner cleanup, exact in-runbook host equality, and administrator-owned outer cleanup/equality. The bounded runbook line named `Q20-02` maps explicitly to Q20-02L; no seed/canary action is permitted. | `live-e2e` — namespace, mount, setuid, prompt-free PAM, teardown, and equality are real host facts. | Exact outer invocation exits 0; inner sudo reaches and passes real PAM without prompting or input; sanitized record and Q20-11 bind to the reviewed program; no residue exists. Documentary Q20-02 must review PASS before canary authorization. Any mismatch is `permission_issue`/`environment_issue`; an interactive path or unreviewable record is `qa_plan_defect`. |
| Q20-03 | In the separately authorized `canary` invocation, require the same launcher checks and exact enumerated mount map: rootfs, `/etc`, `/run`, `/usr/local`, `/var`, `/tmp`, `/dev`, `/artifact`, `/evidence`, and `/input` are tmpfs as declared; host `/usr` and non-merged runtime paths are read-only bindings; `/proc` and minimal devices have the fixed restrictions. Materialize the archive exactly once into tmpfs `/artifact`, recheck six digests, and retain only a mount-map digest. Host equality and absence after cleanup are the negative isolation control; no deliberate host write is authorized. | `live-e2e` — actual kernel namespace/mount and extraction placement require the live OS. | The reviewed map is exact, all writable fixture paths are private, no host payload is materialized, and Q20-11 proves no change. Any extra writable/unoverlaid fixture target, shared propagation, or host difference is blocking `environment_issue`/`qa_plan_defect`. |
| Q20-04 | Validate the genuine archive PAM/sudoers/systemd inputs with `visudo`, the platform systemd verifier, and real PAM; install only the fixed archived declarations, harmless fixed command grant, and checksum-matching three binaries in tmpfs. Bind validator exits and installed digests to Q20-01. | `live-e2e` — genuine platform parsing, PAM load, installation, and binary selection are live facts. | All genuine declarations validate, the command grant is exact, and all executables match the verified archive. A wider/different policy or digest is rejected before broker start. Genuine validator unavailability is `environment_issue`; an artifact mismatch/invalid declaration is classified from evidence. |
| Q20-05 | Create owner UID/GID 42020/42020 and distinct UID/GID 42021/42021 only in tmpfs. Give the distinct identity only owner GID 42020 as a supplementary group so it passes mode-0660 socket DAC. Use a root-owned mode-0600 fixture seed. For each broker process, record a conservative boot-floor counter ceiling after socket readiness and wait at most 31 seconds until the current counter is strictly greater before piping a new TOTP directly to the CLI. Root and the DAC-capable distinct UID must fail SO_PEERCRED pre-admission. For the distinct peer's claimed-owner raw request, accept only EOF, connection reset, or broken pipe as transport rejection; any response byte is failure. Every wrong-peer subcase requires zero backend audit and has its own bounded Q20-05 label. | `live-e2e` — real UID/groups, socket DAC, SO_PEERCRED, transport race outcomes, boot floor, and TOTP require the live fixture. | Owner activation succeeds only above that process's floor; secret/code is absent from argv/environment/output/evidence. Root, distinct, and claimed-owner probes fail closed with zero events; claimed-owner yields only EOF/reset/broken-pipe and never a response byte. Labels identify broker/root/distinct/claimed/audit/activation failure without raw output. Early-counter rejection is expected replay defense; an ambiguous response/count is `qa_plan_defect`/`environment_issue`, not PASS. |
| Q20-06 | During the active lease, make two separate actual fixture `sudo -- /usr/bin/true` calls through the archived prompt-free `pam_exec` stack and helper, each with stdin from `/dev/null` and no password, conversation, terminal, or interactive input path. The outer administrator route remains `sudo -n`; only fixture sudo omits `-n` so PAM executes. The genuine archived `timestamp_timeout=0`, initially empty tmpfs sudo state, one-event count delta per call, and post-call no-helper probe are the cache controls. | `live-e2e` — real prompt-free sudo/PAM/helper execution proves per-call authorization. | Both calls succeed without prompting/input and each adds exactly one authorize allow; no helper survives and no reusable timestamp/local permit is accepted. Any prompt/input path, zero/multiple event delta, or cached allow fails according to the genuine evidence boundary. |
| Q20-07 | Bind the observed immutable allow expiry to the activation audit relationship, use the two Q20-06 calls as pre-deadline controls, and wait naturally until host time is strictly after that deadline, with a 305-second bound and no clock/state/restart/lease shortening. Then make a fresh real sudo with no helper and require one newly admitted authorize denial. | `live-e2e` — the real 300-second deadline and PAM denial cannot be simulated. | Natural post-deadline sudo fails closed and adds exactly one five-field authorize deny with null expiry. Timeout, clock ambiguity, cached success, deadline extension, or wrong count blocks PASS. |
| Q20-08 | Above each process's boot-floor ceiling, perform a second real activation and one-request sudo allow. Stop/reap the broker and require sudo failure with no socket/helper and zero new audit events in the stopped process stream. Start a demonstrably fresh broker; before activation require a fresh sudo denial with exactly one admitted five-field event. Wait until counter > the new process's separately recorded boot-floor ceiling, then perform a third real readiness/TOTP activation and one-request sudo allow. | `live-e2e` — stop, transport failure, new runtime state, replay floor, and PAM outcomes are real process properties. | Stopped broker is a zero-event unavailable transport; fresh broker carries no lease and emits one denial; only the post-floor TOTP restores allow. A surviving lease is `implementation_defect`/`regression`; failure to distinguish processes/floors/events is `qa_plan_defect`. |
| Q20-09 | Strictly parse protected audit lines for backend-admitted operations only and reduce them to bounded counts/types/relationships/digests. Every admitted ready/OTP/authorize has exactly the five named fields and one event; actors, scopes, results, null rules, and related allow expiry match. Separately reconcile Q20-05 wrong-peer pre-admission rejections and Q20-08 stopped-broker transport failure as zero-event cases, while the fresh-broker authorize denial has one event. | `evidence-review` — strict parsing of the genuine live stream and cross-case count reconciliation are independently auditable. | Genuine admitted and zero-event classes reconcile separately. Any schema, relationship, or count discrepancy is artifact defect/regression; an ambiguous/crossed class or non-strict parser is `qa_plan_defect`. |
| Q20-10 | Prove secret-free evidence structurally: seed and OTP exist only in two root-only mode-0600 tmpfs files and a direct pipe; they never enter argv or environment; CLI/PAM stdout/stderr is discarded except the protected five-field broker audit stream; every audit line passes the exact schema/type/relationship checks; and the only retained runbook output matches the fixed `case/result/count/digest` grammar. Treat the built-in literal scan as defense in depth only. Independently inspect retained Markdown and reject raw JSON, raw command output, or any secret-bearing value. | `evidence-review` — the information-flow, exact schema, bounded emit grammar, and retained Markdown are reviewable; no claim is made that a regex detects arbitrary unknown secrets. | Every declared flow terminates in protected tmpfs/discard or bounded public evidence, audit/output grammar is exact, literal scan passes, and independent review finds no forbidden retained value. A genuine leak invalidates evidence and triggers cleanup without quoting the value; an unaccounted output path is `qa_plan_defect`. |
| Q20-11 | For both Q20-02L and canary, exercise the real unconditional trap: stop/reap broker, overwrite/remove transient secret files, unmount the enumerated stack in reverse order, empty rootfs, and compare the complete in-runbook host capture exactly. After invocation, the administrator cleanup rejects a symlink, extra stage member, mounted/nonempty rootfs, live holder, removal error, or parent mismatch; on the genuine path it removes only the enumerated fixed stage and exactly matches the captured `/var/tmp` parent state. | `live-e2e` — real cleanup/equality and the stage program's fail-closed guards are operational facts. | Both inner and outer rollback are exact; no PID, mount, socket, timestamp, log, identity, seed, binary, policy, rootfs content, or staging tree remains. Any guard failure, difference, or residue is blocking `environment_issue`; unknown cleanup is never attempted. |
| Q20-12 | Review `CANARY_RESULT.md` against the exact required set Q20-01, Q20-02, Q20-02L, and Q20-03–Q20-11, current packet/runbook digests, operator/reviewer separation, unchanged product SLOC 1478, no product DEV/Lap/REVIEW_RESULT/QA_RESULT, declared PLAN_REVIEW/STAGE/CANARY evidence paths, and Main-only Git. | `evidence-review` — completion, scope, ownership, and consistency are documentary. | Every required record is present, bound, secret-free, and consistent; all live cases PASS and rollback is exact. Any omission/contradiction prevents `EVIDENCE_REVIEW.md` PASS. |

## Execution order, classification, and stop rules

Q20-01 and documentary Q20-02 phase A are hard preconditions to privileged
work.  A corrected independent plan-review PASS may authorize only Q20-02L.
Main runs that exact preflight, always completes its Q20-11 cleanup, and the
independent Reviewer completes documentary Q20-02 phase B from its sanitized
record.  Only then may Main authorize the separate canary invocation.  The
current Q20-01 external verifier evidence must also PASS before canary
authorization.  Main then runs Q20-03 through Q20-08 in order, reducing protected results for
Q20-09/Q20-10, and must run Q20-11 after success or any post-entry failure.
Q20-12 is the final evidence review.  A live case cannot PASS from unit,
prior-Task, source-built, simulated PAM/sudo, or document-only evidence.  An
evidence-review case cannot PASS from Main's assertion without its case-bound
sanitized record.

Main stops the functional sequence immediately on artifact ambiguity,
namespace leakage, wrong identity acceptance, secret exposure, unexplained
request count, cached allow, post-expiry/restart allow, or product-behavior
defect; cleanup Q20-11 still runs.  Classification is evidence-led:

- `permission_issue`: the approved exact noninteractive invocation is
  unavailable or insufficient; this label does not imply the pre-existing
  host sudo policy is narrow.
- `environment_issue`: safe namespace/tmpfs/PAM/sudo/TOTP/natural-expiry/
  process-control/cleanup/exact-rollback execution is unavailable or
  incomplete.
- `implementation_defect` or `regression`: the exact verified artifact
  violates the fixed product or previously accepted behavior under a valid
  fixture.
- `requirement_gap`: a mandatory v1 property cannot be proven without
  changing scope or weakening the safety boundary.
- `qa_plan_defect`: a case, fixture, parser, negative control, or evidence
  rule cannot prove its stated oracle.

An artifact defect or regression must open a separately approved product
Task.  TASK-0020 must not repair or replace the artifact.  A requirement or
QA-plan gap returns to planning.  Permission/environment failures remain
blocked until the exact safe prerequisite exists; they never become a partial
PASS.  Because this Task has zero product changes, product candidate/tree QA
and `qa_carry_forward` do not apply.

## PLAN reconciliation and planning verdict

The TASK-first matrix was fixed before PLAN review and is now reconciled to
the changed packet and the independent review's findings.  TASK and backlog
declare the same zero-product evidence paths, including `PLAN_REVIEW.md`,
`STAGE_RUNBOOK.sh`, and `CANARY_RUNBOOK.sh`; finding 5 is resolved, and the
current PLAN accurately records that correction.

Finding 1 is resolved in this QA contract by documentary Q20-02 and separately
authorized live Q20-02L.  Planning approval authorizes only the exact
preflight; documentary review of its secret-free PASS and exact rollback is a
hard prerequisite to canary authorization.  `CANARY_RUNBOOK.sh` currently
emits the bounded preflight line as `Q20-02`; `CANARY_RESULT.md` must map that
line explicitly to Q20-02L, or the runbook must be re-digested after relabeling
and independently re-reviewed.

Finding 2 is materially addressed by the declared exact programs.  The
root-owned outer setup/cleanup is fixed by the STAGE runbook digest and literal
paths; the inner launcher, mount map, authentication sequence, trap, and
comparator are fixed by the CANARY runbook digest.  The current broad
pre-existing sudo policy is recorded and unchanged.  PLAN wording that the
"elevation rule admits only" the two vectors is accepted only as a statement
of Main's approved invocation, never as a technical restriction imposed by
sudo.  Q20-02 must verify both source/staged digests and exact invoked argv.

Finding 3 is resolved by three disjoint audit classes: every backend-admitted
operation produces one exact five-field event; root, supplementary-group
distinct UID, and claimed-UID wrong-peer probes pass/bypass pathname DAC but
fail pre-admission with zero backend events; stopped broker is a zero-event
unavailable transport; and fresh-broker authorize denial is admitted and
produces one event.  Finding 4 is resolved by recording a conservative counter
ceiling after each broker socket becomes ready and waiting up to 31 seconds
for `current_counter > ceiling` before every successful activation, including
after restart.  The wait is secret-free and distinct from the natural
300-second expiry.

The latest frozen runbooks also resolve the re-review's additional safety and
oracle findings.  Every mount is registered immediately after creation;
recursive bind targets and submounts are individually remounted and verified
read-only; cleanup uses recursive unmount, proves the entire rootfs mount table
empty, and only then performs `find -xdev`.  The claimed-owner raw peer must
observe only EOF, connection reset, or broken pipe and must reject any
response byte.  Every authorize allow expiry
must equal the activation's exact `LEASE_EXPIRY`, while denials remain null.
These are bound to the current STAGE/CANARY digests above and require the next
independent review before execution.

The first Q20-02L attempt against the preceding reviewed digests is retained
as a safe `environment_issue`: it stopped before PAM because tmpfs
`/usr/local/bin` was absent, while inner and outer Q20-11 rollback both PASSed
and no fixture state remained.  `CANARY_RESULT.md` is the secret-free evidence
record.  This is no product/artifact classification.  The only fixture change
creates `/usr/local/bin` immediately after mounting tmpfs `/usr/local`; that
one-line change produced the current CANARY and STAGE digests above and
invalidated the third review's prior Q20-02L authorization.

The second Q20-02L attempt against the next reviewed digests passed Q20-01 and
Q20-03, then safely failed the real setuid sudo/PAM Q20-02 oracle before any
seed, broker, OTP, lease, or product behavior was reached.  Inner and outer
Q20-11 rollback again PASSed with no residue.  It is retained only as an
`environment_issue`/fixture-feasibility result; it supplies no product or
artifact classification.  The attempt-3 candidate added only secret-free
preflight invariants for sudo mode/owner, absence of `nosuid`, and fixture UID
plus `NoNewPrivs`, together with a tmpfs-only stderr reducer whose bounded
categories are 1 conversation, 2 setuid/NoNewPrivs/nosuid, 3 policy, 4
PAM/account, and 9 unknown.  Raw stderr is never emitted and is destroyed at
rollback.

The third Q20-02L diagnostic attempt against those reviewed digests again
passed Q20-01/Q20-03, then returned bounded category 1: fixture `sudo -n`
rejected before the intentionally prompt-free PAM stack.  Inner and outer
rollback PASSed with no residue and no product path was reached.  This is
retained only as an `environment_issue` in the fixture route, with no
product/artifact classification.  The current candidate removes `-n` only
from fixture sudo and redirects its stdin from `/dev/null`; outer administrator
`sudo -n` is unchanged.  The preflight `pam_permit` and live archived
`pam_exec` stacks contain no prompt module, so there is no password, terminal,
conversation, or interactive input path.  This change produced the current
pre-canary digests and invalidated every earlier authorization.

Q20-02L attempt 4 then PASSed against those independently reviewed digests:
Q20-01, Q20-03, mapped Q20-02 real setuid/prompt-free PAM, inner Q20-11, and
outer rollback all passed; no seed, broker, OTP, lease, or canary action ran.
Documentary Q20-02 phase B also passed.  Before authorizing canary, Main found
that successful `ready`/`otp` discarded stdout but their failure stderr—and
the TOTP generator's stderr—were not all discarded as PLAN's information-flow
oracle requires.  No canary ran.  The stream-safe candidate added only `2>&1` to
`ready` and to both sides of the `totp_pipe | otp` pipeline after the explicit
`[[ $MODE == canary ]] || exit 0` boundary; authentication, preflight,
cleanup, artifact, and all decision oracles are unchanged.

Q20-02L evidence is eligible for documentary carry-forward only if a fresh
independent reviewer verifies the complete old-to-new runbook diff, confirms
every changed command is unreachable in `preflight` mode, binds both new
runbook digests, and confirms the prior PASS/rollback record.  This is
operational-evidence reconciliation, not product `qa_carry_forward`, and it
does not itself authorize any invocation.  If reachability or scope cannot be
proven exactly, Q20-02L must be freshly reviewed and rerun before canary.

Fresh review carried Q20-02L forward and authorized one stream-safe canary
attempt.  That attempt passed Q20-01/Q20-03/carried Q20-02L/Q20-04, then the
composite Q20-05 gate safely stopped before any sudo, natural-expiry, or broker
lifecycle case.  Inner and outer Q20-11 rollback PASSed with no residue.  The
record remains a fixture `environment_issue`; it is insufficient for product
or artifact classification.  The likely failure is a valid pre-admission
transport race: a wrong peer may be rejected by EOF, connection reset, or
broken pipe before/during send.  The current candidate accepts only those
three rejection forms, still rejects every response byte, still requires zero
backend audit, and adds granular bounded Q20-05 labels.  It weakens no product
oracle.  The digest change invalidates the seventh review's single-canary
authorization.  Because all changed lines remain beyond the preflight exit,
Q20-02L may again be carried forward only after a fresh reviewer proves the
exact unreachable diff and binds the current digests; no carry-forward or
execution is authorized by this QA reconciliation itself.

The acceptance matrix uses proportional operational controls for this
zero-product Task: current external artifact/attestation re-verification plus
the already exercised strict wrong-repository/workflow/ref/source/run
controls; digest and wrong-entry-argument checks for the two reviewed
programs; the two genuine sudo calls as the cache control; real wrong-peer,
natural-expiry, stopped-broker, and fresh-broker denials; the genuine secret
information-flow proof, exact audit/output grammar, literal defense-in-depth
scan, and independent Markdown inspection; and fail-closed cleanup guards
with exact equality.  No synthetic
policy/parser/comparator suite or nonexistent negative-control appendix is an
approval gate.

**Verdict: QA PLAN RECONCILED TO THE CURRENT PEER-TRANSPORT FIXTURE DIGESTS;
all further execution remains unauthorized pending fresh independent
review.**  That review may carry Q20-02L attempt 4 forward only under the exact
unreachable-diff conditions above, or require a new Q20-02L.  Current Q20-01,
accepted Q20-02L with exact inner/outer rollback, and documentary Q20-02 PASS
then gate a separately authorized canary.  Final
Task PASS still requires every live/evidence case, a secret-free record, and
exact rollback; there is no partial verdict.

## Post-completion oracle correction

All earlier "current" digest references in this QA history refer to the
reviewed and executed STAGE
`79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`
and CANARY
`4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`
lineage. The repository now contains STAGE
`5a0efd62c4f3393cd59b76930c2b76490fd61d81bfffea548eb9dd1c5d199b9b`
and CANARY
`b71ddfce0afaa8c373e0bcd67ddfd35027ab4010bd5a87a19611921c976a0629`.
The only oracle change removes volatile `/run` root inode size while retaining
the exact fixture-owned path checks. The owner accepted the raw Q20-11 FAIL as
a `qa_plan_defect`, waived a corrected-script rerun, and did not represent the
corrected script as executed. This note preserves that exception and does not
retroactively change any raw case result or execution authorization.
