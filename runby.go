// Package runby detects the coding agent or orchestrator that started the
// current process from environment variables inherited by that process.
//
// The common case is one call, whose result answers every axis:
//
//	result := runby.Current()
//	if result.IsAgent() {
//		log.Printf("run by %s", result.Chain())
//	}
//
// Current caches, so the result is worth passing down rather than re-fetching.
// A program that branches on it is also easier to test that way: Detect with
// WithEnviron builds a Result for any environment, while a test that sets
// variables and calls Current sees whatever the first caller in the test binary
// already cached.
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
// Options divide into three tiers, and most programs stay in the first.
//
//   - No options at all. Current answers for this process, and Detect() with
//     no options does the same without the cache.
//   - Describing something that is not this process: a wrapper classifying
//     another process from its /proc entry or its exec.Cmd.Env, or an
//     environment recorded earlier and analyzed later. WithEnviron supplies
//     the environment, and WithTTY and WithProcessTree supply the two axes
//     that cannot be read from it. WithoutTTY and WithoutProcessTree skip
//     those axes when only the environment matters.
//   - Writing a driver, and testing one: WithEnv is the general form of
//     WithEnviron, WithDrivers adds a driver to the set this call would
//     otherwise run, and WithOnlyDrivers runs an exact set.
//
// Every axis is extensible by the same means. A product this package does not
// support is added by passing a driver — AgentDriver, CIDriver, TerminalDriver,
// RemoteDriver, or RunnerDriver — to Detect through WithDrivers, or to the
// whole process through Register. A driver carries the rule for detecting the
// product together with the facts no environment can supply, such as what a
// detection proves and which binaries the product runs as, so the built-in
// products and yours are declared exactly the same way. Read the environment
// through an EnvReader and the driver reports its own evidence.
package runby

import "sync"

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
//
// Pair it with LookupFunc to inspect an environment through a lookup function
// rather than a slice:
//
//	Detect(WithEnv(LookupFunc(os.LookupEnv)))
func WithEnv(env Env) Option {
	return func(o *options) {
		o.env = env
		o.inspectTTY = false
		o.inspectProcess = false
	}
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

// WithDrivers adds drivers to the set this call would otherwise run: the
// built-in drivers with anything added through Register merged in. A driver
// whose identity matches one already there replaces it rather than running
// beside it, which is the same rule Register follows.
//
// It is the per-call counterpart to Register, for a driver that belongs to one
// call rather than to the whole process:
//
//	Detect(WithDrivers(acme))
//
// Use Register instead when the driver should reach Current and the CLI, which
// take no options. Use WithOnlyDrivers instead when the point is to exclude the
// built-in set rather than extend it.
//
// Drivers are sorted onto their own axis, so one call covers as many axes as
// the drivers span. The agent axis is ordered by the ladder rather than by the
// order given here.
func WithDrivers(drivers ...Driver) Option {
	var added registry
	for _, driver := range drivers {
		if driver == nil {
			panic("runby: WithDrivers called with a nil driver")
		}
		driver.addTo(&added)
	}
	added.check()
	return func(o *options) {
		o.agentDrivers = merge(added.agents, o.agentDrivers)
		o.ciDrivers = merge(added.ci, o.ciDrivers)
		o.terminalDrivers = merge(added.terminals, o.terminalDrivers)
		o.remoteDrivers = merge(added.remotes, o.remoteDrivers)
		o.runnerDrivers = merge(added.runners, o.runnerDrivers)
	}
}

// WithOnlyDrivers runs exactly the drivers given and nothing else: no built-in
// driver and nothing added through Register participates in this call. With no
// drivers at all, no axis derived from the environment is detected.
//
// Reach for it only to run a fixed set. Adding a driver to the default set is
// what WithDrivers is for. The two things this option is for do not divide by
// axis:
//
//   - Testing a driver in isolation, so a fixture cannot be answered by a
//     built-in that happens to match it too.
//   - Pinning a test in a program where something has registered a driver. The
//     registry is process-wide, so without this a blank import anywhere in the
//     build could change what a test observes.
//
// Pair it with BuiltinDrivers to run the built-in set minus one of its own,
// which is chiefly useful in a test standing in for a built-in product; see
// BuiltinDrivers for how to filter it.
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

	result.Agents = detectAgents(config, result.Process)
	result.CI = detectCI(config)
	result.Remotes = detectRemote(config, result.Process)
	result.Runners = detectRunners(config, result.Process)
	applyMultiplexerStaleness(&result)
	result.Terminal = detectTerminal(config, result)
	return result
}

// applyMultiplexerStaleness weakens the environment-derived layers when a
// terminal multiplexer is present. A multiplexer server keeps the environment
// of whichever client started it and cannot refresh a running pane, so an
// agent or runner marker seen through one may have been left by a session that
// has since ended. That is the same caveat detectTerminal applies to the
// terminal identity, moved onto the layers it equally affects.
//
// A live ancestor settles the question the other way, exactly as it does for
// the terminal: a layer corroborated by AncestorPID is running now, so its
// confidence stands. The CI axis is left alone because a CI job is not run
// through a multiplexer's stored environment, and the remote axis is left
// alone because the multiplexer is itself one of its layers.
func applyMultiplexerStaleness(result *Result) {
	if _, muxed := result.Multiplexer(); !muxed {
		return
	}
	for i := range result.Agents {
		if result.Agents[i].AncestorPID == 0 && result.Agents[i].Confidence == ConfidenceDefinite {
			result.Agents[i].Confidence = ConfidenceProbable
		}
	}
	for i := range result.Runners {
		if result.Runners[i].AncestorPID == 0 && result.Runners[i].Confidence == ConfidenceDefinite {
			result.Runners[i].Confidence = ConfidenceProbable
		}
	}
}

// executableLabels gathers the name-to-product mapping from every configured
// driver that names its executables. The CI axis is absent because a CI driver
// names no executables; see CIDriver.
func (config options) executableLabels() executableLabels {
	labels := make(executableLabels)
	addLabels(labels, config.agentDrivers)
	addLabels(labels, config.terminalDrivers)
	addLabels(labels, config.remoteDrivers)
	addLabels(labels, config.runnerDrivers)
	return labels
}

// ancestorLabel is what a live ancestor running one of this driver's
// executables is labelled as: the driver's identity on its own axis's field
// of Process. Keeping it on the driver type puts the axis-specific fact
// beside the identity it labels.
func (d AgentDriver) ancestorLabel() ([]string, Process) {
	return d.Executables, Process{Agent: d.Agent}
}

func (d TerminalDriver) ancestorLabel() ([]string, Process) {
	return d.Executables, Process{Terminal: d.Program}
}

func (d RemoteDriver) ancestorLabel() ([]string, Process) {
	return d.Executables, Process{Remote: d.Platform}
}

func (d RunnerDriver) ancestorLabel() ([]string, Process) {
	return d.Executables, Process{Runner: d.Tool}
}

// addLabels records one axis's executable names into labels.
func addLabels[D interface{ ancestorLabel() ([]string, Process) }](labels executableLabels, drivers []D) {
	for _, driver := range drivers {
		names, product := driver.ancestorLabel()
		labels.add(names, product)
	}
}

// detectAgents reports every agent layer, most specific orchestrator first.
func detectAgents(config options, tree ProcessTree) []Agent {
	var layers []Agent
	for _, driver := range config.agentDrivers {
		detection, ok := driver.Detect(config.env)
		if !ok {
			continue
		}
		// Drivers fill in only what their agent advertises; the identity and
		// the defaults shared by every detection are applied once, here.
		detection.Name = driver.Agent
		detection.Kind = driver.Kind
		detection.Models = driver.Models
		if detection.Kind == "" {
			detection.Kind = KindUnknown
		}
		if detection.Models == "" {
			detection.Models = ModelsUnknown
		}
		if detection.Sandbox.Network == "" {
			detection.Sandbox.Network = NetworkUnknown
		}
		detection.applyDefaults()
		// An ancestor running this agent's executable proves it is still
		// alive, which no environment variable can.
		detection.AncestorPID = tree.pidOf(func(p Process) bool { return p.Agent == detection.Name })
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
	return CI{Provider: CIUnknown, Axis: Axis{Confidence: ConfidenceUnknown}}
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
		terminal.Term, _ = envValue(config.env, "TERM")
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
// so repeated calls are free.
//
// The cache is process-wide and is filled by whichever call comes first, so
// Current cannot see an environment a test sets up afterwards with t.Setenv.
// Build the Result explicitly for those:
//
//	runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true"}))
//
// Use Detect directly, too, to observe changes made by os.Setenv after the
// first call.
func Current() Result {
	currentOnce.Do(func() {
		// Recorded before detecting so that a Register racing this call is
		// reported as the mistake it is rather than being silently dropped.
		detected.Store(true)
		currentResult = Detect()
	})
	return currentResult
}
