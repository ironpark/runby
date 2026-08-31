---
title: GitHub Codespaces
slug: github-codespaces
research_date: 2026-08-31
open_source: false
repository: null
product_type: remote_environment
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 문서가 기본 환경변수 표를 완결된 목록으로 제시하지만 CI·GITHUB_ACTIONS의 부재는 "표에 없다"는 소거법으로만 확인되므로, 실제 Codespace를 띄워 env 전체 덤프로 CI·GITHUB_ACTIONS 부재와 devcontainer.json remoteEnv/containerEnv 주입 결과를 직접 검증해야 함
---

# GitHub Codespaces

GitHub Codespaces는 저장소별로 클라우드에 컨테이너를 띄우고 브라우저·VS Code·SSH로 접속해 개발하는 원격 개발 환경입니다. 공식 문서는 모든 Codespace에 `CODESPACES=true`를 포함한 고정된 기본 환경변수 집합을 주입한다고 명시하며, 이 집합은 `GITHUB_*` 접두사를 다수 포함하지만 GitHub Actions가 사용하는 `GITHUB_ACTIONS`나 범용 CI 마커 `CI`는 이 목록에 등장하지 않습니다. `runby`가 이미 `GITHUB_ACTIONS=true`로 GitHub Actions를, `FORGEJO_ACTIONS=true`(및 구형 Runner의 `GITHUB_*` 별칭)로 Forgejo Actions를 판정하는 상황에서, Codespaces는 같은 `GITHUB_*` 네임스페이스를 세 번째로 소비하는 제품이라는 점에서 이 문서의 핵심은 "겹치는 변수"가 아니라 "겹치지 않는 변수"를 정확히 가려내는 데 있습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CODESPACES` | 불리언 문자열 (`true`) | 실행 식별 | Codespace 내부에서 항상 `true`로 설정되는 존재 마커 | 적합 — 공식 문서가 "Codespace 안에서는 항상 `true`"라고 명시하는 가장 직접적인 마커 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `CODESPACE_NAME` | 문자열 (예: `octocat-literate-space-parakeet-mld5`) | 실행 식별 | 현재 Codespace 인스턴스의 고유 이름 | 적합 — Codespace 인스턴스를 개별적으로 식별하는 고유 식별자이며, 포트 포워딩 URL(`https://<CODESPACE_NAME>-<port>.app.github.dev`) 구성에도 쓰이는 안정적인 이름 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GITHUB_USER` | 사용자명 문자열 (예: `octocat`) | 상태·컨텍스트 | Codespace를 생성(초기화)한 사용자 이름 | 보조 신호 — "누가 실행했는가"를 알려주지만 존재 자체가 Codespace 마커는 아니며 `CODESPACES`로 확정한 뒤 보강용으로만 사용 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |

## 상태·컨텍스트 변수

공식 문서가 열거하는 기본 변수 중 실행 위치·저장소·API 배관에 해당하는 값들입니다. `GIT_COMMITTER_EMAIL`/`GIT_COMMITTER_NAME`은 커밋 메타데이터용, `GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN`은 포트 포워딩 도메인용으로 실행 주체 식별과는 거리가 있어 표에서는 후순위로 다룹니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GITHUB_REPOSITORY` | `owner/repo` 형식 문자열 (예: `octocat/Hello-World`) | 상태·컨텍스트 | 현재 Codespace가 연결된 저장소 | 보조 신호 — 어떤 저장소의 Codespace인지 특정하는 컨텍스트 값. GitHub Actions·Forgejo도 동일 이름의 변수를 쓰므로 단독으로는 세 제품을 구분하지 못함 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GITHUB_SERVER_URL` | URL 문자열 (예: `https://github.com`) | 상태·컨텍스트 | GitHub 서버(GHEC/GHES 인스턴스)의 기본 URL | 보조 신호 — 어느 GitHub 인스턴스에 속한 Codespace인지 알려주는 배관용 값 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GITHUB_API_URL` | URL 문자열 (예: `https://api.github.com`) | 상태·컨텍스트 | REST API 엔드포인트 | 부적합 — API 호출용 배관이며 실행 주체 식별과 무관 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GITHUB_GRAPHQL_URL` | URL 문자열 (예: `https://api.github.com/graphql`) | 상태·컨텍스트 | GraphQL API 엔드포인트 | 부적합 — API 호출용 배관이며 실행 주체 식별과 무관 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN` | 도메인 문자열 (예: `app.github.dev`) | 상태·컨텍스트 | 포트 포워딩에 쓰이는 도메인 | 부적합 — 네트워킹 설정값이며 실행 주체 식별과 무관 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GIT_COMMITTER_NAME` | 문자열 | 설정 | 향후 git 커밋의 committer 이름 필드 기본값 | 부적합 — git 설정값이며 실행 주체 식별과 무관 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GIT_COMMITTER_EMAIL` | 이메일 문자열 | 설정 | 향후 git 커밋의 author 필드 기본값 | 부적합 — git 설정값이며 실행 주체 식별과 무관 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |
| `GITHUB_TOKEN` | 서명된 토큰 문자열 (자격 증명) | 상태·컨텍스트 | Codespace 내부 사용자를 대표하는 서명된 인증 토큰으로, GitHub API 인증 호출에 사용 | 부적합 — 값 자체가 자격 증명이므로 `runby`가 읽거나 로깅해서는 안 되며, 존재 여부만으로는 GitHub Actions·Forgejo에도 동일 이름 변수가 있어 제품 구분에 쓸 수 없음 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |

`GITHUB_TOKEN`은 이름만 언급합니다. 값은 실제 API 인증에 쓰이는 자격 증명이므로 `runby`는 이 변수의 존재 여부만 참고하고 값을 읽거나 로그로 남겨서는 안 됩니다.

공식 문서는 위 표에 실린 변수들을 "모든 Codespace에 기본 제공되는" 값으로 열거하며, `devcontainer.json`에서 별도로 설정하지 않아도 컨테이너 생성 시점부터 존재합니다. 반대로 이 목록에 없는 값(사용자 지정 환경변수, 서비스 접속 정보 등)은 저장소 소유자나 사용자가 `devcontainer.json`의 `remoteEnv`/`containerEnv` 또는 Codespaces secrets로 직접 구성해야 나타납니다. 즉 표의 변수는 "기본값", 그 밖의 값은 "옵트인 설정값"으로 나뉩니다.

## CI와의 구별

`runby`는 이미 `GITHUB_ACTIONS=true`로 GitHub Actions를, `FORGEJO_ACTIONS=true`(및 Forgejo Runner v7 미만의 `GITHUB_*` 별칭)로 Forgejo Actions를 판정합니다. Codespaces는 같은 `GITHUB_*` 네임스페이스의 세 번째 소비자이므로, 이 절이 이 문서에서 가장 중요합니다.

**Codespaces는 `GITHUB_ACTIONS`를 설정하지 않습니다.** 공식 "Default environment variables for your codespace" 문서가 나열하는 기본 변수 표에는 `CODESPACE_NAME`, `CODESPACES`, `GIT_COMMITTER_EMAIL`, `GIT_COMMITTER_NAME`, `GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN`, `GITHUB_API_URL`, `GITHUB_GRAPHQL_URL`, `GITHUB_REPOSITORY`, `GITHUB_SERVER_URL`, `GITHUB_TOKEN`, `GITHUB_USER` 11개만 있고, `GITHUB_ACTIONS`는 등장하지 않습니다. `docs/ci/github-actions.md`가 근거로 삼는 GitHub Actions "Default environment variables" 문서와 비교하면, Codespaces 쪽 목록은 `GITHUB_REPOSITORY`·`GITHUB_SERVER_URL`·`GITHUB_API_URL`·`GITHUB_TOKEN` 네 개만 이름이 겹치고 나머지(`GITHUB_ACTIONS`, `GITHUB_RUN_ID`, `GITHUB_RUN_ATTEMPT`, `GITHUB_JOB`, `GITHUB_SHA`, `GITHUB_REF`, `GITHUB_ACTOR`, `GITHUB_WORKFLOW` 등)는 Codespaces 쪽에 전혀 없습니다. 이 부재가 두 컨텍스트를 갈라놓는 핵심 근거입니다.

**Codespaces가 `CI`를 설정하는지도 공식 문서상 확인되지 않습니다.** 같은 기본 변수 표에 `CI`는 등장하지 않으며, `runby`는 `CI=true`를 보면 제품을 특정하지 못하는 범용 CI로 폴백하므로, 만약 Codespaces가 이 변수를 조용히 설정한다면 대화형 개발 환경이 CI 작업으로 오판될 위험이 있습니다. 공식 문서가 이 변수의 존재를 언급하지 않는다는 사실 자체는 강한 근거이지만, "표에 없다"는 소거법적 확인이므로 절대적 보장은 아닙니다. 이 지점이 `runtime_test_required: true`로 표시한 이유이며, 실제 Codespace에서 `env` 전체 덤프를 떠서 `CI`와 `GITHUB_ACTIONS`가 정말 부재하는지 직접 확인하는 검증이 필요합니다.

겹치는 네 변수(`GITHUB_REPOSITORY`, `GITHUB_SERVER_URL`, `GITHUB_API_URL`, `GITHUB_TOKEN`)는 이름만 같을 뿐 GitHub Actions·Forgejo와 별도로 각 제품이 독립적으로 주입하는 값이므로, 이름이 겹친다는 사실만으로는 어느 제품인지 구분할 수 없습니다. 반드시 존재 마커를 먼저 확인해야 합니다.

**권장 판정 순서**

1. `FORGEJO_ACTIONS == "true"`이면 Forgejo Actions로 판정한다 (`docs/ci/forgejo-runner.md` 규칙).
2. 그렇지 않고 `GITHUB_ACTIONS == "true"`이면 GitHub Actions로 판정한다.
3. `GITHUB_ACTIONS`와 `FORGEJO_ACTIONS`가 모두 없고 `CODESPACES == "true"`이면 GitHub Codespaces로 판정한다. `GITHUB_REPOSITORY`·`GITHUB_SERVER_URL` 등 겹치는 이름의 변수가 있어도 무시하고 `CODESPACES`를 1차 마커로 삼는다.
4. `CI`는 어느 판정에서도 1차 근거로 쓰지 않는다. GitHub Actions·Forgejo 판정 이후의 보조 신호로만 재확인하고, Codespaces 판정에는 아예 관여시키지 않는다.

이 순서를 지키면 `CI`나 `GITHUB_*` 이름 겹침 때문에 Codespace 내부의 대화형 세션이 CI 파이프라인으로 잘못 보고되는 일을 막을 수 있습니다.

## 환경 전달 규칙

Codespace는 사용자의 로컬 머신과는 별개의 원격 컨테이너입니다. 로컬 터미널의 `TERM_PROGRAM`, 로컬에서 실행 중인 에이전트 CLI가 남긴 마커(`CLAUDECODE`, `CURSOR_AGENT` 등) 같은 값은 컨테이너 생성 과정에서 전달되지 않으며, 공식 문서가 나열하는 기본 변수 목록에도 이런 로컬 신호를 그대로 복제하는 메커니즘은 없습니다. 따라서 **Codespace 내부에서 관측되는 터미널·에이전트 마커는 항상 그 컨테이너 자체에서 새로 실행된 프로세스가 만든 값이며, 사용자의 로컬 노트북을 설명하는 값이 아닙니다.** 예를 들어 Codespace 안에서 Claude Code를 실행하면 `CLAUDECODE=1`이 나타나지만, 이는 컨테이너 안에서 Claude Code가 실행됐다는 뜻이지 로컬 머신에 Claude Code가 떠 있다는 뜻이 아닙니다.

접속 수단(브라우저의 웹 기반 VS Code, 로컬 VS Code 데스크톱에서 Codespaces 확장으로 접속, `gh` CLI를 통한 SSH 접속)은 모두 같은 원격 컨테이너에 붙는 서로 다른 클라이언트일 뿐입니다. 공식 문서는 이 세 가지를 모두 "Codespace를 개발하는" 방법으로 나란히 설명하며, 접속 수단에 따라 컨테이너 내부의 기본 환경변수 집합이 달라진다는 근거는 확인되지 않습니다. 즉 `CODESPACES`·`CODESPACE_NAME` 등 표의 값은 접속 클라이언트와 무관하게 컨테이너 자체의 속성입니다. 다만 SSH·VS Code 원격 세션이 각기 별도의 셸 프로세스를 만들므로, 세션별로 사용자가 추가한 셸 설정(dotfiles 등)이 다르게 로드될 여지는 있습니다.

사용자가 기본값 이상으로 환경변수를 주입하는 공식 경로는 `devcontainer.json`의 두 속성입니다.

- **`containerEnv`**: 컨테이너 생성 시점에 컨테이너 전체에 적용되는 변수를 설정합니다. Dev Container 사양상 이미지·Dockerfile 빌드 단계에서부터 유효한 값이며, 컨테이너에 붙는 모든 프로세스(어떤 클라이언트로 접속하든)에 공통으로 반영됩니다.
- **`remoteEnv`**: VS Code(및 그 하위 터미널·태스크·디버그 프로세스) 같은 개발 도구 쪽에서만 설정·override되는 변수입니다. 컨테이너 전체가 아니라 그 도구가 띄우는 프로세스 범위에 한정되며, `${containerEnv:VAR}` 구문으로 컨테이너 변수를 참조하거나 `${localEnv:VAR}` 구문으로 로컬 호스트 값을 끌어올 수 있습니다.

두 속성 모두 사용자(저장소 소유자 또는 Codespace를 만드는 개인)가 명시적으로 `devcontainer.json`에 작성해야 적용되는 **옵트인 설정값**입니다. 이 문서 상단 표의 기본 변수들과 달리, `containerEnv`/`remoteEnv`로 주입된 값은 저장소마다 다르므로 `runby`가 범용 감지 규칙으로 의존할 수 없습니다.

## 다른 축에 미치는 영향

Codespace는 리눅스 컨테이너이므로 그 안에서 실행되는 터미널 에뮬레이터 신호(`docs/terminals/`)는 컨테이너에 실제로 붙은 클라이언트(브라우저 내장 xterm.js 터미널, VS Code 통합 터미널 등)의 것이며, 로컬 터미널의 값과는 무관하게 독립적으로 관측됩니다. 에이전트 축(`docs/agents/`)도 마찬가지로, 컨테이너 안에서 에이전트 CLI를 실행해야만 해당 마커가 나타나고 로컬에서 실행 중인 에이전트는 이 컨테이너에 아무 흔적도 남기지 않습니다. CI 축은 앞 절에서 다룬 대로 `CODESPACES=true`가 `GITHUB_ACTIONS`·`FORGEJO_ACTIONS` 어느 쪽과도 동시에 서지 않는다는 점에서, 세 축(터미널·에이전트·CI) 모두 Codespaces와 원칙적으로 독립적으로 병존할 수 있는 구조입니다.

## 실행 주체 감지에 관한 결론

GitHub Codespaces를 식별하는 가장 신뢰도 높은 마커는 `CODESPACES=true`입니다. 공식 문서가 "Codespace 안에서는 항상 `true`"라고 명시하는 존재 마커이며, `CODESPACE_NAME`을 보조로 사용하면 어떤 Codespace 인스턴스인지까지 특정할 수 있습니다. 이 두 값은 GitHub Actions의 `GITHUB_ACTIONS`/`GITHUB_RUN_ID`, Forgejo의 `FORGEJO_ACTIONS`/`FORGEJO_RUN_ID`와 이름 공간이 겹치지 않으므로 세 제품을 혼동할 위험이 구조적으로 낮습니다.

다만 `GITHUB_REPOSITORY`·`GITHUB_SERVER_URL`·`GITHUB_API_URL`·`GITHUB_TOKEN` 네 변수는 이름이 GitHub Actions와 겹치므로, `runby`가 이 네 변수만으로 제품을 판정해서는 안 되며 반드시 `CODESPACES` 또는 `GITHUB_ACTIONS`/`FORGEJO_ACTIONS` 같은 전용 존재 마커를 먼저 확인한 뒤 보강 정보로만 사용해야 합니다. `CI`는 Codespaces 공식 문서에 아예 등장하지 않으므로 이 컨텍스트에서는 판정 근거로 삼지 않는 편이 안전하며, 이 부재를 실제 런타임에서 재확인하는 것을 이 문서의 `runtime_test_required: true` 근거로 남깁니다.

마지막으로, 이 값들도 일반 환경변수이므로 컨테이너 안에서 만들어진 모든 자식 프로세스에 상속되고 사용자가 동일한 이름으로 직접 export하면 위조될 수 있습니다. `runby`는 이를 절대적 신뢰 경계가 아니라 "이 프로세스 트리가 Codespace 컨테이너 안에서 시작됐다"는 강한 정황 증거로 다뤄야 합니다.

## 공식 문서

- [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace)
- [Introduction to dev containers](https://docs.github.com/en/codespaces/setting-up-your-project-for-codespaces/adding-a-dev-container-configuration/introduction-to-dev-containers)
- [Using GitHub Codespaces in Visual Studio Code](https://docs.github.com/en/codespaces/developing-in-a-codespace/using-github-codespaces-in-visual-studio-code)
- [Using GitHub Codespaces with GitHub CLI](https://docs.github.com/en/codespaces/developing-in-a-codespace/using-github-codespaces-with-github-cli)
- [Dev Container Specification — devcontainer.json reference (containerEnv / remoteEnv)](https://containers.dev/implementors/json_reference/)
