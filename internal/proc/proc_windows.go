//go:build windows

package proc

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Supported reports whether this platform can read the ancestor chain.
func Supported() bool { return true }

func selfPID() int { return os.Getpid() }

// reader holds one process snapshot. Windows offers no per-process lookup in
// the standard library, only a snapshot of every process, so the whole table
// is read once per walk instead of once per ancestor.
type reader struct{ table map[int]Info }

func newReader() reader {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return reader{}
	}
	defer syscall.CloseHandle(snapshot)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err := syscall.Process32First(snapshot, &entry); err != nil {
		return reader{}
	}

	table := make(map[int]Info)
	for {
		pid := int(entry.ProcessID)
		table[pid] = Info{
			PID:  pid,
			PPID: int(entry.ParentProcessID),
			// The snapshot carries the executable's name but not its path.
			Name: normalize(filepath.Base(syscall.UTF16ToString(entry.ExeFile[:]))),
		}
		if err := syscall.Process32Next(snapshot, &entry); err != nil {
			return reader{table: table}
		}
	}
}

// selfPPID reads our own entry from the snapshot already in hand. os.Getppid
// would be simpler but takes a second machine-wide snapshot to do it.
func selfPPID(r reader) (int, bool) {
	info, ok := r.lookup(os.Getpid())
	return info.PPID, ok
}

func (r reader) lookup(pid int) (Info, bool) {
	info, ok := r.table[pid]
	return info, ok
}
