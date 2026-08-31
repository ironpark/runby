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
	// TerminalVTE is a terminal built on the VTE library whose product could
	// not be determined. VTE sets VTE_VERSION for every terminal embedding it
	// (XFCE Terminal, guake, terminator, sakura, and others), so that variable
	// alone names a family rather than a product.
	TerminalVTE TerminalProgram = "vte"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (p TerminalProgram) String() string {
	if p == "" {
		return string(TerminalUnknown)
	}
	return string(p)
}

// TerminalPrograms returns every supported terminal in detection precedence
// order.
func TerminalPrograms() []TerminalProgram {
	programs := make([]TerminalProgram, 0, len(builtinTerminalDetectors))
	for _, detector := range builtinTerminalDetectors {
		programs = append(programs, detector.Program())
	}
	return programs
}

// Multiplexer identifies a terminal multiplexer between this process and the
// emulator.
type Multiplexer string

const (
	MultiplexerNone   Multiplexer = ""
	MultiplexerTmux   Multiplexer = "tmux"
	MultiplexerScreen Multiplexer = "screen"
)

// Terminal identifies the terminal emulator that produced this environment.
//
// Terminal is weaker evidence than the other axes, and deliberately so. Every
// descendant process inherits these variables, so Terminal names the emulator
// that *created this environment*, never the emulator *currently attached to
// this process's TTY*. Four things break the correspondence:
//
//   - A multiplexer server keeps the environment of whichever client started
//     it, so a pane can report a terminal that closed long ago. See
//     Multiplexer, which downgrades Confidence when it is set.
//   - Some terminals deliberately forward their identity across SSH, so a
//     process on a remote host can report a terminal that is not on that
//     machine. iTerm2's LC_TERMINAL is forwarded by OpenSSH's default
//     SendEnv LC_*, which is why this package never uses it for detection.
//   - A daemonized process outlives the terminal that started it.
//   - Any script can export these variables.
//
// Treat Terminal as context for presentation decisions, never as a trust
// boundary.
type Terminal struct {
	Detected   bool            `json:"detected"`
	Program    TerminalProgram `json:"program"`
	Confidence Confidence      `json:"confidence"`

	// Version is the emulator version it advertised, in its own format.
	Version string `json:"version,omitempty"`
	// SessionID identifies the window, tab, or pane, when the emulator
	// advertises one. Several terminals advertise none, and some advertise a
	// reusable index rather than a stable identifier; see the per-terminal
	// notes in docs/terminals.
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

	// Multiplexer names the multiplexer between this process and the
	// emulator. When it is set, Program describes the terminal that started
	// the multiplexer server, which may not be the one displaying this pane,
	// so Confidence is capped at ConfidenceProbable.
	Multiplexer Multiplexer `json:"multiplexer,omitempty"`

	// Extra holds values that only one terminal advertises, keyed by
	// "<terminal-slug>.<name>".
	Extra map[string]string `json:"extra,omitempty"`

	// Evidence lists the environment variable names that produced this
	// result, sorted. Their values may be sensitive and are never copied.
	Evidence []string `json:"evidence"`
}

// TerminalDetector reports whether an environment shows its terminal emulator.
// Implement it to detect a terminal this package does not support, then pass
// it to Detect with WithTerminalDetectors.
type TerminalDetector interface {
	// Program returns the terminal this detector reports.
	Program() TerminalProgram
	// Detect returns the terminal, or false if the environment holds no
	// evidence of it. Implementations must not retain env.
	Detect(env Env) (Terminal, bool)
}

// NewTerminalDetector adapts a function into a TerminalDetector.
func NewTerminalDetector(program TerminalProgram, detect func(env Env) (Terminal, bool)) TerminalDetector {
	return funcTerminalDetector{program: program, detect: detect}
}

type funcTerminalDetector struct {
	program TerminalProgram
	detect  func(Env) (Terminal, bool)
}

func (d funcTerminalDetector) Program() TerminalProgram        { return d.program }
func (d funcTerminalDetector) Detect(env Env) (Terminal, bool) { return d.detect(env) }

// detectMultiplexer reports the multiplexer between this process and the
// emulator. TMUX and STY hold the server socket path and session name, so
// their presence is the signal; TERM is not consulted because a multiplexer
// overwrites it with a value that names no product.
func detectMultiplexer(env Env) (Multiplexer, []string) {
	if _, ok := Value(env, "TMUX"); ok {
		return MultiplexerTmux, []string{"TMUX"}
	}
	if _, ok := Value(env, "STY"); ok {
		return MultiplexerScreen, []string{"STY"}
	}
	return MultiplexerNone, nil
}

// parsePID reads an emulator process ID. Non-numeric and non-positive values
// are reported as 0, meaning unknown.
func parsePID(env Env, name string) int {
	raw, ok := Value(env, name)
	if !ok {
		return 0
	}
	pid, err := strconv.Atoi(raw)
	if err != nil || pid < 1 {
		return 0
	}
	return pid
}
