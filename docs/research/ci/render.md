---
title: Render
slug: render
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Render 공식 환경변수와 service preview 문서가 RENDER 및 IS_PULL_REQUEST를 빌드 환경에 자동 설정한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Render

Render build 환경은 `RENDER=true`를 설정합니다. Pull Request Preview build에서는 `IS_PULL_REQUEST=true`가 추가되고, 일반 build에서는 `false`이므로 두 변수를 각각 실행 마커와 PR 플래그로 사용합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `RENDER` | 불리언 문자열 (`true`) | 실행 식별 | Render 환경 표시 | 적합 — 전용 마커 | [Environment variables](https://render.com/docs/environment-variables) |
| `RENDER_SERVICE_ID` | 문자열 | 실행 식별 | 현재 service 식별자 | 보조 — Extra에 보존 | [Environment variables](https://render.com/docs/environment-variables) |
| `IS_PULL_REQUEST` | `true` 또는 `false` | 상태·컨텍스트 | Preview가 PR에서 생성됐는지 | 적합 — `true`일 때 `PullRequest=true`; ID는 광고하지 않음 | [Service previews](https://render.com/docs/service-previews) |

`IS_PULL_REQUEST`는 요청 번호가 아닌 불리언이므로 `PullRequestID`는 비워 둡니다. 일반 환경변수나 배포 secret은 실행 증거가 아니므로 읽지 않습니다.

## 공식 문서

- [Environment variables](https://render.com/docs/environment-variables)
- [Service previews](https://render.com/docs/service-previews)
