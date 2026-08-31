package runby

import "strings"

func detectClaudeCode(env Env) (Detection, bool) {
	sessionID, hasSessionID := Value(env, "CLAUDE_CODE_SESSION_ID")
	aiAgent, hasAIAgent := Value(env, "AI_AGENT")
	isAIAgent := hasAIAgent && strings.HasPrefix(strings.ToLower(aiAgent), "claude-code")
	if !IsTrue(env, "CLAUDECODE") && !hasSessionID && !isAIAgent {
		return Detection{}, false
	}

	entrypoint, _ := Value(env, "CLAUDE_CODE_ENTRYPOINT")
	childSession, _ := Bool(env, "CLAUDE_CODE_CHILD_SESSION")
	// AI_AGENT is only evidence when its value names Claude Code, so it joins
	// the candidate list conditionally; PresentNames still does the sorting.
	names := []string{"CLAUDECODE", "CLAUDE_CODE_SESSION_ID", "CLAUDE_CODE_ENTRYPOINT", "CLAUDE_CODE_CHILD_SESSION"}
	if isAIAgent {
		names = append(names, "AI_AGENT")
	}
	return Detection{
		Agent:      AgentClaudeCode,
		SessionID:  sessionID,
		Entrypoint: entrypoint,
		Nested:     childSession,
		Evidence:   PresentNames(env, names...),
	}, true
}
