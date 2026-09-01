---
title: Factory Droid
slug: factory-droid
research_date: 2026-09-02
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: 핵심 실행 구현이 비공개이고 공식 문서에는 설정·인증 변수만 있어 자식 환경 계약을 확인할 수 없음
---

# Factory Droid

## 요약

**드라이버 없음 — 조사 결과 제외.** Factory Droid의 핵심 실행 소스는 공개되어 있지 않다. 공식 문서에서 확인되는 `settings.json`과 `FACTORY_API_KEY`는 사용자가 Droid를 구성하거나 인증하기 위한 입력이다. Droid가 자신이 실행하는 셸·도구 프로세스에 고유 환경변수를 자동 주입한다는 공개 계약은 찾지 못했다.

## 조사 경위

공식 Droid CLI settings 문서는 설정 파일의 사용자 설정 항목을 설명하고, Software Factory의 CI/code review 문서는 `FACTORY_API_KEY`를 secret으로 전달하는 방법을 설명한다. 두 문서 모두 자식 프로세스에 실행 ID나 agent marker를 주입한다고 하지 않는다.

| 변수/경로 | 성격 | 판정 |
|---|---|---|
| `FACTORY_API_KEY` | API 인증 secret | **부적합** — 인증 입력이며 실행 주체 마커가 아님 |
| `settings.json` | Droid CLI 사용자 설정 | **부적합** — 실행 전 구성 상태 |
| 비공개 자식 환경 주입 구현 | 공개 근거 없음 | **확인 불가** — 임의의 내부 변수를 채택하지 않음 |

## 결론

공식 문서에 설정·인증 변수만 있고, 비공개 구현을 추측할 수 없으므로 Factory Droid는 agent 축에서 감지하지 않는다. 향후 제작사가 자식 프로세스용 실행 마커와 주입 범위를 문서화하면 재검토한다.

## 공식 문서

- [Droid CLI settings](https://docs.factory.ai/droid-cli/settings)
- [Software Factory code review/CI — FACTORY_API_KEY](https://docs.factory.ai/software-factory/code-review-ci)
