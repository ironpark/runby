# runby

English | [한국어](README.ko.md) | [日本語](README.ja.md)

A Go package that detects **who launched the current process, where it is running, and how it was invoked**.

It reports whether a command was launched by an AI agent or CI job, whether its streams are attached to a terminal, and whether it was invoked through tools such as npm or systemd. It uses only the Go standard library.

## Use cases

- Disable confirmation prompts and progress animations under AI agents or CI.
- Record the actual execution stack, such as `paseo>codex`, in logs.
- Distinguish npm, make, pre-commit, or systemd execution from direct invocation.
- Capture terminal, SSH, tmux, and ancestor-process context in bug reports.

## Installation

Go 1.24 or later is required.

As a library:

```sh
go get github.com/ironpark/runby
```

As a CLI:

```sh
go install github.com/ironpark/runby/cmd/runby@latest
```

## Quick start

Most programs only need `Current()`, which returns a cached description of the current execution environment.

```go
package main

import (
	"log"

	"github.com/ironpark/runby"
)

func main() {
	result := runby.Current()

	if result.IsAgent() {
		log.Printf("launched by an AI agent: %s", result.Chain())
	}
	if result.IsCI() {
		log.Printf("running in CI: %s", result.CI.Provider)
	}
	if !result.TTY.Interactive {
		disableInteractivePrompts()
	}
}
```

Use `Detect()` when the environment may change between calls or when supplying a test environment.

```go
freshResult := runby.Detect()
testResult := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
```

See [Getting started](docs/en/guide/getting-started.md) for the full setup flow and [Recipes](docs/en/guide/recipes.md) for practical patterns.

## Choose the question you need

| Question | API | Typical result |
|---|---|---|
| Did an AI agent launch this? | `result.IsAgent()` | `true` |
| What is the agent stack? | `result.Chain()` | `"paseo>codex"` |
| Is this running in CI? | `result.IsCI()` | `true` |
| Did npm, make, or systemd launch it? | `result.HasRunner()` | `true` |
| Did it pass through SSH, tmux, or a development container? | `result.IsRemote()` | `true` |
| Which terminal environment created it? | `result.HasTerminal()` / `result.Terminal` | `ghostty` |
| Can the process prompt right now? | `result.TTY.Interactive` | `true` |
| Is nobody watching the output? | `result.Unattended()` | `true` |
| Is a detected launcher still a live ancestor? | `AncestorPID` | `2540` |

`Terminal` and `TTY` are different. `Terminal` identifies the emulator that created the environment; `TTY` reports whether the current standard streams are attached to a terminal. Use `TTY.Interactive` when deciding whether to prompt.

See the [API reference](docs/en/guide/api.md) for all fields and options.

## CLI

```console
$ runby
agent     paseo>codex
            paseo          orchestrator  delegated     definite  live ancestor pid=84445
            codex          harness       first-party   probable
ci        -
terminal  ghostty (probable)
remote    tmux (multiplexer)
runner    npm (script) test
tty       interactive (stdin and output are terminals)
process   7 ancestors

warning: running inside tmux. Multiplexers cannot refresh the environment of
an existing pane, so environment-derived axes (terminal, agent, runner) may be stale.
```

`codex` and `ghostty` are `probable` here because of tmux. A multiplexer cannot refresh an existing pane's environment, so findings not corroborated by a live ancestor lose one confidence level. `paseo` remains `definite` because an ancestor process confirms it.

Use the `is` command in shell conditions; it answers through its exit status instead of printing output.

```sh
if runby is agent; then
	export NO_COLOR=1
fi

runby is agent codex   # narrow the agent axis to one product
runby is remote tmux   # the same form works for every named axis
runby is unattended    # the same rule as Result.Unattended()
runby chain            # paseo>codex
runby -v               # evidence variable names and ancestor processes
runby -json            # the complete Result as JSON
```

Product names are the slugs shown by `-json`. A typo returns usage error 2 rather than false (1), preventing scripts from silently taking the wrong branch.

`-json` may contain session identifiers and local paths. For diagnostics intended for sharing, `runby -v` is safer because it prints variable names rather than values. See the [CLI guide](docs/en/guide/cli.md).

## Independent axes

One process can be launched by Codex, run inside GitHub Actions, pass through an npm script, and have tmux attached at the same time. `runby` reports these as independent axes.

| Axis | Question | Result field |
|---|---|---|
| Agent | Who requested the command? | `Agents` |
| CI | Which CI job owns the process? | `CI` |
| Runner | Did npm, make, systemd, or another tool launch it? | `Runners` |
| Remote | What sits between the user and the process? | `Remotes` |
| Terminal | Which emulator created the environment? | `Terminal` |
| Process | Which ancestors are alive now? | `Process` |

Standard-stream attachment is reported separately in `TTY`. See the [concept guides](docs/en/guide/README.md#concept-guides) for semantics and confidence rules.

## Supported products

- **Agents:** Paseo, Orca, Antigravity 2.0, Cursor Agent, OpenCode, Amp, OpenClaw, Auggie, pi, Charm Crush, Roo Code, OpenHands, Cline, OpenAI Codex, Claude Code, Gemini CLI, Grok Build, Qwen Code, DeepSeek Harness
- **CI:** GitHub Actions, Forgejo Actions, Gitea Actions, GitLab CI/CD, CircleCI, Travis CI, Buildkite, Azure Pipelines, Bitbucket Pipelines, Jenkins, Vercel, Netlify, TeamCity, Drone, AppVeyor, Semaphore, Cirrus CI, AWS CodeBuild, Google Cloud Build, Xcode Cloud, Cloudflare Pages, Cloudflare Workers Builds, Woodpecker CI, Bitrise, Render, Harness CI, Bamboo, GoCD, Taskcluster, Sourcehut, Codefresh, Codemagic, Buddy, Screwdriver, Vela, and generic CI conventions
- **Runners:** npm, pnpm, Bun, GNU Make, systemd, pre-commit
- **Remote environments:** tmux, GNU Screen, Zellij, OpenSSH, WSL, GitHub Codespaces, Gitpod, Dev Containers
- **Terminals:** iTerm2, Apple Terminal, WezTerm, Ghostty, Warp, Zed, VS Code family, JetBrains family, kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal, VTE family

Add products not listed here with a [driver](docs/en/guide/drivers.md). Products intentionally excluded and the reasons are recorded in the [research documents](docs/research/).

## Important limitations

Environment-based results are **snapshots taken when a process starts**. A detected agent is not necessarily still alive. Long-lived environments such as tmux can retain values from previous sessions, so a multiplexer lowers uncorroborated layers to `probable`.

Check `AncestorPID` when liveness matters. A nonzero value confirms a live ancestor process; zero does not mean the detection is wrong. See the [process guide](docs/en/guide/process.md).

Environment values may contain tokens or paths, so `Evidence` stores **variable names only**.

## Documentation

- [Getting started](docs/en/guide/getting-started.md) — installation, first detection, and tests
- [Recipes](docs/en/guide/recipes.md) — prompts, logs, CI, and diagnostics
- [CLI guide](docs/en/guide/cli.md) — shell conditions, JSON, and exit codes
- [API reference](docs/en/guide/api.md) — `Result`, options, caching, and driver APIs
- [Concept guides](docs/en/guide/) — agents, CI, runners, remote environments, terminals, and processes
- [Research documents](docs/research/) — official evidence and exclusion rationale for each detector

## Platforms and dependencies

Environment-based detection works the same way on every platform. Ancestor-process inspection supports Linux, macOS, and Windows; elsewhere `Process.Supported == false`. See the [terminal guide](docs/en/guide/terminal.md#tty-versus-the-terminal-axis) for TTY limitations on some Unix platforms.

`runby` has no external Go module dependencies.

## License

[MIT](LICENSE)
