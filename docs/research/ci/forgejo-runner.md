---
title: Forgejo Actions (Forgejo Runner)
slug: forgejo-runner
research_date: 2026-08-31
open_source: true
repository: https://code.forgejo.org/forgejo/runner
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Forgejo Actions 최신 공식 레퍼런스가 Runner v7 이상에서 주입되는 전용 FORGEJO_* 변수, 값의 의미, 재실행 회차와 GITHUB_* 호환 별칭까지 명시하므로 감지 규칙을 문서만으로 확정할 수 있음
---

# Forgejo Actions (Forgejo Runner)

Forgejo Actions는 Forgejo 서버에 내장된 CI 시스템이고, 실제 workflow job은 별도 오픈소스 프로그램인 Forgejo Runner가 실행합니다. 공식 레퍼런스는 각 step의 셸 환경에 `FORGEJO_ACTIONS=true`와 `CI=true`를 주입한다고 명시합니다. 따라서 `runby`가 workflow step의 자식 프로세스로 실행되면 `FORGEJO_ACTIONS`가 가장 직접적인 실행 주체 식별 신호입니다.

이 문서는 현재 최신 문서인 Forgejo v15.0과 Forgejo Runner v7 이상을 기준으로 합니다. v7부터 모든 `FORGEJO_*` 변수가 제공되며, 이전 Runner는 `GITHUB_*` 이름만 정의했습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `FORGEJO_ACTIONS` | 불리언 문자열 (`true`) | 실행 식별 | Forgejo Runner가 workflow를 실행 중임을 표시 | 적합 — 공식 문서가 Runner 실행 중 항상 `true`라고 명시하는 전용 마커 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `CI` | 불리언 문자열 (`true`) | 실행 식별 | 범용 CI 실행 표시 | 보조 신호 — 공식 문서상 항상 `true`지만 여러 CI 제품이 공유하므로 Forgejo를 특정하지 못함 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_RUN_ID` | 정수형 문자열 | 실행 식별 | Forgejo 인스턴스 전체에서 현재 workflow run을 고유하게 식별 | 적합 — pipeline/run 수준의 고유 식별자 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_RUN_ATTEMPT` | 정수형 문자열 (`1`부터 시작) | 실행 식별 | 같은 run의 실행 회차. job 재실행마다 증가 | 적합 — 별도 보정 없이 공통 `Attempt`에 사용할 수 있음 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_RUN_NUMBER` | 정수형 문자열 | 실행 식별 | 해당 workflow 저장소 안에서 현재 run을 식별 | 보조 신호 — 저장소 범위 값이므로 인스턴스 전체 고유 ID인 `FORGEJO_RUN_ID`를 우선함 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_JOB` | workflow YAML의 `job_id` 문자열 | 실행 식별 | 현재 job을 식별 | 적합 — run 내부 job 식별자이며 전역 고유 ID는 아니므로 `FORGEJO_RUN_ID`와 조합해야 함 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_ACTION` | 숫자형 step id 문자열 | 실행 식별 | 현재 step을 식별 | 보조 신호 — step 단위 위치를 보강하지만 job이나 run의 대체 식별자는 아님 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `FORGEJO_EVENT_NAME` | 이벤트명 문자열 (`push`, `pull_request` 등) | 상태·컨텍스트 | workflow를 트리거한 이벤트 | 보조 신호 — 공통 `Trigger`에 적합하지만 존재 마커는 아님 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_REPOSITORY` | `owner/repository` 문자열 | 상태·컨텍스트 | workflow가 실행되는 저장소 | 보조 신호 — 실행 출처 저장소를 보강 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_ACTOR` | 사용자명 문자열 | 상태·컨텍스트 | workflow를 트리거한 사용자 | 보조 신호 — 사람 또는 자동화 주체의 이름을 제공 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_REF` | `refs/...` 문자열 | 상태·컨텍스트 | 트리거와 연결된 완전한 Git ref | 보조 신호 — 실행 대상 브랜치·태그·PR을 보강 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_REF_NAME` | 짧은 ref 이름 | 상태·컨텍스트 | 브랜치 또는 태그 이름 | 보조 신호 — 표시용 컨텍스트 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_SHA` | 커밋 SHA 문자열 | 상태·컨텍스트 | workflow를 트리거한 커밋 | 보조 신호 — 이벤트에 따라 정확한 의미가 달라질 수 있음 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_SERVER_URL` | URL 문자열 | 상태·컨텍스트 | workflow를 제공한 Forgejo 인스턴스 | 보조 신호 — 자체 호스팅 인스턴스 위치를 보강 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_WORKFLOW_REF` | `<owner>/<repo>/.forgejo/workflows/<file>@<ref>` 형식 | 상태·컨텍스트 | 실행된 workflow 정의와 ref | 보조 신호 — 어떤 workflow 정의인지 특정 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `FORGEJO_WORKSPACE` | 디렉터리 경로 | 상태·컨텍스트 | step의 기본 작업 디렉터리 | 보조 신호 — 실행 위치이며 제품 판정에는 사용하지 않음 | [Actions reference: env context](https://forgejo.org/docs/latest/user/actions/reference/#env) |
| `RUNNER_OS` | `Linux`, `Windows`, `macOS`, `undefined` | 상태·컨텍스트 | Runner 운영체제 | 보조 신호 — Forgejo 전용이 아니며 플랫폼 정보에만 사용 | [Actions reference: runner context](https://forgejo.org/docs/latest/user/actions/reference/#runner) |
| `RUNNER_ARCH` | `X86`, `X64`, `ARM`, `ARM64`, `undefined` | 상태·컨텍스트 | Runner CPU 아키텍처 | 보조 신호 — Forgejo 전용이 아니며 플랫폼 정보에만 사용 | [Actions reference: runner context](https://forgejo.org/docs/latest/user/actions/reference/#runner) |

### 의도적으로 제외한 변수

`FORGEJO_ENV`, `FORGEJO_OUTPUT`, `FORGEJO_PATH`, `FORGEJO_EVENT_PATH`는 step 간 workflow command나 이벤트 payload를 전달하는 파일 경로이고, `FORGEJO_API_URL`은 API endpoint이므로 실행 식별 데이터로 수집하지 않습니다. `FORGEJO_ACTION_PATH`와 `FORGEJO_ACTION_REPOSITORY`는 composite 또는 재사용 action에서만 의미가 있어 일반 step의 안정적인 마커로 사용하지 않습니다. PR 이벤트에만 존재하는 `FORGEJO_BASE_REF`와 `FORGEJO_HEAD_REF`도 핵심 감지 규칙에서는 제외합니다.

## GitHub Actions 호환 별칭과 판정 우선순위

공식 문서는 모든 `FORGEJO_*` 변수를 같은 suffix의 `GITHUB_*` 변수로도 제공한다고 명시합니다. 예를 들어 Forgejo에서 `FORGEJO_RUN_ID`와 `GITHUB_RUN_ID`, `FORGEJO_REPOSITORY`와 `GITHUB_REPOSITORY`가 함께 존재합니다. 따라서 `GITHUB_ACTIONS=true`만 먼저 검사하면 Forgejo Actions를 GitHub Actions로 잘못 분류할 수 있습니다.

권장 판정 순서는 다음과 같습니다.

1. `FORGEJO_ACTIONS == "true"`이면 Forgejo Actions로 확정한다.
2. `FORGEJO_RUN_ID`, `FORGEJO_JOB`, `FORGEJO_RUN_ATTEMPT`를 각각 pipeline, job, attempt에 사용한다.
3. 그 다음에만 `GITHUB_ACTIONS == "true"`를 GitHub Actions 후보로 검사한다.
4. `CI=true`와 `RUNNER_*`는 제품을 특정하지 못하므로 보조 신호로만 사용한다.

Forgejo Runner v7 미만에서는 `FORGEJO_*`가 없고 `GITHUB_*`만 정의됩니다. 이 구형 환경은 환경변수만으로 GitHub Actions와 안전하게 구별할 전용 마커가 없으므로, `runby`가 이를 Forgejo로 단정하면 안 됩니다. 최신 규칙의 지원 하한을 Runner v7로 명시하거나, 구형 Runner를 GitHub 호환/불명 CI로 낮은 신뢰도로 보고하는 편이 안전합니다.

## 실행 주체 감지에 관한 결론

Forgejo Runner v7 이상에서는 `FORGEJO_ACTIONS=true`가 가장 강한 존재 마커입니다. `FORGEJO_RUN_ID`는 인스턴스 전체 run ID, `FORGEJO_JOB`은 workflow 내부 job id이므로 두 값을 함께 사용해야 실행 위치를 안정적으로 표현할 수 있습니다. `FORGEJO_RUN_ATTEMPT`는 이미 1부터 시작하므로 정규화할 때 더하지 않습니다.

Docker·Podman·LXC·host 등 Runner 실행 방식은 격리와 경로에 영향을 주지만, 공식 Actions reference는 step 셸에 같은 컨텍스트 환경변수를 제공하는 계약으로 설명합니다. 다만 이 값들도 일반 환경변수여서 자식 프로세스로 상속되고 사용자가 흉내 낼 수 있으므로 보안상 신뢰 경계로 사용하면 안 됩니다.

## 공식 문서와 소스

- [Forgejo Actions overview](https://forgejo.org/docs/latest/user/actions/overview/)
- [Forgejo Actions reference — env 및 runner context](https://forgejo.org/docs/latest/user/actions/reference/#env)
- [Forgejo Runner installation](https://forgejo.org/docs/latest/admin/actions/installation/)
- [Forgejo Runner 공식 소스](https://code.forgejo.org/forgejo/runner)
