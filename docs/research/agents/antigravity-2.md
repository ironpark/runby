---
title: Google Antigravity 2.0
slug: antigravity-2
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_orchestrator
model_source: first-party
executes_agents:
  - antigravity-agent
runtime_test_required: true
runtime_test_reason: sidecar 전용 신호와 일반 shell command 및 hook의 차이를 실제 실행으로 확인해야 함
---

# Google Antigravity 2.0

Antigravity 2.0은 프로젝트, 여러 workspace와 worktree, 병렬 local subagent를 관리하는 standalone desktop agent command center이다. Antigravity CLI와 같은 agent harness 및 공통 설정을 사용하지만 별도의 제품 표면이므로 환경변수 감지 문서도 분리한다.

공식 문서에서 일반 agent shell command 전체에 주입되는 실행 식별자는 확인되지 않았다. 현재 확인되는 강한 신호는 Antigravity 2.0이 lifecycle을 관리하는 sidecar 프로세스에 제공하는 `ANTIGRAVITY_EXECUTABLE_DATA_DIR`이다.

## 공식 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ANTIGRAVITY_EXECUTABLE_DATA_DIR` | 절대 디렉터리 경로 문자열 | 실행 식별/상태 | 현재 sidecar 전용 persistent data 디렉터리. Antigravity가 sidecar를 시작할 때 제공 | 적합 — Antigravity 2.0 관리형 sidecar라는 강한 신호. 일반 agent shell command에는 적용되지 않음 | [Sidecars — Runtime Data](https://antigravity.google/docs/sidecars/#runtime-data) |

## 감지 범위와 주의사항

- `ANTIGRAVITY_EXECUTABLE_DATA_DIR`가 비어 있지 않으면 `Antigravity 2.0 / sidecar`로 감지한다.
- 이 변수의 부재는 Antigravity 2.0이 아니라는 뜻이 아니다. 일반 agent shell command, hook, MCP server에는 공통 주입된다는 공식 보장이 없다.
- sidecar 설정의 `env` 필드는 사용자가 임의 환경변수를 추가하는 기능이므로 해당 사용자 변수는 Antigravity 실행 식별자로 취급하면 안 된다.
- hooks가 제공하는 conversation ID, transcript path, model, workspace 등의 상태는 stdin JSON 계약이며 환경변수가 아니다.
- Antigravity SDK의 `GOOGLE_GENAI_USE_VERTEXAI`, `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`은 별도 SDK 표면의 provider 설정이므로 Antigravity 2.0 감지 목록에서 제외한다.

## 공식 문서

- [Google Antigravity Home](https://antigravity.google/docs/home)
- [Antigravity 2.0 Getting Started](https://antigravity.google/docs/getting-started/)
- [Antigravity 2.0 Sidecars](https://antigravity.google/docs/sidecars/)
- [Antigravity Hooks](https://antigravity.google/docs/hooks/)
