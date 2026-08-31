---
title: npm
slug: npm
research_date: 2026-08-31
open_source: true
repository: https://github.com/npm/cli
product_type: task_runner
executes_agents: []
runtime_test_required: false
runtime_test_reason: 2026-08-31 npm 11.13.0으로 `npm run`을 직접 실행해 npm_config_user_agent·npm_lifecycle_event·INIT_CWD·npm_execpath의 실제 값을 관측함
---

# npm

npm이 `package.json`의 스크립트를 실행할 때 자식 프로세스에 주입하는 변수입니다. 계열 전체가 공유하는 프로토콜과 도구 간 차이는 [`README.md`](README.md)에 정리되어 있습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 주체 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `npm_config_user_agent` | `npm/<버전> node/<버전> <platform> <arch> ...` | 실행 식별 | HTTP User-Agent 헤더 구성값을 스크립트에도 노출 | 적합 — **첫 토큰이 `npm/`인 경우에 한함.** pnpm·Bun이 문자열 뒤쪽에 `npm/?`을 붙이므로 부분 문자열 검색은 오탐 | [npm config: `user-agent` 기본 템플릿](https://docs.npmjs.com/cli/v11/using-npm/config), [npm config: `npm_config_` 접두사 규칙](https://docs.npmjs.com/cli/v11/using-npm/config) |
| `npm_lifecycle_event` | 스크립트/라이프사이클 단계 이름 | 실행 식별 | "실행 중인 사이클 단계" | 보조 신호 — 계열 세 도구가 모두 설정하므로 단독으로는 도구를 특정하지 못함 | [npm scripts](https://docs.npmjs.com/cli/v11/using-npm/scripts) |
| `INIT_CWD` | 절대 경로 | 상태·컨텍스트 | `npm run`을 실행한 시점의 디렉터리 | 보조 신호 — Bun은 설정하지 않으므로 계열 공통 신호가 아님 | [npm scripts](https://docs.npmjs.com/cli/v11/using-npm/scripts) |
| `npm_execpath` | 절대 경로 | 상태·컨텍스트 | 실행 중인 npm CLI 경로 | 보조 신호 — 값이 경로라 판정 근거로는 쓰지 않음 | 실측 |
| `npm_lifecycle_script` | 셸 명령문 전체 | — | 실행될 스크립트 본문 | **사용 금지** — 임의의 셸 명령문이라 자격증명을 포함할 수 있음. `Evidence`·`Extra` 어디에도 넣지 않음 | 실측 |
| `npm_package_*` | package.json에서 파생된 값들 | 설정 | 패키지 메타데이터 | 부적합 — 설정값이며 실행 주체 식별과 무관 | [npm scripts](https://docs.npmjs.com/cli/v11/using-npm/scripts) |

## 2026-08-31 실측

```
INIT_CWD=/private/tmp/pmt
npm_config_user_agent=npm/11.13.0 node/v26.1.0 darwin arm64 workspaces/false
npm_execpath=/Users/…/lib/node_modules/npm/bin/npm-cli.js
npm_lifecycle_event=dump
```

## 실행 파일 이름

`Executables`는 **비워 둡니다.** npm 스크립트를 실행하는 실제 프로세스는 `node`이며(`npm_execpath`가 `npm-cli.js`를 가리키는 데서 확인됨), `node`는 무관한 프로세스를 잘못 라벨링할 만큼 일반적인 이름입니다. `npm` 자체는 셸 래퍼라 조상 체인에 남지 않을 수 있습니다.

## 결론

`npm_config_user_agent`의 첫 토큰이 `npm/`인지가 유일하고 신뢰할 수 있는 판정 기준입니다. `npm_lifecycle_event`는 마커가 결정된 뒤 작업 이름으로 읽습니다.

부재가 "npm이 아님"을 뜻하지는 않습니다 — `npm exec`나 `npx` 경로, 또는 사용자가 스크립트 안에서 `env -i`로 환경을 비운 경우에는 나타나지 않을 수 있습니다.
