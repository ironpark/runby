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
	if !result.IsAgent() {
		t.Fatalf("IsAgent() = false: %#v", result.Layers)
	}
}

func TestDetectUnknownStillInspectsTerminal(t *testing.T) {
	// The zero-detection path must not skip terminal inspection; Terminal is a
	// property of the process, not of a detected layer.
	result := runby.Detect(runby.WithOnlyDrivers())

	if result.IsAgent() {
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

	if result.IsAgent() || result.IsAgent() {
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
	paseo, ok := result.Layer(runby.AgentPaseo)
	if !ok || result.Agent() != runby.AgentPaseo {
		t.Fatalf("primary = %q, want %q", result.Agent(), runby.AgentPaseo)
	}
	if paseo.AgentID != "reviewer" || paseo.Paths.WorkingDirectory != "/work/project" {
		t.Fatalf("Paseo detection = %#v", paseo)
	}
	if paseo.Kind != runby.KindOrchestrator || paseo.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Paseo classification = %#v", paseo)
	}

	codex, ok := result.Layer(runby.AgentCodex)
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
	if !result.HasLayer(runby.AgentCodex) || result.HasLayer(runby.AgentAmp) {
		t.Fatalf("Has is inconsistent: %#v", result.Layers)
	}
}

func TestCodexSessionIDFallsBackToSessionVariable(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"CODEX_SESSION_ID=session-9"}))
	codex, ok := result.Layer(runby.AgentCodex)
	if !ok || codex.SessionID != "session-9" {
		t.Fatalf("Codex detection = %#v", codex)
	}
}

func TestCodexSandboxOnlyIsProbable(t *testing.T) {
	// The sandbox variables describe configuration, not a specific run, so
	// they are a supporting signal rather than an execution marker.
	result := runby.Detect(runby.WithEnviron([]string{"CODEX_SANDBOX=read-only"}))
	codex, ok := result.Layer(runby.AgentCodex)
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

	claude, ok := result.Layer(runby.AgentClaudeCode)
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
	if got := runby.Detect(runby.WithEnviron([]string{"AI_AGENT=claude-code/2.0"})); !got.HasLayer(runby.AgentClaudeCode) {
		t.Fatalf("AI_AGENT=claude-code not detected: %#v", got.Layers)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"AI_AGENT=some-other-agent"})); got.IsAgent() {
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

	if result.IsAgent() {
		t.Fatalf("Found() = true, want false: %#v", result.Layers)
	}
	if !result.HasTerminal() || result.Terminal.Program != runby.TerminalZed {
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
	amp, ok := orb.Layer(runby.AgentAmp)
	if !ok || amp.Entrypoint != "orb" || amp.SessionID != "" {
		t.Fatalf("Amp orb detection = %#v", amp)
	}

	service := runby.Detect(runby.WithEnviron([]string{"AMP_THREAD_ID=T-1"}))
	amp, ok = service.Layer(runby.AgentAmp)
	if !ok || amp.Entrypoint != "orb-service" || amp.SessionID != "T-1" {
		t.Fatalf("Amp service detection = %#v", amp)
	}
}

func TestRemainingDetectors(t *testing.T) {
	cursor := runby.Detect(runby.WithEnviron([]string{"CURSOR_AGENT=1"}))
	if cursor.Agent() != runby.AgentCursor || !cursor.IsAgent() {
		t.Fatalf("Cursor detection = %#v", cursor.Layers)
	}

	opencode := runby.Detect(runby.WithEnviron([]string{"OPENCODE_CLIENT=ACP"}))
	layer, ok := opencode.Layer(runby.AgentOpenCode)
	if !ok || layer.Entrypoint != "acp" || layer.Confidence != runby.ConfidenceProbable {
		t.Fatalf("OpenCode detection = %#v", layer)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"OPENCODE_CLIENT=cli"})); got.IsAgent() {
		t.Fatalf("OPENCODE_CLIENT=cli detected: %#v", got.Layers)
	}

	antigravity := runby.Detect(runby.WithEnviron([]string{"ANTIGRAVITY_EXECUTABLE_DATA_DIR=/data/ag"}))
	layer, ok = antigravity.Layer(runby.AgentAntigravity2)
	if !ok || layer.Paths.DataDirectory != "/data/ag" || layer.Entrypoint != "sidecar" {
		t.Fatalf("Antigravity detection = %#v", layer)
	}
}

func TestEmptyValueIsNotEvidence(t *testing.T) {
	if got := runby.Detect(runby.WithEnviron([]string{"PASEO_AGENT_ID=   "})); got.IsAgent() {
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
	paseo, _ := dup.Layer(runby.AgentPaseo)
	if paseo.AgentID != "second" {
		t.Fatalf("AgentID = %q, want %q", paseo.AgentID, "second")
	}
}

func TestCustomAgentDriverReplacesABuiltin(t *testing.T) {
	const inHouse runby.Agent = "acme-orchestrator"
	driver := runby.AgentDriver{
		Agent: inHouse,
		Kind:  runby.KindOrchestrator,
		Detect: func(env runby.Env) (runby.Detection, bool) {
			id, ok := runby.Value(env, "ACME_RUN_ID")
			if !ok {
				return runby.Detection{}, false
			}
			return runby.Detection{AgentID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_RUN_ID")}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_RUN_ID=run-7", "CODEX_THREAD_ID=t-1"}),
		runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), driver)...),
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
	if !result.IsAgent() {
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
	if runby.IsAgent() != runby.Current().IsAgent() {
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

// TestLevelDerivedForCustomDrivers checks that a driver supplied through
// WithOnlyDrivers is placed on the ladder by the same rule as a built-in one,
// and that an unclassified driver reports unknown rather than guessing.
func TestLevelDerivedForCustomDrivers(t *testing.T) {
	for _, test := range []struct {
		name   string
		kind   runby.Kind
		models runby.ModelSource
		want   runby.Level
	}{
		{"orchestrator delegating", runby.KindOrchestrator, runby.ModelsDelegated, runby.Level3},
		// The case a single ladder cannot hold: an orchestrator over its own
		// vendor's harness, as Antigravity 2.0 is.
		{"orchestrator first party", runby.KindOrchestrator, runby.ModelsFirstParty, runby.Level3},
		{"first party harness", runby.KindHarness, runby.ModelsFirstParty, runby.Level1},
		{"multi vendor harness", runby.KindHarness, runby.ModelsMultiVendor, runby.Level2},
		{"harness with no models declared", runby.KindHarness, "", runby.LevelUnknown},
		{"nothing declared", "", "", runby.LevelUnknown},
	} {
		t.Run(test.name, func(t *testing.T) {
			driver := runby.AgentDriver{
				Agent:  "acme",
				Kind:   test.kind,
				Models: test.models,
				Detect: func(env runby.Env) (runby.Detection, bool) {
					if _, ok := runby.Value(env, "ACME"); !ok {
						return runby.Detection{}, false
					}
					return runby.Detection{Axis: runby.Axis{Evidence: []string{"ACME"}}}, true
				},
			}
			result := runby.Detect(
				runby.WithEnviron([]string{"ACME=1"}),
				runby.WithOnlyDrivers(driver),
			)
			layer, ok := result.Primary()
			if !ok {
				t.Fatal("the custom driver did not match")
			}
			if layer.Level != test.want {
				t.Errorf("Level = %s, want %s", layer.Level, test.want)
			}
			// An undeclared axis must read as unknown rather than empty, so a
			// serialized Detection never carries a blank classification.
			if test.kind == "" && layer.Kind != runby.KindUnknown {
				t.Errorf("Kind = %q, want %s", layer.Kind, runby.KindUnknown)
			}
			if test.models == "" && layer.Models != runby.ModelsUnknown {
				t.Errorf("Models = %q, want %s", layer.Models, runby.ModelsUnknown)
			}
		})
	}
}

// The five agents added from the vercel/detect-agent survey. Each marker was
// verified against the product's own source or reference documentation before
// the driver was written; docs/research/agents/ records which and why.
func TestAgentsAddedFromTheDetectAgentSurvey(t *testing.T) {
	for _, test := range []struct {
		name       string
		environ    []string
		want       runby.Agent
		confidence runby.Confidence
		entrypoint string
	}{
		{"gemini cli", []string{"GEMINI_CLI=1"}, runby.AgentGeminiCLI, runby.ConfidenceDefinite, ""},
		{"auggie", []string{"AUGMENT_AGENT=1"}, runby.AgentAuggie, runby.ConfidenceDefinite, ""},
		{"grok build hook", []string{"GROK_PLUGIN_ROOT=/p"}, runby.AgentGrokBuild, runby.ConfidenceDefinite, "plugin-hook"},
		{"openclaw exec", []string{"OPENCLAW_SHELL=exec"}, runby.AgentOpenClaw, runby.ConfidenceDefinite, "exec"},
		{"openclaw tui", []string{"OPENCLAW_SHELL=tui-local"}, runby.AgentOpenClaw, runby.ConfidenceDefinite, "tui-local"},
		// Cline's marker rides on the terminal it created, and a human can
		// type into that terminal, so it never claims to be proof.
		{"cline", []string{"CLINE_ACTIVE=true"}, runby.AgentCline, runby.ConfidenceProbable, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := runby.Detect(runby.WithEnviron(test.environ))
			layer, ok := result.Layer(test.want)
			if !ok {
				t.Fatalf("%s not detected from %v", test.want, test.environ)
			}
			if layer.Confidence != test.confidence {
				t.Errorf("Confidence = %q, want %q", layer.Confidence, test.confidence)
			}
			if layer.Entrypoint != test.entrypoint {
				t.Errorf("Entrypoint = %q, want %q", layer.Entrypoint, test.entrypoint)
			}
		})
	}
}

// Variables that merely share a prefix with a marker are configuration the user
// sets, not evidence that the product launched anything. OpenClaw's timeout
// setting is the trap: a prefix match would report the agent for it.
func TestConfigurationLookalikesAreNotEvidence(t *testing.T) {
	for _, environ := range [][]string{
		{"OPENCLAW_SHELL_ENV_TIMEOUT_MS=15000"},
		{"GEMINI_CLI_HOME=/home/u/.gemini"},
		{"CLINE_DATA_DIR=/home/u/.cline"},
		{"AUGMENT_API_URL=https://example.invalid"},
		// GOOSE_PROVIDER selects which LLM to use; it is set by the user and
		// says nothing about what launched this process.
		{"GOOSE_PROVIDER=anthropic"},
	} {
		if result := runby.Detect(runby.WithEnviron(environ)); result.IsAgent() {
			t.Errorf("%v was reported as %s", environ, result.Chain())
		}
	}
}

// OpenCode sets OPENCODE on every invocation, before the subcommand runs, so
// an ordinary `opencode run` session is proof and not merely a hint. The lab
// found this by measuring a container; the driver changed only once the CLI
// source confirmed it.
func TestOpenCodeGeneralMarker(t *testing.T) {
	for _, test := range []struct {
		name       string
		environ    []string
		detected   bool
		confidence runby.Confidence
		entrypoint string
	}{
		{"run", []string{"OPENCODE=1", "OPENCODE_PID=17", "AGENT=1"}, true, runby.ConfidenceDefinite, ""},
		{"acp", []string{"OPENCODE=1", "OPENCODE_CLIENT=acp"}, true, runby.ConfidenceDefinite, "acp"},
		// The ACP marker on its own still counts, but only as a hint: an
		// embedding host sets it on the way in.
		{"acp without the launch marker", []string{"OPENCODE_CLIENT=acp"}, true, runby.ConfidenceProbable, "acp"},
		// OPENCODE_CLIENT is read with a default of "cli", so any other value
		// is a user-settable input rather than evidence.
		{"client name alone", []string{"OPENCODE_CLIENT=cli"}, false, "", ""},
		{"zed as host", []string{"OPENCODE_CLIENT=zed"}, false, "", ""},
		// AGENT is not vendor-specific; Goose sets it too.
		{"agent alone", []string{"AGENT=1"}, false, "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := runby.Detect(runby.WithEnviron(test.environ))
			layer, ok := result.Layer(runby.AgentOpenCode)
			if ok != test.detected {
				t.Fatalf("detected = %v, want %v (%#v)", ok, test.detected, result.Layers)
			}
			if !test.detected {
				return
			}
			if layer.Confidence != test.confidence {
				t.Errorf("Confidence = %q, want %q", layer.Confidence, test.confidence)
			}
			if layer.Entrypoint != test.entrypoint {
				t.Errorf("Entrypoint = %q, want %q", layer.Entrypoint, test.entrypoint)
			}
			// OPENCODE_PID and AGENT must never become evidence.
			for _, name := range layer.Evidence {
				if name == "AGENT" || name == "OPENCODE_PID" {
					t.Errorf("%s was used as evidence", name)
				}
			}
		})
	}
}
