package runby

func detectCodex(env Env) (Detection, bool) {
	threadID, hasThreadID := Value(env, "CODEX_THREAD_ID")
	sessionID, hasSessionID := Value(env, "CODEX_SESSION_ID")
	sandbox, hasSandbox := Value(env, "CODEX_SANDBOX")
	ci, hasCI := Bool(env, "CODEX_CI")
	networkDisabled, hasNetwork := Bool(env, "CODEX_SANDBOX_NETWORK_DISABLED")
	if !hasThreadID && !hasSessionID && !hasSandbox && !(hasCI && ci) {
		return Detection{}, false
	}

	if threadID == "" {
		threadID = sessionID
	}
	// A thread or session identifier is set per Codex conversation, so it is an
	// execution marker. The sandbox and CI variables describe the environment
	// Codex was configured with and could plausibly be set by other tooling.
	confidence := ConfidenceProbable
	if hasThreadID || hasSessionID {
		confidence = ConfidenceDefinite
	}

	network := NetworkUnknown
	if hasNetwork {
		network = NetworkEnabled
		if networkDisabled {
			network = NetworkDisabled
		}
	}

	var extra map[string]string
	if hasCI {
		extra = map[string]string{"codex.ci": boolString(ci)}
	}

	return Detection{
		Agent:      AgentCodex,
		Confidence: confidence,
		SessionID:  threadID,
		Sandbox:    Sandbox{Mode: sandbox, Network: network},
		Extra:      extra,
		Evidence:   PresentNames(env, "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_SANDBOX", "CODEX_SANDBOX_NETWORK_DISABLED", "CODEX_CI"),
	}, true
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
