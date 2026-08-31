package runby

import (
	"strings"

	"github.com/ironpark/runby/internal/proc"
)

// Process is one ancestor of this process.
//
// This is the only evidence in a Result that does not come from the
// environment, and it is the strongest kind available here. A parent process
// cannot be produced by an export, it is not inherited by descendants, and if
// it is visible at all it is running now. Where an environment variable can
// only say what was true when this process started, an ancestor says what is
// true at this moment.
type Process struct {
	PID  int `json:"pid"`
	PPID int `json:"ppid"`
	// Name is the executable's base name, lowercased with any .exe suffix
	// removed. It is what matching uses, because Path is often unreadable for
	// a process owned by another user.
	Name string `json:"name"`
	// Path is the executable's full path, empty when it could not be read.
	Path string `json:"path,omitempty"`

	// Agent, Terminal, and Remote name the product this executable is known
	// to belong to, when it is recognized. At most one is set.
	Agent    Agent           `json:"agent,omitempty"`
	Terminal TerminalProgram `json:"terminal,omitempty"`
	Remote   RemotePlatform  `json:"remote,omitempty"`
}

// ProcessTree is the ancestor chain of this process, nearest parent first.
type ProcessTree struct {
	// Inspected reports whether the chain was actually read. It is false when
	// Detect was given an environment that does not necessarily belong to
	// this process, when WithoutProcessTree was used, and on platforms this
	// module cannot read; see Supported.
	Inspected bool `json:"inspected"`
	// Supported reports whether this platform can read ancestors at all. Only
	// Linux, macOS, and Windows can.
	Supported bool `json:"supported"`
	// Ancestors is the chain from the immediate parent upward. It is short or
	// empty when the walk reaches a process owned by another user, which is
	// routine and not an error.
	Ancestors []Process `json:"ancestors,omitempty"`
}

// Find returns the nearest ancestor satisfying match.
func (t ProcessTree) Find(match func(Process) bool) (Process, bool) {
	for _, p := range t.Ancestors {
		if match(p) {
			return p, true
		}
	}
	return Process{}, false
}

// FindAgent returns the nearest ancestor running agent's executable.
func (t ProcessTree) FindAgent(agent Agent) (Process, bool) {
	return t.Find(func(p Process) bool { return p.Agent == agent })
}

// executableLabels collects the name-to-product mapping from a set of drivers.
// Building it per Detect call rather than once at init is what lets a driver
// supplied through WithAgentDrivers and its siblings be corroborated exactly
// like a built-in one.
type executableLabels map[string]Process

func (labels executableLabels) add(executables []string, label Process) {
	for _, name := range executables {
		labels[name] = label
	}
}

// find returns the label for an executable name. A truncated name is matched
// as a prefix, because Linux caps /proc/<pid>/comm and the full name is
// unreadable for a process owned by another user.
func (labels executableLabels) find(info procInfo) (Process, bool) {
	if label, ok := labels[info.Name]; ok {
		return label, true
	}
	if !info.Truncated {
		return Process{}, false
	}
	var match Process
	found := 0
	for name, label := range labels {
		if strings.HasPrefix(name, info.Name) {
			match, found = label, found+1
		}
	}
	// An ambiguous prefix names no single product, so it labels nothing.
	if found == 1 {
		return match, true
	}
	return Process{}, false
}

// procInfo is the shape inspectProcessTree needs from internal/proc, named
// here so the labelling helpers do not import it in their signatures.
type procInfo = proc.Info

// inspectProcessTree reads the ancestor chain and labels each entry with the
// product its executable belongs to.
func inspectProcessTree(labels executableLabels) ProcessTree {
	tree := ProcessTree{Inspected: true, Supported: proc.Supported()}
	if !tree.Supported {
		return tree
	}

	ancestors := proc.Ancestors()
	if len(ancestors) == 0 {
		// Leave Ancestors nil rather than empty: an empty slice and a nil one
		// are the same to a caller but differ across a JSON round trip.
		return tree
	}

	tree.Ancestors = make([]Process, 0, len(ancestors))
	for _, info := range ancestors {
		// A miss yields the zero Process, so the labels stay empty without a
		// second branch, and a new label field needs no change here.
		process, _ := labels.find(info)
		process.PID, process.PPID, process.Name, process.Path = info.PID, info.PPID, info.Name, info.Path
		tree.Ancestors = append(tree.Ancestors, process)
	}
	return tree
}

// labelProcessTree applies labels to a tree that was supplied rather than
// read, so an injected tree is described the same way a live one is.
func labelProcessTree(tree ProcessTree, labels executableLabels) ProcessTree {
	if len(tree.Ancestors) == 0 {
		return tree
	}
	labelled := make([]Process, len(tree.Ancestors))
	for i, process := range tree.Ancestors {
		label, _ := labels.find(procInfo{Name: process.Name})
		label.PID, label.PPID, label.Name, label.Path = process.PID, process.PPID, process.Name, process.Path
		labelled[i] = label
	}
	tree.Ancestors = labelled
	return tree
}
