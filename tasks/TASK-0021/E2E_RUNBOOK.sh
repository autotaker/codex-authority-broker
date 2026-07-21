#!/bin/bash
set -euo pipefail
export LC_ALL=C
export PATH=/usr/sbin:/usr/bin:/sbin:/bin
umask 077

EXPECTED_ARCHIVE_SHA256="ab93c6dcc6d82f7aa93a67c40f132d81bf6d6d8b5b1ff75c230732fd53890979"
MARKER=/run/codex-authority-e2e-fixture
PERSISTENCE=/var/tmp/codex-authority-e2e-root-marker

emit() {
  local id=$1 result=$2 count=$3 label=$4 digest
  digest=$(printf '%s' "$label" | sha256sum | awk '{print $1}')
  printf '%s result=%s count=%s digest=%s\n' "$id" "$result" "$count" "$digest"
}

require_root() {
  [[ $(id -u) -eq 0 ]]
}

capture_host_state() {
  local output=$1
  {
    stat -c '%n %F %a %u %g' /etc/passwd /etc/group /etc/shadow /etc/gshadow
    sha256sum /etc/passwd /etc/group /etc/shadow /etc/gshadow \
      /etc/sudo.conf /etc/nsswitch.conf /etc/pam.d/sudo /etc/sudoers
    find /etc/sudoers.d -xdev -type f -printf '%p %m %U %G\n' -exec sha256sum {} \; | sort
    getent passwd coding-agent || true
    getent group coding-agent || true
    if [[ -d /var/lib/coding-agent ]]; then
      find /var/lib/coding-agent -xdev -printf '%P %y %m %U %G\n' -type f -exec sha256sum {} \; | sha256sum
    else
      printf '%s\n' home-absent
    fi
    systemctl show codex-authority-broker.service --property=LoadState,ActiveState,UnitFileState --value
    for path in /etc/codex-authority /etc/pam.d/codex-authority /etc/sudoers.d/codex-authority \
      /etc/systemd/system/codex-authority-broker.service /var/lib/codex-authority-installer \
      /var/lib/codex-authority-recovery /var/lib/codex-authority-recovery.completed \
      /usr/local/lib/codex-authority /usr/local/bin/codex-authority /usr/local/bin/codex-authority-broker \
      /usr/local/bin/codex-authority-sudo /usr/local/sbin/codex-authority-admin \
      /usr/local/sbin/codex-authority-verify /usr/local/sbin/codex-authority-recover \
      /usr/local/sbin/codex-authority-uninstall; do
      if [[ -e $path || -L $path ]]; then stat -c '%n %F %a %u %g' "$path"; else printf '%s absent\n' "$path"; fi
    done
  } >"$output"
  chmod 0600 "$output"
}

assert_uninstalled() {
  local identity_mode=$1 path uid
  for path in \
    /run/codex-authority.sock /etc/codex-authority /etc/pam.d/codex-authority \
    /etc/sudoers.d/codex-authority /etc/systemd/system/codex-authority-broker.service \
    /var/lib/codex-authority-installer /var/lib/codex-authority-recovery \
    /var/lib/codex-authority-recovery.completed /var/tmp/codex-authority-install \
    /usr/local/lib/codex-authority /usr/local/bin/codex-authority \
    /usr/local/bin/codex-authority-broker /usr/local/bin/codex-authority-sudo \
    /usr/local/sbin/codex-authority-admin /usr/local/sbin/codex-authority-verify \
    /usr/local/sbin/codex-authority-recover /usr/local/sbin/codex-authority-uninstall; do
    [[ ! -e $path && ! -L $path ]]
  done
  if systemctl is-active --quiet codex-authority-broker.service; then return 1; fi
  if systemctl is-enabled --quiet codex-authority-broker.service; then return 1; fi
  if pgrep -f '^/usr/local/bin/codex-authority-broker$' >/dev/null; then return 1; fi
  [[ ! -e /run/sudo/ts/coding-agent ]]
  if [[ -f /var/tmp/task0021-identity.uid ]]; then
    read -r uid </var/tmp/task0021-identity.uid
    [[ $uid =~ ^[0-9]+$ && ! -e /run/sudo/ts/$uid ]]
  fi
  if [[ $identity_mode == created ]]; then
    if getent passwd coding-agent >/dev/null; then return 1; fi
    if getent group coding-agent >/dev/null; then return 1; fi
  else
    getent passwd coding-agent >/dev/null
    getent group coding-agent >/dev/null
  fi
}

case "${1-}" in
  preflight)
    [[ $# -eq 1 ]]
    require_root
    # shellcheck source=/dev/null
    . /etc/os-release
    [[ $ID == ubuntu && $VERSION_ID == 24.04 ]]
    [[ $(dpkg --print-architecture) == amd64 ]]
    command -v visudo >/dev/null
    command -v qrencode >/dev/null
    command -v getcap >/dev/null
    [[ ! -e /var/lib/codex-authority-installer ]]
    [[ ! -e /var/lib/codex-authority-recovery ]]
    [[ ! -e /etc/sudoers.d/codex-authority ]]
    emit Q21-02 PASS 1 ubuntu-24.04-amd64-prerequisites
    ;;
  host-state-before)
    [[ $# -eq 1 ]]
    require_root
    capture_host_state /var/tmp/task0021-host.before
    emit Q21-23 OBSERVED 1 "before:$(sha256sum /var/tmp/task0021-host.before | awk '{print $1}')"
    ;;
  host-state-compare)
    [[ $# -eq 1 ]]
    require_root
    [[ -f /var/tmp/task0021-host.before ]]
    capture_host_state /var/tmp/task0021-host.after
    cmp /var/tmp/task0021-host.before /var/tmp/task0021-host.after
    emit Q21-23 PASS 1 "equal:$(sha256sum /var/tmp/task0021-host.after | awk '{print $1}')"
    ;;
  rollback-compare)
    [[ $# -eq 1 && ! -e $MARKER ]]
    require_root
    [[ -f /var/tmp/task0021-host.before ]]
    capture_host_state /var/tmp/task0021-host.after
    cmp /var/tmp/task0021-host.before /var/tmp/task0021-host.after
    emit Q21-18 PASS 1 "rollback-equal:$(sha256sum /var/tmp/task0021-host.after | awk '{print $1}')"
    ;;
  crash-result)
    [[ $# -eq 2 && $2 =~ ^(created|reuse)$ && ! -e $MARKER ]]
    require_root
    if [[ -x /usr/local/sbin/codex-authority-verify ]]; then
      /usr/local/sbin/codex-authority-verify >/dev/null
      outcome=installed
    else
      assert_uninstalled "$2"
      outcome=uninstalled
    fi
    emit Q21-19 PASS 1 "crash-recovered:$outcome"
    ;;
  verify-archive)
    [[ $# -eq 2 ]]
    require_root
    [[ $EXPECTED_ARCHIVE_SHA256 =~ ^[0-9a-f]{64}$ ]]
    [[ $(sha256sum "$2" | awk '{print $1}') == "$EXPECTED_ARCHIVE_SHA256" ]]
    expected=$(mktemp)
    actual=$(mktemp)
    staging=$(mktemp -d)
    trap 'rm -f "$expected" "$actual"; rm -rf "$staging"' EXIT
    printf '%s\n' SHA256SUMS bin/codex-authority bin/codex-authority-broker \
      bin/codex-authority-sudo deploy/pam/codex-authority deploy/sudo/codex-authority \
      deploy/systemd/codex-authority-broker.service docs/ADMIN_MANUAL.md docs/USER_MANUAL.md \
      install/codex-authority-admin install/codex-authority-install \
      install/codex-authority-recover install/codex-authority-uninstall \
      install/codex-authority-verify install/codex_authority_installer.py | sort >"$expected"
    tar -tzf "$2" | sort >"$actual"
    cmp "$expected" "$actual"
    if tar -tvzf "$2" | awk '$1 !~ /^-/{exit 1}'; then :; else exit 1; fi
    tar -xzf "$2" -C "$staging" --no-same-owner --no-same-permissions
    (cd "$staging" && sha256sum -c SHA256SUMS >/dev/null)
    observation="$EXPECTED_ARCHIVE_SHA256:$(sha256sum "$actual" "$staging/SHA256SUMS" | awk '{print $1}' | paste -sd: -)"
    emit Q21-03 PASS 3 "$observation"
    ;;
  post-install)
    [[ $# -eq 1 ]]
    require_root
    /usr/local/sbin/codex-authority-verify >/dev/null
    account=$(getent passwd coding-agent)
    group=$(getent group coding-agent)
    uid=$(printf '%s' "$account" | cut -d: -f3)
    gid=$(printf '%s' "$account" | cut -d: -f4)
    [[ $uid -ne 0 && $uid -eq $gid ]]
    [[ $(printf '%s' "$group" | cut -d: -f3) -eq $gid ]]
    [[ $(id -G coding-agent) == "$gid" ]]
    printf '%s\n' "$uid" >/var/tmp/task0021-identity.uid
    chmod 0600 /var/tmp/task0021-identity.uid
    visudo -cf /etc/sudoers >/dev/null
    ! grep -q NOPASSWD /etc/sudoers.d/codex-authority
    [[ $(grep -c '^Defaults:coding-agent pam_.*service=codex-authority$' /etc/sudoers.d/codex-authority) -eq 3 ]]
    observation="$uid:$gid:$(sha256sum /etc/sudoers.d/codex-authority /etc/pam.d/codex-authority | awk '{print $1}' | paste -sd: -):$(systemctl is-active codex-authority-broker.service)"
    emit Q21-05 PASS 4 "$observation"
    emit Q21-11 PASS 5 "$observation"
    ;;
  sudo-allow)
    [[ $# -eq 1 ]]
    [[ $(id -un) == coding-agent ]]
    [[ $(sudo -- /usr/bin/id -u) == 0 ]]
    [[ $(sudo -u root -- /usr/bin/id -u) == 0 ]]
    [[ $(sudo -g root -- /usr/bin/id -g) == 0 ]]
    [[ $(sudo -s -- -c '/usr/bin/id -u') == 0 ]]
    [[ $(sudo -i /usr/bin/id -u) == 0 ]]
    [[ $(SUDO_ASKPASS=/bin/false sudo -A -- /usr/bin/id -u) == 0 ]]
    sudo -- /usr/bin/touch "$PERSISTENCE"
    emit Q21-12 PASS 7 uncached-sudo-forms-and-root-marker
    ;;
  sudo-deny)
    [[ $# -eq 1 ]]
    [[ $(id -un) == coding-agent ]]
    denied=0
    if ! timeout 10 sudo -- /usr/bin/true </dev/null >/dev/null 2>&1; then ((denied+=1)); fi
    if ! timeout 10 sudo -u root -- /usr/bin/true </dev/null >/dev/null 2>&1; then ((denied+=1)); fi
    if ! timeout 10 sudo -g root -- /usr/bin/true </dev/null >/dev/null 2>&1; then ((denied+=1)); fi
    if ! timeout 10 sudo -s -- -c /usr/bin/true </dev/null >/dev/null 2>&1; then ((denied+=1)); fi
    if ! timeout 10 sudo -i /usr/bin/true </dev/null >/dev/null 2>&1; then ((denied+=1)); fi
    if ! timeout 10 env SUDO_ASKPASS=/bin/false sudo -A -- /usr/bin/true </dev/null >/dev/null 2>&1; then ((denied+=1)); fi
    [[ $denied -eq 6 ]]
    emit Q21-13 PASS "$denied" "fresh-sudo-denied:$denied"
    ;;
  sudo-deny-one)
    [[ $# -eq 1 && $(id -un) == coding-agent ]]
    if timeout 12 sudo -- /usr/bin/true </dev/null >/dev/null 2>&1; then exit 1; fi
    emit Q21-13 PASS 1 one-faulted-sudo-denied
    ;;
  ready-deny-one)
    [[ $# -eq 1 && $(id -un) == coding-agent ]]
    if timeout 12 /usr/local/bin/codex-authority ready </dev/null >/dev/null 2>&1; then exit 1; fi
    emit Q21-13 PASS 1 audit-failure-ready-denied
    ;;
  expiry-noncontainment)
    [[ $# -eq 1 ]]
    [[ $(id -un) == coding-agent ]]
    sleep 301
    if sudo -- /usr/bin/true >/dev/null 2>&1; then
      emit Q21-13 FAIL 1 post-expiry-allow
      exit 1
    fi
    [[ -f $PERSISTENCE && $(stat -c %u:%g:%a "$PERSISTENCE") == 0:0:644 ]]
    emit Q21-15 PASS 1 root-effect-outlives-lease
    ;;
  inject)
    [[ $# -eq 3 ]]
    require_root
    [[ $2 == fail || $2 == kill ]]
    [[ $3 =~ ^[A-Z_]+$ ]]
    printf '%s %s\n' "$2" "$3" >"$MARKER"
    chown root:root "$MARKER"
    chmod 0600 "$MARKER"
    emit Q21-18 ARMED 1 "fixture:$2:$3"
    ;;
  fault-socket)
    [[ $# -eq 2 && $2 =~ ^(regular|malformed|timeout)$ ]]
    require_root
    systemctl stop codex-authority-broker.service
    rm -f /run/codex-authority.sock /run/codex-authority-e2e-fault.pid
    if [[ $2 == regular ]]; then
      install -o coding-agent -g coding-agent -m 0660 /dev/null /run/codex-authority.sock
    else
      python3 -c 'import os,socket,sys,time
p="/run/codex-authority.sock"
s=socket.socket(socket.AF_UNIX); s.bind(p); s.listen(1)
c,_=s.accept()
if sys.argv[1]=="malformed": c.sendall(b"malformed\n")
else: time.sleep(15)
c.close(); s.close()' "$2" >/dev/null 2>&1 &
      printf '%s\n' "$!" >/run/codex-authority-e2e-fault.pid
      for _ in $(seq 1 50); do [[ -S /run/codex-authority.sock ]] && break; sleep .1; done
      chown coding-agent:coding-agent /run/codex-authority.sock
      chmod 0660 /run/codex-authority.sock
    fi
    emit Q21-13 ARMED 1 "socket-fault:$2"
    ;;
  fault-audit)
    [[ $# -eq 1 ]]
    require_root
    systemctl stop codex-authority-broker.service
    rm -f /run/codex-authority.sock /run/codex-authority-e2e-fault.pid
    /usr/local/bin/codex-authority-broker >/dev/full 2>/dev/null &
    printf '%s\n' "$!" >/run/codex-authority-e2e-fault.pid
    for _ in $(seq 1 50); do [[ -S /run/codex-authority.sock ]] && break; sleep .1; done
    [[ -S /run/codex-authority.sock ]]
    emit Q21-13 ARMED 1 audit-writer-fault
    ;;
  clear-fault)
    [[ $# -eq 1 ]]
    require_root
    if [[ -f /run/codex-authority-e2e-fault.pid ]]; then
      read -r pid </run/codex-authority-e2e-fault.pid
      [[ $pid =~ ^[0-9]+$ ]]
      if kill -0 "$pid" 2>/dev/null; then
        command_line=$(tr '\0' ' ' <"/proc/$pid/cmdline")
        [[ $command_line == *'/run/codex-authority.sock'* || $command_line == '/usr/local/bin/codex-authority-broker ' ]]
        kill "$pid"
      fi
    fi
    rm -f /run/codex-authority-e2e-fault.pid /run/codex-authority.sock
    systemctl start codex-authority-broker.service
    emit Q21-13 CLEARED 1 fault-cleared
    ;;
  audit-count)
    [[ $# -eq 1 ]]
    require_root
    count=$(journalctl -u codex-authority-broker.service --no-pager -o cat | awk 'index($0,"\"scope\":\"authorize\",\"result\":\"allow\""){count++} END{print count+0}')
    emit Q21-12 OBSERVED "$count" "authorize-allow-count:$count"
    ;;
  audit-total)
    [[ $# -eq 1 ]]
    require_root
    count=$(journalctl -u codex-authority-broker.service --no-pager -o cat | awk 'index($0,"\"scope\":"){count++} END{print count+0}')
    emit Q21-13 OBSERVED "$count" "audit-total:$count"
    ;;
  audit-no-delta)
    [[ $# -eq 2 && $2 =~ ^[0-9]+$ ]]
    require_root
    count=$(journalctl -u codex-authority-broker.service --no-pager -o cat | awk 'index($0,"\"scope\":"){count++} END{print count+0}')
    [[ $count -eq $2 ]]
    emit Q21-13 PASS 0 "pre-admission-audit-delta:$2:$count"
    ;;
  audit-delta)
    [[ $# -eq 2 && $2 =~ ^[0-9]+$ ]]
    require_root
    count=$(journalctl -u codex-authority-broker.service --no-pager -o cat | awk 'index($0,"\"scope\":\"authorize\",\"result\":\"allow\""){count++} END{print count+0}')
    [[ $((count - $2)) -eq 7 ]]
    emit Q21-12 PASS 7 "authorize-allow-delta:$2:$count"
    ;;
  clear-inject)
    [[ $# -eq 1 ]]
    require_root
    rm -f "$MARKER"
    emit Q21-18 CLEARED 1 fixture-injection-cleared
    ;;
  cleanup-marker)
    [[ $# -eq 1 ]]
    require_root
    rm -f "$PERSISTENCE" "$MARKER" /var/tmp/task0021-host.before /var/tmp/task0021-host.after \
      /var/tmp/task0021-identity.uid
    ;;
  post-uninstall)
    [[ $# -eq 2 && $2 =~ ^(created|reuse)$ ]]
    require_root
    assert_uninstalled "$2"
    emit Q21-21 PASS 19 "no-residue:$(findmnt -rn -o TARGET /):$(systemctl is-active codex-authority-broker.service || true)"
    ;;
  *)
    exit 64
    ;;
esac
