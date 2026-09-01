# CLI

シェルから同じ判定を行う場合や、バグ報告に実行環境を添付する場合に使います。

```sh
go install github.com/ironpark/runby/cmd/runby@latest
```

## 使用法

```text
runby [-json] [-v]     読みやすい報告、または Result 全体の JSON
runby is <軸> [製品]  終了コードだけで回答
runby chain            "paseo>codex"、未検出なら "unknown"
runby help             ヘルプ（runby -h と同じ）
```

製品名を指定できる軸は `agent`、`ci`、`terminal`、`remote`、`runner` です。`tty` と `unattended` は製品名を取りません。

| フラグ | 説明 |
|---|---|
| `-json` | `Result` 全体を JSON で出力 |
| `-v` | 根拠となった環境変数名と祖先プロセスも出力 |

## シェルから使う

`is` は何も出力せず、終了コードで答えます。

```sh
if runby is agent; then
	export NO_COLOR=1
fi

runby is agent codex
runby is ci github-actions
runby is runner npm
runby is terminal ghostty
runby is remote tmux
```

製品名は `-json` の slug と同じです。エージェント・リモート・ランナーには複数階層があり、どれか一つが一致すれば真です。

| 終了コード | 意味 |
|---|---|
| `0` | 正常。`is` では真 |
| `1` | `is` では偽、または内部エラー |
| `2` | 不明なコマンド・軸・製品・フラグを含む使用法エラー |

入力ミスは偽ではなくエラーです。`runby is agent codexx` は 1 ではなく 2 を返します。

```sh
runby is agent codex
case $? in
	0) echo "codex" ;;
	1) echo "codex ではない" ;;
	*) echo "runby の呼び出しが不正" >&2; exit 2 ;;
esac
```

## 環境変数の値は出力しない

テキストモードと `-v` は環境変数の**名前**だけを表示します。

ただし `-json` の正規化済み結果にはセッション ID、作業ディレクトリ、製品固有値、祖先実行ファイルのフルパスが含まれることがあります。バグ報告には `-json` より `runby -v` が安全です。

```sh
runby -json | jq '{chain: [.agents[]?.name] | join(">"), ci: .ci.provider, tty: .tty.interactive}'
```

各 `is` 判定はライブラリの同名メソッドをそのまま使用し、CLI 独自の検出規則は持ちません。
