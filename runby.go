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

// WithAgentDrivers adds agent drivers ahead of the built-in ones, so a custom
// orchestrator is reported as the primary layer over the runtime it drives.
// Drivers are tried in the order given, and every match is reported.
func WithAgentDrivers(drivers ...AgentDriver) Option {
	return func(o *options) {
		o.agentDrivers = append(append([]AgentDriver{}, drivers...), o.agentDrivers...)
	}
}

// WithOnlyAgentDrivers replaces the built-in agent drivers entirely. Passing
// no drivers disables agent detection.
func WithOnlyAgentDrivers(drivers ...AgentDriver) Option {
	return func(o *options) {
		o.agentDrivers = append([]AgentDriver{}, drivers...)
	}
}

// WithCIDrivers adds CI drivers ahead of the built-in ones, so a platform this
// package does not support is reported over the generic CI convention. Drivers
// are tried in the order given, and the first match wins.
func WithCIDrivers(drivers ...CIDriver) Option {
	return func(o *options) {
		o.ciDrivers = append(append([]CIDriver{}, drivers...), o.ciDrivers...)
	}
}

// WithOnlyCIDrivers replaces the built-in CI drivers entirely. Passing no
// drivers disables CI detection.
func WithOnlyCIDrivers(drivers ...CIDriver) Option {
	return func(o *options) {
		o.ciDrivers = append([]CIDriver{}, drivers...)
	}
}

// WithTerminalDrivers adds terminal drivers ahead of the built-in ones.
// Drivers are tried in the order given, and the first match wins.
func WithTerminalDrivers(drivers ...TerminalDriver) Option {
	return func(o *options) {
		o.terminalDrivers = append(append([]TerminalDriver{}, drivers...), o.terminalDrivers...)
	}
}

// WithOnlyTerminalDrivers replaces the built-in terminal drivers entirely.
// Passing no drivers disables terminal detection.
func WithOnlyTerminalDrivers(drivers ...TerminalDriver) Option {
	return func(o *options) {
		o.terminalDrivers = append([]TerminalDriver{}, drivers...)
	}
}

// WithRemoteDrivers adds remote-layer drivers ahead of the built-in ones.
// Unlike the agent axis, every matching driver is reported, so ordering
// affects only the order of Result.Remote.
func WithRemoteDrivers(drivers ...RemoteDriver) Option {
	return func(o *options) {
		o.remoteDrivers = append(append([]RemoteDriver{}, drivers...), o.remoteDrivers...)
	}
}

// WithOnlyRemoteDrivers replaces the built-in remote drivers entirely. Passing
// no drivers disables remote detection.
func WithOnlyRemoteDrivers(drivers ...RemoteDriver) Option {
	return func(o *options) {
		o.remoteDrivers = append([]RemoteDriver{}, drivers...)
	}
}

// Detect inspects an environment and returns every supported agent found in
// it, plus the CI platform and terminal status. With no options it inspects
// the current process, including its terminal.
func Detect(opts ...Option) Result {
	config := options{
		env:             processEnv{},
		agentDrivers:    builtinAgentDrivers,
		ciDrivers:       builtinCIDrivers,
		terminalDrivers: builtinTerminalDrivers,
		remoteDrivers:   builtinRemoteDrivers,
		inspectTTY:      true,
		inspectProcess:  true,
	}
	for _, opt := range opts {
		opt(&config)
	}

	result := Result{TTY: config.tty}
	if config.inspectTTY {
		result.TTY = InspectTTY()
	}

	// The ancestor chain is labelled from the drivers this call was configured
	// with, so a driver added through WithAgentDrivers is corroborated exactly
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
		if detection.Kind == "" {
			detection.Kind = KindUnknown
		}
		if detection.Confidence == "" {
			detection.Confidence = ConfidenceDefinite
		}
		if detection.Sandbox.Network == "" {
			detection.Sandbox.Network = NetworkUnknown
		}
		// An ancestor running this agent's executable proves it is still
		// alive, which no environment variable can.
		if ancestor, ok := tree.FindAgent(detection.Agent); ok {
			detection.AncestorPID = ancestor.PID
		}
		layers = append(layers, detection)
	}
	return layers
}

// detectCI reports the CI platform. Only the first, most specific match is
// reported; every platform also sets the generic CI variable, so later matches
// are redundant.
func detectCI(config options) CI {
	for _, driver := range config.ciDrivers {
		ci, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		ci.Provider = driver.Provider
		if ci.Confidence == "" {
			ci.Confidence = ConfidenceDefinite
		}
		ci.Detected = true
		return ci
	}
	return CI{Provider: CIProviderUnknown, Confidence: ConfidenceUnknown}
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
		if remote.Confidence == "" {
			remote.Confidence = ConfidenceDefinite
		}
		if ancestor, ok := tree.Find(func(p Process) bool { return p.Remote == remote.Platform }); ok {
			remote.AncestorPID = ancestor.PID
		}
		layers = append(layers, remote)
	}
	return layers
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
		if terminal.Confidence == "" {
			terminal.Confidence = ConfidenceDefinite
		}
		terminal.Detected = true
		break
	}
	if !terminal.Detected {
		terminal.Program = TerminalUnknown
		terminal.Confidence = ConfidenceUnknown
		terminal.Term, _ = Value(config.env, "TERM")
		return terminal
	}

	if ancestor, ok := result.Process.Find(func(p Process) bool { return p.Terminal == terminal.Program }); ok {
		terminal.AncestorPID = ancestor.PID
	}

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
	currentOnce.Do(func() { currentResult = Detect() })
	return currentResult
}

// IsAgent reports whether this process was launched by an AI agent, using the
// cached Current result. Terminal ownership is not agent evidence and is
// reported on the Terminal axis, so it never affects this.
func IsAgent() bool { return Current().Found() }

// IsTerminal reports whether a terminal emulator was identified, using the
// cached Current result. See Terminal for why this is weaker evidence than the
// other axes.
func IsTerminal() bool { return Current().Terminal.Detected }

// Environ returns the current process environment as an Env. It is a
// convenience for callers that build their own driver pipelines.
func Environ() Env { return EnvironEnv(os.Environ()) }
