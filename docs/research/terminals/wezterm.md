---
title: WezTerm
slug: wezterm
research_date: 2026-08-31
open_source: true
repository: https://github.com/wezterm/wezterm
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 소스에서 주입 지점(`config/src/config.rs`, `mux/src/domain.rs`, `env-bootstrap/src/lib.rs`)을 모두 확인했지만, 실제 WezTerm 실행 환경에서 SSH 도메인·tmux 내부·중첩 자식 프로세스로 값이 상속되는 구체적인 동작은 직접 검증하지 않음. 특히 `WEZTERM_PANE`이 중첩 셸에서 갱신되는지, `sudo`/`su` 이후에도 유지되는지는 실행 시험이 필요함
---

# WezTerm

WezTerm은 자신이 직접 실행한 프로세스의 환경에 `TERM`, `COLORTERM`, `TERM_PROGRAM`, `TERM_PROGRAM_VERSION`을 매번 명시적으로 설정합니다. 이는 공식 소스 [`config/src/config.rs`의 `apply_cmd_defaults`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1617-L1622)에서 모든 신규 프로세스 스폰 경로가 공유하는 로직으로 확인됩니다. 여기에 더해 로컬 멀티플렉서가 붙는 pane에는 `WEZTERM_PANE`이, 실행 파일 위치에는 `WEZTERM_EXECUTABLE`/`WEZTERM_EXECUTABLE_DIR`이 주입됩니다. 다만 이 변수들은 안정적으로 문서화된 **공개 API 계약**이 아니라 구현에 근거한 신호이며, 다른 다수의 터미널도 `TERM_PROGRAM`이라는 동일한 이름의 관례를 따르므로 값 자체(`WezTerm`)가 식별의 핵심입니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERM_PROGRAM` | 문자열 (`WezTerm`, 대소문자 고정) | 실행 식별 | 터미널 호스트 프로그램 식별을 위한 업계 관례 값 | 적합 — WezTerm이 스폰하는 모든 자식 프로세스에 매번 `WezTerm`을 재설정함. 값의 대소문자가 고정되어 있어 다른 터미널과 구분됨 | [공식 소스: `apply_cmd_defaults`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1617-L1622) |
| `WEZTERM_PANE` | 정수 문자열 (프로세스 인스턴스 내에서 0부터 증가하는 pane ID) | 실행 식별 | 현재 pane을 `wezterm cli --pane-id` 기본값으로 지정 | 보조 신호 — WezTerm 전용 이름이라 존재 자체는 강한 신호지만, 값은 mux 인스턴스 생명주기에 종속된 카운터일 뿐 전역적으로 유일하거나 영구적인 세션 ID가 아님(아래 상속과 잔존 참고) | [공식 소스: `build_command`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/mux/src/domain.rs#L479-L482), [`PaneId` 정의](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/mux/src/pane.rs#L24-L25), [공식 CLI 문서: Targeting Panes](https://wezterm.org/cli/cli/index.html) |
| `WEZTERM_EXECUTABLE` | 절대 경로 문자열 | 상태·컨텍스트 | 현재 WezTerm 실행 파일의 경로 | 보조 신호 — 실제 wezterm 프로세스만 부트스트랩 시점에 설정하지만, 경로 문자열은 사용자가 셸에서 동일한 이름으로 재정의할 수 있음 | [공식 소스: `set_wezterm_executable`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/env-bootstrap/src/lib.rs#L5-L12) |
| `WEZTERM_EXECUTABLE_DIR` | 디렉터리 경로 | 상태·컨텍스트 | 실행 파일이 위치한 디렉터리 (AppImage 등에서 동봉 도구 PATH 구성에도 사용) | 보조 신호 — `WEZTERM_EXECUTABLE`과 동일한 한계를 가짐 | [공식 소스: `set_wezterm_executable`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/env-bootstrap/src/lib.rs#L5-L12) |
| `TERM_PROGRAM_VERSION` | 버전 문자열 (`wezterm_version()` 반환값) | 상태·컨텍스트 | WezTerm 버전 정보 제공 | 보조 신호 — `TERM_PROGRAM=WezTerm`과 함께 있을 때만 버전 컨텍스트로 의미가 있고, 단독으로는 식별력이 없음 | [공식 소스: `apply_cmd_defaults`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1617-L1622) |
| `TERM` | 문자열 (기본값 `xterm-256color`, `term` 설정으로 변경 가능) | 설정 | 터미널 기능 유형(terminfo) 선택 | 부적합 — 다수의 터미널이 같은 값을 기본값으로 사용하는 매우 일반적인 표준 변수 | [공식 소스: `default_term`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1751-L1753), [`apply_cmd_defaults`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1617), [공식 문서: `term`](https://wezterm.org/config/lua/config/term.html) |

가장 신뢰도 높은 단일 마커는 `TERM_PROGRAM=WezTerm`입니다(대소문자를 그대로 `WezTerm`으로 비교해야 합니다). 이 값은 WezTerm이 스폰하는 모든 프로세스에 대해 소스 코드 상 예외 없이 재설정되며, `TERM_PROGRAM_VERSION`으로 버전까지 교차 확인할 수 있습니다.

`WEZTERM_PANE`은 "이 pane을 대상으로 삼아라"라는 CLI 기본 타겟팅 용도로 설계된 값이며, 공식 CLI 문서는 이를 안정적인 세션 식별자가 아니라 `--pane-id` 인자의 대체값으로만 설명합니다. 소스 상 `PaneId`는 `AtomicUsize`로 0부터 증가하는 **프로세스(mux 인스턴스) 로컬 카운터**이므로, WezTerm 프로세스가 재시작되면 다시 0부터 시작할 수 있고 전역적으로 유일하지 않습니다. 또한 pane이 닫힌 뒤에도 그 pane에서 파생되어 계속 살아있는 자식 프로세스의 환경에는 예전 pane ID가 그대로 남아, 더 이상 존재하지 않는 pane을 가리킬 수 있습니다.

`TERM`의 기본값은 `xterm-256color`이며, 공식 문서는 이를 "추가 terminfo 설치 없이도 좋은 수준의 기능 지원을 제공하기 위한" 선택이라고 명시합니다. 즉 WezTerm 전용 `wezterm` terminfo가 설치되어 있는지 여부와 무관하게 기본값은 `xterm-256color`로 고정되며, 사용자가 `config.term`을 명시적으로 바꾸지 않는 한 `wezterm`이라는 값은 나타나지 않습니다. 이 값만으로는 WezTerm을 다른 xterm 계열 터미널과 구분할 수 없습니다.

## 상태·컨텍스트 변수

`runby`가 안전하게 노출할 만한 범위로 한정하면 다음 두 변수가 있습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `WEZTERM_UNIX_SOCKET` | 유닉스 소켓 경로 문자열 | 상태·컨텍스트 | 로컬 mux 서버와 통신할 소켓 경로 | 보조 신호 — mux 도메인을 사용할 때만 존재하며, 이미 환경에 있을 때만 자식으로 전달됨(아래 `WEZTERM_UNIX_SOCKET`이 없는 최초 GUI 프로세스에서는 mux 서버가 직접 설정) | [공식 소스: `build_command`의 조건부 전달](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/mux/src/domain.rs#L479-L480), [mux 서버의 설정](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/wezterm-mux-server/src/main.rs#L313) |
| `WEZTERM_CONFIG_FILE` | 절대 파일 경로 | 설정 | 현재 로드된 `wezterm.lua` 설정 파일 경로 | 부적합 — 설정 위치일 뿐 실행 주체 식별과 무관하며, lua 설정 파일을 찾지 못하면 환경에서 제거됨 | [공식 소스: 설정 시 주입](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1152-L1155), [설정 실패 시 제거](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/config/src/config.rs#L1079-L1081) |

`WEZTERM_CONFIG_FILE`은 개별 자식 프로세스마다 `cmd.env()`로 명시적으로 주입되는 값이 아니라, WezTerm 프로세스 자체의 환경(`std::env::set_var`)에 설정된 뒤 일반적인 유닉스 fork/exec 상속 경로로 자식에게 전달됩니다. 따라서 WezTerm GUI를 다른 wezterm 실행 파일이 감싸고 있거나 config 파일을 찾지 못한 경우에는 존재하지 않을 수 있습니다. `WEZTERM_EXECUTABLE_DIR`, `WEZTERM_UNIX_SOCKET`과 함께 이 값들은 "실행 주체가 누구인가"보다는 "WezTerm이 어떤 상태로 구성되었는가"를 설명하는 보조 정보로만 취급해야 합니다.

## 상속과 잔존

환경변수는 자식 프로세스 전체에 상속되므로, WezTerm이 어떤 프로세스를 직접 만들었다는 사실과 "현재 이 변수가 보이는 프로세스가 지금 WezTerm 안에서 실행 중이다"라는 결론은 다릅니다. 다음 경로들에서 값이 잔존하거나 훼손될 수 있습니다.

- **SSH (`ssh` 자체)** — OpenSSH 클라이언트는 기본 `SendEnv`/서버의 기본 `AcceptEnv` 설정에서 임의의 환경변수를 원격으로 전달하지 않습니다(기본적으로 `LANG`/`LC_*` 계열 정도만 통과). 따라서 일반적인 `ssh host` 접속에서는 로컬의 `TERM_PROGRAM=WezTerm`, `WEZTERM_PANE` 등이 원격 셸로 넘어가지 않는 것이 기본 동작입니다. WezTerm 소스에서도 `wezterm ssh`(내장 ad-hoc SSH 클라이언트) 경로에서 원격에 `WEZTERM_REMOTE_PANE`을 전달하려고 시도하지만, 주석에 "대부분의 sshd 설정은 임의 환경변수 설정을 금지하므로 보통 설정되지 않는다"고 명시되어 있습니다([`mux/src/ssh.rs`](https://github.com/wezterm/wezterm/blob/08e5e0afc6007567d3d131bc2ec1382423c0d2da/mux/src/ssh.rs#L266-L270)). 즉 SSH를 건너면 이 변수들은 원칙적으로 사라지거나, 서버 설정에 따라 예외적으로만 남습니다.
- **WezTerm 자체 멀티플렉서 (SSH 도메인/`wezterm connect`)** — WezTerm은 `ssh` 프로토콜을 채널로만 사용해 **원격 호스트에 설치된 별도의 wezterm 데몬**에 접속하는 SSH 도메인을 지원합니다([공식 문서: Multiplexing — SSH Domains](https://wezterm.org/multiplexing.html)). 이 경우 원격 wezterm-mux-server가 그 자신의 `apply_cmd_defaults`를 다시 실행하므로, 원격 pane에서 관찰되는 `TERM_PROGRAM=WezTerm`, `WEZTERM_PANE`은 로컬 GUI가 아니라 **원격 호스트의 wezterm 인스턴스가 새로 설정한 값**입니다. 값 자체는 진짜지만 "어느 머신의 WezTerm인지"는 알려주지 않으므로, 로컬 GUI 실행과 원격 mux 세션을 이 변수만으로 구분할 수 없습니다.
- **tmux / screen** — tmux 서버는 그 서버를 최초로 띄운 클라이언트의 환경을 기준으로 세션을 만들고, 이후 다른 터미널에서 붙은(attach) 클라이언트에도 그 환경이 그대로 노출될 수 있습니다. 즉 WezTerm에서 시작한 tmux 세션에 나중에 다른 터미널이 attach해도 pane에는 여전히 옛 `TERM_PROGRAM=WezTerm`, `WEZTERM_PANE=<옛 값>`이 남아 있을 수 있습니다. tmux의 `update-environment` 설정은 `attach` 시점에 지정된 변수 목록만 클라이언트 환경으로 갱신하며 기본 목록에 `TERM_PROGRAM`이나 `WEZTERM_PANE`은 포함되어 있지 않으므로, 별도 설정이 없다면 이 값들은 최초 attach 시점의 스냅샷으로 고정된 채 남습니다.
- **분리된(daemonized)/터미널을 벗어난 프로세스** — `nohup`, `disown`, `setsid`로 백그라운드에 남긴 프로세스나 시스템 서비스로 등록된 프로세스는 WezTerm pane과 터미널 자체가 이미 닫힌 뒤에도 계속 실행되며, 그 시점에 상속받은 `TERM_PROGRAM`/`WEZTERM_PANE` 값을 영구히 들고 있을 수 있습니다. 이때 `WEZTERM_PANE`이 가리키는 pane ID는 더 이상 존재하지 않거나(닫혔거나), 같은 mux 인스턴스에서 다른 pane이 재사용했을 수도 있는 값입니다.
- **위조 가능성** — 이 모든 변수는 이름과 값이 문서화되어 있어 누구나 `export TERM_PROGRAM=WezTerm`, `export WEZTERM_PANE=1`처럼 임의로 설정할 수 있습니다. 환경변수는 신뢰 경계가 아니므로 권한 판단의 근거로 쓸 수 없습니다.

이런 이유로 `runby`가 이 변수들로 정직하게 주장할 수 있는 것은 "이 프로세스의 환경에는 WezTerm(또는 WezTerm SSH 도메인의 원격 호스트)이 한 번은 설정했던 값이 남아 있다"는 것뿐이며, "지금 이 순간 눈앞의 WezTerm GUI 창 안에서 실행 중이다"라는 보장은 아닙니다. SSH·tmux·백그라운드화를 거치지 않은 직계 자식 프로세스에서는 신뢰도가 높지만, 그 경계를 하나라도 넘으면 정확도가 급격히 떨어집니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM=WezTerm`이 가장 신뢰도 높은 단일 마커이며, `TERM_PROGRAM_VERSION`으로 버전을, `WEZTERM_EXECUTABLE`/`WEZTERM_EXECUTABLE_DIR`로 실행 파일 위치를 보강할 수 있습니다. `WEZTERM_PANE`은 WezTerm 전용 이름이라는 점에서 존재 자체는 유용한 보조 신호지만, 값은 mux 인스턴스에 로컬한 재사용 가능한 카운터일 뿐이므로 안정적인 pane·세션 식별자로 취급해서는 안 됩니다. `TERM`은 다른 다수의 터미널과 값이 겹치므로 단독 식별에는 부적합합니다.

WezTerm은 agent를 직접 실행하거나 호스팅하는 제품이 아니라 터미널 에뮬레이터이므로, 이 문서의 신호는 `Result.IsAgent()`가 아니라 "어떤 터미널 애플리케이션이 이 프로세스를 소유했는가"라는 별도의 판정에만 쓰여야 합니다. SSH 도메인을 통한 원격 접속에서는 이 신호가 로컬 GUI가 아니라 원격 호스트의 WezTerm 인스턴스를 가리킬 수 있다는 점, 그리고 SSH·tmux·백그라운드화를 거친 프로세스에서는 값이 상속된 스냅샷일 뿐 실시간 상태가 아니라는 점을 함께 보고해야 합니다.

## 공식 문서

- [WezTerm 공식 문서](https://wezterm.org/)
- [Shell Integration](https://wezterm.org/shell-integration.html)
- [`term` 설정](https://wezterm.org/config/lua/config/term.html)
- [`wezterm cli`](https://wezterm.org/cli/cli/index.html)
- [Multiplexing](https://wezterm.org/multiplexing.html)
- [SSH](https://wezterm.org/ssh.html)
- [`pane:pane_id()`](https://wezterm.org/config/lua/pane/pane_id.html)
- [F.A.Q.](https://wezterm.org/faq.html)
- [공식 WezTerm 소스 저장소 (`wezterm/wezterm`)](https://github.com/wezterm/wezterm)
