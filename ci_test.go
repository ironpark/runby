package runby_test

import (
	"reflect"
	"testing"

	"github.com/ironpark/runby"
)

func TestCINotDetected(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"PATH=/usr/bin"}))
	if result.IsCI() || result.CI.Detected {
		t.Fatalf("CI = %#v, want undetected", result.CI)
	}
	if result.CI.Provider != runby.CIProviderUnknown || result.CI.Confidence != runby.ConfidenceUnknown {
		t.Fatalf("CI = %#v", result.CI)
	}
}

func TestCIGitHubActions(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"CI=true",
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=1658821493",
		"GITHUB_RUN_NUMBER=3",
		"GITHUB_RUN_ATTEMPT=2",
		"GITHUB_JOB=build",
		"GITHUB_ACTION=__run_2",
		"GITHUB_EVENT_NAME=pull_request",
		"RUNNER_NAME=GitHub Actions 4",
		"RUNNER_ENVIRONMENT=github-hosted",
	}))

	ci := result.CI
	if !result.IsCI() || ci.Provider != runby.CIProviderGitHubActions {
		t.Fatalf("CI = %#v", ci)
	}
	if ci.PipelineID != "1658821493" || ci.BuildNumber != "3" || ci.JobID != "build" {
		t.Fatalf("identity = %#v", ci)
	}
	if ci.Attempt != 2 || ci.Trigger != "pull_request" || ci.Runner != "GitHub Actions 4" {
		t.Fatalf("run metadata = %#v", ci)
	}
	if ci.Extra["github-actions.action"] != "__run_2" || ci.Extra["github-actions.runner_environment"] != "github-hosted" {
		t.Fatalf("Extra = %#v", ci.Extra)
	}
	if ci.Confidence != runby.ConfidenceDefinite {
		t.Fatalf("Confidence = %q", ci.Confidence)
	}
}

func TestCIGitLab(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"CI=true",
		"GITLAB_CI=true",
		"CI_PIPELINE_ID=1000",
		"CI_PIPELINE_IID=7",
		"CI_JOB_ID=2000",
		"CI_JOB_NAME=test",
		"CI_PIPELINE_SOURCE=merge_request_event",
		"CI_RUNNER_ID=42",
		"CI_JOB_RETRY_COUNT=1",
	}))

	ci := result.CI
	if ci.Provider != runby.CIProviderGitLab || ci.PipelineID != "1000" || ci.JobID != "2000" {
		t.Fatalf("CI = %#v", ci)
	}
	if ci.BuildNumber != "7" || ci.JobName != "test" || ci.Runner != "42" {
		t.Fatalf("CI = %#v", ci)
	}
	if ci.Trigger != "merge_request_event" {
		t.Fatalf("Trigger = %q", ci.Trigger)
	}
	// CI_JOB_RETRY_COUNT counts retries from 0, so the second run is attempt 2.
	if ci.Attempt != 2 {
		t.Fatalf("Attempt = %d, want 2", ci.Attempt)
	}
}

func TestCIGitLabWithoutRetryCountLeavesAttemptUnknown(t *testing.T) {
	// CI_JOB_RETRY_COUNT only exists on GitLab 19.3 and later.
	result := runby.Detect(runby.WithEnviron([]string{"GITLAB_CI=true", "CI_JOB_ID=1"}))
	if result.CI.Attempt != 0 {
		t.Fatalf("Attempt = %d, want 0", result.CI.Attempt)
	}
}

func TestCIBuildkiteNormalizesRetryCount(t *testing.T) {
	// BUILDKITE_RETRY_COUNT is 0 on the first run, which is attempt 1.
	first := runby.Detect(runby.WithEnviron([]string{
		"BUILDKITE=true",
		"BUILDKITE_BUILD_ID=b-1",
		"BUILDKITE_JOB_ID=j-1",
		"BUILDKITE_RETRY_COUNT=0",
		"BUILDKITE_SOURCE=schedule",
		"BUILDKITE_AGENT_PID=1234",
	}))
	if first.CI.Provider != runby.CIProviderBuildkite || first.CI.Attempt != 1 {
		t.Fatalf("CI = %#v", first.CI)
	}
	if first.CI.Trigger != "schedule" {
		t.Fatalf("Trigger = %q", first.CI.Trigger)
	}
	// The agent PID identifies the long-lived agent daemon, not this process,
	// so it stays context rather than identity.
	if first.CI.Extra["buildkite.agent_pid"] != "1234" {
		t.Fatalf("Extra = %#v", first.CI.Extra)
	}

	retried := runby.Detect(runby.WithEnviron([]string{
		"BUILDKITE=true", "BUILDKITE_RETRY_COUNT=2",
	}))
	if retried.CI.Attempt != 3 {
		t.Fatalf("Attempt = %d, want 3", retried.CI.Attempt)
	}
}

func TestCIBitbucketTrimsUUIDBraces(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"BITBUCKET_BUILD_NUMBER=59",
		"BITBUCKET_PIPELINE_UUID={11d87b82-13c6-47a3-8e28-73a2bc378675}",
		"BITBUCKET_STEP_UUID={c1c4a4c1-2b1e-4a1e-9e5a-1f2b3c4d5e6f}",
		"BITBUCKET_STEP_RUN_NUMBER=1",
		"BITBUCKET_PR_ID=12",
	}))

	ci := result.CI
	if ci.Provider != runby.CIProviderBitbucket {
		t.Fatalf("Provider = %q", ci.Provider)
	}
	if ci.PipelineID != "11d87b82-13c6-47a3-8e28-73a2bc378675" {
		t.Fatalf("PipelineID = %q, want braces trimmed", ci.PipelineID)
	}
	if ci.JobID != "c1c4a4c1-2b1e-4a1e-9e5a-1f2b3c4d5e6f" {
		t.Fatalf("JobID = %q, want braces trimmed", ci.JobID)
	}
	if ci.BuildNumber != "59" || ci.Attempt != 1 || ci.Extra["bitbucket-pipelines.pr_id"] != "12" {
		t.Fatalf("CI = %#v", ci)
	}
}

func TestCIBitbucketNeedsASecondSignal(t *testing.T) {
	// BITBUCKET_BUILD_NUMBER alone is a bare identifier, not a marker.
	only := runby.Detect(runby.WithEnviron([]string{"BITBUCKET_BUILD_NUMBER=59"}))
	if only.CI.Provider == runby.CIProviderBitbucket {
		t.Fatalf("CI = %#v, want no Bitbucket match", only.CI)
	}
	// CI=true is accepted as that second signal when the UUID is absent.
	withCI := runby.Detect(runby.WithEnviron([]string{"BITBUCKET_BUILD_NUMBER=59", "CI=true"}))
	if withCI.CI.Provider != runby.CIProviderBitbucket {
		t.Fatalf("CI = %#v, want Bitbucket", withCI.CI)
	}
}

func TestCIJenkinsRequiresAJenkinsOwnedVariable(t *testing.T) {
	// BUILD_NUMBER and JOB_NAME are generic names that other tools also set,
	// so a Jenkins-owned variable must be present too.
	generic := runby.Detect(runby.WithEnviron([]string{"BUILD_NUMBER=17", "JOB_NAME=nightly"}))
	if generic.CI.Provider == runby.CIProviderJenkins {
		t.Fatalf("CI = %#v, want no Jenkins match", generic.CI)
	}
	// JENKINS_URL is only injected when an administrator configured the root
	// URL, so JENKINS_HOME alone must still be enough.
	home := runby.Detect(runby.WithEnviron([]string{
		"BUILD_NUMBER=17", "JENKINS_HOME=/var/lib/jenkins", "CI=true",
	}))
	if home.CI.Provider != runby.CIProviderJenkins {
		t.Fatalf("CI = %#v, want Jenkins", home.CI)
	}

	result := runby.Detect(runby.WithEnviron([]string{
		"JENKINS_URL=https://jenkins.example.com/",
		"BUILD_ID=17",
		"BUILD_NUMBER=17",
		"BUILD_TAG=jenkins-nightly-17",
		"JOB_NAME=nightly",
		"NODE_NAME=built-in",
	}))
	ci := result.CI
	if ci.Provider != runby.CIProviderJenkins || ci.PipelineID != "17" || ci.JobName != "nightly" {
		t.Fatalf("CI = %#v", ci)
	}
	if ci.Runner != "built-in" || ci.Extra["jenkins.build_tag"] != "jenkins-nightly-17" {
		t.Fatalf("CI = %#v", ci)
	}
	// Jenkins advertises no trigger type without plugins.
	if ci.Trigger != "" {
		t.Fatalf("Trigger = %q, want empty", ci.Trigger)
	}
}

func TestCIGenericFallback(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{"CI=true"}))
	if !result.IsCI() || result.CI.Provider != runby.CIProviderGeneric {
		t.Fatalf("CI = %#v", result.CI)
	}
	// The bare convention is not owned by any platform and local tooling sets
	// it too, so it never rises above probable.
	if result.CI.Confidence != runby.ConfidenceProbable {
		t.Fatalf("Confidence = %q, want %q", result.CI.Confidence, runby.ConfidenceProbable)
	}
	if result.CI.PipelineID != "" {
		t.Fatalf("PipelineID = %q, want empty", result.CI.PipelineID)
	}
	if got := runby.Detect(runby.WithEnviron([]string{"CI=false"})); got.IsCI() {
		t.Fatalf("CI=false detected: %#v", got.CI)
	}
}

func TestCISpecificPlatformBeatsGeneric(t *testing.T) {
	// Every platform also sets CI, so only the most specific one is reported.
	result := runby.Detect(runby.WithEnviron([]string{"CI=true", "CIRCLECI=true", "CIRCLE_PIPELINE_ID=p-1"}))
	if result.CI.Provider != runby.CIProviderCircleCI || result.CI.PipelineID != "p-1" {
		t.Fatalf("CI = %#v", result.CI)
	}
}

func TestCIAgentAndCIAreIndependentAxes(t *testing.T) {
	// An agent running inside CI populates both, which is why CI is not a
	// Layer.
	result := runby.Detect(runby.WithEnviron([]string{
		"CLAUDECODE=1",
		"CLAUDE_CODE_SESSION_ID=s-1",
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=99",
	}))

	if !result.IsAgent() || result.Agent() != runby.AgentClaudeCode {
		t.Fatalf("agent = %#v", result.Layers)
	}
	if !result.IsCI() || result.CI.Provider != runby.CIProviderGitHubActions {
		t.Fatalf("CI = %#v", result.CI)
	}
	// Chain describes who requested the command, not where it runs.
	if result.Chain() != "claude-code" {
		t.Fatalf("Chain() = %q", result.Chain())
	}
}

func TestCIAttemptIgnoresUnparsableValues(t *testing.T) {
	for _, value := range []string{"", "   ", "abc", "-3"} {
		result := runby.Detect(runby.WithEnviron([]string{"GITHUB_ACTIONS=true", "GITHUB_RUN_ATTEMPT=" + value}))
		if result.CI.Attempt != 0 {
			t.Fatalf("GITHUB_RUN_ATTEMPT=%q gave Attempt = %d, want 0", value, result.CI.Attempt)
		}
	}
}

func TestCIEvidenceIsNamesOnly(t *testing.T) {
	result := runby.Detect(runby.WithEnviron([]string{
		"GITHUB_ACTIONS=true",
		"GITHUB_RUN_ID=1",
		"GITHUB_EVENT_NAME=push",
	}))
	want := []string{"GITHUB_ACTIONS", "GITHUB_EVENT_NAME", "GITHUB_RUN_ID"}
	if !reflect.DeepEqual(result.CI.Evidence, want) {
		t.Fatalf("Evidence = %#v, want %#v", result.CI.Evidence, want)
	}
}

func TestCustomCIDriverOutranksTheGenericConvention(t *testing.T) {
	driver := runby.CIDriver{
		Provider: "acme-ci",
		Detect: func(env runby.Env) (runby.CI, bool) {
			id, ok := runby.Value(env, "ACME_CI_BUILD")
			if !ok {
				return runby.CI{}, false
			}
			return runby.CI{PipelineID: id, Axis: runby.Axis{Evidence: runby.PresentNames(env, "ACME_CI_BUILD")}}, true
		},
	}

	result := runby.Detect(
		runby.WithEnviron([]string{"CI=true", "ACME_CI_BUILD=b-9"}),
		// The CI axis takes the first match, so a driver that should outrank
		// the built-ins goes ahead of them rather than after.
		runby.WithOnlyDrivers(append([]runby.Driver{driver}, runby.BuiltinDrivers()...)...),
	)
	if result.CI.Provider != "acme-ci" || result.CI.PipelineID != "b-9" {
		t.Fatalf("CI = %#v", result.CI)
	}
	if result.CI.Confidence != runby.ConfidenceDefinite || !result.CI.Detected {
		t.Fatalf("CI = %#v", result.CI)
	}

	disabled := runby.Detect(runby.WithEnviron([]string{"CI=true"}), runby.WithOnlyDrivers())
	if disabled.IsCI() {
		t.Fatalf("CI = %#v, want detection disabled", disabled.CI)
	}
}

func TestCIProvidersAreOrderedAndGenericIsLast(t *testing.T) {
	providers := runby.CIProviders()
	if len(providers) < 2 {
		t.Fatalf("CIProviders() = %#v", providers)
	}
	// Forgejo leads because its GITHUB_* aliases would otherwise be claimed
	// by the GitHub Actions detector; see TestCIForgejoBeatsGitHubActions.
	if providers[0] != runby.CIProviderForgejo {
		t.Fatalf("providers[0] = %q", providers[0])
	}
	if last := providers[len(providers)-1]; last != runby.CIProviderGeneric {
		t.Fatalf("last provider = %q, want %q", last, runby.CIProviderGeneric)
	}
	if runby.CIProvider("").String() != "unknown" {
		t.Fatalf(`CIProvider("").String() = %q`, runby.CIProvider("").String())
	}
}

func TestCIForgejoBeatsGitHubActions(t *testing.T) {
	// Forgejo Runner v7+ mirrors every FORGEJO_* variable to a GITHUB_* alias,
	// so an environment carrying both must resolve to Forgejo.
	result := runby.Detect(runby.WithEnviron([]string{
		"CI=true",
		"FORGEJO_ACTIONS=true",
		"GITHUB_ACTIONS=true",
		"FORGEJO_RUN_ID=42", "GITHUB_RUN_ID=42",
		"FORGEJO_RUN_NUMBER=7", "FORGEJO_RUN_ATTEMPT=2",
		"FORGEJO_JOB=build", "GITHUB_JOB=build",
		"FORGEJO_EVENT_NAME=push",
		"FORGEJO_SERVER_URL=https://code.example.org",
		"FORGEJO_REPOSITORY=owner/repo",
	}))

	ci := result.CI
	if ci.Provider != runby.CIProviderForgejo {
		t.Fatalf("Provider = %q, want %q", ci.Provider, runby.CIProviderForgejo)
	}
	if ci.PipelineID != "42" || ci.BuildNumber != "7" || ci.JobID != "build" {
		t.Fatalf("identity = %#v", ci)
	}
	// FORGEJO_RUN_ATTEMPT already counts from 1, so it is used unchanged.
	if ci.Attempt != 2 || ci.Trigger != "push" {
		t.Fatalf("run metadata = %#v", ci)
	}
	if ci.Extra["forgejo-runner.server_url"] != "https://code.example.org" {
		t.Fatalf("Extra = %#v", ci.Extra)
	}
}

func TestCIOldForgejoRunnerIsIndistinguishableFromGitHub(t *testing.T) {
	// Runners older than v7 define only the GITHUB_* names. There is no
	// environment signal that separates them from GitHub Actions, so
	// reporting GitHub is the honest outcome rather than guessing Forgejo.
	result := runby.Detect(runby.WithEnviron([]string{
		"CI=true", "GITHUB_ACTIONS=true", "GITHUB_RUN_ID=42",
	}))
	if result.CI.Provider != runby.CIProviderGitHubActions {
		t.Fatalf("Provider = %q, want %q", result.CI.Provider, runby.CIProviderGitHubActions)
	}
}

func TestCIForgejoPrecedesGitHubInRegistry(t *testing.T) {
	providers := runby.CIProviders()
	forgejo, github := -1, -1
	for i, p := range providers {
		switch p {
		case runby.CIProviderForgejo:
			forgejo = i
		case runby.CIProviderGitHubActions:
			github = i
		}
	}
	if forgejo < 0 || github < 0 || forgejo > github {
		t.Fatalf("forgejo=%d github=%d in %v, want forgejo first", forgejo, github, providers)
	}
}
