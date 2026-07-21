# TASK-0021 手動E2E試験

この手順は、snapshot/restore、reboot、out-of-band root consoleを持つ破棄可能な
Ubuntu 24.04 LTS amd64 VM専用です。workstationや代替不能hostでは実行しません。

runbook内のcandidate tree/archive digestが固定されたcommitをcheckoutし、runbookをroot-owned
mode 0500でVMへ配置してsourceとbyte比較します。返却する証跡は
`Q21-NN result=... count=... digest=...`だけです。QR、seed、TOTP、raw journal、shadow、環境、
credential、VM endpointは返却しません。

checkout直後に通常ユーザーで`tasks/TASK-0021/BUILD_CANDIDATE.sh`を実行します。固定digestと一致した
archiveだけが`/tmp/task0021-candidate/codex-authority-linux-amd64.tar.gz`へ生成されます。既存の
`/tmp/task0021-candidate`がある場合は、snapshotを戻してから再開します。

## 基本sequence

1. clean snapshotを作成し、root consoleを実際に確認する。各rollback/uninstall caseの直前にclean
   snapshotで`host-state-before`を実行し、操作後に`host-state-compare`がPASSすることを要求する。
2. `E2E_RUNBOOK.sh preflight`をrootで実行する。
3. candidate archiveをroot-owned stagingへcopyし、`verify-archive`を実行する。
4. archiveのexact manifestと内側checksumを検証し、installerを実行する。表示されたQRを登録し、
   `ENROLLED`を入力する。
5. `post-install`をrootで実行する。
6. rootで`audit-count`を実行してcountを控える。`coding-agent` shellでユーザーマニュアルどおり
   ready/TOTPを行い、`sudo-allow`を実行する。rootで`audit-delta <最初のcount>`を実行し、差が
   正確に7であることをrunbook自身に判定させる。
7. `expiry-noncontainment`を実行し、自然な301秒待機後のdenyとroot marker残存を確認する。
8. root consoleからbroker stop、fresh restartをそれぞれ行い、その都度`coding-agent`で
   `sudo-deny`を実行する。新しいready/TOTPだけがallowを戻すことを確認する。
   別snapshotではrootの`audit-total` countを控えてから`fault-socket regular`、`malformed`、
   `timeout`を一つずつarmし、各回`coding-agent`で`sudo-deny-one`、rootで
   `audit-no-delta <count>`、`clear-fault`を実行する。`fault-audit`では`ready-deny-one`、
   `audit-no-delta <count>`、`clear-fault`の順に実行する。`ARMED`はPASS証跡として扱わない。
9. root consoleで`codex-authority-admin`を実行してQRをrotationし、旧code/旧lease denyと
   新codeによるfresh activationを確認する。
10. reinstall、uninstall、作成identityなら`post-uninstall created`（再利用なら`reuse`）、
    `cleanup-marker`を実行し、snapshotを破棄する。

## identity cases

別snapshotでcompatible reuseを作り、UID/GID、group、locked password、home/shell、home digestが
install/uninstall前後で一致することを確認します。user-only、group-only、UID/GID不一致、wrong
primary group、supplementary group、unlocked password、wrong home/shell、capability、既存grant、
既存timestampを各snapshotで作り、すべてhost mutation前に拒否されることを確認します。

## failure/reboot matrix

次の各phaseをclean snapshotから独立に試験します。

```text
IDENTITY_INTENT IDENTITY_GROUP_MUTATED IDENTITY_USER_MUTATED IDENTITY_HOME_MUTATED
FILE_INTENT FILE_MUTATED SEED_INTENT SEED_MUTATED POLICY_INTENT START_INTENT SERVICE_STARTED_MUTATED
BROKER_STOPPED ROTATION_SERVICE_STOPPED_MUTATED NEW_SEED_INTENT NEW_SEED_PENDING_ACK
ENROLLMENT_ACKED RESTART_INTENT ROTATION_SERVICE_STARTED_MUTATED
UNINSTALL_PREPARED STOP_INTENT SERVICE_STOPPED_MUTATED REMOVE_INTENT REMOVE_MUTATED
TIMESTAMP_INTENT TIMESTAMP_MUTATED IDENTITY_INTENT IDENTITY_USER_REMOVED_MUTATED
IDENTITY_GROUP_REMOVED_MUTATED IDENTITY_HOME_REMOVED_MUTATED RELOAD_INTENT RELOAD_MUTATED VERIFY_DONE
CLEANUP_COMMITTED CLEANUP_STATE_MUTATED CLEANUP_BACKUPS_MUTATED CLEANUP_RECOVERY_STATE_MUTATED CLEANUP_RENAMED
```

`inject`の`ARMED`出力自体は試験PASSではありません。各phaseで`inject fail PHASE`を設定し、操作後の
rollback equality oracleが一致した場合だけQ21-18 PASSとして記録します。別snapshotで`inject kill PHASE`を
設定し、installer/admin/uninstallerがSIGKILLされた後にVMを強制rebootします。boot後は通常操作を
せず`/var/lib/codex-authority-recovery/recover`を実行する。self-cleanupでcanonical pathがrename済みなら
`/var/lib/codex-authority-recovery.completed/recover`を実行する。二回目も同じ結果になること、完全install
または完全uninstallの一方だけになること、root consoleと他ユーザーsudoが維持されることを確認します。
各fail caseは`rollback-compare`、各kill/reboot/recover caseはidentity種別を付けた
`crash-result created`または`crash-result reuse`を実行し、そのrunbook
出力だけをQ21-18/Q21-19 PASS証跡として返します。

## post-merge production bootstrap

implementation merge後のmain artifactでは、管理者マニュアルのroot-side
`gh attestation verify`を含むproduction bootstrapをfresh snapshotで実行し、install、verify、
uninstall、post-uninstallを再実行します。repository、signer workflow、main ref、merge commit、
subject digest、payload checksumのいずれかを変更したnegative verificationはすべて失敗が必要です。
