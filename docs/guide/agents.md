# 에이전트 축

"누가 이 명령을 요청했는가"에 답하는 축입니다. 결과는 `Result.Layers`에 들어갑니다.

에이전트 여부만 필요하면 `IsAgent()`, 로그에 계층 하나를 남기려면 `Chain()`이면 충분합니다. 제품 분류와 신뢰도를 해석해야 할 때 나머지 내용을 참고하세요.

```go
result := runby.Detect()
result.IsAgent()               // AI 에이전트가 실행했는가
result.Chain()                 // "paseo>codex", 감지 실패 시 "unknown"
result.Agent()                 // 최상위 레이어의 Agent
result.Layer(runby.AgentCodex) // (Detection, bool)
```

API 키와 일반 설정 변수는 에이전트가 프로세스를 실행했다는 증거가 아니므로 감지에 사용하지 않습니다.

## 감지 대상

표는 감지 순서, 즉 바깥 계층부터입니다.

| 에이전트 | `Agent` | `Level` | `Kind` | `Models` | 식별 신호 |
|---|---|---|---|---|---|
| Paseo | `AgentPaseo` | `l3` | `orchestrator` | `delegated` | `PASEO_AGENT_ID`, `PASEO_AGENT_CWD` |
| Orca (Stably AI) | `AgentOrca` | `l3` | `orchestrator` | `delegated` | `ORCA_PANE_KEY` 또는 `ORCA_TERMINAL_HANDLE` + `ORCA_TAB_ID` 또는 `ORCA_WORKTREE_ID` |
| Antigravity 2.0 sidecar | `AgentAntigravity2` | `l3` | `orchestrator` | `first-party` | `ANTIGRAVITY_EXECUTABLE_DATA_DIR` |
| Cursor Agent | `AgentCursor` | `l2` | `harness` | `multi-vendor` | `CURSOR_AGENT` |
| OpenCode ACP | `AgentOpenCode` | `l2` | `harness` | `multi-vendor` | `OPENCODE_CLIENT=acp` |
| Amp Orb / Orb 관리형 서비스 | `AgentAmp` | `l2` | `harness` | `multi-vendor` | `AMP_ORB`, `AMP_THREAD_ID` |
| OpenClaw | `AgentOpenClaw` | `l2` | `harness` | `multi-vendor` | `OPENCLAW_SHELL` (값이 `Entrypoint`가 됨) |
| Auggie (Augment Code) | `AgentAuggie` | `l2` | `harness` | `multi-vendor` | `AUGMENT_AGENT` |
| Cline | `AgentCline` | `l2` | `harness` | `multi-vendor` | `CLINE_ACTIVE` (보조 신호 — 아래 참조) |
| OpenAI Codex | `AgentCodex` | `l1` | `harness` | `first-party` | `CODEX_THREAD_ID`, `CODEX_SESSION_ID`, 샌드박스 관련 변수 |
| Claude Code | `AgentClaudeCode` | `l1` | `harness` | `first-party` | `CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`, `AI_AGENT=claude-code*` |
| Gemini CLI | `AgentGeminiCLI` | `l1` | `harness` | `first-party` | `GEMINI_CLI` |
| Grok Build 플러그인 훅 | `AgentGrokBuild` | `l1` | `harness` | `first-party` | `GROK_PLUGIN_ROOT`, `GROK_PLUGIN_DATA` |

**Cline은 `probable`입니다.** 마커가 프로세스가 아니라 Cline이 만든 **터미널**에 붙어 있어, 사람이 그 터미널에 직접 타이핑한 명령도 같은 값을 받기 때문입니다. **Grok Build는 플러그인 훅에만** 마커가 주어지므로, 훅을 실행하지 않은 세션은 감지되지 않습니다.

Antigravity CLI, GitHub Copilot CLI, Junie, Goose, Kimi CLI, Claude Cowork, 일반 OpenCode CLI는 공식적으로 확인된 범용 자식 프로세스 실행 마커가 없어 감지하지 않습니다. 각 제품을 왜 감지하지 않기로 했는지는 [`docs/research/agents/`](../research/agents/)에 제품별로 기록되어 있습니다 — 설정 변수를 마커로 오인한 사례(Goose의 `GOOSE_PROVIDER`)와 공식 문서에 없는 내부 변수(Cowork의 `CLAUDE_CODE_IS_COWORK`)를 포함합니다.

## 분류: Kind, Models, Level

제품 분류는 **서로 독립적인 두 축**이고, `Level`은 그 둘에서 파생된 읽기 편한 사다리입니다.

| | `Kind` — 무엇을 구동하는가 | `Models` — 지능은 어디서 오는가 |
|---|---|---|
| `l1` | `harness` (모델을 구동) | `first-party` (자사 모델) |
| `l2` | `harness` | `multi-vendor` (타사 모델을 등록/API로) |
| `l3` | `orchestrator` (다른 하네스를 구동) | 보통 `delegated`, 그러나 항상은 아님 |

**두 축을 하나로 합칠 수 없는 이유**가 마지막 줄에 있습니다. Antigravity 2.0은 Paseo·Orca처럼 하네스를 구동하는 오케스트레이터지만, 그 하네스와 모델이 모두 자사 것입니다 — `(orchestrator, first-party)`. 단일 사다리에는 이 칸이 없고, 한 벤더가 하네스와 그 위의 오케스트레이터를 함께 내놓는 건 일회성이 아니라 패턴입니다. 그래서 `Level`은 로그·집계용 라벨이고, **"누구 모델인가"의 답은 언제나 `Models`** 입니다.

```go
result.Agent().Level()   // Level, 미지원 에이전트는 LevelUnknown
result.Agent().Kind()    // Kind
result.Agent().Models()  // ModelSource
```

`Models`는 **제품의 성격**이지 이번 실행에 실제로 답한 모델이 아닙니다. `first-party` 하네스도 설정으로 타사 엔드포인트를 바라보게 할 수 있고, 이 패키지가 읽는 환경변수는 그 사실을 말해주지 않습니다.

셋 중 무엇이든 **AI 에이전트가 이 프로세스를 실행했다는 증거**이므로 `IsAgent()`가 그대로 그 질문의 답이 됩니다.

터미널을 소유한다는 사실은 에이전트 실행의 증거가 아닙니다. Zed는 Agent 전용 신호가 없어 이 축이 아니라 [터미널 축](terminal.md)(`Terminal.Program == TerminalZed`)으로 보고합니다.

`Kind`와 `Models`는 환경이 알려줄 수 없는 사실이라 드라이버 테이블에 손으로 적고 [`docs/research/agents/`](../research/agents/)의 `product_type`·`model_source`에도 손으로 적습니다. 두 곳이 어긋나지 않도록 `TestKindsMatchDocs`가 잠급니다.

## 계층은 여러 개일 수 있습니다

Paseo가 Codex를 구동했다면 `Layers`에 둘 다 들어가고, 바깥 계층인 Paseo가 `Primary()`가 됩니다. 순서는 사다리를 따라 `l3` → `l2` → `l1`, 즉 바깥에서 안쪽 방향입니다.

```go
result := runby.Detect()
if codex, ok := result.Layer(runby.AgentCodex); ok && codex.Sandbox.Network == runby.NetworkDisabled {
	skipNetworkTests()
}
```

## Confidence

`ConfidenceDefinite`는 제품이 **자신이 실행한 프로세스에 한해** 설정하는 실행 마커이고, `ConfidenceProbable`은 에이전트 실행과 모순되지 않지만 그것만의 신호는 아닌 보조 신호입니다.

Orca가 항상 `probable`인 것이 그 예입니다. Orca는 자신이 호스팅하는 pane에 표시를 남기므로, 사용자가 Orca 터미널에 직접 입력한 명령과 Orca가 실행한 에이전트가 호출한 명령이 같은 변수를 갖습니다. 실제로 실행한 에이전트는 해당 harness의 고유 신호로 별도 계층에 보고됩니다.

## 사내 에이전트 추가

```go
acme := runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Kind:        runby.KindOrchestrator,
	Executables: []string{"acme-run"},
	Detect: func(env runby.Env) (runby.Detection, bool) {
		id, ok := runby.Value(env, "ACME_RUN_ID")
		if !ok {
			return runby.Detection{}, false
		}
		return runby.Detection{AgentID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_RUN_ID")}}, true
	},
}

result := runby.Detect(runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), acme)...))
```

자세한 규칙은 [`api.md`](api.md#드라이버-확장)에 있습니다.
