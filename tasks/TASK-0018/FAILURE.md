# TASK-0018 planning failure

TASK-0018 stopped before candidate fixation. The first draft duplicated an OTP
pre-admission callback and a whole-Runtime audit transaction and reached +157
production SLOC. A bounded retry removed that duplication and produced the
smallest readable draft observed in this Task:

| File | Canonical delta |
| --- | ---: |
| `internal/backend/runtime.go` | +62 |
| `internal/ipc/server_linux.go` | +27 |
| `internal/lease/lease.go` | +13 |
| **Total** | **+102** |

The draft was never a candidate, tested, reviewed, committed, pushed, or
merged. Main restored only those three Agent-owned paths to planning commit
`0dbeec7`; the planning evidence remains intact.

The nine-line target overflow requires the backlog's ordered shedding audit.
TASK-0019 applies item 3 only: it retains a fresh correlation ID, actor, scope,
result, and expiry but drops the richer cryptographically random fixed-32-hex
ID format. It also uses permanent Runtime closure as the fail-closed state
instead of separately clearing inaccessible lease internals. No mandatory
audit field, redaction, source-free artifact, attestation, or manual-canary
requirement is shed.
