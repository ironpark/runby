# CI 축

**CI는 `Agents`가 아니라 별도 축입니다.** Claude Code가 GitHub Actions 워크플로에서 실행되면 `KindHarness` 레이어와 CI 결과가 **동시에** 채워집니다. `Kind`는 "누가 명령을 요청했는가"를, `CI`는 "어디서 실행되는가"를 답합니다. 그래서 `Chain()`에는 CI가 들어가지 않습니다.

CI 여부만 필요하면 `IsCI()`, 플랫폼이나 재시도 회차가 필요할 때 `CI`의 상세 필드를 사용하세요.

현재 내장 감지 제공자는 GitHub Actions, Forgejo Actions, Gitea Actions, GitLab CI/CD, CircleCI, Travis CI, Buildkite, Azure Pipelines, Bitbucket Pipelines, Jenkins, Vercel, Netlify, TeamCity, Drone, AppVeyor, Semaphore, Cirrus CI, AWS CodeBuild, Google Cloud Build, Xcode Cloud, Cloudflare Pages, Cloudflare Workers Builds, Woodpecker CI, Bitrise, Render, Harness CI, Bamboo, GoCD, Taskcluster, Sourcehut, Codefresh, Codemagic, Buddy, Screwdriver, Vela와 일반 CI 관례입니다. 감지 순서와 제품별 근거는 [`docs/research/ci/`](../../research/ci/)에 있습니다.

```go
result := runby.Detect()
result.IsAgent()            // Claude Code가 실행했는가
result.IsCI()             // CI 잡에서 도는가
result.CI.Provider        // "github-actions"
result.CI.PipelineID      // GITHUB_RUN_ID
result.CI.JobID           // GITHUB_JOB
result.CI.Attempt         // GITHUB_RUN_ATTEMPT (1부터)
result.CI.Trigger         // GITHUB_EVENT_NAME ("push", "pull_request", ...)
result.CI.PullRequest     // PR/MR build로 광고됐는가
result.CI.PullRequestID   // 플랫폼이 광고한 PR/MR 식별자
```

`CI` 구조체는 플랫폼별 용어를 공통 필드로 정규화합니다.

| 필드 | 의미 |
|---|---|
| `PipelineID` | run/build/pipeline 단위 식별자 |
| `BuildNumber` | UI에 보이는 사람용 카운터. 프로젝트 간 유일하지 않음 |
| `JobID` / `JobName` | 파이프라인 내 개별 job·step |
| `Attempt` | **1부터 시작하는** 재시도 회차. 광고하지 않으면 `0` |
| `Trigger` | 플랫폼 자체 용어의 트리거 종류. 없으면 빈 문자열 |
| `Runner` | 잡을 실행 중인 머신·에이전트 |
| `PullRequest` | 플랫폼이 PR/MR 실행이라고 광고했을 때 `true`. 신호가 없으면 `false` |
| `PullRequestID` | 플랫폼이 광고한 PR/MR 식별자. 요청 종류만 알 수 있으면 빈 문자열 |
| `Extra` | 단일 플랫폼 전용 값. 키는 `"<provider-slug>.<name>"` |

정규화 시 두 가지를 처리합니다.

- **`Attempt`는 1-based로 통일합니다.** Buildkite `BUILDKITE_RETRY_COUNT`와 GitLab `CI_JOB_RETRY_COUNT`는 0부터 세는 재시도 횟수이므로 +1 해서 맞춥니다. GitHub `GITHUB_RUN_ATTEMPT`와 Azure `SYSTEM_JOBATTEMPT`는 이미 1부터라 그대로 씁니다.
- **Bitbucket UUID의 중괄호를 벗깁니다.** `BITBUCKET_PIPELINE_UUID`는 `{11d8...}` 형태로 오므로 `PipelineID`/`JobID`, 그리고 해당 플랫폼의 `Extra` 값에는 중괄호를 뺀 값이 들어갑니다.

PR/MR 감지는 플랫폼별 직접 신호만 사용합니다. 예를 들어 `BUILD_REASON=PullRequest`, `PULL_REQUEST=true`, `CI_PIPELINE_EVENT=pull_request`처럼 플랫폼이 요청 종류를 명시하거나, PR 번호·ID 변수를 제공하는 경우에만 `PullRequest`를 채웁니다. 일반 branch 이름이나 commit message는 추정에 사용하지 않습니다. `Evidence`에는 이 판정에 사용한 변수 이름만 들어가며 값은 들어가지 않습니다.

Forgejo Actions는 Runner v7+에서 모든 `FORGEJO_*`를 `GITHUB_*` 별칭으로도 제공하므로 GitHub Actions보다 **먼저** 검사합니다. v7 미만 Runner는 `GITHUB_*`만 제공해 환경변수로는 구별할 수 없어 GitHub Actions로 보고됩니다.

플랫폼별 조사 근거는 [`docs/research/ci/`](../../research/ci/)에 있습니다. 지원하지 않는 플랫폼은 [드라이버](drivers.md)를 만들어 추가할 수 있습니다.

```go
acme := runby.CIDriver{
	Provider: "acme-ci",
	Detect: func(env runby.Env) (runby.CI, bool) {
		r := runby.NewEnvReader(env)
		id, ok := r.Value("ACME_CI_BUILD")
		if !ok {
			return runby.CI{}, false
		}
		return runby.CI{PipelineID: id, Axis: runby.Axis{Evidence: r.Evidence()}}, true
	},
}

runby.Register(acme) // 또는 이 호출에만:
result := runby.Detect(runby.WithOnlyDrivers(append([]runby.Driver{acme}, runby.BuiltinDrivers()...)...))
```
