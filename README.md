# runby

`runby`는 현재 프로세스를 실행한 코딩 에이전트 또는 오케스트레이터를 상속된 환경변수로 감지하는 Go 패키지입니다.

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

`runby`는 서로 **독립적인 세 축**을 보고합니다. 에이전트가 CI 안에서, 또 터미널 안에서 실행될 수 있으므로 세 축이 동시에 참일 수 있습니다.

| 축 | 질문 | 필드 |
|---|---|---|
| 에이전트 | 누가 명령을 요청했는가 | `Layers` |
| CI | 어떤 CI 잡에서 실행되는가 | `CI` |
| 터미널 | 어떤 에뮬레이터가 이 환경을 만들었는가 | `Terminal` |
| 원격·멀티플렉서 | 사용자와 이 프로세스 사이에 무엇이 끼어 있는가 | `Remote` |

여기에 환경변수가 아닌 두 가지가 더해집니다 — 시스템콜에서 오는 `TTY`(표준 스트림이 터미널인가)와, 커널에서 읽는 `Process`(상위 프로세스 체인).

API 키와 일반 설정 변수는 에이전트가 프로세스를 실행했다는 증거가 아니므로 감지에 사용하지 않습니다.

## 감지 대상

| 에이전트 | `Agent` | `Kind` | 식별 신호 |
|---|---|---|---|
| Paseo | `AgentPaseo` | `orchestrator` | `PASEO_AGENT_ID`, `PASEO_AGENT_CWD` |
| Orca (Stably AI) | `AgentOrca` | `orchestrator` | `ORCA_PANE_KEY` 또는 `ORCA_TERMINAL_HANDLE` + `ORCA_TAB_ID` 또는 `ORCA_WORKTREE_ID` |
| OpenAI Codex | `AgentCodex` | `harness` | `CODEX_THREAD_ID`, `CODEX_SESSION_ID`, 샌드박스 관련 변수 |
| Claude Code | `AgentClaudeCode` | `harness` | `CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`, `AI_AGENT=claude-code*` |
| Antigravity 2.0 sidecar | `AgentAntigravity2` | `harness` | `ANTIGRAVITY_EXECUTABLE_DATA_DIR` |
| Amp Orb / Orb 관리형 서비스 | `AgentAmp` | `harness` | `AMP_ORB`, `AMP_THREAD_ID` |
| Cursor Agent | `AgentCursor` | `harness` | `CURSOR_AGENT` |
| OpenCode ACP | `AgentOpenCode` | `harness` | `OPENCODE_CLIENT=acp` |

### CI 플랫폼

| 플랫폼 | `CIProvider` | 식별 신호 |
|---|---|---|
| Forgejo Actions | `CIProviderForgejo` | `FORGEJO_ACTIONS=true` (Runner v7+) |
| GitHub Actions | `CIProviderGitHubActions` | `GITHUB_ACTIONS=true` |
| GitLab CI/CD | `CIProviderGitLab` | `GITLAB_CI=true` |
| CircleCI | `CIProviderCircleCI` | `CIRCLECI=true` |
| Travis CI | `CIProviderTravis` | `TRAVIS=true` |
| Buildkite | `CIProviderBuildkite` | `BUILDKITE=true` |
| Azure Pipelines | `CIProviderAzurePipelines` | `TF_BUILD=True` |
| Bitbucket Pipelines | `CIProviderBitbucket` | `BITBUCKET_BUILD_NUMBER` + (`BITBUCKET_PIPELINE_UUID` 또는 `CI`) |
| Jenkins | `CIProviderJenkins` | `BUILD_NUMBER` + (`JENKINS_URL` 또는 `JENKINS_HOME` 등) |
| 그 외 | `CIProviderGeneric` | `CI=true` (보조 신호) |

Antigravity CLI, GitHub Copilot CLI, Junie, 일반 OpenCode CLI는 공식적으로 확인된 범용 자식 프로세스 실행 마커가 없어 감지하지 않습니다. 자세한 조사 근거는 [`docs/`](docs/)에 있습니다.

### Kind: 오케스트레이터인가, 하네스인가

`KindOrchestrator`는 하위 에이전트를 관리하며 자신의 에이전트 ID를 광고하는 제품(Paseo, Orca), `KindHarness`는 모델이 요청한 명령을 실행하는 런타임입니다. 둘 다 **AI 에이전트가 이 프로세스를 실행했다는 증거**이므로 `Found()`가 그대로 그 질문의 답이 됩니다.

터미널을 소유한다는 사실은 에이전트 실행의 증거가 아닙니다. Zed는 Agent 전용 신호가 없어 이 축이 아니라 **터미널 축**(`Terminal.Program == TerminalZed`)으로 보고합니다.

`Confidence`는 신호의 직접성을 구분합니다. `ConfidenceDefinite`는 제품이 자신이 실행한 프로세스에 한해 설정하는 실행 마커이고, `ConfidenceProbable`은 에이전트 실행과 모순되지 않지만 그것만의 신호는 아닌 보조 신호입니다.

Orca가 `ConfidenceProbable`인 이유가 그 예입니다. Orca는 자신이 호스팅하는 pane에 표시를 남기므로, 사용자가 Orca 터미널에 직접 입력한 명령과 Orca가 실행한 에이전트가 호출한 명령이 같은 변수를 갖습니다. 실제로 실행한 에이전트는 해당 harness의 고유 신호로 별도 계층에 보고됩니다.

## CI

**CI는 `Layers`가 아니라 별도 축입니다.** Claude Code가 GitHub Actions 워크플로에서 실행되면 `KindHarness` 레이어와 CI 결과가 **동시에** 채워집니다. `Kind`는 "누가 명령을 요청했는가"를, `CI`는 "어디서 실행되는가"를 답합니다. 그래서 `Chain()`에는 CI가 들어가지 않습니다.

```go
result := runby.Detect()
result.Found()            // Claude Code가 실행했는가
result.IsCI()             // CI 잡에서 도는가
result.CI.Provider        // "github-actions"
result.CI.PipelineID      // GITHUB_RUN_ID
result.CI.JobID           // GITHUB_JOB
result.CI.Attempt         // GITHUB_RUN_ATTEMPT (1부터)
result.CI.Trigger         // GITHUB_EVENT_NAME ("push", "pull_request", ...)
```

`CI` 구조체는 플랫폼별 용어를 공통 필드로 정규화합니다.

| 필드 | 의미 |
|---|---|
| `PipelineID` | run/build/pipeline 단위 식별자 |
| `BuildNumber` | UI에 보이는 사람용 카운터. 프로젝트 간 유일하지 않음 |
| `JobID` / `JobName` | 파이프라인 내 개별 job·step |
| `Attempt` | **1부터 시작하는** 재시도 회차. 광고하지 않으면 `0` |
| `Trigger` | 플랫폼 자체 용어의 트리거 종류. 없으면 빈 문자열 |
| `Runner` | 잡을 실행 중인 머신·에이전트 |
| `Extra` | 단일 플랫폼 전용 값. 키는 `"<provider-slug>.<name>"` |

정규화 시 두 가지를 처리합니다.

- **`Attempt`는 1-based로 통일합니다.** Buildkite `BUILDKITE_RETRY_COUNT`와 GitLab `CI_JOB_RETRY_COUNT`는 0부터 세는 재시도 횟수이므로 +1 해서 맞춥니다. GitHub `GITHUB_RUN_ATTEMPT`와 Azure `SYSTEM_JOBATTEMPT`는 이미 1부터라 그대로 씁니다.
- **Bitbucket UUID의 중괄호를 벗깁니다.** `BITBUCKET_PIPELINE_UUID`는 `{11d8...}` 형태로 오므로 `PipelineID`/`JobID`에는 중괄호를 뺀 값이 들어갑니다.

Forgejo Actions는 Runner v7+에서 모든 `FORGEJO_*`를 `GITHUB_*` 별칭으로도 제공하므로 GitHub Actions보다 **먼저** 검사합니다. v7 미만 Runner는 `GITHUB_*`만 제공해 환경변수로는 구별할 수 없어 GitHub Actions로 보고됩니다.

플랫폼별 조사 근거는 [`docs/ci/`](docs/ci/)에 있습니다. 지원하지 않는 플랫폼은 `WithCIDetectors`로 추가할 수 있습니다.

```go
detector := runby.NewCIDetector("acme-ci", func(env runby.Env) (runby.CI, bool) {
	id, ok := runby.Value(env, "ACME_CI_BUILD")
	if !ok {
		return runby.CI{}, false
	}
	return runby.CI{PipelineID: id, Evidence: runby.PresentNames(env, "ACME_CI_BUILD")}, true
})

result := runby.Detect(runby.WithCIDetectors(detector))
```

## Terminal

터미널 축은 **의도적으로 가장 약한 신호**입니다. 환경변수는 모든 자손 프로세스가 상속하므로 이 값은 *이 환경을 만든 터미널*을 가리키며, *지금 이 프로세스의 TTY에 붙어 있는 터미널*이 아닙니다.

```go
term := runby.Detect().Terminal
term.Program      // "ghostty", "kitty", "vte", ...
term.SessionID    // 창·탭·pane 식별자 (제공하는 터미널만)
term.Version
term.PID          // kitty만. 0이 아니면 생존 확인 가능
term.Term         // TERM 값. 판정에는 쓰지 않고 컨텍스트로만 기록
term.Multiplexer  // "tmux", "screen", 또는 ""
```

| 터미널 | `TerminalProgram` | 마커 |
|---|---|---|
| iTerm2 | `TerminalITerm2` | `TERM_PROGRAM=iTerm.app` |
| Apple Terminal | `TerminalAppleTerminal` | `TERM_PROGRAM=Apple_Terminal` |
| WezTerm | `TerminalWezTerm` | `TERM_PROGRAM=WezTerm` |
| Ghostty | `TerminalGhostty` | `TERM_PROGRAM=ghostty` |
| Warp | `TerminalWarp` | `TERM_PROGRAM=WarpTerminal` |
| Zed | `TerminalZed` | `ZED_TERM=true` + `TERM_PROGRAM=zed` |
| kitty | `TerminalKitty` | `KITTY_WINDOW_ID` |
| Windows Terminal | `TerminalWindowsTerminal` | `WT_SESSION` |
| Alacritty | `TerminalAlacritty` | `ALACRITTY_LOG` |
| Konsole | `TerminalKonsole` | `KONSOLE_VERSION` 또는 `KONSOLE_DBUS_SESSION` |
| GNOME Terminal | `TerminalGNOMETerminal` | `GNOME_TERMINAL_SCREEN` |
| VTE 계열 | `TerminalVTE` | `VTE_VERSION` (제품이 아니라 계열) |

### 이 축을 신뢰 경계로 쓰면 안 되는 이유

- **멀티플렉서 잔존** — 위 [Remote](#remote) 절을 참고하십시오. `runby`는 멀티플렉서를 감지하면 `Confidence`를 `probable`로 낮춥니다.
- **SSH 정체성 전파** — iTerm2의 `LC_TERMINAL`(기본 켜짐), kitty의 `kitten ssh`, Ghostty의 `ssh-env`가 터미널 정체성을 원격 호스트에 의도적으로 전달합니다. 그래서 `runby`는 `LC_TERMINAL`을 **감지에 전혀 쓰지 않습니다**.
- **데몬화**와 **위조** — 낡은 스냅샷이 남거나 누구나 `export TERM_PROGRAM=...` 할 수 있습니다.

### 알아둘 세부사항

- **`TERM`은 마커가 아닙니다.** 정체성이 아니라 terminfo 능력을 나타내고, 사용자가 덮어쓰며, 멀티플렉서가 교체하고, Alacritty·Ghostty는 terminfo가 없으면 `xterm-256color`로 폴백합니다. 판정이 끝난 뒤 컨텍스트로만 기록합니다.
- **`TERM_PROGRAM`은 절반에서만 쓸 수 있습니다.** kitty, Windows Terminal, Alacritty, Konsole, GNOME Terminal은 설정하지 않습니다.
- **`VTE_VERSION`은 계열입니다.** VTE 라이브러리가 설정하므로 XFCE Terminal·guake·terminator 등이 공유합니다. `GNOME_TERMINAL_*` 없이 이 값만 있으면 `TerminalVTE`로 보고합니다.
- **`Terminal.PID`는 kitty만 제공합니다.** 다른 모든 신호와 달리 프로세스 조회로 낡은 마커와 살아 있는 터미널을 구분할 수 있습니다.
- **Konsole은 항상 `probable`입니다.** `konsolepart`(Dolphin, Kate, KDevelop, Krusader)가 같은 라이브러리를 써서 동일한 변수를 주입하므로, 이 증거는 "Konsole 엔진"을 증명할 뿐 사용자가 Konsole 창을 보고 있다는 뜻은 아닙니다.

터미널별 조사 근거는 [`docs/terminals/`](docs/terminals/)에, 멀티플렉서와 원격 실행 계층은 [`docs/remote/`](docs/remote/)에 있습니다.

## Process

**환경변수가 아닌 유일한 증거이며, 가장 강한 증거입니다.**

환경변수는 자손이 모두 상속하고, 멀티플렉서가 몇 시간 전 값을 물려주며, 누구나 `export`로 위조할 수 있습니다. 상위 프로세스는 다릅니다 — `export`로 만들 수 없고, 상속되지 않으며, **보인다는 것 자체가 지금 살아 있다는 뜻**입니다. 환경변수가 "이 프로세스가 시작될 때 참이었던 것"만 말할 수 있는 반면, 조상 프로세스는 **지금** 참인 것을 말합니다.

```go
tree := runby.Detect().Process
tree.Supported            // Linux·macOS·Windows에서만 true
tree.Ancestors            // 가까운 부모부터
tree.FindAgent(runby.AgentCodex)
```

```
env chain : paseo>claude-code
  pid=3066    zsh
  pid=11904   claude           -> agent=claude-code
  pid=2540    paseo            -> agent=paseo
corroboration:
  paseo          CONFIRMED by live ancestor pid=2540
  claude-code    CONFIRMED by live ancestor pid=11904
```

### 교차 검증

에이전트가 감지되고 그 실행 파일이 조상으로 살아 있으면 `Detection.AncestorPID`가 채워집니다.

```go
for _, layer := range runby.Current().Layers {
	if layer.AncestorPID != 0 {
		// 환경변수가 이 에이전트를 말했고, 그 에이전트가 지금도 조상으로 살아 있음
	}
}
```

**`AncestorPID == 0`은 부정이 아닙니다.** 체인은 다른 사용자 소유 프로세스에서 멈추고, 일부 플랫폼에서는 아예 읽을 수 없으며, 에이전트가 조상으로 남지 않는 방식으로 프로세스를 띄울 수도 있습니다. **긍정을 강화하는 데만 쓰고, 부정의 근거로 쓰면 안 됩니다.** 이 규칙을 테스트로 고정해 두었습니다.

### 플랫폼

| 플랫폼 | 방법 |
|---|---|
| Linux | `/proc/<pid>/stat`(ppid), `/proc/<pid>/exe`, `/proc/<pid>/comm` |
| macOS | `sysctl(KERN_PROC_PID)` + `KERN_PROCARGS2` |
| Windows | `CreateToolhelp32Snapshot` 스냅샷 1회 |
| 그 외 | `Supported == false`, 빈 체인 |

macOS에는 `kinfo_proc` 헤더가 없어 ppid 오프셋을 상수로 두어야 합니다. 그래서 **시작 시 `os.Getppid()`와 대조해 검증**하고, 어긋나면 잘못된 필드를 읽는 대신 기능을 끕니다.

`TTY`와 마찬가지로 이 축은 **이 프로세스**를 설명하므로 `WithEnviron` 계열에서는 읽지 않습니다. 비용이 다른 축보다 크므로 `WithoutProcessTree()`로 끌 수 있습니다.

## Remote

사용자와 이 프로세스 사이에 낀 계층입니다. 이 축이 따로 있는 이유는 여기 있는 것들이 **자기 변수를 추가하는 데 그치지 않고 다른 축의 변수가 살아남을지를 결정**하기 때문입니다 — tmux는 `update-environment`로, OpenSSH는 `SendEnv`/`AcceptEnv`로, WSL은 `WSLENV`로, Dev Containers는 `containerEnv`/`remoteEnv`로 거릅니다. 따라서 이 축의 감지 결과는 독립된 사실이 아니라 **다른 축을 얼마나 믿을 수 있는지에 대한 단서**입니다.

**여러 계층이 동시에 존재할 수 있으므로 슬라이스입니다.** Codespace에 SSH로 붙어 tmux를 쓰면 세 계층이 함께 잡힙니다.

```go
result := runby.Detect()
result.IsRemote()                      // 낀 계층이 있는가
result.HasRemote(runby.RemoteSSH)      // 특정 계층
result.GetRemote(runby.RemoteTmux)     // (Remote, bool)
result.Multiplexer()                   // (Remote, bool) — 잔존 위험의 주 원인
```

| 계층 | `RemotePlatform` | `Kind` | 마커 |
|---|---|---|---|
| tmux | `RemoteTmux` | `multiplexer` | `TMUX` |
| GNU Screen | `RemoteScreen` | `multiplexer` | `STY` |
| Zellij | `RemoteZellij` | `multiplexer` | `ZELLIJ` (값은 리터럴 `"0"`) |
| OpenSSH | `RemoteSSH` | `environment` | `SSH_CONNECTION` |
| WSL | `RemoteWSL` | `environment` | `WSL_DISTRO_NAME` 또는 `WSL_INTEROP` |
| GitHub Codespaces | `RemoteCodespaces` | `environment` | `CODESPACES=true` |
| Gitpod | `RemoteGitpod` | `environment` | `GITPOD_WORKSPACE_ID` |
| Dev Containers | `RemoteDevContainer` | `environment` | `REMOTE_CONTAINERS` 또는 `DEVCONTAINER` |

`Remote`의 순서는 **감지 순서일 뿐 중첩 순서가 아닙니다.** 환경변수로는 어느 계층이 바깥인지 증명할 수 없습니다.

### 멀티플렉서만 신뢰도를 낮춥니다

멀티플렉서 서버는 처음 붙은 클라이언트의 환경을 유지하고 **이미 실행 중인 pane의 환경은 갱신할 수 없습니다.** 그래서 `Multiplexer()`가 잡히면 `Terminal.Confidence`를 `probable`로 낮춥니다. SSH는 다릅니다 — 터미널이 다른 머신에 있을 수 있다는 뜻이지만 값이 낡은 것은 아니므로, 신뢰도를 낮추는 대신 `RemoteSSH` 계층의 존재로 그 사실을 표현합니다.

tmux와 Screen은 실패 방향이 정반대입니다. tmux는 `TERM_PROGRAM`을 `tmux`로 덮어써 그 계열 터미널 6종의 정체성을 **지우고**(거짓 음성), 건드리지 않는 마커만 통과시켜 **잔존**시킵니다(거짓 양성). Screen과 Zellij는 `TERM_PROGRAM`을 덮어쓰지 않고 갱신 기제도 없어 **모든 마커가 잔존 가능**합니다.

잔존 위험은 터미널 축에만 걸리지 않습니다. 오래 사는 서버는 처음 시작될 때의 CI·에이전트 마커도 나중 pane에 물려줍니다. `runby`는 이를 신뢰도로 자동 반영하지 않고 `Multiplexer()`라는 사실만 노출하며, 세 축에 대한 해석은 소비자에게 맡깁니다.

### 감지할 수 없는 것

- **Mosh** — 원리적으로 불가능합니다. `MOSH_KEY`는 제거되는 게 아니라 애초에 원격 셸 환경에 들어간 적이 없고, 나머지 `MOSH_*`는 전부 클라이언트 전용입니다. 정상 세션에는 `MOSH_*`가 하나도 없습니다. `MOSH_KEY`는 **자격증명**이므로 어떤 경우에도 읽거나 기록하면 안 됩니다.
- **컨테이너 일반** — Docker·Podman은 식별 환경변수를 설정하지 않습니다. 관례적 감지는 `/.dockerenv`, `/run/.containerenv`, cgroup 같은 **파일시스템 경로**를 쓰므로 환경변수만 읽는 이 라이브러리의 범위 밖입니다. Dev Containers나 Codespaces처럼 도구가 스스로 광고한 경우만 보입니다. `HOSTNAME`이 짧은 16진 문자열인 것은 근거가 아니라 추측이라 쓰지 않습니다.

### 오탐을 피하려고 일부러 쓰지 않는 변수

- **`SSH_AUTH_SOCK`** — `ssh-agent`가 로컬 데스크톱에서도 설정하므로 SSH 세션 마커가 아닙니다.
- **`WINDOW`** (Screen) — 무조건 설정되지만 다른 소프트웨어와 이름이 겹치기 쉬워 `STY`와 함께일 때만 컨텍스트로 씁니다.
- **`LC_TERMINAL`** — 배포판이 배포하는 `SendEnv LC_*` 설정에 걸려 SSH를 건너가므로 다른 머신의 터미널을 가리킬 수 있습니다.

조사 근거는 [`docs/remote/`](docs/remote/)에 있습니다.

## API

진입점은 `Detect(opts ...Option) Result` 하나입니다.

```go
result := runby.Detect()                                  // 현재 프로세스 (터미널 포함)
result := runby.Detect(runby.WithEnviron(environ))         // 명시적 환경
result := runby.Detect(runby.WithoutTTY())                 // TTY 시스템콜 생략
result := runby.Detect(runby.WithDetectors(myDetector))    // 사내 오케스트레이터 추가
```

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

### Result

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
result.Multiplexer()            // (Remote, bool)
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

### TTY

```go
tty := runby.Detect().TTY
tty.Inspected   // 표준 스트림을 실제로 검사했는가
tty.StdinTTY
tty.StdoutTTY
tty.StderrTTY
tty.Attached    // 세 스트림 중 하나 이상이 터미널
tty.Interactive // stdin과 출력 스트림 하나 이상이 터미널
```

`Interactive`는 프롬프트를 띄우고 응답을 받을 수 있는 형태라는 뜻일 뿐, 사용자가 직접 명령을 호출했다는 증거는 아닙니다. 에이전트나 서브에이전트도 PTY를 할당할 수 있습니다.

`TTY`는 `Result`에서 유일하게 환경변수가 아닌 **시스템콜**로 얻는 값입니다. 그래서 `WithEnviron` 계열 옵션(이 프로세스의 것이 아닐 수도 있는 환경)에서는 검사하지 않으며 `Inspected`가 `false`입니다. 필요하면 `InspectTTY()`를 직접 호출하거나 `WithTTY()`로 주입하십시오. 반대로 `Terminal`은 환경변수 기반이라 `WithEnviron`에서도 그대로 채워집니다.

### 캐시된 진입점

프로세스 환경과 표준 스트림은 실무상 시작 시점에 고정되므로, 대부분의 호출자는 캐시된 진입점을 쓰면 됩니다.

```go
runby.Current()  // Detect()를 1회만 계산해 캐시
runby.IsAgent()  // Current().Found()
runby.IsCI()       // Current().CI.Detected
runby.IsTerminal() // Current().Terminal.Detected
runby.IsRemote()   // len(Current().Remote) > 0
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

## 의존성

**표준 라이브러리만 사용합니다.** `go.mod`에 외부 모듈이 없고 `go.sum`도 없습니다.

TTY 검사는 [`internal/term`](internal/term/)에, 프로세스 체인 읽기는 [`internal/proc`](internal/proc/)에 있습니다. `golang.org/x/term`의 `IsTerminal` 하나만 옮겨와 `golang.org/x/sys` 대신 표준 `syscall`로 다시 작성했으며, 원본의 BSD-3-Clause 라이선스를 함께 보관합니다.

그 대가로 **AIX·Solaris·z/OS에서는 `TTY.Attached`와 `TTY.Interactive`가 항상 `false`**입니다. 표준 `syscall` 패키지가 이 플랫폼들에 `TCGETS`·`SYS_IOCTL`을 노출하지 않기 때문입니다. 환경변수만 읽는 나머지 네 축은 모든 플랫폼에서 동일하게 동작합니다.

## 상태의 의미

환경변수는 프로세스가 시작될 때 상속된 스냅샷입니다. 감지 성공은 에이전트가 이 프로세스를 실행할 당시 활성 상태였다는 뜻입니다. 장시간 실행되는 프로세스에서 부모 에이전트가 현재도 살아 있는지 확인하려면 PID, IPC 또는 에이전트 API 같은 별도의 생존 확인 수단이 필요합니다.
