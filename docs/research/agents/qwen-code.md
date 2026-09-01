---
title: Qwen Code
slug: qwen-code
research_date: 2026-09-02
open_source: true
repository: https://github.com/QwenLM/qwen-code
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 셸 실행 코드가 자식 환경에 QWEN_CODE를 직접 주입하고 컨텍스트 변수 전달 함수도 공개함
---

# Qwen Code

## 요약

Qwen Code는 Qwen 계열 모델을 기본 대상으로 만들어진 오픈소스 터미널 에이전트 하네스다. 제품의 기본 모델 성격은 `first-party`로 분류하지만, 공식 설정 문서상 OpenAI·Anthropic·Gemini 호환 프로토콜과 사용자 지정 OpenAI 호환 endpoint도 사용할 수 있다. endpoint를 바꿀 수 있다는 사실은 제품 분류를 바꾸지 않는다.

## 환경변수

| 변수 | 값/형태 | 분류 | 판정 |
|---|---|---|---|
| `QWEN_CODE` | `1` | 실행 식별 | **적합 — definite**. 일반 셸 실행과 PTY 실행의 자식 환경에 직접 주입 |
| `QWEN_CODE_SESSION_ID` | 세션 식별자 | 컨텍스트 | `SessionID`로 수집. 단독으로는 감지 근거가 아님 |
| `QWEN_CODE_PROJECT_DIR` | 프로젝트 경로 | 컨텍스트 | `Paths.WorkingDirectory`로 수집. 단독으로는 감지 근거가 아님 |

`shellExecutionService.ts`는 일반 자식 프로세스와 PTY 프로세스의 환경에 모두 `QWEN_CODE=1`을 넣는다. `getShellContextEnvVars()`는 현재 세션 ID와 해당 세션의 launch project directory를 `QWEN_CODE_SESSION_ID`·`QWEN_CODE_PROJECT_DIR`로 전달한다. 따라서 두 컨텍스트 변수는 마커가 아니라, `QWEN_CODE`가 성립한 뒤에만 읽는 부가 정보다.

## 실행 파일과 감지 규칙

- 실행 파일은 `qwen`으로 등록해 조상 프로세스 확증에 사용한다.
- `QWEN_CODE=1`이 있으면 Qwen Code로 판정한다.
- 세션 ID가 있으면 `SessionID`, 프로젝트 디렉터리가 있으면 `Paths.WorkingDirectory`에 담는다.
- 컨텍스트 변수만 있고 `QWEN_CODE`가 없으면 감지하지 않는다.

## 공식 소스

- [`shellExecutionService.ts` — 자식 셸 환경의 QWEN_CODE 주입](https://github.com/QwenLM/qwen-code/blob/main/packages/core/src/services/shellExecutionService.ts#L802-L816)
- [`shellExecutionService.ts` — PTY 환경의 QWEN_CODE 주입](https://github.com/QwenLM/qwen-code/blob/main/packages/core/src/services/shellExecutionService.ts#L1505-L1520)
- [`shellContextEnv.ts` — 세션·프로젝트 컨텍스트 전달](https://github.com/QwenLM/qwen-code/blob/main/packages/core/src/services/shellContextEnv.ts#L115-L140)
- [Qwen Code 모델 공급자 문서 — OpenAI 호환 endpoint](https://github.com/QwenLM/qwen-code/blob/main/docs/users/configuration/model-providers.md#openai-compatible-providers-openai)
