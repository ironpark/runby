package runby_test

import (
	"encoding/json"
	"fmt"
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
		t.Fatalf("Chain() = %q, want %q: %#v", got, want, result.Agents)
	}
	if !result.IsAgent() {
		t.Fatalf("IsAgent() = false: %#v", result.Agents)
	}
}

func TestDetectUnknownStillInspectsTerminal(t *testing.T) {
	// The zero-detection path must not skip terminal inspection; Terminal is a
	// property of the process, not of a detected layer.
	result := runby.Detect(runby.WithOnlyDrivers())

	if result.IsAgent() {
		t.Fatalf("Found() = true, want false: %#v", result.Agents)
	}
	if !result.TTY.Inspected {
		t.Fatalf("Terminal = %#v, want inspected", result.Terminal)
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
		t.Fatalf("Found()/IsAgent() = true, want false: %#v", result.Agents)
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

	if len(result.Agents) != 2 {
		t.Fatalf("len(Layers) = %d, want 2: %#v", len(result.Agents), result.Agents)
	}
	if result.Chain() != "paseo>codex" {
		t.Fatalf("Chain() = %q", result.Chain())
	}

	// The orchestrator is primary because it names the logical agent.
	paseo, ok := result.Agent(runby.AgentPaseo)
	if !ok || primaryAgent(result) != runby.AgentPaseo {
		t.Fatalf("primary = %q, want %q", primaryAgent(result), runby.AgentPaseo)
	}
	if paseo.AgentID != "reviewer" || paseo.Paths.WorkingDirectory != "/work/project" {
		t.Fatalf("Paseo detection = %#v", paseo)
	}
	if paseo.Kind != runby.KindOrchestrator || paseo.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Paseo classification = %#v", paseo)
	}

	codex, ok := result.Agent(runby.AgentCodex)
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
	_, hasCodex := result.Agent(runby.AgentCodex)
	_, hasAmp := result.Agent(runby.AgentAmp)
	if !hasCodex || hasAmp {
		t.Fatalf("Layer lookup is inconsistent: %#v", result.Agents)
	}
}

func TestCodexSessionIDFallsBackToSessionVariable(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"CODEX_SESSION_ID=session-9"}))
	codex, ok := result.Agent(runby.AgentCodex)
	if !ok || codex.SessionID != "session-9" {
		t.Fatalf("Codex detection = %#v", codex)
	}
}

func TestCodexSandboxOnlyIsProbable(t *testing.T) {
	// The sandbox variables describe configuration, not a specific run, so
	// they are a supporting signal rather than an execution marker.
	result := runby.Detect(runby.WithEnviron([]string{"CODEX_SANDBOX=read-only"}))
	codex, ok := result.Agent(runby.AgentCodex)
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

	claude, ok := result.Agent(runby.AgentClaudeCode)
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
	got := runby.Detect(runby.WithEnviron([]string{"AI_AGENT=claude-code/2.0"}))
	if _, ok := got.Agent(runby.AgentClaudeCode); !ok {
		t.Fatalf("AI_AGENT=claude-code not detected: %#v", got.Agents)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"AI_AGENT=some-other-agent"})); got.IsAgent() {
		t.Fatalf("AI_AGENT=some-other-agent detected: %#v", got.Agents)
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
		t.Fatalf("Found() = true, want false: %#v", result.Agents)
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
	amp, ok := orb.Agent(runby.AgentAmp)
	if !ok || amp.Entrypoint != "orb" || amp.SessionID != "" {
		t.Fatalf("Amp orb detection = %#v", amp)
	}

	service := runby.Detect(runby.WithEnviron([]string{"AMP_THREAD_ID=T-1"}))
	amp, ok = service.Agent(runby.AgentAmp)
	if !ok || amp.Entrypoint != "orb-service" || amp.SessionID != "T-1" {
		t.Fatalf("Amp service detection = %#v", amp)
	}
}

func TestRemainingDetectors(t *testing.T) {
	cursor := runby.Detect(runby.WithEnviron([]string{"CURSOR_AGENT=1"}))
	if primaryAgent(cursor) != runby.AgentCursor || !cursor.IsAgent() {
		t.Fatalf("Cursor detection = %#v", cursor.Agents)
	}

	opencode := runby.Detect(runby.WithEnviron([]string{"OPENCODE_CLIENT=ACP"}))
	layer, ok := opencode.Agent(runby.AgentOpenCode)
	if !ok || layer.Entrypoint != "acp" || layer.Confidence != runby.ConfidenceProbable {
		t.Fatalf("OpenCode detection = %#v", layer)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"OPENCODE_CLIENT=cli"})); got.IsAgent() {
		t.Fatalf("OPENCODE_CLIENT=cli detected: %#v", got.Agents)
	}

	antigravity := runby.Detect(runby.WithEnviron([]string{"ANTIGRAVITY_EXECUTABLE_DATA_DIR=/data/ag"}))
	layer, ok = antigravity.Agent(runby.AgentAntigravity2)
	if !ok || layer.Paths.DataDirectory != "/data/ag" || layer.Entrypoint != "sidecar" {
		t.Fatalf("Antigravity detection = %#v", layer)
	}
}

func TestEmptyValueIsNotEvidence(t *testing.T) {
	if got := runby.Detect(runby.WithEnviron([]string{"PASEO_AGENT_ID=   "})); got.IsAgent() {
		t.Fatalf("blank PASEO_AGENT_ID detected: %#v", got.Agents)
	}
}

func TestLookupFuncAndLastValueWins(t *testing.T) {
	values := map[string]string{"CURSOR_AGENT": "1"}
	got := runby.Detect(runby.WithEnv(runby.LookupFunc(func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	})))
	if primaryAgent(got) != runby.AgentCursor {
		t.Fatalf("Agent() = %q, want %q", primaryAgent(got), runby.AgentCursor)
	}

	dup := runby.Detect(runby.WithEnviron([]string{"PASEO_AGENT_ID=first", "PASEO_AGENT_ID=second"}))
	paseo, _ := dup.Agent(runby.AgentPaseo)
	if paseo.AgentID != "second" {
		t.Fatalf("AgentID = %q, want %q", paseo.AgentID, "second")
	}
}

func TestCustomAgentDriverReplacesABuiltin(t *testing.T) {
	const inHouse runby.AgentName = "acme-orchestrator"
	driver := runby.AgentDriver{
		Agent: inHouse,
		Kind:  runby.KindOrchestrator,
		Detect: func(env runby.Env) (runby.Agent, bool) {
			id, ok := runby.NewEnvReader(env).Value("ACME_RUN_ID")
			if !ok {
				return runby.Agent{}, false
			}
			return runby.Agent{AgentID: id, Axis: runby.Axis{Evidence: []string{"ACME_RUN_ID"}}}, true
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
	if primary.Name != inHouse || primary.Kind != runby.KindOrchestrator ||
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
	for _, driver := range runby.BuiltinDrivers() {
		if agent, ok := driver.(runby.AgentDriver); ok && agent.Kind == runby.KindUnknown {
			t.Fatalf("%q has no Kind", agent.Agent)
		}
	}
	if runby.AgentName("").String() != "unknown" {
		t.Fatalf(`AgentName("").String() = %q`, runby.AgentName("").String())
	}
}

func TestCurrentIsCached(t *testing.T) {
	if !reflect.DeepEqual(runby.Current(), runby.Current()) {
		t.Fatal("Current() is not stable")
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

// TestLadderPlacesCustomDrivers checks that a driver supplied through
// WithOnlyDrivers is placed on the ladder by the same rule as a built-in one:
// what it declares decides the order, not the order it was given in. It also
// checks that an undeclared classification reads as unknown rather than
// empty, so a serialized Agent never carries a blank field.
func TestLadderPlacesCustomDrivers(t *testing.T) {
	detect := func(name string) func(env runby.Env) (runby.Agent, bool) {
		return func(env runby.Env) (runby.Agent, bool) {
			if _, ok := runby.NewEnvReader(env).Value(name); !ok {
				return runby.Agent{}, false
			}
			return runby.Agent{Axis: runby.Axis{Evidence: []string{name}}}, true
		}
	}
	harness := runby.AgentDriver{
		Agent: "acme-harness", Kind: runby.KindHarness,
		Models: runby.ModelsFirstParty, Detect: detect("ACME_HARNESS"),
	}
	orchestrator := runby.AgentDriver{
		Agent: "acme-orchestrator", Kind: runby.KindOrchestrator,
		Models: runby.ModelsDelegated, Detect: detect("ACME_ORCH"),
	}
	bare := runby.AgentDriver{Agent: "acme-bare", Detect: detect("ACME_BARE")}

	// The harness is given first; the orchestrator still comes out on top,
	// and the unclassified driver sorts last.
	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_HARNESS=1", "ACME_ORCH=1", "ACME_BARE=1"}),
		runby.WithOnlyDrivers(harness, bare, orchestrator),
	)
	if got := result.Chain(); got != "acme-orchestrator>acme-harness>acme-bare" {
		t.Fatalf("Chain() = %q", got)
	}

	layer, ok := result.Agent("acme-bare")
	if !ok {
		t.Fatal("the bare driver did not match")
	}
	if layer.Kind != runby.KindUnknown || layer.Models != runby.ModelsUnknown {
		t.Errorf("bare driver classified as (%s, %s), want unknown", layer.Kind, layer.Models)
	}
}

// The five agents added from the vercel/detect-agent survey. Each marker was
// verified against the product's own source or reference documentation before
// the driver was written; docs/research/agents/ records which and why.
func TestAgentsAddedFromTheDetectAgentSurvey(t *testing.T) {
	for _, test := range []struct {
		name       string
		environ    []string
		want       runby.AgentName
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
			layer, ok := result.Agent(test.want)
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
			layer, ok := result.Agent(runby.AgentOpenCode)
			if ok != test.detected {
				t.Fatalf("detected = %v, want %v (%#v)", ok, test.detected, result.Agents)
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

// TestSessionAndAgentIDAcrossLayers pins the reason the two accessors walk the
// layers and report which agent answered: in a Paseo>Codex stack the
// orchestrator names the agent and the harness names the session, and in an
// Orca>Codex stack two layers carry a session at once.
func TestSessionAndAgentIDAcrossLayers(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"PASEO_AGENT_ID=reviewer",
		"CODEX_THREAD_ID=thread-123",
	}))

	session, ok := result.SessionID()
	if !ok || session.Value != "thread-123" || session.Agent != runby.AgentCodex {
		t.Fatalf("SessionID() = %#v, %v, want the harness thread", session, ok)
	}
	agentID, ok := result.AgentID()
	if !ok || agentID.Value != "reviewer" || agentID.Agent != runby.AgentPaseo {
		t.Fatalf("AgentID() = %#v, %v, want the orchestrator agent", agentID, ok)
	}

	// Two layers advertise a session here. The outermost wins, and the agent
	// says which identifier the caller actually got.
	nested := runby.Detect(runby.WithEnviron([]string{
		"ORCA_PANE_KEY=pane-1",
		"ORCA_TAB_ID=tab-1",
		"CODEX_THREAD_ID=thread-123",
	}))
	session, ok = nested.SessionID()
	if !ok || session.Value != "pane-1" || session.Agent != runby.AgentOrca {
		t.Fatalf("SessionID() = %#v, %v, want the outermost layer", session, ok)
	}
	if codex, found := nested.Agent(runby.AgentCodex); !found || codex.SessionID != "thread-123" {
		t.Fatalf("the inner session is still reachable per layer: %#v", codex)
	}
	if agentID, ok = nested.AgentID(); ok || agentID != (runby.Identifier{}) {
		t.Fatalf("AgentID() = %#v, %v, want the zero Identifier when no layer advertises one", agentID, ok)
	}

	none := runby.Detect(runby.WithEnviron(nil))
	if _, ok := none.SessionID(); ok {
		t.Fatal("SessionID() reported a session with no agent detected")
	}
	if _, ok := none.AgentID(); ok {
		t.Fatal("AgentID() reported an agent id with no agent detected")
	}
}

// Unattended is the one place this package combines axes, so its rule is
// pinned by test as well as by doc comment.
func TestUnattended(t *testing.T) {
	interactive := runby.TTY{Inspected: true, StdinTTY: true, StdoutTTY: true, Attached: true, Interactive: true}
	piped := runby.TTY{Inspected: true}

	for _, test := range []struct {
		name    string
		environ []string
		tty     runby.TTY
		want    bool
	}{
		{"a person at a terminal", nil, interactive, false},
		{"output redirected", nil, piped, true},
		{"an agent that allocated a pty", []string{"CODEX_SANDBOX=seatbelt"}, interactive, true},
		{"CI with a pty", []string{"GITHUB_ACTIONS=true"}, interactive, true},
		{"a systemd unit", []string{"INVOCATION_ID=abc"}, interactive, true},
		// A script runner is not a reason on its own: `npm test` typed at a
		// prompt still has a person in front of it.
		{"an npm script at a terminal", []string{"npm_config_user_agent=npm/10.0.0 node/v22"}, interactive, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := runby.Detect(runby.WithEnviron(test.environ), runby.WithTTY(test.tty))
			if got := result.Unattended(); got != test.want {
				t.Errorf("Unattended() = %v, want %v: %#v", got, test.want, result)
			}
		})
	}
}

// An unread TTY is not evidence. A Result built from a bare environment never
// examined the streams, so the TTY clause must not fire on its zero value.
func TestUnattendedIgnoresAnUninspectedTTY(t *testing.T) {
	result := runby.Detect(runby.WithEnviron(nil))
	if result.TTY.Inspected {
		t.Fatal("WithEnviron inspected the standard streams")
	}
	if result.Unattended() {
		t.Errorf("Unattended() = true from an unread TTY: %#v", result.TTY)
	}
}

// Every identity and classification enum in this package renders its zero value
// as "unknown" rather than as the empty string. It matters on the miss path:
// Primary, Layer, Runner, and Remote all return a zero struct when they find
// nothing, and a caller that logs one without checking ok should still write a
// meaningful field rather than a blank one.
//
// This is a table of fmt.Stringer, so an enum added without a String method is
// caught here rather than by a blank column in someone's logs.
func TestZeroEnumsRenderAsUnknown(t *testing.T) {
	for _, value := range []fmt.Stringer{
		runby.AgentName(""),
		runby.Kind(""),
		runby.ModelSource(""),
		runby.Confidence(""),
		runby.Network(""),
		runby.CIProvider(""),
		runby.TerminalProgram(""),
		runby.RemotePlatform(""),
		runby.RemoteKind(""),
		runby.RunnerTool(""),
		runby.RunnerKind(""),
	} {
		if got := value.String(); got != "unknown" {
			t.Errorf("%T zero value renders as %q, want \"unknown\"", value, got)
		}
	}
}

// The same fields on a zero Agent, which is what Primary returns on a miss.
func TestZeroAgentRendersAsUnknown(t *testing.T) {
	var layer runby.Agent
	got := fmt.Sprintf("%s %s %s %s", layer.Name, layer.Kind, layer.Models, layer.Confidence)
	if want := "unknown unknown unknown unknown"; got != want {
		t.Errorf("zero Agent renders as %q, want %q", got, want)
	}
}

// primaryAgent returns the name of the most specific detected layer, or the
// zero name when nothing was detected.
func primaryAgent(r runby.Result) runby.AgentName {
	primary, _ := r.Primary()
	return primary.Name
}
