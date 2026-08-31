---
title: Alacritty
slug: alacritty
research_date: 2026-08-31
open_source: true
repository: https://github.com/alacritty/alacritty
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: 소스는 어떤 변수가 어떤 플랫폼·설정에서 주입되는지 보여주지만, 실제 자식 셸(특히 tmux/SSH를 거친 프로세스)에서의 상속 결과와 macOS/Windows 바이너리의 IPC 기능 활성화 여부는 직접 실행해 확인해야 함
---

# Alacritty

Alacritty는 "터미널 실행 식별" 전용의 공식 변수 계약을 문서화하지 않습니다. 공식 소스에서 확인되는 것은 GPU 렌더링 설정, 로그, IPC 소켓, 창 ID, `TERM`/`COLORTERM` 같은 터미널 기능 변수뿐이며, 어떤 것도 "이 프로세스는 Alacritty가 실행했다"는 것을 증명하기 위해 설계되지 않았습니다. 그중 일부는 Alacritty가 스폰하는 셸에만 존재하는 이름을 가지므로 실무적으로는 존재 신호로 쓸 수 있지만, 아래에서 보듯 플랫폼과 버전, 설정에 따라 조건부로만 존재합니다.

## 터미널 식별 신호

Alacritty는 다른 여러 터미널이 쓰는 `TERM_PROGRAM` 계열 변수를 **설정하지 않습니다.** 공식 소스 전체(`alacritty`, `alacritty_terminal` 크레이트)에서 `TERM_PROGRAM` 문자열을 검색해도 일치하는 코드가 없습니다. 이는 Zed, iTerm2, VS Code 등 대부분의 터미널이 관행적으로 심는 표준 식별 변수를 Alacritty만 심지 않는다는 뜻이며, 그 결과 "가장 신뢰도 높은 마커가 없다"는 것 자체가 이 터미널의 핵심 특징입니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `ALACRITTY_WINDOW_ID` | 정수 문자열 (예: `1`) | 실행 식별 | 창을 스폰한 자식 셸에 창 ID를 전달, `alacritty msg config --window-id`/`get-config`의 기본값으로도 사용 | 보조 — 이름이 Alacritty 전용이라 존재 자체는 강한 신호지만, Unix 전용이고 v0.11.0(2022-08-31) 이전 버전과 Windows 빌드에는 없음 | [공식 소스: `unix.rs` `from_fd`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty_terminal/src/tty/unix.rs#L229-L230), [`cli.rs`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty/src/cli.rs#L336) |
| `ALACRITTY_SOCKET` | 소켓 파일 경로 | 실행 식별 | `alacritty msg`가 실행 중인 창의 IPC 유닉스 소켓을 찾도록 안내 | 보조 — Unix 전용이며 `general.ipc_socket`(기본 `true`) 설정이나 소켓 바인딩 실패 시 존재하지 않을 수 있어 단독 마커로 부적합 | [공식 소스: `polling/ipc.rs`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty/src/polling/ipc.rs#L21-L44), [`main.rs`의 `#[cfg(unix)] mod polling`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty/src/main.rs#L44) |
| `ALACRITTY_LOG` | 로그 파일 경로 (예: 임시 디렉터리의 `Alacritty-<pid>.log`) | 실행 식별 | 실행 중인 인스턴스의 로그 파일 위치 안내 | 적합 — 이름이 Alacritty 전용이며 v0.2.8부터 플랫폼(Windows 포함)·설정과 무관하게 로거 초기화 시 항상 전역 프로세스 환경에 설정됨 | [공식 소스: `logging.rs` `OnDemandLogFile::new`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty/src/logging.rs#L26-L204), [`main.rs`의 무조건 `logging::initialize` 호출](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty/src/main.rs#L141) |
| `TERM` | 문자열 (`alacritty` 또는 `xterm-256color`) | 설정 | 터미널 기능(terminfo) 광고 | 부적합 — 사용자가 임의로 덮어쓸 수 있고, terminfo가 없으면 다른 터미널과 동일한 `xterm-256color`로 대체되어 식별력이 사라짐 | [공식 소스: `tty/mod.rs` `setup_env`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty_terminal/src/tty/mod.rs#L99-L106) |
| `WINDOWID` | 정수 문자열 | 상태·컨텍스트 | X11 관행에 의존하는 클라이언트를 위한 레거시 창 ID (`$WINDOWID`) | 부적합 — 이름 자체가 X11 관행 표준이라 다른 X11 터미널도 동일하게 설정함 | [공식 소스: `unix.rs` `from_fd`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty_terminal/src/tty/unix.rs#L233-L234) |

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `COLORTERM` | 문자열 (`truecolor`) | 설정 | 24비트 컬러 지원 광고 | 부적합 — 여러 터미널이 동일한 값으로 관행적으로 설정하는 표준 변수 | [공식 소스: `tty/mod.rs` `setup_env`](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/alacritty_terminal/src/tty/mod.rs#L102-L106) |

`[env]` 설정 섹션으로 사용자가 임의 키·값을 자식 프로세스 환경에 추가할 수 있다는 점도 공식 문서에 명시되어 있으나, 이는 사용자 정의 값이라 Alacritty 실행 식별과 무관해 표에서 제외했습니다.

## 상속과 잔존

환경변수는 자식 프로세스 전체에 상속되므로, 위 표의 어떤 "적합"·"보조" 판정도 **부모 Alacritty 프로세스가 지금 살아 있다는 보장이 아닙니다.** 이는 이 문서에서 가장 중요하게 다뤄야 할 지점입니다.

- **SSH**: OpenSSH 클라이언트는 기본적으로 `TERM`만 원격 세션에 전달합니다. 서버의 기본 `AcceptEnv`는 `LANG`/`LC_*` 계열만 허용하고, 클라이언트의 기본 `SendEnv`도 그 계열에 한정되므로 `ALACRITTY_WINDOW_ID`, `ALACRITTY_SOCKET`, `ALACRITTY_LOG`는 원격 셸까지 넘어가지 않습니다. 즉 SSH를 한 번 거치면 표에서 유일하게 남는 신호가 가장 신뢰도 낮은 `TERM`뿐이며, 그마저 원격 로그인 셸 설정에서 재정의될 수 있습니다.
- **tmux / screen**: tmux 서버는 최초로 서버를 띄운 클라이언트의 환경을 캡처해 세션 전역 환경으로 유지합니다. 이후 다른 터미널에서 같은 세션에 붙어도(`attach`) 새 클라이언트의 `ALACRITTY_WINDOW_ID`로 갱신되지 않으므로, 어떤 pane은 지금 화면을 그리고 있지 않은 Alacritty 창을 가리키는 오래된 `ALACRITTY_WINDOW_ID` 값을 계속 노출할 수 있습니다. tmux의 `update-environment` 설정으로 `attach` 시 일부 변수를 갱신할 수 있지만 기본 목록에 `ALACRITTY_*`는 포함되지 않습니다. 더 결정적으로, tmux는 pane 안에서 `TERM`을 자체 값(`screen`, `tmux-256color` 등)으로 덮어써 `TERM=alacritty` 신호 자체를 완전히 지워버립니다. screen도 마찬가지로 자체 `TERM` 값을 강제합니다.
- **디태치·데몬화된 프로세스**: `nohup`, `disown`, 백그라운드 데몬 등으로 부모 Alacritty 창이 이미 닫힌 뒤에도 살아남는 프로세스는 창이 종료되었어도 `ALACRITTY_WINDOW_ID`, `ALACRITTY_SOCKET` 값을 그대로 들고 있습니다. `ALACRITTY_SOCKET`이 가리키는 소켓 파일은 Alacritty 종료 시 정리되지만(`TemporaryFiles::drop`), 환경변수 자체는 그 프로세스가 스스로 종료할 때까지 남아 잘못된 경로를 계속 가리킵니다.
- **위조 가능성**: 이 변수들은 일반 셸 환경변수이므로 사용자나 스크립트가 `export ALACRITTY_WINDOW_ID=1`, `export ALACRITTY_LOG=/tmp/fake.log`처럼 임의로 설정·복제할 수 있습니다. 코드 서명이나 프로세스 트리 검증 같은 신뢰 경계가 전혀 없습니다.

이런 이유로 `runby` 같은 라이브러리가 정직하게 주장할 수 있는 것은 "이 프로세스의 환경에 Alacritty가 심는 이름의 변수가 남아 있다"는 관측뿐이며, "지금 이 명령을 실행 중인 Alacritty 창이 존재한다"는 보장은 아닙니다. SSH·tmux·screen을 하나라도 거쳤다면 이 관측의 신뢰도는 급격히 떨어지고, 특히 `TERM` 단독 신호는 tmux/screen을 거치는 순간 완전히 사라집니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM`이 없는 상태에서 가장 신뢰도 높은 존재 마커는 `ALACRITTY_LOG`입니다. v0.2.8부터 플랫폼(Windows 포함)과 설정에 무관하게 로거 초기화 시 항상 전역 프로세스 환경에 설정되기 때문입니다. `ALACRITTY_WINDOW_ID`는 이름이 더 명확하게 Alacritty 전용이지만 Unix 전용이고 v0.11.0(2022) 이전에는 존재하지 않으므로 버전에 따라 결과가 갈립니다. `ALACRITTY_SOCKET`은 `general.ipc_socket` 설정과 플랫폼(Unix 전용)에 따라 조건부로만 존재해 단독 마커로 쓸 수 없습니다. `TERM=alacritty`는 사용자가 덮어쓸 수 있고 terminfo 부재 시 `xterm-256color`로 대체되어 식별력을 잃으며, tmux/screen을 거치면 값 자체가 사라지므로 보조 신호 이상으로 취급해서는 안 됩니다.

결론적으로 `runby`는 이 변수들 중 하나 이상의 **존재**를 "Alacritty가 이 프로세스 트리 어딘가에 있었다"는 낮은 확신의 신호로만 보고해야 하며, SSH·tmux·screen을 거친 경로에서는 이 신호가 약화되거나 완전히 사라질 수 있다는 점, 그리고 어떤 조합도 위조를 막지 못한다는 점을 함께 명시해야 합니다.

## 공식 문서

- [Alacritty 공식 GitHub 저장소](https://github.com/alacritty/alacritty)
- [Alacritty 웹사이트 — 설정 문서](https://alacritty.org/config-alacritty.html)
- [`alacritty(1)` man 페이지 (scdoc 원본)](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/extra/man/alacritty.1.scd)
- [`alacritty(5)` 설정 man 페이지 (scdoc 원본)](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/extra/man/alacritty.5.scd)
- [`alacritty-msg(1)` man 페이지 (scdoc 원본)](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/extra/man/alacritty-msg.1.scd)
- [CHANGELOG.md](https://github.com/alacritty/alacritty/blob/ede2ac144da4dec4c075bfa803aacf3b3739bce6/CHANGELOG.md)
