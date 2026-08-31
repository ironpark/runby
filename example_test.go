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
// agent rather than a person ran the command, so it lands on the Terminal axis.
func ExampleResult_HasTerminal() {
	result := runby.Detect(runby.WithEnviron([]string{
		"ZED_TERM=true",
		"TERM_PROGRAM=zed",
	}))

	fmt.Println(result.IsAgent(), result.HasTerminal(), result.Terminal.Program)
	// Output: false true zed
}

func ExampleResult_Layer() {
	result := runby.Detect(runby.WithEnviron([]string{
		"CODEX_THREAD_ID=thread-123",
		"CODEX_SANDBOX=workspace-write",
		"CODEX_SANDBOX_NETWORK_DISABLED=true",
	}))

	if codex, ok := result.Layer(runby.AgentCodex); ok {
		fmt.Println(codex.SessionID, codex.Sandbox.Mode, codex.Sandbox.Network)
	}
	// Output: thread-123 workspace-write disabled
}

// A driver for an agent this package does not support runs beside the
// built-in ones for a single call. It carries the agent's identity, what a
// detection of it proves, and the binaries it runs as. Register the same value
// instead to have every caller in the program see it.
func ExampleWithOnlyDrivers() {
	acme := runby.AgentDriver{
		Agent:       "acme-orchestrator",
		Kind:        runby.KindOrchestrator,
		Executables: []string{"acme-run"},
		Detect: func(env runby.Env) (runby.Detection, bool) {
			id, ok := runby.Value(env, "ACME_RUN_ID")
			if !ok {
				return runby.Detection{}, false
			}
			return runby.Detection{AgentID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_RUN_ID")}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=run-7", "CLAUDECODE=1"}),
		runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), acme)...),
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

	mux, _ := result.Multiplexer()
	fmt.Println(result.Terminal.Program, mux.Platform)
	// Confidence drops because the tmux server keeps the environment of
	// whichever client started it and cannot refresh a running pane.
	fmt.Println(result.Terminal.Confidence)
	// Output:
	// ghostty tmux
	// probable
}

// Remote layers coexist. An SSH session into a Codespace running tmux is
// three concurrent facts, and only the multiplexer weakens the terminal.
func ExampleResult_Remote() {
	result := runby.Detect(runby.WithEnviron([]string{
		"SSH_CONNECTION=203.0.113.5 54321 198.51.100.9 22",
		"CODESPACES=true",
		"TMUX=/tmp/tmux-1000/default,9,0",
		"KITTY_WINDOW_ID=4",
	}))

	for _, layer := range result.Remote {
		fmt.Println(layer.Platform, layer.Kind)
	}
	fmt.Println(result.Terminal.Program, result.Terminal.Confidence)
	// Output:
	// tmux multiplexer
	// openssh environment
	// github-codespaces environment
	// kitty probable
}

// The process tree is the one axis that cannot be forged by an export and
// that proves the agent is still running.
func ExampleProcessTree() {
	result := runby.Detect(
		runby.WithEnviron([]string{"CLAUDECODE=1", "PASEO_AGENT_ID=reviewer"}),
		runby.WithProcessTree(runby.ProcessTree{
			Inspected: true,
			Supported: true,
			Ancestors: []runby.Process{
				{PID: 100, PPID: 200, Name: "zsh"},
				{PID: 200, PPID: 300, Name: "claude", Agent: runby.AgentClaudeCode},
				{PID: 300, PPID: 1, Name: "paseo", Agent: runby.AgentPaseo},
			},
		}),
	)

	for _, layer := range result.Layers {
		fmt.Println(layer.Agent, layer.AncestorPID)
	}
	// Output:
	// paseo 300
	// claude-code 200
}
