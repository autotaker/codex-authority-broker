# PLAN — TASK-0020: artifact-only manual canary and exact rollback

## Authority, scope, and present blocker

This is a safety-contract and live operational-evidence plan only. It changes
no product, test, workflow, dependency, generated product, or installation
artifact; production SLOC remains 1478. There is no product DEV, counted Lap,
product `REVIEW_RESULT.md`, or product `QA_RESULT.md`. Main alone may operate
the approved fixture and own evidence Git; an independent Reviewer who did not
operate it writes `EVIDENCE_REVIEW.md` from the approved TASK-first matrix.

Live execution requires Main approval of this PLAN and the TASK-first
`QA_PLAN.md` plus independent plan review. Main's approved host-context
`sudo -n true` returned exit 0 with no output; that establishes availability,
not scope. Approval of the corrected planning packet authorizes only the
digest-pinned `preflight` invocation below as a bounded `live-e2e` feasibility
case. Independent review of its sanitized Q20-02 record must PASS before Main
may authorize the separate `canary` invocation. This removes the circular
gate: documentary Q20-02 review remains a precondition to the canary, while
the explicitly bounded Q20-02L live preflight is authorized by plan approval.

## Immutable artifact gate

Main uses only artifact `codex-authority-linux-amd64` from successful GitHub
Actions run **29720021660**, workflow `.github/workflows/release.yml`, ref
`refs/heads/main`, source commit
`09487b104f32cad23a695ec3f1a0c7e7a68e6163`, and archive SHA-256
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`.
In a new empty protected staging directory, before any root operation, Main:

1. binds the GitHub run/API record to the exact repository, successful
   conclusion, workflow, run/attempt, main ref, full source commit, and exact
   artifact name; downloads that run artifact without a mutable tag;
2. requires the outer artifact to contain only
   `codex-authority-linux-amd64.tar.gz` and its uploaded `SHA256SUMS`, verifies
   the archive digest above, and runs GitHub attestation verification against
   `autotaker/codex-authority-broker`, then checks the attested subject digest,
   signer workflow, run, main ref, and full source commit exactly;
3. lists without extracting and requires seven regular, relative, non-link
   archive members exactly: `SHA256SUMS`, the three `bin/codex-authority*`
   executables, and the PAM, sudoers, and systemd declarative files named by
   TASK-0019; rejects hidden, absolute, traversal, duplicate, or extra members;
4. requires exactly six unique sorted payload entries in archived
   `SHA256SUMS`, obtains that file only with `tar -xOf`, and byte-compares its
   stream with the uploaded checksum file. For each exact allowlisted payload,
   `tar -xOf` streams its bytes directly into SHA-256 verification against the
   corresponding checksum entry. No archive payload is materialized outside
   the namespace; the one and only material extraction occurs later inside
   private tmpfs.

Wrong run/workflow/ref/commit/name/digest, altered bytes, member/checksum
mismatch, an unverified copy, local build, or source-checkout binary is a
pre-entry rejection, never a substitute. Any ambiguity stops the Task.

GitHub run/API and attestation checks remain external Q20-01 gates and execute
before root staging. Main captures the empty staging parent, then a platform
administrator creates `/var/tmp/codex-authority-task0020` and its empty
`rootfs` as root:root mode 0700; places only the verified archive, uploaded
checksum file, and the host mount/PID namespace link texts there as
root-owned, non-group/world-writable fixed inputs; and stages the exact
reviewed runbook at
`/var/tmp/codex-authority-task0020/CANARY_RUNBOOK.sh`, root:root and not
group/world writable. The runbook source is
`tasks/TASK-0020/CANARY_RUNBOOK.sh`, SHA-256
`4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`.
Main byte-compares installed/source copies and verifies this digest before
either invocation. Any setup variance blocks execution.

The elevation rule admits only these two exact argument vectors (with the
root-owned runbook and fixed staging path above), not a shell or another
runbook mode:

```text
/usr/bin/unshare --mount --pid --fork --kill-child --mount-proc=/proc /var/tmp/codex-authority-task0020/CANARY_RUNBOOK.sh preflight /var/tmp/codex-authority-task0020
/usr/bin/unshare --mount --pid --fork --kill-child --mount-proc=/proc /var/tmp/codex-authority-task0020/CANARY_RUNBOOK.sh canary /var/tmp/codex-authority-task0020
```

Before use, Main records that the pre-existing `sudo -n -l` policy is broader
than these vectors. TASK-0020 neither widens that policy nor misrepresents it
as a capability boundary: Main's approved operation is limited to the two
reviewed vectors and invokes them with `sudo -n --` and no redirection to a
secret-bearing file. The administrator-owned setup/removal procedure and its
parent comparator are separately captured; the exact reviewed program, not
namespace isolation or sudo policy, is the control on actual host-root use.

The exact secret-free outer setup/teardown is
`tasks/TASK-0020/STAGE_RUNBOOK.sh`, SHA-256
`79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`,
invoked only as `sudo -n -- /bin/bash` followed by its absolute repository
path and the single literal mode `setup` or `cleanup`. It uses the already
verified download at `/tmp/task0020-artifact-29720021660.AwGYdh` and the
reviewed repository runbook. Main captures the sorted, one-level `/var/tmp` metadata listing before
setup, then performs only these writes as root: create the literal
`/var/tmp/codex-authority-task0020` and empty `rootfs` directories mode 0700;
install the archive, uploaded checksum file, and runbook there root:root with
modes 0600, 0600, and 0500; write the pre-unshare host mount/PID namespace IDs
and the captured parent listing as root-owned non-writable files; byte-compare
the staged inputs with their sources; and recheck the archive and runbook
digests. No glob, caller-selected destination, shell evaluation, or other
host path is admitted. This setup is itself recorded as a bounded host
mutation and is not included in an in-namespace rollback PASS.

After both namespace runs, outer teardown first proves that no process, mount,
or open namespace holder refers to the literal stage path, that `rootfs` is an
ordinary empty root-owned directory, and that every other stage member is one
of the fixed setup inputs. It then removes only
`/var/tmp/codex-authority-task0020` with a one-filesystem boundary and compares
the new sorted one-level `/var/tmp` metadata listing byte-for-byte with the
stored pre-setup listing. A symlink, extra member, mounted subtree, live
holder, nonempty `rootfs`, removal error, or parent mismatch blocks rollback;
cleanup never follows an unknown path. The bounded setup/teardown command text
and exits are retained in `CANARY_RESULT.md`; raw directory output is not.

## Frozen preflight, isolation, and evidence handling

`CANARY_RUNBOOK.sh` is the frozen fail-fast launcher, comparator, canary, and
idempotent cleanup program. Its unconditional trap preserves the incoming
failure, stops/reaps the broker, overwrites/removes the transient secret,
unmounts in reverse order, deletes tmpfs contents, and performs exact host
comparison. The only host filesystem inputs are the fixed root-owned runbook
and staging tree; the script makes no host file write. Immediately after
`unshare` it verifies EUID 0, `NoNewPrivs: 0`, PID 1, and distinct mount/PID
namespace IDs, then makes `/` recursively private before any mount or fixture
mutation.

The `preflight` mode exercises the exact launcher in a disposable namespace and
prove that its effective UID is 0, `/proc/self/status` reports `NoNewPrivs: 0`,
new mount and PID namespace identities differ from the host, mount propagation
is private, and required tmpfs mounts can be created and removed. Before any
seed, broker, readiness, OTP, or lease exists, install a root-owned,
namespace-private, syntactically valid preflight PAM/sudo configuration for
only the harmless fixed command and prove that the platform's real setuid
`sudo` traverses real PAM successfully. Then remove that preflight policy and
prove it absent before installing the archived PAM/helper path. Failure of
EUID, `NoNewPrivs`, namespace, tmpfs, setuid, PAM, ownership, cleanup, or host
equality proof blocks the `canary` mode; the earlier `sudo -n true` is not a
substitute. Preflight evidence is a live-e2e record independently reviewed
under documentary Q20-02 before canary authorization.

If the real sudo/PAM probe fails, its stderr remains mode-0600 in tmpfs and is
never emitted. The runbook may retain only a bounded diagnostic category:
1 password/conversation, 2 setuid/`NoNewPrivs`/`nosuid`, 3 sudoers/policy,
4 PAM/account, or 9 unclassified, plus a digest of that constant category.
This is failure classification only and cannot convert Q20-02L to PASS.

From the host, capture SHA-256 plus type, link target, size, mode, UID/GID, and
sorted directory entries for `/etc/passwd`, `/etc/group`, `/etc/shadow`,
`/etc/gshadow`, `/etc/pam.d`, `/etc/sudoers`, `/etc/sudoers.d`, relevant
`/run` paths, and every fixture installation parent including `/usr/local` and
`/usr/local/bin`. Record absent paths explicitly. Also record bounded host
mount/PID/socket probes. The post-state command and field-for-field comparator
are fixed before elevation; any omitted surface or unknown cleanup blocks
entry as `environment_issue`.

Inside either mode, the runbook mounts tmpfs at the precreated empty `rootfs`,
binds host `/usr` and any non-merged `/bin`, `/sbin`, `/lib`, `/lib64`
read-only, and mounts only enumerated writable tmpfs for `/etc`, `/run`,
`/usr/local`, `/var`, `/tmp`, `/dev`, `/artifact`, `/evidence`, and `/input`;
`/proc` is read-only and the minimal device nodes live on tmpfs. It chroots
before authentication. The sole material archive extraction is to tmpfs
`/artifact`; all six installed inputs are rehashed. It validates sudoers with
`visudo`, systemd input with the platform verifier, and PAM by the real
preflight and artifact-backed executions. The fixed command grant is
authentication-requiring (not `NOPASSWD`) so real PAM must invoke the helper;
the archived `timestamp_timeout=0` supplies no-cache behavior.

Raw audit and command material stays in root-only tmpfs with bounded size and
mode 0600. It is reduced immediately to case IDs, UTC times, bounded probe
names, exits, counts/types/relationship assertions, and SHA-256 digests. Never
retain or print a seed, OTP, request/response payload, token, key, credential,
environment, lease identifier, internal error text, or unbounded output. The
evidence proof is structural: seed and OTP exist only in the two root-only
tmpfs files and a direct pipe; all CLI/PAM command streams are discarded; each
audit line must have exactly the five public fields and constrained types; and
the only retained runbook lines follow the fixed case/result/count/digest
grammar. A literal category scan is defense in depth, not a claim that regexes
can recognize arbitrary unknown secrets. Independent review scans the retained
Markdown and rejects any raw JSON or command output. A genuine leak invalidates
the evidence and triggers protected cleanup.

## Main-owned live sequence

After all gates pass, Main executes one uninterrupted fixture sequence:

1. Create disposable dedicated and distinct nonroot identities only in the
   tmpfs identity databases. Require the dedicated numeric UID = GID != 0.
   Give the distinct UID its distinct primary GID and only the dedicated
   socket GID as a supplementary group: it must pass mode-0660 pathname DAC
   while remaining a different `SO_PEERCRED` UID.
   Independently generate a real TOTP secret, write only its encoded seed to
   `/etc/codex-authority/seed.json` as root-owned mode 0600, never display it,
   and start the checksum-matching broker on fixed
   `/run/codex-authority.sock` with its protected audit stream.
2. Record a conservative TOTP boot-floor ceiling only after each new broker's
   socket is ready. Before every successful activation after that start, wait
   at most 31 seconds until the real 30-second counter is strictly greater
   than the ceiling (therefore greater than the process boot floor). Generate
   the six-digit code independently with Python stdlib from a root-only tmpfs
   secret and pipe it directly to the CLI; neither value enters argv,
   environment, output, or evidence. As the dedicated identity, perform
   actual `ready` and `otp`. Prove via real Unix `SO_PEERCRED` that UID 0, the
   distinct identity, and claimed-UID/payload substitution are denied and
   create no authority.
3. During that lease, run the one narrow harmless command through two separate
   actual `sudo` invocations, the real PAM stack, and the archived helper.
   Fixture sudo receives stdin from `/dev/null` and uses a prompt-free PAM
   stack; it deliberately omits sudo's `-n` shortcut because attempt 3 proved
   that shortcut rejects before PAM with a password/conversation category.
   No password, conversation, terminal, or interactive input is available.
   The fixture's `timestamp_timeout=0`, empty tmpfs timestamp store, process
   probes, and before/after audit counts must prove no sudo timestamp, helper,
   or local permit reuse. Each invocation succeeds and causes exactly one new
   `authorize` request.
4. Preserve the real immutable activation deadline, do not change clocks,
   shorten the lease, restart, or manipulate state, and wait for the natural
   **300-second** expiry. With empty timestamps and no surviving helper, a new
   actual sudo must fail closed and add exactly one fresh authorize request.
5. Complete a new real readiness/TOTP activation and require a fresh sudo
   allow with one request. Stop the broker, clear/verify no timestamps, and
   require fresh sudo denial. Start a new broker process against the same
   fixture installation/seed but no runtime lease and require denial again.
   Only another real readiness/TOTP sequence may restore a one-request allow.
   Record bounded old/new PID and start-time relationships, not lease IDs.
6. Strictly parse every protected audit line. Each backend-admitted operation
   has exactly
   `correlation_id`, numeric `actor_uid`, `scope`, `result`, and
   `lease_expiry`; actor/scope/result/counts match the real calls, ready and
   all denies have null expiry, and related OTP/authorize allows share one
   immutable non-null expiry. Retain only bounded relationships/counts/types
   and digests. Root/distinct/claimed-UID probes must pass socket pathname DAC
   where applicable but fail the pre-admission peer gate with **zero backend
   audit events**. Sudo while the broker is stopped is an unavailable
   transport and also has **zero events**. Fresh-broker authorize denial is
   admitted and therefore has exactly one five-field deny event. These three
   oracles are reconciled separately; zero-event transport failures are not
   mistaken for missing backend audit.

Any wrong-actor allow, cached/no-request allow, count ambiguity, post-expiry
allow, lease surviving stop/restart, or exact verified-artifact behavior
failure stops functional testing and proceeds directly to cleanup. An
artifact `implementation_defect`/`regression` requires a separately approved
product Task; this Task may not patch, rebuild, replace, or weaken a case.

## Mandatory teardown and exact rollback

The cleanup trap runs after success or any post-entry failure: terminate and
reap every fixture PID; remove broker/socket, sudo timestamps, protected logs,
seed, identities, PAM/sudoers/systemd policy, command grant, binaries and
staging material; unmount tmpfs in reverse order; destroy namespace holders;
and remove only the enumerated disposable host-side staging directory. Probes
must show no fixture PID, mount, namespace holder, socket, timestamp, log,
identity, seed, binary, or policy remains.

Re-run the exact frozen host capture and require byte-identical hashes and
field-identical metadata, link targets, absent-path states, and sorted
listings. Controlled tmpfs markers exercise every cleanup class, and a retained
marker or synthetic comparison difference must fail rollback. Any residue,
unknown state, or host difference is a blocking `environment_issue`; continue
safe rollback attempts but never report partial PASS.

## Completion and stop rules

`CANARY_RESULT.md` records Q20-01 through Q20-11 with immutable artifact
identity, UTC ordering, bounded probe/exit/count evidence, sanitized digests,
negative-control results, and exact pre/post equality—never full raw output or
secrets. The independent Reviewer checks it against the already fixed
`QA_PLAN.md`, including mutation detection, operator separation, unchanged
product paths/SLOC 1478, and Main-only evidence Git, then writes
`EVIDENCE_REVIEW.md`.

Completion requires every positive, denial, redaction, natural-expiry,
lifecycle, cleanup, and equality oracle to PASS plus the final v1
requirement-by-requirement audit. Missing narrow elevation is
`permission_issue`; unavailable safe namespace/tmpfs/PAM/sudo/TOTP/expiry or
rollback is `environment_issue`; an unprovable mandatory property is
`requirement_gap` or `qa_plan_defect`. None permits partial, simulated,
unit-only, prior-Task, source-built, or persistent-host substitution.

The runbook owns the live positive path, real wrong-peer and unavailable-
transport zero-event controls, fresh-broker denial, strict real-event parsing,
secret detector, and exact cleanup. Q20-01 reuses the already exercised strict
external verifier and its wrong-repository/workflow/ref/source/run rejection
controls against the fixed main artifact; no privileged dry mutation or
host-side payload extraction is added. Documentary Q20-02 independently checks
the two runbook digests, wrong entry arguments, declared paths, trap, and
comparators before authorization. These are proportional controls for a
zero-product operational evidence Task; synthetic policy/parser mutation
suites are not v1 canary substitutes or extra approval gates.

Main has added `PLAN_REVIEW.md`, `STAGE_RUNBOOK.sh`, and
`CANARY_RUNBOOK.sh` to the identical TASK/backlog zero-product evidence path
sets. A fresh independent review must verify those current paths and replace
the earlier FAIL; the earlier review record is never rewritten or hidden.

**Planner verdict: the digest-bound preflight and canary route is executable,
but only the bounded preflight may follow corrected plan approval. The canary
remains unauthorized until Q20-01, preflight/evidence review, exact inner and
outer rollback, and a fresh independent plan-review PASS.**

## Post-completion oracle correction

References above to the "current" runbook bindings describe the reviewed and
executed authorization lineage at the time of the live attempts: STAGE
`79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`
and CANARY
`4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`.
After the owner-run manual E2E, the repository runbooks became STAGE
`5a0efd62c4f3393cd59b76930c2b76490fd61d81bfffea548eb9dd1c5d199b9b`
and CANARY
`b71ddfce0afaa8c373e0bcd67ddfd35027ab4010bd5a87a19611921c976a0629`.
That correction removes only the invalid `/run` root inode-size input from
Q20-11; it retains explicit fixture-owned path comparison. The corrected
script was not executed, and the owner waived a rerun while preserving the raw
FAIL as an approved `qa_plan_defect` exception. The current manual pins the
corrected hashes; the historical authorization above does not authorize a new
run.
