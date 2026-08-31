---
title: Cline
slug: cline
research_date: 2026-09-01
open_source: true
repository: https://github.com/cline/cline
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 소스에서 터미널 생성 시 주입 지점을 직접 확인했으며, 사람이 같은 터미널에 입력할 수 있다는 한계는 소스만으로 판정 가능
---

# Cline

## 요약

Cline은 VS Code 확장으로 동작하며, 명령을 실행하기 위해 **자신이 만드는 터미널**의 환경에 `CLINE_ACTIVE=true`를 주입합니다. 공식 소스의 두 지점에서 확인됩니다 — 터미널 레지스트리가 터미널을 생성할 때와, hostbridge가 터미널에서 명령을 실행할 때입니다.

## 신뢰도를 `probable`로 낮춘 이유

이 마커는 **터미널에 붙어 있지 프로세스 호출에 붙어 있지 않습니다.** Cline이 만든 터미널은 VS Code UI에 노출되므로 사용자가 그 터미널에 직접 타이핑할 수 있고, 그 명령도 `CLINE_ACTIVE=true`를 받습니다. 즉 "Cline이 만든 터미널에서 실행되었다"는 확실하지만 "Cline이 이 명령을 요청했다"는 확실하지 않습니다.

`runby`의 `ConfidenceProbable` 정의 — "에이전트 실행과 부합하지만 배타적이지 않은 보조 신호" — 에 정확히 해당합니다. 같은 이유로 Zed의 `ZED_TERM`은 에이전트가 아니라 터미널 축으로 분류되어 있으나, Cline의 터미널은 **Cline이 작업 실행을 위해 만든 것**이라는 점에서 사용자가 여는 일반 터미널과 다르므로 에이전트 축에 둡니다.

## 환경변수

| 변수 | 값 | 분류 | 의미 | 판정 | 근거 |
|---|---|---|---|---|---|
| `CLINE_ACTIVE` | `true` | 실행 식별 | Cline이 생성한 터미널 표시 | **적합 (보조)** — 주입 마커이나 사람의 입력과 구분하지 못함 | [`VscodeTerminalRegistry.ts`](https://github.com/cline/cline/blob/main/apps/vscode/src/hosts/vscode/terminal/VscodeTerminalRegistry.ts), [`executeCommandInTerminal.ts`](https://github.com/cline/cline/blob/main/apps/vscode/src/hosts/vscode/hostbridge/workspace/executeCommandInTerminal.ts) |
| `CLINE_DIR`, `CLINE_DATA_DIR`, `CLINE_GLOBAL_SETTINGS_PATH` 등 | 경로 문자열 | 설정 | 설정·상태 디렉터리 재지정 | 부적합 — 사용자가 실행 전에 설정하는 경로 | 저장소 전반 |
| `CLINE_API_KEY` | 토큰 문자열 | 설정 | API 인증 | 부적합 — **자격증명** | 저장소 전반 |

## 실행 파일

없습니다. Cline은 편집기 확장으로 동작하므로 조상 체인에 나타나는 것은 편집기(`code` 등)이며, 이는 Cline 실행의 증거가 아닙니다.
