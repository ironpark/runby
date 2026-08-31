package runby_test

import (
	"reflect"
	"testing"

	"github.com/ironpark/runby"
)

func TestRemoteNotDetected(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"PATH=/usr/bin"}))
	if result.IsRemote() || len(result.Remote) != 0 {
		t.Fatalf("Remote = %#v, want empty", result.Remote)
	}
	if _, ok := result.Multiplexer(); ok {
		t.Fatal("Multiplexer() ok = true, want false")
	}
}

func TestRemoteMultiplexers(t *testing.T) {
	for _, test := range []struct {
		environ   []string
		want      runby.RemotePlatform
		sessionID string
	}{
		{[]string{"TMUX=/tmp/tmux-501/default,123,0", "TMUX_PANE=%3"}, runby.RemoteTmux, "/tmp/tmux-501/default,123,0"},
		{[]string{"STY=1234.pts-0.host", "WINDOW=0"}, runby.RemoteScreen, "1234.pts-0.host"},
		{[]string{"ZELLIJ=0", "ZELLIJ_SESSION_NAME=main"}, runby.RemoteZellij, "main"},
	} {
		result := runby.Detect(runby.WithEnviron(test.environ))
		layer, ok := result.RemoteLayer(test.want)
		if !ok {
			t.Fatalf("%v gave Remote = %#v", test.environ, result.Remote)
		}
		if layer.SessionID != test.sessionID {
			t.Fatalf("%q SessionID = %q, want %q", test.want, layer.SessionID, test.sessionID)
		}
		if layer.Kind != runby.RemoteKindMultiplexer || !layer.IsMultiplexer() {
			t.Fatalf("%q Kind = %q", test.want, layer.Kind)
		}
	}
}

func TestRemoteZellijMarkerIsNotABoolean(t *testing.T) {
	// Zellij sets ZELLIJ to the literal string "0". Parsing it as a boolean
	// yields false, which would silently lose the detection inside a real
	// session, so detection must key on presence.
	result := runby.Detect(runby.WithEnviron([]string{"ZELLIJ=0"}))
	if !result.HasRemoteLayer(runby.RemoteZellij) {
		t.Fatalf("ZELLIJ=0 not detected: %#v", result.Remote)
	}
	if runby.IsTrue(runby.EnvironEnv([]string{"ZELLIJ=0"}), "ZELLIJ") {
		t.Fatal("IsTrue(ZELLIJ=0) = true; the test's premise is wrong")
	}
}

func TestRemoteScreenWindowIsContextNotMarker(t *testing.T) {
	// WINDOW is set unconditionally by Screen but is a very generic name that
	// other software also uses, so it must never match on its own.
	if got := runby.Detect(runby.WithEnviron([]string{"WINDOW=0"})); got.IsRemote() {
		t.Fatalf("WINDOW alone detected: %#v", got.Remote)
	}
	result := runby.Detect(runby.WithEnviron([]string{"STY=1.pts-0.h", "WINDOW=2"}))
	layer, _ := result.RemoteLayer(runby.RemoteScreen)
	if layer.Extra["gnu-screen.window"] != "2" {
		t.Fatalf("Extra = %#v", layer.Extra)
	}
}

func TestRemoteSSH(t *testing.T) {
	// An interactive login has a pty, so SSH_TTY is present.
	login := runby.Detect(runby.WithEnviron([]string{
		"SSH_CONNECTION=203.0.113.5 54321 198.51.100.9 22",
		"SSH_TTY=/dev/pts/0",
	}))
	layer, ok := login.RemoteLayer(runby.RemoteSSH)
	if !ok || layer.Kind != runby.RemoteKindEnvironment {
		t.Fatalf("Remote = %#v", login.Remote)
	}
	if layer.Extra["openssh.tty"] != "/dev/pts/0" {
		t.Fatalf("Extra = %#v", layer.Extra)
	}

	// `ssh host command` allocates no pty, so SSH_TTY is absent.
	command := runby.Detect(runby.WithEnviron([]string{"SSH_CONNECTION=203.0.113.5 54321 198.51.100.9 22"}))
	layer, _ = command.RemoteLayer(runby.RemoteSSH)
	if _, ok := layer.Extra["openssh.tty"]; ok {
		t.Fatalf("Extra = %#v, want no tty", layer.Extra)
	}
}

func TestRemoteSSHAuthSockIsNotAMarker(t *testing.T) {
	// ssh-agent sets SSH_AUTH_SOCK on a purely local desktop session, so
	// treating it as an SSH marker is a classic false positive.
	for _, environ := range [][]string{
		{"SSH_AUTH_SOCK=/tmp/ssh-XXXX/agent.1"},
		{"SSH_AUTH_SOCK=/tmp/ssh-XXXX/agent.1", "SSH_AGENT_PID=123"},
	} {
		if got := runby.Detect(runby.WithEnviron(environ)); got.HasRemoteLayer(runby.RemoteSSH) {
			t.Fatalf("Detect(%v) reported SSH: %#v", environ, got.Remote)
		}
	}
}

func TestRemoteWSLIsProbable(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"WSL_DISTRO_NAME=Ubuntu",
		"WSL_INTEROP=/run/WSL/8_interop",
		"WSLENV=WT_SESSION::WT_PROFILE_ID",
	}))
	layer, ok := result.RemoteLayer(runby.RemoteWSL)
	if !ok || layer.SessionID != "Ubuntu" {
		t.Fatalf("Remote = %#v", result.Remote)
	}
	// WSL_DISTRO_NAME is not a documented stable contract and is absent for
	// root and systemd services, so presence never rises above probable.
	if layer.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q, want %q", layer.Confidence, runby.ConfidenceProbable)
	}
	// Either marker alone is enough, since either can be absent.
	if !runby.Detect(runby.WithEnviron([]string{"WSL_INTEROP=/run/WSL/1_interop"})).HasRemoteLayer(runby.RemoteWSL) {
		t.Fatal("WSL_INTEROP alone not detected")
	}
}

func TestRemoteCodespacesIsNotCI(t *testing.T) {
	// A codespace is an interactive development environment. It does not set
	// GITHUB_ACTIONS, which is what keeps it off the CI axis.
	result := runby.Detect(runby.WithEnviron([]string{
		"CODESPACES=true",
		"CODESPACE_NAME=fluffy-space-parakeet",
		"GITHUB_REPOSITORY=owner/repo",
		"GITHUB_SERVER_URL=https://github.com",
	}))

	layer, ok := result.RemoteLayer(runby.RemoteCodespaces)
	if !ok || layer.SessionID != "fluffy-space-parakeet" {
		t.Fatalf("Remote = %#v", result.Remote)
	}
	if result.IsCI() {
		t.Fatalf("IsCI() = true, want false: %#v", result.CI)
	}
}

func TestRemoteGitpodIsNotCI(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"GITPOD_WORKSPACE_ID=ironpark-runby-abc123",
		"GITPOD_WORKSPACE_URL=https://example.gitpod.io",
		"GITPOD_REPO_ROOT=/workspace/runby",
	}))
	layer, ok := result.RemoteLayer(runby.RemoteGitpod)
	if !ok || layer.SessionID != "ironpark-runby-abc123" {
		t.Fatalf("Remote = %#v", result.Remote)
	}
	if result.IsCI() {
		t.Fatalf("IsCI() = true, want false: %#v", result.CI)
	}
}

func TestRemoteDevContainerIsProbable(t *testing.T) {
	// The Dev Containers specification mandates no marker; REMOTE_CONTAINERS
	// is a VS Code implementation detail and DEVCONTAINER a deferred
	// convention, so neither can be definite.
	for _, entry := range []string{"REMOTE_CONTAINERS=true", "DEVCONTAINER=true"} {
		result := runby.Detect(runby.WithEnviron([]string{entry}))
		layer, ok := result.RemoteLayer(runby.RemoteDevContainer)
		if !ok {
			t.Fatalf("%s not detected: %#v", entry, result.Remote)
		}
		if layer.Confidence != runby.ConfidenceProbable {
			t.Fatalf("%s gave Confidence = %q", entry, layer.Confidence)
		}
	}
}

func TestRemoteLayersCoexist(t *testing.T) {
	// SSH into a Codespace running tmux is three concurrent layers, not a
	// precedence contest.
	result := runby.Detect(runby.WithEnviron([]string{
		"SSH_CONNECTION=203.0.113.5 54321 198.51.100.9 22",
		"CODESPACES=true",
		"CODESPACE_NAME=fluffy",
		"TMUX=/tmp/tmux-1000/default,9,0",
		"TERM_PROGRAM=ghostty",
	}))

	if len(result.Remote) != 3 {
		t.Fatalf("len(Remote) = %d, want 3: %#v", len(result.Remote), result.Remote)
	}
	for _, want := range []runby.RemotePlatform{runby.RemoteTmux, runby.RemoteSSH, runby.RemoteCodespaces} {
		if !result.HasRemoteLayer(want) {
			t.Fatalf("missing %q: %#v", want, result.Remote)
		}
	}
	// The multiplexer is what downgrades the terminal, not the other layers.
	mux, ok := result.Multiplexer()
	if !ok || mux.Platform != runby.RemoteTmux {
		t.Fatalf("Multiplexer() = %#v", mux)
	}
	if result.Terminal.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Terminal.Confidence = %q", result.Terminal.Confidence)
	}
}

func TestRemoteSSHAloneDoesNotDowngradeTerminal(t *testing.T) {
	// SSH means the terminal may be on another machine, but it does not make
	// the evidence stale the way a multiplexer does, so confidence stands and
	// the caveat is expressed by the SSH layer itself.
	result := runby.Detect(runby.WithEnviron([]string{
		"SSH_CONNECTION=203.0.113.5 54321 198.51.100.9 22",
		"TERM_PROGRAM=ghostty",
	}))
	if result.Terminal.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Confidence = %q, want %q", result.Terminal.Confidence, runby.ConfidenceDefinite)
	}
	if !result.HasRemoteLayer(runby.RemoteSSH) {
		t.Fatalf("Remote = %#v", result.Remote)
	}
}

func TestRemoteEvidenceIsNamesOnly(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"SSH_CONNECTION=203.0.113.5 54321 198.51.100.9 22",
		"SSH_TTY=/dev/pts/0",
		"SSH_AUTH_SOCK=/tmp/agent.1",
	}))
	layer, _ := result.RemoteLayer(runby.RemoteSSH)
	want := []string{"SSH_CONNECTION", "SSH_TTY"}
	if !reflect.DeepEqual(layer.Evidence, want) {
		t.Fatalf("Evidence = %#v, want %#v", layer.Evidence, want)
	}
}

func TestWithRemoteDrivers(t *testing.T) {
	driver := runby.RemoteDriver{
		Platform: "acme-vpn",
		Detect: func(env runby.Env) (runby.Remote, bool) {
			id, ok := runby.Value(env, "ACME_VPN_SESSION")
			if !ok {
				return runby.Remote{}, false
			}
			return runby.Remote{SessionID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_VPN_SESSION")}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"ACME_VPN_SESSION=v-1", "TMUX=/tmp/t,1,0"}),
		runby.WithRemoteDrivers(driver),
	)
	layer, ok := result.RemoteLayer("acme-vpn")
	if !ok || layer.SessionID != "v-1" || layer.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Remote = %#v", result.Remote)
	}
	// A driver that names no Kind must not be mistaken for a multiplexer.
	if layer.Kind != runby.RemoteKindUnknown || layer.IsMultiplexer() {
		t.Fatalf("Kind = %q", layer.Kind)
	}
	if mux, _ := result.Multiplexer(); mux.Platform != runby.RemoteTmux {
		t.Fatalf("Multiplexer() = %#v", mux)
	}

	disabled := runby.Detect(runby.WithEnviron([]string{"TMUX=/tmp/t,1,0"}), runby.WithOnlyDrivers())
	if disabled.IsRemote() {
		t.Fatalf("Remote = %#v, want detection disabled", disabled.Remote)
	}
}

func TestRemotePlatformsAndKinds(t *testing.T) {
	platforms := runby.RemotePlatforms()
	if len(platforms) == 0 || platforms[0] != runby.RemoteTmux {
		t.Fatalf("RemotePlatforms() = %#v", platforms)
	}
	for _, p := range platforms {
		if p.Kind() == runby.RemoteKindUnknown {
			t.Fatalf("%q has no Kind", p)
		}
	}
	if runby.RemoteUnknown.Kind() != runby.RemoteKindUnknown {
		t.Fatalf("RemoteUnknown.Kind() = %q", runby.RemoteUnknown.Kind())
	}
	if runby.RemotePlatform("").String() != "unknown" {
		t.Fatalf(`RemotePlatform("").String() = %q`, runby.RemotePlatform("").String())
	}
}

func TestRemoteMoshAndContainersAreNotDetectable(t *testing.T) {
	// Mosh removes nothing because MOSH_KEY never enters the remote shell's
	// environment, and no other MOSH_* variable reaches it either. A Mosh
	// session is therefore indistinguishable from a plain login.
	mosh := runby.Detect(runby.WithEnviron([]string{"TERM=xterm-256color"}))
	if mosh.IsRemote() {
		t.Fatalf("Remote = %#v, want empty", mosh.Remote)
	}

	// SSH_CONNECTION does survive into a Mosh session, but it describes the
	// short-lived bootstrap connection rather than the live UDP one. It is
	// still reported as SSH, which is the honest reading of the evidence.
	bootstrap := runby.Detect(runby.WithEnviron([]string{"SSH_CONNECTION=203.0.113.5 1 198.51.100.9 22"}))
	if !bootstrap.HasRemoteLayer(runby.RemoteSSH) {
		t.Fatalf("Remote = %#v", bootstrap.Remote)
	}

	// Docker and Podman set no identifying variable, so a bare container is
	// invisible to this package. HOSTNAME being a short hex string is a
	// heuristic, not evidence, and is deliberately not used.
	container := runby.Detect(runby.WithEnviron([]string{"HOSTNAME=3f2a1b9c4d5e"}))
	if container.IsRemote() {
		t.Fatalf("Remote = %#v, want empty", container.Remote)
	}
}
