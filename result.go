package runby

import "strings"

// Agent is one agent layer detected in an environment.
type Agent struct {
	Name AgentName `json:"name"`
	// Kind and Models classify the product rather than this run. Kind is what
	// it drives, and Models is whose intelligence is behind it. See the
	// commentary above Kind.
	Kind   Kind        `json:"kind"`
	Models ModelSource `json:"models"`
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
	Paths   Paths   `json:"paths,omitzero"`

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
	// Agents holds every detected agent, ordered from the most specific
	// orchestrator to the underlying runtime. It is empty when nothing was
	// detected.
	Agents []Agent `json:"agents,omitempty"`
	// TTY is the process-level standard stream status. It is present even
	// when Agents is empty; see TTY.Inspected. It is the only field derived
	// from system calls rather than from the environment.
	TTY TTY `json:"tty"`
	// CI describes the continuous integration run that owns this process. It
	// is a separate axis from Agents: an agent can run inside CI, so both can
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
	// Runners holds every tool that ran this process: a package manager
	// script, a build recipe, a service manager. It is the half of "what
	// launched this" that the other axes leave out, and like Remotes more than
	// one can be present at once. See Runner for what it cannot detect.
	Runners []Runner `json:"runners,omitempty"`
	// Remotes holds every layer detected between the user and this process:
	// multiplexers, SSH, and remote or isolated environments. More than one
	// can be present at once, and the order is a detection order rather than
	// a nesting order. See Remote for what a layer here implies about the
	// other axes.
	Remotes []Remote `json:"remotes,omitempty"`
}

// IsAgent reports whether any supported agent was detected, answering "was
// this launched by an AI agent". Terminal ownership is not agent evidence and
// is reported on the Terminal axis instead.
//
// It, IsCI, HasTerminal, IsRemote, and HasRunner are the five axis predicates
// and are named alike on purpose. Each answers "did this axis detect anything".
// The Agent, Runner, and Remote accessors below answer the different question
// of whether one named product is among what it detected; each returns the
// value with an ok, so a bare existence check is the usual Go idiom:
//
//	if _, ok := result.Runner(runby.RunnerNPM); ok {
func (r Result) IsAgent() bool { return len(r.Agents) > 0 }

// Primary returns the most specific detected layer, and whether there was one.
// On a miss it returns the zero Agent, as Agent, Runner, and Remote do. The ok
// is the thing to branch on, but a caller that logs the layer without checking
// it still writes something meaningful: every enum in this package renders its
// zero value as "unknown" rather than as the empty string.
func (r Result) Primary() (Agent, bool) {
	if len(r.Agents) == 0 {
		return Agent{}, false
	}
	return r.Agents[0], true
}

// Identifier is a value one agent layer advertised, paired with the agent that
// advertised it. Result.SessionID and Result.AgentID return it.
//
// The two travel together because more than one layer can advertise the same
// kind of identifier at once, and those identifiers are not interchangeable. A
// caller that logs the value without the agent is writing two products'
// identifiers into one field that claims to hold one thing.
type Identifier struct {
	// Value is the identifier as the agent advertised it.
	Value string `json:"value"`
	// Agent is the layer that advertised it.
	Agent AgentName `json:"agent"`
}

// SessionID returns the conversation or thread identifier of the outermost
// layer that advertised one, and false when no layer did.
//
// The agent comes back with the value because more than one layer can carry a
// session at once — Orca stamps a pane key and the Codex it hosts stamps a
// thread — and the identifiers are not interchangeable. Returning the agent
// keeps a caller from logging two products' identifiers into one field that
// claims to hold one thing, and keeps the meaning of that field stable as
// drivers are added.
//
// It is not read off Primary alone for the same reason. An orchestrator often
// names the logical agent rather than the conversation — Paseo sets an AgentID
// and no session — while the harness it drives carries the session. Walking
// outermost first gives the most specific identifier that exists without the
// caller having to know which layer of a stack publishes it. Read
// Agent(name).SessionID when the question is about one named product, and
// range over Agents when every identifier is wanted.
func (r Result) SessionID() (Identifier, bool) {
	for _, layer := range r.Agents {
		if layer.SessionID != "" {
			return Identifier{Value: layer.SessionID, Agent: layer.Name}, true
		}
	}
	return Identifier{}, false
}

// AgentID returns the logical agent identifier of the outermost layer that
// advertised one, and false when no layer did. It is the counterpart to
// SessionID and resolves layers the same way; an agent identifier is stable
// across the sessions that agent runs, so the two answer different questions
// and neither substitutes for the other.
func (r Result) AgentID() (Identifier, bool) {
	for _, layer := range r.Agents {
		if layer.AgentID != "" {
			return Identifier{Value: layer.AgentID, Agent: layer.Name}, true
		}
	}
	return Identifier{}, false
}

// Agent returns the detected layer for agent, and whether it was detected.
func (r Result) Agent(agent AgentName) (Agent, bool) {
	for _, layer := range r.Agents {
		if layer.Name == agent {
			return layer, true
		}
	}
	return Agent{}, false
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
func (r Result) IsRemote() bool { return len(r.Remotes) > 0 }

// HasRunner reports whether a tool ran this process rather than a person
// invoking it directly. It is named for the axis, like the other predicates.
func (r Result) HasRunner() bool { return len(r.Runners) > 0 }

// Unattended reports whether nothing is in a position to read this process's
// output or answer a prompt. It is the question behind most uses of this
// package: whether to draw a spinner, colour the output, or stop and ask.
//
// It is the one place this package combines axes. Everything else in Result is
// reported per axis on purpose, because the axes are independent facts, so the
// rule here is pinned rather than left implied. Any of these makes it true:
//
//   - An agent layer of ConfidenceDefinite. An agent can allocate a PTY, so
//     the streams may well look interactive, but no person is behind them.
//     Probable layers do not count on their own: probable is what a driver
//     reports when the product owns the environment but a person could still
//     be the one typing — an Orca pane, a Cline terminal, an agent marker
//     seen through a multiplexer that may have outlived its session — and
//     silencing a prompt a person is waiting on is the worse mistake.
//   - IsCI. A CI job's log is read afterwards, if it is read at all.
//   - A runner of RunnerKindService. A service manager started this, so the
//     output goes to a journal and nobody is waiting on it.
//   - The standard streams were examined and cannot carry a prompt. The
//     examined part matters: a Result built from a bare environment never read
//     them, and an unread TTY is not evidence of anything. See TTY.Inspected.
//
// Terminal is deliberately not consulted. It names the emulator that produced
// the environment, never the one attached now, so it cannot say whether anyone
// is watching; see Terminal.
//
// Treat it as the default for a presentation decision, never as a trust
// boundary. Read the axes directly when the policy differs — a program that
// wants to keep prompting under an agent, say, wants IsCI and TTY rather than
// this, and one that wants to go quiet on any agent evidence at all wants
// IsAgent.
func (r Result) Unattended() bool {
	for _, agent := range r.Agents {
		if agent.Confidence == ConfidenceDefinite {
			return true
		}
	}
	if r.IsCI() {
		return true
	}
	if _, ok := r.RunnerOfKind(RunnerKindService); ok {
		return true
	}
	return r.TTY.Inspected && !r.TTY.Interactive
}

// Runner returns the detected runner for tool, and whether it was detected.
func (r Result) Runner(tool RunnerTool) (Runner, bool) {
	for _, runner := range r.Runners {
		if runner.Tool == tool {
			return runner, true
		}
	}
	return Runner{}, false
}

// RunnerOfKind returns the first detected runner of kind. Use it to ask the
// question a kind exists for, such as whether a service manager started this
// and so nobody is watching the output.
func (r Result) RunnerOfKind(kind RunnerKind) (Runner, bool) {
	for _, runner := range r.Runners {
		if runner.Kind == kind {
			return runner, true
		}
	}
	return Runner{}, false
}

// Remote returns the detected layer for platform, and whether it was detected.
func (r Result) Remote(platform RemotePlatform) (Remote, bool) {
	for _, layer := range r.Remotes {
		if layer.Platform == platform {
			return layer, true
		}
	}
	return Remote{}, false
}

// Multiplexer returns the detected terminal multiplexer. A multiplexer server
// keeps the environment of whichever client started it and cannot refresh an
// already running pane, so its presence means every other axis may be
// reporting evidence left by a session that has since ended.
func (r Result) Multiplexer() (Remote, bool) {
	for _, layer := range r.Remotes {
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
	if len(r.Agents) == 0 {
		return AgentUnknown.String()
	}
	var chain strings.Builder
	chain.WriteString(r.Agents[0].Name.String())
	for _, layer := range r.Agents[1:] {
		chain.WriteByte('>')
		chain.WriteString(layer.Name.String())
	}
	return chain.String()
}
