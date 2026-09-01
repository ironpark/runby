# ランナー軸 (`Result.Runners`)

パッケージマネージャーのスクリプト、ビルドレシピ、フック、サービス管理など、何がプロセスを直接実行したかを表します。

```go
result.HasRunner()
result.Runner(runby.RunnerNPM)
result.RunnerOfKind(runby.RunnerKindService)
```

## 検出対象

| ツール | `RunnerTool` | `Kind` | マーカー | `Task` |
|---|---|---|---|---|
| npm | `RunnerNPM` | `script` | npm user agent が `npm/` で始まる | スクリプト名 |
| pnpm | `RunnerPNPM` | `script` | `pnpm/` で始まる | スクリプト名 |
| Bun | `RunnerBun` | `script` | `bun/` で始まる | スクリプト名 |
| GNU Make | `RunnerMake` | `script` | `MAKELEVEL` | なし |
| systemd | `RunnerSystemd` | `service` | `INVOCATION_ID` | なし |
| pre-commit | `RunnerPreCommit` | `hook` | `PRE_COMMIT=1` | なし |

`script` はプロジェクト設定内のコマンド、`hook` はイベントへの反応、`service` はサービス管理下の無人実行を表します。

複数ランナーを同時に検出できます。順序は検出順であり、ネスト順ではありません。CI 軸とも独立しています。

## 検出できないもの

- **通常の Git フック:** 全フックに固有の環境変数がありません。pre-commit はフレームワーク固有マーカーがあるため例外です。
- **cron:** 実行主体を示す信頼できる変数がありません。
- **Yarn:** 公式契約または実測で確認できるまで除外しています。

パッケージマネージャーは user-agent の接頭辞で区別します。`MAKEFLAGS` や `JOURNAL_STREAM` は単独マーカーではありません。認証情報を含む可能性がある `npm_lifecycle_script` は読みません。

根拠は[ランナー調査文書](../../research/runners/)にあります。
