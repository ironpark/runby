# CI axis

CI is independent from agents. Claude Code running inside GitHub Actions populates both `Agents` and `CI`; CI is not included in `Chain()`.

```go
result := runby.Detect()
result.IsCI()
result.CI.Provider
result.CI.PipelineID
result.CI.JobID
result.CI.Attempt
result.CI.Trigger
result.CI.PullRequest
result.CI.PullRequestID
```

Built-in detection covers the providers listed in the root README plus generic CI conventions. Detection order and evidence are documented in [CI research](../../research/ci/).

## Normalized fields

| Field | Meaning |
|---|---|
| `PipelineID` | Run, build, or pipeline identifier |
| `BuildNumber` | Human-facing counter; not globally unique |
| `JobID`, `JobName` | Individual job or step |
| `Attempt` | One-based attempt number; zero if unavailable |
| `Trigger` | Provider-specific trigger name |
| `Runner` | Machine or agent executing the job |
| `PullRequest` | Provider explicitly advertised a PR or MR run |
| `PullRequestID` | Provider-advertised PR or MR identifier |
| `Extra` | Provider-specific normalized values |

Retry counters are normalized to one-based attempts. Bitbucket UUID braces are removed. Pull-request detection uses direct provider signals only; branch names and commit messages are not guessed.

Forgejo is checked before GitHub Actions because modern Forgejo Runner exposes both `FORGEJO_*` and `GITHUB_*` aliases. Older runners that expose only `GITHUB_*` cannot be distinguished and are reported as GitHub Actions.

Add an unsupported platform with a [driver](drivers.md).
