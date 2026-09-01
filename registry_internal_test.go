package runby

import (
	"strings"
	"testing"
)

// The registry is process-wide, so the behaviours that would poison it for
// every other test — replacing a built-in, rejecting a duplicate, refusing a
// late registration — are exercised against the helpers directly rather than
// through Register.

// TestMergeReplacesBuiltins covers the reason a registered driver takes over
// rather than running beside a built-in of the same name: two drivers for one
// product would report the same product twice.
func TestMergeReplacesBuiltins(t *testing.T) {
	id := func(d AgentDriver) AgentName { return d.Agent }
	builtin := []AgentDriver{{Agent: AgentPaseo}, {Agent: AgentCodex}}

	t.Run("nothing registered", func(t *testing.T) {
		merged := merge(nil, builtin, id)
		if len(merged) != 2 || merged[0].Agent != AgentPaseo {
			t.Errorf("merged = %v", merged)
		}
	})

	t.Run("added alongside", func(t *testing.T) {
		merged := merge([]AgentDriver{{Agent: "acme"}}, builtin, id)
		if len(merged) != 3 {
			t.Fatalf("got %d drivers, want 3: %v", len(merged), merged)
		}
		// Registered first, then the built-ins in their own order.
		if merged[0].Agent != "acme" || merged[1].Agent != AgentPaseo {
			t.Errorf("merged = %v", merged)
		}
	})

	t.Run("replacing a built-in", func(t *testing.T) {
		merged := merge([]AgentDriver{{Agent: AgentCodex, Kind: KindOrchestrator}}, builtin, id)
		if len(merged) != 2 {
			t.Fatalf("got %d drivers, want 2 — the built-in codex should be gone: %v", len(merged), merged)
		}
		var seen int
		for _, driver := range merged {
			if driver.Agent == AgentCodex {
				seen++
				if driver.Kind != KindOrchestrator {
					t.Error("the built-in codex survived instead of the registered one")
				}
			}
		}
		if seen != 1 {
			t.Errorf("codex appears %d times, want 1", seen)
		}
	})

	// The built-in table must never be modified in place; callers hold it.
	if len(builtin) != 2 || builtin[1].Kind != "" {
		t.Errorf("merge mutated the built-in table: %v", builtin)
	}
}

func TestCheckUniquePanicsOnDuplicate(t *testing.T) {
	defer func() {
		message, ok := recover().(string)
		if !ok {
			t.Fatal("a duplicate registration did not panic")
		}
		if !strings.Contains(message, "acme") {
			t.Errorf("panic message does not name the product: %q", message)
		}
	}()
	checkUnique("agent", []AgentDriver{{Agent: "acme"}, {Agent: "acme"}},
		func(d AgentDriver) AgentName { return d.Agent })
}

func TestRegisterAfterCurrentPanics(t *testing.T) {
	// Current has not run in this test binary yet, so the flag is set by hand
	// and restored, rather than by calling Current and poisoning the cache for
	// every other test.
	was := detected.Swap(true)
	defer detected.Store(was)

	defer func() {
		message, ok := recover().(string)
		if !ok {
			t.Fatal("registering after Current did not panic")
		}
		if !strings.Contains(message, "init") {
			t.Errorf("panic message does not say what to do instead: %q", message)
		}
	}()
	Register(AgentDriver{Agent: "too-late"})
}

// TestSortByLadder covers the ordering rule on its own, including the driver
// that declares nothing and therefore cannot be placed.
func TestSortByLadder(t *testing.T) {
	l1 := AgentDriver{Agent: "l1", Kind: KindHarness, Models: ModelsFirstParty}
	l2 := AgentDriver{Agent: "l2", Kind: KindHarness, Models: ModelsMultiVendor}
	l3 := AgentDriver{Agent: "l3", Kind: KindOrchestrator, Models: ModelsDelegated}
	bare := AgentDriver{Agent: "bare"}

	sorted := sortByLadder([]AgentDriver{l1, bare, l2, l3})
	var got []string
	for _, driver := range sorted {
		got = append(got, string(driver.Agent))
	}
	want := []string{"l3", "l2", "l1", "bare"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}

	// Stable within a rung: two Level1 drivers keep the order they were given,
	// which is what lets the built-in table stay as written.
	other := AgentDriver{Agent: "l1-other", Kind: KindHarness, Models: ModelsFirstParty}
	pair := sortByLadder([]AgentDriver{other, l1})
	if pair[0].Agent != "l1-other" || pair[1].Agent != "l1" {
		t.Errorf("the sort was not stable within a rung: %v", pair)
	}
}

// TestBuiltinTableIsAlreadyInLadderOrder means sorting never rearranges the
// built-ins, so the table stays readable as the precedence contract.
func TestBuiltinTableIsAlreadyInLadderOrder(t *testing.T) {
	sorted := sortByLadder(builtinAgentDrivers)
	for i, driver := range builtinAgentDrivers {
		if sorted[i].Agent != driver.Agent {
			t.Fatalf("sorting moved %s; the table is no longer in ladder order", driver.Agent)
		}
	}
}
