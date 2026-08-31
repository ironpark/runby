---
title: Cursor Agent
slug: cursor-agent
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_harness
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 문서가 terminal command에 CURSOR_AGENT를 주입한다고 명시함
---

# Cursor Agent

## 요약

Cursor 공식 문서에서 확인되는 `CURSOR_AGENT`는 Cursor Agent가 실행한 터미널 명령의 환경에 주입되는 실행 식별 신호이므로 이 패키지의 주 감지 키로 적합하다.

Cursor 공식 문서에는 세션 ID, 부모 Agent PID, 실행 모드 또는 샌드박스 상태를 자식 프로세스에 전달하는 환경변수가 별도로 공개되어 있지 않다.

## 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CURSOR_AGENT` | 비어 있지 않은 문자열 | 실행 식별 | Cursor Agent가 실행한 셸에서 프롬프트·테마 초기화를 다르게 처리하도록 제공되는 마커 | **적합** — 공식 문서가 Agent 실행 중인지 감지하는 용도라고 명시한다. 값 자체보다 존재 여부를 검사한다. | [Cursor Terminal](https://docs.cursor.com/en/agent/terminal) |

## 권장 감지 규칙

1. `CURSOR_AGENT`가 존재하고 빈 문자열이 아니면 Cursor Agent가 현재 프로세스의 실행 환경을 만든 것으로 판정한다.
2. 공개된 세션 식별·상태 변수가 없으므로 환경변수만으로 실행 중인 Cursor 세션의 세부 상태나 생존 여부는 판정하지 않는다.

## 공식 문서

- [Cursor CLI 개요](https://docs.cursor.com/en/cli/overview)
- [Cursor Agent 터미널](https://docs.cursor.com/en/agent/terminal)
- [Cursor CLI 매개변수](https://docs.cursor.com/en/cli/reference/parameters)
