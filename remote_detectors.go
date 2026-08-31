package runby

// remoteSpec declares a layer whose detection is a marker plus a set of
// variables read by name.
type remoteSpec struct {
	platform RemotePlatform
	// marker reports whether the environment shows this layer.
	marker func(Env) bool
	// markerNames lists the variables marker consults, so that they are
	// reported as evidence alongside the fields read below.
	markerNames []string
	// confidence defaults to ConfidenceDefinite when empty.
	confidence Confidence

	sessionID string

	extra    map[string]string
	evidence []string
}

// remoteSpecs lists multiplexers before remote environments. The order is a
// detection order only: environment variables cannot prove how the layers
// nest, so Result.Remote must not be read as an ordered chain.
var remoteSpecs = []remoteSpec{
	{
		// TMUX holds "<socket path>,<server pid>,<session id>". The man page
		// documents only that tmux "initialises the TMUX variable with some
		// internal information", so the value is kept whole rather than split
		// into fields this package would be asserting a contract about.
		platform:    RemoteTmux,
		marker:      markerSet("TMUX"),
		markerNames: []string{"TMUX"},
		sessionID:   "TMUX",
		extra:       map[string]string{"tmux.pane": "TMUX_PANE"},
	},
	{
		// STY holds "<pid>.<tty>.<host>", or "<pid>.<name>" under -S.
		platform:    RemoteScreen,
		marker:      markerSet("STY"),
		markerNames: []string{"STY"},
		sessionID:   "STY",
		// WINDOW is set unconditionally but is a very generic name that other
		// software also uses, so it is context and never part of the marker.
		extra: map[string]string{"gnu-screen.window": "WINDOW"},
	},
	{
		// ZELLIJ is set to the literal string "0", so it must be tested for
		// presence. Parsing it as a boolean would read false inside a real
		// Zellij session and silently lose the detection.
		platform:    RemoteZellij,
		marker:      markerSet("ZELLIJ"),
		markerNames: []string{"ZELLIJ"},
		sessionID:   "ZELLIJ_SESSION_NAME",
		extra:       map[string]string{"zellij.pane_id": "ZELLIJ_PANE_ID"},
	},
	{
		// SSH_CONNECTION holds "<client ip> <client port> <server ip>
		// <server port>" and is set for every session. SSH_CLIENT is marked
		// deprecated in OpenSSH's own source and is only a fallback here.
		//
		// SSH_AUTH_SOCK is deliberately excluded: ssh-agent sets it on a
		// purely local desktop session, so it is a classic false positive.
		platform:    RemoteSSH,
		marker:      markerSet("SSH_CONNECTION"),
		markerNames: []string{"SSH_CONNECTION"},
		extra: map[string]string{
			// SSH_TTY exists only when a pty was allocated, so its presence
			// separates an interactive login from `ssh host command`.
			"openssh.tty":              "SSH_TTY",
			"openssh.original_command": "SSH_ORIGINAL_COMMAND",
		},
	},
	{
		// WSL_DISTRO_NAME is not a documented stable contract and is absent
		// for root and systemd services, and WSL_INTEROP can be absent when
		// interop is disabled. Presence is good evidence; absence proves
		// nothing, which is why this stays probable.
		platform: RemoteWSL,
		marker: func(env Env) bool {
			if _, ok := Value(env, "WSL_DISTRO_NAME"); ok {
				return true
			}
			_, ok := Value(env, "WSL_INTEROP")
			return ok
		},
		markerNames: []string{"WSL_DISTRO_NAME", "WSL_INTEROP"},
		confidence:  ConfidenceProbable,
		sessionID:   "WSL_DISTRO_NAME",
		extra:       map[string]string{"wsl.wslenv": "WSLENV"},
	},
	{
		// CODESPACES is documented as always true inside a codespace. A
		// codespace does not set GITHUB_ACTIONS, which is what keeps it from
		// being reported as a CI run.
		platform:    RemoteCodespaces,
		marker:      markerTrue("CODESPACES"),
		markerNames: []string{"CODESPACES"},
		sessionID:   "CODESPACE_NAME",
		extra:       map[string]string{"github-codespaces.repository": "GITHUB_REPOSITORY"},
	},
	{
		// Documented for Gitpod Classic. Whether the rebranded successor
		// keeps the same variable names is unverified, so absence of these
		// says nothing about newer generations.
		platform:    RemoteGitpod,
		marker:      markerSet("GITPOD_WORKSPACE_ID"),
		markerNames: []string{"GITPOD_WORKSPACE_ID"},
		sessionID:   "GITPOD_WORKSPACE_ID",
		extra: map[string]string{
			"gitpod.workspace_url": "GITPOD_WORKSPACE_URL",
			"gitpod.repo_root":     "GITPOD_REPO_ROOT",
		},
	},
	{
		// The Dev Containers specification mandates no marker at all.
		// REMOTE_CONTAINERS is a VS Code implementation detail that the
		// devcontainer CLI, JetBrains Gateway, and DevPod do not set, and
		// DEVCONTAINER is a convention that was proposed and deferred. Both
		// are therefore probable, and their absence is not evidence that the
		// process is outside a container.
		platform: RemoteDevContainer,
		marker: func(env Env) bool {
			if IsTrue(env, "REMOTE_CONTAINERS") {
				return true
			}
			return IsTrue(env, "DEVCONTAINER")
		},
		markerNames: []string{"REMOTE_CONTAINERS", "DEVCONTAINER"},
		confidence:  ConfidenceProbable,
	},
}

// detect reads the spec's variables out of env.
func (spec remoteSpec) detect(env Env) (Remote, bool) {
	if !spec.marker(env) {
		return Remote{}, false
	}

	confidence := spec.confidence
	if confidence == "" {
		confidence = ConfidenceDefinite
	}
	result := Remote{
		Platform:   spec.platform,
		Kind:       spec.platform.Kind(),
		Confidence: confidence,
	}

	names := append(append([]string{}, spec.markerNames...), spec.evidence...)
	if spec.sessionID != "" {
		names = append(names, spec.sessionID)
		result.SessionID, _ = Value(env, spec.sessionID)
	}
	for key, name := range spec.extra {
		names = append(names, name)
		value, ok := Value(env, name)
		if !ok {
			continue
		}
		if result.Extra == nil {
			result.Extra = make(map[string]string, len(spec.extra))
		}
		result.Extra[key] = value
	}

	result.Evidence = PresentNames(env, names...)
	return result, true
}

// builtinRemoteDetectors is in detection order. Unlike the agent axis, every
// matching detector is reported: an SSH session into a Codespace running tmux
// is three concurrent layers, not a precedence contest.
var builtinRemoteDetectors = func() []RemoteDetector {
	detectors := make([]RemoteDetector, 0, len(remoteSpecs))
	for _, spec := range remoteSpecs {
		detectors = append(detectors, NewRemoteDetector(spec.platform, spec.detect))
	}
	return detectors
}()

// RemoteDetectors returns the built-in detectors in detection order. The
// returned slice is a copy and may be reordered or filtered before being
// passed back through WithOnlyRemoteDetectors.
func RemoteDetectors() []RemoteDetector {
	detectors := make([]RemoteDetector, len(builtinRemoteDetectors))
	copy(detectors, builtinRemoteDetectors)
	return detectors
}
