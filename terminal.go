package runby

import (
	"os"

	"golang.org/x/term"
)

// Terminal describes terminal attachment for the process standard streams. It
// is a property of the process, not of any single detected layer, which is why
// it lives on Result rather than on Detection.
//
// Terminal describes the execution channel only. Interactive does not prove
// that a person launched the process; agents and subagents can allocate a PTY.
type Terminal struct {
	// Inspected reports whether the standard streams were actually examined.
	// It is false when Detect was given an environment that does not
	// necessarily belong to this process.
	Inspected bool `json:"inspected"`
	StdinTTY  bool `json:"stdin_tty"`
	StdoutTTY bool `json:"stdout_tty"`
	StderrTTY bool `json:"stderr_tty"`
	// Attached reports whether at least one standard stream is a terminal.
	Attached bool `json:"attached"`
	// Interactive reports whether stdin and at least one output stream are
	// terminals, so the process could prompt and read a reply.
	Interactive bool `json:"interactive"`
}

// InspectTerminal checks the current process standard streams using the
// operating system's terminal API.
func InspectTerminal() Terminal {
	return terminalStatus(
		term.IsTerminal(int(os.Stdin.Fd())),
		term.IsTerminal(int(os.Stdout.Fd())),
		term.IsTerminal(int(os.Stderr.Fd())),
	)
}

func terminalStatus(stdin, stdout, stderr bool) Terminal {
	return Terminal{
		Inspected:   true,
		StdinTTY:    stdin,
		StdoutTTY:   stdout,
		StderrTTY:   stderr,
		Attached:    stdin || stdout || stderr,
		Interactive: stdin && (stdout || stderr),
	}
}
