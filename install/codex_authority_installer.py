#!/usr/bin/python3
"""Transactional Ubuntu 24.04 installer for Codex Authority Broker.

The public entrypoints select a fixed operation.  Production never accepts a
caller-selected root, destination, identity, service, or executable path.
"""

from __future__ import annotations

import base64
import grp
import hashlib
import json
import os
import pwd
import re
import secrets
import shutil
import signal
import stat
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Callable, Iterable

IDENTITY = "coding-agent"
IDENTITY_HOME = Path("/var/lib/coding-agent")
STATE_VERSION = 1
ROOT = Path("/")
STATE_DIR = Path("/var/lib/codex-authority-installer")
STATE_PATH = STATE_DIR / "state.json"
RECOVERY_DIR = Path("/var/lib/codex-authority-recovery")
RECOVERY_PATH = RECOVERY_DIR / "recover"
RECOVERY_CORE = RECOVERY_DIR / "codex_authority_installer.py"
RECOVERY_SUMS = RECOVERY_DIR / "SHA256SUMS"
RECOVERY_STATE = RECOVERY_DIR / "state.json"
UNINSTALL_BACKUP_DIR = RECOVERY_DIR / "uninstall-backup"
COMPLETED_RECOVERY_DIR = Path("/var/lib/codex-authority-recovery.completed")
SEED_PATH = Path("/etc/codex-authority/seed.json")
SOCKET_PATH = Path("/run/codex-authority.sock")
FIXTURE_MARKER = Path("/run/codex-authority-e2e-fixture")
UNIT = "codex-authority-broker.service"
UNIT_PATHS = (
    Path("/etc/systemd/system") / UNIT,
    Path("/run/systemd/system") / UNIT,
    Path("/usr/lib/systemd/system") / UNIT,
    Path("/lib/systemd/system") / UNIT,
)
UNIT_DROPINS = tuple(Path(f"{path}.d") for path in UNIT_PATHS)

STATE_KEYS = {
    "version",
    "operation",
    "phase",
    "artifact_digest",
    "identity",
    "pre_state",
    "backup_paths",
    "owned_paths",
    "pending_path",
    "completed_steps",
    "seed_generation",
}

SOURCE_MEMBERS = (
    "SHA256SUMS",
    "bin/codex-authority",
    "bin/codex-authority-broker",
    "bin/codex-authority-sudo",
    "deploy/pam/codex-authority",
    "deploy/sudo/codex-authority",
    "deploy/systemd/codex-authority-broker.service",
    "install/codex-authority-install",
    "install/codex-authority-verify",
    "install/codex-authority-admin",
    "install/codex-authority-recover",
    "install/codex-authority-uninstall",
    "install/codex_authority_installer.py",
    "docs/ADMIN_MANUAL.md",
    "docs/USER_MANUAL.md",
)

INSTALL_MAP = (
    ("bin/codex-authority", "/usr/local/bin/codex-authority", 0o755),
    ("bin/codex-authority-broker", "/usr/local/bin/codex-authority-broker", 0o755),
    ("bin/codex-authority-sudo", "/usr/local/bin/codex-authority-sudo", 0o755),
    ("deploy/pam/codex-authority", "/etc/pam.d/codex-authority", 0o644),
    ("deploy/sudo/codex-authority", "/etc/sudoers.d/codex-authority", 0o440),
    (
        "deploy/systemd/codex-authority-broker.service",
        "/etc/systemd/system/codex-authority-broker.service",
        0o644,
    ),
    ("install/codex-authority-admin", "/usr/local/sbin/codex-authority-admin", 0o755),
    ("install/codex-authority-verify", "/usr/local/sbin/codex-authority-verify", 0o755),
    ("install/codex-authority-recover", "/usr/local/sbin/codex-authority-recover", 0o755),
    ("install/codex-authority-uninstall", "/usr/local/sbin/codex-authority-uninstall", 0o755),
    (
        "install/codex_authority_installer.py",
        "/usr/local/lib/codex-authority/codex_authority_installer.py",
        0o644,
    ),
)

OWNED_DIRECTORIES = (
    Path("/etc/codex-authority"),
    Path("/usr/local/lib/codex-authority"),
)

ALLOWED_MANAGED_PATHS = {
    *(destination for _, destination, _ in INSTALL_MAP),
    str(SEED_PATH),
    str(RECOVERY_PATH),
    str(RECOVERY_CORE),
    str(RECOVERY_SUMS),
    *(str(path) for path in OWNED_DIRECTORIES),
}

ALLOWED_PHASES = {
    "PREPARED", "IDENTITY_INTENT", "IDENTITY_DONE", "FILE_INTENT", "FILE_DONE",
    "SEED_INTENT", "SEED_DONE", "POLICY_INTENT", "POLICY_DONE", "START_INTENT",
    "START_DONE", "ROTATE_PREPARED", "BACKUP_INTENT", "BACKUP_DONE", "BROKER_STOPPED",
    "NEW_SEED_INTENT", "NEW_SEED_PENDING_ACK", "ENROLLMENT_ACKED", "RESTART_INTENT",
    "RESTART_DONE", "UNINSTALL_PREPARED", "STOP_INTENT", "STOP_DONE", "REMOVE_INTENT",
    "REMOVE_DONE", "TIMESTAMP_INTENT", "TIMESTAMP_DONE", "RELOAD_INTENT", "RELOAD_DONE",
    "VERIFY_DONE", "COMMITTED",
}

OPERATION_PHASES = {
    "install": {
        "PREPARED", "IDENTITY_INTENT", "IDENTITY_DONE", "FILE_INTENT", "FILE_DONE",
        "SEED_INTENT", "SEED_DONE", "POLICY_INTENT", "POLICY_DONE", "START_INTENT",
        "START_DONE", "COMMITTED",
    },
    "rotate": {
        "ROTATE_PREPARED", "BACKUP_INTENT", "BACKUP_DONE", "BROKER_STOPPED",
        "NEW_SEED_INTENT", "NEW_SEED_PENDING_ACK", "ENROLLMENT_ACKED", "RESTART_INTENT",
        "RESTART_DONE",
    },
    "uninstall": {
        "UNINSTALL_PREPARED", "STOP_INTENT", "STOP_DONE", "REMOVE_INTENT", "REMOVE_DONE",
        "TIMESTAMP_INTENT", "TIMESTAMP_DONE", "IDENTITY_INTENT", "IDENTITY_DONE",
        "RELOAD_INTENT", "RELOAD_DONE", "VERIFY_DONE", "COMMITTED",
    },
}


class InstallError(Exception):
    """A bounded, non-secret installation failure."""


def _pairs(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if key in result:
            raise InstallError("invalid state")
        result[key] = value
    return result


def strict_json(data: bytes) -> dict[str, object]:
    try:
        value = json.loads(data, object_pairs_hook=_pairs)
    except (ValueError, TypeError, InstallError) as exc:
        raise InstallError("invalid state") from exc
    if not isinstance(value, dict):
        raise InstallError("invalid state")
    return value


def _fsync_dir(path: Path) -> None:
    descriptor = os.open(path, os.O_RDONLY | os.O_DIRECTORY | os.O_CLOEXEC)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)


def ensure_root_directory(path: Path, mode: int = 0o700) -> None:
    missing: list[Path] = []
    current = path
    while not current.exists():
        if current.is_symlink():
            raise InstallError("unsafe directory")
        missing.append(current)
        current = current.parent
    for directory in reversed(missing):
        directory.mkdir(mode=mode)
        os.chown(directory, 0, 0)
        os.chmod(directory, mode)
        _fsync_dir(directory.parent)


def atomic_write(path: Path, data: bytes, mode: int) -> None:
    parent_created = not path.parent.exists()
    ensure_root_directory(path.parent)
    os.chmod(path.parent, 0o700)
    os.chown(path.parent, 0, 0)
    if parent_created:
        _fsync_dir(path.parent.parent)
    descriptor, temporary = tempfile.mkstemp(prefix=".cab-", dir=path.parent)
    try:
        os.fchmod(descriptor, mode)
        os.fchown(descriptor, 0, 0)
        with os.fdopen(descriptor, "wb", closefd=False) as stream:
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
        os.close(descriptor)
        descriptor = -1
        os.replace(temporary, path)
        _fsync_dir(path.parent)
    finally:
        if descriptor >= 0:
            os.close(descriptor)
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass


def state_document(operation: str, digest: str, identity: dict[str, object]) -> dict[str, object]:
    return {
        "version": STATE_VERSION,
        "operation": operation,
        "phase": "PREPARED" if operation == "install" else f"{operation.upper()}_PREPARED",
        "artifact_digest": digest,
        "identity": identity,
        "pre_state": {
            "identity_mode": identity["mode"],
            "managed_paths_absent": True,
            "installed_digests": {},
            "sudo_digests": {"/etc/sudoers": "0" * 64},
            "pam_sudo_digest": "0" * 64,
            "unit_absent": True,
        },
        "backup_paths": [],
        "owned_paths": [],
        "pending_path": None,
        "completed_steps": [],
        "seed_generation": 0,
    }


def validate_state(value: dict[str, object]) -> dict[str, object]:
    if set(value) != STATE_KEYS or value.get("version") != STATE_VERSION:
        raise InstallError("invalid state")
    if value.get("operation") not in {"install", "rotate", "uninstall"}:
        raise InstallError("invalid state")
    operation = value.get("operation")
    phase = value.get("phase")
    if phase not in ALLOWED_PHASES or phase not in OPERATION_PHASES[operation]:
        raise InstallError("invalid state")
    if not re.fullmatch(r"[0-9a-f]{64}", str(value.get("artifact_digest", ""))):
        raise InstallError("invalid state")
    for name in ("identity", "pre_state"):
        if not isinstance(value.get(name), dict):
            raise InstallError("invalid state")
    for name in ("backup_paths", "owned_paths", "completed_steps"):
        items = value.get(name)
        if not isinstance(items, list) or any(not isinstance(item, str) for item in items):
            raise InstallError("invalid state")
    identity = value["identity"]
    assert isinstance(identity, dict)
    if set(identity) != {"mode", "uid", "gid"} or identity.get("mode") not in {"create", "creating", "created", "reuse"}:
        raise InstallError("invalid state")
    if not isinstance(identity.get("uid"), int) or not isinstance(identity.get("gid"), int):
        raise InstallError("invalid state")
    if identity["mode"] == "create":
        if identity["uid"] != 0 or identity["gid"] != 0:
            raise InstallError("invalid state")
    elif identity["uid"] <= 0 or identity["uid"] != identity["gid"]:
        raise InstallError("invalid state")
    pre_state = value["pre_state"]
    assert isinstance(pre_state, dict)
    required_pre_state = {
        "identity_mode", "managed_paths_absent", "installed_digests", "sudo_digests",
        "pam_sudo_digest",
        "unit_absent",
    }
    if not required_pre_state <= set(pre_state) or not set(pre_state) <= required_pre_state | {"seed_backup_digest", "uninstall_backups"}:
        raise InstallError("invalid state")
    if pre_state.get("identity_mode") not in {"create", "reuse"} or pre_state.get("managed_paths_absent") is not True:
        raise InstallError("invalid state")
    if pre_state.get("unit_absent") is not True:
        raise InstallError("invalid state")
    if (identity["mode"] == "reuse") != (pre_state["identity_mode"] == "reuse"):
        raise InstallError("invalid state")
    digests = pre_state.get("installed_digests", {})
    if not isinstance(digests, dict) or not set(digests) <= {destination for _, destination, _ in INSTALL_MAP}:
        raise InstallError("invalid state")
    if any(not isinstance(item, str) or not re.fullmatch(r"[0-9a-f]{64}", item) for item in digests.values()):
        raise InstallError("invalid state")
    sudo_digests = pre_state.get("sudo_digests")
    if not isinstance(sudo_digests, dict):
        raise InstallError("invalid state")
    for name, digest in sudo_digests.items():
        candidate = Path(name)
        if (
            not isinstance(name, str)
            or not candidate.is_absolute()
            or ".." in candidate.parts
            or candidate != Path("/etc/sudoers") and Path("/etc") not in candidate.parents
            or not isinstance(digest, str)
            or not re.fullmatch(r"[0-9a-f]{64}", digest)
        ):
            raise InstallError("invalid state")
    if "/etc/sudoers" not in sudo_digests:
        raise InstallError("invalid state")
    pam_digest = pre_state.get("pam_sudo_digest")
    if not isinstance(pam_digest, str) or not re.fullmatch(r"[0-9a-f]{64}", pam_digest):
        raise InstallError("invalid state")
    uninstall_backups = pre_state.get("uninstall_backups")
    if uninstall_backups is not None:
        if operation != "uninstall" or not isinstance(uninstall_backups, dict):
            raise InstallError("invalid state")
        backup_sources = {destination for _, destination, _ in INSTALL_MAP} | {str(SEED_PATH)}
        for source, record in uninstall_backups.items():
            expected_backup = UNINSTALL_BACKUP_DIR / source.lstrip("/") if isinstance(source, str) else None
            if source not in backup_sources or not isinstance(record, dict) or set(record) != {"path", "digest", "mode"}:
                raise InstallError("invalid state")
            if record.get("path") != str(expected_backup) or not re.fullmatch(r"[0-9a-f]{64}", str(record.get("digest", ""))):
                raise InstallError("invalid state")
            if record.get("mode") not in {0o440, 0o600, 0o644, 0o755}:
                raise InstallError("invalid state")
    backup_paths = value["backup_paths"]
    owned_paths = value["owned_paths"]
    assert isinstance(backup_paths, list) and isinstance(owned_paths, list)
    if len(set(owned_paths)) != len(owned_paths) or not set(owned_paths) <= ALLOWED_MANAGED_PATHS:
        raise InstallError("invalid state")
    recovery_owned = [str(RECOVERY_PATH), str(RECOVERY_CORE), str(RECOVERY_SUMS)]
    installed_owned = recovery_owned + [str(path) for path in OWNED_DIRECTORIES] + [
        destination for _, destination, _ in INSTALL_MAP
    ]
    complete_owned = installed_owned + [str(SEED_PATH)]
    if operation == "install":
        if phase == "PREPARED":
            valid_owned = owned_paths in ([], recovery_owned)
        elif phase in {"IDENTITY_INTENT", "IDENTITY_DONE"}:
            valid_owned = owned_paths == recovery_owned
        elif phase in {"FILE_INTENT", "FILE_DONE"}:
            minimum = len(recovery_owned) + (1 if phase == "FILE_DONE" else 0)
            valid_owned = owned_paths == installed_owned[:len(owned_paths)] and len(owned_paths) >= minimum
        elif phase == "SEED_INTENT":
            valid_owned = owned_paths == installed_owned
        else:
            valid_owned = owned_paths == complete_owned
    else:
        valid_owned = owned_paths == complete_owned
    if not valid_owned:
        raise InstallError("invalid state")
    if operation == "install" and pre_state["identity_mode"] == "create":
        expected_identity_modes = {
            "PREPARED": {"create"},
            "IDENTITY_INTENT": {"creating"},
        }.get(str(phase), {"created"})
        if identity["mode"] not in expected_identity_modes:
            raise InstallError("invalid state")
    elif operation == "install" and phase in {"IDENTITY_INTENT", "IDENTITY_DONE"}:
        raise InstallError("invalid state")
    expected_installed_digests = {
        destination for _, destination, _ in INSTALL_MAP if destination in owned_paths
    }
    if set(digests) != expected_installed_digests:
        raise InstallError("invalid state")
    if backup_paths not in ([], [str(STATE_DIR / "backups/seed.previous")]):
        raise InstallError("invalid state")
    if operation != "rotate" and backup_paths:
        raise InstallError("invalid state")
    if operation == "rotate" and backup_paths != [str(STATE_DIR / "backups/seed.previous")]:
        raise InstallError("invalid state")
    backup_digest = pre_state.get("seed_backup_digest")
    if backup_digest is not None and not re.fullmatch(r"[0-9a-f]{64}", str(backup_digest)):
        raise InstallError("invalid state")
    if operation == "rotate" and phase not in {"ROTATE_PREPARED", "BACKUP_INTENT"} and backup_digest is None:
        raise InstallError("invalid state")
    if operation != "rotate" and backup_digest is not None:
        raise InstallError("invalid state")
    pending = value.get("pending_path")
    if pending is not None and pending not in ALLOWED_MANAGED_PATHS:
        raise InstallError("invalid state")
    if (phase in {"FILE_INTENT", "REMOVE_INTENT", "SEED_INTENT"}) != (pending is not None):
        raise InstallError("invalid state")
    if phase == "SEED_INTENT" and pending != str(SEED_PATH):
        raise InstallError("invalid state")
    if phase == "REMOVE_INTENT" and pending not in owned_paths:
        raise InstallError("invalid state")
    if phase == "FILE_INTENT":
        next_index = len(owned_paths)
        if next_index >= len(installed_owned) or pending != installed_owned[next_index]:
            raise InstallError("invalid state")
    complete_digest_phases = operation != "install" or phase in {
        "SEED_INTENT", "SEED_DONE", "POLICY_INTENT", "POLICY_DONE", "START_INTENT",
        "START_DONE", "COMMITTED",
    }
    if complete_digest_phases and set(digests) != {destination for _, destination, _ in INSTALL_MAP}:
        raise InstallError("invalid state")
    if operation == "uninstall" and phase != "UNINSTALL_PREPARED" and set(uninstall_backups or {}) != ({destination for _, destination, _ in INSTALL_MAP} | {str(SEED_PATH)}):
        raise InstallError("invalid state")
    if value.get("completed_steps") != []:
        raise InstallError("invalid state")
    if not isinstance(value.get("seed_generation"), int):
        raise InstallError("invalid state")
    return value


def load_state(path: Path = STATE_PATH) -> dict[str, object]:
    descriptor = os.open(path, os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW)
    try:
        metadata = os.fstat(descriptor)
        if (
            not stat.S_ISREG(metadata.st_mode)
            or metadata.st_uid != 0
            or stat.S_IMODE(metadata.st_mode) != 0o600
            or metadata.st_size > 131072
        ):
            raise InstallError("invalid state")
        data = b""
        while len(data) <= 131072:
            block = os.read(descriptor, min(131073 - len(data), 16384))
            if not block:
                break
            data += block
        if len(data) > 131072:
            raise InstallError("invalid state")
    finally:
        os.close(descriptor)
    return validate_state(strict_json(data))


def save_state(value: dict[str, object], path: Path = STATE_PATH) -> None:
    validate_state(value)
    data = json.dumps(value, sort_keys=True, separators=(",", ":")).encode() + b"\n"
    atomic_write(path, data, 0o600)


def set_phase(value: dict[str, object], phase: str, pending: str | None = None, path: Path = STATE_PATH) -> None:
    value["phase"] = phase
    value["pending_path"] = pending
    save_state(value, path)
    fixture_checkpoint(phase)


def fixture_checkpoint(phase: str) -> None:
    try:
        metadata = FIXTURE_MARKER.lstat()
    except FileNotFoundError:
        return
    if not stat.S_ISREG(metadata.st_mode) or metadata.st_uid != 0 or stat.S_IMODE(metadata.st_mode) != 0o600:
        raise InstallError("invalid fixture marker")
    instruction = FIXTURE_MARKER.read_text(encoding="ascii").strip()
    match = re.fullmatch(r"(fail|kill) ([A-Z_]+)", instruction)
    if not match or match.group(2) != phase:
        return
    FIXTURE_MARKER.unlink()
    _fsync_dir(FIXTURE_MARKER.parent)
    if match.group(1) == "fail":
        raise InstallError("injected failure")
    os.kill(os.getpid(), 9)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(131072), b""):
            digest.update(block)
    return digest.hexdigest()


def require_root_file_digest(path: Path, expected: str, mode: int) -> None:
    metadata = path.lstat()
    if (
        not stat.S_ISREG(metadata.st_mode)
        or metadata.st_uid != 0
        or stat.S_IMODE(metadata.st_mode) != mode
        or sha256(path) != expected
    ):
        raise InstallError("recovery data invalid")


def source_root() -> Path:
    return Path(__file__).resolve().parent.parent


def parse_sums(root: Path, required_uid: int | None = 0) -> dict[str, str]:
    actual_files: set[str] = set()
    for path in root.rglob("*"):
        metadata = path.lstat()
        if stat.S_ISREG(metadata.st_mode):
            actual_files.add(path.relative_to(root).as_posix())
        elif not stat.S_ISDIR(metadata.st_mode):
            raise InstallError("invalid artifact")
    if actual_files != set(SOURCE_MEMBERS):
        raise InstallError("invalid artifact")
    lines = (root / "SHA256SUMS").read_text(encoding="ascii").splitlines()
    if lines != sorted(lines):
        raise InstallError("invalid artifact")
    result: dict[str, str] = {}
    for line in lines:
        match = re.fullmatch(r"([0-9a-f]{64})  ([A-Za-z0-9_./-]+)", line)
        if not match or match.group(2) in result:
            raise InstallError("invalid artifact")
        result[match.group(2)] = match.group(1)
    expected = set(SOURCE_MEMBERS) - {"SHA256SUMS"}
    if set(result) != expected:
        raise InstallError("invalid artifact")
    for name, expected_digest in result.items():
        path = root / name
        metadata = path.lstat()
        if not stat.S_ISREG(metadata.st_mode) or (required_uid is not None and metadata.st_uid != required_uid) or metadata.st_mode & 0o022:
            raise InstallError("invalid artifact")
        if sha256(path) != expected_digest:
            raise InstallError("invalid artifact")
    return result


def artifact_digest(root: Path) -> str:
    parse_sums(root)
    return sha256(root / "SHA256SUMS")


@dataclass
class Runner:
    call: Callable[..., subprocess.CompletedProcess[str]] = subprocess.run

    def run(self, argv: Iterable[str], *, check: bool = True, input_text: str | None = None) -> subprocess.CompletedProcess[str]:
        return self.call(
            list(argv),
            check=check,
            text=True,
            input=input_text,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env={"PATH": "/usr/sbin:/usr/bin:/sbin:/bin", "HOME": "/root", "LANG": "C"},
        )


def require_root() -> None:
    if os.geteuid() != 0:
        raise InstallError("root required")


def require_platform() -> None:
    values: dict[str, str] = {}
    for line in Path("/etc/os-release").read_text(encoding="utf-8").splitlines():
        if "=" in line:
            key, value = line.split("=", 1)
            values[key] = value.strip('"')
    if values.get("ID") != "ubuntu" or values.get("VERSION_ID") != "24.04":
        raise InstallError("unsupported platform")
    if os.uname().machine != "x86_64":
        raise InstallError("unsupported platform")
    for path in ("/usr/sbin/visudo", "/usr/sbin/getcap", "/usr/bin/systemctl", "/usr/bin/systemd-analyze", "/usr/bin/qrencode"):
        if not Path(path).is_file():
            raise InstallError("missing prerequisite")


def _secure_policy_file(path: Path) -> None:
    metadata = path.lstat()
    if (
        not stat.S_ISREG(metadata.st_mode)
        or metadata.st_uid != 0
        or metadata.st_mode & 0o022
        or path.is_symlink()
    ):
        raise InstallError("incompatible sudo policy")


def sudo_policy_sources() -> list[Path]:
    pending = [Path("/etc/sudoers")]
    sources: list[Path] = []
    seen: set[Path] = set()
    directive = re.compile(r"^\s*(?:#|@)(include|includedir)\s+(.+?)\s*$")
    while pending:
        path = pending.pop(0)
        path = Path(os.path.abspath(path))
        if path in seen:
            continue
        if path != Path("/etc/sudoers") and Path("/etc") not in path.parents:
            raise InstallError("incompatible sudo policy")
        _secure_policy_file(path)
        seen.add(path)
        sources.append(path)
        text = path.read_text(encoding="utf-8", errors="strict")
        for line in text.splitlines():
            match = directive.match(line)
            if not match:
                continue
            target = Path(match.group(2).strip().strip('"'))
            if not target.is_absolute():
                target = path.parent / target
            if match.group(1) == "include":
                pending.append(target)
                continue
            metadata = target.lstat()
            if not stat.S_ISDIR(metadata.st_mode) or metadata.st_uid != 0 or metadata.st_mode & 0o022 or target.is_symlink():
                raise InstallError("incompatible sudo policy")
            pending.extend(sorted(
                item for item in target.iterdir()
                if "." not in item.name and not item.name.endswith("~")
            ))
    return sources


def sudo_policy_text() -> str:
    return "\n".join(path.read_text(encoding="utf-8", errors="strict") for path in sudo_policy_sources())


def sudo_policy_snapshot() -> dict[str, str]:
    return {str(path): sha256(path) for path in sudo_policy_sources()}


def require_safe_sudo_baseline() -> None:
    sudo_conf = Path("/etc/sudo.conf").read_text(encoding="utf-8", errors="strict")
    if any(line.strip() and not line.lstrip().startswith("#") for line in sudo_conf.splitlines()):
        raise InstallError("incompatible sudo policy")
    nsswitch = Path("/etc/nsswitch.conf").read_text(encoding="utf-8", errors="strict")
    for line in nsswitch.splitlines():
        match = re.match(r"^\s*sudoers\s*:\s*(.*?)\s*(?:#.*)?$", line)
        if match and match.group(1).split() != ["files"]:
            raise InstallError("incompatible sudo policy")
    policy = sudo_policy_text()
    lowered = policy.lower()
    forbidden = ("coding-agent", "exempt_group", "pam_service=", "pam_login_service=", "pam_askpass_service=")
    if any(token in lowered for token in forbidden) or re.search(r"(?mi)^\s*Defaults\S*\s+.*!authenticate", policy):
        raise InstallError("incompatible sudo policy")


def require_no_existing_sudo_grant(runner: Runner) -> None:
    result = runner.run(("/usr/bin/sudo", "-ll", "-U", IDENTITY), check=False)
    combined = f"{result.stdout}\n{result.stderr}".lower()
    if result.returncode == 0 and ("sudoers entry:" in combined or re.search(r"(?m)^\s*commands:\s*$", combined)):
        raise InstallError("incompatible sudo policy")


def require_clean_managed_paths(identity: dict[str, object]) -> None:
    paths = [Path(destination) for _, destination, _ in INSTALL_MAP]
    paths.extend((SEED_PATH, STATE_DIR, RECOVERY_DIR, COMPLETED_RECOVERY_DIR))
    paths.extend(OWNED_DIRECTORIES)
    if identity.get("mode") == "create":
        paths.append(IDENTITY_HOME)
    if any(path.exists() or path.is_symlink() for path in paths):
        raise InstallError("managed path collision")
    if any(path.exists() or path.is_symlink() for path in UNIT_PATHS + UNIT_DROPINS):
        raise InstallError("unit collision")
    parents = {path.parent for path in paths}
    for parent in parents:
        current = parent
        while True:
            if not current.exists():
                if current.is_symlink():
                    raise InstallError("unsafe installation parent")
                current = current.parent
                continue
            metadata = current.lstat()
            if not stat.S_ISDIR(metadata.st_mode) or current.is_symlink() or metadata.st_uid != 0 or metadata.st_mode & 0o022:
                raise InstallError("unsafe installation parent")
            if current == Path("/"):
                break
            current = current.parent


def remove_empty_completed_recovery() -> None:
    try:
        metadata = COMPLETED_RECOVERY_DIR.lstat()
    except FileNotFoundError:
        return
    if (
        not stat.S_ISDIR(metadata.st_mode)
        or metadata.st_uid != 0
        or stat.S_IMODE(metadata.st_mode) != 0o700
        or COMPLETED_RECOVERY_DIR.is_symlink()
        or any(COMPLETED_RECOVERY_DIR.iterdir())
    ):
        raise InstallError("recovery cleanup required")
    COMPLETED_RECOVERY_DIR.rmdir()
    _fsync_dir(COMPLETED_RECOVERY_DIR.parent)


def require_unit_absent(runner: Runner) -> None:
    shown = runner.run(("/usr/bin/systemctl", "show", UNIT, "--property=LoadState", "--value"), check=False)
    files = runner.run(("/usr/bin/systemctl", "list-unit-files", UNIT, "--no-legend", "--no-pager"), check=False)
    active = runner.run(("/usr/bin/systemctl", "is-active", UNIT), check=False)
    if (
        shown.returncode != 0
        or shown.stdout.strip() != "not-found"
        or files.returncode != 0
        or files.stdout.strip()
        or active.returncode not in {3, 4}
    ):
        raise InstallError("unit collision")


def remove_managed_path(path: Path) -> None:
    try:
        metadata = path.lstat()
    except FileNotFoundError:
        return
    if stat.S_ISDIR(metadata.st_mode):
        path.rmdir()
    else:
        path.unlink()


def inspect_identity(runner: Runner) -> dict[str, object]:
    try:
        account = pwd.getpwnam(IDENTITY)
    except KeyError:
        try:
            grp.getgrnam(IDENTITY)
        except KeyError:
            return {"mode": "create", "uid": 0, "gid": 0}
        raise InstallError("identity collision")
    try:
        group = grp.getgrnam(IDENTITY)
    except KeyError as exc:
        raise InstallError("identity collision") from exc
    groups = os.getgrouplist(IDENTITY, account.pw_gid)
    status = runner.run(("/usr/bin/passwd", "-S", IDENTITY)).stdout.split()
    if (
        account.pw_uid == 0
        or account.pw_uid != account.pw_gid
        or group.gr_gid != account.pw_gid
        or groups != [account.pw_gid]
        or account.pw_dir != str(IDENTITY_HOME)
        or account.pw_shell != "/bin/bash"
        or len(status) < 2
        or status[1] not in {"L", "LK"}
    ):
        raise InstallError("incompatible identity")
    capability_result = runner.run(("/usr/sbin/getcap", "-r", account.pw_dir), check=False)
    capabilities = capability_result.stdout.strip()
    timestamps = (Path("/run/sudo/ts") / IDENTITY, Path("/run/sudo/ts") / str(account.pw_uid))
    processes = runner.run(("/usr/bin/pgrep", "-u", str(account.pw_uid)), check=False)
    sessions = runner.run(("/usr/bin/loginctl", "list-sessions", "--no-legend"), check=False)
    if (
        capability_result.returncode != 0
        or capabilities
        or any(path.exists() for path in timestamps)
        or processes.returncode not in {0, 1}
        or processes.returncode == 0 and processes.stdout.strip()
        or sessions.returncode != 0
        or any(IDENTITY in line or re.search(rf"\b{account.pw_uid}\b", line) for line in sessions.stdout.splitlines())
        or first_owned_path(runner, account.pw_uid, exclude_home=True)
    ):
        raise InstallError("incompatible identity")
    return {"mode": "reuse", "uid": account.pw_uid, "gid": account.pw_gid}


def local_mount_targets(runner: Runner) -> list[str]:
    result = runner.run(("/usr/bin/findmnt", "-rn", "-o", "TARGET"), check=False)
    targets = [line.strip() for line in result.stdout.splitlines() if line.startswith("/")]
    if result.returncode != 0 or "/" not in targets:
        raise InstallError("filesystem ownership scan failed")
    return targets


def first_owned_path(runner: Runner, uid: int, exclude_home: bool = False) -> str:
    for target in local_mount_targets(runner):
        command = [
            "/usr/bin/find", target, "-xdev", "(", "-uid", str(uid), "-o", "-gid", str(uid), ")",
        ]
        if exclude_home:
            command.extend(("!", "-path", str(IDENTITY_HOME), "!", "-path", f"{IDENTITY_HOME}/*"))
        command.extend(("-print", "-quit"))
        result = runner.run(command, check=False)
        if result.returncode != 0:
            raise InstallError("filesystem ownership scan failed")
        if result.stdout.strip():
            return result.stdout.strip()
    return ""


def allocate_identity(runner: Runner) -> dict[str, object]:
    used_uids = {entry.pw_uid for entry in pwd.getpwall()}
    used_gids = {entry.gr_gid for entry in grp.getgrall()}
    numeric_id = None
    for value in range(999, 99, -1):
        if value in used_uids or value in used_gids:
            continue
        if not first_owned_path(runner, value):
            numeric_id = value
            break
    if numeric_id is None:
        raise InstallError("identity allocation failed")
    return {"mode": "creating", "uid": numeric_id, "gid": numeric_id}


def create_identity(runner: Runner, identity: dict[str, object] | None = None) -> dict[str, object]:
    identity = allocate_identity(runner) if identity is None else identity
    numeric_id = int(identity["uid"])
    runner.run(("/usr/sbin/groupadd", "--system", "--gid", str(numeric_id), IDENTITY))
    fixture_checkpoint("IDENTITY_GROUP_MUTATED")
    runner.run((
        "/usr/sbin/useradd", "--system", "--uid", str(numeric_id), "--gid", str(numeric_id), "--home-dir",
        str(IDENTITY_HOME), "--no-create-home", "--shell", "/bin/bash", IDENTITY,
    ))
    fixture_checkpoint("IDENTITY_USER_MUTATED")
    runner.run(("/usr/bin/passwd", "-l", IDENTITY))
    account = pwd.getpwnam(IDENTITY)
    if account.pw_uid == 0 or account.pw_uid != account.pw_gid:
        raise InstallError("identity creation failed")
    IDENTITY_HOME.mkdir(mode=0o700)
    os.chown(IDENTITY_HOME, account.pw_uid, account.pw_gid)
    _fsync_dir(IDENTITY_HOME.parent)
    fixture_checkpoint("IDENTITY_HOME_MUTATED")
    return {"mode": "created", "uid": account.pw_uid, "gid": account.pw_gid}


def remove_identity(runner: Runner, identity: dict[str, object]) -> None:
    if identity.get("mode") not in {"created", "creating", "create"}:
        return
    home = IDENTITY_HOME
    if home.exists() and any(home.iterdir()):
        raise InstallError("identity home is not empty")
    uid = int(identity["uid"])
    try:
        account = pwd.getpwnam(IDENTITY)
    except KeyError:
        try:
            group = grp.getgrnam(IDENTITY)
        except KeyError:
            try:
                home.rmdir()
            except FileNotFoundError:
                pass
            return
        if identity.get("mode") == "creating":
            if group.gr_gid != uid:
                raise InstallError("identity removal collision")
            if first_owned_path(runner, uid, exclude_home=True):
                raise InstallError("identity group still owns files")
            runner.run(("/usr/sbin/groupdel", IDENTITY))
            return
        raise InstallError("identity removal incomplete")
    try:
        group = grp.getgrnam(IDENTITY)
    except KeyError as exc:
        raise InstallError("identity removal collision") from exc
    if (
        account.pw_uid != uid
        or account.pw_gid != int(identity["gid"])
        or group.gr_gid != int(identity["gid"])
        or os.getgrouplist(IDENTITY, account.pw_gid) != [account.pw_gid]
        or account.pw_dir != str(IDENTITY_HOME)
        or account.pw_shell != "/bin/bash"
    ):
        raise InstallError("identity removal collision")
    processes = runner.run(("/usr/bin/pgrep", "-u", str(uid)), check=False)
    if processes.returncode == 0 and processes.stdout.strip():
        raise InstallError("identity still has processes")
    if processes.returncode not in {0, 1}:
        raise InstallError("process ownership scan failed")
    sessions = runner.run(("/usr/bin/loginctl", "list-sessions", "--no-legend"), check=False)
    if sessions.returncode != 0:
        raise InstallError("session ownership scan failed")
    if sessions.returncode == 0 and any(IDENTITY in line or re.search(rf"\b{uid}\b", line) for line in sessions.stdout.splitlines()):
        raise InstallError("identity still has sessions")
    if first_owned_path(runner, uid, exclude_home=True):
        raise InstallError("identity still owns files")
    runner.run(("/usr/sbin/userdel", IDENTITY))
    fixture_checkpoint("IDENTITY_USER_REMOVED_MUTATED")
    runner.run(("/usr/sbin/groupdel", IDENTITY))
    fixture_checkpoint("IDENTITY_GROUP_REMOVED_MUTATED")
    try:
        home.rmdir()
    except FileNotFoundError:
        pass
    fixture_checkpoint("IDENTITY_HOME_REMOVED_MUTATED")
    try:
        pwd.getpwnam(IDENTITY)
    except KeyError:
        pass
    else:
        raise InstallError("identity removal failed")
    try:
        grp.getgrnam(IDENTITY)
    except KeyError:
        pass
    else:
        raise InstallError("identity removal failed")


def install_file(source: Path, destination: Path, mode: int, *, checkpoint: bool = True) -> None:
    if destination.exists() or destination.is_symlink():
        raise InstallError("managed path collision")
    ensure_root_directory(destination.parent)
    descriptor, temporary = tempfile.mkstemp(prefix=".cab-", dir=destination.parent)
    try:
        with source.open("rb") as incoming, os.fdopen(descriptor, "wb", closefd=False) as outgoing:
            shutil.copyfileobj(incoming, outgoing)
            outgoing.flush()
            os.fsync(outgoing.fileno())
        os.fchmod(descriptor, mode)
        os.fchown(descriptor, 0, 0)
        os.close(descriptor)
        descriptor = -1
        os.replace(temporary, destination)
        _fsync_dir(destination.parent)
        if checkpoint:
            fixture_checkpoint("FILE_MUTATED")
    finally:
        if descriptor >= 0:
            os.close(descriptor)
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass


def enrollment(identity: dict[str, object], runner: Runner, state: dict[str, object], rotation: bool = False) -> None:
    if not sys.stdin.isatty() or not sys.stdout.isatty():
        raise InstallError("controlling terminal required")
    secret = bytearray(secrets.token_bytes(20))
    try:
        encoded = base64.b64encode(secret).decode("ascii")
        base32 = base64.b32encode(secret).decode("ascii").rstrip("=")
        uri = f"otpauth://totp/Coding%20Agent?secret={base32}&issuer=Coding%20Agent"
        document = json.dumps(
            {"totp_secret_b64": encoded, "allowed_uid": identity["uid"]},
            separators=(",", ":"),
        ).encode()
        if rotation:
            set_phase(state, "NEW_SEED_PENDING_ACK")
        completed = subprocess.run(
            ["/usr/bin/qrencode", "-t", "ANSIUTF8"],
            input=uri.encode(),
            stdout=sys.stdout.buffer,
            stderr=subprocess.DEVNULL,
            env={"PATH": "/usr/bin", "LANG": "C"},
            check=False,
        )
        if completed.returncode != 0:
            raise InstallError("enrollment failed")
        answer = input("Authenticatorへ登録後、ENROLLED と入力してください: ")
        if answer != "ENROLLED":
            raise InstallError("enrollment not acknowledged")
        atomic_write(SEED_PATH, document, 0o600)
        fixture_checkpoint("SEED_MUTATED")
    finally:
        for index in range(len(secret)):
            secret[index] = 0


def verify_installation(runner: Runner, service: bool = True) -> None:
    expected = {Path(destination): mode for _, destination, mode in INSTALL_MAP}
    expected[SEED_PATH] = 0o600
    for path, mode in expected.items():
        metadata = path.lstat()
        if not stat.S_ISREG(metadata.st_mode) or metadata.st_uid != 0 or stat.S_IMODE(metadata.st_mode) != mode:
            raise InstallError("verification failed")
    seed = strict_json(SEED_PATH.read_bytes())
    if set(seed) != {"totp_secret_b64", "allowed_uid"} or seed["allowed_uid"] != pwd.getpwnam(IDENTITY).pw_uid:
        raise InstallError("verification failed")
    decoded = base64.b64decode(str(seed["totp_secret_b64"]), validate=True)
    if not 1 <= len(decoded) <= 128 or base64.b64encode(decoded).decode("ascii") != seed["totp_secret_b64"]:
        raise InstallError("verification failed")
    if STATE_PATH.exists():
        state = load_state()
        pre_state = dict(state["pre_state"])
        expected_digests = dict(pre_state.get("installed_digests", {}))
        if set(expected_digests) != {destination for _, destination, _ in INSTALL_MAP}:
            raise InstallError("verification failed")
        for name, expected_digest in expected_digests.items():
            if not re.fullmatch(r"[0-9a-f]{64}", str(expected_digest)) or sha256(Path(name)) != expected_digest:
                raise InstallError("verification failed")
        expected_sudo = dict(pre_state.get("sudo_digests", {}))
        candidate_policy = Path("/etc/sudoers.d/codex-authority")
        expected_sudo[str(candidate_policy)] = sha256(candidate_policy)
        if sudo_policy_snapshot() != expected_sudo:
            raise InstallError("verification failed")
        if sha256(Path("/etc/pam.d/sudo")) != pre_state.get("pam_sudo_digest"):
            raise InstallError("verification failed")
    policy = Path("/etc/sudoers.d/codex-authority").read_text(encoding="ascii")
    required = (
        "Defaults:coding-agent pam_service=codex-authority\n",
        "Defaults:coding-agent pam_login_service=codex-authority\n",
        "Defaults:coding-agent pam_askpass_service=codex-authority\n",
        "Defaults:coding-agent timestamp_timeout=0\n",
        "coding-agent ALL=(root:root) PASSWD: ALL\n",
    )
    if policy != "".join(required) or "NOPASSWD" in policy:
        raise InstallError("verification failed")
    pam = Path("/etc/pam.d/codex-authority").read_text(encoding="ascii")
    if pam != "#%PAM-1.0\nauth required pam_exec.so quiet seteuid /usr/local/bin/codex-authority-sudo\naccount required pam_permit.so\n":
        raise InstallError("verification failed")
    runner.run(("/usr/sbin/visudo", "-cf", "/etc/sudoers"))
    runner.run(("/usr/bin/systemd-analyze", "verify", "/etc/systemd/system/codex-authority-broker.service"))
    effective = runner.run(("/usr/bin/sudo", "-ll", "-U", IDENTITY)).stdout
    if (
        effective.count("Sudoers entry:") != 1
        or "RunAsUsers: root" not in effective
        or "RunAsGroups: root" not in effective
        or not re.search(r"(?m)^\s+ALL\s*$", effective)
        or "authenticate" not in effective
        or "NOPASSWD" in effective
        or "!authenticate" in effective
        or "RunAsUsers: ALL" in effective
        or "RunAsGroups: ALL" in effective
    ):
        raise InstallError("verification failed")
    if service:
        runner.run(("/usr/bin/systemctl", "is-enabled", UNIT))
        runner.run(("/usr/bin/systemctl", "is-active", UNIT))
        metadata = SOCKET_PATH.lstat()
        account = pwd.getpwnam(IDENTITY)
        if not stat.S_ISSOCK(metadata.st_mode) or metadata.st_uid != account.pw_uid or metadata.st_gid != account.pw_gid or stat.S_IMODE(metadata.st_mode) != 0o660:
            raise InstallError("verification failed")


def stop_and_disable_service(runner: Runner, *, unit_expected: bool) -> None:
    stopped = runner.run(("/usr/bin/systemctl", "stop", UNIT), check=False)
    active = runner.run(("/usr/bin/systemctl", "is-active", UNIT), check=False)
    process = runner.run(("/usr/bin/pgrep", "-f", "^/usr/local/bin/codex-authority-broker$"), check=False)
    if active.returncode == 0 or process.returncode == 0:
        raise InstallError("service stop failed")
    if active.returncode not in {1, 3, 4} or process.returncode not in {0, 1}:
        raise InstallError("service stop verification failed")
    if stopped.returncode != 0 and unit_expected:
        raise InstallError("service stop failed")
    disabled = runner.run(("/usr/bin/systemctl", "disable", UNIT), check=False)
    if disabled.returncode != 0 and unit_expected:
        raise InstallError("service disable failed")
    enabled = runner.run(("/usr/bin/systemctl", "is-enabled", UNIT), check=False)
    if enabled.returncode == 0 or enabled.stdout.strip() not in {"disabled", "not-found"}:
        raise InstallError("service disable failed")


def verify_service_absent(runner: Runner) -> None:
    active = runner.run(("/usr/bin/systemctl", "is-active", UNIT), check=False)
    enabled = runner.run(("/usr/bin/systemctl", "is-enabled", UNIT), check=False)
    process = runner.run(("/usr/bin/pgrep", "-f", "^/usr/local/bin/codex-authority-broker$"), check=False)
    if (
        active.returncode not in {3, 4}
        or enabled.returncode == 0
        or enabled.stdout.strip() not in {"disabled", "not-found"}
        or process.returncode != 1
    ):
        raise InstallError("service absence verification failed")


def rollback(state: dict[str, object], runner: Runner) -> None:
    unit_path = str(Path("/etc/systemd/system") / UNIT)
    stop_and_disable_service(runner, unit_expected=unit_path in state.get("owned_paths", []))
    try:
        SOCKET_PATH.unlink()
    except FileNotFoundError:
        pass
    candidates = list(state.get("owned_paths", []))
    pending = state.get("pending_path")
    if isinstance(pending, str) and pending.startswith("/"):
        candidates.append(pending)
    for item in reversed(candidates):
        remove_managed_path(Path(item))
    remove_identity(runner, dict(state.get("identity", {})))
    runner.run(("/usr/bin/systemctl", "daemon-reload"))
    verify_service_absent(runner)


def cleanup_install_transaction() -> None:
    for path in (STATE_PATH, STATE_DIR / "backups/seed.previous"):
        try:
            path.unlink()
        except FileNotFoundError:
            pass
    backups = STATE_DIR / "backups"
    try:
        backups.rmdir()
    except FileNotFoundError:
        pass
    try:
        STATE_DIR.rmdir()
    except FileNotFoundError:
        pass
    for path in (RECOVERY_STATE, RECOVERY_PATH, RECOVERY_CORE, RECOVERY_SUMS):
        try:
            path.unlink()
        except FileNotFoundError:
            pass
    try:
        RECOVERY_DIR.rmdir()
    except FileNotFoundError:
        pass
    _fsync_dir(Path("/var/lib"))


def install(runner: Runner) -> None:
    require_root()
    require_platform()
    remove_empty_completed_recovery()
    root = source_root()
    digest = artifact_digest(root)
    if STATE_PATH.exists():
        existing = load_state()
        if existing["operation"] == "install" and existing["phase"] == "COMMITTED" and existing["artifact_digest"] == digest:
            verify_installation(runner)
            cleanup_stale_uninstall_recovery(existing)
            return
        raise InstallError("recovery required")
    require_safe_sudo_baseline()
    sudo_digests = sudo_policy_snapshot()
    pam_sudo_digest = sha256(Path("/etc/pam.d/sudo"))
    identity = inspect_identity(runner)
    require_clean_managed_paths(identity)
    require_unit_absent(runner)
    state = state_document("install", digest, identity)
    state["pre_state"] = {
        "identity_mode": identity["mode"],
        "managed_paths_absent": True,
        "installed_digests": {},
        "sudo_digests": sudo_digests,
        "pam_sudo_digest": pam_sudo_digest,
        "unit_absent": True,
    }
    save_state(state)
    try:
        RECOVERY_DIR.mkdir(parents=True, exist_ok=False, mode=0o700)
        os.chown(RECOVERY_DIR, 0, 0)
        _fsync_dir(RECOVERY_DIR.parent)
        install_file(root / "install/codex-authority-recover", RECOVERY_PATH, 0o500, checkpoint=False)
        install_file(root / "install/codex_authority_installer.py", RECOVERY_CORE, 0o400, checkpoint=False)
        install_file(root / "SHA256SUMS", RECOVERY_SUMS, 0o400, checkpoint=False)
        state["owned_paths"].append(str(RECOVERY_PATH))
        state["owned_paths"].append(str(RECOVERY_CORE))
        state["owned_paths"].append(str(RECOVERY_SUMS))
        save_state(state)
        if identity["mode"] == "create":
            state["identity"] = allocate_identity(runner)
            set_phase(state, "IDENTITY_INTENT")
            state["identity"] = create_identity(runner, dict(state["identity"]))
            set_phase(state, "IDENTITY_DONE")
        require_no_existing_sudo_grant(runner)
        for directory in OWNED_DIRECTORIES:
            set_phase(state, "FILE_INTENT", str(directory))
            directory.mkdir(mode=0o700)
            os.chown(directory, 0, 0)
            _fsync_dir(directory.parent)
            state["owned_paths"].append(str(directory))
            set_phase(state, "FILE_DONE")
        for source_name, destination_name, mode in INSTALL_MAP:
            destination = Path(destination_name)
            set_phase(state, "FILE_INTENT", destination_name)
            install_file(root / source_name, destination, mode)
            state["owned_paths"].append(destination_name)
            pre_state = state["pre_state"]
            assert isinstance(pre_state, dict)
            installed_digests = pre_state["installed_digests"]
            assert isinstance(installed_digests, dict)
            installed_digests[destination_name] = sha256(destination)
            set_phase(state, "FILE_DONE")
        set_phase(state, "SEED_INTENT", str(SEED_PATH))
        enrollment(dict(state["identity"]), runner, state)
        state["owned_paths"].append(str(SEED_PATH))
        set_phase(state, "SEED_DONE")
        set_phase(state, "POLICY_INTENT")
        verify_installation(runner, service=False)
        set_phase(state, "POLICY_DONE")
        set_phase(state, "START_INTENT")
        runner.run(("/usr/bin/systemctl", "daemon-reload"))
        runner.run(("/usr/bin/systemctl", "enable", "--now", UNIT))
        fixture_checkpoint("SERVICE_STARTED_MUTATED")
        set_phase(state, "START_DONE")
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline and not SOCKET_PATH.is_socket():
            time.sleep(0.1)
        if not SOCKET_PATH.is_socket():
            raise InstallError("service verification failed")
        set_phase(state, "COMMITTED")
    except BaseException:
        rollback(state, runner)
        cleanup_install_transaction()
        raise


def rotate(runner: Runner) -> None:
    require_root()
    state = load_state()
    if state["operation"] != "install" or state["phase"] != "COMMITTED":
        raise InstallError("recovery required")
    verify_installation(runner)
    backup = STATE_DIR / "backups/seed.previous"
    ensure_root_directory(backup.parent)
    state["operation"] = "rotate"
    state["backup_paths"] = [str(backup)]
    state["phase"] = "ROTATE_PREPARED"
    state["pre_state"]["seed_backup_digest"] = sha256(SEED_PATH)
    save_state(state)
    try:
        set_phase(state, "BACKUP_INTENT")
        install_file(SEED_PATH, backup, 0o600, checkpoint=False)
        require_root_file_digest(backup, str(state["pre_state"]["seed_backup_digest"]), 0o600)
        set_phase(state, "BACKUP_DONE")
        runner.run(("/usr/bin/systemctl", "stop", UNIT))
        fixture_checkpoint("ROTATION_SERVICE_STOPPED_MUTATED")
        set_phase(state, "BROKER_STOPPED")
        set_phase(state, "NEW_SEED_INTENT")
        enrollment(dict(state["identity"]), runner, state, rotation=True)
        set_phase(state, "ENROLLMENT_ACKED")
        commit_rotation(state, runner)
    except BaseException:
        if state["phase"] in {"ENROLLMENT_ACKED", "RESTART_INTENT", "RESTART_DONE"}:
            commit_rotation(state, runner)
            raise
        if backup.exists():
            require_root_file_digest(backup, str(dict(state["pre_state"])["seed_backup_digest"]), 0o600)
            os.replace(backup, SEED_PATH)
            _fsync_dir(SEED_PATH.parent)
            runner.run(("/usr/bin/systemctl", "start", UNIT))
            deadline = time.monotonic() + 5
            while time.monotonic() < deadline and not SOCKET_PATH.is_socket():
                time.sleep(0.1)
            verify_installation(runner)
        state["operation"] = "install"
        state["backup_paths"] = []
        state["pre_state"].pop("seed_backup_digest", None)
        set_phase(state, "COMMITTED")
        raise


def commit_rotation(state: dict[str, object], runner: Runner, state_path: Path = STATE_PATH) -> None:
    backup = Path(list(state["backup_paths"])[0])
    if state["phase"] != "RESTART_DONE":
        set_phase(state, "RESTART_INTENT", path=state_path)
        runner.run(("/usr/bin/systemctl", "start", UNIT))
        fixture_checkpoint("ROTATION_SERVICE_STARTED_MUTATED")
        set_phase(state, "RESTART_DONE", path=state_path)
    deadline = time.monotonic() + 5
    while time.monotonic() < deadline and not SOCKET_PATH.is_socket():
        time.sleep(0.1)
    verify_installation(runner)
    try:
        backup.unlink()
        _fsync_dir(backup.parent)
    except FileNotFoundError:
        pass
    state["operation"] = "install"
    state["backup_paths"] = []
    state["pre_state"].pop("seed_backup_digest", None)
    state["seed_generation"] = int(state["seed_generation"]) + 1
    set_phase(state, "COMMITTED", path=state_path)


def recover(runner: Runner) -> None:
    require_root()
    state_path = STATE_PATH if STATE_PATH.exists() else RECOVERY_STATE
    state = load_state(state_path)
    if state["operation"] == "install" and state["phase"] == "COMMITTED":
        verify_installation(runner)
        cleanup_stale_uninstall_recovery(state)
        return
    if state["operation"] == "rotate":
        backup_paths = list(state["backup_paths"])
        if state["phase"] in {"ENROLLMENT_ACKED", "RESTART_INTENT", "RESTART_DONE"}:
            commit_rotation(state, runner, state_path)
            return
        if state["phase"] in {"ROTATE_PREPARED", "BACKUP_INTENT"} and not any(Path(item).exists() for item in backup_paths):
            state["operation"] = "install"
            state["backup_paths"] = []
            state["pre_state"].pop("seed_backup_digest", None)
            set_phase(state, "COMMITTED", path=state_path)
            return
        if backup_paths:
            backup = Path(backup_paths[0])
            if backup.exists():
                require_root_file_digest(backup, str(dict(state["pre_state"])["seed_backup_digest"]), 0o600)
                os.replace(backup, SEED_PATH)
                _fsync_dir(SEED_PATH.parent)
                runner.run(("/usr/bin/systemctl", "start", UNIT))
                deadline = time.monotonic() + 5
                while time.monotonic() < deadline and not SOCKET_PATH.is_socket():
                    time.sleep(0.1)
                verify_installation(runner)
                state["operation"] = "install"
                state["backup_paths"] = []
                state["pre_state"].pop("seed_backup_digest", None)
                set_phase(state, "COMMITTED", path=state_path)
                return
    if state["operation"] == "uninstall":
        if state["phase"] == "COMMITTED":
            finish_uninstall_cleanup(state)
        else:
            rollback_uninstall(state, runner)
        return
    rollback(state, runner)
    cleanup_install_transaction()


def save_uninstall_state(state: dict[str, object]) -> None:
    save_state(state)
    save_state(state, RECOVERY_STATE)


def prepare_uninstall_backups(state: dict[str, object]) -> None:
    pre_state = state["pre_state"]
    assert isinstance(pre_state, dict)
    records = pre_state.setdefault("uninstall_backups", {})
    assert isinstance(records, dict)
    sources = [Path(destination) for _, destination, _ in INSTALL_MAP] + [SEED_PATH]
    for source in sources:
        if str(source) in records:
            continue
        metadata = source.lstat()
        if not stat.S_ISREG(metadata.st_mode) or metadata.st_uid != 0:
            raise InstallError("uninstall backup failed")
        backup = UNINSTALL_BACKUP_DIR / str(source).lstrip("/")
        ensure_root_directory(backup.parent)
        os.chmod(backup.parent, 0o700)
        os.chown(backup.parent, 0, 0)
        install_file(source, backup, stat.S_IMODE(metadata.st_mode), checkpoint=False)
        records[str(source)] = {
            "path": str(backup),
            "digest": sha256(backup),
            "mode": stat.S_IMODE(metadata.st_mode),
        }
        save_uninstall_state(state)


def recreate_identity(identity: dict[str, object], runner: Runner) -> None:
    if identity.get("mode") != "created":
        return
    try:
        pwd.getpwnam(IDENTITY)
    except KeyError:
        pass
    else:
        existing = inspect_identity(runner)
        if existing["uid"] != identity["uid"] or existing["gid"] != identity["gid"]:
            raise InstallError("identity recovery collision")
        return
    uid = str(identity["uid"])
    gid = str(identity["gid"])
    if first_owned_path(runner, int(uid), exclude_home=True):
        raise InstallError("identity recovery collision")
    if IDENTITY_HOME.exists():
        metadata = IDENTITY_HOME.lstat()
        if (
            not stat.S_ISDIR(metadata.st_mode)
            or IDENTITY_HOME.is_symlink()
            or stat.S_IMODE(metadata.st_mode) != 0o700
            or metadata.st_uid != int(uid)
            or metadata.st_gid != int(gid)
            or any(IDENTITY_HOME.iterdir())
        ):
            raise InstallError("identity recovery collision")
    try:
        group = grp.getgrnam(IDENTITY)
    except KeyError:
        runner.run(("/usr/sbin/groupadd", "--system", "--gid", gid, IDENTITY))
    else:
        if group.gr_gid != int(gid):
            raise InstallError("identity recovery collision")
    runner.run((
        "/usr/sbin/useradd", "--system", "--uid", uid, "--gid", gid, "--home-dir",
        str(IDENTITY_HOME), "--no-create-home", "--shell", "/bin/bash", IDENTITY,
    ))
    runner.run(("/usr/bin/passwd", "-l", IDENTITY))
    IDENTITY_HOME.mkdir(mode=0o700, exist_ok=True)
    os.chown(IDENTITY_HOME, int(uid), int(gid))
    os.chmod(IDENTITY_HOME, 0o700)
    _fsync_dir(IDENTITY_HOME.parent)


def rollback_uninstall(state: dict[str, object], runner: Runner) -> None:
    pre_state = state["pre_state"]
    assert isinstance(pre_state, dict)
    records = pre_state.get("uninstall_backups", {})
    if not isinstance(records, dict):
        raise InstallError("uninstall recovery failed")
    expected = {destination for _, destination, _ in INSTALL_MAP} | {str(SEED_PATH)}
    if set(records) != expected:
        if state["phase"] != "UNINSTALL_PREPARED" or any(not Path(item).is_file() for item in expected):
            raise InstallError("uninstall recovery incomplete")
        pre_state.pop("uninstall_backups", None)
        state["operation"] = "install"
        state["phase"] = "COMMITTED"
        state["pending_path"] = None
        save_state(state)
        if UNINSTALL_BACKUP_DIR.exists():
            shutil.rmtree(UNINSTALL_BACKUP_DIR)
            _fsync_dir(RECOVERY_DIR)
        try:
            RECOVERY_STATE.unlink()
        except FileNotFoundError:
            pass
        verify_installation(runner)
        return
    recreate_identity(dict(state["identity"]), runner)
    for directory in OWNED_DIRECTORIES:
        directory.mkdir(mode=0o700, exist_ok=True)
        os.chmod(directory, 0o700)
        os.chown(directory, 0, 0)
    for source_name, record in records.items():
        source = Path(source_name)
        backup = Path(record["path"])
        require_root_file_digest(backup, str(record["digest"]), int(record["mode"]))
        if source.exists():
            require_root_file_digest(source, str(record["digest"]), int(record["mode"]))
        else:
            install_file(backup, source, int(record["mode"]), checkpoint=False)
    runner.run(("/usr/bin/systemctl", "daemon-reload"))
    runner.run(("/usr/bin/systemctl", "enable", "--now", UNIT))
    pre_state.pop("uninstall_backups", None)
    state["operation"] = "install"
    state["phase"] = "COMMITTED"
    state["pending_path"] = None
    save_state(state)
    if UNINSTALL_BACKUP_DIR.exists():
        shutil.rmtree(UNINSTALL_BACKUP_DIR)
        _fsync_dir(RECOVERY_DIR)
    try:
        RECOVERY_STATE.unlink()
    except FileNotFoundError:
        pass
    verify_installation(runner)


def remove_uninstall_backups(state: dict[str, object]) -> None:
    records = dict(dict(state["pre_state"])["uninstall_backups"])
    sources = {destination for _, destination, _ in INSTALL_MAP} | {str(SEED_PATH)}
    expected_files = {UNINSTALL_BACKUP_DIR / source.lstrip("/") for source in sources}
    actual_files: set[Path] = set()
    if UNINSTALL_BACKUP_DIR.exists():
        for path in UNINSTALL_BACKUP_DIR.rglob("*"):
            metadata = path.lstat()
            if stat.S_ISREG(metadata.st_mode) and metadata.st_uid == 0 and not path.is_symlink():
                actual_files.add(path)
            elif not stat.S_ISDIR(metadata.st_mode) or metadata.st_uid != 0 or path.is_symlink():
                raise InstallError("invalid uninstall backup")
    if not actual_files <= expected_files:
        raise InstallError("invalid uninstall backup")
    for source, record in records.items():
        backup = Path(record["path"])
        if backup.exists():
            require_root_file_digest(backup, str(record["digest"]), int(record["mode"]))
            backup.unlink()
    recorded_files = {Path(record["path"]) for record in records.values()}
    for backup in actual_files - recorded_files:
        source = Path("/") / backup.relative_to(UNINSTALL_BACKUP_DIR)
        metadata = source.lstat()
        require_root_file_digest(backup, sha256(source), stat.S_IMODE(metadata.st_mode))
        backup.unlink()
    if UNINSTALL_BACKUP_DIR.exists():
        for directory in sorted((path for path in UNINSTALL_BACKUP_DIR.rglob("*") if path.is_dir()), key=lambda path: len(path.parts), reverse=True):
            directory.rmdir()
        UNINSTALL_BACKUP_DIR.rmdir()
    _fsync_dir(RECOVERY_DIR)


def cleanup_stale_uninstall_recovery(installed_state: dict[str, object]) -> None:
    if not RECOVERY_STATE.exists():
        if UNINSTALL_BACKUP_DIR.exists() or UNINSTALL_BACKUP_DIR.is_symlink():
            raise InstallError("stale uninstall recovery")
        return
    stale = load_state(RECOVERY_STATE)
    if (
        installed_state["operation"] != "install"
        or installed_state["phase"] != "COMMITTED"
        or stale["operation"] != "uninstall"
        or stale["artifact_digest"] != installed_state["artifact_digest"]
        or stale["identity"] != installed_state["identity"]
    ):
        raise InstallError("stale uninstall recovery")
    remove_uninstall_backups(stale)
    RECOVERY_STATE.unlink()
    _fsync_dir(RECOVERY_DIR)


def finish_uninstall_cleanup(state: dict[str, object]) -> None:
    if state["operation"] != "uninstall" or state["phase"] != "COMMITTED":
        raise InstallError("invalid uninstall cleanup")
    if COMPLETED_RECOVERY_DIR.exists():
        finish_completed_recovery()
        return
    try:
        STATE_PATH.unlink()
    except FileNotFoundError:
        pass
    backups = STATE_DIR / "backups"
    if backups.exists():
        backups.rmdir()
    try:
        STATE_DIR.rmdir()
    except FileNotFoundError:
        pass
    fixture_checkpoint("CLEANUP_STATE_MUTATED")
    remove_uninstall_backups(state)
    fixture_checkpoint("CLEANUP_BACKUPS_MUTATED")
    try:
        RECOVERY_STATE.unlink()
    except FileNotFoundError:
        pass
    _fsync_dir(RECOVERY_DIR)
    fixture_checkpoint("CLEANUP_RECOVERY_STATE_MUTATED")
    if COMPLETED_RECOVERY_DIR.exists() or COMPLETED_RECOVERY_DIR.is_symlink():
        raise InstallError("cleanup collision")
    os.replace(RECOVERY_DIR, COMPLETED_RECOVERY_DIR)
    _fsync_dir(COMPLETED_RECOVERY_DIR.parent)
    fixture_checkpoint("CLEANUP_RENAMED")
    finish_completed_recovery()


def finish_completed_recovery() -> None:
    metadata = COMPLETED_RECOVERY_DIR.lstat()
    if not stat.S_ISDIR(metadata.st_mode) or metadata.st_uid != 0 or stat.S_IMODE(metadata.st_mode) != 0o700 or COMPLETED_RECOVERY_DIR.is_symlink():
        raise InstallError("invalid completed recovery")
    allowed = {"recover", "codex_authority_installer.py", "SHA256SUMS"}
    if any(path.name not in allowed or not path.is_file() or path.is_symlink() for path in COMPLETED_RECOVERY_DIR.iterdir()):
        raise InstallError("invalid completed recovery")
    for name in ("codex_authority_installer.py", "SHA256SUMS"):
        try:
            (COMPLETED_RECOVERY_DIR / name).unlink()
        except FileNotFoundError:
            pass
    try:
        (COMPLETED_RECOVERY_DIR / "recover").unlink()
    except FileNotFoundError:
        pass
    COMPLETED_RECOVERY_DIR.rmdir()
    _fsync_dir(COMPLETED_RECOVERY_DIR.parent)


def continue_uninstall(state: dict[str, object], runner: Runner) -> None:
    def phase(name: str, pending: str | None = None) -> None:
        set_phase(state, name, pending)
        save_state(state, RECOVERY_STATE)

    phase("UNINSTALL_PREPARED")
    phase("STOP_INTENT")
    stop_and_disable_service(runner, unit_expected=True)
    fixture_checkpoint("SERVICE_STOPPED_MUTATED")
    phase("STOP_DONE")
    try:
        SOCKET_PATH.unlink()
    except FileNotFoundError:
        pass
    for item in reversed(list(state["owned_paths"])):
        if item in {str(RECOVERY_PATH), str(RECOVERY_CORE), str(RECOVERY_SUMS)}:
            continue
        phase("REMOVE_INTENT", item)
        remove_managed_path(Path(item))
        fixture_checkpoint("REMOVE_MUTATED")
        phase("REMOVE_DONE")
    phase("TIMESTAMP_INTENT")
    uid = int(dict(state["identity"])["uid"])
    for timestamp in (Path("/run/sudo/ts") / IDENTITY, Path("/run/sudo/ts") / str(uid)):
        try:
            timestamp.unlink()
        except FileNotFoundError:
            pass
    fixture_checkpoint("TIMESTAMP_MUTATED")
    phase("TIMESTAMP_DONE")
    phase("IDENTITY_INTENT")
    remove_identity(runner, dict(state["identity"]))
    fixture_checkpoint("IDENTITY_REMOVED_MUTATED")
    phase("IDENTITY_DONE")
    phase("RELOAD_INTENT")
    runner.run(("/usr/bin/systemctl", "daemon-reload"))
    fixture_checkpoint("RELOAD_MUTATED")
    phase("RELOAD_DONE")
    verify_service_absent(runner)
    if SOCKET_PATH.exists() or SOCKET_PATH.is_symlink():
        raise InstallError("uninstall verification failed")
    for item in state["owned_paths"]:
        if item not in {str(RECOVERY_PATH), str(RECOVERY_CORE), str(RECOVERY_SUMS)} and (Path(item).exists() or Path(item).is_symlink()):
            raise InstallError("uninstall verification failed")
    phase("VERIFY_DONE")
    state["phase"] = "COMMITTED"
    save_uninstall_state(state)
    fixture_checkpoint("CLEANUP_COMMITTED")
    finish_uninstall_cleanup(state)


def uninstall(runner: Runner) -> None:
    require_root()
    state = load_state()
    if state["operation"] != "install" or state["phase"] != "COMMITTED":
        raise InstallError("recovery required")
    verify_installation(runner)
    cleanup_stale_uninstall_recovery(state)
    state["operation"] = "uninstall"
    state["phase"] = "UNINSTALL_PREPARED"
    state["pre_state"]["uninstall_backups"] = {}
    save_uninstall_state(state)
    try:
        prepare_uninstall_backups(state)
        continue_uninstall(state, runner)
    except BaseException:
        if state["phase"] == "COMMITTED":
            finish_uninstall_cleanup(state)
        else:
            rollback_uninstall(state, runner)
            raise


def cli(operation: str) -> int:
    if len(sys.argv) != 1:
        print("request denied", file=sys.stderr)
        return 1
    os.umask(0o077)
    def controlled_signal(_number: int, _frame: object) -> None:
        raise InstallError("controlled signal")

    for number in (signal.SIGHUP, signal.SIGINT, signal.SIGQUIT, signal.SIGTERM):
        signal.signal(number, controlled_signal)
    runner = Runner()
    actions = {
        "install": lambda: install(runner),
        "verify": lambda: verify_installation(runner),
        "rotate": lambda: rotate(runner),
        "recover": lambda: recover(runner),
        "uninstall": lambda: uninstall(runner),
    }
    try:
        actions[operation]()
    except (InstallError, OSError, subprocess.SubprocessError, ValueError, KeyError):
        print("request denied", file=sys.stderr)
        return 1
    print(f"{operation} complete")
    return 0
