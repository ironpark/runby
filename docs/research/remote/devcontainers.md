---
title: Dev Containers
slug: devcontainers
research_date: 2026-08-31
open_source: true
repository: https://github.com/devcontainers/spec
product_type: remote_environment
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 문서와 이슈 트래커만으로는 `REMOTE_CONTAINERS`/`REMOTE_CONTAINERS_IPC`/`REMOTE_CONTAINERS_SOCKETS`가 dotfiles 설치 등 라이프사이클의 어느 시점부터 실제로 존재하는지, 현재 VS Code 버전에서도 동일한 이름과 시점을 유지하는지 확정할 수 없음. 실행 검증 범위는 (1) VS Code Dev Containers 확장으로 연 컨테이너에서 `env` 스냅샷을 라이프사이클 스크립트별로 뜨는 것, (2) `devcontainer` CLI로 동일 `devcontainer.json`을 `up`/`exec`한 컨테이너의 `env`와 비교하는 것, (3) `docker exec`로 외부에서 진입한 셸의 `env`를 비교하는 것
---

# Dev Containers

Dev Containers는 Microsoft가 주도하고 여러 도구가 구현하는 [Development Containers 명세](https://containers.dev)와, 그 명세를 구현하는 VS Code Dev Containers 확장, 오픈소스 `devcontainer` CLI, GitHub Codespaces 등을 통칭합니다. 이 문서에서 가장 먼저 확인해야 할 사실은 **명세 자체가 컨테이너 안에서 실행 중임을 알리는 환경변수를 하나도 요구하지 않는다**는 점입니다. `REMOTE_CONTAINERS`류 변수는 VS Code 확장이 자체적으로 도입한 구현 세부사항이며, 같은 컨테이너를 `devcontainer` CLI나 다른 IDE로 열면 나타나지 않습니다. 아래에서 이 구분을 근거와 함께 정리합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `REMOTE_CONTAINERS` | 불리언 문자열 (`true`) | 실행 식별 | VS Code Remote-Containers/Dev Containers 확장이 컨테이너 안에서 실행 중임을 스크립트가 알 수 있도록 설정 | 보조 신호 — VS Code 확장 메인테이너가 명시적으로 구현했지만(0.143.0, 2020-09), Development Containers **명세에는 없는** VS Code 전용 값이라 `devcontainer` CLI나 다른 도구로 열면 설정되지 않음 | [microsoft/vscode-remote-release#3517](https://github.com/microsoft/vscode-remote-release/issues/3517) |
| `REMOTE_CONTAINERS_IPC` | 소켓 경로 문자열 | 상태·컨텍스트 | VS Code 서버와 확장 호스트 간 IPC 통신에 쓰이는 소켓 경로 | 부적합 — 공식 문서화된 계약이 아니라 커뮤니티가 관찰한 내부 구현 상세이며, 그 값 자체가 감지용으로 설계되지 않음 | [microsoft/vscode-remote-release#3517](https://github.com/microsoft/vscode-remote-release/issues/3517) |
| `REMOTE_CONTAINERS_SOCKETS` | 소켓 경로 목록 문자열 | 상태·컨텍스트 | VS Code가 포워딩하는 추가 소켓(예: SSH agent, X11 디스플레이) 경로 목록 | 부적합 — 위와 동일하게 비공식 내부 구현 상세 | [microsoft/vscode-remote-release#3517](https://github.com/microsoft/vscode-remote-release/issues/3517) |
| `VSCODE_REMOTE_CONTAINERS_SESSION` | 세션 식별 문자열 | 상태·컨텍스트 | 이슈 트래커에서 `REMOTE_CONTAINERS`가 아직 설정되지 않는 라이프사이클 초기 구간(예: dotfiles 설치)에 사용자가 대신 참조한 변수로 언급됨 | 부적합 — 공식 문서가 존재를 확인해 주지 않는, 커뮤니티가 발견한 비공식 값 | [microsoft/vscode-remote-release#3517](https://github.com/microsoft/vscode-remote-release/issues/3517) |
| `DEVCONTAINER` | 불리언 문자열 (`true`, 사용자 설정 시) | 실행 식별 | 여러 도구(VS Code, Codespaces, DevPod 등)를 아우르는 공통 감지 변수로 제안됨 | 부적합 — 어떤 공식 도구도 기본으로 설정하지 않음. VS Code 메인테이너가 "지금은 아니다, Codespaces와 조율이 필요하다"고 명시적으로 보류했고 이후 공식 채택 기록도 확인되지 않음. `remoteEnv`로 사용자가 직접 설정해야 하는 **관용구**일 뿐임 | [microsoft/vscode-remote-release#3517](https://github.com/microsoft/vscode-remote-release/issues/3517) |
| `CODESPACES` | 불리언 문자열 (`true`) | 실행 식별 | GitHub Codespaces가 코드스페이스 안에서 항상 설정하는 자체 마커. `REMOTE_CONTAINERS`와는 **별개**의 변수 | 적합 — GitHub 공식 문서가 "코드스페이스 안에 있는 동안 항상 `true`"라고 명시 | [Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace) |

## containerEnv와 remoteEnv

Development Containers 명세는 컨테이너에 값을 주입하는 경로를 두 가지로 나눕니다. 이 둘의 차이가 `runby`가 devcontainer 설정을 근거로 무언가를 감지하려 할 때 부딪히는 핵심 문제입니다.

**`containerEnv`** — 공식 정의는 다음과 같습니다.

> "A set of name-value pairs that sets or overrides environment variables for the container."

[VS Code 문서](https://code.visualstudio.com/remote/advancedcontainers/environment-variables)는 이를 "컨테이너 생성 시점에 컨테이너의 일부가 되어 전체 라이프사이클 동안 존재"한다고 설명합니다. `containerEnv`는 이미지 빌드/컨테이너 생성 단계에 반영되므로 **컨테이너 안에서 실행되는 모든 프로세스**가 상속합니다 — VS Code가 띄운 터미널이든, `docker exec`로 외부에서 들어온 셸이든, 컨테이너 안에서 도는 cron이나 데몬이든 동일합니다. 값을 바꾸려면 컨테이너를 다시 빌드해야 합니다.

**`remoteEnv`** — 공식 정의는 다음과 같습니다.

> "A set of name-value pairs that sets or overrides environment variables for the `devcontainer.json` supporting service / tool (or sub-processes like terminals) but not the container as a whole."

[containers.dev 구현체 문서](https://containers.dev/implementors/json_reference/)는 이를 "lifecycle 스크립트와 원격 에디터/IDE 서버 프로세스를 포함해 컨테이너 안에서 파생된 프로세스에 설정되는 원격 환경변수"라고 설명합니다. 즉 `remoteEnv`는 **연결을 맺은 도구(예: VS Code 서버 프로세스)가 자기 자신과 자신이 자식으로 띄우는 프로세스**에만 적용되는 값이며, 컨테이너 자체의 속성이 아닙니다. 리빌드 없이 값을 바꿀 수 있는 대신, 그 도구의 연결 범위를 벗어난 프로세스는 이 값을 보지 못합니다.

### 실무적 차이

- **VS Code가 띄운 터미널·태스크·디버그 세션** — `containerEnv`와 `remoteEnv` 모두 보입니다.
- **`docker exec`로 외부에서 진입한 셸** — `containerEnv`만 보입니다. `remoteEnv`는 VS Code 서버 연결에 묶인 값이라 이 셸에는 전달되지 않습니다.
- **컨테이너 안에서 자체적으로 도는 cron 잡·데몬** — 컨테이너 생성 시점에 시작된 프로세스라면 `containerEnv`는 보이지만, VS Code 세션이 나중에 붙였다 뗀 `remoteEnv` 값은 그 프로세스의 시작 시점 환경에 없었으므로 보이지 않습니다.
- 위 표의 `REMOTE_CONTAINERS=true`는 VS Code 확장이 내부적으로 이 `remoteEnv` 메커니즘으로 주입하는 값입니다. 실제로 원 이슈에서 사용자가 "dotfiles 설치 단계에서는 아직 설정되어 있지 않다"고 보고했고, 메인테이너가 이를 버그로 인정해 이후 버전에서 고쳤습니다 — 즉 `remoteEnv` 기반 값은 **같은 세션 안에서도 라이프사이클 단계에 따라 존재 여부가 달라질 수 있다**는 것이 공식 이슈 트래커에 기록된 사실입니다.

따라서 `containerEnv` 기반 마커는 "이 프로세스가 이 컨테이너 안에 있다"는 비교적 안정적인 근거가 될 수 있지만(다만 컨테이너 자체를 식별하는 것이지 devcontainer 도구를 식별하는 것은 아닙니다), `remoteEnv` 기반 마커(`REMOTE_CONTAINERS` 포함)는 "지금 이 요청을 만든 도구가 VS Code 서버 연결의 자식 프로세스다"라는 훨씬 좁고 시점에 민감한 근거일 뿐입니다.

## 컨테이너 자체는 환경변수로 감지되지 않습니다

Docker와 Podman은 **컨테이너 런타임을 알리는 환경변수를 아무것도 설정하지 않습니다.** 컨테이너 안에서 실행 중임을 감지하는 통상적인 방법은 전부 파일시스템 경로이지 환경변수가 아닙니다.

- Docker: `/.dockerenv` 파일 존재 여부
- Podman: `/run/.containerenv` 파일 존재 여부
- 범용: `/proc/1/cgroup` 내용, cgroup 계층 구조

이 세 가지 모두 `runby`가 읽는 대상인 "상속되는 환경변수"가 아니라 파일시스템 조회입니다. 따라서 **`runby`는 환경변수만으로는 자신이 컨테이너 안에 있다는 사실 자체를 감지할 수 없습니다.** VS Code Dev Containers 확장이나 GitHub Codespaces처럼 상위 도구가 스스로 `REMOTE_CONTAINERS`나 `CODESPACES` 같은 값을 명시적으로 광고해 줄 때만 그 도구의 존재를 알 수 있고, 이는 "컨테이너 감지"가 아니라 "그 컨테이너를 만든 특정 도구의 감지"입니다. plain `docker run`이나 `podman run`으로 띄운 컨테이너, 혹은 devcontainer 명세를 따르되 이런 광고를 하지 않는 도구(JetBrains Gateway, DevPod 등— 위 표의 `loft-sh/devpod#891`이 이 공백을 그대로 보여줍니다)는 환경변수 축으로는 원천적으로 보이지 않습니다.

이는 메워야 할 구멍이 아니라 이 라이브러리가 환경변수만 읽는다는 설계 범위에서 나오는 **문서화된 한계**입니다.

`HOSTNAME`이 짧은 16진수 문자열(컨테이너 ID 축약형)로 보이는 경우가 흔해 "컨테이너 안인 것 같다"는 추측의 근거로 종종 쓰이지만, 이는 어떤 공식 문서도 계약하지 않은 **휴리스틱**입니다. `HOSTNAME`은 `docker run --hostname`으로 임의 설정 가능하고, 짧은 16진수 문자열은 컨테이너가 아닌 다른 여러 환경(예: 일부 클라우드 VM, CI 러너)에서도 나타나며, devcontainer 명세나 VS Code 문서 어디에도 `HOSTNAME`의 형태를 실행 컨텍스트 감지에 쓰라는 규정이 없습니다. `runby`는 이를 판정 근거로 사용하지 않습니다.

## 다른 축에 미치는 영향

`runby`는 에이전트·CI·터미널 세 축을 보고합니다. devcontainer 경계를 넘나드는 프로세스에서 이 축들이 어떻게 되는지 짚어야 합니다.

- **호스트에서 컨테이너로 exec하는 에이전트** — 호스트에서 실행 중인 에이전트(예: Claude Code, Cursor Agent)가 `docker exec`로 컨테이너 안에 셸을 띄우는 경우, `docker exec`는 **호스트 프로세스의 환경을 컨테이너 프로세스에 상속시키지 않습니다.** 컨테이너 안 프로세스는 그 컨테이너의 `containerEnv`(또는 이미지에 굽힌 값)에서 시작하며, 호스트에서 설정된 에이전트 마커(`CLAUDECODE=1`, `CURSOR_AGENT` 등)는 `docker exec -e VAR=value`로 명시적으로 전달하지 않는 한 **건너가지 않습니다.**
- **VS Code가 원격 세션에 전달하는 값** — VS Code Dev Containers 확장은 예외적으로 이 경계를 스스로 메웁니다. 확장이 컨테이너에 붙일 때 `remoteEnv`(및 자체 IPC 메커니즘)로 자신이 필요로 하는 값을 그 연결이 띄우는 프로세스에만 주입합니다. 이는 일반적인 `docker exec`가 아니라 VS Code 서버 프로세스 자체가 하는 일이므로, VS Code로 연 세션의 터미널에서는 호스트 쪽 VS Code 관련 마커 일부가 보일 수 있지만, 같은 컨테이너에 별도로 `docker exec`로 들어간 셸에는 보이지 않습니다.
- **결론** — devcontainer 경계는 기본적으로 환경변수 상속을 차단하는 벽입니다. 이 벽을 넘는 값은 (a) 이미지/`containerEnv`에 미리 구워둔 값이거나 (b) VS Code처럼 연결 프로토콜이 명시적으로 전달하는 값뿐입니다. `runby`가 호스트에서 관찰한 에이전트·CI 마커가 컨테이너 안 프로세스에서도 그대로 보일 것이라고 가정해서는 안 되며, 반대로 컨테이너 안에서 관찰한 devcontainer 마커가 호스트 프로세스에 영향을 주는 일도 없습니다.

## 실행 주체 감지에 관한 결론

세 가지를 분명히 정리합니다.

1. **명세는 마커를 요구하지 않습니다.** Development Containers 명세(`devcontainer.json` 레퍼런스, `devcontainer-reference.md`)는 `containerEnv`/`remoteEnv`라는 주입 메커니즘만 정의할 뿐, 컨테이너나 실행 도구를 식별하는 환경변수를 하나도 규정하지 않습니다.
2. **`REMOTE_CONTAINERS`는 VS Code 확장의 구현 세부사항입니다.** 원 이슈(`vscode-remote-release#3517`)에서 VS Code 메인테이너가 직접 도입했고(v0.143.0, 2020-09), `remoteEnv` 메커니즘으로 설정됩니다. `devcontainer` CLI(오픈소스 레퍼런스 구현)나 JetBrains Gateway, DevPod 같은 다른 도구는 이 값을 설정하지 않습니다. `DEVCONTAINER=true`는 같은 이슈에서 여러 도구를 아우르는 공통 규약으로 제안됐지만 메인테이너가 "지금은 채택하지 않는다"고 명시적으로 보류했고, 이후 공식 채택 기록도 확인되지 않았습니다 — 즉 사실상 사용자가 `remoteEnv`에 직접 넣어야 하는 관용구이지 어느 도구도 기본 제공하는 값이 아닙니다. `CODESPACES=true`만이 GitHub 공식 문서가 보증하는, Codespaces 자체의 별도 마커입니다.
3. **컨테이너 자체는 감지 대상이 아닙니다.** Docker/Podman 어느 쪽도 환경변수로 컨테이너 런타임을 알리지 않으며, 공식적인 감지 수단(`/.dockerenv`, `/run/.containerenv`, cgroup)은 전부 파일시스템 조회입니다. `runby`는 환경변수만 읽으므로, **devcontainer 도구가 스스로 광고하지 않는 한 컨테이너 실행 여부를 전혀 판정할 수 없습니다.** 이는 라이브러리의 명시적 한계이며, `HOSTNAME` 같은 값을 대체 휴리스틱으로 쓰는 것도 근거가 없어 채택하지 않습니다.

`runby`가 devcontainer 신호를 다룬다면, 그것은 "컨테이너 안에 있다"의 증거가 아니라 "VS Code Dev Containers 확장(또는 GitHub Codespaces)이 이 프로세스 트리를 열었다"는 훨씬 좁은 주장으로만 취급해야 하며, `devcontainer` CLI로 직접 열거나 다른 IDE로 연 devcontainer는 환경변수 축에서 조용히 보이지 않는다는 점을 명시해야 합니다.

## 공식 문서

- [Development Containers Specification — Overview](https://containers.dev)
- [Dev Container metadata reference (`containerEnv`, `remoteEnv`)](https://containers.dev/implementors/json_reference/)
- [devcontainers/spec — `devcontainer-reference.md`](https://github.com/devcontainers/spec/blob/main/docs/specs/devcontainer-reference.md)
- [devcontainers/spec — `devcontainerjson-reference.md`](https://github.com/devcontainers/spec/blob/main/docs/specs/devcontainerjson-reference.md)
- [devcontainers/cli (오픈소스 레퍼런스 구현)](https://github.com/devcontainers/cli)
- [VS Code — Environment variables in Dev Containers](https://code.visualstudio.com/remote/advancedcontainers/environment-variables)
- [VS Code — Developing inside a Container](https://code.visualstudio.com/docs/devcontainers/containers)
- [microsoft/vscode-remote-release#3517 — `REMOTE_CONTAINERS`/`DEVCONTAINER` 도입 논의 및 구현 기록](https://github.com/microsoft/vscode-remote-release/issues/3517)
- [GitHub Codespaces — Default environment variables for your codespace](https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace)
- [loft-sh/devpod#891 — DevPod에 devcontainer 감지 변수가 없다는 이슈](https://github.com/loft-sh/devpod/issues/891)
