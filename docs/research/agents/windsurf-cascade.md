---
title: Windsurf Cascade
slug: windsurf-cascade
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: 핵심 확장이 비공개이고 공식 Cascade·터미널 문서에 자식 프로세스 실행 마커가 없음
---

# Windsurf Cascade

## 요약

**드라이버 없음 — 음성 결과.** Windsurf Cascade는 IDE 안에서 terminal tool을 사용하는 agent 기능이지만, 공식 문서에는 Cascade가 실행한 명령과 사용자가 Windsurf 터미널에 직접 입력한 명령을 구분하는 환경변수가 없다. 제품 구현이 비공개이므로 내부 변수 이름을 추측해 채택하지 않는다.

## 조사 경위

Cascade 공식 문서는 agent와 도구 사용을 설명하고, Windsurf terminal 문서는 전용 터미널과 셸 초기화 환경을 설명한다. 확인된 환경은 `.zshrc` 등 사용자의 셸 설정과 일반 terminal 컨텍스트뿐이며, `WINDSURF_*` 또는 Cascade 전용 child execution marker의 이름·값·주입 범위는 공개되지 않았다.

| 조사 대상 | 확인 내용 | 판정 |
|---|---|---|
| Cascade | IDE agent 기능과 도구 호출 설명 | child 실행 marker 없음 |
| Windsurf terminal | 셸·터미널 사용 설명 | 터미널 소유는 agent 요청 증거가 아님 |
| `WINDSURF_*` 후보 | 공식 문서에서 실행 marker 계약 확인 안 됨 | **부적합** — 비공개 내부 구현 추측 금지 |

## 결론

Windsurf Cascade는 공식적으로 검증된 env marker가 없으므로 감지하지 않는다. 향후 자식 프로세스에 주입되는 전용 변수와 그 범위가 공식 문서 또는 공개 소스로 확인되면 재검토한다.

## 공식 문서

- [Cascade](https://docs.windsurf.com/windsurf/cascade/cascade)
- [Windsurf terminal](https://docs.windsurf.com/windsurf/terminal)
