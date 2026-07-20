# TASK-0020 手動E2Eテスト手順書

この手順書は、エージェントがホストのセキュリティ境界を越えられない場合に、
人間のオペレーターがTASK-0020のリリース成果物canaryを実行するためのものです。
既存のレビュー済みrunbookを変更せず、そのまま呼び出します。

製品を永続インストールする手順ではありません。ローカルビルドやソースツリー内の
バイナリを代用してはいけません。

## 安全上の注意

- 対象は、使い捨て可能なLinux amd64テストホストです。
- オペレーターには、既存の非対話administrator権限が必要です。この手順でsudo
  policyを追加・拡張してはいけません。
- ホスト側で書き込む場所は
  `/var/tmp/codex-authority-task0020` だけです。fixtureの`/etc`、`/run`、
  `/usr/local`、ユーザー、policy、seed、ログ、sudo stateはprivate tmpfs内に
  作られます。
- runbook、引数、成果物を変更しないでください。canary内部のコマンドを抜き出して
  個別に実行してはいけません。
- 実行中に`kill -9`を使わないでください。`HUP`、`INT`、`TERM`ならcleanup trapが
  動きますが、`SIGKILL`では動きません。
- 失敗後にfixtureをその場で修理・再実行しないでください。必ず外側cleanupを完了し、
  clean stateから原因を調べます。
- OTP、seed、raw audit JSON、request/response payload、environment dump、保護された
  stderrを保存しないでください。保存してよいのはrunbookが出力する
  `case/result/count/digest`形式の行だけです。

固定入力は次のとおりです。

| 項目 | 固定値 |
| --- | --- |
| Repository | `autotaker/codex-authority-broker` |
| Actions run | `29720021660`、attempt 1、successful |
| Ref | `refs/heads/main` |
| Source commit | `09487b104f32cad23a695ec3f1a0c7e7a68e6163` |
| Artifact | `codex-authority-linux-amd64` |
| Archive SHA-256 | `5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd` |
| Stage runbook SHA-256 | `79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd` |
| Canary runbook SHA-256 | `4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7` |

一つでも一致しない場合は中止してください。別の成果物を試す場合は、新しいテスト
packetとして別途レビューする必要があります。

## 1. 実行前確認

一般ユーザーのterminalでrepositoryへ移動します。

```bash
cd /home/ubuntu/git/codex-authority-broker
test "$(uname -s)" = Linux
test "$(uname -m)" = x86_64
```

必要なコマンドと既存sudo権限を確認します。

```bash
for command in bash gh jq python3 sha256sum tar unshare findmnt mountpoint chroot setpriv sudo visudo systemd-analyze; do
  command -v "$command" >/dev/null || { echo "missing: $command"; exit 1; }
done
gh --version
gh attestation verify --help >/dev/null || {
  echo 'gh 2.49.0以降のattestation対応版が必要です'
  exit 1
}
sudo -n true
sudo -n -l
```

`sudo -n -l`の内容を自分で確認し、既存権限の範囲を理解できた場合だけ続行します。
`unknown command "attestation"`になる場合は`type -a gh`で古いCLIがPATHの先頭に
ないか確認し、GitHub CLIを2.49.0以降へ更新してください。attestation確認を省略して
先へ進んではいけません。

Ubuntu標準repositoryが`gh 2.45.x`を最新版として返す場合は、GitHub CLI公式apt
repositoryを追加します。repository keyを取得・検証してからinstallしてください。

```bash
sudo mkdir -p -m 755 /etc/apt/keyrings
KEYRING=$(mktemp)
wget -nv -O "$KEYRING" https://cli.github.com/packages/githubcli-archive-keyring.gpg
printf '%s  %s\n' 6084d5d7bd8e288441e0e94fc6275570895da18e6751f70f057485dc2d1a811b "$KEYRING" | sha256sum -c -
sudo install -m 0644 "$KEYRING" /etc/apt/keyrings/githubcli-archive-keyring.gpg
rm -f "$KEYRING"
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list >/dev/null
sudo apt update
apt-cache policy gh
sudo apt install gh
hash -r
gh --version
gh attestation verify --help >/dev/null
gh auth status
```

`apt-cache policy gh`のCandidateが2.49.0以上にならない場合はinstallを続けず、apt
sourceとkey errorを確認してください。

前回のfixtureが残っていないことを確認します。以下はすべて成功し、processやmountを
表示しない必要があります。

```bash
test ! -e /var/tmp/codex-authority-task0020
test ! -L /var/tmp/codex-authority-task0020
test ! -S /run/codex-authority.sock
! findmnt -rn -o TARGET | awk '$0=="/var/tmp/codex-authority-task0020" || index($0,"/var/tmp/codex-authority-task0020/")==1 {found=1} END{exit !found}'
! pgrep -af '/var/tmp/codex-authority-task0020|codex-fixture|codex-distinct'
```

runbookがレビュー済みの内容から変わっていないことを確認します。

```bash
printf '%s  %s\n' \
  79acb81eb39209c966fd183d0925667bb589d208f8f89134bb433fcea7f9e3dd \
  tasks/TASK-0020/STAGE_RUNBOOK.sh | sha256sum -c -
printf '%s  %s\n' \
  4321084a31719ab582a66e0fc1712e3d20685768f8f359c7b20dee40bd9bd5c7 \
  tasks/TASK-0020/CANARY_RUNBOOK.sh | sha256sum -c -
bash -n tasks/TASK-0020/STAGE_RUNBOOK.sh
bash -n tasks/TASK-0020/CANARY_RUNBOOK.sh
```

## 2. リリース成果物の取得と検証

stage runbookが受け付ける入力directoryは固定されています。

```bash
INPUT=/tmp/task0020-artifact-29720021660.AwGYdh
```

このdirectoryに検証済みdownloadが存在しない場合だけ、空のdirectoryを作り、固定
run IDから取得します。既存directoryを上書きしないでください。

```bash
test ! -e "$INPUT"
install -d -m 0700 "$INPUT"
gh run view 29720021660 \
  --repo autotaker/codex-authority-broker \
  --json status,conclusion,headBranch,headSha,workflowName,url
gh run download 29720021660 \
  --repo autotaker/codex-authority-broker \
  --name codex-authority-linux-amd64 \
  --dir "$INPUT"
```

表示結果がcompleted/success、`main`、固定source commit、release workflowであることを
確認します。次に、外側の2ファイル、archive digest、GitHub attestationを確認します。

```bash
test "$(find "$INPUT" -mindepth 1 -maxdepth 1 -printf '%y %f\n' | sort)" = "$(printf 'f %s\n' SHA256SUMS codex-authority-linux-amd64.tar.gz | sort)"
printf '%s  %s\n' \
  5ff05af201284c7581f1a4b9a2c3db5f5fd3102666644039cea56b8b8e4809dd \
  "$INPUT/codex-authority-linux-amd64.tar.gz" | sha256sum -c -
gh attestation verify "$INPUT/codex-authority-linux-amd64.tar.gz" \
  --repo autotaker/codex-authority-broker
```

attestationのrepository、release workflow、`refs/heads/main`、source commit、run、
subject digestが固定値と一致することを目視確認します。曖昧な場合は中止します。

hostへpayloadを展開せず、archive manifestと6つのchecksumを検証します。

```bash
EXPECTED_MANIFEST=$(printf '%s\n' \
  SHA256SUMS \
  bin/codex-authority \
  bin/codex-authority-broker \
  bin/codex-authority-sudo \
  deploy/pam/codex-authority \
  deploy/sudo/codex-authority \
  deploy/systemd/codex-authority-broker.service)
ARCHIVE="$INPUT/codex-authority-linux-amd64.tar.gz"
test "$(tar -tzf "$ARCHIVE" | sort)" = "$(printf '%s\n' "$EXPECTED_MANIFEST" | sort)"
test "$(tar -tvzf "$ARCHIVE" | awk '$1 !~ /^-/{bad=1} END{print bad+0}')" = 0
cmp -s <(tar -xOf "$ARCHIVE" SHA256SUMS) "$INPUT/SHA256SUMS"
test "$(wc -l <"$INPUT/SHA256SUMS")" = 6
test "$(awk '{print $2}' "$INPUT/SHA256SUMS" | sort)" = "$(printf '%s\n' "$EXPECTED_MANIFEST" | sed '/^SHA256SUMS$/d' | sort)"
while read -r expected member; do
  test "$member" != SHA256SUMS
  actual=$(tar -xOf "$ARCHIVE" "$member" | sha256sum | awk '{print $1}')
  test "$actual" = "$expected" || exit 1
done <"$INPUT/SHA256SUMS"
```

## 3. 外側stageの作成

次のレビュー済みコマンドだけを実行します。

```bash
sudo -n -- /bin/bash \
  /home/ubuntu/git/codex-authority-broker/tasks/TASK-0020/STAGE_RUNBOOK.sh setup
```

期待する出力は、次から始まる1行です。

```text
Q20-outer mode=setup result=PASS digest=
```

失敗または別の出力ならnamespaceへ入らないでください。stageが作成されている場合は、
section 6のcleanupへ進みます。

## 4. isolation/PAM preflight

引数やredirectを追加せず、正確に実行します。

```bash
sudo -n -- /usr/bin/unshare \
  --mount --pid --fork --kill-child --mount-proc=/proc \
  /var/tmp/codex-authority-task0020/CANARY_RUNBOOK.sh \
  preflight /var/tmp/codex-authority-task0020
```

exit code 0で、Q20-01、Q20-03、Q20-02、Q20-11のPASSが必要です。Q20-11は内部
fixtureのrollback完了を示します。FAILが一つでもあればcanaryを実行せず、section 6へ
進みます。

## 5. full canary

7分以上確保してください。実際の300秒lease expiryを待つため、途中の待機は正常です。
clock短縮、早期restart、timeout扱いをしてはいけません。

```bash
sudo -n -- /usr/bin/unshare \
  --mount --pid --fork --kill-child --mount-proc=/proc \
  /var/tmp/codex-authority-task0020/CANARY_RUNBOOK.sh \
  canary /var/tmp/codex-authority-task0020
```

成功時はexit code 0で、以下がすべてPASSになります。

- Q20-01: archive identityと6つのstreaming checksum
- Q20-02/Q20-03: real setuid/PAM preflight、private extraction
- Q20-04: artifact binaryとPAM/sudoers/systemd validation
- Q20-05: wrong-peer拒否とreal TOTP activation
- Q20-06: timestamp再利用なしの、独立した2回のreal sudo allow
- Q20-07: 自然なlease expiry後のfresh deny
- Q20-08: reactivation、broker停止時deny、新brokerの未activation deny、再activation
- Q20-09: admitted operationの厳密な5-field audit関係
- Q20-10: secret-freeな保護evidence処理
- Q20-11: 内側の完全rollback

case欠落、FAIL、nonzero exit、予期しないprompt、bounded形式以外の出力は失敗です。
cleanupと原因確認が終わるまで再実行しないでください。

## 6. 必ず外側cleanupを実行する

preflight/canaryの成功、失敗、中断にかかわらず実行します。

```bash
sudo -n -- /bin/bash \
  /home/ubuntu/git/codex-authority-broker/tasks/TASK-0020/STAGE_RUNBOOK.sh cleanup
```

期待する出力は次の1行です。

```text
Q20-outer mode=cleanup result=PASS digest=
```

cleanup digestはsetup digestと一致する必要があります。cleanupがstageを拒否した場合、
`rm -rf`、不明なmountの手動unmount、推測による削除を行わないでください。そのhostの
使用を止め、administratorと拒否理由を確認します。

## 7. cleanup後の残留確認

以下がすべて成功する必要があります。

```bash
test ! -e /var/tmp/codex-authority-task0020
test ! -L /var/tmp/codex-authority-task0020
test ! -S /run/codex-authority.sock
! findmnt -rn -o TARGET | awk '$0=="/var/tmp/codex-authority-task0020" || index($0,"/var/tmp/codex-authority-task0020/")==1 {found=1} END{exit !found}'
! pgrep -af '/var/tmp/codex-authority-task0020|codex-fixture|codex-distinct'
```

runbook自身も、hostのidentity、PAM、sudoers、`/run`、`/usr/local`、stage surfaceを
実行前後でbyte-for-byte比較します。簡易probeがcleanでも、Q20-11または外側cleanupの
比較がFAILならテストは失敗です。

## 8. 結果の記録

保存してよい項目は以下だけです。

- UTCの開始・終了時刻
- credentialではなく、通常のaccount名によるoperator identity
- run ID、source commit、archive/runbook digest
- 各processのexit code
- boundedな`case/result/count/digest`行
- 外側setup/cleanup digest
- cleanup後probeのPASS/FAIL

protected tmpfs fileを保存したり、その内容を再構成したりしないでください。

失敗分類:

- `permission_issue`: 必要な非対話administrator invocationを利用できない
- `environment_issue`: namespace、tmpfs、PAM/sudo、TOTP timing、process control、
  cleanup、完全rollbackを安全に実行できない
- `implementation_defect` / `regression`: fixtureとevidenceが正しいにもかかわらず、
  検証済みartifactが固定仕様に違反する
- `requirement_gap` / `qa_plan_defect`: 安全境界を弱めないと必須propertyを証明できない

すべてのcanary case、Q20-11、外側cleanup、setup/cleanup digest一致、cleanup後probeが
PASSした場合だけテスト全体をPASSとします。partial PASSはありません。
