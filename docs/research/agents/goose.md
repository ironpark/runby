---
title: Goose (Block)
slug: goose
research_date: 2026-09-01
open_source: true
repository: https://github.com/block/goose
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: AGENT_SESSION_ID가 셸 rc에 export된 채 남는지, goose 종료 후에도 잔존하는지 실측해야 판정을 뒤집을 수 있음
---

# Goose (Block)

## 요약

**드라이버 없음 — 음성 결과.** Goose는 오픈소스이며 소스를 직접 확인했으나, `runby` 기준을 통과하는 실행 마커가 없습니다.

이 문서는 `vercel/detect-agent`가 Goose 감지에 `GOOSE_PROVIDER`를 쓰는 것을 검토한 결과이기도 합니다. **그 판정은 이 저장소 기준으로는 부적합**입니다.

## `GOOSE_PROVIDER`가 부적합한 이유

`GOOSE_PROVIDER`는 **어떤 LLM 공급자를 쓸지 고르는 설정값**입니다. 공식 CLI 소스의 `--provider` 플래그 도움말이 "이번 실행에 한해 `GOOSE_PROVIDER` 환경변수를 재정의한다"고 적고 있고, 코드에서도 `config.set_param("GOOSE_PROVIDER", ...)`로 설정 시스템을 통해 읽고 씁니다.

즉 이것은 사용자가 셸 프로파일이나 CI에 미리 넣어 두는 값입니다. `~/.zshrc`에 `export GOOSE_PROVIDER=anthropic`을 넣은 사람은 Goose를 실행하지 않아도 **모든 프로세스가 Goose로 오탐**됩니다. `runby`의 "설정 변수를 근거로 쓰지 마십시오" 규칙에 정면으로 걸립니다.

## 실제 주입 마커와 그 한계

| 변수 | 값 | 분류 | 주입 범위 | 판정 | 근거 |
|---|---|---|---|---|---|
| `GOOSE_PROVIDER` | 공급자 이름 | 설정 | — (사용자 설정) | 부적합 — 실행 전 설정값 | [`goose-cli/src/cli.rs`](https://github.com/block/goose/blob/main/crates/goose-cli/src/cli.rs) |
| `AGENT_SESSION_ID` | 세션 ID | 실행 식별 | developer 셸 도구가 실행하는 명령 | 부적합 — 아래 참조 | [`platform_extensions/developer/shell.rs`](https://github.com/block/goose/blob/main/crates/goose/src/agents/platform_extensions/developer/shell.rs) |
| `GOOSE_TERMINAL`, `AGENT` | `1`, `goose` | 실행 식별 | retry 검증 명령 한정 | 부적합 — 범위가 지나치게 좁음 | [`agents/retry.rs`](https://github.com/block/goose/blob/main/crates/goose/src/agents/retry.rs) |

`AGENT_SESSION_ID`는 실제 주입 마커이지만 두 가지 문제가 있습니다.

1. **이름이 Goose 전용이 아닙니다.** 어떤 벤더에도 묶이지 않은 일반명이라 다른 제품이 같은 이름을 쓸 수 있고, 그러면 오탐이 됩니다.
2. **사용자 셸 rc에 상주합니다.** `goose term init <shell>`은 셸 설정 파일에 `export AGENT_SESSION_ID=...`를 심습니다. 셸을 열 때마다 설정되므로 Goose가 실행 중이 아닐 때도 존재합니다.

`GOOSE_TERMINAL=1`과 `AGENT=goose`는 Goose 전용이지만 `retry.rs`의 재시도 검증 명령에만 붙습니다. 일반 도구 실행 경로에는 없어, 이것만으로는 거의 모든 Goose 세션을 놓칩니다.

## 결론

Goose는 현재 감지하지 않습니다. `AGENT_SESSION_ID`가 rc 파일에 상주하는지 실측으로 확인되고, Goose가 범용 마커를 추가한다면 재검토할 수 있습니다.
