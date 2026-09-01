# Recipes

Common patterns for connecting `runby` results to application behavior.

## Decide interactive behavior safely

`TTY.Interactive` directly answers whether the process can prompt and receive input.

```go
result := runby.Current()
if !result.TTY.Interactive {
	return errors.New("this operation requires interactive input")
}
askForConfirmation()
```

Do not use `IsAgent()` alone to decide whether prompting is possible. Agents can allocate PTYs, while human-invoked commands may run through a pipe or service without a TTY.

Use `Unattended()` when the question is whether anyone is watching the output, such as when deciding whether to draw spinners, color, or progress.

```go
result := runby.Current()
if result.Unattended() {
	disableSpinner()
	disableColor()
}
```

`Unattended()` is true for a `definite` agent layer, CI, a service runner, or an inspected but non-interactive TTY. A `probable` agent alone does not make it true, and an uninspected TTY provides no evidence. Terminal identity is intentionally excluded.

Combine axes directly if your policy differs. To treat hook runners as automated:

```go
automated := result.Unattended()
if _, ok := result.RunnerOfKind(runby.RunnerKindHook); ok {
	automated = true
}
```

These signals strengthen a conclusion; false does not prove that a person launched the process. Cron and ordinary Git hooks, for example, expose no reliable environment marker.

## Change behavior for a specific agent

```go
result := runby.Current()

if result.IsAgent() {
	disableSpinner()
}
if codex, ok := result.Agent(runby.AgentCodex); ok &&
	codex.Sandbox.Network == runby.NetworkDisabled {
	skipNetworkChecks()
}
```

Nested execution can produce several agent layers. `Primary()` returns the outermost representative layer; `Chain()` returns the full stack.

## Record execution context

Prefer stable, selected fields in operational logs.

```go
result := runby.Current()
log.Printf(
	"agent=%s ci=%s interactive=%t remote=%t runner=%t",
	result.Chain(), result.CI.Provider, result.TTY.Interactive,
	result.IsRemote(), result.HasRunner(),
)
```

`Chain()` and an undetected `CI.Provider` render as `"unknown"`, keeping log fields stable. A complete JSON `Result` can include session identifiers, working directories, and executable paths; select only the fields appropriate for external telemetry.

```go
type executionContext struct {
	Agent       string `json:"agent"`
	CI          string `json:"ci"`
	Interactive bool   `json:"interactive"`
}

context := executionContext{
	Agent: result.Chain(), CI: result.CI.Provider.String(),
	Interactive: result.TTY.Interactive,
}
```

## Distinguish CI from local execution

```go
if result.IsCI() {
	configureForCI(result.CI.Provider, result.CI.Attempt)
} else {
	configureForLocalRun()
}
```

An agent can run inside a CI job, so `IsAgent()` and `IsCI()` are not mutually exclusive modes.

## Distinguish scripts, hooks, and services

```go
if npm, ok := result.Runner(runby.RunnerNPM); ok {
	log.Printf("npm script=%s", npm.Task)
}
if _, ok := result.RunnerOfKind(runby.RunnerKindService); ok {
	disableHumanOrientedOutput()
}
```

Several runners may be present. Their array order is detection order, not proof of nesting.

## Detect potentially stale tmux state

```go
if mux, ok := result.Multiplexer(); ok {
	log.Printf("inherited environment may be stale inside %s", mux.Platform)
}
```

A multiplexer can retain a previous client's environment. `runby` lowers the confidence of uncorroborated terminal, agent, and runner findings. Consult [ancestor-process information](process.md) for important decisions. CI is excluded because it is not a layer launched through the multiplexer's stored environment.

## Branch from a shell script

```sh
if runby is agent || runby is ci; then
	export NO_COLOR=1
fi
if runby is runner; then
	echo "launched by another tool"
fi
```

Use `runby -v` for shareable diagnostics and `runby -json` for machine processing. See the [CLI guide](cli.md) for privacy notes and exit codes.
