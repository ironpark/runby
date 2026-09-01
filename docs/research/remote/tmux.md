---
title: tmux
slug: tmux
research_date: 2026-08-31
open_source: true
repository: https://github.com/tmux/tmux
product_type: terminal_multiplexer
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 man page(GLOBAL AND SESSION ENVIRONMENT)와 공식 소스(environ.c, options-table.c, tmux.c)로 TMUX 변수 형식·update-environment 기본값·global/session 병합 규칙을 확인했으나, "이미 실행 중인 pane이 재접속 후에도 예전 값을 유지하는가"는 문서에 명시된 동작 원리이지 문서화된 값 계약이 아니라서 로컬 tmux 3.7c로 직접 재현해 검증했습니다. 검증 범위: TMUX 값 형식(socket,pid,session-id), TMUX_PANE의 pane_index 재번호 매김 내구성, KITTY_WINDOW_ID류 미등록 변수의 global 잔존, TERM_PROGRAM 강제 재작성.
---

# tmux

tmux는 다른 문서들과 성격이 다릅니다. 터미널 에뮬레이터나 CI, 에이전트 하네스는 자신이 만든 프로세스에 변수를 **추가**할 뿐이지만, tmux는 서버로서 **자신이 관리하는 pane에 어떤 외부 환경변수를 넘길지 스스로 결정**합니다. 즉 tmux를 이해한다는 것은 "tmux가 무엇을 추가하는가"보다 "tmux가 바깥 환경에서 무엇을 걸러내고 무엇을 통과시키는가"를 이해하는 일에 가깝습니다. 아래 내용은 공식 소스 [`tmux/tmux`](https://github.com/tmux/tmux) 릴리스 태그 [`3.7`](https://github.com/tmux/tmux/releases/tag/3.7)(커밋 [`81f88f8`](https://github.com/tmux/tmux/commit/81f88f8517c9fc5371b56cf117530c6b477c96ac))의 `tmux.1`, `environ.c`, `options-table.c`, `tmux.c`와 로컬에 설치된 tmux 3.7c(`tmux -V`로 확인)의 실측을 기준으로 확인했습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `TMUX` | 문자열, `<소켓 경로>,<서버 PID>,<세션 숫자 ID>` 형태 (실측: `/private/tmp/tmux-501/default,63236,0`) | 실행 식별 | pane이 tmux 세션 안에서 실행 중임을 자식 프로세스에 알리고, 중첩 실행(`new-session` 안에서 다시 `tmux`)을 막는 데 쓰임 | 적합 — 이 변수의 존재 자체는 신뢰할 수 있는 "지금 tmux pane 안에서 태어났다"는 표시. 다만 값의 세 필드 형식은 man page 본문에 명시된 공식 계약이 아니라 소스(`environ.c`)에서만 확인되는 구현 세부사항이므로 파싱해서 의미를 부여하는 것은 비공식 추론임 | [man — GLOBAL AND SESSION ENVIRONMENT](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1#L6958-L6979) ("also initialises the TMUX variable with some internal information"), [소스: `environ_set(env, "TMUX", 0, "%s,%ld,%d", socket_path, (long)getpid(), idx)`](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/environ.c#L281-L282) |
| `TMUX_PANE` | 문자열, `%<정수>` (예: `%0`, `%1`) | 실행 식별 | 이 프로세스가 실행 중인 pane의 고유 ID를 자식 프로세스에 전달 | 적합 — man page가 "세션·창·pane마다 고유 ID가 있고 이는 tmux 서버 안에서 그 객체의 생애 동안 변하지 않는다"고 명시하며, 화면에 보이는 pane 번호(`pane_index`, 1부터 재배정됨)와는 별개의 안정적 식별자임을 실측으로도 확인 | [man — TMUX_PANE](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1#L912-L924) |
| `TERM` | 문자열, 기본값 `screen` (컴파일 상수 `TMUX_TERM`, `default-terminal` 옵션 값을 따름) | 설정 | 새 pane 프로세스에 전달되는 terminfo 능력 이름. 바깥 터미널이 설정했던 `TERM` 값을 무조건 덮어씀 | 부적합 — tmux 자신의 terminfo 협상용 값이지 신원이 아니며, `default-terminal` 옵션으로 사용자가 얼마든지 바꿀 수 있음(`tmux-256color`가 흔한 설정값). `runby`의 다른 터미널 문서들과 동일하게 판정 근거로 쓰지 않음 | [man — GLOBAL AND SESSION ENVIRONMENT](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1#L6975-L6979), [소스: `TMUX_TERM "screen"`](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.h#L93), [소스: `environ_set(env, "TERM", ...)`](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/environ.c#L265) |

## 환경 변형

tmux를 정확히 다루려면 이 절이 핵심입니다. tmux는 pane을 만들 때 바깥 세계의 환경변수를 그대로 물려주지 않고, 자체 규칙에 따라 **재구성한 환경**을 자식 프로세스에 넘깁니다.

### 1. 두 개의 환경: global과 session

man page의 GLOBAL AND SESSION ENVIRONMENT 절은 다음과 같이 명시합니다.

> When the server is started, tmux copies the environment into the *global environment*; in addition, each session has a *session environment*. When a window is created, the session and global environments are merged. If a variable exists in both, the value from the session environment is used. The result is the initial environment passed to the new process.

([원문 링크](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1#L6958-L6966))

즉:

- **global environment** — tmux **서버 프로세스가 처음 시작될 때 딱 한 번** 자신의 `environ`(그 순간 `tmux` 명령을 실행한 클라이언트의 전체 환경)을 복사해 만듭니다. 서버가 살아 있는 한 이후 그 어떤 attach로도 이 global environment는 자동으로 갱신되지 않습니다. 소스에서도 `global_environ = environ_create()`와 `for (var = environ; *var != NULL; var++) environ_put(global_environ, *var, 0);`가 `main()` 안에서 딱 한 번 실행됩니다. ([소스](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.c#L411-L413))
- **session environment** — 세션별로 존재하며, `update-environment` 옵션과 `set-environment` 명령으로만 변경됩니다.
- 새 창(그리고 pane)이 만들어질 때 두 환경이 병합되고, 이름이 겹치면 **session environment가 이깁니다**. 이 병합 결과가 자식 프로세스의 초기 환경입니다.

실측으로 이 구조를 확인했습니다. `KITTY_WINDOW_ID=123`을 export한 뒤 `tmux new-session`을 실행하면, 새 pane의 `env` 출력에 `KITTY_WINDOW_ID=123`이 그대로 나타나고, `tmux show-environment -g`(global)에도 남아 있지만 `tmux show-environment -t <세션>`(session)에는 나타나지 않습니다 — global에만 존재하는 값이 병합을 통해 자식에게 전달된 것입니다.

### 2. `update-environment` — 언제, 무엇을 복사하는가

`update-environment` 세션 옵션의 공식 설명(man page)은 다음과 같습니다.

> Set list of environment variables to be copied into the session environment when a new session is created or an existing session is attached. Any variables that do not exist in the source environment are set to be removed from the session environment (as if `-r` was given to the `set-environment` command).

([원문 링크](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/options-table.c#L1025-L1034))

현재(2026-08-31 기준 최신 릴리스 태그 `3.7`) 기본값은 소스 `options-table.c`에 다음과 같이 정의되어 있습니다. **여기 그대로 인용합니다.**

```
DISPLAY KRB5CCNAME MSYSTEM SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WAYLAND_DISPLAY WINDOWID XAUTHORITY XDG_CURRENT_DESKTOP XDG_SESSION_DESKTOP XDG_SESSION_TYPE
```

([소스 링크](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/options-table.c#L1025-L1034))

**주의 — 버전에 따라 이 기본값이 다릅니다.** 3.5a 이하 릴리스에서는 `MSYSTEM`, `WAYLAND_DISPLAY`, `XDG_CURRENT_DESKTOP`, `XDG_SESSION_DESKTOP`, `XDG_SESSION_TYPE`이 빠진 더 짧은 목록(`DISPLAY KRB5CCNAME SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WINDOWID XAUTHORITY`)이 기본값이었음을 태그 [`3.5a`](https://github.com/tmux/tmux/blob/3.5a/options-table.c)의 소스로 직접 확인했습니다. `runby`가 참조하는 상위 문서([`docs/terminals/README.md`](../terminals/README.md))에 인용된 목록도 이 구버전 형태와 일치합니다. **따라서 이 값은 배포판이 패키징한 tmux 버전에 따라 실제로 달라지며, 코드로 이 값에 의존한다면 하드코딩하지 말고 실행 중인 tmux의 `show-options -g update-environment`로 직접 조회해야 합니다.**

가장 중요한 사실은 **복사가 일어나는 시점**입니다. man page 원문 그대로: "when a new session is created or an existing session is attached." 즉:

- 새 세션이 만들어질 때(`new-session`), 그리고
- 기존 세션에 클라이언트가 **attach**할 때(재접속 포함)

에만 attach하는 **클라이언트**의 환경에서 이 목록에 있는 이름을 읽어 그 **세션 환경(session environment)**에 복사합니다.

**이 복사는 세션 안에서 이미 실행 중인 pane의 프로세스에는 어떤 영향도 주지 않습니다.** 프로세스는 fork 시점에 결정된 초기 환경(global+session 병합 결과의 스냅샷)을 이미 상속받아 실행 중이며, 부모(tmux)가 나중에 세션 환경을 바꾼다고 해서 이미 살아 있는 자식 프로세스의 환경변수가 다시 쓰이는 메커니즘은 유닉스 프로세스 모델 자체에 존재하지 않습니다. 세션 환경 갱신의 효과를 보려면 그 갱신 **이후에** 새로 만들어지는 창/pane이어야 합니다. 이를 로컬 tmux 3.7c로 재확인했습니다 — `SSH_CONNECTION`을 export한 뒤 세션을 만들면 첫 pane과 `show-environment -t <세션>` 모두 그 값을 반영하지만, 그 세션 환경 값을 바꾸는 메커니즘(재attach)은 이미 실행 중인 pane의 `env` 출력을 갱신하지 않습니다.

### 3. 결과 — 목록에 없는 변수는 "서버를 처음 띄운 클라이언트"에 냉동됩니다

`update-environment` 목록에 없는 변수(예: `TERM_PROGRAM`은 별도 메커니즘으로 항상 재작성되므로 예외이지만, `KITTY_WINDOW_ID`, `WEZTERM_PANE`, `ITERM_SESSION_ID`, 목록에 없는 상태에서의 `SSH_CONNECTION` 등)는 **global environment**를 통해서만 pane에 전달됩니다. global environment는 "서버 프로세스를 최초로 띄운 그 순간의 클라이언트 환경"에 영구히 고정되므로, 다음과 같은 상황이 그대로 재현됩니다.

1. 터미널 A(예: kitty)에서 `tmux new-session`을 실행해 tmux 서버가 처음 뜹니다. 이때 kitty가 심어둔 `KITTY_WINDOW_ID`가 global environment에 박제됩니다.
2. 터미널 A를 완전히 닫습니다. kitty 창 자체는 더 이상 존재하지 않습니다.
3. 나중에 전혀 다른 터미널 B(또는 SSH로 원격 접속한 세션)에서 `tmux attach`로 그 서버에 재접속합니다.
4. 이미 살아 있던 pane은 물론, 이 시점 이후 **새로 만든 pane조차** `update-environment` 목록에 없는 한 여전히 `KITTY_WINDOW_ID`를 물려받습니다 — 존재하지도 않는 kitty 창을 계속 보고합니다.

이 현상을 로컬 tmux 3.7c로 실측했습니다: `KITTY_WINDOW_ID=123`을 export하고 tmux 서버를 띄운 뒤 값을 바꾸거나 unset해도, 그 서버 안에서 새로 만든 pane은 여전히 최초 값 `123`을 `env`에서 보고했습니다(global environment가 서버 재시작 없이는 갱신되지 않으므로).

### 4. `set-environment` / `show-environment` — 수동 탈출구

man page:

> `set-environment [-Fhgru] [-t target-session] variable [value]` (alias: `setenv`) — Set or unset an environment variable. If `-g` is used, the change is made in the global environment; otherwise, it is applied to the session environment for `target-session`. ... `-r` indicates the variable is to be removed from the environment before starting a new process. `-h` marks the variable as hidden.
>
> `show-environment [-hgs] [-t target-session] [variable]` (alias: `showenv`) — Display the environment for `target-session` or the global environment with `-g`. ... Variables removed from the environment are prefixed with `-`.

([원문 링크](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1#L6988-L7020))

`set-environment`는 `update-environment` 목록을 우회해 임의의 변수를 즉시 session(기본) 또는 `-g`로 global 환경에 반영할 수 있는 유일한 공식 수단입니다. 다만 이 역시 **이미 실행 중인 프로세스에는 영향을 주지 않고, 그 이후 새로 만들어지는 pane/window부터** 적용됩니다 — 위 2번 사실과 동일한 제약입니다. `show-environment`는 진단 도구로서 `runby`가 실행 시점의 실제 tmux 세션/global 환경을 (tmux CLI를 통해) 조회하고 싶을 때 쓸 수 있지만, `runby`는 프로세스가 상속받은 환경변수만 읽으므로 이 명령을 직접 호출하지는 않습니다.

### 5. 세션 안에서 `new-session`을 또 실행하면?

기존 세션 안에서 다시 `tmux new-session`을 실행하면 tmux는 기본적으로 이를 **중첩 실행**으로 간주해 "sessions should be nested with care, unset $TMUX to force"라는 오류로 거부합니다(`$TMUX`가 이미 설정되어 있음을 감지) — 이는 `cmd-new-session.c`와 `cmd-attach-session.c` 양쪽에서 동일한 검사로 확인됩니다. 사용자가 `$TMUX`를 명시적으로 unset하고 강제로 진행하면, 새로 뜨는 서버(또는 중첩된 클라이언트)는 그 시점 셸의 환경(이미 바깥쪽 tmux가 병합해 넘긴 환경)을 기준으로 **또 다른** global environment를 구성합니다. 즉 중첩 자체가 값을 특별히 다르게 다루지는 않지만, 중첩 서버의 global environment는 "바깥쪽 tmux pane의 환경"을 물려받으므로 원본 터미널의 신호가 한 단계 더 간접화됩니다.

## 다른 축에 미치는 영향

`runby`는 **에이전트**(누가 명령을 요청했는가), **CI**(어디서 실행되는가), **터미널**(어떤 에뮬레이터가 이 환경을 만들었는가)에 **실행 도구**와 **원격**을 더해 다섯 축을 보고합니다. tmux 자신은 원격 축의 `RemoteTmux`(`RemoteKindMultiplexer`) 계층으로 보고되므로, 아래에서는 tmux가 **나머지 축**을 무엇을 어떻게 오염시키는지 정리합니다.

- **터미널 축 — 직접 오염됨, 방향은 "그럴듯하지만 낡은 값(stale-but-plausible)".** 위에서 확인했듯 `TERM_PROGRAM`/`TERM_PROGRAM_VERSION`/`COLORTERM`은 tmux가 pane마다 스스로 `"tmux"`/tmux 버전/`"truecolor"`로 **강제 재작성**하므로([소스](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/environ.c#L266-L267)) — 이는 문서·소스 모두로 확인했고 실측으로도 재현했습니다(export한 가짜 `TERM_PROGRAM=fakevalue`가 tmux pane 안에서는 무조건 `tmux`로 관측됨) — 오히려 **바깥 터미널의 정체를 완전히 지워버립니다(위양성이 아니라 신호 소실)**. 반면 `KITTY_WINDOW_ID`, `WEZTERM_PANE`, `ITERM_SESSION_ID` 같은 터미널 고유 마커는 tmux가 건드리지 않고 global environment를 통해 그대로 통과시키므로, 이 값들은 **거짓 양성에 가까운 낡은 값**을 만듭니다 — 존재하는 것은 확인되지만 그 값이 가리키는 터미널 창은 이미 닫혔을 수 있습니다. 요약하면 방향이 신호마다 다릅니다: `TERM_PROGRAM` 계열은 "지워짐", 제품별 pane/window ID 계열은 "낡았지만 그럴듯함"입니다.
- **CI 축 — 간접적으로만 영향.** CI 러너 안에서 tmux를 쓰는 경우는 드물지만, 만약 CI 스텝이 tmux 세션 안에서 실행된다면 `GITHUB_ACTIONS`, `CI` 같은 CI 플랫폼 변수는 대개 `update-environment` 기본 목록에 없으므로 global environment를 통해 그대로 통과합니다 — 즉 CI 감지 자체는 tmux로 인해 크게 훼손되지 않습니다. 다만 CI 러너가 세션을 오래 유지하며 재사용하는 구성이라면, 오래된 잡의 환경변수가 global environment에 남아 다음 잡의 pane에 새어 들어갈 이론적 위험이 있습니다(문서로 확인된 사실이 아니라 위 메커니즘에서 유추한 위험이므로, 이런 구성을 쓰는 경우 별도 검증이 필요합니다).
- **에이전트 축 — CI와 동일한 논리.** `PASEO_AGENT_ID`, `CLAUDECODE` 같은 에이전트 실행 식별 변수도 대개 `update-environment` 기본 목록 밖에 있으므로, tmux는 이들을 능동적으로 지우거나 오염시키지 않고 global/session 병합 규칙을 그대로 적용합니다. 다만 "누가 tmux 세션을 시작했는가"와 "지금 이 pane에서 명령을 보낸 주체가 누구인가"는 다른 질문이며, 에이전트가 `tmux send-keys`로 기존 pane에 명령을 주입하는 경우 그 pane의 자식 프로세스 환경에는 애초에 에이전트 신호가 없을 수 있습니다 — 이는 에이전트 신호값의 변형이 아니라 애초에 신호가 전달되는 경로가 다르다는 구조적 한계입니다.

**결론적으로 tmux가 구조적으로 오염시키는 축은 터미널 축 하나뿐입니다.** `runby`가 현재 하는 일 — `TMUX` 존재를 `RemoteTmux` 계층으로 보고하고, 그때 터미널 판정과 (조상으로 확증되지 않은) 에이전트·러너 판정의 `Confidence`를 `probable`로 낮추는 것 — 은 방향은 옳지만 터미널 축의 세밀함이 부족합니다. 위 조사로 확인했듯 tmux 내부에서는 `TERM_PROGRAM` 계열(신호 소실, 검사할 가치가 낮음)과 제품별 pane ID 계열(낡았지만 존재 여부는 유의미, `KITTY_PID`처럼 생존 검사가 가능한 값도 있음)이 서로 다르게 취급됩니다. 더 강한 대응으로 고려할 수 있는 것은:

- `TMUX_PANE`(pane ID)을 함께 노출하는 것 — **이미 반영했습니다.** `RemoteTmux` 계층의 `Extra["tmux.pane"]`로 나오고, 계층의 `SessionID`는 `TMUX` 값입니다. 값 자체는 man page가 "생애 동안 불변"이라고 보장하므로, 같은 프로세스 트리 안에서 pane 재사용 여부를 안정적으로 구분하는 데 쓸 수 있습니다(다만 이는 어떤 터미널이 지금 그 pane을 보고 있는지와는 무관한 정보입니다).
- 다른 터미널 문서들과 마찬가지로, tmux를 경유해 관측된 터미널 마커(`KITTY_WINDOW_ID` 등)는 `probable`보다 더 올릴 수 없다는 규칙을 명시적으로 문서화하는 것 — global environment가 재접속으로 갱신되지 않는다는 사실이 소스 코드 수준에서 확정적이므로, 이는 추정이 아니라 증명된 상한선입니다.
- 그 이상(예: "지금 이 pane에 실제로 attach된 클라이언트가 어떤 터미널인지")은 환경변수만으로는 원리적으로 알 수 없습니다 — tmux 클라이언트-서버 프로토콜 자체를 통해 `list-clients`를 조회해야 하는 영역이며, 이는 `runby`가 다루는 "상속된 환경변수" 범위를 벗어납니다.

## 실행 주체 감지에 관한 결론

tmux는 "자식 프로세스가 상속한 환경변수"라는 `runby`의 기본 가정 자체를 구조적으로 깨뜨리는 유일한 대상입니다. 다른 문서들의 신호는 상속되지만 최소한 처음에는 정확했던 값이 시간이 지나며 낡습니다. tmux는 그보다 한 단계 더 나아가, **애초에 어떤 값이 자식에게 전달될지를 서버가 능동적으로 선별**합니다 — `update-environment` 기본 목록(현재 릴리스 기준 `DISPLAY KRB5CCNAME MSYSTEM SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WAYLAND_DISPLAY WINDOWID XAUTHORITY XDG_CURRENT_DESKTOP XDG_SESSION_DESKTOP XDG_SESSION_TYPE`)에 없는 모든 변수는 세션 attach로 갱신되지 않고, 서버를 최초로 띄운 클라이언트의 환경에 영구히 고정됩니다. 이는 man page의 명시적 서술과 소스 코드(`tmux.c`의 `main()`에서 `global_environ`이 딱 한 번 구성됨, `environ.c`의 `environ_for_session`이 매 pane 생성 시 global+session을 병합)로 이중 확인됐고, 로컬 tmux 3.7c 실측으로도 동일하게 재현됐습니다.

`TMUX`, `TMUX_PANE`은 "지금 tmux 안에 있다"는 사실과 pane의 생애 동안 불변인 식별자를 신뢰성 있게 제공하므로 `실행 식별`로 쓰기에 적합합니다. 그러나 tmux를 경유한 다른 모든 터미널 신호(예: `KITTY_WINDOW_ID`)는 "tmux가 이 값을 어디서 가져왔는지"를 알아야만 신뢰도를 판단할 수 있고, 그 답은 언제나 "서버를 처음 띄운 순간"이라는 과거 시점입니다. 따라서 `runby`가 `TMUX` 존재만으로 터미널 신뢰도를 `probable`로 낮추는 현재 접근은 방향상 올바르며, 이보다 더 강하게(`definite`로) 올릴 수 있는 공식적으로 검증된 방법은 없습니다.

## 공식 문서

- [tmux(1) man page](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1)
- [GLOBAL AND SESSION ENVIRONMENT 절](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.1#L6958-L7020)
- [소스: `environ.c` — `environ_for_session()`](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/environ.c#L251-L284)
- [소스: `options-table.c` — `update-environment` 기본값](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/options-table.c#L1025-L1034)
- [소스: `tmux.c` — `global_environ` 최초 구성](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.c#L411-L413)
- [소스: `tmux.h` — `TMUX_TERM` 기본값](https://github.com/tmux/tmux/blob/81f88f8517c9fc5371b56cf117530c6b477c96ac/tmux.h#L93)
- [공식 소스 저장소 (`tmux/tmux`)](https://github.com/tmux/tmux)
- [릴리스 태그 3.7](https://github.com/tmux/tmux/releases/tag/3.7)
