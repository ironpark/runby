# runby

`runby`는 현재 프로세스를 실행한 코딩 에이전트 또는 오케스트레이터를 상속된 환경변수로 감지하는 Go 패키지입니다.

```go
import "github.com/ironpark/runby"

if runby.IsAgent() {
	log.Printf("run by %s", runby.Current().Chain()) // "paseo>codex"
	disableInteractivePrompts()
}
```

API 키와 일반 설정 변수는 에이전트가 프로세스를 실행했다는 증거가 아니므로 감지에 사용하지 않습니다.

## 감지 대상

| 에이전트 | `Agent` | `Kind` | 식별 신호 |
|---|---|---|---|
| Paseo | `AgentPaseo` | `orchestrator` | `PASEO_AGENT_ID`, `PASEO_AGENT_CWD` |
| OpenAI Codex | `AgentCodex` | `harness` | `CODEX_THREAD_ID`, `CODEX_SESSION_ID`, 샌드박스 관련 변수 |
| Claude Code | `AgentClaudeCode` | `harness` | `CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`, `AI_AGENT=claude-code*` |
| Antigravity 2.0 sidecar | `AgentAntigravity2` | `harness` | `ANTIGRAVITY_EXECUTABLE_DATA_DIR` |
| Amp Orb / Orb 관리형 서비스 | `AgentAmp` | `harness` | `AMP_ORB`, `AMP_THREAD_ID` |
| Cursor Agent | `AgentCursor` | `harness` | `CURSOR_AGENT` |
| OpenCode ACP | `AgentOpenCode` | `harness` | `OPENCODE_CLIENT=acp` |
| Zed 터미널 | `AgentZed` | `host` | `ZED_TERM=true`, `TERM_PROGRAM=zed` |

Antigravity CLI, GitHub Copilot CLI, Junie, 일반 OpenCode CLI는 공식적으로 확인된 범용 자식 프로세스 실행 마커가 없어 감지하지 않습니다. 자세한 조사 근거는 [`docs/`](docs/)에 있습니다.

### Kind: 에이전트인가, 호스트인가

`Kind`는 감지 결과가 무엇을 증명하는지를 구분합니다.

- `KindOrchestrator` / `KindHarness` — **AI 에이전트가** 이 프로세스를 실행했다는 증거입니다.
- `KindHost` — 어떤 애플리케이션이 터미널을 소유하는지만 증명합니다. Zed는 Agent 전용 신호가 없어 여기에 속하며, 사람이 직접 명령을 실행했을 수도 있습니다.

따라서 "AI가 실행했는가"를 판단할 때는 `Found()`가 아니라 `IsAgent()`를 쓰십시오.

```go
result := runby.Detect()
result.Found()   // Zed 터미널만 있어도 true
result.IsAgent() // Zed 터미널만 있으면 false
```

`Confidence`는 신호의 직접성을 구분합니다. `ConfidenceDefinite`는 제품이 자신이 실행한 프로세스에 한해 설정하는 실행 마커이고, `ConfidenceProbable`은 에이전트 실행과 모순되지 않지만 그것만의 신호는 아닌 보조 신호입니다.

## API

진입점은 `Detect(opts ...Option) Result` 하나입니다.

```go
result := runby.Detect()                                  // 현재 프로세스 (터미널 포함)
result := runby.Detect(runby.WithEnviron(environ))         // 명시적 환경
result := runby.Detect(runby.WithoutTerminal())            // TTY 시스템콜 생략
result := runby.Detect(runby.WithDetectors(myDetector))    // 사내 오케스트레이터 추가
```

| 옵션 | 설명 |
|---|---|
| `WithEnviron([]string)` | `"NAME=value"` 슬라이스로 환경 지정 |
| `WithEnv(Env)` / `WithLookup(func)` | 임의의 조회 함수로 환경 지정 |
| `WithoutTerminal()` | 터미널 검사 생략 |
| `WithTerminal(Terminal)` | 터미널 상태를 직접 주입 |
| `WithDetectors(...Detector)` | 내장 detector 앞에 추가 |
| `WithOnlyDetectors(...Detector)` | 내장 detector를 완전히 대체 |

### Result

`Result`는 **프로세스 하나**에 대한 조사 결과입니다. `Terminal`은 프로세스당 하나뿐인 사실이므로 레이어가 아니라 `Result`에 있으며, 아무 에이전트도 감지되지 않아도 채워집니다.

```go
type Result struct {
	Layers   []Detection // 가장 구체적인 오케스트레이터 → 하위 런타임 순
	Terminal Terminal
}

result.Found()                  // 감지된 레이어가 있는가 (host 포함)
result.IsAgent()                // AI 에이전트 실행 증거가 있는가
result.Agent()                  // 최상위 레이어의 Agent, 없으면 AgentUnknown
result.Primary()                // (Detection, bool)
result.Get(runby.AgentCodex)    // (Detection, bool)
result.Has(runby.AgentCodex)    // bool
result.Chain()                  // "paseo>codex", 감지 실패 시 "unknown"
```

여러 계층이 함께 존재할 수 있습니다. Paseo가 Codex를 구동했다면 `Layers`에 둘 다 들어가고, 명시적인 에이전트 ID를 가진 Paseo가 `Primary()`가 됩니다.

```go
result := runby.Detect()
if codex, ok := result.Get(runby.AgentCodex); ok && codex.Sandbox.Network == runby.NetworkDisabled {
	skipNetworkTests()
}
```

### Detection

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

	Evidence []string // 감지에 사용한 변수 "이름"만
}
```

모든 필드에 JSON 태그가 있어 로그·텔레메트리로 그대로 직렬화할 수 있습니다.

`Extra`는 한 에이전트만 광고하는 값을 담아 공용 필드가 무한정 늘어나지 않게 합니다. 현재 키는 `codex.ci`, `zed.version`입니다.

`Evidence`에는 변수 이름만 들어갑니다. 값은 민감할 수 있으므로 어떤 경우에도 복사하지 않으며, `Status`에는 제품이 명시적으로 광고한 실행 메타데이터만 담습니다.

### Terminal

```go
terminal := runby.Detect().Terminal
terminal.Inspected   // 표준 스트림을 실제로 검사했는가
terminal.StdinTTY
terminal.StdoutTTY
terminal.StderrTTY
terminal.Attached    // 세 스트림 중 하나 이상이 터미널
terminal.Interactive // stdin과 출력 스트림 하나 이상이 터미널
```

`Interactive`는 프롬프트를 띄우고 응답을 받을 수 있는 형태라는 뜻일 뿐, 사용자가 직접 명령을 호출했다는 증거는 아닙니다. 에이전트나 서브에이전트도 PTY를 할당할 수 있습니다.

`WithEnviron` 계열 옵션은 이 프로세스의 것이 아닐 수도 있는 환경을 검사하므로 터미널을 검사하지 않으며 `Inspected`가 `false`입니다. 필요하면 `InspectTerminal()`을 직접 호출하거나 `WithTerminal()`로 주입하십시오.

### 캐시된 진입점

프로세스 환경과 표준 스트림은 실무상 시작 시점에 고정되므로, 대부분의 호출자는 캐시된 진입점을 쓰면 됩니다.

```go
runby.Current()  // Detect()를 1회만 계산해 캐시
runby.IsAgent()  // Current().IsAgent()
```

첫 호출 이후의 `os.Setenv`를 반영하려면 `Detect()`를 직접 부르십시오.

### detector 확장

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

`WithDetectors`로 추가한 detector는 내장 detector보다 앞서므로, 사내 오케스트레이터가 그것이 구동한 런타임보다 우선해 보고됩니다. `Value`, `Bool`, `IsTrue`, `EqualsFold`, `PresentNames`는 내장 detector와 동일한 파싱 규칙을 재사용하도록 공개되어 있습니다. `Agent`, `Kind`, `Confidence`, `Sandbox.Network`를 비워두면 `Detect`가 기본값을 채웁니다.

## 상태의 의미

환경변수는 프로세스가 시작될 때 상속된 스냅샷입니다. 감지 성공은 에이전트가 이 프로세스를 실행할 당시 활성 상태였다는 뜻입니다. 장시간 실행되는 프로세스에서 부모 에이전트가 현재도 살아 있는지 확인하려면 PID, IPC 또는 에이전트 API 같은 별도의 생존 확인 수단이 필요합니다.
