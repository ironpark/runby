package runby_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ironpark/runby"
)

// docSlug reads the slug recorded in a research document's YAML front matter.
func docSlug(t *testing.T, path string) string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if slug, ok := strings.CutPrefix(line, "slug:"); ok {
			return strings.TrimSpace(slug)
		}
		if line == "---" && scanner.Text() == "---" {
			continue
		}
	}
	t.Fatalf("%s has no slug in its front matter", path)
	return ""
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
			dir:  "docs/agents",
			slugs: func() (s []string) {
				for _, a := range runby.Agents() {
					s = append(s, string(a))
				}
				return
			}(),
		},
		{
			name: "ci",
			dir:  "docs/ci",
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
			name: "terminals",
			dir:  "docs/terminals",
			slugs: func() (s []string) {
				for _, p := range runby.TerminalPrograms() {
					s = append(s, string(p))
				}
				return
			}(),
			exempt: map[string]string{
				string(runby.TerminalVTE): "VTE names a family, not a product",
				string(runby.TerminalZed): "researched in docs/agents/zed-agent.md",
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
