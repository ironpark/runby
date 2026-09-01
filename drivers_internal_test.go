package runby

import "testing"

// The built-in driver tables are unexported, so the tests holding them to
// their contracts live inside the package. Callers outside reach the same
// drivers through BuiltinDrivers, which flattens all five tables into the one
// type WithOnlyDrivers takes.

// TestDriverTablesMatchTheirIdentityLists holds the tables to the exported
// lists derived from them: a table is the single place a product is
// registered, so the two cannot disagree about what exists or in what order.
func TestDriverTablesMatchTheirIdentityLists(t *testing.T) {
	agents := Agents()
	drivers := builtinAgentDrivers
	if len(agents) != len(drivers) {
		t.Fatalf("Agents() has %d entries, the table has %d", len(agents), len(drivers))
	}
	for i, driver := range drivers {
		if driver.Agent != agents[i] {
			t.Errorf("driver %d is %q, Agents()[%d] is %q", i, driver.Agent, i, agents[i])
		}
		if driver.Detect == nil {
			t.Errorf("%q has no Detect", driver.Agent)
		}
	}

	for i, driver := range builtinRemoteDrivers {
		if got := RemotePlatforms()[i]; driver.Platform != got {
			t.Errorf("driver %d is %q, RemotePlatforms()[%d] is %q", i, driver.Platform, i, got)
		}
	}
	for i, driver := range builtinRunnerDrivers {
		if got := RunnerTools()[i]; driver.Tool != got {
			t.Errorf("driver %d is %q, RunnerTools()[%d] is %q", i, driver.Tool, i, got)
		}
	}
	for i, driver := range builtinTerminalDrivers {
		if got := TerminalPrograms()[i]; driver.Program != got {
			t.Errorf("driver %d is %q, TerminalPrograms()[%d] is %q", i, driver.Program, i, got)
		}
	}
	for i, driver := range builtinCIDrivers {
		if got := CIProviders()[i]; driver.Provider != got {
			t.Errorf("driver %d is %q, CIProviders()[%d] is %q", i, driver.Provider, i, got)
		}
	}
}

// TestBuiltinAgentsHaveALadderPosition means sortByLadder never has to guess
// about a built-in, and that TestBuiltinTableIsAlreadyInLadderOrder is
// checking a real order rather than a table of unplaceable drivers.
func TestBuiltinAgentsHaveALadderPosition(t *testing.T) {
	for _, driver := range builtinAgentDrivers {
		if ladderRank(driver) == 3 {
			t.Errorf("%s has no ladder position; declare Kind and Models", driver.Agent)
		}
	}
}
