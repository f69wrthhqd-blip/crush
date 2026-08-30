package tools

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/i18n"
	"github.com/charmbracelet/crush/internal/question"
)

// PresentPlanToolName is the tool the model calls in plan mode to hand its
// finished plan to the user for approval.
const PresentPlanToolName = "present_plan"

//go:embed present_plan.md
var presentPlanDescription string

// PresentPlanParams defines the parameters for the present_plan tool.
type PresentPlanParams struct {
	// Plan is the full markdown plan the model wants the user to approve.
	Plan string `json:"plan" description:"The complete plan (markdown) you want the user to approve before you start implementing."`
}

// PlanApprovalOption identifies the user's decision on a presented plan.
type PlanApprovalOption string

const (
	// PlanApprovalExecute approves the plan and exits plan mode so the
	// model starts implementing in the current conversation.
	PlanApprovalExecute PlanApprovalOption = "execute"
	// PlanApprovalExecuteFresh approves the plan, summarizes the
	// conversation to free up context, and then starts implementing in
	// the fresh context.
	PlanApprovalExecuteFresh PlanApprovalOption = "execute_fresh"
	// PlanApprovalContinue keeps plan mode active so the plan can be
	// refined further.
	PlanApprovalContinue PlanApprovalOption = "continue"
	// PlanApprovalCancel dismisses the plan and exits plan mode without
	// executing.
	PlanApprovalCancel PlanApprovalOption = "cancel"
)

// PlanApprovalQuestionID is the question ID used for the plan approval
// single-choice question published through the question service. The UI
// special-cases this question and renders a plan-approval dialog.
const PlanApprovalQuestionID = "_plan_approval"

// PlanApprovalBatchID is the batch ID used for plan approval requests.
const PlanApprovalBatchID = "_plan_approval_batch"

// NewPresentPlanTool creates the present_plan tool. It publishes a
// plan-approval question through the question service and blocks until the
// user picks an option. The returned tool response tells the model whether
// to start implementing or keep refining. onExecuteFresh, when non-nil, is
// invoked when the user picks "execute in fresh context" so the caller can
// orchestrate the summarize-then-implement handoff.
func NewPresentPlanTool(svc question.Service, onExecuteFresh func(ctx context.Context, sessionID string)) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		PresentPlanToolName,
		presentPlanDescription,
		func(ctx context.Context, params PresentPlanParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if strings.TrimSpace(params.Plan) == "" {
				return fantasy.NewTextErrorResponse("plan is required and must not be empty"), nil
			}

			sessionID := GetSessionFromContext(ctx)

			// Dialog strings are user-facing and localized; the response
			// strings below stay in English because the model consumes
			// them. The question text stays short: question.Service
			// validates it against MaxQuestionLength, and the full plan
			// is already visible in the conversation transcript where it
			// renders as markdown.
			req := question.Request{
				ID:                 PlanApprovalBatchID,
				SessionID:          sessionID,
				ToolCallID:         call.ID,
				ConfirmTitle:       i18n.T("plan_approval.title"),
				ConfirmDescription: i18n.T("plan_approval.description"),
				Questions: []question.Question{
					{
						ID:          PlanApprovalQuestionID,
						Type:        question.TypeSingleChoice,
						Label:       i18n.T("plan_approval.label"),
						Text:        i18n.T("plan_approval.question_text"),
						Description: i18n.T("plan_approval.question_description"),
						Choices: []question.Choice{
							{
								ID:          string(PlanApprovalExecute),
								Label:       i18n.T("plan_approval.execute"),
								Description: i18n.T("plan_approval.execute_description"),
							},
							{
								ID:          string(PlanApprovalExecuteFresh),
								Label:       i18n.T("plan_approval.execute_fresh"),
								Description: i18n.T("plan_approval.execute_fresh_description"),
							},
							{
								ID:          string(PlanApprovalContinue),
								Label:       i18n.T("plan_approval.continue"),
								Description: i18n.T("plan_approval.continue_description"),
							},
							{
								ID:          string(PlanApprovalCancel),
								Label:       i18n.T("plan_approval.cancel"),
								Description: i18n.T("plan_approval.cancel_description"),
							},
						},
					},
				},
			}

			answers, err := svc.Ask(ctx, req)
			if err != nil {
				if errors.Is(err, question.ErrCancelled) {
					resp := fantasy.NewTextErrorResponse(i18n.T("plan_approval.cancelled"))
					resp.StopTurn = true
					return resp, nil
				}
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			if planApprovalOption(answers) == PlanApprovalExecuteFresh && onExecuteFresh != nil {
				// Detach from the tool call's context: this answer ends the
				// run (and cancels its context), while the orchestration
				// must outlive it. Session-level Cancel still stops the
				// follow-up summarize and turn.
				detached := context.WithoutCancel(ctx)
				go onExecuteFresh(detached, sessionID)
			}

			return formatPlanApproval(answers)
		},
	)
}

// planApprovalOption extracts the chosen PlanApprovalOption from the
// question answers, falling back to cancel for empty or unrecognized
// answers.
func planApprovalOption(answers []question.Answer) PlanApprovalOption {
	if len(answers) == 0 {
		return PlanApprovalCancel
	}
	if len(answers[0].SelectedIDs) > 0 {
		return PlanApprovalOption(answers[0].SelectedIDs[0])
	}
	if answers[0].FillInText != "" {
		return PlanApprovalOption(strings.TrimSpace(strings.ToLower(answers[0].FillInText)))
	}
	return PlanApprovalCancel
}

// formatPlanApproval converts the plan-approval answer into the tool
// response fed back to the model.
func formatPlanApproval(answers []question.Answer) (fantasy.ToolResponse, error) {
	switch planApprovalOption(answers) {
	case PlanApprovalExecute:
		// No StopTurn: the turn continues in the same conversation so
		// the model starts implementing right after the approval.
		return fantasy.NewTextResponse("User approved the plan. Start implementing it now. Begin by updating your todo list if applicable."), nil
	case PlanApprovalExecuteFresh:
		// StopTurn ends this turn; the coordinator summarizes the
		// conversation and re-invokes the model with the approved plan
		// in a fresh context.
		resp := fantasy.NewTextResponse("User approved the plan and chose to execute it in a fresh context. The conversation will be summarized and implementation will start automatically.")
		resp.StopTurn = true
		return resp, nil
	case PlanApprovalContinue:
		return fantasy.NewTextResponse("User wants to keep refining the plan. Stay in plan mode, address their feedback, and call present_plan again once the plan is complete."), nil
	default:
		resp := fantasy.NewTextResponse("User dismissed the plan and exited plan mode. Do not implement the plan; wait for the user's next request.")
		resp.StopTurn = true
		return resp, nil
	}
}

// IsPlanApprovalBatch reports whether a question batch is a plan-approval
// request (used by the UI to render the plan dialog).
func IsPlanApprovalBatch(batch question.Request) bool {
	return batch.ID == PlanApprovalBatchID
}
