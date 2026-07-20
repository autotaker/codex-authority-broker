# Q20-02 phase B interim review

## Decision: PASS

I did not operate the fixture. This independent documentary review is limited
to the current Q20-01 binding, Q20-02L attempt 4, its inner and outer rollback,
and the retained Q20-02L attempt history. It does not review or report a live
canary result.

The exact current reviewed bindings are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `cf381c8b9645ae0f3d573da8562314aa03decae2a6d0b215392766fe2a846215`
- `QA_PLAN.md`: `6b25e62892d2068ff4f3108bbe61cc44e4b64b32fa9ac2f90ba1da89016b4df3`
- `PLAN_REVIEW.md`: `f8278b0f7bd4006747b3e202d707ee02588212b8038bec32676aaffbee88d880`
- `STAGE_RUNBOOK.sh`: `c0a3e04dad22a45befd245e8545030e7af0d7596a58b1a0113e23c6d86d0d6be`
- `CANARY_RUNBOOK.sh`: `2e77e37502bd956daf023f725bc787caba190e55134bea7f4da6778cf4867711`
- `CANARY_RESULT.md`: `6c05af808cb15beea7e582c8fa28b6df40b136386b8116a73904b7a185571b38`

## Q20-01 current binding

The contract, QA plan, backlog, runbook, and retained result consistently bind
artifact `codex-authority-linux-amd64` to successful run `29720021660`, attempt
1, repository `autotaker/codex-authority-broker`, workflow
`.github/workflows/release.yml`, branch `main` / `refs/heads/main`, source
`09487b104f32cad23a695ec3f1a0c7e7a68e6163`, and archive SHA-256
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`.
The current GitHub run and artifact records corroborate that run, attempt,
successful conclusion, workflow, repository, branch, source, and exact
artifact name. Independent read-only inspection of the retained download
found exactly the two outer files, the exact seven allowed archive members,
six payload checksum entries, byte-identical uploaded/archived checksum
files, and six successful streamed payload checks. The retained Main record
also binds the successful GitHub attestation verification to the same
repository, workflow, ref, source, and subject digest, and states that no
payload was extracted on the host. Q20-01 is current and PASS for this phase-B
gate.

## Q20-02L attempt 4 and rollback

The sixth independent plan-review PASS authorized one execution of only the
current digest-pinned `preflight` vector. `CANARY_RESULT.md` records that the
exact retry exited 0 and retained bounded PASS results for Q20-01 (six streamed
payload checks), Q20-03 (the private mount/extraction map), Q20-02 explicitly
mapped to Q20-02L (real setuid sudo through the prompt-free `pam_permit`
stack), and inner Q20-11. It also records fixed outer cleanup PASS with the
same `/var/tmp` parent digest as setup and the prior attempts. No seed, broker,
OTP, lease, archived PAM-helper decision, or canary operation existed in this
preflight. The retained Markdown contains no raw JSON, raw command output,
secret, OTP, request/response payload, credential, environment, lease
identifier, or internal error text.

Independent read-only host probes after teardown found:

- `/var/tmp/codex-authority-task0020` absent, with zero mounts below it and
  zero process cwd/root/fd holders;
- zero processes whose executable is a fixture broker, CLI, or PAM helper;
- `/run/codex-authority.sock` absent;
- zero `codex-fixture` or `codex-distinct` passwd/group entries; and
- the named PAM, sudoers, installed binary, and service residue paths absent.

Attempt 4 therefore satisfies Q20-02L with exact recorded inner and outer
rollback and independently corroborated absence of residue.

## Retained history and authorization boundary

Attempts 1 through 3 remain non-PASS feasibility history: respectively the
missing tmpfs `/usr/local/bin`, the setuid/PAM oracle failure, and bounded
category 1 caused by fixture `sudo -n`. Each stopped before seed, broker, OTP,
lease, or artifact product behavior; each retained inner Q20-11 PASS, outer
cleanup/equality PASS, and no residue. Their successive fixture corrections
changed the operational digest and received fresh independent review before
the next attempt. They are correctly classified as evidence-harness
`environment_issue` history, not product or artifact defects, and none is
being substituted for attempt 4.

Documentary Q20-02 phase B is PASS. With Q20-01 current and Q20-02L attempt 4
plus both rollback layers passing, Main may separately authorize the exact
current digest-pinned `canary` vector described by the approved packet. This
is only a recommendation to authorize that next vector: no canary execution
or canary PASS is claimed here. Any change to the current planning or
operational digests requires fresh review before execution.

## Final Q20-12 evidence review

### Decision: FAIL — final completion evidence is incomplete

I did not operate either canary attempt. I re-read the complete current TASK,
backlog entry, PLAN, QA_PLAN, full PLAN_REVIEW history, both runbooks,
CANARY_RESULT, the interim review above, and the relevant broker IPC, audit,
lease/TOTP, PAM helper, deployment, and release-workflow source. This review
preserves the interim documentary Q20-02 phase-B PASS; it decides only the
final Q20-12 completion gate.

The exact current inputs to this review are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `ab4b48461ba398c0608d3e635e22a00d680b7747ef51d2f6d8475f6f43db41d6`
- `QA_PLAN.md`: `d78c63e613f8135a22b8e539b65b32824156523c80a7a527ff323213191e7a88`
- `PLAN_REVIEW.md`: `a032f67898d48bc3019822cf12065cc5a0e8b12d1f8c77a1ea1cfd32d21fb5f4`
- `STAGE_RUNBOOK.sh`: `79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`
- `CANARY_RUNBOOK.sh`: `4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`
- `CANARY_RESULT.md`: `bd555b41cbd439ce68487a450c1f091619069c6c9f0065e98d84063110cfaaa8`
- interim `EVIDENCE_REVIEW.md` before this append:
  `d0acf07380afc355e32e72cc68b8c39928322fa3d20bc6b13486c4e59134d43c`

### Evidence that does reconcile

Q20-01 remains bound to successful GitHub Actions run `29720021660`, attempt
1, repository `autotaker/codex-authority-broker`, workflow
`.github/workflows/release.yml`, `main` / `refs/heads/main`, source
`09487b104f32cad23a695ec3f1a0c7e7a68e6163`, artifact
`codex-authority-linux-amd64`, and archive SHA-256
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`.
The current GitHub run/artifact records corroborate those fields. Independent
stream-only inspection again found exactly two outer files, seven exact
regular archive members, six checksum entries, byte-identical outer/archived
checksum files, and six matching payload digests. The retained record states
that the GitHub attestation passed against the same repository, workflow,
ref, source, and subject digest.

The eighth independent plan review authorized exactly one retry using the
current STAGE and CANARY digests. It correctly classified canary attempt 1 as
a Q20-05 transport-fixture stop with inner and outer rollback PASS, then
reviewed the narrow EOF/reset/broken-pipe correction without accepting a
response byte or backend audit event. The prior four Q20-02L attempts and
canary attempt 1 remain intact as history; each failed attempt is separated
from the later authorization and has recorded safe rollback rather than being
substituted as a PASS.

The frozen runbook's logic supports the asserted functional matrix. Wrong
root, distinct, and claimed-owner peers are pre-admission zero-event cases.
Every admitted call is checked for exactly one strict five-field event. A
successful complete run necessarily checks 12 admitted events: three
`ready`, three `otp`, and six `authorize`, split as eight events in broker
generation one and four in generation two. The six authorize events comprise
three pre-expiry allows in the first generation, its natural-expiry denial,
the fresh-generation denial, and its final post-activation allow. Every allow
is compared with the activation's captured immutable expiry, every denial and
ready event requires null, both initial sudo allows are distinct calls with
`timestamp_timeout=0`, and stopped-broker failure requires zero added events.
The TOTP counter wait is strictly above each process boot floor. Source review
confirms that wrong peers are rejected before `Backend.Handle`, the lease is
process-local with an immutable 300-second deadline, and the PAM helper makes
one payload-free authorize request after dropping to the socket owner.

The secret-flow review also passes as a design and retained-Markdown check.
Seed and raw TOTP material are confined to root-only tmpfs files/process
memory, OTP travels through a direct pipe with both command streams discarded,
audit stays in protected tmpfs, and successful public output can only use the
fixed bounded emit grammar. `CANARY_RESULT.md` contains no raw JSON, audit
event, seed, OTP, request/response payload, token, credential, environment,
lease identifier, internal error, or unbounded command output.

The current canonical production recount is exactly 1478 nonblank,
non-comment lines over the nine non-test Go production files. TASK-0020 still
declares no production or test path and only the nine evidence/backlog paths;
its result records no product, test, workflow, dependency, product DEV, Lap,
REVIEW_RESULT, or QA_RESULT change. Both current runbooks pass `bash -n`.

Independent post-run host probes found the literal stage absent, zero mounts
at or below it, zero process cwd/root/fd holders, zero fixture executable
processes, the fixed socket absent, zero fixture passwd/group identities, and
zero named residue across PAM, sudoers, installed binaries, service locations,
logs, `/run`, `/var/tmp`, `/usr/local`, and the inspected `/etc` locations.
This corroborates the recorded inner Q20-11 and outer rollback outcome.

### Blocking Q20-12 evidence gap

`CANARY_RESULT.md` says attempt 2 emitted bounded
`case/result/count/digest` records, but it does not retain those records. Its
table contains narrative PASS summaries and a few aggregate counts, not the
case-bound sanitized execution evidence required by QA_PLAN. In particular,
the declared evidence set contains none of the following from canary attempt
2:

- the emitted per-operation Q20-09 digest records or an exact reconciliation
  of the 12 admitted events, their 3/3/6 scope counts, and their 8/4
  broker-generation split;
- the bounded digests emitted for Q20-03 through Q20-10, including the two
  audit-stream relationship digests and the Q20-10 retained-surface digest;
- the actual inner Q20-11 equality digest and outer setup/cleanup parent
  digest that would bind the asserted equality to this attempt;
- per-case UTC start/end, operator/fixture labels, bounded probe names, and
  exits required by the fixed evidence rules; or
- an exact retained output-grammar/count transcript from which an independent
  reviewer can distinguish execution of every oracle from Main's summary of
  those oracles.

This is not repaired by the runbook being capable of producing the evidence,
by source review showing what a successful exit would imply, or by the clean
host now corroborating teardown. QA_PLAN explicitly says an evidence-review
case cannot PASS from Main's assertion without its case-bound sanitized
record, and Q20-12 says any omission prevents `EVIDENCE_REVIEW.md` PASS. The
current table cannot independently establish that the claimed counts,
relationships, digests, and exact equality were the bounded outputs of canary
attempt 2.

Final Q20-12 is therefore FAIL for incomplete documentary evidence. This does
not reclassify the operational behavior as a product defect and does not
invalidate the clean rollback finding. TASK-0020 and v1 completion must not be
claimed. A new sanitized CANARY_RESULT revision would need to retain the
already-produced bounded records (without reconstructing or inventing them),
bind them to the current runbook/artifact and attempt, and then receive a fresh
independent Q20-12 review. If the original bounded records are unavailable,
the gap cannot be converted to PASS by narrative reconstruction.

## Fresh final Q20-12 re-review

### Decision: PASS

The preceding FAIL remains the accurate verdict on the earlier incomplete
record. Main subsequently appended the complete retained 24-row bounded
attempt-2 record to `CANARY_RESULT.md`; no operational program, planning
contract, product path, artifact, or prior history changed. I independently
re-reviewed that new evidence without operating the fixture.

The exact current bindings for this re-review are:

- `TASK.md`: `ed0356fbe8a059017cea0c18f28a91819fef1c5a28a930d2c1180793a3427d67`
- `backlog.json`: `3ae6b76e4f2a802b6ddc5730a1a36160d3000df4bce5c6dd96efc6347438e7fe`
- `PLAN.md`: `ab4b48461ba398c0608d3e635e22a00d680b7747ef51d2f6d8475f6f43db41d6`
- `QA_PLAN.md`: `d78c63e613f8135a22b8e539b65b32824156523c80a7a527ff323213191e7a88`
- `PLAN_REVIEW.md`: `a032f67898d48bc3019822cf12065cc5a0e8b12d1f8c77a1ea1cfd32d21fb5f4`
- `STAGE_RUNBOOK.sh`: `79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd`
- `CANARY_RUNBOOK.sh`: `4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7`
- `CANARY_RESULT.md`: `e88dbee8e2d831610f787a9110ed5595c9b53bbd0c32eef61156263c211c669c`
- `EVIDENCE_REVIEW.md` before this append:
  `5eb7318b9d906dda33c0d7833239c3fd84e185f5fdaa48658477885d5742f08f`

### Ordered record reconciliation

The appended record has exactly 24 consecutively numbered rows; every row has
the required bounded count and a well-formed SHA-256 digest. Its order exactly
matches the frozen runbook:

1. Q20-01 archive verification precedes the private Q20-03 mount/extraction
   map, the real Q20-02/Q20-02L setuid/PAM preflight, and Q20-04 artifact input
   validation.
2. Q20-05 then records all three wrong-peer controls as zero-backend-event
   rejections before the first activation.
3. The first broker generation contains eight admitted events in exact call
   order: ready allow, OTP allow, two sudo authorize allows, natural-expiry
   authorize denial, second ready allow, second OTP allow, and post-reactivation
   authorize allow. The Q20-06 aggregate follows its first two sudo events and
   the Q20-07 aggregate follows the natural-expiry denial.
4. The stopped-broker Q20-08 row follows the reactivation allow and correctly
   has no Q20-09 event. The fresh broker then has exactly one pre-activation
   authorize denial, followed by third ready/OTP allows and the final authorize
   allow. The Q20-08 lifecycle aggregate follows those four lifecycle cases.
5. Q20-10 protected-evidence PASS follows all functional/audit cases, then
   inner Q20-11 rollback and fixed outer setup/cleanup equality close the
   record.

There are exactly 12 Q20-09 rows: three ready, three OTP, and six authorize,
split eight/four across the two broker generations. The authorize sequence is
three first-generation pre-expiry allows, its natural-expiry denial, the fresh
generation's pre-activation denial, and its post-activation allow. This
reconciles the stopped transport as zero event and the fresh-broker denial as
one event rather than crossing those oracle classes.

The successful runbook path makes each Q20-09 row contingent on an exact-one
count delta and strict five-field parsing. Ready and denials require null
expiry; OTP and every authorize allow are bound to the activation's captured
expiry, with later allows compared for exact immutable equality. The two
Q20-06 sudo calls are separate real PAM/helper calls with
`timestamp_timeout=0`, and the stopped-broker check also requires no surviving
helper. The natural-expiry and fresh-process boot-floor ordering therefore
cannot be replaced by a cached permit, synthetic event, or prior lease.

Independent digest reconstruction matches the retained Q20-01 artifact,
Q20-02 fixed preflight relationship, Q20-04 three installed artifact inputs,
Q20-05 wrong-peer relationship, Q20-08 stopped-broker relationship, and
Q20-10 protected evidence filename-set rows. The run-specific mount map,
individual event, audit-stream aggregate, inner equality, and outer parent
digests are present in their exact execution positions. Row 23 supplies the
previously missing inner host-equality digest. Row 24 supplies one identical
outer digest for setup and cleanup, binding exact parent equality rather than
only asserting cleanup.

### Remaining completion gates

The retained 24-row transcript consists only of bounded labels, counts, and
digests. It contains no raw audit JSON, OTP, seed, request/response payload,
token, key, credential, environment, lease identifier, internal error, or
unbounded command output. Together with the already reviewed root-only tmpfs,
direct-pipe/discard, strict audit-schema, and cleanup flows, Q20-10 remains
PASS.

Current independent read-only probes again find the literal stage and socket
absent, zero stage mounts or process holders, zero fixture executable
processes, zero fixture passwd/group identities, and zero named PAM, sudoers,
binary, service, log, `/run`, `/var/tmp`, `/usr/local`, or inspected `/etc`
residue. These probes corroborate the inner Q20-11 and outer equality rows.

The canonical production recount remains exactly 1478 nonblank,
non-comment lines across the nine non-test Go production files. TASK-0020 has
no production or test path and retains only its declared evidence/backlog
scope; no product DEV, counted Lap, product REVIEW_RESULT, product QA_RESULT,
workflow, dependency, or generated-product change is claimed. Both current
runbooks remain syntactically valid.

The newly retained ordered evidence closes every documentary omission that
caused the earlier FAIL without reconstructing a secret or raw event. Q20-01,
Q20-02/Q20-02L, and Q20-03 through Q20-11 are now present, current-bound,
secret-free, internally consistent, and independently reviewable. Final
Q20-12 is PASS for TASK-0020's evidence and exact rollback contract. This
verdict does not erase the earlier FAIL or the safely rolled-back attempt
history.
