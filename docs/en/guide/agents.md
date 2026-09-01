# Agent axis

The agent axis answers “who requested this command?” Results are stored in `Result.Agents`.

```go
result := runby.Detect()
result.IsAgent()
result.Chain()
result.Primary()
result.Agent(runby.AgentCodex)
```

API keys and general configuration variables are not evidence that an agent launched the process.

## Detected agents

| Agent | `Name` | `Kind` | `Models` | Main signal |
|---|---|---|---|---|
| Paseo | `AgentPaseo` | `orchestrator` | `delegated` | `PASEO_AGENT_ID`, `PASEO_AGENT_CWD` |
| Orca | `AgentOrca` | `orchestrator` | `delegated` | Orca pane or terminal ownership markers |
| Antigravity 2.0 | `AgentAntigravity2` | `orchestrator` | `first-party` | `ANTIGRAVITY_EXECUTABLE_DATA_DIR` |
| Cursor Agent | `AgentCursor` | `harness` | `multi-vendor` | `CURSOR_AGENT` |
| OpenCode | `AgentOpenCode` | `harness` | `multi-vendor` | `OPENCODE`, `OPENCODE_CLIENT=acp` |
| Amp | `AgentAmp` | `harness` | `multi-vendor` | `AMP_ORB`, `AMP_THREAD_ID` |
| OpenClaw | `AgentOpenClaw` | `harness` | `multi-vendor` | `OPENCLAW_SHELL` |
| Auggie | `AgentAuggie` | `harness` | `multi-vendor` | `AUGMENT_AGENT` |
| pi | `AgentPi` | `harness` | `multi-vendor` | `PI_SESSION_ID` |
| Charm Crush | `AgentCrush` | `harness` | `multi-vendor` | `CRUSH=1` |
| Roo Code | `AgentRooCode` | `harness` | `multi-vendor` | `ROO_ACTIVE=true` |
| OpenHands | `AgentOpenHands` | `harness` | `multi-vendor` | `AI_AGENT=openhands` |
| Cline | `AgentCline` | `harness` | `multi-vendor` | `CLINE_ACTIVE` |
| OpenAI Codex | `AgentCodex` | `harness` | `first-party` | Codex session and sandbox markers |
| Claude Code | `AgentClaudeCode` | `harness` | `first-party` | Claude Code execution markers |
| Gemini CLI | `AgentGeminiCLI` | `harness` | `first-party` | `GEMINI_CLI` |
| Grok Build | `AgentGrokBuild` | `harness` | `first-party` | Grok plugin-hook markers |
| Qwen Code | `AgentQwenCode` | `harness` | `first-party` | `QWEN_CODE=1` |
| DeepSeek Harness | `AgentDeepSeekHarness` | `harness` | `first-party` | `DSH_SHELL=1` |

Cline and Roo Code are `probable` because their markers describe a product-owned terminal where a person may also type. Grok Build is visible only in plugin hooks. Generic variables such as `AGENT` and `AI_AGENT` are used only when their value names a supported product exactly.

Products without an officially verified, general child-process execution marker are intentionally excluded. Product-by-product rationale is maintained in the [agent research](../../research/agents/).

## `Kind` and `Models`

`Kind` answers what the product drives: a model (`harness`) or another harness (`orchestrator`). `Models` answers where intelligence comes from: `first-party`, `multi-vendor`, or `delegated`.

These classifications describe the product, not necessarily the model selected for this particular run.

## Multiple layers

If Paseo launches Codex, both layers appear and Paseo is `Primary()`. Ordering derives from the classifications: orchestrators, multi-vendor harnesses, then first-party harnesses.

## Session and agent identifiers

Several layers can advertise different session identifiers. `SessionID()` and `AgentID()` scan from outermost to innermost and return an `Identifier` containing both the value and the source agent. Use `Agent(name)` when you need one product's identifier.

Identifiers are omitted from the human-readable CLI report but included in `-json`.

## Confidence

`ConfidenceDefinite` means the product sets an execution marker on processes it launched. `ConfidenceProbable` means the signal is compatible with agent execution but not exclusive to it. A live `AncestorPID` is the strongest corroboration available.

## Add a user-defined agent driver

```go
acme := runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Kind:        runby.KindOrchestrator,
	Executables: []string{"acme-run"},
	Detect: func(env runby.Env) (runby.Agent, bool) {
		r := runby.NewEnvReader(env)
		id, ok := r.Value("ACME_RUN_ID")
		if !ok { return runby.Agent{}, false }
		return runby.Agent{AgentID: id, Axis: runby.Axis{Evidence: r.Evidence()}}, true
	},
}
```

See [Writing drivers](drivers.md) for complete rules.
