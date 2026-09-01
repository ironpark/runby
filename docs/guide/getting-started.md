# 시작하기

이 문서는 Go 프로그램에 `runby`를 처음 연결하고, 필요한 결과만 선택하고, 테스트에서 실행 환경을 재현하는 과정까지 안내합니다.

## 1. 설치하기

기존 Go 모듈에서 다음 명령을 실행합니다.

```sh
go get github.com/ironpark/runby
```

그리고 패키지를 임포트합니다.

```go
import "github.com/ironpark/runby"
```

셸 스크립트에서만 사용할 예정이라면 라이브러리 대신 CLI를 설치해도 됩니다.

```sh
go install github.com/ironpark/runby/cmd/runby@latest
runby
```

CLI의 서브명령과 종료 코드는 [CLI 가이드](cli.md)에 있습니다.

## 2. 현재 실행 환경 읽기

일반 애플리케이션에서는 `Current()`를 사용하세요. 첫 호출에 한 번 감지하고 같은 결과를 재사용합니다.

```go
result := runby.Current()

if result.IsAgent() {
	log.Printf("agent=%s", result.Chain())
}

if result.IsCI() {
	log.Printf("ci=%s", result.CI.Provider)
}
```

`Current()`의 결과에는 에이전트뿐 아니라 CI, 실행 도구, 원격 환경, 터미널, TTY, 상위 프로세스 정보도 함께 들어 있습니다.

## 3. 프로그램 동작 결정하기

사용자가 입력할 수 있는지가 중요하면 에이전트 이름이나 터미널 이름을 추측하지 말고 `TTY.Interactive`를 확인합니다.

```go
result := runby.Current()

if result.TTY.Interactive {
	confirmBeforeDelete()
} else {
	return errors.New("확인이 필요한 작업은 대화형 터미널에서 실행하세요")
}
```

에이전트와 CI에서 출력 형식을 바꾸고 싶다면 해당 축을 직접 확인합니다.

```go
result := runby.Current()

machineRun := result.IsAgent() || result.IsCI()
if _, service := result.RunnerOfKind(runby.RunnerKindService); service {
	machineRun = true
}

if machineRun {
	disableProgressAnimation()
}
```

상황별 권장 조건은 [활용 예제](recipes.md)에 더 정리되어 있습니다.

## 4. 필요한 상세 정보 선택하기

```go
result := runby.Current()

result.Chain()                              // "paseo>codex"
result.CI.Provider                         // "github-actions"
result.Terminal.Program                    // "ghostty"
result.Remote(runby.RemoteTmux)            // (Remote, bool)
result.Runner(runby.RunnerNPM)             // (Runner, bool)
```

특정 에이전트가 감지됐을 때만 그 제품의 필드를 읽을 수 있습니다.

```go
if codex, ok := result.Agent(runby.AgentCodex); ok {
	log.Printf("sandbox=%s network=%s", codex.Sandbox.Mode, codex.Sandbox.Network)
}
```

각 결과 구조체의 전체 필드는 [API 레퍼런스](api.md)에 있습니다.

## `Current()`와 `Detect()` 중 무엇을 쓸까요?

| 상황 | 권장 API |
|---|---|
| 현재 프로세스 정보를 여러 곳에서 읽음 | `Current()` |
| 현재 환경을 매번 새로 읽어야 함 | `Detect()` |
| 다른 프로세스나 기록된 환경을 분류함 | `Detect(WithEnviron(...))` |
| 환경변수 픽스처로 테스트함 | `Detect(WithEnviron(...))` |
| TTY나 프로세스 시스템콜을 생략함 | `Detect(WithoutTTY(), WithoutProcessTree())` |
| 커스텀 드라이버를 한 호출에서 격리함 | `Detect(WithOnlyDrivers(...))` |

`Current()`는 첫 호출 뒤 결과를 캐시합니다. 프로그램 시작 후 `os.Setenv`로 바꾼 값을 반영하려면 `Detect()`를 직접 호출하세요.

## 다른 환경을 넘기기

`WithEnviron`은 **이 프로세스가 아닌 환경**을 분류할 때 씁니다. 용도는 둘입니다.

하나는 다른 프로세스를 기술하는 경우입니다 — `/proc/<pid>/environ`을 읽었거나, `exec.Cmd.Env`를 만들어 두었거나, 환경을 파일로 기록해 두고 나중에 분석하는 래퍼가 여기에 해당합니다.

다른 하나는 실제 머신 환경과 분리된 결정적인 테스트입니다.

```go
func TestGitHubActions(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=1234",
	}))

	if !result.IsCI() || result.CI.Provider != runby.CIGitHubActions {
		t.Fatalf("unexpected result: %#v", result.CI)
	}
}
```

명시적 환경은 현재 프로세스의 환경이 아닐 수 있으므로 `WithEnviron`과 `WithEnv`는 TTY와 상위 프로세스를 자동으로 검사하지 않습니다. 필요한 테스트에서는 `WithTTY()`나 `WithProcessTree()`로 값을 직접 넣을 수 있습니다.

## 결과를 해석할 때 주의하기

환경변수는 프로세스가 시작될 때 상속됩니다. 따라서 `IsAgent()`가 참이라는 것은 해당 에이전트의 실행 신호를 물려받았다는 뜻이며, 에이전트가 현재도 살아 있다는 보장은 아닙니다.

감지 결과의 `AncestorPID`가 0이 아니면 환경 신호와 살아 있는 조상 프로세스가 함께 확인된 것입니다. `0`은 부정이 아니므로 실패 조건으로 사용하지 마세요. 자세한 내용은 [상위 프로세스 가이드](process.md)를 참고하세요.

다음 단계로 [활용 예제](recipes.md)를 살펴보거나, 필요한 [개념별 상세 문서](README.md#개념별-상세)로 이동하세요.
