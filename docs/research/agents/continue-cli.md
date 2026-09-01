---
title: Continue CLI
slug: continue-cli
research_date: 2026-09-02
open_source: true
repository: https://github.com/continuedev/continue
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 runTerminalCommand.ts의 양쪽 spawn 경로가 부모 환경과 색상 변수만 사용하며 Continue 실행 마커를 넣지 않음
---

# Continue CLI

## 요약

**드라이버 없음 — 음성 결과.** Continue의 `runTerminalCommand.ts`는 로컬 명령을 `child_process.spawn`으로 실행한다. `env`가 전달되는 현재 경로도 `getColorEnv()`가 `process.env`를 펼친 뒤 터미널 색상 관련 변수만 추가한다. Continue가 자식 프로세스에 “이 명령은 Continue가 요청했다”는 전용 변수를 넣는 코드는 확인되지 않았다.

## 조사 경위

streaming과 non-streaming 양쪽 spawn을 확인했다. 두 경로 모두 `cwd`와 `getColorEnv()`만 전달하며, `CONTINUE_*` 세션·실행 마커는 없다. `FORCE_COLOR`, `COLORTERM`, `TERM`, `CLICOLOR`, `CLICOLOR_FORCE`는 출력 표시를 위한 일반 터미널 변수라 제품 식별에 사용할 수 없다.

| 조사 대상 | 확인 내용 | 판정 |
|---|---|---|
| `getColorEnv()` | 부모 환경을 복사하고 색상 변수만 추가 | 실행 주체 마커 아님 |
| streaming `spawn` | `cwd`, `env: getColorEnv()`만 사용 | Continue 고유 마커 없음 |
| non-streaming `spawn` | 같은 환경 구성 사용 | Continue 고유 마커 없음 |

## 결론

Continue가 자식 셸에 전용 marker를 주입한다는 공식 소스 근거가 없어 감지하지 않는다. 색상·셸 설정과 부모 환경 상속은 명령 실행 주체를 특정하지 못한다.

## 공식 소스

- [`getColorEnv()` — process.env와 색상 변수](https://github.com/continuedev/continue/blob/main/core/tools/implementations/runTerminalCommand.ts#L90-L99)
- [streaming spawn — Continue marker 없는 환경 옵션](https://github.com/continuedev/continue/blob/main/core/tools/implementations/runTerminalCommand.ts#L147-L152)
- [non-streaming spawn — 같은 환경 구성](https://github.com/continuedev/continue/blob/main/core/tools/implementations/runTerminalCommand.ts#L377-L392)
