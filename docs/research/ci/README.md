# CI 플랫폼 문서

CI 플랫폼별 환경변수 감지 규칙을 정리하는 문서입니다. 양식과 front matter 필드의 의미는 [상위 문서](../README.md)와 같으며, `product_type`은 모두 `ci_platform`입니다.

## 에이전트와 CI는 별개의 축입니다

`runby`는 CI를 `Agent` 레이어로 보고하지 않고 `Result.CI`로 분리합니다. Claude Code가 GitHub Actions 워크플로에서 실행되는 경우처럼 **두 축이 동시에 참일 수 있기** 때문입니다. `Kind`는 "누가 명령을 요청했는가"를, `CI`는 "어디서 실행되는가"를 답합니다.

## 감지 신호 요약

`PR 감지` 열의 값이 있으면 `CI.PullRequest=true`로 정규화하고, 별도 ID가 있는 경우 `CI.PullRequestID`에 넣습니다. 표의 순서는 드라이버 우선순위이며 generic은 항상 마지막입니다.

| 플랫폼 | 실행 마커 | 파이프라인 ID | Job ID | 재시도 회차 | 트리거 | PR 감지 |
|---|---|---|---|---|---|---|
| [Forgejo Actions](forgejo-runner.md) | `FORGEJO_ACTIONS=true` (Runner v7+) | `FORGEJO_RUN_ID` | `FORGEJO_JOB` | `FORGEJO_RUN_ATTEMPT` (1부터) | `FORGEJO_EVENT_NAME` | `FORGEJO_EVENT_NAME=pull_request` |
| [Gitea Actions](gitea-actions.md) | `GITEA_ACTIONS=true` | `GITHUB_RUN_ID` | `GITHUB_JOB` | 없음 | `GITHUB_EVENT_NAME` | `GITHUB_EVENT_NAME=pull_request` |
| [GitHub Actions](github-actions.md) | `GITHUB_ACTIONS=true` | `GITHUB_RUN_ID` | `GITHUB_JOB` | `GITHUB_RUN_ATTEMPT` (1부터) | `GITHUB_EVENT_NAME` | `GITHUB_EVENT_NAME=pull_request` |
| [GitLab CI/CD](gitlab-ci.md) | `GITLAB_CI=true` | `CI_PIPELINE_ID` | `CI_JOB_ID` | `CI_JOB_RETRY_COUNT` (0부터, 19.3+) | `CI_PIPELINE_SOURCE` | `CI_MERGE_REQUEST_ID` |
| [CircleCI](circleci.md) | `CIRCLECI=true` | `CIRCLE_PIPELINE_ID` | `CIRCLE_WORKFLOW_JOB_ID` | 없음 | 환경변수로는 없음 | `CIRCLE_PULL_REQUEST` |
| [Travis CI](travis-ci.md) | `TRAVIS=true` | `TRAVIS_BUILD_ID` | `TRAVIS_JOB_ID` | 없음 (`TRAVIS_JOB_RESTARTED` 플래그만) | `TRAVIS_EVENT_TYPE` | `TRAVIS_PULL_REQUEST≠false` |
| [Buildkite](buildkite.md) | `BUILDKITE=true` | `BUILDKITE_BUILD_ID` | `BUILDKITE_JOB_ID` | `BUILDKITE_RETRY_COUNT` (0부터) | `BUILDKITE_SOURCE` | `BUILDKITE_PULL_REQUEST≠false` |
| [Azure Pipelines](azure-pipelines.md) | `TF_BUILD=True` | `BUILD_BUILDID` | `SYSTEM_JOBID` | `SYSTEM_JOBATTEMPT` (1부터) | `BUILD_REASON` | `BUILD_REASON=PullRequest` |
| [Bitbucket Pipelines](bitbucket-pipelines.md) | 전용 마커 없음 | `BITBUCKET_PIPELINE_UUID` | `BITBUCKET_STEP_UUID` | `BITBUCKET_STEP_RUN_NUMBER` (1부터) | 없음 | `BITBUCKET_PR_ID` |
| [Jenkins](jenkins.md) | 전용 마커 없음 | `BUILD_ID` (= `BUILD_NUMBER`) | 없음 | 없음 | 코어에는 없음 | `ghprbPullId` 또는 `CHANGE_ID` |
| [Vercel](vercel.md) | `VERCEL` 또는 `NOW_BUILDER` | `VERCEL_DEPLOYMENT_ID` | 없음 | 없음 | 없음 | `VERCEL_GIT_PULL_REQUEST_ID` |
| [Netlify](netlify.md) | `NETLIFY` | `DEPLOY_ID` | `BUILD_ID` | 없음 | `CONTEXT` | `PULL_REQUEST≠false` |
| [TeamCity](teamcity.md) | `TEAMCITY_VERSION` | `BUILD_ID` | 없음 | 없음 | 없음 | 없음 |
| [Drone](drone.md) | `DRONE=true` | `DRONE_BUILD_NUMBER` | `DRONE_STAGE_NUMBER` | 없음 | `DRONE_BUILD_EVENT` | `DRONE_BUILD_EVENT=pull_request` |
| [AppVeyor](appveyor.md) | `APPVEYOR=true` | `APPVEYOR_BUILD_ID` | `APPVEYOR_JOB_ID` | 없음 | 없음 | `APPVEYOR_PULL_REQUEST_NUMBER` |
| [Semaphore](semaphore.md) | `SEMAPHORE=true` | `SEMAPHORE_PIPELINE_ID` | `SEMAPHORE_JOB_ID` | 없음 | `SEMAPHORE_GIT_REF_TYPE` | `PULL_REQUEST_NUMBER` 또는 `SEMAPHORE_GIT_PR_NUMBER` |
| [Cirrus CI](cirrus-ci.md) | `CIRRUS_CI` | `CIRRUS_BUILD_ID` | `CIRRUS_TASK_ID` | 없음 | 없음 | `CIRRUS_PR` |
| [AWS CodeBuild](aws-codebuild.md) | `CODEBUILD_BUILD_ARN` | `CODEBUILD_BUILD_ARN` | `CODEBUILD_BUILD_ID` | 없음 | `CODEBUILD_WEBHOOK_EVENT` | `CODEBUILD_WEBHOOK_EVENT`가 `PULL_REQUEST_`로 시작 |
| [Google Cloud Build](google-cloud-build.md) | `BUILDER_OUTPUT` | 없음 | 없음 | 없음 | 없음 | 없음 |
| [Xcode Cloud](xcode-cloud.md) | `CI_XCODE_PROJECT` | `CI_BUILD_ID` | `CI_WORKFLOW_ID` | 없음 | 없음 | `CI_PULL_REQUEST_NUMBER` |
| [Cloudflare Pages](cloudflare-pages.md) | `CF_PAGES` | 없음 | 없음 | 없음 | 없음 | 없음 |
| [Cloudflare Workers Builds](cloudflare-workers.md) | `WORKERS_CI` | `WORKERS_CI_BUILD_UUID` | 없음 | 없음 | 없음 | 없음 |
| [Woodpecker CI](woodpecker.md) | `CI=woodpecker` | `CI_BUILD_NUMBER` | `CI_JOB_NUMBER` | 없음 | `CI_BUILD_EVENT` 또는 `CI_PIPELINE_EVENT` | event 또는 PR 번호 변수 |
| [Bitrise](bitrise.md) | `BITRISE_IO=true` | `BITRISE_BUILD_SLUG` | 없음 | 없음 | `BITRISE_TRIGGER_BY` | `BITRISE_PULL_REQUEST` |
| [Render](render.md) | `RENDER=true` | 없음 | 없음 | 없음 | 없음 | `IS_PULL_REQUEST=true` |
| [Harness CI](harness-ci.md) | `HARNESS_BUILD_ID` | `HARNESS_BUILD_ID` | `HARNESS_STAGE_ID` | 없음 | 없음 | 없음 |
| [Bamboo](bamboo.md) | `bamboo_planKey` | `bamboo_buildResultKey` | `bamboo_buildKey` | 없음 | 없음 | `bamboo_repository_pr_key` |
| [GoCD](gocd.md) | `GO_PIPELINE_LABEL` | `GO_PIPELINE_LABEL` | `GO_JOB_NAME` | 없음 | 없음 | 없음 |
| [Taskcluster](taskcluster.md) | `TASK_ID` + `RUN_ID` | `TASK_ID` | 없음 | `RUN_ID` (0부터) | 없음 | 없음 |
| [Sourcehut builds](sourcehut.md) | `CI_NAME=sourcehut` | `JOB_ID` | 없음 | 없음 | 없음 | 없음 |
| [Codefresh](codefresh.md) | `CF_BUILD_ID` | `CF_BUILD_ID` | 없음 | 없음 | 없음 | `CF_PULL_REQUEST_NUMBER` 또는 `CF_PULL_REQUEST_ID` |
| [Codemagic](codemagic.md) | `CM_BUILD_ID` | `CM_BUILD_ID` | 없음 | 없음 | `CM_TRIGGER_SOURCE` | `CM_PULL_REQUEST=true` + 번호 |
| [Buddy](buddy.md) | `BUDDY_WORKSPACE_ID` | `BUDDY_EXECUTION_ID` | `BUDDY_ACTION_ID` | 없음 | `BUDDY_PIPELINE_TRIGGER_MODE` | `BUDDY_EXECUTION_PULL_REQUEST_*` 또는 `BUDDY_RUN_PR_*` |
| [Screwdriver](screwdriver.md) | `SCREWDRIVER=true` | `SD_BUILD_ID` | `SD_JOB_ID` | 없음 | 없음 | `SD_PULL_REQUEST≠false` |
| [Vela](vela.md) | `VELA` | `VELA_BUILD_NUMBER` | 없음 | 없음 | `VELA_BUILD_EVENT` | `VELA_PULL_REQUEST=1` |
| 일반 CI | `CI=true`, `CONTINUOUS_INTEGRATION=true`, 또는 값이 있는 `CI_NAME` | 없음 | 없음 | 없음 | 없음 | 없음 |

## 전용 마커가 없는 두 플랫폼

Bitbucket Pipelines와 Jenkins만 `<PLATFORM>=true` 형태의 불리언 마커를 제공하지 않아 별도 판정 규칙이 필요합니다.

- **Bitbucket Pipelines** — `BITBUCKET_BUILD_NUMBER`가 항상 존재해 마커 역할을 겸하지만 값이 단순 정수이므로, `BITBUCKET_PIPELINE_UUID` 또는 `CI=true`를 함께 요구해 오탐을 막습니다.
- **Jenkins** — `BUILD_ID`, `BUILD_NUMBER`, `JOB_NAME`, `NODE_NAME`은 다른 도구도 쓰는 매우 일반적인 이름입니다. 따라서 `BUILD_NUMBER`와 함께 Jenkins 소유 변수(`JENKINS_URL`, `JENKINS_HOME`, `HUDSON_URL`, `HUDSON_HOME`) 중 하나가 있을 때만 감지합니다. `JENKINS_URL`만 요구하면 안 되는데, 이 변수는 관리자가 Jenkins root URL을 설정한 경우에만 주입되는 반면 `JENKINS_HOME`은 항상 설정되기 때문입니다.

  `BUILD_ID`는 Jenkins 1.597부터 `BUILD_NUMBER`와 값이 같습니다. Jenkins에는 불투명한 실행 ID가 없어 `PipelineID`도 이 값을 씁니다.

## Forgejo와 GitHub Actions의 판정 순서

Forgejo Runner v7 이상은 모든 `FORGEJO_*` 변수를 같은 suffix의 `GITHUB_*` 별칭으로도 제공합니다. 따라서 `FORGEJO_ACTIONS=true`를 `GITHUB_ACTIONS=true`보다 먼저 검사해야 Forgejo를 GitHub Actions로 오인하지 않습니다. v7 미만 Runner는 `GITHUB_*`만 제공하므로 환경변수만으로 두 제품을 확정적으로 구별할 수 없습니다.

## 정규화 규칙

`runby`는 플랫폼별 용어를 공통 `CI` 구조체로 옮기면서 두 가지를 정규화합니다.

- **`Attempt`는 1부터 세는 회차로 통일합니다.** Buildkite와 GitLab은 0부터 세는 재시도 *횟수*를 제공하므로 +1 합니다. Taskcluster의 `RUN_ID`도 0부터 시작하므로 +1 합니다. GitHub과 Azure는 이미 1부터 세는 *회차*라 그대로 씁니다.
- **Bitbucket UUID의 중괄호를 벗깁니다.** 값이 `{11d8...}` 형태로 오므로 `PipelineID`와 `JobID`에는 중괄호를 제거한 값을 넣습니다.

플랫폼 하나만 광고하는 값은 `CI.Extra`에 `"<slug>.<name>"` 키로 보존합니다.

## generic 휴리스틱 검토

ci-info가 넓게 쓰는 `BUILD_ID`, `BUILD_NUMBER`, `RUN_ID` 계열 중 단독 신호는 runby 기준에서 제외했습니다. 범용 이름이고 일반 build script·Jenkins·다른 제품과 겹쳐 CI 실행을 증명하지 못하기 때문입니다.

다음 관례만 generic으로 채택합니다.

- `CI=true` 또는 `CONTINUOUS_INTEGRATION=true` — 널리 통용되지만 제품을 특정하지 못하므로 `ConfidenceProbable`입니다.
- 값이 있는 `CI_NAME` — 값이 제품명을 담는 관례일 수 있어 generic probable로 보고합니다. `CI_NAME=sourcehut`처럼 제품명이 정확히 확인되는 값은 전용 드라이버가 먼저 가져갑니다.

`CI=false`와 `CONTINUOUS_INTEGRATION=false`는 감지하지 않습니다. 모든 generic Evidence에는 변수 이름만 남기고 값은 보존하지 않습니다.

## 채택하지 않은 후보

- **Heroku** — ci-info의 `NODE` 경로 문자열(`/app/.heroku/node/bin/node`) 매칭은 실행 환경이 주입한 전용 변수나 고정 마커가 아닙니다. 경로가 바뀌거나 사용자가 위조할 수 있어 runby의 증거 기준을 통과하지 못했습니다.
- **Codeship** — `CI_NAME=codeship`은 값 기반으로 generic에 포함할 수 있지만, 현재 활성 CI 제공자로 별도 드라이버를 유지할 공식 실행 환경 근거를 확인하지 못했습니다. 따라서 `CICodeShip`은 만들지 않고 `CIGeneric` probable로만 보고합니다.
- **Magnum, Solano, Strider, dsari 등** — ci-info에 남은 레거시 vendor 항목이지만 서비스가 사실상 사멸했거나 현재 공식 실행 환경 계약을 확인할 수 없습니다. 전용 드라이버와 상수는 추가하지 않습니다.

`CI=true`는 사실상의 업계 관행이지만 어느 플랫폼의 소유도 아니며 로컬 도구도 설정하므로, 플랫폼을 특정하지 못했을 때만 `CIGeneric`으로 보고하고 `ConfidenceProbable`을 넘지 않습니다.

## 공통 주의사항

환경변수는 자식 프로세스에 상속되고 사용자가 위조할 수 있으므로 신뢰 경계가 아닙니다. CI 감지는 실행 컨텍스트를 알리는 용도이며 권한 판단의 근거로 쓰면 안 됩니다.
