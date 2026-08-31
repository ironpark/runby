# API

진입점은 `Detect(opts ...Option) Result` 하나입니다.

```go
result := runby.Detect()                                   // 현재 프로세스
result := runby.Detect(runby.WithEnviron(environ))         // 명시적 환경
result := runby.Detect(runby.WithoutTTY())                 // TTY 시스템콜 생략
runby.Register(myDriver)                                   // 사내 드라이버, 프로세스 전체
```

## 옵션

| 옵션 | 설명 |
|---|---|
| `WithEnviron([]string)` | `"NAME=value"` 슬라이스로 환경 지정 |
| `WithEnv(Env)` / `WithLookup(func)` | 임의의 조회 함수로 환경 지정 |
| `WithoutTTY()` | 표준 스트림 검사 생략 |
| `WithTTY(TTY)` | 표준 스트림 상태를 직접 주입 |
| `WithoutProcessTree()` | 상위 프로세스 체인 읽기 생략 |
| `WithProcessTree(ProcessTree)` | 상위 프로세스 체인을 직접 주입 |
| `WithOnlyDrivers(...Driver)` | **딱 이것만** 실행. 내장도 등록된 것도 무시. 인자가 없으면 전부 비활성화 |

드라이버를 넘기는 옵션은 `WithOnlyDrivers` 하나입니다. 축별 `With*Drivers` 다섯 개는 제거되었고, 그 자리는 `BuiltinDrivers()`와의 조합이 대신합니다 — 순서가 숨지 않고 슬라이스에 드러납니다.

| 하고 싶은 일 | 방법 |
|---|---|
| 프로그램 전체에 드라이버 추가 | `Register(d)` (`init`에서) |
| 이 호출에만 내장 + 커스텀 | `WithOnlyDrivers(append(BuiltinDrivers(), d)...)` |
| 커스텀을 내장보다 **앞**에 | `WithOnlyDrivers(append([]Driver{d}, BuiltinDrivers()...)...)` |
| 내장 하나 끄기 | `BuiltinDrivers()`를 필터해서 `WithOnlyDrivers` |
| 드라이버 격리 테스트 | `WithOnlyDrivers(d)` |

CI·터미널 축은 **첫 매치가 이깁니다.** 커스텀 드라이버가 내장보다 우선해야 한다면 슬라이스 앞에 두십시오. 에이전트 축은 순서와 무관하게 항상 사다리(`l3`→`l2`→`l1`)로 정렬되고, remote·runner 축은 매치 전부를 보고하므로 순서가 결과 순서일 뿐입니다.

`WithEnviron`·`WithEnv`·`WithLookup`은 **이 프로세스의 것이 아닐 수도 있는** 환경을 넘기는 것이므로, 같은 프로세스를 설명해야만 의미가 있는 `TTY`와 `Process` 축은 자동으로 꺼집니다.

`WithOnlyDrivers`는 축별로 나뉘지 않습니다. `Driver` 인터페이스로 받아 드라이버가 속한 축에 알아서 배치하므로 한 번에 여러 축을 덮을 수 있고, 주지 않은 축은 꺼집니다. 용도는 두 가지입니다 — **드라이버를 격리해서 테스트**하는 것, 그리고 어딘가의 `_` 임포트가 `Register`한 드라이버에 좌우되지 않는 **결정적 테스트**를 쓰는 것.

```go
runby.Detect(runby.WithOnlyDrivers(myAgent, myRunner))  // 딱 이 둘만
runby.Detect(runby.WithOnlyDrivers())                    // 환경 기반 축 전부 off
```

## Result

`Result`는 **프로세스 하나**에 대한 조사 결과입니다. `Terminal`은 프로세스당 하나뿐인 사실이므로 레이어가 아니라 `Result`에 있으며, 아무 에이전트도 감지되지 않아도 채워집니다.

```go
type Result struct {
	Layers   []Detection // 가장 구체적인 오케스트레이터 → 하위 런타임 순
	TTY      TTY         // 표준 스트림 상태 (시스템콜 기반)
	Process  ProcessTree // 상위 프로세스 체인 (커널에서 읽음)
	CI       CI          // Layers와 독립된 축
	Terminal Terminal    // Layers와 독립된 축
	Remote   []Remote    // 동시에 여러 계층이 존재할 수 있음
	Runner   []Runner    // 무엇이 직접 실행했는가. 중첩이 정상
}

// 축 술어 다섯은 이름을 맞춰 두었습니다.
result.IsAgent()                       // AI 에이전트가 실행했는가
result.IsCI()                          // CI 잡에서 도는가
result.HasTerminal()                   // 터미널 에뮬레이터를 식별했는가
result.IsRemote()                      // 낀 계층이 있는가
result.HasRunner()                     // 도구가 실행했는가 (스크립트·훅·서비스)

// 특정 제품이 계층에 있는지는 다른 질문이라 이름도 다릅니다.
result.Layer(runby.AgentCodex)         // (Detection, bool)
result.HasLayer(runby.AgentCodex)      // bool
result.RemoteLayer(runby.RemoteTmux)   // (Remote, bool)
result.HasRemoteLayer(runby.RemoteSSH) // bool
result.RunnerBy(runby.RunnerNPM)       // (Runner, bool)
result.HasRunnerBy(runby.RunnerMake)   // bool
result.RunnerOfKind(runby.RunnerKindService) // (Runner, bool) — 데몬인가

result.Agent()                         // 최상위 레이어의 Agent, 없으면 AgentUnknown
result.Primary()                       // (Detection, bool)
result.Chain()                         // "paseo>codex", 감지 실패 시 "unknown"
result.Multiplexer()                   // (Remote, bool) — 잔존 위험의 주 원인
```

## Detection

```go
// Axis는 다섯 축의 결과가 공통으로 지니는 부분이며, 임베드되어 있습니다.
// 임베드이므로 직렬화 형태는 평평합니다 — JSON은 아래 필드를 그대로 갖습니다.
type Axis struct {
	Confidence Confidence
	Extra      map[string]string // 단일 제품 전용 값. 키는 "<slug>.<name>"
	Evidence   []string          // 감지에 사용한 변수 "이름"만
}

type Detection struct {
	Agent  Agent
	Kind   Kind        // orchestrator | harness — 무엇을 구동하는가
	Models ModelSource // first-party | multi-vendor | delegated — 지능의 출처
	Level  Level       // l1 | l2 | l3 — 위 둘에서 파생된 사다리
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

**`Evidence`에는 변수 이름만 들어갑니다.** 값은 민감할 수 있으므로 어떤 경우에도 복사하지 않습니다. 이 규칙은 다섯 축 전부에 동일하게 적용됩니다.

`AncestorPID`는 [`process.md`](process.md)를 참고하십시오. **0은 부정이 아닙니다.** `Terminal`, `Remote[]`, `Runner[]`에도 같은 필드가 있습니다.

드라이버의 `Executables`를 채우면 조상 확증 대상에 포함됩니다. agent·terminal·remote 세 축에서 동작합니다.

## 캐시된 진입점

프로세스 환경과 표준 스트림은 실무상 시작 시점에 고정되므로, 대부분의 호출자는 캐시된 진입점을 쓰면 됩니다.

```go
runby.Current()     // Detect()를 1회만 계산해 캐시
runby.IsAgent()     // Current().IsAgent()
runby.IsCI()        // Current().IsCI()
runby.HasTerminal() // Current().HasTerminal()
runby.IsRemote()    // Current().IsRemote()
runby.HasRunner()   // Current().HasRunner()
```

첫 호출 이후의 `os.Setenv`를 반영하려면 `Detect()`를 직접 부르십시오.

## 드라이버 확장

축마다 드라이버 구조체가 하나씩 있고, 내장 제품과 사내 제품이 **같은 타입**을 씁니다. 드라이버는 감지 규칙과 함께, 환경이 알려줄 수 없는 사실(축위, 실행 파일 이름)까지 한 곳에 담습니다.

드라이버를 **모듈로 배포**해서 `_` 임포트로 쓰게 하려면 [`drivers.md`](drivers.md)를 보십시오. 아래 옵션은 그 `Detect` 호출 하나에만 적용되며, `runby.IsAgent()`·`Current()`·CLI에는 반영되지 않습니다.

```go
acme := runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Kind:        runby.KindOrchestrator,
	Models:      runby.ModelsDelegated,
	Executables: []string{"acme-run"}, // 살아 있는 조상 프로세스로 확증
	Detect: func(env runby.Env) (runby.Detection, bool) {
		id, ok := runby.Value(env, "ACME_RUN_ID")
		if !ok {
			return runby.Detection{}, false
		}
		return runby.Detection{AgentID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_RUN_ID")}}, true
	},
}

result := runby.Detect(runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), acme)...))

// 프로세스 전체에 한 번만 등록하려면 (init에서):
runby.Register(acme)
```

에이전트 축의 순서는 추가 순서가 아니라 **`Kind`와 `Models`에서 파생된 `Level`**로 정해집니다(`l3` → `l2` → `l1`). 사내 오케스트레이터는 그것이 구동한 런타임보다 언제나 앞서고, 둘 다 선언하지 않은 드라이버는 사다리에 자리가 없어 맨 뒤로 갑니다.

`Agent`·`Kind`·`Models`는 드라이버가 제공하므로 `Detect` 안에서 다시 쓸 필요가 없고, `Level`은 그 둘에서 내장 에이전트와 **같은 규칙으로** 파생됩니다. `Confidence`와 `Sandbox.Network`는 비워두면 기본값이 채워집니다.

| 축 | 드라이버 | 식별자 | 축위 | 실행 파일 |
|---|---|---|---|---|
| agent | `AgentDriver` | `Agent` | `Kind` + `Models` | `Executables` |
| CI | `CIDriver` | `Provider` | — | — |
| terminal | `TerminalDriver` | `Program` | — | `Executables` |
| remote | `RemoteDriver` | `Platform` | `Kind` | `Executables` |
| runner | `RunnerDriver` | `Tool` | `Kind` | `Executables` |

CI 드라이버만 `Executables`가 없습니다. CI 잡은 이 프로세스가 이어받은 조상 프로세스가 아니라 러너 위의 작업이므로, 체인에서 대조할 대상이 없습니다.

remote 드라이버의 `Kind`를 `RemoteKindMultiplexer`로 두면 `Result.Multiplexer()`가 그 계층을 보고하고, 다른 모든 축이 낡았을 수 있다는 경고가 함께 적용됩니다. 비워두면 그 처리가 조용히 빠지므로 멀티플렉서라면 반드시 지정하십시오.

### 저작 헬퍼

내장 드라이버와 동일한 파싱 규칙을 재사용하도록 공개되어 있습니다.

| 헬퍼 | 용도 |
|---|---|
| `Value` / `Bool` / `IsTrue` / `EqualsFold` | 변수 하나 읽기 (공백 트림, 빈 값은 근거 아님) |
| `AnyPresent(env, names...)` | 여러 후보 중 하나라도 설정됐는지 |
| `MarkerSet` / `MarkerTrue` / `MarkerTermProgram` | `Marker`(= `func(Env) bool`) 생성 |
| `CollectExtra(env, keys)` | `Extra` 맵 구성. 설정되지 않은 키는 건너뜀 |
| `PresentNames(env, names...)` | `Evidence` 구성. 정렬·중복 제거, **이름만** |

`Evidence`에는 값이 절대 들어가지 않습니다. `PresentNames`에 후보 이름을 전부 넘기면 설정된 것만 남으므로, 값을 직접 다루지 않고 근거 목록을 만들 수 있습니다.

### 드라이버가 환경을 바꿀 수 없는 이유

드라이버의 `Detect`는 `Env` 인터페이스(`Lookup(name) (string, bool)`)만 받습니다. `os.Environ()`을 직접 읽지 않으므로 사내 드라이버가 실수로 환경을 변경하거나 다른 프로세스의 환경을 섞을 수 없고, 같은 드라이버를 테스트에서 임의의 환경으로 그대로 실행할 수 있습니다.
