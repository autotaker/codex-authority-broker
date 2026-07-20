# CANARY_RESULT — TASK-0020

## Main approval and fixed input

Main approved only Q20-02L against the independently reviewed planning packet;
the `canary` vector remained unauthorized. Q20-01 reverified successful run
29720021660 attempt 1, main source
`09487b104f32cad23a695ec3f1a0c7e7a68e6163`, archive SHA-256
`5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd`,
seven exact members, six streaming payload checksums, and the exact GitHub
attestation repository/workflow/ref/source binding. No payload was extracted
on the host.

## Q20-02L attempt 1 — environment fixture failure, safe rollback

The approved preflight emitted bounded Q20-01 PASS, then stopped before PAM,
seed, broker, OTP, or lease work because the tmpfs `/usr/local` did not contain
the required `bin` directory. This is an `environment_issue` in the evidence
harness, not a product or artifact defect. Inner Q20-11 PASS proved exact
namespace cleanup. The fixed outer cleanup PASS matched its pre-setup
`/var/tmp` digest exactly. No process, mount, staging tree, fixture identity,
policy, seed, socket, timestamp, log, or binary remained.

The runbook was corrected only by creating tmpfs `/usr/local/bin` after the
`/usr/local` mount. All prior authorization was invalidated by the digest
change. No second preflight and no canary may run until QA reconciliation and a
fresh independent plan-review PASS bind the new digests.

## Q20-02L attempt 2 — setuid/PAM oracle failure, safe rollback

After fresh QA reconciliation and independent review, Q20-01 and private
mount/extraction Q20-03 passed. The bounded Q20-02 real setuid sudo/PAM oracle
failed before seed, broker, OTP, or lease work. Inner Q20-11 and fixed outer
cleanup both passed, with the same pre/post outer digest as attempt 1 and no
residue. This remains an `environment_issue`/fixture-classification question;
there is no product or artifact failure evidence.

The next runbook candidate adds only secret-free preflight invariants and a
bounded failure-category reducer. Raw sudo/PAM stderr remains mode-0600 in
tmpfs, is never emitted, and is destroyed during rollback. The changed digest
again invalidates all prior authorization; no diagnostic retry or canary is
authorized until QA reconciliation and fresh independent review.

## Q20-02L attempt 3 — category 1, safe rollback

The independently approved diagnostic retry again passed Q20-01 and Q20-03,
then reduced the Q20-02 failure to category 1
(password/conversation) without emitting raw stderr. Inner Q20-11 and outer
cleanup passed with exact parent equality and no residue. This proves that
fixture `sudo -n` rejects before reaching the intentionally prompt-free PAM
stack; it is an evidence-harness defect, not product behavior.

The candidate now removes `-n` only from sudo commands inside the isolated
fixture and supplies stdin from `/dev/null`. The outer administrator launcher
remains noninteractive `sudo -n`. Because both the preflight and archived PAM
stacks contain no prompting module, the revised route permits no password,
terminal, conversation, or interactive secret while allowing real PAM to run.
The digest change invalidates prior authorization pending QA reconciliation
and fresh independent review.

## Q20-02L attempt 4 — PASS with exact rollback

After QA reconciliation and a sixth independent plan review, the exact
digest-bound retry exited 0. Its bounded record contained Q20-01 PASS (six
streamed payload checks), Q20-03 PASS (private mount/extraction map), Q20-02
PASS (mapped to Q20-02L real setuid sudo/PAM preflight), and inner Q20-11 PASS.
No seed, broker, OTP, lease, or canary operation existed. The fixed outer
cleanup then PASSed with the same `/var/tmp` parent digest as setup and all
prior attempts.

Q20-02L therefore demonstrates the reviewed namespace, tmpfs, setuid, PAM,
fail-fast cleanup, and exact inner/outer rollback route. This does not by
itself authorize the `canary` vector. Documentary Q20-02 phase B must
independently bind this result to the reviewed digests and confirm no residue
before Main can separately approve the live canary.

## Pre-canary Main stream audit

Documentary phase B passed, but Main found before canary execution that the
successful `ready`/`otp` commands discarded stdout while their generic failure
stderr could still escape the protected evidence boundary. No canary had
started. The candidate now redirects both CLI streams, including TOTP-generator
stderr, to `/dev/null`; no authentication behavior, preflight path, artifact,
or oracle changed. The digest change invalidates authorization until QA and an
independent reviewer decide whether the already-PASSed Q20-02L can be carried
forward and authorize the new canary candidate.

## Canary attempt 1 — Q20-05 fixture stop, safe rollback

The separately authorized canary passed Q20-01, Q20-03, carried-forward
Q20-02L, and Q20-04 artifact binary/PAM/sudoers/systemd validation, then
stopped at the composite Q20-05 peer/broker/activation gate. No sudo allow,
lease-expiry, or broker lifecycle case ran. Inner Q20-11 and exact outer
cleanup both passed with no residue. The evidence is insufficient to classify
product behavior and remains a fixture `environment_issue`.

The claimed-identity transport probe previously required only EOF; the server
may also reject the wrong peer by connection reset or broken pipe before the
client's send completes. The candidate now treats only EOF/reset/broken-pipe
as transport rejection, still rejects any response byte, still requires zero
backend audit, and emits bounded subcase labels for future classification. No
product input or oracle is weakened. The digest change invalidates
authorization pending QA reconciliation and independent review.

## Canary attempt 2 — PASS pending independent evidence review

The eighth independent review authorized exactly one retry of the current
digest-bound `canary` vector. The run completed successfully and emitted only
bounded case/result/count/digest records:

| Case | Result | Bounded evidence |
| --- | --- | --- |
| Q20-01 | PASS | exact archive digest; six streamed payload checks |
| Q20-02/Q20-03 | PASS | carried real setuid/PAM preflight; private mount and single tmpfs extraction |
| Q20-04 | PASS | three checksum-matching binaries and genuine PAM/sudoers/systemd validation |
| Q20-05 | PASS | root, DAC-capable distinct UID, and claimed-owner probes rejected before backend; zero audit; real post-boot-floor TOTP activation |
| Q20-06 | PASS | two separate actual sudo/PAM/helper allows; one authorize event each; no helper/timestamp reuse |
| Q20-07 | PASS | natural 300-second expiry; fresh actual sudo denied with one fresh authorize event |
| Q20-08 | PASS | reactivation allow; stopped broker denial with zero event; fresh broker denial with one event; only a new post-boot-floor TOTP restored allow |
| Q20-09 | PASS | every backend-admitted operation had one exact five-field event; actor/scope/result/null and immutable allow-expiry relationships reconciled |
| Q20-10 | PASS | structural secret boundary and protected evidence scan passed; retained output remained bounded |
| Q20-11 | PASS | broker/process/socket/secret/policy/binary/mount teardown and exact inner host equality |

The fixed outer cleanup also PASSed and reproduced the same pre-setup
`/var/tmp` parent digest used for setup and every prior rollback. Independent
post-cleanup probes at `2026-07-20T08:49:55Z` found the staging path and fixed
socket absent and zero fixture executable processes. No host identity, PAM,
sudoers, service, seed, timestamp, log, mount, namespace holder, or installed
artifact remains.

No OTP, seed, request/response payload, token, key, credential, environment,
lease identifier, raw audit event, internal error, or unbounded command output
is retained here. Product/test/workflow/dependency paths were unchanged and
production SLOC remains 1478. Main does not mark TASK-0020 or v1 complete until
an independent Reviewer checks this final record against Q20-01–Q20-12.

### Ordered bounded attempt-2 record

The following is the complete retained runbook record in execution order. It
contains only case IDs, bounded counts, and SHA-256 digests; it is not raw
command output or raw audit JSON.

| Sequence | Case / relationship | Count | Digest |
| ---: | --- | ---: | --- |
| 1 | Q20-01 artifact payload checks | 6 | `5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd` |
| 2 | Q20-03 private mount map | 1 | `112a87ad912a12f2574dc99f55a8214f2fe239188651fb632fa6e947a97824e7` |
| 3 | Q20-02 carried Q20-02L | 1 | `c24bf849b707e5ecd8c3decde1b77631fc2ae216cbfee35f3438b1a8e61987c0` |
| 4 | Q20-04 installed artifact inputs | 3 | `3710b8b3f45653f60f1dd18fb3e2e024bed5f939fc2353bf6220bdff764ac077` |
| 5 | Q20-05 three wrong-peer zero-event controls | 3 | `1896de989ddaf06b75034f868fa7752496cf64a907e93594d032a180c6b8eec1` |
| 6 | Q20-09 first-generation ready allow | 1 | `24050752c6e31898787f1cd2ec480287208ae11bc596cfa4694df5569ab72cf3` |
| 7 | Q20-09 first-generation OTP allow | 1 | `4953c135e1c122146274f0266c27f90f158eb3ab9a2dec6bcb43d2784471c66c` |
| 8 | Q20-09 first sudo authorize allow | 1 | `a7c33d926774014d2a2e8b671a046509d77a6d1d5ffcb12923d52e08a612b185` |
| 9 | Q20-09 second sudo authorize allow | 1 | `429028d7daee4ae607db3daa8f500361a8339e1d1fef2425c100c928d31eccdf` |
| 10 | Q20-06 two actual no-cache sudo allows | 2 | `bef65cf109d11f188e766128b045b98983b035eff81ba1dc2070c97489a76c23` |
| 11 | Q20-09 natural-expiry authorize deny | 1 | `674a551a260e2157defe28ad47592ce987559b4c96025ea6a32184a545e5e8a7` |
| 12 | Q20-07 natural-expiry result | 1 | `6ce68a4ae8a1817fe04b11667f573b0eb63df6648f525fc86851bdac9ac52190` |
| 13 | Q20-09 second activation ready allow | 1 | `bcb7fa42909be002ece888b1a0f50da5c1c344a4621ab729e14edfea8e2aaf67` |
| 14 | Q20-09 second activation OTP allow | 1 | `d3d663f424b6b07e05218ecc61ed1ba1c8fe7072bfb915e7d129bcfee3379ebb` |
| 15 | Q20-09 post-reactivation authorize allow | 1 | `55654fc29e420adc96b128da95481276040c7d39512b4e350fc61815818b8573` |
| 16 | Q20-08 stopped-broker zero-event denial | 1 | `27e098e43a3ccf42616438158608f74255db9a170289fd3480b7566497e86685` |
| 17 | Q20-09 fresh-broker authorize deny | 1 | `9663dde46bb4b8ae66b91ae0ff8002ad41252121d155ac727a3741751419992e` |
| 18 | Q20-09 third activation ready allow | 1 | `2c906ee97468c74da85db8e63b902f5d04c46e7ef995dc4f36d9f2c5dc024623` |
| 19 | Q20-09 third activation OTP allow | 1 | `4bc44f1b1fa8a83a675a36268a3ebfbcc20ddd1ae8c1a1925d74c5715e7a125f` |
| 20 | Q20-09 post-restart authorize allow | 1 | `38c9d13de1c143aff182f9e225042ddc0df3044e70e5d8ba38dfecd768ff230e` |
| 21 | Q20-08 lifecycle aggregate | 4 | `3655446a690efca33d51ba7e52431ab408ba2e6ec6f0a96aee56405753ebf5ee` |
| 22 | Q20-10 protected evidence set | 1 | `4c5f6ac1151124bac5ea64a283441d551fa42e02eed48b6d0daa9576d7187d03` |
| 23 | Q20-11 inner rollback | 1 | `26e2142cceff5085b984a7bcb91dc24db36816d2f0acd8cd4f31a7cfb69cf1a9` |
| 24 | outer setup and cleanup parent equality | 1 each | `fe31a9ffed2a5200ea8a636f5a8a990ad170915e5ceaca037b2caaee06f1e6f6` |

The Q20-09 rows are hashes of protected exact five-field events, not the
events themselves. Their labels and order bind each digest to its corresponding
call. The stopped-broker row intentionally has no Q20-09 event; the fresh
broker denial intentionally has one. This complete record is submitted for a
fresh independent Q20-12 review; the earlier FAIL remains part of history.

## Owner-run manual E2E — 2026-07-21

The repository owner subsequently executed the published manual procedure on
Ubuntu. Preflight passed Q20-01, Q20-03, Q20-02, and Q20-11. The canary passed
Q20-01 through Q20-10, including the functional, denial, natural-expiry,
broker-lifecycle, audit, and protected-evidence checks. Its raw terminal record
was:

```text
Q20-11 result=FAIL count=0 digest=e86746208b5a305f8809d498edc80ceefaf12540897beae62828ab09868489c5
```

A bounded diagnostic rerun reproduced the same classification with raw digest
`095c29175d7a46e370f81f2d8343c399998641b6da3fd219c0b8b4140940385a`.
The retained host diff contained exactly one difference:

```diff
-run-root\tdirectory\t/run\t755\t0\t0\t940
+run-root\tdirectory\t/run\t755\t0\t0\t920
```

No fixture-owned path changed. The `/run` root directory inode size is volatile
because unrelated runtime services can add or remove entries during the
canary's natural-expiry wait; it is not evidence of fixture residue. The outer
cleanup passed with the same setup digest
`a8258a8c1b1d68fd299287288e2a3c047c32aaec39e3626b5d05bf58ebfa0bcd`,
and the staging path, socket, stage mounts, and fixture processes were all
absent (`post_cleanup_rc=0`).

The owner accepted the overall run as PASS with a `qa_plan_defect` exception,
retained the raw Q20-11 FAIL, and waived a formal rerun. The oracle-only script
correction removes the `/run` root inode-size comparison while retaining exact
comparisons of `/run/codex-authority.sock`, `/run/sudo`, and every other
declared host boundary. The corrected script was not represented as executed
evidence.
