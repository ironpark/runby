# 端末軸と TTY

端末識別は意図的に弱い信号です。環境変数は継承されるため、`Terminal` は現在接続中の端末ではなく、環境を作ったエミュレーターを示します。

```go
term := runby.Detect().Terminal
term.Program
term.SessionID
term.Version
term.PID
term.Term
term.Confidence
```

iTerm2、Apple Terminal、WezTerm、Ghostty、Warp、Zed、VS Code 系、JetBrains 系、kitty、Windows Terminal、Alacritty、Konsole、GNOME Terminal、VTE 系を識別します。

端末識別を信頼境界に使わないでください。マルチプレクサーによる残存、SSH による転送、ユーザーによる設定が可能です。`TERM` は能力情報であり識別マーカーではありません。

Konsole、VS Code、JetBrains 系は製品ではなくエンジンや系列を示すため `probable` です。`Terminal.PID` を提供するのは kitty だけです。

根拠は[端末調査](../../research/terminals/)と[リモート調査](../../research/remote/)にあります。

## `HasTerminal()` である理由

Go の `IsTerminal` は通常ファイルディスクリプターが TTY かを尋ねます。`HasTerminal()` はエミュレーターを識別できたかを尋ねます。

```go
result.HasTerminal()
result.TTY.Interactive
```

## TTY と端末軸の違い

```go
tty := runby.Detect().TTY
tty.Inspected
tty.StdinTTY
tty.StdoutTTY
tty.StderrTTY
tty.Attached
tty.Interactive
```

`Interactive` はプロンプト可能なストリーム構成を意味し、人が直接実行した証明ではありません。プロンプト可否には `TTY.Interactive`、人向け表示の既定値には `Unattended()` を使います。

`WithEnviron` と `WithEnv` では TTY 検査を無効にします。必要なら `InspectTTY()` または `WithTTY()` を使います。

AIX、Solaris、z/OS では標準 `syscall` に必要な ioctl がないため `Attached` と `Interactive` は常に false です。他の軸には影響しません。
