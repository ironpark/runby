---
title: Ghostty
slug: ghostty
research_date: 2026-08-31
open_source: true
repository: https://github.com/ghostty-org/ghostty
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 소스에서 변수 주입 코드는 확인했지만, TERM_PROGRAM_VERSION의 실제 런타임 포맷, 셸별 자동 감지(zsh/bash/fish/elvish/nushell) 시 GHOSTTY_SHELL_FEATURES 실값, macOS 앱 번들과 Linux 패키지 간 GHOSTTY_RESOURCES_DIR 경로 차이, 그리고 tmux/SSH 중첩 상황에서의 잔존 동작은 아직 실제로 Ghostty를 실행해 검증하지 않음
---

# Ghostty

Ghostty의 공식 소스([`ghostty-org/ghostty`](https://github.com/ghostty-org/ghostty))는 자식 프로세스를 생성할 때 `TERM_PROGRAM`, `TERM_PROGRAM_VERSION`, `TERM`, `GHOSTTY_RESOURCES_DIR`, `GHOSTTY_BIN_DIR`를 코드에서 직접 `env.put`으로 주입합니다. 공식 문서는 이 중 `GHOSTTY_RESOURCES_DIR`만 "셸이 Ghostty 안에서 실행 중인지 감지하는 용도"로 명시적으로 설명하며, 나머지는 소스 코드의 주석과 구현으로만 확인됩니다. Ghostty는 어떤 변수도 세션·창·분할창(pane) 단위로 구분되는 고유 식별자로 주지 않으며, 셸 통합(shell integration) 관련 변수는 통합이 실제로 활성화되었을 때만 존재합니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERM_PROGRAM` | 문자열, 정확히 소문자 `ghostty` | 실행 식별 | neovim 등 애플리케이션이 터미널 종류를 감지하도록 제공 | 적합 — 모든 Ghostty 자식 프로세스에 무조건 주입되는 값이며 대소문자까지 고정된 리터럴 | [공식 소스: `Subprocess.init`](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/termio/Exec.zig#L753-L755) |
| `TERM_PROGRAM_VERSION` | 버전 문자열 (`build_config.version_string`) | 상태·컨텍스트 | Ghostty 버전 제공 | 보조 신호 — `TERM_PROGRAM=ghostty`와 함께 버전 컨텍스트를 주지만 단독으로는 다른 터미널과 구분되지 않음 | [공식 소스: `Subprocess.init`](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/termio/Exec.zig#L753-L755) |
| `TERM` | 문자열. 기본값 `xterm-ghostty`, 리소스 디렉터리를 찾지 못하면 `xterm-256color`로 대체 | 설정 | 터미널 기능(terminfo) 식별 | 부적합 — 값이 조건에 따라 바뀌고 사용자가 `term` 설정으로 임의 값을 지정할 수도 있어 단독 판정에는 쓰지 말 것 | [`term` 설정 기본값](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/config/Config.zig#L3850), [공식 소스: 리소스 디렉터리 유무에 따른 분기](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/termio/Exec.zig#L644-L667) |

가장 신뢰할 수 있는 단일 마커는 **`TERM_PROGRAM=ghostty`(전부 소문자)** 입니다. 대소문자가 다른 값(`Ghostty`, `GHOSTTY` 등)을 공식 소스에서 사용하는 경로는 확인되지 않았으므로, exact-match 비교 시 소문자 리터럴을 그대로 사용해야 합니다.

Ghostty는 창·탭·분할창(split/pane)을 구분하는 세션 식별자를 환경변수로 노출하지 않습니다. 공식 소스의 `Subprocess.init`에서 자식 프로세스에 주입하는 변수 목록 전체를 확인했지만 `WINDOW_ID`, `PANE_ID`, `SURFACE_ID` 류의 값은 없었습니다. 즉 `runby`는 "이 프로세스가 Ghostty에서 실행되었다"는 사실은 판정할 수 있지만, kitty의 `KITTY_WINDOW_ID`나 WezTerm의 `WEZTERM_PANE`처럼 **어느 창·분할창인지는 구분할 수 없습니다.**

## 상태·컨텍스트 변수

이 변수들은 `runby`가 안전하게 노출할 수 있는 수준의 보조 컨텍스트입니다. 셸 통합이 켜져 있을 때만 존재하는 값은 그렇게 표시했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `GHOSTTY_RESOURCES_DIR` | 디렉터리 경로 | 상태·컨텍스트 | 셸 통합 스크립트·terminfo 위치 제공, 문서가 명시하는 "Ghostty 안에서 실행 중"임을 감지하는 용도 | 보조 신호 — 리소스 디렉터리를 찾은 경우에만 무조건 주입되지만, 사용자가 셸에서 같은 이름으로 값을 재설정할 수 있어 단독 신뢰 근거로는 약함 | [공식 문서: Shell Integration](https://ghostty.org/docs/features/shell-integration), [공식 소스: `Subprocess.init`](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/termio/Exec.zig#L636-L638) |
| `GHOSTTY_BIN_DIR` | 디렉터리 경로 | 상태·컨텍스트 | 셸 초기화 스크립트가 PATH를 재설정해도 `ghostty` 실행 파일을 찾을 수 있도록 경로 제공 | 보조 신호 — 실행 파일 경로를 찾을 수 있으면 거의 항상 주입되지만(Flatpak 제외) 사용자가 재정의할 수 있음 | [공식 소스: `Subprocess.init`](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/termio/Exec.zig#L683-L692) |
| `GHOSTTY_SHELL_FEATURES` | 쉼표 구분 기능명 문자열 (예: `cursor:blink,path,sudo,title`) | 상태·컨텍스트 | 현재 활성화된 셸 통합 기능 목록 | 보조 신호 — **셸 통합이 실제로 로드되었을 때만** 존재하며, 통합이 꺼져 있거나(`shell-integration = none`) 감지에 실패하면 아예 설정되지 않으므로 감지용으로 의존해서는 안 됨 | [공식 소스: 기능 목록 직렬화](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/termio/shell_integration.zig#L204-L236), [공식 소스: 기능 플래그 정의](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/config/Config.zig#L8782-L8789) |

`GHOSTTY_SHELL_INTEGRATION_NO_*` 형태의 변수는 공식 문서와 공식 소스 어디에서도 확인되지 않았습니다. 실제로 존재하는 것은 `shell-integration-features` 설정에 `no-cursor`, `no-sudo`, `no-title` 같은 값을 넣는 방식이며, 이는 Ghostty 설정 파일 문법이지 자식 프로세스에 주입되는 환경변수가 아닙니다. 따라서 이 변수명을 감지 로직에 사용해서는 안 됩니다 — 공식 근거: [`shell-integration-features` 설정 문서](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/config/Config.zig#L2933-L2977).

`COLORTERM=truecolor`, `TERMINFO`, `XDG_DATA_DIRS`/`MANPATH`(macOS 전용) 등은 렌더링·리소스 탐색 설정일 뿐 실행 주체 식별과 무관해 표에서 제외했습니다.

## 상속과 잔존

환경변수는 모든 자식 프로세스에 상속되므로, 위 변수들이 존재한다는 사실은 "지금 이 프로세스가 떠 있는 동안 Ghostty가 어딘가에서 이 값을 최초로 설정했다"는 것만 증명하며 "지금 이 프로세스를 화면에 보여주는 터미널이 Ghostty"라는 것을 보장하지 않습니다.

- **SSH**: `shell-integration-features`의 `ssh-env`와 `ssh-terminfo`는 기본값이 `false`인 옵트인 기능입니다(공식 소스: [`ShellIntegrationFeatures` 기본값](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/config/Config.zig#L8782-L8789)). `ssh-env`를 켜면 로컬 셸 통합이 SSH 연결 시 `COLORTERM`, `TERM_PROGRAM`, `TERM_PROGRAM_VERSION`을 원격 세션에 전파하고, `TERM`을 `xterm-ghostty`에서 `xterm-256color`로 변환합니다(공식 소스 주석: [`ssh-env`/`ssh-terminfo` 설명](https://github.com/ghostty-org/ghostty/blob/ec3e384d2d7d86206fc3c71aa23a76b7bdb5eae9/src/config/Config.zig#L2953-L2977)). 원격 `sshd_config`의 `AcceptEnv`가 이 변수들을 허용해야 실제로 전달됩니다. 이는 로컬 터미널의 정체성을 원격 호스트로 **의도적으로 전파**하는 기능이므로, 이 변수들을 "로컬에서 Ghostty가 직접 이 프로세스를 실행했다"는 증거로 취급하면 원격 세션에서 오탐이 발생합니다. `ssh-terminfo`가 성공하면 원격에서도 `TERM=xterm-ghostty`를 유지하고, `tic` 설치에 실패하면 `TERM=xterm-256color`로 대체됩니다. 참고로 이 두 기능이 꺼져 있는 기본 설정에서 `xterm-ghostty` terminfo가 없는 원격 호스트로 평범하게 SSH 접속하면 애플리케이션에 따라 "unknown terminal type" 오류가 날 수 있습니다.
- **tmux / screen**: tmux 서버는 서버를 최초로 띄운 클라이언트의 환경을 복제해 보관하며, 이후 다른 터미널에서 붙은 클라이언트에는 `update-environment` 설정에 나열된 변수만 갱신됩니다. 즉 Ghostty에서 tmux 서버를 처음 띄우고 이후 다른 터미널 앱에서 그 세션에 붙으면, pane 안의 `TERM_PROGRAM`이 `ghostty`로 남아있더라도 화면에 표시 중인 터미널은 Ghostty가 아닐 수 있습니다. 반대의 경우도 마찬가지입니다. `TERM_PROGRAM`은 tmux의 `update-environment` 기본 목록에 포함되지 않는 경우가 많아 특히 오래된 값이 남기 쉽습니다.
- **데몬화·분리된 프로세스**: `nohup`, `setsid`, 백그라운드 서비스 매니저 등으로 터미널과 분리된 프로세스는 시작 당시 물려받은 `TERM_PROGRAM=ghostty`를 터미널 창이 닫힌 뒤에도 계속 들고 있을 수 있습니다. 이런 프로세스에서는 변수가 "현재 Ghostty가 살아있고 이 프로세스를 소유한다"는 뜻이 아니라 "과거 어느 시점에 Ghostty 자식으로 시작됐다"는 뜻일 뿐입니다.
- **위조 가능성**: 이 변수들은 모두 일반 환경변수이므로 어떤 프로세스든 자유롭게 설정·복사·삭제할 수 있습니다. 신뢰 경계나 서명된 토큰이 아니라 관례적 신호입니다.

따라서 `runby`가 정직하게 주장할 수 있는 것은 "이 프로세스의 환경에 `TERM_PROGRAM=ghostty`가 존재한다"까지이며, "지금 이 순간 Ghostty 창이 이 프로세스의 출력을 표시하고 있다"는 사실이나 "사람이 아니라 Ghostty 자체가 이 프로세스를 방금 실행했다"는 사실까지는 보장하지 못합니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM=ghostty`(정확히 소문자)는 공식 소스에서 모든 Ghostty 자식 프로세스에 무조건 주입되는 가장 신뢰도 높은 단일 신호입니다. `TERM_PROGRAM_VERSION`은 버전 정보를 더할 수 있는 보조 신호이며, `TERM=xterm-ghostty`는 사용자가 `term` 설정을 바꾸거나 리소스 디렉터리를 찾지 못하면 값이 달라지므로 단독 판정에는 부적합합니다. `GHOSTTY_RESOURCES_DIR`과 `GHOSTTY_BIN_DIR`은 존재 여부만으로 보강 신호가 될 수 있지만 사용자가 재정의할 수 있어 `TERM_PROGRAM`보다 신뢰도가 낮습니다. `GHOSTTY_SHELL_FEATURES`는 셸 통합이 활성화된 경우에만 존재하므로 감지 실패의 근거로 쓸 수 없습니다.

Ghostty는 창·탭·분할창 단위의 세션 식별자를 전혀 제공하지 않으므로, `runby`는 "Ghostty에서 실행되었다"는 것까지만 보고할 수 있고 어느 창인지는 구분할 수 없습니다. SSH의 `ssh-env` 기능은 원격 호스트에도 같은 마커를 전파하도록 설계돼 있고, tmux/screen은 서버를 띄운 클라이언트의 환경을 오래 유지하므로, 두 경우 모두 값이 있다고 해서 지금 이 프로세스를 실제로 소유한 터미널이 Ghostty라고 확정할 수 없습니다. `runby`는 이를 상속 가능하고 위조 가능한 관례적 신호로 취급해야 합니다.

## 공식 문서

- [Shell Integration — Features](https://ghostty.org/docs/features/shell-integration)
- [SSH — Features](https://ghostty.org/docs/features/ssh)
- [Terminfo — Help](https://ghostty.org/docs/help/terminfo)
- [Option Reference — Configuration](https://ghostty.org/docs/config/reference)
- [공식 Ghostty 소스 저장소 (`ghostty-org/ghostty`)](https://github.com/ghostty-org/ghostty)
