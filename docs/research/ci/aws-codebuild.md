---
title: AWS CodeBuild
slug: aws-codebuild
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: AWS 공식 CodeBuild build environment reference가 CODEBUILD_BUILD_ARN과 webhook event family를 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# AWS CodeBuild

AWS CodeBuild는 build environment에 `CODEBUILD_*` 시스템 환경변수를 주입합니다. `CODEBUILD_BUILD_ARN`은 CodeBuild build의 ARN이므로 가장 직접적인 존재 마커이자 pipeline 식별자입니다. webhook으로 시작된 build에서 `CODEBUILD_WEBHOOK_EVENT`가 `PULL_REQUEST_*` 계열이면 PR build로 보고합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `CODEBUILD_BUILD_ARN` | ARN 문자열 | 실행 식별 | 현재 build의 ARN | 적합 — CodeBuild 전용 build 식별자 | [Environment variables in build environments](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html) |
| `CODEBUILD_BUILD_ID` | 문자열 | 실행 식별 | CodeBuild build ID | 적합 — job/pipeline 보조 식별자 | [Environment variables in build environments](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html) |
| `CODEBUILD_WEBHOOK_EVENT` | 이벤트명 (`PULL_REQUEST_CREATED` 등) | 상태·컨텍스트 | webhook을 시작한 이벤트 | 적합 — `PULL_REQUEST_` 접두사면 `PullRequest=true` | [GitHub webhook events](https://docs.aws.amazon.com/codebuild/latest/userguide/github-webhook.html) |
| `CODEBUILD_SOURCE_VERSION` | source revision 문자열 | 상태·컨텍스트 | build가 사용하는 source 버전 | 보조 신호 — PR에서는 `pr/<번호>` 형식일 수 있지만 값 파싱은 하지 않음 | [Environment variables in build environments](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html) |

CodeBuild의 event 이름은 생성·갱신·재개·병합·종료 등 여러 `PULL_REQUEST_*` 값을 가질 수 있으므로 특정 세 값만 열거하지 않고 공식 접두사 규칙으로 처리합니다. 요청 번호를 독립적인 PR ID 변수로 광고하지 않으므로 `PullRequestID`는 비워 둡니다.

## 공식 문서

- [Environment variables in build environments](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html)
- [GitHub webhook events](https://docs.aws.amazon.com/codebuild/latest/userguide/github-webhook.html)
- [Build environment reference](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref.html)
