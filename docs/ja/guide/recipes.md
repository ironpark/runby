# レシピ

`runby` の結果をアプリケーション動作に接続するための一般的なパターンです。

## 対話機能を安全に判断する

プロンプトを表示して入力を受け取れるかは `TTY.Interactive` で判断します。

```go
if !result.TTY.Interactive {
	return errors.New("対話入力が必要です")
}
askForConfirmation()
```

`IsAgent()` だけでプロンプト可否を決めないでください。エージェントも PTY を割り当てられ、人が実行したコマンドもパイプやサービス内では TTY を持たないことがあります。

出力を見ている人がいるかに応じてスピナーや色を変える場合は `Unattended()` を使います。

```go
if result.Unattended() {
	disableSpinner()
	disableColor()
}
```

`definite` のエージェント、CI、サービスランナー、または検査済みで非対話の TTY のいずれかで真になります。`probable` のエージェントだけでは真にならず、未検査の TTY も証拠になりません。

独自ポリシーでは軸を直接組み合わせます。

```go
automated := result.Unattended()
if _, ok := result.RunnerOfKind(runby.RunnerKindHook); ok {
	automated = true
}
```

## エージェント別に動作を変える

```go
if result.IsAgent() {
	disableSpinner()
}
if codex, ok := result.Agent(runby.AgentCodex); ok &&
	codex.Sandbox.Network == runby.NetworkDisabled {
	skipNetworkChecks()
}
```

複数階層では `Primary()` が最外側の代表階層、`Chain()` が全階層を返します。

## 実行コンテキストを記録する

```go
log.Printf(
	"agent=%s ci=%s interactive=%t remote=%t runner=%t",
	result.Chain(), result.CI.Provider, result.TTY.Interactive,
	result.IsRemote(), result.HasRunner(),
)
```

完全な JSON にはセッション ID、作業ディレクトリ、実行ファイルパスが含まれることがあります。外部テレメトリには必要なフィールドだけを選んでください。

## CI とローカル実行を区別する

```go
if result.IsCI() {
	configureForCI(result.CI.Provider, result.CI.Attempt)
} else {
	configureForLocalRun()
}
```

エージェントは CI 内でも実行できるため、`IsAgent()` と `IsCI()` は排他的ではありません。

## スクリプト・フック・サービスを区別する

```go
if npm, ok := result.Runner(runby.RunnerNPM); ok {
	log.Printf("npm script=%s", npm.Task)
}
if _, ok := result.RunnerOfKind(runby.RunnerKindService); ok {
	disableHumanOrientedOutput()
}
```

複数のランナーが同時に存在できます。配列順は検出順であり、ネスト順の証明ではありません。

## tmux の古い環境を検出する

```go
if mux, ok := result.Multiplexer(); ok {
	log.Printf("%s 内では継承環境が古い可能性があります", mux.Platform)
}
```

裏付けのない端末・エージェント・ランナーは信頼度が下がります。重要な判断では[祖先プロセス](process.md)も確認してください。CI はこの降格対象外です。

## シェルスクリプトで分岐する

```sh
if runby is agent || runby is ci; then
	export NO_COLOR=1
fi
```

共有用診断には `runby -v`、機械処理には `runby -json` を使います。
