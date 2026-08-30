package agent

import (
	"context"
	_ "embed"

	"github.com/charmbracelet/crush/internal/agent/prompt"
	"github.com/charmbracelet/crush/internal/config"
)

//go:embed templates/coder.md.tpl
var coderPromptTmpl []byte

//go:embed templates/task.md.tpl
var taskPromptTmpl []byte

//go:embed templates/initialize.md.tpl
var initializePromptTmpl []byte

//go:embed templates/optimize_prompt.md.tpl
var optimizePromptTmpl []byte

func coderPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("coder", string(coderPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

func taskPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("task", string(taskPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

func InitializePrompt(cfg *config.ConfigStore) (string, error) {
	systemPrompt, err := prompt.NewPrompt("initialize", string(initializePromptTmpl))
	if err != nil {
		return "", err
	}
	return systemPrompt.Build(context.Background(), "", "", cfg)
}

// optimizePromptSystem renders the system prompt for the prompt
// optimization side-call, including project context such as the
// working directory, git status, and context files.
func optimizePromptSystem(ctx context.Context, provider, model string, cfg *config.ConfigStore) (string, error) {
	p, err := prompt.NewPrompt("optimize_prompt", string(optimizePromptTmpl))
	if err != nil {
		return "", err
	}
	return p.Build(ctx, provider, model, cfg)
}

// PlanModeSystemReminder returns the plan-mode guidance appended to the
// system prompt while plan mode is active. It instructs the model to
// investigate read-only, ask clarifying questions, and hand over the
// finished plan: via present_plan in interactive sessions, or as the
// final message in non-interactive runs where the present_plan tool is
// not registered.
func PlanModeSystemReminder(interactive bool) string {
	reminder := `<system-reminder>
Plan mode is active. The user does not want you to execute yet: you MUST NOT edit files, run state-modifying commands, or otherwise change the system. Only read-only tools are available (view, glob, grep, ls, fetch, agentic_fetch, sourcegraph, LSP read-only tools, crush_info, crush_logs, read_mcp_resource, list_mcp_resources, todos, question). A blocked tool means it is not read-only: do NOT retry it, do not work around it, and do not call present_plan just to unblock it.

Work iteratively with the user:
1. Explore the codebase with read-only tools to understand the task and find existing patterns to reuse.
2. When you hit an ambiguity you cannot resolve from code alone, ask the user with the question tool (batch related questions, never ask what you can find by reading code).
3. Keep refining until the plan is complete and unambiguous: what to change, which files to touch, what to reuse (with paths), and how to verify.

`
	if interactive {
		reminder += `Only when the plan is decision-complete, call present_plan with the full markdown plan. The user may approve it (start implementing after approval), ask to keep refining, or dismiss it. Never start implementing until the user explicitly approves the plan.
</system-reminder>`
	} else {
		reminder += `When the plan is decision-complete, output the full markdown plan as your final message. This run is non-interactive: present_plan is unavailable, so do not call it.
</system-reminder>`
	}
	return reminder
}
