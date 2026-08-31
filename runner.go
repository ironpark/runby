package runby

// RunnerTool identifies a tool that runs other programs as its job: a package
// manager executing a script, a build tool executing a recipe, a service
// manager starting a unit.
type RunnerTool string

const (
	RunnerUnknown RunnerTool = "unknown"

	RunnerNPM       RunnerTool = "npm"
	RunnerPNPM      RunnerTool = "pnpm"
	RunnerBun       RunnerTool = "bun"
	RunnerMake      RunnerTool = "gnu-make"
	RunnerSystemd   RunnerTool = "systemd"
	RunnerPreCommit RunnerTool = "pre-commit"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (t RunnerTool) String() string { return slug(t, RunnerUnknown) }

// RunnerTools returns every supported tool in detection order.
func RunnerTools() []RunnerTool {
	return mapSlice(builtinRunnerDrivers, func(d RunnerDriver) RunnerTool { return d.Tool })
}

// RunnerKind separates the kinds of thing that run a program on your behalf.
// It mirrors nothing in the environment; it is the fact about the product that
// makes a detection actionable.
type RunnerKind string

const (
	RunnerKindUnknown RunnerKind = "unknown"
	// RunnerKindScript is a package manager script or build recipe. The
	// command was written down in a project file rather than typed.
	RunnerKindScript RunnerKind = "script"
	// RunnerKindHook is a version control hook, run in response to a repository
	// event rather than to a request.
	RunnerKindHook RunnerKind = "hook"
	// RunnerKindService is a service manager running this as part of a unit,
	// which means no one is watching the output.
	RunnerKindService RunnerKind = "service"
)

// runnerKinds is derived from the built-in driver table, so a tool is
// registered in one place.
var runnerKinds = indexBy(builtinRunnerDrivers, func(d RunnerDriver) (RunnerTool, RunnerKind) {
	return d.Tool, d.Kind
})

// Kind reports what kind of thing t is. It returns RunnerKindUnknown for tools
// this package does not support; a driver supplied through WithRunnerDrivers
// carries its own Kind onto the Runner instead.
func (t RunnerTool) Kind() RunnerKind { return lookupOr(runnerKinds, t, RunnerKindUnknown) }

// Runner is one tool that ran this process.
//
// This axis answers the half of "what launched this process" that the others
// leave out. A process started by `npm test` has a terminal attached, is not in
// CI, and was not launched by an agent, so TTY, CI, and Layers all describe it
// as an interactive command a person typed — and none of them is wrong, but
// none of them says that a script in package.json ran it.
//
// Like Remote, this is a slice: nesting is normal here rather than
// exceptional. A pre-commit hook running an npm script that shells out to make
// is three layers, and the order is a detection order rather than a nesting
// one.
//
// Two limits are structural and documented rather than worked around:
//
//   - Git hooks cannot be detected. A post-checkout hook and a plain git alias
//     receive the same GIT_* variables, so no marker separates a hook from
//     anything else git runs. GIT_EXEC_PATH is present in both and is a
//     documented input setting besides, which this package does not treat as
//     evidence. Only a framework that advertises itself, such as pre-commit,
//     is visible here. See docs/research/runners/README.md for the observed
//     environments.
//   - cron cannot be detected. It sets no identifying variable, and a sparse
//     environment is a guess rather than evidence.
type Runner struct {
	Tool RunnerTool `json:"tool"`
	Kind RunnerKind `json:"kind"`
	Axis

	// Task is the name of the script, target, or unit being run, when the tool
	// advertises one. Only the npm-family package managers do.
	Task string `json:"task,omitempty"`

	// AncestorPID is the PID of a running ancestor process whose executable
	// belongs to this tool, or 0 when none was found. As with
	// Detection.AncestorPID, a non-zero value confirms the environment
	// evidence against a live process, and zero is not a denial.
	AncestorPID int `json:"ancestor_pid,omitempty"`
}

// RunnerDriver detects one tool that runs other programs. It is the unit of
// extension for this axis: the built-in tools are declared as drivers, and a
// tool this package does not support is added by passing another to Detect
// with WithRunnerDrivers.
type RunnerDriver struct {
	// Tool identifies the tool this driver reports. Detect fills it into every
	// Runner the driver returns, so Detect need not repeat it.
	Tool RunnerTool
	// Kind is what kind of thing this tool is, and so what a detection of it
	// means for whether anyone is watching.
	Kind RunnerKind
	// Executables names the binaries this tool runs as, so that a live
	// ancestor process can corroborate an environment detection. Leave it
	// empty when the tool runs under an interpreter's name, such as node or
	// python, which would mislabel unrelated processes.
	Executables []string
	// Detect returns the runner, or false when the environment holds no
	// evidence of this tool. It must not retain env. Tool, Kind, and a missing
	// Confidence are filled in by Detect.
	Detect func(env Env) (Runner, bool)
}
