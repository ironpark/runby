---
title: GitLab CI/CD
slug: gitlab-ci
research_date: 2026-08-31
open_source: true
repository: https://gitlab.com/gitlab-org/gitlab
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 Predefined CI/CD variables 문서가 모든 값과 트리거 소스, 도입 버전을 명시하므로 별도 실행 검증 없이 문서만으로 판정 가능함. 다만 self-managed 인스턴스 및 child/downstream 파이프라인에서의 값 조합은 실제 표시를 다시 확인할 여지가 있음
---

# GitLab CI/CD

GitLab은 CI/CD 파이프라인이 실행하는 모든 잡에 `GITLAB_CI`와 `CI`를 포함한 사전 정의 변수(predefined variables) 집합을 공식적으로 문서화하고 보장합니다. 이 변수들은 GitLab.com(SaaS)과 self-managed 인스턴스 모두에서 GitLab Runner가 잡을 시작할 때 자동으로 주입되며, `.gitlab-ci.yml` 설정 없이도 항상 존재합니다. 단, 일부 변수는 "pre-pipeline"(파이프라인 생성 전 컨텍스트에서도 사용 가능) 범위이고 다른 일부는 "job-only"(실제 잡 실행 중에만 존재) 범위로 나뉘므로, `runby`가 이 변수들을 읽는 시점은 항상 실행 중인 잡의 프로세스 환경이라는 점에 유의해야 합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GITLAB_CI` | 불리언 문자열 (`true`) | 실행 식별 | GitLab CI/CD에서 실행되는 모든 잡에 존재 | 적합 — GitLab이 공식 보장하는 존재 마커이며 오탈자·충돌 위험이 낮은 GitLab 전용 변수명 | [Predefined variables — `GITLAB_CI`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI` | 불리언 문자열 (`true`) | 실행 식별 | GitLab CI/CD에서 실행되는 모든 잡에 존재 | 보조 신호 — GitLab이 공식 보장하지만 `CI`는 GitHub Actions, CircleCI 등 대부분의 CI 플랫폼이 공유하는 범용 관례 변수명이라 단독으로는 GitLab 특정이 불가능 | [Predefined variables — `CI`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_PIPELINE_ID` | 정수 문자열 | 실행 식별 | 현재 파이프라인의 ID. GitLab 인스턴스 전체에서 고유 | 적합 — 파이프라인 단위의 안정적 고유 식별자 | [Predefined variables — `CI_PIPELINE_ID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_PIPELINE_IID` | 정수 문자열 | 실행 식별 | 현재 파이프라인의 내부 ID(IID). 현재 프로젝트 내에서만 고유 | 보조 신호 — 프로젝트 범위에서만 고유하므로 인스턴스 전역 식별에는 `CI_PIPELINE_ID`가 더 적합 | [Predefined variables — `CI_PIPELINE_IID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_JOB_ID` | 정수 문자열 | 실행 식별 | 현재 잡의 내부 ID. GitLab 인스턴스 전체 잡 중 고유 | 적합 — 잡 단위의 안정적 고유 식별자 | [Predefined variables — `CI_JOB_ID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_JOB_RETRY_COUNT` | 정수 문자열 (`0`부터 시작) | 실행 식별 | 현재 파이프라인에서 이 잡이 재시도된 횟수. 최초 실행은 `0`, 재시도마다 1씩 증가 | 적합 — GitLab 19.3에서 도입된 공식 재시도 카운터로, 동일 잡의 시도 회차를 구분하는 유일한 문서화된 신호 | [Predefined variables — `CI_JOB_RETRY_COUNT`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_JOB_TOKEN` | 비밀 토큰 문자열 | 실행 식별 | 잡이 실행되는 동안만 유효한 API 인증 토큰 | 보조 신호 — 잡 실행과 결부된 존재 자체는 강한 신호이지만 민감한 자격 증명이므로 감지 목적으로 값을 로깅·비교해서는 안 됨 | [Predefined variables — `CI_JOB_TOKEN`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_RUNNER_ID` | 정수 문자열 | 실행 식별 | 잡을 처리한 Runner의 고유 ID | 적합 — 어떤 잡이든 Runner가 배정되면 항상 존재하는 실행 인프라 식별자 | [Predefined variables — `CI_RUNNER_ID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_PIPELINE_SOURCE` | 열거형 문자열(`push`, `web`, `schedule`, `api`, `external`, `trigger`, `webide`, `merge_request_event`, `external_pull_request_event`, `parent_pipeline`, `pipeline`, `chat`, `ondemand_dast_scan`, `ondemand_dast_validation`) | 실행 식별 | 파이프라인이 어떤 방식으로 트리거되었는지 표시 | 적합 — GitLab이 공식 열거하는 트리거 유형 신호로, push/schedule/API/트리거 파이프라인을 구분하는 핵심 값 | [Predefined variables — `CI_PIPELINE_SOURCE`](https://docs.gitlab.com/ci/variables/predefined_variables/) |

## 상태·컨텍스트 변수

이 절은 GitLab이 문서화한 방대한 사전 정의 변수 목록 중 `runby`가 실행 메타데이터로 안전하게 노출할 수 있는 항목만 선별했습니다. 프로젝트, 커밋, 트리거 주체, Runner, 서버 식별 정보가 대상입니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CI_SERVER` | 문자열 (`yes`) | 상태·컨텍스트 | `GITLAB_CI`와 동일하게 모든 잡에서 존재를 표시 | 보조 신호 — `GITLAB_CI`와 중복되는 존재 마커이며 값이 `true`가 아닌 `yes`라는 점에서 별도 확인이 필요 | [Predefined variables — `CI_SERVER`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_SERVER_URL` | URL 문자열 | 상태·컨텍스트 | GitLab 인스턴스의 기본 URL(프로토콜·포트 포함) | 보조 신호 — GitLab.com과 self-managed 인스턴스를 구분할 수 있는 유일한 항목이지만 값 자체가 커스터마이즈 가능하므로 단독 감지에는 부적합 | [Predefined variables — `CI_SERVER_URL`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_SERVER_VERSION` | 버전 문자열 | 상태·컨텍스트 | GitLab 인스턴스의 전체 버전 | 보조 신호 — self-managed 인스턴스의 GitLab 버전을 확인하는 보강 정보 | [Predefined variables — `CI_SERVER_VERSION`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_PROJECT_ID` | 정수 문자열 | 상태·컨텍스트 | 현재 프로젝트의 ID. 인스턴스 전체에서 고유 | 보조 신호 — 파이프라인 소속 프로젝트를 식별하는 컨텍스트 정보 | [Predefined variables — `CI_PROJECT_ID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_PROJECT_PATH` | `namespace/project` 문자열 | 상태·컨텍스트 | 네임스페이스를 포함한 프로젝트 전체 경로 | 보조 신호 — 사람이 읽을 수 있는 프로젝트 식별 정보 | [Predefined variables — `CI_PROJECT_PATH`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_COMMIT_SHA` | 전체 Git SHA 문자열 | 상태·컨텍스트 | 빌드 대상 커밋의 리비전 | 보조 신호 — 실행 대상 커밋을 특정하는 컨텍스트 정보 | [Predefined variables — `CI_COMMIT_SHA`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_COMMIT_REF_NAME` | 브랜치·태그명 문자열 | 상태·컨텍스트 | 빌드 대상 브랜치 또는 태그명 | 보조 신호 — 실행 대상 ref를 특정하는 컨텍스트 정보 | [Predefined variables — `CI_COMMIT_REF_NAME`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_JOB_NAME` | 문자열 | 상태·컨텍스트 | 현재 잡의 이름 | 보조 신호 — 잡을 사람이 식별하기 위한 라벨 | [Predefined variables — `CI_JOB_NAME`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_JOB_STAGE` | 문자열 | 상태·컨텍스트 | 현재 잡이 속한 스테이지명 | 보조 신호 — 파이프라인 내 실행 단계 컨텍스트 | [Predefined variables — `CI_JOB_STAGE`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_RUNNER_TAGS` | JSON 배열 문자열 | 상태·컨텍스트 | 잡을 처리한 Runner에 설정된 태그 목록 | 보조 신호 — Runner 구성을 나타낼 뿐 실행 주체 자체의 식별자는 아님 | [Predefined variables — `CI_RUNNER_TAGS`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_RUNNER_DESCRIPTION` | 문자열 | 상태·컨텍스트 | Runner의 설명 텍스트 | 보조 신호 — 사람이 읽는 라벨이며 관리자가 임의로 변경 가능 | [Predefined variables — `CI_RUNNER_DESCRIPTION`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_RUNNER_SHORT_TOKEN` | 문자열 (토큰 앞 17자) | 상태·컨텍스트 | 새 잡 요청 인증에 쓰이는 Runner 고유 토큰의 축약형 | 부적합 — 민감한 인증 관련 값이며 감지 목적으로 노출·비교하지 않는 편이 안전 | [Predefined variables — `CI_RUNNER_SHORT_TOKEN`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `GITLAB_USER_ID` | 정수 문자열 | 상태·컨텍스트 | 파이프라인을 시작한 사용자의 숫자 ID(수동 잡은 실행한 사용자) | 보조 신호 — 트리거 주체(사람)를 식별하는 컨텍스트 정보 | [Predefined variables — `GITLAB_USER_ID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `GITLAB_USER_LOGIN` | 사용자명 문자열 | 상태·컨텍스트 | 파이프라인을 시작한 사용자의 고유 username | 보조 신호 — 트리거 주체를 사람이 읽을 수 있는 형태로 제공 | [Predefined variables — `GITLAB_USER_LOGIN`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_PIPELINE_TRIGGERED` | 불리언 문자열 (`true`) | 상태·컨텍스트 | 트리거 토큰으로 시작된 파이프라인에서만 존재 | 보조 신호 — `CI_PIPELINE_SOURCE == "trigger"`와 상당 부분 중복되는 보강 신호 | [Predefined variables — `CI_PIPELINE_TRIGGERED`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_JOB_MANUAL` | 불리언 문자열 (`true`) | 상태·컨텍스트 | 잡이 수동으로 시작된 경우에만 존재 | 보조 신호 — 잡 시작 방식에 대한 보강 정보이며 파이프라인 트리거 출처와는 별개 | [Predefined variables — `CI_JOB_MANUAL`](https://docs.gitlab.com/ci/variables/predefined_variables/) |
| `CI_CONCURRENT_ID` | 정수 문자열 | 상태·컨텍스트 | 단일 executor 내 빌드 실행의 고유 ID | 보조 신호 — Runner의 동시 실행 슬롯을 식별할 뿐 파이프라인·잡 식별자는 아님 | [Predefined variables — `CI_CONCURRENT_ID`](https://docs.gitlab.com/ci/variables/predefined_variables/) |

의도적으로 제외한 변수도 있습니다. `CI_COMMIT_MESSAGE`, `CI_COMMIT_AUTHOR` 등 커밋 메타데이터는 실행 주체 감지와 무관한 페이로드성 정보라서 제외했습니다. `CI_REGISTRY*`, `CI_DEPENDENCY_PROXY*` 등 Container Registry·Dependency Proxy 관련 변수도 인프라 설정값일 뿐 실행 식별 신호가 아니므로 제외했습니다. `CI_API_V4_URL`, `CI_TEMPLATE_REGISTRY_HOST` 같은 API/템플릿 엔드포인트 변수 역시 설정용이라 제외했습니다. `CI_MERGE_REQUEST_*` 계열은 `CI_PIPELINE_SOURCE == "merge_request_event"`일 때만 존재하는 파생 컨텍스트이므로, 핵심 실행 감지 표에는 넣지 않고 `CI_PIPELINE_SOURCE` 값 해석으로 충분하다고 판단했습니다.

### `CI_PIPELINE_SOURCE`의 전체 값 집합

GitLab 공식 문서가 열거하는 값은 다음과 같습니다.

| 값 | 의미 |
|---|---|
| `push` | git push 이벤트로 트리거 |
| `web` | GitLab 웹 UI에서 수동으로 "Run pipeline" 실행 |
| `schedule` | 예약된(scheduled) 파이프라인 |
| `api` | GitLab API 호출로 트리거 |
| `external` | 외부 CI 시스템에서 트리거 |
| `trigger` | 트리거 토큰을 사용한 트리거 |
| `webide` | Web IDE에서 트리거 |
| `merge_request_event` | 머지 리퀘스트 파이프라인 |
| `external_pull_request_event` | 외부(GitHub 등) 풀 리퀘스트에서 트리거 |
| `parent_pipeline` | 부모 파이프라인이 트리거한 child 파이프라인(같은 프로젝트) |
| `pipeline` | 멀티 프로젝트(다운스트림) 파이프라인 |
| `chat` | ChatOps 명령으로 트리거 |
| `ondemand_dast_scan` | on-demand DAST 스캔 파이프라인 |
| `ondemand_dast_validation` | DAST 사이트 프로파일 검증 파이프라인 |

## 실행 주체 감지에 관한 결론

**가장 신뢰도 높은 존재 마커는 `GITLAB_CI`입니다.** `CI`도 공식 보장되지만 다른 CI 플랫폼과 공유하는 범용 관례 변수명이라 GitLab 특정 신호로는 `GITLAB_CI`가 우선합니다.

식별자 계층은 다음과 같이 정리됩니다.

- **파이프라인 단위 안정 ID**: `CI_PIPELINE_ID`(인스턴스 전역 고유). `CI_PIPELINE_IID`는 프로젝트 범위 고유값이라 보조로만 사용합니다.
- **잡 단위 ID**: `CI_JOB_ID`(인스턴스 전역 고유).
- **재시도(attempt) 카운터**: `CI_JOB_RETRY_COUNT`(GitLab 19.3 도입, `0`부터 시작). 그 이전 GitLab 버전에는 이에 대응하는 공식 변수가 없었으므로, 낮은 버전의 self-managed 인스턴스에서는 이 변수가 존재하지 않을 수 있다는 점을 감안해야 합니다.
- **트리거 유형**: `CI_PIPELINE_SOURCE`. push/schedule/api/trigger/merge_request_event 등 실제 트리거 원인을 구분하는 유일한 공식 열거형 신호입니다.

권장 판정 순서는 다음과 같습니다.

1. `GITLAB_CI`(또는 보조로 `CI`)로 GitLab CI/CD 실행 여부를 판정합니다.
2. `CI_PIPELINE_ID`와 `CI_JOB_ID`를 각각 파이프라인·잡 식별자로 채택합니다.
3. `CI_JOB_RETRY_COUNT`가 존재하면 재시도 회차로 사용하고, 없으면(구버전 인스턴스) 재시도 정보를 `unknown`으로 남깁니다.
4. `CI_PIPELINE_SOURCE` 값으로 트리거 유형(수동/스케줄/API/트리거 토큰/MR/child·downstream)을 분류합니다.
5. `CI_SERVER_URL`, `CI_SERVER_VERSION`으로 GitLab.com인지 self-managed 인스턴스인지 보강 판단합니다. 단, `CI_SERVER_URL`은 인스턴스 운영자가 임의의 도메인으로 설정하므로 GitLab.com(`https://gitlab.com`)과의 문자열 일치만으로 SaaS 여부를 확정하고, 그 외 값은 self-managed로 간주하되 절대적 신뢰 신호로 취급하지 않습니다.

**child/downstream 파이프라인 관련 주의사항**: `CI_PIPELINE_SOURCE`는 같은 프로젝트 내 child 파이프라인에서는 `parent_pipeline`, 멀티 프로젝트(다운스트림) 파이프라인에서는 `pipeline` 값을 가집니다. 두 경우 모두 `CI_PIPELINE_ID`와 `CI_JOB_ID`는 원본 파이프라인/잡과는 별개의 새 값을 받습니다(GitLab 공식 문서는 부모 파이프라인 ID를 자식에 자동 상속한다고 명시하지 않으며, 필요 시 `trigger` 잡의 커스텀 변수로 수동 전달해야 한다고 안내합니다). 따라서 `runby`가 child/downstream 파이프라인에서 부모 파이프라인과의 연결 관계를 확인하려면 GitLab이 자동으로 상속하는 표준 변수가 아니라, 사용자가 명시적으로 전달한 커스텀 변수에 의존해야 합니다.

**self-managed·self-hosted/그룹 Runner 관련 주의사항**: 공식 다운스트림 파이프라인 문서는 GitLab.com, GitLab Self-Managed, GitLab Dedicated 간 이 변수들의 동작 차이를 명시하지 않습니다. `CI_RUNNER_ID`, `CI_RUNNER_TAGS`, `CI_RUNNER_DESCRIPTION`은 GitLab이 관리하는 shared Runner든 사용자가 등록한 self-hosted/그룹 Runner든 동일한 스키마로 제공되지만, 값 자체(태그 구성, 설명 텍스트)는 Runner 관리자가 임의로 설정하므로 인스턴스 간 비교 가능한 표준값이 아닙니다.

마지막으로, 다른 모든 CI 플랫폼과 마찬가지로 이 변수들은 잡 프로세스의 환경변수이므로 자식 프로세스에 상속되거나, 로컬에서 같은 이름의 변수를 수동으로 설정해 위조할 수 있습니다. `runby`는 이를 절대적 신뢰 경계가 아닌 GitLab CI/CD 컨텍스트에 대한 강한 정황 증거로 취급해야 합니다.

## 공식 문서

- [Predefined CI/CD variables reference](https://docs.gitlab.com/ci/variables/predefined_variables/)
- [CI/CD variables](https://docs.gitlab.com/ci/variables/)
- [Downstream pipelines](https://docs.gitlab.com/ci/pipelines/downstream_pipelines/)
- [GitLab 19.3 release notes](https://docs.gitlab.com/releases/19/gitlab-19-3-released/)
- [GitLab 공식 소스 저장소 (`gitlab-org/gitlab`)](https://gitlab.com/gitlab-org/gitlab)

## Merge request 감지

`CI_MERGE_REQUEST_ID`는 merge request 파이프라인에서만 제공되는 사전 정의 변수이므로 존재하면 `PullRequest=true`, 값은 `PullRequestID`로 옮깁니다. `CI_PIPELINE_SOURCE=merge_request_event`도 트리거 종류를 설명하지만, 실행이 merge request 파이프라인이라는 직접 식별자는 ID 변수에 있으므로 ID가 없는 경우에는 PR로 추정하지 않습니다. `CI_MERGE_REQUEST_ID`는 `Evidence`에 이름만 남깁니다.
