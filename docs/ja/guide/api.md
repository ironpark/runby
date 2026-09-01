# API リファレンス

初めて使う場合は先に[はじめに](getting-started.md)を読んでください。

通常はキャッシュ済みの `Current()`、毎回の再検出やオプション適用には `Detect(opts ...Option)`、プロセス全体へのユーザー定義ドライバー追加には `Register()` を使います。

```go
result := runby.Current()
result := runby.Detect()
result := runby.Detect(runby.WithEnviron(environ))
result := runby.Detect(runby.WithoutTTY())
runby.Register(myDriver)
```

## オプション

| オプション | 説明 |
|---|---|
| `WithEnviron([]string)` | `"NAME=value"` 形式の環境を指定 |
| `WithTTY(TTY)` | 標準ストリーム状態を注入 |
| `WithProcessTree(ProcessTree)` | 祖先チェーンを注入 |
| `WithoutTTY()` | TTY システムコールを省略 |
| `WithoutProcessTree()` | 祖先プロセス検査を省略 |
| `WithEnv(Env)` | 任意の環境実装を指定 |
| `WithDrivers(...Driver)` | 既定セットを拡張。同じ識別子は置換 |
| `WithOnlyDrivers(...Driver)` | 指定ドライバーだけを実行 |

環境を明示すると、その環境が別プロセスのものかもしれないため TTY と祖先検査は自動的に無効になります。必要なら直接注入してください。

`Register(d)` はプロセス全体、`WithDrivers(d)` は一回だけの拡張、`WithOnlyDrivers(d)` は分離テスト向けです。同じオプション内の重複識別子は panic します。CI・端末軸は最初の一致、エージェントは分類順、リモート・ランナーは全一致を返します。

## `Result`

```go
type Result struct {
	Agents   []Agent
	TTY      TTY
	Process  ProcessTree
	CI       CI
	Terminal Terminal
	Remotes  []Remote
	Runners  []Runner
}
```

```go
result.IsAgent()
result.IsCI()
result.HasTerminal()
result.IsRemote()
result.HasRunner()

result.Agent(runby.AgentCodex)
result.Remote(runby.RemoteSSH)
result.Runner(runby.RunnerNPM)
result.RunnerOfKind(runby.RunnerKindService)

result.Primary()
result.Chain()
result.SessionID()
result.AgentID()
result.Multiplexer()
result.Unattended()
```

`SessionID()` と `AgentID()` は値だけでなく、その値を提供したエージェントを含む `Identifier` を返します。

## `Unattended()`

軸を組み合わせる唯一のメソッドです。次のいずれかで真になります。

| 条件 | 理由 |
|---|---|
| `definite` のエージェント | PTY があっても人がいない場合がある |
| `IsCI()` | 出力は CI が収集する |
| `RunnerKindService` | サービス管理下で実行される |
| `TTY.Inspected && !TTY.Interactive` | 検査済みストリームでプロンプト不可 |

`probable` のエージェントだけでは真になりません。未検査 TTY も証拠ではありません。表示の既定値には使えますが、セキュリティ境界には使わないでください。

## 共通軸フィールド

```go
type Axis struct {
	Confidence Confidence
	Extra      map[string]string
	Evidence   []string
}
```

`Evidence` には環境変数名だけが入り、値は入りません。`Extra` は製品固有の正規化済み値です。

エージェントには識別子、分類、セッション、sandbox、パス、`AncestorPID` が追加されます。`AncestorPID` は端末・リモート・ランナーにもありますが、CI にはありません。ゼロは否定ではありません。

すべての enum はゼロ値を `"unknown"` として表示します。

## キャッシュ

`Current()` はプロセスごとに `Detect()` を一度だけ計算します。テストではグローバルキャッシュを避け、明示的な結果を作ってください。

```go
result := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
```

## ドライバー拡張

内蔵ドライバーとユーザー定義ドライバーは同じ型を使います。

```go
acme := runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Kind:        runby.KindOrchestrator,
	Models:      runby.ModelsDelegated,
	Executables: []string{"acme-run"},
	Detect: func(env runby.Env) (runby.Agent, bool) {
		r := runby.NewEnvReader(env)
		id, ok := r.Value("ACME_RUN_ID")
		if !ok { return runby.Agent{}, false }
		return runby.Agent{AgentID: id, Axis: runby.Axis{Evidence: r.Evidence()}}, true
	},
}

result := runby.Detect(runby.WithDrivers(acme))
runby.Register(acme)
```

エージェント順は `Kind` と `Models` から決まり、オーケストレーター、マルチベンダーハーネス、ファーストパーティーハーネスの順です。

| 軸 | ドライバー | 識別子 | 分類 | 実行ファイル |
|---|---|---|---|---|
| Agent | `AgentDriver` | `Agent` | `Kind` + `Models` | あり |
| CI | `CIDriver` | `Provider` | — | なし |
| Terminal | `TerminalDriver` | `Program` | — | あり |
| Remote | `RemoteDriver` | `Platform` | `Kind` | あり |
| Runner | `RunnerDriver` | `Tool` | `Kind` | あり |

`EnvReader` は読み取った環境変数名を記録し、検出ロジックと `Evidence` のずれを防ぎます。主なメソッドは `Value`、`Bool`、`Any`、`First`、`Extra`、`Peek`、`Record`、`Evidence` です。同時利用には対応していないため、検出ごとに一つ作ります。

配布と登録の詳細は[ドライバー作成](drivers.md)を参照してください。
