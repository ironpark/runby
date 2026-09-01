# エージェント軸

「誰がこのコマンドを要求したか」に答える軸で、結果は `Result.Agents` に入ります。

```go
result.IsAgent()
result.Chain()
result.Primary()
result.Agent(runby.AgentCodex)
```

API キーや一般設定変数は、エージェントがプロセスを実行した証拠として使いません。

## 検出対象

Paseo、Orca、Antigravity 2.0、Cursor Agent、OpenCode、Amp、OpenClaw、Auggie、pi、Charm Crush、Roo Code、OpenHands、Cline、OpenAI Codex、Claude Code、Gemini CLI、Grok Build、Qwen Code、DeepSeek Harness を検出します。

Cline と Roo Code は製品所有の端末に付くマーカーを使うため `probable` です。Grok Build はプラグインフック内だけで検出できます。一般名の変数は値が対応製品を正確に示す場合だけ使います。

公式に確認できる汎用の子プロセス実行マーカーがない製品は意図的に除外しています。理由は[エージェント調査文書](../../research/agents/)にあります。

## `Kind` と `Models`

`Kind` は何を駆動するか（`harness` または `orchestrator`）、`Models` は知能の提供元（`first-party`、`multi-vendor`、`delegated`）を表します。この分類は製品の性質であり、今回実際に選択されたモデルではありません。

## 複数階層

Paseo が Codex を起動すると両方が入り、Paseo が `Primary()` になります。順序はオーケストレーター、マルチベンダーハーネス、ファーストパーティーハーネスです。

## セッションとエージェント識別子

`SessionID()` と `AgentID()` は外側から最初に値を持つ階層を探し、値と提供元エージェントを組にして返します。特定製品の値には `Agent(name)` を使います。識別子はテキスト報告には出ませんが `-json` には含まれます。

## Confidence

`definite` は製品が実行したプロセスに設定するマーカー、`probable` はエージェント実行と矛盾しない補助信号です。非ゼロの `AncestorPID` が最も強い裏付けです。

## ユーザー定義エージェントドライバーを追加する

`AgentDriver` に識別子、分類、実行ファイル、`Detect` 関数を設定します。詳しくは[ドライバー作成](drivers.md)を参照してください。
