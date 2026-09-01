# Runner axis (`Result.Runners`)

This axis answers what directly launched the process: a package-manager script, build recipe, hook framework, or service manager.

```go
result := runby.Detect()
result.HasRunner()
result.Runner(runby.RunnerNPM)
result.RunnerOfKind(runby.RunnerKindService)
```

## Detected runners

| Tool | `RunnerTool` | `Kind` | Marker | `Task` |
|---|---|---|---|---|
| npm | `RunnerNPM` | `script` | npm user agent begins with `npm/` | Script name |
| pnpm | `RunnerPNPM` | `script` | begins with `pnpm/` | Script name |
| Bun | `RunnerBun` | `script` | begins with `bun/` | Script name |
| GNU Make | `RunnerMake` | `script` | `MAKELEVEL` | Empty |
| systemd | `RunnerSystemd` | `service` | `INVOCATION_ID` | Empty |
| pre-commit | `RunnerPreCommit` | `hook` | `PRE_COMMIT=1` | Empty |

`script` means the command came from project configuration, `hook` responds to an event, and `service` runs under a service manager with nobody directly watching its output.

Several runners can be present simultaneously. Their order is detection order, not proof of nesting. This axis is also independent from CI.

## What cannot be detected

- **Ordinary Git hooks:** Git exposes no environment variable unique to all hook execution. pre-commit is detectable only because the framework adds its own marker.
- **Cron:** It adds no reliable execution identity variable.
- **Yarn:** Excluded until its contract is verified by an official source or measurement.

## Pitfalls

- Match package-manager user-agent prefixes; searching for `npm` as a substring misclassifies pnpm and Bun.
- `MAKEFLAGS` may be empty, so `MAKELEVEL` is the marker.
- `JOURNAL_STREAM` alone is not a systemd marker.
- `npm_lifecycle_script` is never read because it may contain inline credentials.
- Absence is not negative evidence; an environment can be cleared.

See [runner research](../../research/runners/) for source evidence.
