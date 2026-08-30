package dialog

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"
)

const (
	// ThemeID is the identifier for the theme picker dialog.
	ThemeID              = "theme"
	themeDialogMaxWidth  = 62
	themeDialogMaxHeight = 18
)

// Theme represents a dialog for selecting the UI color theme.
type Theme struct {
	com   *common.Common
	help  help.Model
	list  *list.FilterableList
	input textinput.Model

	// termBgDark carries the direction of the terminal's real background
	// color when known (nil otherwise). It drives the low-contrast
	// markers shown while transparency is enabled.
	termBgDark *bool

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// ThemeItem represents a theme list item.
type ThemeItem struct {
	*list.Versioned
	key       string
	name      string
	swatches  []color.Color
	isCurrent bool
	// contrastWarn marks themes whose palette direction conflicts with
	// the terminal's real background while transparency is enabled.
	contrastWarn bool
	t            *styles.Styles
	m            fuzzy.Match
	cache        map[int]string
	focused      bool
}

// Finished implements list.Item. Theme items are render-stable outside
// of explicit SetFocused / SetMatch.
func (l *ThemeItem) Finished() bool {
	return true
}

var (
	_ Dialog   = (*Theme)(nil)
	_ ListItem = (*ThemeItem)(nil)
)

// NewTheme creates a new theme picker dialog. termBgDark carries the
// direction of the terminal's real background color when known (nil
// otherwise); it drives the low-contrast markers shown while
// transparency is enabled.
func NewTheme(com *common.Common, termBgDark *bool) *Theme {
	l := &Theme{com: com, termBgDark: termBgDark}

	h := help.New()
	h.Styles = com.Styles.DialogHelpStyles()
	l.help = h

	l.list = list.NewFilterableList()
	l.list.Focus()

	l.input = textinput.New()
	l.input.SetVirtualCursor(false)
	l.input.Placeholder = i18n.T("commands.filter_placeholder")
	l.input.SetStyles(com.Styles.TextInput)
	l.input.Focus()

	l.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", i18n.T("key.confirm")),
	)
	l.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", i18n.T("key.next_item")),
	)
	l.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", i18n.T("key.previous_item")),
	)
	l.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", i18n.T("key.choose")),
	)
	l.keyMap.Close = CloseKey

	l.setItems()
	return l
}

// ID implements Dialog.
func (l *Theme) ID() string {
	return ThemeID
}

// HandleMsg implements [Dialog].
func (l *Theme) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, l.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, l.keyMap.Previous):
			l.list.Focus()
			if l.list.IsSelectedFirst() {
				l.list.SelectLast()
				l.list.ScrollToBottom()
				break
			}
			l.list.SelectPrev()
			l.list.ScrollToSelected()
		case key.Matches(msg, l.keyMap.Next):
			l.list.Focus()
			if l.list.IsSelectedLast() {
				l.list.SelectFirst()
				l.list.ScrollToTop()
				break
			}
			l.list.SelectNext()
			l.list.ScrollToSelected()
		case key.Matches(msg, l.keyMap.Select):
			selectedItem := l.list.SelectedItem()
			if selectedItem == nil {
				break
			}
			themeItem, ok := selectedItem.(*ThemeItem)
			if !ok {
				break
			}
			return ActionSelectTheme{Key: themeItem.key}
		default:
			prevValue := l.input.Value()
			var cmd tea.Cmd
			l.input, cmd = l.input.Update(msg)
			value := l.input.Value()
			if value != prevValue {
				l.list.SetFilter(value)
				l.list.ScrollToTop()
				l.list.SetSelected(0)
			}

			return ActionCmd{cmd}
		}
	}
	return nil
}

// Cursor returns the cursor position relative to the dialog.
func (l *Theme) Cursor() *tea.Cursor {
	return InputCursor(l.com.Styles, l.input.Cursor())
}

// Draw implements [Dialog].
func (l *Theme) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := l.com.Styles
	width := max(0, min(themeDialogMaxWidth, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(themeDialogMaxHeight, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()

	l.input.SetWidth(dialogInputTextWidth(t, l.input, innerWidth))
	l.list.SetSize(innerWidth, max(0, height-heightOffset))

	rc := NewRenderContext(t, width)
	rc.Title = i18n.T("commands.theme")
	inputView := t.Dialog.InputPrompt.Render(l.input.View())
	rc.AddPart(inputView)

	visibleCount := len(l.list.FilteredItems())
	if l.list.Height() >= visibleCount {
		l.list.ScrollToTop()
	} else {
		l.list.ScrollToSelected()
	}

	listView := t.Dialog.List.Height(l.list.Height()).Render(l.list.Render())
	rc.AddPart(listView)
	rc.Help = renderDialogHelp(t, &l.help, l, innerWidth)

	view := rc.Render()

	cur := l.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// ShortHelp implements [help.KeyMap].
func (l *Theme) ShortHelp() []key.Binding {
	return []key.Binding{
		l.keyMap.UpDown,
		l.keyMap.Select,
		l.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (l *Theme) FullHelp() [][]key.Binding {
	m := [][]key.Binding{}
	slice := []key.Binding{
		l.keyMap.Select,
		l.keyMap.Next,
		l.keyMap.Previous,
		l.keyMap.Close,
	}
	for i := 0; i < len(slice); i += 4 {
		end := min(i+4, len(slice))
		m = append(m, slice[i:end])
	}
	return m
}

// setItems populates the theme list: an auto entry that follows the
// large model's provider, then every registered theme. The item whose
// key matches the configured theme is marked and preselected. While
// transparency is enabled, themes conflicting with the terminal's real
// background direction are flagged so users can see why they would be
// hard to read.
func (l *Theme) setItems() {
	current := ""
	transparent := false
	if cfg := l.com.Config(); cfg != nil {
		current = cfg.TUITheme()
		transparent = cfg.Options != nil && cfg.Options.TUI.IsTransparent()
	}

	items := make([]list.FilterableItem, 0, len(styles.AvailableThemes())+1)
	selectedIndex := 0

	auto := &ThemeItem{
		Versioned: list.NewVersioned(),
		name:      i18n.T("theme.auto"),
		isCurrent: current == "",
		t:         l.com.Styles,
	}
	items = append(items, auto)

	for i, entry := range styles.AvailableThemes() {
		item := &ThemeItem{
			Versioned: list.NewVersioned(),
			key:       entry.Key,
			name:      entry.Name,
			swatches:  entry.Swatches,
			isCurrent: entry.Key == current,
			t:         l.com.Styles,
		}
		if transparent && l.termBgDark != nil && entry.IsLight == *l.termBgDark {
			// Light theme over a dark terminal, or dark theme over a
			// light terminal: the palette contrast inverts.
			item.contrastWarn = true
		}
		items = append(items, item)
		if entry.Key == current {
			selectedIndex = i + 1
		}
	}

	l.list.SetItems(items...)
	l.list.SetSelected(selectedIndex)
	l.list.ScrollToSelected()
}

// Filter returns the filter value for the theme item.
func (l *ThemeItem) Filter() string {
	return l.name
}

// ID returns the unique identifier for the theme.
func (l *ThemeItem) ID() string {
	return l.key
}

// SetFocused sets the focus state of the theme item.
func (l *ThemeItem) SetFocused(focused bool) {
	if l.focused == focused {
		return
	}
	l.cache = nil
	l.focused = focused
	if l.Versioned != nil {
		l.Bump()
	}
}

// SetMatch sets the fuzzy match for the theme item.
func (l *ThemeItem) SetMatch(m fuzzy.Match) {
	if sameFuzzyMatch(l.m, m) {
		return
	}
	l.cache = nil
	l.m = m
	if l.Versioned != nil {
		l.Bump()
	}
}

// Render returns the string representation of the theme item: the theme
// name followed by a row of color swatches drawn in the theme's own
// palette, plus low-contrast and current markers when they apply.
func (l *ThemeItem) Render(width int) string {
	info := l.renderSwatches()
	if l.contrastWarn {
		if info != "" {
			info += " "
		}
		// Rendered last so the segment's own color reset cannot drop the
		// info-column color for anything that follows.
		info += l.t.LSP.WarningDiagnostic.Render("⚠ " + i18n.T("theme.contrast_low"))
	}
	if l.isCurrent {
		if info != "" {
			info += " "
		}
		info += i18n.T("theme.current")
	}
	st := ListItemStyles{
		ItemBlurred:     l.t.Dialog.NormalItem,
		ItemFocused:     l.t.Dialog.SelectedItem,
		InfoTextBlurred: l.t.Dialog.ListItem.InfoBlurred,
		InfoTextFocused: l.t.Dialog.ListItem.InfoFocused,
	}
	return renderItem(st, l.name, info, l.focused, width, l.cache, &l.m)
}

// renderSwatches renders the preview color chips for the theme. Each
// chip uses the theme's own color so the list doubles as a palette
// preview, independent of the active theme.
func (l *ThemeItem) renderSwatches() string {
	if len(l.swatches) == 0 {
		return ""
	}
	var b strings.Builder
	for i, c := range l.swatches {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(lipgloss.NewStyle().Foreground(c).Render("██"))
	}
	return b.String()
}
