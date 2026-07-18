# TASK-0001: Implement a TOTP-backed Codex authority lease

## Objective

Build a separately released authority broker that grants the dedicated `codex` OS user an absolute five-minute full-sudo and restricted GitHub-push lease after a human supplies a TOTP through chat.

## Acceptance criteria

- No challenge is created until the human confirms readiness; a challenge then waits up to five minutes for an OTP.
- A valid, unused TOTP activates one non-renewing 300-second lease; invalid, replayed, rate-limited, expired, or concurrent activation fails closed.
- Every `sudo` call checks the live lease through a dedicated PAM service with sudo timestamp caching disabled.
- The same lease permits only fast-forward pushes to `autotaker/kakesu` branches `main` and `task/TASK-*`; force, tag, delete, oversize, dirty, and remote-race cases fail closed.
- Expiry first blocks new privileged operations, then best-effort terminates tracked root processes.
- OTPs, seeds, GitHub App private keys, and installation tokens never appear in argv, environment, responses, logs, source, or release artifacts.
- Release artifacts are built in GitHub-hosted Actions, carry provenance attestation and checksums, and are installed without a source checkout.
- Unit and integration tests cover the state machine, absolute timeout, replay/rate limiting, authorization boundaries, redaction, and expiry behavior.

## Scope boundaries

- Full-root persistence cannot be prevented after authorization; external trace monitoring and integrity checks are detective controls.
- Host PAM/systemd activation and real GitHub App provisioning are rollout steps, not actions performed by unit tests.
- Linux amd64 and Ubuntu 24.04 are the initial supported platform.

