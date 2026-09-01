---
title: Vercel
slug: vercel
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Vercel 공식 System Environment Variables 문서가 시스템이 자동으로 채우는 VERCEL 계열 변수와 PR ID를 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Vercel

Vercel은 배포 빌드와 런타임에 시스템 환경변수를 자동으로 노출합니다. `VERCEL` 또는 레거시 `NOW_BUILDER`가 존재하면 Vercel 빌드 환경으로 판정하고, `VERCEL_DEPLOYMENT_ID`를 배포 식별자로 사용합니다. `VERCEL_GIT_PULL_REQUEST_ID`가 제공되면 해당 배포가 PR에서 시작되었다고 보고합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `VERCEL` | 값이 있는 시스템 변수 | 실행 식별 | Vercel 환경임을 표시 | 적합 — Vercel이 자동으로 제공하는 전용 이름 | [System environment variables](https://vercel.com/docs/environment-variables/system-environment-variables) |
| `NOW_BUILDER` | 값이 있는 레거시 시스템 변수 | 실행 식별 | 레거시 Vercel builder 환경 표시 | 적합 — ci-info가 호환을 위해 보존한 Vercel 전용 신호 | [System environment variables](https://vercel.com/docs/environment-variables/system-environment-variables) |
| `VERCEL_DEPLOYMENT_ID` | 문자열 | 실행 식별 | 현재 배포의 식별자 | 적합 — 배포 단위 식별자 | [System environment variables](https://vercel.com/docs/environment-variables/system-environment-variables) |
| `VERCEL_GIT_PULL_REQUEST_ID` | PR 번호 문자열 또는 빈 문자열 | 상태·컨텍스트 | 배포를 시작한 PR 번호 | 적합 — 값이 있으면 `PullRequest=true`, `PullRequestID`로 사용 | [System environment variables](https://vercel.com/docs/environment-variables/system-environment-variables) |

`VERCEL`과 `NOW_BUILDER`는 대안 마커이므로 둘 중 하나만 있어도 감지합니다. 일반 프로젝트 설정값과 인증·OIDC 토큰은 실행 마커로 사용하지 않으며, Evidence에는 변수 이름만 기록합니다.

## 공식 문서

- [System environment variables](https://vercel.com/docs/environment-variables/system-environment-variables)
- [Vercel build pipeline](https://vercel.com/docs/deployments/builds)
