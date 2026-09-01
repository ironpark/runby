---
title: Bamboo
slug: bamboo
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Atlassian 공식 Bamboo variables 문서가 bamboo_planKey 계열의 시스템 변수를 build 환경에 제공한다고 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Bamboo

Atlassian Bamboo는 점(`.`)이 있는 내부 변수를 환경변수 이름으로 변환해 step에 제공합니다. `bamboo_planKey`는 Bamboo plan에 고유한 전용 이름이므로 실행 마커로 사용합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `bamboo_planKey` | 문자열 | 실행 식별 | 현재 plan key | 적합 — Bamboo 전용 마커 | [Bamboo variables](https://confluence.atlassian.com/bamboo0800/bamboo-variables-1077779389.html) |
| `bamboo_buildResultKey` | 문자열 | 실행 식별 | build result key | 적합 — `PipelineID` | [Bamboo variables](https://confluence.atlassian.com/bamboo0800/bamboo-variables-1077779389.html) |
| `bamboo_buildNumber` | 정수 문자열 | 실행 식별 | plan build 번호 | 보조 — 사람용 카운터 | [Bamboo variables](https://confluence.atlassian.com/bamboo0800/bamboo-variables-1077779389.html) |
| `bamboo_buildKey` / `bamboo_planName` | 문자열 | 실행 식별 | build key·plan 이름 | 보조 — `JobID`·`JobName` | [Bamboo variables](https://confluence.atlassian.com/bamboo0800/bamboo-variables-1077779389.html) |
| `bamboo_repository_pr_key` | 문자열 | 상태·컨텍스트 | repository pull request key | 적합 — 존재하면 `PullRequest=true`, `PullRequestID`로 사용 | [Bamboo variables](https://confluence.atlassian.com/bamboo0800/bamboo-variables-1077779389.html) |

변수명은 Bamboo의 점 표기(`bamboo.planKey`, `bamboo.repository.pr.key`)가 환경변수로 변환된 형태입니다. Bamboo plan 변수를 사용자가 덮어쓸 수 있다는 일반적 한계가 있으므로 권한 판단의 근거로 사용하지 않습니다.

## 공식 문서

- [Bamboo variables](https://confluence.atlassian.com/bamboo0800/bamboo-variables-1077779389.html)
