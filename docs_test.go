package runby_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ironpark/runby"
)

// docFields reads the YAML front matter of a research document as a flat map.
// The front matter is a fixed set of scalar keys, so it is read line by line
// rather than by pulling in a YAML parser for one test. Nested list items are
// skipped, so executes_agents does not become a field of its own.
func docFields(t *testing.T, path string) map[string]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	fields := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" && len(fields) > 0 {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(key, " ") || strings.HasPrefix(key, "-") {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return fields
}

// docSlug reads the slug recorded in a research document's front matter.
func docSlug(t *testing.T, path string) string {
	t.Helper()
	slug := docFields(t, path)["slug"]
	if slug == "" {
		t.Fatalf("%s has no slug in its front matter", path)
	}
	return slug
}

// TestSlugsMatchDocs keeps every identifier this package reports tied to the
// research document that justifies it. A slug is part of the serialized output,
// so a silent rename would break both consumers and the documentation link.
func TestSlugsMatchDocs(t *testing.T) {
	for _, test := range []struct {
		name  string
		dir   string
		slugs []string
		// exempt lists slugs whose document lives elsewhere or covers a
		// family rather than one product.
		exempt map[string]string
	}{
		{
			name: "agents",
			dir:  "docs/research/agents",
			slugs: func() (s []string) {
				for _, a := range runby.Agents() {
					s = append(s, string(a))
				}
				return
			}(),
		},
		{
			name: "ci",
			dir:  "docs/research/ci",
			slugs: func() (s []string) {
				for _, p := range runby.CIProviders() {
					s = append(s, string(p))
				}
				return
			}(),
			exempt: map[string]string{
				string(runby.CIProviderGeneric): "the bare CI convention belongs to no product",
			},
		},
		{
			name: "remote",
			dir:  "docs/research/remote",
			slugs: func() (s []string) {
				for _, p := range runby.RemotePlatforms() {
					s = append(s, string(p))
				}
				return
			}(),
		},
		{
			name: "runners",
			dir:  "docs/research/runners",
			slugs: func() (s []string) {
				for _, t := range runby.RunnerTools() {
					s = append(s, string(t))
				}
				return
			}(),
		},
		{
			name: "terminals",
			dir:  "docs/research/terminals",
			slugs: func() (s []string) {
				for _, p := range runby.TerminalPrograms() {
					s = append(s, string(p))
				}
				return
			}(),
			exempt: map[string]string{
				string(runby.TerminalVTE): "VTE names a family, not a product",
				string(runby.TerminalZed): "researched in docs/research/agents/zed-agent.md",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if len(test.slugs) == 0 {
				t.Fatal("no slugs to check")
			}
			for _, slug := range test.slugs {
				if reason, ok := test.exempt[slug]; ok {
					t.Logf("skipping %q: %s", slug, reason)
					continue
				}
				path := filepath.Join(test.dir, slug+".md")
				if _, err := os.Stat(path); err != nil {
					t.Errorf("%q has no research document at %s", slug, path)
					continue
				}
				if got := docSlug(t, path); got != slug {
					t.Errorf("%s records slug %q, want %q", path, got, slug)
				}
			}
		})
	}
}

// TestKindsMatchDocs holds every registered agent's classification to the
// research document that justifies it.
//
// Kind and ModelSource are facts about a product that no environment can
// supply, so they are asserted in the driver table by hand and recorded in the
// research front matter by hand, in two files that nothing otherwise connects.
// They did drift: Antigravity 2.0 was documented as an orchestrator and
// registered as a harness. This test is what makes that a failure rather than
// a silent misreport, and it is why Kind's documentation can claim to mirror
// product_type.
func TestKindsMatchDocs(t *testing.T) {
	productTypes := map[string]runby.Kind{
		"agent_orchestrator": runby.KindOrchestrator,
		"agent_harness":      runby.KindHarness,
	}
	modelSources := map[string]runby.ModelSource{
		"first-party":  runby.ModelsFirstParty,
		"multi-vendor": runby.ModelsMultiVendor,
		"delegated":    runby.ModelsDelegated,
	}

	// The driver table is the single place a product's classification lives,
	// and BuiltinDrivers exposes it, so this needs no other access to it.
	for _, driver := range runby.BuiltinDrivers() {
		agent, ok := driver.(runby.AgentDriver)
		if !ok {
			continue
		}
		t.Run(string(agent.Agent), func(t *testing.T) {
			path := filepath.Join("docs/research/agents", string(agent.Agent)+".md")
			fields := docFields(t, path)

			kind, ok := productTypes[fields["product_type"]]
			if !ok {
				t.Fatalf("%s records product_type %q, which maps to no Kind", path, fields["product_type"])
			}
			if kind != agent.Kind {
				t.Errorf("%s records product_type %q (%s), but the driver says %s",
					path, fields["product_type"], kind, agent.Kind)
			}

			models, ok := modelSources[fields["model_source"]]
			if !ok {
				t.Fatalf("%s records model_source %q, which maps to no ModelSource", path, fields["model_source"])
			}
			if models != agent.Models {
				t.Errorf("%s records model_source %q, but the driver says %s",
					path, fields["model_source"], agent.Models)
			}
		})
	}
}
