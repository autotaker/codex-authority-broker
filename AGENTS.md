# Development contract

- The development process exists to improve quality, safety, and delivery speed. It is not an end in itself.
- `docs/development/README.md` is the canonical development process. This file only defines routing and hard boundaries.
- Product and safety-contract changes use PLAN, DEV, independent REVIEW, and QA. DEV, reviewer, and QA roles remain separate.
- Documentation and evidence maintenance use the lightweight path defined by the canonical process.
- Review the development process after every five completed development tasks. A process review is maintenance, not a development task: it needs a concise proposal and independent consistency review, but no TASK, PLAN, DEV, or QA artifacts.
- A serious process defect triggers an immediate review, without waiting for the fifth completed task.
- Child agents must not stage, commit, merge, or write `.git`. Only the main agent owns Git operations and release publication.
- Secrets, TOTP values, GitHub App keys, and installation tokens must never be committed or logged.
