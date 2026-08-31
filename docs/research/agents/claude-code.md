---
title: Claude Code
slug: claude-code
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 문서가 직접 자식과 넓은 실행 컨텍스트의 주입 변수를 명시함
---

# Claude Code

## 요약

Claude Code는 조사 대상 중 자식 프로세스에 가장 풍부한 실행 컨텍스트를 제공한다. `CLAUDE_CODE_CHILD_SESSION=1`은 Claude Code 자체가 Bash·PowerShell·Monitor 도구, hook, status line을 직접 실행했다는 가장 강한 신호다. `CLAUDECODE=1`은 stdio MCP와 IDE 통합 터미널까지 포함하는 더 넓은 신호다. `CLAUDE_CODE_SESSION_ID`, `CLAUDE_PID`, 원격 세션 변수로 세션과 부모 프로세스를 보강할 수 있다.

환경변수는 스냅샷이며 상속·위조될 수 있다. 따라서 여기서 말하는 상태는 “어떤 Claude Code 실행 컨텍스트에서 생성되었는가”이지 Agent가 현재 작업 중이거나 살아 있다는 보장은 아니다.

Anthropic은 별도의 [환경변수 전체 레퍼런스](https://code.claude.com/docs/en/env-vars)를 제공하며 기능 플래그를 매우 자주 추가한다. 아래에서는 **자동 주입되는 실행·세션 변수는 모두 개별 기입**하고, 사용자가 설정하는 방대한 기능 플래그는 감지 의미가 같은 계열끼리 묶어 정리한다.

## 자동 주입되는 실행 식별과 세션 컨텍스트

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CLAUDE_CODE_CHILD_SESSION` | `1` | 실행 식별 | Bash·PowerShell·Monitor 도구, hook, status line 자식에 Claude Code가 직접 설정. stdio MCP에는 설정하지 않음 | **적합** — IDE 확장이 설정하지 않으므로 직접 자식/중첩 세션을 가장 신뢰성 있게 구분한다. v2.1.172 이상. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDECODE` | `1` | 실행 식별 | Claude Code의 도구·tmux·hook·status line·stdio MCP 자식과 IDE 통합 터미널에 설정 | **적합** — Claude Code 계열 실행 환경을 넓게 포착한다. 단, IDE 터미널에서도 설정되어 직접 자식 여부는 구분하지 못한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_SESSION_ID` | 세션 ID 문자열 | 상태/컨텍스트 | Bash·PowerShell·hook·stdio MCP 자식에 현재 세션 ID 설정; `/clear` 때 갱신될 수 있음 | **보조** — 주 실행 마커 확인 후 세션 상관관계에 유용하다. 단독으로는 상속·잔존 가능성이 있다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_PID` | PID 정수 | 상태/컨텍스트 | Bash·PowerShell·hook 자식에 Claude Code 자신의 PID 설정 | **보조** — 부모 프로세스 존재·실행 파일을 별도 확인하는 강한 보강 신호다. PID 재사용과 PID namespace를 고려해야 한다. v2.1.214 이상. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_EFFORT` | `low`, `medium`, `high`, `xhigh`, `max` | 상태/컨텍스트 | 지원 모델 사용 시 Bash 도구와 hook 자식에 실제 적용 중인 effort를 자동 설정 | **보조** — 런타임 값이지만 지원 모델에서만 존재하므로 실행 식별 신호로는 불완전하다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_BRIDGE_SESSION_ID` | `session_...` 문자열 | 상태/컨텍스트 | Remote Control 연결 중 Bash·hook 자식에 해당 원격 세션 ID 설정하고 연결 종료 시 제거 | **보조** — Remote Control 활성 컨텍스트와 세션 링크를 제공하지만 모든 Claude 실행에 존재하지 않는다. v2.1.199 이상. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_REMOTE` | `true` | 상태/컨텍스트 | Claude Code 클라우드 세션에서 자동 설정 | **보조** — 주 신호와 함께 로컬/클라우드 실행 위치를 구분한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_REMOTE_SESSION_ID` | 세션 ID 문자열 | 상태/컨텍스트 | 클라우드 세션의 현재 세션 ID | **보조** — 클라우드 세션 상관관계에 유용하나 로컬 실행에는 없다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_MESSAGING_SOCKET` | Unix socket/pipe 경로 | 상태/컨텍스트 | inbox socket을 바인딩한 세션이 hook과 Bash 자식에 자신의 socket 경로를 설정 | **보조** — 세션별 통신 endpoint이지만 기능 활성 시에만 존재한다. v2.1.224 이상. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_PROVIDER_MANAGED_BY_HOST` | 비어 있지 않은 문자열 | 상태/컨텍스트 | Claude Code를 임베드한 호스트가 provider·모델 라우팅을 관리함을 표시 | **보조** — 임베디드 호스트 컨텍스트만 나타내며 Claude Code 직접 실행의 증거로는 부족하다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |

## 상태와 실행 모드

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CLAUDE_CODE_RESTRICTED` | `1` | 상태/컨텍스트 | 제한 모드로 세션 시작 | **보조** — 주 감지 후 의도된 제한 상태를 설명한다. 실제 OS 격리는 별도 검증 대상이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_SAFE_MODE` | `1` | 상태/컨텍스트 | CLAUDE.md·skills·plugins·hooks·MCP 등 사용자 확장 없이 문제 해결 모드로 시작 | **보조** — 직접 자식에게 상속되므로 모드 단서로 쓸 수 있다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_SIMPLE` | `1` | 상태/컨텍스트 | 최소 시스템 프롬프트와 Bash/Read/Edit 중심의 bare 모드 | **보조** — 실행 모드 보강 정보이며 단독 감지에는 부적합하다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` | `1` | 상태/컨텍스트 | 자식 환경의 민감한 provider 값을 제거하고 Linux에서는 PID namespace 격리도 적용 | **보조** — 보안 모드 단서다. 값이 없다고 기능이 반드시 꺼졌다고 단정하지 않는다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_USE_BEDROCK` | `1` | 설정 | Amazon Bedrock provider 사용 | **보조** — 주 신호 뒤 provider 컨텍스트에만 사용한다. 사용자가 미리 설정하므로 단독 감지는 부적합하다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_USE_MANTLE` | `1` | 설정 | Amazon Bedrock Mantle endpoint 사용 | **보조** — provider 컨텍스트에만 사용한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_USE_VERTEX` | `1` | 설정 | Google Cloud Vertex AI 사용 | **보조** — provider 컨텍스트에만 사용한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_USE_FOUNDRY` | `1` | 설정 | Microsoft Foundry 사용 | **보조** — provider 컨텍스트에만 사용한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_USE_ANTHROPIC_AWS` | `1` | 설정 | Claude Platform on AWS 사용 | **보조** — provider 컨텍스트에만 사용한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_EFFORT_LEVEL` | `low`, `medium`, `high`, `xhigh`, `max`, `auto` | 상태/컨텍스트 | 지원 모델의 effort 수준 재정의 | **보조** — 현재 실행 설정을 설명할 뿐 실행 주체는 증명하지 않는다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `ANTHROPIC_MODEL` | 모델 ID/별칭 문자열 | 설정 | 사용할 모델 선택 | **보조** — 모델 컨텍스트로만 사용하며 여러 Anthropic 클라이언트가 공유할 수 있다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_SUBAGENT_MODEL` | 모델 ID/별칭 문자열 | 설정 | subagent 모델 선택 | **보조** — subagent 구성 단서이나 현재 프로세스가 subagent라는 뜻은 아니다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` | `1` | 디버그/실험 | 실험적 Agent Teams 활성화 | **보조** — 기능 사용 가능성만 나타내며 실제 팀/하위 Agent 실행 여부는 알 수 없다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_FORK_SUBAGENT` | `1` 또는 `0` | 설정 | forked subagent 모드 활성·비활성 제어 | **보조** — 기능 설정이지 현재 프로세스 역할 표지가 아니다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_FORCE_SESSION_PERSISTENCE` | `1` | 설정 | 중첩 세션으로 분류되어도 transcript/history/agent 등록을 유지 | **보조** — `CLAUDE_CODE_CHILD_SESSION` 해석 예외를 설명한다. 단독 감지는 부적합하다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_EXIT_AFTER_STOP_DELAY` | 밀리초 정수 | 설정 | SDK 자동화에서 query loop idle 후 종료 지연 | **보조** — 자동화 실행 컨텍스트의 약한 단서다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_RETRY_WATCHDOG` | `1` | 상태/컨텍스트 | unattended 세션에서 용량 오류 등을 장시간 재시도 | **보조** — CI/원격 worker 의도 상태를 나타내지만 실행 식별은 아니다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_RESUME_INTERRUPTED_TURN` | `1`/`0` | 상태/컨텍스트 | 중단된 turn 자동 재개 | **보조** — SDK·재시작 컨텍스트 단서다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_RESUME_INTERRUPTED_TURN_MAX_AGE_MS` | 밀리초 정수 | 설정 | 자동 재개할 transcript의 최대 나이 | **부적합** — 재개 정책 설정이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_RESUME_PROMPT` | 문자열 | 설정 | 중단 세션 재개 시 주입할 메시지 | **부적합** — 자동화 동작 설정이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CONFIG_DIR` | 디렉터리 경로 | 설정 | 기본 `~/.claude` 구성·세션 저장 위치 변경 | **부적합** — 설치/프로필 구성만 나타내며 경로는 개인 정보가 될 수 있다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_PROJECT_DIR_NAME` | 디렉터리 이름 문자열 | 설정 | `CLAUDE_CONFIG_DIR` 아래 프로젝트 transcript/auto-memory 디렉터리 이름 지정 | **부적합** — 저장 위치 구성이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_ENV_FILE` | 셸 스크립트 경로 | 설정 | 각 Bash 명령 전에 실행하여 환경을 유지; 일부 hook이 동적으로 채울 수 있음 | **부적합** — 실행 식별자도 세션 상태도 아니다. 경로는 외부 출력하지 않는 편이 안전하다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |

## 디버그·관측과 기타 공식 변수 계열

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `DEBUG` | `1`, `true`, `yes`, `on` | 디버그/실험 | Claude Code debug mode 활성화 | **부적합** — 범용 변수라 오탐 가능성이 높다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_DEBUG_LOGS_DIR` | 파일 경로 | 디버그/실험 | debug 로그 파일 경로 변경 | **부적합** — 이름과 달리 디렉터리가 아닌 파일 경로이며 실행 식별자는 아니다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_DEBUG_LOG_LEVEL` | `verbose`, `debug`, `info`, `warn`, `error` | 디버그/실험 | debug 로그 최소 레벨 | **부적합** — 진단 설정이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_ENABLE_TELEMETRY` | `1` | 상태/컨텍스트 | OpenTelemetry 수집 활성화 | **보조** — 감지 뒤 관측 설정을 설명할 수 있으나 실행 주체를 증명하지 않는다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `DISABLE_TELEMETRY` / `DO_NOT_TRACK` | 비어 있지 않은 값 / boolean | 설정 | 텔레메트리 및 일부 feature-flag fetch opt-out | **부적합** — 여러 도구가 공유하는 범용 설정이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_OTEL_*`, `OTEL_*` | boolean·문자열·URL·밀리초 | 디버그/실험 | exporter, header, flush, log-content, metrics 속성 등 OpenTelemetry 세부 설정 | **부적합** — 관측 구성 계열이다. 민감한 prompt/tool/API body 로깅 플래그가 있으므로 값 수집을 피한다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `ANTHROPIC_BASE_URL`, `ANTHROPIC_*_BASE_URL`, `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` | URL·호스트 문자열 | 설정 | API/provider/proxy endpoint 라우팅 | **부적합** — 네트워크 구성이고 값이 내부 인프라를 노출할 수 있다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `ANTHROPIC_DEFAULT_*_MODEL`, `ANTHROPIC_CUSTOM_MODEL_OPTION*` | 모델 ID·표시명·capability 문자열 | 설정 | 기본·사용자 지정 model picker 항목과 capability 설정 | **부적합** — 모델 구성 계열이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_DISABLE_*`, `DISABLE_*` | 주로 boolean/존재 여부 | 설정 | 기능, 트래픽, UI, command, memory, workflow, cache 등을 비활성화 | **부적합** — 기능 플래그 계열이며 일부는 `0`도 활성으로 해석하므로 일반 boolean 파서로 감지하지 않는다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_ENABLE_*`, `ENABLE_*` | 주로 boolean | 디버그/실험 | tasks, telemetry, gateway discovery, streaming, prompt cache 등 기능 활성화 | **부적합** — 기능 플래그 계열이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `API_TIMEOUT_MS`, `BASH_*`, `CLAUDE_*TIMEOUT*`, `MCP_*TIMEOUT*`, `TASK_MAX_OUTPUT_LENGTH` | 정수 | 설정 | API·shell·stream·MCP·subagent timeout 및 출력 한도 | **부적합** — 성능/실행 정책 설정 계열이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_PLUGIN_*`, `MCP_*` | 경로·boolean·정수·protocol 문자열 | 설정 | plugin cache/seed/update와 MCP 연결·discovery 설정 | **부적합** — 확장 시스템 구성 계열이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `CLAUDE_CODE_SHELL`, `CLAUDE_CODE_SHELL_PREFIX`, `CLAUDE_CODE_GIT_BASH_PATH` | 실행 파일 경로·명령 prefix | 설정 | shell 선택·감사 wrapper·Windows Git Bash 지정 | **부적합** — 실행 환경 구성이고 값이 민감한 로컬 경로를 포함할 수 있다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |
| `VERTEX_REGION_*`, `ANTHROPIC_SMALL_FAST_MODEL_AWS_REGION` | 리전 문자열 | 설정 | provider별 모델 리전 재정의 | **부적합** — cloud provider 설정 계열이다. | [Claude Code environment variables](https://code.claude.com/docs/en/env-vars#variables) |

## 권장 감지 규칙

1. `CLAUDE_CODE_CHILD_SESSION == "1"`이면 Claude Code가 직접 만든 자식 실행으로 판정한다.
2. 그렇지 않고 `CLAUDECODE == "1"`이면 Claude Code 계열 컨텍스트로 판정하되 IDE 통합 터미널 또는 stdio MCP일 수 있음을 표시한다.
3. `CLAUDE_CODE_SESSION_ID`와 `CLAUDE_PID`는 주 신호가 확인된 뒤 상관관계 및 선택적 부모 생존 확인에 사용한다.
4. `CLAUDE_CODE_REMOTE`, `CLAUDE_CODE_REMOTE_SESSION_ID`, `CLAUDE_CODE_BRIDGE_SESSION_ID`로 클라우드·Remote Control 컨텍스트를 보강한다.
5. `CLAUDE_CODE_DISABLE_*` 같은 설정 변수를 Claude Code 실행 신호로 사용하지 않는다. 사용자가 셸 프로필이나 `settings.json`에 영구 설정할 수 있기 때문이다.

## 공식 문서

- [Claude Code 환경변수 전체 레퍼런스](https://code.claude.com/docs/en/env-vars)
- [Claude Code 개요](https://code.claude.com/docs/en/overview)
- [Claude Code 설정](https://code.claude.com/docs/en/settings)
- [Claude Code hooks](https://code.claude.com/docs/en/hooks)
