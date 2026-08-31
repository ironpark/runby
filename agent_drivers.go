package runby

import (
	"strconv"
	"strings"
)

// builtinAgentDrivers is ordered from the most specific orchestrator to the
// underlying runtime. Result.Primary reports the first match, so this order is
// the precedence contract.
//
// This table is the single source of truth for what an Agent is: a product is
// registered here and nowhere else.
var builtinAgentDrivers = []AgentDriver{
	{Agent: AgentPaseo, Kind: KindOrchestrator, Executables: []string{"paseo"}, Detect: detectPaseo},
	// Orca's binary shares its name with the GNOME screen reader, so an
	// ancestor running it is not evidence of this orchestrator.
	{Agent: AgentOrca, Kind: KindOrchestrator, Detect: detectOrca},
	{Agent: AgentCodex, Kind: KindHarness, Executables: []string{"codex"}, Detect: detectCodex},
	{Agent: AgentClaudeCode, Kind: KindHarness, Executables: []string{"claude"}, Detect: detectClaudeCode},
	{Agent: AgentCursor, Kind: KindHarness, Executables: []string{"cursor-agent"}, Detect: detectCursor},
	{Agent: AgentOpenCode, Kind: KindHarness, Executables: []string{"opencode"}, Detect: detectOpenCode},
	{Agent: AgentAmp, Kind: KindHarness, Executables: []string{"amp"}, Detect: detectAmp},
	// No Antigravity executable name has been verified against an official
	// source, so there is nothing safe to match in the ancestor chain.
	{Agent: AgentAntigravity2, Kind: KindHarness, Detect: detectAntigravity2},
}

// AgentDrivers returns the built-in agent drivers in precedence order. The
// returned slice is a copy and may be reordered, filtered, or adjusted before
// being passed back through WithOnlyAgentDrivers.
func AgentDrivers() []AgentDriver { return cloneSlice(builtinAgentDrivers) }

// detectPaseo identifies a process launched by a Paseo agent. PASEO_AGENT_ID is
// set per agent, so it names the logical agent rather than a single session.
func detectPaseo(env Env) (Detection, bool) {
	r := newReader(env)
	agentID, ok := r.value("PASEO_AGENT_ID")
	if !ok {
		return Detection{}, false
	}

	workingDirectory, _ := r.value("PASEO_AGENT_CWD")
	return Detection{
		AgentID: agentID,
		Paths:   Paths{WorkingDirectory: workingDirectory},
		Axis:    Axis{Evidence: r.evidence()},
	}, true
}

// orcaMarkers are the variables Orca stamps on every pane it hosts. One from
// each pair must be present: a pane or terminal handle says Orca owns the
// stream, a tab or worktree says where. Requiring both keeps a single stray
// ORCA_ variable from standing in for a detection.
var orcaMarkers = struct{ owner, location []string }{
	owner:    []string{"ORCA_PANE_KEY", "ORCA_TERMINAL_HANDLE"},
	location: []string{"ORCA_TAB_ID", "ORCA_WORKTREE_ID"},
}

// orcaExtra maps Orca's context variables to their stable Extra keys. Orca
// sets these only in specific situations — the hook variables during worktree
// setup and archive scripts, the host variables only across a WSL boundary,
// ORCA_CODEX_HOME only for an Orca-managed Codex — so they are context and
// never grounds for the detection itself.
var orcaExtra = map[string]string{
	"orca.tab_id":           "ORCA_TAB_ID",
	"orca.worktree_id":      "ORCA_WORKTREE_ID",
	"orca.terminal_handle":  "ORCA_TERMINAL_HANDLE",
	"orca.workspace_name":   "ORCA_WORKSPACE_NAME",
	"orca.root_path":        "ORCA_ROOT_PATH",
	"orca.worktree_path":    "ORCA_WORKTREE_PATH",
	"orca.codex_home":       "ORCA_CODEX_HOME",
	"orca.host_kind":        "ORCA_ORCHESTRATION_COMPATIBILITY_HOST_KIND",
	"orca.host_id":          "ORCA_ORCHESTRATION_COMPATIBILITY_HOST_ID",
	"orca.host_incarnation": "ORCA_ORCHESTRATION_COMPATIBILITY_HOST_INCARNATION",
}

// detectOrca identifies a process running in a terminal hosted by Orca
// (github.com/stablyai/orca), a desktop orchestrator that runs CLI agents in
// terminals backed by git worktrees.
//
// Orca stamps the pane, not the agent invocation. A match proves Orca owns the
// terminal but not that an agent rather than a person typed the command, so
// the confidence is never definite. When Orca did launch an agent, that agent
// sets its own marker and is reported as its own layer.
func detectOrca(env Env) (Detection, bool) {
	r := newReader(env)
	if !r.any(orcaMarkers.owner...) || !r.any(orcaMarkers.location...) {
		return Detection{}, false
	}

	// The pane key identifies the pane a session runs in, which is the closest
	// thing Orca advertises to a session identifier. The terminal handle is
	// the stable name Orca's own CLI and daemon use for a terminal.
	sessionID := r.first("ORCA_PANE_KEY", "ORCA_TERMINAL_HANDLE")
	// Orca prefers the worktree it created over the repository it came from.
	workingDirectory := r.first("ORCA_WORKTREE_PATH", "ORCA_ROOT_PATH")
	dataDirectory, _ := r.value("ORCA_USER_DATA_PATH")

	return Detection{
		SessionID: sessionID,
		Paths: Paths{
			WorkingDirectory: workingDirectory,
			DataDirectory:    dataDirectory,
		},
		Axis: Axis{
			Confidence: ConfidenceProbable,
			Extra:      r.extra(orcaExtra),
			Evidence:   r.evidence(),
		},
	}, true
}

func detectCodex(env Env) (Detection, bool) {
	r := newReader(env)
	_, hasThreadID := r.peek("CODEX_THREAD_ID")
	_, hasSessionID := r.peek("CODEX_SESSION_ID")
	// Codex prefers a thread over a session; both are recorded either way.
	threadID := r.first("CODEX_THREAD_ID", "CODEX_SESSION_ID")
	sandbox, hasSandbox := r.value("CODEX_SANDBOX")
	ci, hasCI := r.boolean("CODEX_CI")
	networkDisabled, hasNetwork := r.boolean("CODEX_SANDBOX_NETWORK_DISABLED")
	if !hasThreadID && !hasSessionID && !hasSandbox && !(hasCI && ci) {
		return Detection{}, false
	}

	// A thread or session identifier is set per Codex conversation, so it is an
	// execution marker. The sandbox and CI variables describe the environment
	// Codex was configured with and could plausibly be set by other tooling.
	confidence := ConfidenceProbable
	if hasThreadID || hasSessionID {
		confidence = ConfidenceDefinite
	}

	network := NetworkUnknown
	if hasNetwork {
		network = NetworkEnabled
		if networkDisabled {
			network = NetworkDisabled
		}
	}

	// The value is the parsed boolean rather than the raw one, so that "1" and
	// "true" reach a caller in the same form.
	var extra map[string]string
	if hasCI {
		extra = map[string]string{"codex.ci": strconv.FormatBool(ci)}
	}

	return Detection{
		SessionID: threadID,
		Sandbox:   Sandbox{Mode: sandbox, Network: network},
		Axis: Axis{
			Confidence: confidence,
			Extra:      extra,
			Evidence:   r.evidence(),
		},
	}, true
}

func detectClaudeCode(env Env) (Detection, bool) {
	r := newReader(env)
	claudeCode := r.isTrue("CLAUDECODE")
	sessionID, hasSessionID := r.value("CLAUDE_CODE_SESSION_ID")
	// AI_AGENT is shared with other tooling, so it is evidence only when its
	// value names Claude Code: it is peeked, and recorded once the value has
	// decided. Every other variable here is recorded by the act of reading it.
	aiAgent, _ := r.peek("AI_AGENT")
	isAIAgent := strings.HasPrefix(strings.ToLower(aiAgent), "claude-code")
	if isAIAgent {
		r.record("AI_AGENT")
	}
	if !claudeCode && !hasSessionID && !isAIAgent {
		return Detection{}, false
	}

	entrypoint, _ := r.value("CLAUDE_CODE_ENTRYPOINT")
	nested, _ := r.boolean("CLAUDE_CODE_CHILD_SESSION")
	return Detection{
		SessionID:  sessionID,
		Entrypoint: entrypoint,
		Nested:     nested,
		Axis:       Axis{Evidence: r.evidence()},
	}, true
}

func detectCursor(env Env) (Detection, bool) {
	r := newReader(env)
	if _, ok := r.value("CURSOR_AGENT"); !ok {
		return Detection{}, false
	}
	return Detection{Axis: Axis{Evidence: r.evidence()}}, true
}

// detectOpenCode identifies OpenCode running as an ACP client. OpenCode has no
// general execution marker, so this covers ACP invocations only and is reported
// as a supporting signal rather than proof.
func detectOpenCode(env Env) (Detection, bool) {
	r := newReader(env)
	if !r.equalsFold("OPENCODE_CLIENT", "acp") {
		return Detection{}, false
	}
	return Detection{
		Entrypoint: "acp",
		Axis: Axis{
			Confidence: ConfidenceProbable,
			Evidence:   r.evidence(),
		},
	}, true
}

func detectAmp(env Env) (Detection, bool) {
	r := newReader(env)
	orb := r.isTrue("AMP_ORB")
	threadID, hasThreadID := r.value("AMP_THREAD_ID")
	if !orb && !hasThreadID {
		return Detection{}, false
	}

	entrypoint := "orb"
	if hasThreadID {
		entrypoint = "orb-service"
	}
	return Detection{
		SessionID:  threadID,
		Entrypoint: entrypoint,
		Axis:       Axis{Evidence: r.evidence()},
	}, true
}

// detectAntigravity2 identifies a sidecar whose lifecycle Antigravity 2.0
// manages. Antigravity CLI sets no general execution marker and is not detected.
func detectAntigravity2(env Env) (Detection, bool) {
	r := newReader(env)
	dataDirectory, ok := r.value("ANTIGRAVITY_EXECUTABLE_DATA_DIR")
	if !ok {
		return Detection{}, false
	}
	return Detection{
		Entrypoint: "sidecar",
		Paths:      Paths{DataDirectory: dataDirectory},
		Axis:       Axis{Evidence: r.evidence()},
	}, true
}
