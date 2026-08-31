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
			id, ok := runby.Value(env, "ACME_MUX")
			if !ok {
				return runby.Remote{}, false
			}
			return runby.Remote{SessionID: id, Evidence: runby.PresentNames(env, "ACME_MUX")}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_MUX=s-1", "TERM_PROGRAM=ghostty"}),
		runby.WithRemoteDrivers(acme),
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
			if !runby.IsTrue(env, "ACME_TERM") {
				return runby.Terminal{}, false
			}
			return runby.Terminal{Evidence: runby.PresentNames(env, "ACME_TERM")}, true
		},
	}
	vpn := runby.RemoteDriver{
		Platform:    "acme-vpn",
		Kind:        runby.RemoteKindEnvironment,
		Executables: []string{"acme-vpnd"},
		Detect: func(env runby.Env) (runby.Remote, bool) {
			if !runby.IsTrue(env, "ACME_VPN") {
				return runby.Remote{}, false
			}
			return runby.Remote{Evidence: runby.PresentNames(env, "ACME_VPN")}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_TERM=1", "ACME_VPN=1"}),
		runby.WithTerminalDrivers(term),
		runby.WithRemoteDrivers(vpn),
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
	layer, ok := result.GetRemote("acme-vpn")
	if !ok || layer.AncestorPID != 10 {
		t.Errorf("Remote = %#v, want AncestorPID 10", result.Remote)
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
		runby.WithOnlyAgentDrivers(runby.AgentDriver{
			Agent: "acme",
			Detect: func(env runby.Env) (runby.Detection, bool) {
				return runby.Detection{Evidence: runby.PresentNames(env, "ACME")}, true
			},
		}),
	)
	primary, _ := result.Primary()
	if primary.Kind != runby.KindUnknown {
		t.Fatalf("Kind = %q, want %q", primary.Kind, runby.KindUnknown)
	}
}

func TestMarkerHelpers(t *testing.T) {
	env := runby.EnvironEnv([]string{"A=1", "B=false", "C=x", "EMPTY=", "TERM_PROGRAM=WezTerm"})
	for _, test := range []struct {
		name   string
		marker runby.Marker
		want   bool
	}{
		{"set both", runby.MarkerSet("A", "C"), true},
		{"set missing", runby.MarkerSet("A", "MISSING"), false},
		{"set empty is not set", runby.MarkerSet("EMPTY"), false},
		{"true", runby.MarkerTrue("A"), true},
		{"true on false", runby.MarkerTrue("A", "B"), false},
		{"true on non-boolean", runby.MarkerTrue("C"), false},
		{"term program folds case", runby.MarkerTermProgram("wezterm"), true},
		{"term program mismatch", runby.MarkerTermProgram("ghostty"), false},
	} {
		if got := test.marker(env); got != test.want {
			t.Errorf("%s = %v, want %v", test.name, got, test.want)
		}
	}

	if !runby.AnyPresent(env, "MISSING", "C") {
		t.Error("AnyPresent missed a set name")
	}
	if runby.AnyPresent(env, "MISSING", "EMPTY") {
		t.Error("AnyPresent matched an unset name")
	}
	if runby.AnyPresent(env) {
		t.Error("AnyPresent matched with no names")
	}
}

func TestCollectExtra(t *testing.T) {
	env := runby.EnvironEnv([]string{"A=1", "EMPTY="})
	got := runby.CollectExtra(env, map[string]string{"acme.a": "A", "acme.missing": "MISSING", "acme.empty": "EMPTY"})
	if want := map[string]string{"acme.a": "1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("CollectExtra = %#v, want %#v", got, want)
	}
	// Nothing present carries no map, so a detection without context has none.
	if got := runby.CollectExtra(env, map[string]string{"acme.missing": "MISSING"}); got != nil {
		t.Fatalf("CollectExtra = %#v, want nil", got)
	}
}

// The built-in tables are the single place a product is registered, so the
// exported lists must agree with the drivers they are derived from.
func TestDriverTablesMatchTheirIdentityLists(t *testing.T) {
	agents := runby.Agents()
	drivers := runby.AgentDrivers()
	if len(agents) != len(drivers) {
		t.Fatalf("Agents() has %d entries, AgentDrivers() has %d", len(agents), len(drivers))
	}
	for i, driver := range drivers {
		if driver.Agent != agents[i] {
			t.Errorf("driver %d is %q, Agents()[%d] is %q", i, driver.Agent, i, agents[i])
		}
		if driver.Kind != driver.Agent.Kind() {
			t.Errorf("%q: driver Kind %q, Agent.Kind() %q", driver.Agent, driver.Kind, driver.Agent.Kind())
		}
		if driver.Detect == nil {
			t.Errorf("%q has no Detect", driver.Agent)
		}
	}

	for i, driver := range runby.RemoteDrivers() {
		if got := runby.RemotePlatforms()[i]; driver.Platform != got {
			t.Errorf("driver %d is %q, RemotePlatforms()[%d] is %q", i, driver.Platform, i, got)
		}
		if driver.Kind != driver.Platform.Kind() {
			t.Errorf("%q: driver Kind %q, Platform.Kind() %q", driver.Platform, driver.Kind, driver.Platform.Kind())
		}
	}
	for i, driver := range runby.TerminalDrivers() {
		if got := runby.TerminalPrograms()[i]; driver.Program != got {
			t.Errorf("driver %d is %q, TerminalPrograms()[%d] is %q", i, driver.Program, i, got)
		}
	}
	for i, driver := range runby.CIDrivers() {
		if got := runby.CIProviders()[i]; driver.Provider != got {
			t.Errorf("driver %d is %q, CIProviders()[%d] is %q", i, driver.Provider, i, got)
		}
	}
}

// The returned slices are copies, so filtering one and passing it back cannot
// corrupt the built-in tables.
func TestDriverListsAreCopies(t *testing.T) {
	drivers := runby.AgentDrivers()
	drivers[0] = runby.AgentDriver{Agent: "tampered"}
	if again := runby.AgentDrivers(); again[0].Agent == "tampered" {
		t.Fatal("AgentDrivers() returned the built-in table itself")
	}

	// Dropping a driver disables exactly that agent.
	var withoutCodex []runby.AgentDriver
	for _, driver := range runby.AgentDrivers() {
		if driver.Agent != runby.AgentCodex {
			withoutCodex = append(withoutCodex, driver)
		}
	}
	result := runby.Detect(
		runby.WithEnviron([]string{"CODEX_THREAD_ID=t-1", "CLAUDECODE=1"}),
		runby.WithOnlyAgentDrivers(withoutCodex...),
	)
	if result.Has(runby.AgentCodex) || !result.Has(runby.AgentClaudeCode) {
		t.Fatalf("Layers = %#v", result.Layers)
	}
}
