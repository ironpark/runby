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
func (p RemotePlatform) String() string {
	if p == "" {
		return string(RemoteUnknown)
	}
	return string(p)
}

// RemotePlatforms returns every supported platform in detection order.
func RemotePlatforms() []RemotePlatform {
	platforms := make([]RemotePlatform, 0, len(builtinRemoteDetectors))
	for _, detector := range builtinRemoteDetectors {
		platforms = append(platforms, detector.Platform())
	}
	return platforms
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

// remoteInfo is everything this package knows about a platform that is not
// the environment rule for detecting it.
type remoteInfo struct {
	kind RemoteKind
	// executables names the binaries this platform runs as, so a live
	// ancestor process can corroborate an environment detection. It is empty
	// for a platform that is not a process, such as WSL or a hosted
	// workspace.
	executables []string
}

// remotes is the single source of truth for what a RemotePlatform is.
var remotes = map[RemotePlatform]remoteInfo{
	RemoteTmux:   {kind: RemoteKindMultiplexer, executables: []string{"tmux"}},
	RemoteScreen: {kind: RemoteKindMultiplexer, executables: []string{"screen"}},
	RemoteZellij: {kind: RemoteKindMultiplexer, executables: []string{"zellij"}},
	RemoteSSH:    {kind: RemoteKindEnvironment, executables: []string{"sshd", "sshd-session"}},
	// These describe where the process runs rather than an ancestor process,
	// so there is nothing in the chain to match.
	RemoteWSL:          {kind: RemoteKindEnvironment},
	RemoteCodespaces:   {kind: RemoteKindEnvironment},
	RemoteGitpod:       {kind: RemoteKindEnvironment},
	RemoteDevContainer: {kind: RemoteKindEnvironment},
}

// Kind reports what a detection of p proves. It returns RemoteKindUnknown for
// platforms this package does not support, which is not the map's zero value.
func (p RemotePlatform) Kind() RemoteKind {
	if info, ok := remotes[p]; ok {
		return info.kind
	}
	return RemoteKindUnknown
}

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
	Platform   RemotePlatform `json:"platform"`
	Kind       RemoteKind     `json:"kind"`
	Confidence Confidence     `json:"confidence"`

	// SessionID identifies the multiplexer session, workspace, or distro,
	// when the platform advertises one.
	SessionID string `json:"session_id,omitempty"`

	// AncestorPID is the PID of a running ancestor process whose executable
	// belongs to this layer, or 0 when none was found. As with
	// Detection.AncestorPID, a non-zero value confirms the environment
	// evidence against a live process, and zero is not a denial.
	AncestorPID int `json:"ancestor_pid,omitempty"`

	// Extra holds values that only one platform advertises, keyed by
	// "<platform-slug>.<name>".
	Extra map[string]string `json:"extra,omitempty"`

	// Evidence lists the environment variable names that produced this
	// result, sorted. Their values may be sensitive and are never copied.
	Evidence []string `json:"evidence"`
}

// IsMultiplexer reports whether this layer is a terminal multiplexer.
func (r Remote) IsMultiplexer() bool { return r.Kind == RemoteKindMultiplexer }

// RemoteDetector reports whether an environment shows its platform. Implement
// it to detect a platform this package does not support, then pass it to
// Detect with WithRemoteDetectors.
type RemoteDetector interface {
	// Platform returns the platform this detector reports.
	Platform() RemotePlatform
	// Detect returns the layer, or false if the environment holds no evidence
	// of it. Implementations must not retain env.
	Detect(env Env) (Remote, bool)
}

// NewRemoteDetector adapts a function into a RemoteDetector.
func NewRemoteDetector(platform RemotePlatform, detect func(env Env) (Remote, bool)) RemoteDetector {
	return funcRemoteDetector{platform: platform, detect: detect}
}

type funcRemoteDetector struct {
	platform RemotePlatform
	detect   func(Env) (Remote, bool)
}

func (d funcRemoteDetector) Platform() RemotePlatform      { return d.platform }
func (d funcRemoteDetector) Executables() []string         { return remotes[d.platform].executables }
func (d funcRemoteDetector) Detect(env Env) (Remote, bool) { return d.detect(env) }

// IsRemote reports whether any layer was detected between the user and this
// process, using the cached Current result.
func IsRemote() bool { return len(Current().Remote) > 0 }
