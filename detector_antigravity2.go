package runby

// detectAntigravity2 identifies a sidecar whose lifecycle Antigravity 2.0
// manages. Antigravity CLI sets no general execution marker and is not detected.
func detectAntigravity2(env Env) (Detection, bool) {
	dataDirectory, ok := Value(env, "ANTIGRAVITY_EXECUTABLE_DATA_DIR")
	if !ok {
		return Detection{}, false
	}
	return Detection{
		Agent:      AgentAntigravity2,
		Entrypoint: "sidecar",
		Paths:      Paths{DataDirectory: dataDirectory},
		Evidence:   []string{"ANTIGRAVITY_EXECUTABLE_DATA_DIR"},
	}, true
}
