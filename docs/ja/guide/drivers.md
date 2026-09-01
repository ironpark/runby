# ドライバーの作成と配布

`runby` が認識しない環境を検出するにはドライバーを作ります。

| | `Register(...Driver)` | `WithOnlyDrivers(...Driver)` |
|---|---|---|
| 範囲 | 以後の全 `Detect()` と `Current()` | 一回の `Detect()` |
| 内蔵ドライバー | 拡張。同じ識別子は置換 | 無視 |
| 用途 | 再利用可能なモジュール | テストと完全制御 |

`WithDrivers` は一回だけ既定セットを拡張します。

## ドライバー構造体

- `AgentDriver{Agent, Kind, Models, Executables, Detect}`
- `CIDriver{Provider, Detect}`
- `TerminalDriver{Program, Executables, Detect}`
- `RemoteDriver{Platform, Kind, Executables, Detect}`
- `RunnerDriver{Tool, Kind, Executables, Detect}`

ドライバーが固定の識別子と分類を提供し、`Detect` が今回観測した値を返します。

## 検出ロジック

`EnvReader` を使うと、成功した検出で読んだ環境変数名が `Evidence` に反映されます。

```go
func detectAcme(env runby.Env) (runby.Agent, bool) {
	r := runby.NewEnvReader(env)
	id, ok := r.Value("ACME_RUN_ID")
	if !ok { return runby.Agent{}, false }
	return runby.Agent{
		AgentID: id,
		Axis: runby.Axis{Evidence: r.Evidence()},
	}, true
}
```

API キーやユーザー設定を実行マーカーとして使わないでください。`Evidence` には値を入れません。

## 再利用可能なモジュール

ドライバーパッケージは `init` から登録できます。利用側は最初の `Current()` より前に blank import します。

```go
import _ "example.com/acme-runby"
```

`Current()` がレジストリを初期化した後の `Register` は panic します。同じ軸・識別子の重複も panic しますが、既定セットの一致項目を新しいドライバーで置換することはできます。

## 分離テスト

```go
result := runby.Detect(
	runby.WithEnviron([]string{"ACME_RUN_ID=42"}),
	runby.WithOnlyDrivers(acmeDriver),
)
```

内蔵も必要なら `WithDrivers` を使います。`BuiltinDrivers()` は安全にフィルターできるコピーを返します。

`Executables` を設定すると祖先プロセスによる裏付けを受けます。マルチプレクサーのリモートドライバーは `RemoteKindMultiplexer` を設定し、古い環境への信頼度処理を有効にしてください。

内蔵ドライバーはすべて[調査文書](../../research/)を持ちます。ユーザー定義ドライバーに必須ではありませんが、判断根拠を記録すると環境変数契約の変更を追跡しやすくなります。
