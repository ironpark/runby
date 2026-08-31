package runby

// Detection is one agent layer detected in an environment.
type Detection struct {
	Agent Agent `json:"agent"`
	// Kind, Models, and Level classify the product rather than this run. Kind
	// is what it drives, Models is whose intelligence is behind it, and Level
	// is the ladder position read off the two. See the commentary above Kind.
	Kind   Kind        `json:"kind"`
	Models ModelSource `json:"models"`
	Level  Level       `json:"level"`
	Axis

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

	// AncestorPID is the PID of a running ancestor process whose executable
	// belongs to this agent, or 0 when none was found.
	//
	// A non-zero value is the strongest confirmation this package can offer:
	// the environment said this agent, and the agent is still running as an
	// ancestor. Zero is not a denial. The chain is unreadable on some
	// platforms and stops at the first process owned by another user, and an
	// agent can legitimately launch a process it does not remain an ancestor
	// of. Use it to strengthen a positive, never to reject one.
	AncestorPID int `json:"ancestor_pid,omitempty"`
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

// Result is everything detected about one process environment.
type Result struct {
	// Layers holds every detected agent, ordered from the most specific
	// orchestrator to the underlying runtime. It is empty when nothing was
	// detected.
	Layers []Detection `json:"layers"`
	// TTY is the process-level standard stream status. It is present even
	// when Layers is empty; see TTY.Inspected. It is the only field derived
	// from system calls rather than from the environment.
	TTY TTY `json:"tty"`
	// CI describes the continuous integration run that owns this process. It
	// is a separate axis from Layers: an agent can run inside CI, so both can
	// be populated at once.
	CI CI `json:"ci"`
	// Terminal identifies the terminal emulator that produced this
	// environment. It is a separate axis too, and a deliberately weak one;
	// see Terminal for why it cannot describe the currently attached
	// terminal.
	Terminal Terminal `json:"terminal"`
	// Process is the ancestor chain of this process. It is the only axis not
	// derived from the environment, and the only one that can distinguish a
	// live ancestor from a marker left behind by one that has exited.
	Process ProcessTree `json:"process"`
	// Remote holds every layer detected between the user and this process:
	// multiplexers, SSH, and remote or isolated environments. More than one
	// can be present at once, and the order is a detection order rather than
	// a nesting order. See Remote for what a layer here implies about the
	// other axes.
	Remote []Remote `json:"remote"`
}

// IsAgent reports whether any supported agent was detected, answering "was
// this launched by an AI agent". Terminal ownership is not agent evidence and
// is reported on the Terminal axis instead.
//
// It, IsCI, HasTerminal, and IsRemote are the four axis predicates and are
// named alike on purpose. The Layer and RemoteLayer accessors below answer the
// different question of whether one named product is among the layers.
func (r Result) IsAgent() bool { return len(r.Layers) > 0 }

// Primary returns the most specific detected layer.
func (r Result) Primary() (Detection, bool) {
	if len(r.Layers) == 0 {
		return Detection{Agent: AgentUnknown, Kind: KindUnknown, Axis: Axis{Confidence: ConfidenceUnknown}}, false
	}
	return r.Layers[0], true
}

// Agent returns the most specific detected agent, or AgentUnknown.
func (r Result) Agent() Agent {
	primary, _ := r.Primary()
	return primary.Agent
}

// Layer returns the detected layer for agent.
func (r Result) Layer(agent Agent) (Detection, bool) {
	for _, layer := range r.Layers {
		if layer.Agent == agent {
			return layer, true
		}
	}
	return Detection{}, false
}

// HasLayer reports whether agent is one of the detected layers.
func (r Result) HasLayer(agent Agent) bool {
	_, ok := r.Layer(agent)
	return ok
}

// IsCI reports whether this process is running in a CI job.
func (r Result) IsCI() bool { return r.CI.Detected }

// HasTerminal reports whether a terminal emulator was identified.
//
// It is not IsTerminal on purpose. In Go that name answers "is this file
// descriptor a terminal", which this package answers with TTY.Interactive and
// its neighbours. This one answers a different question: whether the
// environment named the emulator that produced it.
func (r Result) HasTerminal() bool { return r.Terminal.Detected }

// IsRemote reports whether any layer sits between the user and this process.
func (r Result) IsRemote() bool { return len(r.Remote) > 0 }

// RemoteLayer returns the detected layer for platform.
func (r Result) RemoteLayer(platform RemotePlatform) (Remote, bool) {
	for _, layer := range r.Remote {
		if layer.Platform == platform {
			return layer, true
		}
	}
	return Remote{}, false
}

// HasRemoteLayer reports whether platform is one of the detected layers.
func (r Result) HasRemoteLayer(platform RemotePlatform) bool {
	_, ok := r.RemoteLayer(platform)
	return ok
}

// Multiplexer returns the detected terminal multiplexer. A multiplexer server
// keeps the environment of whichever client started it and cannot refresh an
// already running pane, so its presence means every other axis may be
// reporting evidence left by a session that has since ended.
func (r Result) Multiplexer() (Remote, bool) {
	for _, layer := range r.Remote {
		if layer.IsMultiplexer() {
			return layer, true
		}
	}
	return Remote{}, false
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
