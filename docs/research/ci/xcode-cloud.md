---
title: Xcode Cloud
slug: xcode-cloud
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Apple 공식 Xcode Cloud environment variable reference가 CI_XCODE_PROJECT와 build·workflow·PR 변수를 항상 또는 조건부로 제공한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Xcode Cloud

Xcode Cloud는 custom build script가 실행되는 임시 build environment에 `CI=true`, `CI_XCODE_CLOUD=true`, `CI_XCODE_PROJECT`와 build 식별자를 제공합니다. `CI_XCODE_PROJECT`는 Xcode Cloud 전용 프로젝트/workspace 이름으로 문서화된 marker라서 `CI`만으로 감지하지 않습니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CI_XCODE_PROJECT` | 프로젝트/workspace 이름 | 실행 식별 | 현재 Xcode Cloud action의 프로젝트 | 적합 — 전용 marker | [Environment variable reference](https://developer.apple.com/documentation/xcode/environment-variable-reference) |
| `CI_BUILD_ID` | UUID 문자열 | 실행 식별 | 현재 build ID | 적합 — pipeline 식별자 | [Environment variable reference](https://developer.apple.com/documentation/xcode/environment-variable-reference) |
| `CI_BUILD_NUMBER` | 정수 문자열 | 실행 식별 | 현재 build 번호 | 보조 신호 — 사람이 읽는 카운터 | [Environment variable reference](https://developer.apple.com/documentation/xcode/environment-variable-reference) |
| `CI_WORKFLOW_ID` | UUID 문자열 | 실행 식별 | workflow ID | 적합 — job/workflow 식별자 | [Environment variable reference](https://developer.apple.com/documentation/xcode/environment-variable-reference) |
| `CI_PULL_REQUEST_NUMBER` | 정수 문자열 | 상태·컨텍스트 | PR change start condition의 PR 번호 | 적합 — 존재하면 `PullRequest=true`, `PullRequestID`로 사용 | [PR-specific variables](https://developer.apple.com/documentation/xcode/environment-variable-reference?language=objc) |

`CI_PULL_REQUEST_NUMBER`는 PR start condition에서만 제공된다고 문서에 명시되어 있어 별도 boolean 변수 없이 존재 여부를 판정에 사용합니다. `CI_XCODE_CLOUD`는 같은 사실을 말하는 boolean 보조 마커지만 현재 드라이버는 프로젝트 이름을 기준으로 하므로 Evidence에는 실제로 읽은 이름만 남깁니다.

## 공식 문서

- [Environment variable reference](https://developer.apple.com/documentation/xcode/environment-variable-reference)
- [Writing custom build scripts](https://developer.apple.com/documentation/xcode/writing-custom-build-scripts)
