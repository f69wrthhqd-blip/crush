package session

import (
	"testing"

	"github.com/charmbracelet/crush/internal/db"
	"github.com/stretchr/testify/require"
)

func TestEstimatedUsageStateSurvivesFetchModifySave(t *testing.T) {
	dataDir := t.TempDir()
	t.Cleanup(func() {
		require.NoError(t, db.Release(dataDir))
		db.ResetPool()
	})

	conn, err := db.Connect(t.Context(), dataDir)
	require.NoError(t, err)

	sessions := NewService(db.New(conn), conn)

	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)
	created.PromptTokens = 100
	created.CompletionTokens = 50
	created.EstimatedUsage = true

	saved, err := sessions.Save(t.Context(), created)
	require.NoError(t, err)
	require.True(t, saved.EstimatedUsage)

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.True(t, fetched.EstimatedUsage)

	fetched.Todos = []Todo{{
		Content:    "Check estimate state",
		Status:     TodoStatusInProgress,
		ActiveForm: "Checking estimate state",
	}}

	updated, err := sessions.Save(t.Context(), fetched)
	require.NoError(t, err)
	require.True(t, updated.EstimatedUsage)

	refetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.True(t, refetched.EstimatedUsage)
}

func TestEstimatedUsageStateCanBeClearedByExplicitSave(t *testing.T) {
	dataDir := t.TempDir()
	t.Cleanup(func() {
		require.NoError(t, db.Release(dataDir))
		db.ResetPool()
	})

	conn, err := db.Connect(t.Context(), dataDir)
	require.NoError(t, err)

	sessions := NewService(db.New(conn), conn)

	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)
	created.PromptTokens = 100
	created.CompletionTokens = 50
	created.EstimatedUsage = true

	saved, err := sessions.Save(t.Context(), created)
	require.NoError(t, err)
	require.True(t, saved.EstimatedUsage)

	saved.EstimatedUsage = false
	updated, err := sessions.Save(t.Context(), saved)
	require.NoError(t, err)
	require.False(t, updated.EstimatedUsage)

	refetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.False(t, refetched.EstimatedUsage)
}

// newColumnUpdateTestService creates a service backed by a fresh
// temporary database for the column-level update tests.
func newColumnUpdateTestService(t *testing.T) *service {
	t.Helper()
	dataDir := t.TempDir()
	t.Cleanup(func() {
		require.NoError(t, db.Release(dataDir))
		db.ResetPool()
	})

	conn, err := db.Connect(t.Context(), dataDir)
	require.NoError(t, err)

	return NewService(db.New(conn), conn).(*service)
}

func TestUpdateTodosPersistsListAtomically(t *testing.T) {
	sessions := newColumnUpdateTestService(t)

	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)

	todos := []Todo{
		{Content: "First", Status: TodoStatusCompleted, ActiveForm: "Finishing first"},
		{Content: "Second", Status: TodoStatusInProgress, ActiveForm: "Doing second"},
		{Content: "Third", Status: TodoStatusPending, ActiveForm: "Doing third"},
	}
	require.NoError(t, sessions.UpdateTodos(t.Context(), created.ID, todos))

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Equal(t, todos, fetched.Todos)

	// Clearing writes an empty list back to the row.
	require.NoError(t, sessions.UpdateTodos(t.Context(), created.ID, nil))
	refetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Empty(t, refetched.Todos)
}

func TestUpdateUsagePreservesZeroCountersAndAccumulatesCost(t *testing.T) {
	sessions := newColumnUpdateTestService(t)

	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)
	created.PromptTokens = 123
	created.CompletionTokens = 456
	created.Cost = 1.25
	created.EstimatedUsage = true
	_, err = sessions.Save(t.Context(), created)
	require.NoError(t, err)

	// Zero token counters leave the stored values untouched; cost is
	// accumulated; the estimated marker clears for real usage.
	require.NoError(t, sessions.UpdateUsage(t.Context(), created.ID, 0, 2000, 0.05, false))

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Equal(t, int64(123), fetched.PromptTokens)
	require.Equal(t, int64(2000), fetched.CompletionTokens)
	require.InDelta(t, 1.30, fetched.Cost, 1e-9)
	require.False(t, fetched.EstimatedUsage)

	// Non-zero counters overwrite; estimated usage re-marks the session
	// without accruing cost.
	require.NoError(t, sessions.UpdateUsage(t.Context(), created.ID, 789, 0, 0, true))

	refetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Equal(t, int64(789), refetched.PromptTokens)
	require.Equal(t, int64(2000), refetched.CompletionTokens)
	require.InDelta(t, 1.30, refetched.Cost, 1e-9)
	require.True(t, refetched.EstimatedUsage)
}

func TestUpdateSessionSummaryResetsCountersAndRecordsMessage(t *testing.T) {
	sessions := newColumnUpdateTestService(t)

	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)
	created.PromptTokens = 5000
	created.CompletionTokens = 1000
	created.Cost = 2.0
	_, err = sessions.Save(t.Context(), created)
	require.NoError(t, err)

	require.NoError(t, sessions.UpdateSessionSummary(t.Context(), created.ID, "summary-msg", 0, 42, 0.5, false))

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "summary-msg", fetched.SummaryMessageID)
	require.Zero(t, fetched.PromptTokens)
	require.Equal(t, int64(42), fetched.CompletionTokens)
	require.InDelta(t, 2.5, fetched.Cost, 1e-9)
	require.False(t, fetched.EstimatedUsage)
}

func TestAddSessionCostAccumulatesAndErrorsOnMissingSession(t *testing.T) {
	sessions := newColumnUpdateTestService(t)

	created, err := sessions.Create(t.Context(), "test")
	require.NoError(t, err)
	created.Cost = 1.0
	_, err = sessions.Save(t.Context(), created)
	require.NoError(t, err)

	require.NoError(t, sessions.AddSessionCost(t.Context(), created.ID, 0.75))

	fetched, err := sessions.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.InDelta(t, 1.75, fetched.Cost, 1e-9)

	require.Error(t, sessions.AddSessionCost(t.Context(), "non-existent", 1.0))
}
