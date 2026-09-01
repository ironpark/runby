---
title: GoCD
slug: gocd
research_date: 2026-09-02
open_source: true
repository: https://github.com/gocd/gocd
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: GoCD 공식 configuration reference가 GO_PIPELINE_LABEL과 pipeline·stage·agent 환경변수를 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# GoCD

GoCD agent는 pipeline task에 `GO_PIPELINE_LABEL`을 주입합니다. 이는 GoCD pipeline의 label이므로 일반 `BUILD_NUMBER`보다 제품을 직접 식별하는 강한 마커입니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `GO_PIPELINE_LABEL` | 문자열 | 실행 식별 | 현재 pipeline label | 적합 — GoCD 전용 마커이자 `PipelineID` | [Configuration reference](https://docs.gocd.org/current/configuration/configuration_reference.html) |
| `GO_PIPELINE_COUNTER` | 정수 문자열 | 실행 식별 | pipeline 실행 카운터 | 보조 — `BuildNumber` | [Configuration reference](https://docs.gocd.org/current/configuration/configuration_reference.html) |
| `GO_JOB_NAME` / `GO_STAGE_NAME` | 문자열 | 실행 식별 | 현재 job·stage 이름 | 보조 — `JobID`·`JobName` | [Configuration reference](https://docs.gocd.org/current/configuration/configuration_reference.html) |
| `GO_AGENT_HOST` | 호스트명 | 실행 식별 | task를 실행한 agent | 보조 — `Runner` | [Configuration reference](https://docs.gocd.org/current/configuration/configuration_reference.html) |

GoCD의 기본 환경변수 목록에서 PR 여부를 직접 나타내는 공통 마커는 확인하지 못했으므로 `PullRequest`는 추정하지 않습니다. 사용자 정의 환경변수와 agent secret은 읽지 않습니다.

## 공식 문서

- [Configuration reference](https://docs.gocd.org/current/configuration/configuration_reference.html)
- [GoCD 공식 저장소](https://github.com/gocd/gocd)
