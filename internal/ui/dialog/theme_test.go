package dialog

import (
	"image"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/crush/internal/workspace"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type themeTestWorkspace struct {
	workspace.Workspace
	cfg *config.Config
}

func (w *themeTestWorkspace) Config() *config.Config { return w.cfg }

func newThemeTestDialog(t *testing.T, theme string) *Theme {
	t.Helper()
	return newThemeTestDialogWith(t, theme, false, nil)
}

func newThemeTestDialogWith(t *testing.T, theme string, transparent bool, termBgDark *bool) *Theme {
	t.Helper()

	cfg := &config.Config{
		Options: &config.Options{TUI: &config.TUIOptions{Theme: theme, Transparent: &transparent}},
	}
	sty := styles.CharmtonePantera()
	return NewTheme(&common.Common{
		Workspace: &themeTestWorkspace{cfg: cfg},
		Styles:    &sty,
	}, termBgDark)
}

// themeItemsByKey maps the dialog's items by theme key; the auto entry
// lives under the empty key.
func themeItemsByKey(t *testing.T, d *Theme) map[string]*ThemeItem {
	t.Helper()
	byKey := make(map[string]*ThemeItem)
	for _, item := range d.list.FilteredItems() {
		themeItem, ok := item.(*ThemeItem)
		require.True(t, ok)
		byKey[themeItem.key] = themeItem
	}
	return byKey
}

func TestThemeDialog(t *testing.T) {
	t.Parallel()

	t.Run("lists auto plus every registered theme", func(t *testing.T) {
		t.Parallel()
		d := newThemeTestDialog(t, "")
		items := d.list.FilteredItems()
		require.Len(t, items, len(styles.AvailableThemes())+1)

		first, ok := items[0].(*ThemeItem)
		require.True(t, ok)
		require.Empty(t, first.key, "first item must be the auto entry")
	})

	t.Run("preselects the configured theme", func(t *testing.T) {
		t.Parallel()
		d := newThemeTestDialog(t, "nord")
		selected, ok := d.list.SelectedItem().(*ThemeItem)
		require.True(t, ok)
		require.Equal(t, "nord", selected.key)
		require.True(t, selected.isCurrent)
	})

	t.Run("preselects auto when no theme configured", func(t *testing.T) {
		t.Parallel()
		d := newThemeTestDialog(t, "")
		selected, ok := d.list.SelectedItem().(*ThemeItem)
		require.True(t, ok)
		require.Empty(t, selected.key)
		require.True(t, selected.isCurrent)
	})

	t.Run("enter returns the selected theme", func(t *testing.T) {
		t.Parallel()
		d := newThemeTestDialog(t, "dracula")
		action := d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
		selectAction, ok := action.(ActionSelectTheme)
		require.True(t, ok)
		require.Equal(t, "dracula", selectAction.Key)
	})

	t.Run("renders without overflow", func(t *testing.T) {
		t.Parallel()
		d := newThemeTestDialog(t, "catppuccin-mocha")
		scr := uv.NewScreenBuffer(80, 24)
		require.NotPanics(t, func() {
			d.Draw(scr, image.Rect(0, 0, 80, 24))
		})
	})

	t.Run("flags light themes as low contrast over a dark terminal", func(t *testing.T) {
		t.Parallel()
		dark := true
		d := newThemeTestDialogWith(t, "", true, &dark)
		byKey := themeItemsByKey(t, d)

		for _, key := range []string{"catppuccin-latte", "solarized-light"} {
			require.True(t, byKey[key].contrastWarn, "light theme %q must be flagged", key)
			require.Contains(t, byKey[key].Render(80), i18n.T("theme.contrast_low"))
		}
		for _, key := range []string{"nord", "gruvbox-dark", "pantera"} {
			require.False(t, byKey[key].contrastWarn, "dark theme %q must not be flagged", key)
		}
	})

	t.Run("flags dark themes as low contrast over a light terminal", func(t *testing.T) {
		t.Parallel()
		light := false
		d := newThemeTestDialogWith(t, "", true, &light)
		byKey := themeItemsByKey(t, d)

		require.True(t, byKey["nord"].contrastWarn)
		require.False(t, byKey["catppuccin-latte"].contrastWarn)
	})

	t.Run("no flags when transparency is off or terminal bg unknown", func(t *testing.T) {
		t.Parallel()
		dark := true

		d := newThemeTestDialogWith(t, "", false, &dark)
		for key, item := range themeItemsByKey(t, d) {
			require.False(t, item.contrastWarn, "item %q must not be flagged while opaque", key)
		}

		d = newThemeTestDialogWith(t, "", true, nil)
		for key, item := range themeItemsByKey(t, d) {
			require.False(t, item.contrastWarn, "item %q must not be flagged with unknown terminal bg", key)
		}
	})
}
