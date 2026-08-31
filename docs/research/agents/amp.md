---
title: Amp
slug: amp
research_date: 2026-08-30
open_source: false
repository: null
product_type: agent_harness
model_source: multi-vendor
executes_agents: []
runtime_test_required: true
runtime_test_reason: 로컬 CLI와 Orb 및 Orb 관리형 서비스의 환경 주입 범위를 각각 확인해야 함
---

# Amp

Amp의 최신 공식 문서에는 로컬 CLI가 실행한 모든 자식 프로세스에 공통으로 주입되는 전용 실행 마커가 공개되어 있지 않습니다. 다만 원격 실행 환경인 Orb에는 `AMP_ORB=1`이 설정되고, Orb의 관리형 서비스에는 `AMP_THREAD_ID`가 추가로 주입됩니다. 따라서 로컬 CLI와 Orb를 구분해서 판단해야 합니다.

## 공식 확인 환경변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 프로세스 실행 주체 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `AMP_ORB` | 정수형 불리언 (`1`) | 실행 식별 | 현재 환경이 Amp가 만든 Orb임을 표시 | 적합 — 값이 `1`이면 Amp Orb 내부라는 강한 신호. 단, 로컬 Amp CLI에는 적용되지 않고 Orb 안의 모든 프로세스가 상속할 수 있음 | [Amp Portals — Handle Development Sign-In](https://ampcode.com/docs/orbs/portals#handle-development-sign-in) |
| `AMP_THREAD_ID` | 스레드 ID 문자열 | 상태/컨텍스트 | `.amp/services.yaml`로 실행한 모든 관리형 서비스에 소유 스레드 ID 제공 | 적합 — Amp가 관리하는 Orb 서비스에서는 직접 주입되는 강한 신호. 일반 Agent 도구 프로세스나 로컬 CLI 전체에 대한 계약은 아님 | [Amp Portals — Environment variables](https://ampcode.com/docs/orbs/portals) |
| `PORT` | 포트 번호 (`1`–`65535`) | 상태/컨텍스트 | Orb 관리형 서비스에 할당된 로컬 포트 | 부적합 — Amp가 주입하지만 매우 일반적인 이름이며 다른 실행 환경에서도 흔함 | [Amp Portals — Service configuration](https://ampcode.com/docs/orbs/portals) |
| `PUBLIC_URL` | HTTP(S) URL 문자열 | 상태/컨텍스트 | Portal 링크가 있는 Orb 서비스에 외부 URL 제공 | 보조 — `AMP_ORB` 또는 `AMP_THREAD_ID`와 함께만 의미가 있으며 단독으로는 일반적인 변수 | [Amp Portals — Service configuration](https://ampcode.com/docs/orbs/portals) |
| `AMP_REMOTE_CONTROL_TERMINAL` | 불리언 (`1` 또는 `0`) | 설정 | 웹·앱에서 실행 중인 CLI의 터미널 제어 허용 여부 | 부적합 — 사용자 설정이며 `amp` 실행 전부터 셸에 존재할 수 있음 | [Amp Remote Control — Remote Terminal Access](https://ampcode.com/docs/cli/remote-control#remote-terminal-access) |
| `AMP_DISABLE_AMP_THREAD_TRAILER` | 정수형 불리언 (`1`) | 설정 | Git 커밋의 `Amp-Thread-ID` 트레일러 비활성화 | 부적합 — 커밋 설정이며 실행 마커가 아님 | [Amp Configuration — Settings](https://ampcode.com/docs/cli/settings#settings) |
| `AMP_DISABLE_AMP_COAUTHOR_TRAILER` | 정수형 불리언 (`1`) | 설정 | Git 커밋의 Amp 공동 작성자 트레일러 비활성화 | 부적합 — 커밋 설정이며 실행 마커가 아님 | [Amp Configuration — Settings](https://ampcode.com/docs/cli/settings#settings) |
| `AMP_SKIP_UPDATE_CHECK` | 정수형 불리언 (`1`) | 설정 | 업데이트 확인 비활성화 | 부적합 — 일반 실행 설정이며 자식 프로세스 주입 계약이 없음 | [Amp Configuration — Updates](https://ampcode.com/docs/cli/settings#settings) |
| `AMP_FORCE_BEL` | 존재 여부 | 설정 | 완료·입력 대기 알림에 호스트 오디오 대신 터미널 벨 사용 | 부적합 — 사용자 설정이며 실행 주체나 상태를 나타내지 않음 | [Amp Configuration — Notifications](https://ampcode.com/docs/cli/settings#settings) |
| `HTTP_PROXY` | 프록시 URL 문자열 | 설정 | Amp CLI의 HTTP 프록시 | 부적합 — Node.js와 여러 프로그램이 공유하는 일반 변수 | [Amp Configuration — Proxies and Certificates](https://ampcode.com/docs/cli/settings#proxies-and-certificates) |
| `HTTPS_PROXY` | 프록시 URL 문자열 | 설정 | Amp CLI의 HTTPS 프록시 | 부적합 — Node.js와 여러 프로그램이 공유하는 일반 변수 | [Amp Configuration — Proxies and Certificates](https://ampcode.com/docs/cli/settings#proxies-and-certificates) |
| `EDITOR` | 실행 파일명 또는 경로 | 설정 | `Ctrl+G`로 프롬프트를 외부 편집기에서 열 때 사용할 편집기 | 부적합 — 셸 전반에서 사용하는 표준 변수 | [Amp CLI Keybindings — Writing Prompts](https://ampcode.com/docs/cli/keybindings#writing-prompts) |

## 실행 주체 감지에 관한 결론

- Orb에서는 `AMP_ORB=1`을 1차 신호로 사용하고, 관리형 서비스라면 `AMP_THREAD_ID`를 스레드 컨텍스트로 함께 기록할 수 있습니다.
- 로컬 Amp CLI에 대해서는 2026-08-30 현재 공식 문서가 보장하는 실행 식별 환경변수가 없습니다. 설정 변수만으로 Amp가 현재 프로세스를 실행했다고 판단하면 오탐이 발생합니다.
- 공식 문서의 `$AMP_USER_EMAIL`은 Portal 설정의 URL·제목·설명에서 치환되는 **플레이스홀더**이며, 서비스 프로세스에 주입된다고 명시된 환경변수가 아니므로 표에서 제외했습니다.
- 공개 Amp 스레드에서 보이는 `AMP_CURRENT_THREAD_ID` 사례는 사용자 콘텐츠이지 제품의 공식 문서나 공개 공식 소스 계약이 아닙니다. 따라서 안정적인 감지 변수로 채택하지 않았습니다.

## 공식 문서

- [Amp 문서](https://ampcode.com/docs)
- [Amp CLI Configuration](https://ampcode.com/docs/cli/settings)
- [Amp Execute Mode](https://ampcode.com/docs/cli/execute-mode)
- [Amp Remote Control](https://ampcode.com/docs/cli/remote-control)
- [Amp Orbs: Portals](https://ampcode.com/docs/orbs/portals)
