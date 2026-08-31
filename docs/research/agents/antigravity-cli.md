---
title: Google Antigravity CLI
slug: antigravity-cli
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식적인 범용 자식 프로세스 마커가 없어 실제 shell command 환경 확인이 필요함
---

# Google Antigravity CLI

Antigravity CLI는 Antigravity 2.0과 동일한 에이전트 하네스를 사용하는 터미널용 TUI이며 실행 명령은 `agy`이다. Gemini CLI에서 이전하는 사용자를 위한 공식 마이그레이션 경로가 제공된다.

공식 문서와 공식 저장소의 변경 기록에서 아래 환경변수를 확인했지만, CLI가 일반 셸 명령의 자식 프로세스에 자동 주입하는 실행 식별 환경변수는 확인하지 못했다. 따라서 현재 `runby`는 Antigravity CLI를 환경변수만으로 자동 감지하지 않는다.

## 공식 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GOOGLE_GEMINI_BASE_URL` | URL 문자열 | 설정 | Gemini 호환 사용자 지정 endpoint 지정 | 부적합 — 실행 전 provider 설정이며 자식 프로세스 기원을 나타내지 않음 | [Installation & Auth](https://antigravity.google/docs/cli/install/#point-the-cli-to-a-custom-endpoint) |
| `AGY_CLI_DISABLE_AUTO_UPDATE` | boolean (`true`) | 설정 | CLI의 background self-update 비활성화 | 부적합 — 셸 profile에 지속할 수 있는 사용자 설정 | [CLI Troubleshooting](https://antigravity.google/docs/cli/troubleshooting/#resolve-self-updater-locks-and-failures) |
| `ANTIGRAVITY_MIC` | `host:port` 문자열 | 설정 | SSH 원격 CLI의 음성 입력 서버 주소 | 부적합 — 사용자가 `agy` 시작 전에 지정 | [Voice over SSH](https://antigravity.google/docs/cli/commands/voice/#3-start-the-cli-with-antigravity-mic) |
| `AGY_CLI_HIDE_LOGO` | 설정 여부 | 설정 | 시작 banner의 logo art 숨김 | 부적합 — 표시 설정이며 실행 마커가 아님 | [공식 changelog 1.1.19](https://github.com/google-antigravity/antigravity-cli/blob/556846a4bb94117222f53846896c7eb0d645307e/CHANGELOG.md#L48-L51) |
| `AGY_CLI_DISABLE_ESCAPE_SEQUENCE_OPTIMIZATIONS` | 설정 여부 | 설정 | renderer의 dirty-rectangle 및 diff 최적화 우회 | 부적합 — renderer 설정 | [공식 changelog 1.1.19](https://github.com/google-antigravity/antigravity-cli/blob/556846a4bb94117222f53846896c7eb0d645307e/CHANGELOG.md#L48-L51) |
| `AGY_CLI_CMD_OUTPUT_PERCENTAGE` | 백분율 정수 | 설정 | TUI 명령 출력의 최대 높이를 터미널 높이 비율로 제한 | 부적합 — 표시 설정 | [공식 changelog](https://github.com/google-antigravity/antigravity-cli/blob/556846a4bb94117222f53846896c7eb0d645307e/CHANGELOG.md#L427) |
| `AGY_CLI_DISABLE_LATEX` | 설정 여부 | 설정 | LaTeX 수식 rendering 비활성화 | 부적합 — 표시 설정 | [공식 changelog](https://github.com/google-antigravity/antigravity-cli/blob/556846a4bb94117222f53846896c7eb0d645307e/CHANGELOG.md#L532) |
| `AGY_CLI_HIDE_ACCOUNT_INFO` | 설정 여부 | 설정 | header에서 email과 plan tier 숨김 | 부적합 — privacy/UI 설정 | [공식 changelog](https://github.com/google-antigravity/antigravity-cli/blob/556846a4bb94117222f53846896c7eb0d645307e/CHANGELOG.md#L558) |

## 실행 주체 감지에 관한 결론

- `AGY_CLI_*`와 `ANTIGRAVITY_MIC`은 Antigravity CLI에 특화된 설정 단서지만 사용자가 부모 셸에 미리 설정하므로 현재 프로세스를 CLI가 실행했다는 증거가 아니다.
- `GOOGLE_GEMINI_BASE_URL`은 provider 설정이므로 Antigravity CLI 외의 프로세스에서도 나타날 수 있다.
- 공식 자식 프로세스 실행 마커가 추가되기 전까지 환경변수 기반 자동 감지는 하지 않는다.
- hooks와 status line이 제공하는 conversation 상태는 환경변수가 아니라 JSON 입력이므로 이 패키지의 감지 대상이 아니다.

## Gemini CLI 전환 상태

Google의 공식 발표에 따르면 Gemini CLI는 2026-06-18부터 Google AI Pro·Ultra 및 무료 개인 계정 요청 제공을 중단했으며, 개인 사용자의 공식 터미널 제품 경로는 Antigravity CLI로 전환되었다.

## 공식 문서와 소스

- [Antigravity CLI Overview](https://antigravity.google/docs/cli/overview/)
- [Gemini CLI Migration](https://antigravity.google/docs/cli/gcli-migration/)
- [Antigravity CLI 공식 저장소](https://github.com/google-antigravity/antigravity-cli)
- [Gemini CLI 서비스 전환 공지](https://github.com/google-gemini/gemini-cli/discussions/28017)
