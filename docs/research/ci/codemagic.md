---
title: Codemagic
slug: codemagic
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Codemagic 공식 built-in environment variables 문서가 CM_BUILD_ID와 CM_PULL_REQUEST 계열을 build 환경에 자동 제공한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Codemagic

Codemagic은 build에 `CM_BUILD_ID`를 자동으로 설정합니다. `CM_PULL_REQUEST`는 PR build에서 `true`, 그 외에는 `false`이고, `CM_PULL_REQUEST_NUMBER`가 요청 번호를 제공합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CM_BUILD_ID` | UUID 문자열 | 실행 식별 | 현재 build UUID | 적합 — 전용 마커이자 `PipelineID` | [Built-in environment variables](https://docs.codemagic.io/yaml-basic-configuration/environment-variables/) |
| `BUILD_NUMBER` | 정수 문자열 | 실행 식별 | workflow의 build 번호 | 보조 — Codemagic 마커가 맞은 뒤에만 읽음 | [Built-in environment variables](https://docs.codemagic.io/yaml-basic-configuration/environment-variables/) |
| `CM_WORKFLOW_NAME` | 문자열 | 실행 식별 | 현재 workflow 이름 | 보조 — `JobName` | [Built-in environment variables](https://docs.codemagic.io/yaml-basic-configuration/environment-variables/) |
| `CM_PULL_REQUEST` | `true` 또는 `false` | 상태·컨텍스트 | 현재 build의 PR 여부 | 적합 — `true`일 때 `PullRequest=true` | [Built-in environment variables](https://docs.codemagic.io/yaml-basic-configuration/environment-variables/) |
| `CM_PULL_REQUEST_NUMBER` | 정수 문자열 | 상태·컨텍스트 | PR 번호 | 적합 — `PullRequestID` | [Built-in environment variables](https://docs.codemagic.io/yaml-basic-configuration/environment-variables/) |

`BUILD_NUMBER`는 일반 이름이지만 Codemagic 전용 `CM_BUILD_ID`가 먼저 맞은 경우에만 컨텍스트로 읽습니다. signing secret·API token은 읽지 않습니다.

## 공식 문서

- [Built-in environment variables](https://docs.codemagic.io/yaml-basic-configuration/environment-variables/)
