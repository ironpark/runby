package main

import (
	"bytes"
	"encoding/json"
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
		{[]string{"is", "agent", "extra"}, 2, "축이 둘"},
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
		"agent":    result.Found(),
		"ci":       result.IsCI(),
		"terminal": result.IsTerminal(),
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
	for _, layer := range result.Layers {
		names = append(names, layer.Evidence...)
	}
	names = append(names, result.CI.Evidence...)
	names = append(names, result.Terminal.Evidence...)
	for _, layer := range result.Remote {
		names = append(names, layer.Evidence...)
	}

	for _, name := range names {
		if !strings.Contains(verbose, name) {
			t.Errorf("-v omits evidence variable %q", name)
		}
		// The name is printed, so the value must not be.
		value, ok := runby.Value(runby.Environ(), name)
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
