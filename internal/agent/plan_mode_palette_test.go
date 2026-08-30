package agent

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/permission"
)

// TestBuildToolsCoderPalette is a regression test for the palette as the
// model actually sees it. It caught present_plan being dropped by the
// AllowedTools filter because it was missing from config.allToolNames,
// which silently disabled the whole plan-approval flow while the
// isPlanModeTool unit test still passed.
func TestBuildToolsCoderPalette(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfgStore, err := config.Init(tmp, "", false)
	require.NoError(t, err)
	cfgStore.SetupAgents()

	c := &coordinator{
		cfg:         cfgStore,
		permissions: permission.NewPermissionService(tmp, true, nil),
		interactive: true,
	}

	agentCfg := cfgStore.Config().Agents[config.AgentCoder]
	// The agent and agentic_fetch tools need sessions/messages wiring the
	// palette assertions don't exercise; drop them from the test config.
	agentCfg.AllowedTools = slices.DeleteFunc(slices.Clone(agentCfg.AllowedTools), func(name string) bool {
		return name == AgentToolName || name == tools.AgenticFetchToolName
	})

	buildNames := paletteNames(t, c, agentCfg, false)
	require.Contains(t, buildNames, tools.PresentPlanToolName)
	require.Contains(t, buildNames, tools.QuestionToolName)
	require.Contains(t, buildNames, tools.BashToolName)
	require.Contains(t, buildNames, tools.EditToolName)
	require.Contains(t, buildNames, tools.WriteToolName)
	require.Contains(t, buildNames, tools.ViewToolName)

	c.planMode.Store(true)
	planNames := paletteNames(t, c, agentCfg, false)
	require.Contains(t, planNames, tools.PresentPlanToolName)
	require.Contains(t, planNames, tools.ViewToolName)
	require.Contains(t, planNames, tools.QuestionToolName)
	require.Contains(t, planNames, tools.TodosToolName)
	for _, name := range []string{
		tools.BashToolName,
		tools.EditToolName,
		tools.MultiEditToolName,
		tools.WriteToolName,
		tools.DownloadToolName,
		tools.RenameToolName,
		tools.ReplaceSymbolToolName,
		tools.LSPRestartToolName,
		tools.JobKillToolName,
		tools.JobOutputToolName,
		tools.AgenticFetchToolName,
	} {
		require.NotContains(t, planNames, name)
	}
}

// TestBuildToolsTaskPalette checks the task sub-agent keeps its read-only
// palette and never receives the plan lifecycle or question tools.
func TestBuildToolsTaskPalette(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfgStore, err := config.Init(tmp, "", false)
	require.NoError(t, err)
	cfgStore.SetupAgents()

	c := &coordinator{
		cfg:         cfgStore,
		permissions: permission.NewPermissionService(tmp, true, nil),
	}

	agentCfg := cfgStore.Config().Agents[config.AgentTask]
	names := paletteNames(t, c, agentCfg, true)

	require.Contains(t, names, tools.ViewToolName)
	require.Contains(t, names, tools.GrepToolName)
	for _, name := range []string{
		tools.PresentPlanToolName,
		tools.QuestionToolName,
		tools.BashToolName,
		tools.EditToolName,
		tools.WriteToolName,
		tools.AgenticFetchToolName,
	} {
		require.NotContains(t, names, name)
	}
}

func paletteNames(t *testing.T, c *coordinator, agentCfg config.Agent, isSubAgent bool) []string {
	t.Helper()

	built, err := c.buildTools(context.Background(), agentCfg, isSubAgent)
	require.NoError(t, err)

	names := make([]string, 0, len(built))
	for _, tool := range built {
		names = append(names, tool.Info().Name)
	}
	slices.Sort(names)
	return names
}
