#!/bin/bash
set -euo pipefail
GO=$(command -v go)
export LC_ALL=C
export PATH=/usr/sbin:/usr/bin:/sbin:/bin
umask 077

EXPECTED=ab93c6dcc6d82f7aa93a67c40f132d81bf6d6d8b5b1ff75c230732fd53890979
ROOT=$(cd "$(dirname "$0")/../.." && pwd -P)
OUTPUT=/tmp/task0021-candidate
STAGING=$OUTPUT/staging
ARCHIVE=$OUTPUT/codex-authority-linux-amd64.tar.gz
[[ $# -eq 0 && ! -e $OUTPUT ]]
[[ $($GO version) == 'go version go1.23.12 linux/amd64' ]]
mkdir -p "$STAGING/bin" "$STAGING/deploy/pam" "$STAGING/deploy/sudo" \
  "$STAGING/deploy/systemd" "$STAGING/install" "$STAGING/docs" "$OUTPUT/go-cache"

for target in codex-authority-broker codex-authority codex-authority-sudo; do
  (cd "$ROOT" && GOCACHE="$OUTPUT/go-cache" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    "$GO" build -buildvcs=false -trimpath -o "$STAGING/bin/$target" "./cmd/$target")
done
install -m 0644 "$ROOT/deploy/pam/codex-authority" "$STAGING/deploy/pam/codex-authority"
install -m 0644 "$ROOT/deploy/sudo/codex-authority" "$STAGING/deploy/sudo/codex-authority"
install -m 0644 "$ROOT/deploy/systemd/codex-authority-broker.service" \
  "$STAGING/deploy/systemd/codex-authority-broker.service"
install -m 0755 "$ROOT/install/codex-authority-install" "$ROOT/install/codex-authority-verify" \
  "$ROOT/install/codex-authority-admin" "$ROOT/install/codex-authority-recover" \
  "$ROOT/install/codex-authority-uninstall" "$STAGING/install/"
install -m 0755 "$ROOT/install/codex_authority_installer.py" "$STAGING/install/"
install -m 0644 "$ROOT/docs/ADMIN_MANUAL.md" "$ROOT/docs/USER_MANUAL.md" "$STAGING/docs/"
(cd "$STAGING" && sha256sum bin/codex-authority bin/codex-authority-broker \
  bin/codex-authority-sudo deploy/pam/codex-authority deploy/sudo/codex-authority \
  deploy/systemd/codex-authority-broker.service install/codex-authority-install \
  install/codex-authority-verify install/codex-authority-admin install/codex-authority-recover \
  install/codex-authority-uninstall install/codex_authority_installer.py docs/ADMIN_MANUAL.md \
  docs/USER_MANUAL.md | sort >SHA256SUMS)
tar --sort=name --format=gnu --mtime=@0 --owner=0 --group=0 --numeric-owner -cf - \
  -C "$STAGING" SHA256SUMS bin/codex-authority bin/codex-authority-broker \
  bin/codex-authority-sudo deploy/pam/codex-authority deploy/sudo/codex-authority \
  deploy/systemd/codex-authority-broker.service install/codex-authority-install \
  install/codex-authority-verify install/codex-authority-admin install/codex-authority-recover \
  install/codex-authority-uninstall install/codex_authority_installer.py docs/ADMIN_MANUAL.md \
  docs/USER_MANUAL.md | gzip -n >"$ARCHIVE"
actual=$(sha256sum "$ARCHIVE" | awk '{print $1}')
[[ $actual == "$EXPECTED" ]]
printf 'Q21-build result=PASS count=15 digest=%s archive=%s\n' "$actual" "$ARCHIVE"
