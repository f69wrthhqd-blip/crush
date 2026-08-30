package dialog

import (
	"image"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/crush/internal/workspace"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
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

	t.Run("renders the list on first open with the auto entry selected", func(t *testing.T) {
		t.Parallel()
		// Regression: a fresh FilterableList starts with a -1 scroll
		// offset (from its constructor-time empty SetItems). When the
		// selected item was already in view, ScrollToSelected never
		// corrected it and the whole list rendered blank.
		d := newThemeTestDialog(t, "")
		scr := uv.NewScreenBuffer(113, 37)
		d.Draw(scr, image.Rect(0, 0, 113, 37))

		out := d.list.Render()
		require.NotEmpty(t, out, "list must render on first open")
		for i, line := range strings.Split(out, "\n") {
			require.Greater(t, ansi.StringWidth(line), 0,
				"viewport row %d is blank", i)
		}
	})

	t.Run("fills the viewport when the selection is near the end", func(t *testing.T) {
		t.Parallel()
		// Regression: opening the picker with a late-list theme selected
		// used to pin the selection to the top of the viewport, leaving
		// blank rows below the last item.
		d := newThemeTestDialog(t, "github-light")
		scr := uv.NewScreenBuffer(80, 24)
		d.Draw(scr, image.Rect(0, 0, 80, 24))

		rendered := strings.Split(d.list.Render(), "\n")
		require.NotEmpty(t, rendered)
		for i, line := range rendered {
			require.Greater(t, ansi.StringWidth(line), 0,
				"viewport row %d is blank", i)
		}
		// The viewport must be full: rows == viewport height while items
		// remain (22 entries, height < 22 here).
		require.Len(t, rendered, d.list.Height())
	})

	t.Run("focused row keeps its highlight across the swatches", func(t *testing.T) {
		t.Parallel()
		// Regression: the swatch chips' embedded color resets used to
		// drop the focused row's highlight background for everything
		// rendered after them.
		d := newThemeTestDialog(t, "nord")
		selected, ok := d.list.SelectedItem().(*ThemeItem)
		require.True(t, ok)

		// Draw sizes the viewport and runs the list's render callback,
		// which is what marks the selected item focused.
		d.Draw(uv.NewScreenBuffer(80, 24), image.Rect(0, 0, 80, 24))

		focused := selected.Render(80)
		require.GreaterOrEqual(t, strings.Count(focused, "\x1b[48;2;"), 6,
			"focused row must re-apply its background after the swatch chips")

		first, ok := d.list.FilteredItems()[0].(*ThemeItem)
		require.True(t, ok)
		require.Equal(t, 0, strings.Count(first.Render(80), "\x1b[48;2;"),
			"blurred rows carry no explicit background")
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
