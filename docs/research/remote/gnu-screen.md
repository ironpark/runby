---
title: GNU Screen
slug: gnu-screen
research_date: 2026-08-31
open_source: true
repository: https://git.savannah.gnu.org/cgit/screen.git
product_type: terminal_multiplexer
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 매뉴얼(Environment, Setenv/New Window, Session Name, Invoking Screen 챕터)이 STY·WINDOW·TERM·SCREENDIR·SCREENRC·SYSSCREENRC의 의미를 명시하고, 공식 소스(savannah.gnu.org `screen.git`, master 브랜치 v5.0.0, 커밋 `9d8b0ff3901bdcb8d3bc05d94fce2ef987562768`)의 `screen.c`(`MakeNewEnv`, `main`), `window.c`(새 창 `execvpe` 경로), `process.c`(`DoCommandSetenv`/`DoCommandUnsetenv`/`DoCommandSessionname`), `attacher.c`(재접속 경로)에서 환경 스냅샷 생성·갱신·비갱신 지점을 코드 줄 단위로 직접 확인해 표를 확정할 수 있었음. 배포판이 패치하는 `TERM` 기본값 차이 등 비공식 영역만 존재하며, 이는 별도 실행 검증 없이도 "비공식"이라고 명시하는 것으로 충분함
---

# GNU Screen

아래 내용은 조사 시점의 공식 매뉴얼([GNU Screen 5.0.0, Aug 2024](https://www.gnu.org/software/screen/manual/screen.html))과 공식 소스 [`git.savannah.gnu.org/cgit/screen.git`](https://git.savannah.gnu.org/cgit/screen.git) master 브랜치(v5.0.0, 커밋 [`9d8b0ff`](https://git.savannah.gnu.org/cgit/screen.git/commit/?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768))를 기준으로 확인했습니다. GNU Screen은 [터미널 에뮬레이터 문서](../terminals/README.md)가 다루는 대상과 근본적으로 다른 종류의 소프트웨어입니다 — 터미널은 프로세스를 시작할 때 환경변수를 **주입**만 하고 이후에는 관여하지 않지만, Screen은 서버 프로세스가 살아있는 동안 **어떤 환경을 새 창에 물려줄지 자신이 계속 결정**합니다. 이 문서의 핵심은 마커 변수 자체보다, 그 결정 로직이 언제 갱신되고(혹은 갱신되지 않고) 그것이 `runby`의 나머지 축(agent·CI·terminal)에 어떤 방향의 오류를 만드는지입니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `STY` | 문자열, 형식 `<pid>.<tty>.<host>` (기본) 또는 `-S <name>` 사용 시 `<pid>.<name>` | 실행 식별 | 세션(소켓) 식별자. `screen` 실행 시 `STY`가 이미 설정돼 있으면 새 세션을 만드는 대신 그 세션에 창 하나만 추가함 | 적합 — 공식 매뉴얼이 이 변수를 세션 식별·재진입 판단에 쓴다고 명시하고, 소스에서도 서버가 시작될 때 `MakeNewEnv()`가 `STY=<SocketPath 마지막 세그먼트>`를 새 창 환경의 첫 슬롯에 항상 채워 넣는 것을 확인함. 다만 매뉴얼이 "자식 프로세스가 이 변수로 Screen 여부를 판정하라"고 계약으로 규정하지는 않으므로, 이는 Screen 자신의 내부 메커니즘을 관측한 결과이지 공식 감지 API는 아님 | [Environment — `STY`](https://www.gnu.org/software/screen/manual/html_node/Environment.html), [Session Name](https://www.gnu.org/software/screen/manual/html_node/Session-Name.html), [소스: `socknamebuf` 구성](https://git.savannah.gnu.org/cgit/screen.git/tree/src/screen.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n1012), [소스: `MakeNewEnv()`의 `STY=` 대입](https://git.savannah.gnu.org/cgit/screen.git/tree/src/screen.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n1457) |
| `WINDOW` | 정수 문자열 (창 번호, 0부터 시작) | 실행 식별 | 그 창이 만들어질 당시 부여된 창 번호 | 보조 신호 — 소스 확인 결과 새 창을 만드는 `execvpe` 경로에서 예외 없이 `WINDOW=<w_number>`가 대입됨(조건부 아님). 그러나 "WINDOW"는 X11 유틸리티나 사용자 스크립트가 재사용하기 쉬운 매우 일반적인 이름이라, 이 변수의 **존재만으로** GNU Screen을 특정하는 것은 위험함 — 다른 도구가 같은 이름을 우연히 쓰면 오탐(false positive)으로 이어질 수 있음. `STY`와 함께 있을 때만 보강 신호로 쓸 가치가 있음 | [Environment — `WINDOW`](https://www.gnu.org/software/screen/manual/html_node/Environment.html), [소스: `WINDOW=%d` 대입 (무조건)](https://git.savannah.gnu.org/cgit/screen.git/tree/src/window.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n1170) |
| `TERM` | 문자열, 컴파일 기본값 리터럴 `"screen"` | 설정 | 창 프로세스가 보게 되는 terminfo 능력 이름. 새 창을 만들 때 부모 환경의 `TERM`을 이 값으로 덮어씀 | 부적합 단독 사용 시 — 공식 소스가 보장하는 기본값은 컴파일 타임 문자열 `"screen"`뿐이며, 사용자가 `-T`나 `term` 명령으로 얼마든지 덮어쓸 수 있음. 배포판이 패치를 통해 `screen.xterm-256color` 같은 값을 기본으로 쓰는 경우가 실무에서 흔하지만, 이는 GNU 공식 소스가 아니라 배포판 벤더가 붙인 비공식 변경이므로 이 문서에서 계약으로 취급하지 않음 | [Environment — `TERM`](https://www.gnu.org/software/screen/manual/html_node/Environment.html), [소스: `screenterm` 기본값 `"screen"`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/screen.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768), [소스: 창 생성 시 `TERM=` 덮어쓰기](https://git.savannah.gnu.org/cgit/screen.git/tree/src/window.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n1155) |

## 환경 변형

이 절이 이 문서의 핵심입니다. 터미널 에뮬레이터나 CI 플랫폼 문서와 달리, Screen은 환경변수를 한 번 붙여주고 끝나는 것이 아니라 **새 창이 무엇을 물려받을지를 서버가 능동적으로 결정**합니다. 공식 소스를 줄 단위로 추적한 결과는 다음과 같습니다.

### 새 창은 서버가 "시작된 시점"의 환경을 물려받습니다

새 창의 프로세스는 클라이언트(지금 화면에 attach해 들어온 터미널)의 환경이 아니라, **screen 서버 프로세스가 최초로 기동됐을 때의 환경**을 물려받습니다. 소스에서 이는 정확히 다음과 같이 구현되어 있습니다.

- 서버는 `main()` 안에서 세션을 초기화하는 도중 [`MakeNewEnv()`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/screen.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n1046)를 **정확히 한 번** 호출합니다. 이 함수는 그 순간의 전역 `environ`을 훑어 `TERM`/`TERMCAP`/`STY`/`WINDOW`/`SCREENCAP`/`SHELL`/`LINES`/`COLUMNS`를 제외한 나머지를 복사하고, `STY`·`TERM`·`SHELL`·`TERMCAP`·`WINDOW`는 별도 슬롯을 예약해 둔 `NewEnv[]` 배열을 만듭니다.
- 이후 새 창을 만들 때마다([`window.c`의 창 생성 경로](https://git.savannah.gnu.org/cgit/screen.git/tree/src/window.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n1170))는 이 `NewEnv[]`를 그대로 `execvpe(proc, args, NewEnv)`에 넘겨 자식 셸을 실행합니다.
- 즉 `NewEnv[]`는 **서버 기동 시점의 스냅샷**이며, 그 이후 어떤 클라이언트가 attach하든 그 클라이언트 프로세스의 환경은 이 스냅샷에 전혀 반영되지 않습니다.

### `setenv`/`unsetenv`만이 공식적인 변경 수단이며, 새 창에만 적용됩니다

Screen 자체 명령어 `setenv var string`/`unsetenv var`는 매뉴얼이 "환경은 이후에 fork되는 모든 셸에 상속된다(The environment is inherited by all subsequently forked shells)"고 명시하는, 유일하게 문서화된 변경 수단입니다. 소스에서 이 동작은 다음과 같이 구현됩니다.

- [`DoCommandSetenv()`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/process.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n3327)는 libc `setenv()`를 호출한 **직후 `MakeNewEnv()`를 다시 호출**해 `NewEnv[]` 스냅샷을 새로 만듭니다.
- [`DoCommandUnsetenv()`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/process.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n3339)도 동일하게 `unsetenv()` 뒤 `MakeNewEnv()`를 다시 호출합니다.
- 세션 이름을 바꾸는 `sessionname` 명령([`DoCommandSessionname()`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/process.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768#n3293))도 `SocketPath`를 바꾼 뒤 `MakeNewEnv()`를 호출합니다 — 즉 `STY` 값조차 세션 이름을 바꿔야 갱신되며, 그 갱신도 이후 새로 만들어지는 창에만 적용됩니다.

여기서 결정적인 사실은, `MakeNewEnv()`가 다시 만드는 것은 **"다음번 창 생성에 쓸 스냅샷"**이지 이미 실행 중인 창의 셸 프로세스가 아니라는 점입니다. 이미 fork/exec된 프로세스는 OS 프로세스 모델상 자신만의 독립된 `environ` 사본을 이미 가지고 있으므로, 서버가 `NewEnv[]`를 아무리 갱신해도 그 갱신은 절대 소급 적용되지 않습니다. 매뉴얼의 "subsequently forked shells"라는 표현은 이 구현 사실을 정확히 반영합니다.

### tmux의 `update-environment`에 대응하는 기능이 Screen에는 없습니다

tmux는 `update-environment` 옵션으로 지정한 변수 목록(기본값 `DISPLAY KRB5CCNAME SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WINDOWID XAUTHORITY`)을 **attach할 때마다** attach하는 클라이언트의 값으로 갱신합니다. 이는 [터미널 문서](../terminals/README.md)에도 기록된 tmux의 공식 동작입니다.

GNU Screen 공식 소스에서 재접속(attach/reattach) 경로를 담당하는 [`attacher.c`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/attacher.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768) 전체를 확인한 결과, `MakeNewEnv()` 호출은 물론 `environ`을 다루는 코드가 **단 한 줄도 없습니다**. `MakeNewEnv()`를 호출하는 지점은 앞서 확인한 서버 기동 시점(`screen.c`)과 `setenv`/`unsetenv`/`sessionname` 명령(`process.c`) 세 곳뿐이며, 그 어디에도 "attach 시점에 클라이언트 환경을 반영"하는 코드 경로가 없습니다.

즉 **GNU Screen에는 tmux `update-environment`에 대응하는 기능이 공식적으로 존재하지 않습니다.** 이것이 이 조사의 가장 중요한 발견입니다: tmux의 낡음(staleness)은 사용자가 `update-environment` 목록에 원하는 변수를 추가해 완화할 수 있는 반면, Screen의 낡음은 **설정으로 고칠 수 있는 여지가 원천적으로 없습니다.** 유일한 수단은 사용자가 매 세션마다 수동으로 `setenv` 명령을 실행하는 것뿐이며, 그마저도 이미 떠 있는 창에는 적용되지 않고 그 이후 새로 만드는 창에만 적용됩니다.

이 한계는 실사용에서도 잘 알려진 문제로 드러납니다. `ssh -X`로 X11 포워딩을 연결한 뒤 Screen 세션을 시작하고, 나중에 다른 SSH 연결(다른 `DISPLAY` 포트)로 재접속하면 이미 떠 있던 창의 `DISPLAY`는 예전 연결을 계속 가리키고 있어 그래픽 프로그램이 실패합니다 — 사용자가 `setenv DISPLAY ...`로 수동 갱신해야 합니다. 이는 위에서 코드로 확인한 구조("attach 경로에 환경 갱신 코드 없음")가 실제로 야기하는 현상입니다.

### `screen -d -r`은 기존 창의 어떤 것도 갱신하지 않습니다

다른 터미널에서 `screen -d -r`로 세션을 detach 후 reattach해도, 위에서 확인했듯 `attacher.c` 경로에는 환경 갱신 코드가 없으므로 이미 존재하는 창의 환경변수는 **아무것도 바뀌지 않습니다.** 갱신되는 것은 화면에 다시 그려지는 디스플레이·PTY 연결뿐이며, 각 창 프로세스가 들고 있는 `environ`은 그 창이 처음 만들어졌을 때(서버 기동 시점 스냅샷 기준) 그대로 남습니다.

### `SCREENDIR`는 `STY` 값 자체를 바꾸지 않습니다

`SCREENDIR`는 세션 소켓 파일을 저장하는 **디렉터리 경로**만 바꿉니다. 소스에서 `STY`에 대입되는 값은 소켓 경로 전체가 아니라 `socknamebuf`(`<pid>.<tty>.<host>` 또는 `<pid>.<name>`)라는 마지막 세그먼트뿐이므로, `SCREENDIR`를 바꿔도 `STY`의 **값 형식이나 내용에는 영향이 없습니다.** 다만 `screen -ls`나 `-r` 같은 명령이 세션을 찾는 위치는 바뀌므로, 여러 `SCREENDIR`를 쓰는 환경에서는 같은 `STY` 문자열을 가진 세션이 이론상 서로 다른 디렉터리에 존재할 수 있어 `STY` 값만으로는 세션을 완전히 특정하지 못할 수 있습니다.

### `SCREENRC`·`SYSSCREENRC`

`SCREENRC`는 사용자 `screenrc` 파일 경로를, `SYSSCREENRC`(빌드 시 `ALLOW_SYSSCREENRC`가 켜져 있을 때만 유효)는 시스템 전역 `screenrc` 파일 경로를 재정의합니다. 두 변수 모두 설정 파일을 어디서 읽을지에만 관여하며, 세션 식별이나 새 창의 실행 환경 구성 로직에는 관여하지 않습니다.

## 다른 축에 미치는 영향

`runby`는 에이전트·CI·터미널·실행 도구·원격 다섯 축을 보고합니다. GNU Screen 자신은 원격 축의 `RemoteScreen`(`RemoteKindMultiplexer`) 계층으로 보고되므로, 아래에서는 Screen이 **나머지 축**을 오염시키는 방식을 정리합니다 — 축마다 방식이 다릅니다.

- **terminal 축 — 가장 심각하게 오염됩니다.** `STY`/`WINDOW`가 가리키는 것은 "지금 이 프로세스에 붙어 있는 터미널"이 아니라 "이 Screen 서버를 최초로 기동한 터미널"입니다. 게다가 위에서 확인했듯 Screen에는 tmux `update-environment` 같은 새로고침 수단이 전혀 없으므로, 이 오염은 tmux보다 구조적으로 더 심하고 사용자가 설정으로 고칠 수도 없습니다. 오류 방향은 **낡았지만 그럴듯함(stale-but-plausible)**입니다 — 서버가 기동될 당시에는 사실이었던 정보가, 시간이 지나 다른 사람이 다른 터미널에서 재접속해 만든 새 창에도 계속 나타납니다.
- **CI 축 — stale-but-plausible 오류를 만듭니다.** CI 러너 안에서 시작된 Screen 세션이 잡 종료 후에도 데몬처럼 계속 살아있다면, `GITHUB_ACTIONS` 같은 CI 식별 변수가 서버 기동 시점 스냅샷에 그대로 박혀, 잡이 끝난 한참 뒤 사람이 attach해서 만든 새 창에도 나타날 수 있습니다. "그 서버가 CI 환경에서 시작됐다"는 사실 자체는 참이지만, "지금 이 명령이 CI에서 실행되고 있다"는 판정으로 오독하면 오탐입니다.
- **agent 축 — 동일한 메커니즘으로 오탐 방향의 오류를 만듭니다.** 만약 어떤 에이전트(예: Claude Code)가 실행한 셸 안에서 Screen 서버가 시작됐다면, 그 에이전트가 주입한 실행 식별 변수(`CLAUDECODE=1` 등)가 서버 기동 시점 스냅샷에 포함되어, 이후 완전히 다른 사람이 대화형으로 만든 새 창에서도 계속 보고될 수 있습니다. Screen 자체는 "누가 명령을 요청했는가"를 판단할 자체 신호를 전혀 제공하지 않으므로, agent 축에 대한 오염은 전적으로 Screen이 상위 프로세스의 에이전트 마커를 오래도록 보존한다는 부작용을 통해서만 발생합니다.

`runby`는 `STY`를 `RemoteScreen` 계층으로 보고하고, 멀티플렉서가 잡히면 터미널 판정의 `Confidence`를 `probable`로 낮춥니다. 살아 있는 조상으로 확증되지 않은 에이전트·러너 계층도 함께 낮아집니다(`runby.go`의 `applyMultiplexerStaleness`). 방향은 옳지만, 이번 조사로 확인된 사실 — **Screen은 tmux와 달리 attach 시점 갱신 수단이 아예 없다** — 을 반영하면 이 처리로 충분한지 재검토할 필요가 있습니다. tmux는 사용자가 `update-environment`를 설정하면 적어도 일부 변수는 최신 값을 반영할 수 있는 반면, Screen 세션은 서버가 오래 살아있을수록 낡음이 무조건 누적되며 사용자가 이를 완화할 공식 수단이 `setenv`를 매번 수동 실행하는 것 말고는 없습니다. 따라서 `runby`가 `STY`를 볼 때는 단순히 `probable`로 낮추는 것에 더해, 그 판정이 "창이 만들어진 시점이 아니라 Screen 서버가 최초로 기동된 시점"의 스냅샷에 근거한다는 점, 그리고 그 시점과 현재 사이의 간격을 좁힐 어떤 공식 메커니즘도 없다는 점을 문서·주석 수준에서라도 명시하는 것이 정직합니다.

## 실행 주체 감지에 관한 결론

`STY`의 존재는 "이 프로세스가 GNU Screen이 관리하는 창 안에서 실행되고 있다"는 사실을 신뢰도 높게 알려주지만, 그것이 가리키는 터미널·시점은 서버가 처음 기동된 순간에 고정되어 있으며 그 이후 어떤 재접속으로도 갱신되지 않습니다. `WINDOW`는 창 생성 시 예외 없이 설정되는 값이지만 이름 자체가 지나치게 일반적이어서 단독으로는 오탐 위험이 있으므로 `STY`와 결합해야만 의미가 있습니다. `TERM`은 공식적으로 보장되는 기본값이 `"screen"`이라는 문자열뿐이고 사용자·배포판이 쉽게 바꾸므로 판정에는 쓸 수 없습니다.

가장 중요한 결론은 환경 변형 메커니즘 자체에 있습니다. Screen은 서버 기동 시점의 환경을 새 창에 계속 물려주며, 이를 바꾸는 유일한 공식 수단(`setenv`/`unsetenv`)조차 새로 만드는 창에만 적용되고, tmux의 `update-environment`에 해당하는 attach-시점 자동 갱신 기능은 공식 소스 어디에도 없습니다. 따라서 Screen이 보고하는 모든 환경변수는 원리적으로 "그때는 참이었던, 지금은 확인되지 않는" 스냅샷이며, 이 낡음은 사용자 설정으로 해소할 수 없는 구조적 한계입니다. `runby`는 이 사실을 근거로 `STY` 기반 판정을 `probable` 이상으로 올리지 말아야 하며, 가능하다면 tmux와 Screen을 구분해 "Screen은 새로고침 수단이 전혀 없다"는 더 강한 경고를 별도로 표시하는 편이 정직합니다.

## 공식 문서

- [Screen User's Manual (v5.0.0, Aug 2024)](https://www.gnu.org/software/screen/manual/screen.html)
- [Environment — 전체 환경변수 목록](https://www.gnu.org/software/screen/manual/html_node/Environment.html)
- [Setenv (New Window 챕터) — `setenv`/`unsetenv` 명령](https://www.gnu.org/software/screen/manual/html_node/Setenv.html)
- [Session Name — `STY`/세션 이름 형식](https://www.gnu.org/software/screen/manual/html_node/Session-Name.html)
- [Invoking Screen — `-S`, `-d -r`, `-D -R` 등 명령행 옵션](https://www.gnu.org/software/screen/manual/html_node/Invoking-Screen.html)
- 공식 소스 저장소 [`git.savannah.gnu.org/cgit/screen.git`](https://git.savannah.gnu.org/cgit/screen.git) (master, v5.0.0, 커밋 [`9d8b0ff`](https://git.savannah.gnu.org/cgit/screen.git/commit/?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768))
  - [`src/screen.c` — `MakeNewEnv()`, `main()`의 서버 기동 시 1회 호출, `socknamebuf`/`STY=` 구성](https://git.savannah.gnu.org/cgit/screen.git/tree/src/screen.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768)
  - [`src/window.c` — 새 창 `execvpe` 경로의 `WINDOW=`/`TERM=` 대입](https://git.savannah.gnu.org/cgit/screen.git/tree/src/window.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768)
  - [`src/process.c` — `DoCommandSetenv`/`DoCommandUnsetenv`/`DoCommandSessionname`](https://git.savannah.gnu.org/cgit/screen.git/tree/src/process.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768)
  - [`src/attacher.c` — 재접속 경로 (환경 갱신 코드 없음)](https://git.savannah.gnu.org/cgit/screen.git/tree/src/attacher.c?id=9d8b0ff3901bdcb8d3bc05d94fce2ef987562768)
