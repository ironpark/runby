package runby_test

import (
	"testing"

	"github.com/ironpark/runby"
)

// The user agent strings below are the ones observed on 2026-08-31 and
// recorded in docs/research/runners/README.md. They are used verbatim so that
// the tests fail if the parsing stops matching what the tools actually emit.
const (
	npmAgent  = "npm/11.13.0 node/v26.1.0 darwin arm64 workspaces/false"
	pnpmAgent = "pnpm/11.24.0 npm/? node/v26.1.0 darwin arm64"
	bunAgent  = "bun/1.3.5 npm/? node/v24.3.0 darwin arm64"
)

func TestNoRunner(t *testing.T) {
	result := runby.Detect(runby.WithEnviron(nil))
	if result.HasRunner() || len(result.Runner) != 0 {
		t.Errorf("empty environment reported %v", result.Runner)
	}
}

// TestNPMFamilyToldApartByUserAgent is the point of this axis's trickiest
// rule. pnpm and Bun put "npm/?" in their own user agent, so a substring
// search for "npm" reports all three as npm. Only a prefix match works.
func TestNPMFamilyToldApartByUserAgent(t *testing.T) {
	for _, test := range []struct {
		name    string
		environ []string
		want    runby.RunnerTool
	}{
		{
			name:    "npm",
			environ: []string{"npm_config_user_agent=" + npmAgent, "npm_lifecycle_event=test", "INIT_CWD=/w"},
			want:    runby.RunnerNPM,
		},
		{
			name:    "pnpm",
			environ: []string{"npm_config_user_agent=" + pnpmAgent, "npm_lifecycle_event=test", "INIT_CWD=/w"},
			want:    runby.RunnerPNPM,
		},
		{
			// Bun sets no INIT_CWD, which is why the family marker cannot be
			// that variable.
			name:    "bun",
			environ: []string{"npm_config_user_agent=" + bunAgent, "npm_lifecycle_event=test"},
			want:    runby.RunnerBun,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := runby.Detect(runby.WithEnviron(test.environ))
			if len(result.Runner) != 1 {
				t.Fatalf("got %d runners, want 1: %v", len(result.Runner), result.Runner)
			}
			runner := result.Runner[0]
			if runner.Tool != test.want {
				t.Errorf("tool = %s, want %s", runner.Tool, test.want)
			}
			if runner.Kind != runby.RunnerKindScript {
				t.Errorf("kind = %s, want script", runner.Kind)
			}
			if runner.Task != "test" {
				t.Errorf("task = %q, want test", runner.Task)
			}
			if !result.HasRunnerBy(test.want) {
				t.Errorf("HasRunnerBy(%s) = false", test.want)
			}
		})
	}
}

// TestLifecycleEventIsNeverAMarker pins that the script name cannot decide
// which tool ran it: every member of the family sets it.
func TestLifecycleEventIsNeverAMarker(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"npm_lifecycle_event=test", "INIT_CWD=/w"}))
	if result.HasRunner() {
		t.Errorf("npm_lifecycle_event alone detected %v", result.Runner)
	}
}

// TestLifecycleScriptIsNeverEvidence holds the one variable this axis refuses
// to touch. npm_lifecycle_script holds the script body, which is an arbitrary
// shell command that can carry an inline credential, so it must not appear in
// Evidence or Extra even though it is set on every detected run.
func TestLifecycleScriptIsNeverEvidence(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"npm_config_user_agent=" + npmAgent,
		"npm_lifecycle_event=deploy",
		"npm_lifecycle_script=deploy --token=SECRET",
	}))
	if len(result.Runner) != 1 {
		t.Fatalf("got %d runners, want 1", len(result.Runner))
	}
	runner := result.Runner[0]
	for _, name := range runner.Evidence {
		if name == "npm_lifecycle_script" {
			t.Error("npm_lifecycle_script was named as evidence")
		}
	}
	for key, value := range runner.Extra {
		if value == "deploy --token=SECRET" {
			t.Errorf("the script body reached Extra under %q", key)
		}
	}
}

// TestMakeUsesLevelNotFlags covers the finding that made MAKELEVEL the marker:
// a plain make exports MAKEFLAGS as the empty string, which this package reads
// as unset, so it cannot decide anything.
func TestMakeUsesLevelNotFlags(t *testing.T) {
	// Observed: a top-level recipe sees MAKELEVEL=1 and MAKEFLAGS empty.
	recipe := runby.Detect(runby.WithEnviron([]string{"MAKELEVEL=1", "MAKEFLAGS="}))
	runner, ok := recipe.RunnerBy(runby.RunnerMake)
	if !ok {
		t.Fatal("a make recipe was not detected")
	}
	if runner.Kind != runby.RunnerKindScript {
		t.Errorf("kind = %s, want script", runner.Kind)
	}
	if runner.Extra["gnu-make.level"] != "1" {
		t.Errorf("level = %q, want 1", runner.Extra["gnu-make.level"])
	}
	for _, name := range runner.Evidence {
		if name == "MAKEFLAGS" {
			t.Error("an empty MAKEFLAGS was counted as evidence")
		}
	}

	// A sub-make recipe carries a deeper level and real flags.
	sub := runby.Detect(runby.WithEnviron([]string{"MAKELEVEL=2", "MAKEFLAGS= --no-print-directory"}))
	if got, _ := sub.RunnerBy(runby.RunnerMake); got.Extra["gnu-make.level"] != "2" {
		t.Errorf("sub-make level = %q, want 2", got.Extra["gnu-make.level"])
	}

	// MAKEFLAGS alone must not detect make, since it is empty in the common case.
	if flags := runby.Detect(runby.WithEnviron([]string{"MAKEFLAGS= -j2"})); flags.HasRunner() {
		t.Errorf("MAKEFLAGS alone detected %v", flags.Runner)
	}
}

func TestSystemdService(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"INVOCATION_ID=cc8fdc149b2b4ca698d4f259f4054236",
		"JOURNAL_STREAM=8:1234",
	}))
	runner, ok := result.RunnerBy(runby.RunnerSystemd)
	if !ok {
		t.Fatal("a systemd unit was not detected")
	}
	if runner.Kind != runby.RunnerKindService {
		t.Errorf("kind = %s, want service", runner.Kind)
	}
	if runner.Extra["systemd.journal_stream"] != "8:1234" {
		t.Errorf("journal stream = %q", runner.Extra["systemd.journal_stream"])
	}

	// systemd itself warns that JOURNAL_STREAM being set is not sufficient to
	// conclude anything, so it must never decide on its own.
	if only := runby.Detect(runby.WithEnviron([]string{"JOURNAL_STREAM=8:1234"})); only.HasRunner() {
		t.Errorf("JOURNAL_STREAM alone detected %v", only.Runner)
	}

	// A service is the one kind that answers "is anybody watching".
	if _, ok := result.RunnerOfKind(runby.RunnerKindService); !ok {
		t.Error("RunnerOfKind(service) found nothing")
	}
}

func TestPreCommitHook(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"PRE_COMMIT=1"}))
	runner, ok := result.RunnerBy(runby.RunnerPreCommit)
	if !ok {
		t.Fatal("pre-commit was not detected")
	}
	if runner.Kind != runby.RunnerKindHook {
		t.Errorf("kind = %s, want hook", runner.Kind)
	}

	// SKIP is an input pre-commit reads rather than sets, and the name is far
	// too generic to be evidence of anything.
	if skip := runby.Detect(runby.WithEnviron([]string{"SKIP=flake8"})); skip.HasRunner() {
		t.Errorf("SKIP alone detected %v", skip.Runner)
	}
}

// TestGitHooksAreNotDetected pins a deliberate absence. Observed on git 2.55.0,
// a post-checkout hook and a plain git alias receive the same GIT_* variables,
// so no marker separates a hook from anything else git runs. Adding a driver
// keyed on any of them would report aliases, pagers, and filters as hooks.
func TestGitHooksAreNotDetected(t *testing.T) {
	for _, test := range []struct {
		name    string
		environ []string
	}{
		{
			name: "post-checkout hook",
			environ: []string{
				"GIT_EDITOR=:",
				"GIT_EXEC_PATH=/usr/libexec/git-core",
				"GIT_PREFIX=",
			},
		},
		{
			// Identical to the hook above but for a git alias, which is the
			// whole problem.
			name: "git alias",
			environ: []string{
				"GIT_CONFIG_PARAMETERS=''",
				"GIT_EDITOR=:",
				"GIT_EXEC_PATH=/usr/libexec/git-core",
				"GIT_PREFIX=",
			},
		},
		{
			name: "pre-commit hook",
			environ: []string{
				"GIT_AUTHOR_NAME=t",
				"GIT_EXEC_PATH=/usr/libexec/git-core",
				"GIT_INDEX_FILE=.git/index",
				"GIT_PREFIX=",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if result := runby.Detect(runby.WithEnviron(test.environ)); result.HasRunner() {
				t.Errorf("GIT_* variables detected %v", result.Runner)
			}
		})
	}
}

// TestRunnersNest covers the reason this axis is a slice: a pre-commit hook
// running an npm script that shells out to make is three concurrent layers,
// not a precedence contest.
func TestRunnersNest(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"PRE_COMMIT=1",
		"npm_config_user_agent=" + npmAgent,
		"npm_lifecycle_event=lint",
		"MAKELEVEL=1",
	}))
	if len(result.Runner) != 3 {
		t.Fatalf("got %d runners, want 3: %v", len(result.Runner), result.Runner)
	}
	for _, tool := range []runby.RunnerTool{runby.RunnerNPM, runby.RunnerMake, runby.RunnerPreCommit} {
		if !result.HasRunnerBy(tool) {
			t.Errorf("%s was not among the layers", tool)
		}
	}
	if _, ok := result.RunnerOfKind(runby.RunnerKindHook); !ok {
		t.Error("the hook layer was not reported")
	}
}

// TestRunnerIsIndependentOfCIAndTTY is why the axis was added. A script run in
// CI is both, and a script run locally with a terminal attached is neither CI
// nor an agent — which is exactly the case the other axes describe as an
// interactive command a person typed.
func TestRunnerIsIndependentOfCIAndTTY(t *testing.T) {
	local := runby.Detect(
		runby.WithEnviron([]string{"npm_config_user_agent=" + npmAgent, "npm_lifecycle_event=test"}),
		runby.WithTTY(runby.TTY{Inspected: true, StdinTTY: true, StdoutTTY: true, Attached: true, Interactive: true}),
	)
	if local.IsCI() || local.IsAgent() {
		t.Fatal("the local fixture was expected to look like a person at a terminal")
	}
	if !local.TTY.Interactive {
		t.Fatal("the local fixture was expected to be interactive")
	}
	// Only the runner axis knows a script ran this.
	if !local.HasRunner() {
		t.Error("an interactive npm script reported no runner")
	}

	inCI := runby.Detect(runby.WithEnviron([]string{
		"GITHUB_ACTIONS=true", "CI=true",
		"npm_config_user_agent=" + npmAgent, "npm_lifecycle_event=test",
	}))
	if !inCI.IsCI() {
		t.Error("CI was not detected alongside the runner")
	}
	if !inCI.HasRunnerBy(runby.RunnerNPM) {
		t.Error("the runner was not detected alongside CI")
	}
}

// TestCustomRunnerDriver covers the extension path, including that an
// undeclared kind reads as unknown rather than empty.
func TestCustomRunnerDriver(t *testing.T) {
	acme := runby.RunnerDriver{
		Tool: "acme-task",
		Kind: runby.RunnerKindScript,
		Detect: func(env runby.Env) (runby.Runner, bool) {
			task, ok := runby.Value(env, "ACME_TASK")
			if !ok {
				return runby.Runner{}, false
			}
			return runby.Runner{
				Task: task,
				Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_TASK")},
			}, true
		},
	}
	result := runby.Detect(runby.WithEnviron([]string{"ACME_TASK=build"}), runby.WithOnlyDrivers(append(runby.BuiltinDrivers(), acme)...))
	runner, ok := result.RunnerBy("acme-task")
	if !ok {
		t.Fatal("the custom driver did not match")
	}
	if runner.Task != "build" || runner.Confidence != runby.ConfidenceDefinite {
		t.Errorf("task = %q, confidence = %s", runner.Task, runner.Confidence)
	}

	bare := runby.RunnerDriver{
		Tool: "bare",
		Detect: func(env runby.Env) (runby.Runner, bool) {
			return runby.Runner{}, true
		},
	}
	only := runby.Detect(runby.WithEnviron(nil), runby.WithOnlyDrivers(bare))
	if len(only.Runner) != 1 || only.Runner[0].Kind != runby.RunnerKindUnknown {
		t.Errorf("an undeclared kind did not default to unknown: %v", only.Runner)
	}

	// Passing no drivers disables the axis.
	off := runby.Detect(runby.WithEnviron([]string{"MAKELEVEL=1"}), runby.WithOnlyDrivers())
	if off.HasRunner() {
		t.Errorf("WithOnlyDrivers() left the axis on: %v", off.Runner)
	}
}
