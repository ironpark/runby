---
title: Trae
slug: trae
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: TRAE_AI_SHELL_ID는 제3자 detector 논의에서만 보았고 Trae 공식 문서·소스에서 확인하지 못함
---

# Trae

## 상태: 보류

Trae는 IDE 기반 AI 개발 도구지만, `TRAE_AI_SHELL_ID`를 agent가 자식 프로세스에 설정한다는 Trae 공식 문서나 공개 소스 근거를 확인하지 못했다. 현재 이 변수는 agents.md의 표준화 논의에서 제3자 detector 후보로만 등장한다. 후보 이름만으로는 값 형식, 설정 시점, 자식 셸 전체 적용 여부를 알 수 없다.

## 조사 경위

Trae 공식 사이트와 문서 포털은 제품 기능·설정을 설명하지만 `TRAE_AI_SHELL_ID`의 정의나 자식 프로세스 주입 계약을 공개하지 않는다. agents.md issue #136은 여러 도구별 변수의 현황을 정리하면서 Trae 후보를 언급하지만, 이는 Trae 제작사의 확인이 아니다.

| 후보 | 공식 확인 | 판정 |
|---|---|---|
| `TRAE_AI_SHELL_ID` | Trae 공식 문서·소스에서 미확인 | **보류** — marker로 채택하지 않음 |

## 재검토 조건

다음 근거가 생기면 재검토한다.

1. Trae 공식 문서가 `TRAE_AI_SHELL_ID`의 값과 주입 범위를 설명한다.
2. Trae 공개 소스가 자신이 실행하는 셸·프로세스에 이 변수를 설정하는 지점을 보여 준다.
3. 실제 실행 실측이 위 공식 근거와 일치한다.

## 공식 문서·논의

- [Trae 공식 사이트](https://www.trae.ai/)
- [Trae 문서 포털](https://docs.trae.ai/)
- [agents.md 표준화 논의 — tool-specific 후보](https://github.com/agentsmd/agents.md/issues/136#tool-specific-variables-status-quo)
