# 활용 예제

`runby` 결과를 애플리케이션 동작에 연결할 때 자주 쓰는 패턴입니다. 필요한 예제만 골라 적용하세요.

## 대화형 기능을 안전하게 결정하기

프롬프트를 띄울 수 있는지는 `TTY.Interactive`가 가장 직접적으로 답합니다.

```go
result := runby.Current()
if !result.TTY.Interactive {
	return errors.New("이 작업은 대화형 입력이 필요합니다")
}

askForConfirmation()
```

`result.IsAgent()`만으로 프롬프트 가능 여부를 정하지 마세요. 에이전트도 PTY를 할당할 수 있고, 사람이 실행한 명령도 파이프나 서비스 안에서는 TTY가 없을 수 있습니다.

반대로 “아무도 이 출력을 보고 있지 않은가?”가 중요하면 — 스피너, 색상, 진행률 표시를 켤지 정할 때 — `Unattended()`를 쓰십시오. 이 패키지에서 축을 합치는 **유일한** 메서드이고, 그래서 규칙이 doc comment에 못박혀 있습니다.

```go
result := runby.Current()
if result.Unattended() {
	disableSpinner()
	disableColor()
}
```

`Unattended()`가 참이 되는 조건은 넷입니다 — `definite` 신뢰도의 에이전트 계층, `IsCI()`, `RunnerKindService` 러너, 그리고 **검사된** 표준 스트림이 대화형이 아닌 경우. `probable` 계층은 사람이 타이핑하고 있을 가능성을 남기므로 그것만으로는 참이 되지 않습니다. `TTY.Inspected`가 거짓이면(예: `WithEnviron`으로 만든 결과) TTY 조건은 발동하지 않습니다. 읽지 않은 TTY는 근거가 아니기 때문입니다. `Terminal` 축은 보지 않습니다 — 지금 붙어 있는 에뮬레이터가 아니라 환경을 만든 에뮬레이터를 가리키므로 누가 보고 있는지 답할 수 없습니다.

정책이 다르면 축을 직접 조합하십시오. 예를 들어 git 훅까지 자동 실행으로 치고 싶다면:

```go
automated := result.Unattended()
if _, ok := result.RunnerOfKind(runby.RunnerKindHook); ok {
	automated = true
}
```

어느 쪽이든 알려진 자동 실행 신호를 모은 것이지, `false`일 때 사람이 직접 실행했다고 증명하는 값은 아닙니다. cron이나 일반 git 훅처럼 환경변수만으로 식별할 수 없는 실행 방식도 있습니다.

## 에이전트별 동작 바꾸기

에이전트가 하나라도 있는지만 확인하려면 `IsAgent()`, 특정 제품의 상세 정보가 필요하면 `Agent()`를 사용합니다.

```go
result := runby.Current()

if result.IsAgent() {
	disableSpinner()
}

if codex, ok := result.Agent(runby.AgentCodex); ok &&
	codex.Sandbox.Network == runby.NetworkDisabled {
	skipNetworkChecks()
}
```

중첩된 실행에서는 여러 에이전트가 함께 감지될 수 있습니다. `result.Primary()`는 가장 바깥의 대표 계층을 `(Agent, bool)`로, `result.Chain()`은 전체 계층을 반환합니다.

## 실행 맥락을 로그로 남기기

운영 로그에는 안정적인 식별자만 골라 기록하는 편이 좋습니다.

```go
result := runby.Current()

log.Printf(
	"agent=%s ci=%s interactive=%t remote=%t runner=%t",
	result.Chain(),
	result.CI.Provider,
	result.TTY.Interactive,
	result.IsRemote(),
	result.HasRunner(),
)
```

`Chain()`은 감지 실패 시 빈 문자열 대신 `"unknown"`을 반환합니다. `CI.Provider`도 감지되지 않으면 `"unknown"`이므로 로그 필드를 일정하게 유지할 수 있습니다.

전체 `Result`를 JSON으로 직렬화할 수도 있지만 세션 ID, 작업 디렉터리, 실행 파일 전체 경로가 포함될 수 있습니다. 외부 텔레메트리에는 필요한 필드만 선택하세요.

```go
type executionContext struct {
	Agent       string `json:"agent"`
	CI          string `json:"ci"`
	Interactive bool   `json:"interactive"`
}

context := executionContext{
	Agent:       result.Chain(),
	CI:          result.CI.Provider.String(),
	Interactive: result.TTY.Interactive,
}
```

## CI와 로컬 실행 구분하기

```go
result := runby.Current()

if result.IsCI() {
	configureForCI(result.CI.Provider, result.CI.Attempt)
} else {
	configureForLocalRun()
}
```

AI 에이전트가 CI 잡 안에서 실행되면 `IsAgent()`와 `IsCI()`가 둘 다 참입니다. 둘을 상호 배타적인 모드처럼 `switch`로 처리하지 말고 필요한 동작을 각각 적용하세요.

## 스크립트·훅·서비스 구분하기

```go
result := runby.Current()

if npm, ok := result.Runner(runby.RunnerNPM); ok {
	log.Printf("npm script=%s", npm.Task)
}

if _, ok := result.RunnerOfKind(runby.RunnerKindService); ok {
	disableHumanOrientedOutput()
}
```

`Runners`에는 여러 항목이 동시에 들어갈 수 있습니다. 예를 들어 pre-commit이 npm 스크립트를 부르고 그 스크립트가 make를 실행하면 세 도구가 모두 감지될 수 있습니다. 배열 순서는 중첩 순서를 증명하지 않습니다.

## tmux 안에서 낡은 환경 감지하기

```go
result := runby.Current()

if mux, ok := result.Multiplexer(); ok {
	log.Printf("%s 안에서는 상속된 환경 정보가 오래됐을 수 있음", mux.Platform)
}
```

멀티플렉서 서버는 이전 클라이언트의 환경을 오래 유지할 수 있습니다. `runby`는 터미널 신뢰도를 낮춰 표시하지만, 에이전트와 CI를 포함한 다른 환경변수도 낡을 수 있으므로 중요한 판단에서는 [상위 프로세스 정보](process.md)를 함께 사용하세요.

## 셸 스크립트에서 분기하기

CLI의 `is` 명령은 아무것도 출력하지 않고 종료 코드로 답합니다.

```sh
if runby is agent || runby is ci; then
	export NO_COLOR=1
fi

if runby is runner; then
	echo "다른 도구가 실행한 작업입니다"
fi
```

공유 가능한 진단 보고서는 `runby -v`, 자동 처리할 데이터는 `runby -json`을 사용하세요. JSON의 개인정보 주의사항과 종료 코드 전체는 [CLI 가이드](cli.md)에 있습니다.
