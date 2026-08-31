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
			return runby.Remote{SessionID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_MUX")}}, true
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
			if !runby.IsTrue(env, "ACME_TERM") {
				return runby.Terminal{}, false
			}
			return runby.Terminal{Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_TERM")}}, true
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
			return runby.Remote{Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_VPN")}}, true
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
	layer, ok := result.RemoteLayer("acme-vpn")
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
		runby.WithOnlyDrivers(runby.AgentDriver{
			Agent: "acme",
			Detect: func(env runby.Env) (runby.Detection, bool) {
				return runby.Detection{Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME")}}, true
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
	if result.HasLayer(runby.AgentCodex) {
		t.Error("codex was detected after its driver was filtered out")
	}
	if !result.HasLayer(runby.AgentClaudeCode) {
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
	want := len(runby.Agents()) + len(runby.CIProviders()) + len(runby.TerminalPrograms()) +
		len(runby.RemotePlatforms()) + len(runby.RunnerTools())
	if got := len(runby.BuiltinDrivers()); got != want {
		t.Fatalf("BuiltinDrivers has %d drivers, the identity lists name %d products", got, want)
	}
}
