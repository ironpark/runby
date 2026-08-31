package runby

import (
	"strings"
	"testing"
)

// unmatchableProducts lists the products that deliberately have no entry in
// the executables table, each with the reason. The table is a second place
// where a product must be registered, so this pins the gap list: adding a
// product forces a decision here rather than silently leaving it
// uncorroborated forever.
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

func TestExecutablesCoverEveryProduct(t *testing.T) {
	matched := map[string]bool{}
	for _, p := range executables {
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
			t.Errorf("%s has no executables entry: add one so a live ancestor "+
				"can corroborate it, or list it in unmatchableProducts with the reason",
				product)
		}
		if matched[product] && exempt {
			t.Errorf("%s is listed in unmatchableProducts but the executables "+
				"table matches it; drop the exemption", product)
		}
	}

	for product := range unmatchableProducts {
		if !seen[product] {
			t.Errorf("unmatchableProducts lists %q, which is no longer a product", product)
		}
	}
}

func TestExecutableKeysAreNormalized(t *testing.T) {
	// Lookups run every name through normalizeExecutable, so a key that is not
	// already in that form can never be hit.
	for name := range executables {
		if got := normalizeExecutable(name); got != name {
			t.Errorf("key %q normalizes to %q and can never match", name, got)
		}
		if strings.ContainsAny(name, `/\`) {
			t.Errorf("key %q contains a path separator; keys are base names", name)
		}
	}
}
