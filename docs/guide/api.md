# API

처음 사용하는 경우에는 이 레퍼런스보다 [시작하기](getting-started.md)를 먼저 읽는 편이 빠릅니다. 여기서는 공개 옵션과 결과 구조 전체를 설명합니다.

진입점은 `Detect(opts ...Option) Result` 하나입니다.

```go
result := runby.Detect()                                   // 현재 프로세스
result := runby.Detect(runby.WithEnviron(environ))         // 명시적 환경
result := runby.Detect(runby.WithoutTTY())                 // TTY 시스템콜 생략
runby.Register(myDriver)                                   // 사내 드라이버, 프로세스 전체
```

## 옵션

옵션은 세 층으로 나뉩니다. **대부분의 프로그램은 1층에서 끝납니다** — 옵션을 하나도 쓰지 않습니다.

**1층 — 일반 호출자.** `Current()` 또는 인자 없는 `Detect()`. 아래 표는 볼 일이 없습니다.

**2층 — 이 프로세스가 아닌 무언가를 기술하는 래퍼.** `/proc/<pid>/environ`이나 `exec.Cmd.Env`를 읽어 **다른 프로세스**를 분류하거나, 기록해 둔 환경을 나중에 분석하는 경우입니다.

| 옵션 | 설명 |
|---|---|
| `WithEnviron([]string)` | `"NAME=value"` 슬라이스로 환경 지정 |
| `WithTTY(TTY)` | 표준 스트림 상태를 직접 주입 |
| `WithProcessTree(ProcessTree)` | 상위 프로세스 체인을 직접 주입 |
| `WithoutTTY()` | 표준 스트림 검사 생략 (시스템콜 절약) |
| `WithoutProcessTree()` | 상위 프로세스 체인 읽기 생략 |

환경을 넘기면 `TTY`와 `Process` 축은 자동으로 꺼집니다 — 아래 [무엇이 꺼지는가](#무엇이-꺼지는가)를 보십시오. 그 두 축까지 채우려면 `WithTTY`·`WithProcessTree`로 **직접** 주어야 합니다. 기술 대상 프로세스의 상태를 아는 쪽은 래퍼이지 이 패키지가 아니기 때문입니다.

**3층 — 드라이버 작성자와 테스트.**

| 옵션 | 설명 |
|---|---|
| `WithEnv(Env)` | 임의의 `Env` 구현으로 환경 지정. `WithEnviron`의 일반형 |
| `WithEnv(LookupFunc(f))` | 조회 함수(`os.LookupEnv` 등)로 환경 지정 |
| `WithDrivers(...Driver)` | 기본 세트(내장 + `Register`된 것)에 **추가**. 같은 식별자는 교체 |
| `WithOnlyDrivers(...Driver)` | **딱 이것만** 실행. 내장도 등록된 것도 무시. 인자가 없으면 전부 비활성화 |

드라이버를 넘기는 옵션은 둘입니다. 확장이 목적이면 `WithDrivers`, 내장을 **배제**하는 것이 목적이면 `WithOnlyDrivers`입니다.

| 하고 싶은 일 | 방법 |
|---|---|
| 프로그램 전체에 드라이버 추가 | `Register(d)` (`init`에서) |
| 이 호출에만 드라이버 추가 | `WithDrivers(d)` |
| 내장 하나를 이 호출에서만 교체 | `WithDrivers(d)` — 식별자가 같으면 교체됩니다 |
| 내장 하나 끄기 (이 호출만) | 같은 식별자 + 절대 매치 안 하는 `Detect`를 `WithDrivers` |
| 내장 하나 끄기 (프로세스 전체) | 같은 식별자 + 절대 매치 안 하는 `Detect`를 `Register` |
| 드라이버 격리 테스트 | `WithOnlyDrivers(d)` |
| 내장 세트에서 하나만 빼고 실행 | `BuiltinDrivers()`를 필터해서 `WithOnlyDrivers` |

`WithDrivers`와 `WithOnlyDrivers` 모두 **한 호출 안에** 같은 축·같은 식별자를 둘 넘기면 panic합니다. `WithDrivers`가 기존 세트의 항목을 교체하는 것은 이와 별개이며 panic하지 않습니다.

CI·터미널 축은 **첫 매치가 이깁니다.** 커스텀 드라이버가 내장보다 우선해야 한다면 슬라이스 앞에 두십시오. 에이전트 축은 순서와 무관하게 항상 오케스트레이터 → 멀티 벤더 하네스 → 자사 모델 하네스 순으로 정렬되고, remote·runner 축은 매치 전부를 보고하므로 순서가 결과 순서일 뿐입니다.

### 무엇이 꺼지는가

`WithEnviron`과 `WithEnv`는 **이 프로세스의 것이 아닐 수도 있는** 환경을 넘기는 것이므로, 같은 프로세스를 설명해야만 의미가 있는 `TTY`와 `Process` 축은 자동으로 꺼집니다. 남의 환경에 이 프로세스의 파일 디스크립터와 조상 체인을 섞어 한 결과로 내놓으면, 그 결과는 어떤 프로세스도 설명하지 않습니다.

`WithOnlyDrivers`는 축별로 나뉘지 않습니다. `Driver` 인터페이스로 받아 드라이버가 속한 축에 알아서 배치하므로 한 번에 여러 축을 덮을 수 있고, 주지 않은 축은 꺼집니다. 용도는 두 가지입니다 — **드라이버를 격리해서 테스트**하는 것, 그리고 어딘가의 `_` 임포트가 `Register`한 드라이버에 좌우되지 않는 **결정적 테스트**를 쓰는 것.

```go
runby.Detect(runby.WithDrivers(myAgent, myRunner))      // 내장 + 이 둘
runby.Detect(runby.WithOnlyDrivers(myAgent, myRunner))  // 딱 이 둘만
runby.Detect(runby.WithOnlyDrivers())                   // 환경 기반 축 전부 off
```

`WithDrivers`도 축별로 나뉘지 않습니다. 넘긴 드라이버를 각자의 축에 배치하되, 나머지 축과 나머지 드라이버는 그대로 둡니다.

## Result

`Result`는 **프로세스 하나**에 대한 조사 결과입니다. `Terminal`은 프로세스당 하나뿐인 사실이므로 레이어가 아니라 `Result`에 있으며, 아무 에이전트도 감지되지 않아도 채워집니다.

```go
type Result struct {
	Agents   []Agent     // 가장 구체적인 오케스트레이터 → 하위 런타임 순
	TTY      TTY         // 표준 스트림 상태 (시스템콜 기반)
	Process  ProcessTree // 상위 프로세스 체인 (커널에서 읽음)
	CI       CI          // Agents와 독립된 축
	Terminal Terminal    // Agents와 독립된 축
	Remotes  []Remote    // 동시에 여러 계층이 존재할 수 있음
	Runners  []Runner    // 무엇이 직접 실행했는가. 중첩이 정상
}

// 축 술어 다섯은 이름을 맞춰 두었습니다.
result.IsAgent()                       // AI 에이전트가 실행했는가
result.IsCI()                          // CI 잡에서 도는가
result.HasTerminal()                   // 터미널 에뮬레이터를 식별했는가
result.IsRemote()                      // 낀 계층이 있는가
result.HasRunner()                     // 도구가 실행했는가 (스크립트·훅·서비스)

// 축을 합치는 유일한 메서드. 규칙이 doc comment에 못박혀 있습니다.
result.Unattended()                    // definite 에이전트 ∨ IsCI ∨ service 러너 ∨ (TTY 검사됨 ∧ 비대화형)

// 특정 제품이 계층에 있는지는 다른 질문입니다. 셋 다 축 이름을 그대로 쓰고
// (값, ok)를 돌려주므로, 존재 여부만 필요하면 값을 버리면 됩니다.
result.Agent(runby.AgentCodex)         // (Agent, bool)
result.Remote(runby.RemoteTmux)        // (Remote, bool)
result.Runner(runby.RunnerNPM)         // (Runner, bool)
result.RunnerOfKind(runby.RunnerKindService) // (Runner, bool) — 데몬인가

if _, ok := result.Remote(runby.RemoteSSH); ok {
	// SSH 계층이 있다
}

result.SessionID()                     // (Identifier, bool) — 값과 그 값을 광고한 에이전트
result.AgentID()                       // (Identifier, bool) — 논리적 에이전트 식별자, 같은 방식

result.Primary()                       // (Agent, bool) — 없으면 제로 Agent, ok로 판단
result.Unattended()                    // 아무도 출력을 보고 있지 않은가 (아래 참고)
result.Chain()                         // "paseo>codex", 감지 실패 시 "unknown"
result.Multiplexer()                   // (Remote, bool) — 잔존 위험의 주 원인
```

### `Identifier`

`SessionID()`와 `AgentID()`가 돌려주는 값입니다. 값과 **그 값을 광고한 에이전트**가 항상 함께 다닙니다 — 한 계층 스택에서 여러 제품이 같은 종류의 식별자를 광고할 수 있고, 그 식별자들은 서로 호환되지 않기 때문입니다.

```go
type Identifier struct {
	Value string    `json:"value"`
	Agent AgentName `json:"agent"`
}

if session, ok := result.SessionID(); ok {
	log.Printf("session=%s from=%s", session.Value, session.Agent)
}
```

### `Unattended()`

**이 패키지에서 축을 합치는 유일한 메서드입니다.** 나머지는 전부 축별로 따로 보고합니다 — 축은 서로 독립된 사실이기 때문입니다. 그래서 이 하나는 규칙을 doc comment와 테스트 양쪽에 못박아 두었습니다. 넷 중 하나라도 참이면 참입니다.

| 조건 | 이유 |
|---|---|
| `definite` 신뢰도의 에이전트 계층 | 에이전트도 PTY를 할당하므로 스트림은 대화형으로 보일 수 있지만, 뒤에 사람이 없습니다 |
| `IsCI()` | CI 로그는 나중에, 읽힌다면, 읽힙니다 |
| `RunnerKindService` 러너 | 서비스 관리자가 실행했으므로 출력은 저널로 갑니다 |
| `TTY.Inspected && !TTY.Interactive` | 스트림을 **검사했고** 프롬프트를 띄울 수 없습니다 |

첫 조건이 `IsAgent()`가 아니라는 점이 중요합니다. `probable`은 드라이버가 "제품이 환경을 소유하지만 타이핑하는 건 사람일 수 있다"고 말하는 방식입니다 — Orca가 소유한 페인, Cline이 연 터미널, 멀티플렉서 너머로 보인 (세션이 끝났을 수도 있는) 에이전트 마커. 사람이 기다리는 프롬프트를 꺼버리는 쪽이 더 나쁜 실수이므로, probable 계층만으로는 참이 되지 않습니다. 에이전트 흔적이 조금이라도 있으면 조용해지고 싶다면 `IsAgent()`를 직접 읽으십시오.

마지막 조건의 `Inspected`가 중요합니다. `WithEnviron`으로 만든 결과는 스트림을 읽은 적이 없고, **읽지 않은 TTY는 근거가 아닙니다.**

`Terminal` 축은 일부러 보지 않습니다. 지금 붙어 있는 에뮬레이터가 아니라 환경을 만든 에뮬레이터를 가리키므로, 누가 보고 있는지 답할 수 없습니다.

표시 방식을 정하는 기본값으로 쓰십시오. **신뢰 경계로 쓰면 안 됩니다.** 정책이 다르면 축을 직접 읽으십시오.

## Agent

```go
// Axis는 다섯 축의 결과가 공통으로 지니는 부분이며, 임베드되어 있습니다.
// 임베드이므로 직렬화 형태는 평평합니다 — JSON은 아래 필드를 그대로 갖습니다.
type Axis struct {
	Confidence Confidence
	Extra      map[string]string // 단일 제품 전용 값. 키는 "<slug>.<name>"
	Evidence   []string          // 감지에 사용한 변수 "이름"만
}

type Agent struct {
	Name   AgentName
	Kind   Kind        // orchestrator | harness — 무엇을 구동하는가
	Models ModelSource // first-party | multi-vendor | delegated — 지능의 출처
	Axis

	SessionID  string // 대화/스레드 식별자
	AgentID    string // 오케스트레이터의 논리적 에이전트 식별자
	Entrypoint string // "cli", "acp", "sidecar" 등 제품 자체 용어
	Nested     bool   // 최상위 세션이 아닌 자식 세션

	Sandbox Sandbox // {Mode string, Network Network}
	Paths   Paths   // {WorkingDirectory, DataDirectory}

	AncestorPID int // 살아 있는 조상으로 확인되면 그 PID
}
```

모든 필드에 JSON 태그가 있어 로그·텔레메트리로 그대로 직렬화할 수 있습니다.

`Axis`는 `CI`, `Terminal`, `Remote`, `Runner`에도 똑같이 임베드되어 있어, 어느 축의 결과든 `Confidence`·`Extra`·`Evidence`를 같은 이름으로 읽을 수 있습니다. `AncestorPID`는 `Axis`에 없습니다 — CI 잡은 이 프로세스가 파생된 프로세스가 아니라 러너 위의 작업이라 조상 체인에 대조할 대상이 없기 때문입니다.

`Extra`는 한 제품만 광고하는 값을 담아 공용 필드가 무한정 늘어나지 않게 합니다. 현재 키는 `codex.ci`와 `orca.*` 계열입니다.

이 패키지의 **모든 enum은 제로값을 `"unknown"`으로 렌더링합니다** — `AgentName`, `Kind`, `ModelSource`, `Confidence`, `Network`, `CIProvider`, `TerminalProgram`, `RemotePlatform`, `RemoteKind`, `RunnerTool`, `RunnerKind`. `Primary()`·`Agent()`·`Runner()`·`Remote()`가 미스에서 돌려주는 제로 구조체를 `ok` 확인 없이 그대로 로그에 찍어도 빈 칸이 아니라 `unknown`이 남습니다.

**`Evidence`에는 변수 이름만 들어갑니다.** 값은 민감할 수 있으므로 어떤 경우에도 복사하지 않습니다. 이 규칙은 다섯 축 전부에 동일하게 적용됩니다.

`AncestorPID`는 [`process.md`](process.md)를 참고하십시오. **0은 부정이 아닙니다.** `Terminal`, `Remotes[]`, `Runners[]`에도 같은 필드가 있습니다.

드라이버의 `Executables`를 채우면 조상 확증 대상에 포함됩니다. agent·terminal·remote·runner 네 축에서 동작하며, CI 드라이버에만 이 필드가 없습니다.

## 캐시된 진입점

프로세스 환경과 표준 스트림은 실무상 시작 시점에 고정되므로, 대부분의 호출자는 `Current()`만 쓰면 됩니다.

```go
result := runby.Current() // Detect()를 1회만 계산해 캐시
```

`Result`를 아래로 넘겨 쓰십시오. 축 질문은 전부 `Result`의 메서드입니다.

첫 호출 이후의 `os.Setenv`를 반영하려면 `Detect()`를 직접 부르십시오.

### 캐시와 테스트

캐시는 프로세스 전역이고 **가장 먼저 호출한 쪽이 채웁니다.** 그래서 `t.Setenv`로 환경을 꾸민 뒤 `Current()`를 부르면, 같은 테스트 바이너리에서 앞서 `Current()`를 건드린 테스트가 있었는지에 따라 결과가 달라집니다. 순서 의존 플레이크가 되기 쉬우므로, 테스트에서는 `Result`를 명시적으로 만드십시오.

```go
result := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
```

이 때문에 애플리케이션 코드도 `Current()`를 여기저기서 부르는 대신, 진입점에서 한 번 받아 `Result`를 인자로 넘기는 편이 테스트하기 쉽습니다.

## 드라이버 확장

축마다 드라이버 구조체가 하나씩 있고, 내장 제품과 사내 제품이 **같은 타입**을 씁니다. 드라이버는 감지 규칙과 함께, 환경이 알려줄 수 없는 사실(축위, 실행 파일 이름)까지 한 곳에 담습니다.

드라이버를 **모듈로 배포**해서 `_` 임포트로 쓰게 하려면 [`drivers.md`](drivers.md)를 보십시오. 아래 옵션은 그 `Detect` 호출 하나에만 적용되며, `Current()`와 CLI에는 반영되지 않습니다.

```go
acme := runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Kind:        runby.KindOrchestrator,
	Models:      runby.ModelsDelegated,
	Executables: []string{"acme-run"}, // 살아 있는 조상 프로세스로 확증
	Detect: func(env runby.Env) (runby.Agent, bool) {
		r := runby.NewEnvReader(env)
		id, ok := r.Value("ACME_RUN_ID")
		if !ok {
			return runby.Agent{}, false
		}
		return runby.Agent{AgentID: id, Axis: runby.Axis{Evidence: r.Evidence()}}, true
	},
}

result := runby.Detect(runby.WithDrivers(acme))

// 프로세스 전체에 한 번만 등록하려면 (init에서):
runby.Register(acme)
```

에이전트 축의 순서는 추가 순서가 아니라 **`Kind`와 `Models`에서 파생**됩니다 — 오케스트레이터, 그다음 멀티 벤더 하네스, 그다음 자사 모델 하네스 순. 사내 오케스트레이터는 그것이 구동한 런타임보다 언제나 앞서고, 둘 다 선언하지 않은 드라이버는 사다리에 자리가 없어 맨 뒤로 갑니다.

`Name`·`Kind`·`Models`는 드라이버가 제공하므로 `Detect` 안에서 다시 쓸 필요가 없습니다. `Confidence`와 `Sandbox.Network`는 비워두면 기본값이 채워집니다.

| 축 | 드라이버 | 식별자 | 축위 | 실행 파일 |
|---|---|---|---|---|
| agent | `AgentDriver` | `Agent` | `Kind` + `Models` | `Executables` |
| CI | `CIDriver` | `Provider` | — | — |
| terminal | `TerminalDriver` | `Program` | — | `Executables` |
| remote | `RemoteDriver` | `Platform` | `Kind` | `Executables` |
| runner | `RunnerDriver` | `Tool` | `Kind` | `Executables` |

CI 드라이버만 `Executables`가 없습니다. CI 잡은 이 프로세스가 이어받은 조상 프로세스가 아니라 러너 위의 작업이므로, 체인에서 대조할 대상이 없습니다.

remote 드라이버의 `Kind`를 `RemoteKindMultiplexer`로 두면 `Result.Multiplexer()`가 그 계층을 보고하고, 다른 모든 축이 낡았을 수 있다는 경고가 함께 적용됩니다. 비워두면 그 처리가 조용히 빠지므로 멀티플렉서라면 반드시 지정하십시오.

### `EnvReader` — 읽으면 근거가 따라옵니다

드라이버는 자신이 참조한 변수 이름을 `Evidence`에 보고해야 합니다. 이 목록을 손으로 유지하면 조용히 썩습니다. 조회를 하나 추가하고 목록 넓히기를 잊어도 **테스트는 전부 통과하고**, 감지 결과만 자기 근거를 축소 보고합니다.

`EnvReader`는 이 틈을 막습니다. 값을 물으면 이름이 기록되므로, `Evidence()`는 별도로 관리하는 두 번째 목록이 아니라 **실제로 읽은 것**입니다. 내장 에이전트 드라이버가 전부 이 방식으로 작성되어 있습니다.

```go
r := runby.NewEnvReader(env)
id, ok := r.Value("ACME_RUN_ID")     // 읽고 기록
if !ok {
	return runby.Agent{}, false      // 조기 반환. 기록만 하고 보고는 없음
}
home := r.First("ACME_HOME", "ACME_ROOT") // 선호 소스 + 폴백, 둘 다 기록
return runby.Agent{
	SessionID: id,
	Paths:     runby.Paths{DataDirectory: home},
	Axis:      runby.Axis{Evidence: r.Evidence()},
}, true
```

| 메서드 | 용도 |
|---|---|
| `Value(name)` | 변수 하나 읽기 (공백 트림, 빈 값은 근거 아님) |
| `Bool` / `IsTrue` / `EqualsFold` | 불리언·값 비교 |
| `Any(names...)` | 여러 후보 중 하나라도 설정됐는지. 전부 기록 |
| `First(names...)` | 가장 먼저 설정된 값. 전부 기록 |
| `Extra(keys)` | `Extra` 맵 구성. 참조한 변수를 전부 기록 |
| `Peek(name)` | **기록하지 않고** 읽기. 값이 특정할 때만 근거인 변수용 |
| `Record(names...)` | 읽지 않고 기록만 |
| `Evidence()` | 기록된 것 중 설정된 것만, 정렬·중복 제거해서 반환 |

`EnvReader`는 동시 사용에 안전하지 않습니다. `Detect` 호출마다 하나 만들고 호출 밖으로 넘기지 마십시오.

`Evidence`에는 값이 절대 들어가지 않습니다. `EnvReader.Evidence()`는 이름만 돌려줍니다.

### 드라이버가 환경을 바꿀 수 없는 이유

드라이버의 `Detect`는 `Env` 인터페이스(`Lookup(name) (string, bool)`)만 받습니다. `os.Environ()`을 직접 읽지 않으므로 사내 드라이버가 실수로 환경을 변경하거나 다른 프로세스의 환경을 섞을 수 없고, 같은 드라이버를 테스트에서 임의의 환경으로 그대로 실행할 수 있습니다.
