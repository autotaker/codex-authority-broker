#!/bin/bash
# Exact outer staging/rollback harness for TASK-0020. It handles no secret.
set -Eeuo pipefail
umask 077
export LC_ALL=C PATH=/usr/sbin:/usr/bin:/sbin:/bin

readonly STAGE=/var/tmp/codex-authority-task0020
readonly INPUT=/tmp/task0020-artifact-29720021660.AwGYdh
readonly SOURCE=/home/ubuntu/git/codex-authority-broker/tasks/TASK-0020/CANARY_RUNBOOK.sh
readonly ARCHIVE_SHA=5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd
readonly RUNBOOK_SHA=4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7

[[ $# -eq 1 && ( $1 == setup || $1 == cleanup ) && $EUID -eq 0 ]] || exit 2
readonly MODE=$1

parent_snapshot() {
  find /var/tmp -mindepth 1 -maxdepth 1 -printf '%y\t%f\t%m\t%U\t%G\t%l\n' | sort
}

emit() {
  printf 'Q20-outer mode=%s result=%s digest=%s\n' "$1" "$2" "$3"
}

if [[ $MODE == setup ]]; then
  [[ ! -e $STAGE && ! -L $STAGE ]] || exit 2
  [[ -d $INPUT && ! -L $INPUT && -f $INPUT/codex-authority-linux-amd64.tar.gz && -f $INPUT/SHA256SUMS ]] || exit 2
  [[ -f $SOURCE && ! -L $SOURCE ]] || exit 2
  before=$(parent_snapshot)
  rollback_setup() {
    trap - EXIT
    set +e
    if [[ -d $STAGE && ! -L $STAGE ]]; then
      find "$STAGE" -xdev -depth -delete
    fi
    [[ $(parent_snapshot) == "$before" ]] || exit 1
  }
  trap rollback_setup EXIT
  install -d -o 0 -g 0 -m 0700 "$STAGE" "$STAGE/rootfs"
  install -o 0 -g 0 -m 0600 "$INPUT/codex-authority-linux-amd64.tar.gz" "$STAGE/codex-authority-linux-amd64.tar.gz"
  install -o 0 -g 0 -m 0600 "$INPUT/SHA256SUMS" "$STAGE/SHA256SUMS"
  install -o 0 -g 0 -m 0500 "$SOURCE" "$STAGE/CANARY_RUNBOOK.sh"
  readlink /proc/self/ns/mnt >"$STAGE/host.mount-ns"
  readlink /proc/self/ns/pid >"$STAGE/host.pid-ns"
  printf '%s' "$before" >"$STAGE/var-tmp.before"
  chown 0:0 "$STAGE"/host.*-ns "$STAGE/var-tmp.before"
  chmod 0400 "$STAGE"/host.*-ns "$STAGE/var-tmp.before"
  [[ $(sha256sum "$STAGE/codex-authority-linux-amd64.tar.gz" | awk '{print $1}') == "$ARCHIVE_SHA" ]]
  [[ $(sha256sum "$STAGE/CANARY_RUNBOOK.sh" | awk '{print $1}') == "$RUNBOOK_SHA" ]]
  cmp -s "$INPUT/SHA256SUMS" "$STAGE/SHA256SUMS"
  cmp -s "$SOURCE" "$STAGE/CANARY_RUNBOOK.sh"
  trap - EXIT
  emit setup PASS "$(printf '%s' "$before" | sha256sum | awk '{print $1}')"
  exit 0
fi

[[ -d $STAGE && ! -L $STAGE && $(stat -c '%u:%g:%a' "$STAGE") == 0:0:700 ]] || exit 2
[[ -f $STAGE/var-tmp.before && ! -L $STAGE/var-tmp.before ]] || exit 2
before=$(<"$STAGE/var-tmp.before")
[[ ! -L $STAGE/rootfs && -d $STAGE/rootfs && -z $(find "$STAGE/rootfs" -mindepth 1 -print -quit) ]] || exit 2
mountpoint -q "$STAGE/rootfs" && exit 2
actual=$(find "$STAGE" -mindepth 1 -maxdepth 1 -printf '%f\n' | sort)
expected=$'CANARY_RUNBOOK.sh\nSHA256SUMS\ncodex-authority-linux-amd64.tar.gz\nhost.mount-ns\nhost.pid-ns\nrootfs\nvar-tmp.before'
[[ $actual == "$expected" ]] || exit 2
for link in /proc/[0-9]*/cwd /proc/[0-9]*/root /proc/[0-9]*/fd/*; do
  target=$(readlink "$link" 2>/dev/null || :)
  [[ $target != "$STAGE" && $target != "$STAGE"/* ]] || exit 2
done
find "$STAGE" -xdev -depth -delete
after=$(parent_snapshot)
[[ $after == "$before" ]] || { emit cleanup FAIL "$(printf '%s' "$after" | sha256sum | awk '{print $1}')"; exit 1; }
emit cleanup PASS "$(printf '%s' "$after" | sha256sum | awk '{print $1}')"
