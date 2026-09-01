---
title: Charm Crush
slug: crush
research_date: 2026-09-02
open_source: true
repository: https://github.com/charmbracelet/crush
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 소스의 공통 셸 생성 경로가 모든 자식 셸에 실행 마커를 붙인다고 명시함
---

# Charm Crush

## 요약

Charm Crush는 여러 모델 공급자를 연결하는 터미널 코딩 에이전트 하네스다. 공식 소스의 `CrushEnvMarkers()`는 Crush가 실행하는 **모든 셸**에 `CRUSH=1`, `AGENT=crush`, `AI_AGENT=crush`를 무조건 붙인다고 설명한다. 대화형 bash 셸과 hook runner 양쪽에서 같은 목록을 쓰므로 `CRUSH=1`을 definite 실행 마커로 채택한다.

## 환경변수

| 변수 | 값 | 분류 | 판정 |
|---|---|---|---|
| `CRUSH` | `1` | 실행 식별 | **적합 — definite**. Crush가 spawn하는 모든 셸에 무조건 설정 |
| `AGENT` | `crush` | 보조 식별 | `CRUSH=1`이 이미 성립할 때 값이 정확히 `crush`인 경우에만 evidence로 기록 |
| `AI_AGENT` | `crush` | 보조 식별 | `CRUSH=1`이 이미 성립할 때 값이 정확히 `crush`인 경우에만 evidence로 기록 |

`AGENT`는 제품 벤더를 특정하지 않는 일반 이름이다. 따라서 `AGENT=1`이나 `AGENT=crush`만으로는 Crush를 판정하지 않는다. OpenCode 조사 문서가 지적하듯 Goose도 `AGENT`를 설정하므로, 이 드라이버는 `AGENT`와 `AI_AGENT`를 먼저 엿보고 값이 Crush를 지목할 때만 evidence에 추가한다. 환경 evidence에는 변수 이름만 노출하며 값은 노출하지 않는다.

## 실행 파일과 감지 규칙

- 실행 파일은 `crush`로 등록해 환경 마커와 함께 살아 있는 조상 프로세스 확증에 사용한다.
- `CRUSH=1`이 있으면 Crush로 판정하고, `AGENT`·`AI_AGENT`는 값이 `crush`일 때만 보조 evidence로 기록한다.
- `CRUSH` 없이 일반 변수만 있는 경우는 감지하지 않는다.

## 공식 소스

- [`CrushEnvMarkers()` — 모든 spawn 셸에 넣는 세 변수](https://github.com/charmbracelet/crush/blob/main/internal/shell/shell.go#L878-L901)
- [`NewShell()` — 셸 환경에 마커를 추가](https://github.com/charmbracelet/crush/blob/main/internal/shell/shell.go#L950-L985)
