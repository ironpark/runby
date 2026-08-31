package runby

// detectPaseo identifies a process launched by a Paseo agent. PASEO_AGENT_ID is
// set per agent, so it names the logical agent rather than a single session.
func detectPaseo(env Env) (Detection, bool) {
	agentID, ok := Value(env, "PASEO_AGENT_ID")
	if !ok {
		return Detection{}, false
	}

	workingDirectory, _ := Value(env, "PASEO_AGENT_CWD")
	return Detection{
		Agent:    AgentPaseo,
		AgentID:  agentID,
		Paths:    Paths{WorkingDirectory: workingDirectory},
		Evidence: PresentNames(env, "PASEO_AGENT_ID", "PASEO_AGENT_CWD"),
	}, true
}
