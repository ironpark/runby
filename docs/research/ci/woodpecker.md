---
title: Woodpecker CI
slug: woodpecker
research_date: 2026-09-02
open_source: true
repository: https://github.com/woodpecker-ci/woodpecker
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Woodpecker 공식 환경변수 문서가 CI=woodpecker와 pipeline·step·pull request 변수를 실행 컨테이너에 주입한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Woodpecker CI

Woodpecker는 모든 pipeline step에 `CI=woodpecker`를 주입합니다. 현재 문서는 `CI_PIPELINE_*`·`CI_COMMIT_PULL_REQUEST` 이름을 사용하고, 이전 ci-info 호환 환경은 `CI_BUILD_EVENT`·`CI_PULL_REQUEST` 이름을 사용하므로 `runby`는 두 세대를 함께 읽습니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CI` | `woodpecker` | 실행 식별 | Woodpecker pipeline 환경임을 표시 | 적합 — 제품명이 들어간 전용 마커 | [Environment variables](https://woodpecker-ci.org/docs/next/usage/environment) |
| `CI_PIPELINE_NUMBER` / `CI_BUILD_NUMBER` | 정수 문자열 | 실행 식별 | 현재 pipeline 번호 | 보조 — `PipelineID`는 호환 이름인 `CI_BUILD_NUMBER` 우선 | [Environment variables](https://woodpecker-ci.org/docs/next/usage/environment) |
| `CI_PIPELINE_EVENT` / `CI_BUILD_EVENT` | 이벤트명 (`push`, `pull_request` 등) | 상태·컨텍스트 | pipeline을 시작한 이벤트 | 적합 — 어느 이름이든 `pull_request`이면 `PullRequest=true` | [Environment variables](https://woodpecker-ci.org/docs/next/usage/environment) |
| `CI_COMMIT_PULL_REQUEST` / `CI_PULL_REQUEST` | 정수 문자열 | 상태·컨텍스트 | 현재 pull request 번호 | 적합 — 값이 있으면 `PullRequestID`로 사용 | [Environment variables](https://woodpecker-ci.org/docs/next/usage/environment) |
| `CI_STEP_NAME` | 문자열 | 실행 식별 | 현재 step 이름 | 보조 — `JobName`으로 사용 | [Environment variables](https://woodpecker-ci.org/docs/next/usage/environment) |

현재 공식 이름은 `CI_COMMIT_PULL_REQUEST`이며 PR 이벤트에서만 설정됩니다. 과거 이름도 탐지해 기존 Woodpecker/ci-info 환경을 놓치지 않지만, 값은 Evidence에 기록하지 않고 변수 이름만 남깁니다.

## 공식 문서

- [Environment variables](https://woodpecker-ci.org/docs/next/usage/environment)
- [Woodpecker 공식 저장소](https://github.com/woodpecker-ci/woodpecker)
