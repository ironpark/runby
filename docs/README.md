# Agent documentation

에이전트별 환경변수 감지 규칙과 상태 해석을 정리하는 문서입니다.

각 문서의 조사 기준일과 제품·소스 특성은 YAML front matter에 기록합니다. 제작사의 최신 공식 문서를 우선하고, 문서에 없는 구현 세부사항은 제작사가 공개한 공식 소스에서만 보충했습니다. 환경변수는 상속·잔존·위조될 수 있으므로 `적합` 판정도 부모 에이전트의 현재 생존 상태까지 보장하지는 않습니다.

## Front matter

| 필드 | 의미 |
|---|---|
| `title` | 문서에 표시되는 제품명 |
| `slug` | `runby`에서 사용하는 안정적인 제품 식별자 |
| `research_date` | 공식 문서와 소스를 마지막으로 확인한 날짜 (`YYYY-MM-DD`) |
| `open_source` | 실행 제품의 핵심 구현이 오픈소스 라이선스로 공개되는지 여부. 공개 저장소나 issue tracker만 있는 경우는 `false` |
| `repository` | `open_source: true`일 때의 공식 소스 저장소. 그 외에는 `null` |
| `product_type` | `agent_harness`, 여러 하위 agent를 관리하는 `agent_orchestrator`, agent를 내장·연결하는 `agent_host`, CI 플랫폼인 `ci_platform`, 터미널 에뮬레이터인 `terminal_emulator`, 터미널 멀티플렉서인 `terminal_multiplexer`, 또는 원격·격리 실행 환경인 `remote_environment` |
| `executes_agents` | 제품이 실행하거나 호스팅하는 별도 agent/harness 식별자. 자체 하네스만 실행하면 빈 배열 |
| `runtime_test_required` | 공식 문서·소스 조사 외에 제품을 직접 실행하는 추가 검증이 필요한지 여부 |
| `runtime_test_reason` | 직접 실행 검증이 필요하거나 불필요하다고 판정한 근거와 핵심 시험 범위 |

`product_type`은 코드의 `Kind`와 대응합니다. `agent_harness`는 `KindHarness`, `agent_orchestrator`는 `KindOrchestrator`입니다. `ci_platform`과 `terminal_emulator`는 예외로, `Kind`가 아니라 각각 `Result.CI`와 `Result.Terminal`이라는 **별도 축**이 됩니다. 에이전트가 CI 안에서, 또 터미널 안에서 실행될 수 있어 세 축이 동시에 참일 수 있기 때문입니다.

`agent_host`는 현재 코드에 대응하는 `Kind`가 없습니다. 터미널을 소유한다는 사실은 에이전트 실행의 증거가 아니므로, Zed처럼 여기 해당하던 제품은 터미널 축으로 옮겼으므로, `agent_host` 문서는 `Result.Terminal`을 통해 반영됩니다.

`provider`라는 말은 LLM API 공급자와 agent 실행 구현 양쪽에 쓰여 혼동될 수 있으므로, front matter에서는 후자를 `executes_agents`로 명시합니다. Paseo처럼 orchestrator 신호와 하위 harness 신호가 함께 존재할 수 있으며, 이 경우 둘을 각각 보존해야 합니다.

| 에이전트 | 공식 실행 식별 신호 | 적용 범위 |
|---|---|---|
| Paseo | `PASEO_AGENT_ID` | Paseo 에이전트가 실행한 프로세스 |
| OpenAI Codex | `CODEX_THREAD_ID`, `CODEX_SESSION_ID` | 모델이 접근할 수 있는 셸 명령; 공식 CLI 소스에서 주입을 확인했으나 안정적 공개 환경변수 계약에는 미포함 |
| Claude Code | `CLAUDE_CODE_CHILD_SESSION=1`, `CLAUDECODE=1` | 직접 자식 및 넓은 Claude Code 실행 컨텍스트 |
| Antigravity CLI | 없음 | 공식 변수는 표시·업데이트 설정용이며 범용 자식 프로세스 실행 마커는 확인되지 않음 |
| Antigravity 2.0 | `ANTIGRAVITY_EXECUTABLE_DATA_DIR` | Antigravity 2.0이 lifecycle을 관리하는 sidecar에 한정 |
| Amp | `AMP_ORB=1`, `AMP_THREAD_ID` | Orb 및 Orb 관리형 서비스에 한정 |
| Zed Agent | Agent 전용 신호 없음 | `ZED_TERM=true`는 Zed 터미널만 식별 |
| Cursor Agent | `CURSOR_AGENT` | Cursor Agent가 실행한 셸 |
| GitHub Copilot CLI | 없음 | 공개 변수는 설정·관측용 |
| OpenCode | 일반 실행 신호 없음 | `OPENCODE_CLIENT=acp`는 ACP 실행의 보조 신호 |
| JetBrains Junie | 없음 | 공개 변수는 CLI 입력 설정용 |

## CI 플랫폼별 조사

CI 감지는 "누가 명령을 요청했는가"가 아니라 "어디서 실행되는가"를 답하므로 `Agent`가 아닌 `Result.CI`로 보고합니다. 문서는 [`ci/`](ci/)에 있습니다.

- [Forgejo Actions (Forgejo Runner)](ci/forgejo-runner.md)

## 멀티플렉서·원격 실행 계층 조사

사용자와 프로세스 사이에 끼어 있는 계층입니다. 다른 범주와 달리 자기 변수를 추가하는 데 그치지 않고 **다른 축의 변수가 살아남을지, 어떻게 변형될지를 결정**하므로(`update-environment`, `SendEnv`, `WSLENV`), 나머지 문서의 신뢰도 해석에 전제가 됩니다. 문서는 [`remote/`](remote/)에 있습니다.

## 터미널 에뮬레이터별 조사

터미널 식별은 상속·SSH 전파·멀티플렉서 잔존 때문에 구조적으로 현재 상태를 보장할 수 없는 가장 약한 축입니다. 문서는 [`terminals/`](terminals/)에 있습니다.

## 에이전트별 조사

- [Paseo](agents/paseo.md)
- [OpenAI Codex](agents/codex.md)
- [Claude Code](agents/claude-code.md)
- [Antigravity CLI](agents/antigravity-cli.md)
- [Antigravity 2.0](agents/antigravity-2.md)
- [Amp](agents/amp.md)
- [Zed Agent](agents/zed-agent.md)
- [Cursor Agent](agents/cursor-agent.md)
- [GitHub Copilot CLI](agents/github-copilot-cli.md)
- [OpenCode](agents/opencode.md)
- [JetBrains Junie](agents/junie.md)
