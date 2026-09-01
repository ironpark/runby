package runby_test

import (
	"testing"

	"github.com/ironpark/runby"
)

func TestOrcaLayersWithTheAgentItLaunched(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"ORCA_PANE_KEY=w1:t2:p3",
		"ORCA_TAB_ID=1e9d2c2a-0f1b-4a4e-9b3d-7c5f2a1b8e40",
		"ORCA_WORKTREE_ID=wt-42",
		"ORCA_WORKTREE_PATH=/work/trees/wt-42",
		"ORCA_ROOT_PATH=/work/project",
		"ORCA_USER_DATA_PATH=/home/dev/.orca",
		"ORCA_CODEX_HOME=/home/dev/.orca/codex",
		"CODEX_THREAD_ID=thread-9",
	}))

	if result.Chain() != "orca>codex" {
		t.Fatalf("Chain() = %q, want the orchestrator outside the harness", result.Chain())
	}

	orca, ok := result.Agent(runby.AgentOrca)
	if !ok || primaryAgent(result) != runby.AgentOrca {
		t.Fatalf("primary = %q, want %q", primaryAgent(result), runby.AgentOrca)
	}
	if orca.Kind != runby.KindOrchestrator {
		t.Fatalf("Kind = %q, want orchestrator", orca.Kind)
	}
	// Orca stamps the pane, not the command, so it can never be definite: a
	// person typing in an Orca terminal produces the same variables.
	if orca.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q, want probable", orca.Confidence)
	}
	if orca.SessionID != "w1:t2:p3" {
		t.Fatalf("SessionID = %q, want the pane key", orca.SessionID)
	}
	// The worktree is where the command actually runs; the root is the repo it
	// was branched from.
	if orca.Paths.WorkingDirectory != "/work/trees/wt-42" {
		t.Fatalf("WorkingDirectory = %q, want the worktree", orca.Paths.WorkingDirectory)
	}
	if orca.Paths.DataDirectory != "/home/dev/.orca" {
		t.Fatalf("DataDirectory = %q", orca.Paths.DataDirectory)
	}
	if orca.Extra["orca.worktree_id"] != "wt-42" || orca.Extra["orca.root_path"] != "/work/project" {
		t.Fatalf("Extra = %#v", orca.Extra)
	}
	// ORCA_CODEX_HOME corroborates an Orca-managed Codex but must never be a
	// detection on its own.
	if orca.Extra["orca.codex_home"] != "/home/dev/.orca/codex" {
		t.Fatalf("Extra = %#v", orca.Extra)
	}

	// The harness underneath is detected independently, at its own confidence.
	codex, ok := result.Agent(runby.AgentCodex)
	if !ok || codex.SessionID != "thread-9" || codex.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Codex layer = %#v", codex)
	}

	assertJSONRoundTrip(t, result)
}

func TestOrcaFallsBackToTerminalHandleAndRootPath(t *testing.T) {
	// A pane outside a worktree: no pane key, no worktree path.
	result := runby.Detect(runby.WithEnviron([]string{
		"ORCA_TERMINAL_HANDLE=term_17",
		"ORCA_TAB_ID=tab-1",
		"ORCA_ROOT_PATH=/work/project",
	}))

	orca, ok := result.Agent(runby.AgentOrca)
	if !ok {
		t.Fatal("Get(AgentOrca) = false")
	}
	if orca.SessionID != "term_17" {
		t.Fatalf("SessionID = %q, want the terminal handle", orca.SessionID)
	}
	if orca.Paths.WorkingDirectory != "/work/project" {
		t.Fatalf("WorkingDirectory = %q, want the root path", orca.Paths.WorkingDirectory)
	}
	if _, ok := orca.Extra["orca.worktree_id"]; ok {
		t.Fatalf("Extra carries a key for an unset variable: %#v", orca.Extra)
	}
}

func TestOrcaRequiresAnOwnerAndALocationMarker(t *testing.T) {
	// Orca sets an owner marker and a location marker together. Either alone
	// is not enough, so a lone stray ORCA_ variable cannot fake a detection.
	for _, environ := range [][]string{
		{"ORCA_PANE_KEY=w1:t2:p3"},
		{"ORCA_TERMINAL_HANDLE=term_17"},
		{"ORCA_TAB_ID=tab-1"},
		{"ORCA_WORKTREE_ID=wt-42"},
		{"ORCA_WORKTREE_PATH=/work/trees/wt-42", "ORCA_WORKSPACE_NAME=demo"},
		{"ORCA_USER_DATA_PATH=/home/dev/.orca", "ORCA_CODEX_HOME=/home/dev/.orca/codex"},
	} {
		if _, ok := runby.Detect(runby.WithEnviron(environ)).Agent(runby.AgentOrca); ok {
			t.Errorf("environ %v detected Orca, want no detection", environ)
		}
	}
}

func TestOrcaEvidenceNamesTheVariablesItRead(t *testing.T) {
	orca, ok := runby.Detect(runby.WithEnviron([]string{
		"ORCA_PANE_KEY=w1:t2:p3",
		"ORCA_WORKTREE_ID=wt-42",
		"ORCA_ORCHESTRATION_COMPATIBILITY_HOST_KIND=wsl",
		"ORCA_ORCHESTRATION_COMPATIBILITY_HOST_INCARNATION=Ubuntu-24.04",
	})).Agent(runby.AgentOrca)
	if !ok {
		t.Fatal("Get(AgentOrca) = false")
	}

	want := []string{
		"ORCA_ORCHESTRATION_COMPATIBILITY_HOST_INCARNATION",
		"ORCA_ORCHESTRATION_COMPATIBILITY_HOST_KIND",
		"ORCA_PANE_KEY",
		"ORCA_WORKTREE_ID",
	}
	if len(orca.Evidence) != len(want) {
		t.Fatalf("Evidence = %#v, want %#v", orca.Evidence, want)
	}
	for i, name := range want {
		if orca.Evidence[i] != name {
			t.Fatalf("Evidence = %#v, want %#v", orca.Evidence, want)
		}
	}
	// Only WSL sets the host variables, so their presence is context, not a
	// stronger claim about who ran the command.
	if orca.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q", orca.Confidence)
	}
	if orca.Extra["orca.host_kind"] != "wsl" {
		t.Fatalf("Extra = %#v", orca.Extra)
	}
}
