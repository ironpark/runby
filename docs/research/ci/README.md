# CI 플랫폼 문서

CI 플랫폼별 환경변수 감지 규칙을 정리하는 문서입니다. 양식과 front matter 필드의 의미는 [상위 문서](../README.md)와 같으며, `product_type`은 모두 `ci_platform`입니다.

## 에이전트와 CI는 별개의 축입니다

`runby`는 CI를 `Agent` 레이어로 보고하지 않고 `Result.CI`로 분리합니다. Claude Code가 GitHub Actions 워크플로에서 실행되는 경우처럼 **두 축이 동시에 참일 수 있기** 때문입니다. `Kind`는 "누가 명령을 요청했는가"를, `CI`는 "어디서 실행되는가"를 답합니다.

## 감지 신호 요약

| 플랫폼 | 실행 마커 | 파이프라인 ID | Job ID | 재시도 회차 | 트리거 |
|---|---|---|---|---|---|
| [Forgejo Actions](forgejo-runner.md) | `FORGEJO_ACTIONS=true` (Runner v7+) | `FORGEJO_RUN_ID` | `FORGEJO_JOB` | `FORGEJO_RUN_ATTEMPT` (1부터) | `FORGEJO_EVENT_NAME` |
| [GitHub Actions](github-actions.md) | `GITHUB_ACTIONS=true` | `GITHUB_RUN_ID` | `GITHUB_JOB` | `GITHUB_RUN_ATTEMPT` (1부터) | `GITHUB_EVENT_NAME` |
| [GitLab CI/CD](gitlab-ci.md) | `GITLAB_CI=true` | `CI_PIPELINE_ID` | `CI_JOB_ID` | `CI_JOB_RETRY_COUNT` (0부터, 19.3+) | `CI_PIPELINE_SOURCE` |
| [CircleCI](circleci.md) | `CIRCLECI=true` | `CIRCLE_PIPELINE_ID` | `CIRCLE_WORKFLOW_JOB_ID` | 없음 | 환경변수로는 없음 |
| [Travis CI](travis-ci.md) | `TRAVIS=true` | `TRAVIS_BUILD_ID` | `TRAVIS_JOB_ID` | 없음 (`TRAVIS_JOB_RESTARTED` 플래그만) | `TRAVIS_EVENT_TYPE` |
| [Buildkite](buildkite.md) | `BUILDKITE=true` | `BUILDKITE_BUILD_ID` | `BUILDKITE_JOB_ID` | `BUILDKITE_RETRY_COUNT` (0부터) | `BUILDKITE_SOURCE` |
| [Azure Pipelines](azure-pipelines.md) | `TF_BUILD=True` | `BUILD_BUILDID` | `SYSTEM_JOBID` | `SYSTEM_JOBATTEMPT` (1부터) | `BUILD_REASON` |
| [Bitbucket Pipelines](bitbucket-pipelines.md) | 전용 마커 없음 | `BITBUCKET_PIPELINE_UUID` | `BITBUCKET_STEP_UUID` | `BITBUCKET_STEP_RUN_NUMBER` (1부터) | 없음 |
| [Jenkins](jenkins.md) | 전용 마커 없음 | `BUILD_ID` (= `BUILD_NUMBER`) | 없음 | 없음 | 코어에는 없음 |

## 전용 마커가 없는 두 플랫폼

Bitbucket Pipelines와 Jenkins만 `<PLATFORM>=true` 형태의 불리언 마커를 제공하지 않아 별도 판정 규칙이 필요합니다.

- **Bitbucket Pipelines** — `BITBUCKET_BUILD_NUMBER`가 항상 존재해 마커 역할을 겸하지만 값이 단순 정수이므로, `BITBUCKET_PIPELINE_UUID` 또는 `CI=true`를 함께 요구해 오탐을 막습니다.
- **Jenkins** — `BUILD_ID`, `BUILD_NUMBER`, `JOB_NAME`, `NODE_NAME`은 다른 도구도 쓰는 매우 일반적인 이름입니다. 따라서 `BUILD_NUMBER`와 함께 Jenkins 소유 변수(`JENKINS_URL`, `JENKINS_HOME`, `HUDSON_URL`, `HUDSON_HOME`) 중 하나가 있을 때만 감지합니다. `JENKINS_URL`만 요구하면 안 되는데, 이 변수는 관리자가 Jenkins root URL을 설정한 경우에만 주입되는 반면 `JENKINS_HOME`은 항상 설정되기 때문입니다.

  `BUILD_ID`는 Jenkins 1.597부터 `BUILD_NUMBER`와 값이 같습니다. Jenkins에는 불투명한 실행 ID가 없어 `PipelineID`도 이 값을 씁니다.

## Forgejo와 GitHub Actions의 판정 순서

Forgejo Runner v7 이상은 모든 `FORGEJO_*` 변수를 같은 suffix의 `GITHUB_*` 별칭으로도 제공합니다. 따라서 `FORGEJO_ACTIONS=true`를 `GITHUB_ACTIONS=true`보다 먼저 검사해야 Forgejo를 GitHub Actions로 오인하지 않습니다. v7 미만 Runner는 `GITHUB_*`만 제공하므로 환경변수만으로 두 제품을 확정적으로 구별할 수 없습니다.

## 정규화 규칙

`runby`는 플랫폼별 용어를 공통 `CI` 구조체로 옮기면서 두 가지를 정규화합니다.

- **`Attempt`는 1부터 세는 회차로 통일합니다.** Buildkite와 GitLab은 0부터 세는 재시도 *횟수*를 제공하므로 +1 합니다. GitHub과 Azure는 이미 1부터 세는 *회차*라 그대로 씁니다.
- **Bitbucket UUID의 중괄호를 벗깁니다.** 값이 `{11d8...}` 형태로 오므로 `PipelineID`와 `JobID`에는 중괄호를 제거한 값을 넣습니다.

플랫폼 하나만 광고하는 값은 `CI.Extra`에 `"<slug>.<name>"` 키로 보존합니다.

## 공통 주의사항

`CI=true`는 사실상의 업계 관행이지만 어느 플랫폼의 소유도 아니며 로컬 도구도 설정하므로, 플랫폼을 특정하지 못했을 때만 `CIGeneric`으로 보고하고 `ConfidenceProbable`을 넘지 않습니다.

환경변수는 자식 프로세스에 상속되고 사용자가 위조할 수 있으므로 신뢰 경계가 아닙니다. CI 감지는 실행 컨텍스트를 알리는 용도이며 권한 판단의 근거로 쓰면 안 됩니다.
