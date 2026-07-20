# Codex Authority Broker v1 ユーザーマニュアル

## このマニュアルについて

Codex Authority Brokerは、ツール非依存の専用OSユーザー`coding-agent`に対して、TOTPで有効化する5分間のauthority leaseを提供します。sudoを実行するたびにPAM helperがbrokerへ問い合わせ、leaseがその時点で有効な場合だけ処理を許可します。

この文書は、サイト管理者が導入とsudo/PAM設定を完了した環境を利用するユーザー向けです。v1には本番インストーラー、PAM統合ツール、sudoコマンドgrant、アンインストーラーは含まれていません。

> [!WARNING]
> `tasks/TASK-0020`の手順は、隔離されたtmpfs fixtureで製品を検証するためのE2E手順です。本番ホストへインストールする手順ではありません。実ホストの`/etc/pam.d/sudo`へそのまま適用しないでください。

## v1でできること

- TOTPを使って300秒間のleaseを有効化する
- sudoを実行するたびに、キャッシュを使わずbrokerへ再確認する
- lease期限切れ、broker停止、broker再起動、通信異常時にfail closedで拒否する
- allow/denyを、秘密情報を含まない固定形式のaudit eventとして記録する

v1ではGitHub push権限を扱いません。また、leaseの状態表示、残り時間表示、ユーザー操作による即時revokeはありません。

## 利用開始前の確認

サイト管理者から、次の事項がすべて完了していることを確認してください。

- Linux amd64ホストへ、検証済みの`codex-authority`、`codex-authority-broker`、`codex-authority-sudo`が配置されている
- rootでbroker serviceが起動している
- 自分が専用OSユーザー`coding-agent`として登録され、数値UIDとGIDが同じ非zero値になっている
- `/run/codex-authority.sock`のowner UID、group GIDがその専用ユーザーと一致し、modeが`0660`になっている
- Authenticatorへ、管理者がbrokerと同じTOTP enrollmentを安全に登録している
- full sudo grantが`coding-agent`だけに限定され、すべてのsudo呼び出しがCodex Authority PAM serviceを通ることがセキュリティレビュー済みである
- sudo timestamp cacheが`coding-agent`に対して無効化されている
- 緊急時に使える、Codex Authority Brokerを経由しない管理者復旧経路がある

同梱の`deploy/sudo/codex-authority`にある`codex-fixture`は、隔離E2E fixtureだけで使うテスト名です。本番ユーザー名ではありません。このfragmentは`timestamp_timeout=0`だけを設定し、sudo commandを許可しません。TASK-0021で計画している本番インストーラーは、`coding-agent`だけを対象とするfull sudo grantとlive PAM checkを一体で導入します。

一般ユーザーで次を確認できます。

```bash
test -x /usr/local/bin/codex-authority
test -S /run/codex-authority.sock
id
stat -c 'socket owner_uid=%u group_gid=%g mode=%a' /run/codex-authority.sock
```

UID/GIDやsocket metadataが管理者から示された値と異なる場合は、操作を続けないでください。rootや別ユーザーからのCLI接続は、意図的に拒否されます。

## Leaseを有効化する

### 1. broker再起動直後はAuthenticatorの表示更新を待つ

brokerは起動時点の30秒TOTP windowをreplay防止のため受け付けません。brokerの起動・再起動直後は、Authenticatorに表示される6桁コードが少なくとも一度切り替わってから操作してください。

### 2. readinessを開始する

`coding-agent`のshellで実行します。

```bash
/usr/local/bin/codex-authority ready
```

成功時は次の1行が表示されます。

```text
ready accepted
```

readinessの有効時間は開始から300秒です。すでにleaseが有効な場合、`ready`は拒否されます。

### 3. TOTPを入力する

TOTPをコマンドライン引数、environment、shell historyへ入れてはいけません。Bashでは、次の関数をそのまま定義して実行できます。

```bash
activate_codex_authority() {
  local code rc
  IFS= read -r -s -p 'TOTP: ' code
  printf '\n'
  builtin printf '%s\n' "$code" | /usr/local/bin/codex-authority otp
  rc=$?
  unset code
  return "$rc"
}

activate_codex_authority
unset -f activate_codex_authority
```

入力した6桁コードは画面へ表示されません。成功時は次の1行が表示されます。

```text
otp accepted
```

成功時点から300秒間のleaseが始まります。期限は延長されません。leaseが有効な間の再認証や、同じTOTPの再利用は拒否されます。

TOTP入力は60秒間に5回までです。繰り返し失敗した場合は総当たりをせず、後述の確認を行ってください。

## sudoを使う

本番インストール完了後は、lease有効中に任意のroot commandを通常どおり実行できます。

```bash
sudo -- /usr/bin/id -u
```

成功時、上の例は`0`を表示します。`coding-agent`にはfull sudoが付与されるため、実行するcommand自体のallowlistはありません。

> [!CAUTION]
> authority leaseはroot権限を新たに取得する入口だけを制御します。root取得後に作成したservice、cron、setuid file、SSH key、sudo/PAM policy変更、常駐processなどは300秒後も残り得ます。leaseはsandboxではなく、期限切れ時に作用を巻き戻す仕組みでもありません。

各sudo processは、固定socketに対して1回だけ新しいauthorize要求を送ります。次の状態では、以前の成功に関係なく拒否されます。

- leaseの300秒期限を過ぎた
- brokerが停止している、応答しない、または再起動された
- socketの型、owner、group、配置が安全条件を満たさない
- PAM helperが専用UID/GIDへ安全に権限dropできない
- broker応答が不正、期限切れ、またはdenyである
- audit eventを完全に記録できない

続けてsudoを実行しても、前回の結果は再利用されません。

## Leaseを終了する

通常は、認証成功から300秒後に自動終了します。ユーザー向けの`revoke`コマンドはありません。

期限前に無効化する必要がある場合は、サイト管理者へbrokerの停止または再起動を依頼してください。brokerの停止・再起動でprocess-localなleaseは失われ、再起動後は新しいTOTP windowで再認証するまでsudoが拒否されます。

## エラー時の確認

CLIは内部情報を開示せず、失敗時は原則として次だけを表示します。

```text
request denied
```

### `ready`が拒否される

- broker serviceが起動しているか管理者へ確認する
- 専用OSユーザーで実行しているか確認する
- socketのUID/GIDが自分の数値UID/GIDと一致するか確認する
- すでにleaseが有効なら、期限切れまで待つ
- stale socketや不正なsocketを自分で削除しない

### `otp`が拒否される

- 先に`ready`が成功しているか確認する
- `ready`から300秒以内か確認する
- broker起動後にAuthenticatorのコードが一度更新されたか確認する
- ホストとAuthenticatorの時刻が同期しているか確認する
- 6桁コードの前後に空白が入っていないか確認する
- 同じコードを再利用していないか確認する
- 5回失敗した場合は少なくとも60秒待ち、原因を確認してからやり直す

### sudoが拒否される

- leaseの開始から300秒以内か確認する
- `coding-agent`だけにfull sudoが設定され、live PAM checkとtimestamp無効化が維持されているか管理者へ確認する
- broker停止・再起動がなかったか管理者へ確認する
- `ready`とTOTP認証を新しいwindowでやり直す

denyを回避するためにPAM、sudoers、socket権限、systemd unitをユーザー自身で変更しないでください。fail closedは正常な安全動作です。

## Audit log

brokerは、backendへ到達した各操作について次の5フィールドだけをJSON 1行で記録します。

- `correlation_id`
- `actor_uid`
- `scope`（`ready`、`otp`、`authorize`）
- `result`（`allow`または`deny`）
- `lease_expiry`

seed、TOTP、request payload、credential、environment、内部エラーは記録しません。systemd環境では通常journalへ送られます。閲覧権限と保存期間はサイト管理者が管理します。

wrong peerの接続など、backendへ到達する前に拒否された処理にはaudit eventがありません。audit書き込みに失敗した場合、brokerは安全側に閉じて以後のauthority要求を拒否します。

## セキュリティ上の禁止事項

- TOTPを`codex-authority otp 123456`のような引数へ入れない
- TOTPを`echo 123456 | ...`としてshell historyへ残さない
- seed、TOTP、raw audit、environment dumpをissue、チャット、ログへ貼らない
- `/etc/codex-authority/seed.json`を読み出したりコピーしたりしない
- `/run/codex-authority.sock`を別ユーザーへchown/chmodしない
- PAMやsudoersの拒否を、`NOPASSWD`、他ユーザーへのgrant、global default、またはlive lease確認の迂回で回避しない
- lease中に取得したroot権限で永続化を行う場合は、lease期限後にも影響が残ることを理解し、サイトの管理・監査方針に従う
- broker障害時に古いsocketを確認せず削除しない
- TASK-0020の隔離fixture用設定を本番ホストへコピーしない

## 管理者に伝えるv1の制限

次の機能はv1に含まれていません。

- 本番インストーラーとアンインストーラー
- OS/distribution別のPAM統合
- `coding-agent`専用full sudo grantの作成
- 専用UID/GIDの割当て
- TOTP enrollment、QR生成、secret rotation
- leaseのstatus、残り時間表示、ユーザーrevoke
- audit viewer、export、rotation
- stale socketの自動復旧
- GitHub push authority

本番導入では、サイト管理者が既存のroot復旧経路を維持したまま、`coding-agent`専用full sudo、毎回のlive PAM check、timestamp cache無効化を一体として設計・レビュー・rollback検証する必要があります。製品に同梱されたPAM/sudo fragmentsだけでは、本番sudo環境は完成しません。
