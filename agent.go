package runby

// Agent identifies a supported agent runtime, orchestrator, or host.
type Agent string

const (
	AgentUnknown      Agent = "unknown"
	AgentPaseo        Agent = "paseo"
	AgentCodex        Agent = "codex"
	AgentClaudeCode   Agent = "claude-code"
	AgentAntigravity2 Agent = "antigravity-2.0"
	AgentAmp          Agent = "amp"
	AgentCursor       Agent = "cursor-agent"
	AgentOpenCode     Agent = "opencode"
	AgentZed          Agent = "zed"
)

// Kind separates products that prove an AI agent launched the process from
// products that only prove which application owns the terminal. It mirrors the
// product_type field recorded in docs/agents.
type Kind string

const (
	KindUnknown Kind = "unknown"
	// KindOrchestrator manages other agents and advertises its own agent
	// identity, so it is reported ahead of the runtime it drives.
	KindOrchestrator Kind = "orchestrator"
	// KindHarness is an agent runtime that executes model-requested commands.
	KindHarness Kind = "harness"
	// KindHost owns the terminal but does not prove that an agent, rather than
	// a person, requested the command.
	KindHost Kind = "host"
)

// agentKinds is the single source of truth for Agent classification. Agents
// added here must also be added to detectors in runby.go.
var agentKinds = map[Agent]Kind{
	AgentPaseo:        KindOrchestrator,
	AgentCodex:        KindHarness,
	AgentClaudeCode:   KindHarness,
	AgentAntigravity2: KindHarness,
	AgentAmp:          KindHarness,
	AgentCursor:       KindHarness,
	AgentOpenCode:     KindHarness,
	AgentZed:          KindHost,
}

// Kind reports how much a detection of a proves. It returns KindUnknown for
// agents this package does not support.
func (a Agent) Kind() Kind {
	if kind, ok := agentKinds[a]; ok {
		return kind
	}
	return KindUnknown
}

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (a Agent) String() string {
	if a == "" {
		return string(AgentUnknown)
	}
	return string(a)
}

// Agents returns every supported agent in detection precedence order.
func Agents() []Agent {
	agents := make([]Agent, 0, len(builtinDetectors))
	for _, detector := range builtinDetectors {
		agents = append(agents, detector.Agent())
	}
	return agents
}

// Confidence records how directly the evidence ties the process to the agent.
type Confidence string

const (
	ConfidenceUnknown Confidence = "unknown"
	// ConfidenceDefinite means the evidence is an execution marker that the
	// product sets specifically for processes it launches.
	ConfidenceDefinite Confidence = "definite"
	// ConfidenceProbable means the evidence is a supporting signal that is
	// consistent with, but not exclusive to, agent execution.
	ConfidenceProbable Confidence = "probable"
)

// Network describes the network access advertised by the agent environment.
type Network string

const (
	NetworkUnknown  Network = "unknown"
	NetworkEnabled  Network = "enabled"
	NetworkDisabled Network = "disabled"
)
