---
title: Zed Agent
slug: zed-agent
research_date: 2026-08-30
open_source: true
repository: https://github.com/zed-industries/zed
product_type: agent_host
model_source: multi-vendor
executes_agents:
  - zed-agent
  - acp-agents
runtime_test_required: true
runtime_test_reason: Zed terminal 및 Task 신호만으로 Agent 요청과 사용자 직접 실행을 구분할 수 있는지 확인해야 함
---

# Zed Agent

Zed Agent 전용의 세션·스레드 실행 마커는 최신 공식 문서와 공개 공식 소스에서 확인되지 않았습니다. 공식 소스는 Zed가 만든 로컬·원격 터미널에 `ZED_TERM=true`와 터미널 식별값을 주입하고, 공식 문서는 Zed Task에 편집기·worktree 컨텍스트 변수를 제공합니다. 이 신호들은 **Zed가 만든 터미널 또는 Task**는 식별하지만, 그 명령을 Zed Agent가 요청했는지 사용자가 직접 실행했는지는 구분하지 못합니다.

## 터미널 실행 신호 — 공식 소스

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ZED_TERM` | 불리언 문자열 (`true`) | 실행 식별 | Zed의 로컬·원격 터미널 세션 표시 | 보조 — Zed 터미널의 강한 신호지만 Zed Agent 전용은 아님 | [공식 소스: `insert_zed_terminal_env`](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/terminal/src/terminal.rs#L654-L664) |
| `TERM_PROGRAM` | 문자열 (`zed`) | 실행 식별 | 터미널 호스트 프로그램 식별 | 보조 — `ZED_TERM=true`와 함께 Zed 터미널을 확인하는 교차 신호. 다른 터미널도 같은 표준 변수를 사용함 | [공식 소스: `insert_zed_terminal_env`](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/terminal/src/terminal.rs#L654-L664) |
| `TERM_PROGRAM_VERSION` | 버전 문자열 | 상태/컨텍스트 | Zed 버전 제공 | 보조 — `TERM_PROGRAM=zed`와 함께 버전 컨텍스트 제공. 단독 감지에는 부적합 | [공식 소스: `insert_zed_terminal_env`](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/terminal/src/terminal.rs#L654-L664) |
| `TERM` | 문자열 (`xterm-256color`) | 설정 | 터미널 기능 유형 설정 | 부적합 — 매우 일반적인 표준 터미널 변수 | [공식 소스: `insert_zed_terminal_env`](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/terminal/src/terminal.rs#L654-L664) |
| `COLORTERM` | 문자열 (`truecolor`) | 설정 | true-color 지원 표시 | 부적합 — 여러 터미널이 동일 값을 사용 | [공식 소스: `insert_zed_terminal_env`](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/terminal/src/terminal.rs#L654-L664) |

## Task 컨텍스트 변수

이 변수들은 Zed Task 템플릿을 해석하고 Task 프로세스를 만들 때 제공됩니다. 현재 버퍼나 선택 영역이 없으면 일부 변수는 존재하지 않을 수 있으며, 일반 내장 터미널 전체에 항상 주입되는 계약은 아닙니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ZED_COLUMN` | 정수 | 상태/컨텍스트 | 현재 줄의 열 위치 | 보조 — Zed Task 컨텍스트이지만 사용자가 같은 이름을 설정할 수 있고 Agent 전용이 아님 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_ROW` | 정수 | 상태/컨텍스트 | 현재 줄 번호 | 보조 — Zed Task 컨텍스트이지만 Agent 전용이 아님 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_FILE` | 절대 파일 경로 | 상태/컨텍스트 | 현재 열린 파일 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_FILENAME` | 파일명 문자열 | 상태/컨텍스트 | 현재 열린 파일의 이름 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_DIRNAME` | 절대 디렉터리 경로 | 상태/컨텍스트 | 현재 파일의 상위 디렉터리 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_RELATIVE_FILE` | 상대 파일 경로 | 상태/컨텍스트 | worktree 루트 기준 현재 파일 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_RELATIVE_DIR` | 상대 디렉터리 경로 | 상태/컨텍스트 | worktree 루트 기준 현재 파일 디렉터리 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_STEM` | 문자열 | 상태/컨텍스트 | 확장자를 제외한 현재 파일명 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_SYMBOL` | 문자열 | 상태/컨텍스트 | 현재 선택된 코드 심볼 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_SELECTED_TEXT` | 문자열 | 상태/컨텍스트 | 현재 선택한 텍스트 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_LANGUAGE` | 언어명 문자열 | 상태/컨텍스트 | 현재 버퍼의 언어 | 보조 — Zed Task 신호이나 Agent 실행 여부를 구분하지 못함 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_WORKTREE_ROOT` | 절대 디렉터리 경로 | 상태/컨텍스트 | 현재 worktree 루트 | 보조 — Task나 worktree hook의 강한 Zed 컨텍스트지만 일반 터미널에는 항상 제공되지 않고 Agent 전용도 아님 | [Zed Tasks — Variables and hooks](https://zed.dev/docs/tasks#variables) |
| `ZED_MAIN_GIT_WORKTREE` | 절대 디렉터리 경로 | 상태/컨텍스트 | linked worktree의 원본 Git worktree | 보조 — `create_worktree` hook에서 유용한 Zed 컨텍스트이나 Agent 전용이 아님 | [Zed Tasks — Variables and hooks](https://zed.dev/docs/tasks#hooks) |
| `ZED_CUSTOM_RUST_PACKAGE` | 패키지명 문자열 | 상태/컨텍스트 | Rust 소스 파일의 상위 패키지 | 보조 — 언어별 Task 컨텍스트이며 실행 식별자는 아님 | [Zed Tasks — Variables](https://zed.dev/docs/tasks#variables) |
| `ZED_CUSTOM_<CAPTURE_NAME>` | 문자열 | 상태/컨텍스트 | 언어 확장의 runnable query capture를 Task 변수로 노출 | 보조 — Zed가 해석한 Task 컨텍스트이나 변수명이 동적이고 Agent 전용이 아님 | [Zed Extensions — Runnable code detection](https://zed.dev/docs/extensions/languages#runnable-code-detection) |
| `ZED_GIT_SHA` | 전체 Git SHA 문자열 | 상태/컨텍스트 | Git Graph에서 선택한 커밋 | 보조 — Git Graph command Task에만 제공 | [Zed Tasks — Custom Git Commands](https://zed.dev/docs/tasks#custom-git-commands) |
| `ZED_GIT_SHA_SHORT` | 짧은 Git SHA 문자열 | 상태/컨텍스트 | 선택 커밋의 짧은 SHA | 보조 — Git Graph command Task에만 제공 | [Zed Tasks — Custom Git Commands](https://zed.dev/docs/tasks#custom-git-commands) |
| `ZED_GIT_REPOSITORY_NAME` | 저장소명 문자열 | 상태/컨텍스트 | 선택 Git 저장소의 이름 | 보조 — Git Graph command Task에만 제공 | [Zed Tasks — Custom Git Commands](https://zed.dev/docs/tasks#custom-git-commands) |
| `ZED_GIT_REPOSITORY_PATH` | 절대 디렉터리 경로 | 상태/컨텍스트 | 선택 Git 저장소의 worktree 경로 | 보조 — Git Graph command Task에만 제공 | [Zed Tasks — Custom Git Commands](https://zed.dev/docs/tasks#custom-git-commands) |
| `ZED_GIT_REF` | Git ref 문자열 | 상태/컨텍스트 | 클릭한 branch·tag·remote ref | 보조 — ref에서 연 Git Graph command Task에만 제공 | [Zed Tasks — Custom Git Commands](https://zed.dev/docs/tasks#custom-git-commands) |

## 개발·평가 전용 변수 — 공식 소스

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ZED_AGENT_MODEL` | `provider/model-id` 문자열 | 디버그/실험 | Zed Agent 도구·inline assistant 평가 테스트에서 실제 평가 모델 선택 | 부적합 — 테스트 코드가 읽는 개발용 입력이며 실행 프로세스에 주입되는 마커가 아님 | [공식 소스: terminal tool eval](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/agent/src/tools/evals/terminal_tool.rs#L170-L174) |
| `ZED_JUDGE_MODEL` | `provider/model-id` 문자열 | 디버그/실험 | Agent edit-file 평가의 판정 모델 선택 | 부적합 — 평가 테스트 전용 입력 | [공식 소스: edit-file eval](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/agent/src/tools/evals/edit_file.rs#L264-L272) |
| `ZED_SIMULATE_NO_LLM_PROVIDER` | 존재 여부 | 디버그/실험 | debug build에서 LLM 공급자가 없는 상태 시뮬레이션 | 부적합 — `debug_assertions`에서만 읽는 개발용 토글 | [공식 소스: language model registry](https://github.com/zed-industries/zed/blob/399258feeaf90ad8a3a208c99221ee87b6452f38/crates/language_model/src/registry.rs#L459-L463) |

## 실행 주체 감지에 관한 결론

`ZED_TERM=true`와 `TERM_PROGRAM=zed`의 조합은 현재 프로세스가 Zed 터미널에서 실행되었다는 가장 신뢰도 높은 공개 신호입니다. 여기에 `ZED_WORKTREE_ROOT` 같은 Task 컨텍스트가 있으면 Zed Task라는 판단을 강화할 수 있습니다. 그러나 어느 신호도 Zed Agent가 명령을 요청했는지, 사용자가 터미널이나 Task UI에서 직접 실행했는지를 구분하지 않습니다. 따라서 에이전트 이름은 `zed` 또는 `zed-terminal` 수준으로 보고하고, `zed-agent` 확정 판정에는 사용하지 않는 편이 안전합니다.

설치(`ZED_VERSION`, `ZED_CHANNEL`), 앱 개발·그래픽·로그, 프록시 변수는 Zed 제품 전체의 설정일 뿐 Zed Agent 실행과 직접 관계가 없어 이 문서의 표에서 제외했습니다.

## 공식 문서

- [Zed Agent](https://zed.dev/docs/ai/zed-agent)
- [AI Agents in Zed](https://zed.dev/docs/ai/agents)
- [Zed Tasks](https://zed.dev/docs/tasks)
- [Environment Variables](https://zed.dev/docs/environment)
- [Zed 공식 GitHub 저장소](https://github.com/zed-industries/zed)
