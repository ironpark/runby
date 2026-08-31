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

// builtinDetectors is ordered from the most specific orchestrator to the
// underlying runtime. Result.Primary reports the first match, so this order is
// the precedence contract.
var builtinDetectors = []Detector{
	NewDetector(AgentPaseo, detectPaseo),
	NewDetector(AgentCodex, detectCodex),
	NewDetector(AgentClaudeCode, detectClaudeCode),
	NewDetector(AgentCursor, detectCursor),
	NewDetector(AgentOpenCode, detectOpenCode),
	NewDetector(AgentZed, detectZed),
	NewDetector(AgentAmp, detectAmp),
	NewDetector(AgentAntigravity2, detectAntigravity2),
}

// Detectors returns the built-in detectors in precedence order. The returned
// slice is a copy and may be reordered or filtered before being passed back
// through WithDetectors.
func Detectors() []Detector {
	detectors := make([]Detector, len(builtinDetectors))
	copy(detectors, builtinDetectors)
	return detectors
}

type options struct {
	env       Env
	detectors []Detector
	terminal  Terminal
	// inspectTerminal is only true when the environment being inspected is
	// this process's own, so that the standard streams describe the same
	// process as the detected layers.
	inspectTerminal bool
}

// Option configures Detect.
type Option func(*options)

// WithEnviron inspects an explicit environment given as "NAME=value" entries,
// instead of the process environment. Because that environment does not
// necessarily belong to this process, the terminal is not inspected unless
// WithTerminal is also given.
func WithEnviron(environ []string) Option {
	return WithEnv(EnvironEnv(environ))
}

// WithEnv inspects an explicit Env instead of the process environment. As with
// WithEnviron, the terminal is not inspected.
func WithEnv(env Env) Option {
	return func(o *options) {
		o.env = env
		o.inspectTerminal = false
	}
}

// WithLookup inspects an environment through a lookup function such as
// os.LookupEnv. As with WithEnviron, the terminal is not inspected.
func WithLookup(lookup func(name string) (string, bool)) Option {
	return WithEnv(lookupEnv(lookup))
}

// WithoutTerminal skips terminal inspection, avoiding its system calls when
// only the agent layers are needed.
func WithoutTerminal() Option {
	return func(o *options) {
		o.inspectTerminal = false
		o.terminal = Terminal{}
	}
}

// WithTerminal sets the terminal status explicitly instead of inspecting the
// standard streams. It is intended for wrappers that already know the terminal
// status of the environment they are describing, and for tests.
func WithTerminal(terminal Terminal) Option {
	return func(o *options) {
		o.inspectTerminal = false
		o.terminal = terminal
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

// Detect inspects an environment and returns every supported agent found in
// it. With no options it inspects the current process, including its terminal.
func Detect(opts ...Option) Result {
	config := options{
		env:             processEnv{},
		detectors:       builtinDetectors,
		inspectTerminal: true,
	}
	for _, opt := range opts {
		opt(&config)
	}

	result := Result{Terminal: config.terminal}
	if config.inspectTerminal {
		result.Terminal = InspectTerminal()
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
		result.Layers = append(result.Layers, detection)
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
// cached Current result. It is false when only a KindHost layer, such as an
// editor-owned terminal, was detected.
func IsAgent() bool { return Current().IsAgent() }

// Environ returns the current process environment as an Env. It is a
// convenience for callers that build their own detector pipelines.
func Environ() Env { return EnvironEnv(os.Environ()) }
