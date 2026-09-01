# ユーザーガイド

[English](../../en/guide/) | [한국어](../../ko/guide/) | 日本語

初めて使う場合は[はじめに](getting-started.md)から、すぐに使える実装例が必要なら[レシピ](recipes.md)から読んでください。

## 目的別に探す

| 目的 | 最初に読む文書 |
|---|---|
| Go プログラムに導入する | [はじめに](getting-started.md) |
| プロンプト・色・進捗表示を安全に制御する | [レシピ](recipes.md#対話機能を安全に判断する) |
| シェルスクリプトで分岐する | [CLI ガイド](cli.md#シェルから使う) |
| 結果をログや JSON に残す | [レシピ](recipes.md#実行コンテキストを記録する) |
| テストで環境を再現する | [はじめに](getting-started.md#別の環境を渡す) |
| 未対応の環境を追加する | [ドライバー作成](drivers.md) |
| 全フィールドとオプションを調べる | [API リファレンス](api.md) |
| 検出規則の根拠を確認する | [調査文書](../../research/) |

## 推奨する読み順

1. [はじめに](getting-started.md)で `Current()` と `Detect()` を選びます。
2. [レシピ](recipes.md)から用途に合うパターンを選びます。
3. 必要な軸の概念ガイドだけを読みます。
4. 高度な設定や拡張には [API](api.md) と[ドライバー](drivers.md)を参照します。

## 概念別ガイド

| ガイド | 答える質問 | 主な結果 |
|---|---|---|
| [エージェント](agents.md) | AI エージェントが実行したか、階層は何か | `Agents`, `Chain()` |
| [CI](ci.md) | どの CI プラットフォームとジョブか | `CI` |
| [ランナー](runner.md) | npm・make・systemd など何が実行したか | `Runners` |
| [リモート環境](remote.md) | SSH・tmux・開発コンテナを経由したか | `Remotes` |
| [端末と TTY](terminal.md) | どのアプリが環境を作り、現在対話可能か | `Terminal`, `TTY` |
| [祖先プロセス](process.md) | 検出した実行主体が現在も祖先か | `Process`, `AncestorPID` |

複数の軸は同時に成立します。Claude Code が GitHub Actions 内で `npm test` を実行すれば、`Agents`、`CI`、`Runners` がすべて埋まります。

## 解釈の原則

- 環境変数による検出はプロセス開始時に継承した状態です。現在の生存確認には[祖先プロセスによる裏付け](process.md#裏付け)を使います。
- `Terminal` は環境を作ったエミュレーター、`TTY` は現在のストリーム接続状態です。プロンプトには `TTY.Interactive` を使います。
- `AncestorPID != 0` は強い肯定的証拠です。ゼロは否定ではありません。
- `-json` にはセッション ID やローカルパスが含まれることがあります。共有用診断には `runby -v` がより安全です。

公式根拠、検証日、除外製品は[調査文書](../../research/)にあります。
