package runby

import "strings"

// ciSpec describes a platform whose detection is a presence marker plus a set
// of variables read by name. Platforms differ only in those names, so they are
// declared as data rather than as one function each.
type ciSpec struct {
	provider CIProvider
	// marker reports whether the environment shows this platform is running.
	marker func(Env) bool
	// markerNames lists the variables marker consults, so that they are
	// reported as evidence alongside the fields read below.
	markerNames []string
	// confidence defaults to ConfidenceDefinite when empty.
	confidence Confidence

	pipelineID  string
	buildNumber string
	jobID       string
	jobName     string
	trigger     string
	runner      string

	// attempt names the retry counter, and attemptOffset normalizes a
	// 0-based retry count to the 1-based CI.Attempt form.
	attempt       string
	attemptOffset int

	// trimBraces strips the surrounding curly braces that some platforms
	// include in their UUID values, so PipelineID and JobID are usable as-is.
	trimBraces bool

	// extra maps a CI.Extra key to the variable supplying it.
	extra map[string]string
	// evidence lists further variables that count as evidence when set. The
	// marker and the fields named above are added automatically.
	evidence []string
}

// markerTrue matches when every name holds a parsable boolean that is true.
func markerTrue(names ...string) func(Env) bool {
	return func(env Env) bool {
		for _, name := range names {
			if !IsTrue(env, name) {
				return false
			}
		}
		return true
	}
}

// markerSet matches when every name is set to a non-empty value.
func markerSet(names ...string) func(Env) bool {
	return func(env Env) bool {
		for _, name := range names {
			if _, ok := Value(env, name); !ok {
				return false
			}
		}
		return true
	}
}

// ciSpecs is ordered so that a specific platform is reported ahead of the
// generic CI convention, which every entry here also sets.
var ciSpecs = []ciSpec{
	{
		// Forgejo Runner v7 and later mirrors every FORGEJO_* variable to a
		// GITHUB_* alias, so it must be tried before GitHub Actions or a
		// Forgejo job is misreported as a GitHub one. Runners older than v7
		// define only the GITHUB_* names and are indistinguishable from
		// GitHub Actions by environment alone; they fall through on purpose.
		provider:    CIProviderForgejo,
		marker:      markerTrue("FORGEJO_ACTIONS"),
		markerNames: []string{"FORGEJO_ACTIONS"},
		pipelineID:  "FORGEJO_RUN_ID",
		buildNumber: "FORGEJO_RUN_NUMBER",
		jobID:       "FORGEJO_JOB",
		trigger:     "FORGEJO_EVENT_NAME",
		// FORGEJO_RUN_ATTEMPT already counts from 1.
		attempt: "FORGEJO_RUN_ATTEMPT",
		extra: map[string]string{
			"forgejo-runner.repository":   "FORGEJO_REPOSITORY",
			"forgejo-runner.server_url":   "FORGEJO_SERVER_URL",
			"forgejo-runner.workflow_ref": "FORGEJO_WORKFLOW_REF",
			"forgejo-runner.step":         "FORGEJO_ACTION",
			"forgejo-runner.runner_os":    "RUNNER_OS",
		},
		evidence: []string{"CI"},
	},
	{
		provider:    CIProviderGitHubActions,
		marker:      markerTrue("GITHUB_ACTIONS"),
		markerNames: []string{"GITHUB_ACTIONS"},
		pipelineID:  "GITHUB_RUN_ID",
		buildNumber: "GITHUB_RUN_NUMBER",
		jobID:       "GITHUB_JOB",
		trigger:     "GITHUB_EVENT_NAME",
		runner:      "RUNNER_NAME",
		attempt:     "GITHUB_RUN_ATTEMPT",
		extra: map[string]string{
			"github-actions.action":             "GITHUB_ACTION",
			"github-actions.workflow":           "GITHUB_WORKFLOW",
			"github-actions.repository":         "GITHUB_REPOSITORY",
			"github-actions.runner_environment": "RUNNER_ENVIRONMENT",
		},
		evidence: []string{"CI"},
	},
	{
		provider:    CIProviderGitLab,
		marker:      markerTrue("GITLAB_CI"),
		markerNames: []string{"GITLAB_CI"},
		pipelineID:  "CI_PIPELINE_ID",
		buildNumber: "CI_PIPELINE_IID",
		jobID:       "CI_JOB_ID",
		jobName:     "CI_JOB_NAME",
		trigger:     "CI_PIPELINE_SOURCE",
		runner:      "CI_RUNNER_ID",
		// CI_JOB_RETRY_COUNT was added in GitLab 19.3 and counts from 0, so
		// it is absent on older instances and Attempt stays 0 there.
		attempt:       "CI_JOB_RETRY_COUNT",
		attemptOffset: 1,
		extra: map[string]string{
			"gitlab-ci.project_path": "CI_PROJECT_PATH",
			"gitlab-ci.server_url":   "CI_SERVER_URL",
			"gitlab-ci.stage":        "CI_JOB_STAGE",
		},
		evidence: []string{"CI"},
	},
	{
		provider:    CIProviderCircleCI,
		marker:      markerTrue("CIRCLECI"),
		markerNames: []string{"CIRCLECI"},
		pipelineID:  "CIRCLE_PIPELINE_ID",
		buildNumber: "CIRCLE_BUILD_NUM",
		jobID:       "CIRCLE_WORKFLOW_JOB_ID",
		jobName:     "CIRCLE_JOB",
		// CircleCI documents no rerun counter, and its trigger source is a
		// config-time pipeline value rather than a job environment variable.
		extra: map[string]string{
			"circleci.workflow_id":  "CIRCLE_WORKFLOW_ID",
			"circleci.project_repo": "CIRCLE_PROJECT_REPONAME",
			"circleci.node_index":   "CIRCLE_NODE_INDEX",
			"circleci.node_total":   "CIRCLE_NODE_TOTAL",
		},
		evidence: []string{"CI"},
	},
	{
		provider:    CIProviderTravis,
		marker:      markerTrue("TRAVIS"),
		markerNames: []string{"TRAVIS"},
		pipelineID:  "TRAVIS_BUILD_ID",
		buildNumber: "TRAVIS_BUILD_NUMBER",
		jobID:       "TRAVIS_JOB_ID",
		jobName:     "TRAVIS_JOB_NAME",
		trigger:     "TRAVIS_EVENT_TYPE",
		// Travis documents no numeric attempt counter, only a restart flag.
		extra: map[string]string{
			"travis-ci.job_number":    "TRAVIS_JOB_NUMBER",
			"travis-ci.repo_slug":     "TRAVIS_REPO_SLUG",
			"travis-ci.job_restarted": "TRAVIS_JOB_RESTARTED",
		},
		evidence: []string{"CI", "CONTINUOUS_INTEGRATION"},
	},
	{
		provider:      CIProviderBuildkite,
		marker:        markerTrue("BUILDKITE"),
		markerNames:   []string{"BUILDKITE"},
		pipelineID:    "BUILDKITE_BUILD_ID",
		buildNumber:   "BUILDKITE_BUILD_NUMBER",
		jobID:         "BUILDKITE_JOB_ID",
		jobName:       "BUILDKITE_LABEL",
		trigger:       "BUILDKITE_SOURCE",
		runner:        "BUILDKITE_AGENT_NAME",
		attempt:       "BUILDKITE_RETRY_COUNT",
		attemptOffset: 1, // Buildkite counts retries from 0; Attempt is 1-based.
		extra: map[string]string{
			"buildkite.pipeline_slug":     "BUILDKITE_PIPELINE_SLUG",
			"buildkite.organization_slug": "BUILDKITE_ORGANIZATION_SLUG",
			"buildkite.agent_id":          "BUILDKITE_AGENT_ID",
			"buildkite.compute_type":      "BUILDKITE_COMPUTE_TYPE",
			// BUILDKITE_AGENT_PID is the long-lived agent daemon's PID, not
			// this job's process, so it is context rather than identity.
			"buildkite.agent_pid": "BUILDKITE_AGENT_PID",
		},
		evidence: []string{"CI"},
	},
	{
		provider: CIProviderAzurePipelines,
		// The documented value is "True"; IsTrue parses case-insensitively.
		marker:      markerTrue("TF_BUILD"),
		markerNames: []string{"TF_BUILD"},
		pipelineID:  "BUILD_BUILDID",
		buildNumber: "BUILD_BUILDNUMBER",
		jobID:       "SYSTEM_JOBID",
		jobName:     "SYSTEM_JOBDISPLAYNAME",
		trigger:     "BUILD_REASON",
		runner:      "AGENT_NAME",
		attempt:     "SYSTEM_JOBATTEMPT",
		extra: map[string]string{
			// Azure documents no stage ID, only a stage name.
			"azure-pipelines.stage_name":     "SYSTEM_STAGENAME",
			"azure-pipelines.stage_attempt":  "SYSTEM_STAGEATTEMPT",
			"azure-pipelines.definition":     "BUILD_DEFINITIONNAME",
			"azure-pipelines.collection_uri": "SYSTEM_COLLECTIONURI",
			"azure-pipelines.host_type":      "SYSTEM_HOSTTYPE",
		},
		evidence: []string{"CI"},
	},
	{
		provider: CIProviderBitbucket,
		// Bitbucket Pipelines has no dedicated boolean marker, so the always
		// present build number doubles as one. A second signal is required so
		// that a stray BITBUCKET_BUILD_NUMBER alone does not match.
		marker: func(env Env) bool {
			if _, ok := Value(env, "BITBUCKET_BUILD_NUMBER"); !ok {
				return false
			}
			_, hasUUID := Value(env, "BITBUCKET_PIPELINE_UUID")
			return hasUUID || IsTrue(env, "CI")
		},
		markerNames: []string{"BITBUCKET_BUILD_NUMBER", "BITBUCKET_PIPELINE_UUID", "CI"},
		// Bitbucket wraps its UUID values in curly braces.
		trimBraces:  true,
		pipelineID:  "BITBUCKET_PIPELINE_UUID",
		buildNumber: "BITBUCKET_BUILD_NUMBER",
		jobID:       "BITBUCKET_STEP_UUID",
		// BITBUCKET_STEP_RUN_NUMBER already counts from 1.
		attempt: "BITBUCKET_STEP_RUN_NUMBER",
		extra: map[string]string{
			"bitbucket-pipelines.repo_slug":  "BITBUCKET_REPO_SLUG",
			"bitbucket-pipelines.workspace":  "BITBUCKET_WORKSPACE",
			"bitbucket-pipelines.deployment": "BITBUCKET_DEPLOYMENT_ENVIRONMENT",
			// Bitbucket exposes no trigger-type enum; a PR ID is the only
			// direct signal that this run came from a pull request.
			"bitbucket-pipelines.pr_id": "BITBUCKET_PR_ID",
		},
	},
	{
		provider: CIProviderJenkins,
		// Jenkins has no boolean marker either, and BUILD_NUMBER alone is far
		// too generic a name, so a Jenkins-owned variable is required as
		// well. JENKINS_URL is not enough on its own: it is only injected
		// when an administrator has configured the Jenkins root URL, while
		// JENKINS_HOME is always set. HUDSON_* are core legacy aliases.
		marker: func(env Env) bool {
			if _, ok := Value(env, "BUILD_NUMBER"); !ok {
				return false
			}
			for _, name := range []string{"JENKINS_URL", "JENKINS_HOME", "HUDSON_URL", "HUDSON_HOME"} {
				if _, ok := Value(env, name); ok {
					return true
				}
			}
			return false
		},
		markerNames: []string{"BUILD_NUMBER", "JENKINS_URL", "JENKINS_HOME", "HUDSON_URL", "HUDSON_HOME"},
		// Jenkins exposes no opaque run identifier. BUILD_ID has been
		// identical to BUILD_NUMBER since Jenkins 1.597.
		pipelineID:  "BUILD_ID",
		buildNumber: "BUILD_NUMBER",
		jobName:     "JOB_NAME",
		runner:      "NODE_NAME",
		// Core Jenkins advertises no trigger type; that needs a plugin.
		extra: map[string]string{
			"jenkins.build_tag": "BUILD_TAG",
			"jenkins.url":       "JENKINS_URL",
			"jenkins.home":      "JENKINS_HOME",
			// EXECUTOR_NUMBER, NODE_NAME, and NODE_LABELS are only injected
			// inside an Executor thread, so Pipeline code outside a node
			// block will not have them.
			"jenkins.executor_number": "EXECUTOR_NUMBER",
			"jenkins.node_labels":     "NODE_LABELS",
		},
		evidence: []string{"CI", "WORKSPACE"},
	},
	{
		provider: CIProviderGeneric,
		// The bare CI convention is widely honored but is owned by no
		// platform, and local tooling sets it too, so it is only probable.
		marker:      markerTrue("CI"),
		markerNames: []string{"CI"},
		confidence:  ConfidenceProbable,
	},
}

// detect reads the spec's variables out of env.
func (spec ciSpec) detect(env Env) (CI, bool) {
	if !spec.marker(env) {
		return CI{}, false
	}

	confidence := spec.confidence
	if confidence == "" {
		confidence = ConfidenceDefinite
	}
	result := CI{
		Detected:   true,
		Provider:   spec.provider,
		Confidence: confidence,
		Attempt:    parseAttempt(env, spec.attempt, spec.attemptOffset),
	}

	names := append(append([]string{}, spec.markerNames...), spec.evidence...)
	if spec.attempt != "" {
		names = append(names, spec.attempt)
	}
	for _, field := range []struct {
		name string
		into *string
	}{
		{spec.pipelineID, &result.PipelineID},
		{spec.buildNumber, &result.BuildNumber},
		{spec.jobID, &result.JobID},
		{spec.jobName, &result.JobName},
		{spec.trigger, &result.Trigger},
		{spec.runner, &result.Runner},
	} {
		if field.name == "" {
			continue
		}
		names = append(names, field.name)
		value, _ := Value(env, field.name)
		if spec.trimBraces {
			value = strings.TrimSuffix(strings.TrimPrefix(value, "{"), "}")
		}
		*field.into = value
	}

	for key, name := range spec.extra {
		names = append(names, name)
		value, ok := Value(env, name)
		if !ok {
			continue
		}
		if result.Extra == nil {
			result.Extra = make(map[string]string, len(spec.extra))
		}
		result.Extra[key] = value
	}

	result.Evidence = PresentNames(env, names...)
	return result, true
}

// builtinCIDetectors is ordered from the most specific platform to the generic
// CI convention. Detect reports the first match.
var builtinCIDetectors = func() []CIDetector {
	detectors := make([]CIDetector, 0, len(ciSpecs))
	for _, spec := range ciSpecs {
		detectors = append(detectors, NewCIDetector(spec.provider, spec.detect))
	}
	return detectors
}()

// CIDetectors returns the built-in CI detectors in precedence order. The
// returned slice is a copy and may be reordered or filtered before being
// passed back through WithOnlyCIDetectors.
func CIDetectors() []CIDetector {
	detectors := make([]CIDetector, len(builtinCIDetectors))
	copy(detectors, builtinCIDetectors)
	return detectors
}
