---
title: iTerm2
slug: iterm2
research_date: 2026-08-31
open_source: true
repository: https://github.com/gnachman/iTerm2
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: LC_TERMINAL의 SSH 전달, tmux 세션 재부착 시 TERM_PROGRAM/ITERM_SESSION_ID 잔존 여부, Shell Integration 미설치 환경에서의 변수 유무를 실제 iTerm2·SSH·tmux 조합으로 확인해야 함
---

# iTerm2

iTerm2는 세션을 새로 만들 때 `TERM_PROGRAM`, `TERM_PROGRAM_VERSION`, `ITERM_SESSION_ID`, `TERM_SESSION_ID`, `ITERM_PROFILE`을 프로세스 환경에 직접 주입하며, 이는 공식 소스 `PTYSession.m`의 세션 환경 구성 코드에서 확인됩니다. 다만 이 계약은 macOS에서 iTerm2가 직접 새 셸을 실행할 때만 성립하며, 문서에 명시된 안정된 공개 API가 아니라 구현 세부사항입니다. `LC_TERMINAL`/`LC_TERMINAL_VERSION`은 별도의 "Experimental" 설정(`Set LC_TERMINAL=iTerm2`)에 의해 주입되고, Shell Integration이 추가하는 변수는 사용자가 스크립트를 설치했을 때만 존재합니다. 즉 iTerm2는 "이 변수들이 항상 있다"를 보장하지 않고, 각각 다른 조건에서 켜지는 여러 신호의 집합을 제공할 뿐입니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERM_PROGRAM` | 문자열 (`iTerm.app`) | 실행 식별 | 세션을 만든 터미널 프로그램 식별 | 적합 — iTerm2가 새 셸마다 하드코딩하는 값이며, 값 자체의 대소문자(`iTerm.app`)까지 고정되어 가장 신뢰도 높은 단일 마커임 | [공식 소스: `PTYSession.m#L3290`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3290) |
| `ITERM_SESSION_ID` | 문자열, `w<창번호>t<탭번호>p<창번호>:<GUID>` (예: `w0t1p12:C3D91F33-...`) | 실행 식별 | 창·탭·창(pane) 위치와 세션 GUID를 함께 인코딩한 세션 식별자 | 적합 — iTerm2가 세션 생성 시 직접 채번하며, `TERM_PROGRAM`과 교차 검증 시 특정 pane까지 특정 가능 | [공식 소스: `PTYSession.m#L3110-L3115`, `L3287`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3110-L3115) |
| `TERM_SESSION_ID` | `ITERM_SESSION_ID`와 동일 형식·동일 값 | 실행 식별 | macOS `Terminal.app` 계열과 호환되는 별칭 변수 | 보조 — `ITERM_SESSION_ID`와 값이 같아 정보량은 동일하지만, 변수명 자체는 다른 터미널도 관례적으로 채택할 수 있어 `TERM_PROGRAM`과의 조합을 권장 | [공식 소스: `PTYSession.m#L3289`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3289) |
| `TERM_PROGRAM_VERSION` | 버전 문자열 (`CFBundleShortVersionString`) | 상태·컨텍스트 | iTerm2 앱 버전 | 보조 — `TERM_PROGRAM=iTerm.app`과 함께일 때만 버전 컨텍스트로 유효하며 단독 식별에는 부적합 | [공식 소스: `PTYSession.m#L3288`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3288) |
| `ITERM_PROFILE` | 프로필 이름 문자열 | 상태·컨텍스트 | 세션이 사용한 iTerm2 프로필 이름 | 보조 — iTerm2 세션이라는 전제 하에서만 의미가 있고, 사용자가 프로필 이름을 임의로 짓기 때문에 그 값 자체로는 아무것도 증명하지 않음 | [공식 소스: `PTYSession.m#L3295-L3296`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3296) |
| `LC_TERMINAL` | 문자열 (`iTerm2`) | 실행 식별 | `LC_*` 로케일 네임스페이스를 이용해 SSH로 전파되는 터미널 식별자 | 보조 — 로컬에서는 `TERM_PROGRAM`만큼 신뢰도가 높지만, 이름 자체가 SSH 전달을 노린 설계이므로 원격 호스트에서 관측되면 "지금 여기가 iTerm2"라는 뜻이 아님(아래 상속과 잔존 참고) | [공식 소스: `PTYSession.m#L3264-L3266`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3264-L3266), [Advanced Settings: `shouldSetLCTerminal`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/Settings/iTermAdvancedSettingsModel.m#L882) |
| `LC_TERMINAL_VERSION` | 버전 문자열 | 상태·컨텍스트 | `LC_TERMINAL`과 짝을 이루는 버전 정보 | 보조 — `LC_TERMINAL`과 동일한 조건·동일한 SSH 전파 특성을 가지므로 로컬 식별 보조 신호로만 사용 가능 | [공식 소스: `PTYSession.m#L3266`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3266) |
| `TERM` | 문자열, 기본값 `xterm-256color` | 설정 | 터미널 기능(terminfo) 식별 | 부적합 — 프로필의 "Terminal type" 설정값이며 사용자가 임의로 바꾸고, `xterm-256color`는 다른 터미널도 널리 채택하는 값이라 단독으로는 의미가 없음 | [공식 소스: 프로필 기본값 `ITAddressBookMgr.m#L545`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/Settings/Profiles/ITAddressBookMgr.m#L545) |

가장 신뢰도 높은 단일 마커는 **`TERM_PROGRAM=iTerm.app`**입니다. iTerm2가 새 셸 프로세스를 만들 때마다 예외 없이 하드코딩하는 값이며, 대소문자까지 고정되어 있어 (`iTerm.app`, 앞의 `i`는 소문자) 비교 시 정확히 맞춰야 합니다. `ITERM_SESSION_ID`는 창(window)·탭(tab)·창(pane) 위치와 세션 GUID를 `w<N>t<N>p<N>:<GUID>` 형태로 함께 담으므로, 같은 iTerm2 인스턴스 안에서 정확히 어느 pane인지까지 구분할 수 있는 유일한 변수입니다.

## 상태·컨텍스트 변수

Shell Integration(사용자가 `.zshrc` 등에 스크립트를 직접 설치해야 활성화됨)은 위 표에 없는 추가 변수를 셸 세션에 더합니다. 대표적으로 설치 여부 자체를 나타내는 `ITERM_SHELL_INTEGRATION_INSTALLED`가 있으며, 공식 스크립트는 이 값이 비어 있을 때만 스스로를 설치합니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `ITERM_SHELL_INTEGRATION_INSTALLED` | 문자열 (`Yes`) | 상태·컨텍스트 | Shell Integration 스크립트가 이미 로드되었음을 표시하는 재진입 방지 플래그 | 부적합 — 사용자가 별도로 설치해야만 존재하므로 iTerm2 세션 전반에 보장되지 않고, 변수명 자체가 셸 초기화 스크립트가 자유롭게 지정하는 값이라 위조도 쉬움 | [공식 소스: `iterm2_shell_integration.zsh#L18-L19`](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/Resources/shell_integration/iterm2_shell_integration.zsh#L18-L19) |

`runby`가 안전하게 노출할 수 있는 상태 정보는 사실상 `TERM_PROGRAM_VERSION`뿐입니다. iTerm2가 Python API·Scripting을 통해 관리하는 `session.*`, `user.*` 같은 "Variables" 체계는 공식 문서(iTerm2 Variables)에 별도로 설명되어 있지만, 이는 iTerm2 내부 상태·API 조회용 개념이며 자식 프로세스에 자동으로 환경변수로 전달되지 않으므로 이 문서의 범위에서 제외합니다.

## 상속과 잔존

환경변수는 자식 프로세스 전체에 상속되며 프로세스가 스스로 지우지 않는 한 계속 남습니다. 이 때문에 위 표의 "적합" 판정도 "지금 이 프로세스가 iTerm2가 방금 만든 셸"이라는 뜻일 뿐, "부모 iTerm2 세션이 여전히 살아 있다"는 것을 보장하지 않습니다. 구체적으로:

- **SSH**: OpenSSH는 기본 설정에서 임의 환경변수를 원격으로 전달하지 않습니다. `TERM_PROGRAM`, `ITERM_SESSION_ID`는 클라이언트의 `SendEnv`/서버의 `AcceptEnv` 설정에 명시적으로 포함되지 않는 한 원격 호스트로 넘어가지 않습니다. 반면 `LC_TERMINAL`은 iTerm2가 **의도적으로** `LC_*` 로케일 네임스페이스에 넣은 값입니다. OpenSSH의 기본 `SendEnv`에는 관례적으로 `LANG`, `LC_*` 패턴이 포함되어 있어(배포판별 `ssh_config` 기본값에 따라 다르지만 흔한 기본값입니다), 별도 설정 없이도 `LC_TERMINAL=iTerm2`가 원격 호스트로 전달되는 경우가 많습니다. iTerm2 자체의 "Set LC_TERMINAL=iTerm2" 설정 설명도 "openssh and mosh pass this to hosts you connect to"라고 명시합니다. 결과적으로 **원격 호스트에서 실행 중인 프로세스가 `LC_TERMINAL=iTerm2`를 관측할 수 있는데, 이 값은 "지금 이 원격 셸이 iTerm2에서 열렸다"를 증명하지 않고 "로컬에서 접속해 온 클라이언트가 iTerm2였다"만 증명합니다.** 이 둘을 혼동하면, 원격 서버에서 실행되는 모든 후속 자식 프로세스가 (SSH 세션이 끝나기 전까지) 로컬 iTerm2로 오탐될 수 있습니다. `LC_TERMINAL`을 로컬 터미널 식별 신호로 쓰는 라이브러리는 이 오탐 경로를 반드시 문서화해야 합니다.
- **tmux/screen**: tmux는 서버를 최초로 띄운 클라이언트의 환경을 기준으로 새 세션·창의 환경을 구성합니다. `update-environment` 옵션의 기본 목록(`DISPLAY`, `KRB5CCNAME`, `SSH_ASKPASS`, `SSH_AUTH_SOCK`, `SSH_AGENT_PID`, `SSH_CONNECTION`, `WINDOWID`, `XAUTHORITY`)에는 `TERM_PROGRAM`도 `ITERM_SESSION_ID`도 포함되어 있지 않습니다. 즉 tmux는 클라이언트가 재접속(attach)할 때 이 변수들을 자동으로 새로고침하지 않으므로, 어떤 pane의 `TERM_PROGRAM=iTerm.app` / `ITERM_SESSION_ID=w0t1p1:...`은 그 tmux 세션을 **처음 만든** 클라이언트가 iTerm2였다는 화석일 뿐, 지금 그 pane을 보고 있는 클라이언트가 iTerm2라는 뜻이 아닙니다. tmux 세션이 iTerm2가 아닌 다른 터미널(또는 다른 머신)에서 attach된 상태로 남아 있어도 값은 그대로입니다.
- **분리·데몬화된 프로세스**: `nohup`, `disown`, `setsid`로 띄운 백그라운드 프로세스나 아예 터미널에서 분리되어 상주하는 데몬은 시작 시점에 물려받은 `TERM_PROGRAM`/`ITERM_SESSION_ID`를 그대로 유지한 채, 만들어 준 iTerm2 창·탭·세션이 이미 닫힌 뒤에도 계속 실행될 수 있습니다.
- **위조 가능성**: 이 변수들은 표준 환경변수이므로 어떤 프로세스든 `export TERM_PROGRAM=iTerm.app`으로 값을 자유롭게 지정할 수 있습니다. 서명이나 커널 수준의 보증이 없어 신뢰 경계로 쓸 수 없습니다.

결론적으로 이 변수들은 **"이 프로세스 트리가 한때 iTerm2가 시작한 세션에서 나왔다"**는 이력 신호이지, **"지금 이 프로세스를 보고 있는 터미널이 iTerm2다"**를 보장하는 신호가 아닙니다. 라이브러리는 이 구분을 명시하고, 특히 `LC_TERMINAL`이 SSH로 넘어온 값일 가능성을 항상 열어 두어야 합니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM=iTerm.app`은 로컬에서 관측되는 한 가장 신뢰도 높은 단일 마커이며, `ITERM_SESSION_ID`(형식 `w<창>t<탭>p<창(pane)>:<GUID>`)와 결합하면 어느 window/tab/pane에서 시작됐는지까지 구분할 수 있습니다. 하지만 `runby`가 이 신호로 답할 수 있는 것은 "이 프로세스가 (한때) iTerm2 세션에서 상속된 환경을 가지고 있다"까지이며, "지금 살아 있는 iTerm2 창이 이 프로세스를 소유한다"는 보장은 아닙니다. SSH로 넘어간 `LC_TERMINAL`, tmux에 잔존하는 `TERM_PROGRAM`/`ITERM_SESSION_ID`, 분리된 데몬 프로세스가 모두 같은 이유로 오탐을 만들 수 있습니다. iTerm2는 `agent_host`가 아니라 순수 터미널 에뮬레이터이므로 `executes_agents`는 비워 두며, 이 문서의 신호는 `Result.IsAgent()` 판정에 포함하지 않고 터미널 호스트 식별(`Kind`가 아닌 별도 표시)에만 사용하는 것이 안전합니다.

## 공식 문서

- [iTerm2 Documentation — Shell Integration](https://iterm2.com/documentation-shell-integration.html)
- [iTerm2 Documentation — Variables](https://iterm2.com/documentation-variables.html)
- [iTerm2 공식 GitHub 저장소](https://github.com/gnachman/iTerm2)
- [공식 소스: `PTYSession.m` — 세션 환경변수 구성](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/PTYSession/PTYSession.m#L3260-L3297)
- [공식 소스: `iTermAdvancedSettingsModel.m` — `shouldSetLCTerminal` 기본값과 설명](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/sources/Settings/iTermAdvancedSettingsModel.m#L882)
- [공식 소스: `iterm2_shell_integration.zsh` — 설치 감지 플래그](https://github.com/gnachman/iTerm2/blob/5ff63dade30865fe9faf2ac7003971dd55c46c88/Resources/shell_integration/iterm2_shell_integration.zsh#L18-L19)
