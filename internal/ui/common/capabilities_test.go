package common

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestTerminalBackgroundIsDark(t *testing.T) {
	t.Parallel()

	t.Run("unknown when no background reported", func(t *testing.T) {
		t.Parallel()
		var caps Capabilities
		isDark, known := caps.TerminalBackgroundIsDark()
		require.False(t, known)
		require.False(t, isDark)
	})

	t.Run("reports direction after a background message", func(t *testing.T) {
		t.Parallel()

		var caps Capabilities
		caps.Update(tea.BackgroundColorMsg{Color: color.Black})
		isDark, known := caps.TerminalBackgroundIsDark()
		require.True(t, known)
		require.True(t, isDark)

		caps.Update(tea.BackgroundColorMsg{Color: color.White})
		isDark, known = caps.TerminalBackgroundIsDark()
		require.True(t, known)
		require.False(t, isDark)
	})
}
