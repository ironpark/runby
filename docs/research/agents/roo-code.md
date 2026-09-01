---
title: Roo Code
slug: roo-code
research_date: 2026-09-02
open_source: true
repository: https://github.com/RooCodeInc/Roo-Code
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 VS Code 터미널 생성 코드에서 ROO_ACTIVE 주입을 직접 확인했으며 터미널 소유 신호의 한계는 소스 구조로 판정 가능
---

# Roo Code

## 요약

Roo Code는 여러 모델 공급자를 연결하는 VS Code 확장형 에이전트 하네스다. `Terminal.getEnv()`가 `ROO_ACTIVE=true`를 만들고, 생성자가 그 환경을 VS Code 터미널 생성 API에 넘긴다. 즉 이 값은 Roo Code가 만든 터미널을 식별하는 신호다.

## 신뢰도를 `probable`로 낮춘 이유

`ROO_ACTIVE`는 개별 프로세스를 실행하는 순간의 환경에 붙은 것이 아니라 **터미널 소유자가 만든 환경**에 붙는다. 사용자가 같은 Roo Code 터미널에 직접 입력하면 그 명령도 같은 값을 상속받는다. 따라서 “Roo Code가 만든 터미널에서 실행되었다”는 것은 확인하지만 “Roo Code가 이 명령을 요청했다”는 것까지 증명하지는 못한다.

이 구조는 Cline의 `CLINE_ACTIVE`와 같다. 그래서 `ConfidenceProbable`로 보고하며, Roo Code는 VS Code 확장이므로 안전하게 매칭할 별도 실행 파일도 등록하지 않는다.

## 환경변수

| 변수 | 값 | 분류 | 판정 |
|---|---|---|---|
| `ROO_ACTIVE` | `true` | 터미널 소유 식별 | **적합 — probable**. Roo Code가 만드는 터미널에 주입되지만 사람의 입력과 구분하지 못함 |

`ROO_ACTIVE`만으로 definite로 올리거나, 다른 제품이 이 값을 상속받았다고 가정하지 않는다. 값은 evidence에 복사하지 않고 변수 이름만 기록한다.

## 실행 파일과 감지 규칙

- Roo Code는 VS Code 확장으로 실행되므로 `Executables`는 비워 둔다.
- `ROO_ACTIVE=true`가 있으면 Roo Code 터미널의 probable 감지로 보고한다.
- 값이 없거나 false이면 감지하지 않는다.

## 공식 소스

- [`Terminal` 생성 — Roo 환경을 VS Code 터미널에 전달](https://github.com/RooCodeInc/Roo-Code/blob/main/src/integrations/terminal/Terminal.ts#L15-L20)
- [`Terminal.getEnv()` — ROO_ACTIVE=true 설정](https://github.com/RooCodeInc/Roo-Code/blob/main/src/integrations/terminal/Terminal.ts#L153-L161)
- [Cline 조사 문서 — 동일한 터미널 소유 신호의 판정](cline.md)
