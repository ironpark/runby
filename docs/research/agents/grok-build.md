---
title: Grok Build
slug: grok-build
research_date: 2026-09-01
open_source: false
repository: null
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 문서가 플러그인 훅 주입 사실만 밝히고 변수의 값 형식을 기술하지 않아 Extra 필드 확장 전에 실측이 필요함
---

# Grok Build

## 요약

Grok Build(xAI)의 감지 가능 범위는 **플러그인 훅에 한정**됩니다. 공식 문서는 "플러그인 훅은 추가로 `GROK_PLUGIN_ROOT`와 `GROK_PLUGIN_DATA`를 환경에서 받는다"고 명시합니다. 이는 훅이 실행되었다는 것, 즉 Grok Build가 그 프로세스를 실행했다는 것의 직접 증거이므로 `definite`입니다.

**중요한 한계:** 이 마커는 훅에만 주어지며, Grok Build가 실행하는 일반 셸 명령에는 주어지지 않습니다. 공식 설정 레퍼런스에는 모든 자식 프로세스에 공통으로 주입되는 범용 마커(`GROK=1` 같은)가 문서화되어 있지 않습니다. 따라서 **훅을 한 번도 실행하지 않은 Grok Build 세션은 감지되지 않습니다.** Amp의 Orb 한정 감지, Antigravity 2.0의 sidecar 한정 감지와 같은 성격의 부분 커버리지입니다.

## 환경변수

| 변수 | 값 | 분류 | 의미 | 판정 | 근거 |
|---|---|---|---|---|---|
| `GROK_PLUGIN_ROOT` | 미문서화 (경로로 추정) | 실행 식별 | 플러그인 훅에 제공되는 플러그인 루트 | **적합 (범위 한정)** — 훅 실행의 직접 증거 | [Skills, Plugins & Marketplaces](https://docs.x.ai/build/features/skills-plugins-marketplaces) |
| `GROK_PLUGIN_DATA` | 미문서화 | 실행 식별 | 플러그인 훅에 제공되는 데이터 위치 | **적합 (범위 한정)** — 위와 동일 | [Skills, Plugins & Marketplaces](https://docs.x.ai/build/features/skills-plugins-marketplaces) |

공식 문서가 값의 형식을 기술하지 않으므로 `Extra`나 `Paths`로 값을 노출하지 않습니다. 값이 경로라면 홈 디렉터리가 드러날 수 있어, 실측으로 형식이 확인되기 전까지는 이름만 근거로 씁니다.

## 실행 파일

`grok`.
