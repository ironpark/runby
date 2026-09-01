---
title: Semaphore
slug: semaphore
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Semaphore 공식 Environment variables reference가 SEMAPHORE 마커와 pipeline·job·PR 변수의 제공 조건을 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Semaphore

Semaphore CI는 job 환경에 `SEMAPHORE=true`와 `CI=true`를 자동으로 설정합니다. `SEMAPHORE_PIPELINE_ID`, `SEMAPHORE_JOB_ID` 등은 실행 계층을 식별하고, 현재 문서와 이전 호환 환경에 존재할 수 있는 PR 번호 변수는 PR 판정에 사용합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `SEMAPHORE` | `true` | 실행 식별 | Semaphore 환경 표시 | 적합 — 전용 마커 | [Environment variables](https://docs.semaphoreci.com/EE/reference/env-vars) |
| `SEMAPHORE_PIPELINE_ID` | UUID 문자열 | 실행 식별 | 현재 pipeline ID | 적합 — pipeline 식별자 | [Environment variables](https://docs.semaphoreci.com/EE/reference/env-vars) |
| `SEMAPHORE_JOB_ID` | UUID 문자열 | 실행 식별 | 현재 job ID | 적합 — job 식별자 | [Environment variables](https://docs.semaphoreci.com/EE/reference/env-vars) |
| `PULL_REQUEST_NUMBER` | 정수 문자열 | 상태·컨텍스트 | PR 번호를 제공하는 호환 변수 | 적합 — 존재하면 `PullRequest=true`, `PullRequestID`로 사용 | [ci-info compatibility](https://github.com/watson/ci-info/blob/master/vendors.json) |
| `SEMAPHORE_GIT_PR_NUMBER` | 정수 문자열 | 상태·컨텍스트 | 현재 PR 번호 | 적합 — 최신 문서의 이름. 위 호환 변수와 대안으로 사용 | [Environment variables](https://docs.semaphoreci.com/EE/reference/env-vars) |

공식 최신 이름은 `SEMAPHORE_GIT_PR_NUMBER`이며, ci-info가 사용해 온 `PULL_REQUEST_NUMBER`도 이미 배포된 Semaphore 환경과의 호환을 위해 함께 인식합니다. `SEMAPHORE_GIT_REF_TYPE=pull-request`만으로는 번호가 없는 실행을 구분할 수 있어, 번호 변수가 있을 때만 PR로 보고합니다.

## 공식 문서

- [Environment variables](https://docs.semaphoreci.com/EE/reference/env-vars)
- [Semaphore pipelines](https://docs.semaphoreci.com/ci-cd/pipelines)
- [ci-info vendor definition](https://github.com/watson/ci-info/blob/master/vendors.json)
