---
title: OpenHands
slug: openhands
research_date: 2026-09-02
open_source: true
repository: https://github.com/OpenHands/software-agent-sdk
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: SDK가 terminal과 hook subprocess 환경에 AI_AGENT 기본값을 직접 설정하는 경로를 공개함
---

# OpenHands

## 요약

OpenHands SDK는 여러 모델 공급자를 연결하는 에이전트 하네스·도구 SDK다. `sanitized_env()`는 비어 있는 `AI_AGENT`를 `openhands`로 채우고, terminal 도구는 이 정리된 환경을 bash/PTY 세션에 사용한다. 그러므로 `AI_AGENT=openhands`의 **값을 대소문자 무시로 정확히 매칭**하면 OpenHands가 실행한 프로세스를 definite로 판정할 수 있다.

`AI_AGENT`는 다른 제품도 사용할 수 있는 표준화 후보 변수이므로, 알 수 없는 값을 generic agent로 처리하지 않는다. 이 드라이버는 OpenHands라는 제품명이 정확히 들어온 경우에만 기록한다. Claude Code의 `AI_AGENT=claude-code*` 규칙과도 값 범위가 겹치지 않는다.

## 환경변수

| 변수 | 값/형태 | 범위 | 판정 |
|---|---|---|---|
| `AI_AGENT` | `openhands` | terminal·hook subprocess | **적합 — definite**. OpenHands 환경 헬퍼가 비어 있을 때 제품명을 설정 |
| `OPENHANDS_SESSION_ID` | 세션 식별자 | hook subprocess | `SessionID`로 수집. hook 전용이므로 단독으로는 감지 근거가 아님 |
| `OPENHANDS_PROJECT_DIR` | 작업 디렉터리 | hook subprocess | `Extra["openhands.project_dir"]`로 수집 |
| `OPENHANDS_EVENT_TYPE` | hook 이벤트 | hook subprocess | `Extra["openhands.event_type"]`로 수집 |
| `OPENHANDS_TOOL_NAME` | hook을 부른 도구명 | hook subprocess, 선택적 | `Extra["openhands.tool_name"]`로 수집 |

`OPENHANDS_*` 네 변수는 hook 실행기에만 추가된다. 따라서 이 드라이버는 `AI_AGENT`로 먼저 판정한 뒤, 존재하는 hook 컨텍스트를 `SessionID`와 `Extra`로만 보탠다. hook 변수가 없는 일반 terminal 자식도 `AI_AGENT=openhands`만으로 정상 감지된다.

## 실행 파일과 감지 규칙

- 실행 파일은 `openhands`로 등록해 조상 프로세스 확증에 사용한다.
- `AI_AGENT`가 공백을 제거한 뒤 `openhands`와 대소문자 무시로 같을 때만 OpenHands로 판정한다.
- `OPENHANDS_SESSION_ID`는 `SessionID`로, 나머지 hook 변수는 `Extra`로 수집한다.
- `AI_AGENT=claude-code...`나 알 수 없는 `AI_AGENT` 값은 OpenHands로 보고하지 않는다.

## 공식 소스

- [`sanitized_env()` — AI_AGENT 기본값](https://github.com/OpenHands/software-agent-sdk/blob/main/openhands-sdk/openhands/sdk/utils/command.py#L31-L70)
- [`build_terminal_env()` — terminal 환경에 sanitized_env 적용](https://github.com/OpenHands/software-agent-sdk/blob/main/openhands-tools/openhands/tools/terminal/env.py#L32-L38)
- [`subprocess_terminal.py` — PTY bash에 환경 전달](https://github.com/OpenHands/software-agent-sdk/blob/main/openhands-tools/openhands/tools/terminal/terminal/subprocess_terminal.py#L144-L168)
- [`HookExecutor.execute()` — hook 전용 OPENHANDS_* 변수](https://github.com/OpenHands/software-agent-sdk/blob/main/openhands-sdk/openhands/sdk/hooks/executor.py#L460-L511)
