package runby

// RemotePlatform identifies a layer sitting between the user and this process.
type RemotePlatform string

const (
	RemoteUnknown RemotePlatform = "unknown"

	RemoteTmux         RemotePlatform = "tmux"
	RemoteScreen       RemotePlatform = "gnu-screen"
	RemoteZellij       RemotePlatform = "zellij"
	RemoteSSH          RemotePlatform = "openssh"
	RemoteWSL          RemotePlatform = "wsl"
	RemoteCodespaces   RemotePlatform = "github-codespaces"
	RemoteGitpod       RemotePlatform = "gitpod"
	RemoteDevContainer RemotePlatform = "devcontainers"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (p RemotePlatform) String() string { return slug(p, RemoteUnknown) }

// RemotePlatforms returns every built-in platform in detection order. As with
// Agents, registered drivers are not included.
func RemotePlatforms() []RemotePlatform {
	return mapSlice(builtinRemoteDrivers, func(d RemoteDriver) RemotePlatform { return d.Platform })
}

// RemoteKind separates a terminal multiplexer from a remote or isolated
// execution environment. It mirrors the product_type recorded in docs/research/remote.
type RemoteKind string

const (
	RemoteKindUnknown RemoteKind = "unknown"
	// RemoteKindMultiplexer multiplexes one terminal into many panes. Its
	// server keeps the environment of whichever client started it, so it is
	// the main source of stale evidence on every other axis.
	RemoteKindMultiplexer RemoteKind = "multiplexer"
	// RemoteKindEnvironment is a remote or isolated environment the process
	// runs in, rather than on the user's own machine.
	RemoteKindEnvironment RemoteKind = "environment"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (k RemoteKind) String() string { return slug(k, RemoteKindUnknown) }

// remoteKinds is derived from the built-in driver table, so a driver is the
// one place a platform is registered.
// Remote is one layer detected between the user and this process.
//
// This axis exists because these layers do more than add variables of their
// own: they decide which variables from the outer environment survive, and in
// what form. tmux filters through update-environment, OpenSSH through SendEnv
// and AcceptEnv, WSL through WSLENV, and Dev Containers through containerEnv
// and remoteEnv. A detection here is therefore a caveat on how much the other
// axes can be trusted, not an independent fact about them.
//
// Two limits are structural and documented rather than worked around:
//
//   - Mosh cannot be detected. MOSH_KEY never enters the remote shell's
//     environment at all, and every other MOSH_* variable is client-side, so
//     a normal Mosh session carries no marker. MOSH_KEY is a credential; if
//     it is ever present it must not be read or logged.
//   - Container runtimes cannot be detected. Docker and Podman set no
//     identifying variable; the conventional checks read /.dockerenv,
//     /run/.containerenv, or cgroups, which are files rather than
//     environment variables. Only a tool that advertises itself, such as
//     Dev Containers or Codespaces, is visible here.
type Remote struct {
	Platform RemotePlatform `json:"platform"`
	Kind     RemoteKind     `json:"kind"`
	Axis

	// SessionID identifies the multiplexer session, workspace, or distro,
	// when the platform advertises one.
	SessionID string `json:"session_id,omitempty"`

	// AncestorPID is the PID of a running ancestor process whose executable
	// belongs to this layer, or 0 when none was found. As with
	// Agent.AncestorPID, a non-zero value confirms the environment
	// evidence against a live process, and zero is not a denial.
	AncestorPID int `json:"ancestor_pid,omitempty"`
}

// IsMultiplexer reports whether this layer is a terminal multiplexer.
func (r Remote) IsMultiplexer() bool { return r.Kind == RemoteKindMultiplexer }

// RemoteDriver detects one layer between the user and this process. It is the
// unit of extension for this axis: the built-in platforms are declared as
// drivers, and a platform this package does not support is added by passing
// another through Register or WithOnlyDrivers.
type RemoteDriver struct {
	// Platform identifies the layer this driver reports. Detect fills it into
	// every Remote the driver returns, so Detect need not repeat it.
	Platform RemotePlatform
	// Kind is what a detection of this platform proves. Setting it to
	// RemoteKindMultiplexer is what makes Result.Multiplexer report the layer,
	// and with it the caveat that every other axis may be stale.
	Kind RemoteKind
	// Executables names the binaries this layer runs as, so that a live
	// ancestor process can corroborate an environment detection. Leave it
	// empty for a layer that is not a process.
	Executables []string
	// Detect returns the layer, or false when the environment holds no
	// evidence of it. It must not retain env. Platform, Kind, and a missing
	// Confidence are filled in by Detect.
	Detect func(env Env) (Remote, bool)
}
