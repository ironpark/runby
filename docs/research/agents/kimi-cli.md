---
title: Kimi CLI
slug: kimi-cli
research_date: 2026-09-01
open_source: true
repository: https://github.com/MoonshotAI/kimi-cli
product_type: agent_harness
model_source: first-party
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공개 저장소에는 주입 마커가 없으나 IDE 플러그인 배포본이 별도 변수를 심을 가능성이 남아 실측 필요
---

# Kimi CLI

## 요약

**드라이버 없음 — 음성 결과.** 공식 저장소를 확인했으나 Kimi CLI가 자식 프로세스에 심는 식별 마커가 없습니다.

셸 도구(`src/kimi_cli/tools/shell/__init__.py`)는 자식 환경을 만들 때 `SHELL`만 자신이 실행하는 bash 경로로 덮어씁니다. 이는 `$SHELL`을 읽는 명령이 부모의 낡은 값을 보지 않게 하려는 것이며 식별 목적이 아닙니다. `KIMI_*` 변수는 전부 CLI가 **읽는** 설정값입니다 — `KIMI_API_KEY`(자격증명), `KIMI_SHARE_DIR`·`KIMI_WORK_DIR`(경로 설정), `KIMI_MODEL_NAME`·`KIMI_BASE_URL`(모델 설정).

## `KIMI_PLUGIN_ROOT`에 대하여

`vercel/detect-agent`는 `KIMI_PLUGIN_ROOT`로 Kimi를 감지합니다. **이 변수는 공식 `kimi-cli` 저장소 어디에도 없습니다.** IDE 플러그인 배포본에서 오는 값으로 추정되나 공식 문서·소스로 확인하지 못했습니다.

설령 존재하더라도 `*_PLUGIN_ROOT`라는 이름은 플러그인 설치 경로를 가리키는 설정값일 가능성이 높고, 그렇다면 Grok과 마찬가지로 훅 실행 범위에 한정될 것입니다. 확인 없이 채택하지 않습니다.

## 결론

Kimi CLI는 현재 감지하지 않습니다. 공식 소스에 범용 마커가 추가되거나, `KIMI_PLUGIN_ROOT`의 출처와 주입 범위가 공식 문서로 확인되면 재검토합니다.
