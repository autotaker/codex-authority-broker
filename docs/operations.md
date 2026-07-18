# Authority operations

Install a checksum- and provenance-verified Debian release on Ubuntu 24.04
amd64; a source checkout is not a host prerequisite. Keep config, base32 TOTP
seed, independent replay HMAC key, and GitHub App key in distinct root-owned
mode 0600 files. Host PAM activation and App provisioning are manual rollout.

`confirm_ready` opens one absolute 300-second challenge. A valid unused OTP
opens one non-renewing 300-second monotonic lease. Restart loses authority.
Stop the daemon, revoke the App, rotate secrets, and inspect external audit and
integrity telemetry on suspected compromise.

Full-root authority can persist or escape tracked cleanup. Cleanup is best
effort; this product makes no containment claim for such persistence.
