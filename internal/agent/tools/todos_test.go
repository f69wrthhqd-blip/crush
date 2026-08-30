package tools

import (
	"context"
	"encoding/json"
	"testing"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/db"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/stretchr/testify/require"
)

func newTodosTestService(t *testing.T) session.Service {
	t.Helper()
	dataDir := t.TempDir()
	t.Cleanup(func() {
		require.NoError(t, db.Release(dataDir))
		db.ResetPool()
	})

	conn, err := db.Connect(t.Context(), dataDir)
	require.NoError(t, err)

	return session.NewService(db.New(conn), conn)
}

func todosToolCallInput(t *testing.T, todos []TodoItem) string {
	t.Helper()
	b, err := json.Marshal(TodosParams{Todos: todos})
	require.NoError(t, err)
	return string(b)
}

func TestTodosTool_RejectsEmptyList(t *testing.T) {
	t.Parallel()

	sessions := newTodosTestService(t)
	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)

	// Seed an existing list so the rejection is observable.
	require.NoError(t, sessions.UpdateTodos(t.Context(), created.ID, []session.Todo{
		{Content: "existing task", Status: session.TodoStatusPending, ActiveForm: "Doing existing task"},
	}))

	tool := NewTodosTool(sessions)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, created.ID)

	resp, err := tool.Run(ctx, fantasy.ToolCall{
		ID:    "call-1",
		Name:  TodosToolName,
		Input: todosToolCallInput(t, nil),
	})
	require.NoError(t, err)
	require.True(t, resp.IsError, "empty todo list must be rejected")

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Len(t, fetched.Todos, 1, "existing todos must survive a rejected empty update")
}

func TestTodosTool_SavesListAtomically(t *testing.T) {
	t.Parallel()

	sessions := newTodosTestService(t)
	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)

	tool := NewTodosTool(sessions)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, created.ID)

	resp, err := tool.Run(ctx, fantasy.ToolCall{
		ID:   "call-1",
		Name: TodosToolName,
		Input: todosToolCallInput(t, []TodoItem{
			{Content: "first", Status: "completed", ActiveForm: "Finishing first"},
			{Content: "second", Status: "in_progress", ActiveForm: "Doing second"},
			{Content: "third", Status: "pending", ActiveForm: "Doing third"},
		}),
	})
	require.NoError(t, err)
	require.False(t, resp.IsError)

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Len(t, fetched.Todos, 3)
	require.Equal(t, session.TodoStatusCompleted, fetched.Todos[0].Status)
	require.Equal(t, session.TodoStatusInProgress, fetched.Todos[1].Status)
	require.Equal(t, session.TodoStatusPending, fetched.Todos[2].Status)
}

func TestTodosTool_RejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	sessions := newTodosTestService(t)
	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)

	tool := NewTodosTool(sessions)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, created.ID)

	resp, err := tool.Run(ctx, fantasy.ToolCall{
		ID:   "call-1",
		Name: TodosToolName,
		Input: todosToolCallInput(t, []TodoItem{
			{Content: "task", Status: "done", ActiveForm: "Doing task"},
		}),
	})
	require.Error(t, err, "invalid status must be rejected")
	require.False(t, resp.IsError)

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Empty(t, fetched.Todos)
}
