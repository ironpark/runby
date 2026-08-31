# runby

`runby`는 현재 프로세스를 **무엇이 실행했는지** 감지하는 Go 패키지입니다. 코딩 에이전트, CI 잡, 터미널 에뮬레이터, 그 사이에 낀 원격 계층을 구분해서 보고합니다.

```
go get github.com/ironpark/runby
```

```go
import "github.com/ironpark/runby"

if runby.IsAgent() {
	log.Printf("run by %s", runby.Current().Chain()) // "paseo>codex"
	disableInteractivePrompts()
}
if runby.IsCI() {
	log.Printf("ci: %s", runby.Current().CI.Provider) // "github-actions"
}
```

**표준 라이브러리만 사용합니다.** `go.mod`에 외부 모듈이 없고 `go.sum`도 없습니다.

## 여섯 개의 독립된 축

에이전트가 CI 안에서, 또 터미널 안에서 실행될 수 있으므로 여러 축이 동시에 참일 수 있습니다.

| 축 | 질문 | 필드 | 문서 |
|---|---|---|---|
| 에이전트 | 누가 명령을 요청했는가 | `Layers` | [agents](docs/guide/agents.md) |
| CI | 어떤 CI 잡에서 실행되는가 | `CI` | [ci](docs/guide/ci.md) |
| 터미널 | 어떤 에뮬레이터가 이 환경을 만들었는가 | `Terminal` | [terminal](docs/guide/terminal.md) |
| 원격·멀티플렉서 | 사용자와 이 프로세스 사이에 무엇이 끼어 있는가 | `Remote` | [remote](docs/guide/remote.md) |
| 실행 주체 | 무엇이 이 프로세스를 직접 실행했는가 | `Runner` | [runner](docs/guide/runner.md) |
| 프로세스 | 지금 살아 있는 조상은 무엇인가 | `Process` | [process](docs/guide/process.md) |

앞의 다섯 축은 상속된 환경변수를 읽고, `Process`는 커널에서 상위 프로세스 체인을 읽습니다. 여기에 표준 스트림이 터미널인지 시스템콜로 확인하는 `TTY`가 더해집니다.

```go
result := runby.Detect()

result.IsAgent()               // AI 에이전트가 실행했는가
result.Chain()                 // "paseo>codex"
result.CI.Provider             // "github-actions"
result.Terminal.Program        // "ghostty"
result.Multiplexer()           // (Remote, bool) — tmux 등
result.HasRunner()             // npm 스크립트·make·systemd가 실행했는가
result.TTY.Interactive         // 프롬프트를 띄울 수 있는가
```

전체 API는 [`docs/guide/api.md`](docs/guide/api.md)에 있습니다.

## 드라이버 확장

지원하지 않는 제품은 드라이버로 추가합니다. 내장 제품과 **같은 타입**을 쓰고, `init`에서 등록하면 `_` 임포트만으로 프로그램 전체에 적용됩니다.

```go
// example.com/runby-acme
func init() {
	runby.Register(runby.AgentDriver{
		Agent: "acme", Kind: runby.KindOrchestrator, Models: runby.ModelsDelegated,
		Detect: func(env runby.Env) (runby.Detection, bool) { … },
	})
}
```

```go
import _ "example.com/runby-acme"  // 이것만으로 runby.IsAgent()가 acme를 압니다
```

자세한 내용과 주의점은 [`docs/guide/drivers.md`](docs/guide/drivers.md)에 있습니다.

## CLI

라이브러리가 먼저지만, 셸에서 바로 쓸 수 있는 명령도 있습니다.

```
go install github.com/ironpark/runby/cmd/runby@latest
```

```sh
runby                 # 사람이 읽는 요약
runby -json           # Result 전체 JSON
runby chain           # "paseo>codex"

if runby is agent; then export NO_COLOR=1; fi    # 종료 코드로만 답합니다
```

자세한 사용법은 [`docs/guide/cli.md`](docs/guide/cli.md)에 있습니다.

## 감지 대상

**에이전트** — Paseo, Orca, OpenAI Codex, Claude Code, Antigravity 2.0, Amp, Cursor Agent, OpenCode ACP

**CI** — Forgejo Actions, GitHub Actions, GitLab CI/CD, CircleCI, Travis CI, Buildkite, Azure Pipelines, Bitbucket Pipelines, Jenkins, 그 외 `CI=true`

**터미널** — iTerm2, Apple Terminal, WezTerm, Ghostty, Warp, Zed, kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal, VTE 계열

**원격·멀티플렉서** — tmux, GNU Screen, Zellij, OpenSSH, WSL, GitHub Codespaces, Gitpod, Dev Containers

목록에 없는 제품은 [드라이버](docs/guide/drivers.md)를 만들어 `Register`로 추가하면 지원됩니다. 지원하지 않기로 한 제품과 그 이유는 [`docs/research/`](docs/research/)에 제품별로 기록되어 있습니다.

## 알아야 할 한 가지

**환경변수는 프로세스가 시작될 때 상속된 스냅샷입니다.** 감지 성공은 "이 프로세스가 시작될 당시 그 에이전트가 활성 상태였다"는 뜻이지, "지금도 살아 있다"는 뜻이 아닙니다.

그래서 `runby`는 축마다 신뢰도를 구분해서 보고합니다. `Confidence`는 신호가 얼마나 직접적인지를, `Process` 축은 **지금** 참인 것을 말합니다. 에이전트가 감지되고 그 실행 파일이 조상으로 살아 있으면 `Detection.AncestorPID`가 채워집니다 — 이 패키지가 제공할 수 있는 가장 강한 확증입니다.

```go
for _, layer := range runby.Current().Layers {
	if layer.AncestorPID != 0 {
		// 환경변수가 이 에이전트를 말했고, 그 에이전트가 지금도 조상으로 살아 있음
	}
}
```

`AncestorPID == 0`은 **부정이 아닙니다.** 자세한 내용은 [`docs/guide/process.md`](docs/guide/process.md)에 있습니다.

## 문서

- [사용자 문서](docs/guide/) — 축별 상세, API 레퍼런스
- [조사 문서](docs/research/) — 각 판정을 어떤 공식 문서·소스에서 확인했는지

## 플랫폼

`Process` 축은 Linux·macOS·Windows에서 동작하고, 그 외에서는 `Supported == false`로 빈 체인을 반환합니다. `TTY`는 AIX·Solaris·z/OS에서 항상 `false`입니다 — 표준 `syscall`이 이 플랫폼들에 `TCGETS`·`SYS_IOCTL`을 노출하지 않기 때문입니다. 환경변수를 읽는 다섯 축은 모든 플랫폼에서 동일하게 동작합니다.

TTY 검사는 [`internal/term`](internal/term/)에, 프로세스 체인 읽기는 [`internal/proc`](internal/proc/)에 있습니다. `golang.org/x/term`의 `IsTerminal` 하나만 옮겨와 `golang.org/x/sys` 대신 표준 `syscall`로 다시 작성했으며, 원본의 BSD-3-Clause 라이선스를 함께 보관합니다.
