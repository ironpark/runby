# API

진입점은 `Detect(opts ...Option) Result` 하나입니다.

```go
result := runby.Detect()                                   // 현재 프로세스
result := runby.Detect(runby.WithEnviron(environ))         // 명시적 환경
result := runby.Detect(runby.WithoutTTY())                 // TTY 시스템콜 생략
result := runby.Detect(runby.WithDetectors(myDetector))    // 사내 오케스트레이터 추가
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
| `WithDetectors(...Detector)` | 내장 agent detector 앞에 추가 |
| `WithOnlyDetectors(...Detector)` | 내장 agent detector를 완전히 대체 |
| `WithCIDetectors(...CIDetector)` | 내장 CI detector 앞에 추가 |
| `WithOnlyCIDetectors(...CIDetector)` | 내장 CI detector를 대체. 인자가 없으면 CI 감지 비활성화 |
| `WithTerminalDetectors(...TerminalDetector)` | 내장 터미널 detector 앞에 추가 |
| `WithOnlyTerminalDetectors(...)` | 내장 터미널 detector를 대체. 인자가 없으면 비활성화 |
| `WithRemoteDetectors(...RemoteDetector)` | 내장 remote detector 앞에 추가 |
| `WithOnlyRemoteDetectors(...)` | 내장 remote detector를 대체. 인자가 없으면 비활성화 |

`WithEnviron`·`WithEnv`·`WithLookup`은 **이 프로세스의 것이 아닐 수도 있는** 환경을 넘기는 것이므로, 같은 프로세스를 설명해야만 의미가 있는 `TTY`와 `Process` 축은 자동으로 꺼집니다.

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
}

result.Found()                  // AI 에이전트가 실행했는가
result.Agent()                  // 최상위 레이어의 Agent, 없으면 AgentUnknown
result.Primary()                // (Detection, bool)
result.Get(runby.AgentCodex)    // (Detection, bool)
result.Has(runby.AgentCodex)    // bool
result.Chain()                  // "paseo>codex", 감지 실패 시 "unknown"
result.IsCI()                   // CI 잡에서 도는가
result.IsTerminal()             // 터미널 에뮬레이터를 식별했는가
result.IsRemote()               // 낀 계층이 있는가
result.HasRemote(runby.RemoteSSH)
result.GetRemote(runby.RemoteTmux)
result.Multiplexer()            // (Remote, bool) — 잔존 위험의 주 원인
```

## Detection

```go
type Detection struct {
	Agent      Agent
	Kind       Kind
	Confidence Confidence

	SessionID  string  // 대화/스레드 식별자
	AgentID    string  // 오케스트레이터의 논리적 에이전트 식별자
	Entrypoint string  // "cli", "acp", "sidecar" 등 제품 자체 용어
	Nested     bool    // 최상위 세션이 아닌 자식 세션

	Sandbox Sandbox           // {Mode string, Network Network}
	Paths   Paths             // {WorkingDirectory, DataDirectory}
	Extra   map[string]string // 단일 에이전트 전용 값. 키는 "<slug>.<name>"

	Evidence    []string // 감지에 사용한 변수 "이름"만
	AncestorPID int      // 살아 있는 조상으로 확인되면 그 PID
}
```

모든 필드에 JSON 태그가 있어 로그·텔레메트리로 그대로 직렬화할 수 있습니다.

`Extra`는 한 에이전트만 광고하는 값을 담아 공용 필드가 무한정 늘어나지 않게 합니다. 현재 키는 `codex.ci`와 `orca.*` 계열입니다.

**`Evidence`에는 변수 이름만 들어갑니다.** 값은 민감할 수 있으므로 어떤 경우에도 복사하지 않습니다. 이 규칙은 네 축 전부에 동일하게 적용됩니다.

`AncestorPID`는 [`process.md`](process.md)를 참고하십시오. **0은 부정이 아닙니다.** `Terminal`과 `Remote[]`에도 같은 필드가 있습니다.

검출기가 `Executables() []string`을 구현하면 조상 확증 대상에 포함됩니다. 이 선택적 인터페이스는 네 축 모두에서 동작합니다.

## 캐시된 진입점

프로세스 환경과 표준 스트림은 실무상 시작 시점에 고정되므로, 대부분의 호출자는 캐시된 진입점을 쓰면 됩니다.

```go
runby.Current()    // Detect()를 1회만 계산해 캐시
runby.IsAgent()    // Current().Found()
runby.IsCI()       // Current().CI.Detected
runby.IsTerminal() // Current().Terminal.Detected
runby.IsRemote()   // len(Current().Remote) > 0
```

첫 호출 이후의 `os.Setenv`를 반영하려면 `Detect()`를 직접 부르십시오.

## detector 확장

```go
detector := runby.NewDetector("acme-orchestrator", func(env runby.Env) (runby.Detection, bool) {
	id, ok := runby.Value(env, "ACME_RUN_ID")
	if !ok {
		return runby.Detection{}, false
	}
	return runby.Detection{
		Kind:     runby.KindOrchestrator,
		AgentID:  id,
		Evidence: runby.PresentNames(env, "ACME_RUN_ID"),
	}, true
})

result := runby.Detect(runby.WithDetectors(detector))
```

`WithDetectors`로 추가한 detector는 내장 detector보다 앞서므로, 사내 오케스트레이터가 그것이 구동한 런타임보다 우선해 보고됩니다.

`Value`, `Bool`, `IsTrue`, `EqualsFold`, `PresentNames`는 내장 detector와 동일한 파싱 규칙을 재사용하도록 공개되어 있습니다. `Agent`, `Kind`, `Confidence`, `Sandbox.Network`를 비워두면 `Detect`가 기본값을 채웁니다.

CI·터미널·remote 축도 각각 `NewCIDetector`, `NewTerminalDetector`, `NewRemoteDetector`로 같은 방식으로 확장합니다.

### detector가 환경을 바꿀 수 없는 이유

detector는 `Env` 인터페이스(`Lookup(name) (string, bool)`)만 받습니다. `os.Environ()`을 직접 읽지 않으므로 사내 detector가 실수로 환경을 변경하거나 다른 프로세스의 환경을 섞을 수 없고, 같은 detector를 테스트에서 임의의 환경으로 그대로 실행할 수 있습니다.
