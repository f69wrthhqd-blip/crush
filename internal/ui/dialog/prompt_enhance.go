package dialog

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/ui/common"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// PromptEnhanceID is the identifier for the prompt-enhance dialog.
const PromptEnhanceID = "prompt_enhance"

// Limits for the text previews shown in the dialog. Full text lives in
// the editor; the dialog is a decision aid, not a reader.
const (
	promptEnhanceMaxLinesOriginal  = 5
	promptEnhanceMaxLinesOptimized = 12
	promptEnhanceMaxWidth          = 90
	promptEnhanceMinWidth          = 40
)

// PromptEnhance is a confirmation dialog showing the original draft and
// the optimized prompt stacked. Applying replaces the editor content
// with the optimized prompt.
type PromptEnhance struct {
	com        *common.Common
	original   string
	optimized  string
	selectedNo bool
	help       help.Model

	keyMap struct {
		LeftRight,
		EnterSpace,
		Yes,
		No,
		Tab,
		Close key.Binding
	}
}

var _ Dialog = (*PromptEnhance)(nil)

// NewPromptEnhance creates a new prompt-enhance confirmation dialog.
func NewPromptEnhance(com *common.Common, original, optimized string) *PromptEnhance {
	d := &PromptEnhance{
		com:       com,
		original:  original,
		optimized: optimized,
		help:      help.New(),
	}
	d.help.Styles = com.Styles.Help
	d.keyMap.LeftRight = key.NewBinding(
		key.WithKeys("left", "right"),
		key.WithHelp("←/→", i18n.T("key.switch_options")),
	)
	d.keyMap.EnterSpace = key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", i18n.T("key.confirm")),
	)
	d.keyMap.Yes = key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y/Y", i18n.T("key.yes")),
	)
	d.keyMap.No = key.NewBinding(
		key.WithKeys("n", "N"),
		key.WithHelp("n/N", i18n.T("key.no")),
	)
	d.keyMap.Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", i18n.T("key.switch_options")),
	)
	d.keyMap.Close = CloseKey
	return d
}

// ID implements [Dialog].
func (*PromptEnhance) ID() string {
	return PromptEnhanceID
}

// HandleMsg implements [Dialog].
func (d *PromptEnhance) HandleMsg(msg tea.Msg) Action {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, d.keyMap.Close):
			return ActionClose{}
		case key.Matches(keyMsg, d.keyMap.LeftRight, d.keyMap.Tab):
			d.selectedNo = !d.selectedNo
		case key.Matches(keyMsg, d.keyMap.EnterSpace):
			if !d.selectedNo {
				return ActionPromptEnhanceApply{Optimized: d.optimized}
			}
			return ActionClose{}
		case key.Matches(keyMsg, d.keyMap.Yes):
			return ActionPromptEnhanceApply{Optimized: d.optimized}
		case key.Matches(keyMsg, d.keyMap.No):
			return ActionClose{}
		}
	}
	return nil
}

// promptEnhancePreview wraps text to width (wide-character aware) and
// truncates the result to maxLines with an ellipsis marker.
func promptEnhancePreview(text string, width, maxLines int) string {
	text = strings.ReplaceAll(text, "\t", "  ")
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		lines = append(lines, strings.Split(ansi.WrapWc(paragraph, width, ""), "\n")...)
	}
	if len(lines) > maxLines {
		lines = append(lines[:maxLines-1], "…")
	}
	return strings.Join(lines, "\n")
}

// Draw implements [Dialog].
func (d *PromptEnhance) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	s := d.com.Styles

	width := min(promptEnhanceMaxWidth, max(promptEnhanceMinWidth, area.Dx()-s.Dialog.View.GetHorizontalFrameSize()-2))
	// Leave room for the row padding so text never touches the panel
	// edge.
	textWidth := max(20, width-4)

	original := promptEnhancePreview(d.original, textWidth, promptEnhanceMaxLinesOriginal)
	optimized := promptEnhancePreview(d.optimized, textWidth, promptEnhanceMaxLinesOptimized)

	// Render each panel row at the full panel width with the panel
	// background baked in. Composing shorter styled rows through
	// lipgloss.JoinVertical instead leaves the filler cells unstyled,
	// which shows the terminal background as black stripes.
	bg := s.Dialog.ContentPanel.GetBackground()
	panelRow := func(base lipgloss.Style) lipgloss.Style {
		return base.Background(bg).Width(width).Padding(0, 2)
	}
	labelRow := panelRow(lipgloss.NewStyle().Foreground(s.Dialog.Quit.Hint.GetForeground()))
	textRow := panelRow(lipgloss.NewStyle().Foreground(s.Dialog.Quit.Content.GetForeground()))
	blankRow := panelRow(lipgloss.NewStyle())

	buttons := common.ButtonGroup(s, []common.ButtonOpts{
		{Text: i18n.T("dialog.prompt_enhance.apply"), Selected: !d.selectedNo, Padding: 3},
		{Text: i18n.T("dialog.prompt_enhance.cancel"), Selected: d.selectedNo, Padding: 3},
	}, " ")

	blank := blankRow.Render("")
	panel := lipgloss.JoinVertical(
		lipgloss.Left,
		blank,
		labelRow.Render(i18n.T("dialog.prompt_enhance.original")),
		textRow.Render(original),
		blank,
		labelRow.Render(i18n.T("dialog.prompt_enhance.optimized")),
		textRow.Render(optimized),
		blank,
		blankRow.Render(buttons),
		blank,
	)

	header := common.DialogTitle(s, i18n.T("dialog.prompt_enhance.title"), width, s.Dialog.TitleGradFromColor, s.Dialog.TitleGradToColor)
	helpView := renderDialogHelp(s, &d.help, d, width)

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		s.Dialog.Title.Render(header),
		"",
		panel,
		"",
		helpView,
	)

	DrawCenter(scr, area, s.Dialog.View.Render(view))
	return nil
}

// ShortHelp implements [help.KeyMap].
func (d *PromptEnhance) ShortHelp() []key.Binding {
	return []key.Binding{
		d.keyMap.LeftRight,
		d.keyMap.EnterSpace,
		d.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (d *PromptEnhance) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{d.keyMap.LeftRight, d.keyMap.EnterSpace, d.keyMap.Yes, d.keyMap.No},
		{d.keyMap.Tab, d.keyMap.Close},
	}
}
