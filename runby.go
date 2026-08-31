// Package runby detects the coding agent or orchestrator that started the
// current process from environment variables inherited by that process.
//
// The common case is two calls:
//
//	if runby.IsAgent() {
//		log.Printf("run by %s", runby.Current().Chain())
//	}
//
// Environment variables are only a snapshot taken when the process started. A
// detection means that the agent was active when it launched the process; it
// does not prove that the parent agent is still alive when the result is
// inspected later. Confirming that requires a separate liveness channel such as
// a PID, IPC, or the agent's own API.
//
// API keys and general configuration variables are not evidence that an agent
// launched the process and are never used for detection.
//
// Every axis is extensible by the same means. A product this package does not
// support is added by passing a driver — AgentDriver, CIDriver,
// TerminalDriver, or RemoteDriver — to Detect. A driver carries the rule for
// detecting the product together with the facts no environment can supply,
// such as what a detection proves and which binaries the product runs as, so
// the built-in products and yours are declared exactly the same way.
package runby

import (
	"os"
	"sync"
)

type options struct {
	env             Env
	agentDrivers    []AgentDriver
	ciDrivers       []CIDriver
	terminalDrivers []TerminalDriver
	remoteDrivers   []RemoteDriver
	runnerDrivers   []RunnerDriver
	tty             TTY
	process         ProcessTree
	// inspectTTY and inspectProcess are only true when the environment being
	// inspected is this process's own, so that the standard streams and the
	// ancestor chain describe the same process as the detected layers.
	inspectTTY     bool
	inspectProcess bool
}

// Option configures Detect.
type Option func(*options)

// WithEnviron inspects an explicit environment given as "NAME=value" entries,
// instead of the process environment. Because that environment does not
// necessarily belong to this process, the standard streams are not inspected
// unless WithTTY is also given.
func WithEnviron(environ []string) Option {
	return WithEnv(EnvironEnv(environ))
}

// WithEnv inspects an explicit Env instead of the process environment. As with
// WithEnviron, the standard streams are not inspected.
func WithEnv(env Env) Option {
	return func(o *options) {
		o.env = env
		o.inspectTTY = false
		o.inspectProcess = false
	}
}

// WithLookup inspects an environment through a lookup function such as
// os.LookupEnv. As with WithEnviron, the standard streams are not inspected.
func WithLookup(lookup func(name string) (string, bool)) Option {
	return WithEnv(lookupEnv(lookup))
}

// WithoutTTY skips standard stream inspection, avoiding its system calls when
// only the environment-derived axes are needed. It does not affect Terminal,
// which is derived from the environment rather than from file descriptors.
func WithoutTTY() Option {
	return func(o *options) {
		o.inspectTTY = false
		o.tty = TTY{}
	}
}

// WithTTY sets the standard stream status explicitly instead of inspecting
// them. It is intended for wrappers that already know the status of the
// environment they are describing, and for tests.
func WithTTY(tty TTY) Option {
	return func(o *options) {
		o.inspectTTY = false
		o.tty = tty
	}
}

// WithoutProcessTree skips reading the ancestor process chain. The walk costs
// a few file reads or system calls per ancestor, which is more than the other
// axes cost; skip it when only the environment matters.
func WithoutProcessTree() Option {
	return func(o *options) {
		o.inspectProcess = false
		o.process = ProcessTree{}
	}
}

// WithProcessTree sets the ancestor chain explicitly instead of reading it.
// It is intended for wrappers describing another process, and for tests.
func WithProcessTree(tree ProcessTree) Option {
	return func(o *options) {
		o.inspectProcess = false
		o.process = tree
	}
}

// WithOnlyDrivers runs exactly the drivers given and nothing else: no built-in
// driver and nothing added through Register participates in this call. With no
// drivers at all, no axis derived from the environment is detected.
//
// It is the only per-call driver option, and the two things it is for do not
// divide by axis:
//
//   - Testing a driver in isolation, so a fixture cannot be answered by a
//     built-in that happens to match it too.
//   - Pinning a test in a program where something has registered a driver. The
//     registry is process-wide, so without this a blank import anywhere in the
//     build could change what a test observes.
//
// Pair it with BuiltinDrivers to run the built-in set plus a custom driver, or
// minus one of its own:
//
//	Detect(WithOnlyDrivers(append(BuiltinDrivers(), acme)...))
//
// Drivers are sorted onto their own axis, so one call covers as many axes as
// the drivers span. The agent axis is still ordered by the ladder rather than
// by the order given here.
func WithOnlyDrivers(drivers ...Driver) Option {
	var only registry
	for _, driver := range drivers {
		if driver == nil {
			panic("runby: WithOnlyDrivers called with a nil driver")
		}
		driver.addTo(&only)
	}
	only.check()
	return func(o *options) {
		o.agentDrivers = only.agents
		o.ciDrivers = only.ci
		o.terminalDrivers = only.terminals
		o.remoteDrivers = only.remotes
		o.runnerDrivers = only.runners
	}
}

// Detect inspects an environment and returns every supported agent found in
// it, plus the CI platform and terminal status. With no options it inspects
// the current process, including its terminal.
func Detect(opts ...Option) Result {
	config := defaultOptions()
	for _, opt := range opts {
		opt(&config)
	}
	// The agent axis is ordered by the ladder rather than by who was added
	// first, so neither package initialization order nor the order options
	// were passed in can report a runtime as the layer above its orchestrator.
	config.agentDrivers = sortByLadder(config.agentDrivers)

	result := Result{TTY: config.tty}
	if config.inspectTTY {
		result.TTY = InspectTTY()
	}

	// The ancestor chain is labelled from the drivers this call was configured
	// with, so a driver added through Register is corroborated exactly
	// like a built-in one.
	labels := config.executableLabels()
	if config.inspectProcess {
		result.Process = inspectProcessTree(labels)
	} else {
		result.Process = labelProcessTree(config.process, labels)
	}

	result.Layers = detectAgents(config, result.Process)
	result.CI = detectCI(config)
	result.Remote = detectRemote(config, result.Process)
	result.Runner = detectRunners(config, result.Process)
	result.Terminal = detectTerminal(config, result)
	return result
}

// executableLabels gathers the name-to-product mapping from every configured
// driver that names its executables.
func (config options) executableLabels() executableLabels {
	labels := make(executableLabels)
	for _, driver := range config.agentDrivers {
		labels.add(driver.Executables, Process{Agent: driver.Agent})
	}
	for _, driver := range config.terminalDrivers {
		labels.add(driver.Executables, Process{Terminal: driver.Program})
	}
	for _, driver := range config.remoteDrivers {
		labels.add(driver.Executables, Process{Remote: driver.Platform})
	}
	for _, driver := range config.runnerDrivers {
		labels.add(driver.Executables, Process{Runner: driver.Tool})
	}
	return labels
}

// detectAgents reports every agent layer, most specific orchestrator first.
func detectAgents(config options, tree ProcessTree) []Detection {
	var layers []Detection
	for _, driver := range config.agentDrivers {
		detection, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		// Drivers fill in only what their agent advertises; the identity and
		// the defaults shared by every detection are applied once, here.
		detection.Agent = driver.Agent
		detection.Kind = driver.Kind
		detection.Models = driver.Models
		if detection.Kind == "" {
			detection.Kind = KindUnknown
		}
		if detection.Models == "" {
			detection.Models = ModelsUnknown
		}
		detection.Level = level(detection.Kind, detection.Models)
		if detection.Sandbox.Network == "" {
			detection.Sandbox.Network = NetworkUnknown
		}
		detection.applyDefaults()
		// An ancestor running this agent's executable proves it is still
		// alive, which no environment variable can.
		detection.AncestorPID = tree.pidOf(func(p Process) bool { return p.Agent == detection.Agent })
		layers = append(layers, detection)
	}
	return layers
}

// detectCI reports the CI platform. Only the first, most specific match is
// reported; every platform also sets the generic CI variable, so later matches
// are redundant. Nothing in the ancestor chain can corroborate a CI run, so
// unlike the other axes this one has no live process to consult.
func detectCI(config options) CI {
	for _, driver := range config.ciDrivers {
		ci, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		ci.Provider = driver.Provider
		ci.Detected = true
		ci.applyDefaults()
		return ci
	}
	return CI{Provider: CIProviderUnknown, Axis: Axis{Confidence: ConfidenceUnknown}}
}

// detectRemote reports every layer between the user and this process. Unlike
// the other axes it does not stop at the first match: an SSH session into a
// Codespace running tmux is three concurrent layers, not a precedence contest.
func detectRemote(config options, tree ProcessTree) []Remote {
	var layers []Remote
	for _, driver := range config.remoteDrivers {
		remote, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		remote.Platform = driver.Platform
		remote.Kind = driver.Kind
		if remote.Kind == "" {
			remote.Kind = RemoteKindUnknown
		}
		remote.applyDefaults()
		remote.AncestorPID = tree.pidOf(func(p Process) bool { return p.Remote == remote.Platform })
		layers = append(layers, remote)
	}
	return layers
}

// detectRunners reports every tool that ran this process. Like the remote axis
// it does not stop at the first match: a pre-commit hook running an npm script
// that shells out to make is three concurrent layers.
func detectRunners(config options, tree ProcessTree) []Runner {
	var runners []Runner
	for _, driver := range config.runnerDrivers {
		runner, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		runner.Tool = driver.Tool
		runner.Kind = driver.Kind
		if runner.Kind == "" {
			runner.Kind = RunnerKindUnknown
		}
		runner.applyDefaults()
		runner.AncestorPID = tree.pidOf(func(p Process) bool { return p.Runner == runner.Tool })
		runners = append(runners, runner)
	}
	return runners
}

// detectTerminal reports the terminal emulator. It runs last because its
// confidence depends on the remote layers and the ancestor chain that the
// earlier steps produced.
func detectTerminal(config options, result Result) Terminal {
	var terminal Terminal
	for _, driver := range config.terminalDrivers {
		found, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		terminal = found
		terminal.Program = driver.Program
		terminal.Detected = true
		terminal.applyDefaults()
		break
	}
	if !terminal.Detected {
		terminal.Program = TerminalUnknown
		terminal.Confidence = ConfidenceUnknown
		terminal.Term, _ = Value(config.env, "TERM")
		return terminal
	}

	terminal.AncestorPID = result.Process.pidOf(func(p Process) bool { return p.Terminal == terminal.Program })

	// A multiplexer server keeps the environment of whichever client started
	// it and cannot refresh a running pane, so the identity here may name a
	// terminal that is not the one displaying it.
	//
	// A live ancestor settles the question the other way. A multiplexer server
	// daemonizes and is reparented away from the terminal that started it, so
	// the terminal cannot appear in a pane's ancestor chain; finding it there
	// means this process is not behind a stale pane.
	if _, muxed := result.Multiplexer(); muxed && terminal.AncestorPID == 0 &&
		terminal.Confidence == ConfidenceDefinite {
		terminal.Confidence = ConfidenceProbable
	}
	return terminal
}

var (
	currentOnce   sync.Once
	currentResult Result
)

// Current returns Detect for this process, computed once and cached. The
// process environment and standard streams are fixed at startup in practice,
// so repeated calls are free. Use Detect directly to observe changes made by
// os.Setenv after the first call.
func Current() Result {
	currentOnce.Do(func() {
		// Recorded before detecting so that a Register racing this call is
		// reported as the mistake it is rather than being silently dropped.
		detected.Store(true)
		currentResult = Detect()
	})
	return currentResult
}

// The four axis predicates below are the shorthand for the common case, each
// answering one axis of the cached Current result. They live together, and are
// named like their Result counterparts, so that the set is visible at a glance.

// IsAgent reports whether this process was launched by an AI agent. Terminal
// ownership is not agent evidence and is reported on the Terminal axis, so it
// never affects this.
func IsAgent() bool { return Current().IsAgent() }

// IsCI reports whether this process is running in a CI job.
func IsCI() bool { return Current().IsCI() }

// HasTerminal reports whether a terminal emulator was identified. See Terminal
// for why this is weaker evidence than the other axes, and Result.HasTerminal
// for why it is not called IsTerminal.
func HasTerminal() bool { return Current().HasTerminal() }

// IsRemote reports whether any layer was detected between the user and this
// process.
func IsRemote() bool { return Current().IsRemote() }

// HasRunner reports whether a tool ran this process rather than a person
// invoking it directly.
func HasRunner() bool { return Current().HasRunner() }

// Environ returns the current process environment as an Env. It is a
// convenience for callers that build their own driver pipelines.
func Environ() Env { return EnvironEnv(os.Environ()) }
