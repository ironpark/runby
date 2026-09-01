# プロセス軸

祖先プロセスは環境変数以外から得る唯一の証拠であり、最も強い裏付けです。

```go
tree := runby.Detect().Process
tree.Supported
tree.Ancestors
tree.FindAgent(runby.AgentCodex)
```

`runby -v` も同じチェーンを表示します。軸結果の `AncestorPID` は、環境による検出を生存中の祖先実行ファイルが裏付けたことを意味します。

## 裏付け

```go
for _, agent := range runby.Current().Agents {
	if agent.AncestorPID != 0 {
		// 環境信号と生存中の祖先が一致
	}
}
```

エージェント、端末、リモート、ランナーは裏付け可能です。CI ジョブは祖先実行ファイルではないため対象外です。

生存中の端末祖先が見つかれば、マルチプレクサーによる信頼度降格を防ぎます。ユーザー定義ドライバーも `Executables` を指定すれば同じ裏付けを受けます。

`AncestorPID == 0` は否定ではありません。他ユーザーのプロセスで検査が止まる場合、未対応プラットフォーム、実行主体が祖先として残らない起動方式があります。

## プラットフォーム

| プラットフォーム | 方法 |
|---|---|
| Linux | `/proc/<pid>/stat`, `/proc/<pid>/exe`, `/proc/<pid>/comm` |
| macOS | `sysctl(KERN_PROC_PID)` と `KERN_PROCARGS2` |
| Windows | `CreateToolhelp32Snapshot` |
| その他 | `Supported == false`、空のチェーン |

この軸は現在のプロセスを記述するため、`WithEnviron` と `WithEnv` では無効です。`WithoutProcessTree()` で明示的に省略できます。
