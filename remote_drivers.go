package runby

// remoteSpec declares a layer whose detection is a marker plus a set of
// variables read by name. See spec.go for the part shared with the CI and
// terminal axes.
type remoteSpec struct {
	specCore
	platform RemotePlatform
	kind     RemoteKind
	// executables names the binaries this layer runs as, so that a live
	// ancestor process can corroborate an environment detection. It is empty
	// for a layer that is not a process, such as WSL or a hosted workspace:
	// those describe where the process runs, so there is nothing in the chain
	// to match.
	executables []string

	sessionID string
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
		kind:        RemoteKindMultiplexer,
		executables: []string{"tmux"},
		specCore: specCore{
			marker:      MarkerSet("TMUX"),
			markerNames: []string{"TMUX"},
			extra:       map[string]string{"tmux.pane": "TMUX_PANE"},
		},
		sessionID: "TMUX",
	},
	{
		// STY holds "<pid>.<tty>.<host>", or "<pid>.<name>" under -S.
		platform:    RemoteScreen,
		kind:        RemoteKindMultiplexer,
		executables: []string{"screen"},
		specCore: specCore{
			marker:      MarkerSet("STY"),
			markerNames: []string{"STY"},
			// WINDOW is set unconditionally but is a very generic name that other
			// software also uses, so it is context and never part of the marker.
			extra: map[string]string{"gnu-screen.window": "WINDOW"},
		},
		sessionID: "STY",
	},
	{
		// ZELLIJ is set to the literal string "0", so it must be tested for
		// presence. Parsing it as a boolean would read false inside a real
		// Zellij session and silently lose the detection.
		platform:    RemoteZellij,
		kind:        RemoteKindMultiplexer,
		executables: []string{"zellij"},
		specCore: specCore{
			marker:      MarkerSet("ZELLIJ"),
			markerNames: []string{"ZELLIJ"},
			extra:       map[string]string{"zellij.pane_id": "ZELLIJ_PANE_ID"},
		},
		sessionID: "ZELLIJ_SESSION_NAME",
	},
	{
		// SSH_CONNECTION holds "<client ip> <client port> <server ip>
		// <server port>" and is set for every session. SSH_CLIENT is marked
		// deprecated in OpenSSH's own source and is only a fallback here.
		//
		// SSH_AUTH_SOCK is deliberately excluded: ssh-agent sets it on a
		// purely local desktop session, so it is a classic false positive.
		platform:    RemoteSSH,
		kind:        RemoteKindEnvironment,
		executables: []string{"sshd", "sshd-session"},
		specCore: specCore{
			marker:      MarkerSet("SSH_CONNECTION"),
			markerNames: []string{"SSH_CONNECTION"},
			extra: map[string]string{
				// SSH_TTY exists only when a pty was allocated, so its presence
				// separates an interactive login from `ssh host command`.
				"openssh.tty":              "SSH_TTY",
				"openssh.original_command": "SSH_ORIGINAL_COMMAND",
			},
		},
	},
	{
		// WSL_DISTRO_NAME is not a documented stable contract and is absent
		// for root and systemd services, and WSL_INTEROP can be absent when
		// interop is disabled. Presence is good evidence; absence proves
		// nothing, which is why this stays probable.
		platform: RemoteWSL,
		kind:     RemoteKindEnvironment,
		specCore: specCore{
			marker: func(env Env) bool {
				if _, ok := Value(env, "WSL_DISTRO_NAME"); ok {
					return true
				}
				_, ok := Value(env, "WSL_INTEROP")
				return ok
			},
			markerNames: []string{"WSL_DISTRO_NAME", "WSL_INTEROP"},
			confidence:  ConfidenceProbable,
			extra:       map[string]string{"wsl.wslenv": "WSLENV"},
		},
		sessionID: "WSL_DISTRO_NAME",
	},
	{
		// CODESPACES is documented as always true inside a codespace. A
		// codespace does not set GITHUB_ACTIONS, which is what keeps it from
		// being reported as a CI run.
		platform: RemoteCodespaces,
		kind:     RemoteKindEnvironment,
		specCore: specCore{
			marker:      MarkerTrue("CODESPACES"),
			markerNames: []string{"CODESPACES"},
			extra:       map[string]string{"github-codespaces.repository": "GITHUB_REPOSITORY"},
		},
		sessionID: "CODESPACE_NAME",
	},
	{
		// Documented for Gitpod Classic. Whether the rebranded successor
		// keeps the same variable names is unverified, so absence of these
		// says nothing about newer generations.
		platform: RemoteGitpod,
		kind:     RemoteKindEnvironment,
		specCore: specCore{
			marker:      MarkerSet("GITPOD_WORKSPACE_ID"),
			markerNames: []string{"GITPOD_WORKSPACE_ID"},
			extra: map[string]string{
				"gitpod.workspace_url": "GITPOD_WORKSPACE_URL",
				"gitpod.repo_root":     "GITPOD_REPO_ROOT",
			},
		},
		sessionID: "GITPOD_WORKSPACE_ID",
	},
	{
		// The Dev Containers specification mandates no marker at all.
		// REMOTE_CONTAINERS is a VS Code implementation detail that the
		// devcontainer CLI, JetBrains Gateway, and DevPod do not set, and
		// DEVCONTAINER is a convention that was proposed and deferred. Both
		// are therefore probable, and their absence is not evidence that the
		// process is outside a container.
		platform: RemoteDevContainer,
		kind:     RemoteKindEnvironment,
		specCore: specCore{
			marker: func(env Env) bool {
				if IsTrue(env, "REMOTE_CONTAINERS") {
					return true
				}
				return IsTrue(env, "DEVCONTAINER")
			},
			markerNames: []string{"REMOTE_CONTAINERS", "DEVCONTAINER"},
			confidence:  ConfidenceProbable,
		},
	},
}

// detect reads the spec's variables out of env.
func (spec remoteSpec) detect(env Env) (Remote, bool) {
	result := Remote{Platform: spec.platform, Kind: spec.kind}
	values, ok := spec.read(env, specField{spec.sessionID, &result.SessionID})
	if !ok {
		return Remote{}, false
	}

	result.Confidence = values.confidence
	result.Extra = values.extra
	result.Evidence = values.evidence(env)
	return result, true
}

// builtinRemoteDrivers is in detection order. Unlike the agent axis, every
// matching driver is reported: an SSH session into a Codespace running tmux is
// three concurrent layers, not a precedence contest.
var builtinRemoteDrivers = func() []RemoteDriver {
	drivers := make([]RemoteDriver, 0, len(remoteSpecs))
	for _, spec := range remoteSpecs {
		drivers = append(drivers, RemoteDriver{
			Platform:    spec.platform,
			Kind:        spec.kind,
			Executables: spec.executables,
			Detect:      spec.detect,
		})
	}
	return drivers
}()

// RemoteDrivers returns the built-in remote drivers in detection order. The
// returned slice is a copy and may be reordered, filtered, or adjusted before
// being passed back through WithOnlyRemoteDrivers.
func RemoteDrivers() []RemoteDriver {
	drivers := make([]RemoteDriver, len(builtinRemoteDrivers))
	copy(drivers, builtinRemoteDrivers)
	return drivers
}
