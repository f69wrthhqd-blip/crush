package agent

import (
	"testing"

	"github.com/charmbracelet/crush/internal/agent/tools"
)

func TestIsPlanModeTool(t *testing.T) {
	readOnly := []string{
		tools.ViewToolName,
		tools.GlobToolName,
		tools.GrepToolName,
		tools.LSToolName,
		tools.FetchToolName,
		tools.AgenticFetchToolName,
		tools.SourcegraphToolName,
		tools.TodosToolName,
		tools.QuestionToolName,
		tools.CrushInfoToolName,
		tools.CrushLogsToolName,
		tools.ReadMCPResourceToolName,
		tools.ListMCPResourcesToolName,
		tools.DiagnosticsToolName,
		tools.ReferencesToolName,
		tools.SymbolsToolName,
		tools.DefinitionToolName,
		tools.CallHierarchyToolName,
		tools.PresentPlanToolName,
	}
	for _, name := range readOnly {
		if !isPlanModeTool(name) {
			t.Errorf("expected %q to be allowed in plan mode", name)
		}
	}

	writeTools := []string{
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
	}
	for _, name := range writeTools {
		if isPlanModeTool(name) {
			t.Errorf("expected %q to be blocked in plan mode", name)
		}
	}

	// Not first-class tools at all: web_fetch and web_search only exist
	// inside the agentic_fetch sub-runner, so they must never be advertised
	// as part of the plan-mode palette.
	for _, name := range []string{tools.WebFetchToolName, tools.WebSearchToolName} {
		if isPlanModeTool(name) {
			t.Errorf("expected non-first-class tool %q to be blocked in plan mode", name)
		}
	}
}
