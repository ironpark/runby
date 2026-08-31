---
title: OpenCode
slug: opencode
research_date: 2026-08-30
open_source: true
repository: https://github.com/anomalyco/opencode
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: ACP 경로의 소스 기반 신호와 일반 CLI 자식 환경을 각각 확인해야 함
---

# OpenCode

## 요약

OpenCode 공식 CLI 문서는 일반 환경변수 26개와 실험 환경변수 18개를 공개한다. 대부분은 사용자가 OpenCode 시작 전에 넣는 설정값이므로 실행 주체 감지에는 적합하지 않다. `OPENCODE_CLIENT`는 client 종류를 나타내지만 일반 CLI의 기본값이 환경변수로 자식에게 반드시 주입된다고 문서화되어 있지 않다. 다만 공식 소스에서 `opencode acp`가 런타임에 `OPENCODE_CLIENT=acp`를 설정하는 것은 확인된다.

## 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
| --- | --- | --- | --- | --- | --- |
| `OPENCODE_AUTO_SHARE` | boolean | 설정 | 세션 자동 공유 | 부적합 — 사용자 입력 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_GIT_BASH_PATH` | 경로 문자열 | 설정 | Windows Git Bash 실행 파일 경로 | 부적합 — 설치 환경 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_CONFIG` | 파일 경로 문자열 | 설정 | 사용자 지정 config 파일 | 부적합 — 실행 전에 설정 | [Config custom path](https://opencode.ai/docs/config/#custom-path) |
| `OPENCODE_TUI_CONFIG` | 파일 경로 문자열 | 설정 | TUI config 파일 | 부적합 — 실행 전에 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_CONFIG_DIR` | 디렉터리 경로 문자열 | 설정 | agent·command·mode·plugin을 찾을 추가 config 디렉터리 | 부적합 — 실행 전에 설정 | [Config custom directory](https://opencode.ai/docs/config/#custom-directory) |
| `OPENCODE_CONFIG_CONTENT` | JSON 문자열 | 설정 | inline config 내용 | 부적합 — 실행 전에 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_AUTOUPDATE` | boolean | 설정 | 자동 업데이트 확인 비활성화 | 부적합 — 지속 설정 가능 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_PRUNE` | boolean | 설정 | 오래된 데이터 정리 비활성화 | 부적합 — 기능 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_TERMINAL_TITLE` | boolean | 설정 | 터미널 제목 자동 변경 비활성화 | 부적합 — UI 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_PERMISSION` | inline JSON 문자열 | 설정 | permission config | 부적합 — 정책 입력값 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_DEFAULT_PLUGINS` | boolean | 설정 | 기본 플러그인 비활성화 | 부적합 — 기능 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_LSP_DOWNLOAD` | boolean | 설정 | LSP server 자동 다운로드 비활성화 | 부적합 — 기능 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_ENABLE_EXPERIMENTAL_MODELS` | boolean | 디버그/실험 | 실험 모델 활성화 | 부적합 — 실험 플래그 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_AUTOCOMPACT` | boolean | 상태/컨텍스트 | context 자동 compaction 비활성화 | 부적합 — 컨텍스트 처리 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_CLAUDE_CODE` | boolean | 설정 | `.claude` prompt·skill 읽기 전체 비활성화 | 부적합 — 호환성 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_CLAUDE_CODE_PROMPT` | boolean | 설정 | `~/.claude/CLAUDE.md` 읽기 비활성화 | 부적합 — 호환성 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_CLAUDE_CODE_SKILLS` | boolean | 설정 | `.claude/skills` 로딩 비활성화 | 부적합 — 호환성 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_MODELS_FETCH` | boolean | 설정 | 원격 model 정보 fetch 비활성화 | 부적합 — 네트워크 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_DISABLE_MOUSE` | boolean | 설정 | TUI mouse capture 비활성화 | 부적합 — UI 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_FAKE_VCS` | 문자열 | 디버그/실험 | 테스트용 가짜 VCS provider | 부적합 — 테스트 플래그 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_CLIENT` | 문자열, 기본 client ID `cli`; ACP에서는 `acp` | 상태/컨텍스트 | OpenCode client 식별 | 보조 — 공식 소스에서 ACP가 `acp`를 런타임 설정하지만 일반 CLI 자식에 안정적으로 자동 주입된다는 보장은 없음 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables), [ACP 공식 소스](https://github.com/anomalyco/opencode/blob/dev/packages/opencode/src/cli/cmd/acp.ts#L21) |
| `OPENCODE_ENABLE_EXA` | boolean | 설정 | Exa web search tool 활성화 | 부적합 — 기능 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_ENABLE_PARALLEL` | boolean | 설정 | Parallel web search tool 활성화 | 부적합 — 기능 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_MODELS_URL` | URL 문자열 | 설정 | model config fetch URL 변경 | 부적합 — 네트워크 설정 | [CLI 환경변수](https://opencode.ai/docs/cli/#environment-variables) |
| `OPENCODE_EXPERIMENTAL` | boolean | 디버그/실험 | 실험 기능 umbrella flag | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_ICON_DISCOVERY` | boolean | 디버그/실험 | icon discovery 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_DISABLE_COPY_ON_SELECT` | boolean | 디버그/실험 | TUI copy-on-select 비활성화 | 부적합 — 실험 UI 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_BASH_DEFAULT_TIMEOUT_MS` | 정수 ms | 디버그/실험 | bash command 기본 timeout | 부적합 — 실험 실행 설정 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX` | 정수 token 수 | 디버그/실험 | LLM 응답 최대 output token | 부적합 — 실험 모델 설정 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_FILEWATCHER` | boolean | 디버그/실험 | 전체 디렉터리 file watcher 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_OXFMT` | boolean | 디버그/실험 | oxfmt formatter 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_LSP_TOOL` | boolean | 디버그/실험 | 실험 LSP tool 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_DISABLE_FILEWATCHER` | boolean | 디버그/실험 | file watcher 비활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_EXA` | boolean | 디버그/실험 | 실험 Exa 기능 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_LSP_TY` | boolean | 디버그/실험 | Python 파일용 TY LSP 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_PLAN_MODE` | boolean | 디버그/실험 | plan mode 활성화 | 부적합 — 실행 모드 입력값 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_BACKGROUND_SUBAGENTS` | boolean | 디버그/실험 | background subagent task 활성화 | 부적합 — 기능 플래그이며 활성 agent 상태를 나타내지 않음 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_EVENT_SYSTEM` | boolean | 디버그/실험 | 실험 event system 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_NATIVE_LLM` | boolean | 디버그/실험 | native LLM request path 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_PARALLEL` | boolean | 디버그/실험 | Parallel web search 병렬 실행 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_SCOUT` | boolean | 디버그/실험 | Scout subagent 활성화 | 부적합 — 기능 활성화 여부일 뿐 현재 agent 상태가 아님 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |
| `OPENCODE_EXPERIMENTAL_WORKSPACES` | boolean | 디버그/실험 | workspace support 활성화 | 부적합 — 실험 플래그 | [CLI experimental variables](https://opencode.ai/docs/cli/#experimental) |

## 감지 권고

- `OPENCODE_CLIENT=acp`는 `opencode acp` 프로세스 및 그 환경을 상속한 자식에서 유용한 보조 신호다.
- `OPENCODE_CLIENT=cli`는 사용자가 직접 설정할 수 있고 일반 CLI가 항상 자식에게 생성·주입한다고 보장되지 않으므로 단독 확정 신호로 사용하지 않는다.
- 나머지 `OPENCODE_*` 변수는 대체로 사전 설정값이다. 여러 변수가 함께 있으면 “OpenCode용으로 구성된 환경”이라는 약한 신호만 제공한다.
- 공식 문서에서 session ID, 현재 agent ID, 실행 단계나 active 상태를 자식 프로세스에 전달하는 환경변수는 확인되지 않았다.

## 공식 문서 및 소스

- [OpenCode CLI 문서](https://opencode.ai/docs/cli/)
- [OpenCode config 문서](https://opencode.ai/docs/config/)
- [OpenCode CLI 문서 원본](https://github.com/anomalyco/opencode/blob/dev/packages/web/src/content/docs/cli.mdx)
- [OpenCode ACP command source](https://github.com/anomalyco/opencode/blob/dev/packages/opencode/src/cli/cmd/acp.ts)
