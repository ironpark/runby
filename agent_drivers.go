package runby

import (
	"strconv"
	"strings"
)

// builtinAgentDrivers is ordered from the outermost layer to the innermost:
// orchestrators, then multi-vendor harnesses, then first-party. Result.Primary
// reports the first match, so this order is the precedence contract, and it
// follows the ladder because an outer layer is the more specific answer to
// "what launched this".
//
// This table is the single source of truth for what an Agent is: a product is
// registered here and nowhere else, with the facts no environment can supply —
// what it drives, whose models are behind it, and which binaries it runs as.
// TestKindsMatchDocs holds Kind and Models to the research documents.
var builtinAgentDrivers = []AgentDriver{
	// Orchestrators — drive other agent harnesses.
	{Agent: AgentPaseo, Kind: KindOrchestrator, Models: ModelsDelegated, Executables: []string{"paseo"}, Detect: detectPaseo},
	// Orca's binary shares its name with the GNOME screen reader, so an
	// ancestor running it is not evidence of this orchestrator.
	{Agent: AgentOrca, Kind: KindOrchestrator, Models: ModelsDelegated, Detect: detectOrca},
	// Antigravity 2.0 orchestrates its own harness over its own vendor's
	// models, so unlike Paseo and Orca its models are not delegated. No
	// executable name has been verified against an official source, so there
	// is nothing safe to match in the ancestor chain.
	{Agent: AgentAntigravity2, Kind: KindOrchestrator, Models: ModelsFirstParty, Detect: detectAntigravity2},

	// Multi-vendor — a harness of their own, reaching other vendors' models.
	{Agent: AgentCursor, Kind: KindHarness, Models: ModelsMultiVendor, Executables: []string{"cursor-agent"}, Detect: detectCursor},
	{Agent: AgentOpenCode, Kind: KindHarness, Models: ModelsMultiVendor, Executables: []string{"opencode"}, Detect: detectOpenCode},
	{Agent: AgentAmp, Kind: KindHarness, Models: ModelsMultiVendor, Executables: []string{"amp"}, Detect: detectAmp},
	{Agent: AgentOpenClaw, Kind: KindHarness, Models: ModelsMultiVendor, Executables: []string{"openclaw"}, Detect: detectOpenClaw},
	{Agent: AgentAuggie, Kind: KindHarness, Models: ModelsMultiVendor, Executables: []string{"auggie"}, Detect: detectAuggie},
	// Cline runs inside a code editor rather than as a binary of its own, so
	// there is no ancestor name that would corroborate it.
	{Agent: AgentCline, Kind: KindHarness, Models: ModelsMultiVendor, Detect: detectCline},

	// First-party — a harness built around its own vendor's model.
	{Agent: AgentCodex, Kind: KindHarness, Models: ModelsFirstParty, Executables: []string{"codex"}, Detect: detectCodex},
	{Agent: AgentClaudeCode, Kind: KindHarness, Models: ModelsFirstParty, Executables: []string{"claude"}, Detect: detectClaudeCode},
	{Agent: AgentGeminiCLI, Kind: KindHarness, Models: ModelsFirstParty, Executables: []string{"gemini"}, Detect: detectGeminiCLI},
	{Agent: AgentGrokBuild, Kind: KindHarness, Models: ModelsFirstParty, Executables: []string{"grok"}, Detect: detectGrokBuild},
}

// detectPaseo identifies a process launched by a Paseo agent. PASEO_AGENT_ID is
// set per agent, so it names the logical agent rather than a single session.
func detectPaseo(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	agentID, ok := r.Value("PASEO_AGENT_ID")
	if !ok {
		return Agent{}, false
	}

	workingDirectory, _ := r.Value("PASEO_AGENT_CWD")
	return Agent{
		AgentID: agentID,
		Paths:   Paths{WorkingDirectory: workingDirectory},
		Axis:    Axis{Evidence: r.Evidence()},
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
func detectOrca(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	if !r.Any(orcaMarkers.owner...) || !r.Any(orcaMarkers.location...) {
		return Agent{}, false
	}

	// The pane key identifies the pane a session runs in, which is the closest
	// thing Orca advertises to a session identifier. The terminal handle is
	// the stable name Orca's own CLI and daemon use for a terminal.
	sessionID := r.First("ORCA_PANE_KEY", "ORCA_TERMINAL_HANDLE")
	// Orca prefers the worktree it created over the repository it came from.
	workingDirectory := r.First("ORCA_WORKTREE_PATH", "ORCA_ROOT_PATH")
	dataDirectory, _ := r.Value("ORCA_USER_DATA_PATH")

	return Agent{
		SessionID: sessionID,
		Paths: Paths{
			WorkingDirectory: workingDirectory,
			DataDirectory:    dataDirectory,
		},
		Axis: Axis{
			Confidence: ConfidenceProbable,
			Extra:      r.Extra(orcaExtra),
			Evidence:   r.Evidence(),
		},
	}, true
}

func detectCodex(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	_, hasThreadID := r.Peek("CODEX_THREAD_ID")
	_, hasSessionID := r.Peek("CODEX_SESSION_ID")
	// Codex prefers a thread over a session; both are recorded either way.
	threadID := r.First("CODEX_THREAD_ID", "CODEX_SESSION_ID")
	sandbox, hasSandbox := r.Value("CODEX_SANDBOX")
	ci, hasCI := r.Bool("CODEX_CI")
	networkDisabled, hasNetwork := r.Bool("CODEX_SANDBOX_NETWORK_DISABLED")
	if !hasThreadID && !hasSessionID && !hasSandbox && !(hasCI && ci) {
		return Agent{}, false
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

	return Agent{
		SessionID: threadID,
		Sandbox:   Sandbox{Mode: sandbox, Network: network},
		Axis: Axis{
			Confidence: confidence,
			Extra:      extra,
			Evidence:   r.Evidence(),
		},
	}, true
}

func detectClaudeCode(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	claudeCode := r.IsTrue("CLAUDECODE")
	sessionID, hasSessionID := r.Value("CLAUDE_CODE_SESSION_ID")
	// AI_AGENT is shared with other tooling, so it is evidence only when its
	// value names Claude Code: it is peeked, and recorded once the value has
	// decided. Every other variable here is recorded by the act of reading it.
	aiAgent, _ := r.Peek("AI_AGENT")
	isAIAgent := strings.HasPrefix(strings.ToLower(aiAgent), "claude-code")
	if isAIAgent {
		r.Record("AI_AGENT")
	}
	if !claudeCode && !hasSessionID && !isAIAgent {
		return Agent{}, false
	}

	entrypoint, _ := r.Value("CLAUDE_CODE_ENTRYPOINT")
	nested, _ := r.Bool("CLAUDE_CODE_CHILD_SESSION")
	return Agent{
		SessionID:  sessionID,
		Entrypoint: entrypoint,
		Nested:     nested,
		Axis:       Axis{Evidence: r.Evidence()},
	}, true
}

func detectCursor(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	if _, ok := r.Value("CURSOR_AGENT"); !ok {
		return Agent{}, false
	}
	return Agent{Axis: Axis{Evidence: r.Evidence()}}, true
}

// detectOpenCode identifies a process OpenCode launched. OPENCODE is set on
// the process environment during CLI startup, before the subcommand runs, so
// every child of every invocation inherits it.
//
// OPENCODE_CLIENT is subtler and only counts when it says "acp". The ACP
// command sets it to that, but everywhere else OpenCode READS it — with a
// default of "cli" — as the name an embedding host gives itself. Treating any
// value as proof would report the agent for a variable the user can export.
//
// AGENT is set alongside OPENCODE and is deliberately ignored: the name
// belongs to no vendor, and Goose sets it too.
func detectOpenCode(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	launched := r.IsTrue("OPENCODE")
	acp := r.EqualsFold("OPENCODE_CLIENT", "acp")
	if !launched && !acp {
		return Agent{}, false
	}

	detection := Agent{}
	if acp {
		detection.Entrypoint = "acp"
	}
	// Without OPENCODE this is the ACP marker alone, which an embedding host
	// could also have set on its way in.
	if !launched {
		detection.Confidence = ConfidenceProbable
	}
	detection.Evidence = r.Evidence()
	return detection, true
}

func detectAmp(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	orb := r.IsTrue("AMP_ORB")
	threadID, hasThreadID := r.Value("AMP_THREAD_ID")
	if !orb && !hasThreadID {
		return Agent{}, false
	}

	entrypoint := "orb"
	if hasThreadID {
		entrypoint = "orb-service"
	}
	return Agent{
		SessionID:  threadID,
		Entrypoint: entrypoint,
		Axis:       Axis{Evidence: r.Evidence()},
	}, true
}

// detectAntigravity2 identifies a sidecar whose lifecycle Antigravity 2.0
// manages. Antigravity CLI sets no general execution marker and is not detected.
func detectAntigravity2(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	dataDirectory, ok := r.Value("ANTIGRAVITY_EXECUTABLE_DATA_DIR")
	if !ok {
		return Agent{}, false
	}
	return Agent{
		Entrypoint: "sidecar",
		Paths:      Paths{DataDirectory: dataDirectory},
		Axis:       Axis{Evidence: r.Evidence()},
	}, true
}

// detectGeminiCLI identifies a process Gemini CLI launched. The CLI names the
// variable its identification marker in source and sets it on every shell
// command and stdio MCP server it starts, so its presence is proof rather than
// a supporting signal.
func detectGeminiCLI(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	if !r.IsTrue("GEMINI_CLI") {
		return Agent{}, false
	}
	return Agent{Axis: Axis{Evidence: r.Evidence()}}, true
}

// detectGrokBuild identifies a Grok Build plugin hook. The documented scope is
// narrow — only hooks receive these variables, not the shell commands the agent
// runs — so a Grok Build session that never fires a hook is not reported.
func detectGrokBuild(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	if !r.Any("GROK_PLUGIN_ROOT", "GROK_PLUGIN_DATA") {
		return Agent{}, false
	}
	return Agent{
		Entrypoint: "plugin-hook",
		Axis:       Axis{Evidence: r.Evidence()},
	}, true
}

// detectAuggie identifies a command Auggie ran through its launch-process tool.
// The official reference documents the variable for exactly this purpose.
func detectAuggie(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	if !r.IsTrue("AUGMENT_AGENT") {
		return Agent{}, false
	}
	return Agent{Axis: Axis{Evidence: r.Evidence()}}, true
}

// detectOpenClaw identifies a process OpenClaw spawned. The variable carries
// which runtime started it, so the value becomes the entrypoint rather than
// being discarded.
func detectOpenClaw(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	entrypoint, ok := r.Value("OPENCLAW_SHELL")
	if !ok {
		return Agent{}, false
	}
	return Agent{
		Entrypoint: strings.ToLower(entrypoint),
		Axis:       Axis{Evidence: r.Evidence()},
	}, true
}

// detectCline identifies a terminal Cline created. Cline sets the marker on the
// terminals it opens to run its commands, but a human can type into the same
// terminal, so this is a supporting signal rather than proof that the agent
// issued the command.
func detectCline(env Env) (Agent, bool) {
	r := NewEnvReader(env)
	if !r.IsTrue("CLINE_ACTIVE") {
		return Agent{}, false
	}
	return Agent{
		Axis: Axis{
			Confidence: ConfidenceProbable,
			Evidence:   r.Evidence(),
		},
	}, true
}
