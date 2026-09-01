package runby

// ciSpec describes a platform whose detection is a presence marker plus a set
// of variables read by name. Platforms differ only in those names, so they are
// declared as data rather than as one function each. See spec.go for the part
// shared with the terminal and remote axes.
type ciSpec struct {
	specCore
	provider CIProvider

	pipelineID  string
	buildNumber string
	jobID       string
	jobName     string
	trigger     string
	runner      string
	pullRequest ciPRSpec

	// attempt names the retry counter, and attemptOffset normalizes a
	// 0-based retry count to the 1-based CI.Attempt form.
	attempt       string
	attemptOffset int
}

// ciPRSpec describes the optional pull/merge-request signal for a CI
// platform. A platform may advertise only the request kind, or both the kind
// and one or more identifiers. The names are kept separately because a
// condition such as BUILD_REASON=PullRequest is not itself an identifier.
type ciPRSpec struct {
	marker      marker
	markerNames []string
	idNames     []string
}

func ciPR(name string) ciPRSpec {
	return ciPRSpec{marker: markerSet(name), markerNames: []string{name}, idNames: []string{name}}
}

func ciPRAny(names ...string) ciPRSpec {
	return ciPRSpec{marker: func(env Env) bool { return anyPresent(env, names...) }, markerNames: names, idNames: names}
}

func ciPRWhen(m marker, markerNames []string, idNames ...string) ciPRSpec {
	return ciPRSpec{marker: m, markerNames: markerNames, idNames: idNames}
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
		provider: CIForgejo,
		specCore: specCore{
			marker:      markerTrue("FORGEJO_ACTIONS"),
			markerNames: []string{"FORGEJO_ACTIONS"},
			extra: map[string]string{
				"forgejo-runner.repository":   "FORGEJO_REPOSITORY",
				"forgejo-runner.server_url":   "FORGEJO_SERVER_URL",
				"forgejo-runner.workflow_ref": "FORGEJO_WORKFLOW_REF",
				"forgejo-runner.step":         "FORGEJO_ACTION",
				"forgejo-runner.runner_os":    "RUNNER_OS",
			},
			evidence: []string{"CI"},
		},
		pipelineID:  "FORGEJO_RUN_ID",
		buildNumber: "FORGEJO_RUN_NUMBER",
		jobID:       "FORGEJO_JOB",
		trigger:     "FORGEJO_EVENT_NAME",
		pullRequest: ciPRWhen(markerEquals("FORGEJO_EVENT_NAME", "pull_request"), []string{"FORGEJO_EVENT_NAME"}),
		// FORGEJO_RUN_ATTEMPT already counts from 1.
		attempt: "FORGEJO_RUN_ATTEMPT",
	},
	{
		provider: CIGitHubActions,
		specCore: specCore{
			marker:      markerTrue("GITHUB_ACTIONS"),
			markerNames: []string{"GITHUB_ACTIONS"},
			extra: map[string]string{
				"github-actions.action":             "GITHUB_ACTION",
				"github-actions.workflow":           "GITHUB_WORKFLOW",
				"github-actions.repository":         "GITHUB_REPOSITORY",
				"github-actions.runner_environment": "RUNNER_ENVIRONMENT",
			},
			evidence: []string{"CI"},
		},
		pipelineID:  "GITHUB_RUN_ID",
		buildNumber: "GITHUB_RUN_NUMBER",
		jobID:       "GITHUB_JOB",
		trigger:     "GITHUB_EVENT_NAME",
		runner:      "RUNNER_NAME",
		pullRequest: ciPRWhen(markerEquals("GITHUB_EVENT_NAME", "pull_request"), []string{"GITHUB_EVENT_NAME"}),
		attempt:     "GITHUB_RUN_ATTEMPT",
	},
	{
		provider: CIGitLab,
		specCore: specCore{
			marker:      markerTrue("GITLAB_CI"),
			markerNames: []string{"GITLAB_CI"},
			extra: map[string]string{
				"gitlab-ci.project_path": "CI_PROJECT_PATH",
				"gitlab-ci.server_url":   "CI_SERVER_URL",
				"gitlab-ci.stage":        "CI_JOB_STAGE",
			},
			evidence: []string{"CI"},
		},
		pipelineID:  "CI_PIPELINE_ID",
		buildNumber: "CI_PIPELINE_IID",
		jobID:       "CI_JOB_ID",
		jobName:     "CI_JOB_NAME",
		trigger:     "CI_PIPELINE_SOURCE",
		runner:      "CI_RUNNER_ID",
		pullRequest: ciPR("CI_MERGE_REQUEST_ID"),
		// CI_JOB_RETRY_COUNT was added in GitLab 19.3 and counts from 0, so
		// it is absent on older instances and Attempt stays 0 there.
		attempt:       "CI_JOB_RETRY_COUNT",
		attemptOffset: 1,
	},
	{
		provider: CICircleCI,
		specCore: specCore{
			marker:      markerTrue("CIRCLECI"),
			markerNames: []string{"CIRCLECI"},
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
		pipelineID:  "CIRCLE_PIPELINE_ID",
		buildNumber: "CIRCLE_BUILD_NUM",
		jobID:       "CIRCLE_WORKFLOW_JOB_ID",
		jobName:     "CIRCLE_JOB",
		pullRequest: ciPR("CIRCLE_PULL_REQUEST"),
	},
	{
		provider: CITravis,
		specCore: specCore{
			marker:      markerTrue("TRAVIS"),
			markerNames: []string{"TRAVIS"},
			// Travis documents no numeric attempt counter, only a restart flag.
			extra: map[string]string{
				"travis-ci.job_number":    "TRAVIS_JOB_NUMBER",
				"travis-ci.repo_slug":     "TRAVIS_REPO_SLUG",
				"travis-ci.job_restarted": "TRAVIS_JOB_RESTARTED",
			},
			evidence: []string{"CI", "CONTINUOUS_INTEGRATION"},
		},
		pipelineID:  "TRAVIS_BUILD_ID",
		buildNumber: "TRAVIS_BUILD_NUMBER",
		jobID:       "TRAVIS_JOB_ID",
		jobName:     "TRAVIS_JOB_NAME",
		trigger:     "TRAVIS_EVENT_TYPE",
		pullRequest: ciPRWhen(markerNotEquals("TRAVIS_PULL_REQUEST", "false"), []string{"TRAVIS_PULL_REQUEST"}, "TRAVIS_PULL_REQUEST"),
	},
	{
		provider: CIBuildkite,
		specCore: specCore{
			marker:      markerTrue("BUILDKITE"),
			markerNames: []string{"BUILDKITE"},
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
		pipelineID:    "BUILDKITE_BUILD_ID",
		buildNumber:   "BUILDKITE_BUILD_NUMBER",
		jobID:         "BUILDKITE_JOB_ID",
		jobName:       "BUILDKITE_LABEL",
		trigger:       "BUILDKITE_SOURCE",
		runner:        "BUILDKITE_AGENT_NAME",
		pullRequest:   ciPRWhen(markerNotEquals("BUILDKITE_PULL_REQUEST", "false"), []string{"BUILDKITE_PULL_REQUEST"}, "BUILDKITE_PULL_REQUEST"),
		attempt:       "BUILDKITE_RETRY_COUNT",
		attemptOffset: 1, // Buildkite counts retries from 0; Attempt is 1-based.
	},
	{
		provider: CIAzurePipelines,
		specCore: specCore{
			// The documented value is "True"; IsTrue parses case-insensitively.
			marker:      markerTrue("TF_BUILD"),
			markerNames: []string{"TF_BUILD"},
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
		pipelineID:  "BUILD_BUILDID",
		buildNumber: "BUILD_BUILDNUMBER",
		jobID:       "SYSTEM_JOBID",
		jobName:     "SYSTEM_JOBDISPLAYNAME",
		trigger:     "BUILD_REASON",
		runner:      "AGENT_NAME",
		pullRequest: ciPRWhen(markerEquals("BUILD_REASON", "PullRequest"), []string{"BUILD_REASON"}, "SYSTEM_PULLREQUEST_PULLREQUESTID"),
		attempt:     "SYSTEM_JOBATTEMPT",
	},
	{
		provider: CIBitbucket,
		specCore: specCore{
			// Bitbucket Pipelines has no dedicated boolean marker, so the always
			// present build number doubles as one. A second signal is required so
			// that a stray BITBUCKET_BUILD_NUMBER alone does not match.
			marker: func(env Env) bool {
				if _, ok := envValue(env, "BITBUCKET_BUILD_NUMBER"); !ok {
					return false
				}
				_, hasUUID := envValue(env, "BITBUCKET_PIPELINE_UUID")
				return hasUUID || envIsTrue(env, "CI")
			},
			markerNames: []string{"BITBUCKET_BUILD_NUMBER", "BITBUCKET_PIPELINE_UUID", "CI"},
			// Bitbucket wraps its UUID values in curly braces.
			trimBraces: true,
			extra: map[string]string{
				"bitbucket-pipelines.repo_slug":  "BITBUCKET_REPO_SLUG",
				"bitbucket-pipelines.workspace":  "BITBUCKET_WORKSPACE",
				"bitbucket-pipelines.deployment": "BITBUCKET_DEPLOYMENT_ENVIRONMENT",
				// Bitbucket exposes no trigger-type enum; a PR ID is the only
				// direct signal that this run came from a pull request.
				"bitbucket-pipelines.pr_id": "BITBUCKET_PR_ID",
			},
		},
		pipelineID:  "BITBUCKET_PIPELINE_UUID",
		buildNumber: "BITBUCKET_BUILD_NUMBER",
		jobID:       "BITBUCKET_STEP_UUID",
		pullRequest: ciPR("BITBUCKET_PR_ID"),
		// BITBUCKET_STEP_RUN_NUMBER already counts from 1.
		attempt: "BITBUCKET_STEP_RUN_NUMBER",
	},
	{
		provider: CIJenkins,
		specCore: specCore{
			// Jenkins has no boolean marker either, and BUILD_NUMBER alone is far
			// too generic a name, so a Jenkins-owned variable is required as
			// well. JENKINS_URL is not enough on its own: it is only injected
			// when an administrator has configured the Jenkins root URL, while
			// JENKINS_HOME is always set. HUDSON_* are core legacy aliases.
			marker: func(env Env) bool {
				if _, ok := envValue(env, "BUILD_NUMBER"); !ok {
					return false
				}
				for _, name := range []string{"JENKINS_URL", "JENKINS_HOME", "HUDSON_URL", "HUDSON_HOME"} {
					if _, ok := envValue(env, name); ok {
						return true
					}
				}
				return false
			},
			markerNames: []string{"BUILD_NUMBER", "JENKINS_URL", "JENKINS_HOME", "HUDSON_URL", "HUDSON_HOME"},
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
		// Jenkins exposes no opaque run identifier. BUILD_ID has been
		// identical to BUILD_NUMBER since Jenkins 1.597.
		pipelineID:  "BUILD_ID",
		buildNumber: "BUILD_NUMBER",
		jobName:     "JOB_NAME",
		runner:      "NODE_NAME",
		pullRequest: ciPRAny("ghprbPullId", "CHANGE_ID"),
	},
	{
		provider: CIGeneric,
		specCore: specCore{
			// The bare CI convention is widely honored but is owned by no
			// platform, and local tooling sets it too, so it is only probable.
			marker:      markerTrue("CI"),
			markerNames: []string{"CI"},
			confidence:  ConfidenceProbable,
		},
	},
}

// detect reads the spec's variables out of env.
func (spec ciSpec) detect(env Env) (CI, bool) {
	result := CI{Detected: true, Provider: spec.provider}
	values, ok := spec.read(env,
		specField{spec.pipelineID, &result.PipelineID},
		specField{spec.buildNumber, &result.BuildNumber},
		specField{spec.jobID, &result.JobID},
		specField{spec.jobName, &result.JobName},
		specField{spec.trigger, &result.Trigger},
		specField{spec.runner, &result.Runner},
	)
	if !ok {
		return CI{}, false
	}

	// Attempt is read on its own because it is the one field that is parsed
	// rather than copied.
	result.Attempt = parseAttempt(env, spec.attempt, spec.attemptOffset)
	values.add(spec.attempt)

	if spec.pullRequest.marker != nil {
		values.add(spec.pullRequest.markerNames...)
		values.add(spec.pullRequest.idNames...)
		result.PullRequest = spec.pullRequest.marker(env)
		for _, name := range spec.pullRequest.idNames {
			if value, ok := envValue(env, name); ok {
				result.PullRequestID = value
				break
			}
		}
	}

	values.apply(env, &result.Axis)
	return result, true
}

// builtinCIDrivers is ordered from the most specific platform to the generic
// CI convention. Detect reports the first match.
var builtinCIDrivers = mapSlice(ciSpecs, func(spec ciSpec) CIDriver {
	return CIDriver{Provider: spec.provider, Detect: spec.detect}
})
