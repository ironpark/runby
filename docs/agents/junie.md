---
title: JetBrains Junie
slug: junie
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_harness
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공개 변수는 실행 입력이며 자식 프로세스 주입 계약이 없어 실제 환경 검증이 필요함
---

# JetBrains Junie

Junie는 공식 환경변수 레퍼런스를 제공합니다. 아래 변수는 CLI 인수에 대응하는 입력 설정·작업 컨텍스트이며, Junie가 실행한 자식 프로세스에 주입하는 세션 ID나 실행 마커는 없습니다. 즉 `JUNIE_*`가 보이더라도 사용자가 Junie를 실행하기 전에 셸이나 CI에 설정한 값일 수 있습니다.

## 작업·프로젝트 및 모델

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `JUNIE_TASK` | 자연어 문자열 | 상태/컨텍스트 | 비대화형 작업 설명 (`--task`) | 보조 — Junie 실행 입력이라는 문맥은 강하지만 사용자가 미리 설정하는 값이고 자식 프로세스 주입 계약은 없음 | [Junie Environment variables — Project and task](https://junie.jetbrains.com/docs/environment-variables.html#project-and-task) |
| `JUNIE_PROMPT` | 자연어 문자열 | 상태/컨텍스트 | 대화형 TUI 시작 시 자동 제출할 초기 프롬프트 | 보조 — Junie 시작 입력이지만 실행 식별자가 아니며 민감한 프롬프트 내용을 기록해서는 안 됨 | [Junie Environment variables — Project and task](https://junie.jetbrains.com/docs/environment-variables.html#project-and-task) |
| `JUNIE_PROJECT` | 디렉터리 경로 | 상태/컨텍스트 | 작업할 프로젝트 디렉터리; 기본값은 현재 디렉터리 | 보조 — Junie 전용 프로젝트 입력이지만 사용자가 설정하며 현재 프로세스의 실행자를 보장하지 않음 | [Junie Environment variables — Project and task](https://junie.jetbrains.com/docs/environment-variables.html#project-and-task) |
| `JUNIE_MODEL` | 모델 ID 문자열 | 설정 | 사용할 LLM 모델; 미설정 시 `Default` | 부적합 — 모델 선택 설정 | [Junie Environment variables — Model selection](https://junie.jetbrains.com/docs/environment-variables.html#model-selection) |
| `JUNIE_LLM_PROVIDER` | 열거 문자열 (`openai`, `anthropic`, `google`, `xai`, `openrouter`, `copilot`, `litellm`) | 설정 | 모델 공급자 강제 선택 | 부적합 — 모델 공급자 설정 | [Junie Environment variables — Model selection](https://junie.jetbrains.com/docs/environment-variables.html#model-selection) |
| `JUNIE_EFFORT` | 열거 문자열 (`minimal`, `low`, `medium`, `high`, `xhigh`, `max`) | 설정 | 지원 모델의 추론 노력 수준 | 부적합 — 모델 실행 설정 | [Junie Model selection](https://junie.jetbrains.com/docs/junie-cli-model-selection.html) |
| `JUNIE_MODEL_DEFAULT_LOCATIONS` | 불리언 (`true`/`false`, 기본 `true`) | 설정 | 사용자·프로젝트의 기본 custom model 위치 탐색 여부 | 부적합 — 검색 경로 설정 | [Junie Environment variables — Custom model discovery](https://junie.jetbrains.com/docs/environment-variables.html#custom-model-discovery) |
| `JUNIE_MODEL_LOCATIONS` | 경로 문자열; 복수 위치 지정 가능 | 설정 | custom model profile 추가 탐색 위치 | 부적합 — 검색 경로 설정 | [Junie Environment variables — Custom model discovery](https://junie.jetbrains.com/docs/environment-variables.html#custom-model-discovery) |

## 설정 파일과 Junie 홈

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `JUNIE_CONFIG_LOCATION` | 파일 경로 문자열; 복수 지정 가능 | 설정 | 추가 `config.json` 위치 | 부적합 — 설정 파일 검색 경로 | [Junie Environment variables — Configuration files](https://junie.jetbrains.com/docs/environment-variables.html#configuration-files) |
| `JUNIE_CONFIG_DEFAULT_LOCATIONS` | 불리언 (`true`/`false`, 기본 `true`) | 설정 | 기본 사용자·프로젝트 설정 위치 로드 여부 | 부적합 — 설정 로딩 토글 | [Junie Environment variables — Configuration files](https://junie.jetbrains.com/docs/environment-variables.html#configuration-files) |
| `JUNIE_HOME` | 디렉터리 경로 | 설정 | Junie CLI 홈 디렉터리; 기본 `~/.junie` 대체 | 부적합 — 상태 저장 위치 설정이며 셸에 상시 존재할 수 있음 | [Junie Environment variables — Configuration files](https://junie.jetbrains.com/docs/environment-variables.html#configuration-files) |

## 지침·MCP·스킬·확장

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `JUNIE_GUIDELINES_FILENAME` | 파일명 또는 경로 | 설정 | 기본값 대신 사용할 `.junie/` 아래 지침 파일 선택 | 부적합 — 지침 검색 설정 | [Junie Environment variables — Guidelines file](https://junie.jetbrains.com/docs/environment-variables.html#guidelines-file) |
| `JUNIE_MCP_DEFAULT_LOCATIONS` | 불리언 (`true`/`false`, 기본 `true`) | 설정 | 기본 사용자·프로젝트 MCP 설정 위치 탐색 여부 | 부적합 — MCP 검색 설정 | [Junie Environment variables — MCP servers](https://junie.jetbrains.com/docs/environment-variables.html#mcp-servers) |
| `JUNIE_MCP_LOCATIONS` | 디렉터리 경로 문자열; 복수 지정 가능 | 설정 | 추가 MCP 설정 폴더 | 부적합 — MCP 검색 경로 설정 | [Junie Environment variables — MCP servers](https://junie.jetbrains.com/docs/environment-variables.html#mcp-servers) |
| `JUNIE_SKILL_DEFAULT_LOCATIONS` | 불리언 (`true`/`false`, 기본 `true`) | 설정 | 기본 agent skill 위치 탐색 여부 | 부적합 — skill 검색 설정 | [Junie Environment variables — Agent skills](https://junie.jetbrains.com/docs/environment-variables.html#agent-skills) |
| `JUNIE_SKILL_LOCATIONS` | 디렉터리 경로 문자열; 복수 지정 가능 | 설정 | 추가 agent skill 위치 | 부적합 — skill 검색 경로 설정 | [Junie Environment variables — Agent skills](https://junie.jetbrains.com/docs/environment-variables.html#agent-skills) |
| `JUNIE_COMMAND_DEFAULT_LOCATIONS` | 불리언 (`true`/`false`, 기본 `true`) | 설정 | 기본 custom slash command 위치 탐색 여부 | 부적합 — 명령 검색 설정 | [Junie Environment variables — Custom slash commands](https://junie.jetbrains.com/docs/environment-variables.html#custom-slash-commands) |
| `JUNIE_COMMAND_LOCATIONS` | 경로 문자열; 복수 지정 가능 | 설정 | 추가 custom slash command 위치 | 부적합 — 명령 검색 경로 설정 | [Junie Environment variables — Custom slash commands](https://junie.jetbrains.com/docs/environment-variables.html#custom-slash-commands) |
| `JUNIE_AGENT_DEFAULT_LOCATIONS` | 불리언 (`true`/`false`, 기본 `true`) | 설정 | 기본 custom agent 위치 탐색 여부 | 부적합 — agent 정의 검색 설정 | [Junie Environment variables — Custom agents](https://junie.jetbrains.com/docs/environment-variables.html#custom-agents) |
| `JUNIE_AGENT_LOCATIONS` | 경로 문자열; 복수 지정 가능 | 설정 | 추가 custom agent 위치 | 부적합 — agent 정의 검색 경로 설정 | [Junie Environment variables — Custom agents](https://junie.jetbrains.com/docs/environment-variables.html#custom-agents) |
| `JUNIE_EXTENSIONS_DEFAULT_LOCATION` | 디렉터리 경로 | 설정 | 기본 extensions 디렉터리(`~/.junie/extensions`) 대체 | 부적합 — 확장 검색 위치 설정 | [Junie Environment variables — Extensions](https://junie.jetbrains.com/docs/environment-variables.html#extensions) |
| `JUNIE_SHARE_ANONYMOUS_STATISTICS` | 불리언 (`true`/`false`) | 설정 | 현재 실행의 익명 사용 통계 공유 여부 덮어쓰기 | 부적합 — telemetry 설정이며 실행 마커가 아님 | [Junie Environment variables — Usage statistics](https://junie.jetbrains.com/docs/environment-variables.html#usage-statistics) |

## 실행 주체 감지에 관한 결론

2026-08-30 현재 공식 목록에는 `JUNIE_SESSION_ID`, `JUNIE_THREAD_ID`, `JUNIE_AGENT` 같은 자식 프로세스 실행 식별 변수가 없습니다. `JUNIE_TASK`, `JUNIE_PROMPT`, `JUNIE_PROJECT`는 Junie 실행 입력이라는 보조 정황으로만 사용할 수 있고, 단독 확정 판정에는 부적합합니다.

공식 문서는 같은 항목이 환경변수와 CLI flag 양쪽에 있으면 CLI flag가 우선한다고 명시합니다. 따라서 환경변수 값이 실제 실행 상태와 다를 수도 있습니다.

## 공식 문서

- [Junie 환경변수 전체 목록](https://junie.jetbrains.com/docs/environment-variables.html)
- [Junie CLI reference](https://junie.jetbrains.com/docs/parameters.html)
- [Junie CLI configuration](https://junie.jetbrains.com/docs/junie-cli-configuration.html)
- [Junie Model selection](https://junie.jetbrains.com/docs/junie-cli-model-selection.html)
- [Junie 문서](https://junie.jetbrains.com/docs/)
