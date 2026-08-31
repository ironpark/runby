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

	fmt.Println(result.Found())
	fmt.Println(result.Chain())
	// Output:
	// true
	// paseo>codex
}

// A Zed-owned terminal proves which application owns the terminal, not that an
// agent rather than a person ran the command, so it lands on the Terminal axis.
func ExampleResult_IsTerminal() {
	result := runby.Detect(runby.WithEnviron([]string{
		"ZED_TERM=true",
		"TERM_PROGRAM=zed",
	}))

	fmt.Println(result.Found(), result.IsTerminal(), result.Terminal.Program)
	// Output: false true zed
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

// CI is a separate axis from the agent layers. An agent invoked from a
// workflow populates both at once.
func ExampleResult_IsCI() {
	result := runby.Detect(runby.WithEnviron([]string{
		"CLAUDECODE=1",
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=1658821493",
		"GITHUB_RUN_ATTEMPT=2",
		"GITHUB_EVENT_NAME=push",
	}))

	fmt.Println(result.Chain(), result.IsCI(), result.CI.Provider)
	fmt.Println(result.CI.PipelineID, result.CI.Attempt, result.CI.Trigger)
	// Output:
	// claude-code true github-actions
	// 1658821493 2 push
}

// Terminal identifies the emulator that produced the environment, never the
// one currently attached. Inside a multiplexer that gap is explicit.
func ExampleTerminal() {
	result := runby.Detect(runby.WithEnviron([]string{
		"TERM_PROGRAM=ghostty",
		"TERM=xterm-ghostty",
		"TMUX=/tmp/tmux-501/default,123,0",
	}))

	fmt.Println(result.Terminal.Program, result.Terminal.Multiplexer)
	// Confidence drops because the tmux server keeps the environment of
	// whichever client started it.
	fmt.Println(result.Terminal.Confidence)
	// Output:
	// ghostty tmux
	// probable
}
