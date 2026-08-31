package runby

import "strconv"

// CIProvider identifies a continuous integration platform.
type CIProvider string

const (
	CIProviderUnknown CIProvider = "unknown"
	// CIProviderGeneric is a platform that advertises only the conventional
	// CI variable, without a marker this package recognizes.
	CIProviderGeneric CIProvider = "generic"

	CIProviderGitHubActions  CIProvider = "github-actions"
	CIProviderForgejo        CIProvider = "forgejo-runner"
	CIProviderGitLab         CIProvider = "gitlab-ci"
	CIProviderCircleCI       CIProvider = "circleci"
	CIProviderTravis         CIProvider = "travis-ci"
	CIProviderBuildkite      CIProvider = "buildkite"
	CIProviderAzurePipelines CIProvider = "azure-pipelines"
	CIProviderBitbucket      CIProvider = "bitbucket-pipelines"
	CIProviderJenkins        CIProvider = "jenkins"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (p CIProvider) String() string {
	if p == "" {
		return string(CIProviderUnknown)
	}
	return string(p)
}

// CIProviders returns every supported provider in detection precedence order.
// CIProviderGeneric is last because it is the fallback.
func CIProviders() []CIProvider {
	providers := make([]CIProvider, 0, len(builtinCIDetectors))
	for _, detector := range builtinCIDetectors {
		providers = append(providers, detector.Provider())
	}
	return providers
}

// CI describes the continuous integration run that owns this process.
//
// CI is a separate axis from the agent Layers, not one of them. An agent can
// run inside CI: Claude Code invoked from a GitHub Actions workflow produces
// both a KindHarness layer and a CI result. Kind answers who requested the
// command; CI answers where it runs.
type CI struct {
	Detected   bool       `json:"detected"`
	Provider   CIProvider `json:"provider"`
	Confidence Confidence `json:"confidence"`

	// PipelineID is the platform's identifier for the whole run, called a
	// run, build, or pipeline depending on the platform.
	PipelineID string `json:"pipeline_id,omitempty"`
	// BuildNumber is the human-facing counter shown in the platform's UI. It
	// is not unique across projects.
	BuildNumber string `json:"build_number,omitempty"`
	// JobID identifies the individual job or step within the pipeline.
	JobID string `json:"job_id,omitempty"`
	// JobName is the configured name of that job or step.
	JobName string `json:"job_name,omitempty"`
	// Attempt is the 1-based attempt number for retried runs, or 0 when the
	// platform does not advertise one. Platforms that expose a 0-based retry
	// count are normalized to this 1-based form.
	Attempt int `json:"attempt,omitempty"`
	// Trigger is what started the run, in the platform's own vocabulary, such
	// as "push", "schedule", or "pull_request". It is empty when the platform
	// does not advertise one.
	Trigger string `json:"trigger,omitempty"`
	// Runner names the machine or agent executing the job.
	Runner string `json:"runner,omitempty"`

	// Extra holds values that only one platform advertises, keyed by
	// "<provider-slug>.<name>".
	Extra map[string]string `json:"extra,omitempty"`

	// Evidence lists the environment variable names that produced this
	// result, sorted. Their values may be sensitive and are never copied.
	Evidence []string `json:"evidence"`
}

// CIDetector reports whether an environment shows a CI run of its platform.
// Implement it to detect a platform this package does not support, then pass
// it to Detect with WithCIDetectors.
type CIDetector interface {
	// Provider returns the platform this detector reports.
	Provider() CIProvider
	// Detect returns the CI result, or false if the environment holds no
	// evidence of this platform. Implementations must not retain env.
	Detect(env Env) (CI, bool)
}

// NewCIDetector adapts a function into a CIDetector.
func NewCIDetector(provider CIProvider, detect func(env Env) (CI, bool)) CIDetector {
	return funcCIDetector{provider: provider, detect: detect}
}

type funcCIDetector struct {
	provider CIProvider
	detect   func(Env) (CI, bool)
}

func (d funcCIDetector) Provider() CIProvider      { return d.provider }
func (d funcCIDetector) Detect(env Env) (CI, bool) { return d.detect(env) }

// IsCI reports whether this process is running in a CI job, using the cached
// Current result.
func IsCI() bool { return Current().CI.Detected }

// parseAttempt reads a 1-based attempt counter, adding offset first so that
// platforms exposing a 0-based retry count normalize to the same form.
// Non-numeric and negative results are reported as 0, meaning unknown.
func parseAttempt(env Env, name string, offset int) int {
	raw, ok := Value(env, name)
	if !ok {
		return 0
	}
	attempt, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	attempt += offset
	if attempt < 1 {
		return 0
	}
	return attempt
}
