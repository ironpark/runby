package runby_test

import (
	"reflect"
	"testing"

	"github.com/ironpark/runby"
)

// A remote driver that names its Kind is what makes Result.Multiplexer report
// the layer, and with it the staleness caveat that caps terminal confidence.
// Before drivers carried their own metadata this was reachable only for the
// built-in platforms, and a custom multiplexer was silently ignored.
func TestCustomMultiplexerCapsTerminalConfidence(t *testing.T) {
	acme := runby.RemoteDriver{
		Platform:    "acme-mux",
		Kind:        runby.RemoteKindMultiplexer,
		Executables: []string{"acme-mux"},
		Detect: func(env runby.Env) (runby.Remote, bool) {
			id, ok := runby.NewEnvReader(env).Value("ACME_MUX")
			if !ok {
				return runby.Remote{}, false
			}
			return runby.Remote{SessionID: id, Axis: runby.Axis{Evidence: []string{"ACME_MUX"}}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_MUX=s-1", "TERM_PROGRAM=ghostty"}),
		runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), acme)...),
	)

	mux, ok := result.Multiplexer()
	if !ok || mux.Platform != "acme-mux" {
		t.Fatalf("Multiplexer() = %#v, %v", mux, ok)
	}
	if result.Terminal.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Terminal = %#v, want confidence capped to probable", result.Terminal)
	}
}

// The same driver's Executables reach the ancestor chain, so a live server
// cancels the downgrade exactly as tmux does.
func TestCustomTerminalAndRemoteDriversAreCorroborated(t *testing.T) {
	term := runby.TerminalDriver{
		Program:     "acme-term",
		Executables: []string{"acme-term"},
		Detect: func(env runby.Env) (runby.Terminal, bool) {
			if !runby.NewEnvReader(env).IsTrue("ACME_TERM") {
				return runby.Terminal{}, false
			}
			return runby.Terminal{Axis: runby.Axis{Evidence: []string{"ACME_TERM"}}}, true
		},
	}
	vpn := runby.RemoteDriver{
		Platform:    "acme-vpn",
		Kind:        runby.RemoteKindEnvironment,
		Executables: []string{"acme-vpnd"},
		Detect: func(env runby.Env) (runby.Remote, bool) {
			if !runby.NewEnvReader(env).IsTrue("ACME_VPN") {
				return runby.Remote{}, false
			}
			return runby.Remote{Axis: runby.Axis{Evidence: []string{"ACME_VPN"}}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_TERM=1", "ACME_VPN=1"}),
		runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), term, vpn)...),
		runby.WithProcessTree(runby.ProcessTree{
			Inspected: true,
			Supported: true,
			Ancestors: []runby.Process{
				{PID: 10, PPID: 20, Name: "acme-vpnd"},
				{PID: 20, PPID: 1, Name: "acme-term"},
			},
		}),
	)

	if result.Terminal.AncestorPID != 20 {
		t.Errorf("Terminal.AncestorPID = %d, want 20", result.Terminal.AncestorPID)
	}
	layer, ok := result.Remote("acme-vpn")
	if !ok || layer.AncestorPID != 10 {
		t.Errorf("Remote = %#v, want AncestorPID 10", result.Remotes)
	}
	// The injected tree carried no labels; Detect applied them from the drivers.
	if got := result.Process.Ancestors[1].Terminal; got != "acme-term" {
		t.Errorf("ancestor label = %q, want acme-term", got)
	}
}

// A driver that names no Kind gets the unknown sentinel rather than the zero
// value, so the serialized output never carries an empty string.
func TestDriversWithoutKindReportUnknown(t *testing.T) {
	result := runby.Detect(
		runby.WithEnviron([]string{"ACME=1"}),
		runby.WithOnlyDrivers(runby.AgentDriver{
			Agent: "acme",
			Detect: func(env runby.Env) (runby.Agent, bool) {
				return runby.Agent{Axis: runby.Axis{Evidence: []string{"ACME"}}}, true
			},
		}),
	)
	primary, _ := result.Primary()
	if primary.Kind != runby.KindUnknown {
		t.Fatalf("Kind = %q, want %q", primary.Kind, runby.KindUnknown)
	}
}

// BuiltinDrivers is the raw material WithOnlyDrivers needs to mean anything
// other than "these drivers and nothing else". Filtering it drops a built-in
// from one call; internal/silencetest covers dropping one for the whole
// process by registering a driver that never matches.
func TestBuiltinDriversCanDropOneProduct(t *testing.T) {
	var drivers []runby.Driver
	for _, driver := range runby.BuiltinDrivers() {
		if agent, ok := driver.(runby.AgentDriver); ok && agent.Agent == runby.AgentCodex {
			continue
		}
		drivers = append(drivers, driver)
	}

	env := runby.WithEnviron([]string{"CODEX_THREAD_ID=t-1", "CLAUDECODE=1", "GITHUB_ACTIONS=true"})
	result := runby.Detect(env, runby.WithOnlyDrivers(drivers...))
	if _, ok := result.Agent(runby.AgentCodex); ok {
		t.Error("codex was detected after its driver was filtered out")
	}
	if _, ok := result.Agent(runby.AgentClaudeCode); !ok {
		t.Error("dropping codex also silenced claude-code")
	}
	// Filtering the agent axis leaves the other four intact, which passing a
	// hand-built slice per axis would not.
	if !result.IsCI() {
		t.Error("filtering an agent driver disabled the CI axis")
	}
}

// The returned slice is the caller's, so appending to it or reordering it
// cannot reach the next call.
func TestBuiltinDriversAreCopied(t *testing.T) {
	first := runby.BuiltinDrivers()
	if len(first) == 0 {
		t.Fatal("BuiltinDrivers returned nothing")
	}
	first[0] = runby.AgentDriver{Agent: "tampered"}
	// Drivers hold func fields, so they are compared by the identity they
	// carry rather than with ==.
	if agent, ok := runby.BuiltinDrivers()[0].(runby.AgentDriver); !ok || agent.Agent == "tampered" {
		t.Fatal("BuiltinDrivers handed out a shared slice")
	}
	// Every built-in product is present, on every axis. The internal table
	// tests pin each axis to its identity list; this pins that BuiltinDrivers
	// actually reaches all five, which adding a sixth axis would break.
	want := len(runby.AgentNames()) + len(runby.CIProviders()) + len(runby.TerminalPrograms()) +
		len(runby.RemotePlatforms()) + len(runby.RunnerTools())
	if got := len(runby.BuiltinDrivers()); got != want {
		t.Fatalf("BuiltinDrivers has %d drivers, the identity lists name %d products", got, want)
	}
}

// WithDrivers adds to the set a call would otherwise run, so a custom driver
// and the built-in ones are both in play. Before it existed the only way to say
// this was WithOnlyDrivers(append(BuiltinDrivers(), acme)...), which also
// silently dropped anything added through Register.
func TestWithDriversExtendsTheDefaultSet(t *testing.T) {
	acme := runby.AgentDriver{
		Agent:  "acme",
		Kind:   runby.KindHarness,
		Models: runby.ModelsMultiVendor,
		Detect: func(env runby.Env) (runby.Agent, bool) {
			r := runby.NewEnvReader(env)
			id, ok := r.Value("ACME_RUN_ID")
			if !ok {
				return runby.Agent{}, false
			}
			return runby.Agent{SessionID: id, Axis: runby.Axis{Evidence: r.Evidence()}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=r-1", "CODEX_SANDBOX=seatbelt"}),
		runby.WithDrivers(acme),
	)

	custom, ok := result.Agent("acme")
	if !ok || custom.SessionID != "r-1" {
		t.Fatalf("Layer(acme) = %#v, %v", custom, ok)
	}
	if !reflect.DeepEqual(custom.Evidence, []string{"ACME_RUN_ID"}) {
		t.Errorf("Evidence = %v, want the name the reader recorded", custom.Evidence)
	}
	// The built-in axis still runs, which is the whole difference from
	// WithOnlyDrivers. Codex is Level1 and acme is Level2, so the ladder puts
	// the custom harness first whatever order the drivers arrived in.
	if _, ok := result.Agent(runby.AgentCodex); !ok {
		t.Fatalf("the built-in drivers were dropped: %#v", result.Agents)
	}
	if got := result.Chain(); got != "acme>codex" {
		t.Errorf("Chain() = %q, want acme>codex", got)
	}
}

// A driver whose identity matches one already in the set replaces it rather
// than running beside it, so overriding a stale built-in yields one layer.
func TestWithDriversReplacesAMatchingBuiltin(t *testing.T) {
	silenced := runby.AgentDriver{
		Agent:  runby.AgentCodex,
		Kind:   runby.KindHarness,
		Models: runby.ModelsFirstParty,
		Detect: func(runby.Env) (runby.Agent, bool) { return runby.Agent{}, false },
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"CODEX_SANDBOX=seatbelt"}),
		runby.WithDrivers(silenced),
	)
	if result.IsAgent() {
		t.Fatalf("the replaced built-in still ran: %#v", result.Agents)
	}
}
