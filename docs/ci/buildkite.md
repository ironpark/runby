---
title: Buildkite
slug: buildkite
research_date: 2026-08-31
open_source: true
repository: https://github.com/buildkite/agent
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 환경변수 레퍼런스가 모든 변수를 "read-only"로 명시하고 `buildkite-agent` 소스에서 PID·UUID 주입 지점을 직접 확인할 수 있어 별도 실행 검증 없이도 값의 형태와 주입 주체를 판정할 수 있음
---

# Buildkite

Buildkite는 파이프라인 실행을 관리하는 SaaS 컨트롤 플레인과, 사용자의 인프라(자체 서버, 컨테이너, Hosted Agents 등)에서 실제 잡을 실행하는 오픈소스 `buildkite-agent`로 구성됩니다. 잡 환경에 주입되는 `BUILDKITE_*` 변수는 공식 문서에서 대부분 "read-only"로 표기되어 있으며, 이는 에이전트가 부트스트랩 시점에 값을 설정한다는 의미이지 셸 사용자가 물리적으로 값을 덮어쓸 수 없다는 뜻은 아닙니다. 즉 이 변수들은 **에이전트가 신뢰할 수 있는 소스에서 주입한다는 계약**은 있지만, 상속된 자식 프로세스 관점에서는 다른 CI 변수와 마찬가지로 위조 가능한 평문 환경변수입니다.

`open_source: true`로 표기했지만 정확히는 **실행 에이전트(`buildkite-agent`)만 오픈소스**이며, 파이프라인 정의·빌드 스케줄링·웹 UI를 담당하는 Buildkite 컨트롤 플레인 자체는 비공개 SaaS입니다. 따라서 이 문서의 "공식 소스" 인용은 어디까지나 에이전트가 잡 프로세스에 환경변수를 주입하는 부트스트랩 로직에 한정됩니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `BUILDKITE` | 문자열 (`true`) | 실행 식별 | Buildkite 잡 환경에서 항상 설정되는 존재 마커 | 적합 — Buildkite 전용 변수명이며 모든 잡에 항상 주입됨. 가장 신뢰할 수 있는 존재 신호 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `CI` | 문자열 (`true`) | 실행 식별 | 범용 CI 여부 표시 | 보조 신호 — 거의 모든 CI 플랫폼이 동일한 이름과 값을 설정하므로 단독으로는 Buildkite를 특정하지 못함 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_BUILD_ID` | UUID 문자열 | 실행 식별 | 빌드를 식별하는 안정적 UUID | 적합 — 빌드 단위의 고유하고 안정적인 식별자 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_JOB_ID` | UUID 문자열 | 실행 식별 | 개별 잡(단계 실행 인스턴스)을 식별하는 내부 UUID | 적합 — 잡 단위의 고유 식별자이며 같은 빌드 내 병렬·재시도 잡을 구분할 수 있음 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_AGENT_ID` | UUID 문자열 | 실행 식별 | 잡을 실행한 에이전트의 UUID | 보조 신호 — 에이전트 인스턴스 식별에는 유용하지만 그 자체로 잡 존재를 보장하는 값은 아니며 `BUILDKITE_BUILD_ID`/`BUILDKITE_JOB_ID`와 함께 봐야 의미가 큼 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_AGENT_NAME` | 문자열 (예: `elastic-builders-088264dc4f9`) | 실행 식별 | 잡을 실행한 에이전트의 표시 이름 | 보조 신호 — 사람이 읽을 수 있는 식별자이지만 조직에서 임의로 이름을 정할 수 있어 UUID보다 신뢰도가 낮음 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_AGENT_PID` | 정수 문자열 (예: `6`) | 실행 식별 | 잡을 실행 중인 `buildkite-agent` 프로세스 자신의 PID | 보조 신호 — 아래 "BUILDKITE_AGENT_PID에 대한 참고"에서 설명 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables), [공식 소스: `os.Getpid()` 주입](https://github.com/buildkite/agent/blob/main/agent/register.go) |

### BUILDKITE_AGENT_PID에 대한 참고

공식 문서는 이 값을 "The process ID of the agent"라고만 설명합니다. `buildkite/agent` 소스에서 에이전트가 API에 자신을 등록할 때 `os.Getpid()`로 얻은 PID를 사용하는 지점을 확인했으며, 이는 **잡 셸 프로세스의 PID가 아니라 그 잡을 관리하는 장수(long-lived) `buildkite-agent` 데몬 프로세스의 PID**입니다. 즉 이 값은 "어떤 에이전트 프로세스가 이 잡을 소유했는가"라는 컨텍스트는 제공하지만, 값 자체는 호스트마다 임의의 정수이고 재시작마다 바뀌므로 단독으로 Buildkite 실행을 판정하는 근거로는 약합니다. `BUILDKITE=true` 및 `BUILDKITE_JOB_ID`와 함께 있을 때만 보조적 컨텍스트로 사용하는 것이 안전합니다.

Buildkite 에이전트는 원래 사용자가 직접 프로비저닝하는 self-hosted 모델(자체 서버, 컨테이너, VM)로 설계되었습니다. 공식 문서에는 Buildkite가 컴퓨트를 대신 관리하는 Hosted Agents 옵션도 있으며, 이 경우 잡 환경에 `BUILDKITE_COMPUTE_TYPE`(`hosted` 또는 `self-hosted`) 변수가 추가로 설정됩니다. 어느 경우든 `BUILDKITE_AGENT_PID`가 가리키는 것은 항상 "그 컴퓨트 위에서 도는 에이전트 프로세스"이지 잡의 실행 파일 자체가 아니므로, `runby`가 확인해야 할 실제 자식 프로세스와는 별개의 값이라는 점을 유의해야 합니다.

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `BUILDKITE_BUILD_NUMBER` | 정수 문자열 (예: `1514`) | 상태·컨텍스트 | 파이프라인 내에서 매 빌드마다 증가하는 순번 | 보조 신호 — 파이프라인 범위 안에서만 고유하며 전역적으로 유일하지 않아 `BUILDKITE_BUILD_ID`의 보조 정보로만 사용 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_PIPELINE_SLUG` | 슬러그 문자열 | 상태·컨텍스트 | URL에 쓰이는 파이프라인 슬러그 | 보조 신호 — 조직 구성에 따라 흔한 이름일 수 있어 단독 감지에는 부적합하지만 어떤 파이프라인이 실행 중인지 보고하는 데 유용 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_ORGANIZATION_SLUG` | 슬러그 문자열 | 상태·컨텍스트 | URL에 쓰이는 조직 슬러그 | 보조 신호 — 어느 Buildkite 조직에서 실행되는지 식별하는 컨텍스트 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_BRANCH` | 브랜치명 문자열 | 상태·컨텍스트 | 빌드 대상 브랜치 | 보조 신호 — 매우 흔한 이름 형태라 감지보다는 컨텍스트 보고용 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_COMMIT` | 40자리 Git SHA-1 또는 `HEAD` 같은 심볼릭 이름 | 상태·컨텍스트 | 빌드 대상 커밋 | 보조 신호 — 커밋 컨텍스트 보고용, 감지 신호는 아님 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_SOURCE` | 열거형 문자열: `webhook`, `api`, `ui`, `trigger_job`, `schedule` | 상태·컨텍스트 | 빌드를 발생시킨 트리거 종류 | 보조 신호 — Buildkite 내부 트리거 분류이며 감지가 아닌 실행 경위 보고용 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_RETRY_COUNT` | 정수 문자열, 기본 `0` | 상태·컨텍스트 | 현재 잡이 몇 번 재시도되었는지 | 보조 신호 — 재시도 횟수 컨텍스트이며 그 자체로 존재 여부를 증명하지 않음 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |
| `BUILDKITE_COMPUTE_TYPE` | 열거형 문자열: `hosted`, `self-hosted` | 상태·컨텍스트 | 잡이 Buildkite Hosted Agents 위에서 도는지, 사용자 인프라(self-hosted)에서 도는지 | 보조 신호 — 컴퓨트 모델 구분이며 Buildkite 실행 자체의 판정 근거는 아님 | [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables) |

### 의도적으로 제외한 변수

공식 레퍼런스에는 위 표보다 훨씬 많은 `BUILDKITE_*` 변수(120개 이상)가 있습니다. 다음 카테고리는 `runby`가 노출할 "실행 메타데이터"로서 의미가 낮거나 민감해 의도적으로 제외했습니다.

- **에이전트 설정 미러 변수** (`BUILDKITE_AGENT_ENDPOINT`, `BUILDKITE_HOOKS_PATH`, `BUILDKITE_BUILD_PATH`, `BUILDKITE_SHELL`, `BUILDKITE_GIT_*_FLAGS` 등) — 대부분 에이전트 설정 옵션 값을 그대로 반영한 것으로, 실행 감지나 실행 컨텍스트가 아니라 로컬 에이전트 구성입니다.
- **자격 증명·시크릿 관련 변수** (`BUILDKITE_AGENT_ACCESS_TOKEN`, `BUILDKITE_S3_ACCESS_KEY_ID`, `BUILDKITE_S3_SECRET_ACCESS_KEY` 등) — 노출 시 보안 위험이 있어 애초에 라이브러리가 다루지 않아야 합니다.
- **VCS 제공자별 부가 변수** (`BUILDKITE_GITHUB_DEPLOYMENT_ID`, `BUILDKITE_PULL_REQUEST_*`, `BUILDKITE_TAG`, `BUILDKITE_REPO` 등) — GitHub/GitLab 등 특정 소스 제공자·이벤트 유형에서만 존재하며, 플랫폼 감지보다는 파이프라인 트리거 세부사항에 해당합니다.
- **재빌드·트리거 계보 변수** (`BUILDKITE_REBUILT_FROM_BUILD_ID`, `BUILDKITE_TRIGGERED_FROM_BUILD_ID` 등) — 유용한 계보 정보지만 존재 여부가 조건부(재빌드/트리거된 빌드에서만 값이 채워짐)라 일반적인 실행 감지 신호로 부적합합니다.
- **그룹·병렬 단계 변수** (`BUILDKITE_GROUP_ID`, `BUILDKITE_PARALLEL_JOB`, `BUILDKITE_PARALLEL_JOB_COUNT`) — 특정 파이프라인 구조(그룹/병렬 스텝)를 쓸 때만 존재하는 조건부 값입니다.

## 실행 주체 감지에 관한 결론

`BUILDKITE`가 가장 신뢰할 수 있는 단일 존재 마커입니다. Buildkite 전용 이름을 쓰고, 잡이 실행되는 한 항상 `true`로 설정된다고 공식 문서가 명시하기 때문입니다. `CI`는 값과 이름이 사실상 모든 CI 플랫폼에서 동일하므로 `BUILDKITE`가 없을 때 일반 CI 여부를 보강하는 정도로만 사용해야 합니다.

권장 판정 순서:

1. `BUILDKITE == "true"`를 확인해 Buildkite 실행임을 판정합니다.
2. `BUILDKITE_BUILD_ID`(빌드 단위 UUID)와 `BUILDKITE_JOB_ID`(잡 단위 UUID)를 안정적 식별자로 함께 노출합니다. 둘은 계층이 다르므로 혼동하지 않아야 합니다 — 하나의 빌드 안에 여러 잡이 존재할 수 있습니다.
3. `BUILDKITE_RETRY_COUNT`를 재시도 횟수로 노출하되, 문서화된 시도 횟수 변수는 이것 하나뿐이며 자동/수동 재시도를 구분하는 별도의 공식 변수는 확인되지 않았습니다.
4. `BUILDKITE_SOURCE`로 트리거 종류(`webhook`/`api`/`ui`/`trigger_job`/`schedule`)를 보고할 수 있으며, 값 집합은 공식 문서에 열거된 다섯 가지로 한정됩니다.
5. `BUILDKITE_AGENT_PID`는 잡을 소유한 에이전트 데몬의 PID일 뿐 잡 프로세스 자신의 PID가 아니므로, 존재 판정이 아닌 부가 컨텍스트로만 다룹니다.

환경변수는 자식 프로세스에 상속되거나 사용자가 셸에서 동일한 이름으로 재설정할 수 있으므로, 이 신호들은 "Buildkite가 이 프로세스를 만들었다"는 강한 보장이 아니라 "이 프로세스 트리 어딘가에서 Buildkite 잡 환경이 상속되었다"는 근거로 해석해야 합니다. 특히 `BUILDKITE_BUILD_ID`/`BUILDKITE_JOB_ID`처럼 UUID 형태의 값은 사용자가 우연히 같은 값을 설정할 가능성이 낮아 위조 위험이 상대적으로 작지만, `BUILDKITE=true`나 `CI=true`처럼 단순한 문자열 값은 다른 도구나 로컬 셸 설정에서도 얼마든지 재현될 수 있습니다.

## 공식 문서

- [Environment variables](https://buildkite.com/docs/pipelines/configure/environment-variables)
- [Retry](https://buildkite.com/docs/pipelines/configure/retry)
- [Buildkite agent configuration (v3)](https://buildkite.com/docs/agent/v3/configuration)
- [공식 `buildkite-agent` 소스 저장소 (`buildkite/agent`)](https://github.com/buildkite/agent)
- [공식 소스: 에이전트 등록 시 PID 주입 (`agent/register.go`)](https://github.com/buildkite/agent/blob/main/agent/register.go)
