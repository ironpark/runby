---
title: Paseo
slug: paseo
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_orchestrator
executes_agents:
  - claude-code
  - codex
runtime_test_required: false
runtime_test_reason: 2026-08-30 실제 Paseo 내부 Codex 환경에서 두 계층의 동시 상속과 감지를 확인함
---

# Paseo

Paseo는 단독 agent harness가 아니라 Claude Code, Codex 등의 agent provider를 실행·관리하는 orchestrator입니다. 따라서 `PASEO_AGENT_ID`는 **Paseo가 관리하는 에이전트 실행**을 식별하고, 실제 하위 harness는 해당 provider의 환경변수로 별도 판별해야 합니다. 공식 CLI 문서는 이 값을 호출 에이전트 식별과 부모-자식 연결에 사용한다고 설명합니다.

## 에이전트 실행 및 워크스페이스 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `PASEO_AGENT_ID` | 에이전트 ID 문자열 | 실행 식별 | 현재 Paseo 에이전트를 식별하고 그 에이전트가 만든 작업을 하위 에이전트로 연결 | 적합 — 공식 문서가 호출 에이전트 인식 용도를 명시 | [CLI: Running agents](https://paseo.sh/docs/cli#running-agents) |
| `PASEO_SOURCE_CHECKOUT_PATH` | 절대 디렉터리 경로 | 상태/컨텍스트 | 원본 저장소 루트 | 보조 — Paseo 워크트리임은 나타내지만 특정 에이전트 실행은 보장하지 않음 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_WORKTREE_PATH` | 절대 디렉터리 경로 | 상태/컨텍스트 | Paseo가 만든 워크트리 경로 | 보조 — 워크스페이스 컨텍스트 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_BRANCH_NAME` | Git 브랜치 문자열 | 상태/컨텍스트 | 워크트리 브랜치 이름 | 보조 — 워크스페이스 컨텍스트 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_WORKTREE_PORT` | 정수형 포트 문자열 | 상태/컨텍스트 | 레거시 워크트리 포트 | 보조 — 특정 서비스/워크트리 컨텍스트이며 에이전트 ID는 아님 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_PORT` | 정수형 포트 문자열 | 상태/컨텍스트 | 서비스에 할당된 포트 | 보조 — Paseo 서비스 컨텍스트 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_URL` | URL 문자열 | 상태/컨텍스트 | 서비스 프록시 URL | 보조 — Paseo 서비스 컨텍스트 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_SERVICE_<NAME>_PORT` | 정수형 포트 문자열 | 상태/컨텍스트 | 동료 서비스의 할당 포트 | 보조 — 동적 변수명이며 서비스 컨텍스트만 나타냄 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |
| `PASEO_SERVICE_<NAME>_URL` | URL 문자열 | 상태/컨텍스트 | 동료 서비스의 프록시 URL | 보조 — 동적 변수명이며 서비스 컨텍스트만 나타냄 | [Worktrees: Environment variables](https://paseo.sh/docs/worktrees#environment-variables) |

`PASEO_AGENT_CWD`는 현재 일부 실행 환경에서 관찰되지만, 조사 기준일의 Paseo 공식 문서에서는 공개 환경변수 계약을 확인하지 못했습니다. 따라서 공식적으로 문서화되기 전까지는 호환성 휴리스틱으로만 취급해야 합니다.

## 런타임 및 연결 설정

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `PASEO_HOME` | 디렉터리 경로 | 설정 | Paseo 설정과 로컬 상태 루트 변경 | 부적합 — 셸에 상시 설정 가능 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_HOST` | 호스트 또는 URI 문자열 | 설정 | CLI가 연결할 데몬 지정 | 부적합 — 연결 설정 | [CLI: Daemon management](https://paseo.sh/docs/cli#daemon-management) |
| `PASEO_HUB_URL` | URL 문자열 | 설정 | Paseo Hub 원본 주소 지정 | 부적합 — 연결 설정 | [Hub public API](https://paseo.sh/docs/hub/api) |
| `PASEO_LISTEN` | 주소 문자열 | 설정 | 데몬 수신 주소 재정의 | 부적합 — 데몬 설정 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_RELAY_ENABLED` | 불리언 | 설정 | 해당 데몬 실행의 릴레이 활성화 여부 | 부적합 — 데몬 설정 | [Relay](https://paseo.sh/docs/configuration#relay) |
| `PASEO_HOSTNAMES` | 호스트명 목록 | 설정 | 허용 호스트명 확장 또는 재정의 | 부적합 — 데몬 설정 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_ALLOWED_HOSTS` | 호스트명 목록 | 설정 | `PASEO_HOSTNAMES`의 폐기 예정 별칭 | 부적합 — 데몬 설정 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_WEB_UI_ENABLED` | 불리언 | 설정 | 데몬 제공 웹 UI 활성화 | 부적합 — 데몬 설정 | [Bundled web UI](https://paseo.sh/docs/configuration#bundled-web-ui) |
| `PASEO_WEB_UI_DIST_DIR` | 디렉터리 경로 | 설정 | 웹 UI 빌드 디렉터리 재정의 | 부적합 — 데몬 설정 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_TRUSTED_PROXIES` | 프록시 범위 목록 | 설정 | 신뢰하는 역방향 프록시 범위 설정 | 부적합 — 데몬 설정 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |

## 로그와 음성 기능 변수

다음 변수도 공식 설정 문서에 공개되어 있으나 에이전트 실행 식별에는 모두 부적합합니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `PASEO_LOG_CONSOLE_LEVEL` | 로그 수준 문자열 | 디버그/실험 | 콘솔 로그 수준 | 부적합 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_LOG_FILE_LEVEL` | 로그 수준 문자열 | 디버그/실험 | 파일 로그 수준 | 부적합 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_LOG_FILE_PATH` | 파일 경로 | 디버그/실험 | 로그 파일 경로 | 부적합 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_LOG_FILE_ROTATE_SIZE` | 크기 문자열 | 디버그/실험 | 로그 회전 최대 크기 | 부적합 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_LOG_FILE_ROTATE_COUNT` | 정수 | 디버그/실험 | 유지할 로그 파일 수 | 부적합 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_LOG`, `PASEO_LOG_FORMAT` | 로그 수준/형식 문자열 | 디버그/실험 | 레거시 로그 재정의 | 부적합 | [Configuration](https://paseo.sh/docs/configuration#common-env-vars) |
| `PASEO_VOICE_LLM_PROVIDER` | `claude`, `codex`, `opencode` | 설정 | 음성 에이전트 공급자 선택 | 부적합 — 실행 중인 공급자와 일치한다고 보장할 수 없음 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_DICTATION_STT_PROVIDER` | `local` 또는 `openai` | 설정 | 받아쓰기 STT 공급자 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_VOICE_STT_PROVIDER` | `local` 또는 `openai` | 설정 | 음성 모드 STT 공급자 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_VOICE_TTS_PROVIDER` | `local` 또는 `openai` | 설정 | 음성 모드 TTS 공급자 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `OPENAI_STT_BASE_URL`, `OPENAI_TTS_BASE_URL` | URL 문자열 | 설정 | OpenAI 호환 음성 엔드포인트 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_LOCAL_MODELS_DIR` | 디렉터리 경로 | 설정 | 로컬 음성 모델 저장 위치 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_DICTATION_LOCAL_STT_MODEL`, `PASEO_VOICE_LOCAL_STT_MODEL`, `PASEO_VOICE_LOCAL_TTS_MODEL` | 모델 ID 문자열 | 설정 | 로컬 음성 모델 선택 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_DICTATION_LANGUAGE`, `PASEO_VOICE_LANGUAGE` | 언어 코드 문자열 | 설정 | STT 언어 선택 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_VOICE_LOCAL_TTS_SPEAKER_ID` | 화자 ID | 설정 | 로컬 TTS 화자 선택 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |
| `PASEO_VOICE_LOCAL_TTS_SPEED` | 숫자형 문자열 | 설정 | 로컬 TTS 속도 조정 | 부적합 | [Voice](https://paseo.sh/docs/voice#environment-variables) |

## 공식 문서

- [Paseo CLI](https://paseo.sh/docs/cli)
- [Paseo Configuration](https://paseo.sh/docs/configuration)
- [Paseo Git worktrees](https://paseo.sh/docs/worktrees)
- [Paseo Voice](https://paseo.sh/docs/voice)
