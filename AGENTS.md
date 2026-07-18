# Development contract

- Work proceeds through PLAN, DEV, independent REVIEW, and QA gates.
- DEV, reviewer, and QA roles must remain separate.
- Child agents must not stage, commit, merge, or write `.git`.
- Only the main agent owns Git operations and release publication.
- Secrets, TOTP values, GitHub App keys, and installation tokens must never be committed or logged.

