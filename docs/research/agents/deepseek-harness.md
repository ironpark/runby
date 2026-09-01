---
title: DeepSeek Harness
slug: deepseek-harness
research_date: 2026-09-02
open_source: true
repository: https://github.com/deepseek-ai/deepseek-harness
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 shell-env와 bash executor가 매 모델 셸 호출에 DSH_SHELL을 주입하고 상속된 DSH_*를 제거한다고 명시함
---

# DeepSeek Harness

## 요약

DeepSeek Harness(`dsh`)는 DeepSeek AI가 개발한 오픈소스 agent harness다. 공식 `shell-env` 플러그인은 모델이 요청한 bash·PowerShell 셸 호출마다 Harness가 관리하는 `DSH_*` 환경을 만들고, `DSH_SHELL=1`을 공통으로 설정한다. 이 값은 제품이 자신이 실행하는 셸에 직접 붙이는 전용 실행 마커이므로 `definite`로 채택한다.

`model_source`는 제품의 기본 성격을 나타내므로 `first-party`로 분류한다. 설정 가능한 provider/model이 있다는 사실은 실제 응답 모델과 제품 분류가 다를 수 있다는 기존 `runby` 규칙에 해당한다.

## 환경변수

| 변수 | 값/형태 | 분류 | 판정 |
|---|---|---|---|
| `DSH_SHELL` | `1` | 실행 식별 | **적합 — definite**. 모든 모델 shell call의 관리 환경에 설정 |
| `DSH_SESSION_ID` | 세션 ID | 컨텍스트 | agent 호출에만 존재하며 `SessionID`로 수집. 단독으로는 감지 근거가 아님 |
| `DSH_SESSION_JSONL` | JSONL 경로 | 컨텍스트 | active persistence backend가 위치를 제공할 때 `Extra["deepseek-harness.session_jsonl"]`로 수집 |
| `DSH_HOME` | Harness home 경로 | 설정/컨텍스트 | 부적합 — config 또는 ambient 환경에서 정해질 수 있어 감지 근거로 사용하지 않음 |

## 상속 격리와 신뢰도

DeepSeek Harness의 subprocess 계층은 부모 환경에서 모든 `DSH_*`를 제거한 뒤 명시적인 `dshEnv` 스냅샷을 병합한다. bash executor도 이 스냅샷을 명시적 환경으로 전달한다. 따라서 과거 Harness나 중첩된 다른 프로세스가 남긴 `DSH_SHELL`을 그대로 믿는 구조가 아니라 현재 셸 호출의 값을 재생성한다.

이 근거는 `runby`가 환경변수 값 자체를 검증할 수 있다는 뜻은 아니다. 호출자가 임의로 `DSH_SHELL=1`을 넣은 환경은 여전히 위조할 수 있으므로, 결과의 일반적인 환경 스냅샷 한계는 유지된다. 다만 공식 실행 경계가 상속된 namespace를 제거한다는 점에서 제품의 정상 실행에서는 definite가 적절하다.

## 실행 파일과 감지 규칙

- 공식 CLI manifest가 `dsh` 실행 파일을 `lib/bin.js`에 매핑하므로 조상 프로세스 확증 이름으로 `dsh`를 등록한다.
- `DSH_SHELL=1`이 있을 때만 DeepSeek Harness로 판정한다.
- `DSH_SESSION_ID`는 `SessionID`로, `DSH_SESSION_JSONL`은 선택적 `Extra`로 수집한다.
- `DSH_HOME`이나 세션 변수만 있는 환경은 감지하지 않는다.

## 공식 소스·문서

- [DeepSeek Harness README — DeepSeek AI의 오픈소스 agent harness와 `dsh`](https://github.com/deepseek-ai/deepseek-harness#readme)
- [`shell-env` README — 모든 모델 shell call의 DSH_* 환경](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/shell/shell-env/README.md#what-every-shell-call-receives)
- [`ShellEnvRegistry.collect()` — DSH_SHELL·DSH_SESSION_ID 생성](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/shell/shell-env/src/index.ts#L152-L159)
- [`shell-env` persistence contributor — DSH_SESSION_JSONL](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/shell/shell-env/src/index.ts#L201-L215)
- [`bash-local` — dshEnv를 명시적 child env로 병합](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/shell/bash-local/src/index.ts#L195-L199)
- [`subprocess` — 상속된 DSH_* 제거](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/subprocess/subprocess/src/index.ts#L46-L65)
- [`apps/cli/package.json` — `dsh` bin](https://github.com/deepseek-ai/deepseek-harness/blob/master/apps/cli/package.json#L14-L16)
- [config catalog — provider/model 설정](https://github.com/deepseek-ai/deepseek-harness/blob/master/docs/config-catalog.md#L2417-L2420)
