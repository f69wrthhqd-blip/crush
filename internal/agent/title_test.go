package agent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleFromPrompt(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{
			name:   "plain prompt is kept as is",
			prompt: "Fix the login bug",
			want:   "Fix the login bug",
		},
		{
			name:   "newlines are flattened to spaces",
			prompt: "line one\nline two\nline three",
			want:   "line one line two line three",
		},
		{
			name:   "multiline whitespace is collapsed",
			prompt: "first\n\n\nsecond",
			want:   "first second",
		},
		{
			name:   "thinking tags are stripped",
			prompt: "<think>internal reasoning</think>Refactor the parser",
			want:   "Refactor the parser",
		},
		{
			name:   "orphan thinking tags are stripped",
			prompt: "Summarize </think> this file",
			want:   "Summarize this file",
		},
		{
			name:   "surrounding whitespace is trimmed",
			prompt: "  Add unit tests  ",
			want:   "Add unit tests",
		},
		{
			name:   "long prompt is truncated to 50 display chars",
			prompt: strings.Repeat("a", 100),
			want:   strings.Repeat("a", 49) + "…",
		},
		{
			name:   "long prompt truncation includes the ellipsis",
			prompt: strings.Repeat("b", 60) + "end",
			want:   strings.Repeat("b", 49) + "…",
		},
		{
			name:   "empty prompt falls back to the default name",
			prompt: "",
			want:   DefaultSessionName,
		},
		{
			name:   "whitespace-only prompt falls back to the default name",
			prompt: "   \n\t ",
			want:   DefaultSessionName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TitleFromPrompt(tt.prompt))
		})
	}
}

func TestTitleFromPromptNeverReturnsUntitledForRealInput(t *testing.T) {
	prompt := "Explain the cache invalidation strategy for the session store"
	title := TitleFromPrompt(prompt)
	assert.NotEqual(t, DefaultSessionName, title)
	assert.LessOrEqual(t, len([]rune(title)), 50)
}
