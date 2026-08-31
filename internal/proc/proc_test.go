package proc_test

import (
	"os"
	"testing"

	"github.com/ironpark/runby/internal/proc"
)

func TestAncestorsStartAtTheRealParent(t *testing.T) {
	if !proc.Supported() {
		t.Skipf("ancestor reading is unsupported on %s", os.Getenv("GOOS"))
	}

	chain := proc.Ancestors()
	if len(chain) == 0 {
		t.Skip("no ancestor was readable")
	}

	// The kernel already tells us the immediate parent, so the first entry
	// must agree with it. On darwin this also proves the kinfo_proc offset
	// this package hardcodes is still correct.
	if got, want := chain[0].PID, os.Getppid(); got != want {
		t.Fatalf("chain[0].PID = %d, want os.Getppid() = %d", got, want)
	}
}

func TestAncestorsAreAWellFormedChain(t *testing.T) {
	if !proc.Supported() {
		t.Skip("ancestor reading is unsupported")
	}
	chain := proc.Ancestors()
	if len(chain) == 0 {
		t.Skip("no ancestor was readable")
	}

	seen := make(map[int]bool, len(chain))
	for i, p := range chain {
		if p.PID <= 0 {
			t.Fatalf("chain[%d].PID = %d, want positive", i, p.PID)
		}
		if seen[p.PID] {
			t.Fatalf("chain[%d].PID = %d repeats; the walk must not loop", i, p.PID)
		}
		seen[p.PID] = true

		// Each entry must name the next one, which is what makes this a
		// chain rather than a list of unrelated processes.
		if i+1 < len(chain) && chain[i+1].PID != p.PPID {
			t.Fatalf("chain[%d].PPID = %d but chain[%d].PID = %d", i, p.PPID, i+1, chain[i+1].PID)
		}
	}

	// The walk is bounded, so it cannot run away on a corrupt chain.
	if len(chain) > 24 {
		t.Fatalf("len(chain) = %d, want at most the depth limit", len(chain))
	}
}

func TestAncestorsAreStable(t *testing.T) {
	if !proc.Supported() {
		t.Skip("ancestor reading is unsupported")
	}
	first, second := proc.Ancestors(), proc.Ancestors()
	if len(first) != len(second) {
		t.Fatalf("len differs between calls: %d then %d", len(first), len(second))
	}
	for i := range first {
		if first[i].PID != second[i].PID {
			t.Fatalf("chain[%d].PID differs between calls: %d then %d", i, first[i].PID, second[i].PID)
		}
	}
}
