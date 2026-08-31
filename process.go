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

// executables maps a normalized executable name to the product it belongs to.
// Only names specific enough to be worth acting on are listed; a generic name
// would turn every unrelated program into a false positive.
var executables = map[string]Process{
	// Agents.
	"claude":       {Agent: AgentClaudeCode},
	"codex":        {Agent: AgentCodex},
	"cursor-agent": {Agent: AgentCursor},
	"amp":          {Agent: AgentAmp},
	"opencode":     {Agent: AgentOpenCode},
	"paseo":        {Agent: AgentPaseo},

	// Terminals.
	"kitty":                 {Terminal: TerminalKitty},
	"alacritty":             {Terminal: TerminalAlacritty},
	"ghostty":               {Terminal: TerminalGhostty},
	"wezterm":               {Terminal: TerminalWezTerm},
	"wezterm-gui":           {Terminal: TerminalWezTerm},
	"iterm2":                {Terminal: TerminalITerm2},
	"warp":                  {Terminal: TerminalWarp},
	"konsole":               {Terminal: TerminalKonsole},
	"gnome-terminal-server": {Terminal: TerminalGNOMETerminal},
	"windowsterminal":       {Terminal: TerminalWindowsTerminal},
	"zed":                   {Terminal: TerminalZed},

	// Multiplexers and remote layers.
	"tmux":         {Remote: RemoteTmux},
	"screen":       {Remote: RemoteScreen},
	"zellij":       {Remote: RemoteZellij},
	"sshd":         {Remote: RemoteSSH},
	"sshd-session": {Remote: RemoteSSH},
}

// normalizeExecutable lowercases a base name and drops a Windows .exe suffix,
// so one table serves every platform.
func normalizeExecutable(name string) string {
	name = strings.ToLower(name)
	return strings.TrimSuffix(name, ".exe")
}

// inspectProcessTree reads the ancestor chain and annotates each entry with
// the product its executable belongs to.
func inspectProcessTree() ProcessTree {
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
		name := normalizeExecutable(info.Name)
		// A miss yields the zero Process, so the labels stay empty without a
		// second branch, and a new label field needs no change here.
		process := executables[name]
		process.PID, process.PPID, process.Name, process.Path = info.PID, info.PPID, name, info.Path
		tree.Ancestors = append(tree.Ancestors, process)
	}
	return tree
}
