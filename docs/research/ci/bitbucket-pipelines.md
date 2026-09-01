---
title: Bitbucket Pipelines
slug: bitbucket-pipelines
research_date: 2026-08-31
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 Atlassian Support 문서의 "Variables and secrets" 페이지가 각 변수의 제공 범위(전체 파이프라인 / 배포 스텝 전용 / PR 전용)를 명시하므로 문서 조사만으로 감지 신호를 확정할 수 있음. 단, 셀프 호스티드 러너 환경에서의 실제 주입 여부는 문서상 예외가 없다고 확인했을 뿐 직접 검증하지 않음
---

# Bitbucket Pipelines

Bitbucket Pipelines는 파이프라인이 실행될 때마다 `bitbucket-pipelines.yml`의 각 스텝에 일련의 기본 변수(default variables)를 자동으로 주입한다고 공식 문서에서 보장합니다. 이 변수들은 사용자가 정의한 저장소·워크스페이스 변수보다 낮은 우선순위를 가지며(우선순위: Pipeline > Deployment > Repository > Workspace > Default variables), 사용자가 같은 이름으로 재정의하면 값이 덮어써질 수 있습니다. 따라서 이 변수들은 "Bitbucket이 이 프로세스를 실행했다"는 강한 계약이지만, 정확한 값까지 위조 불가능하다고 보장하지는 않습니다.

## 실행 식별 신호

Bitbucket Pipelines에는 다른 플랫폼의 `XXX_CI=true`류 전용 불리언 마커가 없습니다. 공식 문서가 보장하는 `CI` 변수는 기본값이 `true`이고 파이프라인 실행 시마다 설정된다고 명시되어 있어 사실상 "이 프로세스가 어떤 CI에서 실행 중"이라는 존재 마커 역할을 하지만, `CI`라는 이름 자체는 사실상 모든 CI 플랫폼이 공유하는 업계 관행 변수이므로 Bitbucket 고유성이 없습니다. 반면 `BITBUCKET_BUILD_NUMBER`는 값 자체가 불리언이 아닌 증가하는 정수 식별자이지만, 변수명이 `BITBUCKET_` 접두사를 가진 유일한 플랫폼 고유 신호이기 때문에 존재 여부 자체가 가장 신뢰할 수 있는 Bitbucket Pipelines 마커입니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CI` | 문자열 `true` | 실행 식별 | 파이프라인이 실행 중임을 표시하는 기본값 `true` 변수 | 보조 신호 — 거의 모든 CI 플랫폼이 같은 이름을 관행적으로 사용하므로 단독으로는 Bitbucket 특정이 불가능함 | [Variables and secrets — CI](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_BUILD_NUMBER` | 증가하는 정수 문자열 | 실행 식별 | 빌드마다 증가하는 고유 식별자, 아티팩트 이름 생성 등에 사용 | 적합 — `BITBUCKET_` 접두사를 가진 Bitbucket 고유 변수이며 모든 파이프라인에서 존재가 보장됨. `CI`와 함께 존재를 확인하면 오탐 가능성이 낮아짐 | [Variables and secrets — BITBUCKET_BUILD_NUMBER](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_PIPELINE_UUID` | UUID 문자열, 중괄호 `{}`로 감싸짐 (예: `{11d87b82-13c6-47a3-8e28-73a2bc378675}`) | 실행 식별 | 현재 파이프라인 실행의 UUID | 적합 — 파이프라인 단위의 안정적 식별자. API 조회에 쓰려면 중괄호 포함 값을 그대로 사용하거나 URL 인코딩해야 함 | [Variables and secrets — BITBUCKET_PIPELINE_UUID](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/), [중괄호 형식 확인](https://support.atlassian.com/bitbucket-cloud/kb/bitbucket-cloud-pipelines-api-error-the-value-provided-is-not-a-valid-uuid/) |
| `BITBUCKET_STEP_UUID` | UUID 문자열, 중괄호 `{}`로 감싸짐 | 실행 식별 | 현재 스텝의 UUID | 적합 — 스텝 단위의 안정적 식별자로 파이프라인 내 특정 스텝을 구분 | [Variables and secrets — BITBUCKET_STEP_UUID](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/), [중괄호 형식 확인](https://support.atlassian.com/bitbucket-cloud/kb/bitbucket-cloud-pipelines-api-error-the-value-provided-is-not-a-valid-uuid/) |
| `BITBUCKET_STEP_RUN_NUMBER` | 정수 (1부터 시작) | 실행 식별 | 같은 스텝이 재실행(retry)된 횟수를 나타내며 최초 실행은 1 | 적합 — 스텝 재시도 판별에 쓸 수 있는 유일한 공식 카운터. 단독으로는 파이프라인 실행 여부를 알려주지 않으므로 `BITBUCKET_BUILD_NUMBER`와 함께 사용 | [Variables and secrets — BITBUCKET_STEP_RUN_NUMBER](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |

## 상태·컨텍스트 변수

아래 변수들은 실행 여부 자체보다 "어떤 저장소·브랜치·트리거로 실행되었는가"라는 부가 정보를 제공합니다. `runby`가 안전하게 노출할 수 있는 실행 메타데이터로 판단한 항목만 선별했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `BITBUCKET_REPO_SLUG` | URL 친화적 저장소 이름 문자열 | 상태·컨텍스트 | 파이프라인이 속한 저장소의 slug | 보조 신호 — 저장소 컨텍스트를 제공하지만 사용자가 임의로 같은 이름의 변수를 저장소/워크스페이스 변수로 재정의할 수 있음 | [Variables and secrets — BITBUCKET_REPO_SLUG](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_WORKSPACE` | 워크스페이스 이름 문자열 | 상태·컨텍스트 | 저장소가 속한 워크스페이스 이름 | 보조 신호 — 조직 단위 컨텍스트이며 실행 주체 확정에는 부족함 | [Variables and secrets — BITBUCKET_WORKSPACE](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_BRANCH` | 브랜치명 문자열 | 상태·컨텍스트 | 빌드를 유발한 소스 브랜치. 브랜치 파이프라인에서만 존재 | 보조 신호 — 태그 기반 빌드에서는 존재하지 않으므로 부재가 곧 비실행을 의미하지 않음 | [Variables and secrets — BITBUCKET_BRANCH](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_COMMIT` | Git 커밋 해시 문자열 | 상태·컨텍스트 | 빌드를 유발한 커밋 해시 | 보조 신호 — 모든 파이프라인에서 제공되지만 커밋 해시만으로는 Bitbucket 실행 여부를 증명하지 못함 | [Variables and secrets — BITBUCKET_COMMIT](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_PR_ID` | 정수 | 상태·컨텍스트 | 풀 리퀘스트로 트리거된 빌드에서만 제공되는 PR 번호 | 보조 신호 — 존재 자체가 "PR 트리거"라는 간접 신호이지만, 공식 문서에 전용 트리거 타입 enum 변수는 없음(아래 결론 참조) | [Variables and secrets — BITBUCKET_PR_ID](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_DEPLOYMENT_ENVIRONMENT` | URL 친화적 환경 이름 문자열 | 상태·컨텍스트 | 배포 스텝에서만 제공되는 배포 환경 이름 | 보조 신호 — 배포 스텝 한정이며 일반 빌드/테스트 스텝에는 존재하지 않음 | [Variables and secrets — BITBUCKET_DEPLOYMENT_ENVIRONMENT](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |
| `BITBUCKET_STEP_TRIGGERER_UUID` | UUID 문자열 | 상태·컨텍스트 | 빌드를 유발한 사용자(push, merge 등)의 UUID. 예약 실행에서는 Pipelines 시스템 사용자 UUID | 보조 신호 — 예약 실행을 간접적으로 구분할 수 있는 유일한 단서이지만 고정된 시스템 UUID 값이 공식 문서에 명시되지 않아 신뢰성 있는 enum 비교가 어려움 | [Variables and secrets — BITBUCKET_STEP_TRIGGERER_UUID](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/) |

다음 변수들은 의도적으로 표에서 제외했습니다.

- **`BITBUCKET_TAG`** — `BITBUCKET_BRANCH`와 상호 배타적으로 태그 빌드에서만 제공되는 동일 성격의 변수라 별도 행으로 추가해도 판정 로직에 새로운 정보를 주지 않습니다.
- **`BITBUCKET_DEPLOYMENT_ENVIRONMENT_UUID`** — `BITBUCKET_DEPLOYMENT_ENVIRONMENT`와 같은 배포 스텝 한정 변수의 UUID 버전으로, `runby`가 노출할 실행 메타데이터로는 이름 쪽이 더 유용해 중복 수록을 피했습니다.
- **`BITBUCKET_SSH_KEY_FILE`** — Bitbucket Cloud의 Linux Docker 러너 한정 파일 경로이며, 실행 주체 판정과 무관한 SSH 설정용 변수입니다.
- **`BITBUCKET_STEP_OIDC_TOKEN`** — OIDC를 명시적으로 활성화한 스텝에만 존재하는 JWT이며, 활성화하지 않은 절대다수의 파이프라인에는 없어 일반 감지 신호로 부적합합니다.
- **`BITBUCKET_TRIGGER_PIPELINE_UUID`** — 파이프라인 완료·배포 완료·패키지 아티팩트 생성 이벤트로 연쇄 트리거된 파이프라인에서만 존재하는 예외적 변수라 일반적인 존재 신호로 다루지 않았습니다.

## 실행 주체 감지에 관한 결론

**존재 마커 문제.** Bitbucket Pipelines는 다른 CI 플랫폼처럼 전용 불리언 존재 마커(`XXX_CI=true` 형태)를 제공하지 않습니다. 공식 문서가 보장하는 `CI=true`는 업계 공용 관행 변수라 단독으로는 Bitbucket 여부를 증명하지 못합니다. 따라서 `runby`는 **`BITBUCKET_BUILD_NUMBER`의 존재(값 형식 검증 포함)를 1차 마커**로 삼고, `CI=true`를 함께 확인하는 방식을 권장합니다. `BITBUCKET_PIPELINE_UUID`와 `BITBUCKET_STEP_UUID`도 함께 존재하면 오탐 가능성이 더 낮아지지만, 이 셋 모두 사용자가 저장소/워크스페이스 변수로 재정의할 수 있는 이름이므로 절대적 신뢰 경계는 아닙니다.

**계층 식별자.** 파이프라인 단위 식별자는 `BITBUCKET_PIPELINE_UUID`, 스텝 단위 식별자는 `BITBUCKET_STEP_UUID`이며, 스텝 재시도 횟수는 `BITBUCKET_STEP_RUN_NUMBER`(1부터 시작하는 정수)로 확인할 수 있습니다. 파이프라인 전체의 재실행(rerun) 횟수를 나타내는 별도의 공식 카운터 변수는 문서에서 확인되지 않았습니다.

**UUID 형식 주의.** `BITBUCKET_PIPELINE_UUID`와 `BITBUCKET_STEP_UUID`의 값은 공식 지원 문서(중괄호 포함 UUID API 오류 KB)에서 확인한 대로 중괄호 `{}`를 포함한 채로 제공됩니다(예: `{11d87b82-13c6-47a3-8e28-73a2bc378675}`). 이 값을 Bitbucket REST API 경로에 그대로 사용하려면 중괄호를 포함하거나 URL 인코딩해야 하므로, `runby`가 이 값을 파싱해 순수 UUID로 정규화하려면 중괄호를 명시적으로 제거해야 합니다.

**트리거 타입 미노출.** Bitbucket Pipelines는 push/PR/manual/scheduled를 구분하는 전용 enum 환경변수를 공식적으로 제공하지 않습니다. 유일하게 확인되는 간접 신호는 PR 빌드에서만 존재하는 `BITBUCKET_PR_ID`와, 예약 실행에서 고정된 시스템 사용자 UUID가 들어간다고 문서화된 `BITBUCKET_STEP_TRIGGERER_UUID`뿐입니다. 두 신호 모두 값 비교(고정 시스템 UUID)나 완전한 커버리지(수동 실행과 push 실행 구분 불가)가 공식적으로 보장되지 않으므로, `runby`는 트리거 타입을 판정하지 않고 실행 여부와 파이프라인/스텝 식별에만 집중하는 편이 안전합니다.

**셀프 호스티드 러너·배포 스텝의 차이.** 공식 문서상 `BITBUCKET_BUILD_NUMBER`, `BITBUCKET_PIPELINE_UUID`, `BITBUCKET_STEP_UUID`, `BITBUCKET_STEP_RUN_NUMBER`, `BITBUCKET_REPO_SLUG`, `BITBUCKET_WORKSPACE`, `BITBUCKET_COMMIT`은 "모든 파이프라인"에서 제공된다고 명시되어 있고, 러너 종류(Bitbucket Cloud 관리형, Linux Docker/Shell, Windows, macOS 셀프 호스티드)에 따른 예외는 이 변수들에 대해 문서화되지 않았습니다. 확인된 유일한 러너별 차이는 (1) `BITBUCKET_SSH_KEY_FILE`이 Bitbucket Cloud + Linux Docker 러너에만 제공된다는 점과 (2) Windows 셀프 호스티드 러너에서는 변수 접근 문법이 PowerShell 방식(`$env:VAR`)으로 달라진다는 점입니다. 배포 스텝에서는 `BITBUCKET_DEPLOYMENT_ENVIRONMENT`와 `BITBUCKET_DEPLOYMENT_ENVIRONMENT_UUID`가 추가로 주입되지만, 이는 배포 스텝 여부를 알려줄 뿐 파이프라인 자체의 실행 여부 판정에는 영향이 없습니다.

**상속과 위조 가능성.** 이 문서의 모든 변수는 스텝 프로세스와 그 자식 프로세스에 환경변수로 상속되며, 사용자가 로컬 셸이나 다른 실행 환경에서 같은 이름의 값을 직접 설정할 수 있습니다. 따라서 `runby`의 `적합` 판정도 절대적 신뢰 경계가 아니라, 공식적으로 Bitbucket Pipelines가 항상 주입한다고 보장하는 값이라는 의미로 한정해서 해석해야 합니다.

## 공식 문서

- [Variables and secrets](https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/)
- [Scheduled and manually triggered pipelines](https://support.atlassian.com/bitbucket-cloud/docs/pipeline-triggers/)
- [Bitbucket Cloud - Pipelines API Error: The value provided is not a valid uuid](https://support.atlassian.com/bitbucket-cloud/kb/bitbucket-cloud-pipelines-api-error-the-value-provided-is-not-a-valid-uuid/)
- [How to access environment variables in Windows Runners on Bitbucket cloud pipelines?](https://support.atlassian.com/bitbucket-cloud/kb/how-to-access-environment-variables-in-windows-runners-on-bitbucket-cloud-pipelines/)
- [Runners](https://support.atlassian.com/bitbucket-cloud/docs/runners/)
- [Step options](https://support.atlassian.com/bitbucket-cloud/docs/step-options/)

## Pull request 감지

`BITBUCKET_PR_ID`는 PR로 트리거된 파이프라인에서만 제공되는 PR 번호입니다. 따라서 존재하면 `PullRequest=true`, 값은 `PullRequestID`로 보고합니다. Bitbucket Pipelines는 별도의 트리거 enum을 보장하지 않으므로 이 직접적인 PR 변수만 사용하며, 변수 이름은 `Evidence`에 기록합니다.
