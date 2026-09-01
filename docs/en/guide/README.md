# User guide

English | [한국어](../../ko/guide/) | [日本語](../../ja/guide/)

If you are new to `runby`, start with [Getting started](getting-started.md). For code you can apply directly, choose a pattern from [Recipes](recipes.md).

## Find documentation by task

| Goal | Start here |
|---|---|
| Add `runby` to a Go program | [Getting started](getting-started.md) |
| Control prompts, color, and progress safely | [Recipes](recipes.md#decide-interactive-behavior-safely) |
| Branch in a shell script | [CLI guide](cli.md#use-it-from-a-shell) |
| Record results in logs or JSON | [Recipes](recipes.md#record-execution-context) |
| Reproduce an environment in tests | [Getting started](getting-started.md#supply-another-environment) |
| Add an unsupported environment | [Writing drivers](drivers.md) |
| Find every field and option | [API reference](api.md) |
| Check the basis for a detection rule | [Research documents](../../research/) |

## Suggested reading order

1. Use [Getting started](getting-started.md) to choose between `Current()` and `Detect()`.
2. Pick an application pattern from [Recipes](recipes.md).
3. Read only the concept guides for the axes you need.
4. Consult [API](api.md) and [Drivers](drivers.md) for advanced configuration and extension.

## Concept guides

| Guide | Question | Main result |
|---|---|---|
| [Agents](agents.md) | Did an AI agent launch this, and what is the stack? | `Agents`, `Chain()` |
| [CI](ci.md) | Which CI platform and job own this process? | `CI` |
| [Runners](runner.md) | Did npm, make, systemd, or another tool launch it? | `Runners` |
| [Remote environments](remote.md) | Did it pass through SSH, tmux, or a development container? | `Remotes` |
| [Terminal and TTY](terminal.md) | Which app created the environment, and is it interactive now? | `Terminal`, `TTY` |
| [Ancestor processes](process.md) | Is a detected launcher still a live ancestor? | `Process`, `AncestorPID` |

Several axes can be true simultaneously. If Claude Code runs `npm test` inside GitHub Actions, `Agents`, `CI`, and `Runners` are all populated. Select the fields that answer your question instead of looking for one execution mode.

## Interpretation rules

- Environment-based detection reflects inherited state from process startup. Use [ancestor-process corroboration](process.md#corroboration) when current liveness matters.
- `Terminal` identifies the emulator that created the environment; `TTY` reports current stream attachment. Use `TTY.Interactive` for prompts.
- `AncestorPID != 0` is strong positive evidence. Zero can result from permissions or process topology and is not negative evidence.
- `-json` can contain session identifiers and local paths. `runby -v` is safer for diagnostics intended for sharing.

Official sources, verification dates, and excluded products are recorded in the [research documents](../../research/).
