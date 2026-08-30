package agent

import (
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/stretchr/testify/require"
)

func todoFixture(status session.TodoStatus, content string) session.Todo {
	return session.Todo{Content: content, Status: status, ActiveForm: content + " (active)"}
}

func TestTodoNudgePrompt(t *testing.T) {
	t.Parallel()

	incomplete := []session.Todo{
		todoFixture(session.TodoStatusCompleted, "done task"),
		todoFixture(session.TodoStatusInProgress, "current task"),
		todoFixture(session.TodoStatusPending, "later task"),
	}

	t.Run("nudges on natural end with incomplete todos", func(t *testing.T) {
		t.Parallel()
		prompt, ok := todoNudgePrompt(incomplete, message.FinishReasonEndTurn, false, 0)
		require.True(t, ok)
		require.Contains(t, prompt, "current task")
		require.Contains(t, prompt, "later task")
		require.NotContains(t, prompt, "done task")
	})

	t.Run("no nudge when all todos are completed", func(t *testing.T) {
		t.Parallel()
		_, ok := todoNudgePrompt([]session.Todo{todoFixture(session.TodoStatusCompleted, "done")}, message.FinishReasonEndTurn, false, 0)
		require.False(t, ok)
	})

	t.Run("no nudge when already nudged once", func(t *testing.T) {
		t.Parallel()
		_, ok := todoNudgePrompt(incomplete, message.FinishReasonEndTurn, false, 1)
		require.False(t, ok)
	})

	t.Run("no nudge when a tool result halted the turn", func(t *testing.T) {
		t.Parallel()
		_, ok := todoNudgePrompt(incomplete, message.FinishReasonEndTurn, true, 0)
		require.False(t, ok)
	})

	t.Run("no nudge for non-natural finish reasons", func(t *testing.T) {
		t.Parallel()
		for _, reason := range []message.FinishReason{
			message.FinishReasonToolUse,
			message.FinishReasonCanceled,
			message.FinishReasonUnknown,
		} {
			_, ok := todoNudgePrompt(incomplete, reason, false, 0)
			require.False(t, ok, "reason %v should not nudge", reason)
		}
	})
}

func TestTodoListReminder(t *testing.T) {
	t.Parallel()

	t.Run("empty list keeps the create-one invitation", func(t *testing.T) {
		t.Parallel()
		reminder := todoListReminder(nil)
		require.Contains(t, reminder, "todo list is currently empty")
	})

	t.Run("non-empty list restates real progress", func(t *testing.T) {
		t.Parallel()
		reminder := todoListReminder([]session.Todo{
			todoFixture(session.TodoStatusCompleted, "done task"),
			todoFixture(session.TodoStatusInProgress, "current task"),
			todoFixture(session.TodoStatusPending, "later task"),
		})
		require.Contains(t, reminder, "1 completed, 1 in progress, 1 pending")
		require.Contains(t, reminder, "- current task")
		require.Contains(t, reminder, "- later task")
		require.Contains(t, reminder, `the "todos" tool`)
		require.False(t, strings.Contains(reminder, "currently empty"))
	})
}
