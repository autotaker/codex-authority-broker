# PLAN — TASK-0021: 本番インストーラー、復旧、アンインストール

## 対象と境界

対応対象は **Ubuntu 24.04 LTS amd64** の通常のsystemd/PAM/sudo構成に限定する。
本番専用OSユーザー/グループは `coding-agent` とし、UIDとGIDは同じ非zero値とする。
既存の `/etc/pam.d/sudo` は変更せず、ユーザー限定sudoers Defaultsの
`pam_service`、`pam_login_service`、`pam_askpass_service`をすべて
`codex-authority`へ固定する。Ubuntu 24.04のsudo 1.9.15系はこれらをサポートする。
`coding-agent`には `(root:root) PASSWD: ALL` のroot full sudoを付与し、
`NOPASSWD`は使用しない。`timestamp_timeout=0`により各sudo processがPAM helperを実行する。
`sudo -i`、`sudo -A`、`sudo -s`、`sudo -u root`、`sudo -g root`も同じPAMへ通す。
root以外のrunas指定はroot取得後に実現できるためsudoersでは許可しない。

preflightはlocal sudoers pluginだけを許し、LDAP等の追加policy source、`exempt_group`、
global/user/host/command Defaultsの`!authenticate`、coding-agentへ一致する既存
`NOPASSWD`/grant、専用PAM serviceを後勝ちで上書きするDefaultsを拒否する。
install後は`sudo -ll -U coding-agent`と全include fileの正規化検査を併用し、
有効policyがcandidateのroot-only rule一件と専用Defaultsだけであることを要求する。

leaseはroot取得の入口だけを制御する。root取得後に作成されたservice、cron、
setuid file、鍵、policy、process等は期限後も残り得る。installer、manual、testは
containmentや自動巻戻しを主張しない。

## 配布とtrust chain

release artifactへ次を追加し、内側の`SHA256SUMS`で全payloadを列挙する。

- `install/codex-authority-install`
- `install/codex-authority-verify`
- `install/codex-authority-admin`
- `install/codex-authority-recover`
- `install/codex-authority-uninstall`
- `install/codex_authority_installer.py`
- `docs/ADMIN_MANUAL.md`
- `docs/USER_MANUAL.md`

利用者は非特権の空ディレクトリへartifactを取得し、公式GitHub CLIのattestation機能で
repository `autotaker/codex-authority-broker`、signer workflow
`.github/workflows/release.yml`、`refs/heads/main`、source commit、artifact subject digest、
GitHub-hosted runnerを固定して検証する。`gh`はGitHub公式apt repositoryから取得した
attestation subcommand搭載versionを使い、package source/key/fingerprint/versionを記録する。

productionでは、検証済みarchiveを第一の固定privileged stepでhostのtrusted `/usr/bin/install`だけを使い、
fixed path `/var/tmp/codex-authority-install/archive.tar.gz`へroot:root 0600でcopyする。
このstepはarchiveを実行・展開しない。第二の固定stepはrootの`gh attestation verify`で
root-owned copyを同じworkflow/ref/source/digestへ再bindし、trusted `/usr/bin/tar`で
fixed root-owned stagingへ安全に展開し、exact manifest/checksumを検証してから
`/usr/bin/python3 -I`でfixed installerを実行するroot-owned bootstrap runbookとする。
runbook自体はrepositoryから直接実行せず、管理者が内容とdigestを事前固定した短い
host-side bootstrapとする。環境は`env -i`の固定`PATH`, `HOME`, `LANG`だけ、argvとdestinationは固定する。
二つのprivileged step、bootstrap digest、source FD/path、copy→attest→extract→hash→exec順序を
管理者manualへexact commandとして記載する。mutable tag、source checkout、未検証copyは使わない。

pre-merge functional E2Eだけはproduction bootstrapと分離したfixture modeを使う。fixture runbookへ
candidate commit/tree/local archive SHA-256をliteral固定し、copy-only step後にroot側でdigest、
exact manifest、payload checksumを再検証して実行する。このphaseは`gh attestation verify`を行わず、
provenance PASSを主張しない。post-mergeにはmain artifactを使い、fresh snapshot上でproductionの
root re-attest bootstrapによるinstall/verify/uninstallを再実行する。Q21-04はfixture phaseと
post-merge production phaseの両方がPASSして初めてfinal PASSとする。

## installer構造

実装はshell/Pythonのinstaller familyに限定し、既存Go runtimeを変更しない。
installer familyのSLOCは個別に実測するが1500/1800 budgetへ加算しない。
可読性、shellcheck相当の静的検査、失敗処理、no-compression reviewは維持する。

固定配置先は次のとおり。

- binaries: `/usr/local/bin/codex-authority*`
- admin tool: `/usr/local/sbin/codex-authority-admin`
- PAM: `/etc/pam.d/codex-authority`
- sudoers: `/etc/sudoers.d/codex-authority`
- unit: `/etc/systemd/system/codex-authority-broker.service`
- seed: `/etc/codex-authority/seed.json`
- durable state: `/var/lib/codex-authority-installer/state.json`
- installed recovery/uninstall entrypoints: `/usr/local/sbin/codex-authority-recover`、`/usr/local/sbin/codex-authority-uninstall`
- lifetime recovery: `/var/lib/codex-authority-recovery/recover` と `/var/lib/codex-authority-recovery/state.json`

preflightはEUID、OS/version、architecture、必要command、PAM module、sudo version、
sudoers include、filesystem type/owner/mode、path collision、identity collision、
root復旧確認フラグをhost mutation前に検査する。`coding-agent`が存在する場合は、
同名group、同一非zeroUID/GID、primary group、非特権identityという完全一致時だけ再利用し、
それ以外は拒否する。存在しない場合だけsystem identityを作り、所有権をjournalへ記録する。

既存の管理対象pathは、同じinstaller stateが所有するidempotent reinstall以外では上書きしない。
共有のsudo/PAMファイルは変更しない。新規ファイルは同一filesystem内のtemporary fileへ
安全なmode/ownerで書き、検証後にrenameする。sudoersは`visudo -cf`、unitは
`systemd-analyze verify`、seed schemaとfile metadataは専用verifierで検査する。

## transaction、crash recovery、uninstall

`/var/lib/codex-authority-installer`はroot:root 0700、`state.json`はroot:root 0600の
regular fileとする。schemaはexact keys `version`, `operation`, `phase`, `artifact_digest`,
`identity`, `pre_state`, `backup_paths`, `owned_paths`, `pending_path`, `completed_steps`,
`seed_generation` とし、unknown、
duplicate、trailing、wrong type/path、symlink、owner/mode不一致を拒否する。state temporaryは
同directoryの固定prefixでopenat `O_CREAT|O_EXCL|O_NOFOLLOW`し、file fsync→rename→parent fsyncする。

install phaseは `PREPARED`, `IDENTITY_INTENT`, `IDENTITY_DONE`, 各pathごとの
`FILE_INTENT`/`FILE_DONE`, `SEED_INTENT`/`SEED_DONE`, `POLICY_INTENT`/`POLICY_DONE`,
`START_INTENT`/`START_DONE`, `COMMITTED` とする。intentをfile+parent fsyncした後だけmutationし、
mutation verification後だけdoneをfsyncする。journal directory作成後state未作成、truncated temp、
intent前、intent後mutation前、mutation後done前、done後の全gapへ一意なrollback oracleを定義する。
通常error/catch可能signalはreverse rollback、SIGKILL/kernel failure/電源断/reboot後は
`codex-authority-recover`だけを許し、未完installは常にidempotent rollbackする。
unknown/corrupt stateでは推測削除せずfail closedとする。

artifact manifestに含む`install/codex-authority-recover`を最初のmutation前に
`/var/lib/codex-authority-recovery/recover`へroot:root 0500でatomic install+fsyncし、directoryを
root:root 0700でinstallation lifetime中保持する。installed recovery entrypointが未配置・破損・
uninstall中に削除済みでも、このdurable copyをartifact checksumへ再検証してrecoverできる。

uninstall operationは `UNINSTALL_PREPARED`, `STOP_INTENT`/`STOP_DONE`, 各owned pathの
`REMOVE_INTENT`/`REMOVE_DONE`, `TIMESTAMP_INTENT`/`TIMESTAMP_DONE`,
`IDENTITY_INTENT`/`IDENTITY_DONE`, `RELOAD_INTENT`/`RELOAD_DONE`, `VERIFY_DONE`,
`COMMITTED`を同じdurable rulesで記録する。固定installed entrypointでrunning serviceを停止し、socketを除去し、
installer-owned fileだけをreverse orderで削除する。installerが作ったidentityだけを、
process・file ownership・login sessionがないことを確認して削除する。再利用identityは保持する。
system journalとsecurity auditはsite retention policyに従い、削除・改変しない。
最後にdaemon-reloadとabsence oracleを実行し、成功時だけdurable stateを削除する。
installed recover/uninstall/coreと通常stateは最後まで残す。uninstall開始前にlifetime recovery
directoryへuninstall stateとartifact digestをfsyncし、以後のrecovery sourceとする。
`VERIFY_DONE`後、lifetime recoveryが通常stateとinstalled toolsを削除して`COMMITTED`を自directoryへ
fsyncする。COMMITTED後のself-cleanupはauthority stateを変更しないidempotent residue cleanupとし、
crash時はartifact bootstrapまたは残存recoverを再実行してdirectoryを除去する。

compatible identity reuseは、user/group同名、同じ非zero UID/GID、primary group一致、supplementary
groupなし、password locked、home `/var/lib/coding-agent`、shell `/bin/bash`であり、home内容を変更せず、
Linux/file capability、sudo grant、sudo timestampを持たないidentityだけを許す。
存在する場合はpre-state保持を証明できないためmutation前に拒否する。absence、identity metadata、
unit active/enabled、全managed path状態を`pre_state`へ記録する。rotation旧seed backupは
`/var/lib/codex-authority-installer/backups/seed.previous` root:root 0600、digestはstateへ記録する。
install後に生成されたcoding-agent timestampはbaseline absenceとUIDによりtransaction ownershipを証明し、
uninstall時に除去する。他identity timestampへは触れない。

## enrollmentとrotation

admin toolはPython標準libraryのCSPRNGで20-byte secretを生成し、base64 seed JSONを
root-owned mode 0600でatomic作成する。TOTP secret/URIはargv、environment、shell history、
journal、証跡へ渡さない。`qrencode`へstdinだけでotpauth URIを渡し、管理者の制御terminalへ
QRを一回表示する。非terminal、pipeされたstdout、`qrencode`不在ではsecret作成前に拒否する。

rotationは同じdurable journalのoperation `rotate`を使う。root-only backupへ旧seedをcopy+fsyncし、
`ROTATE_PREPARED`, `BROKER_STOPPED`, `NEW_SEED_INTENT`, `NEW_SEED_PENDING_ACK`,
`ENROLLMENT_ACKED`, `RESTART_INTENT`, `RESTART_DONE`, `COMMITTED`をdurable記録する。
QR表示後、管理者がcontrolling terminalで明示ackするまでは旧seedへrollback可能な状態を保つ。
ack前のcrash/rebootは旧seed復元+旧broker restart、ack後はnew seedを保持してbroker restartし、
verify失敗ならfail closedでrecoveryを要求する。backupはcommit後だけsecure removalする。
各phaseのSIGKILL/reboot recoveryを独立testする。broker restartで旧leaseは失われ、旧codeは拒否される。

## DEV scope

- `install/` のinstaller、verifier、admin、uninstallerと共通library
- `install/tests/` のstatic、transaction、failure-injection、fixture tests
- `install/bootstrap/` の固定copy/bootstrap/recovery runbook
- `deploy/pam/codex-authority`、`deploy/sudo/codex-authority`、systemd unit
- `.github/workflows/release.yml` と既存workflow/package test
- `docs/ADMIN_MANUAL.md`、`docs/USER_MANUAL.md`、`README.md`
- `tasks/TASK-0021/MANUAL_E2E_TEST.md`とdigest-bound operator runbook
- TASK-0021の計画、レビュー、QA、完了証跡と`backlog.json`

既存broker/CLI/PAM helper/IPC/lease/audit runtimeの変更が必要になった場合は停止して分割する。

## 検証と完了順序

1. TASK-first QA_PLANとこのPLANを独立reviewする。
2. DEV後、shell/Python syntax、Go tests、format、JSON、manifest、secret scanを実行する。
3. pre-merge candidateからrelease workflowと同じ手順でdeterministic local archiveを生成し、
   candidate commit/tree/archive/runbook digestへbindする。これはfunctional E2E入力であり、
   GitHub provenance PASSとは主張しない。ユーザーが用意する、snapshot/restoreとout-of-band root consoleを持つUbuntu 24.04
   disposable VMで、Mainが固定した`MANUAL_E2E_TEST.md`とdigest-bound runbookだけを
   ユーザーが実行する。clean install、reinstall、identity reuse/collision、全sudo form、
   自然expiry、broker stop/restart、rotation、controlled failure rollback、各declared crash
   pointのSIGKILL/reboot recovery、uninstallを行う。VM endpoint、credential、secret、raw TOTPは
   repository、chat、evidenceへ渡さず、ユーザーは固定grammarのPASS/FAIL/count/digestだけを返す。
4. retained evidenceはcase ID、result、count、digestだけとし、secret、URI、raw command outputを残さない。
5. 同一candidateを独立REVIEWし、static QAとfunctional E2EをPASSする。Q21-03のpost-merge
   GitHub provenance portionとQ21-04のpost-merge production-bootstrap phaseだけはpost-merge
   operational gateとしてpendingを許し、他のcaseを代替しない。
6. Mainがimplementation candidateをmainへmergeする。main-bound release workflowのartifactを、
   repository、signer workflow、main ref、merge source commit、subject digestへ再bindして検証する。
   pre-merge functional archiveとpayload checksumが一致することも要求する。
7. 独立QAがpost-merge provenanceを含むQ21-01〜27を最終PASSした後、MainがTASK completedと
   evidenceだけの最終commitをmainへ追加する。fixture/evidence未完了ならimplementation mergeもしない。

実ホストや代替不能hostでの初回実行、root復旧経路のないlive mutation、test failureのwaive、
部分成功、秘密を含む証跡は完了根拠にできない。
