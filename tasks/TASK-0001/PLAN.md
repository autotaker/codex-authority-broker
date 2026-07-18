# PLAN — TASK-0001

## Delivery decision

Implement a Linux amd64 / Ubuntu 24.04 authority broker as a separately
released Debian package.  The package contains a root-owned Go daemon, a
minimal PAM shared object, an MCP-compatible local chat bridge, a constrained
GitHub push client, systemd units, sudoers/PAM configuration snippets, and an
operator install/rollback guide.  It is installed from the release `.deb`;
the target host neither needs nor receives a source checkout.

**DEV profile:** `dev-sol` / `high`.  This task crosses PAM, privileged IPC,
process lifecycle, and GitHub transport boundaries; it is not suitable for the
lower-risk Luna profile.

## Fixed design decisions

### Authority state machine

The daemon owns the only mutable authority state, protected by a single
transactional state machine and a monotonic-clock deadline:

```text
idle --human readiness--> challenge(open, deadline=now+300s)
challenge --valid unused OTP--> lease(active, deadline=activation+300s)
challenge --bad/replayed/rate-limited/timeout--> idle
lease --deadline--> expired --cleanup complete--> idle
```

- `readiness` is an explicit, authenticated human-confirmation tool call.  It
  creates no lease and returns an opaque challenge handle; duplicate readiness
  calls are idempotent for the currently open challenge.
- An OTP is accepted only while that handle is open, belongs to the current
  challenge, validates against the configured TOTP secret within the configured
  skew window, and has not previously been consumed.  Store an HMAC of the
  accepted time-step (not the OTP) in durable state through the relevant replay
  window.  Serialize compare-and-consume with the state transition, so two
  concurrent submissions cannot both activate a lease.
- Apply per-challenge and global failed-attempt limits before validation;
  return one generic denial for bad, replayed, expired, and rate-limited input.
  A failure closes the challenge where required by the limit; no operation can
  extend either deadline.
- Activation sets an immutable `lease_id`, `activated_monotonic`, and
  `expires_monotonic = activated_monotonic + 300 seconds`.  A readiness or OTP
  submission during an active lease cannot renew, replace, or lengthen it.
  Reboot or loss of monotonic continuity fails closed (no restored active
  lease).

### Chat and secret handling

`cmd/authority-chat-bridge` exposes exactly `confirm_ready` and
`submit_otp` to the local Codex/MCP integration over a root-managed Unix
socket.  The bridge streams OTP bytes directly to the daemon; it never places
them in command arguments, environment variables, JSON responses, audit
records, or error text.  Tool telemetry uses redacted parameter schemas.

The daemon reads the TOTP seed and GitHub App private key from separate
root-owned files opened after start-up, retains their bytes only in memory,
and zeroes temporary buffers where the runtime permits.  Configuration holds
only file paths, GitHub App ID, installation ID, allowed repository, and
branch policy.  Installation tokens are minted on demand, held only by the
in-process GitHub HTTP transport, and are never handed to `git`, a credential
helper, or a child process.

### Live sudo authorization

`pam/pam_codex_lease.c` is a deliberately small PAM module for a dedicated
`codex-authority` PAM service.  On every `sudo` authentication it verifies
`PAM_USER == codex`, connects to the daemon's root-only Unix `SOCK_SEQPACKET`
socket, and asks whether the current monotonic time is before the live lease
deadline.  The daemon derives identity from `SO_PEERCRED` and permits PAM
queries only from root; malformed, unavailable, stale, or unauthorized IPC
denies access.  The PAM protocol carries no OTP or secret material.

Package-provided `/etc/sudoers.d/codex-authority` applies `timestamp_timeout=0`
and `pam_service=codex-authority` to `codex`; it grants no command exemption.
Thus every sudo invocation re-enters PAM and observes expiry.  The package
does not alter other users' sudo policy.

### Restricted push authorization

`cmd/authority-push` is the only push entry point and talks to the daemon over
the same peer-credentialed local API.  The daemon authorizes a request only
when the caller uid is `codex` and the live lease exists.  It accepts a
configured, canonical worktree root and permits only a clean worktree (no
tracked, staged, or untracked changes), remote exactly
`github.com/autotaker/kakesu`, and destination exactly `main` or
`task/TASK-*`.

The daemon performs Git inspection and the smart-HTTP push through an
in-process Go client:

- permit one branch ref update only; reject tag, deletion, symbolic ref,
  force, wildcard, URL override, and multiple-ref requests;
- resolve the local commit, fetch the authoritative remote head immediately
  before push, require the remote head to be an ancestor of the local commit,
  and submit a non-force update with the observed old object ID;
- treat a changed remote advertisement, non-fast-forward result, transport
  ambiguity, oversized object/pack, or any repository cleanliness failure as
  denial/failure without retrying as force.

The GitHub App token is generated in the daemon and supplied solely to its
custom HTTPS transport.  Request/transport logging is disabled or redacted;
the push result exposes only branch and commit IDs.  This is intentionally not
a general-purpose shell or Git credential service.

### Expiry and root-process cleanup

At the absolute deadline the daemon atomically makes new PAM and push checks
fail.  It records root process groups spawned via the lease-aware execution
path (pid, start time, process group) and then sends `SIGTERM`, waits a bounded
period, and sends `SIGKILL` to still-matching groups.  PID start-time checks
avoid killing a reused PID.  Cleanup is best effort and auditable; it is not
claimed to contain processes that escaped tracking or persistent full-root
changes, which remain explicitly detective-control risks.

## Owned implementation paths and interfaces

| Path | Ownership / public interface |
| --- | --- |
| `go.mod`, `go.sum` | Go module and pinned dependencies. |
| `cmd/authorityd/` | Root daemon entry point, config loading, systemd readiness. |
| `cmd/authority-chat-bridge/` | MCP/local chat tools: `confirm_ready`, `submit_otp`. |
| `cmd/authority-push/` | Codex-only restricted push CLI; no general Git passthrough. |
| `internal/lease/` | State machine, monotonic deadlines, durable replay/rate-limit state. |
| `internal/totp/` | Validation and non-secret replay fingerprinting. |
| `internal/ipc/` | Versioned Unix seqpacket protocol, peer credentials, authorization. |
| `internal/githubpush/` | Repository/ref validation, clean-tree checks, GitHub App HTTPS transport and conditional fast-forward push. |
| `internal/process/` | Tracked root process-group registry and expiry termination. |
| `internal/redact/` | Structured logging/response redaction and secret-safe error policy. |
| `pam/pam_codex_lease.c`, `pam/Makefile` | PAM client module and ABI-focused tests. |
| `packaging/debian/`, `packaging/systemd/`, `packaging/pam/`, `packaging/sudoers/` | Debian artifact and installed service/PAM/sudo configuration. |
| `.github/workflows/release.yml` | GitHub-hosted release build, test, checksums, provenance attestation, release upload. |
| `docs/install.md`, `docs/operations.md`, `docs/threat-model.md` | Offline install/rollback, host rollout, incident response, and residual-risk documentation. |
| `*_test.go`, `tests/integration/` | Unit, protocol, fake-GitHub, PAM/sudo, and expiry integration coverage. |

No source checkout, host PAM activation, or GitHub App provisioning is
performed by tests.  Documentation makes those operator-owned rollout steps
explicit.

## Threat boundaries and fail-closed rules

| Boundary | Control and failure behavior |
| --- | --- |
| Human chat → bridge | Readiness is explicit; OTP input is streamed/redacted and generic failures reveal no validity signal. |
| Bridge/codex → root daemon | Root-managed socket plus `SO_PEERCRED`, fixed message sizes/version, schema validation, and `codex` UID checks. |
| PAM/sudo → daemon | PAM module is root-peer-only, checks the live deadline each call, and denies on IPC/config/service error. |
| Daemon → secrets | Root-only files, memory-only secret material, no argv/env/response/log/artifact disclosure. |
| Daemon → GitHub | App JWT and installation token stay inside the HTTPS transport; remote/repo/refs/old-OID are allowlisted and checked. |
| Lease → root processes | Expiry revokes new authority before bounded best-effort tracked-process termination; persistence remains out of scope and documented. |
| CI/release → host | GitHub-hosted build, pinned action revisions, checksums and provenance verified before install; no build-from-checkout host path. |

## DEV sequence and focused evidence

1. Establish the Go module, config schema/permissions validation, redacted
   logger, Unix protocol, and state-machine unit tests.  Test every transition,
   exactly-300-second absolute timeout, restart fail-closed behavior,
   duplicate/concurrent OTP submission, replay, and rate limits using an
   injectable monotonic clock.
2. Implement the chat bridge and daemon secret boundary.  Add tests that scan
   captured logs, errors, JSON/tool responses, child argv, and environment for
   OTP/seed/key/token fixtures.
3. Implement the PAM module and package sudo/PAM files.  Compile it with
   warnings-as-errors and run an isolated sudo/PAM integration fixture proving
   an allowed sudo call, no sudo timestamp reuse, lease expiry denial, and
   daemon-unavailable denial.
4. Implement restricted push with a local bare-repository/GitHub-API fixture.
   Cover allowed `main` and `task/TASK-*` fast-forward pushes plus force, tag,
   delete, non-FF, dirty, oversized, wrong remote, multiple-ref, remote-race,
   and expired-lease denials.  Assert token redaction across HTTP diagnostics.
5. Implement process tracking/expiry and integration tests that verify new
   operations are denied before cleanup, tracked root groups receive
   termination, and reused-PID safeguards do not target an unrelated process.
6. Add Debian packaging, release workflow, SBOM/checksum/provenance validation,
   and install documentation.  Test package contents in a clean Ubuntu 24.04
   container/VM without the source tree present.

Run focused tests during each step, then `go test ./...`, PAM/package build
checks, static analysis (`go vet ./...` and configured linter), integration
suite, secret-artifact scan, and release-workflow validation.  The QA plan must
independently include a clean-host package installation and an adversarial
authorization matrix; host rollout remains manual evidence, not a unit-test
side effect.

## Estimate and completion evidence

Estimate: **10–14 engineering days** (state/IPC 2–3, PAM/sudo 2–3, push and
GitHub transport 3–4, packaging/release/docs 1–2, integration hardening 2–3).

DEV is complete only with the checks above passing and evidence that the
release package contains binaries/configuration/docs but no source tree or
secrets.  Independent REVIEW must inspect the security boundaries and run the
full check suite; independent QA must execute the approved adversarial matrix
and classify any post-install failure before attribution.
