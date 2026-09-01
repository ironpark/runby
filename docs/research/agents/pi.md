---
title: pi
slug: pi
research_date: 2026-09-02
open_source: true
repository: https://github.com/badlogic/pi-mono
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 소스가 bash·powershell 도구의 매 spawn마다 PI_SESSION_ID를 설정한다고 명시함
---

# pi

## 요약

pi(badlogic/pi-mono)는 여러 벤더의 모델에 연결하는 오픈소스 코딩 에이전트 하네스다. bash·powershell 도구가 **자신이 실행하는 모든 명령의 환경에** `PI_*` 세션 변수를 주입하며, 매 spawn마다 기존 값을 지우고 다시 설정하므로 상속된 잔존 값이 아니라 실행 마커다. 도구 가이드라인 자체가 "PI_* 환경변수로 현재 모델·세션 정보를 확인하라"고 안내한다.

주입은 `exposeSessionEnvironment` 옵션으로 제어되며 **기본값이 켜짐**이다. pi는 제품명을 바꾼 화이트라벨 배포를 지원하고 그 경우 변수 접두어가 달라질 수 있으므로, 이 감지는 stock `pi` 구성에만 적용된다.

## 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `PI_SESSION_ID` | 세션 식별자 | 실행 식별 | pi가 실행한 명령에 현재 세션을 알려주는 마커 | **적합** — 매 spawn마다 설정되는 세션 단위 실행 마커 | [bash.ts](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/src/core/tools/bash.ts) |
| `PI_SESSION_FILE` | 파일 경로 | 컨텍스트 | 세션 기록 파일 위치 | 부적합 — 세션 ID 없이 단독 설정되지 않으며 컨텍스트로만 기록 | 위와 동일 |
| `PI_PROVIDER` | 프로바이더 슬러그 | 컨텍스트 | 현재 세션을 서빙하는 모델 프로바이더 | 부적합 — 실행 증거가 아니라 라이브 모델 정보 | 위와 동일 |
| `PI_MODEL` | 모델 식별자 | 컨텍스트 | 현재 세션을 서빙하는 모델 | 부적합 — 위와 동일 | 위와 동일 |
| `PI_REASONING_LEVEL` | thinking 레벨 | 컨텍스트 | 세션의 추론 강도 설정 | 부적합 — 설정 상태 | 위와 동일 |

`PI_PROVIDER`/`PI_MODEL`은 이 패키지가 다루는 에이전트 중 드물게 **실제 서빙 중인 모델**을 광고하는 사례다. `model_source` 분류(제품 성격)는 multi-vendor로 유지하고, 라이브 정보는 `Extra`(`pi.provider`, `pi.model`)로 노출한다.

## 권장 감지 규칙

1. `PI_SESSION_ID`가 존재하고 빈 문자열이 아니면 pi가 현재 프로세스를 실행한 것으로 판정한다(definite). 값은 `SessionID`로 승격한다.
2. `PI_SESSION_FILE`·`PI_PROVIDER`·`PI_MODEL`·`PI_REASONING_LEVEL`은 감지 근거로 쓰지 않고 `Extra`로만 기록한다.
3. 실행 파일 이름은 `pi`이므로 조상 체인 확증(`AncestorPID`)에 사용한다.

## 공식 문서

- [pi coding-agent README](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/README.md)
- [bash 도구 소스 — PI_* 주입](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/src/core/tools/bash.ts)
- [powershell 도구 소스](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/src/core/tools/powershell.ts)
