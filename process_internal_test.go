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
}

// builtinLabels is the mapping Detect builds from the built-in drivers, which
// are the only place a product's executable names live.
func builtinLabels() executableLabels {
	return options{
		agentDrivers:    builtinAgentDrivers,
		terminalDrivers: builtinTerminalDrivers,
		remoteDrivers:   builtinRemoteDrivers,
	}.executableLabels()
}

func TestExecutablesCoverEveryProduct(t *testing.T) {
	labels := builtinLabels()

	matched := map[string]bool{}
	for _, p := range labels {
		matched[string(p.Agent)] = true
		matched[string(p.Terminal)] = true
		matched[string(p.Remote)] = true
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

	if label, ok := labels.find(procInfo{Name: truncated, Truncated: true}); !ok ||
		label.Terminal != TerminalGNOMETerminal {
		t.Fatalf("truncated %q did not resolve to GNOME Terminal: %#v", truncated, label)
	}
	// The same prefix without the flag must not match: a process really named
	// "gnome-terminal" is not the server.
	if _, ok := labels.find(procInfo{Name: truncated}); ok {
		t.Errorf("%q matched without the truncation flag", truncated)
	}
	// An ambiguous prefix names no single product, so it labels nothing.
	ambiguous := executableLabels{"aaa-one": {Agent: AgentCodex}, "aaa-two": {Agent: AgentAmp}}
	if _, ok := ambiguous.find(procInfo{Name: "aaa", Truncated: true}); ok {
		t.Error("an ambiguous prefix produced a label")
	}
}
