---
title: Jules
slug: jules
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 환경 문서는 task VM 구성과 사용자 환경변수만 설명하고 VM 또는 agent 자식용 마커를 제공하지 않음
---

# Jules

## 요약

**드라이버 없음 — 음성 결과.** Jules는 작업마다 보안 단기 VM을 만들지만, 공식 환경 문서에는 그 VM이나 Jules Agent가 실행한 프로세스를 표시하는 환경변수가 없다. 설치 도구와 repo-level 환경변수는 작업 환경을 구성하는 입력일 뿐 실행 주체 마커가 아니다.

## 조사 경위

Jules 환경 문서는 task VM의 수명, 기본 도구, setup script를 설명하고, changelog는 repo-level environment variables를 task 동안 사용할 수 있게 된 기능을 설명한다. 어느 항목도 `JULES_*` VM marker 또는 자식 프로세스용 session ID를 계약하지 않는다.

| 조사 대상 | 공식 의미 | 판정 |
|---|---|---|
| task VM | Jules 작업을 위한 원격 격리 환경 | agent 축 부적합 — remote 환경의 성격 |
| setup/preinstalled tools | VM 준비 과정 | 실행 주체 marker 아님 |
| repo-level environment variables | 사용자가 작업에 제공하는 설정 | 부적합 — 입력값이며 Jules 자식 전용이 아님 |

## 결론

공식 환경 문서만으로 Jules가 현재 프로세스를 실행했다는 것을 판별할 수 없다. VM이 생겼다는 사실이 외부에서 관찰 가능해지더라도 그것은 remote 축의 신호이며, agent 축에 넣으려면 별도의 자식 환경 마커가 필요하다.

## 공식 문서

- [Jules environments](https://jules.google/docs/environment/)
- [Jules changelog — repo-level environment variables](https://jules.google/docs/changelog)
