package dialog

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"
)

const (
	// LanguageID is the identifier for the language picker dialog.
	LanguageID              = "language"
	languageDialogMaxWidth  = 50
	languageDialogMaxHeight = 12
)

// Language represents a dialog for selecting the UI display language.
type Language struct {
	com   *common.Common
	help  help.Model
	list  *list.FilterableList
	input textinput.Model

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// LanguageItem represents a language list item.
type LanguageItem struct {
	*list.Versioned
	locale    i18n.Locale
	isCurrent bool
	t         *styles.Styles
	m         fuzzy.Match
	cache     map[int]string
	focused   bool
}

// Finished implements list.Item. Language items are render-stable
// outside of explicit SetFocused / SetMatch.
func (l *LanguageItem) Finished() bool {
	return true
}

var (
	_ Dialog   = (*Language)(nil)
	_ ListItem = (*LanguageItem)(nil)
)

// NewLanguage creates a new language picker dialog.
func NewLanguage(com *common.Common) *Language {
	l := &Language{com: com}

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
func (l *Language) ID() string {
	return LanguageID
}

// HandleMsg implements [Dialog].
func (l *Language) HandleMsg(msg tea.Msg) Action {
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
			langItem, ok := selectedItem.(*LanguageItem)
			if !ok {
				break
			}
			return ActionSelectLanguage{Code: langItem.locale.Code}
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
func (l *Language) Cursor() *tea.Cursor {
	return InputCursor(l.com.Styles, l.input.Cursor())
}

// Draw implements [Dialog].
func (l *Language) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := l.com.Styles
	width := max(0, min(languageDialogMaxWidth, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(languageDialogMaxHeight, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()

	l.input.SetWidth(dialogInputTextWidth(t, l.input, innerWidth))
	l.list.SetSize(innerWidth, max(0, height-heightOffset))

	rc := NewRenderContext(t, width)
	rc.Title = i18n.T("commands.language")
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
func (l *Language) ShortHelp() []key.Binding {
	return []key.Binding{
		l.keyMap.UpDown,
		l.keyMap.Select,
		l.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (l *Language) FullHelp() [][]key.Binding {
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

func (l *Language) setItems() {
	cfg := l.com.Config()
	current := i18n.CurrentLocale()
	if cfg != nil && cfg.Options != nil && cfg.Options.TUI.Locale != "" {
		current = cfg.Options.TUI.Locale
	}

	items := make([]list.FilterableItem, 0, len(i18n.Available))
	selectedIndex := 0
	for i, locale := range i18n.Available {
		item := &LanguageItem{
			Versioned: list.NewVersioned(),
			locale:    locale,
			isCurrent: locale.Code == current,
			t:         l.com.Styles,
		}
		if locale.Code == current {
			selectedIndex = i
		}
		items = append(items, item)
	}

	l.list.SetItems(items...)
	l.list.SetSelected(selectedIndex)
	l.list.ScrollToSelected()
}

// Filter returns the filter value for the language item.
func (l *LanguageItem) Filter() string {
	return l.locale.Name
}

// ID returns the unique identifier for the language.
func (l *LanguageItem) ID() string {
	return l.locale.Code
}

// SetFocused sets the focus state of the language item.
func (l *LanguageItem) SetFocused(focused bool) {
	if l.focused == focused {
		return
	}
	l.cache = nil
	l.focused = focused
	if l.Versioned != nil {
		l.Bump()
	}
}

// SetMatch sets the fuzzy match for the language item.
func (l *LanguageItem) SetMatch(m fuzzy.Match) {
	if sameFuzzyMatch(l.m, m) {
		return
	}
	l.cache = nil
	l.m = m
	if l.Versioned != nil {
		l.Bump()
	}
}

// Render returns the string representation of the language item.
func (l *LanguageItem) Render(width int) string {
	info := ""
	if l.isCurrent {
		info = i18n.T("lang.current")
	}
	st := ListItemStyles{
		ItemBlurred:     l.t.Dialog.NormalItem,
		ItemFocused:     l.t.Dialog.SelectedItem,
		InfoTextBlurred: l.t.Dialog.ListItem.InfoBlurred,
		InfoTextFocused: l.t.Dialog.ListItem.InfoFocused,
	}
	return renderItem(st, l.locale.Name, info, l.focused, width, l.cache, &l.m)
}
