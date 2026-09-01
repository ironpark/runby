# runby

현재 프로세스를 **누가, 어디서, 어떤 방식으로 실행했는지** 감지하는 Go 패키지입니다.

AI 에이전트가 실행한 명령인지, CI 잡인지, 현재 스트림이 터미널에 연결됐는지, npm 스크립트나 systemd가 실행했는지를 하나의 결과로 확인할 수 있습니다. 외부 의존성 없이 표준 라이브러리만 사용합니다.

## 이럴 때 사용하세요

- AI 에이전트나 CI에서는 확인 프롬프트와 진행 애니메이션을 끄고 싶을 때
- 로그에 `paseo>codex`처럼 실제 실행 계층을 남기고 싶을 때
- npm·make·pre-commit·systemd가 실행한 작업을 직접 실행과 구분하고 싶을 때
- 버그 리포트에서 터미널, SSH, tmux, 상위 프로세스 정보를 한 번에 확인하고 싶을 때

## 설치

라이브러리로 사용하려면:

```sh
go get github.com/ironpark/runby
```

셸에서 사용할 CLI가 필요하면:

```sh
go install github.com/ironpark/runby/cmd/runby@latest
```

## 빠른 시작

대부분의 프로그램은 캐시된 현재 실행 정보인 `Current()`만 호출하면 됩니다.

```go
package main

import (
	"log"

	"github.com/ironpark/runby"
)

func main() {
	result := runby.Current()

	if result.IsAgent() {
		log.Printf("AI 에이전트가 실행함: %s", result.Chain())
	}
	if result.IsCI() {
		log.Printf("CI에서 실행 중: %s", result.CI.Provider)
	}
	if !result.TTY.Interactive {
		disableInteractivePrompts()
	}
}
```

환경이 바뀔 때마다 다시 검사해야 하거나 테스트용 환경을 넣어야 한다면 `Detect()`를 사용합니다.

```go
freshResult := runby.Detect()
testResult := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
```

더 자세한 설치·사용 흐름은 [시작하기](docs/guide/getting-started.md), 실제 적용 예제는 [활용 예제](docs/guide/recipes.md)를 참고하세요.

## 필요한 질문 고르기

| 알고 싶은 것 | 사용할 API | 대표 결과 |
|---|---|---|
| AI 에이전트가 실행했나? | `result.IsAgent()` | `true` |
| 어떤 에이전트 계층인가? | `result.Chain()` | `"paseo>codex"` |
| CI 잡 안인가? | `result.IsCI()` | `true` |
| npm·make·systemd 등이 실행했나? | `result.HasRunner()` | `true` |
| SSH·tmux·컨테이너 개발 환경을 거쳤나? | `result.IsRemote()` | `true` |
| 어떤 터미널 환경에서 시작됐나? | `result.HasTerminal()` / `result.Terminal` | `ghostty` |
| 지금 프롬프트를 사용할 수 있나? | `result.TTY.Interactive` | `true` |
| 아무도 출력을 보고 있지 않나? | `result.Unattended()` | `true` |
| 감지된 실행 주체가 현재도 조상인가? | `AncestorPID` | `2540` |

`Terminal`과 `TTY`는 다릅니다. `Terminal`은 환경을 만든 에뮬레이터의 이름이고, `TTY`는 현재 표준 스트림이 터미널에 연결됐는지 나타냅니다. 프롬프트 여부를 정할 때는 `TTY.Interactive`를 사용하세요.

전체 필드와 옵션은 [API 레퍼런스](docs/guide/api.md)에 있습니다.

## CLI

```console
$ runby
agent     paseo>codex
            paseo          orchestrator  delegated     definite  살아 있는 조상 pid=84445
            codex          harness       first-party   probable
ci        -
terminal  ghostty (probable)
remote    tmux (multiplexer)
runner    npm (script) test
tty       대화형 (stdin과 출력이 터미널)
process   조상 7개

주의: tmux 안에서 실행 중입니다. 멀티플렉서는 이미 열린 pane의 환경을
갱신하지 않으므로 터미널 축의 값이 낡았을 수 있습니다.
```

여기서 `codex`와 `ghostty`가 `probable`인 것은 tmux 때문입니다. 멀티플렉서는 이미 열린 pane의 환경을 갱신하지 못하므로, 살아 있는 조상으로 확증되지 않은 판정은 신뢰도가 한 단계 낮아집니다. `paseo`는 조상 프로세스로 확증돼서 `definite`로 남았습니다.

셸 조건문에서는 출력 대신 종료 코드로 답하는 `is` 명령을 사용합니다.

```sh
if runby is agent; then
	export NO_COLOR=1
fi

runby is agent codex   # 제품 이름을 덧붙여 좁힙니다
runby is remote tmux   # 축마다 같은 방식으로 동작합니다
runby chain            # paseo>codex
runby -v               # 감지 근거가 된 환경변수 이름과 조상 프로세스
runby -json            # Result 전체 JSON
```

제품 이름은 `-json`에 나오는 슬러그와 같습니다. 오타는 거짓(1)이 아니라 사용법 오류(2)로 답하므로, 스크립트가 조용히 잘못된 분기를 타지 않습니다.

`-json`에는 세션 ID와 로컬 경로가 포함될 수 있습니다. 공유용 진단 정보에는 값 대신 변수 이름만 출력하는 `runby -v`가 더 안전합니다. 자세한 내용은 [CLI 가이드](docs/guide/cli.md)를 참고하세요.

## 결과는 서로 독립된 정보입니다

한 프로세스는 동시에 Codex가 실행했고, GitHub Actions 안에 있으며, npm 스크립트를 거쳐, tmux가 붙어 있을 수 있습니다. `runby`는 이 정보를 하나로 뭉개지 않고 다음 축으로 나눠 보고합니다.

| 축 | 답하는 질문 | 결과 필드 |
|---|---|---|
| 에이전트 | 누가 명령을 요청했나? | `Agents` |
| CI | 어느 CI 잡에서 실행 중인가? | `CI` |
| 실행 도구 | npm·make·systemd 같은 무엇이 실행했나? | `Runners` |
| 원격 환경 | 사용자와 프로세스 사이에 무엇이 있나? | `Remotes` |
| 터미널 | 어떤 에뮬레이터가 환경을 만들었나? | `Terminal` |
| 프로세스 | 현재 살아 있는 조상은 무엇인가? | `Process` |

표준 스트림의 연결 상태는 별도의 `TTY` 필드로 제공합니다. 각 축의 의미와 신뢰도는 [개념별 가이드](docs/guide/README.md#개념별-상세)를 참고하세요.

## 지원 범위

- **에이전트:** Paseo, Orca, Antigravity 2.0, Cursor Agent, OpenCode, Amp, OpenClaw, Auggie, Cline, OpenAI Codex, Claude Code, Gemini CLI, Grok Build
- **CI:** GitHub Actions, Forgejo Actions, GitLab CI/CD, CircleCI, Travis CI, Buildkite, Azure Pipelines, Bitbucket Pipelines, Jenkins, 일반 `CI=true`
- **실행 도구:** npm, pnpm, Bun, GNU Make, systemd, pre-commit
- **원격 환경:** tmux, GNU Screen, Zellij, OpenSSH, WSL, GitHub Codespaces, Gitpod, Dev Containers
- **터미널:** iTerm2, Apple Terminal, WezTerm, Ghostty, Warp, Zed, VS Code 계열, JetBrains 계열, kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal, VTE 계열

목록에 없는 제품은 [드라이버](docs/guide/drivers.md)로 추가할 수 있습니다. 감지하지 않는 제품과 그 이유도 [조사 문서](docs/research/)에 기록합니다.

## 꼭 알아둘 한계

환경변수 기반 결과는 **프로세스 시작 시점의 스냅샷**입니다. 감지된 에이전트가 지금도 살아 있다는 뜻은 아닙니다. tmux처럼 오래 유지되는 환경에서는 이전 세션의 값이 남을 수도 있어, 멀티플렉서가 감지되면 살아 있는 조상으로 확증되지 않은 계층의 신뢰도를 `probable`로 낮춥니다.

현재 생존 여부가 중요하면 `AncestorPID`를 확인하세요. 값이 있으면 살아 있는 조상 프로세스로 확증된 것이지만, 값이 `0`이라고 해서 감지 결과가 틀렸다는 뜻은 아닙니다. 자세한 해석은 [프로세스 가이드](docs/guide/process.md)에 있습니다.

환경변수 값은 토큰이나 경로를 포함할 수 있으므로 감지 근거인 `Evidence`에는 **변수 이름만** 저장합니다.

## 문서

- [시작하기](docs/guide/getting-started.md) — 설치부터 첫 판정과 테스트까지
- [활용 예제](docs/guide/recipes.md) — 프롬프트, 로그, CI, 진단 정보 처리
- [CLI 가이드](docs/guide/cli.md) — 셸 조건문, JSON, 종료 코드
- [API 레퍼런스](docs/guide/api.md) — `Result`, 옵션, 캐시, 드라이버 API
- [개념별 가이드](docs/guide/) — 에이전트·CI·실행 도구·원격·터미널·프로세스
- [조사 문서](docs/research/) — 각 감지 규칙의 공식 근거와 제외 사유

## 플랫폼과 의존성

환경변수 기반 감지는 모든 플랫폼에서 동일하게 동작합니다. 상위 프로세스 체인은 Linux·macOS·Windows에서 지원하며, 그 외 플랫폼에서는 `Process.Supported == false`입니다. 일부 Unix 플랫폼의 TTY 제한은 [터미널 가이드](docs/guide/terminal.md#tty-터미널-축과-무엇이-다른가)에 정리되어 있습니다.

`runby`는 외부 Go 모듈에 의존하지 않습니다.
