package dialog

import (
	"image"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
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

	cfg := &config.Config{}
	if theme != "" {
		cfg.Options = &config.Options{TUI: &config.TUIOptions{Theme: theme}}
	}
	sty := styles.CharmtonePantera()
	return NewTheme(&common.Common{
		Workspace: &themeTestWorkspace{cfg: cfg},
		Styles:    &sty,
	})
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
}
