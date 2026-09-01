---
title: Sourcehut builds
slug: sourcehut
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: builds.sr.ht 공식 문서가 build task 환경과 JOB_ID를 확인하고 ci-info가 CI_NAME=sourcehut를 제품명 마커로 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Sourcehut builds

Sourcehut의 builds.sr.ht는 build task 컨테이너에 작업 식별자를 제공합니다. `CI_NAME=sourcehut`는 제품명을 값으로 명시하는 전용 관례이며, `JOB_ID`는 실행 식별자로 읽습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CI_NAME` | 고정 문자열 `sourcehut` | 실행 식별 | Sourcehut CI 이름 | 적합 — 값까지 제품명을 명시하는 전용 마커 | [ci-info vendor definition](https://github.com/watson/ci-info/blob/master/vendors.json) |
| `JOB_ID` | 문자열 | 실행 식별 | 현재 builds.sr.ht job 식별자 | 적합 — `PipelineID` | [builds.sr.ht](https://man.sr.ht/builds.sr.ht/) |

Sourcehut 공식 task 문서는 `JOB_ID`를 제공하지만 `CI_NAME`이라는 이름을 표준 변수 목록으로 별도 열거하지 않습니다. 따라서 `CI_NAME=sourcehut` 채택 근거는 ci-info의 vendor 정의와 그 값이 Sourcehut 제품명을 직접 명시한다는 점이며, 단순히 `CI_NAME`이 존재하는 것만으로는 Sourcehut로 판정하지 않습니다. PR 전용 공통 마커는 확인하지 못했습니다.

## 공식 문서

- [builds.sr.ht manual](https://man.sr.ht/builds.sr.ht/)
- [ci-info vendor definitions](https://github.com/watson/ci-info/blob/master/vendors.json)
