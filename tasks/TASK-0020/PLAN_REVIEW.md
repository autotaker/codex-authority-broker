# PLAN_REVIEW — TASK-0020

## Decision: FAIL

Independent plan review fails the current TASK-0020 planning packet. Live
fixture execution is not authorized. The artifact identity is consistently
bound across `TASK.md`, `PLAN.md`, `QA_PLAN.md`, and `backlog.json` to run
`29720021660`, source commit
`09487b104f32cad23a695ec3f1a0c7e7a68e6163`, and archive SHA-256
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`;
the QA packet's recorded TASK/backlog hashes also match the current files.
Those bindings do not resolve the blocking execution and oracle defects below.

## Blocking findings

1. **Q20-02 has a circular authorization gate and the wrong execution mode.**
   PLAN says no namespace operation may run until Q20-02 passes, but its own
   Q20-02 procedure must run the exact launcher in a disposable mount/PID
   namespace, create/remove tmpfs mounts, and traverse real setuid sudo and
   real PAM. QA labels Q20-02 `evidence-review`, describes it as ending before
   any live fixture starts, and only assigns `live-e2e` beginning at Q20-03.
   Thus the real privileged preflight is simultaneously forbidden until it
   passes and is not owned by a live QA case. Split authorization from
   execution: the approved plan review may authorize a precisely bounded
   preflight, and a `live-e2e` case must execute and retain evidence for that
   preflight before the canary sequence.

2. **The limited root authority and cleanup program are not exact,
   reviewable artifacts.** `sudo -n true` proves only that one sudo call
   succeeded; it proves neither that the canary launcher is authorized nor
   that its privilege is limited to the declared namespace and paths. A mount
   and PID namespace does not itself constrain host-root file capabilities;
   any un-overlaid path remains host-visible. PLAN defers the launcher,
   complete tmpfs/copy/install map, sudo timestamp location, teardown trap,
   and comparator to an operator-time “freeze,” while final evidence retains
   only a bounded command/probe name and digest. An independent reviewer
   therefore cannot verify command scope, path completeness, stop-on-error
   behavior, or that the same reviewed cleanup/comparator was executed. Bind
   a secret-free exact launcher/cleanup/comparator text (or an immutable
   digest plus independently available exact text) before approval, enumerate
   every writable path and staging parent, and prove the elevation rule admits
   only that entrypoint. Namespace isolation must not be treated as a root
   capability boundary.

3. **The audit oracle contradicts the shipped IPC boundary.** Q20-09 requires
   an audit event for every exercised ready/OTP/authorize denial, and TASK's
   live acceptance broadly requires audit evidence for exercised denials.
   However `internal/ipc/server_linux.go` rejects a wrong `SO_PEERCRED` UID
   before request decoding and before `Backend.Handle`, so root/distinct-peer
   probes produce no backend audit event. TASK-0019 explicitly preserved
   wrong-UID requests as “no-backend/no-audit.” QA reconciliation creates a
   zero-event exception only for a stopped broker, not for wrong peers. The
   current Q20-05/Q20-09 count reconciliation therefore cannot pass the exact
   artifact. Define separate, exact oracles for admitted backend operations
   (one five-field event), pre-admission peer rejection (zero backend events
   plus a bounded transport/count proof), and broker-unavailable transport
   failure (zero events). Also ensure the distinct UID can reach the socket
   DAC boundary so the claimed SO_PEERCRED test is not merely a mode-0660
   filesystem denial.

4. **Fresh-process TOTP timing is underspecified and will be flaky or fail.**
   `internal/lease.New` fixes `bootReplayFloor` to the current 30-second TOTP
   counter and `VerifyAndActivate` rejects every candidate at or below that
   floor. Consequently a freshly generated current TOTP immediately after the
   initial broker start, and especially immediately after the restart in
   Q20-08, is not acceptable merely because it is “fresh.” The plan must
   establish, without logging the secret/code, that activation uses a counter
   strictly greater than the new process's boot floor, include the bounded
   wait in timeouts, and distinguish this expected replay defense from an
   artifact defect. The existing 300-second natural-expiry wait does not solve
   the post-restart case because restart establishes a new current floor.

5. **The mandatory plan-review record is outside the declared evidence
   boundary.** Both TASK metadata and `backlog.json` require independent plan
   review but their `evidence_paths` omit
   `tasks/TASK-0020/PLAN_REVIEW.md`. Add the mandatory record to the declared
   zero-product evidence scope before approval; otherwise the required gate
   itself is an undeclared path change.

## Reviewed safety and governance boundaries

The proposed artifact-only route, one in-namespace extraction, source/local
substitution rejection, tmpfs seed mode/ownership, actual PAM/sudo/helper
calls, `timestamp_timeout=0`, natural 300-second expiry, fresh-process restart,
protected transient audit handling, secret nonlogging, unconditional teardown,
and exact host pre/post comparison are appropriate goals. Product DEV,
product REVIEW/QA results, counted product Lap, product mutation, GitHub push,
and child-agent Git remain correctly excluded. The five blockers above must
be corrected and independently re-reviewed; none can be deferred to canary
evidence while retaining an approved plan verdict.

## Re-review — 2026-07-20 current runbook candidate

### Decision: FAIL — Q20-02L is not authorized

This re-review preserves the original FAIL and evaluates the complete revised
packet plus both newly bound operational programs. The current bindings are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `17486172793458423bf9786846b925dcbeaf504f1410af3319be775c2281597f`
- `QA_PLAN.md`: `5ff18556bdb8e055ac8c2868fe2b3e59b4fa1f8281a127c89c8a0bcb53548ca0`
- `STAGE_RUNBOOK.sh`: `bc716b3cb3881aadb9a9da432bc894e9d75710a70fb13a432dc954e18f9e0fb9`
- `CANARY_RUNBOOK.sh`: `a5ba4284f06f96d5f6158dec4720382ead7edcb906a808732d52b75ca90f9a81`

The TASK/backlog hashes match the QA packet, the runbook hashes match PLAN and
QA, and TASK/backlog now declare identical evidence paths. `bash -n` passes
for both runbooks. Shellcheck reports only SC2317 informational findings for
the EXIT-trap callback in the stage script. These checks do not overcome the
unsafe failure cleanup below.

### Mapping of the five original findings

1. **Circular Q20-02 gate: resolved in the contract.** Documentary Q20-02 is
   separated from live Q20-02L, and a plan-review PASS could authorize only
   the exact `preflight` vector. Independent review of Q20-02L and its inner
   and outer rollback remains a hard gate before a separately authorized
   `canary` vector. The bounded runbook output named `Q20-02` must be mapped to
   Q20-02L as QA requires, but that label does not recreate the former cycle.

2. **Exact root program and capability account: only partially resolved and
   still blocking.** The revised packet honestly records that the host sudo
   policy is pre-existing and broader; no policy widening is proposed and
   namespace isolation is no longer represented as a root capability
   boundary. Literal paths, argument vectors, staging, digests, and outer
   parent comparison are reviewable. However the exact inner cleanup is
   unsafe on a partial mount failure:

   - `bind_ro` adds a mount to `MOUNTS` only after `--rbind`, propagation
     change, and read-only remount all succeed. If `--rbind` succeeds and a
     later command fails, the live bind is untracked.
   - The two fixed-file bind sequences have the same register-after-remount
     window.
   - Cleanup notes an unmount failure but still executes recursive
     `find "$ROOT" -mindepth 1 -delete` without `-xdev` and without first
     proving every nested mount absent. It can therefore descend through an
     untracked or failed-to-unmount bind. For `/usr`, that is a view of the
     host filesystem; top-level read-only remount is not an acceptable
     deletion guard and recursive `--rbind` submounts are not made read-only
     by the single top-mount remount.

   This violates fail-safe rollback and the rule against unknown cleanup. It
   blocks even Q20-02L. Register each mount immediately after successful
   creation, make every recursive bind and submount verifiably read-only, and
   never descend for deletion until a mount-table check proves no nested mount
   remains; a safe `-xdev` deletion of only the disposable root filesystem may
   then follow. The revised digest must receive another independent review.

3. **Audit classes: contract resolved, frozen canary oracle incomplete.** PLAN
   and QA now correctly separate one-event backend admission, zero-event
   wrong-peer pre-admission, zero-event stopped transport, and one-event
   fresh-broker denial. The distinct UID receives only the owner socket group,
   so its CLI probe can cross pathname DAC. But the claimed-owner raw-socket
   probe merely calls `recv(1)` and exits successfully for either EOF or any
   received byte; it never asserts EOF/denial. Zero audit alone does not prove
   the probe was rejected. In addition, `event_check` checks only that every
   allowed OTP/authorize expiry is a string. It never compares later allow
   values with the activation's `LEASE_EXPIRY`, so the required shared,
   immutable expiry relationship is not enforced. These are canary blockers
   even after a safe preflight exists.

4. **Fresh-process TOTP floor: resolved.** Each broker start records a
   conservative counter ceiling after socket readiness, and every activation
   waits until the current counter is greater. The 31-second bound covers one
   30-second transition and is separate from the natural-expiry wait. The code
   pipes the generated value from a root-only tmpfs secret without placing the
   value in argv or environment. This matches `bootReplayFloor` behavior,
   including restart.

5. **Plan-review evidence path: resolved.** TASK and backlog now include
   `PLAN_REVIEW.md`, both runbooks, and the later canary/evidence-review records
   in the same zero-product evidence set.

### Additional blocking canary evidence gap

The built-in Q20-10 detector does not implement the forbidden-category oracle
claimed by PLAN/QA. It scans only files below `/evidence` (the audit streams)
and recognizes a short literal regex. It cannot detect an arbitrary six-digit
OTP, raw/base64 seed, generic token/key/credential/environment value, request
payload not containing the literal `"code"`, or arbitrary internal error; it
also does not scan retained runbook output or final Markdown. Independent
Markdown inspection is useful but cannot establish that genuine transient or
retained output passed the stated exact detector. Replace this with a
reviewable category-complete sentinel/detector design that scans every
declared retained surface without printing matches, or narrow the oracle
honestly and add an independent mechanism for the uncovered surfaces.

### Re-review conclusion

The stage program's fixed-path setup, ownership/mode checks, exact-member
cleanup guard, open-holder probes, one-filesystem deletion, and `/var/tmp`
parent equality are proportionate when executed around a safely terminating
inner program. Real setuid sudo/PAM in the private chroot, the archived helper,
`timestamp_timeout=0`, two one-event sudo allows, natural expiry, broker
stop/restart, tmpfs identity/seed/audit state, and exact host comparison remain
feasible. Role separation and zero-product classification are compliant.

Nevertheless, the partial-mount cleanup path can turn a setup failure into a
host deletion attempt. That material safety defect prevents authorization of
Q20-02L. The claimed-peer rejection, immutable-expiry comparison, and secrecy
detector also prevent later canary approval. No namespace, preflight, or
canary invocation is authorized by this re-review; correct and re-digest the
runbook, reconcile PLAN/QA hashes, and obtain another fresh independent plan
review.

## Third re-review — 2026-07-20 current runbook candidate

### Decision: PASS — authorizes Q20-02L preflight only

This third review preserves both prior FAIL decisions and evaluates the newly
digested packet from the current files. The exact reviewed bindings are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `99f820338e6d193e40a6ba608bb0b33eb392fe123746effd2a54476a25e8bc3a`
- `QA_PLAN.md`: `45ecdb2ad0c95f7974165ef99a355da54c65dc518edde808d9813d3075332982`
- `STAGE_RUNBOOK.sh`: `68e975f406e108ec6b434f7c440249d6e3317936662337a31ef209ba7a34eac0`
- `CANARY_RUNBOOK.sh`: `ed36bc7a389f44c55b216d0adc4f4ee3a5ef57c68da0e4326ea123297d31dc59`

The TASK/backlog hashes match QA's TASK-first binding. PLAN, QA, and the stage
program embed the exact current runbook digests. TASK and backlog retain the
same declared zero-product evidence paths. `bash -n` passes for both scripts;
Shellcheck reports only the known SC2317 informational analysis of the stage
EXIT-trap callback. `git diff --check` passes.

### Prior findings and current disposition

1. **Q20-02 cycle — RESOLVED.** Documentary Q20-02 phase A plus this review
   may authorize only Q20-02L. Q20-02L's sanitized result and complete inner
   and outer rollback must then receive independent documentary Q20-02 phase B
   review before Main may separately authorize `canary`. The script's bounded
   `Q20-02` output is explicitly mapped to Q20-02L; it does not authorize any
   seed, broker, OTP, lease, or canary action.

2. **Exact root program and partial-mount safety — RESOLVED for Q20-02L.** The
   packet now states honestly that existing noninteractive sudo authority is
   broad and unchanged. Safety rests on Main approving the exact staged,
   root-owned, digest-matching `preflight` vector, not on sudo or namespaces
   being a technical capability boundary.

   The prior destructive-cleanup window is closed. Each recursive or file
   bind is added to `MOUNTS` immediately after successful creation. Every
   target in a recursive bind tree is remounted read-only deepest-first and
   every resulting options record must contain `ro`. Cleanup attempts
   recursive unmount in reverse registration order, then independently scans
   the mount table for the rootfs or any descendant. If one remains, cleanup
   records failure and performs no deletion. Only an empty mount-table result
   permits `find -xdev` within the disposable rootfs. Thus a partial remount or
   unmount failure cannot recurse into host `/usr` or another bind source.

   The outer stage cleanup is also appropriately fail closed: it requires the
   literal root-owned stage, an ordinary unmounted empty `rootfs`, the exact
   seven fixed members, and no process root/cwd/fd below the stage before its
   one-filesystem deletion. It then requires exact equality with the captured
   one-level `/var/tmp` state. Setup rollback has no mount operation and uses
   the same literal one-filesystem boundary.

3. **Peer/audit/expiry oracles — RESOLVED.** The contract retains separate
   one-event backend admission, zero-event wrong-peer pre-admission,
   zero-event stopped transport, and one-event fresh-broker denial classes.
   The distinct peer has only the owner socket supplementary group and
   therefore crosses pathname DAC while preserving a different peer UID. The
   claimed-owner raw probe now requires `recv(1) == EOF`; any response byte is
   failure, and the audit count must remain unchanged. Audit parsing requires
   the exact five fields and tuple. Activation captures its exact expiry;
   every later authorize allow must equal that value, while ready and every
   denial require null. Count deltas bind each admitted operation separately.

4. **TOTP boot floor — RESOLVED.** Each broker generation records a ceiling
   only after socket readiness and every successful activation waits until the
   current real counter is strictly greater. The bounded 31-second transition
   applies again after restart and remains distinct from the natural
   300-second lease-expiry wait. The generated OTP travels by direct pipe from
   the root-only tmpfs secret and enters neither argv nor environment.

5. **Evidence scope — RESOLVED.** `PLAN_REVIEW.md`, both exact runbooks,
   `CANARY_RESULT.md`, and `EVIDENCE_REVIEW.md` are declared identically in
   TASK and backlog. Product/test/workflow/dependency changes, product DEV,
   product REVIEW/QA results, counted Lap, child Git, and GitHub push remain
   excluded.

6. **Q20-10 evidence boundary — RESOLVED by an honest proportional oracle.**
   PLAN and QA no longer claim that a literal regex can discover arbitrary
   unknown secrets. The reviewable proof instead follows the actual flows:
   seed material is confined to root-only mode-0600 tmpfs files and process
   memory; OTP is generated from that file and sent only through a direct
   pipe; CLI/PAM streams are discarded; broker output is the protected strict
   five-field audit stream; successful retained runbook output is restricted
   to the fixed `case/result/count/digest` grammar; and final Markdown receives
   independent inspection rejecting raw JSON, command output, and forbidden
   values. The literal audit scan is correctly defense in depth. A command
   failure or unexpected output cannot be converted into PASS by later
   redaction; cleanup still runs and the evidence review must fail or classify
   the run from secret-free bounded facts.

### Scoped authorization and remaining gates

This PASS authorizes Main to perform only the exact Q20-02L route already
specified in PLAN: current Q20-01/documentary Q20-02 phase A prerequisites;
the digest-bound outer `setup`; the single staged `preflight` unshare argument
vector; unconditional inner Q20-11 cleanup; and the exact outer `cleanup` and
parent comparison. The pre-existing sudo policy must be recorded as broad and
unchanged, both source/staged digests and invoked argv must match, and any
setup, mount, PAM/sudo, cleanup, output, or equality discrepancy is a blocking
non-PASS.

This review does **not** authorize the `canary` argument vector. Main may
consider that separately only after Q20-02L exits successfully, its bounded
record explicitly maps the emitted `Q20-02` line to Q20-02L, inner and outer
rollback both pass, and an independent Reviewer completes documentary Q20-02
phase B. Q20-01 must also be current and passing. Any file or digest change
invalidates this PASS and requires another independent plan review.

## Fourth re-review — 2026-07-20 Q20-02L retry candidate

### Decision: PASS — authorizes one Q20-02L retry only

This review preserves all earlier FAIL and PASS history. It reviews the
post-attempt-one fixture correction and does not carry forward the superseded
runbook authorization. The current exact bindings are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `b1a2c044c2960f0edd333e93c5efe4ab95de8bb4119f43bb277c1eb2902b53f8`
- `QA_PLAN.md`: `6c7097ce2a40d14e195b496221f6ada7a21f58b6320d6d23530e8050e79c1409`
- `STAGE_RUNBOOK.sh`: `745e0d1852369bf3da834df1d3bfd12c170078aef7f09475c4d0a992e19d78df`
- `CANARY_RUNBOOK.sh`: `3d9519eb3eff3af1330d411d4741be034189ae46e02db725af142eb5344b91ad`
- `CANARY_RESULT.md`: `ef2aee01f4d734c0708d2208a78c08660caf706193156abaf6079b8bc8c6288f`

TASK/backlog retain the QA-bound hashes and identical evidence path set.
PLAN and QA name the current STAGE/CANARY digests, and the stage program pins
the same current CANARY digest. Both scripts pass `bash -n`; Shellcheck again
reports only the known SC2317 informational findings for the stage EXIT-trap
callback. `git diff --check` passes.

### Attempt-one classification and rollback

`CANARY_RESULT.md` records that Main invoked only the then-authorized
Q20-02L preflight and kept `canary` unauthorized. Q20-01 reached its bounded
PASS, then executable staging stopped because the new tmpfs `/usr/local` did
not contain its required `bin` directory. That failure occurs at the first
binary install, before the identity-creation block and before PAM, seed,
broker, OTP, or lease operations. It is therefore an evidence-harness
`environment_issue`, not an artifact implementation defect, product
regression, or permission widening.

The failed run's unconditional trap reported inner Q20-11 PASS. The fixed
outer cleanup then reported PASS using the exact pre-setup `/var/tmp`
comparison. The retained result states that no process, mount, staging tree,
identity, policy, seed, socket, timestamp, log, or binary remained. Current
read-only probes independently find the literal stage absent, no mount below
it, and no matching fixture process. Attempt one remains a failed feasibility
run rather than a Q20-02L PASS and cannot satisfy documentary phase B or
authorize the canary, but its exact inner/outer rollback permits a corrected
retry.

### Correction review

The functional fixture correction is one line immediately after mounting
tmpfs at `$ROOT/usr/local`:

```sh
mkdir -p "$ROOT/usr/local/bin"
```

The directory is consequently created inside the already registered private
tmpfs, before any archive installation. It neither creates a host path nor
changes archive identity, extraction, identities, PAM/sudo policy, TOTP,
audit, namespace scope, read-only host bindings, mount registration, unmount
order, mount-absence deletion gate, or host comparator. It directly addresses
the observed missing installation parent and introduces no source/local-build
substitution. The stage script's operational logic is unchanged apart from
pinning this new inner digest. All earlier partial-mount, recursive read-only,
no-delete-on-mounted-rootfs, peer-EOF, immutable-expiry, boot-floor, and
structural nonlogging controls remain present in the newly digested script.

### Scoped retry authorization

This PASS authorizes exactly one retry of Q20-02L using the current digests:
the fixed outer `setup`, the single PLAN-listed staged `preflight` unshare
vector, unconditional inner cleanup, and fixed outer `cleanup` with exact
parent comparison. Main must recheck source/staged equality, ownership/modes,
the current argv and digests, and the still-broad unchanged pre-existing sudo
policy before invocation. Any failure must again stop, run both cleanup
layers, and remain a non-PASS.

The `canary` vector remains unauthorized. Only a successful retry with the
bounded Q20-02L mapping, Q20-11 PASS, outer cleanup/equality PASS, and an
independent documentary Q20-02 phase B review may permit Main to consider a
separate canary authorization. Any further file or digest change invalidates
this retry PASS and requires another independent review.

## Fifth re-review — 2026-07-20 diagnostic Q20-02L candidate

### Decision: PASS — authorizes one diagnostic Q20-02L retry only

This review preserves every prior finding and verdict. It does not carry
forward the fourth review's now-superseded digest authorization. The current
exact bindings are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `de6dfa25548b36e558641f180f0712120e32be45725e48c9e8ede1c2f77d423b`
- `QA_PLAN.md`: `4c663b8e09b4e2faeb35f9eff85a0272a69e1a11bf2bec8c3d74cf514a295843`
- `STAGE_RUNBOOK.sh`: `13e249d51dcbe5086e796da6528ef40ac07fb00ff7106bb6545a9c45e8ef066c`
- `CANARY_RUNBOOK.sh`: `32b946eace48b8016c4b028684e0829d152043f3aa010b8e086d1cfa4297e617`
- `CANARY_RESULT.md`: `cbff743c07ef052c9ce3ad298cf67446edc0c8916b07ab2c99c481a806a6f294`

TASK/backlog retain their QA-bound hashes and the same zero-product evidence
scope. PLAN and QA bind the current STAGE/CANARY digests, and STAGE pins the
same current CANARY digest. Both scripts pass `bash -n`. Shellcheck reports
only informational SC2317 findings for the stage EXIT-trap callback and
SC2016 for the deliberately single-quoted shell executed inside the chroot;
the latter correctly defers `id` and `awk` expansion to the fixture process.
`git diff --check` passes.

### Attempt-two evidence and classification

`CANARY_RESULT.md` preserves attempt one and records that attempt two used
only the then-authorized Q20-02L vector. Q20-01 and private
mount/extraction Q20-03 passed. The real setuid sudo/PAM oracle then failed
before seed, broker, OTP, lease, or any artifact product behavior was reached.
This remains a fixture feasibility `environment_issue`; it is not evidence of
an implementation defect, regression, product change, or sudo-policy
widening.

The attempt-two trap reported inner Q20-11 PASS, and the fixed outer cleanup
reported PASS with the same exact pre/post parent digest used for attempt one.
The result records no residue. Current read-only probes independently find the
literal stage absent, no mount below it, and no matching fixture process.
Attempt two is not Q20-02L PASS and supplies no canary authorization, but its
exact inner and outer rollback make one bounded diagnostic retry safe.

### Diagnostic change review

The preflight now establishes four secret-free facts before interpreting the
sudo failure:

1. `/usr/bin/sudo` in the chroot is exactly mode 4755 and owned by UID/GID 0.
2. The bound `/usr` mount does not carry `nosuid`.
3. The pre-sudo fixture process has real UID 42020.
4. That process observes `NoNewPrivs: 0`.

The first two are read-only metadata/mount checks; the latter two execute only
the fixture `id`, `/proc` read, and shell after the existing irreversible
`setpriv` transition. A failed invariant remains Q20-02 failure and triggers
the unchanged safe cleanup. None creates authority, alters the host, weakens
PAM/sudo, or reaches a product binary decision.

Only the real sudo/PAM probe's stderr is redirected to
`/evidence/preflight.stderr`. The evidence tmpfs is root-only and mounted mode
0700; the established umask creates the file mode 0600. The raw text is never
printed: quiet case-insensitive searches reduce it to exactly one numeric
category—1 conversation/password, 2 setuid/`NoNewPrivs`/`nosuid`, 3
sudoers/policy, 4 PAM/account, or 9 unknown—and the only digest is derived
from the corresponding constant category label. Sequential matches choose a
single bounded final category and disclose no matched content. The raw file
is destroyed with the registered tmpfs during the same mount-absence-gated
rollback. This diagnostic can classify a failure only; it cannot turn the
oracle into PASS.

All earlier controls remain intact: literal staged invocation, broad existing
sudo policy recorded but unchanged, recursive read-only binds, immediate
mount registration, reverse recursive unmount, no deletion while any rootfs
mount remains, `find -xdev`, exact inner/outer comparators, artifact-only
extraction, no seed/TOTP in preflight, peer EOF, expiry equality, boot-floor
wait, and bounded successful output.

### Scoped diagnostic authorization

This PASS authorizes exactly one execution of the current Q20-02L diagnostic
route: the current fixed STAGE `setup`, the single PLAN-listed staged
`preflight` unshare vector, unconditional inner cleanup, and fixed outer
`cleanup` plus parent equality. Main must bind the exact current digests,
source/staged equality, owners/modes, argv, and unchanged pre-existing sudo
policy before execution. Whether it passes or produces one bounded diagnostic
category, both rollback layers remain mandatory and any discrepancy is a
blocking non-PASS.

The `canary` vector is not authorized. A diagnostic category may inform a
later fixture correction but confers no product classification and no further
retry authority. A successful Q20-02L still requires independent documentary
phase B before Main may separately consider canary execution. Any file or
digest change invalidates this PASS and requires another independent review.

## Sixth re-review — 2026-07-20 prompt-free Q20-02L candidate

### Decision: PASS — authorizes one Q20-02L retry only

This review preserves the complete prior history and supersedes only the fifth
review's old digest authorization. The current exact bindings are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `cf381c8b9645ae0f3d573da8562314aa03decae2a6d0b215392766fe2a846215`
- `QA_PLAN.md`: `6b25e62892d2068ff4f3108bbe61cc44e4b64b32fa9ac2f90ba1da89016b4df3`
- `STAGE_RUNBOOK.sh`: `c0a3e04dad22a45befd245e8545030e7af0d7596a58b1a0113e23c6d86d0d6be`
- `CANARY_RUNBOOK.sh`: `2e77e37502bd956daf023f725bc787caba190e55134bea7f4da6778cf4867711`
- `CANARY_RESULT.md`: `ed422a4e072a50865f78a6afd3cdac16783241cdf7a906306918061adc243772`

The TASK/backlog QA binding and evidence paths remain unchanged. PLAN, QA,
and STAGE consistently pin the new operational digest. Both scripts pass
`bash -n`; Shellcheck has only the previously reviewed informational SC2317
and intentional inner-shell SC2016 reports. `git diff --check` passes.

### Attempt-three classification and rollback

The independently authorized diagnostic attempt again ran only Q20-02L.
Q20-01 and Q20-03 passed, then the reducer returned bounded category 1
(password/conversation) at the preflight sudo oracle without emitting raw
stderr. The failure preceded seed, broker, OTP, lease, archived PAM-helper
execution, and product behavior. It therefore demonstrates a fixture route
problem—sudo's `-n` shortcut rejecting an authentication-required policy
before the intentionally prompt-free PAM stack—not an artifact defect,
regression, or product classification.

Inner Q20-11 and fixed outer cleanup both reported PASS with exact parent
equality and no residue. Current read-only probes find the literal stage
absent, no mount beneath it, and no matching fixture process. Attempt three
does not become Q20-02L PASS, but its safe rollback permits one corrected
retry.

### Noninteractive-boundary review

The current candidate removes `-n` only from the three forms of sudo executed
inside the private chroot: the harmless preflight, ordinary canary
`sudo_expect`, and stopped-broker denial. Every such call receives stdin from
`/dev/null`; stdout/stderr is either discarded or, for the diagnostic
preflight failure only, captured mode 0600 in root-only tmpfs and reduced to a
bounded category. The private `/dev` contains no tty or pty, so there is no
terminal input path.

The preflight PAM stack contains only `pam_permit` for auth/account and cannot
prompt. The archived live stack contains `pam_exec.so quiet seteuid` invoking
the fixed payload-free helper plus `pam_permit` for account; neither module
requests a password or interactive conversation. The helper accepts no
authority input and has its existing bounded broker timeout. Thus omitting
inner `-n` permits sudo to traverse real PAM but adds no password, terminal,
conversation, credential, or unbounded user-input route. If platform behavior
nevertheless requests input, EOF makes the call fail, its bounded diagnostic
or generic denial remains non-PASS, and cleanup is unconditional.

All host privilege entry remains explicitly noninteractive: STAGE setup and
cleanup and the exact PLAN-listed `unshare` launcher continue to be invoked
with outer `sudo -n --`. The host's pre-existing broad sudo policy is still
recorded honestly and unchanged. Removing an option from a command already
inside the private root-owned fixture neither changes that policy nor expands
the authorized host argument vectors.

The modification leaves artifact binding, mount/read-only/cleanup safety,
fixture identities, command grant, no-cache policy, audit reconciliation,
TOTP floor, natural expiry, restart, and secret-flow controls unchanged.

### Scoped authorization

This PASS authorizes one execution only of the current Q20-02L route: fixed
STAGE `setup`, the exact staged `preflight` unshare vector under outer
`sudo -n --`, unconditional inner cleanup, and fixed outer `cleanup` with
exact parent equality. Main must re-bind all current digests, installed/source
equality, owners/modes, argv, and unchanged host sudo policy before executing.
Any prompt, hang, unexpected output, diagnostic failure, cleanup discrepancy,
or equality mismatch is a blocking non-PASS and grants no retry authority.

The `canary` vector remains unauthorized. Only a successful Q20-02L record,
inner and outer rollback PASS, and independent documentary Q20-02 phase B may
allow Main to consider a separately authorized canary. Any file or digest
change invalidates this PASS and requires another independent review.

## Seventh re-review — 2026-07-20 canary authorization candidate

### Decision: PASS — authorizes one current `canary` execution only

I independently re-read the current task packet, both runbooks,
`CANARY_RESULT.md`, and `EVIDENCE_REVIEW.md`. The evidence now closes the
precondition left by the sixth review: attempt four ran the then-authorized
Q20-02L vector to completion, recorded bounded Q20-01 and Q20-03 PASS, mapped
the run to Q20-02L PASS, completed inner Q20-11, and completed the exact outer
parent-equality cleanup. The independent evidence review corroborates Q20-01
and Q20-02 phase B, including archive/run/attestation equality and the absence
of stage, mount, process, socket, identity, PAM, sudoers, binary, and service
residue. Attempts one through three remain documented fixture-environment
failures with successful inner and outer rollback; none reached canary
activation.

The exact current binding reviewed here is:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `PLAN.md`: `e076dc3f4e00b2fca27ed7a0e33e58519f62f5d964303c3d92f0f1a279534f42`
- `QA_PLAN.md`: `0745952ab34f30b4b8edf155d7525cdff7ef3b6ee4fe4cb1e65e5cf5e38f0500`
- `STAGE_RUNBOOK.sh`: `61de9f76f47fd4ad4a3a1b966506ccdd3530dfc7c3b716b1e422c4d47b011478`
- `CANARY_RUNBOOK.sh`: `544915baed67292fc082f15429e6de2276e17f7c0c55d36ac9f3431f0f66af3e`
- `CANARY_RESULT.md`: `95a1fd920fd13d0b0dd35fe0e0c3f3b581418d3cb538c71c88e4e70de89d5cb4`
- `EVIDENCE_REVIEW.md`: `d0acf07380afc355e32e72cc68b8c39928322fa3d20bc6b13486c4e59134d43c`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`

### Q20-02L semantic carry-forward and stream review

I reconstructed the sixth-review/preflight-PASS runbook from the current
runbook by reversing only the two post-mode-boundary stderr redirections. Its
SHA-256 is exactly
`2e77e37502bd956daf023f725bc787caba190e55134bea7f4da6778cf4867711`,
the digest bound to the successful attempt-four preflight. The entire
old-to-current executable delta is therefore confined to the ready probe and
real-OTP pipeline after `[[ $MODE == canary ]] || exit 0`; it was unreachable
in `preflight`. The STAGE change binds only the new inner runbook digest.
Accordingly the successful Q20-02L semantics and its independently reviewed
evidence carry forward without rerunning the privileged preflight.

The current ready probe discards both streams. The OTP generator now discards
stderr, and the OTP CLI discards both streams while receiving the code only
through the pipe. These changes close the last identified paths by which a
bounded failure diagnostic or generator diagnostic could escape to the outer
operator stream. They do not weaken an oracle: `pipefail`, command status,
and the exact-one-event reconciliation still make generation, transport, or
authorization failure fail closed. No seed, encoded secret, OTP, lease token,
raw audit line, or broker diagnostic is selected into emitted evidence.

### Canary safety and oracle review

The current canary route retains the previously reviewed containment and
rollback controls: exact mode/argument/root/stage gates; root, NNP, mount/PID
namespace, PID 1, and private-propagation checks; immediate reverse-order mount
registration; read-only recursive host binds and proc; tmpfs mutable trees;
one material archive extraction; exact host snapshots; holder checks; and
one-filesystem cleanup followed by exact parent equality. Cleanup remains
unconditional on every inner exit, with fixed outer cleanup required as a
second boundary.

The authorization oracles are sufficiently closed and discriminating for the
single canary run. The fixture uses distinct owner/actor identities, a narrow
`/usr/bin/true` sudo grant, `timestamp_timeout=0`, no tty, and stdin from
`/dev/null`; outer privilege entry remains `sudo -n --`. Wrong-root,
distinct-actor, and claimed-UID probes must deny with zero audit events. Two
authorized calls must each produce exactly one matching allow event and no
helper route. The run then proves the audited immutable expiry, natural-expiry
denial, reactivation, broker-stop denial, fresh-broker denial before the TOTP
floor, and a final real-TOTP activation and allow. Strict typed five-field
audit parsing, exact actor/scope/result matching, the shared captured lease
expiry, exact event counts, boot-counter floor, bounded waits, and literal
defense scans prevent a generic success, cached sudo state, fabricated expiry,
or uncontrolled output from satisfying PASS. Seeds, OTP material, lease state,
and raw broker audit remain root-only in tmpfs and are destroyed by rollback.

### Scoped canary authorization

This PASS authorizes exactly one execution of the current canary vector, plus
the fixed STAGE `setup`, unconditional inner cleanup, and fixed outer
`cleanup`. Main must first re-bind every digest above, source/staged equality,
owners/modes, the exact PLAN-listed canary argv, the current Q20-01 artifact
and attestation, and unchanged host sudo policy. It may then invoke only that
vector under outer `sudo -n --`. Any prompt, hang, unexpected output,
diagnostic failure, oracle mismatch, cleanup discrepancy, or equality mismatch
is a blocking non-PASS and grants no retry authority. No source substitution,
local build, interactive recovery, service enablement, product fix, or second
canary execution is authorized.

After the run, Q20-11 and outer rollback must pass and an independent evidence
review must decide the remaining canary evidence gates. This authorization is
not a final TASK-0020 PASS. Any bound-file or digest change invalidates it and
requires another independent review.

## Eighth re-review — 2026-07-20 Q20-05 transport retry candidate

### Decision: PASS — authorizes one current `canary` retry only

I independently re-read the current packet, both runbooks, retained result,
and prior documentary evidence. Canary attempt one used the exact vector
authorized by the seventh review. It passed Q20-01, Q20-03, carried-forward
Q20-02L, and Q20-04, then stopped at the composite Q20-05 gate before any
sudo authorization, natural-expiry, or broker-lifecycle case. Inner Q20-11
and fixed outer cleanup both PASSed with no residue. This is a fixture
`environment_issue`, not evidence of a product or artifact defect. The
existing `EVIDENCE_REVIEW.md` remains the independent Q20-01/Q20-02 phase-B
record; it does not purport to turn the failed canary into a Q20-05 PASS.

The exact current binding reviewed here is:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `PLAN.md`: `ab4b48461ba398c0608d3e635e22a00d680b7747ef51d2f6d8475f6f43db41d6`
- `QA_PLAN.md`: `d78c63e613f8135a22b8e539b65b32824156523c80a7a527ff323213191e7a88`
- `STAGE_RUNBOOK.sh`: `79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`
- `CANARY_RUNBOOK.sh`: `4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`
- `CANARY_RESULT.md`: `90d5d9c61f7b9c179ce201136354cf31cd2d81ed8b1a56a5f7539b69333858bc`
- `EVIDENCE_REVIEW.md`: `d0acf07380afc355e32e72cc68b8c39928322fa3d20bc6b13486c4e59134d43c`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`

### Failure classification and exact semantic change

The prior composite label cannot distinguish whether Q20-05 stopped on
broker start, a root/distinct peer, the claimed-identity transport, an audit
delta, or activation. The retained record therefore correctly makes no
product-behavior claim. The candidate changes no artifact and supplies no
product workaround. It makes the fixture's pre-admission transport oracle
match the server-side race: after connection, the wrong peer may be rejected
by EOF, connection reset, or broken pipe before or during the client's send.

The raw claimed-identity probe accepts exactly those three transport
rejection forms. A response byte makes `rejected` false; an unlisted exception
or nonzero probe status reaches the failure label; and the separate exact
before/after audit comparison must still show zero backend events. Root and
DAC-capable distinct-peer calls retain their denial and zero-event checks.
Thus the candidate recognizes equivalent closed-connection transports but
does not accept a response, admitted request, wrong-actor authority, or
ambiguous audit outcome. The new broker/root/distinct/claimed/audit/activation
labels are bounded constants which improve failure localization without
changing any command, success condition, secret flow, or emitted raw data.

All executable changes are inside the `canary` branch after
`[[ $MODE == canary ]] || exit 0`, and specifically after Q20-01, Q20-03,
Q20-02/Q20-02L, and Q20-04 have completed. Q20-02L attempt four and its
independent documentary phase-B PASS therefore remain semantically valid.
Canary attempt one's current-artifact Q20-01, isolation Q20-03, and genuine
archive declaration/binary Q20-04 results also carry forward as prerequisite
evidence; the retry will execute and must PASS those gates again. Replacing
the current inner digest with the seventh-review digest in the current STAGE
runbook reproduces the previously reviewed STAGE digest
`61de9f76f47fd4ad4a3a1b966506ccdd3530dfc7c3b716b1e422c4d47b011478`,
confirming its only operational change is the inner binding.

### Safety and retry scope

The candidate preserves the seventh review's containment, secrecy, and
functional oracles: exact entry and ownership checks; private mount/PID
namespaces; read-only host bindings and tmpfs mutation; genuine artifact-only
installation; root-only seed/audit files; direct-pipe TOTP with discarded
streams; strict typed five-field audit parsing; exact event counts and shared
expiry; prompt-free no-cache sudo; natural expiry; stopped/fresh-broker
separation; bounded waits; unconditional reverse cleanup; exact host snapshot;
holder checks; and fixed outer parent equality. The live stage path is absent,
no stage mount was observed, both scripts pass `bash -n`, and the packet passes
`git diff --check` at review time.

This PASS authorizes at most one execution of the current `canary` vector,
plus the fixed STAGE `setup`, unconditional inner cleanup, and fixed outer
`cleanup`. Main must first re-bind every digest above, source/staged equality,
owners/modes, the exact PLAN-listed argv, current Q20-01 artifact and
attestation, unchanged host sudo policy, and absence of fixture residue. The
invocation remains outer `sudo -n --`; no alternative vector or interactive
recovery is permitted.

Any failure—including another Q20-05 label, prompt, hang, unexpected output,
response byte, audit delta, later oracle mismatch, cleanup discrepancy, or
parent-equality mismatch—is a blocking non-PASS and grants no further retry
authority. No product edit, local build, artifact substitution, weakened
oracle, service enablement, or persistent installation is authorized. Q20-11
and outer cleanup are mandatory on every outcome, and a new independent
evidence review must decide the completed canary record. This is not final
TASK-0020 PASS; any bound-file or digest change invalidates this authorization.
