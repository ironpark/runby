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
		provider: CIGiteaActions,
		specCore: specCore{
			// Gitea publishes GitHub-compatible aliases as well as its own
			// marker. Keep it ahead of GitHub Actions for the same reason Forgejo
			// is ahead of GitHub.
			marker:      markerTrue("GITEA_ACTIONS"),
			markerNames: []string{"GITEA_ACTIONS"},
			extra: map[string]string{
				"gitea-actions.runner_version": "GITEA_ACTIONS_RUNNER_VERSION",
				"gitea-actions.repository":     "GITHUB_REPOSITORY",
			},
			evidence: []string{"CI"},
		},
		pipelineID:  "GITHUB_RUN_ID",
		buildNumber: "GITHUB_RUN_NUMBER",
		jobID:       "GITHUB_JOB",
		trigger:     "GITHUB_EVENT_NAME",
		runner:      "RUNNER_NAME",
		pullRequest: ciPRWhen(markerEquals("GITHUB_EVENT_NAME", "pull_request"), []string{"GITHUB_EVENT_NAME"}),
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
		provider: CIVercel,
		specCore: specCore{
			marker:      func(env Env) bool { return anyPresent(env, "VERCEL", "NOW_BUILDER") },
			markerNames: []string{"VERCEL", "NOW_BUILDER"},
			extra: map[string]string{
				"vercel.environment":    "VERCEL_ENV",
				"vercel.url":            "VERCEL_URL",
				"vercel.git_repo_slug":  "VERCEL_GIT_REPO_SLUG",
				"vercel.git_commit_ref": "VERCEL_GIT_COMMIT_REF",
				"vercel.git_commit_sha": "VERCEL_GIT_COMMIT_SHA",
			},
		},
		pipelineID:  "VERCEL_DEPLOYMENT_ID",
		pullRequest: ciPR("VERCEL_GIT_PULL_REQUEST_ID"),
	},
	{
		provider: CINetlify,
		specCore: specCore{
			marker:      markerSet("NETLIFY"),
			markerNames: []string{"NETLIFY"},
			extra: map[string]string{
				"netlify.url":    "URL",
				"netlify.branch": "BRANCH",
			},
		},
		pipelineID:  "DEPLOY_ID",
		jobID:       "BUILD_ID",
		trigger:     "CONTEXT",
		pullRequest: ciPRWhen(markerNotEquals("PULL_REQUEST", "false"), []string{"PULL_REQUEST"}),
	},
	{
		provider: CITeamCity,
		specCore: specCore{
			marker:      markerSet("TEAMCITY_VERSION"),
			markerNames: []string{"TEAMCITY_VERSION"},
			extra:       map[string]string{"teamcity.version": "TEAMCITY_VERSION"},
		},
		pipelineID:  "BUILD_ID",
		buildNumber: "BUILD_NUMBER",
		jobName:     "TEAMCITY_BUILDCONF_NAME",
	},
	{
		provider: CIDrone,
		specCore: specCore{
			marker:      markerTrue("DRONE"),
			markerNames: []string{"DRONE"},
			extra: map[string]string{
				"drone.repository": "DRONE_REPO",
				"drone.system":     "DRONE_SYSTEM_HOST",
			},
			evidence: []string{"CI"},
		},
		pipelineID:  "DRONE_BUILD_NUMBER",
		buildNumber: "DRONE_BUILD_NUMBER",
		jobID:       "DRONE_STAGE_NUMBER",
		jobName:     "DRONE_STEP_NAME",
		trigger:     "DRONE_BUILD_EVENT",
		runner:      "DRONE_SYSTEM_HOSTNAME",
		pullRequest: ciPRWhen(markerEquals("DRONE_BUILD_EVENT", "pull_request"), []string{"DRONE_BUILD_EVENT"}, "DRONE_PULL_REQUEST"),
	},
	{
		provider: CIAppVeyor,
		specCore: specCore{
			marker:      markerTrue("APPVEYOR"),
			markerNames: []string{"APPVEYOR"},
			evidence:    []string{"CI"},
		},
		pipelineID:  "APPVEYOR_BUILD_ID",
		buildNumber: "APPVEYOR_BUILD_NUMBER",
		jobID:       "APPVEYOR_JOB_ID",
		jobName:     "APPVEYOR_JOB_NAME",
		runner:      "APPVEYOR_BUILD_WORKER_IMAGE",
		pullRequest: ciPR("APPVEYOR_PULL_REQUEST_NUMBER"),
	},
	{
		provider: CISemaphore,
		specCore: specCore{
			marker:      markerTrue("SEMAPHORE"),
			markerNames: []string{"SEMAPHORE"},
			evidence:    []string{"CI"},
		},
		pipelineID:  "SEMAPHORE_PIPELINE_ID",
		buildNumber: "SEMAPHORE_WORKFLOW_NUMBER",
		jobID:       "SEMAPHORE_JOB_ID",
		jobName:     "SEMAPHORE_JOB_NAME",
		trigger:     "SEMAPHORE_GIT_REF_TYPE",
		pullRequest: ciPRAny("PULL_REQUEST_NUMBER", "SEMAPHORE_GIT_PR_NUMBER"),
	},
	{
		provider: CICirrus,
		specCore: specCore{
			marker:      markerSet("CIRRUS_CI"),
			markerNames: []string{"CIRRUS_CI"},
			evidence:    []string{"CI", "CONTINUOUS_INTEGRATION"},
		},
		pipelineID:  "CIRRUS_BUILD_ID",
		jobID:       "CIRRUS_TASK_ID",
		jobName:     "CIRRUS_TASK_NAME",
		runner:      "CIRRUS_OS",
		pullRequest: ciPR("CIRRUS_PR"),
	},
	{
		provider: CIAWSCodeBuild,
		specCore: specCore{
			marker:      markerSet("CODEBUILD_BUILD_ARN"),
			markerNames: []string{"CODEBUILD_BUILD_ARN"},
		},
		pipelineID:  "CODEBUILD_BUILD_ARN",
		buildNumber: "CODEBUILD_BUILD_NUMBER",
		jobID:       "CODEBUILD_BUILD_ID",
		trigger:     "CODEBUILD_WEBHOOK_EVENT",
		pullRequest: ciPRWhen(markerHasPrefix("CODEBUILD_WEBHOOK_EVENT", "PULL_REQUEST_"), []string{"CODEBUILD_WEBHOOK_EVENT"}),
	},
	{
		provider: CIGoogleCloudBuild,
		specCore: specCore{
			// BUILDER_OUTPUT is exported by Google's Cloud Build builder
			// environment. It is a product-specific marker, unlike generic
			// BUILD_ID or the user-defined env entries in a build config.
			marker:      markerSet("BUILDER_OUTPUT"),
			markerNames: []string{"BUILDER_OUTPUT"},
		},
	},
	{
		provider: CIXcodeCloud,
		specCore: specCore{
			marker:      markerSet("CI_XCODE_PROJECT"),
			markerNames: []string{"CI_XCODE_PROJECT"},
			evidence:    []string{"CI"},
		},
		pipelineID:  "CI_BUILD_ID",
		buildNumber: "CI_BUILD_NUMBER",
		jobID:       "CI_WORKFLOW_ID",
		jobName:     "CI_XCODE_SCHEME",
		pullRequest: ciPR("CI_PULL_REQUEST_NUMBER"),
	},
	{
		provider: CICloudflarePages,
		specCore: specCore{
			marker:      markerSet("CF_PAGES"),
			markerNames: []string{"CF_PAGES"},
			evidence:    []string{"CI"},
			extra: map[string]string{
				"cloudflare-pages.branch":     "CF_PAGES_BRANCH",
				"cloudflare-pages.commit_sha": "CF_PAGES_COMMIT_SHA",
				"cloudflare-pages.url":        "CF_PAGES_URL",
			},
		},
	},
	{
		provider: CICloudflareWorkers,
		specCore: specCore{
			marker:      markerSet("WORKERS_CI"),
			markerNames: []string{"WORKERS_CI"},
			evidence:    []string{"CI"},
			extra: map[string]string{
				"cloudflare-workers.branch":     "WORKERS_CI_BRANCH",
				"cloudflare-workers.commit_sha": "WORKERS_CI_COMMIT_SHA",
			},
		},
		pipelineID: "WORKERS_CI_BUILD_UUID",
	},
	{
		provider: CIWoodpecker,
		specCore: specCore{
			// Woodpecker sets CI=woodpecker, which is more specific than the
			// generic CI=true convention and therefore needs an equality marker.
			marker:      markerEquals("CI", "woodpecker"),
			markerNames: []string{"CI"},
			extra: map[string]string{
				"woodpecker.repository":      "CI_REPO",
				"woodpecker.system":          "CI_SYSTEM_HOST",
				"woodpecker.pipeline_number": "CI_PIPELINE_NUMBER",
				"woodpecker.pipeline_event":  "CI_PIPELINE_EVENT",
				"woodpecker.pull_request":    "CI_COMMIT_PULL_REQUEST",
				"woodpecker.step_number":     "CI_STEP_NUMBER",
			},
		},
		pipelineID:  "CI_BUILD_NUMBER",
		buildNumber: "CI_BUILD_NUMBER",
		jobID:       "CI_JOB_NUMBER",
		jobName:     "CI_STEP_NAME",
		trigger:     "CI_BUILD_EVENT",
		runner:      "CI_SYSTEM_HOST",
		pullRequest: ciPRWhen(func(env Env) bool {
			return markerEquals("CI_BUILD_EVENT", "pull_request")(env) ||
				markerEquals("CI_PIPELINE_EVENT", "pull_request")(env) ||
				anyPresent(env, "CI_PULL_REQUEST", "CI_COMMIT_PULL_REQUEST")
		}, []string{"CI_BUILD_EVENT", "CI_PIPELINE_EVENT", "CI_PULL_REQUEST", "CI_COMMIT_PULL_REQUEST"}, "CI_PULL_REQUEST", "CI_COMMIT_PULL_REQUEST"),
	},
	{
		provider: CIBitrise,
		specCore: specCore{
			marker:      markerTrue("BITRISE_IO"),
			markerNames: []string{"BITRISE_IO"},
			evidence:    []string{"CI"},
		},
		pipelineID:  "BITRISE_BUILD_SLUG",
		buildNumber: "BITRISE_BUILD_NUMBER",
		jobName:     "BITRISE_TRIGGERED_WORKFLOW_ID",
		trigger:     "BITRISE_TRIGGER_BY",
		pullRequest: ciPR("BITRISE_PULL_REQUEST"),
	},
	{
		provider: CIRender,
		specCore: specCore{
			marker:      markerTrue("RENDER"),
			markerNames: []string{"RENDER"},
			extra: map[string]string{
				"render.service_id": "RENDER_SERVICE_ID",
				"render.branch":     "RENDER_GIT_BRANCH",
				"render.commit":     "RENDER_GIT_COMMIT",
			},
		},
		pullRequest: ciPRWhen(markerEquals("IS_PULL_REQUEST", "true"), []string{"IS_PULL_REQUEST"}),
	},
	{
		provider: CIHarness,
		specCore: specCore{
			marker:      markerSet("HARNESS_BUILD_ID"),
			markerNames: []string{"HARNESS_BUILD_ID"},
			extra: map[string]string{
				"harness.org_id":      "HARNESS_ORG_ID",
				"harness.project_id":  "HARNESS_PROJECT_ID",
				"harness.pipeline_id": "HARNESS_PIPELINE_ID",
			},
		},
		pipelineID: "HARNESS_BUILD_ID",
		jobID:      "HARNESS_STAGE_ID",
		jobName:    "HARNESS_STAGE_NAME",
	},
	{
		provider: CIBamboo,
		specCore: specCore{
			marker:      markerSet("bamboo_planKey"),
			markerNames: []string{"bamboo_planKey"},
			extra: map[string]string{
				"bamboo.plan_name": "bamboo_planName",
				"bamboo.agent_id":  "bamboo_agentId",
			},
		},
		pipelineID:  "bamboo_buildResultKey",
		buildNumber: "bamboo_buildNumber",
		jobID:       "bamboo_buildKey",
		jobName:     "bamboo_planName",
		runner:      "bamboo_agentId",
		pullRequest: ciPR("bamboo_repository_pr_key"),
	},
	{
		provider: CIGoCD,
		specCore: specCore{
			marker:      markerSet("GO_PIPELINE_LABEL"),
			markerNames: []string{"GO_PIPELINE_LABEL"},
			extra:       map[string]string{"gocd.pipeline_name": "GO_PIPELINE_NAME"},
		},
		pipelineID:  "GO_PIPELINE_LABEL",
		buildNumber: "GO_PIPELINE_COUNTER",
		jobID:       "GO_JOB_NAME",
		jobName:     "GO_STAGE_NAME",
		runner:      "GO_AGENT_HOST",
	},
	{
		provider: CITaskCluster,
		specCore: specCore{
			marker:      markerSet("TASK_ID", "RUN_ID"),
			markerNames: []string{"TASK_ID", "RUN_ID"},
		},
		pipelineID:    "TASK_ID",
		attempt:       "RUN_ID",
		attemptOffset: 1,
	},
	{
		provider: CISourcehut,
		specCore: specCore{
			// CI_NAME is a generic-looking variable, but sourcehut's build
			// service assigns the literal product name to it. Keep this specific
			// equality ahead of the generic CI_NAME heuristic.
			marker:      markerEquals("CI_NAME", "sourcehut"),
			markerNames: []string{"CI_NAME"},
		},
		pipelineID: "JOB_ID",
	},
	{
		provider: CICodefresh,
		specCore: specCore{
			marker:      markerSet("CF_BUILD_ID"),
			markerNames: []string{"CF_BUILD_ID"},
			extra: map[string]string{
				"codefresh.branch":   "CF_BRANCH",
				"codefresh.revision": "CF_REVISION",
			},
		},
		pipelineID:  "CF_BUILD_ID",
		jobName:     "CF_STEP_NAME",
		pullRequest: ciPRAny("CF_PULL_REQUEST_NUMBER", "CF_PULL_REQUEST_ID"),
	},
	{
		provider: CICodemagic,
		specCore: specCore{
			marker:      markerSet("CM_BUILD_ID"),
			markerNames: []string{"CM_BUILD_ID"},
		},
		pipelineID:  "CM_BUILD_ID",
		buildNumber: "BUILD_NUMBER",
		jobName:     "CM_WORKFLOW_NAME",
		trigger:     "CM_TRIGGER_SOURCE",
		pullRequest: ciPRWhen(markerEquals("CM_PULL_REQUEST", "true"), []string{"CM_PULL_REQUEST"}, "CM_PULL_REQUEST_NUMBER"),
	},
	{
		provider: CIBuddy,
		specCore: specCore{
			marker:      markerSet("BUDDY_WORKSPACE_ID"),
			markerNames: []string{"BUDDY_WORKSPACE_ID"},
		},
		pipelineID:  "BUDDY_EXECUTION_ID",
		buildNumber: "BUDDY_EXECUTION_NUMBER",
		jobID:       "BUDDY_ACTION_ID",
		jobName:     "BUDDY_ACTION_NAME",
		trigger:     "BUDDY_PIPELINE_TRIGGER_MODE",
		pullRequest: ciPRAny("BUDDY_EXECUTION_PULL_REQUEST_ID", "BUDDY_EXECUTION_PULL_REQUEST_NO", "BUDDY_RUN_PR_ID", "BUDDY_RUN_PR_NO"),
	},
	{
		provider: CIScrewdriver,
		specCore: specCore{
			marker:      markerTrue("SCREWDRIVER"),
			markerNames: []string{"SCREWDRIVER"},
			evidence:    []string{"CI", "CONTINUOUS_INTEGRATION"},
		},
		pipelineID:  "SD_BUILD_ID",
		jobID:       "SD_JOB_ID",
		jobName:     "SD_JOB_NAME",
		pullRequest: ciPRWhen(markerNotEquals("SD_PULL_REQUEST", "false"), []string{"SD_PULL_REQUEST"}, "SD_PULL_REQUEST"),
	},
	{
		provider: CIVela,
		specCore: specCore{
			marker:      markerSet("VELA"),
			markerNames: []string{"VELA"},
		},
		pipelineID:  "VELA_BUILD_NUMBER",
		trigger:     "VELA_BUILD_EVENT",
		runner:      "VELA_BUILD_HOST",
		pullRequest: ciPRWhen(markerEquals("VELA_PULL_REQUEST", "1"), []string{"VELA_PULL_REQUEST"}, "VELA_PULL_REQUEST"),
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
