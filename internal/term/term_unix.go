//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package term

import (
	"syscall"
	"unsafe"
)

// isTerminal asks the kernel for the file descriptor's terminal attributes.
// Only a terminal answers; every other kind of file fails with ENOTTY. This is
// why the check cannot be replaced by a mode test: os.ModeCharDevice is also
// set for /dev/null and would report it as a terminal.
func isTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(ioctlReadTermios),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
	return errno == 0
}
