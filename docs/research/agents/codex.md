---
title: OpenAI Codex
slug: codex
research_date: 2026-08-30
open_source: true
repository: https://github.com/openai/codex
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: false
runtime_test_reason: 2026-08-30 실제 Paseo 내부 Codex와 서브에이전트에서 환경 상속 및 rollout 메타데이터를 확인함
---

# OpenAI Codex

OpenAI의 공식 환경변수 문서는 Codex가 직접 읽는 안정적 공개 변수만 열거합니다. 이 공개 목록에는 자식 프로세스의 실행 주체를 식별하는 변수가 없지만, OpenAI가 공개한 Codex CLI 소스에서는 셸 도구의 자식 프로세스에 실행 식별자와 런타임 상태를 주입하는 구현을 확인할 수 있습니다.

따라서 이 문서는 **공개 문서로 보장되는 설정 변수**와 **공식 클라이언트 소스에서 확인된 런타임 변수**를 구분합니다. 후자는 실제 Codex 구현에 근거한 감지 신호이지만 안정적 공개 API 계약은 아니므로 버전 변경 가능성을 고려해야 합니다.

## 공식 공개 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CODEX_HOME` | 디렉터리 경로 | 설정 | 설정, 로그, 세션 및 스킬을 저장하는 Codex 상태 루트 | 부적합 — 사용자가 셸에 상시 설정할 수 있음 | [Core locations](https://learn.chatgpt.com/docs/config-file/environment-variables#core-locations) |
| `CODEX_SQLITE_HOME` | 디렉터리 경로 | 설정 | SQLite 기반 상태 저장 위치 | 부적합 — 실행 주체가 아닌 저장 위치 설정 | [Core locations](https://learn.chatgpt.com/docs/config-file/environment-variables#core-locations) |
| `CODEX_NON_INTERACTIVE` | 불리언 (`1`, `true`, `yes`) | 설정 | 독립 실행형 설치 스크립트의 프롬프트 생략 | 부적합 — 설치 과정에만 적용 | [Installer variables](https://learn.chatgpt.com/docs/config-file/environment-variables#installer-variables) |
| `CODEX_INSTALL_DIR` | 디렉터리 경로 | 설정 | `codex` 실행 파일 설치 위치 변경 | 부적합 — 설치 과정에만 적용 | [Installer variables](https://learn.chatgpt.com/docs/config-file/environment-variables#installer-variables) |
| `RUST_LOG` | 로그 필터 문자열 (`error`, `warn`, `info`, `debug`, `trace` 등) | 디버그/실험 | CLI와 app-server의 Rust 로그 수준 및 대상 설정 | 부적합 — Rust 프로그램 전반에서 쓰는 일반 변수 | [Diagnostics](https://learn.chatgpt.com/docs/config-file/environment-variables#diagnostics) |

## 공식 클라이언트 소스에서 확인된 런타임 변수

아래 내용은 조사 시점의 공식 `openai/codex` 저장소 커밋 [`63d2138`](https://github.com/openai/codex/commit/63d213884daea50e4f74efc192cdc44f549b67d5)을 기준으로 확인했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CODEX_THREAD_ID` | 스레드 ID 문자열 | 실행 식별 | 현재 Codex 스레드 ID를 셸 명령 환경에 주입 | 적합 — Codex가 런타임 값으로 덮어써 주입하는 가장 구체적인 신호 | [변수 정의와 환경 주입](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/protocol/src/shell_environment.rs#L6-L7), [unified exec 주입](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/unified_exec/process_manager.rs#L1344-L1351) |
| `CODEX_SESSION_ID` | 세션 ID 문자열 | 실행 식별 | 모델이 접근할 수 있는 셸 명령에 공유 루트 세션 ID를 주입 | 적합 — Codex 세션 계층을 식별하는 런타임 신호 | [세션 ID 주입 함수](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/exec_env.rs#L38-L40), [unified exec 호출부](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/unified_exec/process_manager.rs#L1347-L1352) |
| `CODEX_PERMISSION_PROFILE` | 권한 프로필 ID 문자열 | 상태/컨텍스트 | 선택된 명명형 권한 프로필을 셸 도구 환경에 주입 | 보조 신호 — 소스도 정보용 값이며 실제 권한 적용의 증거로 취급하지 말라고 명시 | [정의와 주입](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/exec_env.rs#L16-L18), [주입 구현](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/exec_env.rs#L43-L61) |
| `CODEX_SANDBOX` | 현재 macOS에서 `seatbelt` | 상태/컨텍스트 | macOS Seatbelt 샌드박스 아래에서 실행됨을 표시 | 보조 신호 — macOS Seatbelt 경로에 한정되며 값은 향후 변경될 수 있음 | [정의와 주의사항](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/spawn.rs#L20-L23), [Seatbelt 환경 주입](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/sandboxing/mod.rs#L180-L183) |
| `CODEX_SANDBOX_NETWORK_DISABLED` | `1` | 상태/컨텍스트 | Codex 셸 도구가 네트워크 제한 정책으로 자식 프로세스를 실행했음을 표시 | 보조 신호 — 제한된 셸 호출에서만 존재하는 실험적 변수 | [의미와 실험 상태](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/spawn.rs#L11-L23), [정책 기반 주입](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/sandboxing/mod.rs#L173-L183) |
| `CODEX_CI` | `1` | 실행 컨텍스트 | unified exec의 비대화형·결정적 명령 환경을 표시 | 보조 신호 — unified exec 경로에는 주입되지만 스레드/세션 식별자는 아님 | [unified exec 고정 환경](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/unified_exec/process_manager.rs#L88-L99), [환경 적용 함수](https://github.com/openai/codex/blob/63d213884daea50e4f74efc192cdc44f549b67d5/codex-rs/core/src/unified_exec/process_manager.rs#L123-L127) |

## 서브에이전트 판정과 rollout 메타데이터

2026-08-30 실제 Paseo 내부 Codex에서 루트 에이전트와 스폰된 서브에이전트의 전체 환경변수를 이름별 값 해시로 비교했다. 두 환경은 같은 `CODEX_SESSION_ID`를 유지했지만 `CODEX_THREAD_ID`는 서로 달랐다. 새로 추가된 서브에이전트 전용 환경변수는 없었으므로 **환경변수만으로는 현재 thread가 루트인지 서브에이전트인지 확정할 수 없다.**

현재 Codex 구현은 `CODEX_THREAD_ID`와 일치하는 로컬 rollout 파일을 다음 형태로 저장한다.

```text
$CODEX_HOME/sessions/YYYY/MM/DD/rollout-<timestamp>-<CODEX_THREAD_ID>.jsonl
```

`CODEX_HOME`이 설정되지 않은 기본 설치에서는 `~/.codex/sessions/...` 아래에서 찾을 수 있다. 파일명의 thread ID를 기준으로 찾아야 하며, 단순히 파일 내용 전체를 검색하면 다른 thread가 참조한 ID까지 함께 검색될 수 있다.

rollout의 첫 `session_meta` 레코드에서 실제로 확인한 서브에이전트 관련 필드는 다음과 같다.

| 필드 | 확인된 의미 | 판정 용도 |
|---|---|---|
| `thread_source` | 서브에이전트 thread에서는 `subagent` | 가장 직접적인 서브에이전트 판정 신호 |
| `parent_thread_id` | 이 thread를 스폰한 부모 Codex thread ID | 부모-자식 연결 |
| `source.subagent.thread_spawn.parent_thread_id` | spawn 이벤트에 기록된 부모 thread ID | 구조화된 부모 관계 교차 확인 |
| `source.subagent.thread_spawn.depth` | 루트로부터의 spawn 깊이 | 중첩 서브에이전트 깊이 |
| `agent_path` | 오케스트레이터가 부여한 agent 경로 | agent 트리 내 논리적 위치 |
| `agent_nickname` | 서브에이전트 표시 이름 | 사용자 표시용 보조 메타데이터 |
| `multi_agent_version` | multi-agent 메타데이터 형식 버전 | 파서 호환성 판단 |

권장 판정 순서는 다음과 같다.

1. 환경에서 `CODEX_THREAD_ID`를 읽어 Codex 실행을 감지한다.
2. 해당 thread ID로 rollout 파일을 찾는다.
3. 첫 레코드가 `session_meta`인지 확인한다.
4. `thread_source == "subagent"` 또는 `source.subagent`의 존재를 확인한다.
5. 존재한다면 `parent_thread_id`, `depth`, `agent_path`를 선택적 상태 정보로 읽는다.

rollout 경로와 JSONL 필드는 조사 시점의 실제 Codex 클라이언트 동작에서 확인한 **로컬 구현 세부사항**이며 안정적인 공개 API 계약이 아니다. 파일이 없거나 읽을 수 없거나 스키마가 달라져도 Codex 실행 감지 자체를 실패시키지 말고, 서브에이전트 상태만 `unknown`으로 남겨야 한다. 파일 내용에는 대화와 실행 정보가 포함될 수 있으므로 판정 시 첫 `session_meta` 외의 레코드는 읽거나 외부로 출력하지 않는 편이 안전하다.

## 실행 주체 감지에 관한 결론

`CODEX_THREAD_ID` 또는 `CODEX_SESSION_ID`가 있으면 공식 Codex CLI 구현에 근거해 Codex가 실행한 프로세스로 감지할 수 있습니다. `CODEX_SANDBOX`, `CODEX_SANDBOX_NETWORK_DISABLED`, `CODEX_CI`, `CODEX_PERMISSION_PROFILE`은 실행 경로와 플랫폼에 따라 없을 수 있으므로 상태 보강 신호로 사용하는 편이 안전합니다.

환경변수는 자식 프로세스에 상속되거나 사용자가 위조할 수 있으므로 절대적인 신뢰 경계는 아닙니다. 또한 이 런타임 변수들은 공식 소스에서 확인되지만 [안정적 공개 환경변수 목록](https://learn.chatgpt.com/docs/config-file/environment-variables)에는 포함되지 않습니다. `runby`는 이를 **공식 구현 기반의 비공개 계약 신호**로 취급하고, Codex 버전별 변경에 대비해야 합니다.

## 공식 문서

- [Codex 환경변수](https://learn.chatgpt.com/docs/config-file/environment-variables)
- [Codex CLI](https://learn.chatgpt.com/docs/codex/cli)
- [Codex 오픈 소스 구성요소](https://learn.chatgpt.com/docs/open-source)
- [공식 Codex CLI 소스 (`openai/codex`)](https://github.com/openai/codex)
