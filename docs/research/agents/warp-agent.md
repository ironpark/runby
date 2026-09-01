---
title: Warp Agent
slug: warp-agent
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: Warp 클라이언트와 Agent Mode가 비공개이고 공식 문서에 agent 명령용 실행 마커가 없음
---

# Warp Agent

## 요약

**드라이버 없음 — 음성 결과.** Warp는 Agent Mode와 외부 코딩 에이전트 실행을 제공하지만, Agent Mode가 요청한 명령과 사용자가 Warp 터미널에 직접 입력한 명령을 구분하는 공식 환경변수는 확인되지 않았다. Warp 터미널 자체의 `TERM_PROGRAM=WarpTerminal`은 이미 terminal 축에서 처리한다.

## 조사 경위

Warp의 Terminal/Agent modes, Warp Drive environment variables, Cloud Agents environments 문서를 확인했다. Warp Drive 변수는 사용자가 정의해 명령에 주입하는 값이고, cloud agent 환경 변수도 환경 이미지·런타임을 구성하는 값이다. `WARP_API_KEY` 등 Warp Agent CLI 설정값 역시 인증·동작 입력이지 하위 셸에 실행 주체를 표시하는 출력값이 아니다.

| 신호 | 확인 내용 | 판정 |
|---|---|---|
| `TERM_PROGRAM=WarpTerminal` | Warp가 만든 터미널을 식별하는 공식 신호 | **agent 축 부적합** — 사람 입력과 Agent Mode를 구분하지 못함; terminal 축에서 감지 |
| `WARP_API_KEY` 등 | Agent CLI 인증·설정 입력 | 부적합 — 실행 마커가 아님 |
| `WARP_*` Agent Mode marker | 공식 문서에서 이름·주입 범위 확인 안 됨 | **보류하지 않고 제외** — 비공개 구현 추측 금지 |

## 결론

Warp Agent의 실행 마커가 공식적으로 공개되기 전에는 agent 드라이버를 추가하지 않는다. `TERM_PROGRAM=WarpTerminal`은 `docs/research/terminals/warp.md`에 기록된 대로 터미널 호스트 신호이며, 이를 agent로 승격하면 사용자가 직접 실행한 명령까지 오탐한다.

## 공식 문서·기존 조사

- [Terminal and Agent modes](https://docs.warp.dev/agents/local-agents/interacting-with-agents/terminal-and-agent-modes/)
- [Warp Drive environment variables](https://docs.warp.dev/knowledge-and-collaboration/warp-drive/environment-variables)
- [Cloud Agent environments](https://docs.warp.dev/agent-platform/cloud-agents/environments)
- [Warp 터미널 조사 — 에이전트 기능과 TERM_PROGRAM](../../research/terminals/warp.md#warp의-에이전트-기능)
