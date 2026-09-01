---
title: Codefresh
slug: codefresh
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Codefresh 공식 pipeline variables 문서가 CF_BUILD_ID와 pull request 변수를 기본값으로 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Codefresh

Codefresh pipeline은 `CF_BUILD_ID`를 모든 build에 주입합니다. PR build에서는 `CF_PULL_REQUEST_NUMBER` 또는 `CF_PULL_REQUEST_ID`를 제공하므로 해당 변수가 있을 때만 PR로 보고합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CF_BUILD_ID` | 문자열 | 실행 식별 | 현재 Codefresh build 식별자 | 적합 — 전용 마커이자 `PipelineID` | [Pipeline variables](https://codefresh.io/docs/docs/pipelines/variables/) |
| `CF_PULL_REQUEST_NUMBER` | 정수 문자열 | 상태·컨텍스트 | PR 번호 | 적합 — `PullRequestID` 우선값 | [Pipeline variables](https://codefresh.io/docs/docs/pipelines/variables/) |
| `CF_PULL_REQUEST_ID` | 문자열 | 상태·컨텍스트 | PR 식별자 | 적합 — 번호가 없을 때 `PullRequestID` | [Pipeline variables](https://codefresh.io/docs/docs/pipelines/variables/) |
| `CF_BRANCH` / `CF_REVISION` | 문자열 | 상태·컨텍스트 | branch·revision | 보조 — Extra에 보존 | [Pipeline variables](https://codefresh.io/docs/docs/pipelines/variables/) |

`CF_BUILD_ID`는 사용자 정의 build argument가 아니라 Codefresh가 주입하는 build 식별자입니다. API token과 secret 변수는 읽지 않습니다.

## 공식 문서

- [Pipeline variables](https://codefresh.io/docs/docs/pipelines/variables/)
- [Codefresh CI integrations](https://codefresh.io/docs/docs/gitops-integrations/ci-integrations/codefresh-classic/)
