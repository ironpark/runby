package runby

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// Driver is a driver for any one axis. It exists so that Register can take all
// five kinds in a single call:
//
//	func init() {
//		runby.Register(
//			runby.AgentDriver{Agent: "acme", Kind: runby.KindOrchestrator, …},
//			runby.RunnerDriver{Tool: "acme-task", Kind: runby.RunnerKindScript, …},
//		)
//	}
//
// The interface is closed: its only method is unexported, so exactly the five
// driver types in this package implement it. A new axis cannot be added from
// outside, because an axis is not just a driver — it is a field on Result, a
// contribution to the ancestor labels, and a rule for how matches combine.
type Driver interface {
	addTo(*registry)
}

func (d AgentDriver) addTo(r *registry)    { r.agents = append(r.agents, d) }
func (d CIDriver) addTo(r *registry)       { r.ci = append(r.ci, d) }
func (d TerminalDriver) addTo(r *registry) { r.terminals = append(r.terminals, d) }
func (d RemoteDriver) addTo(r *registry)   { r.remotes = append(r.remotes, d) }
func (d RunnerDriver) addTo(r *registry)   { r.runners = append(r.runners, d) }

type registry struct {
	mu        sync.RWMutex
	agents    []AgentDriver
	ci        []CIDriver
	terminals []TerminalDriver
	remotes   []RemoteDriver
	runners   []RunnerDriver
}

var (
	registered registry
	// detected records that Current has already computed and cached a result,
	// which is what makes a later Register a programming error rather than a
	// late but harmless addition.
	detected atomic.Bool
)

// Register adds drivers to the process-wide set that every Detect call uses,
// including the zero-argument one behind Current, IsAgent, and their
// neighbours. It is what lets a driver live in its own module and reach every
// caller through a blank import:
//
//	import _ "example.com/runby-acme"
//
// Registered drivers are tried ahead of the built-in ones on their axis, and a
// registered driver whose identity matches a built-in replaces it rather than
// running beside it — so a stale built-in can be corrected without waiting for
// a release of this package.
//
// Call it from an init function. Registering the same identity twice, or
// registering after Current has already computed its cached result, panics:
// both are mistakes whose quiet form is a detection that silently disagrees
// with the code that asked for it.
//
// # This is process-wide state
//
// A blank import anywhere in the build changes detection for the whole
// program, including for code that never asked. That is inherent to the
// pattern rather than a flaw in this implementation, and it is the same
// bargain database/sql makes. Two consequences are worth planning around:
//
//   - Do the blank import from a main package. A library that imports a driver
//     imposes it on every program that depends on the library, transitively and
//     invisibly.
//   - Tests that need a fixed driver set should use WithOnlyDrivers, which
//     ignores the registry and the built-ins alike and runs exactly what it is
//     given.
func Register(drivers ...Driver) {
	if detected.Load() {
		panic("runby: Register called after Current already computed its cached " +
			"result; register drivers from an init function so they are in place " +
			"before the first detection")
	}
	registered.mu.Lock()
	defer registered.mu.Unlock()
	for _, driver := range drivers {
		if driver == nil {
			panic("runby: Register called with a nil driver")
		}
		driver.addTo(&registered)
	}
	registered.check()
}

// check panics on a duplicate identity within the registered drivers of one
// axis. Two drivers for the same product would both match and report the same
// thing twice, so the ambiguity is refused at registration rather than carried
// into every result.
func (r *registry) check() {
	checkUnique("agent", r.agents, func(d AgentDriver) Agent { return d.Agent })
	checkUnique("CI", r.ci, func(d CIDriver) CIProvider { return d.Provider })
	checkUnique("terminal", r.terminals, func(d TerminalDriver) TerminalProgram { return d.Program })
	checkUnique("remote", r.remotes, func(d RemoteDriver) RemotePlatform { return d.Platform })
	checkUnique("runner", r.runners, func(d RunnerDriver) RunnerTool { return d.Tool })
}

func checkUnique[D any, ID comparable](axis string, drivers []D, identify func(D) ID) {
	seen := make(map[ID]bool, len(drivers))
	for _, driver := range drivers {
		id := identify(driver)
		if seen[id] {
			panic(fmt.Sprintf("runby: two %s drivers registered for %v", axis, id))
		}
		seen[id] = true
	}
}

// merge puts the registered drivers ahead of the built-in ones and drops any
// built-in a registered driver has taken over, so that replacing a built-in
// yields one layer rather than two.
func merge[D any, ID comparable](registeredDrivers, builtin []D, identify func(D) ID) []D {
	if len(registeredDrivers) == 0 {
		return builtin
	}
	replaced := make(map[ID]bool, len(registeredDrivers))
	for _, driver := range registeredDrivers {
		replaced[identify(driver)] = true
	}
	merged := cloneSlice(registeredDrivers)
	for _, driver := range builtin {
		if !replaced[identify(driver)] {
			merged = append(merged, driver)
		}
	}
	return merged
}

// defaults returns the driver set a Detect call starts from: the built-in
// drivers with the registered ones merged in. Detect and its tests both go
// through here so the two cannot describe different defaults.
func defaultOptions() options {
	registered.mu.RLock()
	defer registered.mu.RUnlock()
	return options{
		env: processEnv{},
		agentDrivers: merge(registered.agents, builtinAgentDrivers,
			func(d AgentDriver) Agent { return d.Agent }),
		ciDrivers: merge(registered.ci, builtinCIDrivers,
			func(d CIDriver) CIProvider { return d.Provider }),
		terminalDrivers: merge(registered.terminals, builtinTerminalDrivers,
			func(d TerminalDriver) TerminalProgram { return d.Program }),
		remoteDrivers: merge(registered.remotes, builtinRemoteDrivers,
			func(d RemoteDriver) RemotePlatform { return d.Platform }),
		runnerDrivers: merge(registered.runners, builtinRunnerDrivers,
			func(d RunnerDriver) RunnerTool { return d.Tool }),
		inspectTTY:     true,
		inspectProcess: true,
	}
}

// The Kind, Models, and Level methods on a product identity answer from the
// built-in tables first and fall through to here, so that a registered driver
// answers them the same way a built-in one does. Without this a caller would
// see Layer.Level report l3 while Agent.Level on the same agent reported
// unknown, which is the kind of split a registry exists to avoid.
//
// The registered slices hold a handful of entries, so a scan is cheaper than
// the map that would have to be invalidated on every Register.

// The driver slice is selected inside the lock rather than passed in, so that
// reading it is never racing a concurrent Register.
func lookupRegistered[D any, ID comparable, V any](
	slice func(*registry) []D, id ID, identify func(D) ID, value func(D) V, fallback V,
) V {
	registered.mu.RLock()
	defer registered.mu.RUnlock()
	for _, driver := range slice(&registered) {
		if identify(driver) == id {
			return value(driver)
		}
	}
	return fallback
}

func registeredAgentKind(a Agent) Kind {
	return lookupRegistered(func(r *registry) []AgentDriver { return r.agents }, a,
		func(d AgentDriver) Agent { return d.Agent },
		func(d AgentDriver) Kind { return d.Kind }, KindUnknown)
}

func registeredAgentModels(a Agent) ModelSource {
	return lookupRegistered(func(r *registry) []AgentDriver { return r.agents }, a,
		func(d AgentDriver) Agent { return d.Agent },
		func(d AgentDriver) ModelSource { return d.Models }, ModelsUnknown)
}

func registeredRemoteKind(p RemotePlatform) RemoteKind {
	return lookupRegistered(func(r *registry) []RemoteDriver { return r.remotes }, p,
		func(d RemoteDriver) RemotePlatform { return d.Platform },
		func(d RemoteDriver) RemoteKind { return d.Kind }, RemoteKindUnknown)
}

func registeredRunnerKind(t RunnerTool) RunnerKind {
	return lookupRegistered(func(r *registry) []RunnerDriver { return r.runners }, t,
		func(d RunnerDriver) RunnerTool { return d.Tool },
		func(d RunnerDriver) RunnerKind { return d.Kind }, RunnerKindUnknown)
}

// ladderRank orders the agent axis from the outermost layer inward, which is
// what Result.Primary means by "most specific".
//
// The order cannot be left to whoever registered or passed a driver first.
// Package initialization order is not something a driver author controls, and
// prepending would put a custom Level1 harness ahead of a built-in Level3
// orchestrator, which reports the runtime as the primary layer over the
// orchestrator driving it — exactly backwards. Sorting by the level a driver
// declares makes the order a property of what the product is.
//
// A driver that declares neither Kind nor Models has no place on the ladder
// and sorts last, because this package cannot claim it is an orchestrator.
// Declare both to be placed correctly.
func ladderRank(driver AgentDriver) int {
	switch level(driver.Kind, driver.Models) {
	case Level3:
		return 0
	case Level2:
		return 1
	case Level1:
		return 2
	default:
		return 3
	}
}

// sortByLadder orders agent drivers outermost first. The sort is stable, so
// drivers on the same rung keep the order they were given, and the built-in
// table — already written in ladder order — is left untouched.
func sortByLadder(drivers []AgentDriver) []AgentDriver {
	sorted := cloneSlice(drivers)
	sort.SliceStable(sorted, func(i, j int) bool {
		return ladderRank(sorted[i]) < ladderRank(sorted[j])
	})
	return sorted
}

// BuiltinDrivers returns every driver this package ships, as a fresh slice the
// caller may filter. It is the raw material for WithOnlyDrivers, which
// otherwise runs nothing but the drivers given.
//
// Adding a driver to the built-in set needs neither: that is what WithDrivers
// does for one call, and Register for the whole process. What this is for is
// dropping a built-in from one call, which is chiefly useful in a test that
// wants the built-in set minus the product it is standing in for:
//
//	var drivers []Driver
//	for _, driver := range BuiltinDrivers() {
//		if agent, ok := driver.(AgentDriver); ok && agent.Agent == AgentCodex {
//			continue
//		}
//		drivers = append(drivers, driver)
//	}
//
// To silence a built-in for the whole program, including Current and the CLI,
// register a driver with the same identity whose Detect never matches. That
// reaches the entry points which take no options, and this does not.
//
// The order is each axis's own: the agent axis is in ladder order, and the CI
// and terminal axes are in the order their first match is decided.
func BuiltinDrivers() []Driver {
	drivers := make([]Driver, 0, len(builtinAgentDrivers)+len(builtinCIDrivers)+
		len(builtinTerminalDrivers)+len(builtinRemoteDrivers)+len(builtinRunnerDrivers))
	drivers = appendDrivers(drivers, builtinAgentDrivers)
	drivers = appendDrivers(drivers, builtinCIDrivers)
	drivers = appendDrivers(drivers, builtinTerminalDrivers)
	drivers = appendDrivers(drivers, builtinRemoteDrivers)
	drivers = appendDrivers(drivers, builtinRunnerDrivers)
	return drivers
}

// appendDrivers widens one axis's table into the common Driver slice. Each
// conversion copies the driver, which is what keeps the shipped tables beyond
// a caller's reach.
func appendDrivers[D Driver](dst []Driver, src []D) []Driver {
	for _, driver := range src {
		dst = append(dst, driver)
	}
	return dst
}
