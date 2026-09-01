package runby

import "strconv"

// TerminalProgram identifies a terminal emulator.
type TerminalProgram string

const (
	TerminalUnknown TerminalProgram = "unknown"

	TerminalAppleTerminal   TerminalProgram = "apple-terminal"
	TerminalITerm2          TerminalProgram = "iterm2"
	TerminalWezTerm         TerminalProgram = "wezterm"
	TerminalKitty           TerminalProgram = "kitty"
	TerminalGhostty         TerminalProgram = "ghostty"
	TerminalWarp            TerminalProgram = "warp"
	TerminalAlacritty       TerminalProgram = "alacritty"
	TerminalWindowsTerminal TerminalProgram = "windows-terminal"
	TerminalKonsole         TerminalProgram = "konsole"
	TerminalGNOMETerminal   TerminalProgram = "gnome-terminal"
	// TerminalZed is the terminal Zed owns. It identifies the application
	// that owns the terminal, not that Zed Agent rather than a person ran the
	// command; Zed exposes no agent-specific marker.
	TerminalZed TerminalProgram = "zed"
	// TerminalVSCode is a Visual Studio Code integrated terminal. The value
	// VS Code sets is a literal in source every fork inherits, so this names
	// the family — VS Code, Cursor, Windsurf, and the rest — rather than one
	// product. Cursor advertises its agent separately, so a command Cursor
	// Agent ran appears on the agent axis as well.
	TerminalVSCode TerminalProgram = "vscode"
	// TerminalJetBrains is the JediTerm terminal built into IntelliJ platform
	// IDEs. The marker names the engine, so it covers every JetBrains IDE and
	// third-party IntelliJ platform IDEs such as Android Studio, and it
	// cannot say which one.
	TerminalJetBrains TerminalProgram = "jetbrains"
	// TerminalVTE is a terminal built on the VTE library whose product could
	// not be determined. VTE sets VTE_VERSION for every terminal embedding it
	// (XFCE Terminal, guake, terminator, sakura, and others), so that variable
	// alone names a family rather than a product.
	TerminalVTE TerminalProgram = "vte"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (p TerminalProgram) String() string { return slug(p, TerminalUnknown) }

// TerminalPrograms returns every built-in terminal in detection precedence
// order. As with Agents, registered drivers are not included.
func TerminalPrograms() []TerminalProgram {
	return mapSlice(builtinTerminalDrivers, func(d TerminalDriver) TerminalProgram { return d.Program })
}

// Terminal identifies the terminal emulator that produced this environment.
//
// Terminal is weaker evidence than the other axes, and deliberately so. Every
// descendant process inherits these variables, so Terminal names the emulator
// that *created this environment*, never the emulator *currently attached to
// this process's TTY*. Four things break the correspondence:
//
//   - A multiplexer server keeps the environment of whichever client started
//     it and cannot refresh an already running pane. tmux overwrites
//     TERM_PROGRAM with its own name, erasing the identity of the terminals
//     that use it, while markers it does not touch, such as KITTY_WINDOW_ID,
//     pass through and can name a terminal that closed long ago. GNU Screen
//     and Zellij do not overwrite TERM_PROGRAM and have no refresh mechanism
//     at all, so every marker can go stale there. Result.Multiplexer reports
//     this, and Confidence is capped at ConfidenceProbable when it is set.
//   - Some terminals deliberately forward their identity across SSH, so a
//     process on a remote host can report a terminal that is not on that
//     machine. iTerm2 puts LC_TERMINAL in the LC_* namespace because the
//     ssh_config and sshd_config that mainstream distributions ship enable
//     SendEnv LC_* and AcceptEnv LC_*, so it crosses on a typical host. That
//     is a shipped-configuration convention, not an OpenSSH default: OpenSSH
//     itself documents sending and accepting nothing. Either way the variable
//     can name a terminal on another machine, so it is never used here.
//   - A daemonized process outlives the terminal that started it.
//   - Any script can export these variables.
//
// Treat Terminal as context for presentation decisions, never as a trust
// boundary.
type Terminal struct {
	Detected bool            `json:"detected"`
	Program  TerminalProgram `json:"program"`
	Axis

	// Version is the emulator version it advertised, in its own format.
	Version string `json:"version,omitempty"`
	// SessionID identifies the window, tab, or pane, when the emulator
	// advertises one. Several terminals advertise none, and some advertise a
	// reusable index rather than a stable identifier; see the per-terminal
	// notes in docs/research/terminals.
	SessionID string `json:"session_id,omitempty"`
	// PID is the emulator's process ID, or 0 when it is not advertised. Only
	// kitty advertises one. It is the single signal in this struct that
	// permits a liveness check rather than a snapshot: the process can be
	// looked up to tell a live terminal from a stale marker.
	PID int `json:"pid,omitempty"`
	// Term is the TERM value. It describes terminfo capability rather than
	// identity, users override it, and multiplexers replace it, so it is
	// reported for context but never used as a sole detection signal.
	Term string `json:"term,omitempty"`

	// AncestorPID is the PID of a running ancestor process whose executable
	// belongs to this terminal, or 0 when none was found. As with
	// Agent.AncestorPID, a non-zero value confirms the environment
	// evidence against a live process, and zero is not a denial.
	AncestorPID int `json:"ancestor_pid,omitempty"`
}

// TerminalDriver detects one terminal emulator. It is the unit of extension
// for this axis: the built-in terminals are declared as drivers, and a
// terminal this package does not support is added by passing another to Detect
// through Register or WithOnlyDrivers.
type TerminalDriver struct {
	// Program identifies the terminal this driver reports. Detect fills it
	// into every Terminal the driver returns, so Detect need not repeat it.
	Program TerminalProgram
	// Executables names the binaries this terminal runs as, so that a live
	// ancestor process can corroborate an environment detection. It is the
	// only thing that can tell a live terminal from a marker left behind by
	// one that has closed, so it is worth setting wherever a name is specific
	// enough to match safely.
	Executables []string
	// Detect returns the terminal, or false when the environment holds no
	// evidence of it. It must not retain env. Program, Detected, and a
	// missing Confidence are filled in by Detect.
	Detect func(env Env) (Terminal, bool)
}

// parsePID reads an emulator process ID. Non-numeric and non-positive values
// are reported as 0, meaning unknown.
func parsePID(env Env, name string) int {
	raw, ok := envValue(env, name)
	if !ok {
		return 0
	}
	pid, err := strconv.Atoi(raw)
	if err != nil || pid < 1 {
		return 0
	}
	return pid
}
