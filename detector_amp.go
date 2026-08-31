package runby

func detectAmp(env Env) (Detection, bool) {
	threadID, hasThreadID := Value(env, "AMP_THREAD_ID")
	if !IsTrue(env, "AMP_ORB") && !hasThreadID {
		return Detection{}, false
	}

	entrypoint := "orb"
	if hasThreadID {
		entrypoint = "orb-service"
	}
	return Detection{
		Agent:      AgentAmp,
		SessionID:  threadID,
		Entrypoint: entrypoint,
		Evidence:   PresentNames(env, "AMP_ORB", "AMP_THREAD_ID"),
	}, true
}
