package runby

// Agent identifies a supported agent runtime, orchestrator, or host.
type Agent string

const (
	AgentUnknown      Agent = "unknown"
	AgentPaseo        Agent = "paseo"
	AgentOrca         Agent = "orca"
	AgentCodex        Agent = "codex"
	AgentClaudeCode   Agent = "claude-code"
	AgentAntigravity2 Agent = "antigravity-2"
	AgentAmp          Agent = "amp"
	AgentCursor       Agent = "cursor-agent"
	AgentOpenCode     Agent = "opencode"
)

// Kind separates an orchestrator that manages other agents from the agent
// runtime it drives. It mirrors the product_type field recorded in docs/research/agents.
//
// There is no host kind. A product that only proves which application owns the
// terminal, such as Zed, is reported on the Terminal axis instead, because
// owning the terminal is not evidence that an agent rather than a person ran
// the command.
type Kind string

const (
	KindUnknown Kind = "unknown"
	// KindOrchestrator manages other agents and advertises its own agent
	// identity, so it is reported ahead of the runtime it drives.
	KindOrchestrator Kind = "orchestrator"
	// KindHarness is an agent runtime that executes model-requested commands.
	KindHarness Kind = "harness"
)

// agentInfo is everything this package knows about an agent that is not the
// environment rule for detecting it.
type agentInfo struct {
	kind Kind
	// executables names the binaries this agent runs as, so a live ancestor
	// process can corroborate an environment detection. Leave it empty when
	// no name is specific enough to match safely; a generic name would
	// mislabel unrelated processes.
	executables []string
}

// agents is the single source of truth for what an Agent is. A product is
// registered here and detected in agent_detectors.go, and nowhere else.
var agents = map[Agent]agentInfo{
	AgentPaseo:      {kind: KindOrchestrator, executables: []string{"paseo"}},
	AgentOrca:       {kind: KindOrchestrator},
	AgentCodex:      {kind: KindHarness, executables: []string{"codex"}},
	AgentClaudeCode: {kind: KindHarness, executables: []string{"claude"}},
	AgentAmp:        {kind: KindHarness, executables: []string{"amp"}},
	AgentCursor:     {kind: KindHarness, executables: []string{"cursor-agent"}},
	AgentOpenCode:   {kind: KindHarness, executables: []string{"opencode"}},
	// Orca's binary shares its name with the GNOME screen reader, and no
	// Antigravity executable name has been verified against an official
	// source, so neither can be corroborated by an ancestor.
	AgentAntigravity2: {kind: KindHarness},
}

// Kind reports how much a detection of a proves. It returns KindUnknown for
// agents this package does not support, which is not the map's zero value.
func (a Agent) Kind() Kind {
	if info, ok := agents[a]; ok {
		return info.kind
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
