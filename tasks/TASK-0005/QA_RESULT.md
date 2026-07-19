# QA RESULT — TASK-0005

## Decision: PASS

Independent QA began after REVIEW PASS and independently passed Q5-01 through
Q5-07. This evidence authorizes only the Main-owned Git gate.

## Independent evidence

- Fixed `ready` and `otp` actions each dispatched once through a real Unix
  socket. Unsupported and secret-bearing argument/environment paths denied
  locally with zero backend calls.
- OTP input was stdin-only. Capture results were clean for argv, environment,
  stdout, stderr, returned errors, and logs; the non-vacuous negative control
  detected its injected hit. No synthetic secret value is recorded here.
- Real kernel peer UID matching succeeded; deliberately mismatched
  `AllowedUID` denied with backend count zero.
- Default `0600` and numeric-owner/group provisioned `0660` socket modes
  passed. Provisioning failure removed only the owned socket. The host has no
  `codex` account and production performs no account-name lookup.
- Focused/full/race/vet/gofmt/diff checks passed with no P0 skips. Scope,
  retained prior controls, idiomatic structure, and no-compression checks pass.

Exact cumulative production SLOC is 751/820 (83 CLI, 35 client, 117 protocol,
283 server, 173 lease, 60 TOTP), below the Revision 2 stop above 790. Tests are
1,257 physical lines and 1,169 nonblank/non-line-comment lines.

QA timing was `active_ms=20168`, `wait_ms=0`. One retry followed the classified
sandbox AF_UNIX `EPERM` `environment_issue`; the approved socket-capable retry
passed. Candidate/flaky retries were zero. No product, plan, test, Git state,
or Lap30 log was changed by QA.
