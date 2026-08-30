package chat

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

// -----------------------------------------------------------------------------
// Present Plan Tool
// -----------------------------------------------------------------------------

// PresentPlanToolMessageItem is a message item that represents a
// present_plan tool call. It renders the plan as markdown instead of the
// generic raw-JSON parameter dump.
type PresentPlanToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*PresentPlanToolMessageItem)(nil)

// NewPresentPlanToolMessageItem creates a new [PresentPlanToolMessageItem].
func NewPresentPlanToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &PresentPlanToolRenderContext{}, canceled)
}

// PresentPlanToolRenderContext renders present_plan tool messages.
type PresentPlanToolRenderContext struct{}

// RenderTool implements the [ToolRenderer] interface.
func (p *PresentPlanToolRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	title := i18n.T("chat.present_plan")
	if opts.IsPending() {
		return pendingTool(sty, title, opts.Anim, opts.Compact)
	}

	var params tools.PresentPlanParams
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil || strings.TrimSpace(params.Plan) == "" {
		return toolErrorContent(sty, &message.ToolResult{Content: i18n.T("chat.invalid_parameters")}, cappedWidth)
	}

	header := toolHeader(sty, opts.Status, title, cappedWidth, opts)
	if opts.Compact {
		return header
	}

	if earlyState, ok := toolEarlyStateContent(sty, opts, cappedWidth); ok {
		return joinToolParts(header, earlyState)
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	renderer := common.MarkdownRenderer(sty, bodyWidth)
	mu := common.LockMarkdownRenderer(renderer)
	mu.Lock()
	rendered, err := renderer.Render(params.Plan)
	mu.Unlock()
	if err != nil {
		return joinToolParts(header, sty.Tool.Body.Render(params.Plan))
	}

	return joinToolParts(header, sty.Tool.Body.Render(strings.TrimSuffix(rendered, "\n")))
}
