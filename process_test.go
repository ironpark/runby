package runby_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ironpark/runby"
)

func TestProcessTreeIsInspectedForThisProcess(t *testing.T) {
	tree := runby.Detect().Process
	if !tree.Inspected {
		t.Fatal("Inspected = false, want true for the current process")
	}
	if !tree.Supported {
		t.Skipf("ancestor reading is unsupported on this platform")
	}
	if len(tree.Ancestors) == 0 {
		t.Skip("no ancestor was readable")
	}
	if got, want := tree.Ancestors[0].PID, os.Getppid(); got != want {
		t.Fatalf("Ancestors[0].PID = %d, want %d", got, want)
	}
	// Names are normalized so one match table serves every platform.
	for _, p := range tree.Ancestors {
		if p.Name != strings.ToLower(p.Name) {
			t.Fatalf("Name = %q is not normalized", p.Name)
		}
	}
}

func TestProcessTreeNotInspectedForExplicitEnvironment(t *testing.T) {
	// An explicit environment does not necessarily describe this process, so
	// the ancestor chain must not be mixed in, exactly as with TTY.
	tree := runby.Detect(runby.WithEnviron([]string{"CLAUDECODE=1"})).Process
	if tree.Inspected {
		t.Fatalf("Process = %#v, want uninspected", tree)
	}
	if len(tree.Ancestors) != 0 {
		t.Fatalf("Ancestors = %#v, want empty", tree.Ancestors)
	}
}

func TestWithoutProcessTree(t *testing.T) {
	if tree := runby.Detect(runby.WithoutProcessTree()).Process; tree.Inspected {
		t.Fatalf("Process = %#v, want uninspected", tree)
	}
}

func TestWithProcessTreeCorroboratesAgents(t *testing.T) {
	// The environment says Claude Code inside Paseo. An ancestor running each
	// agent's executable proves both are still alive, which no environment
	// variable can.
	tree := runby.ProcessTree{
		Inspected: true,
		Supported: true,
		Ancestors: []runby.Process{
			{PID: 100, PPID: 200, Name: "zsh"},
			{PID: 200, PPID: 300, Name: "claude", Agent: runby.AgentClaudeCode},
			{PID: 300, PPID: 1, Name: "paseo", Agent: runby.AgentPaseo},
		},
	}
	result := runby.Detect(
		runby.WithEnviron([]string{"CLAUDECODE=1", "PASEO_AGENT_ID=reviewer"}),
		runby.WithProcessTree(tree),
	)

	paseo, _ := result.Get(runby.AgentPaseo)
	if paseo.AncestorPID != 300 {
		t.Fatalf("Paseo AncestorPID = %d, want 300", paseo.AncestorPID)
	}
	claude, _ := result.Get(runby.AgentClaudeCode)
	if claude.AncestorPID != 200 {
		t.Fatalf("Claude Code AncestorPID = %d, want 200", claude.AncestorPID)
	}
}

func TestAncestorPIDZeroIsNotADenial(t *testing.T) {
	// An agent can launch a process it does not remain an ancestor of, and
	// the chain stops at the first process owned by another user. A zero
	// must therefore never suppress the detection itself.
	result := runby.Detect(
		runby.WithEnviron([]string{"CLAUDECODE=1", "CLAUDE_CODE_SESSION_ID=s-1"}),
		runby.WithProcessTree(runby.ProcessTree{
			Inspected: true,
			Supported: true,
			Ancestors: []runby.Process{{PID: 10, PPID: 1, Name: "sh"}},
		}),
	)

	claude, ok := result.Get(runby.AgentClaudeCode)
	if !ok || !result.Found() {
		t.Fatalf("detection was suppressed: %#v", result.Layers)
	}
	if claude.AncestorPID != 0 {
		t.Fatalf("AncestorPID = %d, want 0", claude.AncestorPID)
	}
	if claude.SessionID != "s-1" {
		t.Fatalf("SessionID = %q", claude.SessionID)
	}
}

func TestProcessTreeFind(t *testing.T) {
	tree := runby.ProcessTree{
		Inspected: true,
		Supported: true,
		Ancestors: []runby.Process{
			{PID: 1, Name: "tmux", Remote: runby.RemoteTmux},
			{PID: 2, Name: "kitty", Terminal: runby.TerminalKitty},
			{PID: 3, Name: "codex", Agent: runby.AgentCodex},
		},
	}

	if p, ok := tree.FindAgent(runby.AgentCodex); !ok || p.PID != 3 {
		t.Fatalf("FindAgent = %#v, %v", p, ok)
	}
	if _, ok := tree.FindAgent(runby.AgentAmp); ok {
		t.Fatal("FindAgent(AgentAmp) = true, want false")
	}
	p, ok := tree.Find(func(p runby.Process) bool { return p.Terminal == runby.TerminalKitty })
	if !ok || p.PID != 2 {
		t.Fatalf("Find = %#v, %v", p, ok)
	}
	// Find returns the nearest match, so ordering is part of the contract.
	if p, _ := tree.Find(func(p runby.Process) bool { return p.PID > 0 }); p.PID != 1 {
		t.Fatalf("Find returned PID %d, want the nearest ancestor", p.PID)
	}
}

func TestProcessTreeSurvivesJSON(t *testing.T) {
	result := runby.Detect(
		runby.WithEnviron([]string{"CODEX_THREAD_ID=t-1"}),
		runby.WithProcessTree(runby.ProcessTree{
			Inspected: true, Supported: true,
			Ancestors: []runby.Process{{PID: 7, PPID: 1, Name: "codex", Path: "/usr/local/bin/codex", Agent: runby.AgentCodex}},
		}),
	)
	codex, _ := result.Get(runby.AgentCodex)
	if codex.AncestorPID != 7 {
		t.Fatalf("AncestorPID = %d, want 7", codex.AncestorPID)
	}
	assertJSONRoundTrip(t, result)
}

func TestCustomDetectorGetsAncestorCorroboration(t *testing.T) {
	// A detector supplied through WithDetectors can name its executables, and
	// then gets the same live-ancestor confirmation the built-in ones do.
	// Before the labels were derived from the configured detectors this was
	// impossible: the name table was closed.
	const acme runby.Agent = "acme-orchestrator"
	detector := acmeDetector{}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=run-7"}),
		runby.WithDetectors(detector),
		runby.WithProcessTree(runby.ProcessTree{
			Inspected: true,
			Supported: true,
			Ancestors: []runby.Process{
				{PID: 10, PPID: 20, Name: "sh"},
				{PID: 20, PPID: 1, Name: "acme-run"},
			},
		}),
	)

	layer, ok := result.Get(acme)
	if !ok {
		t.Fatalf("custom agent not detected: %#v", result.Layers)
	}
	if layer.AncestorPID != 20 {
		t.Fatalf("AncestorPID = %d, want 20", layer.AncestorPID)
	}
	// The injected tree carried no labels; Detect applied them.
	if got := result.Process.Ancestors[1].Agent; got != acme {
		t.Fatalf("ancestor label = %q, want %q", got, acme)
	}
}

// acmeDetector implements the optional ExecutableNamer interface.
type acmeDetector struct{}

func (acmeDetector) Agent() runby.Agent    { return "acme-orchestrator" }
func (acmeDetector) Executables() []string { return []string{"acme-run"} }
func (acmeDetector) Detect(env runby.Env) (runby.Detection, bool) {
	id, ok := runby.Value(env, "ACME_RUN_ID")
	if !ok {
		return runby.Detection{}, false
	}
	return runby.Detection{Kind: runby.KindOrchestrator, AgentID: id}, true
}

func TestTerminalAndRemoteAreCorroboratedToo(t *testing.T) {
	result := runby.Detect(
		runby.WithEnviron([]string{"KITTY_WINDOW_ID=3", "TMUX=/tmp/t,1,0"}),
		runby.WithProcessTree(runby.ProcessTree{
			Inspected: true,
			Supported: true,
			Ancestors: []runby.Process{
				{PID: 10, PPID: 20, Name: "kitty"},
				{PID: 20, PPID: 1, Name: "tmux"},
			},
		}),
	)

	if result.Terminal.AncestorPID != 10 {
		t.Fatalf("Terminal.AncestorPID = %d, want 10", result.Terminal.AncestorPID)
	}
	tmux, ok := result.GetRemote(runby.RemoteTmux)
	if !ok || tmux.AncestorPID != 20 {
		t.Fatalf("tmux layer = %#v, want AncestorPID 20", tmux)
	}
}

func TestLiveTerminalAncestorOutranksTheMultiplexerDowngrade(t *testing.T) {
	// A multiplexer server daemonizes and is reparented away from the terminal
	// that started it, so the terminal cannot appear in a pane's ancestor
	// chain. Measured with tmux 3.7c: inside a pane the chain ends at the tmux
	// server, never at the launching terminal. Finding the terminal alive
	// therefore means this process is not behind a stale pane.
	tree := runby.ProcessTree{
		Inspected: true,
		Supported: true,
		Ancestors: []runby.Process{{PID: 10, PPID: 1, Name: "kitty"}},
	}
	environ := []string{"KITTY_WINDOW_ID=3", "TMUX=/tmp/t,1,0"}

	withAncestor := runby.Detect(runby.WithEnviron(environ), runby.WithProcessTree(tree))
	if withAncestor.Terminal.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Confidence = %q, want definite when the terminal is a live ancestor",
			withAncestor.Terminal.Confidence)
	}

	// Without the ancestor the downgrade still applies, because the evidence
	// really could have been left by a terminal that has since closed.
	withoutAncestor := runby.Detect(runby.WithEnviron(environ))
	if withoutAncestor.Terminal.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q, want probable without a live ancestor",
			withoutAncestor.Terminal.Confidence)
	}
}
