package styles

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAvailableThemes(t *testing.T) {
	t.Parallel()

	themes := AvailableThemes()
	require.NotEmpty(t, themes)

	seen := make(map[string]bool, len(themes))
	for _, theme := range themes {
		require.NotEmpty(t, theme.Key)
		require.NotEmpty(t, theme.Name)
		require.NotNil(t, theme.Build)
		require.False(t, seen[theme.Key], "duplicate theme key %q", theme.Key)
		seen[theme.Key] = true

		// Every listed theme must be resolvable through the registry and
		// build without panicking.
		_, ok := ThemeByKey(theme.Key)
		require.True(t, ok, "theme %q missing from registry", theme.Key)
	}

	// The default theme must be part of the picker.
	require.True(t, seen["pantera"])
}

func TestThemeByKey(t *testing.T) {
	t.Parallel()

	t.Run("known key resolves", func(t *testing.T) {
		t.Parallel()
		s, ok := ThemeByKey("gruvbox-dark")
		require.True(t, ok)
		require.NotNil(t, s.Dialog.View)
	})

	t.Run("provider alias resolves", func(t *testing.T) {
		t.Parallel()
		_, ok := ThemeByKey("hyper")
		require.True(t, ok)
	})

	t.Run("unknown key falls back to default", func(t *testing.T) {
		t.Parallel()
		s, ok := ThemeByKey("not-a-theme")
		require.False(t, ok)
		require.Equal(t, CharmtonePantera(), s)
	})
}

func TestEffectiveThemeKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configKey  string
		providerID string
		want       string
	}{
		{"config wins over provider", "nord", "hyper", "nord"},
		{"empty config follows provider", "", "hyper", "hyper"},
		{"empty config defaults", "", "anthropic", "pantera"},
		{"unknown config follows provider", "bogus", "hyper", "hyper"},
		{"unknown config defaults", "bogus", "", "pantera"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, EffectiveThemeKey(tt.configKey, tt.providerID))
		})
	}
}

func TestThemeFromOptions(t *testing.T) {
	t.Parallel()

	t.Run("user theme wins", func(t *testing.T) {
		t.Parallel()
		s, key := ThemeFromOptions("dracula", "hyper")
		require.Equal(t, "dracula", key)
		require.Equal(t, Dracula(), s)
	})

	t.Run("provider fallback", func(t *testing.T) {
		t.Parallel()
		s, key := ThemeFromOptions("", "hyper")
		require.Equal(t, "hyper", key)
		require.Equal(t, HypercrushObsidiana(), s)
	})

	t.Run("default fallback", func(t *testing.T) {
		t.Parallel()
		s, key := ThemeFromOptions("", "")
		require.Equal(t, "pantera", key)
		require.Equal(t, CharmtonePantera(), s)
	})
}

func TestThemeName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "Nord", ThemeName("nord"))
	require.Equal(t, "bogus", ThemeName("bogus"))
}

// lightThemeBgLuminance asserts that the light themes ship with a light
// background and dark foreground: a swap of the dark-theme contrast
// direction. Guard against regressions when tuning the palettes.
func lightThemeBgLuminance(t *testing.T, s Styles) {
	t.Helper()
	// Background carries the theme's bgBase; the editor's base text
	// style carries fgBase.
	bg := s.Background
	fg := s.TextInput.Focused.Text.GetForeground()
	require.NotNil(t, bg, "light theme must set a background")
	require.NotNil(t, fg, "light theme must set a foreground")
	require.Greater(t, luminance(bg), luminance(fg),
		"light theme background must be lighter than its foreground")
}

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Rec. 709 relative luminance over 8-bit channels.
	return 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8)
}

func TestLightThemes(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"catppuccin-latte", "solarized-light"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			s, _ := ThemeByKey(key)
			lightThemeBgLuminance(t, s)
		})
	}
}
