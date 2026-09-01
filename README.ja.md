# runby

[English](README.md) | [한국어](README.ko.md) | 日本語

現在のプロセスを**誰が、どこで、どのように実行したか**を検出する Go パッケージです。

AI エージェントが実行したコマンドか、CI ジョブ内か、現在のストリームが端末に接続されているか、npm スクリプトや systemd を経由したかを一つの結果で確認できます。外部依存はなく、Go 標準ライブラリだけを使用します。

## 主な用途

- AI エージェントや CI では確認プロンプトと進捗アニメーションを無効にする
- ログに `paseo>codex` のような実際の実行階層を残す
- npm・make・pre-commit・systemd からの実行を直接実行と区別する
- バグ報告に端末、SSH、tmux、親プロセス情報をまとめて添付する

## インストール

Go 1.24 以降が必要です。

ライブラリとして使用する場合:

```sh
go get github.com/ironpark/runby
```

CLI を使用する場合:

```sh
go install github.com/ironpark/runby/cmd/runby@latest
```

## クイックスタート

ほとんどのプログラムでは、現在の実行情報をキャッシュする `Current()` だけで十分です。

```go
package main

import (
	"log"

	"github.com/ironpark/runby"
)

func main() {
	result := runby.Current()

	if result.IsAgent() {
		log.Printf("AI エージェントが実行: %s", result.Chain())
	}
	if result.IsCI() {
		log.Printf("CI で実行中: %s", result.CI.Provider)
	}
	if !result.TTY.Interactive {
		disableInteractivePrompts()
	}
}
```

環境を毎回検出し直す場合やテスト用環境を渡す場合は `Detect()` を使います。

```go
freshResult := runby.Detect()
testResult := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
```

詳しい導入手順は[はじめに](docs/ja/guide/getting-started.md)、実用例は[レシピ](docs/ja/guide/recipes.md)を参照してください。

## 必要な質問を選ぶ

| 知りたいこと | API | 代表的な結果 |
|---|---|---|
| AI エージェントが実行したか | `result.IsAgent()` | `true` |
| エージェント階層は何か | `result.Chain()` | `"paseo>codex"` |
| CI ジョブ内か | `result.IsCI()` | `true` |
| npm・make・systemd などが実行したか | `result.HasRunner()` | `true` |
| SSH・tmux・開発コンテナを経由したか | `result.IsRemote()` | `true` |
| どの端末環境で開始されたか | `result.HasTerminal()` / `result.Terminal` | `ghostty` |
| 現在プロンプトを使用できるか | `result.TTY.Interactive` | `true` |
| 出力を見ている人がいないか | `result.Unattended()` | `true` |
| 検出した実行主体が現在も祖先か | `AncestorPID` | `2540` |

`Terminal` と `TTY` は異なります。`Terminal` は環境を作ったエミュレーター、`TTY` は現在の標準ストリームが端末に接続されているかを表します。プロンプトの可否には `TTY.Interactive` を使ってください。

全フィールドとオプションは [API リファレンス](docs/ja/guide/api.md)にあります。

## CLI

```console
$ runby
agent     paseo>codex
            paseo          orchestrator  delegated     definite  生存中の祖先 pid=84445
            codex          harness       first-party   probable
ci        -
terminal  ghostty (probable)
remote    tmux (multiplexer)
runner    npm (script) test
tty       対話的 (stdin と出力が端末)
process   祖先 7 個

注意: tmux 内で実行中です。マルチプレクサーは既存 pane の環境を更新できないため、
環境由来の軸（端末・エージェント・ランナー）の値が古い可能性があります。
```

ここで `codex` と `ghostty` が `probable` なのは tmux のためです。生存中の祖先で裏付けられない判定は信頼度が一段下がります。`paseo` は祖先プロセスで裏付けられるため `definite` のままです。

シェル条件では、出力ではなく終了コードで答える `is` を使います。

```sh
if runby is agent; then
	export NO_COLOR=1
fi

runby is agent codex
runby is remote tmux
runby is unattended
runby chain
runby -v
runby -json
```

製品名は `-json` に表示される slug と同じです。入力ミスは偽（1）ではなく使用法エラー（2）となり、スクリプトが誤った分岐を黙って選ぶことを防ぎます。

`-json` にはセッション ID やローカルパスが含まれることがあります。共有用の診断には値ではなく変数名だけを表示する `runby -v` の方が安全です。詳しくは [CLI ガイド](docs/ja/guide/cli.md)を参照してください。

## 独立した情報軸

一つのプロセスが Codex に実行され、GitHub Actions 内にあり、npm スクリプトを経由し、tmux が接続されていることがあります。`runby` はこれらを独立した軸として報告します。

| 軸 | 答える質問 | 結果フィールド |
|---|---|---|
| エージェント | 誰がコマンドを要求したか | `Agents` |
| CI | どの CI ジョブ内か | `CI` |
| ランナー | npm・make・systemd など何が実行したか | `Runners` |
| リモート | ユーザーとプロセスの間に何があるか | `Remotes` |
| 端末 | どのエミュレーターが環境を作ったか | `Terminal` |
| プロセス | 現在生存している祖先は何か | `Process` |

標準ストリームの接続状態は別の `TTY` フィールドで提供します。各軸の意味と信頼度は[概念別ガイド](docs/ja/guide/README.md#概念別ガイド)を参照してください。

## 対応範囲

- **エージェント:** Paseo, Orca, Antigravity 2.0, Cursor Agent, OpenCode, Amp, OpenClaw, Auggie, pi, Charm Crush, Roo Code, OpenHands, Cline, OpenAI Codex, Claude Code, Gemini CLI, Grok Build, Qwen Code, DeepSeek Harness
- **CI:** GitHub Actions, Forgejo Actions, Gitea Actions, GitLab CI/CD, CircleCI, Travis CI, Buildkite, Azure Pipelines, Bitbucket Pipelines, Jenkins, Vercel, Netlify, TeamCity, Drone, AppVeyor, Semaphore, Cirrus CI, AWS CodeBuild, Google Cloud Build, Xcode Cloud, Cloudflare Pages, Cloudflare Workers Builds, Woodpecker CI, Bitrise, Render, Harness CI, Bamboo, GoCD, Taskcluster, Sourcehut, Codefresh, Codemagic, Buddy, Screwdriver, Vela, 一般的な CI 規約
- **ランナー:** npm, pnpm, Bun, GNU Make, systemd, pre-commit
- **リモート環境:** tmux, GNU Screen, Zellij, OpenSSH, WSL, GitHub Codespaces, Gitpod, Dev Containers
- **端末:** iTerm2, Apple Terminal, WezTerm, Ghostty, Warp, Zed, VS Code 系, JetBrains 系, kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal, VTE 系

一覧にない製品は[ドライバー](docs/ja/guide/drivers.md)で追加できます。検出対象外の製品と理由も[調査文書](docs/research/)に記録しています。

## 重要な制限

環境変数ベースの結果は**プロセス開始時点のスナップショット**です。検出したエージェントが現在も生存しているとは限りません。tmux のような長時間維持される環境では古い値が残ることがあるため、マルチプレクサーを検出すると生存中の祖先で裏付けられない階層を `probable` に下げます。

生存確認が重要なら `AncestorPID` を確認してください。非ゼロなら生存中の祖先で裏付けられていますが、`0` は検出が誤りという意味ではありません。詳しくは[プロセスガイド](docs/ja/guide/process.md)を参照してください。

環境変数の値にはトークンやパスが含まれることがあるため、`Evidence` には**変数名だけ**を保存します。

## ドキュメント

- [はじめに](docs/ja/guide/getting-started.md) — インストール、最初の判定、テスト
- [レシピ](docs/ja/guide/recipes.md) — プロンプト、ログ、CI、診断
- [CLI ガイド](docs/ja/guide/cli.md) — シェル条件、JSON、終了コード
- [API リファレンス](docs/ja/guide/api.md) — `Result`、オプション、キャッシュ、ドライバー API
- [概念別ガイド](docs/ja/guide/) — エージェント、CI、ランナー、リモート、端末、プロセス
- [調査文書](docs/research/) — 各検出規則の公式根拠と除外理由

## プラットフォームと依存関係

環境変数ベースの検出はすべてのプラットフォームで同様に動作します。祖先プロセスチェーンは Linux・macOS・Windows をサポートし、それ以外では `Process.Supported == false` です。一部 Unix の TTY 制限は[端末ガイド](docs/ja/guide/terminal.md#tty-と端末軸の違い)を参照してください。

`runby` は外部 Go モジュールに依存しません。

## ライセンス

[MIT](LICENSE)
