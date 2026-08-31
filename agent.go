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

// AgentDriver detects one agent. It is the unit of extension for this axis:
// the built-in agents are declared as drivers, and an agent this package does
// not support is added by passing another to Detect with WithAgentDrivers.
//
// A driver carries both the rule for detecting the agent and the facts about
// it that no environment can supply, so registering an agent is one step and
// there is nothing to keep in sync elsewhere.
type AgentDriver struct {
	// Agent identifies the agent this driver reports. Detect fills it into
	// every Detection the driver returns, so Detect need not repeat it.
	Agent Agent
	// Kind is how much a detection of this agent proves. Detect fills it in
	// the same way, and Agent.Kind answers it for the built-in agents.
	Kind Kind
	// Executables names the binaries this agent runs as, so that a live
	// ancestor process can corroborate an environment detection. Leave it
	// empty when no name is specific enough to match safely; a generic name
	// would mislabel unrelated processes.
	Executables []string
	// Detect returns the detection, or false when the environment holds no
	// evidence of this agent. It must not retain env. Agent, Kind, and a
	// missing Confidence are filled in by Detect, so an implementation sets
	// only what its agent actually advertises.
	Detect func(env Env) (Detection, bool)
}

// agentKinds is derived from the built-in driver table, so a driver is the one
// place an agent is registered.
var agentKinds = indexBy(builtinAgentDrivers, func(d AgentDriver) (Agent, Kind) {
	return d.Agent, d.Kind
})

// Kind reports how much a detection of a proves. It returns KindUnknown for
// agents this package does not support; a driver supplied through
// WithAgentDrivers carries its own Kind onto the Detection instead.
func (a Agent) Kind() Kind { return lookupOr(agentKinds, a, KindUnknown) }

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (a Agent) String() string { return slug(a, AgentUnknown) }

// Agents returns every supported agent in detection precedence order.
func Agents() []Agent {
	return mapSlice(builtinAgentDrivers, func(d AgentDriver) Agent { return d.Agent })
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
