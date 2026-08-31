package runby

import "strings"

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
func AgentDrivers() []AgentDriver {
	drivers := make([]AgentDriver, len(builtinAgentDrivers))
	copy(drivers, builtinAgentDrivers)
	return drivers
}

// detectPaseo identifies a process launched by a Paseo agent. PASEO_AGENT_ID is
// set per agent, so it names the logical agent rather than a single session.
func detectPaseo(env Env) (Detection, bool) {
	agentID, ok := Value(env, "PASEO_AGENT_ID")
	if !ok {
		return Detection{}, false
	}

	workingDirectory, _ := Value(env, "PASEO_AGENT_CWD")
	return Detection{
		AgentID:  agentID,
		Paths:    Paths{WorkingDirectory: workingDirectory},
		Evidence: PresentNames(env, "PASEO_AGENT_ID", "PASEO_AGENT_CWD"),
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

// orcaNames is every variable detectOrca consults, so Evidence and the lookup
// set cannot drift apart. The marker variables appear in orcaExtra too;
// PresentNames deduplicates the union.
var orcaNames = func() []string {
	names := append([]string{"ORCA_USER_DATA_PATH"}, orcaMarkers.owner...)
	names = append(names, orcaMarkers.location...)
	for _, name := range orcaExtra {
		names = append(names, name)
	}
	return names
}()

// detectOrca identifies a process running in a terminal hosted by Orca
// (github.com/stablyai/orca), a desktop orchestrator that runs CLI agents in
// terminals backed by git worktrees.
//
// Orca stamps the pane, not the agent invocation. A match proves Orca owns the
// terminal but not that an agent rather than a person typed the command, so
// the confidence is never definite. When Orca did launch an agent, that agent
// sets its own marker and is reported as its own layer.
func detectOrca(env Env) (Detection, bool) {
	if !AnyPresent(env, orcaMarkers.owner...) || !AnyPresent(env, orcaMarkers.location...) {
		return Detection{}, false
	}

	// The pane key identifies the pane a session runs in, which is the closest
	// thing Orca advertises to a session identifier. The terminal handle is
	// the stable name Orca's own CLI and daemon use for a terminal.
	sessionID, ok := Value(env, "ORCA_PANE_KEY")
	if !ok {
		sessionID, _ = Value(env, "ORCA_TERMINAL_HANDLE")
	}

	// Orca prefers the worktree it created over the repository it came from.
	workingDirectory, ok := Value(env, "ORCA_WORKTREE_PATH")
	if !ok {
		workingDirectory, _ = Value(env, "ORCA_ROOT_PATH")
	}
	dataDirectory, _ := Value(env, "ORCA_USER_DATA_PATH")

	return Detection{
		Confidence: ConfidenceProbable,
		SessionID:  sessionID,
		Paths: Paths{
			WorkingDirectory: workingDirectory,
			DataDirectory:    dataDirectory,
		},
		Extra:    CollectExtra(env, orcaExtra),
		Evidence: PresentNames(env, orcaNames...),
	}, true
}

func detectCodex(env Env) (Detection, bool) {
	threadID, hasThreadID := Value(env, "CODEX_THREAD_ID")
	sessionID, hasSessionID := Value(env, "CODEX_SESSION_ID")
	sandbox, hasSandbox := Value(env, "CODEX_SANDBOX")
	ci, hasCI := Bool(env, "CODEX_CI")
	networkDisabled, hasNetwork := Bool(env, "CODEX_SANDBOX_NETWORK_DISABLED")
	if !hasThreadID && !hasSessionID && !hasSandbox && !(hasCI && ci) {
		return Detection{}, false
	}

	if threadID == "" {
		threadID = sessionID
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

	var extra map[string]string
	if hasCI {
		extra = map[string]string{"codex.ci": boolString(ci)}
	}

	return Detection{
		Confidence: confidence,
		SessionID:  threadID,
		Sandbox:    Sandbox{Mode: sandbox, Network: network},
		Extra:      extra,
		Evidence:   PresentNames(env, "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_SANDBOX", "CODEX_SANDBOX_NETWORK_DISABLED", "CODEX_CI"),
	}, true
}

func detectClaudeCode(env Env) (Detection, bool) {
	sessionID, hasSessionID := Value(env, "CLAUDE_CODE_SESSION_ID")
	aiAgent, hasAIAgent := Value(env, "AI_AGENT")
	isAIAgent := hasAIAgent && strings.HasPrefix(strings.ToLower(aiAgent), "claude-code")
	if !IsTrue(env, "CLAUDECODE") && !hasSessionID && !isAIAgent {
		return Detection{}, false
	}

	entrypoint, _ := Value(env, "CLAUDE_CODE_ENTRYPOINT")
	childSession, _ := Bool(env, "CLAUDE_CODE_CHILD_SESSION")
	// AI_AGENT is only evidence when its value names Claude Code, so it joins
	// the candidate list conditionally; PresentNames still does the sorting.
	names := []string{"CLAUDECODE", "CLAUDE_CODE_SESSION_ID", "CLAUDE_CODE_ENTRYPOINT", "CLAUDE_CODE_CHILD_SESSION"}
	if isAIAgent {
		names = append(names, "AI_AGENT")
	}
	return Detection{
		SessionID:  sessionID,
		Entrypoint: entrypoint,
		Nested:     childSession,
		Evidence:   PresentNames(env, names...),
	}, true
}

func detectCursor(env Env) (Detection, bool) {
	if _, ok := Value(env, "CURSOR_AGENT"); !ok {
		return Detection{}, false
	}
	return Detection{
		Evidence: []string{"CURSOR_AGENT"},
	}, true
}

// detectOpenCode identifies OpenCode running as an ACP client. OpenCode has no
// general execution marker, so this covers ACP invocations only and is reported
// as a supporting signal rather than proof.
func detectOpenCode(env Env) (Detection, bool) {
	if !EqualsFold(env, "OPENCODE_CLIENT", "acp") {
		return Detection{}, false
	}
	return Detection{
		Confidence: ConfidenceProbable,
		Entrypoint: "acp",
		Evidence:   []string{"OPENCODE_CLIENT"},
	}, true
}

func detectAmp(env Env) (Detection, bool) {
	threadID, hasThreadID := Value(env, "AMP_THREAD_ID")
	if !IsTrue(env, "AMP_ORB") && !hasThreadID {
		return Detection{}, false
	}

	entrypoint := "orb"
	if hasThreadID {
		entrypoint = "orb-service"
	}
	return Detection{
		SessionID:  threadID,
		Entrypoint: entrypoint,
		Evidence:   PresentNames(env, "AMP_ORB", "AMP_THREAD_ID"),
	}, true
}

// detectAntigravity2 identifies a sidecar whose lifecycle Antigravity 2.0
// manages. Antigravity CLI sets no general execution marker and is not detected.
func detectAntigravity2(env Env) (Detection, bool) {
	dataDirectory, ok := Value(env, "ANTIGRAVITY_EXECUTABLE_DATA_DIR")
	if !ok {
		return Detection{}, false
	}
	return Detection{
		Entrypoint: "sidecar",
		Paths:      Paths{DataDirectory: dataDirectory},
		Evidence:   []string{"ANTIGRAVITY_EXECUTABLE_DATA_DIR"},
	}, true
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
