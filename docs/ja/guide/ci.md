# CI 軸

CI はエージェントとは独立した軸です。Claude Code が GitHub Actions 内で動けば `Agents` と `CI` の両方が埋まり、CI は `Chain()` には入りません。

```go
result.IsCI()
result.CI.Provider
result.CI.PipelineID
result.CI.JobID
result.CI.Attempt
result.CI.Trigger
result.CI.PullRequest
result.CI.PullRequestID
```

対応プロバイダーはルート README に記載しています。検出順と根拠は [CI 調査文書](../../research/ci/)にあります。

## 正規化フィールド

| フィールド | 意味 |
|---|---|
| `PipelineID` | run・build・pipeline の識別子 |
| `BuildNumber` | 人向けカウンター |
| `JobID`, `JobName` | 個別 job・step |
| `Attempt` | 1 始まりの試行番号。未提供なら 0 |
| `Trigger` | プロバイダー固有のトリガー名 |
| `Runner` | ジョブを実行するマシン・エージェント |
| `PullRequest` | PR/MR 実行を明示したか |
| `PullRequestID` | PR/MR 識別子 |
| `Extra` | プロバイダー固有値 |

再試行回数は 1 始まりに統一し、Bitbucket UUID の波括弧は除去します。PR/MR はプロバイダーの直接信号だけを使い、ブランチ名やコミットメッセージから推測しません。

Forgejo は `GITHUB_*` の別名も提供するため GitHub Actions より先に検査します。古い Runner は区別できず GitHub Actions として報告されます。
