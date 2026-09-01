# はじめに

`runby` のインストール、必要な結果の選択、テストでの実行環境再現までを説明します。

## 1. インストール

Go 1.24 以降が必要です。

```sh
go get github.com/ironpark/runby
```

```go
import "github.com/ironpark/runby"
```

シェルだけで使う場合は CLI をインストールします。

```sh
go install github.com/ironpark/runby/cmd/runby@latest
runby
```

サブコマンドと終了コードは [CLI ガイド](cli.md)を参照してください。

## 2. 現在の実行環境を読む

通常のアプリケーションでは `Current()` を使います。最初の呼び出しで一度検出し、その結果を再利用します。

```go
result := runby.Current()
if result.IsAgent() {
	log.Printf("agent=%s", result.Chain())
}
if result.IsCI() {
	log.Printf("ci=%s", result.CI.Provider)
}
```

結果にはランナー、リモート環境、端末、TTY、祖先プロセスも含まれます。

## 3. プログラムの動作を決める

入力可能かどうかはエージェント名や端末名ではなく `TTY.Interactive` で判断します。

```go
if !result.TTY.Interactive {
	return errors.New("この操作には対話端末が必要です")
}
confirmBeforeDelete()
```

エージェントや CI で表示を変える場合は、それぞれの軸を確認します。

```go
machineRun := result.IsAgent() || result.IsCI()
if _, service := result.RunnerOfKind(runby.RunnerKindService); service {
	machineRun = true
}
if machineRun {
	disableProgressAnimation()
}
```

## 4. 詳細を選ぶ

```go
result.Chain()
result.CI.Provider
result.Terminal.Program
result.Remote(runby.RemoteTmux)
result.Runner(runby.RunnerNPM)
```

製品固有フィールドは検出を確認してから読みます。

```go
if codex, ok := result.Agent(runby.AgentCodex); ok {
	log.Printf("sandbox=%s network=%s", codex.Sandbox.Mode, codex.Sandbox.Network)
}
```

## `Current()` と `Detect()` の選択

| 状況 | 推奨 API |
|---|---|
| 現在のプロセス情報を複数箇所で読む | `Current()` |
| 毎回環境を読み直す | `Detect()` |
| 別プロセスや記録済み環境を分類する | `Detect(WithEnviron(...))` |
| 環境フィクスチャでテストする | `Detect(WithEnviron(...))` |
| TTY・プロセスのシステムコールを省略する | `Detect(WithoutTTY(), WithoutProcessTree())` |
| ユーザー定義ドライバーを一回だけ分離する | `Detect(WithOnlyDrivers(...))` |

`Current()` は最初の結果をキャッシュします。後から変更した環境を反映するには `Detect()` を呼びます。

## 別の環境を渡す

`WithEnviron` は現在のプロセスとは限らない環境を分類します。ラッパーや決定的なテストに適しています。

```go
func TestGitHubActions(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=1234",
	}))
	if !result.IsCI() || result.CI.Provider != runby.CIGitHubActions {
		t.Fatalf("unexpected result: %#v", result.CI)
	}
}
```

明示的な環境を渡すと TTY と祖先プロセスの自動検査は無効になります。必要なら `WithTTY()` と `WithProcessTree()` で注入してください。

環境変数は開始時に継承されます。`IsAgent()` はエージェント信号を継承したことを示しますが、現在も生存している保証ではありません。非ゼロの `AncestorPID` は生存中の祖先による裏付けです。ゼロは否定ではありません。
