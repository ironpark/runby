package runby_test

import (
	"fmt"

	"github.com/ironpark/runby"
)

// The common case: decide whether to drop interactive behavior, and record the
// full execution chain in one log field.
func Example() {
	result := runby.Detect(runby.WithEnviron([]string{
		"PASEO_AGENT_ID=reviewer",
		"CODEX_THREAD_ID=thread-123",
		"CODEX_SANDBOX_NETWORK_DISABLED=true",
	}))

	fmt.Println(result.IsAgent())
	fmt.Println(result.Chain())
	// Output:
	// true
	// paseo>codex
}

// A Zed-owned terminal proves which application owns the terminal, not that an
// agent rather than a person ran the command.
func ExampleResult_IsAgent() {
	result := runby.Detect(runby.WithEnviron([]string{
		"ZED_TERM=true",
		"TERM_PROGRAM=zed",
	}))

	fmt.Println(result.Found(), result.IsAgent())
	// Output: true false
}

func ExampleResult_Get() {
	result := runby.Detect(runby.WithEnviron([]string{
		"CODEX_THREAD_ID=thread-123",
		"CODEX_SANDBOX=workspace-write",
		"CODEX_SANDBOX_NETWORK_DISABLED=true",
	}))

	if codex, ok := result.Get(runby.AgentCodex); ok {
		fmt.Println(codex.SessionID, codex.Sandbox.Mode, codex.Sandbox.Network)
	}
	// Output: thread-123 workspace-write disabled
}

// Detectors for agents this package does not support are added ahead of the
// built-in ones.
func ExampleWithDetectors() {
	acme := runby.NewDetector("acme-orchestrator", func(env runby.Env) (runby.Detection, bool) {
		id, ok := runby.Value(env, "ACME_RUN_ID")
		if !ok {
			return runby.Detection{}, false
		}
		return runby.Detection{
			Kind:     runby.KindOrchestrator,
			AgentID:  id,
			Evidence: runby.PresentNames(env, "ACME_RUN_ID"),
		}, true
	})

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=run-7", "CLAUDECODE=1"}),
		runby.WithDetectors(acme),
	)

	fmt.Println(result.Chain())
	// Output: acme-orchestrator>claude-code
}
