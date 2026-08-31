---
title: pnpm
slug: pnpm
research_date: 2026-08-31
open_source: true
repository: https://github.com/pnpm/pnpm
product_type: task_runner
executes_agents: []
runtime_test_required: false
runtime_test_reason: 2026-08-31 pnpm 11.24.0으로 `pnpm run`을 직접 실행해 npm_config_user_agent의 첫 토큰이 pnpm/이고 npm/?이 뒤따라 붙는 것을 관측함
---

# pnpm

pnpm은 npm이 정의한 스크립트 환경변수 계약을 호환 목적으로 구현합니다. 계열 공통 사항은 [`README.md`](README.md)에 있습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 실행 주체 식별 | 출처 |
|---|---|---|---|---|
| `npm_config_user_agent` | `pnpm/<버전> npm/? node/<버전> …` | 실행 식별 | 적합 — **첫 토큰이 `pnpm/`** | 실측 |
| `npm_lifecycle_event` | 스크립트 이름 | 실행 식별 | 보조 신호 — 계열 공통 | 실측 |
| `INIT_CWD` | 절대 경로 | 상태·컨텍스트 | 보조 신호 | 실측 |
| `npm_execpath` | 절대 경로 (`/opt/homebrew/bin/pnpm`) | 상태·컨텍스트 | 보조 신호 — 값이 경로라 판정에는 미사용 | 실측 |

## 2026-08-31 실측

```
INIT_CWD=/private/tmp/pmt
npm_config_user_agent=pnpm/11.24.0 npm/? node/v26.1.0 darwin arm64
npm_execpath=/opt/homebrew/bin/pnpm
npm_lifecycle_event=dump
```

**`npm/?`이 값 안에 들어 있다는 점이 이 문서의 핵심입니다.** `strings.Contains(ua, "npm")`으로 판정하면 pnpm 스크립트가 npm으로 오보고됩니다. 반드시 접두사 비교여야 합니다.

## 실행 파일 이름

`pnpm`은 자기 이름의 실행 파일로 돌고 그 이름이 충분히 특정적이므로 `Executables`에 채웁니다.

## 결론

`npm_config_user_agent`의 첫 토큰이 `pnpm/`인 것이 판정 기준입니다. 나머지는 npm과 같습니다.
