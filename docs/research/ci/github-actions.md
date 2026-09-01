---
title: GitHub Actions
slug: github-actions
research_date: 2026-08-31
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 Variables reference 문서가 모든 기본 환경변수의 이름, 형식, 재실행 시 안정성을 명시적으로 문서화하고 있어 별도 실행 검증 없이도 감지 규칙을 확정할 수 있음
---

# GitHub Actions

GitHub Actions는 워크플로 실행 중 모든 스텝에 `GITHUB_*`와 `RUNNER_*` 접두사를 가진 기본 환경변수 집합을 항상 주입하며, 공식 문서는 이 값들을 사용자가 덮어쓸 수 없다고 명시합니다. 다만 `CI` 변수는 "현재는" 덮어쓸 수 있지만 이 보장이 계속 유지되지는 않는다고 문서에 별도로 적혀 있습니다. `GITHUB_ACTIONS`, `GITHUB_RUN_ID`, `GITHUB_JOB` 같은 값은 러너의 셸 환경에 프로세스 생성 시점부터 존재하므로, `runby`가 자식 프로세스에서 이를 읽으면 GitHub Actions 러너가 이 프로세스를 실행했다는 강한 근거가 됩니다. 단, 이 값들은 일반 환경변수이므로 상속되거나(예: 컨테이너를 벗어난 별도 백그라운드 데몬으로 유출) 로컬 재현 스크립트에서 사용자가 그대로 흉내 낼 수 있다는 한계는 다른 플랫폼과 동일합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GITHUB_ACTIONS` | 불리언 문자열 (`true`) | 실행 식별 | GitHub Actions가 워크플로를 실행 중임을 표시 | 적합 — 문서가 "GitHub Actions가 워크플로를 실행할 때 항상 `true`로 설정된다"고 명시하는 가장 직접적인 존재 마커 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `CI` | 불리언 문자열 (`true`) | 실행 식별 | 로컬 실행과 GitHub Actions 실행을 구분하는 범용 CI 마커 | 보조 신호 — 문서 스스로 "현재는 덮어쓸 수 있으며 앞으로도 그렇다는 보장은 없다"고 밝힌 유일한 기본 변수이므로 `GITHUB_ACTIONS`보다 신뢰도가 낮음 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_RUN_ID` | 정수형 문자열 | 실행 식별 | 저장소 내 워크플로 실행(run)을 고유하게 식별. 재실행해도 값이 바뀌지 않음 | 적합 — 파이프라인(run) 단위의 안정적 식별자로 로그·API 조회에 직접 대응됨 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_RUN_ATTEMPT` | 정수형 문자열 (`1`부터 시작) | 실행 식별 | 같은 run에 대한 재시도(재실행) 횟수. 최초 시도는 `1`, 재실행마다 증가 | 적합 — `GITHUB_RUN_ID`와 결합해 정확히 어떤 시도(attempt)인지까지 특정 가능 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_RUN_NUMBER` | 정수형 문자열 (`1`부터 시작) | 실행 식별 | 저장소별 워크플로 실행 순번. 재실행해도 값이 바뀌지 않음 | 보조 신호 — 사람이 읽기 쉬운 실행 번호이지만 `GITHUB_RUN_ID`처럼 API 조회에 바로 쓰이는 고유 키는 아님 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_JOB` | 워크플로 YAML의 job id 문자열 | 실행 식별 | 현재 실행 중인 job의 `job_id` | 적합 — run 내에서 어떤 job이 이 프로세스를 실행했는지 특정하는 job 단위 식별자 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_ACTION` | 문자열 (`__run`, `__run_2`, 또는 `__<owner>_<action-repo>` 형태) | 실행 식별 | 현재 실행 중인 action의 이름 또는 id 없는 스크립트 스텝의 자동 생성 이름 | 보조 신호 — 스텝 단위 실행 위치를 나타내지만 값 형식이 스텝 종류(스크립트/action)에 따라 달라지고 재사용 워크플로에서는 값이 겹칠 수 있음 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_ACTION_PATH` | 절대 디렉터리 경로 | 실행 식별 | 현재 action이 위치한 경로 | 보조 신호 — 문서가 "composite action에서만 지원된다"고 명시해 일반 스크립트 스텝에는 존재하지 않음 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |

## 상태·컨텍스트 변수

`runby`가 실행 메타데이터로 노출할 만한 값만 선별했습니다. 저장소·ref·트리거·러너 신원을 대표하는 변수 위주이며, 파일 경로 계열(`GITHUB_ENV`, `GITHUB_PATH`, `GITHUB_OUTPUT`, `GITHUB_STEP_SUMMARY`)과 API 엔드포인트 계열(`GITHUB_API_URL`, `GITHUB_GRAPHQL_URL`, `GITHUB_SERVER_URL`)은 워크플로 명령이 파일을 주고받거나 API를 호출하기 위한 배관(plumbing)일 뿐 실행 주체 식별에 쓰이지 않으므로 제외했습니다. `GITHUB_BASE_REF`/`GITHUB_HEAD_REF`(PR 이벤트에서만 존재), `GITHUB_WORKFLOW_REF`/`GITHUB_WORKFLOW_SHA`(워크플로 정의 자체의 메타데이터), `GITHUB_RETENTION_DAYS`(보존 정책 설정)도 이벤트 종류에 따라 조건부로만 존재하거나 실행 주체와 직접 관련이 없어 표에서는 생략했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GITHUB_EVENT_NAME` | 문자열 (`push`, `pull_request`, `schedule`, `workflow_dispatch` 등) | 상태·컨텍스트 | 워크플로를 트리거한 이벤트 종류 | 보조 신호 — 트리거 유형을 알려주지만 GitHub Actions 자체 실행 여부는 `GITHUB_ACTIONS`가 먼저 확정해야 함 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_REPOSITORY` | `owner/repo` 형식 문자열 | 상태·컨텍스트 | 현재 워크플로가 실행 중인 저장소 | 보조 신호 — 어떤 저장소의 실행인지 특정하는 컨텍스트 값 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_SHA` | 40자 커밋 SHA 문자열 | 상태·컨텍스트 | 트리거된 커밋의 SHA (이벤트 종류에 따라 의미가 달라짐) | 보조 신호 — 실행 대상 커밋을 식별하는 컨텍스트 값 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_REF` | `refs/heads/<branch>`, `refs/pull/<n>/merge`, `refs/tags/<tag>` 등 | 상태·컨텍스트 | 워크플로를 트리거한 완전한 ref | 보조 신호 — 브랜치/태그/PR 컨텍스트를 나타내는 보강 정보 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_ACTOR` | 사용자명 문자열 | 상태·컨텍스트 | 워크플로 실행을 트리거한 사용자 | 보조 신호 — 실행 주체(누가)를 나타내지만 값 자체는 위조·재사용 워크플로 상황에서 문맥에 따라 달라질 수 있음 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `GITHUB_WORKFLOW` | 워크플로 이름 또는 파일 경로 문자열 | 상태·컨텍스트 | 현재 실행 중인 워크플로의 이름 | 보조 신호 — 어떤 워크플로 정의가 이 프로세스를 만들었는지 알려주는 컨텍스트 값 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `RUNNER_ENVIRONMENT` | 문자열 (`github-hosted` 또는 `self-hosted`) | 상태·컨텍스트 | 실행 중인 러너가 GitHub 호스팅형인지 자체 호스팅형인지 표시 | 보조 신호 — GitHub Actions 실행이라는 확정 이후 러너 종류를 구분하는 보강 정보 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `RUNNER_OS` | 문자열 (`Linux`, `Windows`, `macOS`) | 상태·컨텍스트 | 러너의 운영체제 | 보조 신호 — 플랫폼별 분기에 쓸 수 있는 컨텍스트 값 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |
| `RUNNER_DEBUG` | 문자열 (`1`) | 디버그·실험 | 러너 디버그 로깅 활성화 여부. 디버그 로깅이 켜졌을 때만 존재 | 부적합 — 디버그 모드가 아니면 존재하지 않는 조건부 값이며 실행 주체 식별과 무관 | [Default environment variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) |

## 실행 주체 감지에 관한 결론

가장 신뢰도 높은 존재 마커는 `GITHUB_ACTIONS=true`입니다. 공식 문서가 GitHub Actions 실행 시 항상 `true`로 설정된다고 명시하며, `GITHUB_*`/`RUNNER_*` 계열 기본 변수는 사용자가 덮어쓸 수 없다고 별도로 못박고 있어 다른 플랫폼의 유사 마커보다 위조 저항성이 높습니다. `CI`는 여러 CI 플랫폼이 공유하는 범용 변수이자 GitHub 문서 스스로 "현재는 덮어쓸 수 있다"고 인정한 유일한 기본 변수이므로, `runby`는 `GITHUB_ACTIONS`를 1차 판정 기준으로 삼고 `CI`는 존재 여부를 재확인하는 보조 신호로만 사용해야 합니다.

단, Forgejo Runner는 호환성을 위해 동일한 `GITHUB_*` 변수들을 함께 주입합니다. 그러므로 전체 CI 레지스트리에서는 먼저 `FORGEJO_ACTIONS=true`를 검사하고, Forgejo 전용 마커가 없을 때 `GITHUB_ACTIONS=true`를 GitHub Actions 신호로 해석해야 합니다. Forgejo Runner v7 미만은 `FORGEJO_*` 없이 `GITHUB_*`만 제공하므로 환경변수만으로 GitHub Actions와 확정적으로 구분할 수 없습니다.

권장 판정 순서는 다음과 같습니다.

1. `FORGEJO_ACTIONS == "true"`가 아닌지 먼저 확인한 뒤, `GITHUB_ACTIONS == "true"`이면 GitHub Actions로 판정한다.
2. `GITHUB_RUN_ID`를 파이프라인(run) 수준 식별자로, `GITHUB_RUN_ATTEMPT`를 재시도 횟수로 함께 읽는다. 두 값을 조합하면 재실행 여부와 몇 번째 시도인지까지 구분된다.
3. `GITHUB_JOB`을 job 수준 식별자로 읽는다. `GITHUB_RUN_ID`+`GITHUB_JOB` 조합이 이 프로세스가 속한 job을 특정한다.
4. 필요하면 `GITHUB_EVENT_NAME`으로 트리거 유형(`push`/`pull_request`/`schedule`/`workflow_dispatch` 등)을, `GITHUB_REPOSITORY`·`GITHUB_REF`·`GITHUB_SHA`·`GITHUB_ACTOR`로 상태 컨텍스트를 보강한다.

**self-hosted 러너, 컨테이너 job, composite/재사용 워크플로에 대한 참고 사항**: 공식 Variables reference 문서는 `GITHUB_ACTIONS`, `GITHUB_RUN_ID`, `GITHUB_RUN_ATTEMPT`, `GITHUB_JOB` 같은 핵심 식별 변수가 self-hosted 러너나 컨테이너 job에서 다르게 동작한다고 별도로 명시하지 않습니다. `RUNNER_ENVIRONMENT`만 호스팅 방식을 구분하는 값을 제공하고, `RUNNER_TOOL_CACHE`는 문서상 "GitHub 호스팅 러너에서만" 유효하다고 명시되어 self-hosted 환경에서는 신뢰할 수 없습니다. composite action 내부에서는 `GITHUB_ACTION_PATH`가 추가로 존재하지만(문서: "composite action에서만 지원") 이는 스텝 단위 보조 정보이며 `GITHUB_ACTIONS`/`GITHUB_RUN_ID` 판정 자체에는 영향을 주지 않습니다. 재사용 워크플로(reusable workflow) 호출 시 `GITHUB_ACTION` 값의 정확한 형식 차이는 이번 조사에서 사용한 공식 문서에 세부 기술이 없어, `runby`는 이를 실행 식별의 1차 근거로 사용하지 않고 보조 신호로만 취급합니다.

마지막으로, 이 모든 값은 일반 환경변수이므로 러너가 만든 자식 프로세스 트리 전체에 상속되고, 로컬에서 같은 이름의 변수를 그대로 설정하면 위조될 수 있습니다. `runby`는 이를 절대적 신뢰 경계가 아니라 GitHub Actions 러너가 이 프로세스 트리를 만들었다는 강한 정황 증거로 다뤄야 합니다.

## 공식 문서

- [Variables reference (Default environment variables)](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables)
- [Store information in variables](https://docs.github.com/en/actions/learn-github-actions/environment-variables)
- [Contexts reference](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts)

## Pull request 감지

`GITHUB_EVENT_NAME`이 `pull_request`일 때 `runby`는 `PullRequest=true`로 정규화합니다. 이 변수는 이벤트 종류만 광고하고 PR 번호를 기본 환경변수로 직접 제공하지 않으므로 `PullRequestID`는 비워 둡니다. `GITHUB_EVENT_PATH`의 JSON payload를 추가로 읽지 않는 이유는 파일 경로와 payload가 실행 식별 변수 계약보다 불안정하고, 현재 요구 범위를 넘기 때문입니다. 이벤트 변수 이름은 `Evidence`에 기록하되 값은 기록하지 않습니다.
