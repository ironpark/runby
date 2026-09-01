# Process axis

Ancestor processes are the only evidence not derived from environment variables and provide the strongest available corroboration.

```go
tree := runby.Detect().Process
tree.Supported
tree.Ancestors
tree.FindAgent(runby.AgentCodex)
```

`runby -v` prints the same chain. A labeled ancestor connects an executable to a product; an axis result's `AncestorPID` means that live process corroborated the environment detection.

## Corroboration

```go
for _, agent := range runby.Current().Agents {
	if agent.AncestorPID != 0 {
		// Environment and a live ancestor agree.
	}
}
```

Agents, terminals, remote layers, and runners can be corroborated. CI cannot, because a CI job is not an ancestor executable.

A live terminal ancestor prevents multiplexer confidence degradation. Multiplexer servers daemonize and detach from their original terminal, so seeing the terminal as an ancestor shows that the result is not merely stale state behind a pane.

User-defined drivers participate when they declare `Executables`.

`AncestorPID == 0` is not negative evidence. Inspection can stop at another user's process, be unsupported, or miss launchers that no longer remain ancestors.

## Platforms

| Platform | Method |
|---|---|
| Linux | `/proc/<pid>/stat`, `/proc/<pid>/exe`, `/proc/<pid>/comm` |
| macOS | `sysctl(KERN_PROC_PID)` and `KERN_PROCARGS2` |
| Windows | One `CreateToolhelp32Snapshot` snapshot |
| Other | `Supported == false`, empty chain |

Like TTY state, the process axis describes the current process and is disabled by `WithEnviron` and `WithEnv`. Use `WithoutProcessTree()` to skip its cost explicitly.
