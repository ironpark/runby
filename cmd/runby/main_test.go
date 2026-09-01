package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ironpark/runby"
)

// exec runs the command and returns its exit code with both streams.
func exec(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestExitCodesAreTheDocumentedContract(t *testing.T) {
	// Shell scripts branch on these, so they are the command's real API.
	for _, test := range []struct {
		args []string
		want int
		why  string
	}{
		{nil, 0, "기본 보고"},
		{[]string{"-json"}, 0, "JSON 보고"},
		{[]string{"chain"}, 0, "체인 한 줄"},
		{[]string{"is", "bogus"}, 2, "알 수 없는 축"},
		{[]string{"is"}, 2, "축 누락"},
		{[]string{"is", "agent", "extra"}, 2, "알 수 없는 제품"},
		{[]string{"is", "agent", "codex", "more"}, 2, "인자가 셋"},
		{[]string{"is", "tty", "ghostty"}, 2, "tty 축은 제품을 받지 않음"},
		{[]string{"bogus"}, 2, "알 수 없는 명령"},
		{[]string{"chain", "extra"}, 2, "chain은 인자를 받지 않음"},
		{[]string{"extra"}, 2, "예상하지 못한 인자"},
		{[]string{"-nosuchflag"}, 2, "알 수 없는 플래그"},
	} {
		if code, _, _ := exec(t, test.args...); code != test.want {
			t.Errorf("run(%v) = %d, want %d (%s)", test.args, code, test.want, test.why)
		}
	}
}

func TestIsAnswersWithTheExitCodeOnly(t *testing.T) {
	for _, axis := range []string{"agent", "ci", "terminal", "remote", "tty"} {
		code, stdout, stderr := exec(t, "is", axis)
		if code != 0 && code != 1 {
			t.Errorf("is %s = %d, want 0 or 1", axis, code)
		}
		if stdout != "" || stderr != "" {
			t.Errorf("is %s printed %q / %q, want silence", axis, stdout, stderr)
		}
	}
}

func TestIsMatchesTheLibrary(t *testing.T) {
	// The command must not develop its own opinion about the axes.
	result := runby.Current()
	for axis, want := range map[string]bool{
		"agent":    result.IsAgent(),
		"ci":       result.IsCI(),
		"terminal": result.HasTerminal(),
		"remote":   result.IsRemote(),
		"tty":      result.TTY.Interactive,
	} {
		code, _, _ := exec(t, "is", axis)
		if got := code == 0; got != want {
			t.Errorf("is %s = %v, library says %v", axis, got, want)
		}
	}
}

func TestJSONRoundTripsToTheSameResult(t *testing.T) {
	code, stdout, _ := exec(t, "-json")
	if code != 0 {
		t.Fatalf("-json = %d", code)
	}
	var round runby.Result
	if err := json.Unmarshal([]byte(stdout), &round); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if round.Chain() != runby.Current().Chain() {
		t.Errorf("decoded chain = %q, want %q", round.Chain(), runby.Current().Chain())
	}
	if strings.Contains(stdout, `"evidence": null`) {
		t.Errorf("-json emitted null evidence:\n%s", stdout)
	}
}

func TestChainPrintsOneLine(t *testing.T) {
	code, stdout, _ := exec(t, "chain")
	if code != 0 {
		t.Fatalf("chain = %d", code)
	}
	if got := strings.Count(stdout, "\n"); got != 1 {
		t.Errorf("chain printed %d lines, want 1: %q", got, stdout)
	}
	if strings.TrimSpace(stdout) != runby.Current().Chain() {
		t.Errorf("chain = %q, want %q", strings.TrimSpace(stdout), runby.Current().Chain())
	}
}

func TestReportNamesEveryAxis(t *testing.T) {
	_, stdout, _ := exec(t)
	for _, axis := range []string{"agent", "ci", "terminal", "remote", "tty", "process"} {
		if !strings.Contains(stdout, axis) {
			t.Errorf("report omits the %s axis:\n%s", axis, stdout)
		}
	}
}

func TestVerboseAddsEvidenceNamesAndNeverValues(t *testing.T) {
	// Evidence is variable names only; a value may hold a token. This is the
	// one property of the command that matters for security, so it is pinned
	// against the real environment rather than a fixture.
	_, plain, _ := exec(t)
	_, verbose, _ := exec(t, "-v")
	if len(verbose) < len(plain) {
		t.Fatalf("-v produced less output than the default report")
	}

	result := runby.Current()
	var names []string
	for _, layer := range result.Agents {
		names = append(names, layer.Evidence...)
	}
	names = append(names, result.CI.Evidence...)
	names = append(names, result.Terminal.Evidence...)
	for _, layer := range result.Remotes {
		names = append(names, layer.Evidence...)
	}

	for _, name := range names {
		if !strings.Contains(verbose, name) {
			t.Errorf("-v omits evidence variable %q", name)
		}
		// The name is printed, so the value must not be.
		value, ok := runby.NewEnvReader(runby.EnvironEnv(os.Environ())).Value(name)
		if !ok || len(value) < 12 {
			// Short values collide with ordinary words; skip them rather than
			// assert something the report cannot control.
			continue
		}
		if strings.Contains(verbose, value) {
			t.Errorf("-v printed the VALUE of %s, which may be sensitive", name)
		}
	}
}

func TestVerboseLabelsRunnerAncestors(t *testing.T) {
	result := runby.Result{Process: runby.ProcessTree{
		Inspected: true,
		Ancestors: []runby.Process{{PID: 42, Name: "npm", Runner: runby.RunnerNPM}},
	}}
	var output bytes.Buffer
	report(&output, result, true)
	if !strings.Contains(output.String(), "runner=npm") {
		t.Errorf("-v omits runner ancestor label:\n%s", output.String())
	}
}

func TestMultiplexerWarningNamesEveryDowngradedAxis(t *testing.T) {
	result := runby.Result{Remotes: []runby.Remote{{
		Platform: runby.RemoteTmux,
		Kind:     runby.RemoteKindMultiplexer,
	}}}
	var output bytes.Buffer
	report(&output, result, false)
	if !strings.Contains(output.String(), "환경에서 파생된 축(터미널·에이전트·러너)") {
		t.Errorf("멀티플렉서 경고가 강등 대상 축을 모두 언급하지 않음:\n%s", output.String())
	}
}

func TestUsageIsPrintedToStderrOnMisuse(t *testing.T) {
	code, stdout, stderr := exec(t, "bogus")
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("usage went to stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "runby is <축>") {
		t.Errorf("stderr does not carry usage:\n%s", stderr)
	}
}

// The named form is the same contract as the bare one — exit code only, and no
// opinion of the command's own. Every built-in product on every axis is checked
// against the library answer, so a mapping that drifts is caught here.
func TestIsNamedProductMatchesTheLibrary(t *testing.T) {
	result := runby.Current()

	check := func(axis, product string, want bool) {
		t.Helper()
		code, stdout, stderr := exec(t, "is", axis, product)
		if code != 0 && code != 1 {
			t.Fatalf("is %s %s = %d, want 0 or 1: %s", axis, product, code, stderr)
		}
		if stdout != "" || stderr != "" {
			t.Errorf("is %s %s printed %q / %q, want silence", axis, product, stdout, stderr)
		}
		if got := code == 0; got != want {
			t.Errorf("is %s %s = %v, library says %v", axis, product, got, want)
		}
	}

	for _, agent := range runby.Agents() {
		_, want := result.Agent(agent)
		check("agent", string(agent), want)
	}
	for _, provider := range runby.CIProviders() {
		check("ci", string(provider), result.IsCI() && result.CI.Provider == provider)
	}
	for _, program := range runby.TerminalPrograms() {
		check("terminal", string(program), result.HasTerminal() && result.Terminal.Program == program)
	}
	for _, platform := range runby.RemotePlatforms() {
		_, want := result.Remote(platform)
		check("remote", string(platform), want)
	}
	for _, tool := range runby.RunnerTools() {
		_, want := result.Runner(tool)
		check("runner", string(tool), want)
	}
}

// A typo must not read as a confident "no". A script branching on exit 1 would
// take the wrong path forever and nothing would say why, so an unknown product
// is a usage error that names the valid set.
func TestIsRefusesAnUnknownProduct(t *testing.T) {
	code, stdout, stderr := exec(t, "is", "agent", "codexx")
	if code != 2 {
		t.Fatalf("is agent codexx = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("usage went to stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "codexx") {
		t.Errorf("stderr does not name the typo:\n%s", stderr)
	}
	for _, agent := range runby.Agents() {
		if !strings.Contains(stderr, string(agent)) {
			t.Errorf("stderr omits %q from the valid set:\n%s", agent, stderr)
		}
	}
}

// The tty axis reports whether the standard streams can carry a prompt, which
// is not a product anyone can name.
func TestIsRefusesAProductOnTheTTYAxis(t *testing.T) {
	code, _, stderr := exec(t, "is", "tty", "ghostty")
	if code != 2 {
		t.Fatalf("is tty ghostty = %d, want 2", code)
	}
	if !strings.Contains(stderr, "tty") {
		t.Errorf("stderr does not name the axis:\n%s", stderr)
	}
}
