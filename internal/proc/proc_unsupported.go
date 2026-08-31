//go:build !darwin && !linux && !windows

package proc

// Supported reports whether this platform can read the ancestor chain. The
// standard library exposes no portable way to read another process's parent
// or executable, so every platform this module has not implemented reports
// nothing rather than guessing.
func Supported() bool { return false }

func selfPID() int { return 0 }

type reader struct{}

func newReader() reader                { return reader{} }
func (reader) close()                  {}
func (reader) lookup(int) (Info, bool) { return Info{}, false }
