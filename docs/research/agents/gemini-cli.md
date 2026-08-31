---
title: Gemini CLI
slug: gemini-cli
research_date: 2026-09-01
open_source: true
repository: https://github.com/google-gemini/gemini-cli
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 소스가 상수 이름부터 식별 목적을 명시하고 셸 실행·MCP 서버 양쪽 주입 지점을 직접 확인함
---

# Gemini CLI

## 요약

Gemini CLI는 자신이 실행하는 모든 셸 명령과 stdio MCP 서버 환경에 `GEMINI_CLI=1`을 주입합니다. 공식 소스에서 이 변수는 `GEMINI_CLI_IDENTIFICATION_ENV_VAR`라는 상수로 선언되어 있어, **식별 목적임이 제품 자신에 의해 명시**되어 있습니다. 사용자가 설정하는 값이 아니라 제품이 자식에게 심는 마커이므로 `definite`입니다.

`GEMINI_CLI_*` 접두사를 가진 다른 변수(`GEMINI_CLI_HOME`, `GEMINI_CLI_IDE_*`, `GEMINI_CLI_TRUST_WORKSPACE` 등)는 전부 CLI가 **읽는** 설정값이며 주입 마커가 아닙니다. 이름이 비슷하다는 이유로 근거에 넣어서는 안 됩니다.

## 환경변수

| 변수 | 값 | 분류 | 의미 | 판정 | 근거 |
|---|---|---|---|---|---|
| `GEMINI_CLI` | `1` | 실행 식별 | Gemini CLI가 실행한 셸 명령·MCP 서버에 주입 | **적합** — 제품이 식별 상수로 선언한 주입 마커 | [`shellExecutionService.ts` — `GEMINI_CLI_IDENTIFICATION_ENV_VAR`](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/services/shellExecutionService.ts) |
| `GEMINI_CLI_HOME` | 경로 문자열 | 설정 | CLI 홈 디렉터리 재지정 | 부적합 — 사용자가 실행 전에 설정 | [`paths.ts`](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/utils/paths.ts) |
| `GEMINI_CLI_IDE_*` | 포트·토큰·경로 | 설정 | IDE 컴패니언 연결 정보 | 부적합 — 연결 설정이며 `GEMINI_CLI_IDE_AUTH_TOKEN`은 자격증명 | [`ide-client.ts`](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/ide/ide-client.ts) |
| `GEMINI_CLI_TRUST_WORKSPACE` | 불리언 | 설정 | 워크스페이스 신뢰 여부 | 부적합 — 사용자 설정 | [`trust.ts`](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/utils/trust.ts) |

## 주입 범위

공식 소스의 `shellExecutionService.ts`는 자식 환경을 만들 때 정리된 부모 환경 위에 `GEMINI_CLI` 상수를 얹습니다. `mcp-client.ts`도 stdio MCP 서버를 띄울 때 같은 상수를 넣습니다. 즉 **셸 도구와 MCP 서버 양쪽**이 대상입니다.

## 실행 파일

`gemini`. 조상 체인에서 이 이름을 찾아 환경 판정을 확증합니다.
