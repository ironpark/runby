---
title: AppVeyor
slug: appveyor
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: AppVeyor 공식 Environment Variables 문서가 APPVEYOR 전용 마커와 build·job·PR 번호를 모든 build의 기본 변수로 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# AppVeyor

AppVeyor는 build worker가 저장소 checkout 직후 기본 환경변수를 설정합니다. `APPVEYOR=true`가 AppVeyor 전용 실행 마커이고, build·job 식별자와 PR 번호를 별도 변수로 제공합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `APPVEYOR` | `True` 또는 `true` | 실행 식별 | AppVeyor 환경 표시 | 적합 — 전용 마커 | [Environment variables](https://www.appveyor.com/docs/environment-variables/) |
| `APPVEYOR_BUILD_ID` | 문자열 | 실행 식별 | AppVeyor 고유 build ID | 적합 — pipeline 식별자 | [Environment variables](https://www.appveyor.com/docs/environment-variables/) |
| `APPVEYOR_JOB_ID` | 문자열 | 실행 식별 | AppVeyor 고유 job ID | 적합 — job 식별자 | [Environment variables](https://www.appveyor.com/docs/environment-variables/) |
| `APPVEYOR_PULL_REQUEST_NUMBER` | 정수 문자열 | 상태·컨텍스트 | Pull/Merge Request 번호 | 적합 — 존재하면 `PullRequest=true`, `PullRequestID`로 사용 | [Environment variables](https://www.appveyor.com/docs/environment-variables/) |

`CI`도 AppVeyor가 설정하지만 범용 변수이므로 단독 마커로 쓰지 않습니다. `APPVEYOR_PULL_REQUEST_NUMBER`는 PR 실행에서만 제공된다는 공식 설명에 따라, 변수 이름만 Evidence에 남깁니다.

## 공식 문서

- [Environment variables](https://www.appveyor.com/docs/environment-variables/)
- [Build configuration](https://www.appveyor.com/docs/build-configuration/)
