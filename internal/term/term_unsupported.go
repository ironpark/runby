//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !windows

package term

// isTerminal reports false on platforms where the standard library exposes no
// way to ask. golang.org/x/term answers this for AIX, Solaris, and z/OS by
// going through golang.org/x/sys, which this module deliberately does not
// depend on; on those platforms TTY.Attached and TTY.Interactive are therefore
// always false. Every other axis is unaffected, since they read only the
// environment.
func isTerminal(int) bool { return false }
