//go:build darwin

package proc

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

const (
	ctlKern       = 1
	kernProc      = 14
	kernProcPID   = 1
	kernProcArgs2 = 49

	// ppidOffset is the offset of kp_eproc.e_ppid within struct kinfo_proc on
	// 64-bit darwin. Apple ships no header this module can consult, so the
	// offset is verified once at startup against a value the kernel reports
	// another way; see verifyLayout.
	ppidOffset = 560
	// kinfoProcSize is the struct's documented size on 64-bit darwin. A
	// different size means a layout this code was not written against.
	kinfoProcSize = 648
)

var (
	layoutOnce sync.Once
	layoutOK   bool
)

// Supported reports whether this platform can read the ancestor chain. On
// darwin that depends on the kinfo_proc layout still matching, which is
// checked once rather than assumed.
func Supported() bool {
	layoutOnce.Do(verifyLayout)
	return layoutOK
}

// verifyLayout reads our own process record and confirms that the hardcoded
// offset yields the parent PID the kernel already gave us through getppid.
// If Apple ever changes the layout this fails closed, so the package reports
// nothing rather than reporting a number read from the wrong field.
func verifyLayout() {
	buf, err := sysctl(ctlKern, kernProc, kernProcPID, int32(os.Getpid()))
	if err != nil || len(buf) != kinfoProcSize {
		return
	}
	layoutOK = int(*(*int32)(unsafe.Pointer(&buf[ppidOffset]))) == os.Getppid()
}

// selfPPID answers from the runtime rather than from a process record: the
// kernel already knows our parent, so the walk starts without any I/O.
func selfPPID(reader) (int, bool) { return os.Getppid(), true }

// reader is stateless here: this platform can look up one process directly.
type reader struct{}

func newReader() reader { return reader{} }

func (reader) lookup(pid int) (Info, bool) {
	buf, err := sysctl(ctlKern, kernProc, kernProcPID, int32(pid))
	if err != nil || len(buf) < ppidOffset+4 {
		return Info{}, false
	}
	info := Info{
		PID:  pid,
		PPID: int(*(*int32)(unsafe.Pointer(&buf[ppidOffset]))),
	}
	if path, ok := execPath(pid); ok {
		info.Path = path
		info.Name = filepath.Base(path)
	}
	return info, true
}

// execPath reads the executable path from the process argument area. It is
// unreadable for processes owned by another user, which is expected.
func execPath(pid int) (string, bool) {
	buf, err := sysctl(ctlKern, kernProcArgs2, int32(pid))
	if err != nil || len(buf) < 4 {
		return "", false
	}
	// The buffer starts with argc, then the NUL-terminated executable path.
	path := buf[4:]
	for i, b := range path {
		if b == 0 {
			return string(path[:i]), i > 0
		}
	}
	return "", false
}

// sysctl performs a sysctl by MIB and returns the result buffer.
func sysctl(mib ...int32) ([]byte, error) {
	var size uintptr
	if _, _, err := syscall.Syscall6(syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])), uintptr(len(mib)),
		0, uintptr(unsafe.Pointer(&size)), 0, 0); err != 0 {
		return nil, err
	}
	if size == 0 {
		return nil, syscall.ENOENT
	}

	buf := make([]byte, size)
	if _, _, err := syscall.Syscall6(syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])), uintptr(len(mib)),
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0, 0); err != 0 {
		return nil, err
	}
	return buf[:size], nil
}
