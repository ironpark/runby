---
title: Claude Cowork
slug: cowork
research_date: 2026-09-01
open_source: false
repository: null
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: true
runtime_test_reason: 유일한 후보 변수가 공식 문서에 없고 제품이 폐쇄소스라, 실제 Cowork 세션에서 변수 존재와 값을 실측하는 것 외에 확인 경로가 없음
---

# Claude Cowork

## 요약

**드라이버 없음 — 조사 보류.** Claude Cowork는 `runby`의 다중 레이어 모델에 가장 잘 맞는 후보였으나, 현재 조사 기준을 충족하는 근거가 없습니다.

Cowork는 Claude Code와 같은 아키텍처를 터미널 없는 지식 작업에 적용한 제품입니다. Claude Code 위에 얹히는 구조이므로, 감지된다면 `runby`는 `cowork > claude-code`처럼 **두 레이어를 모두** 보고할 수 있습니다. `vercel/detect-agent`는 결과가 하나뿐인 구조라 `COWORK` 규칙을 `CLAUDE` 앞에 놓아 하나만 보고합니다.

## 왜 보류인가

`vercel/detect-agent`가 쓰는 `CLAUDE_CODE_IS_COWORK`는 **Anthropic 공식 환경변수 문서에 없습니다.** 공식 문서는 100개가 넘는 변수를 열거하면서 이 변수를 포함하지 않습니다. 제3자 정리 문서는 이를 내부·비공개(internal/hidden) 변수로 분류하며, 용도도 감지가 아니라 Cowork 전용 flush 동작 제어라고 기술합니다.

Claude Code는 폐쇄소스이므로 공식 소스로 교차 확인할 경로도 없습니다. 즉 이 저장소의 기준 — **제작사 공식 문서 또는 제작사가 공개한 공식 소스** — 어느 쪽도 만족하지 못합니다.

내부 변수는 예고 없이 사라지거나 의미가 바뀔 수 있고, 감지 목적으로 제공된 것이 아니므로 계약이 없습니다. `CLAUDECODE`가 공식 문서에서 "Claude Code가 생성한 자식 프로세스에 `1`로 설정된다"고 명시적으로 계약된 것과 대비됩니다.

## 환경변수

| 변수 | 값 | 분류 | 판정 | 근거 |
|---|---|---|---|---|
| `CLAUDE_CODE_IS_COWORK` | 미확인 | 내부 동작 제어 | **보류** — 공식 문서에 없고 감지 목적으로 제공된 계약이 아님 | 공식 [Environment variables](https://code.claude.com/docs/en/env-vars) 미수록 |
| `CLAUDECODE` | `1` | 실행 식별 | 적합 — [`claude-code.md`](claude-code.md) 참조 | [Environment variables](https://code.claude.com/docs/en/env-vars) |

## 재검토 조건

다음 중 하나가 충족되면 `cowork` 레이어를 추가합니다.

1. Anthropic 공식 문서가 Cowork 실행 마커를 문서화한다.
2. 실제 Cowork 세션에서 변수의 존재·값·주입 범위를 실측하고, 그 실측을 근거로 `probable` 신뢰도로 등록한다.

현재 Cowork에서 실행된 프로세스는 `claude-code` 레이어로 보고됩니다 — 틀린 답은 아니며, 다만 한 단계 덜 구체적입니다.
