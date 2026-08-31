package runby

import "strings"

// npm, pnpm, Bun, and Yarn are separate products that implement one de facto
// contract: npm defined the script environment and the others follow it for
// compatibility. The tool's own name is the first token of
// npm_config_user_agent, whose default template npm documents as
// "npm/{npm-version} node/{node-version} ...".
//
// It must be matched as a prefix, never as a substring. pnpm and Bun append
// "npm/?" to their own user agent, so searching the value for "npm" reports
// every one of them as npm. Observed 2026-08-31:
//
//	npm    npm/11.13.0 node/v26.1.0 darwin arm64 workspaces/false
//	pnpm   pnpm/11.24.0 npm/? node/v26.1.0 darwin arm64
//	bun    bun/1.3.5 npm/? node/v24.3.0 darwin arm64
//
// See docs/research/runners/README.md.
const npmUserAgent = "npm_config_user_agent"

// npmLifecycleEvent holds the script name. Every tool in the family sets it,
// so it identifies the task but never the tool.
const npmLifecycleEvent = "npm_lifecycle_event"

// markerUserAgent matches when npm_config_user_agent names tool as its first
// token, which is what tells the family members apart.
func markerUserAgent(tool string) Marker {
	prefix := tool + "/"
	return func(env Env) bool {
		agent, ok := Value(env, npmUserAgent)
		return ok && strings.HasPrefix(strings.ToLower(agent), prefix)
	}
}

// npmFamily declares one member of the family. Only the name differs.
//
// npm_lifecycle_script is deliberately absent from every spec here. It holds
// the script body, which is an arbitrary shell command that can carry an
// inline credential, so this package neither copies it nor names it as
// evidence.
func npmFamily(tool RunnerTool, executables []string, extra map[string]string) runnerSpec {
	return runnerSpec{
		tool:        tool,
		kind:        RunnerKindScript,
		executables: executables,
		specCore: specCore{
			marker:      markerUserAgent(string(tool)),
			markerNames: []string{npmUserAgent},
			extra:       extra,
		},
		task: npmLifecycleEvent,
	}
}

// runnerSpec declares a tool whose detection is a marker plus a set of
// variables read by name. See spec.go for the part shared with the other
// spec-driven axes.
type runnerSpec struct {
	specCore
	tool RunnerTool
	kind RunnerKind
	// executables names the binaries this tool runs as. It is empty where the
	// tool runs under an interpreter's name that would mislabel unrelated
	// processes: npm scripts run as node, and pre-commit as python.
	executables []string

	task string
}

// runnerSpecs is in detection order. Unlike the agent and CI axes this is not
// a precedence contest: nesting is normal here, so every match is reported.
var runnerSpecs = []runnerSpec{
	// npm scripts run as node, so npm names no executable: node would mislabel
	// every unrelated node process. pnpm and Bun run under their own names.
	npmFamily(RunnerNPM, nil, map[string]string{
		// INIT_CWD is npm and pnpm only; Bun does not set it, which is why it
		// is context here rather than part of the family marker.
		"npm.init_cwd": "INIT_CWD",
	}),
	npmFamily(RunnerPNPM, []string{"pnpm"}, map[string]string{"pnpm.init_cwd": "INIT_CWD"}),
	npmFamily(RunnerBun, []string{"bun"}, nil),
	{
		// MAKELEVEL rather than MAKEFLAGS. The GNU Make manual says MAKEFLAGS
		// "is always exported", but a plain make with no flags exports it as
		// the empty string, which this package treats as unset. MAKELEVEL is
		// always at least 1 in a recipe: the manual documents the top-level
		// make as 0, and that the increment happens when make sets up the
		// environment for a recipe, so a process started by make never sees 0.
		//
		// BSD make is believed to set MAKELEVEL too and would be reported as
		// gnu-make; see docs/research/runners/gnu-make.md.
		tool:        RunnerMake,
		kind:        RunnerKindScript,
		executables: []string{"make", "gmake"},
		specCore: specCore{
			marker:      MarkerSet("MAKELEVEL"),
			markerNames: []string{"MAKELEVEL"},
			// The recursion depth is context: 1 is a top-level recipe, 2 or
			// more is a sub-make. Make advertises no target name.
			extra: map[string]string{"gnu-make.level": "MAKELEVEL"},
		},
	},
	{
		// INVOCATION_ID is documented as passed to every process run as part
		// of a unit, which makes it the one signal here that answers "is this
		// running as a daemon" from the environment alone.
		//
		// JOURNAL_STREAM is context, never the marker: systemd itself warns
		// that checking whether it is set at all is not sufficient, because a
		// service can replace the standard streams of what it invokes. Using
		// it properly means comparing its device and inode against a live file
		// descriptor, which is a system call rather than an environment read.
		tool:        RunnerSystemd,
		kind:        RunnerKindService,
		executables: []string{"systemd"},
		specCore: specCore{
			marker:      MarkerSet("INVOCATION_ID"),
			markerNames: []string{"INVOCATION_ID"},
			extra:       map[string]string{"systemd.journal_stream": "JOURNAL_STREAM"},
		},
	},
	{
		// pre-commit is the only way this package can report a hook. Git hooks
		// themselves are indistinguishable from anything else git runs; see
		// the Runner documentation and docs/research/runners/README.md.
		//
		// The variable arrived in pre-commit 2.5.0, so its absence does not
		// mean pre-commit is not running. SKIP is not used: pre-commit reads
		// it rather than setting it, and the name is far too generic.
		tool: RunnerPreCommit,
		kind: RunnerKindHook,
		specCore: specCore{
			marker:      MarkerTrue("PRE_COMMIT"),
			markerNames: []string{"PRE_COMMIT"},
		},
	},
}

// detect reads the spec's variables out of env.
func (spec runnerSpec) detect(env Env) (Runner, bool) {
	result := Runner{Tool: spec.tool, Kind: spec.kind}
	values, ok := spec.read(env, specField{spec.task, &result.Task})
	if !ok {
		return Runner{}, false
	}
	values.apply(env, &result.Axis)
	return result, true
}

// builtinRunnerDrivers is in detection order. Every matching driver is
// reported: a pre-commit hook running an npm script that shells out to make is
// three layers, not a precedence contest.
var builtinRunnerDrivers = mapSlice(runnerSpecs, func(spec runnerSpec) RunnerDriver {
	return RunnerDriver{
		Tool:        spec.tool,
		Kind:        spec.kind,
		Executables: spec.executables,
		Detect:      spec.detect,
	}
})

// runnerDrivers returns the built-in runner drivers in detection order. It is
// unexported: the only reason to hand out the built-in table was to filter it
// and pass it back, and WithOnlyDrivers took that job. The copy keeps a caller
// inside this package from reordering the table itself.
func runnerDrivers() []RunnerDriver { return cloneSlice(builtinRunnerDrivers) }
