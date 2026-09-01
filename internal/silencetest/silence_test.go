// Package silencetest holds the one test that must not share a test binary
// with the rest of the suite. The driver registry is process-wide, so a
// driver registered to silence a built-in stays silenced for every other test
// in the same binary — the hazard docs/en/guide/drivers.md warns about, and the
// reason WithOnlyDrivers exists.
package silencetest_test

import (
	"testing"

	"github.com/ironpark/runby"
)

// Silencing a built-in for the whole process is expressible with the two
// driver paths already documented: registering the same identity replaces the
// built-in, and a Detect that never matches makes the replacement report
// nothing. Filtering BuiltinDrivers does the same for one call, but it cannot
// reach Current, IsAgent, or the CLI, which take no options.
func TestRegisteringANeverMatchingDriverSilencesABuiltin(t *testing.T) {
	runby.Register(runby.AgentDriver{
		Agent:  runby.AgentClaudeCode,
		Kind:   runby.KindHarness,
		Models: runby.ModelsFirstParty,
		Detect: func(runby.Env) (runby.Agent, bool) { return runby.Agent{}, false },
	})

	result := runby.Detect(runby.WithEnviron([]string{"CLAUDECODE=1", "CODEX_THREAD_ID=t-1"}))
	if _, ok := result.Agent(runby.AgentClaudeCode); ok {
		t.Error("claude-code was reported after a never-matching driver replaced it")
	}
	if _, ok := result.Agent(runby.AgentCodex); !ok {
		t.Error("silencing claude-code also silenced codex")
	}
}
