---
title: Vela
slug: vela
research_date: 2026-09-02
open_source: true
repository: https://github.com/go-vela/vela
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Vela 공식 environment variables 문서가 VELA와 VELA_PULL_REQUEST를 event별 기본 변수로 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Vela

Vela pipeline은 `VELA` 계열의 build 환경변수를 주입합니다. Pull request event에서는 `VELA_PULL_REQUEST=1`이므로 이 고정값을 확인해 PR 실행을 판정합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `VELA` | 값이 있는 Vela 시스템 변수 | 실행 식별 | Vela 환경 표시 | 적합 — 전용 마커 | [Environment variables](https://go-vela.github.io/docs/reference/environment/variables) |
| `VELA_BUILD_NUMBER` | 정수 문자열 | 실행 식별 | 현재 build 번호 | 적합 — `PipelineID` | [Environment variables](https://go-vela.github.io/docs/reference/environment/variables) |
| `VELA_BUILD_EVENT` | 이벤트명 (`push`, `pull_request` 등) | 상태·컨텍스트 | build를 시작한 event | 보조 — `Trigger` | [Environment variables](https://go-vela.github.io/docs/reference/environment/variables) |
| `VELA_PULL_REQUEST` | PR event에서 `1` | 상태·컨텍스트 | 현재 실행의 PR 여부·식별자 | 적합 — `1`일 때 `PullRequest=true`, 값은 `PullRequestID` | [Environment variables](https://go-vela.github.io/docs/reference/environment/variables) |

`VELA_PULL_REQUEST`는 Vela가 pull request event에서 정해진 `1`로 설정하는 마커입니다. token·workspace secret은 읽지 않습니다.

## 공식 문서

- [Environment variables](https://go-vela.github.io/docs/reference/environment/variables)
- [Vela 공식 저장소](https://github.com/go-vela/vela)
