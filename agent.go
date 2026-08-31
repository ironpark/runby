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

// A product on this axis is classified along two independent axes, because one
// enum cannot hold both facts without a cell that fits nothing.
//
// Kind answers what a product drives: a model, or another agent harness.
// ModelSource answers where its intelligence comes from. Level is the familiar
// ladder read off the pair, for a single field to log or group by.
//
// The pair matters because the two are genuinely independent. Google
// Antigravity 2.0 orchestrates a harness, like Paseo and Orca do, but the
// harness and the model behind it are its own vendor's — an orchestrator whose
// models are first-party. A single ladder has no rung for that, and a vendor
// shipping both a harness and an orchestrator over it is a pattern, not a
// one-off.

// Kind separates an orchestrator that manages other agents from the agent
// runtime it drives. It mirrors the product_type field recorded in
// docs/research/agents, and TestKindsMatchDocs holds the two together.
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

// ModelSource says where the intelligence behind a product comes from. It
// mirrors the model_source field recorded in docs/research/agents.
//
// It describes the product, not the model actually serving a given run. A
// first-party harness can usually be pointed at another vendor's endpoint by
// configuration, and no environment variable this package reads would say so.
// Treat it as what the product is built around, never as a claim about which
// model answered.
type ModelSource string

const (
	ModelsUnknown ModelSource = "unknown"
	// ModelsFirstParty is a product built around its own vendor's model, with
	// a harness written for it: Codex on OpenAI, Claude Code on Anthropic,
	// Antigravity on Google.
	ModelsFirstParty ModelSource = "first-party"
	// ModelsMultiVendor is a product that brings a harness but not a model,
	// reaching other vendors' models by registration or API key.
	ModelsMultiVendor ModelSource = "multi-vendor"
	// ModelsDelegated is a product that runs no model of its own and drives
	// other agents instead, inheriting whatever they are configured with.
	ModelsDelegated ModelSource = "delegated"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (m ModelSource) String() string { return slug(m, ModelsUnknown) }

// Level is the layer a product occupies in the agent stack, read off Kind and
// ModelSource. It is the shorthand for the pair, for logging and grouping.
type Level string

const (
	LevelUnknown Level = "unknown"
	// Level1 is a harness built around its own vendor's model.
	Level1 Level = "l1"
	// Level2 is a harness that brings no model and reaches other vendors'.
	Level2 Level = "l2"
	// Level3 drives a harness rather than a model. Its models may still be
	// first-party, as Antigravity 2.0's are, so the ladder alone does not
	// answer whose model is behind it; ModelSource does.
	Level3 Level = "l3"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (l Level) String() string { return slug(l, LevelUnknown) }

// level derives the ladder position from the pair. It is one function so that
// a driver supplied through Register or WithOnlyDrivers is placed by the same rule as a
// built-in one, and so the rule exists in exactly one place.
func level(kind Kind, models ModelSource) Level {
	switch {
	case kind == KindOrchestrator:
		// Driving a harness is what puts a product on the top rung, whoever
		// owns the model it ends up running.
		return Level3
	case kind == KindHarness && models == ModelsFirstParty:
		return Level1
	case kind == KindHarness && models == ModelsMultiVendor:
		return Level2
	default:
		return LevelUnknown
	}
}

// AgentDriver detects one agent. It is the unit of extension for this axis:
// the built-in agents are declared as drivers, and an agent this package does
// not support is added through Register or WithOnlyDrivers.
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
	// Models is where this agent's intelligence comes from. Detect fills it
	// in too, and derives Detection.Level from it and Kind together, so a
	// custom agent is placed on the ladder by the same rule as a built-in one.
	Models ModelSource
	// Executables names the binaries this agent runs as, so that a live
	// ancestor process can corroborate an environment detection. Leave it
	// empty when no name is specific enough to match safely; a generic name
	// would mislabel unrelated processes.
	Executables []string
	// Detect returns the detection, or false when the environment holds no
	// evidence of this agent. It must not retain env. Agent, Kind, Models,
	// Level, and a missing Confidence are filled in by Detect, so an
	// implementation sets only what its agent actually advertises.
	Detect func(env Env) (Detection, bool)
}

// agentKinds is derived from the built-in driver table, so a driver is the one
// place an agent is registered.
var agentKinds = indexBy(builtinAgentDrivers, func(d AgentDriver) (Agent, Kind) {
	return d.Agent, d.Kind
})

// agentModels is derived the same way, from the same table.
var agentModels = indexBy(builtinAgentDrivers, func(d AgentDriver) (Agent, ModelSource) {
	return d.Agent, d.Models
})

// Kind reports how much a detection of a proves. It answers for the built-in
// agents and for those added through Register, and returns KindUnknown for
// anything else — including a driver passed to a single Detect call through
// WithOnlyDrivers, which is not visible outside that call. Detection.Kind
// always carries the driver's own answer.
func (a Agent) Kind() Kind {
	if kind, ok := agentKinds[a]; ok {
		return kind
	}
	return registeredAgentKind(a)
}

// Models reports where a's intelligence comes from. As with Kind it answers for
// the built-in agents and for registered ones, and returns ModelsUnknown
// otherwise.
func (a Agent) Models() ModelSource {
	if models, ok := agentModels[a]; ok {
		return models
	}
	return registeredAgentModels(a)
}

// Level reports the layer a occupies in the agent stack.
func (a Agent) Level() Level { return level(a.Kind(), a.Models()) }

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (a Agent) String() string { return slug(a, AgentUnknown) }

// Agents returns every built-in agent in detection precedence order. Drivers
// added through Register or WithOnlyDrivers are deliberately not included:
// every name returned here has a research document in this repository
// justifying it, and TestSlugsMatchDocs enforces that.
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
