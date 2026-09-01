package runby

// AgentName identifies a supported agent runtime, orchestrator, or host.
type AgentName string

const (
	AgentUnknown      AgentName = "unknown"
	AgentPaseo        AgentName = "paseo"
	AgentOrca         AgentName = "orca"
	AgentCodex        AgentName = "codex"
	AgentClaudeCode   AgentName = "claude-code"
	AgentAntigravity2 AgentName = "antigravity-2"
	AgentAmp          AgentName = "amp"
	AgentCursor       AgentName = "cursor-agent"
	AgentOpenCode     AgentName = "opencode"
	AgentGeminiCLI    AgentName = "gemini-cli"
	AgentCline        AgentName = "cline"
	AgentOpenClaw     AgentName = "openclaw"
	AgentAuggie       AgentName = "auggie"
	AgentGrokBuild    AgentName = "grok-build"
	AgentPi           AgentName = "pi"
)

// A product on this axis is classified along two independent axes, because one
// enum cannot hold both facts without a cell that fits nothing.
//
// Kind answers what a product drives: a model, or another agent harness.
// ModelSource answers where its intelligence comes from. The pair also decides
// the detection ladder: orchestrators are reported ahead of the harnesses they
// drive; see ladderRank.
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

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (k Kind) String() string { return slug(k, KindUnknown) }

// ModelSource says where the intelligence behind a product comes from. It
// mirrors the model_source field recorded in docs/research/agents.
//
// It describes the product, not the model actually serving a given run. A
// first-party harness can usually be pointed at another vendor's endpoint by
// configuration, and almost no environment variable this package reads would
// say so; the rare agent that does advertise the live model — pi stamps
// PI_PROVIDER and PI_MODEL — surfaces it through Extra, not here. Treat this
// as what the product is built around, never as a claim about which model
// answered.
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

// AgentDriver detects one agent. It is the unit of extension for this axis:
// the built-in agents are declared as drivers, and an agent this package does
// not support is added through Register or WithOnlyDrivers.
//
// A driver carries both the rule for detecting the agent and the facts about
// it that no environment can supply, so registering an agent is one step and
// there is nothing to keep in sync elsewhere.
type AgentDriver struct {
	// Agent identifies the agent this driver reports. Detect fills it into
	// every Agent the driver returns, so Detect need not repeat it.
	Agent AgentName
	// Kind is how much a detection of this agent proves. Detect fills it in
	// the same way.
	Kind Kind
	// Models is where this agent's intelligence comes from. Detect fills it
	// in too, and with Kind it places the agent on the detection ladder by
	// the same rule as a built-in one.
	Models ModelSource
	// Executables names the binaries this agent runs as, so that a live
	// ancestor process can corroborate an environment detection. Leave it
	// empty when no name is specific enough to match safely; a generic name
	// would mislabel unrelated processes.
	Executables []string
	// Detect returns the detection, or false when the environment holds no
	// evidence of this agent. It must not retain env. Agent, Kind, Models,
	// and a missing Confidence are filled in by Detect, so an implementation
	// sets only what its agent actually advertises.
	Detect func(env Env) (Agent, bool)
}

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (a AgentName) String() string { return slug(a, AgentUnknown) }

// AgentNames returns every built-in agent in detection precedence order. Drivers
// added through Register or WithOnlyDrivers are deliberately not included:
// every name returned here has a research document in this repository
// justifying it, and TestSlugsMatchDocs enforces that.
func AgentNames() []AgentName {
	return mapSlice(builtinAgentDrivers, func(d AgentDriver) AgentName { return d.Agent })
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

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (c Confidence) String() string { return slug(c, ConfidenceUnknown) }

// Network describes the network access advertised by the agent environment.
type Network string

const (
	NetworkUnknown  Network = "unknown"
	NetworkEnabled  Network = "enabled"
	NetworkDisabled Network = "disabled"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (n Network) String() string { return slug(n, NetworkUnknown) }
