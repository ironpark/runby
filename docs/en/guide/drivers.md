# Writing and distributing drivers

Create a driver when `runby` does not recognize an environment.

There are two ways to use drivers:

| | `Register(...Driver)` | `WithOnlyDrivers(...Driver)` |
|---|---|---|
| Scope | Every later `Detect()` and `Current()` | One `Detect()` call |
| Built-ins | Extended; matching identities are replaced | Ignored |
| Best for | Reusable driver modules | Tests and fully controlled sets |

`WithDrivers` is the one-call counterpart to `Register`: it extends the default set for one detection.

```go
runby.Detect(runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), acme)...))
runby.Detect(runby.WithOnlyDrivers(append([]runby.Driver{acme}, runby.BuiltinDrivers()...)...))
```

CI and terminal axes use the first match. Agent order derives from classification; remote and runner axes report every match.

## Driver structures

```go
type Driver interface { /* implemented by the five driver structs */ }

type AgentDriver struct {
	Agent AgentName
	Kind Kind
	Models ModelSource
	Executables []string
	Detect func(Env) (Agent, bool)
}
```

The other structures follow the same shape:

- `CIDriver{Provider, Detect}`
- `TerminalDriver{Program, Executables, Detect}`
- `RemoteDriver{Platform, Kind, Executables, Detect}`
- `RunnerDriver{Tool, Kind, Executables, Detect}`

The driver supplies stable identity and classification; `Detect` supplies values observed for this run. Empty confidence and network fields receive defaults.

## Write detection logic

Use `EnvReader` so every environment variable read by a successful detector is reflected in `Evidence`.

```go
func detectAcme(env runby.Env) (runby.Agent, bool) {
	r := runby.NewEnvReader(env)
	id, ok := r.Value("ACME_RUN_ID")
	if !ok {
		return runby.Agent{}, false
	}
	return runby.Agent{
		AgentID: id,
		Axis: runby.Axis{Evidence: r.Evidence()},
	}, true
}
```

Never use API keys or user configuration as execution markers. A marker should be set by the product on child processes it launched. Never put environment values in `Evidence`.

## Register a reusable module

A driver package can register itself from `init`:

```go
package acmerunby

import "github.com/ironpark/runby"

func init() {
	runby.Register(runby.AgentDriver{
		Agent: "acme-agent",
		Kind: runby.KindHarness,
		Models: runby.ModelsMultiVendor,
		Executables: []string{"acme"},
		Detect: detectAcme,
	})
}
```

Consumers activate it with a blank import before the first call to `Current()`:

```go
import _ "example.com/acme-runby"
```

Calling `Register` after `Current()` has initialized the global registry panics. Duplicate axis identities also panic, except that a newly supplied driver intentionally replaces a matching driver already in the default set.

## Test in isolation

```go
func TestAcme(t *testing.T) {
	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=42"}),
		runby.WithOnlyDrivers(acmeDriver),
	)
	if agent, ok := result.Agent("acme-agent"); !ok || agent.AgentID != "42" {
		t.Fatalf("unexpected agent: %#v", agent)
	}
}
```

Use `WithDrivers` when a test needs the built-ins as well. `BuiltinDrivers()` returns a copy so it can be filtered safely.

## Ancestor corroboration and multiplexers

Declare normalized executable base names in `Executables` to enable live-ancestor labeling. For remote drivers, set `RemoteKindMultiplexer` when the layer can retain stale environments; this enables confidence degradation for uncorroborated environment-derived axes.

## Research notes

Every built-in driver has a source-backed document under [research](../../research/), and tests require matching slugs. User-defined drivers do not have this repository requirement, but recording the rationale makes future environment-contract changes much easier to audit.
