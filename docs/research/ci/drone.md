---
title: Drone
slug: drone
research_date: 2026-09-02
open_source: true
repository: https://github.com/harness/drone
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Drone 공식 environment reference가 자동 주입 변수와 DRONE_BUILD_EVENT의 pull_request 값을 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Drone

Drone은 pipeline step에 `DRONE=true`와 `CI=true`를 포함한 환경변수를 자동으로 주입합니다. `DRONE`이 Drone 전용 마커이며, build·stage·step 정보를 같은 환경에서 제공하므로 `CI`는 보조 증거로만 수집합니다.

## 실행 식별과 Pull request

| 환경변수 | 값/자료형 | 종류 | 용도 | `runby` 판정 | 공식 출처 |
|---|---|---|---|---|---|
| `DRONE` | `true` | 실행 식별 | Drone pipeline 환경 표시 | 적합 — 전용 마커 | [Environment reference](https://docs.drone.io/pipeline/environment/reference/) |
| `DRONE_BUILD_NUMBER` | 정수 문자열 | 실행 식별 | pipeline build 번호 | 적합 — pipeline 식별·표시값 | [Environment reference](https://docs.drone.io/pipeline/environment/reference/) |
| `DRONE_STAGE_NUMBER` | 정수 문자열 | 실행 식별 | 현재 stage 번호 | 적합 — job 식별자 | [Environment reference](https://docs.drone.io/pipeline/environment/reference/) |
| `DRONE_BUILD_EVENT` | 이벤트명 (`push`, `pull_request` 등) | 상태·컨텍스트 | pipeline을 시작한 이벤트 | 적합 — `pull_request`이면 `PullRequest=true` | [DRONE_BUILD_EVENT](https://docs.drone.io/pipeline/environment/reference/drone-build-event/) |
| `DRONE_PULL_REQUEST` | PR 번호 문자열 | 상태·컨텍스트 | 현재 PR 번호 | 적합 — PR 이벤트에서 `PullRequestID`로 사용 | [Environment reference](https://docs.drone.io/pipeline/environment/reference/) |

`DRONE_BUILD_EVENT`가 `pull_request`이면 번호 변수가 없어도 PR 판정을 유지합니다. Evidence에는 `DRONE_BUILD_EVENT`와 실제로 제공된 `DRONE_PULL_REQUEST`의 이름만 기록합니다.

## 공식 문서

- [Environment reference](https://docs.drone.io/pipeline/environment/reference/)
- [DRONE_BUILD_EVENT](https://docs.drone.io/pipeline/environment/reference/drone-build-event/)
- [공식 Drone 저장소](https://github.com/harness/drone)
