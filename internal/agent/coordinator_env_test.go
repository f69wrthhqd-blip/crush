package agent

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests mutate the process environment, so they must not run in
// parallel with any other test.

func TestMaskAnthropicAPIKeyEnv(t *testing.T) {
	t.Run("restores previously set value", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-original")

		restore := maskAnthropicAPIKeyEnv()
		require.Empty(t, os.Getenv("ANTHROPIC_API_KEY"))
		restore()

		require.Equal(t, "sk-original", os.Getenv("ANTHROPIC_API_KEY"))
	})

	t.Run("unsets previously absent variable", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "")
		os.Unsetenv("ANTHROPIC_API_KEY")

		restore := maskAnthropicAPIKeyEnv()
		require.Empty(t, os.Getenv("ANTHROPIC_API_KEY"))
		restore()

		_, ok := os.LookupEnv("ANTHROPIC_API_KEY")
		require.False(t, ok, "expected ANTHROPIC_API_KEY to remain unset after restore")
	})
}

// TestMaskAnthropicAPIKeyEnv_Concurrent verifies the mutex keeps concurrent
// mask/restore cycles from clobbering each other's captured value.
func TestMaskAnthropicAPIKeyEnv_Concurrent(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-original")

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			restore := maskAnthropicAPIKeyEnv()
			require.Empty(t, os.Getenv("ANTHROPIC_API_KEY"))
			restore()
		})
	}
	wg.Wait()

	require.Equal(t, "sk-original", os.Getenv("ANTHROPIC_API_KEY"))
}
