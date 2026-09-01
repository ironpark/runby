---
title: Bitrise
slug: bitrise
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Bitrise 공식 환경변수 문서가 BITRISE_IO와 build·workflow·pull request 변수를 기본값으로 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Bitrise

Bitrise는 build 컨테이너에 `BITRISE_IO=true`와 `CI=true`를 설정합니다. `BITRISE_IO`는 범용 `CI`와 달리 Bitrise가 직접 제공하는 실행 마커입니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `BITRISE_IO` | 불리언 문자열 (`true`) | 실행 식별 | Bitrise 환경 표시 | 적합 — 전용 마커 | [Available environment variables](https://docs.bitrise.io/en/bitrise-ci/references/available-environment-variables) |
| `BITRISE_BUILD_SLUG` | UUID 문자열 | 실행 식별 | 현재 build의 고유 slug | 적합 — `PipelineID` | [Available environment variables](https://docs.bitrise.io/en/bitrise-ci/references/available-environment-variables) |
| `BITRISE_BUILD_NUMBER` | 정수 문자열 | 실행 식별 | 프로젝트 build 번호 | 보조 — 사람용 build 카운터 | [Available environment variables](https://docs.bitrise.io/en/bitrise-ci/references/available-environment-variables) |
| `BITRISE_TRIGGERED_WORKFLOW_ID` | 문자열 | 실행 식별 | 실행된 workflow | 보조 — `JobName` | [Available environment variables](https://docs.bitrise.io/en/bitrise-ci/references/available-environment-variables) |
| `BITRISE_PULL_REQUEST` | 정수 문자열 | 상태·컨텍스트 | 현재 pull request 번호 | 적합 — 존재하면 `PullRequest=true`, `PullRequestID`로 사용 | [Available environment variables](https://docs.bitrise.io/en/bitrise-ci/references/available-environment-variables) |

`CI`와 `CONTINUOUS_INTEGRATION`도 Bitrise가 제공하지만 범용 마커이므로 단독 판정에 쓰지 않습니다. 인증서·서명·API 관련 변수는 읽지 않습니다.

## 공식 문서

- [Available environment variables](https://docs.bitrise.io/en/bitrise-ci/references/available-environment-variables)
- [Environment variables](https://docs.bitrise.io/en/bitrise-ci/configure-builds/environment-variables)
