---
title: Taskcluster
slug: taskcluster
research_date: 2026-09-02
open_source: true
repository: https://github.com/taskcluster/taskcluster
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Taskcluster 공식 task environment 문서가 TASK_ID와 RUN_ID를 task 실행 환경의 식별자로 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Taskcluster

Taskcluster task environment에는 `TASK_ID`와 `RUN_ID`가 함께 존재합니다. `RUN_ID`는 task 재실행 번호이고 단독 `RUN_ID`는 다른 CI에도 쓰일 수 있으므로, 두 변수를 모두 요구해 Taskcluster 판정의 증거를 강화합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `TASK_ID` | UUID 문자열 | 실행 식별 | 현재 task 식별자 | 적합 — `RUN_ID`와 함께 전용 실행 마커이자 `PipelineID` | [Environment variables](https://docs.taskcluster.net/docs/manual/design/env-vars) |
| `RUN_ID` | 0부터 시작하는 정수 | 실행 식별 | task run 번호 | 적합 — `TASK_ID`와 함께 요구; `Attempt`는 1부터로 정규화 | [Environment variables](https://docs.taskcluster.net/docs/manual/design/env-vars) |

`RUN_ID` 단독은 범용 이름이고 `BUILD_NUMBER`와 같은 일반 휴리스틱과 겹칠 수 있으므로 감지하지 않습니다. Taskcluster 공식 환경변수 목록에서 PR 실행을 광고하는 마커는 확인하지 못했습니다.

## 공식 문서

- [Task environment variables](https://docs.taskcluster.net/docs/manual/design/env-vars)
- [Taskcluster 공식 저장소](https://github.com/taskcluster/taskcluster)
