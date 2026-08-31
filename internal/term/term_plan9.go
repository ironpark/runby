package term

import "syscall"

// isTerminal resolves the descriptor back to its path, because on Plan 9 the
// console is identified by name rather than by a terminal attribute call.
func isTerminal(fd int) bool {
	path, err := syscall.Fd2path(fd)
	if err != nil {
		return false
	}
	return path == "/dev/cons" || path == "/mnt/term/dev/cons"
}
