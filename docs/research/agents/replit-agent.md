---
title: Replit Agent
slug: replit-agent
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 문서가 REPL_ID·REPLIT_*를 workspace/app 런타임 변수로 정의하며 agent 실행 마커로 정의하지 않음
---

# Replit Agent

## 요약

**드라이버 없음 — 축을 달리해 기록.** Replit Agent가 동작하는 Replit workspace에는 `REPL_ID`와 `REPLIT_*` 변수가 있을 수 있지만, 공식 문서는 이를 agent가 실행한 프로세스의 표지가 아니라 앱·workspace·배포 런타임의 기본 환경으로 설명한다. 이 변수는 사람이 같은 workspace에서 실행한 명령에도 존재할 수 있으므로 agent 축에서 구분력이 없다.

## 조사 경위

Replit의 project configuration 문서에서 `REPL_ID`는 앱 UUID이고, `REPLIT_*`는 도메인·사용자·배포 같은 workspace/app 정보를 나타낸다. Secrets 문서도 사용자가 만든 secret과 미리 정의된 앱 환경변수를 실행 환경에 제공하는 방식으로 설명한다. 이는 Replit Agent의 자식 프로세스에만 붙는 신호가 아니다.

| 변수군 | 공식 의미 | 판정 |
|---|---|---|
| `REPL_ID` | Repl/app UUID | 부적합 — workspace 식별자이지 agent 실행 식별자가 아님 |
| `REPLIT_DOMAINS`, `REPLIT_DEV_DOMAIN` 등 | 앱 도메인·배포 런타임 컨텍스트 | 부적합 — workspace 전체에 존재 가능 |
| 사용자 Secrets | 앱에 전달하는 비밀 설정 | 부적합 — 사용자 입력값 |

## 결론

`REPL_ID`·`REPLIT_*`는 remote/workspace 축에 가까운 환경 설명이며, Replit Agent가 요청했는지 사람이나 다른 도구가 요청했는지 구분하지 못한다. agent 전용 자식 프로세스 마커가 공식적으로 공개되기 전에는 Replit Agent 드라이버를 추가하지 않는다.

## 공식 문서

- [Replit project configuration — REPL_ID와 기본 변수](https://docs.replit.com/references/project-setup/configuration)
- [Replit Secrets와 앱 환경변수](https://docs.replit.com/core-concepts/project-editor/app-setup/secrets)
- [Replit MCP server — replId의 앱 식별 용도](https://docs.replit.com/platforms/mcp-server)
