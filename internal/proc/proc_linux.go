//go:build linux

package proc

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Supported reports whether this platform can read the ancestor chain.
func Supported() bool { return true }

// selfPPID answers from the runtime rather than from /proc, so the walk starts
// without any I/O.
func selfPPID(reader) (int, bool) { return os.Getppid(), true }

// reader is stateless here: this platform can look up one process directly.
type reader struct{}

func newReader() reader { return reader{} }

func (reader) lookup(pid int) (Info, bool) {
	dir := "/proc/" + strconv.Itoa(pid)

	ppid, ok := readPPID(dir + "/stat")
	if !ok {
		return Info{}, false
	}
	info := Info{PID: pid, PPID: ppid}

	// The exe symlink is the authoritative path but is readable only for our
	// own processes, so a failure here is expected rather than exceptional.
	if path, err := os.Readlink(dir + "/exe"); err == nil {
		info.Path = path
		info.Name = normalize(filepath.Base(path))
	}
	if info.Name == "" {
		// comm is world readable, which is why it is the fallback, but the
		// kernel caps it at CommLimit bytes. A name at exactly that length
		// may be a prefix, so it is flagged rather than trusted whole.
		if raw, err := os.ReadFile(dir + "/comm"); err == nil {
			name := strings.TrimSpace(string(raw))
			info.Name = normalize(name)
			info.Truncated = len(name) >= CommLimit
		}
	}
	return info, true
}

// readPPID parses the fourth field of /proc/<pid>/stat. The second field is
// the executable name in parentheses and may itself contain spaces and
// parentheses, so the scan starts after the last ')' rather than splitting the
// whole line.
func readPPID(path string) (int, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	line := string(raw)
	close := strings.LastIndexByte(line, ')')
	if close < 0 || close+2 >= len(line) {
		return 0, false
	}
	fields := strings.Fields(line[close+2:])
	// After the name, the fields are state then ppid.
	if len(fields) < 2 {
		return 0, false
	}
	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, false
	}
	return ppid, true
}
