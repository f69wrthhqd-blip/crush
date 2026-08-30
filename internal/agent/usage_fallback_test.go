package agent

import (
	"errors"
	"testing"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/stretchr/testify/require"
)

func TestUsageIsZero(t *testing.T) {
	t.Parallel()

	require.True(t, usageIsZero(fantasy.Usage{}))
	require.False(t, usageIsZero(fantasy.Usage{InputTokens: 1}))
	require.False(t, usageIsZero(fantasy.Usage{OutputTokens: 1}))
	require.False(t, usageIsZero(fantasy.Usage{TotalTokens: 1}))
	require.False(t, usageIsZero(fantasy.Usage{ReasoningTokens: 1}))
	require.False(t, usageIsZero(fantasy.Usage{CacheCreationTokens: 1}))
	require.False(t, usageIsZero(fantasy.Usage{CacheReadTokens: 1}))
}

func TestFallbackStepUsageKeepsProviderUsage(t *testing.T) {
	t.Parallel()

	usage := fantasy.Usage{
		InputTokens:  10,
		OutputTokens: 5,
		TotalTokens:  15,
	}
	step := fantasy.StepResult{
		Response: fantasy.Response{Usage: usage},
	}

	fallbackUsage, estimated := fallbackStepUsage(nil, step)
	require.False(t, estimated)
	require.Equal(t, usage, fallbackUsage)
}

func TestFallbackStepUsageEstimatesPromptAndAssistantText(t *testing.T) {
	t.Parallel()

	messages := []fantasy.Message{
		fantasy.NewUserMessage("please explain the implementation details"),
	}
	step := fantasy.StepResult{
		Response: fantasy.Response{
			Content: fantasy.ResponseContent{
				fantasy.TextContent{Text: "the implementation stores state safely"},
			},
		},
	}

	usage, estimated := fallbackStepUsage(messages, step)
	require.True(t, estimated)
	require.Positive(t, usage.InputTokens)
	require.Positive(t, usage.OutputTokens)
	require.Equal(t, usage.InputTokens+usage.OutputTokens, usage.TotalTokens)
}

func TestFallbackStepUsageEstimatesReasoning(t *testing.T) {
	t.Parallel()

	messages := []fantasy.Message{
		{
			Role: fantasy.MessageRoleAssistant,
			Content: []fantasy.MessagePart{
				fantasy.ReasoningPart{Text: "first reason about the request"},
			},
		},
	}
	step := fantasy.StepResult{
		Response: fantasy.Response{
			Content: fantasy.ResponseContent{
				fantasy.ReasoningContent{Text: "second reason about the answer"},
			},
		},
	}

	usage, estimated := fallbackStepUsage(messages, step)
	require.True(t, estimated)
	require.Positive(t, usage.InputTokens)
	require.Positive(t, usage.OutputTokens)
}

func TestFallbackStepUsageEstimatesToolCalls(t *testing.T) {
	t.Parallel()

	step := fantasy.StepResult{
		Response: fantasy.Response{
			Content: fantasy.ResponseContent{
				fantasy.ToolCallContent{
					ToolCallID: "tool-call-1",
					ToolName:   "view",
					Input:      `{"file_path":"/tmp/example.go"}`,
				},
			},
		},
	}

	usage, estimated := fallbackStepUsage(nil, step)
	require.True(t, estimated)
	require.Zero(t, usage.InputTokens)
	require.Positive(t, usage.OutputTokens)
	require.Equal(t, usage.OutputTokens, usage.TotalTokens)
}

func TestFallbackStepUsageEstimatesToolResults(t *testing.T) {
	t.Parallel()

	messages := []fantasy.Message{
		{
			Role: fantasy.MessageRoleTool,
			Content: []fantasy.MessagePart{
				fantasy.ToolResultPart{
					ToolCallID: "tool-call-1",
					Output: fantasy.ToolResultOutputContentText{
						Text: "file contents returned by the tool",
					},
				},
				fantasy.ToolResultPart{
					ToolCallID: "tool-call-2",
					Output: fantasy.ToolResultOutputContentError{
						Error: errors.New("permission denied"),
					},
				},
				fantasy.ToolResultPart{
					ToolCallID: "tool-call-3",
					Output: fantasy.ToolResultOutputContentMedia{
						MediaType: "image/png",
						Text:      "screenshot",
						Data:      "abc123",
					},
				},
			},
		},
	}

	usage, estimated := fallbackStepUsage(messages, fantasy.StepResult{})
	require.True(t, estimated)
	require.Positive(t, usage.InputTokens)
	require.Zero(t, usage.OutputTokens)
	require.Equal(t, usage.InputTokens, usage.TotalTokens)
}

func TestFallbackStepUsageSkipsClientToolResultsAsOutput(t *testing.T) {
	t.Parallel()

	step := fantasy.StepResult{
		Response: fantasy.Response{
			Content: fantasy.ResponseContent{
				fantasy.ToolResultContent{
					ToolCallID: "tool-call-1",
					ToolName:   "bash",
					Result: fantasy.ToolResultOutputContentText{
						Text: "large client-executed payload that should not count as model output tokens",
					},
				},
			},
		},
	}

	usage, estimated := fallbackStepUsage(nil, step)
	require.False(t, estimated)
	require.Zero(t, usage.OutputTokens)
}

func TestFallbackStepUsageCountsProviderToolResultsAsOutput(t *testing.T) {
	t.Parallel()

	step := fantasy.StepResult{
		Response: fantasy.Response{
			Content: fantasy.ResponseContent{
				fantasy.ToolResultContent{
					ToolCallID:       "tool-call-1",
					ToolName:         "web_search",
					ProviderExecuted: true,
					ClientMetadata:   "provider metadata",
					Result:           fantasy.ToolResultOutputContentText{Text: "provider-executed result"},
				},
			},
		},
	}

	usage, estimated := fallbackStepUsage(nil, step)
	require.True(t, estimated)
	require.Positive(t, usage.OutputTokens)
	require.Equal(t, usage.OutputTokens, usage.TotalTokens)
}

func TestFallbackStepUsageReturnsZeroWithoutContent(t *testing.T) {
	t.Parallel()

	usage, estimated := fallbackStepUsage(nil, fantasy.StepResult{})
	require.False(t, estimated)
	require.True(t, usageIsZero(usage))
}

func TestSessionUsageCostSkipsEstimatedCost(t *testing.T) {
	t.Parallel()

	agent := &sessionAgent{}
	model := Model{CatwalkCfg: catwalk.Model{CostPer1MIn: 10, CostPer1MOut: 20}}
	usage := fantasy.Usage{InputTokens: 1000, OutputTokens: 2000}

	require.Zero(t, agent.sessionUsageCost(model, "session-id", usage, nil, true))
}

func TestSessionUsageCostComputesFromModelRates(t *testing.T) {
	t.Parallel()

	agent := &sessionAgent{}
	model := Model{CatwalkCfg: catwalk.Model{CostPer1MIn: 10, CostPer1MOut: 20}}
	usage := fantasy.Usage{InputTokens: 1000, OutputTokens: 2000}

	// 1000 input tokens at 10/1M plus 2000 output tokens at 20/1M.
	require.Equal(t, 0.05, agent.sessionUsageCost(model, "session-id", usage, nil, false))
}

func TestSessionUsageCostAppliesProviderOverride(t *testing.T) {
	t.Parallel()

	agent := &sessionAgent{}
	model := Model{CatwalkCfg: catwalk.Model{CostPer1MIn: 10, CostPer1MOut: 20}}
	usage := fantasy.Usage{InputTokens: 1000, OutputTokens: 2000}
	override := 9.99

	require.Equal(t, 9.99, agent.sessionUsageCost(model, "session-id", usage, &override, false))
}

func TestSessionUsageCostSkipsFlatRate(t *testing.T) {
	t.Parallel()

	agent := &sessionAgent{}
	model := Model{CatwalkCfg: catwalk.Model{CostPer1MIn: 10, CostPer1MOut: 20}, FlatRate: true}
	usage := fantasy.Usage{InputTokens: 1000, OutputTokens: 2000}

	require.Zero(t, agent.sessionUsageCost(model, "session-id", usage, nil, false))
}

func TestSummaryCompletionTokens(t *testing.T) {
	t.Parallel()

	summaryMessage := message.Message{
		Parts: []message.ContentPart{
			message.TextContent{Text: "summary text"},
			message.ReasoningContent{Thinking: "reasoning text"},
		},
	}

	require.Equal(t, int64(42), summaryCompletionTokens(fantasy.Usage{OutputTokens: 42}, summaryMessage))
	require.Equal(t, approxTokenCount("summary text")+approxTokenCount("reasoning text"), summaryCompletionTokens(fantasy.Usage{}, summaryMessage))
	require.Zero(t, summaryCompletionTokens(fantasy.Usage{}, message.Message{}))
}
