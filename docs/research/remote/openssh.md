---
title: OpenSSH
slug: openssh
research_date: 2026-08-31
open_source: true
repository: https://github.com/openssh/openssh-portable
product_type: remote_environment
executes_agents: []
runtime_test_required: false
runtime_test_reason: ssh(1)·sshd(8)·ssh_config(5)·sshd_config(5) 공식 man 페이지가 SSH_CLIENT/SSH_CONNECTION/SSH_TTY/SSH_ORIGINAL_COMMAND의 값 형식과 조건부 존재 규칙을 명시하고, OpenSSH 공식 소스(session.c의 do_setup_env)에서 실제 주입 코드와 값 포맷을 1:1로 확인했으므로 별도 실행 검증 없이 표를 확정할 수 있음. 다만 SendEnv/AcceptEnv의 "배포판이 실제로 무엇을 켜서 출하하는가"는 배포판마다 달라 이 문서가 인용한 특정 스냅샷(OpenSSH 컴파일 기본값, Debian 계열 예시) 밖의 배포판에 대해서는 해당 배포판의 실제 설정 파일을 직접 확인해야 함
---

# OpenSSH

`runby`가 다루는 축들(에이전트·CI·터미널·실행 도구·원격)은 모두 "이 프로세스를 누가 만들었는가"를 로컬 환경변수로 추론합니다. OpenSSH는 이 추론이 성립하는 **경계 자체**입니다 — SSH는 로컬 프로세스를 만드는 제품이 아니라, 한 머신의 프로세스 트리를 다른 머신으로 이어붙이면서 그 경계를 넘는 환경변수를 의도적으로 걸러내는 프로토콜입니다. 그래서 이 문서는 다른 문서들과 달리 "OpenSSH를 어떻게 식별하는가"보다 "OpenSSH를 넘으면 다른 축의 신호들에게 무슨 일이 일어나는가"를 다룹니다. 아래 내용은 조사 시점 기준 공식 소스 [`openssh/openssh-portable`](https://github.com/openssh/openssh-portable)와 OpenBSD 공식 man 페이지([`ssh(1)`](https://man.openbsd.org/ssh.1), [`sshd(8)`](https://man.openbsd.org/sshd.8), [`ssh_config(5)`](https://man.openbsd.org/ssh_config.5), [`sshd_config(5)`](https://man.openbsd.org/sshd_config.5))를 기준으로 확인했습니다.

## 실행 식별 신호

sshd는 세션의 자식 프로세스 환경을 `session.c`의 `do_setup_env()`에서 직접 구성합니다. 이 함수가 정확히 어떤 변수를 어떤 조건에서 어떤 형식으로 넣는지가 이 절의 근거입니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `SSH_CONNECTION` | 공백으로 구분된 4개 값: 클라이언트 IP, 클라이언트 포트, 서버 IP, 서버 포트 | 실행 식별 | 클라이언트-서버 연결 양 끝을 식별 | 적합 — sshd가 세션마다 무조건 설정하며(`ssh(1)` ENVIRONMENT), 소스에서도 `child_set_env(&env, &envsize, "SSH_CONNECTION", buf)`가 조건 없이 실행됨을 확인. 현재 OpenSSH가 공식적으로 유지하는 연결 식별자 | [`ssh(1)` ENVIRONMENT — SSH_CONNECTION](https://man.openbsd.org/ssh.1#ENVIRONMENT), [소스 `session.c`](https://github.com/openssh/openssh-portable/blob/master/session.c) |
| `SSH_CLIENT` | 공백으로 구분된 3개 값: 클라이언트 IP, 클라이언트 포트, **서버 포트만**(서버 IP 없음) | 실행 식별 | 레거시 클라이언트 연결 식별자 | 보조 신호 — 존재 자체는 SSH 세션임을 강하게 시사하지만, OpenSSH 소스 코드가 이 변수를 설정하는 줄 바로 위에 `/* SSH_CLIENT deprecated */` 주석을 달아 두었고 현재 `ssh(1)` man 페이지 ENVIRONMENT 절에는 아예 등재되어 있지 않음(호환성을 위해 계속 설정만 함). 새 코드는 `SSH_CONNECTION`을 써야 함 | [소스 `session.c` — `/* SSH_CLIENT deprecated */`](https://github.com/openssh/openssh-portable/blob/master/session.c) |
| `SSH_TTY` | 문자열, 할당된 pty 디바이스 경로(예: `/dev/pts/N` 형태) | 실행 식별 | 현재 셸/명령에 연결된 tty 식별 | 적합, 단 조건부 — `ssh(1)`이 "현재 세션에 tty가 없으면 이 변수는 설정되지 않는다"고 명시하고, 소스에서도 `if (s->ttyfd != -1)`일 때만 주입됨. 즉 `ssh host command`처럼 tty를 요청하지 않은 비대화형 실행에는 `SSH_CONNECTION`은 있어도 `SSH_TTY`는 없음 — "SSH를 통해 들어왔는가"와 "대화형 tty가 붙어 있는가"를 분리해서 알려주는 유일한 변수 | [`ssh(1)` ENVIRONMENT — SSH_TTY](https://man.openbsd.org/ssh.1#ENVIRONMENT), [소스 `session.c` — `if (s->ttyfd != -1)`](https://github.com/openssh/openssh-portable/blob/master/session.c) |
| `SSH_ORIGINAL_COMMAND` | 문자열, 클라이언트가 원래 요청한 커맨드라인 | 실행 식별 | `authorized_keys`의 `command=` 강제 커맨드나 `sshd_config`의 `ForceCommand`가 실제 실행 커맨드를 덮어썼을 때, 클라이언트가 원래 보내려던 커맨드를 보존 | 보조 신호 — 값이 있으면 "강제 커맨드 경유"라는 강한 정황이지만, 소스에서 `if (original_command)`일 때만 조건부로 설정되므로 부재가 "SSH가 아니다"를 뜻하지 않음. 일반 대화형 SSH 로그인에는 아예 설정되지 않음 | [`sshd(8)` AUTHORIZED_KEYS FILE FORMAT — command=](https://man.openbsd.org/sshd.8), [소스 `session.c` — `if (original_command)`](https://github.com/openssh/openssh-portable/blob/master/session.c) |

**가장 신뢰할 수 있는 마커는 `SSH_CONNECTION`입니다.** sshd가 예외 없이 설정하고, 현재 공식 man 페이지가 유지하는 변수이며, `SSH_CLIENT`처럼 소스 코드 주석으로 명시적 폐기(deprecated) 딱지가 붙어 있지 않습니다. `SSH_CLIENT`는 여전히 설정되므로 존재 여부만으로는 여전히 유용한 보조 신호지만, 새 판정 로직의 1차 근거로 쓰기에는 OpenSSH 자신이 이미 "쓰지 말라"는 신호를 코드에 남겨 두었습니다.

`SSH_TTY`의 유무는 판정을 한 단계 더 세분화합니다. `SSH_CONNECTION`은 있지만 `SSH_TTY`가 없다면 "SSH로 들어왔지만 tty가 없는 비대화형 실행"(`ssh host 'command'`, CI에서 흔한 원격 배치 실행, `-T` 옵션 등)이라는 뜻이고, 둘 다 있다면 대화형 로그인 셸이라는 뜻입니다. 이 구분은 `runby`가 "SSH 세션인가"와 "그 세션에 사람이 타이핑할 수 있는 tty가 있는가"를 분리해서 보고할 수 있게 해 주는 유일한 공식 근거입니다.

## 환경 전달 규칙

다른 모든 문서([`docs/research/terminals/README.md`](../terminals/README.md) 등)가 "OpenSSH 기본값은 `TERM`과 `LANG`/`LC_*` 외에는 전달하지 않는다"고 인용하는 근거가 바로 이 절입니다. 아래는 그 주장을 man 페이지 원문으로 직접 검증한 결과입니다.

### `SendEnv` (클라이언트, `ssh_config(5)`)

> "Specifies what variables from the local `environ(7)` should be sent to the server. [...] The default is not to send any environment variables."

**OpenSSH 자체의 컴파일 기본값은 "아무 변수도 보내지 않음"입니다.** 그런데 이 컴파일 기본값과 실제로 사용자 머신에 깔리는 설정 파일은 다릅니다. Debian/Ubuntu 계열이 배포하는 `/etc/ssh/ssh_config`에는 다음 줄이 기본으로 **활성화된 채** 들어 있습니다.

```
SendEnv LANG LC_*
```

즉 "OpenSSH는 기본적으로 `LANG`/`LC_*`를 보낸다"는 흔한 서술은 정확히는 **OpenSSH 소프트웨어 자체의 기본값이 아니라 주요 리눅스 배포판이 출하하는 설정 파일의 기본값**입니다. 이 구분은 실무적으로 중요합니다 — macOS에 내장된 OpenSSH나 `ssh_config`를 직접 최소 구성으로 컴파일한 환경에서는 `LC_*`조차 전달되지 않을 수 있습니다.

### `AcceptEnv` (서버, `sshd_config(5)`)

> "Specifies what environment variables sent by the client will be copied into the session's `environ(7)`. [...]" 기본값은 "not to accept any environment variables"입니다.

여기도 마찬가지로 컴파일 기본값은 "아무것도 받지 않음"이지만, Debian 계열이 배포하는 `/etc/ssh/sshd_config`에는 다음 줄이 활성화되어 있습니다.

```
AcceptEnv LANG LC_*
```

(최신 Debian 개발 브랜치는 여기에 `COLORTERM NO_COLOR`까지 추가하는 등, 배포판별로 이 줄의 내용이 계속 넓어지는 추세입니다.) 즉 실제로 변수가 전달되려면 **클라이언트의 `SendEnv`와 서버의 `AcceptEnv` 양쪽이 같은 이름을 허용해야** 하며, 둘 다 배포판이 채워 넣은 값이지 OpenSSH 프로토콜/소프트웨어 자체가 강제하는 기본값이 아닙니다.

### `SetEnv` — `SendEnv`/`AcceptEnv`와의 차이

클라이언트 쪽 `ssh_config`의 `SetEnv`는 `NAME=VALUE` 형태로 값 자체를 직접 지정해 서버로 보냅니다. `SendEnv`와 달리 로컬 환경에 그 변수가 있어야 할 필요가 없고, `TERM`을 제외하면 서버가 여전히 이를 받아들이도록 설정돼 있어야 합니다. 서버 쪽 `sshd_config`의 `SetEnv`는 반대 방향으로, sshd가 세션 시작 시 강제로 심는 값이며 공식 문서가 "`SetEnv`로 설정된 변수는 기본 환경 및 `AcceptEnv`/`PermitUserEnvironment`로 사용자가 지정한 값을 덮어쓴다"고 명시합니다. 즉 `SendEnv`/`AcceptEnv`는 "로컬에 있는 값을 골라 전달"하는 메커니즘이고, `SetEnv`는 "관리자가 값 자체를 주입"하는 별개의 상위 메커니즘입니다.

### `PermitUserEnvironment`와 `~/.ssh/environment` — 또 다른 주입 경로

`sshd_config(5)`의 `PermitUserEnvironment`는 `~/.ssh/environment` 파일과 `~/.ssh/authorized_keys`의 `environment=` 옵션을 sshd가 처리할지를 결정합니다. 기본값은 **`no`**이며, 문서는 이를 켜면 `LD_PRELOAD` 등을 이용한 접근 제한 우회가 가능해질 수 있다는 보안 경고를 함께 달아 두었습니다. 기본 `no` 상태에서는 이 경로로 변수가 전달되지 않으므로, `SendEnv`/`AcceptEnv`/`SetEnv`와 별개로 존재하는 세 번째 주입 경로이지만 기본 설정에서는 닫혀 있습니다.

### `TERM`은 예외 — `SendEnv`를 타지 않고 프로토콜 자체로 전달됩니다

`ssh_config(5)`의 `SendEnv` 설명 안에는 "pty가 요청되었을 때는 `TERM` 변수가 항상 전송된다"는 문장이 포함돼 있고, `sshd_config(5)`의 `AcceptEnv` 설명에도 대칭적으로 "pty가 요청되면 `TERM`은 항상 받아들여진다"는 문장이 있습니다. 이는 `TERM`이 `SendEnv`/`AcceptEnv` 필터를 우회한다는 뜻이 아니라,애초에 그 필터가 적용되는 채널(SSH 프로토콜의 환경변수 요청 메시지)과 **다른 채널**을 탄다는 뜻입니다. `TERM`은 SSH Connection Protocol의 pty-req 메시지 자체에 실려 전달되는 필드이기 때문에, `SendEnv`가 아무것도 지정하지 않은 최소 설정에서도 대화형 세션이라면 `TERM`은 항상 원격에 도달합니다. 이것이 다른 모든 변수가 막힌 기본 SSH 구성에서도 `TERM`만은 항상 원격 셸에 살아남는 이유입니다.

### 결론 — 기본 설정에서 실제로 경계를 넘는 것

- **OpenSSH 소프트웨어 자체의 컴파일 기본값**: 아무 환경변수도 넘지 않습니다(`TERM`은 pty-req로 별도 전달되므로 예외).
- **주요 배포판(Debian/Ubuntu 계열)이 출하하는 기본 설정**: 양쪽에 `SendEnv LANG LC_*` / `AcceptEnv LANG LC_*`가 모두 켜져 있어, `LANG`과 모든 `LC_*` 변수가 실제로 원격 호스트에 도달합니다.
- 이 `LC_*` 채널이 넓게 열려 있다는 사실이 바로 [`docs/research/terminals/iterm2.md`](../terminals/iterm2.md)가 경고하는 문제의 근본 원인입니다 — iTerm2는 터미널 식별자를 `LC_TERMINAL`/`LC_TERMINAL_VERSION`이라는 `LC_*` 네임스페이스 변수에 담기 때문에, 이 변수들은 SSH가 다른 모든 것을 막아도 배포판 기본 `SendEnv LC_*` 규칙에 걸려 아무 특별한 설정 없이 원격으로 전파됩니다.

## SSH_AUTH_SOCK·SSH_AGENT_PID·TERM

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `SSH_AUTH_SOCK` | 유닉스 도메인 소켓 경로 문자열 | 상태·컨텍스트 | ssh-agent 또는 에이전트 포워딩된 소켓 위치 | 부적합 — **SSH 세션 마커가 아닙니다.** `ssh-agent(1)`은 로컬에서 에이전트를 띄우기만 해도 이 변수를 설정하며, 이는 SSH 연결과 무관하게 로컬 데스크톱 세션(예: macOS 로그인 시 자동 실행되는 launchd ssh-agent)에서도 흔히 존재합니다. 반대로 `ssh -A`로 에이전트 포워딩이 켜진 SSH 세션에서는 원격 호스트에도 이 변수가 존재하게 됩니다. 즉 이 변수의 존재는 "SSH 세션 안에 있다"도 "SSH 세션 밖에 있다"도 증명하지 못하는 전형적인 오탐(false positive) 유발 신호입니다 | [`ssh-agent(1)` — SSH_AUTH_SOCK](https://man.openbsd.org/ssh-agent.1), [`ssh(1)` ENVIRONMENT — SSH_AUTH_SOCK](https://man.openbsd.org/ssh.1#ENVIRONMENT) |
| `SSH_AGENT_PID` | 정수 문자열 (ssh-agent 프로세스 PID) | 상태·컨텍스트 | 실행 중인 ssh-agent 프로세스 식별 | 부적합 — `SSH_AUTH_SOCK`과 동일한 이유로 SSH 세션 여부와 무관합니다. `ssh-agent`가 로컬에서 시작될 때 설정되는 값이며 원격 세션 진입과 직접 연결되지 않습니다 | [`ssh-agent(1)` — SSH_AGENT_PID](https://man.openbsd.org/ssh-agent.1) |
| `TERM` | 문자열, terminfo 이름 | 설정 | 터미널 능력 협상값. SSH pty-req로 전달됨 | 부적합 — SSH 여부와 무관하게 모든 tty 세션에 존재하고, `SendEnv`가 하나도 켜지지 않은 최소 설정에서도 살아남기 때문에 "SSH를 통과했는가"의 근거가 되지 못합니다. 오히려 `TERM_PROGRAM`류 값이 `SendEnv`로 새어 나가면(kitty의 `kitten ssh`, Ghostty의 `ssh-env`) 원격 호스트에 로컬에 존재하지도 않는 터미널을 보고하는 오탐 원인이 됩니다([`docs/terminals/README.md`](../terminals/README.md) 참고) | [`ssh_config(5)` — SendEnv (pty 요청 시 TERM 항상 전송)](https://man.openbsd.org/ssh_config.5) |

## 다른 축에 미치는 영향

`runby`는 에이전트·CI·터미널·실행 도구·원격 다섯 축을 보고합니다. SSH 자신은 원격 축의 `RemoteSSH`(`RemoteKindEnvironment`) 계층으로 보고되며, SSH 경계를 넘을 때 나머지 축이 받는 영향은 방향이 다릅니다.

- **가장 크게 위협받는 축은 터미널 축이며, 방향은 두 갈래로 갈립니다.** 기본 구성에서는 SSH가 원격 호스트에 도달하는 변수 수를 극단적으로 줄이므로(`SendEnv`가 비었거나 `LANG`/`LC_*`만 열림), 원격 프로세스는 로컬에서라면 존재했을 `TERM_PROGRAM`, `KITTY_WINDOW_ID`, `WT_SESSION` 같은 터미널 식별 변수를 대부분 **못 받습니다** — 이는 위양성이 아니라 **위음성**(실제로는 특정 터미널 안에서 실행 중인데 감지하지 못함) 방향입니다. 그런데 정확히 두 경로가 이 결과를 뒤집습니다. 첫째는 `LC_*` 채널 — iTerm2의 `LC_TERMINAL`처럼 `LC_*` 네임스페이스에 정체성을 숨긴 변수는 배포판 기본 `SendEnv LC_*` 규칙에 걸려 아무 특별한 설정 없이 새어 나갑니다. 둘째는 터미널이 SSH의 기본 원칙을 의도적으로 뒤집는 자체 기능 — kitty의 `kitten ssh`, Ghostty의 `ssh-env` 셸 통합은 로컬 터미널 상태를 원격 호스트에 능동적으로 복제합니다. 이 두 경로는 "원격 호스트의 프로세스가 그 머신에 존재하지도 않는 터미널을 보고"하는 **위양성**을 만듭니다. 요컨대 SSH는 터미널 축을 기본적으로 약화(위음성)시키지만, 일부 변수·일부 터미널에 한해 정반대의 오탐(위양성)을 만드는 비대칭적 위협입니다.
- **에이전트 축**은 상대적으로 덜 위협받습니다. `CLAUDECODE`, `CURSOR_AGENT` 같은 에이전트 마커는 `LC_*`/`LANG` 네임스페이스를 쓰지 않고 배포판 기본 `SendEnv`/`AcceptEnv` 목록에도 없으므로, 기본 SSH 구성을 통과한 프로세스라면 이런 변수가 없는 쪽이 정상입니다. 다만 에이전트가 스스로 `SendEnv`/`SetEnv`를 설정해 자기 마커를 원격까지 전파하도록 구성할 가능성은 이 문서가 확인한 범위 밖입니다.
- **CI 축**도 마찬가지로 CI 마커 이름(`CI`, `GITHUB_ACTIONS` 등)이 배포판 기본 허용 목록에 없어 기본적으로 막힙니다. 다만 CI 러너 자체가 SSH로 원격 빌드 호스트에 접속하는 구성이라면, 러너가 자기 설정으로 `SendEnv`를 확장해 CI 변수를 명시적으로 전달하는 경우가 실무에서 존재할 수 있으며, 이는 사이트별 설정이라 OpenSSH 기본값 조사만으로는 확인할 수 없습니다.

**`runby`가 SSH를 별도로 감지·보고해야 하는가.** 예. 지금은 에이전트 축의 한 종류가 아니라 **원격 축의 독립된 계층**(`Remote{Platform: RemoteSSH}`)으로 보고합니다 — 이 문서가 권고한 "컨텍스트 플래그"를 축으로 승격한 형태입니다. 다른 축의 신뢰도는 낮추지 않습니다: 값이 낡은 것이 아니라 다른 머신을 가리킬 수 있다는 뜻이므로, 계층의 존재 자체로 그 사실을 표현합니다. `SSH_CONNECTION`(및 `SSH_TTY` 유무)의 존재는 "이 프로세스가 로컬에서 시작되지 않았다"는 사실 하나만 확정하며, 이는 터미널 축의 판정에 대해 서로 다른 두 가지를 동시에 말해 줍니다 — (1) 터미널 축의 위음성 가능성이 높아진다는 경고(로컬 터미널 마커가 없다고 해서 터미널 밖에서 실행됐다고 단정하면 안 됨), (2) `LC_TERMINAL`류나 `kitten ssh`/`ssh-env`로 얻은 터미널 판정이라면 그 판정이 가리키는 터미널이 **다른 머신**에 있을 수 있다는 경고(위양성 방향의 원산지 표시)입니다. 이미 [`docs/research/terminals/README.md`](../terminals/README.md)의 "SSH가 더 이상 방벽이 아닙니다" 절이 이 결론을 전제로 삼고 있으므로, 이 문서는 그 전제를 공식 소스로 뒷받침하는 역할을 합니다. 반대로 SSH 존재를 CI나 에이전트 판정에 직접 개입시키는 것은 근거가 없습니다 — 두 축의 공식 마커는 애초에 SSH 기본 전달 규칙 밖에 있기 때문입니다.

## 실행 주체 감지에 관한 결론

"이 프로세스가 SSH 세션 안에 있는가"에 대한 가장 신뢰할 수 있는 단일 마커는 **`SSH_CONNECTION`**입니다 — sshd가 모든 세션에 예외 없이 설정하고, 현재 공식 `ssh(1)` man 페이지가 유지하는 변수이며, 레거시 딱지가 붙은 `SSH_CLIENT`와 달리 계속 쓰도록 문서화되어 있습니다. 여기에 `SSH_TTY`의 유무를 더하면 "SSH이면서 tty가 있는 대화형 세션"과 "SSH이지만 tty가 없는 비대화형 실행"을 구분할 수 있고, 이는 두 변수 모두 sshd 소스에서 확인한 조건부 주입 규칙(`SSH_TTY`는 `s->ttyfd != -1`일 때만)에 근거합니다.

반대로 `SSH_AUTH_SOCK`/`SSH_AGENT_PID`는 SSH 세션 마커로 **써서는 안 됩니다.** 두 변수 모두 `ssh-agent(1)`이 SSH 연결과 무관하게 로컬에서도 설정하며, 순수 로컬 데스크톱 세션에 흔히 존재합니다. 이 둘을 SSH 판정에 쓰면 로컬 세션을 SSH 세션으로 오판하는 고전적인 오탐이 발생합니다.

전달 규칙 쪽에서는, "OpenSSH는 기본적으로 `TERM`과 `LANG`/`LC_*`만 넘긴다"는 통념을 정밀하게 나눠야 합니다. **OpenSSH 소프트웨어 자체의 컴파일 기본값은 `TERM`(pty-req로 별도 전달)을 제외하면 아무것도 넘기지 않는 것**이고, `LANG`/`LC_*`가 넘어가는 것은 **Debian/Ubuntu 등 주요 배포판이 출하 시 `SendEnv`/`AcceptEnv`에 미리 넣어 둔 설정** 때문입니다. 이 구분이 흐려진 채로 통용되고 있으나, 이 문서가 공식 man 페이지와 배포판 설정 스냅샷으로 확인한 바로는 실제로 그렇게 나뉩니다. 결과적으로 `runby`는 SSH 자체를 별도 `Kind`로 승격하기보다, `SSH_CONNECTION`/`SSH_TTY`를 세 기존 축(에이전트·CI·터미널) 판정에 곁들이는 **컨텍스트 플래그**로 노출해, 특히 터미널 축에서 "이 판정이 다른 머신의 상태를 가리킬 수 있다"는 경고를 함께 전달하는 방식을 권장합니다.

## 공식 문서

- [`ssh(1)` — OpenSSH SSH client (remote login program)](https://man.openbsd.org/ssh.1)
- [`sshd(8)` — OpenSSH SSH daemon](https://man.openbsd.org/sshd.8)
- [`ssh_config(5)` — OpenSSH SSH client configuration files](https://man.openbsd.org/ssh_config.5)
- [`sshd_config(5)` — OpenSSH SSH daemon configuration file](https://man.openbsd.org/sshd_config.5)
- [`ssh-agent(1)` — OpenSSH authentication agent](https://man.openbsd.org/ssh-agent.1)
- [공식 소스 저장소 (`openssh/openssh-portable`) — `session.c`](https://github.com/openssh/openssh-portable/blob/master/session.c)
- [Debian 기본 `sshd_config` 스냅샷 예시](https://gist.github.com/mh61503891/98be80c6e18cb9f037b65324393d4177)
