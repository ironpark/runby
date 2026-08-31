package runby

import "testing"

// The built-in driver tables are unexported, so the tests that hold them to
// their contracts live inside the package. They used to reach the tables
// through exported accessors, which existed only to be filtered and handed
// back to a per-axis WithOnly option; WithOnlyDrivers took that job, and the
// accessors went with it.

// TestDriverTablesMatchTheirIdentityLists holds the tables to the exported
// lists derived from them: a table is the single place a product is
// registered, so the two cannot disagree about what exists or in what order.
func TestDriverTablesMatchTheirIdentityLists(t *testing.T) {
	agents := Agents()
	drivers := agentDrivers()
	if len(agents) != len(drivers) {
		t.Fatalf("Agents() has %d entries, the table has %d", len(agents), len(drivers))
	}
	for i, driver := range drivers {
		if driver.Agent != agents[i] {
			t.Errorf("driver %d is %q, Agents()[%d] is %q", i, driver.Agent, i, agents[i])
		}
		if driver.Kind != driver.Agent.Kind() {
			t.Errorf("%q: driver Kind %q, Agent.Kind() %q", driver.Agent, driver.Kind, driver.Agent.Kind())
		}
		if driver.Models != driver.Agent.Models() {
			t.Errorf("%q: driver Models %q, Agent.Models() %q", driver.Agent, driver.Models, driver.Agent.Models())
		}
		if driver.Detect == nil {
			t.Errorf("%q has no Detect", driver.Agent)
		}
	}

	for i, driver := range remoteDrivers() {
		if got := RemotePlatforms()[i]; driver.Platform != got {
			t.Errorf("driver %d is %q, RemotePlatforms()[%d] is %q", i, driver.Platform, i, got)
		}
		if driver.Kind != driver.Platform.Kind() {
			t.Errorf("%q: driver Kind %q, Platform.Kind() %q", driver.Platform, driver.Kind, driver.Platform.Kind())
		}
	}
	for i, driver := range runnerDrivers() {
		if got := RunnerTools()[i]; driver.Tool != got {
			t.Errorf("driver %d is %q, RunnerTools()[%d] is %q", i, driver.Tool, i, got)
		}
		if driver.Kind != driver.Tool.Kind() {
			t.Errorf("%q: driver Kind %q, Tool.Kind() %q", driver.Tool, driver.Kind, driver.Tool.Kind())
		}
	}
	for i, driver := range terminalDrivers() {
		if got := TerminalPrograms()[i]; driver.Program != got {
			t.Errorf("driver %d is %q, TerminalPrograms()[%d] is %q", i, driver.Program, i, got)
		}
	}
	for i, driver := range ciDrivers() {
		if got := CIProviders()[i]; driver.Provider != got {
			t.Errorf("driver %d is %q, CIProviders()[%d] is %q", i, driver.Provider, i, got)
		}
	}
}

// TestDriverTablesAreCopied keeps a caller inside this package from reordering
// or truncating a table that every other caller shares.
func TestDriverTablesAreCopied(t *testing.T) {
	drivers := agentDrivers()
	drivers[0] = AgentDriver{Agent: "tampered"}
	if again := agentDrivers(); again[0].Agent == "tampered" {
		t.Fatal("agentDrivers() returned the built-in table itself")
	}
}

// TestOnlyDriversCanDropABuiltin covers the capability the removed per-axis
// WithOnly options provided: running the built-in set minus one product. It is
// how a caller silences a driver rather than replacing it, which Register
// cannot express.
func TestOnlyDriversCanDropABuiltin(t *testing.T) {
	var withoutCodex []Driver
	for _, driver := range agentDrivers() {
		if driver.Agent != AgentCodex {
			withoutCodex = append(withoutCodex, driver)
		}
	}

	result := Detect(
		WithEnviron([]string{"CODEX_THREAD_ID=t-1", "CLAUDECODE=1"}),
		WithOnlyDrivers(withoutCodex...),
	)
	if result.HasLayer(AgentCodex) {
		t.Error("codex was detected after its driver was dropped")
	}
	if !result.HasLayer(AgentClaudeCode) {
		t.Error("dropping codex also silenced claude-code")
	}
	// Only agent drivers were passed, so every other axis is off.
	if result.IsCI() || result.HasTerminal() || result.IsRemote() || result.HasRunner() {
		t.Error("WithOnlyDrivers left an axis on that was given no drivers")
	}
}

// TestBuiltinAgentsHaveALadderPosition means sortByLadder never has to guess
// about a built-in, and that TestBuiltinTableIsAlreadyInLadderOrder is
// checking a real order rather than a table of unplaceable drivers.
func TestBuiltinAgentsHaveALadderPosition(t *testing.T) {
	for _, driver := range builtinAgentDrivers {
		if got := driver.Agent.Level(); got == LevelUnknown {
			t.Errorf("%s has no ladder position; declare Kind and Models", driver.Agent)
		}
	}
}
