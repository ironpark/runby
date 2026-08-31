package runby

// terminalSpec declares a terminal whose detection is a marker plus a set of
// variables read by name. See spec.go for the part shared with the CI and
// remote axes.
type terminalSpec struct {
	specCore
	program TerminalProgram
	// executables names the binaries this terminal runs as, so that a live
	// ancestor process can corroborate an environment detection. It is empty
	// where no name is specific enough to match safely: Apple Terminal's
	// binary is called Terminal, and VTE is a widget library, not a process.
	executables []string

	version   string
	sessionID string
	pid       string
}

// MarkerTermProgram matches TERM_PROGRAM against value, case-insensitively.
// Casing differs across terminals (Apple_Terminal, iTerm.app, WezTerm,
// ghostty, WarpTerminal), and matching loosely avoids depending on it.
func MarkerTermProgram(value string) Marker {
	return func(env Env) bool { return EqualsFold(env, "TERM_PROGRAM", value) }
}

// terminalSpecs is ordered so that a product-specific terminal is reported
// ahead of the generic VTE family, which several products share.
var terminalSpecs = []terminalSpec{
	{
		// Zed owns the terminal but exposes no marker distinguishing a
		// command Zed Agent requested from one a person typed.
		program:     TerminalZed,
		executables: []string{"zed"},
		specCore: specCore{
			marker:      func(env Env) bool { return IsTrue(env, "ZED_TERM") && EqualsFold(env, "TERM_PROGRAM", "zed") },
			markerNames: []string{"ZED_TERM", "TERM_PROGRAM"},
		},
		version: "TERM_PROGRAM_VERSION",
	},
	{
		// VS Code writes TERM_PROGRAM = 'vscode' as a literal, so every fork
		// that does not patch that line reports the same value. The evidence
		// therefore proves a VS Code engine rather than Microsoft's build,
		// which is the same situation as Konsole below and carries the same
		// confidence.
		//
		// The VSCODE_* shell-integration variables are not used as a marker:
		// they exist only when shell integration was injected, so their
		// absence would not mean this is not VS Code. VSCODE_GIT_ASKPASS_NODE
		// would identify the fork, but its value is a filesystem path rather
		// than a product name, and reading values to decide is what this
		// package avoids. See docs/research/terminals/vscode.md.
		program: TerminalVSCode,
		specCore: specCore{
			marker:      MarkerTermProgram("vscode"),
			markerNames: []string{"TERM_PROGRAM"},
			confidence:  ConfidenceProbable,
			extra: map[string]string{
				"vscode.injection": "VSCODE_INJECTION",
				"vscode.stable":    "VSCODE_STABLE",
			},
		},
		version: "TERM_PROGRAM_VERSION",
		// VS Code advertises no window, tab, or terminal identifier.
	},
	{
		// JetBrains sets no TERM_PROGRAM. TERMINAL_EMULATOR names the JediTerm
		// engine shared by every IntelliJ platform IDE, including third-party
		// ones such as Android Studio, so it identifies a family.
		//
		// TERM_SESSION_ID is read only after the marker has decided, because
		// Apple Terminal uses the same variable name for its own session
		// identifier. See docs/research/terminals/jetbrains.md.
		program: TerminalJetBrains,
		specCore: specCore{
			marker:      func(env Env) bool { return EqualsFold(env, "TERMINAL_EMULATOR", "JetBrains-JediTerm") },
			markerNames: []string{"TERMINAL_EMULATOR"},
			confidence:  ConfidenceProbable,
		},
		sessionID: "TERM_SESSION_ID",
	},
	{
		program:     TerminalITerm2,
		executables: []string{"iterm2"},
		specCore: specCore{
			marker:      MarkerTermProgram("iTerm.app"),
			markerNames: []string{"TERM_PROGRAM"},
			extra:       map[string]string{"iterm2.profile": "ITERM_PROFILE"},
		},
		version: "TERM_PROGRAM_VERSION",
		// Shaped w<window>t<tab>p<pane>:<GUID>.
		sessionID: "ITERM_SESSION_ID",
		// LC_TERMINAL is deliberately excluded. iTerm2 puts it in the LC_*
		// namespace because the ssh_config most distributions ship enables
		// SendEnv LC_*, so it crosses SSH on a typical host and can name a
		// terminal on another machine entirely. See docs/research/remote/openssh.md.
	},
	{
		program: TerminalAppleTerminal,
		specCore: specCore{
			marker:      MarkerTermProgram("Apple_Terminal"),
			markerNames: []string{"TERM_PROGRAM"},
		},
		version:   "TERM_PROGRAM_VERSION",
		sessionID: "TERM_SESSION_ID",
	},
	{
		program:     TerminalWezTerm,
		executables: []string{"wezterm", "wezterm-gui"},
		specCore: specCore{
			marker:      MarkerTermProgram("WezTerm"),
			markerNames: []string{"TERM_PROGRAM"},
			extra:       map[string]string{"wezterm.executable": "WEZTERM_EXECUTABLE"},
		},
		version: "TERM_PROGRAM_VERSION",
		// WEZTERM_PANE is a per-mux-instance counter that restarts at 0 and
		// is reused, so it locates a pane but does not identify one durably.
		sessionID: "WEZTERM_PANE",
	},
	{
		program:     TerminalGhostty,
		executables: []string{"ghostty"},
		specCore: specCore{
			marker:      MarkerTermProgram("ghostty"),
			markerNames: []string{"TERM_PROGRAM"},
			// Ghostty advertises no window, tab, or pane identifier.
			extra: map[string]string{"ghostty.resources_dir": "GHOSTTY_RESOURCES_DIR"},
		},
		version: "TERM_PROGRAM_VERSION",
	},
	{
		program:     TerminalWarp,
		executables: []string{"warp"},
		specCore: specCore{
			marker:      MarkerTermProgram("WarpTerminal"),
			markerNames: []string{"TERM_PROGRAM"},
			// WARP_IS_LOCAL_SHELL_SESSION appears only in an unanswered issue,
			// not in Warp's documentation, and Warp is closed source, so it is
			// carried as unverified context rather than used for detection.
			extra: map[string]string{"warp.is_local_shell_session": "WARP_IS_LOCAL_SHELL_SESSION"},
		},
		version: "TERM_PROGRAM_VERSION",
	},
	{
		// kitty deliberately does not set TERM_PROGRAM; its author points to
		// KITTY_WINDOW_ID instead, which is documented as always defined.
		program:     TerminalKitty,
		executables: []string{"kitty"},
		specCore: specCore{
			marker:      MarkerSet("KITTY_WINDOW_ID"),
			markerNames: []string{"KITTY_WINDOW_ID"},
			extra:       map[string]string{"kitty.installation_dir": "KITTY_INSTALLATION_DIR"},
		},
		sessionID: "KITTY_WINDOW_ID",
		pid:       "KITTY_PID",
	},
	{
		// Windows Terminal sets no TERM_PROGRAM either. WT_SESSION is both
		// the marker and the session GUID.
		program:     TerminalWindowsTerminal,
		executables: []string{"windowsterminal"},
		specCore: specCore{
			marker:      MarkerSet("WT_SESSION"),
			markerNames: []string{"WT_SESSION"},
			trimBraces:  true, // WT_PROFILE_ID carries braces; WT_SESSION does not.
			extra:       map[string]string{"windows-terminal.profile_id": "WT_PROFILE_ID"},
		},
		sessionID: "WT_SESSION",
	},
	{
		// Alacritty sets no TERM_PROGRAM. ALACRITTY_WINDOW_ID is gated on
		// version 0.11.0 and on Unix, and ALACRITTY_SOCKET on a build feature
		// and a config option, so neither can be the marker. ALACRITTY_LOG is
		// set unconditionally on every platform at logger initialization.
		program:     TerminalAlacritty,
		executables: []string{"alacritty"},
		specCore: specCore{
			marker:      MarkerSet("ALACRITTY_LOG"),
			markerNames: []string{"ALACRITTY_LOG"},
			extra:       map[string]string{"alacritty.socket": "ALACRITTY_SOCKET"},
		},
		sessionID: "ALACRITTY_WINDOW_ID",
	},
	{
		// Konsole sets no TERM_PROGRAM. KONSOLE_VERSION is a dot-stripped
		// release number, and the KONSOLE_DBUS_* values are live D-Bus paths.
		//
		// Confidence stays probable because konsolepart, the KPart embedded in
		// Dolphin, Kate, KDevelop, and Krusader, links the same library and
		// injects the identical variables. The evidence therefore proves a
		// Konsole engine, not that the user is looking at a Konsole window.
		program:     TerminalKonsole,
		executables: []string{"konsole"},
		specCore: specCore{
			confidence: ConfidenceProbable,
			marker: func(env Env) bool {
				if _, ok := Value(env, "KONSOLE_VERSION"); ok {
					return true
				}
				_, ok := Value(env, "KONSOLE_DBUS_SESSION")
				return ok
			},
			markerNames: []string{"KONSOLE_VERSION", "KONSOLE_DBUS_SESSION"},
			extra: map[string]string{
				"konsole.dbus_service": "KONSOLE_DBUS_SERVICE",
				"konsole.dbus_window":  "KONSOLE_DBUS_WINDOW",
			},
		},
		version:   "KONSOLE_VERSION",
		sessionID: "KONSOLE_DBUS_SESSION",
	},
	{
		// GNOME Terminal sets no TERM_PROGRAM. Only the GNOME_TERMINAL_*
		// D-Bus values are product-specific; VTE_VERSION is not.
		program:     TerminalGNOMETerminal,
		executables: []string{"gnome-terminal-server"},
		specCore: specCore{
			marker:      MarkerSet("GNOME_TERMINAL_SCREEN"),
			markerNames: []string{"GNOME_TERMINAL_SCREEN"},
			extra: map[string]string{
				"gnome-terminal.dbus_service": "GNOME_TERMINAL_SERVICE",
				"gnome-terminal.vte_version":  "VTE_VERSION",
			},
		},
		sessionID: "GNOME_TERMINAL_SCREEN",
	},
	{
		// VTE_VERSION is set by the VTE library itself, so it names the
		// family shared by XFCE Terminal, guake, terminator, sakura, and
		// others. Reporting the family is more honest than guessing a
		// product, so this runs last, after GNOME Terminal has had its turn.
		program: TerminalVTE,
		specCore: specCore{
			marker:      MarkerSet("VTE_VERSION"),
			markerNames: []string{"VTE_VERSION"},
			confidence:  ConfidenceProbable,
		},
		version: "VTE_VERSION",
	},
}

// detect reads the spec's variables out of env.
func (spec terminalSpec) detect(env Env) (Terminal, bool) {
	result := Terminal{Detected: true, Program: spec.program}
	values, ok := spec.read(env,
		specField{spec.version, &result.Version},
		specField{spec.sessionID, &result.SessionID},
	)
	if !ok {
		return Terminal{}, false
	}

	// PID is read on its own because it is the one field that is parsed
	// rather than copied.
	result.PID = parsePID(env, spec.pid)
	values.add(spec.pid)

	// TERM is context only. It is reported after the marker has already
	// decided, never as part of deciding.
	result.Term, _ = Value(env, "TERM")
	values.add("TERM")

	values.apply(env, &result.Axis)
	return result, true
}

// builtinTerminalDrivers is ordered from the most product-specific terminal to
// the generic VTE family. Detect reports the first match.
var builtinTerminalDrivers = mapSlice(terminalSpecs, func(spec terminalSpec) TerminalDriver {
	return TerminalDriver{
		Program:     spec.program,
		Executables: spec.executables,
		Detect:      spec.detect,
	}
})

// TerminalDrivers returns the built-in terminal drivers in precedence order.
// The returned slice is a copy and may be reordered, filtered, or adjusted
// before being passed back through WithOnlyTerminalDrivers.
func TerminalDrivers() []TerminalDriver { return cloneSlice(builtinTerminalDrivers) }
