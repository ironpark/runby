package runby

import "strings"

// terminalSpec declares a terminal whose detection is a marker plus a set of
// variables read by name.
type terminalSpec struct {
	program TerminalProgram
	// marker reports whether the environment shows this terminal.
	marker func(Env) bool
	// markerNames lists the variables marker consults, so that they are
	// reported as evidence alongside the fields read below.
	markerNames []string
	// confidence defaults to ConfidenceDefinite when empty.
	confidence Confidence

	version   string
	sessionID string
	pid       string

	// trimBraces strips the curly braces some terminals wrap GUIDs in.
	trimBraces bool

	extra    map[string]string
	evidence []string
}

// markerTermProgram matches TERM_PROGRAM against value, case-insensitively.
// Casing differs across terminals (Apple_Terminal, iTerm.app, WezTerm,
// ghostty, WarpTerminal), and matching loosely avoids depending on it.
func markerTermProgram(value string) func(Env) bool {
	return func(env Env) bool { return EqualsFold(env, "TERM_PROGRAM", value) }
}

// terminalSpecs is ordered so that a product-specific terminal is reported
// ahead of the generic VTE family, which several products share.
var terminalSpecs = []terminalSpec{
	{
		// Zed owns the terminal but exposes no marker distinguishing a
		// command Zed Agent requested from one a person typed.
		program:     TerminalZed,
		marker:      func(env Env) bool { return IsTrue(env, "ZED_TERM") && EqualsFold(env, "TERM_PROGRAM", "zed") },
		markerNames: []string{"ZED_TERM", "TERM_PROGRAM"},
		version:     "TERM_PROGRAM_VERSION",
	},
	{
		program:     TerminalITerm2,
		marker:      markerTermProgram("iTerm.app"),
		markerNames: []string{"TERM_PROGRAM"},
		version:     "TERM_PROGRAM_VERSION",
		// Shaped w<window>t<tab>p<pane>:<GUID>.
		sessionID: "ITERM_SESSION_ID",
		extra:     map[string]string{"iterm2.profile": "ITERM_PROFILE"},
		// LC_TERMINAL is deliberately excluded. iTerm2 puts it in the LC_*
		// namespace because the ssh_config most distributions ship enables
		// SendEnv LC_*, so it crosses SSH on a typical host and can name a
		// terminal on another machine entirely. See docs/research/remote/openssh.md.
	},
	{
		program:     TerminalAppleTerminal,
		marker:      markerTermProgram("Apple_Terminal"),
		markerNames: []string{"TERM_PROGRAM"},
		version:     "TERM_PROGRAM_VERSION",
		sessionID:   "TERM_SESSION_ID",
	},
	{
		program:     TerminalWezTerm,
		marker:      markerTermProgram("WezTerm"),
		markerNames: []string{"TERM_PROGRAM"},
		version:     "TERM_PROGRAM_VERSION",
		// WEZTERM_PANE is a per-mux-instance counter that restarts at 0 and
		// is reused, so it locates a pane but does not identify one durably.
		sessionID: "WEZTERM_PANE",
		extra:     map[string]string{"wezterm.executable": "WEZTERM_EXECUTABLE"},
	},
	{
		program:     TerminalGhostty,
		marker:      markerTermProgram("ghostty"),
		markerNames: []string{"TERM_PROGRAM"},
		version:     "TERM_PROGRAM_VERSION",
		// Ghostty advertises no window, tab, or pane identifier.
		extra: map[string]string{"ghostty.resources_dir": "GHOSTTY_RESOURCES_DIR"},
	},
	{
		program:     TerminalWarp,
		marker:      markerTermProgram("WarpTerminal"),
		markerNames: []string{"TERM_PROGRAM"},
		version:     "TERM_PROGRAM_VERSION",
		// WARP_IS_LOCAL_SHELL_SESSION appears only in an unanswered issue,
		// not in Warp's documentation, and Warp is closed source, so it is
		// carried as unverified context rather than used for detection.
		extra: map[string]string{"warp.is_local_shell_session": "WARP_IS_LOCAL_SHELL_SESSION"},
	},
	{
		// kitty deliberately does not set TERM_PROGRAM; its author points to
		// KITTY_WINDOW_ID instead, which is documented as always defined.
		program:     TerminalKitty,
		marker:      markerSet("KITTY_WINDOW_ID"),
		markerNames: []string{"KITTY_WINDOW_ID"},
		sessionID:   "KITTY_WINDOW_ID",
		pid:         "KITTY_PID",
		extra:       map[string]string{"kitty.installation_dir": "KITTY_INSTALLATION_DIR"},
	},
	{
		// Windows Terminal sets no TERM_PROGRAM either. WT_SESSION is both
		// the marker and the session GUID.
		program:     TerminalWindowsTerminal,
		marker:      markerSet("WT_SESSION"),
		markerNames: []string{"WT_SESSION"},
		sessionID:   "WT_SESSION",
		trimBraces:  true, // WT_PROFILE_ID carries braces; WT_SESSION does not.
		extra:       map[string]string{"windows-terminal.profile_id": "WT_PROFILE_ID"},
	},
	{
		// Alacritty sets no TERM_PROGRAM. ALACRITTY_WINDOW_ID is gated on
		// version 0.11.0 and on Unix, and ALACRITTY_SOCKET on a build feature
		// and a config option, so neither can be the marker. ALACRITTY_LOG is
		// set unconditionally on every platform at logger initialization.
		program:     TerminalAlacritty,
		marker:      markerSet("ALACRITTY_LOG"),
		markerNames: []string{"ALACRITTY_LOG"},
		sessionID:   "ALACRITTY_WINDOW_ID",
		extra:       map[string]string{"alacritty.socket": "ALACRITTY_SOCKET"},
	},
	{
		// Konsole sets no TERM_PROGRAM. KONSOLE_VERSION is a dot-stripped
		// release number, and the KONSOLE_DBUS_* values are live D-Bus paths.
		//
		// Confidence stays probable because konsolepart, the KPart embedded in
		// Dolphin, Kate, KDevelop, and Krusader, links the same library and
		// injects the identical variables. The evidence therefore proves a
		// Konsole engine, not that the user is looking at a Konsole window.
		program:    TerminalKonsole,
		confidence: ConfidenceProbable,
		marker: func(env Env) bool {
			if _, ok := Value(env, "KONSOLE_VERSION"); ok {
				return true
			}
			_, ok := Value(env, "KONSOLE_DBUS_SESSION")
			return ok
		},
		markerNames: []string{"KONSOLE_VERSION", "KONSOLE_DBUS_SESSION"},
		version:     "KONSOLE_VERSION",
		sessionID:   "KONSOLE_DBUS_SESSION",
		extra: map[string]string{
			"konsole.dbus_service": "KONSOLE_DBUS_SERVICE",
			"konsole.dbus_window":  "KONSOLE_DBUS_WINDOW",
		},
	},
	{
		// GNOME Terminal sets no TERM_PROGRAM. Only the GNOME_TERMINAL_*
		// D-Bus values are product-specific; VTE_VERSION is not.
		program:     TerminalGNOMETerminal,
		marker:      markerSet("GNOME_TERMINAL_SCREEN"),
		markerNames: []string{"GNOME_TERMINAL_SCREEN"},
		sessionID:   "GNOME_TERMINAL_SCREEN",
		extra: map[string]string{
			"gnome-terminal.dbus_service": "GNOME_TERMINAL_SERVICE",
			"gnome-terminal.vte_version":  "VTE_VERSION",
		},
	},
	{
		// VTE_VERSION is set by the VTE library itself, so it names the
		// family shared by XFCE Terminal, guake, terminator, sakura, and
		// others. Reporting the family is more honest than guessing a
		// product, so this runs last, after GNOME Terminal has had its turn.
		program:     TerminalVTE,
		marker:      markerSet("VTE_VERSION"),
		markerNames: []string{"VTE_VERSION"},
		confidence:  ConfidenceProbable,
		version:     "VTE_VERSION",
	},
}

// detect reads the spec's variables out of env.
func (spec terminalSpec) detect(env Env) (Terminal, bool) {
	if !spec.marker(env) {
		return Terminal{}, false
	}

	confidence := spec.confidence
	if confidence == "" {
		confidence = ConfidenceDefinite
	}
	result := Terminal{
		Detected:   true,
		Program:    spec.program,
		Confidence: confidence,
		PID:        parsePID(env, spec.pid),
	}

	names := append(append([]string{}, spec.markerNames...), spec.evidence...)
	if spec.pid != "" {
		names = append(names, spec.pid)
	}
	for _, field := range []struct {
		name string
		into *string
	}{
		{spec.version, &result.Version},
		{spec.sessionID, &result.SessionID},
	} {
		if field.name == "" {
			continue
		}
		names = append(names, field.name)
		value, _ := Value(env, field.name)
		if spec.trimBraces {
			value = strings.TrimSuffix(strings.TrimPrefix(value, "{"), "}")
		}
		*field.into = value
	}

	for key, name := range spec.extra {
		names = append(names, name)
		value, ok := Value(env, name)
		if !ok {
			continue
		}
		if spec.trimBraces {
			value = strings.TrimSuffix(strings.TrimPrefix(value, "{"), "}")
		}
		if result.Extra == nil {
			result.Extra = make(map[string]string, len(spec.extra))
		}
		result.Extra[key] = value
	}

	// TERM is context only. It is reported after the marker has already
	// decided, never as part of deciding.
	result.Term, _ = Value(env, "TERM")
	if result.Term != "" {
		names = append(names, "TERM")
	}

	result.Evidence = PresentNames(env, names...)
	return result, true
}

// builtinTerminalDetectors is ordered from the most product-specific terminal
// to the generic VTE family. Detect reports the first match.
var builtinTerminalDetectors = func() []TerminalDetector {
	detectors := make([]TerminalDetector, 0, len(terminalSpecs))
	for _, spec := range terminalSpecs {
		detectors = append(detectors, NewTerminalDetector(spec.program, spec.detect))
	}
	return detectors
}()

// TerminalDetectors returns the built-in terminal detectors in precedence
// order. The returned slice is a copy and may be reordered or filtered before
// being passed back through WithOnlyTerminalDetectors.
func TerminalDetectors() []TerminalDetector {
	detectors := make([]TerminalDetector, len(builtinTerminalDetectors))
	copy(detectors, builtinTerminalDetectors)
	return detectors
}
