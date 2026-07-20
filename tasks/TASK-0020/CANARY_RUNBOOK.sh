#!/bin/bash
# TASK-0020 evidence harness. Product input is the fixed verified archive only.
set -Eeuo pipefail
umask 077
export LC_ALL=C PATH=/usr/sbin:/usr/bin:/sbin:/bin

readonly STAGE=/var/tmp/codex-authority-task0020
readonly SELF="$STAGE/CANARY_RUNBOOK.sh"
readonly ROOT="$STAGE/rootfs"
readonly ARCHIVE="$STAGE/codex-authority-linux-amd64.tar.gz"
readonly OUTER_SUMS="$STAGE/SHA256SUMS"
readonly HOST_MNTNS="$STAGE/host.mount-ns"
readonly HOST_PIDNS="$STAGE/host.pid-ns"
readonly ARCHIVE_SHA=5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd
readonly OWNER_UID=42020 OWNER_GID=42020 OTHER_UID=42021 OTHER_GID=42021
readonly EXPECTED_MANIFEST=$'SHA256SUMS\nbin/codex-authority\nbin/codex-authority-broker\nbin/codex-authority-sudo\ndeploy/pam/codex-authority\ndeploy/sudo/codex-authority\ndeploy/systemd/codex-authority-broker.service'

[[ $# -eq 2 && ( $1 == preflight || $1 == canary ) && $2 == "$STAGE" && $EUID -eq 0 && $0 == "$SELF" ]] || exit 2
readonly MODE=$1
[[ -d $STAGE && -d $ROOT && -f $ARCHIVE && -f $OUTER_SUMS ]] || exit 2
[[ $(stat -c '%u:%g:%a' "$STAGE" "$ROOT") == $'0:0:700\n0:0:700' ]] || exit 2
for fixed_input in "$SELF" "$ARCHIVE" "$OUTER_SUMS" "$HOST_MNTNS" "$HOST_PIDNS"; do
  [[ $(stat -c '%u:%g' "$fixed_input") == 0:0 && $(( 8#$(stat -c '%a' "$fixed_input") & 8#022 )) -eq 0 ]] || exit 2
done
[[ $(awk '/^NoNewPrivs:/{print $2}' /proc/self/status) == 0 ]] || exit 2
[[ $(readlink /proc/self/ns/mnt) != "$(<"$HOST_MNTNS")" ]] || exit 2
[[ $(readlink /proc/self/ns/pid) != "$(<"$HOST_PIDNS")" ]] || exit 2
[[ $$ -eq 1 ]] || exit 2

# This is deliberately the first privileged mutation after unshare.
mount --make-rprivate /

emit() { printf '%s result=%s count=%s digest=%s\n' "$1" "$2" "$3" "$4"; }
die() { emit "$1" FAIL 0 none; exit 1; }

capture_host() {
  local path
  for path in \
    /etc/passwd /etc/group /etc/shadow /etc/gshadow /etc/pam.d \
    /etc/sudoers /etc/sudoers.d /run/codex-authority.sock /run/sudo \
    /usr/local /usr/local/bin "$STAGE"; do
    if [[ ! -e $path && ! -L $path ]]; then
      printf 'absent\t%s\n' "$path"
      continue
    fi
    find -H "$path" -xdev -printf '%y\t%p\t%m\t%U\t%G\t%s\t%l\n' 2>/dev/null | sort
    while IFS= read -r -d '' file; do
      printf 'sha256\t%s\t' "$file"
      sha256sum "$file" | awk '{print $1}'
    done < <(find -H "$path" -xdev -type f -print0 2>/dev/null | sort -z)
  done
  stat -c 'run-root\t%F\t%n\t%a\t%u\t%g\t%s' /run
}

HOST_BEFORE="$(capture_host)"
readonly HOST_BEFORE
HOST_BEFORE_DIGEST="$(printf '%s' "$HOST_BEFORE" | sha256sum | awk '{print $1}')"
readonly HOST_BEFORE_DIGEST
declare -a MOUNTS=()
BROKER_PID=
BROKER_GEN=0
AUDIT=
CLEANUP_FAIL=0

# shellcheck disable=SC2317
cleanup() {
  local incoming=$?
  local host_after host_after_digest index
  trap - EXIT HUP INT TERM
  set +e
  [[ -z ${BROKER_PID:-} ]] || { kill "$BROKER_PID" 2>/dev/null; wait "$BROKER_PID" 2>/dev/null; }
  BROKER_PID=
  [[ ! -d $ROOT/run/private ]] || find "$ROOT/run/private" -type f -exec sh -c 'for f; do size=$(stat -c %s "$f"); dd if=/dev/zero of="$f" bs=1 count="$size" conv=notrunc status=none; rm -f "$f"; done' sh {} +
  for ((index=${#MOUNTS[@]}-1; index>=0; index--)); do
    if mountpoint -q "${MOUNTS[index]}"; then umount -R "${MOUNTS[index]}" || :; fi
  done
  if findmnt -rn -o TARGET | awk -v root="$ROOT" '$0==root || index($0,root"/")==1 {found=1} END{exit !found}'; then
    CLEANUP_FAIL=1
  else
    find "$ROOT" -xdev -mindepth 1 -delete 2>/dev/null || CLEANUP_FAIL=1
  fi
  host_after="$(capture_host)"
  host_after_digest="$(printf '%s' "$host_after" | sha256sum | awk '{print $1}')"
  if [[ $HOST_BEFORE != "$host_after" || $HOST_BEFORE_DIGEST != "$host_after_digest" || $CLEANUP_FAIL -ne 0 ]]; then
    emit Q20-11 FAIL 0 "$host_after_digest"
    exit 1
  fi
  emit Q20-11 PASS 1 "$host_after_digest"
  (( incoming == 0 )) || exit "$incoming"
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

verify_archive_streaming() {
  local manifest sums member expected actual
  [[ $(sha256sum "$ARCHIVE" | awk '{print $1}') == "$ARCHIVE_SHA" ]] || return 1
  manifest="$(tar -tzf "$ARCHIVE" | sort)"
  [[ $manifest == "$EXPECTED_MANIFEST" ]] || return 1
  [[ $(tar -tvzf "$ARCHIVE" | awk '$1 !~ /^-/{bad=1} END{print bad+0}') == 0 ]] || return 1
  sums="$(tar -xOf "$ARCHIVE" SHA256SUMS)"
  [[ $(printf '%s\n' "$sums" | sed '/^$/d' | wc -l) -eq 6 ]] || return 1
  [[ $sums == "$(<"$OUTER_SUMS")" ]] || return 1
  while read -r expected member; do
    [[ $member != SHA256SUMS && $EXPECTED_MANIFEST == *$'\n'"$member"* ]] || return 1
    actual="$(tar -xOf "$ARCHIVE" "$member" | sha256sum | awk '{print $1}')"
    [[ $actual == "$expected" ]] || return 1
  done <<<"$sums"
}
verify_archive_streaming || die Q20-01
emit Q20-01 PASS 6 "$ARCHIVE_SHA"

mount_tmpfs() {
  mkdir -p "$1"
  mount -t tmpfs -o "$2" tmpfs "$1"
  MOUNTS+=("$1")
}
bind_ro() {
  local target options
  mkdir -p "$2"
  mount --rbind "$1" "$2"
  MOUNTS+=("$2")
  mount --make-rslave "$2"
  while IFS= read -r target; do mount -o remount,bind,ro "$target"; done < <(findmnt -Rrn -o TARGET --mountpoint "$2" | awk '{print length, $0}' | sort -rn | cut -d' ' -f2-)
  while IFS= read -r options; do [[ ,$options, == *,ro,* ]] || return 1; done < <(findmnt -Rrn -o OPTIONS --mountpoint "$2")
}

mount -t tmpfs -o mode=0755,nodev,nosuid tmpfs "$ROOT"
MOUNTS+=("$ROOT")
mkdir -p "$ROOT"/{usr,etc,run,var,tmp,dev,proc,artifact,evidence,input}
bind_ro /usr "$ROOT/usr"
[[ $(findmnt -n -o OPTIONS "$ROOT/usr") == *ro* ]] || die Q20-03
for path in bin sbin lib lib64; do
  if [[ -L /$path ]]; then ln -s "$(readlink /$path)" "$ROOT/$path"; else bind_ro "/$path" "$ROOT/$path"; fi
done
mount_tmpfs "$ROOT/etc" mode=0755,nodev,nosuid
mount_tmpfs "$ROOT/run" mode=0755,nodev,nosuid
mount_tmpfs "$ROOT/usr/local" mode=0755,nodev,nosuid
mkdir -p "$ROOT/usr/local/bin"
mount_tmpfs "$ROOT/var" mode=0755,nodev,nosuid
mount_tmpfs "$ROOT/tmp" mode=1777,nodev,nosuid
mount_tmpfs "$ROOT/dev" mode=0755,nosuid
mount_tmpfs "$ROOT/artifact" mode=0700,nodev,nosuid
mount_tmpfs "$ROOT/evidence" mode=0700,nodev,nosuid,noexec
mount_tmpfs "$ROOT/input" mode=0700,nodev,nosuid,noexec
mknod -m 666 "$ROOT/dev/null" c 1 3
mknod -m 666 "$ROOT/dev/zero" c 1 5
mknod -m 444 "$ROOT/dev/random" c 1 8
mknod -m 444 "$ROOT/dev/urandom" c 1 9
mount -t proc -o ro,nosuid,nodev,noexec proc "$ROOT/proc"
MOUNTS+=("$ROOT/proc")
touch "$ROOT/input/archive" "$ROOT/input/SHA256SUMS"
mount --bind "$ARCHIVE" "$ROOT/input/archive"; MOUNTS+=("$ROOT/input/archive"); mount -o remount,bind,ro "$ROOT/input/archive"
mount --bind "$OUTER_SUMS" "$ROOT/input/SHA256SUMS"; MOUNTS+=("$ROOT/input/SHA256SUMS"); mount -o remount,bind,ro "$ROOT/input/SHA256SUMS"
cp -a /etc/. "$ROOT/etc/"
mkdir -p "$ROOT/run/private" "$ROOT/var/log" "$ROOT/var/lib/sudo" "$ROOT/etc/sudoers.d" "$ROOT/etc/pam.d"
chmod 0700 "$ROOT/run/private" "$ROOT/var/lib/sudo"

# The sole material extraction is here, after every destination is tmpfs.
tar -xzf "$ROOT/input/archive" -C "$ROOT/artifact" --no-same-owner --no-same-permissions
(cd "$ROOT/artifact" && sha256sum -c SHA256SUMS >/dev/null) || die Q20-03
install -o 0 -g 0 -m 0755 "$ROOT/artifact/bin/codex-authority" "$ROOT/usr/local/bin/codex-authority"
install -o 0 -g 0 -m 0755 "$ROOT/artifact/bin/codex-authority-broker" "$ROOT/usr/local/bin/codex-authority-broker"
install -o 0 -g 0 -m 0755 "$ROOT/artifact/bin/codex-authority-sudo" "$ROOT/usr/local/bin/codex-authority-sudo"
emit Q20-03 PASS 1 "$(findmnt -R -n -o TARGET,OPTIONS "$ROOT" | sha256sum | awk '{print $1}')"

python3 - "$ROOT" "$OWNER_UID" "$OTHER_UID" <<'PY'
import pathlib, sys
r=pathlib.Path(sys.argv[1]); owner=sys.argv[2]; other=sys.argv[3]
for name in ('passwd','group'):
    text=(r/'etc'/name).read_text()
    if any(line.split(':')[2] in (owner,other) for line in text.splitlines() if len(line.split(':'))>2): raise SystemExit(1)
with (r/'etc/passwd').open('a') as f:
    f.write(f'codex-fixture:x:{owner}:{owner}::/nonexistent:/usr/sbin/nologin\n')
    f.write(f'codex-distinct:x:{other}:{other}::/nonexistent:/usr/sbin/nologin\n')
with (r/'etc/group').open('a') as f:
    f.write(f'codex-fixture:x:{owner}:codex-distinct\n')
    f.write(f'codex-distinct:x:{other}:\n')
with (r/'etc/shadow').open('a') as f:
    f.write('codex-fixture:!:1:0:99999:7:::\n')
    f.write('codex-distinct:!:1:0:99999:7:::\n')
with (r/'etc/gshadow').open('a') as f:
    f.write('codex-fixture:!::codex-distinct\n')
    f.write('codex-distinct:!::\n')
PY
chown 0:0 "$ROOT/etc/passwd" "$ROOT/etc/group" "$ROOT/etc/shadow" "$ROOT/etc/gshadow"
chmod 0644 "$ROOT/etc/passwd" "$ROOT/etc/group"; chmod 0600 "$ROOT/etc/shadow" "$ROOT/etc/gshadow"

cat >"$ROOT/etc/sudoers" <<'EOF'
Defaults env_reset
root ALL=(ALL:ALL) ALL
@includedir /etc/sudoers.d
EOF
cat >"$ROOT/etc/sudoers.d/codex-authority-command" <<'EOF'
codex-fixture ALL=(root) /usr/bin/true
EOF
cat >"$ROOT/etc/pam.d/sudo" <<'EOF'
#%PAM-1.0
auth required pam_permit.so
account required pam_permit.so
EOF
chown 0:0 "$ROOT/etc/sudoers" "$ROOT/etc/sudoers.d/codex-authority-command" "$ROOT/etc/pam.d/sudo"
chmod 0440 "$ROOT/etc/sudoers" "$ROOT/etc/sudoers.d/codex-authority-command"; chmod 0644 "$ROOT/etc/pam.d/sudo"
chroot "$ROOT" /usr/sbin/visudo -cf /etc/sudoers >/dev/null || die Q20-02
[[ $(stat -c '%a:%u:%g' "$ROOT/usr/bin/sudo") == 4755:0:0 ]]
[[ ,$(findmnt -n -o OPTIONS --mountpoint "$ROOT/usr"), != *,nosuid,* ]]
chroot "$ROOT" /usr/bin/env -i PATH=/usr/sbin:/usr/bin HOME=/nonexistent USER=codex-fixture LOGNAME=codex-fixture \
  /usr/bin/setpriv --reuid="$OWNER_UID" --regid="$OWNER_GID" --clear-groups \
  /bin/sh -c 'test "$(id -u)" -eq 42020 && test "$(awk '\''/^NoNewPrivs:/{print $2}'\'' /proc/self/status)" -eq 0' || die Q20-02
preflight_error="$ROOT/evidence/preflight.stderr"
if ! chroot "$ROOT" /usr/bin/env -i PATH=/usr/sbin:/usr/bin HOME=/nonexistent USER=codex-fixture LOGNAME=codex-fixture \
  /usr/bin/setpriv --reuid="$OWNER_UID" --regid="$OWNER_GID" --clear-groups /usr/bin/sudo -- /usr/bin/true \
  </dev/null >/dev/null 2>"$preflight_error"; then
  reason=9
  grep -Eqi 'password|conversation' "$preflight_error" && reason=1
  grep -Eqi 'no new privileges|effective uid|nosuid|setuid' "$preflight_error" && reason=2
  grep -Eqi 'not allowed|not in the sudoers|policy plugin|sudoers' "$preflight_error" && reason=3
  grep -Eqi 'PAM|account validation|authentication failure' "$preflight_error" && reason=4
  emit Q20-02-diagnostic FAIL "$reason" "$(printf 'preflight-category-%s' "$reason" | sha256sum | awk '{print $1}')"
  exit 1
fi
rm -f "$preflight_error"
rm -f "$ROOT/etc/pam.d/sudo"
emit Q20-02 PASS 1 "$(printf '%s' 'euid0-nnp0-private-mnt-pid-tmpfs-setuid-pam' | sha256sum | awk '{print $1}')"
[[ $MODE == canary ]] || exit 0

install -o 0 -g 0 -m 0644 "$ROOT/artifact/deploy/pam/codex-authority" "$ROOT/etc/pam.d/sudo"
install -o 0 -g 0 -m 0440 "$ROOT/artifact/deploy/sudo/codex-authority" "$ROOT/etc/sudoers.d/codex-authority"
chroot "$ROOT" /usr/sbin/visudo -cf /etc/sudoers >/dev/null || die Q20-04
install -o 0 -g 0 -m 0644 "$ROOT/artifact/deploy/systemd/codex-authority-broker.service" "$ROOT/usr/local/codex-authority-broker.service"
chroot "$ROOT" /usr/bin/systemd-analyze verify /usr/local/codex-authority-broker.service >/dev/null 2>&1 || die Q20-04
emit Q20-04 PASS 3 "$(sha256sum "$ROOT/usr/local/bin/"codex-authority* | sha256sum | awk '{print $1}')"

python3 - "$ROOT" "$OWNER_UID" <<'PY'
import base64,json,os,pathlib,sys
r=pathlib.Path(sys.argv[1]); secret=os.urandom(20)
(r/'run/private/totp.raw').write_bytes(secret)
(r/'etc').joinpath('codex-authority').mkdir(mode=0o700,exist_ok=True)
(r/'etc/codex-authority/seed.json').write_text(json.dumps({'totp_secret_b64':base64.b64encode(secret).decode(),'allowed_uid':int(sys.argv[2])},separators=(',',':')))
os.chmod(r/'run/private/totp.raw',0o600); os.chmod(r/'etc/codex-authority/seed.json',0o600)
PY
chown 0:0 "$ROOT/run/private/totp.raw" "$ROOT/etc/codex-authority/seed.json"

audit_count() { [[ -f $AUDIT ]] && wc -l <"$AUDIT" || printf '0\n'; }
event_check() {
  local line=$1 actor=$2 scope=$3 result=$4 expiry=$5 digest
  jq -e --argjson actor "$actor" --arg scope "$scope" --arg result "$result" --arg expiry "$expiry" '
    (keys|sort)==["actor_uid","correlation_id","lease_expiry","result","scope"] and
    (.correlation_id|type)=="string" and (.correlation_id|test("^[0-9a-f]+$")) and
    .actor_uid==$actor and .scope==$scope and .result==$result and
    (if $expiry=="null" then .lease_expiry==null elif $expiry=="string" then (.lease_expiry|type)=="string" else .lease_expiry==$expiry end)
  ' >/dev/null <<<"$line" || return 1
  digest="$(printf '%s' "$line" | sha256sum | awk '{print $1}')"
  printf '%s\n' "$digest"
}
one_event() {
  local before=$1 actor=$2 scope=$3 result=$4 expiry=$5 after line digest
  after=$(audit_count); (( after == before + 1 )) || return 1
  line=$(sed -n "${after}p" "$AUDIT")
  digest=$(event_check "$line" "$actor" "$scope" "$result" "$expiry") || return 1
  emit Q20-09 PASS 1 "$digest"
}
as_owner() { chroot "$ROOT" /usr/bin/env -i PATH=/usr/sbin:/usr/bin HOME=/nonexistent USER=codex-fixture LOGNAME=codex-fixture /usr/bin/setpriv --reuid="$OWNER_UID" --regid="$OWNER_GID" --clear-groups "$@"; }
as_other() { chroot "$ROOT" /usr/bin/env -i PATH=/usr/sbin:/usr/bin HOME=/nonexistent USER=codex-distinct LOGNAME=codex-distinct /usr/bin/setpriv --reuid="$OTHER_UID" --regid="$OTHER_GID" --groups="$OWNER_GID" "$@"; }
no_helper() {
  local exe
  for exe in "$ROOT"/proc/[0-9]*/exe; do
    [[ -e $exe ]] || continue
    [[ $(readlink "$exe" 2>/dev/null) != /usr/local/bin/codex-authority-sudo ]] || return 1
  done
}
start_broker() {
  local waited=0 mode uid gid
  BROKER_GEN=$((BROKER_GEN+1)); AUDIT="$ROOT/evidence/audit.$BROKER_GEN.jsonl"; : >"$AUDIT"; chmod 0600 "$AUDIT"
  chroot "$ROOT" /usr/local/bin/codex-authority-broker 2>"$AUDIT" & BROKER_PID=$!
  until [[ -S $ROOT/run/codex-authority.sock ]]; do sleep 1; waited=$((waited+1)); (( waited <= 5 )) || return 1; done
  read -r mode uid gid < <(stat -c '%a %u %g' "$ROOT/run/codex-authority.sock")
  [[ $mode == 660 && $uid == "$OWNER_UID" && $gid == "$OWNER_GID" ]] || return 1
  BOOT_FLOOR_CEILING=$(( $(date +%s) / 30 ))
}
stop_broker() { kill "$BROKER_PID"; wait "$BROKER_PID"; BROKER_PID=; [[ ! -e $ROOT/run/codex-authority.sock ]]; }
wait_after_boot_floor() {
  local waited=0 now
  while :; do now=$(( $(date +%s) / 30 )); (( now > BOOT_FLOOR_CEILING )) && return 0; sleep 1; waited=$((waited+1)); (( waited <= 31 )) || return 1; done
}
totp_pipe() {
  python3 - "$ROOT/run/private/totp.raw" <<'PY'
import hashlib,hmac,struct,sys,time
s=open(sys.argv[1],'rb').read(); c=int(time.time())//30
d=hmac.new(s,struct.pack('>Q',c),hashlib.sha1).digest(); o=d[-1]&15
print(f'{(struct.unpack(">I",d[o:o+4])[0]&0x7fffffff)%1000000:06d}')
PY
}
activate() {
  local before expiry_line
  wait_after_boot_floor || return 1
  before=$(audit_count); as_owner /usr/local/bin/codex-authority ready >/dev/null 2>&1; one_event "$before" "$OWNER_UID" ready allow null || return 1
  before=$(audit_count); totp_pipe 2>/dev/null | as_owner /usr/local/bin/codex-authority otp >/dev/null 2>&1; one_event "$before" "$OWNER_UID" otp allow string || return 1
  expiry_line=$(sed -n "$(audit_count)p" "$AUDIT")
  LEASE_EXPIRY=$(jq -er .lease_expiry <<<"$expiry_line")
}
sudo_expect() {
  local expected=$1 before after result expiry
  before=$(audit_count)
  if as_owner /usr/bin/sudo -- /usr/bin/true </dev/null >/dev/null 2>&1; then result=allow; else result=deny; fi
  [[ $result == "$expected" ]] || return 1
  no_helper || return 1
  one_event "$before" "$OWNER_UID" authorize "$result" "$([[ $result == allow ]] && printf '%s' "$LEASE_EXPIRY" || echo null)"
}

start_broker || die Q20-05-broker
# Both wrong peers pass pathname DAC (root bypass; distinct has socket group) but
# fail SO_PEERCRED before Backend.Handle, so the exact oracle is zero audit.
before=$(audit_count); chroot "$ROOT" /usr/local/bin/codex-authority ready >/dev/null 2>&1 && die Q20-05-root-peer; [[ $(audit_count) == "$before" ]] || die Q20-05-root-audit
before=$(audit_count); as_other /usr/local/bin/codex-authority ready >/dev/null 2>&1 && die Q20-05-distinct-peer; [[ $(audit_count) == "$before" ]] || die Q20-05-distinct-audit
before=$(audit_count)
as_other /usr/bin/python3 -c 'import json,socket,struct,sys; b=json.dumps({"version":1,"operation":"ready","actor_uid":42020},separators=(",",":")).encode(); s=socket.socket(socket.AF_UNIX); s.connect("/run/codex-authority.sock");
try:
 s.sendall(struct.pack(">I",len(b))+b); rejected=s.recv(1)==b""
except (BrokenPipeError,ConnectionResetError): rejected=True
sys.exit(0 if rejected else 1)' >/dev/null 2>&1 || die Q20-05-claimed-peer
[[ $(audit_count) == "$before" ]] || die Q20-05-claimed-audit
emit Q20-05 PASS 3 "$(printf '%s' 'root-distinct-claimed-peer-zero-audit' | sha256sum | awk '{print $1}')"

activate || die Q20-05-activate
sudo_expect allow || die Q20-06
sudo_expect allow || die Q20-06
emit Q20-06 PASS 2 "$(sha256sum "$AUDIT" | awk '{print $1}')"

expiry_epoch=$(date -u -d "$LEASE_EXPIRY" +%s); waited=0
while (( $(date +%s) <= expiry_epoch )); do sleep 1; waited=$((waited+1)); (( waited <= 305 )) || die Q20-07; done
sudo_expect deny || die Q20-07
emit Q20-07 PASS 1 "$(sha256sum "$AUDIT" | awk '{print $1}')"

activate || die Q20-08
sudo_expect allow || die Q20-08
old_count=$(audit_count); stop_broker || die Q20-08
if as_owner /usr/bin/sudo -- /usr/bin/true </dev/null >/dev/null 2>&1; then die Q20-08; fi
no_helper || die Q20-08
[[ $(wc -l <"$ROOT/evidence/audit.$BROKER_GEN.jsonl") == "$old_count" ]] || die Q20-08
emit Q20-08 PASS 1 "$(printf '%s' 'broker-unavailable-zero-audit' | sha256sum | awk '{print $1}')"
start_broker || die Q20-08
sudo_expect deny || die Q20-08
activate || die Q20-08
sudo_expect allow || die Q20-08
stop_broker || die Q20-08
emit Q20-08 PASS 4 "$(sha256sum "$ROOT/evidence/"audit.*.jsonl | sha256sum | awk '{print $1}')"

# Protected raw streams never leave tmpfs. Retained output above is bounded.
if grep -R -E -q 'totp_secret_b64|"code"|request denied|internal error|LEASE' "$ROOT/evidence"; then die Q20-10; fi
emit Q20-10 PASS 1 "$(find "$ROOT/evidence" -type f -printf '%f\n' | sort | sha256sum | awk '{print $1}')"
exit 0
