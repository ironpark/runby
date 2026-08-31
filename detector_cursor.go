package runby

func detectCursor(env Env) (Detection, bool) {
	if _, ok := Value(env, "CURSOR_AGENT"); !ok {
		return Detection{}, false
	}
	return Detection{
		Agent:    AgentCursor,
		Evidence: []string{"CURSOR_AGENT"},
	}, true
}
