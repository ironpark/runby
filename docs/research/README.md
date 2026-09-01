# 조사 문서

`runby`가 감지하는 각 제품의 환경변수를 공식 문서·소스에서 확인한 기록입니다. 사용법이 아니라 판정의 **근거**를 담으므로, 라이브러리를 사용하기만 한다면 루트 [`README.md`](../../README.md)를 보세요. 이 디렉터리의 위치와 역할은 [`docs/README.md`](../README.md)에 정리되어 있습니다.

이 문서는 모든 조사 문서가 공유하는 양식과 front matter 필드를 정의하고, 축별 문서 목록을 제공합니다.

각 문서의 조사 기준일과 제품·소스 특성은 YAML front matter에 기록합니다. 제작사의 최신 공식 문서를 우선하고, 문서에 없는 구현 세부사항은 제작사가 공개한 공식 소스에서만 보충했습니다. 환경변수는 상속·잔존·위조될 수 있으므로 `적합` 판정도 부모 에이전트의 현재 생존 상태까지 보장하지는 않습니다.

## Front matter

| 필드 | 의미 |
|---|---|
| `title` | 문서에 표시되는 제품명 |
| `slug` | `runby`에서 사용하는 안정적인 제품 식별자 |
| `research_date` | 공식 문서와 소스를 마지막으로 확인한 날짜 (`YYYY-MM-DD`) |
| `open_source` | 실행 제품의 핵심 구현이 오픈소스 라이선스로 공개되는지 여부. 공개 저장소나 issue tracker만 있는 경우는 `false` |
| `repository` | `open_source: true`일 때의 공식 소스 저장소. 그 외에는 `null` |
| `product_type` | `agent_harness`, 여러 하위 agent를 관리하는 `agent_orchestrator`, agent를 내장·연결하는 `agent_host`, CI 플랫폼인 `ci_platform`, 터미널 에뮬레이터인 `terminal_emulator`, 터미널 멀티플렉서인 `terminal_multiplexer`, 원격·격리 실행 환경인 `remote_environment`, 또는 다른 프로그램을 실행하는 것이 본업인 `task_runner` |
| `model_source` | agent 문서 전용. 자사 모델을 쓰는 `first-party`, 타사 모델을 등록·API로 쓰는 `multi-vendor`, 모델 없이 다른 agent를 구동하는 `delegated`. 코드의 `ModelSource`와 대응하며 `TestKindsMatchDocs`가 일치를 강제 |
| `executes_agents` | 제품이 실행하거나 호스팅하는 별도 agent/harness 식별자. 자체 하네스만 실행하면 빈 배열 |
| `runtime_test_required` | 공식 문서·소스 조사 외에 제품을 직접 실행하는 추가 검증이 필요한지 여부 |
| `runtime_test_reason` | 직접 실행 검증이 필요하거나 불필요하다고 판정한 근거와 핵심 시험 범위 |

`product_type`은 코드의 `Kind`와 대응합니다. `agent_harness`는 `KindHarness`, `agent_orchestrator`는 `KindOrchestrator`입니다. `ci_platform`과 `terminal_emulator`는 예외로, `Kind`가 아니라 각각 `Result.CI`와 `Result.Terminal`이라는 **별도 축**이 됩니다. 에이전트가 CI 안에서, 또 터미널 안에서 실행될 수 있어 축들이 동시에 참일 수 있기 때문입니다. 같은 이유로 `Result.Runners`와 `Result.Remotes`도 별도 축입니다.

`agent_host`는 현재 코드에 대응하는 `Kind`가 없습니다. 터미널을 소유한다는 사실은 에이전트 실행의 증거가 아니므로, Zed처럼 여기 해당하던 제품은 터미널 축으로 옮겼으므로, `agent_host` 문서는 `Result.Terminal`을 통해 반영됩니다.

`provider`라는 말은 LLM API 공급자와 agent 실행 구현 양쪽에 쓰여 혼동될 수 있으므로, front matter에서는 후자를 `executes_agents`로 명시합니다. Paseo처럼 orchestrator 신호와 하위 harness 신호가 함께 존재할 수 있으며, 이 경우 둘을 각각 보존해야 합니다.

| 에이전트 | 공식 실행 식별 신호 | 적용 범위 |
|---|---|---|
| Paseo | `PASEO_AGENT_ID` | Paseo 에이전트가 실행한 프로세스 |
| Orca (Stably AI) | `ORCA_PANE_KEY` + `ORCA_TAB_ID`/`ORCA_WORKTREE_ID` | Orca 관리형 pane/worktree. 실제 하위 agent는 agent별 신호로 별도 판별 |
| OpenAI Codex | `CODEX_THREAD_ID`, `CODEX_SESSION_ID` | 모델이 접근할 수 있는 셸 명령; 공식 CLI 소스에서 주입을 확인했으나 안정적 공개 환경변수 계약에는 미포함 |
| Claude Code | `CLAUDE_CODE_CHILD_SESSION=1`, `CLAUDECODE=1` | 직접 자식 및 넓은 Claude Code 실행 컨텍스트 |
| Antigravity CLI | 없음 | 공식 변수는 표시·업데이트 설정용이며 범용 자식 프로세스 실행 마커는 확인되지 않음 |
| Antigravity 2.0 | `ANTIGRAVITY_EXECUTABLE_DATA_DIR` | Antigravity 2.0이 lifecycle을 관리하는 sidecar에 한정 |
| Amp | `AMP_ORB=1`, `AMP_THREAD_ID` | Orb 및 Orb 관리형 서비스에 한정 |
| Zed Agent | Agent 전용 신호 없음 | `ZED_TERM=true`는 Zed 터미널만 식별 |
| Cursor Agent | `CURSOR_AGENT` | Cursor Agent가 실행한 셸 |
| GitHub Copilot CLI | 없음 | 공개 변수는 설정·관측용 |
| OpenCode | `OPENCODE=1` | 모든 OpenCode 실행. `OPENCODE_CLIENT=acp`는 ACP 진입점을 덧붙이며, 이 변수만 단독으로 있으면 `probable` |
| Gemini CLI | `GEMINI_CLI` | Gemini CLI가 실행한 셸 |
| Grok Build | `GROK_PLUGIN_ROOT`, `GROK_PLUGIN_DATA` | Grok Build 플러그인이 실행한 프로세스 |
| OpenClaw | `OPENCLAW_SHELL` | OpenClaw가 실행한 셸 |
| Auggie | `AUGMENT_AGENT` | Auggie가 실행한 프로세스 |
| Cline | `CLINE_ACTIVE` | Cline이 실행한 셸. 에디터 안에서 돌므로 조상 실행 파일로는 확증할 수 없음 |
| JetBrains Junie | 없음 | 공개 변수는 CLI 입력 설정용 |

## CI 플랫폼별 조사

CI 감지는 "누가 명령을 요청했는가"가 아니라 "어디서 실행되는가"를 답하므로 `Agent`가 아닌 `Result.CI`로 보고합니다. 문서는 [`ci/`](ci/)에 있습니다.

- [GitHub Actions](ci/github-actions.md)
- [Forgejo Actions (Forgejo Runner)](ci/forgejo-runner.md)
- [GitLab CI/CD](ci/gitlab-ci.md)
- [CircleCI](ci/circleci.md)
- [Travis CI](ci/travis-ci.md)
- [Buildkite](ci/buildkite.md)
- [Azure Pipelines](ci/azure-pipelines.md)
- [Bitbucket Pipelines](ci/bitbucket-pipelines.md)
- [Jenkins](ci/jenkins.md)

## 멀티플렉서·원격 실행 계층 조사

사용자와 프로세스 사이에 끼어 있는 계층입니다. 다른 범주와 달리 자기 변수를 추가하는 데 그치지 않고 **다른 축의 변수가 살아남을지, 어떻게 변형될지를 결정**하므로(`update-environment`, `SendEnv`, `WSLENV`), 나머지 문서의 신뢰도 해석에 전제가 됩니다. 문서는 [`remote/`](remote/)에 있습니다.

- [tmux](remote/tmux.md)
- [GNU Screen](remote/gnu-screen.md)
- [Zellij](remote/zellij.md)
- [OpenSSH](remote/openssh.md)
- [WSL](remote/wsl.md)
- [GitHub Codespaces](remote/github-codespaces.md)
- [Gitpod](remote/gitpod.md)
- [Dev Containers](remote/devcontainers.md)
- [Mosh](remote/mosh.md) — 감지하지 않는 이유

## 실행 주체 조사

"무엇이 이 프로세스를 직접 실행했는가"의 흔한 답 하나 — **스크립트가 실행했다** — 를 담는 축입니다. `npm test`로 시작된 프로세스는 터미널이 붙어 있고 CI도 에이전트도 아니라서 기존 축들은 모두 "사람이 대화형으로 실행했다"고 답하지만, 실제로는 `package.json`의 스크립트가 실행했습니다. 문서는 [`runners/`](runners/)에 있습니다.

이 축의 조사에서 나온 가장 중요한 결과는 **git 훅을 환경변수로 감지할 수 없다**는 음성 결과입니다. `post-checkout` 훅과 평범한 git 별칭이 동일한 `GIT_*` 집합을 받는 것을 실측으로 확인했습니다. 근거는 [`runners/README.md`](runners/README.md)에 있습니다.

## 터미널 에뮬레이터별 조사

터미널 식별은 상속·SSH 전파·멀티플렉서 잔존 때문에 구조적으로 현재 상태를 보장할 수 없는 가장 약한 축입니다. 문서는 [`terminals/`](terminals/)에 있습니다.

## 에이전트별 조사

- [Paseo](agents/paseo.md)
- [Orca (Stably AI)](agents/orca.md)
- [OpenAI Codex](agents/codex.md)
- [Claude Code](agents/claude-code.md)
- [Antigravity CLI](agents/antigravity-cli.md)
- [Antigravity 2.0](agents/antigravity-2.md)
- [Amp](agents/amp.md)
- [Zed Agent](agents/zed-agent.md)
- [Cursor Agent](agents/cursor-agent.md)
- [GitHub Copilot CLI](agents/github-copilot-cli.md)
- [OpenCode](agents/opencode.md)
- [Gemini CLI](agents/gemini-cli.md)
- [Grok Build](agents/grok-build.md)
- [OpenClaw](agents/openclaw.md)
- [Auggie](agents/auggie.md)
- [Cline](agents/cline.md)
- [JetBrains Junie](agents/junie.md) — 감지하지 않는 이유
- [Goose](agents/goose.md) — 감지하지 않는 이유
- [Cowork](agents/cowork.md) — 감지하지 않는 이유
- [Kimi CLI](agents/kimi-cli.md) — 감지하지 않는 이유
