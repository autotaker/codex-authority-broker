# REVIEW RESULT — TASK-0001

## Decision: FAIL

Independent review performed against `TASK.md`, approved `PLAN.md`, and
`QA_PLAN.md`.  The candidate is an incomplete foundation and does not meet the
release acceptance criteria.

## Findings

### Critical

1. **Restricted GitHub push is not implemented** (P0-14 through P0-18).
   `cmd/authority-push/main.go:20-24` asks only whether a lease is live and
   then exits successfully after printing `push authority granted; transport
   unavailable`.  It performs no worktree canonicalization/cleanliness check,
   remote/repository check, local or remote ref/OID inspection, conditional
   non-force push, size check, race handling, or GitHub App token transport.
   A live lease therefore reports a successful authority-push command without
   making the required push, and none of the fail-closed push boundaries exists.

2. **The installed service cannot be reached by the required `codex` client**
   (P0-02, P0-04, P0-10, P0-14).  The unit creates
   `/run/codex-authority` as `root:root` mode `0750`
   (`packaging/systemd/codex-authority.service:6-7`).  The daemon creates the
   socket as root mode `0660` and never assigns a `codex`-accessible group
   (`cmd/authorityd/main.go:75`, `internal/ipc/socket.go:24-31`).  A normal
   `codex` process cannot traverse the directory or connect to the socket, so
   it cannot call the chat bridge or push entry point.  No documented/setup
   artifact creates a compatible group or changes ownership.

### High

3. **Expiry process cleanup is absent** (P0-20 and P0-21).  The daemon only
   clears in-memory lease state in `internal/lease/manager.go:54-62`; there is
   no process registry, start-time verification, SIGTERM/SIGKILL sequence, or
   associated test.  This directly contradicts the approved expiry behavior.

4. **Required release/package delivery is absent** (P0-23 through P0-25).
   There is no Debian packaging metadata/build, release workflow, checksum or
   provenance generation/verification, artifact inspection, install/rollback
   guide, or clean Ubuntu package-install evidence.  The checked-in
   `packaging/` directory has only PAM, sudoers, and systemd snippets.

5. **Secret-input boundary does not satisfy the approved bridge design or its
   evidence requirement** (P0-19).  The bridge reads whole JSON requests into
   a scanner and unmarshals the OTP into a Go string
   (`cmd/authority-chat-bridge/main.go:21-37`); the daemon converts the full
   payload to another string (`cmd/authorityd/main.go:104-110`).  These copies
   cannot be zeroed, so the OTP is neither streamed directly as planned nor
   reliably cleared.  No redaction, argv/environment, HTTP diagnostic, source,
   package, SBOM, or artifact secret-scan tests are present.

6. **Mandatory test layers and adversarial matrix evidence are missing.**
   The current tests cover only a small subset of state, TOTP, and protocol
   behavior.  There are no daemon/socket peer, PAM/sudo, push/fake-GitHub,
   process-expiry, package/release, or secret-redaction integration tests.  In
   particular, the review cannot establish P0-09 through P0-25.  This agrees
   with the DEV handover's explicit list of incomplete work.

## Verification evidence

- `GOCACHE=/tmp/codex-authority-review-go-cache make check` — **PASS**:
  `go test ./...` and `go vet ./...` passed.
- PAM compilation was **not performed**: `pam/Makefile` emitted `SKIP: PAM
  headers unavailable`; therefore it provides no PAM ABI/build evidence.
- `git diff --check` — **PASS**.  The candidate files are untracked in this
  worktree, so review inspected all listed implementation files directly.

No code, Git metadata, staging area, commit, or merge was changed by review.
