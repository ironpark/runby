---
title: Bun
slug: bun
research_date: 2026-08-31
open_source: true
repository: https://github.com/oven-sh/bun
product_type: task_runner
executes_agents: []
runtime_test_required: false
runtime_test_reason: 2026-08-31 bun 1.3.5로 `bun run`을 직접 실행해 npm_config_user_agent의 첫 토큰이 bun/이고 INIT_CWD가 설정되지 않는 것을 관측함
---

# Bun

Bun의 `bun run`도 npm 계약을 호환 구현하지만 **완전히 같지는 않습니다.** 계열 공통 사항은 [`README.md`](README.md)에 있습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 실행 주체 식별 | 출처 |
|---|---|---|---|---|
| `npm_config_user_agent` | `bun/<버전> npm/? node/<버전> …` | 실행 식별 | 적합 — **첫 토큰이 `bun/`** | 실측 |
| `npm_lifecycle_event` | 스크립트 이름 | 실행 식별 | 보조 신호 — 계열 공통 | 실측 |
| `npm_execpath` | 절대 경로 (`~/.bun/bin/bun`) | 상태·컨텍스트 | 보조 신호 — 값이 경로라 판정에는 미사용 | 실측 |
| `INIT_CWD` | — | — | **설정하지 않음** | 실측 |

## 2026-08-31 실측

```
npm_config_user_agent=bun/1.3.5 npm/? node/v24.3.0 darwin arm64
npm_execpath=/Users/…/.bun/bin/bun
npm_lifecycle_event=dump
```

`INIT_CWD`가 출력에 없습니다. npm·pnpm과 달리 Bun은 이 변수를 설정하지 않으므로, **`INIT_CWD`를 계열 공통 마커로 삼으면 Bun을 놓칩니다.** `runby`가 `npm_config_user_agent`를 유일한 마커로 삼는 이유 중 하나입니다.

## 실행 파일 이름

`bun`은 자기 이름의 단일 바이너리로 돌고 이름이 충분히 특정적이므로 `Executables`에 채웁니다.

## 결론

`npm_config_user_agent`의 첫 토큰이 `bun/`인 것이 판정 기준입니다. `INIT_CWD` 부재는 Bun의 정상 동작이며 감지 실패의 근거가 아닙니다.
