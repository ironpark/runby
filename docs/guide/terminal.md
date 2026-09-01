# 터미널 축과 TTY

터미널 축은 **의도적으로 가장 약한 신호**입니다. 환경변수는 모든 자손 프로세스가 상속하므로 이 값은 *이 환경을 만든 터미널*을 가리키며, *지금 이 프로세스의 TTY에 붙어 있는 터미널*이 아닙니다.

프롬프트를 띄울지 결정하려는 경우에는 터미널 제품명이 아니라 아래의 `TTY.Interactive`를 사용하세요.

```go
term := runby.Detect().Terminal
term.Program      // "ghostty", "kitty", "vte", ...
term.SessionID    // 창·탭·pane 식별자 (제공하는 터미널만)
term.Version
term.PID          // kitty만. 0이 아니면 생존 확인 가능
term.Term         // TERM 값. 판정에는 쓰지 않고 컨텍스트로만 기록
term.Confidence   // 멀티플렉서가 감지되면 probable로 낮아짐
```

| 터미널 | `TerminalProgram` | 마커 |
|---|---|---|
| iTerm2 | `TerminalITerm2` | `TERM_PROGRAM=iTerm.app` |
| Apple Terminal | `TerminalAppleTerminal` | `TERM_PROGRAM=Apple_Terminal` |
| WezTerm | `TerminalWezTerm` | `TERM_PROGRAM=WezTerm` |
| Ghostty | `TerminalGhostty` | `TERM_PROGRAM=ghostty` |
| Warp | `TerminalWarp` | `TERM_PROGRAM=WarpTerminal` |
| Zed | `TerminalZed` | `ZED_TERM=true` + `TERM_PROGRAM=zed` |
| VS Code 계열 | `TerminalVSCode` | `TERM_PROGRAM=vscode` (제품이 아니라 계열) |
| JetBrains 계열 | `TerminalJetBrains` | `TERMINAL_EMULATOR=JetBrains-JediTerm` (계열) |
| kitty | `TerminalKitty` | `KITTY_WINDOW_ID` |
| Windows Terminal | `TerminalWindowsTerminal` | `WT_SESSION` |
| Alacritty | `TerminalAlacritty` | `ALACRITTY_LOG` |
| Konsole | `TerminalKonsole` | `KONSOLE_VERSION` 또는 `KONSOLE_DBUS_SESSION` |
| GNOME Terminal | `TerminalGNOMETerminal` | `GNOME_TERMINAL_SCREEN` |
| VTE 계열 | `TerminalVTE` | `VTE_VERSION` (제품이 아니라 계열) |

## 이 축을 신뢰 경계로 쓰면 안 되는 이유

- **멀티플렉서 잔존** — [`remote.md`](remote.md)를 참고하십시오. `runby`는 멀티플렉서를 감지하면 `Confidence`를 `probable`로 낮춥니다. 단 그 터미널이 살아 있는 조상으로 확인되면(`Terminal.AncestorPID != 0`) 낮추지 않습니다 — 멀티플렉서 서버는 데몬화되며 재부모화되므로 pane 뒤에서는 원래 터미널이 조상으로 보이지 않기 때문입니다.
- **SSH 정체성 전파** — iTerm2의 `LC_TERMINAL`(기본 켜짐), kitty의 `kitten ssh`, Ghostty의 `ssh-env`가 터미널 정체성을 원격 호스트에 의도적으로 전달합니다. 그래서 `runby`는 `LC_TERMINAL`을 **감지에 전혀 쓰지 않습니다**.
- **데몬화**와 **위조** — 낡은 스냅샷이 남거나 누구나 `export TERM_PROGRAM=...` 할 수 있습니다.

## 알아둘 세부사항

- **`TERM`은 마커가 아닙니다.** 정체성이 아니라 terminfo 능력을 나타내고, 사용자가 덮어쓰며, 멀티플렉서가 교체하고, Alacritty·Ghostty는 terminfo가 없으면 `xterm-256color`로 폴백합니다. 판정이 끝난 뒤 컨텍스트로만 기록합니다.
- **`TERM_PROGRAM`은 절반에서만 쓸 수 있습니다.** kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal, JetBrains는 설정하지 않습니다.
- **`VTE_VERSION`은 계열입니다.** VTE 라이브러리가 설정하므로 XFCE Terminal·guake·terminator 등이 공유합니다. `GNOME_TERMINAL_*` 없이 이 값만 있으면 `TerminalVTE`로 보고합니다.
- **`Terminal.PID`는 kitty만 제공합니다.** 다른 모든 신호와 달리 프로세스 조회로 낡은 마커와 살아 있는 터미널을 구분할 수 있습니다.
- **Konsole은 항상 `probable`입니다.** `konsolepart`(Dolphin, Kate, KDevelop, Krusader)가 같은 라이브러리를 써서 동일한 변수를 주입하므로, 이 증거는 "Konsole 엔진"을 증명할 뿐 사용자가 Konsole 창을 보고 있다는 뜻은 아닙니다.
- **VS Code와 JetBrains도 같은 이유로 `probable`입니다.** `TERM_PROGRAM=vscode`는 공유 소스에 박힌 리터럴이라 Cursor·Windsurf 같은 포크가 그대로 내보내고, `JetBrains-JediTerm`은 IDE가 아니라 터미널 엔진 이름이라 Android Studio 같은 서드파티 IntelliJ 플랫폼 IDE도 포함합니다. 두 값 모두 **엔진**을 증명하지 제품을 증명하지 않습니다.
- **`TERM_SESSION_ID`는 두 제품이 공유합니다.** Apple Terminal과 JetBrains가 같은 이름을 서로 다른 형식으로 씁니다. 그래서 이 변수는 어느 쪽에서도 마커가 아니며, 각 축의 마커가 결정된 **뒤에만** 세션 식별자로 읽힙니다.

터미널별 조사 근거는 [`docs/research/terminals/`](../research/terminals/)에, 멀티플렉서와 원격 실행 계층은 [`docs/research/remote/`](../research/remote/)에 있습니다.

## `HasTerminal()`이지 `IsTerminal()`이 아닙니다

Go에서 `IsTerminal`은 관례적으로 "이 파일 디스크립터가 tty인가"를 뜻합니다(`x/term.IsTerminal(fd)`). 이 패키지에서 그 질문에 답하는 것은 `TTY.Interactive` 쪽이고, 터미널 축은 전혀 다른 질문 — "환경을 만든 에뮬레이터를 식별했는가" — 에 답합니다. 두 뜻이 정면으로 충돌해서, 축 쪽은 `HasTerminal()`로 부릅니다.

```go
result.HasTerminal()     // 에뮬레이터를 식별했는가 (환경변수 기반, 약함)
result.TTY.Interactive   // 내 스트림이 지금 터미널에 붙어 있는가 (시스템콜, 확실함)
```

## TTY: 터미널 축과 무엇이 다른가

두 가지는 이름이 비슷하지만 전혀 다른 질문에 답합니다.

- `Terminal` — **어떤 에뮬레이터가 이 환경을 만들었는가.** 환경변수 기반, 약함.
- `TTY` — **내 표준 스트림이 지금 터미널에 붙어 있는가.** 시스템콜 기반, 확실함.

```go
tty := runby.Detect().TTY
tty.Inspected   // 표준 스트림을 실제로 검사했는가
tty.StdinTTY
tty.StdoutTTY
tty.StderrTTY
tty.Attached    // 세 스트림 중 하나 이상이 터미널
tty.Interactive // stdin과 출력 스트림 하나 이상이 터미널
```

`Interactive`는 프롬프트를 띄우고 응답을 받을 수 있는 형태라는 뜻일 뿐, **사용자가 직접 명령을 호출했다는 증거는 아닙니다.** 에이전트나 서브에이전트도 PTY를 할당할 수 있습니다. 프롬프트를 사용할 수 있는지는 `TTY.Interactive`로 판단하고, 스피너·색상처럼 사람이 출력을 보고 있는지에 따라 달라지는 표시에는 `Unattended()`를 사용하십시오. 정책이 다르면 `IsAgent()`, `IsCI()`, 러너 종류 같은 축을 직접 조합해야 합니다.

`TTY`는 `Result`에서 유일하게 환경변수가 아닌 시스템콜로 얻는 값입니다. 그래서 `WithEnviron` 계열 옵션(이 프로세스의 것이 아닐 수도 있는 환경)에서는 검사하지 않으며 `Inspected`가 `false`입니다. 필요하면 `InspectTTY()`를 직접 호출하거나 `WithTTY()`로 주입하십시오. 반대로 `Terminal`은 환경변수 기반이라 `WithEnviron`에서도 그대로 채워집니다.

**AIX·Solaris·z/OS에서는 `Attached`와 `Interactive`가 항상 `false`입니다.** 표준 `syscall` 패키지가 이 플랫폼들에 `TCGETS`·`SYS_IOCTL`을 노출하지 않기 때문입니다. 환경변수만 읽는 나머지 축은 모든 플랫폼에서 동일하게 동작합니다.
