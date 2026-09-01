# Terminal axis and TTY

Terminal identity is intentionally a weak signal. Environment variables are inherited, so `Terminal` identifies the emulator that created the environment, not necessarily the emulator attached now.

```go
term := runby.Detect().Terminal
term.Program
term.SessionID
term.Version
term.PID
term.Term
term.Confidence
```

Supported identities include iTerm2, Apple Terminal, WezTerm, Ghostty, Warp, Zed, VS Code and JetBrains families, kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal, and the VTE family.

Do not use terminal identity as a trust boundary. Multiplexers can retain it, SSH tools can forward it, and users can export markers themselves. `TERM` describes terminal capabilities and is context only, not an identity marker.

Konsole, VS Code, and JetBrains family detections are `probable` because their variables identify an engine or family rather than one application. `Terminal.PID` is available only for kitty.

See [terminal research](../../research/terminals/) and [remote research](../../research/remote/).

## Why `HasTerminal()` instead of `IsTerminal()`

In Go, `IsTerminal` conventionally asks whether a file descriptor is a TTY. `runby.HasTerminal()` instead asks whether an emulator identity was detected.

```go
result.HasTerminal()    // emulator identity from environment variables
result.TTY.Interactive  // current stream attachment from system calls
```

## TTY versus the terminal axis

```go
tty := runby.Detect().TTY
tty.Inspected
tty.StdinTTY
tty.StdoutTTY
tty.StderrTTY
tty.Attached
tty.Interactive
```

`Interactive` means the streams can support a prompt, not that a person directly launched the command. Use it for prompt capability and `Unattended()` for human-oriented display defaults. Combine axes directly for a different policy.

TTY inspection is disabled when `WithEnviron` or `WithEnv` supplies an environment that may describe another process. Use `InspectTTY()` or inject `WithTTY()` when appropriate.

On AIX, Solaris, and z/OS, `Attached` and `Interactive` are always false because the standard `syscall` package does not expose the required ioctl constants. Other axes are unaffected.
