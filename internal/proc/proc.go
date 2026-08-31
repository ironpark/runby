// Package proc reads the ancestor process chain using only the standard
// library.
//
// It exists because environment variables are a weak kind of evidence: every
// descendant inherits them, a multiplexer can hand a pane values captured
// hours earlier, and any script can export whatever it likes. A parent process
// is different. It cannot be forged by an export, it is not inherited, and if
// it is there at all it is running right now. That makes this the only source
// in this module that can tell a live ancestor from a stale marker.
//
// It is also the least portable. See Supported.
package proc

import "strings"

// Info describes one process in the ancestor chain.
type Info struct {
	PID  int
	PPID int
	// Name is the executable's base name, lowercased with any .exe suffix
	// removed, so one match table serves every platform. It is the value to
	// match on: Path is often unavailable for a process owned by another user.
	Name string
	// Truncated reports that the source of Name imposes a length limit and
	// Name reached it, so it may be a prefix of the real name. Linux falls
	// back to /proc/<pid>/comm when the exe link is unreadable, and comm is
	// capped at CommLimit bytes. A matcher should treat a truncated name as a
	// prefix rather than an exact value.
	Truncated bool
	// Path is the executable's full path, empty when it could not be read.
	Path string
}

// CommLimit is the byte length Linux truncates /proc/<pid>/comm to. It is
// exported so a matcher can reason about Truncated names without repeating
// the constant.
const CommLimit = 15

// normalize puts a base name into the form Info.Name documents.
func normalize(name string) string {
	name = strings.ToLower(name)
	return strings.TrimSuffix(name, ".exe")
}

// maxDepth bounds the walk. Ancestor chains are short in practice, and a
// bound keeps a corrupted or looping chain from spinning.
const maxDepth = 24

// Ancestors returns the chain from this process's parent upward, nearest
// first. It stops at the root, at the depth limit, or at the first process it
// cannot read, which happens routinely once the chain reaches a process owned
// by another user. A short chain is normal and is not an error.
func Ancestors() []Info {
	if !Supported() {
		return nil
	}

	// A reader lets a platform prefetch once per walk. Windows has no
	// per-process lookup, so without this it would snapshot every process on
	// the machine once per ancestor.
	r := newReader()

	pid, ok := selfPPID(r)
	if !ok {
		return nil
	}

	var chain []Info
	seen := make(map[int]bool)
	for depth := 0; depth < maxDepth; depth++ {
		if pid <= 0 || seen[pid] {
			// A repeat means the chain is corrupt; stop rather than loop.
			break
		}
		seen[pid] = true

		info, ok := r.lookup(pid)
		if !ok {
			break
		}
		chain = append(chain, info)
		pid = info.PPID
	}
	return chain
}
