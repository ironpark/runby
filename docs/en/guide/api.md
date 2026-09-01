# API reference

Read [Getting started](getting-started.md) first if you are new to the package. This page describes the public options and result model.

Use cached `Current()` in normal applications, `Detect(opts ...Option)` for a fresh or configured detection, and `Register()` to add process-wide user-defined drivers.

```go
result := runby.Current()
result := runby.Detect()
result := runby.Detect(runby.WithEnviron(environ))
result := runby.Detect(runby.WithoutTTY())
runby.Register(myDriver)
```

## Options

Most callers need no options.

| Option | Meaning |
|---|---|
| `WithEnviron([]string)` | Supply a `"NAME=value"` environment |
| `WithTTY(TTY)` | Inject standard-stream state |
| `WithProcessTree(ProcessTree)` | Inject an ancestor chain |
| `WithoutTTY()` | Skip TTY system calls |
| `WithoutProcessTree()` | Skip ancestor inspection |
| `WithEnv(Env)` | Supply an arbitrary environment implementation |
| `WithDrivers(...Driver)` | Extend the built-in and registered set; replace matching identities |
| `WithOnlyDrivers(...Driver)` | Run exactly these drivers; ignore built-in and registered drivers |

Supplying an environment automatically disables TTY and process inspection because those facts may belong to another process. Inject them explicitly when they are known.

Use `Register(d)` in `init` for process-wide extension, `WithDrivers(d)` for one call, and `WithOnlyDrivers(d)` for isolated tests. Passing duplicate identities in one option call panics. CI and terminal axes use the first match; agents are sorted by their classification; remote and runner axes report every match.

## `Result`

```go
type Result struct {
	Agents   []Agent
	TTY      TTY
	Process  ProcessTree
	CI       CI
	Terminal Terminal
	Remotes  []Remote
	Runners  []Runner
}
```

Axis predicates answer whether anything was detected:

```go
result.IsAgent()
result.IsCI()
result.HasTerminal()
result.IsRemote()
result.HasRunner()
```

Named accessors return `(value, ok)`:

```go
result.Agent(runby.AgentCodex)
result.Remote(runby.RemoteSSH)
result.Runner(runby.RunnerNPM)
result.RunnerOfKind(runby.RunnerKindService)
```

Additional helpers:

```go
result.Primary()      // outermost representative agent
result.Chain()        // "paseo>codex" or "unknown"
result.SessionID()    // Identifier and source agent
result.AgentID()      // Identifier and source agent
result.Multiplexer()  // first multiplexer layer
result.Unattended()   // cross-axis unattended policy
```

### `Identifier`

Session and logical-agent identifiers travel with the layer that advertised them because identifiers from different products are not interchangeable.

```go
type Identifier struct {
	Value string
	Agent AgentName
}
```

### `Unattended()`

This is the only method that combines axes. It is true if any of these hold:

| Condition | Reason |
|---|---|
| A `definite` agent layer | Agents may allocate a PTY even when no person is behind it |
| `IsCI()` | Output is collected by CI |
| A `RunnerKindService` runner | Output belongs to a service manager or journal |
| `TTY.Inspected && !TTY.Interactive` | Streams were inspected and cannot support a prompt |

A `probable` agent alone is intentionally insufficient: the marker may describe a product-owned terminal where a person is typing. An uninspected TTY is also not evidence. Use this method for display defaults, not as a security boundary.

## Common axis fields

Every environment-derived axis embeds:

```go
type Axis struct {
	Confidence Confidence
	Extra      map[string]string
	Evidence   []string
}
```

`Evidence` contains environment-variable names only. `Extra` stores product-specific normalized values under keys such as `"<slug>.<name>"`.

An agent layer additionally contains its identity and classification, session fields, sandbox information, paths, and optional live-ancestor corroboration:

```go
type Agent struct {
	Name   AgentName
	Kind   Kind
	Models ModelSource
	Axis

	SessionID  string
	AgentID    string
	Entrypoint string
	Nested     bool
	Sandbox    Sandbox
	Paths      Paths
	AncestorPID int
}
```

`AncestorPID` also exists on terminal, remote, and runner results. Zero is not negative evidence. CI has no ancestor field because a CI job is not an ancestor executable.

Every enum renders its zero value as `"unknown"`, including agent, classification, confidence, network, CI provider, terminal, remote, and runner enums.

## Caching

```go
result := runby.Current() // computes Detect() once per process
```

The first caller fills the process-global cache. Tests should avoid setting environment variables and then calling `Current()`, because another test may have populated it first. Construct a result explicitly instead:

```go
result := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
```

Application code is easier to test when it obtains one `Result` at the entry point and passes it downward.

## Driver extension

Each axis has a driver structure, and built-in and user-defined drivers use the same types.

```go
acme := runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Kind:        runby.KindOrchestrator,
	Models:      runby.ModelsDelegated,
	Executables: []string{"acme-run"},
	Detect: func(env runby.Env) (runby.Agent, bool) {
		r := runby.NewEnvReader(env)
		id, ok := r.Value("ACME_RUN_ID")
		if !ok {
			return runby.Agent{}, false
		}
		return runby.Agent{AgentID: id, Axis: runby.Axis{Evidence: r.Evidence()}}, true
	},
}

result := runby.Detect(runby.WithDrivers(acme))
runby.Register(acme) // process-wide; call from init
```

Agent order derives from `Kind` and `Models`: orchestrators, multi-vendor harnesses, then first-party harnesses. A driver with neither classification sorts last.

| Axis | Driver | Identity | Classification | Executables |
|---|---|---|---|---|
| Agent | `AgentDriver` | `Agent` | `Kind` + `Models` | Yes |
| CI | `CIDriver` | `Provider` | — | No |
| Terminal | `TerminalDriver` | `Program` | — | Yes |
| Remote | `RemoteDriver` | `Platform` | `Kind` | Yes |
| Runner | `RunnerDriver` | `Tool` | `Kind` | Yes |

A remote driver with `RemoteKindMultiplexer` enables stale-environment handling. See [Writing drivers](drivers.md) for packaging and registration.

### `EnvReader`

`EnvReader` records the names it reads so evidence cannot silently drift from detection logic.

```go
r := runby.NewEnvReader(env)
id, ok := r.Value("ACME_RUN_ID")
if !ok {
	return runby.Agent{}, false
}
home := r.First("ACME_HOME", "ACME_ROOT")
return runby.Agent{
	SessionID: id,
	Paths: runby.Paths{DataDirectory: home},
	Axis: runby.Axis{Evidence: r.Evidence()},
}, true
```

| Method | Purpose |
|---|---|
| `Value` | Read one nonempty value |
| `Bool`, `IsTrue`, `EqualsFold` | Parse or compare a value |
| `Any` | Test candidates and record each name |
| `First` | Return the first set value |
| `Extra` | Build an `Extra` map |
| `Peek` | Read without recording |
| `Record` | Record without reading |
| `Evidence` | Return sorted, deduplicated names that were set |

An `EnvReader` is not safe for concurrent use. Create one per detection call and do not retain it.
