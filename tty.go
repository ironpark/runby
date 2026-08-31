package runby

import (
	"os"

	"github.com/ironpark/runby/internal/term"
)

// TTY describes terminal attachment for the process standard streams. It is
// the only part of a Result that comes from system calls rather than from the
// environment, which is why it is separate from Terminal: TTY answers whether
// this process's streams are terminals, while Terminal answers which emulator
// produced the environment.
//
// TTY describes the execution channel only. Interactive does not prove that a
// person launched the process; agents and subagents can allocate a PTY.
type TTY struct {
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

// InspectTTY checks the current process standard streams using the operating
// system's terminal API.
//
// On AIX, Solaris, z/OS, and WebAssembly the standard library exposes no way
// to ask, so every field except Inspected is false there. See
// internal/term.
func InspectTTY() TTY {
	return ttyStatus(
		term.IsTerminal(int(os.Stdin.Fd())),
		term.IsTerminal(int(os.Stdout.Fd())),
		term.IsTerminal(int(os.Stderr.Fd())),
	)
}

func ttyStatus(stdin, stdout, stderr bool) TTY {
	return TTY{
		Inspected:   true,
		StdinTTY:    stdin,
		StdoutTTY:   stdout,
		StderrTTY:   stderr,
		Attached:    stdin || stdout || stderr,
		Interactive: stdin && (stdout || stderr),
	}
}
