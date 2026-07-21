# Codex Authority Broker 管理者マニュアル

## 対応環境

- Ubuntu 24.04 LTS amd64
- systemd、PAM、sudo 1.9.15系
- `qrencode`、`libcap2-bin`
- GitHub artifact attestationに対応した公式GitHub CLI
- console等、通常のsudo/PAMを経由しないroot復旧経路

本番専用ユーザー/グループは`coding-agent`です。root full sudoを持つため、専用hostまたは
十分に分離されたVMで使用してください。leaseはroot取得時の認証であり、root取得後の作用を
隔離・取消・巻戻しません。

## インストール前の準備

1. root復旧consoleへ実際に接続できることを確認します。
2. `/etc/sudoers`とinclude treeが`coding-agent`へ既存grantを持たず、global/command
   `!authenticate`、`exempt_group`、専用PAM serviceのoverrideを持たないことを確認します。
3. 既存`coding-agent`を再利用する場合、UID=GID!=0、primary group一致、supplementary groupなし、
   password locked、home `/var/lib/coding-agent`、shell `/bin/bash`、capability/grant/timestampなし、実行中
   process/login sessionなし、home外にそのUIDの所有fileなしを確認します。
4. 次を導入します。

```bash
sudo apt-get update
sudo apt-get install --yes qrencode libcap2-bin
```

Ubuntu archiveの2.45系`gh`には`attestation` commandがありません。GitHub CLI公式の
[Linux installation手順](https://github.com/cli/cli/blob/trunk/docs/install_linux.md)で公式apt
repositoryを追加し、attestation対応版を導入してください。repository key fingerprintは公式手順に
掲載された`2C6106201985B60E6C7AC87323F3D4EA75716059`または
`7F38BBB59D064DBCB3D84D725612B36462313325`と照合し、次が成功することを確認します。

```bash
gh attestation verify --help
```

## Artifactの検証

空の非特権directoryへ、mainの成功したrelease workflowから
`codex-authority-linux-amd64.tar.gz`と`codex-authority-bootstrap`を取得します。対象main commitを
`SOURCE_COMMIT`へ設定し、両方について
repository、workflow、ref、source commit、GitHub-hosted runnerを同時に検証します。

```bash
SOURCE_COMMIT='64文字のmain commit SHA'
for subject in codex-authority-linux-amd64.tar.gz codex-authority-bootstrap; do
  gh attestation verify "$subject" \
    --repo autotaker/codex-authority-broker \
    --signer-workflow autotaker/codex-authority-broker/.github/workflows/release.yml \
    --source-ref refs/heads/main \
    --source-digest "$SOURCE_COMMIT" \
    --deny-self-hosted-runners
done
```

`SOURCE_COMMIT`はGitHub上のmain workflow runと照合してください。verification outputをinstallerへ
pipeしません。失敗、複数候補、別ref/workflow/repository、mutable tagの場合は停止します。

## Root-owned stagingとインストール

次のcopy phaseはarchiveを実行・展開しません。

```bash
sudo /usr/bin/install -d -o root -g root -m 0700 /var/tmp/codex-authority-install
sudo /usr/bin/install -o root -g root -m 0600 \
  codex-authority-linux-amd64.tar.gz \
  /var/tmp/codex-authority-install/archive.tar.gz
sudo /usr/bin/install -o root -g root -m 0500 \
  codex-authority-bootstrap \
  /var/tmp/codex-authority-install/codex-authority-bootstrap
```

release workflowが生成したbootstrapにはmain commitとarchive SHA-256がliteralで固定されています。
bootstrapはroot-owned copyに対して両attestationを再検証し、archive digest、exact manifest、member type、
内側`SHA256SUMS`を検証してからだけ、固定pathへ展開してinstallerを実行します。
実行終了時は成功・失敗を問わず`/var/tmp/codex-authority-install`を削除します。

```bash
sudo /usr/bin/env -i PATH=/usr/sbin:/usr/bin:/sbin:/bin HOME=/root LANG=C \
  /var/tmp/codex-authority-install/codex-authority-bootstrap
```

QRは制御terminalだけへ一度表示されます。Authenticatorへ登録後、同じterminalで`ENROLLED`と
入力します。QR、URI、seed、TOTPをログ、画面収録、chat、issueへ保存しないでください。

## 検証

```bash
sudo /usr/local/sbin/codex-authority-verify
sudo systemctl status --no-pager codex-authority-broker.service
sudo visudo -cf /etc/sudoers
```

`coding-agent`でユーザーマニュアルのreadiness/TOTPを完了し、別processのsudoを2回実行します。
`sudo`、`sudo -i`、`sudo -A`、`sudo -s`、`sudo -u root`、`sudo -g root`のすべてが専用PAMを
通ることを受入試験で確認します。

## Rotation

root復旧consoleが利用可能な状態で実行します。

```bash
sudo /usr/local/sbin/codex-authority-admin
```

新QRの登録確認前に中断した場合、recoveryは旧seedを復元します。`ENROLLED`確認後は新seedを
正本とし、brokerをfresh startします。いずれも旧leaseは失われます。

## Recovery

install、rotation、uninstallが中断した場合、通常操作を再開せず次を実行します。

```bash
sudo /var/lib/codex-authority-recovery/recover
```

uninstallの最終self-cleanup中に電源断し、通常pathが存在せず
`/var/lib/codex-authority-recovery.completed/recover`だけが存在する場合は、その固定fallbackをroot
consoleから実行します。両方がないのにresidueがある場合は推測して削除しません。

stateがcorrupt、symlink、owner/mode不正の場合は推測復旧しません。root consoleからhostを隔離し、
artifact digestとstate metadataを確認してください。stateやseedの内容をissueへ貼らないでください。

## Uninstall

`coding-agent`のhomeに利用者ファイルがある場合、uninstallはidentity削除前に停止します。
必要なデータをサイト方針に従って退避・削除してから実行します。

```bash
sudo /usr/local/sbin/codex-authority-uninstall
```

installer-created identityだけを削除し、compatible pre-existing identityは保持します。system journalと
security auditはsite retention policyに従って残し、installerはvacuum、truncate、rewriteしません。

## Incident対応

broker/PAM/sudo異常時に`NOPASSWD`やglobal Defaultsで迂回しないでください。root復旧consoleから
brokerを停止し、必要ならnetworkを隔離してrecoveryまたはuninstallを行います。root権限で作成された
service、cron、setuid file、鍵、policy、processはlease期限後も残るため、通常のhost incident responseで
調査・除去します。
