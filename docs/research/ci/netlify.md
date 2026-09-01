---
title: Netlify
slug: netlify
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Netlify 공식 Build environment variables 문서가 NETLIFY와 PULL_REQUEST의 자동 제공·값 의미를 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Netlify

Netlify 빌드 시스템은 빌드 환경에 `NETLIFY`를 포함한 기본 환경변수를 자동으로 주입합니다. `NETLIFY`가 있으면 Netlify CI로 보고하며, `DEPLOY_ID`와 `BUILD_ID`를 각각 배포·빌드 식별자로 사용합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `NETLIFY` | 불리언 문자열 | 실행 식별 | Netlify 빌드임을 표시 | 적합 — Netlify가 기본으로 제공하는 전용 마커 | [Build environment variables](https://docs.netlify.com/build/configure-builds/environment-variables/) |
| `DEPLOY_ID` | 문자열 | 실행 식별 | 현재 배포 ID | 적합 — 배포 단위 식별자 | [Build environment variables](https://docs.netlify.com/build/configure-builds/environment-variables/) |
| `BUILD_ID` | 문자열 | 실행 식별 | 현재 빌드 ID | 적합 — Netlify 마커가 확인된 뒤에만 읽는 job 식별자 | [Build environment variables](https://docs.netlify.com/build/configure-builds/environment-variables/) |
| `PULL_REQUEST` | `true` 또는 `false` | 상태·컨텍스트 | PR/merge request 빌드 여부 | 적합 — `false`가 아니면 `PullRequest=true`; 값은 `PullRequestID`로 쓰지 않고 이름만 Evidence에 기록 | [Build environment variables](https://docs.netlify.com/build/configure-builds/environment-variables/) |

Netlify의 `PULL_REQUEST`는 번호가 아니라 요청 여부 플래그이므로 `PullRequestID`는 비워 둡니다. `PULL_REQUEST=false`는 PR이 아닌 실행으로 처리하고, `NETLIFY`가 사용자 설정으로 덮어써질 수 있다는 문서의 한계는 다른 환경변수 마커와 동일하게 적용됩니다.

## 공식 문서

- [Build environment variables](https://docs.netlify.com/build/configure-builds/environment-variables/)
- [Continuous deployment overview](https://docs.netlify.com/site-deploys/overview/)
