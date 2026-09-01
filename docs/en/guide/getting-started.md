# Getting started

This guide covers installing `runby`, selecting the result you need, and reproducing an execution environment in tests.

## 1. Install

Go 1.24 or later is required. In an existing Go module, run:

```sh
go get github.com/ironpark/runby
```

Then import the package:

```go
import "github.com/ironpark/runby"
```

If you only need shell integration, install the CLI instead:

```sh
go install github.com/ironpark/runby/cmd/runby@latest
runby
```

See the [CLI guide](cli.md) for subcommands and exit codes.

## 2. Read the current execution environment

Use `Current()` in normal applications. It detects once on the first call and reuses the result.

```go
result := runby.Current()

if result.IsAgent() {
	log.Printf("agent=%s", result.Chain())
}
if result.IsCI() {
	log.Printf("ci=%s", result.CI.Provider)
}
```

The result also includes runners, remote environments, terminal identity, TTY state, and ancestor processes.

## 3. Decide application behavior

When user input is required, inspect `TTY.Interactive` instead of guessing from an agent or terminal name.

```go
result := runby.Current()

if !result.TTY.Interactive {
	return errors.New("this operation requires an interactive terminal")
}
confirmBeforeDelete()
```

Check the corresponding axes when changing output for agents or CI.

```go
result := runby.Current()

machineRun := result.IsAgent() || result.IsCI()
if _, service := result.RunnerOfKind(runby.RunnerKindService); service {
	machineRun = true
}
if machineRun {
	disableProgressAnimation()
}
```

See [Recipes](recipes.md) for recommended conditions by use case.

## 4. Select details

```go
result := runby.Current()

result.Chain()
result.CI.Provider
result.Terminal.Program
result.Remote(runby.RemoteTmux)
result.Runner(runby.RunnerNPM)
```

Read product-specific fields only after checking that the product was detected.

```go
if codex, ok := result.Agent(runby.AgentCodex); ok {
	log.Printf("sandbox=%s network=%s", codex.Sandbox.Mode, codex.Sandbox.Network)
}
```

See the [API reference](api.md) for every result field.

## Choose between `Current()` and `Detect()`

| Situation | Recommended API |
|---|---|
| Read the current process from several places | `Current()` |
| Re-read the current environment every time | `Detect()` |
| Classify another or recorded environment | `Detect(WithEnviron(...))` |
| Use an environment fixture in tests | `Detect(WithEnviron(...))` |
| Skip TTY and process system calls | `Detect(WithoutTTY(), WithoutProcessTree())` |
| Isolate user-defined drivers for one call | `Detect(WithOnlyDrivers(...))` |

`Current()` caches its first result. Call `Detect()` directly if a later `os.Setenv` must be observed.

## Supply another environment

`WithEnviron` classifies an environment that may not belong to this process. This is useful for wrappers that read `/proc/<pid>/environ`, prepare `exec.Cmd.Env`, or analyze a recorded environment, and for deterministic tests.

```go
func TestGitHubActions(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=1234",
	}))

	if !result.IsCI() || result.CI.Provider != runby.CIGitHubActions {
		t.Fatalf("unexpected result: %#v", result.CI)
	}
}
```

Because an explicit environment may describe another process, `WithEnviron` and `WithEnv` disable automatic TTY and ancestor inspection. Supply those explicitly with `WithTTY()` or `WithProcessTree()` when needed.

## Interpret results carefully

Environment variables are inherited at process startup. `IsAgent()` therefore means that the process inherited a supported agent signal, not that the agent is necessarily still alive.

A nonzero `AncestorPID` means the environment signal was corroborated by a live ancestor. Zero is not negative evidence. See the [process guide](process.md).

Continue with [Recipes](recipes.md) or select a [concept guide](README.md#concept-guides).
