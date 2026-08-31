# 에이전트 축

"누가 이 명령을 요청했는가"에 답하는 축입니다. 결과는 `Result.Layers`에 들어갑니다.

```go
result := runby.Detect()
result.Found()                // AI 에이전트가 실행했는가
result.Chain()                // "paseo>codex", 감지 실패 시 "unknown"
result.Agent()                // 최상위 레이어의 Agent
result.Get(runby.AgentCodex)  // (Detection, bool)
```

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

Antigravity CLI, GitHub Copilot CLI, Junie, 일반 OpenCode CLI는 공식적으로 확인된 범용 자식 프로세스 실행 마커가 없어 감지하지 않습니다. 각 제품을 왜 감지하지 않기로 했는지는 [`docs/research/agents/`](../research/agents/)에 제품별로 기록되어 있습니다.

## Kind: 오케스트레이터인가, 하네스인가

`KindOrchestrator`는 하위 에이전트를 관리하며 자신의 에이전트 ID를 광고하는 제품(Paseo, Orca), `KindHarness`는 모델이 요청한 명령을 실행하는 런타임입니다. 둘 다 **AI 에이전트가 이 프로세스를 실행했다는 증거**이므로 `Found()`가 그대로 그 질문의 답이 됩니다.

터미널을 소유한다는 사실은 에이전트 실행의 증거가 아닙니다. Zed는 Agent 전용 신호가 없어 이 축이 아니라 [터미널 축](terminal.md)(`Terminal.Program == TerminalZed`)으로 보고합니다.

## 계층은 여러 개일 수 있습니다

Paseo가 Codex를 구동했다면 `Layers`에 둘 다 들어가고, 명시적인 에이전트 ID를 가진 Paseo가 `Primary()`가 됩니다. 순서는 가장 구체적인 오케스트레이터에서 하위 런타임 방향입니다.

```go
result := runby.Detect()
if codex, ok := result.Get(runby.AgentCodex); ok && codex.Sandbox.Network == runby.NetworkDisabled {
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
		return runby.Detection{AgentID: id, Evidence: runby.PresentNames(env, "ACME_RUN_ID")}, true
	},
}

result := runby.Detect(runby.WithAgentDrivers(acme))
```

자세한 규칙은 [`api.md`](api.md#드라이버-확장)에 있습니다.
