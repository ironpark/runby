---
title: Devin
slug: devin
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 문서에는 workspace VM 설명만 있고 후보 식별자는 환경변수가 아닌 파일 경로임
---

# Devin

## 요약

**드라이버 없음 — 범위 밖.** Devin은 원격 workspace/VM에서 셸과 IDE를 제공하지만, agent가 실행한 자식 프로세스를 식별하는 공식 환경변수는 확인되지 않았다. 후보로 조사된 `/opt/.devin`은 환경변수가 아니라 파일 경로 존재 여부를 보는 filesystem 신호다. `runby`의 agent 축은 환경 기반 마커를 다루므로 이 신호를 구현하지 않는다.

## 조사 경위

Devin 공식 onboarding 문서는 workspace 환경과 VM에서의 개발 도구를 설명하지만 `DEVIN_*` 실행 marker를 공개하지 않는다. `/opt/.devin`은 agents.md 표준화 논의에서 제3자 detector가 사용한 filesystem 후보로 언급된 값이며, 제작사가 자식 프로세스용 환경변수 계약으로 문서화한 것이 아니다.

| 후보 | 형태 | 판정 |
|---|---|---|
| `/opt/.devin` | 파일/디렉터리 경로 | **범위 밖** — env가 아니며 원격 workspace 이미지에도 존재할 수 있음 |
| `DEVIN_*` | 공식 실행 marker | 확인되지 않음 |
| Devin workspace VM | remote 실행 환경 | agent 축이 아니라 remote 축 성격 |

## 결론

filesystem marker를 env 기반 감지에 섞으면 권한·이미지·잔존 상태에 따른 오탐과 플랫폼 종속이 생긴다. Devin이 자식 프로세스용 환경변수를 공식적으로 제공하기 전에는 드라이버를 추가하지 않는다.

## 공식 문서·논의

- [Devin environment](https://docs.devin.ai/onboard-devin/environment)
- [Devin 소개 — shell/IDE workspace](https://docs.devin.ai/get-started/devin-intro)
- [agents.md 표준화 논의 — filesystem 후보 `/opt/.devin`](https://github.com/agentsmd/agents.md/issues/136#tool-specific-variables-status-quo)
