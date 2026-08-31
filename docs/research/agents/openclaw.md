---
title: OpenClaw
slug: openclaw
research_date: 2026-09-01
open_source: true
repository: https://github.com/openclaw/openclaw
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 문서가 주입 사실과 런타임별 값을 모두 명시하고 저장소 문서·CHANGELOG가 교차 확인됨
---

# OpenClaw

## 요약

OpenClaw는 자신이 실행하는 셸 명령 환경에 `OPENCLAW_SHELL`을 주입하며, **값이 어느 런타임에서 실행되었는지를 담습니다.** 공식 문서는 exec 도구가 PTY·샌드박스 실행을 포함해 `OPENCLAW_SHELL=exec`을 설정한다고 명시하고, TUI 문서는 로컬 셸 명령이 `OPENCLAW_SHELL=tui-local`을 받는다고 적고 있습니다.

값이 진입점을 담으므로 `runby`는 이를 버리지 않고 `Entrypoint`로 보존합니다. 단순히 "설정되었는가"만 보는 것보다 많은 정보를 줍니다.

## 환경변수

| 변수 | 값 | 분류 | 의미 | 판정 | 근거 |
|---|---|---|---|---|---|
| `OPENCLAW_SHELL` | `exec`, `acp`, `acp-client`, `tui-local` | 실행 식별 | OpenClaw가 실행한 셸의 런타임 종류 | **적합** — 셸 규칙이 OpenClaw 컨텍스트를 감지하라는 명시적 목적의 주입 마커 | [exec 도구 문서](https://github.com/openclaw/openclaw/blob/main/docs/tools/exec.md), [TUI 문서](https://github.com/openclaw/openclaw/blob/main/docs/web/tui.md) |
| `OPENCLAW_SHELL_ENV_TIMEOUT_MS` | 정수 | 설정 | 셸 환경 수집 타임아웃 | 부적합 — 사용자 설정이며 이름만 유사 | [`.env.example`](https://github.com/openclaw/openclaw/blob/main/.env.example) |

## 주입 범위

CHANGELOG는 이 마커가 `exec`·`acp`·`acp-client`·`tui-local` 네 런타임에 일관되게 설정되도록 도입되었다고 기록합니다. 목적이 "셸 시작·설정 규칙이 OpenClaw 컨텍스트를 겨냥할 수 있게" 하는 것이므로, 감지 용도가 제품 측 의도와 일치합니다.

## 실행 파일

`openclaw`.

## 주의

`OPENCLAW_SHELL_ENV_TIMEOUT_MS`는 접두사가 같지만 사용자 설정입니다. 접두사 매칭이 아니라 **정확한 이름 비교**로만 판정해야 합니다.
