---
title: Cloudflare Workers Builds
slug: cloudflare-workers
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Cloudflare 공식 Workers Builds configuration 문서가 WORKERS_CI와 build UUID·branch·commit 변수를 기본 주입 시스템 변수로 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Cloudflare Workers Builds

Cloudflare Workers Builds는 연결된 Git 저장소의 build/deploy command를 실행할 때 `CI=true`와 `WORKERS_CI=1`을 주입합니다. `WORKERS_CI`는 Workers Builds 전용 marker이고, `WORKERS_CI_BUILD_UUID`는 현재 build를 식별합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `WORKERS_CI` | `1` | 실행 식별 | Workers Builds 환경 표시 | 적합 — Cloudflare가 기본 주입하는 전용 marker | [Workers Builds configuration](https://developers.cloudflare.com/workers/ci-cd/builds/configuration/) |
| `WORKERS_CI_BUILD_UUID` | UUID 문자열 | 실행 식별 | 현재 build UUID | 적합 — pipeline 식별자 | [Workers Builds configuration](https://developers.cloudflare.com/workers/ci-cd/builds/configuration/) |
| `WORKERS_CI_BRANCH` | branch 이름 | 상태·컨텍스트 | push event의 source branch | 보조 신호 | [Workers Builds configuration](https://developers.cloudflare.com/workers/ci-cd/builds/configuration/) |
| `WORKERS_CI_COMMIT_SHA` | Git SHA | 상태·컨텍스트 | 현재 build commit | 보조 신호 | [Workers Builds configuration](https://developers.cloudflare.com/workers/ci-cd/builds/configuration/) |

공식 기본 변수 목록에는 PR 여부나 PR ID가 없으므로 `PullRequest`와 `PullRequestID`는 채우지 않습니다. API token·사용자 정의 build variable은 실행 마커가 아니므로 읽지 않습니다.

## 공식 문서

- [Workers Builds configuration](https://developers.cloudflare.com/workers/ci-cd/builds/configuration/)
- [Workers Builds overview](https://developers.cloudflare.com/workers/ci-cd/builds/)
