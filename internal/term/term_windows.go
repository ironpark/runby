package term

import "syscall"

// isTerminal asks for the console mode of the handle. Only a console handle
// has one; every other kind of handle fails.
func isTerminal(fd int) bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Handle(fd), &mode) == nil
}
