---
title: Harness CI
slug: harness-ci
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Harness 공식 CI 환경변수 문서가 HARNESS_BUILD_ID와 execution·stage 식별자를 기본 변수로 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Harness CI

Harness CI는 step 환경에 `HARNESS_BUILD_ID`를 주입합니다. 이 이름은 사용자 설정값보다 Harness가 생성하는 build 식별자이므로, 범용 `CI`나 비밀 변수 없이도 실행 주체를 판정할 수 있습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `HARNESS_BUILD_ID` | 문자열 | 실행 식별 | 현재 Harness build 식별자 | 적합 — 전용 build 마커이자 `PipelineID` | [CI environment variables](https://developer.harness.io/docs/continuous-integration/troubleshoot-ci/ci-env-var/) |
| `HARNESS_EXECUTION_ID` | 문자열 | 실행 식별 | Harness execution 식별자 | 보조 — 현재 버전에서 별도 필드로 읽지 않음 | [CI environment variables](https://developer.harness.io/docs/continuous-integration/troubleshoot-ci/ci-env-var/) |
| `HARNESS_STAGE_ID` / `HARNESS_STAGE_NAME` | 문자열 | 실행 식별 | 현재 stage 식별자·이름 | 보조 — `JobID`·`JobName` | [CI environment variables](https://developer.harness.io/docs/continuous-integration/troubleshoot-ci/ci-env-var/) |
| `HARNESS_ORG_ID` / `HARNESS_PROJECT_ID` / `HARNESS_PIPELINE_ID` | 문자열 | 상태·컨텍스트 | Harness 리소스 식별자 | 보조 — Extra에 보존 | [CI environment variables](https://developer.harness.io/docs/continuous-integration/troubleshoot-ci/ci-env-var/) |

공식 목록에는 build 식별자는 있지만 모든 실행에서 PR 여부를 나타내는 독립적인 Harness 전용 변수가 확인되지 않았습니다. 따라서 `PullRequest`는 추정하지 않습니다. API key·delegate token 등 인증 변수는 읽지 않습니다.

## 공식 문서

- [CI environment variables](https://developer.harness.io/docs/continuous-integration/troubleshoot-ci/ci-env-var/)
