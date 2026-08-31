---
title: Konsole
slug: konsole
research_date: 2026-08-31
open_source: true
repository: https://invent.kde.org/utilities/konsole
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: D-Bus 오브젝트 경로가 실제로 생존한 Konsole 세션을 가리키는지, tmux/SSH를 거친 뒤에도 값이 잔존만 하는지는 실제 Konsole·tmux·SSH 조합에서 확인이 필요함
---

# Konsole

Konsole은 셸 프로세스에 여러 `KONSOLE_*` 환경변수를 주입하지만, 이 중 어느 것도 "이 프로세스는 반드시 Konsole이 지금 이 순간 실행했다"를 공식적으로 보증하는 안정적 공개 계약으로 문서화되어 있지 않습니다. 공식 핸드북은 `KONSOLE_DBUS_SESSION` 등을 `qdbus` 스크립팅 예제로만 설명하며, 나머지 변수는 공식 소스([`invent.kde.org/utilities/konsole`](https://invent.kde.org/utilities/konsole), GitHub 미러 [`KDE/konsole`](https://github.com/KDE/konsole))에서만 확인할 수 있습니다. 또한 Konsole은 **`TERM_PROGRAM`을 설정하지 않습니다** — 저장소 전체를 검색해도 `TERM_PROGRAM`을 다루는 코드가 없으므로, 다른 다수 터미널이 쓰는 이 관용적 마커에 의존할 수 없습니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `KONSOLE_VERSION` | 점을 제거한 숫자 문자열 (예: `18.04.12` → `180412`, `18.08.0` → `180800`) | 실행 식별 | 셸 프로그램이 산술 비교로 Konsole 버전·기능 지원 여부를 판단하도록 제공 | 적합 — Konsole 프로필의 `Environment` 속성이 적용될 때마다 코드가 직접 계산해 주입하는 값으로, 존재 자체가 Konsole(또는 konsolepart 임베디드 인스턴스) 세션임을 강하게 시사함. 다만 프로필 설정으로 덮어써질 수 있음 | [`SessionManager::applyProfile`](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/session/SessionManager.cpp#L188-L201) |
| `KONSOLE_DBUS_SESSION` | D-Bus 오브젝트 경로 (`/Sessions/<정수 ID>`) | 실행 식별 | 현재 셸을 만든 Konsole 세션의 D-Bus 오브젝트 경로 노출 | 적합 — Konsole이 세션 생성 시 직접 주입. 값 자체보다 이 오브젝트가 실제로 응답하는지가 더 강한 신호(아래 [상속과 잔존](#상속과-잔존) 참고) | [`Session::run` D-Bus 블록](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/session/Session.cpp#L610-L619) |
| `KONSOLE_DBUS_SERVICE` | D-Bus 서비스(버스) 이름 (예: `:1.23` 또는 `org.kde.konsole-<pid>`) | 실행 식별 | `KONSOLE_DBUS_SESSION`을 어느 Konsole 프로세스의 세션 버스에서 찾을지 지정 | 적합 — `KONSOLE_DBUS_SESSION`과 함께 있어야 완전한 D-Bus 목적지가 됨 | [`Session::run` D-Bus 블록](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/session/Session.cpp#L610-L619) |
| `KONSOLE_DBUS_WINDOW` | D-Bus 오브젝트 경로 (`/Windows/<관리자 ID>`) | 실행 식별 | 현재 세션이 속한 Konsole 창의 D-Bus 오브젝트 경로 노출 | 적합 — 창 단위 컨텍스트를 더해주는 보완 신호. 세션 식별자와 결합해 사용하는 것이 안전함 | [`ViewManager::createSession` 계열](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/ViewManager.cpp#L658-L662) |
| `KONSOLE_DBUS_ACTIVATION_COOKIE` | 세션별 무작위 시크릿 문자열 | 실행 식별 | Wayland `xdg-activation` 토큰을 Konsole D-Bus에 요청할 때 제시하는 인증 쿠키 | 보조 신호 — 존재하면 강한 Konsole 증거이지만, Konsole 자신은 이 값을 D-Bus로 노출하지 않도록 명시적으로 "secret" 취급하므로 값 자체를 로그·원격 전송에 노출해서는 안 됨 | [`Session::run` 시크릿 처리](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/session/Session.cpp#L607-L618) |
| `TERM` | 문자열, 기본값 `xterm-256color` | 설정 | 터미널 기능 프로필 설정 | 부적합 — 다수 터미널이 동일 기본값을 쓰고, 프로필에서 자유롭게 덮어쓸 수 있음 | [`Profile` 기본값 테이블](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/profile/Profile.cpp#L71), [Flatpak 샌드박스 강제 지정](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/session/Session.cpp#L478-L482) |
| `COLORTERM` | 문자열, 기본값 `truecolor` | 설정 | 트루컬러 지원 표시 | 부적합 — 다수 터미널이 같은 값을 쓰는 표준 관용 변수이며 프로필에서 덮어쓸 수 있음 | [`Profile` 기본값 테이블](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/profile/Profile.cpp#L71) |

`KONSOLE_PROFILE_NAME`은 공식 소스 저장소 전체 검색(`invent.kde.org/utilities/konsole` GitHub 미러 기준)에서 확인되지 않았습니다. 프로필 이름을 셸에 노출하는 별도 환경변수는 존재하지 않는 것으로 보이며, 이 문서에서는 이를 확인된 변수로 다루지 않습니다.

## 상태·컨텍스트 변수

Konsole은 위 표 외에도 `COLORFGBG`(밝은/어두운 배경 근사값), `SHELL_SESSION_ID`, `WINDOWID`, `PROFILEHOME` 같은 변수를 셸 환경에 추가합니다. 이들은 [`Session::run`](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/session/Session.cpp#L598-L605)에서 함께 주입되지만, `WINDOWID`나 `COLORFGBG`는 X11 환경 전반에서 통용되는 일반 관용 변수이고 `SHELL_SESSION_ID`도 다른 데스크톱 환경에서 쓰이는 이름이라 Konsole 고유 신호로 볼 수 없어 이 문서의 핵심 표에서는 제외했습니다.

## 상속과 잔존

`KONSOLE_*` 변수는 일반 환경변수이므로 자식 프로세스 전체에 상속되며, 이 때문에 존재만으로 "지금 Konsole이 이 프로세스를 실행했다"를 증명하지 못합니다.

- **SSH**: OpenSSH 클라이언트의 `SendEnv` 기본값은 **아무 변수도 전송하지 않음**이며(`TERM`만 pty 요청 시 프로토콜상 항상 전송됨), 서버도 `AcceptEnv`로 명시 허용해야 합니다. 따라서 `KONSOLE_DBUS_*`나 `KONSOLE_VERSION`이 원격 SSH 세션까지 자동으로 전달되는 경우는 거의 없으며, 전달되려면 클라이언트가 `SendEnv KONSOLE_*`를, 서버가 `AcceptEnv KONSOLE_*`를 명시적으로 설정해야 합니다. 다만 `TERM`은 항상 전송되므로 그 값(`xterm-256color`)만으로는 로컬 Konsole인지 원격에서 재설정된 값인지 구분할 수 없습니다.
- **tmux/screen**: tmux 서버는 세션을 처음 만든 클라이언트의 환경을 그대로 물려받고, 이후 attach 시점의 환경으로 자동 갱신하지 않습니다. `update-environment` 세션 옵션에 `KONSOLE_DBUS_SESSION` 같은 이름을 명시적으로 추가하지 않는 한, tmux 팬에 남아있는 `KONSOLE_DBUS_SESSION`/`KONSOLE_DBUS_SERVICE`는 **그 세션을 처음 만들었을 때의 D-Bus 오브젝트를 가리키는 스냅숏**일 뿐입니다. 원래 Konsole 창이 이미 닫혔거나, 다른 창·다른 머신에서 tmux에 재접속했다면 이 경로는 더 이상 존재하지 않는(또는 전혀 다른 세션을 가리키는) D-Bus 목적지가 됩니다. `screen`도 동일하게 세션 시작 시점 환경을 유지하는 구조라 같은 문제가 발생합니다.
- **분리·데몬화된 프로세스**: `nohup`, `disown`, `setsid` 등으로 터미널과 분리되어 오래 살아남는 프로세스는 시작 시점에 상속받은 `KONSOLE_*` 값을 그대로 들고 있지만, 그 시점의 Konsole 창·세션·D-Bus 서비스가 이미 종료되었을 수 있습니다. 값이 있다고 해서 Konsole이 지금도 그 프로세스를 감독하고 있다는 뜻은 아닙니다.
- **위조 가능성**: 모든 `KONSOLE_*` 변수는 일반 환경변수이므로 사용자나 스크립트가 임의로 `export KONSOLE_VERSION=...`, `export KONSOLE_DBUS_SESSION=/Sessions/1` 등을 설정해 Konsole처럼 보이게 만들 수 있습니다. 값 형식이 그럴듯하다고 해서 실제 Konsole 세션이라는 보장은 없습니다.

이 구조적 한계 때문에 정직하게 주장할 수 있는 것은 "**이 값들이 존재하면 최소한 과거 어느 시점에 Konsole(또는 konsolepart 임베디드 인스턴스)이 이 프로세스 트리를 시작했을 가능성이 높다**"는 정도이며, "현재 이 순간 Konsole이 이 프로세스를 소유하고 있다"는 결론은 낼 수 없습니다. 후자를 확인하려면 `KONSOLE_DBUS_SERVICE`/`KONSOLE_DBUS_SESSION`이 가리키는 D-Bus 목적지에 실제로 메서드 호출을 시도해 응답을 받는 **런타임 생존 확인**이 필요합니다.

### D-Bus 변수는 살아있는 IPC 핸들이다

`KONSOLE_DBUS_SESSION`과 `KONSOLE_DBUS_WINDOW`는 단순한 식별 문자열이 아니라 실제 D-Bus 오브젝트 경로이고, `KONSOLE_DBUS_SERVICE`는 그 오브젝트가 있는 버스 이름입니다. 공식 핸드북은 이를 이용해 `qdbus org.kde.konsole $KONSOLE_DBUS_SESSION`처럼 **현재 세션을 제어하는 메서드를 실시간으로 호출**하는 스크립팅 예제를 공식적으로 제공합니다. Konsole 자신도 이 패턴을 실제로 사용합니다 — Wayland에서 `XDG_ACTIVATION_TOKEN`이 없을 때, Konsole은 `KONSOLE_DBUS_SERVICE`/`KONSOLE_DBUS_SESSION`/`KONSOLE_DBUS_ACTIVATION_COOKIE`를 읽어 그 세션에 `activationToken` 메서드를 D-Bus로 호출하고 응답을 기다립니다([`src/main.cpp`](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/main.cpp#L198-L216)).

이는 다른 대부분의 터미널 마커(예: `TERM_PROGRAM`, `ZED_TERM`)와 다른 흥미로운 지점입니다 — 값이 스냅숏이 아니라 **지금 응답 가능한 살아있는 핸들**일 수 있으므로, 원칙적으로는 소비자가 그 경로에 D-Bus 호출을 시도해 "이 Konsole 세션이 지금도 살아있는가"를 스냅숏 신뢰가 아니라 실측으로 확인할 수 있습니다. 다만 이 용도는 **공식적으로 "터미널 생존 확인 API"로 문서화되어 있지 않습니다.** 핸드북이 보여주는 것은 사용자가 `qdbus`로 세션을 제어하는 스크립팅 예제이고, 소스 코드가 보여주는 것은 Konsole 자신의 내부 activation-token 왕복뿐입니다. 외부 프로그램이 임의로 이 경로에 메서드를 호출해도 되는지, 호출 실패를 "세션 종료"로 해석해도 되는지는 공식적으로 보장된 계약이 아니라 D-Bus 인터페이스의 일반적 동작(존재하지 않는 서비스·오브젝트 호출 시 오류 응답)에 의존한 추론입니다.

## konsolepart — 임베디드 터미널과의 구분 불가

Konsole의 핵심 로직은 `konsoleprivate`라는 공유 라이브러리로 빌드되며, `Session.cpp`·`SessionManager.cpp`·`ViewManager.cpp`가 모두 여기에 포함됩니다([`src/CMakeLists.txt`](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/CMakeLists.txt)). Konsole은 이 라이브러리를 **KPart 플러그인 `konsolepart`로도 빌드**하며, `konsolepart` 타깃 역시 `konsoleprivate`에 직접 링크됩니다([`add_library(konsolepart ...)`](https://github.com/KDE/konsole/blob/0768ffeec48f5b88c640a0385b637755d527e3eb/src/CMakeLists.txt#L384-L396)).

`konsolepart`는 Dolphin(F4 터미널 패널), Kate/KDevelop(내장 터미널 도구창), Konqueror, Krusader 등 다른 KDE 애플리케이션이 자기 창 안에 터미널을 넣을 때 사용하는 표준 방식입니다. 같은 `Session`·`SessionManager`·`ViewManager` 코드 경로를 그대로 쓰기 때문에, **konsolepart로 임베딩된 터미널도 독립 실행형 Konsole과 완전히 동일한 `KONSOLE_VERSION`, `KONSOLE_DBUS_SESSION`, `KONSOLE_DBUS_SERVICE`, `KONSOLE_DBUS_WINDOW` 변수를 주입합니다.** 따라서 이 문서의 변수만으로는 "독립 실행형 Konsole 창"과 "Dolphin·Kate 안에 내장된 터미널 패널"을 구분할 수 없습니다. `runby`가 답해야 할 질문이 "누가(어떤 애플리케이션이) 이 프로세스를 실행했는가"라면, `KONSOLE_*` 변수의 판정 결과는 "Konsole 코드베이스가 만든 터미널"까지만 좁혀주고, 그 호스트가 `konsole` 바이너리 자신인지 `dolphin`/`kate`인지는 별도 신호(예: 프로세스 트리의 부모 실행 파일명)로 보강해야 합니다.

## 실행 주체 감지에 관한 결론

`KONSOLE_VERSION`, `KONSOLE_DBUS_SESSION`, `KONSOLE_DBUS_SERVICE`는 Konsole(또는 konsolepart 기반 임베디드 터미널) 코드가 세션 생성 시 직접 계산·주입하는 값으로, 이들이 함께 존재하면 Konsole 계열 터미널이 프로세스를 시작했다는 신뢰도 높은 신호로 볼 수 있습니다. `KONSOLE_DBUS_WINDOW`와 `KONSOLE_DBUS_ACTIVATION_COOKIE`는 보완적 컨텍스트를 더합니다. 반대로 Konsole은 `TERM_PROGRAM`을 설정하지 않으므로, 이 관용적 마커에 의존하는 범용 감지 로직은 Konsole을 놓칩니다. `TERM=xterm-256color`와 `COLORTERM=truecolor`는 흔한 기본값이라 단독으로는 식별력이 없습니다.

가장 중요한 제약은 두 가지입니다. 첫째, 이 변수들은 존재해도 **Konsole 임베디드 인스턴스(Dolphin/Kate/Konqueror 등)와 독립 실행형 Konsole을 구분하지 못합니다.** 둘째, 모든 값이 상속되는 일반 환경변수이므로, SSH·tmux·데몬화를 거치면 **한때 Konsole이 있었다는 흔적**과 **지금 이 순간 Konsole이 살아있다는 사실**이 분리됩니다. `runby`가 정직하게 주장할 수 있는 최댓값은 "이 프로세스 트리는 어느 시점에 Konsole 계열 터미널에서 시작되었다"이며, `KONSOLE_DBUS_*`가 가리키는 D-Bus 오브젝트에 실제로 질의해 응답을 받기 전까지는 그 Konsole 세션이 지금도 존재한다고 단정해서는 안 됩니다.

## 공식 문서

- [Konsole 공식 사이트](https://konsole.kde.org/)
- [The Konsole Handbook — Scripting Konsole (D-Bus)](https://docs.kde.org/stable_kf6/en/konsole/konsole/scripting.html)
- [공식 소스 저장소 (`invent.kde.org/utilities/konsole`)](https://invent.kde.org/utilities/konsole)
- [공식 GitHub 미러 (`KDE/konsole`)](https://github.com/KDE/konsole)
- [KDE 코드 리뷰: KONSOLE_VERSION 환경변수 도입 (Phabricator D12621)](https://phabricator.kde.org/D12621)
