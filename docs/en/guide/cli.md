# CLI

Use the CLI to ask the same questions from shell scripts or attach one execution-environment summary to a bug report. See [Getting started](getting-started.md) for Go integration.

```sh
go install github.com/ironpark/runby/cmd/runby@latest
```

## Usage

```text
runby [-json] [-v]     human-readable report or complete Result JSON
runby is <axis> [name] answer through the exit status only
runby chain            one line such as "paseo>codex"; "unknown" on no detection
runby help             show help (same as runby -h)
```

| Axis | Accepts a product name? |
|---|---|
| `agent` `ci` `terminal` `remote` `runner` | Yes |
| `tty` `unattended` | No |

`unattended` is the library's one cross-axis question: “is nobody watching this output?” Prefer it over manually combining agent, CI, and TTY checks when deciding whether to draw spinners or prompts; the library rule intentionally excludes `probable` agents.

| Flag | Meaning |
|---|---|
| `-json` | Print the complete `Result` as JSON |
| `-v` | Include evidence variable names and the ancestor-process chain |

## Default report

```text
$ runby
agent     paseo>claude-code
            paseo          orchestrator  delegated     definite  live ancestor pid=2540
            claude-code    harness       first-party   definite  live ancestor pid=4344
ci        -
terminal  ghostty (definite)
remote    tmux (multiplexer), openssh (environment)
runner    npm (script) test, gnu-make (script)
tty       interactive (stdin and output are terminals)
process   7 ancestors
```

With `-v`, each axis includes a `←` line listing matched environment-variable names, followed by the ancestor-process chain.

## Use it from a shell

`is` prints nothing and answers with its exit status.

```sh
if runby is agent; then
	export NO_COLOR=1
fi
if runby is ci; then
	go test ./... -race
fi
```

Append a product slug to narrow the question.

```sh
runby is agent codex
runby is ci github-actions
runby is runner npm
runby is terminal ghostty
runby is remote tmux
```

Product names are the slugs shown by `-json`, such as `claude-code`, `github-actions`, and `gnu-make`. Agent, remote, and runner axes can contain several layers, so every detected matching layer returns true.

| Exit code | Meaning |
|---|---|
| `0` | Success; true for `is` |
| `1` | False for `is`, or an internal error |
| `2` | Usage error, including unknown commands, axes, products, or flags |

A typo is an error, not false: `runby is agent codexx` returns 2. Handle the distinction when required:

```sh
runby is agent codex
case $? in
	0) echo "codex" ;;
	1) echo "not codex" ;;
	*) echo "invalid runby invocation" >&2; exit 2 ;;
esac
```

JSON output works directly with `jq`:

```sh
runby -json | jq -r '.agents[]? | select(.ancestor_pid != null) | .name'
```

## Environment values are not printed

Text modes and `-v` print environment-variable **names**, never their values. Values may contain tokens.

`-json` is different: it does not dump the environment, but normalized result fields may still contain product-advertised identifiers and paths.

| Field | Potential content |
|---|---|
| `agents[].agent_id`, `agents[].session_id` | Agent and session UUIDs |
| `agents[].paths.*` | Working and data directories |
| `*.extra` | Product-specific values such as worktree paths or host IDs |
| `process.ancestors[].path` | Full ancestor executable paths |

For a bug report, `runby -v` is safer than `-json`. For logs and telemetry, select only required fields.

```sh
runby -json | jq '{chain: [.agents[]?.name] | join(">"), ci: .ci.provider, tty: .tty.interactive}'
```

## Library equivalents

| CLI | Library (`result := runby.Current()`) |
|---|---|
| `runby is agent` | `result.IsAgent()` |
| `runby is agent codex` | `result.Agent(runby.AgentCodex)` |
| `runby is ci` | `result.IsCI()` |
| `runby is terminal` | `result.HasTerminal()` |
| `runby is remote` | `result.IsRemote()` |
| `runby is runner` | `result.HasRunner()` |
| `runby is tty` | `result.TTY.Interactive` |
| `runby is unattended` | `result.Unattended()` |

The CLI has no separate detection policy; tests pin these mappings to the library.
