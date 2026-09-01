---
title: Buddy
slug: buddy
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Buddy 공식 default variables 문서가 BUDDY_WORKSPACE_ID와 execution·pull request 변수를 자동 제공한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Buddy

Buddy는 pipeline 실행에 `BUDDY_WORKSPACE_ID`를 자동으로 설정합니다. 공식 문서에는 최신 `BUDDY_RUN_*` 이름과 하위 호환용 `BUDDY_EXECUTION_*` 이름이 함께 안내되며, `runby`는 ci-info가 사용한 구형 PR ID를 포함해 두 세대를 읽습니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `BUDDY_WORKSPACE_ID` | 정수 문자열 | 실행 식별 | 현재 workspace 식별자 | 적합 — 전용 마커 | [Default variables](https://buddy.works/docs/basics/environment-variables/default-variables) |
| `BUDDY_EXECUTION_ID` / `BUDDY_RUN_ID` | 문자열 또는 정수 | 실행 식별 | 현재 pipeline run 식별자 | 적합 — `PipelineID`는 legacy 이름 우선 | [Default variables](https://buddy.works/docs/basics/environment-variables/default-variables) |
| `BUDDY_ACTION_ID` | 정수 문자열 | 실행 식별 | 현재 action 식별자 | 보조 — `JobID` | [Default variables](https://buddy.works/docs/basics/environment-variables/default-variables) |
| `BUDDY_EXECUTION_PULL_REQUEST_ID` / `BUDDY_RUN_PR_ID` | 문자열 | 상태·컨텍스트 | 현재 PR 식별자 | 적합 — 존재하면 `PullRequest=true`, `PullRequestID` | [Default variables](https://buddy.works/docs/basics/environment-variables/default-variables) |
| `BUDDY_EXECUTION_PULL_REQUEST_NO` / `BUDDY_RUN_PR_NO` | 정수 문자열 | 상태·컨텍스트 | 현재 PR 번호 | 적합 — ID가 없을 때 사용 | [Default variables](https://buddy.works/docs/basics/environment-variables/default-variables) |

공식 문서가 제공하는 `BUDDY` 불리언도 있지만 ci-info와의 호환성을 위해 workspace ID를 실행 마커로 사용합니다. 사용자 secret·webhook payload는 읽지 않습니다.

## 공식 문서

- [Default variables](https://buddy.works/docs/basics/environment-variables/default-variables)
- [Introducing default variables](https://buddy.works/blog/introducing-default-variables)
