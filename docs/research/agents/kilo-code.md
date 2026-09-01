---
title: Kilo Code
slug: kilo-code
research_date: 2026-09-02
open_source: true
repository: https://github.com/Kilo-Org/kilocode
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: bash 도구의 ChildProcess 옵션과 환경 모델에 제품 실행 마커 주입이 없음을 소스에서 확인함
---

# Kilo Code

## 요약

**드라이버 없음 — 음성 결과.** Kilo Code의 V2 bash 도구는 명령·작업 디렉터리·timeout을 입력으로 받고 `ChildProcess.make`에 shell과 cwd를 전달하지만, Kilo Code 식별 환경변수는 추가하지 않는다. 따라서 Roo Code의 `ROO_ACTIVE`를 Kilo Code의 마커로 재사용할 근거가 없다.

## 조사 경위

`packages/core/src/tool/bash.ts`의 입력 스키마에는 환경변수 입력이 없고, 실제 child process 생성 옵션에도 `env`로 Kilo 또는 Roo 값을 넣는 코드가 없다. 같은 저장소의 process environment helper는 부모 `process.env`를 일반적으로 복사하고 Kilo 내부 서버·설정 변수만 제거한다. 이는 부모에 우연히 있던 `ROO_ACTIVE`를 Kilo가 생성·상속하는 제품 계약이 있다는 뜻이 아니다. Roo 터미널에서 Kilo를 실행했을 때 값이 남을 수 있더라도, 그 값은 Roo 터미널의 ambient marker이며 Kilo 실행의 증거가 아니다.

| 조사 대상 | 확인 내용 | 판정 |
|---|---|---|
| `packages/core/src/tool/bash.ts` | bash 입력은 command/workdir/timeout이고 child 옵션에 제품 마커 없음 | **부적합** — Kilo가 자식에 marker를 주입하지 않음 |
| `packages/opencode/src/kilocode/process/env.ts` | `process.env`와 추가 env를 복사하고 Kilo 내부 변수만 삭제 | `ROO_ACTIVE`를 Kilo 전용으로 만들거나 보장하지 않음 |
| `ROO_ACTIVE` | Roo Code가 만든 터미널의 소유 신호 | Kilo Code 근거로 사용하지 않음 |

## 결론

Kilo Code에는 공식 소스에서 확인되는 고유 실행 마커가 없고, Roo Code의 터미널 변수도 Kilo로 귀속할 수 없다. `ROO_ACTIVE`가 부모 환경에서 보이더라도 Kilo와 Roo가 중첩된 경우 어느 제품이 현재 명령을 요청했는지 구분할 수 없으므로 agent 드라이버를 추가하지 않는다.

## 공식 소스

- [`bash.ts` — 입력 스키마와 shell 실행 경계](https://github.com/Kilo-Org/kilocode/blob/main/packages/core/src/tool/bash.ts#L23-L33)
- [`bash.ts` — 제품 환경 주입 없는 ChildProcess.make](https://github.com/Kilo-Org/kilocode/blob/main/packages/core/src/tool/bash.ts#L151-L160)
- [`env.ts` — 일반 부모 환경 복사 및 Kilo 변수 정리](https://github.com/Kilo-Org/kilocode/blob/main/packages/opencode/src/kilocode/process/env.ts#L1-L14)
