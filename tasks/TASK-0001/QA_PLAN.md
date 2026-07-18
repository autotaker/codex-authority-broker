# QA PLAN — TASK-0001

## QA decision and evidence rules

QA accepts this task only when every **P0 acceptance** row below passes on the
released Debian artifact and its source/test evidence.  A skipped P0 row is a
FAIL, not a conditional pass.  Manual host activation and GitHub App
provisioning are rollout evidence; they are not unit-test side effects and do
not weaken the package or fixture requirements.

Fixtures use unique synthetic OTPs, TOTP seeds, PEM-like private-key strings,
and installation-token strings.  Capture process arguments, environments,
tool/JSON responses, structured logs, package contents, HTTP diagnostics, and
CI artifacts, then assert that none contains a fixture secret.  Preserve only
redacted evidence.

`make check` and the complete project test suite are required QA evidence in
addition to the focused tests below.  Run the release/package checks against a
freshly built release artifact, not a developer checkout.

## Adversarial acceptance matrix

| ID | Area / setup | Adversarial action | Required observation / decision |
| --- | --- | --- | --- |
| P0-01 | Readiness; daemon idle | Invoke any OTP endpoint before `confirm_ready`. | Deny without creating a challenge or lease. |
| P0-02 | Readiness; authenticated caller | Invoke `confirm_ready`; repeat it while the challenge is open. | The first call creates one opaque handle with an absolute `now + 300 s` challenge deadline; repeat is idempotent and does not extend it or create a second challenge. |
| P0-03 | Challenge | Wait past its 300 s deadline, then submit a valid OTP. | Generic denial, no active lease, and no deadline extension. |
| P0-04 | TOTP | Submit a valid unused OTP for the current handle before challenge expiry. | Exactly one lease becomes active with immutable ID and `expires_monotonic = activation + 300 s`. |
| P0-05 | TOTP/replay | Submit the same OTP again, use a prior-window consumed OTP, or submit an OTP for a different/stale handle. | Generic denial in every case; no second/replacement/extended lease; durable replay state survives the relevant window. |
| P0-06 | TOTP failures | Submit malformed, invalid, expired, and rate-limited OTPs; exercise per-challenge and global limits. | Limits are applied before validation; externally indistinguishable generic denial; challenge closes when specified by the limit; no lease. |
| P0-07 | Concurrent activation | Race two valid submissions for one open handle. | At most one succeeds; state/replay record and lease transition are atomic. |
| P0-08 | Absolute lease | During an active lease, call readiness and submit valid OTPs; use an injectable monotonic clock at `activation+299.999 s`, `+300 s`, and afterwards. | No operation renews/replaces/lengthens the lease; access can succeed only strictly before 300 s and fails at/after the deadline. Wall-clock changes cannot extend it. |
| P0-09 | Restart continuity | Restart daemon or simulate lost monotonic-clock continuity with a previously active lease. | No active lease is restored; PAM and push fail closed. |
| P0-10 | PAM/sudo allow path | With a live lease, run an allowed `sudo` command twice through the package PAM service. | Both invocations independently query live daemon state; only `codex` is authorized; no command exemption or sudo timestamp reuse exists. |
| P0-11 | PAM/sudo deny paths | Run as another user; stop daemon; corrupt/version-mismatch/oversize IPC request; make socket unavailable; use expired/stale lease. | Every case denies sudo. PAM never permits on IPC, configuration, peer-identity, protocol, or service failure. |
| P0-12 | PAM policy scope | Inspect installed sudoers and PAM configuration and exercise a non-`codex` sudo user. | `timestamp_timeout=0` and `pam_service=codex-authority` apply only to `codex`; package adds no privilege to other users. |
| P0-13 | IPC boundary | Connect as non-root/non-`codex` as appropriate, use wrong socket ownership/mode, unknown version/type, truncated/malformed/oversize messages, or forged claimed identity. | Root-managed socket and `SO_PEERCRED` govern authorization; schema/size/version failures deny; protocol carries no OTP or secret. |
| P0-14 | Push happy paths | As `codex` with a live lease and canonical clean worktree, push one fast-forward branch update to `github.com/autotaker/kakesu` `main`, then `task/TASK-*`. | Each permitted update is non-force, single-ref, and uses observed remote old OID; result reveals at most branch and commit IDs. |
| P0-15 | Push authorization | Invoke push as another uid, without/after a lease, or outside configured canonical worktree. | Deny before transport; no token is issued or disclosed. |
| P0-16 | Push ref/remote constraints | Attempt force, tag, deletion, symbolic/wildcard ref, multiple refs, URL override, wrong repository, and non-allowed branch. | Deny every request; no general Git passthrough is available. |
| P0-17 | Push repository safety | Use tracked/staged/untracked-dirty worktree, oversized object/pack, local non-fast-forward history, or invalid local commit. | Deny every request before a successful push. |
| P0-18 | Push race/transport | Change remote head after advertisement; return stale-old-OID/non-fast-forward, ambiguous transport result, or HTTP failure. | No retry as force and no success claim; operation fails closed. |
| P0-19 | Secret redaction | Exercise readiness, bad/good OTP, daemon/PAM/IPC/push failures, HTTP transport errors, and package/release build with synthetic secret fixtures. | Fixtures occur in none of argv, environment, response, logs, audit data, error text, source, package, SBOM, checksum, or release artifact. Config contains paths/IDs only; token never reaches git/helper/child process. |
| P0-20 | Expiry ordering | Hold a tracked privileged process group until lease expiry and issue PAM/push requests at expiry while cleanup is pending. | New privileged operations fail before cleanup begins or completes. |
| P0-21 | Expiry cleanup | Register matching root process group, unrelated process, and reused-PID simulation; expire the lease. | Matching tracked group receives SIGTERM, bounded wait, then SIGKILL if needed; start-time mismatch/reused PID and unrelated process are untouched. Evidence labels cleanup best effort. |
| P0-22 | Residual-risk boundary | Attempt/inspect escaped tracking and persistent full-root change scenarios; review operations/threat documentation. | Product makes no containment claim beyond tracked groups; documentation states detective controls and incident response. This does not convert unsupported persistence prevention into an acceptance requirement. |
| P0-23 | Debian contents | Build release `.deb`; inspect ownership, modes, dependencies, maintainer scripts, installed units, PAM module, sudoers snippet, docs, and package file list. | Root-owned secret/config paths and required runtime files are present; package contains no source checkout or secrets and does not broaden sudo policy. |
| P0-24 | Provenance/release | Inspect GitHub-hosted release workflow and a produced release: pinned actions, checksums, SBOM if supplied, provenance attestation, uploaded `.deb`. Verify checksums and provenance using documented tooling. | Artifact is traceable to the release build and checksum-valid; provenance/checksum failure blocks installation acceptance. |
| P0-25 | Clean Ubuntu rollout | In a clean Ubuntu 24.04 amd64 VM/container with no repository/source checkout, install only the released `.deb` and documented prerequisites; configure operator-owned secrets/App values; start service and run smoke tests. | Install/start/rollback instructions work; files/services/PAM/sudo integration are installed from package; no build-from-checkout path is required. Real host PAM activation/App provisioning is recorded separately as rollout evidence. |

## Required test layers

| Layer | Required evidence |
| --- | --- |
| Unit | State transitions, monotonic 300-second edges, replay persistence, limits, concurrent compare-and-consume, redaction, config permissions, protocol parser, ref/remote validation, and PID start-time checks. |
| Integration | Root-managed Unix-socket peer checks, isolated PAM/sudo fixture, local bare-repository/fake GitHub API fixture, conditional remote race, process-group expiry, and secret scans of captured diagnostics. |
| Package/release | PAM compilation with warnings-as-errors, Go tests/static analysis/linter, Debian build, workflow validation, checksum/provenance validation, package file/secret scan, and clean Ubuntu install/rollback. |

## Failure classification and handoff

Classify each failure before attribution; preserve command, artifact version,
redacted excerpt, reproduction conditions, and affected matrix IDs.

| Classification | Use when | Next disposition |
| --- | --- | --- |
| Implementation defect | Delivered behavior conflicts with an approved P0 row. | FAIL; return to DEV with minimal reproduction. |
| Regression | Previously passing approved behavior fails on the candidate artifact/baseline comparison. | FAIL; return to DEV/release owner with comparison evidence. |
| QA-plan defect | Test is ambiguous, invalid, or contradicts TASK/PLAN while product behavior satisfies them. | Pause attribution; amend and re-approve QA plan. |
| Requirement/design gap | TASK/PLAN omits a decision needed to judge observed behavior. | Escalate to task authority; do not infer a product failure. |
| Environment/rollout | Clean fixture or operator prerequisite is demonstrably wrong/unavailable, while artifact behavior cannot yet be assessed. | Record blocker and rerun in corrected environment; do not assign DEV fault. |

## Session efficiency note (not acceptance)

The delivery process may target a **30-minute QA session** for focused setup,
execution, and evidence handoff.  This is an operational planning heuristic
only: it is not a release criterion, acceptance threshold, SLO, or reason to
skip/reduce any matrix row, full check, clean-host test, or failure
classification.
