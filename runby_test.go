package runby_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/ironpark/runby"
)

func TestDetectCurrentEnvironmentPaseoCodex(t *testing.T) {
	if os.Getenv("PASEO_AGENT_ID") == "" ||
		(os.Getenv("CODEX_THREAD_ID") == "" && os.Getenv("CODEX_SESSION_ID") == "") {
		t.Skip("requires Codex running inside Paseo")
	}

	result := runby.Detect()
	if got, want := result.Chain(), "paseo>codex"; got != want {
		t.Fatalf("Chain() = %q, want %q: %#v", got, want, result.Layers)
	}
	if !result.Found() {
		t.Fatalf("IsAgent() = false: %#v", result.Layers)
	}
}

func TestDetectUnknownStillInspectsTerminal(t *testing.T) {
	// The zero-detection path must not skip terminal inspection; Terminal is a
	// property of the process, not of a detected layer.
	result := runby.Detect(runby.WithOnlyAgentDrivers())

	if result.Found() {
		t.Fatalf("Found() = true, want false: %#v", result.Layers)
	}
	if !result.TTY.Inspected {
		t.Fatalf("Terminal = %#v, want inspected", result.Terminal)
	}
	if result.Agent() != runby.AgentUnknown {
		t.Fatalf("Agent() = %q, want %q", result.Agent(), runby.AgentUnknown)
	}
	if result.Chain() != "unknown" {
		t.Fatalf("Chain() = %q, want %q", result.Chain(), "unknown")
	}
	if _, ok := result.Primary(); ok {
		t.Fatal("Primary() ok = true, want false")
	}
}

func TestInspectTTY(t *testing.T) {
	terminal := runby.InspectTTY()
	if !terminal.Inspected {
		t.Fatal("Inspected = false, want true")
	}
	if terminal.Attached != (terminal.StdinTTY || terminal.StdoutTTY || terminal.StderrTTY) {
		t.Fatalf("Attached is inconsistent: %#v", terminal)
	}
	if terminal.Interactive != (terminal.StdinTTY && (terminal.StdoutTTY || terminal.StderrTTY)) {
		t.Fatalf("Interactive is inconsistent: %#v", terminal)
	}
	if got := runby.Detect().TTY; got != terminal {
		t.Fatalf("Detect().Terminal = %#v, want %#v", got, terminal)
	}
}

func TestTTYOptions(t *testing.T) {
	if got := runby.Detect(runby.WithEnviron(nil)).TTY; got.Inspected {
		t.Fatalf("WithEnviron tty = %#v, want uninspected", got)
	}
	if got := runby.Detect(runby.WithoutTTY()).TTY; got.Inspected {
		t.Fatalf("WithoutTTY tty = %#v, want uninspected", got)
	}

	want := runby.TTY{Inspected: true, StdinTTY: true, StdoutTTY: true, Attached: true, Interactive: true}
	got := runby.Detect(runby.WithEnviron(nil), runby.WithTTY(want)).TTY
	if got != want {
		t.Fatalf("WithTTY tty = %#v, want %#v", got, want)
	}
}

func TestDetectUnknownEnviron(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"PATH=/usr/bin",
		"ANTHROPIC_API_KEY=not-an-execution-marker",
		"GEMINI_API_KEY=not-an-execution-marker",
		"COPILOT_MODEL=gpt-5",
	}))

	if result.Found() || result.Found() {
		t.Fatalf("Found()/IsAgent() = true, want false: %#v", result.Layers)
	}
}

func TestDetectLayers(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"PASEO_AGENT_ID=reviewer",
		"PASEO_AGENT_CWD=/work/project",
		"CODEX_THREAD_ID=thread-123",
		"CODEX_SANDBOX=workspace-write",
		"CODEX_SANDBOX_NETWORK_DISABLED=true",
		"CODEX_CI=1",
	}))

	if len(result.Layers) != 2 {
		t.Fatalf("len(Layers) = %d, want 2: %#v", len(result.Layers), result.Layers)
	}
	if result.Chain() != "paseo>codex" {
		t.Fatalf("Chain() = %q", result.Chain())
	}

	// The orchestrator is primary because it names the logical agent.
	paseo, ok := result.Get(runby.AgentPaseo)
	if !ok || result.Agent() != runby.AgentPaseo {
		t.Fatalf("primary = %q, want %q", result.Agent(), runby.AgentPaseo)
	}
	if paseo.AgentID != "reviewer" || paseo.Paths.WorkingDirectory != "/work/project" {
		t.Fatalf("Paseo detection = %#v", paseo)
	}
	if paseo.Kind != runby.KindOrchestrator || paseo.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Paseo classification = %#v", paseo)
	}

	codex, ok := result.Get(runby.AgentCodex)
	if !ok {
		t.Fatal("Get(AgentCodex) = false")
	}
	if codex.SessionID != "thread-123" || codex.Kind != runby.KindHarness {
		t.Fatalf("Codex detection = %#v", codex)
	}
	if codex.Sandbox.Mode != "workspace-write" || codex.Sandbox.Network != runby.NetworkDisabled {
		t.Fatalf("Codex sandbox = %#v", codex.Sandbox)
	}
	if codex.Extra["codex.ci"] != "true" {
		t.Fatalf("Extra = %#v", codex.Extra)
	}
	if !result.Has(runby.AgentCodex) || result.Has(runby.AgentAmp) {
		t.Fatalf("Has is inconsistent: %#v", result.Layers)
	}
}

func TestCodexSessionIDFallsBackToSessionVariable(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"CODEX_SESSION_ID=session-9"}))
	codex, ok := result.Get(runby.AgentCodex)
	if !ok || codex.SessionID != "session-9" {
		t.Fatalf("Codex detection = %#v", codex)
	}
}

func TestCodexSandboxOnlyIsProbable(t *testing.T) {
	// The sandbox variables describe configuration, not a specific run, so
	// they are a supporting signal rather than an execution marker.
	result := runby.Detect(runby.WithEnviron([]string{"CODEX_SANDBOX=read-only"}))
	codex, ok := result.Get(runby.AgentCodex)
	if !ok || codex.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Codex detection = %#v", codex)
	}
	if codex.Sandbox.Network != runby.NetworkUnknown {
		t.Fatalf("Network = %q, want %q", codex.Sandbox.Network, runby.NetworkUnknown)
	}
}

func TestClaudeCodeDetection(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"CLAUDECODE=1",
		"CLAUDE_CODE_SESSION_ID=session-abc",
		"CLAUDE_CODE_ENTRYPOINT=cli",
		"CLAUDE_CODE_CHILD_SESSION=true",
	}))

	claude, ok := result.Get(runby.AgentClaudeCode)
	if !ok {
		t.Fatal("Get(AgentClaudeCode) = false")
	}
	if claude.SessionID != "session-abc" || claude.Entrypoint != "cli" || !claude.Nested {
		t.Fatalf("Claude Code detection = %#v", claude)
	}
	want := []string{"CLAUDECODE", "CLAUDE_CODE_CHILD_SESSION", "CLAUDE_CODE_ENTRYPOINT", "CLAUDE_CODE_SESSION_ID"}
	if !reflect.DeepEqual(claude.Evidence, want) {
		t.Fatalf("Evidence = %#v, want %#v", claude.Evidence, want)
	}
}

func TestClaudeCodeAIAgentIsEvidenceOnlyWhenItNamesClaudeCode(t *testing.T) {
	if got := runby.Detect(runby.WithEnviron([]string{"AI_AGENT=claude-code/2.0"})); !got.Has(runby.AgentClaudeCode) {
		t.Fatalf("AI_AGENT=claude-code not detected: %#v", got.Layers)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"AI_AGENT=some-other-agent"})); got.Found() {
		t.Fatalf("AI_AGENT=some-other-agent detected: %#v", got.Layers)
	}
}

func TestZedIsATerminalNotAnAgent(t *testing.T) {
	// A Zed-owned terminal does not prove that Zed Agent, rather than a
	// person, ran the command, so Zed is reported on the Terminal axis only.
	result := runby.Detect(runby.WithEnviron([]string{
		"ZED_TERM=true",
		"TERM_PROGRAM=zed",
		"TERM_PROGRAM_VERSION=0.100.0",
	}))

	if result.Found() {
		t.Fatalf("Found() = true, want false: %#v", result.Layers)
	}
	if !result.IsTerminal() || result.Terminal.Program != runby.TerminalZed {
		t.Fatalf("Terminal = %#v", result.Terminal)
	}
	if result.Terminal.Version != "0.100.0" {
		t.Fatalf("Version = %q", result.Terminal.Version)
	}
}

func TestZedRequiresBothVariables(t *testing.T) {
	for _, environ := range [][]string{
		{"ZED_TERM=true"},
		{"ZED_TERM=false", "TERM_PROGRAM=zed"},
	} {
		got := runby.Detect(runby.WithEnviron(environ))
		if got.Terminal.Program == runby.TerminalZed {
			t.Fatalf("Detect(%v) = %#v, want no Zed", environ, got.Terminal)
		}
	}
}

func TestAmpEntrypoints(t *testing.T) {
	orb := runby.Detect(runby.WithEnviron([]string{"AMP_ORB=1"}))
	amp, ok := orb.Get(runby.AgentAmp)
	if !ok || amp.Entrypoint != "orb" || amp.SessionID != "" {
		t.Fatalf("Amp orb detection = %#v", amp)
	}

	service := runby.Detect(runby.WithEnviron([]string{"AMP_THREAD_ID=T-1"}))
	amp, ok = service.Get(runby.AgentAmp)
	if !ok || amp.Entrypoint != "orb-service" || amp.SessionID != "T-1" {
		t.Fatalf("Amp service detection = %#v", amp)
	}
}

func TestRemainingDetectors(t *testing.T) {
	cursor := runby.Detect(runby.WithEnviron([]string{"CURSOR_AGENT=1"}))
	if cursor.Agent() != runby.AgentCursor || !cursor.Found() {
		t.Fatalf("Cursor detection = %#v", cursor.Layers)
	}

	opencode := runby.Detect(runby.WithEnviron([]string{"OPENCODE_CLIENT=ACP"}))
	layer, ok := opencode.Get(runby.AgentOpenCode)
	if !ok || layer.Entrypoint != "acp" || layer.Confidence != runby.ConfidenceProbable {
		t.Fatalf("OpenCode detection = %#v", layer)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"OPENCODE_CLIENT=cli"})); got.Found() {
		t.Fatalf("OPENCODE_CLIENT=cli detected: %#v", got.Layers)
	}

	antigravity := runby.Detect(runby.WithEnviron([]string{"ANTIGRAVITY_EXECUTABLE_DATA_DIR=/data/ag"}))
	layer, ok = antigravity.Get(runby.AgentAntigravity2)
	if !ok || layer.Paths.DataDirectory != "/data/ag" || layer.Entrypoint != "sidecar" {
		t.Fatalf("Antigravity detection = %#v", layer)
	}
}

func TestEmptyValueIsNotEvidence(t *testing.T) {
	if got := runby.Detect(runby.WithEnviron([]string{"PASEO_AGENT_ID=   "})); got.Found() {
		t.Fatalf("blank PASEO_AGENT_ID detected: %#v", got.Layers)
	}
}

func TestWithLookupAndLastValueWins(t *testing.T) {
	values := map[string]string{"CURSOR_AGENT": "1"}
	got := runby.Detect(runby.WithLookup(func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}))
	if got.Agent() != runby.AgentCursor {
		t.Fatalf("Agent() = %q, want %q", got.Agent(), runby.AgentCursor)
	}

	dup := runby.Detect(runby.WithEnviron([]string{"PASEO_AGENT_ID=first", "PASEO_AGENT_ID=second"}))
	paseo, _ := dup.Get(runby.AgentPaseo)
	if paseo.AgentID != "second" {
		t.Fatalf("AgentID = %q, want %q", paseo.AgentID, "second")
	}
}

func TestWithAgentDriversTakePrecedence(t *testing.T) {
	const inHouse runby.Agent = "acme-orchestrator"
	driver := runby.AgentDriver{
		Agent: inHouse,
		Kind:  runby.KindOrchestrator,
		Detect: func(env runby.Env) (runby.Detection, bool) {
			id, ok := runby.Value(env, "ACME_RUN_ID")
			if !ok {
				return runby.Detection{}, false
			}
			return runby.Detection{AgentID: id, Evidence: runby.PresentNames(env, "ACME_RUN_ID")}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=run-7", "CODEX_THREAD_ID=t-1"}),
		runby.WithAgentDrivers(driver),
	)

	if got, want := result.Chain(), "acme-orchestrator>codex"; got != want {
		t.Fatalf("Chain() = %q, want %q", got, want)
	}
	// Agent and Kind come from the driver, and Confidence defaults, so Detect
	// need not repeat any of them.
	primary, _ := result.Primary()
	if primary.Agent != inHouse || primary.Kind != runby.KindOrchestrator ||
		primary.AgentID != "run-7" || primary.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("primary = %#v", primary)
	}
	if !result.Found() {
		t.Fatal("IsAgent() = false, want true")
	}
}

func TestAgentsAndKinds(t *testing.T) {
	agents := runby.Agents()
	if len(agents) == 0 || agents[0] != runby.AgentPaseo {
		t.Fatalf("Agents() = %#v", agents)
	}
	for _, agent := range agents {
		if agent.Kind() == runby.KindUnknown {
			t.Fatalf("%q has no Kind", agent)
		}
	}
	if runby.AgentUnknown.Kind() != runby.KindUnknown {
		t.Fatalf("AgentUnknown.Kind() = %q", runby.AgentUnknown.Kind())
	}
	if runby.Agent("").String() != "unknown" {
		t.Fatalf(`Agent("").String() = %q`, runby.Agent("").String())
	}
}

func TestCurrentIsCached(t *testing.T) {
	if !reflect.DeepEqual(runby.Current(), runby.Current()) {
		t.Fatal("Current() is not stable")
	}
	if runby.IsAgent() != runby.Current().Found() {
		t.Fatal("IsAgent() disagrees with Current()")
	}
}

// assertJSONRoundTrip checks that a Result survives serialization unchanged,
// which is what lets consumers log it and read it back.
func assertJSONRoundTrip(t *testing.T, result runby.Result) {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var round runby.Result
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(round, result) {
		t.Fatalf("round trip = %#v, want %#v", round, result)
	}
}

func TestResultJSONShape(t *testing.T) {
	assertJSONRoundTrip(t, runby.Detect(
		runby.WithEnviron([]string{"CURSOR_AGENT=1"}),
		runby.WithTTY(runby.TTY{Inspected: true}),
	))
}
