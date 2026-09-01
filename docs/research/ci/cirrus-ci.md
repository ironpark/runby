---
title: Cirrus CI
slug: cirrus-ci
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Cirrus CI 공식 Writing Tasks 문서가 CIRRUS_CI와 CIRRUS_PR을 기본 환경변수로 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Cirrus CI

Cirrus CI는 task가 실행되는 환경에 Cirrus 전용 `CIRRUS_*` 변수를 자동으로 주입합니다. `CIRRUS_CI`가 존재하면 Cirrus CI로 판정하고, task·build·PR 식별자를 읽습니다. `CI`와 `CONTINUOUS_INTEGRATION`은 Cirrus도 설정하지만 범용 신호이므로 보조 Evidence로만 취급합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CIRRUS_CI` | 값이 있는 문자열 | 실행 식별 | Cirrus CI 환경 표시 | 적합 — 전용 마커 | [Writing Tasks — Environment Variables](https://cirrus-ci.org/guide/writing-tasks/) |
| `CIRRUS_BUILD_ID` | 문자열 | 실행 식별 | 현재 build ID | 적합 — pipeline 식별자 | [Writing Tasks — Environment Variables](https://cirrus-ci.org/guide/writing-tasks/) |
| `CIRRUS_TASK_ID` | 문자열 | 실행 식별 | 현재 task ID | 적합 — job 식별자 | [Writing Tasks — Environment Variables](https://cirrus-ci.org/guide/writing-tasks/) |
| `CIRRUS_PR` | PR 번호 문자열 | 상태·컨텍스트 | PR에서 시작된 build의 번호 | 적합 — 존재하면 `PullRequest=true`, `PullRequestID`로 사용 | [Writing Tasks — Environment Variables](https://cirrus-ci.org/guide/writing-tasks/) |

`CIRRUS_PR`는 PR build에서만 제공되며, 일반 branch/tag build에서는 비어 있습니다. Evidence에는 실제로 설정된 변수 이름만 기록하고 PR 번호 값은 기록하지 않습니다.

## 공식 문서

- [Writing Tasks — Environment Variables](https://cirrus-ci.org/guide/writing-tasks/)
- [Programming Tasks in Starlark](https://cirrus-ci.org/guide/programming-tasks/)
