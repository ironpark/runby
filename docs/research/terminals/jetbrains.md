---
title: JetBrains IDE 터미널 (JediTerm)
slug: jetbrains
research_date: 2026-08-31
open_source: true
repository: https://github.com/JetBrains/intellij-community
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 소스에서 `TERMINAL_EMULATOR`·`TERM_SESSION_ID` 주입 코드와 WSLENV 등록 코드는 확인했지만, (1) 2025년 도입된 신형 터미널(reworked terminal)과 구형 Classic 터미널이 같은 변수 집합을 주입하는지, (2) Android Studio 등 서드파티 IntelliJ 플랫폼 IDE에서도 같은 값이 나오는지, (3) IDE 실행 파일 이름이 플랫폼별로 무엇인지, (4) `TERM_SESSION_ID`가 터미널 탭마다 새로 생성되는지 창 단위인지는 실제 실행으로 확인하지 않음
---

# JetBrains IDE 터미널 (JediTerm)

IntelliJ IDEA·PyCharm·GoLand·WebStorm·Rider·CLion 등 JetBrains IDE의 내장 터미널은 [JediTerm](https://github.com/JetBrains/jediterm) 기반이며, 셸을 띄울 때 `TERMINAL_EMULATOR`와 `TERM_SESSION_ID`를 주입합니다.

```java
    if (...) {
      envs.put("TERM", "xterm-256color");
    }
    envs.put(TERMINAL_EMULATOR, "JetBrains-JediTerm");
    envs.put(TERM_SESSION_ID, UUID.randomUUID().toString());
```

— [공식 소스: `LocalOptionsConfigurer.getTerminalEnvironment`](https://github.com/JetBrains/intellij-community/blob/2dad0c70ad8269c90f2d1101749347e3e0241213/plugins/terminal/src/org/jetbrains/plugins/terminal/runner/LocalOptionsConfigurer.java#L163-L166)

상수 이름은 별도 파일에 정의되어 있고 값이 변수명과 같습니다.

```kotlin
  const val TERMINAL_EMULATOR: String = "TERMINAL_EMULATOR"
  const val TERM_SESSION_ID: String = "TERM_SESSION_ID"
```

— [공식 소스: `TerminalEnvironment`](https://github.com/JetBrains/intellij-community/blob/702e47ed87bec16b74ed60fd5200398fcc822902/plugins/terminal/src/org/jetbrains/plugins/terminal/util/TerminalEnvironment.kt#L20-L21)

## 이 문서의 핵심: 값이 IDE가 아니라 엔진을 가리킵니다

`JetBrains-JediTerm`은 IDE 이름이 아니라 **터미널 엔진 이름**입니다. 이 코드는 `plugins/terminal`이라는 IntelliJ 플랫폼 공통 플러그인에 있으므로, IntelliJ 플랫폼 위에 만들어진 모든 IDE가 정확히 같은 값을 내보냅니다. 여기에는 JetBrains 제품 전부(IDEA, PyCharm, GoLand, WebStorm, Rider, CLion, RubyMine, PhpStorm, DataGrip)뿐 아니라 **Google의 Android Studio처럼 JetBrains가 만들지 않은 IntelliJ 플랫폼 IDE**도 포함됩니다.

즉 이 마커가 증명하는 것은 "JetBrains IDE의 내장 터미널"이 아니라 "**IntelliJ 플랫폼의 JediTerm 터미널**"입니다. `runby`는 이를 `jetbrains`라는 계열 이름으로 보고하고 `Confidence`를 `probable`로 둡니다 — Konsole이 `konsolepart`를 임베드한 다른 앱과 구분되지 않아 `probable`인 것, VS Code가 포크와 구분되지 않아 `probable`인 것과 같은 판단입니다.

**어느 IDE인지 알아낼 방법은 환경변수에 없습니다.** 위 코드가 주입하는 전체 목록을 확인했지만 제품명·버전·프로젝트 경로를 담은 변수는 없습니다. 이는 VS Code가 `TERM_PROGRAM_VERSION`이라도 주는 것보다 약한 상황입니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERMINAL_EMULATOR` | 문자열, 정확히 `JetBrains-JediTerm` | 실행 식별 | 셸·프로그램이 IDE 터미널을 감지 | 적합(계열 한정) — 모든 로컬 터미널 자식에 조건 없이 주입되지만, IntelliJ 플랫폼 IDE 전체가 같은 값을 냄 | [소스 L165](https://github.com/JetBrains/intellij-community/blob/2dad0c70ad8269c90f2d1101749347e3e0241213/plugins/terminal/src/org/jetbrains/plugins/terminal/runner/LocalOptionsConfigurer.java#L165) |
| `TERM_SESSION_ID` | UUID 문자열 | 실행 식별 | 터미널 세션 구분 | 보조 신호 — **단독으로는 절대 마커가 될 수 없음.** Apple Terminal이 같은 이름을 쓰므로 이름만으로는 두 터미널을 구분하지 못함 | [소스 L166](https://github.com/JetBrains/intellij-community/blob/2dad0c70ad8269c90f2d1101749347e3e0241213/plugins/terminal/src/org/jetbrains/plugins/terminal/runner/LocalOptionsConfigurer.java#L166) |
| `TERM` | `xterm-256color` (조건부) | 설정 | terminfo 기능 식별 | 부적합 — 조건부로만 설정되고 값이 범용적이라 식별력이 없음 | [소스 L163](https://github.com/JetBrains/intellij-community/blob/2dad0c70ad8269c90f2d1101749347e3e0241213/plugins/terminal/src/org/jetbrains/plugins/terminal/runner/LocalOptionsConfigurer.java#L163) |

### `TERM_SESSION_ID` 이름 충돌

`TERM_SESSION_ID`는 **Apple Terminal도 세션 식별자로 쓰는 이름**입니다([`apple-terminal.md`](apple-terminal.md)). 두 제품이 같은 변수 이름을 서로 다른 형식의 값으로 채우므로, 이 변수는 어느 쪽에서도 마커가 될 수 없고 마커가 먼저 결정된 뒤에만 읽어야 합니다.

`runby`의 구조에서는 이 충돌이 문제가 되지 않습니다. Apple Terminal의 마커는 `TERM_PROGRAM=Apple_Terminal`이고 JetBrains의 마커는 `TERMINAL_EMULATOR=JetBrains-JediTerm`이라 서로 겹치지 않으며, `TERM_SESSION_ID`는 양쪽 모두 마커가 결정된 **뒤에** 세션 식별자로 읽힙니다. 이는 `specCore`가 "마커가 결정하고, 그 다음에 이름으로 값을 읽는다"를 구조로 강제한 덕분입니다.

다만 이 충돌 때문에 **`TERM_SESSION_ID`의 값 형식으로 제품을 추론해서는 안 됩니다.** JetBrains는 UUID를, Apple Terminal은 자체 형식을 쓰지만, 형식 매칭은 이 저장소가 금지하는 종류의 추측입니다.

## WSL 경계를 의도적으로 넘습니다

JetBrains는 이 두 변수를 `WSLENV`에 등록해 Windows에서 WSL로 넘깁니다.

```kotlin
    val envNamesToPass = buildList {
      addAll(userDefinedEnvData?.envs?.keys.orEmpty())
      add(TERMINAL_EMULATOR)
      add(TERM_SESSION_ID)
    }.distinctBy { it.lowercase() }

    val newItems = envNamesToPass.filter { it != WSLENV }.map { "$it/u" }
```

— [공식 소스: `TerminalEnvironment.doSetWslEnv`](https://github.com/JetBrains/intellij-community/blob/702e47ed87bec16b74ed60fd5200398fcc822902/plugins/terminal/src/org/jetbrains/plugins/terminal/util/TerminalEnvironment.kt#L59-L70)

`/u` 플래그는 "Win32에서 WSL을 호출할 때만 포함"을 뜻하므로([`../remote/wsl.md`](../remote/wsl.md)의 플래그 표), Windows에서 IDE 터미널로 WSL 셸을 열면 **Linux 프로세스가 Windows 쪽 IDE의 터미널 마커를 그대로 봅니다.**

이는 Windows Terminal이 `WT_SESSION`을 플래그 없이 넘기는 것과 같은 부류이며, `remote/wsl.md`가 이미 지적한 "터미널 변수가 더 이상 같은 OS를 뜻하지 않는다"의 두 번째 실제 사례입니다. 차이는 Windows Terminal이 양방향인 반면 JetBrains는 `/u`라 Win32→WSL 한 방향이라는 점입니다.

## 상속과 잔존

다른 터미널과 같은 네 가지 한계(멀티플렉서 잔존, 데몬화, 위조, 상속)가 그대로 적용되며, JetBrains에는 다음 특성이 있습니다.

- **tmux 안에서 살아남습니다.** JetBrains의 마커는 `TERM_PROGRAM`이 아니라 `TERMINAL_EMULATOR`이므로, tmux가 `TERM_PROGRAM`을 자기 이름으로 덮어써도 이 값은 통과합니다. 즉 kitty·Windows Terminal·Alacritty·Konsole·GNOME Terminal과 같은 부류로, tmux 안에서 **거짓 음성이 아니라 낡은 거짓 양성**을 만드는 쪽입니다. `runby`가 멀티플렉서 감지 시 신뢰도를 낮추는 규칙이 그대로 적용됩니다.
- **SSH로 새어 나가지 않습니다.** `LC_*` 네임스페이스가 아니고 JetBrains가 `SendEnv`에 등록하는 기능도 확인되지 않았으므로, iTerm2의 `LC_TERMINAL` 같은 원격 전파 경로는 없습니다. 다만 WSL 경계는 위와 같이 의도적으로 넘습니다.

## 실행 파일 이름

`runby`의 `Executables`는 **비워 둡니다.**

IDE 프로세스의 실행 파일 이름은 제품마다(`idea`, `pycharm`, `goland`, `webstorm`, `rider`, `clion`, `studio`) 다르고, 플랫폼과 설치 형태(네이티브 런처, JVM 직접 실행, Toolbox 설치, snap/flatpak)에 따라 `java`로 보일 수도 있습니다. 계열 전체를 덮으려면 제품별 이름을 모두 공식 자료로 확인해야 하는데 아직 하지 않았고, `java`처럼 일반적인 이름은 무관한 프로세스를 잘못 라벨링합니다. 확인되지 않은 이름을 넣는 것보다 비워 두는 편이 안전하다는 이 패키지의 기존 판단(Apple Terminal, Orca)을 따릅니다.

## 실행 주체 감지에 관한 결론

`TERMINAL_EMULATOR=JetBrains-JediTerm`은 공식 소스에서 모든 로컬 IDE 터미널 자식에 조건 없이 주입되는 유일한 마커입니다. 값이 엔진 이름이라 IntelliJ 플랫폼 IDE 전체(서드파티인 Android Studio 포함)를 가리키므로, `runby`는 `jetbrains`를 계열 이름으로 보고하고 신뢰도를 `probable`로 둡니다.

`TERM_SESSION_ID`는 터미널마다 새 UUID라 세션 식별자로 유용하지만 Apple Terminal과 이름이 겹치므로 마커로는 쓸 수 없고, 마커가 결정된 뒤에만 읽습니다. 어느 IDE인지, 어떤 프로젝트인지는 환경변수만으로 알 수 없습니다.

**JetBrains Junie는 이 축이 아닙니다.** Junie는 별도의 에이전트 제품이며 자체 실행 마커가 확인되지 않아 감지 대상이 아닙니다([`../agents/junie.md`](../agents/junie.md)). 이 문서의 마커는 "IDE 터미널에서 실행됐다"만 증명하며, 사람이 쳤는지 Junie가 요청했는지는 구분하지 못합니다 — Zed·VS Code와 같은 이유로 터미널 축에만 나타납니다.

## 공식 문서

- [Terminal | IntelliJ IDEA Documentation](https://www.jetbrains.com/help/idea/terminal-emulator.html) — 내장 터미널 개요
- [JetBrains/jediterm](https://github.com/JetBrains/jediterm) — 터미널 엔진
- [공식 소스: `LocalOptionsConfigurer`](https://github.com/JetBrains/intellij-community/blob/2dad0c70ad8269c90f2d1101749347e3e0241213/plugins/terminal/src/org/jetbrains/plugins/terminal/runner/LocalOptionsConfigurer.java#L163-L166) — `TERMINAL_EMULATOR`·`TERM_SESSION_ID` 주입
- [공식 소스: `TerminalEnvironment`](https://github.com/JetBrains/intellij-community/blob/702e47ed87bec16b74ed60fd5200398fcc822902/plugins/terminal/src/org/jetbrains/plugins/terminal/util/TerminalEnvironment.kt#L20-L70) — 상수 정의와 WSLENV 등록
