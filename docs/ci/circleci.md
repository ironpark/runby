---
title: CircleCI
slug: circleci
research_date: 2026-08-31
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 빌트인 환경변수 문서(circleci.com/docs)와 ID 계층 가이드로 판정 근거가 충분함. 실행 식별자(CIRCLECI, CIRCLE_PIPELINE_ID, CIRCLE_WORKFLOW_ID, CIRCLE_WORKFLOW_JOB_ID)는 job 컨테이너에 항상 주입되는 안정적 공개 계약이므로 별도 실제 파이프라인 실행 검증 없이도 감지 규칙을 확정할 수 있음
---

# CircleCI

CircleCI는 모든 job 컨테이너에 `CI`, `CIRCLECI`를 포함한 빌트인 환경변수 세트를 job 수준으로 주입한다고 공식 문서에 명시합니다. 이 변수들은 파이프라인(pipeline) 단계에서는 존재하지 않고 **job이 시작된 이후에만** 사용 가능하며, 파이프라인이나 워크플로 수준 로직에는 쓸 수 없습니다. 반면 `.circleci/config.yml`에서 참조하는 `pipeline.*` 값(예: `pipeline.trigger_source`)은 설정 처리 시점에만 존재하는 별개의 개념이며, 명시적으로 `environment` 키에 대입하지 않는 한 job 런타임 환경변수로 자동 노출되지 않습니다. 이 문서는 이 두 계층을 구분해서 다룹니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CIRCLECI` | 불리언 문자열 (`true`) | 실행 식별 | CircleCI 환경에서 실행 중임을 표시 | 적합 — CircleCI 전용 최상위 presence 마커. 값은 항상 `true`로 문서화됨 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CI` | 불리언 문자열 (`true`) | 실행 식별 | 일반 CI 환경에서 실행 중임을 표시 | 보조 신호 — 대부분의 CI 플랫폼이 공통으로 설정하는 범용 변수이므로 CircleCI 단독 식별에는 `CIRCLECI`와 함께 사용해야 함 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_PIPELINE_ID` | 문자열(고유 식별자) | 실행 식별 | 트리거된 파이프라인의 고유 ID. 계층 최상위(pipeline) 식별자 | 적합 — 파이프라인 단위 실행을 식별하는 안정적인 빌트인 변수 | [Built-in environment variables](https://circleci.com/docs/variables/), [How to find IDs](https://circleci.com/docs/guides/toolkit/how-to-find-ids/) |
| `CIRCLE_WORKFLOW_ID` | 문자열(고유 식별자) | 실행 식별 | 트리거된 워크플로 인스턴스의 고유 ID. 같은 워크플로에 속한 모든 job에서 동일한 값 | 적합 — 워크플로 단위 실행을 식별하며 job 간 그룹핑 키로 사용 가능 | [Built-in environment variables](https://circleci.com/docs/variables/), [How to find IDs](https://circleci.com/docs/guides/toolkit/how-to-find-ids/) |
| `CIRCLE_WORKFLOW_JOB_ID` | 문자열(고유 식별자) | 실행 식별 | 현재 job 실행의 고유 ID. 계층 최하위(job) 식별자 | 적합 — job 단위 최소 실행 단위를 식별하는 가장 구체적인 변수 | [Built-in environment variables](https://circleci.com/docs/variables/), [How to find IDs](https://circleci.com/docs/guides/toolkit/how-to-find-ids/) |
| `CIRCLE_JOB` | 문자열 (`config.yml`에 정의된 job 이름) | 실행 식별 | 현재 실행 중인 job의 이름 | 적합 — 사람이 읽을 수 있는 job 식별자로, ID 변수와 병행해 파이프라인 구성상의 위치를 보강함 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_BUILD_NUM` | 정수 | 실행 식별 | 프로젝트 내에서 job마다 부여되는 고유 빌드 번호 | 보조 신호 — job 고유 값이지만 workflow/pipeline 계층을 표현하지 않는 레거시 스타일 번호이며, 근접 변수인 `CIRCLE_PREVIOUS_BUILD_NUM`은 self-hosted runner executor에서 설정되지 않는다고 공식 문서에 명시되어 있어 `CIRCLE_BUILD_NUM` 계열 전체를 유일한 판정 근거로 삼지 않는 편이 안전함 | [Built-in environment variables](https://circleci.com/docs/variables/) |

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CIRCLE_PROJECT_REPONAME` | 문자열 | 상태·컨텍스트 | 현재 파이프라인이 속한 저장소 이름 | 보조 신호 — 실행 여부보다 대상 프로젝트를 식별하는 컨텍스트 정보 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_PROJECT_USERNAME` | 문자열 | 상태·컨텍스트 | 저장소 소유자(사용자/조직) 이름 | 보조 신호 — 프로젝트 소속 컨텍스트 정보 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_BRANCH` | 문자열 | 상태·컨텍스트 | 현재 git 브랜치 이름 | 보조 신호 — 태그 기반 빌드에서는 존재하지 않을 수 있음 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_TAG` | 문자열 | 상태·컨텍스트 | 태그 기반 빌드의 git 태그 이름 | 보조 신호 — 브랜치 기반 빌드에서는 존재하지 않음 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_SHA1` | 40자 SHA1 문자열 | 상태·컨텍스트 | 빌드 대상 커밋의 전체 SHA1 해시 | 보조 신호 — 커밋 컨텍스트일 뿐 실행 주체 판정에는 직접 쓰이지 않음 | [Built-in environment variables](https://circleci.com/docs/variables/) |
| `CIRCLE_NODE_INDEX` | 정수 (0부터 시작) | 상태·컨텍스트 | `parallelism` 사용 시 현재 컨테이너의 인덱스 | 보조 신호 — `parallelism`을 설정한 job에서만 의미가 있으며, 단독 값으로는 CircleCI 실행 여부를 알려주지 않음 | [Test splitting and parallelism](https://circleci.com/docs/guides/optimize/parallelism-faster-jobs/) |
| `CIRCLE_NODE_TOTAL` | 정수 | 상태·컨텍스트 | `parallelism`으로 분할된 총 컨테이너 수 | 보조 신호 — `CIRCLE_NODE_INDEX`와 함께 분할 위치를 계산하는 용도이며 단독 판정력은 낮음 | [Test splitting and parallelism](https://circleci.com/docs/guides/optimize/parallelism-faster-jobs/) |
| `CIRCLE_WORKFLOW_WORKSPACE_ID` | 문자열(고유 식별자) | 상태·컨텍스트 | 워크플로에 걸쳐 일관되게 유지되는 workspace 식별자 | 보조 신호 — `CIRCLE_WORKFLOW_ID`를 보강하는 워크플로 컨텍스트이며, 아래 결론에서 다루는 rerun 관련 동작은 공식 문서로 확정되지 않음 | [Built-in environment variables](https://circleci.com/docs/variables/) |

### 의도적으로 제외한 변수

다음 공식 빌트인 변수들은 `runby`가 실행 메타데이터로 안전하게 노출하기에 부적절하다고 판단해 표에서 제외했습니다.

- `CIRCLE_PR_NUMBER`, `CIRCLE_PR_REPONAME`, `CIRCLE_PR_USERNAME`, `CIRCLE_PULL_REQUEST`, `CIRCLE_PULL_REQUESTS` — GitHub/Bitbucket 특정 VCS 조합의 forked PR 빌드에서만 존재하고, GitLab 등 일부 VCS에서는 지원이 종료(deprecated)되어 플랫폼 전반의 일관된 실행 신호로 쓸 수 없습니다.
- `CIRCLE_REPOSITORY_URL`, `CIRCLE_USERNAME` — 일부 VCS(GitLab SaaS, GitHub App, Bitbucket Data Center 등)에서 지원이 종료되었고, `CIRCLE_USERNAME`은 트리거한 사용자 식별 정보로 실행 판정과 무관합니다.
- `CIRCLE_PREVIOUS_BUILD_NUM` — 공식 문서가 "비결정적이며 새 워크플로에서 사용을 피해야 한다"고 명시하고, self-hosted runner executor에서는 설정되지 않습니다.
- `CIRCLE_OIDC_TOKEN`, `CIRCLE_OIDC_TOKEN_V2` — 인증 토큰이므로 노출 자체가 보안 위험이며 실행 식별 메타데이터가 아닙니다.
- `CIRCLE_WORKING_DIRECTORY`, `CIRCLE_INTERNAL_TASK_DATA` — 각각 job 설정 값과 "향후 변경될 수 있는 내부 데이터"로 문서화되어 있어 실행 감지에 부적합합니다.
- `CIRCLE_ORGANIZATION_ID`, `CIRCLE_PROJECT_ID` — 조직/프로젝트를 식별하는 유효한 컨텍스트 변수이지만, 이 문서는 실행 계층(pipeline/workflow/job) 식별에 선택과 집중했으므로 별도 표에 포함하지 않았습니다.

## 실행 주체 감지에 관한 결론

**presence 마커**는 `CIRCLECI=true`가 가장 신뢰할 수 있는 단일 신호입니다. `CI=true`는 CircleCI를 포함한 거의 모든 CI 플랫폼이 공통으로 설정하므로 보조 신호로만 취급해야 합니다.

**계층 구조**는 pipeline → workflow → job 3단계이며, 공식 ID 가이드는 각 단계를 다음과 같이 매핑합니다.

- pipeline 수준 ID: `CIRCLE_PIPELINE_ID`
- workflow 수준 ID: `CIRCLE_WORKFLOW_ID`
- job 수준 ID: `CIRCLE_WORKFLOW_JOB_ID`

`runby`는 **`CIRCLE_PIPELINE_ID`를 pipeline 수준 ID로, `CIRCLE_WORKFLOW_JOB_ID`를 job 수준 ID로** 채택해야 합니다. `CIRCLE_WORKFLOW_ID`는 그 사이의 workflow 수준 그룹핑 키로 별도 보관하는 것이 계층 정보를 온전히 보존하는 방법입니다. `CIRCLE_BUILD_NUM`과 `CIRCLE_JOB`은 사람이 읽기 쉬운 보조 식별자로 취급하되, ID 계열을 대체하지는 않아야 합니다.

**attempt/retry 카운터**: 공식 빌트인 환경변수 문서에는 워크플로 rerun이나 job 재시도 횟수를 나타내는 전용 카운터 변수가 없습니다. `CIRCLE_WORKFLOW_WORKSPACE_ID`가 워크플로 rerun 시 `CIRCLE_WORKFLOW_ID`와 다르게 유지되는지 여부는 일부 커뮤니티 자료에서 언급되지만, 조사 시점의 공식 문서에서는 이 rerun 시 동작 차이를 명시적으로 확인할 수 없었습니다. 따라서 `runby`는 rerun/attempt 횟수를 CircleCI 환경변수만으로 판정하려 하지 말아야 합니다.

**트리거 유형(push/API/scheduled) 노출 여부**: 트리거 출처는 `pipeline.trigger_source`(config.yml에서 `<< pipeline.trigger_source >>`로 참조하는 파이프라인 값)로 존재하지만, 이는 파이프라인 설정 처리 시점에만 쓸 수 있는 값이며 job 컨테이너의 환경변수로 자동 주입되지 않습니다. 프로젝트가 `environment:` 키에 명시적으로 대입하지 않는 한 `runby`가 관찰하는 job 런타임 환경에는 트리거 유형 정보가 존재하지 않습니다.

**self-hosted runner / executor 차이**: 공식 문서에서 명확히 확인된 차이는 `CIRCLE_PREVIOUS_BUILD_NUM`이 self-hosted runner executor에서 설정되지 않는다는 점뿐입니다. Docker, machine, macOS 실행기 간의 폭넓은 변수 차이는 공식 문서에서 별도로 다루지 않으므로, 이 문서의 실행 식별 표에 있는 핵심 변수(`CIRCLECI`, `CIRCLE_PIPELINE_ID`, `CIRCLE_WORKFLOW_ID`, `CIRCLE_WORKFLOW_JOB_ID`)는 실행기 종류와 무관하게 job 수준에서 항상 존재한다고 간주할 수 있습니다. `CIRCLE_NODE_INDEX`/`CIRCLE_NODE_TOTAL`은 `parallelism`을 설정한 job에서만 의미 있는 값을 가지므로 실행기 차이가 아니라 job 설정 차이로 존재 여부가 갈립니다.

**상속·위조 주의**: 다른 CI 문서와 마찬가지로 이 변수들은 job 컨테이너 안의 자식 프로세스에 상속되며, 사용자가 로컬 셸에서 같은 이름으로 값을 임의로 설정할 수도 있습니다. `적합` 판정은 CircleCI가 공식적으로 이 변수들을 job 시작 시 주입한다는 계약을 근거로 한 것이지, 절대적인 위조 방지 신뢰 경계를 의미하지 않습니다.

## 공식 문서

- [Project values and variables (Built-in environment variables)](https://circleci.com/docs/variables/)
- [How to find IDs](https://circleci.com/docs/guides/toolkit/how-to-find-ids/)
- [Test splitting and parallelism](https://circleci.com/docs/guides/optimize/parallelism-faster-jobs/)
- [Pipelines and triggers overview](https://circleci.com/docs/guides/orchestrate/pipelines.html)
- [Trigger options](https://circleci.com/docs/guides/orchestrate/triggers-overview/)
- [Pipeline values and parameters](https://circleci.com/docs/guides/orchestrate/pipeline-variables/)
- [Introduction to environment variables](https://circleci.com/docs/guides/security/env-vars/)
