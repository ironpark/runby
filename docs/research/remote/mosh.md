---
title: Mosh
slug: mosh
research_date: 2026-08-31
open_source: true
repository: https://github.com/mobile-shell/mosh
product_type: remote_environment
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 소스(고정 커밋 decd9b705eb81626f694335b8d5940538beb06da)와 세 man 페이지의 ENVIRONMENT VARIABLES 절만으로도 각 MOSH_* 변수가 어디서 getenv/setenv/unsetenv되는지는 명확히 특정됐지만, mosh.1의 ENVIRONMENT VARIABLES 절은 MOSH_ESCAPE_KEY·MOSH_PREDICTION_DISPLAY·MOSH_TITLE_NOPREFIX를 "mosh-server(1)로 전달된다"고 적어 실제 소스(모두 클라이언트 쪽 stmclient.cc/mosh-client.cc에서만 읽음)와 어긋난다. 이런 문서-소스 불일치가 있는 채로 "원격 셸에는 MOSH_* 가 사실상 남지 않는다"는 보안 관련 결론을 내리므로, 실제 mosh 세션 안에서 `env`/`cat /proc/self/environ`을 실행해 (a) MOSH_KEY가 정말 전혀 보이지 않는지, (b) MOSH_SERVER_NETWORK_TMOUT/SIGNAL_TMOUT을 미리 export해 두었을 때 실제로 셸까지 전달되는지, (c) 배포판별 sshd 설정에 따라 SSH_CONNECTION이 항상 셸까지 살아남는지를 한 번은 실측으로 재확인할 가치가 있다.
---

# Mosh

Mosh(mobile shell)는 SSH로 원격 호스트에 `mosh-server`를 부트스트랩한 뒤, 이후의 실제 세션은 SSH가 아니라 자체 UDP 프로토콜(SSP)로 이어가는 원격 셸 도구입니다. 로컬에서 실행되는 것은 `mosh`(부트스트랩 스크립트)와 `mosh-client`(터미널 프런트엔드)이고, 원격에는 `mosh-server`가 사용자의 로그인 셸을 실행합니다. 아래 내용은 공식 소스 [`mobile-shell/mosh`](https://github.com/mobile-shell/mosh) 커밋 [`decd9b705`](https://github.com/mobile-shell/mosh/commit/decd9b705eb81626f694335b8d5940538beb06da)의 `src/frontend/mosh-server.cc`, `src/frontend/mosh-client.cc`, `src/frontend/stmclient.cc`와 공식 man 페이지 `mosh.1`, `mosh-server.1`, `mosh-client.1`([mosh.org](https://mosh.org/#getting))을 기준으로 확인했습니다.

이 문서의 결론은 다른 문서들과 방향이 다릅니다. 대부분의 터미널·에이전트 문서는 "신호가 있지만 상속·잔존 때문에 신뢰도가 떨어진다"는 형태지만, Mosh는 **애초에 검사할 신호가 원격 셸의 환경에 거의 남지 않도록 설계**되어 있습니다.

## 실행 식별 신호

공개 환경변수만으로 "이 프로세스는 Mosh 세션 안에서 실행 중이다"를 정적으로 확정하는 신뢰 가능한 마커는 확인되지 않았습니다. 이유는 세 갈래입니다.

1. **`MOSH_KEY`는 원격 셸 쪽 환경에 애초에 존재하지 않습니다.** 세션 키는 `mosh-server`가 프로세스 내부 메모리(`network->get_key()`)에서 생성해 `MOSH CONNECT <port> <key>` 형태로 표준출력(SSH 부트스트랩 채널)에 한 번 인쇄할 뿐([`mosh-server.cc`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc), `printf( "MOSH CONNECT %s %s\n", ... )`), 서버 프로세스 자신의 환경변수로 `setenv`된 적이 없습니다. `mosh-server.1`의 `ENVIRONMENT VARIABLES` 절에도 `MOSH_KEY`는 등장하지 않고 `MOSH_SERVER_NETWORK_TMOUT`/`MOSH_SERVER_SIGNAL_TMOUT` 둘뿐입니다. 즉 로그인 셸을 만드는 `forkpty()` 호출([L526](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L526))이 상속하는 부모 환경에 `MOSH_KEY`가 들어갈 경로 자체가 없습니다.
2. **클라이언트 쪽에서는 존재하지만 즉시 지워집니다.** 로컬 `mosh` 스크립트가 SSH로 키를 받아 `mosh-client`를 실행할 때만 `MOSH_KEY` 환경변수로 넘기고([`mosh-client.1`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-client.1#L21) `MOSH_KEY=KEY`), `mosh-client`는 시작하자마자 이를 읽어 문자열로 복사한 뒤 `unsetenv("MOSH_KEY")`를 호출합니다([`mosh-client.cc` L167-184](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-client.cc#L167-L184)). `mosh-client`는 셸을 fork하지 않는 터미널 프런트엔드 프로그램이므로, 이 값이 어떤 자손 프로세스로도 상속되지 않습니다.
3. **나머지 클라이언트 전용 변수(`MOSH_ESCAPE_KEY`, `MOSH_PREDICTION_DISPLAY`, `MOSH_PREDICTION_OVERWRITE`, `MOSH_TITLE_NOPREFIX`)도 원격 셸과 무관합니다.** 모두 로컬 터미널 프런트엔드인 `stmclient.cc`/`mosh-client.cc`에서만 `getenv`로 읽히며([`stmclient.cc` L126, L132](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/stmclient.cc#L126-L132)), 원격 `mosh-server`나 사용자의 로그인 셸에는 전달되지 않습니다.

**따라서 `MOSH_*` 이름의 변수가 하나도 없는 것 자체가 Mosh 세션의 정상적인 모습입니다.** 유일하게 원격 셸에 실제로 도달할 수 있는 `MOSH_*` 변수는 `MOSH_SERVER_NETWORK_TMOUT`/`MOSH_SERVER_SIGNAL_TMOUT`이며, 이마저도 Mosh가 스스로 설정하는 것이 아니라 관리자·사용자가 로그인 rc 파일 등에 미리 `export`해 둔 경우에만 존재하는 **선택적** 설정값입니다(`mosh-server.1`, "mosh-server passes these variables to the login session and shell that it starts").

**`MOSH_KEY`를 절대 읽거나 로그에 남기지 마십시오.** 이 변수는 세션 전체의 기밀성과 무결성을 보장하는 128비트 AES 키 그 자체이며(`mosh-client.1`, "This represents a 128-bit AES key that protects the integrity and confidentiality of the session"), 자격 증명입니다. 위에서 확인했듯 정상 경로에서는 `runby`가 실행되는 시점(원격 셸 또는 그 자손 프로세스)에는 이미 존재하지 않아야 하지만, 만약 어떤 비정상적인 경로(예: 사용자가 직접 `env MOSH_KEY=... mosh-server` 형태로 실험)로 이 변수가 관측 가능한 위치에 남아 있더라도, `runby`는 그 값을 읽거나 결과 구조체·로그·에러 메시지 어디에도 복사해서는 안 됩니다. 이는 이 라이브러리의 다른 모든 신호와 다른, 자격 증명 취급 규칙입니다.

## 환경 전달 규칙

Mosh의 실행 경로는 세 단계로 나뉘고, 각 단계마다 환경이 전달되는 방식이 다릅니다.

1. **SSH 부트스트랩 단계.** `mosh` 스크립트는 `ssh`(또는 `--ssh`로 지정한 명령)로 원격 호스트에 접속해 `mosh-server new`를 실행합니다([`mosh.1`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh.1)). 이때 실제 SSH 연결이 잠깐 맺어지며, `sshd`는 일반적인 SSH 세션과 동일하게 `SSH_CONNECTION`을 그 연결의 로컬/원격 주소·포트로 채워 `mosh-server` 프로세스의 환경에 심습니다. 이 SSH 연결은 `mosh-server`가 포트와 키를 출력한 직후 끊어지고, 이후의 실제 세션은 UDP로 이어집니다.
2. **`mosh-server` 자신의 설정 읽기.** `mosh-server`는 자신의 프로세스 환경에서 `MOSH_SERVER_NETWORK_TMOUT`([`mosh-server.cc` L395-406](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L395-L406))과 `MOSH_SERVER_SIGNAL_TMOUT`([L409-420](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L409-L420))을 `getenv`로 읽어 `strtol`로 파싱합니다. 이 값들은 사용자의 로그인 rc 파일이나 관리자가 시스템 설정에 미리 넣어둔 것이며, Mosh 자신이 부트스트랩 과정에서 주입하는 값이 아닙니다.
3. **로그인 셸로 넘어가는 지점.** `mosh-server`는 UDP 핸드셰이크가 끝난 뒤 `forkpty()`로 자식을 만들고([L526](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L526)), 그 자식이 `TERM`을 `xterm` 또는 `xterm-256color`로 강제 설정하고([L575](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L575)), `NCURSES_NO_UTF8_ACS=1`을 설정하고([L581](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L581)), `STY`를 `unsetenv`한 뒤([L587](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L587)) `execvp()`로 로그인 셸을 실행합니다([L624](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L624)). 이 세 개의 `setenv`/`unsetenv` 외에는 `mosh-server` 자신이 물려받은 환경(1단계의 SSH 부트스트랩 환경 + 2단계에서 읽었을 뿐 지우지는 않은 `MOSH_SERVER_*_TMOUT`)이 **그대로** 로그인 셸에 상속됩니다. 특히 `SSH_CONNECTION`을 지우는 코드는 소스 어디에도 없으므로, `sshd`가 이를 심어 두었다면 로그인 셸에도 그대로 남습니다.

**Mosh는 SSH의 `SendEnv`/`AcceptEnv`에 해당하는, 클라이언트 쪽 임의 환경변수를 원격으로 전달하는 메커니즘을 제공하지 않습니다.** `mosh` 스크립트가 클라이언트→서버 방향으로 넘기는 것은 접속 인자(호스트, 포트, `--server`/`--ssh` 커맨드라인 등)뿐이며, 로컬 셸의 임의 환경변수를 원격에 실어 보내는 옵션은 man 페이지 어디에도 없습니다. `MOSH_ESCAPE_KEY`/`MOSH_PREDICTION_DISPLAY`/`MOSH_PREDICTION_OVERWRITE`/`MOSH_TITLE_NOPREFIX`는 이름이 비슷해 혼동하기 쉽지만 전부 로컬 `mosh-client`만 소비하는 **클라이언트 측 UI 설정**이지, SSH의 `SendEnv`처럼 원격으로 전파되는 값이 아닙니다.

### 로밍이 만드는 특수한 종류의 신선도 문제

Mosh의 핵심 기능은 클라이언트의 IP가 바뀌거나(Wi-Fi ↔ 셀룰러 전환 등) 장시간 연결이 끊겨도 세션을 그대로 유지하는 것입니다. 이는 다른 문서들이 다루는 "상속된 값이 오래됐을 수 있다"는 일반적 SSH 신선도 문제보다 한 단계 더 강한 형태의 문제를 만듭니다.

- SSH에서는 세션이 끊기면 보통 그 프로세스 트리도 함께 끝나므로, 살아 있는 프로세스가 들고 있는 `SSH_CONNECTION` 값은 적어도 "그 프로세스를 만든 연결"을 가리킵니다.
- Mosh에서는 **연결이 완전히 끊겼다가 몇 시간 뒤 전혀 다른 네트워크(다른 IP, 다른 지역)에서 재개돼도 로그인 셸 프로세스는 죽지 않고 그대로 이어집니다.** 그런데 그 셸이 fork 시점에 물려받은 `SSH_CONNECTION`은 세션을 시작한 **최초의 짧은 SSH 부트스트랩 연결**의 로컬/원격 주소일 뿐이고, 로밍으로 클라이언트가 이동한 뒤에도 절대 갱신되지 않습니다(`mosh-server`가 이 값을 다시 쓰거나 지우는 코드가 없음을 소스에서 확인했습니다).
- 즉 이 변수가 남아 있다고 해서 "지금 이 프로세스에 붙어 있는 사람이 그 IP에 있다"는 뜻이 전혀 아니며, 오히려 세션이 오래 유지될수록 이 값이 현재 위치를 나타낼 가능성은 낮아집니다. 이는 tmux의 클라이언트별 서버 환경 잔존과 비슷한 모양이지만, Mosh는 "네트워크가 바뀌어도 안 죽는다"는 것 자체가 제품의 존재 이유이므로 이 신선도 붕괴가 예외가 아니라 **정상 동작의 일부**라는 점이 다릅니다.

### `TERM`과 터미널 에뮬레이션

Mosh는 원격 호스트에서 실제 터미널 에뮬레이터를 흉내 내지 않고, 오히려 `mosh-server`가 로그인 셸에 넘기는 `TERM` 값을 자신이 직접 결정합니다: 기본값은 `xterm`이고, `-c 256`(또는 그에 준하는 색상 협상) 시 `xterm-256color`로 강제 설정합니다([`mosh-server.cc` L570-578](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L570-L578), `mosh-server.1`의 `-c` 옵션 설명). 클라이언트가 SSH 부트스트랩 당시 어떤 `TERM`을 가지고 있었는지와 무관하게 이 값으로 덮어써지므로, 사용자 로컬 터미널의 실제 정체성(kitty, iTerm2 등)은 `TERM`에서 사라지고 Mosh가 스스로 에뮬레이션 능력을 광고하는 값으로 대체됩니다. 다만 `xterm`/`xterm-256color`는 매우 흔한 기본값이라(SSH로 아무 터미널도 없이 접속했을 때의 폴백과도 같은 값), `TERM`만으로 "이것은 Mosh다"라고 판단할 근거는 없습니다.

## 환경변수 표

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `MOSH_KEY` | Base64 22바이트 문자열(128비트 AES 키) | 실행 식별 | 클라이언트-서버 UDP 트래픽을 암호화하는 세션 공유키. `mosh-server`가 생성해 SSH 부트스트랩 출력으로만 전달 | 부적합 — 원격 셸의 환경에는 애초에 존재하지 않고(서버가 `setenv`한 적이 없음), 클라이언트 쪽에서도 `mosh-client`가 읽자마자 `unsetenv`한다. **절대 읽거나 로그에 남기면 안 되는 자격 증명** | [`mosh-client.1` ENVIRONMENT](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-client.1#L58-L64), [소스: `getenv`/`unsetenv`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-client.cc#L167-L184) |
| `MOSH_PREDICTION_DISPLAY` | 문자열(`adaptive`/`always`/`never`/`experimental` 등) | 설정 | 로컬 에코(입력 예측 표시) 모드 제어 | 부적합 — `mosh-client`(로컬 터미널 프런트엔드) 자신의 표시 설정이며 원격 셸에는 전달되지 않음 | [`mosh.1` L129, L309](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh.1#L129), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-client.cc#L174) |
| `MOSH_PREDICTION_OVERWRITE` | 문자열/존재 여부 | 설정 | 로컬 에코의 삽입(insert) vs 덮어쓰기(overwrite) 방식 제어 | 부적합 — 동일하게 클라이언트 전용, 원격 셸과 무관 | [`mosh.1` L150](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh.1#L150), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-client.cc#L178) |
| `MOSH_ESCAPE_KEY` | 단일 ASCII 문자(1-127) | 설정 | 로컬 명령(연결 종료, 클라이언트 일시정지 등) 진입용 이스케이프 문자 지정 | 부적합 — 클라이언트 전용 UI 설정, 원격 셸에 전달되지 않음 | [`mosh.1` ESCAPE SEQUENCES](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh.1#L272), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/stmclient.cc#L132) |
| `MOSH_TITLE_NOPREFIX` | 존재 여부(값 무관) | 설정 | 설정 시 창 제목에 `[mosh] ` 접두어를 붙이지 않음 | 부적합 — 클라이언트(로컬 터미널) 창 제목 설정이며 원격 셸과 무관 | [`mosh.1` L314](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh.1#L314), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/stmclient.cc#L126) |
| `MOSH_SERVER_NETWORK_TMOUT` | 양의 정수(초) | 설정 | 클라이언트로부터 갱신을 받지 못했을 때 `mosh-server`가 종료 대기하는 시간 | 보조 신호 — Mosh가 스스로 만드는 값이 아니라 관리자/사용자가 로그인 rc 파일에 미리 넣어야만 존재하는 선택적 설정. 존재한다면 man 페이지가 명시하듯 로그인 셸까지 전달되지만, 부재가 기본값이라 없다고 해서 Mosh가 아니라는 뜻은 아님 | [`mosh-server.1` ENVIRONMENT VARIABLES](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-server.1#L87-L107), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L395-L406) |
| `MOSH_SERVER_SIGNAL_TMOUT` | 양의 정수(초) | 설정 | 클라이언트 갱신 대기 중 `SIGUSR1`을 무시하는 시간 (연결 끊긴 세션 정리용) | 보조 신호 — 위와 동일한 이유로 선택적이며 opt-in일 때만 로그인 셸에 도달 | [`mosh-server.1` L109-121](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-server.1#L109-L121), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L409-L420) |
| `TERM` | 문자열(`xterm` 또는 `xterm-256color`) | 상태·컨텍스트 | `mosh-server`가 로그인 셸에 강제로 설정하는 terminfo 값 | 부적합 — `mosh-server`가 무조건 덮어쓰긴 하지만 값 자체가 SSH 등 다른 다수 도구의 기본 폴백과 동일해, 존재만으로 Mosh를 특정할 수 없음 | [`mosh-server.1` `-c` 옵션](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-server.1), [소스](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L570-L578) |
| `SSH_CONNECTION` (Mosh 고유 변수 아님) | `<local-ip> <local-port> <remote-ip> <remote-port>` | 상태·컨텍스트 | 원래는 `sshd`가 SSH 연결 주소를 알리는 표준 변수. `mosh-server`는 `--bind-server=ssh`(기본값) 판단에 이 값을 읽어 사용 | 부적합 — Mosh가 만든 값이 아니라 부트스트랩용 SSH 연결에서 상속됐을 뿐이며, 로밍 후에도 갱신되지 않아 최초 접속 당시 위치만 가리키는 정적 스냅샷. 어떤 SSH 세션에도 등장하는 값이라 Mosh 특정 신호가 아님 | [`mosh-server.1` `-i` 옵션](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-server.1#L58-L61), [소스: `get_SSH_IP()`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc#L154-L163) |

## 다른 축에 미치는 영향

`runby`는 agent(누가 실행을 요청했는가) / CI(어디서 실행되는가) / terminal(어떤 터미널이 소유하는가) / runner(무엇이 직접 실행했는가) / remote(사용자와 프로세스 사이에 무엇이 있는가) 다섯 축을 보고합니다. Mosh는 마커가 없어 remote 축에도 나타나지 않으며, 그러면서 특히 **terminal 축을 조용히 무력화**합니다.

- **Terminal 축.** SSH 위에서 동작하는 다른 터미널들(iTerm2의 `LC_TERMINAL`, kitty의 `kitten ssh` 등, [`terminals/README.md`](../terminals/README.md) 참고)은 최소한 "낡았을 수 있는 값"이라도 원격에 전달합니다. Mosh는 그런 전달 메커니즘 자체가 없고, 오히려 `TERM`을 `xterm`/`xterm-256color`로 강제 재작성해 로컬 터미널의 정체성을 지워버립니다. 즉 Mosh를 거친 원격 셸에서는 `TERM_PROGRAM`류 변수가 아예 도착하지 않을 뿐 아니라, 도착했더라도 `TERM` 자체가 로컬 터미널을 가리키지 않게 됩니다. 로밍까지 겹치면 유일하게 살아남는 `SSH_CONNECTION`조차 "지금 여기"가 아니라 "몇 시간/며칠 전 부트스트랩 시점"을 가리키므로, 터미널 축의 신뢰도는 이 문서가 다루는 어떤 다른 제품보다도 낮습니다.
- **Agent 축.** Mosh 자체는 에이전트를 실행하거나 구분하는 기능이 없습니다(`executes_agents: []`). Mosh 세션 안에서 Claude Code나 Codex 같은 에이전트가 실행된다면, 그 에이전트들의 고유 마커(`CLAUDECODE=1` 등)는 Mosh와 무관하게 정상적으로 상속·관측됩니다. Mosh는 이 축에 신호를 추가하지도, 지우지도 않습니다.
- **CI 축.** Mosh는 대화형 원격 셸 도구이므로 CI 플랫폼과 겹치는 시나리오 자체가 드뭅니다. 영향 없음으로 판단합니다.

**정직한 결론**: 원격 셸 안의 프로세스가 "나는 Mosh 세션 안에 있다"는 사실을 스스로도 재구성하기 어렵게 설계되어 있으므로, `runby`가 프로세스 환경변수만으로 이를 감지할 수 있는 방법은 사실상 없습니다. 라이브러리가 볼 수 없는 경계에 대해 경고를 낼 수는 없습니다. 따라서 `runby`가 취할 수 있는 정직한 태도는 다음과 같습니다.

1. Mosh 전용 감지 로직(마커 검사)을 추가하지 않는다 — 추가해봐야 상시 `false negative`만 만드는 죽은 코드가 된다.
2. 대신 이 문서와 같은 형태로 "Mosh는 구조적으로 탐지 불가능하며, 그 이유는 설계상 의도된 것"이라는 사실을 명시적으로 문서화해, 사용자가 `runby`의 터미널/원격 판정 결과를 "Mosh가 아니었다"가 아니라 "판단할 수 없었다"로 읽도록 안내한다.
3. 이미 확보한 `SSH_CONNECTION`/`TERM` 감지 로직(있다면)이 있는 문서(예: 터미널 축)에서, 그 값들이 Mosh 부트스트랩을 거쳤을 가능성까지 배제하지 못한다는 한계를 함께 언급한다 — 즉 "SSH를 거쳤다"는 판정에 Mosh의 짧은 부트스트랩 SSH 연결도 포함될 수 있음을 인지한다.
4. 무엇보다, `MOSH_KEY`라는 이름의 변수가 어떤 경로로든 관측된다면 이는 정상 경로를 벗어난 상황이므로, `runby`는 그 변수를 결과에 포함시키지 않고 값도 절대 로깅하지 않는다.

## 실행 주체 감지에 관한 결론

Mosh는 세션의 공유 비밀(`MOSH_KEY`)이 원격 로그인 셸의 환경에 애초에 도달하지 않도록 설계됐고, 클라이언트 쪽에서도 `mosh-client`가 그 값을 읽는 즉시 `unsetenv`로 지웁니다. 나머지 `MOSH_ESCAPE_KEY`/`MOSH_PREDICTION_DISPLAY`/`MOSH_PREDICTION_OVERWRITE`/`MOSH_TITLE_NOPREFIX`도 전부 로컬 터미널 프런트엔드 전용 설정이라 원격 셸에는 원천적으로 전달되지 않습니다. 원격 셸에 실제로 도달할 수 있는 `MOSH_*` 변수는 `MOSH_SERVER_NETWORK_TMOUT`/`MOSH_SERVER_SIGNAL_TMOUT` 뿐이며, 이마저도 Mosh가 스스로 주입하는 것이 아니라 사용자가 미리 `export`해 둔 경우에만 존재하는 선택적 신호이므로 "보조 신호"이지 마커가 아닙니다. `TERM`은 `mosh-server`가 무조건 재작성하지만 값 자체가 SSH의 일반적인 폴백과 동일해 단독으로는 무의미합니다. `SSH_CONNECTION`은 상속되어 살아남지만 Mosh 고유 변수가 아니고, 로밍이라는 Mosh의 핵심 기능 때문에 세션이 오래갈수록 실제 클라이언트 위치와 점점 더 멀어지는 정적 스냅샷일 뿐입니다.

결론적으로, **정상적인 Mosh 세션 안의 프로세스는 `MOSH_*` 환경변수를 하나도 가지고 있지 않은 것이 오히려 정상**이며, 환경변수만으로 "이 프로세스는 Mosh 세션 안에서 실행 중이다"를 판정할 신뢰 가능한 방법은 존재하지 않습니다. `runby`는 Mosh에 대해 별도의 긍정 판정 로직을 두지 않고, 이 구조적 한계를 문서로 명시하는 쪽을 선택해야 합니다.

## 공식 문서

- [Mosh 공식 사이트](https://mosh.org/)
- [`mosh(1)` man 페이지](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh.1)
- [`mosh-server(1)` man 페이지](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-server.1)
- [`mosh-client(1)` man 페이지](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/man/mosh-client.1)
- [공식 소스 저장소 (`mobile-shell/mosh`)](https://github.com/mobile-shell/mosh)
- [소스: `src/frontend/mosh-server.cc`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-server.cc)
- [소스: `src/frontend/mosh-client.cc`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/mosh-client.cc)
- [소스: `src/frontend/stmclient.cc`](https://github.com/mobile-shell/mosh/blob/decd9b705eb81626f694335b8d5940538beb06da/src/frontend/stmclient.cc)
