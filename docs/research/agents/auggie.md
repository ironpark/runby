---
title: Auggie (Augment Code CLI)
slug: auggie
research_date: 2026-09-01
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 CLI 레퍼런스가 변수의 값·주입 시점·감지 용도를 모두 명시함
---

# Auggie (Augment Code CLI)

## 요약

Auggie는 `launch-process` 도구로 셸 명령을 실행할 때 `AUGMENT_AGENT=1`을 주입합니다. 공식 CLI 레퍼런스는 "명령이 Auggie에 의해 실행될 때 `1`로 설정되며, 스크립트는 이 변수를 확인해 에이전트가 실행 중인지 감지할 수 있다"고 명시합니다. **감지 용도가 제품 문서에 명시된** 드문 사례입니다.

## 환경변수

| 변수 | 값 | 분류 | 의미 | 판정 | 근거 |
|---|---|---|---|---|---|
| `AUGMENT_AGENT` | `1` | 실행 식별 | Auggie가 실행한 명령임을 표시 | **적합** — 공식 문서가 감지 목적으로 제공한다고 명시한 주입 마커 | [CLI Flags and Options](https://docs.augmentcode.com/cli/reference) |
| `AUGMENT_API_TOKEN` | 토큰 문자열 | 설정 | CI/CD 인증 토큰 | 부적합 — **자격증명**. 값은 물론 이름도 근거에 넣지 않음 | [CLI Flags and Options](https://docs.augmentcode.com/cli/reference) |
| `AUGMENT_API_URL` | URL 문자열 | 설정 | API 엔드포인트 지정 | 부적합 — 실행 전 사용자 설정 | [CLI Flags and Options](https://docs.augmentcode.com/cli/reference) |

## 실행 파일

`auggie`.
