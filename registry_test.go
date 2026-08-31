package runby_test

import (
	"testing"

	"github.com/ironpark/runby"
)

// This file registers drivers from init, which is exactly what a third-party
// driver module does and what a blank import triggers. Registering here rather
// than inside a test is deliberate: it is the only way to be sure the drivers
// are in place before anything calls Current, and it makes the test binary a
// faithful stand-in for a program that blank-imports a driver.
//
// The markers are variables nothing else in this repository sets, so these
// drivers never match in another test's fixture. Neither declares Executables,
// so they stay out of the ancestor labels that TestExecutablesCoverEveryProduct
// checks against the built-in products.
func init() {
	runby.Register(
		runby.AgentDriver{
			Agent:  "acme-orchestrator",
			Kind:   runby.KindOrchestrator,
			Models: runby.ModelsDelegated,
			Detect: func(env runby.Env) (runby.Detection, bool) {
				id, ok := runby.Value(env, "ACME_REGISTERED_RUN_ID")
				if !ok {
					return runby.Detection{}, false
				}
				return runby.Detection{
					AgentID: id,
					Axis:    runby.Axis{Evidence: runby.PresentNames(env, "ACME_REGISTERED_RUN_ID")},
				}, true
			},
		},
		// A Level1 harness, used to check that the ladder decides the order
		// rather than whoever registered first.
		runby.AgentDriver{
			Agent:  "acme-harness",
			Kind:   runby.KindHarness,
			Models: runby.ModelsFirstParty,
			Detect: func(env runby.Env) (runby.Detection, bool) {
				if _, ok := runby.Value(env, "ACME_REGISTERED_HARNESS"); !ok {
					return runby.Detection{}, false
				}
				return runby.Detection{
					Axis: runby.Axis{Evidence: []string{"ACME_REGISTERED_HARNESS"}},
				}, true
			},
		},
		runby.RunnerDriver{
			Tool: "acme-task",
			Kind: runby.RunnerKindScript,
			Detect: func(env runby.Env) (runby.Runner, bool) {
				task, ok := runby.Value(env, "ACME_REGISTERED_TASK")
				if !ok {
					return runby.Runner{}, false
				}
				return runby.Runner{
					Task: task,
					Axis: runby.Axis{Evidence: []string{"ACME_REGISTERED_TASK"}},
				}, true
			},
		},
	)
}

// TestRegisteredDriversReachDetect is the promise a blank import makes: a
// driver in someone else's module participates in a plain Detect, with no
// option threaded through the call.
func TestRegisteredDriversReachDetect(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"ACME_REGISTERED_RUN_ID=r1",
		"ACME_REGISTERED_TASK=build",
	}))

	layer, ok := result.Layer("acme-orchestrator")
	if !ok {
		t.Fatalf("the registered agent driver did not run: %v", result.Layers)
	}
	if layer.AgentID != "r1" {
		t.Errorf("agent id = %q, want r1", layer.AgentID)
	}
	// Detect fills the identity and the ladder for a registered driver exactly
	// as it does for a built-in one.
	if layer.Kind != runby.KindOrchestrator || layer.Level != runby.Level3 {
		t.Errorf("kind = %s, level = %s", layer.Kind, layer.Level)
	}

	runner, ok := result.RunnerBy("acme-task")
	if !ok {
		t.Fatalf("the registered runner driver did not run: %v", result.Runner)
	}
	if runner.Task != "build" || runner.Kind != runby.RunnerKindScript {
		t.Errorf("task = %q, kind = %s", runner.Task, runner.Kind)
	}
}

// TestLadderOrdersRegisteredAgents is why registration cannot simply prepend.
// Package initialization order is not something a driver author controls, so a
// registered Level1 harness must not displace a built-in Level3 orchestrator
// as the primary layer.
func TestLadderOrdersRegisteredAgents(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"ACME_REGISTERED_HARNESS=1",
		"PASEO_AGENT_ID=p1",
	}))
	if len(result.Layers) != 2 {
		t.Fatalf("got %d layers, want 2: %v", len(result.Layers), result.Layers)
	}
	if got := result.Agent(); got != runby.AgentPaseo {
		t.Errorf("primary = %s, want paseo: a Level1 harness outranked a Level3 orchestrator", got)
	}
	if result.Chain() != "paseo>acme-harness" {
		t.Errorf("chain = %q, want paseo>acme-harness", result.Chain())
	}

	// And a registered orchestrator does come first, which is the case the
	// option documentation has always promised.
	outer := runby.Detect(runby.WithEnviron([]string{
		"ACME_REGISTERED_RUN_ID=r1",
		"CODEX_THREAD_ID=t1",
	}))
	if outer.Chain() != "acme-orchestrator>codex" {
		t.Errorf("chain = %q, want acme-orchestrator>codex", outer.Chain())
	}
}

// TestWithOnlyIgnoresRegistry pins the escape hatch. A test that needs a fixed
// driver set must not be at the mercy of what a blank import elsewhere in the
// build has registered.
func TestWithOnlyIgnoresRegistry(t *testing.T) {
	environ := []string{"ACME_REGISTERED_RUN_ID=r1", "ACME_REGISTERED_TASK=build"}

	agents := runby.Detect(runby.WithEnviron(environ), runby.WithOnlyDrivers())
	if agents.IsAgent() {
		t.Errorf("WithOnlyDrivers() still ran a registered driver: %v", agents.Layers)
	}
	runners := runby.Detect(runby.WithEnviron(environ), runby.WithOnlyDrivers())
	if runners.HasRunner() {
		t.Errorf("WithOnlyDrivers() still ran a registered driver: %v", runners.Runner)
	}
}

// TestIdentityListsStayBuiltIn holds a boundary that TestSlugsMatchDocs depends
// on: every name these return has a research document in this repository, so a
// third-party driver must not appear in them.
func TestIdentityListsStayBuiltIn(t *testing.T) {
	for _, agent := range runby.Agents() {
		if agent == "acme-orchestrator" || agent == "acme-harness" {
			t.Errorf("Agents() returned the registered %s", agent)
		}
	}
	for _, tool := range runby.RunnerTools() {
		if tool == "acme-task" {
			t.Error("RunnerTools() returned the registered acme-task")
		}
	}
}

// TestRegisteredIdentityAnswersLikeBuiltin closes the gap that made this worth
// doing: a registered agent must answer Kind, Models, and Level the same way a
// built-in one does, so that Agent.Level and Detection.Level cannot disagree
// about the same agent.
func TestRegisteredIdentityAnswersLikeBuiltin(t *testing.T) {
	const acme = runby.Agent("acme-orchestrator")
	if got := acme.Kind(); got != runby.KindOrchestrator {
		t.Errorf("Kind() = %s, want orchestrator", got)
	}
	if got := acme.Models(); got != runby.ModelsDelegated {
		t.Errorf("Models() = %s, want delegated", got)
	}
	if got := acme.Level(); got != runby.Level3 {
		t.Errorf("Level() = %s, want l3", got)
	}

	// And the answer matches what a detection of it carries.
	result := runby.Detect(runby.WithEnviron([]string{"ACME_REGISTERED_RUN_ID=r1"}))
	layer, ok := result.Layer(acme)
	if !ok {
		t.Fatal("the registered driver did not match")
	}
	if layer.Level != acme.Level() || layer.Kind != acme.Kind() {
		t.Errorf("Detection says (%s, %s) but Agent says (%s, %s)",
			layer.Kind, layer.Level, acme.Kind(), acme.Level())
	}

	if got := runby.RunnerTool("acme-task").Kind(); got != runby.RunnerKindScript {
		t.Errorf("RunnerTool.Kind() = %s, want script", got)
	}

	// An identity nobody registered still reads as unknown.
	if got := runby.Agent("never-registered").Level(); got != runby.LevelUnknown {
		t.Errorf("unknown agent Level() = %s", got)
	}
}
