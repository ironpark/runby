---
title: Screwdriver
slug: screwdriver
research_date: 2026-09-02
open_source: true
repository: https://github.com/screwdriver-cd/screwdriver
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Screwdriver 공식 environment variables 문서가 SCREWDRIVER와 build·job·pull request 변수를 모든 build에 제공한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Screwdriver

Screwdriver는 build 환경에 `SCREWDRIVER=true`를 설정합니다. PR build에서는 `SD_PULL_REQUEST`에 요청 번호를 넣고, 비-PR build에서는 `false` 또는 빈 값이므로 `false`가 아닌 비어 있지 않은 값만 PR로 인정합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `SCREWDRIVER` | 불리언 문자열 (`true`) | 실행 식별 | Screwdriver 환경 표시 | 적합 — 전용 마커 | [Environment variables](https://docs.screwdriver.cd/user-guide/environment-variables) |
| `SD_BUILD_ID` | 정수 문자열 | 실행 식별 | 현재 build ID | 적합 — `PipelineID` | [Environment variables](https://docs.screwdriver.cd/user-guide/environment-variables) |
| `SD_JOB_ID` / `SD_JOB_NAME` | 문자열 | 실행 식별 | 현재 job 식별자·이름 | 보조 — `JobID`·`JobName` | [Environment variables](https://docs.screwdriver.cd/user-guide/environment-variables) |
| `SD_PULL_REQUEST` | PR 번호 또는 `false`/빈 값 | 상태·컨텍스트 | 현재 PR 번호 | 적합 — `false`가 아니고 값이 있으면 `PullRequest=true`, `PullRequestID` | [Environment variables](https://docs.screwdriver.cd/user-guide/environment-variables) |

Screwdriver도 `CI`와 `CONTINUOUS_INTEGRATION`을 제공하지만 범용 이름이므로 전용 `SCREWDRIVER`가 먼저 맞아야 합니다. secret·metadata payload는 읽지 않습니다.

## 공식 문서

- [Environment variables](https://docs.screwdriver.cd/user-guide/environment-variables)
- [Screwdriver 공식 저장소](https://github.com/screwdriver-cd/screwdriver)
