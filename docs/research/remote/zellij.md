---
title: Zellij
slug: zellij
research_date: 2026-08-31
open_source: true
repository: https://github.com/zellij-org/zellij
product_type: terminal_multiplexer
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 사용자 가이드([Integration](https://zellij.dev/documentation/integration.html))가 `ZELLIJ`가 세션 안에서 `0`으로 설정된다고 명시하고, 공식 소스([`envs.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/envs.rs), pty spawn 지점)에서 `ZELLIJ`·`ZELLIJ_SESSION_NAME`·`ZELLIJ_PANE_ID`·`ZELLIJ_SOCKET_DIR` 주입 지점과 값을 직접 확인했으므로 별도 실행 검증 없이 표를 확정할 수 있음. 다만 서버-클라이언트 분리 구조로 인한 환경 잔존 범위(오래 떠 있는 세션에서 서버가 실제로 어떤 시점의 환경을 들고 있는지)는 장기 실행 세션에서 직접 확인하면 더 확실해짐
---

# Zellij

아래 내용은 조사 시점의 공식 소스 [`zellij-org/zellij`](https://github.com/zellij-org/zellij) 커밋 [`e839bff`](https://github.com/zellij-org/zellij/commit/e839bfffa586992364309a685b2c71f3b23c247e)와 공식 사용자 가이드 [Integration](https://zellij.dev/documentation/integration.html)을 기준으로 확인했습니다. Zellij는 Rust로 작성된 오픈소스 터미널 멀티플렉서로, tmux·screen과 마찬가지로 서버 프로세스가 세션 상태를 들고 있고 클라이언트가 거기 붙었다 떨어졌다 하는 client-server 구조를 씁니다. `runby`가 이미 `TMUX`/`STY`로 감지하는 멀티플렉서 축에 Zellij가 빠져 있으며, 이 문서는 그 공백을 메우는 데 필요한 사실을 정리합니다.

## 실행 식별 신호

Zellij 세션 안에서 실행되는 모든 프로세스는 최소 `ZELLIJ`와 `ZELLIJ_SESSION_NAME`을 물려받습니다. `ZELLIJ`는 세션의 **존재**를 나타내는 마커이고, `ZELLIJ_SESSION_NAME`은 그 세션의 사람이 읽을 수 있는 식별자입니다. 세션 안에서 새로 열리는 개별 pane은 여기에 더해 `ZELLIJ_PANE_ID`를 받습니다. 소스 전체에서 `ZELLIJ*` 접두사를 가진 변수를 검색한 결과, 자식 프로세스 환경에 실제로 주입되는 것은 이 넷뿐입니다(`ZELLIJ_SOCKET_DIR`는 조회 함수만 확인되며 어디서 설정되는지는 소스에서 직접 확인하지 못했습니다 — 실행 환경(플랫폼 기본 소켓 디렉터리)에서 유래하는 값으로 보입니다).

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `ZELLIJ` | 문자열 리터럴 `"0"` (고정값) | 실행 식별 | Zellij 세션 안에서 실행 중임을 나타내는 존재 마커 | 적합 — 네 곳의 소스 진입점 모두에서 세션 시작 시 반드시 설정됨. 단, 값이 `"0"`이므로 **반드시 존재 여부로만 판정해야 하며 불리언 파싱 결과(예: Go의 `strconv.ParseBool`)를 신뢰해서는 안 됨** | [Integration](https://zellij.dev/documentation/integration.html), [`envs.rs#L11-L17`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/envs.rs), [`zellij-server/src/lib.rs#L901`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-server/src/lib.rs#L901) |
| `ZELLIJ_SESSION_NAME` | 문자열 (사용자 지정 또는 자동 생성 형용사-명사 조합) | 상태·컨텍스트 | 현재 세션의 이름 | 보조 신호 — 세션 존재를 보강하지만 이름은 사용자가 선택·재사용할 수 있어 유일 식별자가 아니고, 세션 이름 변경 시 이미 열린 pane에는 갱신되지 않음 | [Integration](https://zellij.dev/documentation/integration.html), [`envs.rs#L19-L27`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/envs.rs) |
| `ZELLIJ_PANE_ID` | 정수 문자열 (`terminal_id` 카운터) | 실행 식별 | 현재 프로세스가 실행 중인 pane의 id. `zellij action`류 CLI가 `--pane-id` 생략 시 기본 대상으로 사용 | 보조 신호 — pane 단위로는 구체적이나 값이 세션 내부 카운터라 pane이 닫히고 새로 열리면 재사용될 수 있고, 전역 유일성이 없음 | [CLI Actions](https://zellij.dev/documentation/cli-actions), [`os_input_output_unix.rs#L226`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-server/src/os_input_output_unix.rs#L226), [`actions.rs#L2118-L2119`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/input/actions.rs) |
| `ZELLIJ_SOCKET_DIR` | 디렉터리 경로 문자열 | 설정 | 세션 제어 소켓이 위치한 디렉터리 | 부적합 — 조회 함수만 공개 소스에서 확인되며 실행 주체 식별과 무관한 배관용 경로 정보 | [`envs.rs#L29-L32`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/envs.rs) |
| `TERM` | 문자열, Zellij가 재작성하지 않음(원본 터미널 값 유지) | 설정 | 터미널이 지원을 주장하는 terminfo capability 이름 | 부적합 — 정체성이 아니라 능력 협상용 값이며, 사용자가 덮어쓸 수 있고, Zellij는 이 값을 자신의 것으로 바꾸지 않으므로 Zellij 감지에는 쓸 수 없음(다만 tmux/screen과 달리 뭉개지지 않는다는 사실 자체는 기록할 가치가 있음) | 소스 전체 검색 결과 명시적 재작성 지점 없음(부재 증거, 공식 문서 페이지는 미확인) |

### `ZELLIJ`는 값이 `0`입니다 — 존재만 확인해야 합니다

공식 문서는 "`ZELLIJ` gets set to `0` inside a zellij session"이라고 명시적으로 씁니다([Integration](https://zellij.dev/documentation/integration.html)). 소스에서도 이를 그대로 확인할 수 있습니다. `zellij-utils/src/envs.rs`는 키 상수만 정의하고,

```rust
pub const ZELLIJ_ENV_KEY: &str = "ZELLIJ";
pub fn set_zellij(v: String) {
    set_var(ZELLIJ_ENV_KEY, v);
}
```

실제 값은 클라이언트·서버 양쪽의 시작 경로 네 곳 모두에서 리터럴 문자열 `"0"`으로 고정 호출됩니다.

- [`zellij-client/src/lib.rs#L863`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-client/src/lib.rs#L863) — 중첩 세션(nested session) 진입 경로
- [`zellij-client/src/lib.rs#L991`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-client/src/lib.rs#L991) — `start_client`
- [`zellij-client/src/lib.rs#L1645`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-client/src/lib.rs#L1645) — detach 시작 경로
- [`zellij-server/src/lib.rs#L901`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-server/src/lib.rs#L901) — 서버 시작 경로

**이 사실은 `runby`에 직접적인 함정입니다.** `runby`는 불리언 값을 Go의 `strconv.ParseBool`로 해석합니다. `ParseBool`은 `"0"`을 `false`로 읽습니다. 따라서 `ZELLIJ`를 "불리언으로 파싱해서 참이면 감지"하는 방식으로 구현하면 Zellij 세션 안에 있는데도 항상 `false`가 나와 감지에 **실패**합니다. 이 마커는 값의 참·거짓이 아니라 **변수 자체의 존재 여부**로만 판정해야 합니다(`os.LookupEnv`에 해당하는 존재 확인). kitty의 `KITTY_WINDOW_ID`나 tmux의 `TMUX`와 동일한 패턴이지만, Zellij는 그 값이 하필 `"0"`이라는 점에서 순진한 불리언 파싱 구현에 훨씬 잘 걸리는 함정입니다.

### `ZELLIJ_SESSION_NAME` — 사용자가 정하는, 유일하지 않은 이름

`ZELLIJ_SESSION_NAME`은 세션 이름을 담습니다. 이름은 `zellij --session <name>`으로 사용자가 직접 지정하거나, 지정하지 않으면 Zellij가 임의 생성한 형용사-명사 조합(예: `sad-panda`)이 됩니다([Session Management](https://zellij.dev/tutorials/session-management/), [MANPAGE](https://github.com/zellij-org/zellij/blob/main/docs/MANPAGE.md)). 즉 tmux의 세션 이름과 마찬가지로 **사람이 선택하거나 재사용할 수 있는 문자열**이며, 시스템이 보장하는 전역 유일 식별자가 아닙니다. 같은 이름의 세션을 지웠다가 다시 만들면 완전히 다른 세션이 같은 `ZELLIJ_SESSION_NAME` 값을 갖습니다.

또한 공식 문서는 세션 이름을 바꿔도(`zellij action rename-session`) **이미 열려 있는 pane의 `ZELLIJ_SESSION_NAME`은 갱신되지 않고, 새로 여는 pane에만 새 이름이 반영된다**고 명시합니다([Integration](https://zellij.dev/documentation/integration.html)). 같은 세션 안에서도 pane마다 서로 다른(낡은) 값을 들고 있을 수 있다는 뜻이므로, `runby`가 이 값을 세션의 "현재" 이름으로 신뢰하려면 이 잔존 가능성을 함께 알려야 합니다.

### `ZELLIJ_PANE_ID` — pane 단위로 존재, 서버가 아니라 pane 생성 시점에 주입

`ZELLIJ_PANE_ID`는 공식 문서와 소스 양쪽에서 확인됩니다. CLI 문서는 "Defaults to `$ZELLIJ_PANE_ID` if not provided"라고 pane-targeting 명령의 기본값으로 이 변수를 언급하고([CLI Actions](https://zellij.dev/documentation/cli-actions)), 소스에서는 pty를 열 때마다 그 pane의 자식 프로세스 환경에 직접 주입합니다.

```rust
command
    .args(&cmd.args)
    .env("ZELLIJ_PANE_ID", &format!("{}", terminal_id))
```

([`zellij-server/src/os_input_output_unix.rs#L226`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-server/src/os_input_output_unix.rs#L226))

값은 정수형 `terminal_id`이며, `zellij action` 계열 CLI 명령이 `--pane-id`를 생략했을 때 "지금 이 pane"을 가리키는 기본값으로 그대로 쓰입니다([`zellij-utils/src/input/actions.rs#L2118-L2119`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/input/actions.rs)). Windows 쪽 구현(`os_input_output_windows.rs`)도 동일한 이름의 변수를 pane 진입점 환경(`pane_entry`)에 주입합니다.

### `TERM` — Zellij는 값을 덮어쓰지 않습니다

소스 전체를 검색해도 Zellij가 pane을 열 때 `TERM`을 특정 값으로 강제로 설정하는 코드 경로는 없습니다(`os_input_output_unix.rs`의 pty 생성 코드는 `ZELLIJ_PANE_ID`만 명시적으로 `.env()`하고 `TERM`은 건드리지 않습니다). 즉 새 pane의 `TERM`은 Zellij 서버 프로세스가 시작될 때 물려받은 값을 그대로 이어받습니다 — 사용자의 원래 터미널이 설정한 `TERM`(예: `xterm-256color`, `alacritty` 등)이 그대로 보입니다. 이는 **tmux/screen이 자신의 `tmux-256color`/`screen`으로 `TERM`을 명시적으로 재작성하는 것과 다른 동작**입니다. 다만 이 사실은 소스 코드 검색으로 확인한 부재(negative) 증거이며, `TERM`을 다루는 공식 참조 문서 페이지를 별도로 찾지는 못했습니다.

## 환경 변형

이 절이 이 문서의 핵심입니다. 멀티플렉서는 터미널이나 CI와 달리 "새 pane이 무엇을 물려받는가"를 자신이 능동적으로 결정하는 소프트웨어입니다.

### 새 pane은 서버 프로세스의 환경을 물려받습니다 — 최초 시작 시점의 스냅샷

Zellij는 client-server 구조입니다. 세션을 처음 만들 때 클라이언트가 `zellij-server` 실행 파일을 자식 프로세스로 spawn합니다.

```rust
pub fn spawn_server(socket_path: &Path, debug: bool) -> io::Result<()> {
    let mut cmd = Command::new(current_exe()?);
    ...
```

([`zellij-client/src/lib.rs#L464`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-client/src/lib.rs#L464), 유닉스/윈도우 두 구현 모두 동일 패턴)

이 호출에는 `.env_clear()`나 명시적 환경 재구성이 없으므로, Rust `std::process::Command`의 기본 동작에 따라 서버 프로세스는 **자신을 띄운 최초 클라이언트 프로세스의 환경변수 전체를 그대로 물려받습니다**. 이후 그 서버 프로세스 안에서 pane을 새로 열 때마다(`os_input_output_unix.rs`의 pty open 코드) 또 다른 `Command`가 생성되는데, 여기서도 `ZELLIJ_PANE_ID`만 명시적으로 덮어쓸 뿐 나머지는 **서버 프로세스 자신의 환경**을 그대로 물려줍니다.

결론적으로 새 pane이 받는 환경은 "지금 붙어 있는(attach한) 클라이언트의 환경"이 아니라 **"세션을 최초로 만든 클라이언트가 서버를 띄운 그 순간의 환경 스냅샷"**입니다. 세션이 오래 살아 있을수록, 그 스냅샷은 지금 사용자가 보고 있는 실제 터미널·셸 상태와 점점 더 멀어질 수 있습니다.

### tmux의 `update-environment`에 해당하는 메커니즘이 없습니다

tmux는 `update-environment` 옵션(기본값 `DISPLAY KRB5CCNAME SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WINDOWID XAUTHORITY`)으로 클라이언트가 attach할 때마다 지정된 변수를 그 클라이언트의 현재 값으로 세션 환경에 갱신합니다. Zellij 소스와 공식 문서 어디에도 이에 대응하는 메커니즘을 찾지 못했습니다 — attach 시점에 pane 환경의 일부를 새 클라이언트 값으로 동기화하는 코드 경로가 없습니다. 사용자가 직접 이 기능을 요청한 [GitHub Issue #1987](https://github.com/zellij-org/zellij/issues/1987)("set an environment variable for the newly created pane or tab")이 조사 시점 기준 **미해결(open)** 상태로 남아 있다는 사실이 이를 뒷받침합니다.

**분명히 말해두면: Zellij에는 tmux의 `update-environment`에 해당하는 기능이 없습니다.** 이는 세션 안 pane의 환경이 재접속으로도 갱신되지 않는다는 뜻이며, 앞서 설명한 "최초 시작 시점 스냅샷" 문제를 완화할 공식 수단이 없다는 뜻이기도 합니다. 즉 오래된 세션의 pane일수록 환경이 더 낡아 있을 수 있고, Zellij는 tmux보다 이를 보정할 수단이 더 적습니다.

### `set-environment`/`show-environment`에 대응하는 기능도 없습니다

tmux는 `set-environment`로 세션 전역 환경변수를 세션 안에서 직접 설정하고 `show-environment`로 조회할 수 있습니다. Zellij 공식 CLI 문서([CLI Actions](https://zellij.dev/documentation/cli-actions), [Commands](https://zellij.dev/documentation/plugin-api-commands.html))와 소스 어디에도 실행 중인 세션의 환경변수를 조회하거나 설정하는 `zellij action` 하위 명령을 확인하지 못했습니다. 대신 Zellij는 **세션 시작 전** 설정 파일(`config.kdl`)이나 레이아웃 파일(`.kdl`)에 `env { ... }` 블록으로 키-값을 정적으로 선언하는 방식만 제공합니다. 이 값은 `zellij-utils/src/envs.rs`의 `EnvironmentVariables::set_vars()`가 세션이 뜰 때 `std::env::set_var`로 한 번 적용하는 것으로 끝이며, 이후 실행 중에 이를 다시 읽거나 고치는 공식 명령은 없습니다.

```rust
pub fn set_vars(&self) {
    for (k, v) in &self.env {
        set_var(k, v);
    }
}
```

([`zellij-utils/src/envs.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/envs.rs))

정리하면 Zellij는 "설정 파일로 시작 시점에 정적으로 주입"만 지원하고, tmux의 `set-environment`/`show-environment`처럼 **실행 중인 세션의 환경을 동적으로 조회·수정하는 공식 인터페이스는 제공하지 않습니다.**

### 다른 터미널에서 attach해도 기존 pane은 바뀌지 않습니다

위 사실들을 종합하면 답은 명확합니다. 세션 안 기존 pane이 물려받은 환경은 서버 프로세스가 시작될 때의 스냅샷이고, 이를 갱신할 `update-environment` 대응 기능도 없으므로, **터미널 A에서 시작한 Zellij 세션을 터미널 B에서 나중에 attach해도 이미 열려 있던 pane들의 환경변수는 바뀌지 않습니다.** 오직 그 이후 attach된 클라이언트 쪽에서 **새로** 여는 pane만 attach 시점 클라이언트의 조건(예: 클라이언트가 보낸 초기 터미널 크기 등 IPC로 별도 전달되는 정보)의 영향을 받을 수 있고, 환경변수 자체의 상속 경로는 여전히 서버 프로세스 스냅샷입니다. 이는 `docs/research/terminals/README.md`가 이미 tmux/screen에 대해 정리한 "멀티플렉서 잔존" 문제가 Zellij에도 그대로, 오히려 완화 수단 없이 적용됨을 의미합니다.

## 다른 축에 미치는 영향

`runby`는 세 축(agent/CI/terminal)을 보고합니다. Zellij는 이 중 **터미널(terminal) 축을 구조적으로 오염시키는 멀티플렉서**이며, 나머지 두 축에는 직접적인 신호를 제공하지 않습니다.

- **터미널 축(`Result.Terminal`)** — 오염 방향은 명확합니다. `docs/research/terminals/README.md`가 정리한 "멀티플렉서 잔존" 문제와 동일하게, Zellij 세션 안 프로세스가 보고하는 `TERM_PROGRAM`/`KITTY_WINDOW_ID` 등 터미널 식별 변수는 **세션을 최초로 만든 클라이언트가 실행되던 터미널**을 가리킵니다. 지금 실제로 화면을 보고 있는 터미널 에뮬레이터와 무관할 수 있고, 앞 절에서 확인했듯 Zellij에는 이를 갱신할 tmux식 `update-environment`조차 없어 tmux보다 더 오래 낡은 값을 들고 있을 수 있습니다. 다만 `TERM` 자체는 Zellij가 재작성하지 않으므로, 적어도 `TERM`이 가리키는 terminfo 능력 정보는 원본 터미널의 것이 그대로 유지됩니다(tmux/screen처럼 `tmux-256color`/`screen`으로 뭉개지지 않음).
- **에이전트 축(`Kind`)** — Zellij 공식 변수 중 "누가(사람 또는 어떤 agent 하네스가) 이 명령을 요청했는가"를 나타내는 것은 없습니다. `ZELLIJ`/`ZELLIJ_SESSION_NAME`/`ZELLIJ_PANE_ID`는 모두 실행 공간(멀티플렉서 세션·pane)을 가리킬 뿐 요청 주체를 가리키지 않으므로, `runby`가 이미 다른 문서에서 확인한 원칙("터미널을 소유한다는 사실은 에이전트 실행의 증거가 아니다")이 그대로 적용됩니다. Zellij는 agent 관련 축에 어떤 신호도 제공하지 않습니다.
- **CI 축(`Result.CI`)** — Zellij는 대화형 터미널 멀티플렉서이며 CI 실행기가 아니므로 CI 축과 무관합니다. CI 파이프라인 안에서 우연히 Zellij가 실행될 이유도, 공식적으로 그런 사용 사례를 문서화한 근거도 없습니다.

### Zellij를 `runby`에 추가해야 하는가

**추가해야 합니다.** `runby`는 이미 `TMUX`와 `STY`로 `Terminal.Multiplexer`를 채우고 그때 `Confidence`를 `probable`로 낮추는 정책을 갖고 있습니다(`terminal.go`의 `detectMultiplexer`). Zellij는 이 두 제품과 정확히 같은 구조적 위치에 있는 세 번째 멀티플렉서이므로, 같은 정책을 그대로 적용하는 것이 일관적입니다.

- **마커**: `ZELLIJ` 환경변수의 **존재 여부**만 검사해야 합니다. 값이 `"0"`이라는 사실 때문에, `TMUX`/`STY`와 달리 이 변수는 절대 `strconv.ParseBool`로 판정해서는 안 됩니다 — 존재 확인(예: `Value(env, "ZELLIJ")`가 반환하는 `ok`)만으로 감지하고, 값 자체는 검사에 쓰지 않아야 합니다.
- **부가 정보**: `ZELLIJ_SESSION_NAME`은 tmux 세션 이름과 동일한 성격(사용자 선택, 비유일)이므로 있다면 참고 정보로만 노출하고 신뢰 경계로 쓰지 않아야 합니다.
- **Confidence**: tmux/screen과 동일하게, Zellij가 감지되면 `Confidence`를 `probable`로 낮추는 것이 이 문서에서 확인한 사실(서버 스냅샷 상속, `update-environment` 부재)과 정확히 부합합니다.

## 실행 주체 감지에 관한 결론

Zellij 세션의 존재 자체는 `ZELLIJ` 환경변수의 **존재**로 신뢰성 있게 판정할 수 있습니다 — 공식 문서와 네 곳의 소스 주입 지점 모두 세션 안에서 이 변수가 항상 설정됨을 보여줍니다. 다만 그 값이 리터럴 `"0"`으로 고정되어 있다는 점은 구현상 치명적인 함정이며, `runby`처럼 불리언 파싱을 쓰는 라이브러리는 존재 확인과 값 파싱을 반드시 분리해야 합니다.

`ZELLIJ_SESSION_NAME`과 `ZELLIJ_PANE_ID`는 각각 세션과 pane 수준에서 유용한 보조 식별자이지만, 전자는 사용자가 정하고 재사용 가능한 이름이라 유일성이 없고, 후자는 정수 카운터라 pane이 닫히고 새로 열리면 재사용될 수 있습니다. 두 값 모두 "지금 이 세션·pane"을 가리키는 스냅샷이지 전역적으로 안정된 식별자가 아닙니다.

가장 중요한 구조적 사실은, Zellij의 새 pane이 물려받는 환경이 **세션을 최초로 만든 시점의 서버 프로세스 스냅샷**이며, tmux의 `update-environment`나 `set-environment`/`show-environment`에 해당하는 어떤 공식 갱신·조회 수단도 없다는 점입니다. 이는 Zellij 세션 안 프로세스가 보고하는 터미널 식별 신호(`TERM_PROGRAM` 등, `ZELLIJ` 자체는 예외)가 tmux보다도 더 쉽게, 그리고 더 오래 낡을 수 있음을 뜻합니다. 반면 `TERM` 값 자체는 Zellij가 재작성하지 않으므로 원본 터미널의 terminfo 정보는 그대로 보존됩니다.

`runby`는 `ZELLIJ`의 존재를 `TMUX`/`STY`와 같은 층위의 멀티플렉서 마커로 추가하고, 값을 절대 불리언으로 해석하지 않으며, 감지 시 `Confidence`를 `probable`로 낮추는 기존 정책을 그대로 적용해야 합니다.

## 공식 문서

- [Integration — Environment Variables](https://zellij.dev/documentation/integration.html)
- [Session Management with Zellij](https://zellij.dev/tutorials/session-management/)
- [CLI Actions](https://zellij.dev/documentation/cli-actions)
- [Commands (plugin API)](https://zellij.dev/documentation/plugin-api-commands.html)
- [MANPAGE.md](https://github.com/zellij-org/zellij/blob/main/docs/MANPAGE.md)
- 공식 소스 저장소 [`zellij-org/zellij`](https://github.com/zellij-org/zellij), 조사 시점 커밋 [`e839bff`](https://github.com/zellij-org/zellij/commit/e839bfffa586992364309a685b2c71f3b23c247e)
  - [`zellij-utils/src/envs.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/envs.rs)
  - [`zellij-server/src/os_input_output_unix.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-server/src/os_input_output_unix.rs)
  - [`zellij-client/src/lib.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-client/src/lib.rs)
  - [`zellij-server/src/lib.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-server/src/lib.rs)
  - [`zellij-utils/src/input/actions.rs`](https://github.com/zellij-org/zellij/blob/e839bfffa586992364309a685b2c71f3b23c247e/zellij-utils/src/input/actions.rs)
- [GitHub Issue #1987 — 환경변수 설정 기능 요청 (미해결)](https://github.com/zellij-org/zellij/issues/1987)
