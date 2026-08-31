---
title: GitHub Copilot CLI
slug: github-copilot-cli
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_harness
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 문서에서 자식 프로세스 실행 마커를 확인하지 못해 실제 환경 검증이 필요함
---

# GitHub Copilot CLI

## 요약

GitHub 공식 문서는 Copilot CLI가 읽는 설정, UI 및 OpenTelemetry 환경변수를 공개한다. 그러나 Copilot CLI가 자신이 실행한 셸 명령이나 MCP 프로세스에 **의도적으로 자동 주입하는 안정적인 실행 식별 환경변수는 공식 문서에서 확인되지 않는다**. 따라서 아래 변수는 Copilot CLI의 구성 상태를 해석하는 데에는 쓸 수 있지만, 현재 프로세스를 Copilot CLI가 실행했다는 단독 증거로 사용하면 안 된다.

## 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
| --- | --- | --- | --- | --- | --- |
| `COPILOT_ALLOW_ALL` | boolean (`true`) | 설정 | 모든 도구·경로·URL 권한을 자동 허용 | 부적합 — 실행 전에 사용자가 설정하며 다른 프로세스에도 남을 수 있음 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_AUTO_UPDATE` | boolean (`false`로 비활성화) | 설정 | CLI와 1st-party 플러그인 자동 업데이트 제어 | 부적합 — 지속 설정일 수 있으며 실행 주체를 나타내지 않음 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_CACHE_HOME` | 경로 문자열 | 설정 | 마켓플레이스·업데이트 등의 캐시 디렉터리 변경 | 부적합 — 사용자 지정 경로일 뿐 실행 증거가 아님 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_CUSTOM_INSTRUCTIONS_DIRS` | 쉼표 구분 경로 문자열 | 설정 | 추가 custom instructions 디렉터리 지정 | 부적합 — 실행 전 입력 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_EDITOR` | 명령 문자열 | 설정 | 대화형 편집기 지정 (`VISUAL`, `EDITOR` 다음 우선순위) | 부적합 — 일반 셸 설정으로도 유지 가능 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_ENABLE_HTTP2` | boolean (`1`, `true`) | 디버그/실험 | HTTP/2 전송 opt-in | 부적합 — 기능 플래그이며 실행 식별자가 아님 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_GH_HOST` | 호스트명 문자열 | 설정 | Copilot CLI 전용 GitHub 호스트 지정 | 부적합 — 실행 전 구성값 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_HOME` | 경로 문자열 | 설정 | 설정·상태 디렉터리 변경, 기본값 `$HOME/.copilot` | 부적합 — 사용자 설정이며 자동 실행 표지가 아님 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_LARGE_OUTPUT_THRESHOLD_BYTES` | 정수, 기본 `20480` | 설정 | 모델에 직접 전달할 도구 출력 크기 상한 | 부적합 — 튜닝 값 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_MCP_TOOL_CACHE` | boolean (`false`로 비활성화) | 설정 | 로컬 MCP 도구 스냅샷 캐시 제어 | 부적합 — MCP 설정 상태만 나타냄 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_MODEL` | 모델 ID 문자열 | 설정 | 사용할 AI 모델 지정 | 부적합 — 다른 실행을 위해 미리 설정될 수 있음 | [Programmatic reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference#environment-variables) |
| `COPILOT_PLAN_THEN_AUTOPILOT` | boolean (`1`, `true`, `yes`, `on`) | 상태/컨텍스트 | plan 후 autopilot 모드 요청 | 부적합 — 실행 모드 입력값이지 현재 실행 주체의 출력값이 아님 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_PROMPT_FRAME` | boolean (`1`/`0`) | 디버그/실험 | 입력 프롬프트 장식 프레임 제어 | 부적합 — UI 기능 플래그 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_SKILLS_DIRS` | 쉼표 구분 경로 문자열 | 설정 | 추가 skills 디렉터리 지정 | 부적합 — 실행 전 입력 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `PLUGINS_DASHBOARD` | boolean (`false`로 비활성화) | 설정 | 플러그인 대시보드와 비대화형 플러그인 명령 제어 | 부적합 — 일반 기능 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_STRIP_REASONING_ON_RESUME` | boolean (`0`/`false`로 유지) | 상태/컨텍스트 | 세션 재개 시 reasoning token 제거 여부 | 부적합 — 재개 동작 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_SUBAGENT_MAX_CONCURRENT` | 정수, 기본 `32`, 범위 `1`–`256` | 설정 | 세션당 동시 subagent 상한 | 부적합 — 동시성 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_SUBAGENT_MAX_DEPTH` | 정수, 기본 `4`, 범위 `1`–`128` | 설정 | subagent 중첩 깊이 상한 | 부적합 — 동시성 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_TASK_WAIT_TIMEOUT_SECONDS` | 정수 초, 기본 `600` | 설정 | prompt 모드 종료 전 background agent·shell 대기 시간 | 부적합 — 시간 제한 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `GH_HOST` | 호스트명 문자열 | 설정 | GitHub CLI와 Copilot CLI 공통 GitHub 호스트 | 부적합 — GitHub CLI도 사용하는 범용 변수 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `GITHUB_COPILOT_PROMPT_MODE_EXTENSIONS` | boolean (`true`) | 설정 | prompt 모드에서 프로젝트 extension 로드·관리 허용 | 부적합 — 실행 전 기능 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `GITHUB_COPILOT_PROMPT_MODE_REPO_HOOKS` | boolean (`true`) | 설정 | prompt 모드에서 저장소 hook 로드 | 부적합 — 실행 전 기능 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `GITHUB_COPILOT_PROMPT_MODE_WORKSPACE_MCP` | boolean (`true`) | 설정 | prompt 모드에서 workspace MCP source 로드 | 부적합 — 실행 전 기능 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `PLAIN_DIFF` | boolean (`true`) | 설정 | rich diff 렌더링 비활성화 | 부적합 — 출력 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `USE_BUILTIN_RIPGREP` | boolean (`false`로 시스템 rg 사용) | 설정 | 번들 ripgrep 사용 여부 | 부적합 — 검색 도구 설정 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `USE_TGREP` | boolean | 디버그/실험 | tgrep 강제 사용 또는 ripgrep 강제 사용 | 부적합 — 검색 구현 선택 플래그 | [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#environment-variables) |
| `COPILOT_OTEL_ENABLED` | boolean, 기본 `false` | 디버그/실험 | OpenTelemetry 명시적 활성화 | 부적합 — 관측 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | URL 문자열 | 디버그/실험 | OTLP endpoint 지정, 설정 시 OTel 자동 활성화 | 부적합 — 여러 애플리케이션이 쓰는 표준 OTel 변수 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `COPILOT_OTEL_EXPORTER_TYPE` | enum: `otlp-http`(기본), `file` | 디버그/실험 | Copilot OTel exporter 선택 | 부적합 — 관측 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | enum: `http/json`(기본), `http/protobuf` | 디버그/실험 | OTLP HTTP wire protocol 지정 | 부적합 — 표준 OTel 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` | protocol 문자열 | 디버그/실험 | trace protocol만 override | 부적합 — 표준 OTel 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` | protocol 문자열 | 디버그/실험 | metric protocol만 override | 부적합 — 표준 OTel 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_SERVICE_NAME` | 문자열, 기본 `github-copilot` | 디버그/실험 | OTel resource service name | 부적합 — 기본값은 환경에 자동 export되지 않고 사용자가 바꿀 수도 있음 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_RESOURCE_ATTRIBUTES` | 쉼표 구분 `key=value` 문자열 | 디버그/실험 | 추가 OTel resource attribute | 부적합 — 표준 OTel 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT` | boolean, 기본 `false` | 디버그/실험 | prompt·response 전체 내용 수집 | 부적합 — 표준 관측 플래그이며 민감 내용 노출 가능 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `OTEL_LOG_LEVEL` | enum: `NONE`–`ALL` | 디버그/실험 | OTel 진단 로그 수준 | 부적합 — 표준 OTel 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `COPILOT_OTEL_FILE_EXPORTER_PATH` | 파일 경로 문자열 | 디버그/실험 | JSON Lines signal 출력 파일, 설정 시 OTel 활성화 | 부적합 — 관측 출력 설정 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |
| `COPILOT_OTEL_SOURCE_NAME` | 문자열, 기본 `github.copilot` | 디버그/실험 | tracer·meter instrumentation scope 이름 | 부적합 — 관측 설정이며 자동 주입된 프로세스 표지가 아님 | [OTel 환경변수](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference#otel-environment-variables) |

## 감지 권고

- 공식 환경변수만으로는 GitHub Copilot CLI가 현재 프로세스를 실행했다고 확정하지 않는다.
- `COPILOT_*`가 여러 개 존재하더라도 “Copilot CLI용으로 구성된 환경” 정도의 낮은 신뢰도 보조 신호로만 취급한다.
- 세션 ID, agent 상태, 현재 실행 모드를 자식 프로세스에 전달하는 공식 환경변수는 확인되지 않았다.

## 공식 문서

- [GitHub Copilot CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference)
- [GitHub Copilot CLI programmatic reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference)
