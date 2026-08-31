package runby_test

import (
	"reflect"
	"testing"

	"github.com/ironpark/runby"
)

func TestTerminalNotDetected(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"PATH=/usr/bin", "TERM=xterm-256color"}))
	if result.IsTerminal() || result.Terminal.Detected {
		t.Fatalf("Terminal = %#v, want undetected", result.Terminal)
	}
	if result.Terminal.Program != runby.TerminalUnknown || result.Terminal.Confidence != runby.ConfidenceUnknown {
		t.Fatalf("Terminal = %#v", result.Terminal)
	}
	// TERM is still reported as context even when nothing was identified.
	if result.Terminal.Term != "xterm-256color" {
		t.Fatalf("Term = %q", result.Terminal.Term)
	}
}

func TestTerminalTermProgramFamily(t *testing.T) {
	for _, test := range []struct {
		value   string
		program runby.TerminalProgram
	}{
		{"iTerm.app", runby.TerminalITerm2},
		{"Apple_Terminal", runby.TerminalAppleTerminal},
		{"WezTerm", runby.TerminalWezTerm},
		{"ghostty", runby.TerminalGhostty},
		{"WarpTerminal", runby.TerminalWarp},
	} {
		result := runby.Detect(runby.WithEnviron([]string{
			"TERM_PROGRAM=" + test.value, "TERM_PROGRAM_VERSION=1.2.3",
		}))
		if result.Terminal.Program != test.program {
			t.Fatalf("TERM_PROGRAM=%s gave %q, want %q", test.value, result.Terminal.Program, test.program)
		}
		if result.Terminal.Version != "1.2.3" {
			t.Fatalf("TERM_PROGRAM=%s version = %q", test.value, result.Terminal.Version)
		}
	}
}

func TestTerminalKittyExposesPIDForLiveness(t *testing.T) {
	// kitty sets no TERM_PROGRAM; KITTY_WINDOW_ID is the documented marker.
	result := runby.Detect(runby.WithEnviron([]string{
		"KITTY_WINDOW_ID=3", "KITTY_PID=4242", "TERM=xterm-kitty",
	}))

	term := result.Terminal
	if term.Program != runby.TerminalKitty || term.SessionID != "3" {
		t.Fatalf("Terminal = %#v", term)
	}
	// The PID is the only signal here that permits a liveness check.
	if term.PID != 4242 {
		t.Fatalf("PID = %d, want 4242", term.PID)
	}
	if term.Term != "xterm-kitty" {
		t.Fatalf("Term = %q", term.Term)
	}
}

func TestTerminalPIDIgnoresUnparsableValues(t *testing.T) {
	for _, value := range []string{"", "  ", "abc", "0", "-1"} {
		result := runby.Detect(runby.WithEnviron([]string{"KITTY_WINDOW_ID=1", "KITTY_PID=" + value}))
		if result.Terminal.PID != 0 {
			t.Fatalf("KITTY_PID=%q gave PID = %d, want 0", value, result.Terminal.PID)
		}
	}
}

func TestTerminalAlacrittyUsesLogAsMarker(t *testing.T) {
	// ALACRITTY_WINDOW_ID is gated on version 0.11.0 and Unix, and
	// ALACRITTY_SOCKET on a build feature, so neither can be the marker.
	// ALACRITTY_LOG is set unconditionally on every platform.
	logOnly := runby.Detect(runby.WithEnviron([]string{"ALACRITTY_LOG=/tmp/Alacritty-1.log"}))
	if logOnly.Terminal.Program != runby.TerminalAlacritty {
		t.Fatalf("Terminal = %#v", logOnly.Terminal)
	}
	if logOnly.Terminal.SessionID != "" {
		t.Fatalf("SessionID = %q, want empty", logOnly.Terminal.SessionID)
	}

	full := runby.Detect(runby.WithEnviron([]string{
		"ALACRITTY_LOG=/tmp/Alacritty-1.log",
		"ALACRITTY_WINDOW_ID=12345",
		"ALACRITTY_SOCKET=/tmp/Alacritty-wayland-0.sock",
	}))
	if full.Terminal.SessionID != "12345" {
		t.Fatalf("SessionID = %q", full.Terminal.SessionID)
	}
	if full.Terminal.Extra["alacritty.socket"] == "" {
		t.Fatalf("Extra = %#v", full.Terminal.Extra)
	}
}

func TestTerminalWindowsTerminalTrimsProfileBraces(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"WT_SESSION=5720ee6d-6474-47b0-88db-fa7e10e60d37",
		"WT_PROFILE_ID={2ece5bfe-50ed-5f3a-ab87-5cd4baafed2b}",
	}))

	term := result.Terminal
	if term.Program != runby.TerminalWindowsTerminal {
		t.Fatalf("Program = %q", term.Program)
	}
	// WT_SESSION has no braces; WT_PROFILE_ID does.
	if term.SessionID != "5720ee6d-6474-47b0-88db-fa7e10e60d37" {
		t.Fatalf("SessionID = %q", term.SessionID)
	}
	if got := term.Extra["windows-terminal.profile_id"]; got != "2ece5bfe-50ed-5f3a-ab87-5cd4baafed2b" {
		t.Fatalf("profile_id = %q, want braces trimmed", got)
	}
}

func TestTerminalVTEFamilyIsNotGNOMETerminal(t *testing.T) {
	// VTE sets VTE_VERSION for every terminal embedding it, so on its own it
	// names a family. Guessing a product would be dishonest.
	family := runby.Detect(runby.WithEnviron([]string{"VTE_VERSION=8500", "TERM=xterm-256color"}))
	if family.Terminal.Program != runby.TerminalVTE {
		t.Fatalf("Program = %q, want %q", family.Terminal.Program, runby.TerminalVTE)
	}
	if family.Terminal.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q, want %q", family.Terminal.Confidence, runby.ConfidenceProbable)
	}
	if family.Terminal.Version != "8500" {
		t.Fatalf("Version = %q", family.Terminal.Version)
	}

	// Only the GNOME_TERMINAL_* D-Bus values name the product.
	gnome := runby.Detect(runby.WithEnviron([]string{
		"VTE_VERSION=8500",
		"GNOME_TERMINAL_SCREEN=/org/gnome/Terminal/screen/abc",
		"GNOME_TERMINAL_SERVICE=:1.23",
	}))
	if gnome.Terminal.Program != runby.TerminalGNOMETerminal {
		t.Fatalf("Program = %q", gnome.Terminal.Program)
	}
	if gnome.Terminal.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Confidence = %q", gnome.Terminal.Confidence)
	}
	if gnome.Terminal.Extra["gnome-terminal.vte_version"] != "8500" {
		t.Fatalf("Extra = %#v", gnome.Terminal.Extra)
	}
}

func TestTerminalKonsole(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"KONSOLE_VERSION=230800",
		"KONSOLE_DBUS_SESSION=/Sessions/1",
		"KONSOLE_DBUS_SERVICE=:1.42",
		"KONSOLE_DBUS_WINDOW=/Windows/1",
	}))

	term := result.Terminal
	if term.Program != runby.TerminalKonsole || term.Version != "230800" {
		t.Fatalf("Terminal = %#v", term)
	}
	// konsolepart, embedded in Dolphin, Kate, and KDevelop, injects the same
	// variables, so this can never rise above probable.
	if term.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q, want %q", term.Confidence, runby.ConfidenceProbable)
	}
	if term.SessionID != "/Sessions/1" || term.Extra["konsole.dbus_window"] != "/Windows/1" {
		t.Fatalf("Terminal = %#v", term)
	}

	// KONSOLE_DBUS_SESSION alone is enough; KONSOLE_VERSION can be overridden
	// by a profile's Environment setting.
	dbusOnly := runby.Detect(runby.WithEnviron([]string{"KONSOLE_DBUS_SESSION=/Sessions/2"}))
	if dbusOnly.Terminal.Program != runby.TerminalKonsole {
		t.Fatalf("Program = %q", dbusOnly.Terminal.Program)
	}
}

func TestTerminalLCTerminalIsNeverEvidence(t *testing.T) {
	// iTerm2 puts LC_TERMINAL in the LC_* namespace precisely so OpenSSH's
	// default SendEnv LC_* forwards it. A remote process therefore sees it
	// while the terminal is on another machine, so it must not identify one.
	result := runby.Detect(runby.WithEnviron([]string{
		"LC_TERMINAL=iTerm2", "LC_TERMINAL_VERSION=3.5.0", "TERM=xterm-256color",
	}))
	if result.IsTerminal() {
		t.Fatalf("Terminal = %#v, want undetected", result.Terminal)
	}
}

func TestTerminalMultiplexerDowngradesConfidence(t *testing.T) {
	plain := runby.Detect(runby.WithEnviron([]string{"TERM_PROGRAM=ghostty"}))
	if plain.Terminal.Multiplexer != runby.MultiplexerNone || plain.Terminal.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Terminal = %#v", plain.Terminal)
	}

	// Inside tmux the identity describes whichever client started the server,
	// which may not be the terminal displaying this pane.
	for _, test := range []struct {
		entry string
		want  runby.Multiplexer
	}{
		{"TMUX=/tmp/tmux-501/default,123,0", runby.MultiplexerTmux},
		{"STY=1234.pts-0.host", runby.MultiplexerScreen},
	} {
		result := runby.Detect(runby.WithEnviron([]string{"TERM_PROGRAM=ghostty", test.entry}))
		if result.Terminal.Multiplexer != test.want {
			t.Fatalf("%s gave Multiplexer = %q, want %q", test.entry, result.Terminal.Multiplexer, test.want)
		}
		if result.Terminal.Confidence != runby.ConfidenceProbable {
			t.Fatalf("%s gave Confidence = %q, want %q", test.entry, result.Terminal.Confidence, runby.ConfidenceProbable)
		}
		if result.Terminal.Program != runby.TerminalGhostty {
			t.Fatalf("Program = %q", result.Terminal.Program)
		}
	}
}

func TestTerminalMultiplexerWithoutEmulatorIdentity(t *testing.T) {
	// A multiplexer with no surviving terminal identity is still worth
	// reporting, because it explains why the identity is missing.
	result := runby.Detect(runby.WithEnviron([]string{"TMUX=/tmp/tmux-501/default,123,0", "TERM=tmux-256color"}))
	if result.IsTerminal() {
		t.Fatalf("Terminal = %#v, want undetected", result.Terminal)
	}
	if result.Terminal.Multiplexer != runby.MultiplexerTmux {
		t.Fatalf("Multiplexer = %q", result.Terminal.Multiplexer)
	}
}

func TestTerminalMultiplexerIsInEvidence(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"TERM_PROGRAM=WezTerm", "WEZTERM_PANE=0", "TMUX=/tmp/tmux-501/default,1,0",
	}))
	want := []string{"TERM_PROGRAM", "TMUX", "WEZTERM_PANE"}
	if !reflect.DeepEqual(result.Terminal.Evidence, want) {
		t.Fatalf("Evidence = %#v, want %#v", result.Terminal.Evidence, want)
	}
}

func TestTerminalIsIndependentOfAgentAndCI(t *testing.T) {
	// All three axes can be populated at once.
	result := runby.Detect(runby.WithEnviron([]string{
		"CLAUDECODE=1",
		"GITHUB_ACTIONS=true", "GITHUB_RUN_ID=1",
		"TERM_PROGRAM=ghostty",
	}))
	if !result.Found() || !result.IsCI() || !result.IsTerminal() {
		t.Fatalf("result = %#v", result)
	}
	if result.Chain() != "claude-code" {
		t.Fatalf("Chain() = %q", result.Chain())
	}
}

func TestTerminalDetectionUsesEnvironNotTTY(t *testing.T) {
	// Terminal comes from the environment, so WithEnviron still yields it,
	// unlike TTY which needs this process's own file descriptors.
	result := runby.Detect(runby.WithEnviron([]string{"TERM_PROGRAM=WezTerm"}))
	if !result.IsTerminal() {
		t.Fatalf("Terminal = %#v", result.Terminal)
	}
	if result.TTY.Inspected {
		t.Fatalf("TTY = %#v, want uninspected", result.TTY)
	}
}

func TestWithTerminalDetectors(t *testing.T) {
	detector := runby.NewTerminalDetector("acme-term", func(env runby.Env) (runby.Terminal, bool) {
		id, ok := runby.Value(env, "ACME_TERM_SESSION")
		if !ok {
			return runby.Terminal{}, false
		}
		return runby.Terminal{SessionID: id, Evidence: runby.PresentNames(env, "ACME_TERM_SESSION")}, true
	})

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_TERM_SESSION=s-1", "TERM_PROGRAM=ghostty"}),
		runby.WithTerminalDetectors(detector),
	)
	if result.Terminal.Program != "acme-term" || result.Terminal.SessionID != "s-1" {
		t.Fatalf("Terminal = %#v", result.Terminal)
	}

	disabled := runby.Detect(
		runby.WithEnviron([]string{"TERM_PROGRAM=ghostty"}),
		runby.WithOnlyTerminalDetectors(),
	)
	if disabled.IsTerminal() {
		t.Fatalf("Terminal = %#v, want detection disabled", disabled.Terminal)
	}
}

func TestTerminalProgramsAreOrderedAndVTEIsLast(t *testing.T) {
	programs := runby.TerminalPrograms()
	if len(programs) < 2 {
		t.Fatalf("TerminalPrograms() = %#v", programs)
	}
	if last := programs[len(programs)-1]; last != runby.TerminalVTE {
		t.Fatalf("last program = %q, want %q", last, runby.TerminalVTE)
	}
	// GNOME Terminal must precede the VTE family it belongs to.
	gnome, vte := -1, -1
	for i, p := range programs {
		switch p {
		case runby.TerminalGNOMETerminal:
			gnome = i
		case runby.TerminalVTE:
			vte = i
		}
	}
	if gnome < 0 || gnome > vte {
		t.Fatalf("gnome=%d vte=%d in %v", gnome, vte, programs)
	}
	if runby.TerminalProgram("").String() != "unknown" {
		t.Fatalf(`TerminalProgram("").String() = %q`, runby.TerminalProgram("").String())
	}
}
