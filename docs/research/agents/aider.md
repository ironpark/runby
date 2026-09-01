---
title: Aider
slug: aider
research_date: 2026-09-02
open_source: true
repository: https://github.com/Aider-AI/aider
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 run_cmd.py의 두 실행 경로가 호출자 환경을 그대로 상속할 뿐 Aider 마커를 추가하지 않음
---

# Aider

## 요약

**드라이버 없음 — 음성 결과.** Aider의 `run_cmd.py`는 TTY 여부에 따라 `subprocess.Popen` 또는 `pexpect.spawn`을 사용한다. 두 경로 모두 `env`를 넘겨 Aider 전용 변수를 주입하지 않고, `cwd`·출력 옵션만 설정한다. 따라서 자식 명령이 부모 환경을 상속한다는 사실은 확인되지만, Aider가 실행 주체임을 식별할 수 있는 마커는 아니다.

## `AIDER_*` 변수는 설정 입력

Aider 공식 설정 문서는 `AIDER_*`를 셸에서 `export`하는 방식으로 옵션을 지정한다고 설명한다. `AIDER_MODEL`, `AIDER_DARK_MODE`, `AIDER_API_KEY` 같은 이름은 Aider를 실행하기 전에 사용자가 넣는 모델·UI·인증 설정이지, Aider가 자식 프로세스에 남기는 실행 표지가 아니다. 설정값은 Aider를 실행하지 않는 다른 프로세스에도 상속될 수 있으므로 감지에 쓰지 않는다.

| 조사 대상 | 확인 내용 | 판정 |
|---|---|---|
| `run_cmd_subprocess()` | `Popen`에 `cwd`는 있지만 Aider 전용 `env`가 없음 | 자식 환경에 실행 마커를 주입하지 않음 |
| `run_cmd_pexpect()` | `pexpect.spawn`에 `cwd`는 있지만 Aider 전용 `env`가 없음 | 자식 환경에 실행 마커를 주입하지 않음 |
| `AIDER_*` | 공식 설정 문서가 사용자의 shell 환경변수 입력으로 설명 | **부적합** — 실행 전 설정값 |

## 결론

Aider가 자식에게 전용 세션 ID나 실행 마커를 추가한다는 공식 계약이 없으므로 `runby`에서는 감지하지 않는다. `AIDER_*`의 존재를 실행 증거로 해석하면 환경을 미리 구성한 일반 셸까지 Aider로 오탐한다.

## 공식 소스·문서

- [`run_cmd()` — TTY별 실행 경로 선택](https://github.com/Aider-AI/aider/blob/main/aider/run_cmd.py#L11-L16)
- [`run_cmd_subprocess()` — env override 없는 Popen](https://github.com/Aider-AI/aider/blob/main/aider/run_cmd.py#L42-L73)
- [`run_cmd_pexpect()` — env override 없는 pexpect.spawn](https://github.com/Aider-AI/aider/blob/main/aider/run_cmd.py#L89-L121)
- [Aider 설정 — 환경변수 입력](https://aider.chat/docs/config.html)
- [Aider 옵션 목록](https://aider.chat/docs/config/options.html)
