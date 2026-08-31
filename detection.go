package runby

// Detection is one agent layer detected in an environment.
type Detection struct {
	Agent      Agent      `json:"agent"`
	Kind       Kind       `json:"kind"`
	Confidence Confidence `json:"confidence"`

	// SessionID is the agent's identifier for the conversation or thread that
	// launched the process.
	SessionID string `json:"session_id,omitempty"`
	// AgentID is an orchestrator's identifier for the logical agent, which is
	// stable across the sessions that agent runs.
	AgentID string `json:"agent_id,omitempty"`
	// Entrypoint is how the agent was invoked, using the agent's own
	// vocabulary, such as "cli", "acp", or "sidecar".
	Entrypoint string `json:"entrypoint,omitempty"`
	// Nested reports that the agent advertised this process as belonging to a
	// child session rather than a top-level one.
	Nested bool `json:"nested,omitempty"`

	Sandbox Sandbox `json:"sandbox"`
	Paths   Paths   `json:"paths"`

	// Extra holds values that only one agent advertises, keyed by
	// "<agent-slug>.<name>", so that agent-specific metadata does not widen
	// the shared fields above. Keys are stable; treat missing keys as unset.
	Extra map[string]string `json:"extra,omitempty"`

	// Evidence lists the environment variable names that produced this
	// detection, sorted. Their values may be sensitive and are never copied.
	Evidence []string `json:"evidence"`
}

// Sandbox describes the isolation the agent advertised for this process.
type Sandbox struct {
	// Mode is the agent's own sandbox name, such as "workspace-write". It is
	// deliberately not normalized across agents.
	Mode    string  `json:"mode,omitempty"`
	Network Network `json:"network"`
}

// Paths holds filesystem locations the agent advertised. They are not
// validated; the directories may not exist.
type Paths struct {
	WorkingDirectory string `json:"working_directory,omitempty"`
	DataDirectory    string `json:"data_directory,omitempty"`
}

// IsAgent reports whether this layer is evidence that an AI agent, rather than
// a person, launched the process. It is false for KindHost layers such as a
// terminal owned by an editor.
func (d Detection) IsAgent() bool {
	return d.Kind == KindOrchestrator || d.Kind == KindHarness
}

// Result is everything detected about one process environment.
type Result struct {
	// Layers holds every detected agent, ordered from the most specific
	// orchestrator to the underlying runtime. It is empty when nothing was
	// detected.
	Layers []Detection `json:"layers"`
	// Terminal is the process-level terminal status. It is present even when
	// Layers is empty; see Terminal.Inspected.
	Terminal Terminal `json:"terminal"`
}

// Found reports whether any supported agent was detected, including KindHost
// layers. Use IsAgent to exclude those.
func (r Result) Found() bool { return len(r.Layers) > 0 }

// IsAgent reports whether any layer is evidence of AI agent execution.
func (r Result) IsAgent() bool {
	for _, layer := range r.Layers {
		if layer.IsAgent() {
			return true
		}
	}
	return false
}

// Primary returns the most specific detected layer.
func (r Result) Primary() (Detection, bool) {
	if len(r.Layers) == 0 {
		return Detection{Agent: AgentUnknown, Kind: KindUnknown, Confidence: ConfidenceUnknown}, false
	}
	return r.Layers[0], true
}

// Agent returns the most specific detected agent, or AgentUnknown.
func (r Result) Agent() Agent {
	primary, _ := r.Primary()
	return primary.Agent
}

// Get returns the detected layer for agent.
func (r Result) Get(agent Agent) (Detection, bool) {
	for _, layer := range r.Layers {
		if layer.Agent == agent {
			return layer, true
		}
	}
	return Detection{}, false
}

// Has reports whether agent is one of the detected layers.
func (r Result) Has(agent Agent) bool {
	_, ok := r.Get(agent)
	return ok
}

// Chain renders the detected layers as "paseo>codex", most specific first, for
// use as a single log or telemetry field. It returns "unknown" when nothing was
// detected, so the field is never empty.
func (r Result) Chain() string {
	if len(r.Layers) == 0 {
		return AgentUnknown.String()
	}
	chain := r.Layers[0].Agent.String()
	for _, layer := range r.Layers[1:] {
		chain += ">" + layer.Agent.String()
	}
	return chain
}
