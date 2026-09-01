package runby

import "strconv"

// CIProvider identifies a continuous integration platform.
type CIProvider string

const (
	CIUnknown CIProvider = "unknown"
	// CIGeneric is a platform that advertises only the conventional
	// CI variable, without a marker this package recognizes.
	CIGeneric CIProvider = "generic"

	CIGitHubActions  CIProvider = "github-actions"
	CIForgejo        CIProvider = "forgejo-runner"
	CIGitLab         CIProvider = "gitlab-ci"
	CICircleCI       CIProvider = "circleci"
	CITravis         CIProvider = "travis-ci"
	CIBuildkite      CIProvider = "buildkite"
	CIAzurePipelines CIProvider = "azure-pipelines"
	CIBitbucket      CIProvider = "bitbucket-pipelines"
	CIJenkins        CIProvider = "jenkins"
)

// String returns the stable slug used across this package, its documentation,
// and its serialized output.
func (p CIProvider) String() string { return slug(p, CIUnknown) }

// CIProviders returns every built-in provider in detection precedence order.
// CIGeneric is last because it is the fallback. As with AgentNames,
// registered drivers are not included.
func CIProviders() []CIProvider {
	return mapSlice(builtinCIDrivers, func(d CIDriver) CIProvider { return d.Provider })
}

// CI describes the continuous integration run that owns this process.
//
// CI is a separate axis from the agent layers, not one of them. An agent can
// run inside CI: Claude Code invoked from a GitHub Actions workflow produces
// both a KindHarness layer and a CI result. Kind answers who requested the
// command; CI answers where it runs.
type CI struct {
	Detected bool       `json:"detected"`
	Provider CIProvider `json:"provider"`
	Axis

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
	// PullRequest reports that the run was triggered by a pull or merge request,
	// when the platform advertises it. It is false when the platform does not
	// expose a reliable signal for this distinction.
	PullRequest bool `json:"pull_request,omitempty"`
	// PullRequestID is the platform's identifier for that request, when
	// advertised. It may be empty even when PullRequest is true if the platform
	// only exposes the request kind.
	PullRequestID string `json:"pull_request_id,omitempty"`
}

// CIDriver detects one CI platform. It is the unit of extension for this
// axis: the built-in platforms are declared as drivers, and a platform this
// package does not support is added by passing another to Detect with
// Register or WithOnlyDrivers.
//
// Unlike the other axes a CI driver names no executables. A CI run is a job on
// a runner rather than a process this one descends from, so there is nothing
// in the ancestor chain to corroborate it against.
type CIDriver struct {
	// Provider identifies the platform this driver reports. Detect fills it
	// into every CI the driver returns, so Detect need not repeat it.
	Provider CIProvider
	// Detect returns the CI result, or false when the environment holds no
	// evidence of this platform. It must not retain env. Provider, Detected,
	// and a missing Confidence are filled in by Detect.
	Detect func(env Env) (CI, bool)
}

// parseAttempt reads a 1-based attempt counter, adding offset first so that
// platforms exposing a 0-based retry count normalize to the same form.
// Non-numeric and negative results are reported as 0, meaning unknown.
func parseAttempt(env Env, name string, offset int) int {
	raw, ok := envValue(env, name)
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
