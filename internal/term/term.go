// Package term reports whether a file descriptor is a terminal.
//
// It exists so that runby depends only on the standard library. The logic is
// derived from golang.org/x/term, which is BSD-3-Clause licensed by The Go
// Authors, reduced to the single IsTerminal check this module needs and
// rewritten against the standard syscall package rather than golang.org/x/sys.
package term

// IsTerminal reports whether fd is a terminal. It returns false on platforms
// where this module cannot determine the answer from the standard library
// alone; see the per-platform files.
func IsTerminal(fd int) bool { return isTerminal(fd) }
