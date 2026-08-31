package runby

import (
	"strings"
	"testing"

	"github.com/ironpark/runby/internal/proc"
)

// unmatchableProducts lists the products whose driver deliberately names no
// executable, each with the reason. This pins the gap list: adding a product
// forces a decision rather than silently leaving it uncorroborated forever.
var unmatchableProducts = map[string]string{
	// Agents.
	"antigravity-2": "no executable name verified against an official source",
	"orca":          "orca is also the GNOME screen reader, so the name would mislabel an unrelated process",

	// Terminals.
	"apple-terminal": "the binary is named Terminal, too generic to match safely",
	"vte":            "a widget library, never a process of its own",

	// Remote layers. These describe where the process runs, not an ancestor
	// process, so there is nothing in the chain to match.
	"wsl":               "a kernel interop layer, not an ancestor process",
	"github-codespaces": "a hosting environment, not an ancestor process",
	"gitpod":            "a hosting environment, not an ancestor process",
	"devcontainers":     "a container spec, not an ancestor process",
	// The terminal's parent is a pty host rather than the app, and its name
	// varies by platform and packaging across Code, Code Helper, code,
	// electron, and node. The last three are generic enough to mislabel
	// unrelated processes, and the rest are unverified.
	"vscode": "the pty host's executable name is unverified, and its candidates are generic",
	// Every IntelliJ platform IDE ships this terminal under its own binary
	// name, and a Toolbox or JVM launch can appear as java. Covering the
	// family would mean asserting a list of names no official source verifies.
	"jetbrains": "one binary name per IDE, none verified, and java is too generic",
	// npm scripts run as node, which would mislabel every unrelated node
	// process. pnpm and bun run under their own names and do name them.
	"npm": "npm scripts run as node, which is too generic to match safely",
	// pre-commit is a Python entry point, so the ancestor is python or python3.
	"pre-commit": "a Python entry point, so the ancestor name is python",
}

// builtinLabels is the mapping Detect builds from the built-in drivers, which
// are the only place a product's executable names live.
func builtinLabels() executableLabels {
	return options{
		agentDrivers:    builtinAgentDrivers,
		terminalDrivers: builtinTerminalDrivers,
		remoteDrivers:   builtinRemoteDrivers,
		runnerDrivers:   builtinRunnerDrivers,
	}.executableLabels()
}

func TestExecutablesCoverEveryProduct(t *testing.T) {
	labels := builtinLabels()

	matched := map[string]bool{}
	for _, p := range labels {
		matched[string(p.Agent)] = true
		matched[string(p.Terminal)] = true
		matched[string(p.Remote)] = true
		matched[string(p.Runner)] = true
	}

	var products []string
	for _, a := range Agents() {
		products = append(products, string(a))
	}
	for _, x := range TerminalPrograms() {
		products = append(products, string(x))
	}
	for _, x := range RemotePlatforms() {
		products = append(products, string(x))
	}
	for _, x := range RunnerTools() {
		products = append(products, string(x))
	}

	seen := map[string]bool{}
	for _, product := range products {
		seen[product] = true
		_, exempt := unmatchableProducts[product]
		if !matched[product] && !exempt {
			t.Errorf("%s names no executables: add them to its driver so a live "+
				"ancestor can corroborate it, or list it in unmatchableProducts "+
				"with the reason", product)
		}
		if matched[product] && exempt {
			t.Errorf("%s is listed in unmatchableProducts but its driver names "+
				"an executable; drop the exemption", product)
		}
	}

	for product := range unmatchableProducts {
		if !seen[product] {
			t.Errorf("unmatchableProducts lists %q, which is no longer a product", product)
		}
	}
}

func TestExecutableKeysAreNormalized(t *testing.T) {
	// internal/proc lowercases names and strips any .exe suffix, so a key not
	// already in that form can never be hit.
	labels := builtinLabels()

	for name := range labels {
		if got := strings.TrimSuffix(strings.ToLower(name), ".exe"); got != name {
			t.Errorf("key %q normalizes to %q and can never match", name, got)
		}
		if strings.ContainsAny(name, `/\`) {
			t.Errorf("key %q contains a path separator; keys are base names", name)
		}
	}
}

func TestTruncatedNamesMatchByPrefix(t *testing.T) {
	// Linux caps /proc/<pid>/comm at CommLimit bytes, and comm is the only
	// readable source for a process owned by another user — the common case
	// for a terminal ancestor. gnome-terminal-server is 21 characters, so
	// without prefix matching it could never be corroborated on Linux.
	labels := builtinLabels()

	full := "gnome-terminal-server"
	truncated := full[:proc.CommLimit]

	if label := labels.find(procInfo{Name: truncated, Truncated: true}); label.Terminal != TerminalGNOMETerminal {
		t.Fatalf("truncated %q did not resolve to GNOME Terminal: %#v", truncated, label)
	}
	// The same prefix without the flag must not match: a process really named
	// "gnome-terminal" is not the server.
	if label := labels.find(procInfo{Name: truncated}); label != (Process{}) {
		t.Errorf("%q matched without the truncation flag: %#v", truncated, label)
	}
	// An ambiguous prefix names no single product, so it labels nothing.
	ambiguous := executableLabels{"aaa-one": {Agent: AgentCodex}, "aaa-two": {Agent: AgentAmp}}
	if label := ambiguous.find(procInfo{Name: "aaa", Truncated: true}); label != (Process{}) {
		t.Errorf("an ambiguous prefix produced a label: %#v", label)
	}
}
