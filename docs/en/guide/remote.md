# Remote axis

Remote layers sit between the user and the process. They also explain how other environment variables may be filtered or retained.

```go
result := runby.Detect()
result.IsRemote()
result.Remote(runby.RemoteTmux)
result.Multiplexer()
```

Several layers can coexist, and their order is detection order rather than nesting order.

| Layer | `RemotePlatform` | `Kind` | Marker |
|---|---|---|---|
| tmux | `RemoteTmux` | `multiplexer` | `TMUX` |
| GNU Screen | `RemoteScreen` | `multiplexer` | `STY` |
| Zellij | `RemoteZellij` | `multiplexer` | `ZELLIJ="0"` |
| OpenSSH | `RemoteSSH` | `environment` | `SSH_CONNECTION` |
| WSL | `RemoteWSL` | `environment` | WSL distribution or interop markers |
| GitHub Codespaces | `RemoteCodespaces` | `environment` | `CODESPACES=true` |
| Gitpod | `RemoteGitpod` | `environment` | `GITPOD_WORKSPACE_ID` |
| Dev Containers | `RemoteDevContainer` | `environment` | Dev Container markers |

## Only multiplexers lower confidence

A multiplexer retains the environment of a previous client and cannot refresh an existing pane. When one is detected, uncorroborated terminal, agent, and runner findings are lowered to `probable`. A nonzero `AncestorPID` preserves confidence. CI and remote layers are not lowered.

SSH indicates that a terminal may live on another machine, not that inherited values are stale.

## What cannot be detected

- **Mosh:** Normal remote shells receive no `MOSH_*` identity variable. `MOSH_KEY` is a credential and must never be read.
- **Generic containers:** Docker and Podman add no standard identity environment variable. Files such as `/.dockerenv` are outside this environment-only scope.

`SSH_AUTH_SOCK`, `WINDOW`, and `LC_TERMINAL` are intentionally not used as standalone markers because they are ambiguous or can cross an SSH boundary.

See [remote research](../../research/remote/).
