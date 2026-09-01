---
title: Cloudflare Pages
slug: cloudflare-pages
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Cloudflare 공식 Pages build configuration 문서가 CF_PAGES를 기본 주입 시스템 변수로 명시하고 branch·commit·URL 컨텍스트도 함께 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Cloudflare Pages

Cloudflare Pages의 Git 연동 build system은 모든 build step에 `CI=true`와 `CF_PAGES=1`을 기본으로 주입합니다. `CF_PAGES`가 Pages 전용 marker이며, branch·commit·배포 URL은 실행 컨텍스트로만 수집합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CF_PAGES` | `1` | 실행 식별 | Cloudflare Pages build임을 표시 | 적합 — Pages가 기본 주입하는 전용 marker | [Pages build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration/) |
| `CF_PAGES_BRANCH` | branch 이름 | 상태·컨텍스트 | 현재 배포 branch | 보조 신호 — 실행 위치 컨텍스트 | [Pages build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration/) |
| `CF_PAGES_COMMIT_SHA` | Git SHA | 상태·컨텍스트 | 현재 build commit | 보조 신호 — pipeline ID로 오인하지 않도록 Extra로만 보존 | [Pages build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration/) |
| `CF_PAGES_URL` | URL 문자열 | 상태·컨텍스트 | 현재 배포 URL | 보조 신호 — 실행 결과 컨텍스트 | [Pages build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration/) |

Cloudflare Pages 공식 기본 변수 목록에는 PR 번호나 PR 여부를 직접 광고하는 변수가 없으므로 `PullRequest`와 `PullRequestID`는 채우지 않습니다. `CF_PAGES` 값은 `1`이지만 값 자체를 Evidence에 복사하지 않고 이름만 기록합니다.

## 공식 문서

- [Pages build configuration — Environment variables](https://developers.cloudflare.com/pages/configuration/build-configuration/)
- [Pages Git integration](https://developers.cloudflare.com/pages/get-started/git-integration/)
