package runby

// detectOpenCode identifies OpenCode running as an ACP client. OpenCode has no
// general execution marker, so this covers ACP invocations only and is reported
// as a supporting signal rather than proof.
func detectOpenCode(env Env) (Detection, bool) {
	if !EqualsFold(env, "OPENCODE_CLIENT", "acp") {
		return Detection{}, false
	}
	return Detection{
		Agent:      AgentOpenCode,
		Confidence: ConfidenceProbable,
		Entrypoint: "acp",
		Evidence:   []string{"OPENCODE_CLIENT"},
	}, true
}
