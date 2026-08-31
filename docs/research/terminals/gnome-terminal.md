---
title: GNOME Terminal
slug: gnome-terminal
research_date: 2026-08-31
open_source: true
repository: https://gitlab.gnome.org/GNOME/gnome-terminal
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: 소스 코드는 GNOME Terminal 고유 변수와 VTE가 강제로 설정하는 변수를 명확히 구분하지만, 배포판 패치나 실제 버전별로 필터 목록·값 형식이 달라질 수 있어 실제 GNOME Terminal과 Tilix·Terminator 등 다른 VTE 기반 터미널에서 `env` 출력을 직접 비교해 GNOME_TERMINAL_* 유무만으로 제품이 갈리는지 확인해야 함
---

# GNOME Terminal

GNOME Terminal 자체 공식 문서(Help/manpage)는 자식 프로세스에 주입하는 환경변수 계약을 표 형태로 명시하지 않습니다. 이 문서의 근거는 대부분 공식 소스 코드(`gitlab.gnome.org/GNOME/gnome-terminal`, `gitlab.gnome.org/GNOME/vte`)입니다. 소스를 확인한 결과 GNOME Terminal이 자식 프로세스에 보장하는 것은 두 가지로 나뉩니다. 하나는 **VTE 라이브러리가 강제로 설정하는 값**(`VTE_VERSION`, `TERM`, `COLORTERM`)으로, VTE를 임베드하는 모든 터미널에 공통으로 나타납니다. 다른 하나는 **GNOME Terminal 프로세스 자신이 설정하는 값**(`GNOME_TERMINAL_SCREEN`, `GNOME_TERMINAL_SERVICE`)으로, gnome-terminal-server의 D-Bus 단일 인스턴스 아키텍처에서만 나타납니다. GNOME Terminal은 `TERM_PROGRAM`을 설정하지 않으며, 소스 전체를 검색해도 해당 변수를 다루는 코드가 없습니다.

## 터미널 식별 신호

`TERM_PROGRAM`이 없기 때문에, GNOME Terminal이 직접 만든 프로세스임을 가리키는 가장 신뢰도 높은 존재 마커는 `GNOME_TERMINAL_SCREEN`(또는 `GNOME_TERMINAL_SERVICE`)입니다. 이 값들은 gnome-terminal-server가 새 탭/창을 열 때 사용하는 D-Bus 통신 정보 자체이므로, 다른 터미널이 우연히 같은 이름으로 임의의 값을 설정할 이유가 없습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `GNOME_TERMINAL_SCREEN` | D-Bus 오브젝트 경로 문자열 (`/org/gnome/Terminal/screen/<uuid, '-'를 '_'로 치환>`) | 실행 식별 | 이 프로세스를 소유한 gnome-terminal 화면(탭)의 D-Bus 오브젝트 경로 | 적합 — GNOME Terminal 소스가 `TERMINAL_ENV_SCREEN` 상수로 정의하고 `terminal_app_dup_screen_object_path()`가 생성한 값만 넣으며, 수신 측(`terminal-options.cc`)도 `g_variant_is_object_path()`로 유효한 D-Bus 오브젝트 경로인지 검증함 | [공식 소스: `terminal-defines.hh#L51`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-defines.hh#L51), [`terminal-screen.cc#L1955-1958`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-screen.cc#L1955-L1958), [`terminal-app.cc#L1214-1223`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-app.cc#L1214-L1223) |
| `GNOME_TERMINAL_SERVICE` | D-Bus unique name 문자열 (예: `:1.23`) | 실행 식별 | 이 화면을 소유한 gnome-terminal-server 인스턴스의 D-Bus 고유 이름 | 적합 — `g_dbus_connection_get_unique_name()`으로 얻은 실제 서비스 연결 이름을 넣고, 수신 측도 `g_dbus_is_unique_name()`으로 검증함. `GNOME_TERMINAL_SCREEN`과 결합하면 어떤 서버의 어떤 탭인지까지 특정됨 | [공식 소스: `terminal-defines.hh#L50`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-defines.hh#L50), [`terminal-screen.cc#L1954-1955`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-screen.cc#L1954-L1955), [`terminal-options.cc#L1003-1011`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-options.cc#L1003-L1011) |

두 변수 모두 gnome-terminal-server의 서버/클라이언트 D-Bus 아키텍처(단일 인스턴스가 모든 창·탭을 관리하고, 새 `gnome-terminal` 호출은 이 서버에 탭 생성을 요청)에서 나온 **살아있는 IPC 핸들**입니다. 즉 값 자체가 "이 프로세스는 GNOME Terminal이 만들었다"는 서명이자, 동시에 "이 D-Bus 이름·오브젝트 경로로 그 GNOME Terminal 인스턴스에 말을 걸 수 있다"는 실시간 연결 정보이기도 합니다. 소스에서 이 아키텍처가 도입된 정확한 GNOME Terminal 버전(예: 3.12 서버/클라이언트 분리)은 이번 조사에서 확인한 공식 changelog에 명시적으로 기록되어 있지 않아 단정하지 않습니다. 현재 `main` 브랜치 기준 동작만 확정된 사실로 다룹니다.

## VTE 기반 터미널의 공유 신호

`VTE_VERSION`은 GNOME Terminal이 아니라 **VTE 라이브러리 자체**가 pty에서 자식 프로세스를 스폰할 때 강제로 설정하는 값입니다. VTE는 GNOME Terminal 외에도 다수의 터미널 에뮬레이터가 위젯으로 임베드하는 라이브러리이므로, `VTE_VERSION`은 **제품이 아니라 계열(family)**을 가리킵니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `VTE_VERSION` | 정수 문자열 (`MAJOR*10000 + MINOR*100 + MICRO`, 예: VTE 0.85.0 → `8500`) | 실행 식별 | 자식 프로세스를 실행 중인 VTE 라이브러리의 버전 | 보조 신호 — VTE 기반 터미널이라는 사실은 강하게 확인되지만, VTE를 쓰는 어떤 제품인지는 특정하지 못함. 소스 주석이 "envp로 덮어쓰기를 허용하지 않고 항상 스스로 설정한다"고 명시할 만큼 강제되는 값이라 위조 저항성은 높음 | [공식 소스: `vtedefines.hh#L132`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/vtedefines.hh#L132), [`spawn.cc#L279-281`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/spawn.cc#L279-L281) |
| `TERM` | 문자열 (`xterm-256color`, `VTE_TERMINFO_NAME` 상수) | 설정 | 기본 terminfo 종류 지정. envp로 재정의 가능 | 부적합 — VTE가 기본값을 강제하지만 다른 수많은 터미널도 같은 값을 쓰고, 호출자가 자유롭게 덮어쓸 수 있음 | [공식 소스: `vtedefines.hh#L134`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/vtedefines.hh#L134), [`spawn.cc#L251-252`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/spawn.cc#L251-L252) |
| `COLORTERM` | 문자열 (`truecolor`) | 설정 | true-color 지원 표시. `VTE_VERSION`과 같은 코드 경로에서 "항상 스스로 설정" 방식으로 강제됨 | 부적합 — VTE 계열 전체가 공유하는 값이며, 다른 터미널(예: iTerm2, Windows Terminal 등)도 동일 값을 관행적으로 씀 | [공식 소스: `spawn.cc#L279-281`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/spawn.cc#L279-L281) |

VTE 소스(`src/app/app.cc`)의 환경변수 정리 로직 자체가 이 구분을 명확히 보여줍니다. 새 세션을 스폰하기 전 지워야 할 "다른 터미널이 남긴" 변수 목록에 `TERM`, `VTE_VERSION`, `WINDOWID`와 함께 `GNOME_TERMINAL_`으로 시작하는 접두사 전체가 별도로 나열되어 있고, 주석은 "gnome-terminal에서 그대로 가져온 목록"이라고 밝힙니다. 즉 VTE 개발자들 스스로도 `GNOME_TERMINAL_*` 접두사를 "GNOME Terminal이 남긴 흔적"으로 취급해 다른 VTE 기반 터미널에서 지워야 할 대상으로 분류하고 있습니다.

- [공식 소스: `app.cc#L260-268` (COLORTERM/TERM/VTE_VERSION/WINDOWID 등 개별 변수 제거 목록)](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/app/app.cc#L260-L268)
- [공식 소스: `app.cc#L270-283` (`GNOME_TERMINAL_` 및 다른 터미널 접두사 제거 목록)](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/app/app.cc#L270-L283)
- [공식 소스: `terminal-client-utils.cc#L270-296` (gnome-terminal 측의 동일한 `GNOME_TERMINAL_` 접두사 필터)](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-client-utils.cc#L270-L296)

GNOME 공식 위키의 VTE 페이지는 VTE를 사용하는 프로그램 목록에 GNOME Terminal 외에 **XFCE Terminal, guake, terminator, sakura, evilvte, ROX Terminal, vala-terminal**을 명시적으로 들고 있습니다. Tilix, MATE Terminal 등은 일반적으로 VTE 기반으로 알려져 있지만 이번 조사에서 확인한 공식 GNOME 소스·위키에는 나열되어 있지 않아 이 목록에는 포함하지 않았습니다.

**`runby`에 대한 권장 사항**: `VTE_VERSION`만 있고 `GNOME_TERMINAL_SCREEN`/`GNOME_TERMINAL_SERVICE`가 없으면, 결과를 `gnome-terminal`로 단정하지 말고 "VTE 기반 터미널"이라는 일반화된 계열로 보고해야 합니다. `VTE_VERSION`만으로 GNOME Terminal을 추정하는 것은 Tilix·Terminator·XFCE Terminal 등 다른 제품을 GNOME Terminal로 오탐할 위험이 있어 과신입니다. `GNOME_TERMINAL_*` 변수가 함께 있을 때만 제품을 GNOME Terminal로 좁혀야 합니다.

## 상태·컨텍스트 변수

GNOME Terminal이 자식 프로세스 환경을 구성할 때 `NCURSES_NO_UTF8_ACS=1`, 프록시 변수(`http_proxy` 등), `TERMINFO_DIRS` 보강 같은 값도 다루지만, 이는 터미널 기능 설정이거나 배관(plumbing)일 뿐 실행 주체 식별과 무관해 표에서 제외했습니다. 아래는 실행 주체 감지에 참고할 만한 상태 신호입니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `WINDOWID` | 정수 문자열 (X11 창 ID) | 상태·컨텍스트 | 터미널이 실행 중인 X11 창의 ID | 부적합 — X11 하위의 여러 터미널이 공통으로 설정하는 관행적 값이며, VTE 자체도 "다른 터미널이 남긴 값"으로 취급해 새 세션 스폰 시 제거 대상 목록에 넣음. Wayland 세션에서는 애초에 의미가 없음 | [공식 소스: `app.cc#L253` (VTE의 제거 목록)](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/app/app.cc#L253), [`terminal-client-utils.cc#L254` (gnome-terminal의 동일 제거 목록)](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-client-utils.cc#L254) |

이 조사에서는 GNOME Terminal 또는 VTE가 `WINDOWID`를 직접 *설정*하는 코드는 확인하지 못했습니다. 두 소스 모두에서 `WINDOWID`는 오직 "새 세션을 스폰하기 전에 지워야 할 외부 변수" 목록에서만 등장하므로, 이 변수가 존재한다면 그 값은 X11 창 시스템이나 다른 도구가 남긴 것이지 GNOME Terminal/VTE가 보장하는 계약이 아닙니다.

## 상속과 잔존

`GNOME_TERMINAL_SCREEN`·`GNOME_TERMINAL_SERVICE`는 특정 시점에 살아있던 D-Bus 핸들을 값으로 담고 있을 뿐인 평범한 환경변수이므로, 다른 환경변수와 똑같이 모든 자식 프로세스에 상속되고 그 프로세스가 원래 터미널보다 오래 살아남으면 **가리키는 대상이 이미 사라진 값**으로 잔존합니다. 이는 이 변수들을 실행 주체 감지에 쓸 때 구조적인 한계입니다.

- **SSH**: OpenSSH 클라이언트는 기본 `SendEnv`(및 서버의 기본 `AcceptEnv`)에 `LANG`, `LC_*` 계열만 포함하므로, `GNOME_TERMINAL_SCREEN`·`GNOME_TERMINAL_SERVICE`·`VTE_VERSION`은 관리자가 별도로 `SendEnv`/`AcceptEnv`를 확장하지 않는 한 원격 세션으로 전달되지 않습니다. 다만 `TERM`은 `SendEnv` 메커니즘이 아니라 SSH 프로토콜의 pty 요청 자체에 포함되어 전달되므로 동작이 다릅니다. 즉 SSH로 로그인한 세션에서 `GNOME_TERMINAL_*`이 보인다면 그 값은 SSH가 전달한 것이 아니라 원격 셸이 그 안에서 우연히 설정한(또는 위조한) 값일 가능성이 높습니다.
- **tmux/screen**: tmux 서버는 그 서버를 최초로 띄운 클라이언트의 환경을 그대로 물려받아 보관하며, 이후 붙는 클라이언트나 새로 만든 pane은 서버 시작 시점의 환경을 기본으로 물려받습니다. tmux의 `update-environment` 옵션은 새 세션 생성·재접속 시 일부 변수(`DISPLAY`, `SSH_AUTH_SOCK`, `SSH_CONNECTION`, `WINDOWID` 등 기본 목록)만 갱신하며, 이 기본 목록에는 `GNOME_TERMINAL_SCREEN`/`GNOME_TERMINAL_SERVICE`/`VTE_VERSION`이 포함되어 있지 않습니다. 따라서 원래 GNOME Terminal 창을 닫은 뒤에도 tmux pane은 이제는 존재하지 않는 D-Bus 오브젝트를 가리키는 `GNOME_TERMINAL_SCREEN` 값을 계속 노출할 수 있습니다. `runby`가 이 값을 "현재 GNOME Terminal이 살아있다"는 증거로 쓰면 안 되는 이유입니다.
- **데몬화·분리된 프로세스**: `nohup`, `setsid`, 백그라운드 서비스 등으로 터미널에서 분리되어 원래 터미널보다 오래 실행되는 프로세스는 시작 시점의 환경을 그대로 유지합니다. 원래 GNOME Terminal 창이 닫혀도 이 값들은 프로세스가 종료될 때까지 남습니다.
- **위조 가능성**: 어떤 환경변수든 `export GNOME_TERMINAL_SCREEN=...`으로 임의의 값을 직접 설정할 수 있습니다. `g_dbus_is_unique_name()`/`g_variant_is_object_path()` 같은 형식 검증은 GNOME Terminal 자신이 자기 자식 프로세스를 다시 처리할 때만 수행되며, `runby`처럼 외부에서 값을 읽는 라이브러리는 형식이 맞아도 그 D-Bus 이름·경로가 실제로 응답하는지까지 확인하지 않는 한 신뢰할 수 없습니다.

결론적으로 라이브러리가 정직하게 주장할 수 있는 것은 "이 값들은 어느 시점에 GNOME Terminal(또는 VTE 기반 터미널)이 이 프로세스 트리를 만들었거나 그 흔적을 남겼다는 정황 증거"까지입니다. "지금 이 순간 GNOME Terminal이 이 프로세스를 소유하고 있다"는 보장은 D-Bus로 실제 핸들에 접속해 응답을 받아야만 가능하며, 이는 환경변수만 읽는 정적 감지의 범위를 벗어납니다.

## 실행 주체 감지에 관한 결론

1. `TERM_PROGRAM`이 없으므로 GNOME Terminal 자체를 가리키는 가장 신뢰도 높은 존재 마커는 `GNOME_TERMINAL_SCREEN`(및 `GNOME_TERMINAL_SERVICE`)입니다. 이 값들은 GNOME Terminal 소스에서만 정의·주입되며, 형식이 유효한 D-Bus 오브젝트 경로/고유 이름인지까지 자체적으로 검증됩니다.
2. `VTE_VERSION`은 VTE 라이브러리가 강제로 설정하는 값이라 위조 저항성은 높지만, GNOME Terminal 전용이 아니라 XFCE Terminal·guake·terminator 등 VTE를 임베드하는 모든 터미널이 공유합니다. 따라서 `runby`는 `VTE_VERSION`만 있고 `GNOME_TERMINAL_*`이 없을 때 "GNOME Terminal"로 단정하지 말고 "VTE 기반 터미널(제품 미상)"로 보고해야 합니다.
3. `GNOME_TERMINAL_*` 두 값이 함께 있으면 GNOME Terminal이라는 판정 신뢰도는 높지만, 값 자체가 살아있는 D-Bus 핸들을 가리킬 뿐인 평범한 환경변수이므로 상속·잔존·위조에 취약합니다. SSH는 기본 설정상 이 값들을 전달하지 않고, tmux/screen pane은 서버가 시작된 시점의 값을 그대로 들고 있어 원래 터미널이 사라진 뒤에도 잔존할 수 있으며, 데몬화된 프로세스는 종료 전까지 값을 유지합니다.
4. 따라서 `runby`는 `GNOME_TERMINAL_SCREEN`/`GNOME_TERMINAL_SERVICE` 존재를 "이 프로세스 트리가 GNOME Terminal에 의해 만들어졌거나 그 환경을 물려받았다"는 강한 정황 증거로 다루되, "지금 그 GNOME Terminal 인스턴스가 살아있다"는 보장으로 확대 해석해서는 안 됩니다.

## 공식 문서

- [GNOME Terminal 공식 소스 저장소](https://gitlab.gnome.org/GNOME/gnome-terminal)
- [VTE 공식 소스 저장소](https://gitlab.gnome.org/GNOME/vte)
- [공식 소스: `terminal-defines.hh`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-defines.hh)
- [공식 소스: `terminal-screen.cc`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-screen.cc)
- [공식 소스: `terminal-app.cc`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-app.cc)
- [공식 소스: `terminal-options.cc`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-options.cc)
- [공식 소스: `terminal-client-utils.cc`](https://gitlab.gnome.org/GNOME/gnome-terminal/-/blob/fce542e9e08abe16719ffeef2e7c122bd7c8d17f/src/terminal-client-utils.cc)
- [공식 소스: `vtedefines.hh`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/vtedefines.hh)
- [공식 소스: `spawn.cc`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/spawn.cc)
- [공식 소스: `app/app.cc`](https://gitlab.gnome.org/GNOME/vte/-/blob/1f53b4811edf8832439856d5ee359e4b15ca4f2b/src/app/app.cc)
- [GNOME 공식 위키: Apps/Terminal/VTE (VTE 사용 프로그램 목록)](https://wiki.gnome.org/Apps/Terminal/VTE)
- [GNOME 공식 위키: Apps/Terminal/FAQ](https://wiki.gnome.org/Apps/Terminal/FAQ)
- [OpenSSH `ssh_config` 매뉴얼 — `SendEnv` 기본값](https://man.openbsd.org/ssh_config#SendEnv)
- [OpenSSH `sshd_config` 매뉴얼 — `AcceptEnv` 기본값](https://man.openbsd.org/sshd_config#AcceptEnv)
- [tmux 매뉴얼 — `update-environment` 기본값](https://man.openbsd.org/tmux.1#update-environment)
