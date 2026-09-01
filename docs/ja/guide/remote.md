# リモート軸

ユーザーとプロセスの間にある層を表し、他の環境変数が保持・除去される理由も示します。

```go
result.IsRemote()
result.Remote(runby.RemoteTmux)
result.Multiplexer()
```

複数の層を同時に検出でき、順序はネスト順ではなく検出順です。

| 層 | `RemotePlatform` | `Kind` | マーカー |
|---|---|---|---|
| tmux | `RemoteTmux` | `multiplexer` | `TMUX` |
| GNU Screen | `RemoteScreen` | `multiplexer` | `STY` |
| Zellij | `RemoteZellij` | `multiplexer` | `ZELLIJ="0"` |
| OpenSSH | `RemoteSSH` | `environment` | `SSH_CONNECTION` |
| WSL | `RemoteWSL` | `environment` | WSL 固有マーカー |
| GitHub Codespaces | `RemoteCodespaces` | `environment` | `CODESPACES=true` |
| Gitpod | `RemoteGitpod` | `environment` | `GITPOD_WORKSPACE_ID` |
| Dev Containers | `RemoteDevContainer` | `environment` | Dev Container マーカー |

## マルチプレクサーだけが信頼度を下げる

マルチプレクサーは以前のクライアント環境を保持し、既存 pane を更新できません。検出すると、生存中の祖先で裏付けられない端末・エージェント・ランナーを `probable` に下げます。CI とリモート軸は対象外です。

SSH は端末が別マシンにある可能性を示しますが、値が古いことは意味しません。

## 検出できないもの

- **Mosh:** 通常のリモートシェルには識別用 `MOSH_*` がありません。`MOSH_KEY` は認証情報なので読みません。
- **一般コンテナ:** Docker・Podman は標準の識別環境変数を設定しません。

`SSH_AUTH_SOCK`、`WINDOW`、`LC_TERMINAL` は曖昧または SSH を越えるため、単独マーカーとして使いません。

根拠は[リモート調査文書](../../research/remote/)にあります。
