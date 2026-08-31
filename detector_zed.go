package runby

// detectZed identifies a Zed-owned terminal. Zed Agent sets no marker of its
// own, so this proves which application owns the terminal, not that Zed Agent
// rather than a person ran the command. It is reported as KindHost for that
// reason, and IsAgent stays false when it is the only layer.
func detectZed(env Env) (Detection, bool) {
	if !IsTrue(env, "ZED_TERM") || !EqualsFold(env, "TERM_PROGRAM", "zed") {
		return Detection{}, false
	}

	var extra map[string]string
	if version, ok := Value(env, "TERM_PROGRAM_VERSION"); ok {
		extra = map[string]string{"zed.version": version}
	}
	return Detection{
		Agent:      AgentZed,
		Confidence: ConfidenceProbable,
		Entrypoint: "terminal",
		Extra:      extra,
		Evidence:   PresentNames(env, "ZED_TERM", "TERM_PROGRAM", "TERM_PROGRAM_VERSION"),
	}, true
}
