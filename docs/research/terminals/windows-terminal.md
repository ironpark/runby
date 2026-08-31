---
title: Windows Terminal
slug: windows-terminal
research_date: 2026-08-31
open_source: true
repository: https://github.com/microsoft/terminal
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: WSL 안에서 실행되는 Linux 바이너리가 실제로 WT_SESSION·WT_PROFILE_ID를 관측하는지, 그리고 conhost.exe(레거시 콘솔)와 Windows Terminal에서 값이 실제로 달라지는지는 Windows 환경에서 직접 실행해 확인해야 함
---

# Windows Terminal

Windows Terminal의 공식 문서는 `WT_SESSION`·`WT_PROFILE_ID` 환경변수를 안정적 공개 계약으로 문서화하지 않습니다. 두 변수는 공식 오픈소스 저장소 `microsoft/terminal`의 `ConptyConnection::_LaunchAttachedClient()` 구현에서 확인되며, Windows Terminal이 ConPTY로 새 클라이언트 프로세스를 시작할 때마다 그 프로세스의 환경 블록에 직접 주입됩니다. 즉 Windows Terminal이 보장하는 것은 "문서화된 API"가 아니라 "현재 구현이 매 연결마다 만들어 주입하는 값"이며, 이 문서는 공식 소스에서 확인한 이 런타임 동작을 근거로 감지 신호를 정리합니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `WT_SESSION` | 중괄호 없는 GUID 문자열 (예: `5720ee6d-6474-47b0-88db-fa7e10e60d37`) | 실행 식별 | Windows Terminal의 연결(탭/창) 세션을 식별하는 고유 GUID를 클라이언트 프로세스 환경에 주입 | 적합 — Windows Terminal이 매 ConPTY 연결마다 런타임에 덮어써 주입하는 값이며, 존재 자체가 가장 신뢰도 높은 단일 마커 | [공식 소스: `ConptyConnection::_LaunchAttachedClient`](https://github.com/microsoft/terminal/blob/3b8c5606ee1866e7bf5c4e84e962798492282097/src/cascadia/TerminalConnection/ConptyConnection.cpp#L61) |
| `WT_PROFILE_ID` | 중괄호 포함 GUID 문자열 (예: `{2ece5bfe-50ed-5f3a-ab87-5cd4baafed2b}`) | 상태·컨텍스트 | 현재 탭/창이 사용하는 `settings.json` 프로필의 GUID | 보조 신호 — `WT_SESSION`과 함께 있으면 Windows Terminal 확인을 보강하지만, 프로필 GUID 하나만으로는 세션 존재 자체를 증명하지 못함 | [공식 소스: `ConptyConnection::_LaunchAttachedClient`](https://github.com/microsoft/terminal/blob/3b8c5606ee1866e7bf5c4e84e962798492282097/src/cascadia/TerminalConnection/ConptyConnection.cpp#L64) |

`WT_SESSION` 자체가 마커와 세션 식별자를 겸합니다. `CIRCLECI=true`나 `ZED_TERM=true`처럼 "존재/불리언 마커"와 "식별자"가 분리된 다른 도구들과 달리, Windows Terminal은 별도의 불리언 플래그를 두지 않고 GUID 값이 있다는 사실 자체를 신호로 씁니다. 따라서 감지 로직은 값의 내용이 아니라 **변수의 존재 여부**만 확인하면 되고, 값은 세션을 구분하는 용도로 별도 사용할 수 있습니다.

**`TERM_PROGRAM`은 Windows Terminal이 설정하지 않습니다.** 이 변수를 Windows Terminal이 새 프로세스에 주입하도록 해 달라는 기능 요청([microsoft/terminal#840](https://github.com/microsoft/terminal/issues/840))이 있었지만, 실제로 병합된 구현은 `TERM_PROGRAM`이 아니라 지금의 `WT_SESSION`/`WT_PROFILE_ID` 주입이었습니다. `ConptyConnection.cpp`의 환경변수 주입 코드에도 `TERM_PROGRAM`이나 `TERM`을 설정하는 부분은 없습니다. 따라서 macOS 터미널들(`TERM_PROGRAM=Apple_Terminal` 등)과 달리 Windows Terminal에서는 이 변수로 감지할 수 없습니다. 사용자가 프로필의 `environment` 설정으로 `TERM_PROGRAM`을 수동 지정하면 나타날 수 있지만, 이는 Windows Terminal이 아니라 사용자 설정에 의한 값입니다.

## 상태·컨텍스트 변수

Windows Terminal이 클라이언트 프로세스에 항상 주입하는 변수는 `WT_SESSION`, `WT_PROFILE_ID`, 그리고 이 둘을 WSL로 전달하기 위한 `WSLENV` 조정뿐입니다. 창 ID(`WT_WINDOWID`)처럼 커뮤니티에서 요청된 값들은 공식 소스에서 안정적으로 확인되지 않아 이 문서에서는 다루지 않습니다. `settings.json`의 프로필별 `"environment"` 항목으로 사용자가 임의의 변수를 추가할 수 있지만, 이는 Windows Terminal이 자동으로 제공하는 신호가 아니라 사용자가 직접 설정한 값이므로 감지 근거로 쓸 수 없습니다.

## 상속과 잔존

`WT_SESSION`/`WT_PROFILE_ID`는 다른 환경변수와 마찬가지로 프로세스가 fork/exec할 때마다 자식에게 그대로 상속되는 일반 상속 규칙을 따릅니다. Windows Terminal은 프로세스를 새로 만들 때 이 값을 **덮어써서 주입**하지만, 이미 만들어진 자식 프로세스 트리 안에서는 그 값이 그대로 복제되어 흘러갈 뿐 다시 검증되지 않습니다. 즉 "이 변수가 존재한다"는 사실은 "이 프로세스가 Windows Terminal의 직계 클라이언트다"를 보장하지 않고 "이 프로세스 트리 어딘가의 조상이 한 번은 Windows Terminal이 시작한 프로세스였다"만 보장합니다.

- **WSL 경계 넘기.** 이것이 이 라이브러리의 실제 타깃(WSL에서 빌드된 Linux용 Go 바이너리)에 가장 중요한 지점입니다. 공식 소스에서 확인한 바에 따르면, Windows Terminal은 `WT_SESSION`과 `WT_PROFILE_ID`를 클라이언트 환경에 넣는 동시에 `WSLENV`에도 플래그 없이 추가합니다([WSLENV.3 블록](https://github.com/microsoft/terminal/blob/3b8c5606ee1866e7bf5c4e84e962798492282097/src/cascadia/TerminalConnection/ConptyConnection.cpp#L87-L91)). `WSLENV`는 Windows→WSL 방향으로 어떤 변수를 Linux 프로세스 환경에 복사할지 지정하는 콜론 구분 목록이며([Microsoft Learn: WSL interop — 환경변수 공유](https://learn.microsoft.com/en-us/windows/dev-environment/wsl-interop)), `/p`(경로 변환), `/l`(경로 목록), `/u`(WSL→Windows 단방향), `/w`(Windows→WSL 단방향) 같은 플래그를 붙일 수 있습니다. `WT_SESSION`/`WT_PROFILE_ID`는 GUID 문자열이라 경로 변환이 필요 없으므로 플래그 없이 등록되며, 이는 값이 변환 없이 그대로 Linux 프로세스 환경에 복사된다는 뜻입니다.

  결과적으로, 사용자가 Windows Terminal의 "WSL" 프로필로 `wsl.exe`를 직접 실행했거나, Windows Terminal이 시작한 셸 안에서 사용자가 `wsl`을 실행한 경우 모두, 그 WSL 세션 안에서 실행되는 Linux 프로세스(이 라이브러리가 타깃으로 하는 Linux용 Go 바이너리 포함)는 **`WT_SESSION`과 `WT_PROFILE_ID`를 자신의 환경변수로 그대로 관측할 수 있습니다.** 다만 이는 그 `wsl.exe` 호출이 Windows Terminal이 만든 프로세스 트리 안에서 실행되었을 때만 성립하며, WSL 자체를 다른 방식(예: 작업 스케줄러, 서비스, 원격 SSH로 진입한 세션)으로 시작했다면 `WSLENV`에 값이 실려 오지 않으므로 두 변수는 나타나지 않습니다. 이 동작은 `microsoft/terminal` 저장소의 과거 이슈([#3948](https://github.com/microsoft/terminal/issues/3948), [#7130](https://github.com/microsoft/terminal/issues/7130))에서도 다뤄졌고 현재 `main`의 구현에 반영되어 있습니다.

- **conhost.exe(레거시 콘솔)와의 구분.** 위 주입 코드는 Windows Terminal 고유의 `TerminalConnection::ConptyConnection` 클래스에 있습니다. 사용자가 Windows Terminal을 거치지 않고 레거시 `conhost.exe` 콘솔 창(또는 콘솔을 새로 만들지 않는 `CREATE_NO_WINDOW` 방식)에서 프로세스를 실행하면 이 주입 경로를 타지 않으므로 `WT_SESSION`/`WT_PROFILE_ID`가 생기지 않습니다. 단, Windows 11의 "기본 터미널 앱"을 Windows Terminal로 지정한 경우에도 일부 실행 경로(예: 특정 방식의 프로세스 생성)에서는 이 변수가 주입되지 않는다는 보고가 있어([microsoft/terminal#13006](https://github.com/microsoft/terminal/issues/13006)), 변수가 없다고 해서 반드시 Windows Terminal이 아니라고 단정할 수는 없습니다. 즉 존재는 강한 긍정 신호이지만, 부재는 약한 부정 신호입니다.

- **SSH.** OpenSSH는 기본적으로 클라이언트 환경변수를 서버로 전달하지 않습니다. `ssh_config`의 `SendEnv` 기본값은 비어 있고, 서버(`sshd_config`)도 `AcceptEnv`로 명시 허용한 변수만 받습니다. `WT_SESSION`/`WT_PROFILE_ID`는 OpenSSH의 기본 `SendEnv`/`AcceptEnv` 목록(`LANG`, `LC_*` 등)에 포함되지 않으므로, 기본 설정의 SSH 세션에서는 원격 프로세스가 로컬 Windows Terminal 세션의 이 값들을 보지 못합니다. 사용자가 양쪽 설정 파일을 명시적으로 편집해야만 전달됩니다.

- **tmux / screen (WSL 안에서 특히 중요).** tmux 서버는 그것을 처음 띄운 클라이언트의 환경을 통째로 물려받아 세션 환경으로 고정합니다. 이후 다른 창에서 같은 서버에 `attach`해도 세션 환경은 자동으로 갱신되지 않으며, tmux는 `update-environment` 옵션에 나열된 변수(기본값은 `DISPLAY`, `SSH_AUTH_SOCK`, `SSH_AGENT_PID`, `SSH_CONNECTION`, `WINDOWID`, `XAUTHORITY`, `KRB5CCNAME` 등)만 attach 시점에 갱신합니다. `WT_SESSION`/`WT_PROFILE_ID`는 이 기본 목록에 없으므로, WSL 안에서 tmux 세션을 만든 뒤 그 창을 닫고 다른 Windows Terminal 창에서 같은 tmux 세션에 다시 attach해도 pane 안의 `WT_SESSION`은 **처음 세션을 만들었던(지금은 존재하지 않을 수도 있는) 창의 GUID를 그대로 가리키는 잔존 값**으로 남습니다. 사용자가 `set-option -g update-environment "... WT_SESSION WT_PROFILE_ID"`를 직접 추가하지 않는 한 이 값은 현재 보고 있는 창과 무관해질 수 있습니다. screen도 비슷한 구조적 문제를 가지지만 tmux처럼 표준화된 `update-environment` 메커니즘은 없습니다.

- **디태치·데몬화·서비스 프로세스.** `nohup`, `setsid`, `&`, systemd 사용자 서비스처럼 터미널에서 시작되었지만 터미널 세션 종료 후에도 살아남는 프로세스는 시작 시점의 `WT_SESSION`/`WT_PROFILE_ID`를 프로세스 종료 시까지 그대로 들고 있습니다. 그 사이 원래 창이 닫히거나 탭이 재사용되어 같은 GUID가 다른 세션에 재할당될 여지는 없지만(GUID는 매 연결마다 새로 생성됨), 값이 가리키는 창 자체는 이미 사라졌을 수 있어 "현재 어떤 Windows Terminal 창이 이 프로세스를 실행 중인가"라는 질문에는 답하지 못합니다.

- **위조 가능성.** `WT_SESSION`과 `WT_PROFILE_ID`는 형식이 단순한 GUID 문자열이라 어떤 프로세스든 `export WT_SESSION=...`(또는 Windows에서 `set`)으로 값을 자유롭게 흉내 낼 수 있습니다. 서명이나 커널 수준 증명이 없으므로 이 값을 신뢰 경계나 보안 판단에 사용해서는 안 됩니다.

이런 이유로 `runby`가 정직하게 주장할 수 있는 것은 "현재 프로세스 트리의 어느 조상이 Windows Terminal ConPTY 연결로 시작되었을 가능성이 높다"는 **확률적 신호**이지, "지금 이 순간 이 프로세스를 보여주는 창이 Windows Terminal이다"라는 **확정적 사실**이 아닙니다. SSH로 넘어가거나, tmux/screen을 오래 유지하거나, 프로세스를 디태치하면 이 신호는 쉽게 stale해지거나 사라집니다.

## 실행 주체 감지에 관한 결론

`WT_SESSION`의 존재는 Windows Terminal을 식별하는 가장 신뢰도 높은 단일 신호이며, 별도의 `<PRODUCT>=true` 형태 마커 없이 GUID 값 자체가 마커와 세션 식별자를 겸합니다. `WT_PROFILE_ID`는 이를 보강하는 보조 신호로만 사용해야 하며, `TERM_PROGRAM`은 Windows Terminal이 설정하지 않으므로 감지에 쓸 수 없습니다. WSL 안에서 실행되는 Linux 프로세스(이 라이브러리의 주요 타깃)는 Windows Terminal이 `WSLENV`에 두 변수를 자동으로 추가해 주므로, `wsl.exe`가 Windows Terminal이 만든 프로세스 트리 안에서 실행된 경우에 한해 이 값들을 그대로 관측할 수 있습니다. 그러나 이 신호는 상속되는 환경변수의 구조적 한계를 그대로 가지므로, SSH/tmux/screen/디태치 경계를 넘으면 신뢰도가 급격히 떨어지거나 stale해질 수 있고 누구나 위조할 수 있습니다. `runby`는 이 신호를 "확률적 보조 신호"로만 보고해야 하며, 보안이나 권한 판단의 근거로 쓰면 안 됩니다.

## 공식 문서

- [공식 소스: `ConptyConnection::_LaunchAttachedClient` (WT_SESSION/WT_PROFILE_ID 주입, WSLENV 조정)](https://github.com/microsoft/terminal/blob/3b8c5606ee1866e7bf5c4e84e962798492282097/src/cascadia/TerminalConnection/ConptyConnection.cpp#L58-L149)
- [microsoft/terminal 공식 저장소](https://github.com/microsoft/terminal)
- [microsoft/terminal#840 — TERM_PROGRAM 요청, 실제로는 WT_SESSION 도입으로 귀결](https://github.com/microsoft/terminal/issues/840)
- [microsoft/terminal#3948 — WT_SESSION이 WSL에 나타나지 않던 문제](https://github.com/microsoft/terminal/issues/3948)
- [microsoft/terminal#7130 — WSLENV 중복 항목 이슈](https://github.com/microsoft/terminal/issues/7130)
- [microsoft/terminal#13006 — 기본 터미널 앱 경로에서 env var가 설정되지 않는 사례 보고](https://github.com/microsoft/terminal/issues/13006)
- [Microsoft Learn: WSL interop — Windows and Linux integration (WSLENV, 환경변수 공유)](https://learn.microsoft.com/en-us/windows/dev-environment/wsl-interop)
- [OpenSSH `ssh_config(5)` — `SendEnv` 기본값](https://man.openbsd.org/ssh_config)
- [OpenSSH `sshd_config(5)` — `AcceptEnv`](https://man.openbsd.org/sshd_config)
- [tmux(1) — `update-environment`](https://man7.org/linux/man-pages/man1/tmux.1.html)
