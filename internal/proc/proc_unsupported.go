//go:build !darwin && !linux && !windows

package proc

// Supported reports whether this platform can read the ancestor chain. The
// standard library exposes no portable way to read another process's parent
// or executable, so every platform this module has not implemented reports
// nothing rather than guessing.
func Supported() bool { return false }

func selfPPID(reader) (int, bool) { return 0, false }

type reader struct{}

func newReader() reader                { return reader{} }
func (reader) lookup(int) (Info, bool) { return Info{}, false }
