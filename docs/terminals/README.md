# 터미널 에뮬레이터 문서

터미널 에뮬레이터별 환경변수 식별 규칙을 정리하는 문서입니다. 양식과 front matter 필드의 의미는 [상위 문서](../README.md)와 같으며, `product_type`은 모두 `terminal_emulator`입니다.

## 이 축은 의도적으로 약한 신호입니다

에이전트·CI 감지와 달리 터미널 식별은 **구조적으로** 현재 상태를 보장할 수 없습니다. 환경변수는 모든 자손 프로세스가 상속하므로, 이 값들이 가리키는 것은 언제나 *이 환경을 만든 터미널*이지 *지금 이 프로세스의 TTY에 붙어 있는 터미널*이 아닙니다. 표시 방식 결정에는 써도 되지만 신뢰 경계로는 쓸 수 없습니다.

대응 관계를 깨뜨리는 네 가지가 모든 문서에서 확인됐습니다.

1. **멀티플렉서 잔존** — tmux·screen 서버는 *처음 붙은 클라이언트*의 환경을 유지합니다. tmux 기본 `update-environment`(`DISPLAY KRB5CCNAME SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WINDOWID XAUTHORITY`)에 터미널 식별 변수가 하나도 없어, 이미 실행 중인 pane의 값은 재접속해도 갱신되지 않습니다. 이미 닫힌 터미널을 계속 보고할 수 있습니다.
2. **SSH 정체성 전파** — 아래 별도 절 참고.
3. **데몬화** — 터미널보다 오래 사는 프로세스는 낡은 스냅샷을 계속 들고 있습니다.
4. **위조** — 누구나 `export TERM_PROGRAM=...` 할 수 있습니다.

`runby`는 이 중 멀티플렉서만 감지할 수 있으므로, `TMUX`나 `STY`가 있으면 `Terminal.Multiplexer`를 채우고 `Confidence`를 `probable`로 낮춥니다.

## SSH가 더 이상 방벽이 아닙니다

OpenSSH 기본값은 `TERM`(pty 요청으로 항상 전송)과 `LANG`/`LC_*` 외에는 전달하지 않습니다. 그런데 최신 터미널들이 이 전제를 의도적으로 깨고 있습니다.

| 터미널 | 전파 수단 | 기본값 |
|---|---|---|
| iTerm2 | `LC_TERMINAL`·`LC_TERMINAL_VERSION` — `LC_*` 네임스페이스라 `SendEnv LC_*`로 자동 전송 | **켜짐** |
| kitty | `kitten ssh` — terminfo와 셸 통합을 원격에 부트스트랩 | 옵트인 |
| Ghostty | `ssh-env` 셸 통합 — `TERM_PROGRAM`까지 원격에 전달 | 꺼짐 |

결과적으로 **원격 호스트의 프로세스가 그 머신에 존재하지도 않는 터미널을 보고**할 수 있습니다. `runby`가 `LC_TERMINAL`을 식별에 전혀 쓰지 않는 이유입니다.

## 감지 신호 요약

| 터미널 | 마커 | 세션 식별자 | 비고 |
|---|---|---|---|
| [iTerm2](iterm2.md) | `TERM_PROGRAM=iTerm.app` | `ITERM_SESSION_ID` (`w<창>t<탭>p<pane>:<GUID>`) | `LC_TERMINAL`은 SSH 전파되어 사용 금지 |
| [Apple Terminal](apple-terminal.md) | `TERM_PROGRAM=Apple_Terminal` | `TERM_SESSION_ID` | Apple 공식 문서에 변수 언급 자체가 없음 |
| [WezTerm](wezterm.md) | `TERM_PROGRAM=WezTerm` | `WEZTERM_PANE` (재사용되는 카운터) | `TERM` 기본값은 `xterm-256color` |
| [Ghostty](ghostty.md) | `TERM_PROGRAM=ghostty` | **없음** | `TERM`은 `xterm-ghostty`, 리소스 없으면 폴백 |
| [Warp](warp.md) | `TERM_PROGRAM=WarpTerminal` | 없음 | 에이전트/사람 구분 변수 없음 |
| [Zed](../agents/zed-agent.md) | `ZED_TERM=true` + `TERM_PROGRAM=zed` | 없음 | Agent 전용 신호 없음 |
| [kitty](kitty.md) | `KITTY_WINDOW_ID` | `KITTY_WINDOW_ID` | `TERM_PROGRAM` 미설정. `KITTY_PID`로 생존 확인 가능 |
| [Windows Terminal](windows-terminal.md) | `WT_SESSION` | `WT_SESSION` (GUID) | `TERM_PROGRAM` 미설정. WSL 경계를 넘음 |
| [Alacritty](alacritty.md) | `ALACRITTY_LOG` | `ALACRITTY_WINDOW_ID` (v0.11+, Unix 전용) | `TERM_PROGRAM` 미설정 |
| [Konsole](konsole.md) | `KONSOLE_VERSION` 또는 `KONSOLE_DBUS_SESSION` | `KONSOLE_DBUS_SESSION` | `TERM_PROGRAM` 미설정. 임베디드와 구분 불가라 `probable` 고정 |
| [GNOME Terminal](gnome-terminal.md) | `GNOME_TERMINAL_SCREEN` | `GNOME_TERMINAL_SCREEN` | `TERM_PROGRAM` 미설정 |
| VTE 계열 | `VTE_VERSION` | 없음 | 제품이 아니라 **계열**만 식별 |

### `TERM_PROGRAM`은 절반에서만 쓸 수 있습니다

kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal은 `TERM_PROGRAM`을 설정하지 않습니다. kitty는 작성자가 이슈 #957에서 의도적 결정임을 밝혔고 `KITTY_WINDOW_ID`를 대신 쓰라고 권합니다. 따라서 마커 전략은 터미널마다 다릅니다.

### `TERM`은 마커가 아닙니다

`TERM`은 정체성이 아니라 terminfo 능력을 나타내고, 사용자가 덮어쓰며, 멀티플렉서가 `tmux-256color`/`screen`으로 교체합니다. Alacritty는 terminfo가 없으면 `xterm-256color`로, Ghostty도 리소스를 못 찾으면 같은 값으로 폴백합니다. `runby`는 `TERM`을 **판정이 끝난 뒤 컨텍스트로만** 기록하고 판정에는 쓰지 않습니다.

### VTE 계열 처리

`VTE_VERSION`은 GNOME Terminal이 아니라 **VTE 라이브러리**가 설정합니다. XFCE Terminal, guake, terminator, sakura 등이 같은 값을 냅니다. 결정적 근거로 VTE 자체 소스가 새 세션을 띄우기 전 `GNOME_TERMINAL_` 접두사를 "다른 터미널의 잔여물"로 보고 제거합니다. 제품 특정은 `GNOME_TERMINAL_SCREEN`/`_SERVICE`로만 가능하므로, `VTE_VERSION`만 있으면 `TerminalVTE`(계열)로 보고합니다.

## 생존 확인이 가능한 신호

거의 모든 변수는 과거 사실의 스냅샷이지만 셋은 예외입니다.

- **kitty `KITTY_PID`** — 실제 OS PID. 프로세스 조회로 낡은 마커와 살아 있는 터미널을 구분할 수 있어 `Terminal.PID`로 노출합니다.
- **Konsole `KONSOLE_DBUS_*`**, **GNOME Terminal `GNOME_TERMINAL_SCREEN`/`_SERVICE`** — 살아 있는 D-Bus 객체 경로. 다만 두 제품 모두 이를 공식 감지 API로 문서화하지는 않았으므로, D-Bus 호출 실패를 "세션 종료"의 증거로 삼는 것은 합리적 추론일 뿐 보장된 계약이 아닙니다.

## 구분할 수 없는 경우

- **Konsole vs 임베디드 터미널** — `konsolepart`(Dolphin F4 패널, Kate·KDevelop 터미널 뷰, Krusader)가 동일한 `konsoleprivate` 라이브러리를 사용해 같은 `KONSOLE_*` 변수를 주입합니다. 환경변수만으로는 독립 실행 Konsole과 에디터 내장 터미널을 구분할 수 없습니다.

  이 신호가 증명하는 것은 "Konsole 엔진"이지 "사용자가 Konsole 창을 보고 있다"가 아닙니다. 그래서 `runby`는 Konsole의 `Confidence`를 `definite`로 올리지 않고 **`probable`로 고정**합니다. 마커 자체는 확실하지만 그것이 지목하는 대상이 확실하지 않은 경우이며, 마커가 약해서 `probable`인 `VTE_VERSION`과는 이유가 다릅니다.
- **Warp의 에이전트 실행 vs 사람 입력** — Warp는 AI 에이전트 기능을 제공하지만 에이전트가 실행한 명령을 표시하는 변수를 문서화하지 않습니다. Zed와 동일하게 터미널로만 처리합니다.
