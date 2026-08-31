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

// Info describes one process in the ancestor chain.
type Info struct {
	PID  int
	PPID int
	// Name is the executable's base name. It is the value to match on: Path
	// is often unavailable for processes owned by another user.
	Name string
	// Path is the executable's full path, empty when it could not be read.
	Path string
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
	defer r.close()

	var chain []Info
	seen := make(map[int]bool)
	self, ok := r.lookup(selfPID())
	if !ok {
		return nil
	}
	pid := self.PPID
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
