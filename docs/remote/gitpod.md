---
title: Gitpod
slug: gitpod
research_date: 2026-08-31
open_source: true
repository: https://github.com/gitpod-io/gitpod
product_type: remote_environment
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 "Default Environment Variables" 페이지가 워크스페이스에 기본 제공되는 값으로 GITPOD_WORKSPACE_ID·GITPOD_WORKSPACE_URL·GITPOD_REPO_ROOT 세 개만 명시적으로 열거하고 CI 관련 언급이 전혀 없어 그 부재를 소거법으로만 확인했으며, 커뮤니티에서 언급되나 현재 공식 레퍼런스에는 없는 GITPOD_GIT_USER_NAME·GITPOD_INSTANCE_ID의 실재 여부와 SSH 대 브라우저 세션 간 변수 차이를 실제 워크스페이스에서 `env` 전체 덤프로 검증해야 함
---

# Gitpod

Gitpod은 브라우저·로컬 에디터·SSH로 접속하는 원격 컨테이너 개발 환경입니다. 이 제품은 짧은 기간에 두 번 세대가 바뀌었습니다. `gitpod.io`에서 서비스되던 **Gitpod Classic**은 `.gitpod.yml`로 워크스페이스를 정의했고, 2024년 10월 출시된 **Gitpod Flex**(2025년 4월부터 "Flex" 명칭을 떼고 그냥 "Gitpod"으로 GA, `app.gitpod.io`)는 `devcontainer.json` 기반의 새 아키텍처입니다. 2025년 9월 회사 자체가 **Ona**로 리브랜딩되며 "소프트웨어 엔지니어링 에이전트를 위한 미션 컨트롤"을 표방하기 시작했고, Gitpod Classic PAYG(pay-as-you-go)는 2025년 10월 15일에 로그인·신규 워크스페이스 생성이 모두 막히는 방식으로 종료됐습니다(Enterprise 고객은 별도 마이그레이션 일정 적용). 즉 이 문서의 조사 시점(2026-08-31) 기준으로 "Gitpod"이라는 이름으로 남아 있는 것은 사실상 Ona 산하의 새 아키텍처이며, `GITPOD_*` 환경변수 계약이 공식 문서에 구체적으로 정의된 세대는 **Gitpod Classic**입니다.

이 문서는 Gitpod Classic의 공식 "Environment Variables" 레퍼런스(`gitpod.io`/`www.gitpod.io` 문서가 현재 `ona.com/docs`로 308 리다이렉트되지만 페이지 본문은 여전히 "Gitpod"·"Gitpod Classic" 용어를 그대로 씁니다)를 기준으로 삼습니다. 새 세대(Flex/현재 Gitpod, Ona)는 환경변수를 "Secrets"로 취급해 Project 단위로 구성하고 devcontainer를 통해 파일 마운트나 환경변수로 노출한다고만 설명할 뿐, Classic의 `GITPOD_WORKSPACE_ID`에 대응하는 고정 이름의 존재 마커를 이번 조사에서 확인하지 못했습니다. 따라서 아래 표와 판정 규칙은 **Gitpod Classic에 한정**되며, 신세대(현재 Gitpod/Ona) 워크스페이스에도 동일한 변수가 남아 있는지는 별도 실행 검증 없이는 확정할 수 없습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GITPOD_WORKSPACE_ID` | UUID 문자열 | 실행 식별 | 워크스페이스를 고유하게 식별하는 UUID | 적합 — 공식 문서가 "모든 워크스페이스에 자동으로 설정되는" 핵심 기본 변수 세 개 중 하나로 명시하는 가장 직접적인 존재 마커 | [Environment Variables (Gitpod Classic)](https://www.gitpod.io/docs/configure/workspaces/environment-variables) |
| `GITPOD_WORKSPACE_URL` | URL 문자열 (워크스페이스 고유 URL) | 실행 식별 | 해당 워크스페이스에 접속하는 고유 URL | 적합 — `GITPOD_WORKSPACE_ID`와 함께 공식 문서가 열거하는 기본 변수이며, URL 자체가 워크스페이스 인스턴스를 가리키는 별도 식별자 역할을 함 | [Environment Variables (Gitpod Classic)](https://www.gitpod.io/docs/configure/workspaces/environment-variables) |
| `GITPOD_REPO_ROOT` | 절대 경로 문자열 (예: `/workspace/<저장소 이름>`) | 실행 식별 | 워크스페이스 안에 클론된 git 저장소의 경로 | 보조 신호 — 존재 자체는 Gitpod 워크스페이스임을 뒷받침하지만, 값의 형태(저장소 이름 포함 여부)가 과거 변경된 이력이 있어(GitHub 이슈 보고) 경로 형식만으로 버전을 특정하기는 어려움 | [Environment Variables (Gitpod Classic)](https://www.gitpod.io/docs/configure/workspaces/environment-variables) |

### 공식 문서에서 확인되지 않은 변수: `GITPOD_GIT_USER_NAME`, `GITPOD_INSTANCE_ID`

이 두 변수는 조사 요청 범위에 포함되어 있었지만, 이번 조사에서 접근 가능했던 공식 "Environment Variables" 레퍼런스 페이지(Classic·Enterprise 양쪽 모두 확인)에는 **등장하지 않습니다**. 공식 문서가 git 커밋 신원을 다루는 방식은 `GITPOD_GIT_USER_NAME`이 아니라 표준 git 환경변수인 `GIT_AUTHOR_NAME`/`GIT_AUTHOR_EMAIL`/`GIT_COMMITTER_NAME`/`GIT_COMMITTER_EMAIL`로 기본값(연결된 SCM 계정 정보)을 덮어쓰는 방식입니다. `GITPOD_GIT_USER_NAME`은 일부 GitHub 이슈(예: gitpod-io/gitpod #1800)에서 워크스페이스 부트스트랩 스크립트가 내부적으로 참조한다는 언급이 있으나, 공식 레퍼런스 문서에 계약으로 명시되지 않았고 이번 조사에서 소스 코드로 직접 검증하지도 못했습니다. `GITPOD_INSTANCE_ID`도 마찬가지로 워크스페이스의 실행 인스턴스(재시작마다 바뀌는 값으로 추정)를 가리킨다는 커뮤니티 논의는 있으나, 공식 문서의 "기본 환경변수" 목록에는 없습니다.

이 두 변수는 **`runby`의 실행 식별 신호로 채택하지 않는 것이 안전**합니다. 존재하더라도 공식적으로 보장된 계약이 아니므로 버전·플랜에 따라 사라지거나 이름이 바뀔 수 있습니다.

## CI로 오인되지 않아야 하는 이유

**Gitpod은 `CI` 환경변수를 설정한다는 근거를 공식 문서 어디에서도 찾을 수 없습니다.** Classic의 공식 "Environment Variables" 페이지는 "Default Environment Variables" 절에서 워크스페이스에 자동으로 설정되는 값을 `GITPOD_WORKSPACE_ID`, `GITPOD_WORKSPACE_URL`, `GITPOD_REPO_ROOT` 세 개로 완결된 목록처럼 제시하며, 이 중 `CI`는 없습니다. 문서는 이 외에도 `DOCKERD_ARGS`(도커 인자 커스터마이즈)와 `GITPOD_IMAGE_AUTH`(프라이빗 레지스트리 인증)를 특수 목적 변수로 별도 설명하지만 이 역시 사용자가 명시적으로 설정해야 하는 옵트인 값이며 `CI`와 무관합니다.

이 부재는 `runby` 입장에서 중요한 근거입니다. `runby`는 `CI=true`를 보면 특정 제품을 알 수 없는 범용 CI로 폴백하는데, Gitpod은 애초에 브라우저·에디터·SSH로 **사람이 접속해 있는 것을 전제로 하는 대화형(interactive) 개발 환경**입니다. CI 파이프라인처럼 무인으로 커밋마다 자동 실행되는 것이 아니라, 사용자가 워크스페이스를 열고 그 안에서 셸을 사용하는 구조이므로 "CI가 이 프로세스를 실행했다"는 결론과 정반대의 상황입니다. 공식 문서가 `CI`를 언급하지 않는다는 사실 자체는 "표에 없다"는 소거법적 확인이라 절대적 보장은 아니며, 이 지점이 `runtime_test_required: true`로 표시한 이유 중 하나입니다. 만약 실제 워크스페이스에서 `CI`가 관측된다면 그것은 Gitpod이 설정한 값이 아니라 사용자가 `.gitpod.yml`의 `env`나 사용자/저장소 환경변수 설정으로 직접 넣은 값일 가능성이 높으므로, `runby`는 이를 Gitpod 자체의 신호로 취급해서는 안 됩니다.

## 환경 전달 규칙

Gitpod 워크스페이스는 사용자의 로컬 머신과 완전히 분리된 원격 컨테이너입니다.

- **로컬 머신의 환경은 워크스페이스에 전달되지 않습니다.** 로컬 터미널의 `TERM_PROGRAM`이나 로컬에서 실행 중인 에이전트 CLI가 남긴 마커(`CLAUDECODE`, `CURSOR_AGENT` 등)를 컨테이너 생성 과정에서 그대로 복제하는 메커니즘은 공식 문서 어디에도 없습니다. 따라서 **워크스페이스 내부에서 관측되는 터미널·에이전트 마커는 항상 그 컨테이너 자체에서 새로 실행된 프로세스가 만든 값이며, 사용자의 로컬 노트북을 설명하는 값이 아닙니다.** 예를 들어 워크스페이스 안에서 Claude Code를 실행하면 `CLAUDECODE=1`이 나타나지만, 이는 컨테이너 안에서 Claude Code가 실행됐다는 뜻이지 로컬 머신에 떠 있다는 뜻이 아닙니다.
- **사용자 설정·저장소 설정 환경변수가 주입 경로입니다.** 공식 문서는 워크스페이스 환경변수의 우선순위 체계를 이미지(Dockerfile `ENV`) → `.gitpod.yml[env]`(워크스페이스 전역) → 사용자 설정(계정 대시보드에 저장, 모든 워크스페이스에 적용) → 저장소 설정(조직 대시보드에 저장, 해당 저장소 위에서 시작되는 모든 워크스페이스에 적용하며 `.gitpod.yml[env]`보다 우선) → 컨텍스트 URL 일회성 변수 → 태스크별 변수(`.gitpod.yml[tasks][n][env]`) 순으로 명시합니다. 사용자 설정 값은 "저장은 암호화되지만 워크스페이스 내부에서는 평문으로 노출된다"고 문서가 밝히고 있어, 워크스페이스에 붙는 모든 프로세스에 넓게 반영되는 것으로 보이지만, 어느 시점(부팅 초기 vs 셸 진입 시점)에 정확히 주입되는지를 프로세스별로 구분해 명시하지는 않습니다. `gp env` CLI로 확인·수정할 수 있으나, 문서는 "현재 터미널 세션은 바꾸지 않고 다음 워크스페이스부터 반영된다"고 별도로 경고합니다.
- **SSH 접속과 브라우저 접속이 다른 워크스페이스를 만들지는 않습니다.** 브라우저 IDE, 로컬 에디터의 원격 확장, `gp ssh`/직접 SSH 키 업로드를 통한 SSH 접속은 모두 이미 떠 있는 같은 컨테이너에 붙는 서로 다른 클라이언트일 뿐이며, 접속 수단에 따라 `GITPOD_*` 기본 변수 집합 자체가 달라진다는 근거는 공식 문서에서 확인되지 않았습니다. 다만 커뮤니티에서 "SSH 세션과 워크스페이스 터미널 사이에 `PATH`가 다르게 나타난다"는 보고가 있어(공식 문서로 확정된 사실은 아님), SSH 로그인 셸이 dotfiles·셸 초기화 스크립트를 워크스페이스 터미널과 다르게 로드할 가능성은 남아 있습니다. `runby`가 `GITPOD_WORKSPACE_ID` 같은 컨테이너 속성 변수를 신뢰하는 것과, `PATH`처럼 셸 초기화에 좌우되는 값을 신뢰하는 것은 별개로 다뤄야 합니다.

## 다른 축에 미치는 영향

`runby`가 보고하는 세 축(에이전트·CI·터미널) 중 Gitpod 자체의 공식 변수가 직접 값을 채우는 축은 없습니다.

- **CI 축** — 위에서 확인한 대로 Gitpod은 `CI`를 설정하지 않습니다. `GITPOD_WORKSPACE_ID` 같은 마커가 있다고 해서 `Result.CI`를 채울 근거는 없으며, 오히려 대화형 환경이라는 사실이 CI 판정과 상충합니다.
- **터미널 축** — 공식 문서에 `TERM_PROGRAM`이나 터미널 에뮬레이터 식별용 변수를 Gitpod이 자체적으로 설정한다는 언급이 없습니다. 워크스페이스 내부에서 관측되는 터미널 신호는 그 컨테이너에 실제로 붙은 클라이언트(브라우저 내장 xterm.js, VS Code 통합 터미널 등)가 독립적으로 만드는 값이며, `docs/terminals/`의 감지 규칙과 동일하게 다뤄야 합니다.
- **에이전트 축** — Ona로의 리브랜딩 이후 회사는 "에이전트를 위한 미션 컨트롤"을 표방하지만, 이번 조사에서 접근 가능했던 공식 문서 범위 안에서는 Gitpod/Ona 워크스페이스가 특정 에이전트 실행을 표시하는 고정 이름의 환경변수를 문서화한 근거를 찾지 못했습니다. 워크스페이스 안에서 별도 에이전트 CLI(Claude Code, Codex 등)를 실행하면 그 CLI 자체의 마커가 나타날 뿐이며, 이는 `docs/agents/`의 각 문서가 이미 다루는 영역입니다. 그래서 이 문서의 `executes_agents`는 빈 배열입니다.

즉 `runby`가 Gitpod 마커에서 정직하게 내릴 수 있는 결론은 "이 프로세스가 Gitpod(Classic) 워크스페이스 컨테이너 안에서 실행되고 있다"는 원격 실행 위치 정보뿐이며, 그 안에서 무엇이 그 프로세스를 만들었는지(사람의 셸 입력인지, 에이전트인지)나 그것이 CI인지는 별도 축의 신호로 독립적으로 판단해야 합니다.

## 실행 주체 감지에 관한 결론

가장 신뢰도 높은 존재 마커는 `GITPOD_WORKSPACE_ID`입니다. 공식 문서가 "모든 워크스페이스에 자동으로 설정된다"고 명시하는 UUID 값이며, `GITPOD_WORKSPACE_URL`을 함께 확인하면 어떤 워크스페이스 인스턴스인지까지 보강할 수 있습니다. `GITPOD_REPO_ROOT`는 셋째 기본 변수로 존재를 보강하지만 과거 값의 형태가 바뀐 이력이 있어 경로 파싱에 의존하지 않는 편이 안전합니다.

`GITPOD_GIT_USER_NAME`과 `GITPOD_INSTANCE_ID`는 이번 조사에서 공식 레퍼런스로 확인되지 않았으므로 `runby`의 판정 근거로 채택하지 않습니다. 이 두 값을 지지하는 근거는 커뮤니티 GitHub 이슈뿐이며, 공식 "기본 환경변수" 목록과 대조했을 때 없는 값입니다.

`CI`는 Gitpod이 설정한다는 근거가 전혀 없고, Gitpod은 애초에 사람이 접속해 있는 대화형 환경이므로 CI 판정과 구조적으로 반대되는 제품입니다. 마지막으로, 이 문서의 모든 결론은 **Gitpod Classic**(`gitpod.io`, 2025년 10월 PAYG 종료)에 한정됩니다. 리브랜딩된 Ona/현재 Gitpod(`app.gitpod.io`, Flex 아키텍처)이 동일한 `GITPOD_*` 변수 계약을 유지하는지, 아니면 새 이름의 변수로 교체했는지는 이번 조사에서 공식적으로 확인하지 못했으며, 이를 가정하지 않고 별도 실행 검증이 필요한 미확인 영역으로 남깁니다. 이 값들도 결국 일반 환경변수이므로 워크스페이스 안의 모든 자식 프로세스에 상속되고 사용자가 동일한 이름으로 직접 export하면 위조될 수 있다는 한계는 다른 플랫폼과 같습니다.

## 공식 문서

- [Environment Variables (Gitpod Classic)](https://www.gitpod.io/docs/configure/workspaces/environment-variables)
- [Environment Variables (Gitpod Enterprise)](https://www.gitpod.io/docs/enterprise/configure/workspaces/environment-variables)
- [Environment Variables in Gitpod (User settings)](https://www.gitpod.io/docs/configure/user-settings/environment-variables)
- [Environment Variables on Repository level (Gitpod Classic)](https://www.gitpod.io/docs/configure/repositories/environment-variables)
- [Secrets & Environment Variables (Gitpod Flex)](https://www.gitpod.io/docs/flex/secrets)
- [Gitpod Classic PAYG sunset on October 15th · Ona](https://ona.com/stories/gitpod-classic-payg-sunset)
- [Naming is hard: taking 'Flex' into the future · Ona](https://ona.com/stories/naming-is-hard)
- [gitpod-io/gitpod (source repository, AGPL-3.0, README states the project is superseded by Ona)](https://github.com/gitpod-io/gitpod)
- [Environment Variables in Workspaces · Issue #2374 · gitpod-io/gitpod](https://github.com/gitpod-io/gitpod/issues/2374)
