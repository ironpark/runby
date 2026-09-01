---
title: TeamCity
slug: teamcity
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: JetBrains 공식 Predefined Build Parameters 문서가 TEAMCITY_VERSION을 빌드 환경 판정 변수로 직접 설명하고 빌드·구성 ID를 함께 정의하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# TeamCity

TeamCity는 빌드 프로세스에 서버·에이전트의 predefined build parameter를 환경변수로 전달합니다. `TEAMCITY_VERSION`은 공식 문서가 “build runs within TeamCity”를 판별하는 용도로 직접 설명하는 전용 마커입니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `TEAMCITY_VERSION` | 버전 문자열 | 실행 식별 | TeamCity 서버 버전 | 적합 — TeamCity 빌드임을 판정하는 공식 신호 | [Predefined build parameters](https://www.jetbrains.com/help/teamcity/predefined-build-parameters.html) |
| `BUILD_ID` | 문자열 | 실행 식별 | 현재 빌드 내부 ID | 적합 — TeamCity 마커를 확인한 뒤 pipeline 식별자로 사용 | [Predefined build parameters](https://www.jetbrains.com/help/teamcity/predefined-build-parameters.html) |
| `BUILD_NUMBER` | 문자열 | 실행 식별 | 사람이 읽는 빌드 번호 | 보조 신호 — TeamCity가 제공하지만 Jenkins도 같은 일반 이름을 사용함 | [Predefined build parameters](https://www.jetbrains.com/help/teamcity/predefined-build-parameters.html) |
| `TEAMCITY_BUILDCONF_NAME` | 문자열 | 상태·컨텍스트 | 빌드 구성 이름 | 보조 신호 — job 이름으로 사용 | [Predefined build parameters](https://www.jetbrains.com/help/teamcity/predefined-build-parameters.html) |

TeamCity에는 이 환경변수 목록에서 PR 전용 마커를 확인하지 못했으므로 `PullRequest`는 항상 false로 둡니다. `BUILD_ID`·`BUILD_NUMBER`는 `TEAMCITY_VERSION`이 먼저 맞은 경우에만 읽어 일반 변수 단독 오탐을 막습니다.

## 공식 문서

- [List of predefined build parameters](https://www.jetbrains.com/help/teamcity/predefined-build-parameters.html)
- [Configuring build parameters](https://www.jetbrains.com/help/teamcity/configuring-build-parameters.html)
