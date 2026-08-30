package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestConfig_TUITheme(t *testing.T) {
	t.Parallel()

	t.Run("loads theme from options", func(t *testing.T) {
		t.Parallel()
		cfg, err := loadFromBytes([][]byte{[]byte(`{"options": {"tui": {"theme": "nord"}}}`)})
		require.NoError(t, err)
		require.Equal(t, "nord", cfg.TUITheme())
	})

	t.Run("empty when unset", func(t *testing.T) {
		t.Parallel()
		cfg, err := loadFromBytes([][]byte{[]byte(`{}`)})
		require.NoError(t, err)
		require.Empty(t, cfg.TUITheme())
	})

	t.Run("nil-safe on empty config", func(t *testing.T) {
		t.Parallel()
		var cfg *Config
		require.Empty(t, cfg.TUITheme())
	})
}

// TestSetConfigField_TUITheme verifies that the theme picker's
// options.tui.theme write lands in the config file and auto-reloads the
// in-memory snapshot.
func TestSetConfigField_TUITheme(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "crush.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"options": {"tui": {"locale": "en"}}}`), 0o600))

	store, err := Load(dir, dir, false)
	require.NoError(t, err)

	store.globalDataPath = configPath
	store.CaptureStalenessSnapshot([]string{configPath})

	require.NoError(t, store.SetConfigField(ScopeGlobal, "options.tui.theme", "gruvbox-dark"))

	require.Equal(t, "gruvbox-dark", store.config.TUITheme(),
		"expected config to auto-reload with the new theme")

	raw, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Equal(t, "gruvbox-dark", gjson.GetBytes(raw, "options.tui.theme").String(),
		"expected options.tui.theme persisted to the config file")

	// Selecting "auto" clears the key back to an empty string.
	require.NoError(t, store.SetConfigField(ScopeGlobal, "options.tui.theme", ""))
	require.Empty(t, store.config.TUITheme())
}
