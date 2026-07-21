import importlib.util
import json
import os
import stat
import sys
import tempfile
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest import mock


MODULE = Path(__file__).parents[1] / "codex_authority_installer.py"
SPEC = importlib.util.spec_from_file_location("cab_installer", MODULE)
cab = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
sys.modules[SPEC.name] = cab
SPEC.loader.exec_module(cab)


class InstallerTests(unittest.TestCase):
    def valid_state(self):
        return cab.state_document("install", "a" * 64, {"mode": "create", "uid": 0, "gid": 0})

    def committed_state(self):
        state = cab.state_document("install", "a" * 64, {"mode": "reuse", "uid": 998, "gid": 998})
        state["phase"] = "COMMITTED"
        state["pre_state"]["identity_mode"] = "reuse"
        state["pre_state"]["installed_digests"] = {
            destination: "b" * 64 for _, destination, _ in cab.INSTALL_MAP
        }
        state["owned_paths"] = [
            str(cab.RECOVERY_PATH), str(cab.RECOVERY_CORE), str(cab.RECOVERY_SUMS),
            *(str(path) for path in cab.OWNED_DIRECTORIES),
            *(destination for _, destination, _ in cab.INSTALL_MAP), str(cab.SEED_PATH),
        ]
        return state

    def test_state_schema_is_exact(self):
        state = self.valid_state()
        self.assertIs(cab.validate_state(state), state)
        for mutation in (
            lambda value: value.update(extra=True),
            lambda value: value.pop("phase"),
            lambda value: value.update(version=2),
            lambda value: value.update(artifact_digest="bad"),
            lambda value: value.update(owned_paths=[1]),
            lambda value: value.update(owned_paths=["/etc/shadow"]),
            lambda value: value.update(backup_paths=["/tmp/seed"]),
            lambda value: value.update(pending_path=1),
            lambda value: value.update(owned_paths=["/etc/pam.d/codex-authority"]),
            lambda value: value.update(operation="rotate", phase="FILE_DONE"),
            lambda value: value.update(operation="uninstall", phase="COMMITTED", pending_path="/etc/shadow"),
            lambda value: value["pre_state"].pop("managed_paths_absent"),
        ):
            changed = json.loads(json.dumps(state))
            mutation(changed)
            with self.assertRaises(cab.InstallError):
                cab.validate_state(changed)

    def test_duplicate_json_key_is_rejected(self):
        with self.assertRaises(cab.InstallError):
            cab.strict_json(b'{"version":1,"version":1}')

    def test_committed_rotation_and_uninstall_state_invariants(self):
        committed = self.committed_state()
        self.assertIs(cab.validate_state(committed), committed)
        mismatched = json.loads(json.dumps(committed))
        mismatched["pre_state"]["identity_mode"] = "create"
        with self.assertRaises(cab.InstallError):
            cab.validate_state(mismatched)
        missing_owned = json.loads(json.dumps(committed))
        missing_owned["owned_paths"].pop()
        with self.assertRaises(cab.InstallError):
            cab.validate_state(missing_owned)

    def test_uninstall_identity_phases_accept_committed_created_and_reused_modes(self):
        for mode in ("reuse", "created"):
            with self.subTest(mode=mode):
                state = self.committed_state()
                state["operation"] = "uninstall"
                state["phase"] = "IDENTITY_INTENT"
                if mode == "created":
                    state["identity"] = {"mode": "created", "uid": 998, "gid": 998}
                    state["pre_state"]["identity_mode"] = "create"
                backups = {}
                for _, source, file_mode in cab.INSTALL_MAP:
                    backups[source] = {
                        "path": str(cab.UNINSTALL_BACKUP_DIR / source.lstrip("/")),
                        "digest": "c" * 64,
                        "mode": file_mode,
                    }
                backups[str(cab.SEED_PATH)] = {
                    "path": str(cab.UNINSTALL_BACKUP_DIR / str(cab.SEED_PATH).lstrip("/")),
                    "digest": "d" * 64,
                    "mode": 0o600,
                }
                state["pre_state"]["uninstall_backups"] = backups
                self.assertIs(cab.validate_state(state), state)

    def test_atomic_state_has_root_only_mode_and_round_trips(self):
        if os.geteuid() != 0:
            self.skipTest("metadata test requires root fixture")
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "state.json"
            state = self.valid_state()
            cab.save_state(state, path)
            self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)
            self.assertEqual(cab.load_state(path), state)

    def test_sudo_policy_is_root_only_and_uncached(self):
        policy = (Path(__file__).parents[2] / "deploy/sudo/codex-authority").read_text()
        self.assertEqual(
            policy.splitlines(),
            [
                "Defaults:coding-agent pam_service=codex-authority",
                "Defaults:coding-agent pam_login_service=codex-authority",
                "Defaults:coding-agent pam_askpass_service=codex-authority",
                "Defaults:coding-agent timestamp_timeout=0",
                "coding-agent ALL=(root:root) PASSWD: ALL",
            ],
        )
        self.assertNotIn("NOPASSWD", policy)
        self.assertNotIn("codex-fixture", policy)
        self.assertNotIn("(ALL:ALL)", policy)

    def test_unsafe_sudo_baselines_are_rejected(self):
        for policy in (
            "coding-agent ALL=(ALL) ALL\n",
            "Defaults !authenticate\n",
            "Defaults exempt_group=sudo\n",
            "Defaults pam_login_service=sudo-i\n",
        ):
            with self.subTest(policy=policy), mock.patch.object(cab, "sudo_policy_text", return_value=policy):
                with self.assertRaises(cab.InstallError):
                    cab.require_safe_sudo_baseline()

    def test_existing_effective_sudo_grant_is_rejected(self):
        runner = mock.Mock()
        runner.run.return_value = SimpleNamespace(
            returncode=0, stdout="Sudoers entry:\n    Commands:\n        ALL\n", stderr="",
        )
        with self.assertRaises(cab.InstallError):
            cab.require_no_existing_sudo_grant(runner)

    def test_service_stop_must_be_observed_before_removal(self):
        runner = mock.Mock()
        runner.run.side_effect = (
            SimpleNamespace(returncode=0, stdout="", stderr=""),
            SimpleNamespace(returncode=0, stdout="active\n", stderr=""),
            SimpleNamespace(returncode=1, stdout="", stderr=""),
        )
        with self.assertRaises(cab.InstallError):
            cab.stop_and_disable_service(runner, unit_expected=True)

    def test_successful_service_stop_checks_disable_and_process_absence(self):
        runner = mock.Mock()
        runner.run.side_effect = (
            SimpleNamespace(returncode=0, stdout="", stderr=""),
            SimpleNamespace(returncode=3, stdout="inactive\n", stderr=""),
            SimpleNamespace(returncode=1, stdout="", stderr=""),
            SimpleNamespace(returncode=0, stdout="", stderr=""),
            SimpleNamespace(returncode=1, stdout="disabled\n", stderr=""),
        )
        cab.stop_and_disable_service(runner, unit_expected=True)
        self.assertEqual(runner.run.call_count, 5)

    def test_mount_scan_checks_uid_and_gid_on_every_mount(self):
        runner = mock.Mock()
        def result(argv, **kwargs):
            if argv[0] == "/usr/bin/findmnt":
                return SimpleNamespace(returncode=0, stdout="/\n/run\n", stderr="")
            self.assertIn("-uid", argv)
            self.assertIn("-gid", argv)
            found = "/run/orphan\n" if argv[1] == "/run" else ""
            return SimpleNamespace(returncode=0, stdout=found, stderr="")
        runner.run.side_effect = result
        self.assertEqual(cab.first_owned_path(runner, 998), "/run/orphan")

    def test_unit_dropin_collision_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            dropin = root / "unit.service.d"
            dropin.mkdir()
            identity = {"mode": "reuse", "uid": 998, "gid": 998}
            with mock.patch.object(cab, "INSTALL_MAP", ()), \
                mock.patch.object(cab, "STATE_DIR", root / "state"), \
                mock.patch.object(cab, "RECOVERY_DIR", root / "recovery"), \
                mock.patch.object(cab, "COMPLETED_RECOVERY_DIR", root / "completed"), \
                mock.patch.object(cab, "SEED_PATH", root / "seed"), \
                mock.patch.object(cab, "OWNED_DIRECTORIES", ()), \
                mock.patch.object(cab, "UNIT_PATHS", ()), \
                mock.patch.object(cab, "UNIT_DROPINS", (dropin,)):
                with self.assertRaises(cab.InstallError):
                    cab.require_clean_managed_paths(identity)

    def test_rotation_commit_is_forward_and_removes_old_backup(self):
        state = self.committed_state()
        state["operation"] = "rotate"
        state["phase"] = "ENROLLMENT_ACKED"
        state["pre_state"]["seed_backup_digest"] = "c" * 64
        with tempfile.TemporaryDirectory() as directory:
            backup = Path(directory) / "seed.previous"
            backup.write_bytes(b"old")
            state["backup_paths"] = [str(backup)]
            runner = mock.Mock()
            runner.run.return_value = SimpleNamespace(returncode=0, stdout="", stderr="")
            socket = mock.Mock()
            socket.is_socket.return_value = True
            def phase(value, name, pending=None, path=None):
                value["phase"] = name
            with mock.patch.object(cab, "SOCKET_PATH", socket), \
                mock.patch.object(cab, "set_phase", side_effect=phase), \
                mock.patch.object(cab, "verify_installation"), \
                mock.patch.object(cab, "fixture_checkpoint"), \
                mock.patch.object(cab, "_fsync_dir"):
                cab.commit_rotation(state, runner)
            self.assertEqual((state["operation"], state["phase"], state["seed_generation"]), ("install", "COMMITTED", 1))
            self.assertFalse(backup.exists())

    def test_completed_recovery_self_cleanup_removes_last_residue(self):
        with tempfile.TemporaryDirectory() as directory:
            completed = Path(directory) / "completed"
            completed.mkdir(mode=0o700)
            for name, mode in (("recover", 0o500), ("codex_authority_installer.py", 0o400), ("SHA256SUMS", 0o400)):
                path = completed / name
                path.write_bytes(b"fixture")
                path.chmod(mode)
            original_lstat = Path.lstat
            def root_lstat(path):
                metadata = original_lstat(path)
                return SimpleNamespace(st_mode=metadata.st_mode, st_uid=0)
            with mock.patch.object(cab, "COMPLETED_RECOVERY_DIR", completed), \
                mock.patch.object(Path, "lstat", root_lstat), mock.patch.object(cab, "_fsync_dir"):
                cab.finish_completed_recovery()
            self.assertFalse(completed.exists())

    def test_uninstall_backup_state_rejects_arbitrary_restore_path(self):
        state = self.valid_state()
        state["operation"] = "uninstall"
        state["phase"] = "UNINSTALL_PREPARED"
        state["pre_state"]["uninstall_backups"] = {
            str(cab.SEED_PATH): {"path": "/tmp/seed", "digest": "a" * 64, "mode": 0o600},
        }
        with self.assertRaises(cab.InstallError):
            cab.validate_state(state)

    def test_artifact_manifest_rejects_missing_and_changed_payload(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            lines = []
            for name in cab.SOURCE_MEMBERS[1:]:
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(name.encode())
                path.chmod(0o644)
                lines.append(f"{cab.sha256(path)}  {name}\n")
            (root / "SHA256SUMS").write_text("".join(sorted(lines)))
            (root / "SHA256SUMS").chmod(0o644)
            self.assertEqual(len(cab.parse_sums(root, None)), len(cab.SOURCE_MEMBERS) - 1)
            with self.assertRaises(cab.InstallError):
                (root / cab.SOURCE_MEMBERS[-1]).write_bytes(b"changed")
                cab.parse_sums(root, None)

    def test_cli_rejects_arguments_without_running(self):
        with mock.patch.object(cab.sys, "argv", ["tool", "unexpected"]):
            self.assertEqual(cab.cli("verify"), 1)

    def test_rollback_removes_only_recorded_paths(self):
        with tempfile.TemporaryDirectory() as directory:
            owned = Path(directory) / "owned"
            unrelated = Path(directory) / "unrelated"
            owned.write_text("owned")
            unrelated.write_text("keep")
            state = self.valid_state()
            state["owned_paths"] = [str(owned)]
            runner = mock.Mock()
            runner.run.return_value = mock.Mock(returncode=0, stdout="", stderr="")
            with mock.patch.object(cab, "SOCKET_PATH", Path(directory) / "socket"), \
                mock.patch.object(cab, "remove_identity"), mock.patch.object(cab, "stop_and_disable_service"), \
                mock.patch.object(cab, "verify_service_absent"):
                cab.rollback(state, runner)
            self.assertFalse(owned.exists())
            self.assertTrue(unrelated.exists())

    def test_phase_names_cover_install_rotation_and_uninstall(self):
        source = MODULE.read_text()
        required = {
            "IDENTITY_INTENT", "FILE_INTENT", "SEED_INTENT", "START_INTENT",
            "ROTATE_PREPARED", "NEW_SEED_PENDING_ACK", "ENROLLMENT_ACKED",
            "UNINSTALL_PREPARED", "REMOVE_INTENT", "VERIFY_DONE", "COMMITTED",
        }
        self.assertFalse(required - set(item for item in required if item in source))

    def test_fixture_checkpoint_requires_root_regular_marker(self):
        with tempfile.TemporaryDirectory() as directory:
            marker = Path(directory) / "marker"
            marker.write_text("fail FILE_INTENT\n")
            marker.chmod(0o644)
            with mock.patch.object(cab, "FIXTURE_MARKER", marker):
                with self.assertRaises(cab.InstallError):
                    cab.fixture_checkpoint("FILE_INTENT")

    def test_fixture_checkpoint_injects_bounded_failure(self):
        with tempfile.TemporaryDirectory() as directory:
            marker = Path(directory) / "marker"
            marker.write_text("fail FILE_INTENT\n")
            metadata = SimpleNamespace(st_mode=stat.S_IFREG | 0o600, st_uid=0)
            with mock.patch.object(cab, "FIXTURE_MARKER", marker), mock.patch.object(Path, "lstat", return_value=metadata):
                with self.assertRaisesRegex(cab.InstallError, "injected failure"):
                    cab.fixture_checkpoint("FILE_INTENT")

    def test_create_identity_selects_equal_free_uid_gid(self):
        runner = mock.Mock()
        runner.run.side_effect = lambda argv, **kwargs: SimpleNamespace(
            stdout="/\n" if "findmnt" in argv[0] else "", stderr="", returncode=0,
        )
        account = SimpleNamespace(pw_uid=998, pw_gid=998)
        with tempfile.TemporaryDirectory() as directory, \
            mock.patch.object(cab.pwd, "getpwall", return_value=[SimpleNamespace(pw_uid=999)]), \
            mock.patch.object(cab.grp, "getgrall", return_value=[SimpleNamespace(gr_gid=999)]), \
            mock.patch.object(cab.pwd, "getpwnam", return_value=account), \
            mock.patch.object(cab, "IDENTITY_HOME", Path(directory) / "coding-agent"), \
            mock.patch.object(cab, "_fsync_dir"), \
            mock.patch.object(cab.os, "chown"):
            identity = cab.create_identity(runner)
        self.assertEqual(identity, {"mode": "created", "uid": 998, "gid": 998})
        calls = [item.args[0] for item in runner.run.call_args_list]
        groupadd = next(item for item in calls if item[0] == "/usr/sbin/groupadd")
        useradd = next(item for item in calls if item[0] == "/usr/sbin/useradd")
        self.assertIn("998", groupadd)
        self.assertEqual(useradd[useradd.index("--uid") + 1], "998")
        self.assertEqual(useradd[useradd.index("--gid") + 1], "998")
        self.assertIn("--no-create-home", useradd)

    def test_remove_created_identity_refuses_nonempty_home(self):
        with tempfile.TemporaryDirectory() as directory:
            home = Path(directory)
            (home / "user-data").write_text("keep")
            with mock.patch.object(cab, "IDENTITY_HOME", home):
                with self.assertRaises(cab.InstallError):
                    cab.remove_identity(mock.Mock(), {"mode": "created", "uid": 998, "gid": 998})

    def test_remove_identity_rejects_rebound_name(self):
        with tempfile.TemporaryDirectory() as directory:
            home = Path(directory)
            account = SimpleNamespace(pw_uid=997, pw_gid=997, pw_dir=str(home), pw_shell="/bin/bash")
            group = SimpleNamespace(gr_gid=997)
            with mock.patch.object(cab, "IDENTITY_HOME", home), \
                mock.patch.object(cab.pwd, "getpwnam", return_value=account), \
                mock.patch.object(cab.grp, "getgrnam", return_value=group):
                with self.assertRaises(cab.InstallError):
                    cab.remove_identity(mock.Mock(), {"mode": "created", "uid": 998, "gid": 998})

    def test_recreate_identity_rejects_incompatible_existing_name_before_mutation(self):
        runner = mock.Mock()
        with mock.patch.object(cab.pwd, "getpwnam", return_value=SimpleNamespace()), \
            mock.patch.object(cab, "inspect_identity", return_value={"mode": "reuse", "uid": 997, "gid": 997}):
            with self.assertRaises(cab.InstallError):
                cab.recreate_identity({"mode": "created", "uid": 998, "gid": 998}, runner)
        runner.run.assert_not_called()

    def test_recreate_identity_rejects_empty_unsafe_existing_home(self):
        runner = mock.Mock()
        with tempfile.TemporaryDirectory() as directory:
            home = Path(directory)
            home.chmod(0o777)
            with mock.patch.object(cab, "IDENTITY_HOME", home), \
                mock.patch.object(cab.pwd, "getpwnam", side_effect=KeyError), \
                mock.patch.object(cab, "first_owned_path", return_value=""):
                with self.assertRaises(cab.InstallError):
                    cab.recreate_identity({"mode": "created", "uid": 998, "gid": 998}, runner)
        runner.run.assert_not_called()


if __name__ == "__main__":
    unittest.main()
