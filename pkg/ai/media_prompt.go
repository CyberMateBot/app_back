package ai

import "strings"

// enrichMediaPrompt prepends prior user/assistant turns so follow-up requests keep scene context.
func enrichMediaPrompt(prompt string, messages []ChatMessage) string {
	prompt = strings.TrimSpace(prompt)
	if len(messages) == 0 {
		return prompt
	}

	var b strings.Builder
	for _, m := range messages {
		role := strings.TrimSpace(m.Role)
		content := strings.TrimSpace(m.Content)
		if content == "" {
			content = strings.TrimSpace(m.Text)
		}
		if content == "" {
			continue
		}
		switch role {
		case "user":
			b.WriteString("User: ")
			b.WriteString(content)
			b.WriteString("\n")
		case "assistant":
			b.WriteString("Scene: ")
			b.WriteString(content)
			b.WriteString("\n")
		}
	}

	if b.Len() == 0 {
		return prompt
	}

	b.WriteString("\nCurrent request: ")
	b.WriteString(prompt)
	return b.String()
}
