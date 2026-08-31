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
package runby

import (
	"os"
	"sync"
)

// Detector reports whether an environment shows that its agent launched the
// process. Implement it to detect an agent this package does not support, then
// pass it to Detect with WithDetectors.
type Detector interface {
	// Agent returns the agent this detector reports.
	Agent() Agent
	// Detect returns the detection, or false if the environment holds no
	// evidence of this agent. Implementations must not retain env.
	Detect(env Env) (Detection, bool)
}

// NewDetector adapts a function into a Detector.
func NewDetector(agent Agent, detect func(env Env) (Detection, bool)) Detector {
	return funcDetector{agent: agent, detect: detect}
}

type funcDetector struct {
	agent  Agent
	detect func(Env) (Detection, bool)
}

func (d funcDetector) Agent() Agent                     { return d.agent }
func (d funcDetector) Detect(env Env) (Detection, bool) { return d.detect(env) }

type options struct {
	env               Env
	detectors         []Detector
	ciDetectors       []CIDetector
	terminalDetectors []TerminalDetector
	remoteDetectors   []RemoteDetector
	tty               TTY
	process           ProcessTree
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

// WithTerminalDetectors adds terminal detectors ahead of the built-in ones.
// Detectors are tried in the order given, and the first match wins.
func WithTerminalDetectors(detectors ...TerminalDetector) Option {
	return func(o *options) {
		o.terminalDetectors = append(append([]TerminalDetector{}, detectors...), o.terminalDetectors...)
	}
}

// WithOnlyTerminalDetectors replaces the built-in terminal detectors entirely.
// Passing no detectors disables terminal detection.
func WithOnlyTerminalDetectors(detectors ...TerminalDetector) Option {
	return func(o *options) {
		o.terminalDetectors = append([]TerminalDetector{}, detectors...)
	}
}

// WithRemoteDetectors adds remote-layer detectors ahead of the built-in ones.
// Unlike the agent axis, every matching detector is reported, so ordering
// affects only the order of Result.Remote.
func WithRemoteDetectors(detectors ...RemoteDetector) Option {
	return func(o *options) {
		o.remoteDetectors = append(append([]RemoteDetector{}, detectors...), o.remoteDetectors...)
	}
}

// WithOnlyRemoteDetectors replaces the built-in remote detectors entirely.
// Passing no detectors disables remote detection.
func WithOnlyRemoteDetectors(detectors ...RemoteDetector) Option {
	return func(o *options) {
		o.remoteDetectors = append([]RemoteDetector{}, detectors...)
	}
}

// WithDetectors adds detectors ahead of the built-in ones, so a custom
// orchestrator is reported as the primary layer over the runtime it drives.
// Detectors are tried in the order given.
func WithDetectors(detectors ...Detector) Option {
	return func(o *options) {
		o.detectors = append(append([]Detector{}, detectors...), o.detectors...)
	}
}

// WithOnlyDetectors replaces the built-in detectors entirely.
func WithOnlyDetectors(detectors ...Detector) Option {
	return func(o *options) {
		o.detectors = append([]Detector{}, detectors...)
	}
}

// WithCIDetectors adds CI detectors ahead of the built-in ones, so a platform
// this package does not support is reported over the generic CI convention.
// Detectors are tried in the order given, and the first match wins.
func WithCIDetectors(detectors ...CIDetector) Option {
	return func(o *options) {
		o.ciDetectors = append(append([]CIDetector{}, detectors...), o.ciDetectors...)
	}
}

// WithOnlyCIDetectors replaces the built-in CI detectors entirely. Passing no
// detectors disables CI detection.
func WithOnlyCIDetectors(detectors ...CIDetector) Option {
	return func(o *options) {
		o.ciDetectors = append([]CIDetector{}, detectors...)
	}
}

// Detect inspects an environment and returns every supported agent found in
// it, plus the CI platform and terminal status. With no options it inspects
// the current process, including its terminal.
func Detect(opts ...Option) Result {
	config := options{
		env:               processEnv{},
		detectors:         builtinDetectors,
		ciDetectors:       builtinCIDetectors,
		terminalDetectors: builtinTerminalDetectors,
		remoteDetectors:   builtinRemoteDetectors,
		inspectTTY:        true,
		inspectProcess:    true,
	}
	for _, opt := range opts {
		opt(&config)
	}

	result := Result{TTY: config.tty, Process: config.process}
	if config.inspectTTY {
		result.TTY = InspectTTY()
	}
	if config.inspectProcess {
		result.Process = inspectProcessTree()
	}
	for _, detector := range config.detectors {
		detection, ok := detector.Detect(config.env)
		if !ok {
			continue
		}
		// Detectors fill in only what their agent advertises; the defaults
		// shared by every detection are applied once, here.
		if detection.Agent == "" {
			detection.Agent = detector.Agent()
		}
		if detection.Kind == "" {
			detection.Kind = detection.Agent.Kind()
		}
		if detection.Confidence == "" {
			detection.Confidence = ConfidenceDefinite
		}
		if detection.Sandbox.Network == "" {
			detection.Sandbox.Network = NetworkUnknown
		}
		// An ancestor running this agent's executable proves it is still
		// alive, which no environment variable can.
		if ancestor, ok := result.Process.FindAgent(detection.Agent); ok {
			detection.AncestorPID = ancestor.PID
		}
		result.Layers = append(result.Layers, detection)
	}
	for _, detector := range config.ciDetectors {
		ci, ok := detector.Detect(config.env)
		if !ok {
			continue
		}
		// Only the first, most specific platform is reported; every platform
		// also sets the generic CI variable, so later matches are redundant.
		if ci.Provider == "" {
			ci.Provider = detector.Provider()
		}
		if ci.Confidence == "" {
			ci.Confidence = ConfidenceDefinite
		}
		ci.Detected = true
		result.CI = ci
		break
	}
	if !result.CI.Detected {
		result.CI.Provider = CIProviderUnknown
		result.CI.Confidence = ConfidenceUnknown
	}

	for _, detector := range config.remoteDetectors {
		remote, ok := detector.Detect(config.env)
		if !ok {
			continue
		}
		// Every match is reported: an SSH session into a Codespace running
		// tmux is three concurrent layers, not a precedence contest.
		if remote.Platform == "" {
			remote.Platform = detector.Platform()
		}
		if remote.Kind == "" {
			remote.Kind = remote.Platform.Kind()
		}
		if remote.Confidence == "" {
			remote.Confidence = ConfidenceDefinite
		}
		result.Remote = append(result.Remote, remote)
	}

	for _, detector := range config.terminalDetectors {
		terminal, ok := detector.Detect(config.env)
		if !ok {
			continue
		}
		if terminal.Program == "" {
			terminal.Program = detector.Program()
		}
		if terminal.Confidence == "" {
			terminal.Confidence = ConfidenceDefinite
		}
		terminal.Detected = true
		result.Terminal = terminal
		break
	}
	if !result.Terminal.Detected {
		result.Terminal.Program = TerminalUnknown
		result.Terminal.Confidence = ConfidenceUnknown
		result.Terminal.Term, _ = Value(config.env, "TERM")
	}
	if _, ok := result.Multiplexer(); ok && result.Terminal.Confidence == ConfidenceDefinite {
		// The multiplexer server keeps the environment of whichever client
		// started it, so any terminal identity here may name a terminal that
		// is not the one displaying this pane.
		result.Terminal.Confidence = ConfidenceProbable
	}
	return result
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
// convenience for callers that build their own detector pipelines.
func Environ() Env { return EnvironEnv(os.Environ()) }
