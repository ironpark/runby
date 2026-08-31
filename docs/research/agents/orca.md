---
title: Orca (Stably AI)
slug: orca
research_date: 2026-08-31
open_source: true
repository: https://github.com/stablyai/orca
product_type: agent_orchestrator
executes_agents:
  - claude-code
  - codex
  - cursor-agent
  - github-copilot-cli
  - opencode
  - antigravity-cli
  - amp
  - custom-cli-agent
runtime_test_required: true
runtime_test_reason: 소스에서 Orca pane 식별 변수의 PTY 주입과 WSL 전달은 확인했지만, 일반 shell pane과 agent pane의 차이 및 local·WSL·SSH·paired server에서의 보존 범위를 실제 실행으로 구분해야 함
---

# Orca (Stably AI)

이 문서의 Orca는 [`stablyai/orca`](https://github.com/stablyai/orca), 공식 사이트 [`onorca.dev`](https://www.onorca.dev/)의 제품을 뜻합니다. 같은 이름의 별도 오픈소스 agent orchestrator가 여러 개 있으므로 저장소 소유자를 식별자에 포함합니다.

Orca는 자체 agent harness가 아니라 여러 CLI agent를 terminal과 git worktree 안에서 실행·관리하는 오픈소스 데스크톱 IDE/ADE이자 orchestrator입니다. 공식 문서는 agent session을 “하나의 worktree에 있는 하나의 terminal에서 실행되는 하나의 CLI agent”로 정의하고, Orca가 CLI를 spawn한 뒤 OSC title과 agent hook으로 상태를 추적한다고 설명합니다. 따라서 Orca 계층과 실제 하위 harness 계층을 동시에 보존해야 합니다. 예를 들어 Orca가 Codex를 실행하면 `orca (agent_orchestrator)`와 `codex (agent_harness)`가 함께 검출되는 것이 올바른 결과입니다.

## Orca pane 및 worktree 식별 변수

Orca 공식 사용자 문서에는 아래 환경변수 계약이 별도 표로 정리되어 있지 않습니다. 변수명·전달 경계·용도는 공식 오픈소스의 PTY와 WSL 구현에서 확인했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ORCA_PANE_KEY` | pane 좌표 문자열 | 실행 식별/상태 | hook 상태를 특정 Orca pane에 연결 | 보조 — Orca가 관리하는 pane이라는 강한 신호지만 일반 shell pane에도 적용될 수 있어 agent 실행을 단독 확정하지 않음 | [공식 소스: remote pane env 처리](https://github.com/stablyai/orca/blob/main/src/main/ipc/pty.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_TAB_ID` | UUID 형식 문자열 | 상태/컨텍스트 | pane이 속한 Orca tab 식별 | 보조 — Orca UI 위치를 식별하지만 agent 종류나 상태는 나타내지 않음 | [공식 소스: remote pane env 처리](https://github.com/stablyai/orca/blob/main/src/main/ipc/pty.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_WORKTREE_ID` | worktree 식별 문자열 | 상태/컨텍스트 | pane과 agent hook을 Orca worktree에 연결 | 보조 — 어떤 격리 worktree에서 실행되는지 나타내지만 plain terminal에서도 존재할 수 있음 | [공식 소스: remote pane env 처리](https://github.com/stablyai/orca/blob/main/src/main/ipc/pty.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_TERMINAL_HANDLE` | `term_<id>` 형태의 문자열 | 실행 식별/상태 | Orca CLI·daemon이 terminal을 지칭하는 안정적인 handle | 보조 — Orca가 관리하는 terminal임을 나타내지만 agent 전용 신호는 아님 | [공식 소스: local PTY 상태 등록](https://github.com/stablyai/orca/blob/main/src/main/providers/local-pty-provider.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_USER_DATA_PATH` | 디렉터리 경로 | 상태/컨텍스트 | Orca 사용자 데이터 위치 | 보조 — Orca runtime 컨텍스트지만 경로 설정·상속 가능성이 있어 실행 판정의 단독 근거로 부적합 | [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |

`ORCA_PANE_KEY`와 `ORCA_TERMINAL_HANDLE` 중 하나가 있고 `ORCA_TAB_ID` 또는 `ORCA_WORKTREE_ID`가 함께 있으면 Orca가 호스팅하는 terminal/pane으로 판단할 수 있습니다. 그러나 이 조합만으로 사용자가 Orca terminal에 직접 입력한 명령인지, Orca가 선택 메뉴에서 실행한 agent가 호출한 명령인지는 구분할 수 없습니다. 실제 agent harness는 `CODEX_THREAD_ID`, `CLAUDECODE`, `CURSOR_AGENT` 등 해당 agent 고유 신호를 별도 탐지해야 합니다.

## 조건부 컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ORCA_CODEX_HOME` | 디렉터리 경로 | 설정/상태 | Orca가 선택한 Codex runtime home의 소유권을 표시 | 보조 — Orca가 관리하는 Codex 환경에서는 강한 신호지만 Codex에만 적용되고 system-default 또는 remote 구성에 따라 없을 수 있음 | [공식 소스: Codex home 처리](https://github.com/stablyai/orca/tree/main/src/main/codex), [공식 이슈의 소스 추적](https://github.com/stablyai/orca/issues/8612) |
| `ORCA_ROOT_PATH` | 절대 디렉터리 경로 | hook 컨텍스트 | 원본 repository root | 부적합 — worktree setup/archive script용 조건부 변수이며 일반 agent shell 전체의 마커가 아님 | [공식 소스: setup hook env](https://github.com/stablyai/orca/blob/main/src/main/setup-hook-env-vars.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_WORKTREE_PATH` | 절대 디렉터리 경로 | hook 컨텍스트 | 현재 worktree 경로 | 부적합 — setup/archive 실행 컨텍스트에 한정 | [공식 소스: setup hook env](https://github.com/stablyai/orca/blob/main/src/main/setup-hook-env-vars.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_WORKSPACE_NAME` | 표시 이름 문자열 | hook 컨텍스트 | setup hook에서 현재 workspace 이름 제공 | 부적합 — agent 실행 여부가 아닌 workspace setup 컨텍스트 | [공식 소스: setup hook env](https://github.com/stablyai/orca/blob/main/src/main/setup-hook-env-vars.ts), [공식 소스: WSL 전달](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_ORCHESTRATION_COMPATIBILITY_HOST_KIND` | 현재 `wsl` 문자열 | 원격/호스트 컨텍스트 | orchestration compatibility host 종류 | 보조 — WSL 경계에서만 설정되는 조건부 host 정보 | [공식 소스: WSL host stamp](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_ORCHESTRATION_COMPATIBILITY_HOST_ID` | host ID 문자열 | 원격/호스트 컨텍스트 | WSL 실행을 소유한 Orca host 식별 | 보조 — host 연결 정보이며 agent 종류는 나타내지 않음 | [공식 소스: WSL host stamp](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |
| `ORCA_ORCHESTRATION_COMPATIBILITY_HOST_INCARNATION` | WSL distro 문자열 | 원격/호스트 컨텍스트 | host 실행 incarnation/distro 식별 | 보조 — WSL 실행 위치를 보강 | [공식 소스: WSL host stamp](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts) |

hook transport, shell-ready wrapper, history, provider별 설정 overlay와 디버그 변수는 Orca 내부 배관 또는 특정 기능 설정이므로 실행 식별 표에 포함하지 않았습니다.

## 상태 감지의 한계

공식 문서상 Orca의 `working`, `needs you`, `done`, `blocked`, `idle` 상태는 환경변수가 아니라 agent hook과 terminal OSC title 이벤트를 Orca 자체가 수집해 계산한 결과입니다. 따라서 `runby`가 Orca 아래에서 실행될 때 환경변수만 읽어서는 Orca UI가 표시하는 현재 상태를 재구성할 수 없습니다.

또한 공식 소스는 remote agent hook이 꺼진 SSH 연결에서 pane 관련 환경변수를 제거하는 경로를 갖고 있고, WSL은 `WSLENV` allowlist를 통해 선택한 변수만 전달합니다. 그러므로 변수 부재는 Orca 외부 실행을 뜻하지 않습니다. local·WSL·SSH·paired Orca Server에 따라 감지 가능 범위가 달라집니다.

## 권장 판정

1. `ORCA_PANE_KEY` 또는 `ORCA_TERMINAL_HANDLE`이 있고 `ORCA_TAB_ID` 또는 `ORCA_WORKTREE_ID`가 함께 있으면 Orca host/orchestrator 계층을 `ConfidenceProbable`로 추가한다.
2. 하위 agent 전용 변수가 함께 있으면 해당 harness 계층을 별도로 추가하고 Orca보다 안쪽 계층으로 보존한다.
3. `ORCA_CODEX_HOME`은 Orca-managed Codex라는 보강 증거로만 사용하고 단독 판정하지 않는다.
4. setup hook 전용 변수와 WSL host 변수는 `Extra` 컨텍스트로만 보존한다.
5. Orca의 working/done 상태는 환경변수로 추정하지 않는다.

실행 검증에서는 최소한 다음 네 경우를 비교해야 합니다.

- Orca의 일반 terminal에서 사용자가 직접 실행
- Orca가 Codex·Claude Code를 agent tab으로 실행한 뒤 그 agent가 실행
- Windows host와 WSL worktree
- SSH worktree에서 remote agent hook을 켠 경우와 끈 경우

## 공식 문서와 소스

- [Orca 공식 문서](https://www.onorca.dev/docs)
- [Agents & sessions](https://www.onorca.dev/docs/model/agents-sessions)
- [Supported agents](https://www.onorca.dev/docs/agents/supported)
- [Terminal](https://www.onorca.dev/docs/terminal)
- [Orca 공식 소스](https://github.com/stablyai/orca)
- [PTY WSL 환경변수 전달 소스](https://github.com/stablyai/orca/blob/main/src/main/pty/wsl-orca-env.ts)
