# TASK-0021: production installer, rollback, and installation manuals

**Depends on:** TASK-0020 completed with manual canary, exact rollback evidence, and final v1 audit.

**Status:** planned goal; `executable:false` until the activation requirements below receive independent approval.

## Contract metadata

```json
{
  "id": "TASK-0021",
  "title": "production installer, rollback, and installation manuals",
  "status": "in_progress",
  "executable": true,
  "work_classification": "safety_contract_and_product_delivery",
  "depends_on": ["TASK-0020"],
  "baseline_production_sloc": 1478,
  "expected_production_sloc": null,
  "expected_cumulative_production_sloc": 1478,
  "target_cumulative_cap": 1500,
  "hard_cumulative_guard": 1800,
  "production_paths": ["install/codex-authority-install", "install/codex-authority-verify", "install/codex-authority-admin", "install/codex-authority-recover", "install/codex-authority-uninstall", "install/codex_authority_installer.py", "install/bootstrap/codex-authority-bootstrap.in", "deploy/pam/codex-authority", "deploy/sudo/codex-authority", "deploy/systemd/codex-authority-broker.service", ".github/workflows/release.yml", "docs/ADMIN_MANUAL.md", "docs/USER_MANUAL.md", "README.md"],
  "test_paths": ["install/tests/test_installer.py", "cmd/codex-authority-broker/main_test.go", "tasks/TASK-0021/BUILD_CANDIDATE.sh", "tasks/TASK-0021/MANUAL_E2E_TEST.md", "tasks/TASK-0021/E2E_RUNBOOK.sh"],
  "evidence_paths": ["tasks/TASK-0021/TASK.md", "tasks/TASK-0021/PLAN.md", "tasks/TASK-0021/QA_PLAN.md", "tasks/TASK-0021/PLAN_REVIEW.md", "backlog.json"],
  "goal": "Deliver a supported, transactional production installation path that verifies the official artifact, provisions the tool-neutral dedicated OS identity coding-agent and TOTP enrollment, integrates PAM with dedicated-identity full sudo gated by a fresh live lease on every invocation, installs and starts the broker, validates live behavior, and provides deterministic rollback, uninstall, administrator documentation, and an end-user manual.",
  "activation_requirements": ["select and document an explicitly supported OS/version and PAM/sudo model", "use coding-agent as the canonical production OS user and group name; keep product- or vendor-named identities confined to disposable fixtures", "design a PAM integration that never blindly replaces the host sudo stack and preserves an independently tested root recovery path", "define an explicit root full-sudo grant for coding-agent only, with normal, login, askpass, shell, and runas forms all routed through the Codex Authority PAM service and timestamp caching disabled; never claim the lease contains root effects after acquisition", "define secret-safe enrollment, rotation, and root-owned seed creation without argv/environment/log disclosure", "define the installer and verifier trust chain: pinned distribution, checksum/provenance verification before elevation, a copy-only fixed privileged step into root-owned non-writable staging, then a separate fixed root re-attest/extract/hash/execute step with fixed paths, argv, and environment", "freeze installer, verifier, rollback, uninstall, ownership, and durable recovery-state paths plus a disposable reboot-capable destructive-test fixture", "separate controlled error/signal automatic rollback from journaled recovery after SIGKILL, power loss, kernel failure, or reboot", "measure installer-family SLOC for visibility and readability without applying the 1500/1800 production budget; preserve no-compression review and keep every existing runtime change inside the budget", "approve TASK, PLAN, TASK-first QA_PLAN, and independent plan review"],
  "completion_requires": ["idempotent fail-closed install and verify flow for the supported platform", "automatic rollback for every injected controlled error and catchable signal plus deterministic resume, rollback, or uninstall after every tested abrupt interruption and reboot point", "independently runnable uninstall that restores captured shared-file pre-state, preserves compatible pre-existing identities, and leaves no installer-owned process, socket, identity, policy, seed, binary, unit, timestamp, temporary file, secret-bearing log, or recovery state", "real disposable-host E2E covering clean enrollment, install, compatible identity reuse, UID/GID/name collision rejection, dedicated-identity full sudo for arbitrary root commands only during a fresh live authorization check, two uncached sudo calls, expiry, secret rotation with old secret/code rejection and old lease invalidation, broker stop/restart, controlled failed-install rollback, abrupt interruption and reboot recovery at every declared crash point, reinstall, and uninstall", "artifact and installer provenance/checksum verification before privileged execution or writes", "shared system journal and security audit remain unmodified and follow site retention policy while installer-owned temporary and secret-bearing logs are removed", "administrator installation/recovery/uninstall manual and updated end-user manual that explicitly state root-acquired persistence can outlive the lease", "independent REVIEW and QA PASS with secret-free evidence"],
  "fixture_elevation_needs": "Disposable supported-platform VM or equivalent isolated host with snapshot/restore and scripted reboot/power-loss testing, console or other out-of-band root recovery, real systemd/PAM/sudo, network only for pinned artifact/provenance retrieval, and authority to create/remove the dedicated identity and exact declared files. Never develop or first-run the installer against a workstation or irreplaceable host.",
  "sloc_budget_exception": "Installer, verifier, rollback, recovery, enrollment, rotation, and uninstall implementation is measured but excluded from the 1500 target and 1800 hard production SLOC budget by explicit owner decision. Existing runtime/product changes remain budgeted.",
  "exclusions": ["blind replacement of /etc/pam.d/sudo", "production OS user or group names tied to Codex, a specific coding tool, or a vendor", "sudo grant for any identity other than coding-agent", "global sudo defaults that affect other users", "NOPASSWD or any route that bypasses the live PAM lease check", "claiming that lease expiry reverses or contains effects produced after root acquisition", "secret in argv, environment, shell history, logs, evidence, or repository", "manual-only steps presented as a complete installer", "silent mutation without captured pre-state and rollback", "GitHub push authority", "multi-distribution support before one platform is fully accepted"],
  "blocked_reason": "The repository has no trusted production installer distribution path, supported-host dedicated-identity full-sudo PAM contract, enrollment/rotation tool, crash-recovery model, identity-lifecycle proof, or production uninstall proof. Scope and installer trust boundaries must be planned before executable work.",
  "split_stop_rule": "Do not enable DEV or mutate a live host until every activation requirement is approved. Stop and split if dedicated-only full sudo cannot be routed exclusively through a fresh live PAM lease check with timestamp caching disabled, or if out-of-band recovery, exact rollback, secret-safe enrollment, or crash recovery cannot be proven. Never weaken fail-closed behavior, hide installer size, compress code to meet a non-applicable budget, or move existing runtime changes outside the production SLOC guard.",
  "contract_path": "tasks/TASK-0021/TASK.md"
}
```

## ゴール

ユーザーが検証済みの公式artifactから、安全にインストール、初期登録、利用、復旧、アンインストールできる本番導入経路を提供する。単にバイナリや設定fragmentをコピーする手順ではなく、部分失敗時の自動rollbackと、導入前状態へ戻す独立uninstallまでを製品として扱う。

完成時には、少なくとも次を提供する。

- 対応OS/versionを明示したinstaller、preflight、post-install verifier
- 公式artifactのchecksumとprovenanceを特権書き込み前に検証する経路
- installer/verifier自体を非特権で真正性確認し、root-owned stagingから固定path/argvだけを最小限の権限移行で実行するtrust chain
- ツールやベンダー名に依存しない専用OSユーザー/グループ`coding-agent`を、非zero・同一UID/GIDで衝突なく作成または安全に再利用する経路
- TOTP enrollment、root-owned mode `0600` seed生成、rotationを秘密非開示で行う経路
- 既存sudo/PAMを盲目的に置換せず、root復旧経路を維持する統合方式
- `coding-agent`だけにfull sudoを付与し、毎回の実行をfreshなlive lease確認へ通し、sudo timestamp cacheを無効化する統合方式
- systemd service、socket、audit、no-cacheの検証
- 部分失敗時の自動rollback、再install、独立uninstall
- SIGKILL、電源断、kernel failure、reboot後に再開・rollback・uninstallできるdurable recovery state
- 管理者向けinstallation/recovery/uninstall manualとユーザーマニュアル

## 実行可能化の条件

このタスクは、次をPLANとTASK-first QA_PLANで固定し、独立レビューがPASSするまでDEVを開始しない。

1. 対応するOS/versionと、そのOSにおけるsudo/PAM integration model。
2. 既存のsudo認証を破壊せず、console等の独立root復旧経路を事前確認する方式。
3. `coding-agent`だけにfull sudoを付与し、すべてのsudo呼び出しをCodex Authority PAM serviceへ通し、timestamp cacheを無効化する方式。leaseはroot取得の入口を制御するだけで、取得後の作用を封じたり期限後に巻き戻したりしない。
4. secretをargv、environment、history、log、証跡へ出さないenrollment/rotation方式。
5. installer/verifier自身のpinned distribution、checksum/provenance、root-owned staging、固定entrypoint/argv、最小root transition。
6. install前状態、各変更のownership、shared journal/auditの非改変、失敗注入点、rollback順序、durable recovery state、uninstall後のabsence/equality oracle。
7. snapshot/restoreとreboot/power-loss試験が可能な使い捨てhostでの破壊的E2E fixture。
8. installer/verifier/rollback/recovery/enrollment/rotation/uninstallのSLOCを可読性確認用に実測・報告する方式。これらは1500/1800 budgetから除外するが、既存broker、CLI、PAM helper、IPC、lease、audit runtimeの変更は引き続きbudget対象とする。

## 完了条件

- install、同一入力での再実行、制御可能な失敗の自動rollback、異常中断・reboot後の復旧、再install、uninstallが決定的にPASSする。
- installerの各制御可能な失敗注入点で自動rollbackし、各SIGKILL/電源断/reboot pointからjournaled stateを使って安全に再開、rollback、またはuninstallできる。
- clean enrollmentとsecret rotationを実行し、rotation後は旧secret/codeと旧leaseを拒否する。
- compatibleな既存identityは再利用・uninstall時保持し、UID/GID/name collisionはhost mutation前に拒否する。
- 実systemd/PAM/sudo hostで、TOTP activation、独立した2回のno-cache sudo、自然期限切れ、broker停止、fresh restart denialを確認する。
- lease中に`coding-agent`が任意のroot commandを実行でき、leaseなし・期限切れ・broker停止時には新しいsudo呼び出しを拒否することを確認する。root取得後に作成したservice、cron、setuid file、鍵、policy変更等はlease期限後も残り得るため、leaseをsandboxや自動rollbackとして扱わない。
- uninstall後、installerが所有したprocess、socket、identity、policy、seed、binary、unit、timestamp、temporary file、secret-bearing log、recovery stateが残らず、共有ファイルはcaptured pre-stateへ復元される。shared system journalとsecurity auditはsite retention policyに従って保持し、installerは改変しない。
- 管理者とユーザーがraw seed/TOTPを表示・保存せず、enrollment QRを制御terminalだけで一度扱って導入・利用・復旧できる文書が完成する。
- 独立REVIEWとQAが、実装、fixture、rollback、秘密境界、文書を同一candidateでPASSする。

## DEV measurement

installer-familyは、実装候補で物理1797行、空行と先頭`#` commentを除く可視性測定1646行。
明示合意どおり1500/1800 production budgetの対象外だが、reviewから隠さず全体を同一candidateで扱う。
既存runtimeのproduction SLOCは1478のままである。

## 禁止事項

- `/etc/pam.d/sudo`の盲目的な置換
- `coding-agent`以外へのsudo grant、他ユーザーへ影響するglobal sudo default、`NOPASSWD`、またはlive PAM lease確認を迂回する設定
- 本番OSユーザー/グループへCodex、特定coding tool、vendorに依存する名前を使うこと
- seed/TOTPを引数、environment、shell history、log、証跡、repositoryへ記録
- pre-stateを取得しないhost mutation
- workstationや代替不能hostでの初回実行
- manual stepsだけをinstaller完成として扱うこと
- GitHub push authorityをこのタスクへ混在させること
