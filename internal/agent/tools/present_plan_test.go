package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/question"
)

// stubQuestionService is a minimal question.Service that records the
// published request and returns canned answers.
type stubQuestionService struct {
	req     *question.Request
	answers []question.Answer
}

func (s *stubQuestionService) Ask(ctx context.Context, req question.Request) ([]question.Answer, error) {
	s.req = &req
	return s.answers, nil
}

func (s *stubQuestionService) Answer([]question.Answer) bool { return false }
func (s *stubQuestionService) Cancel() bool                  { return false }
func (s *stubQuestionService) Subscribe(ctx context.Context) <-chan pubsub.Event[question.Request] {
	return make(<-chan pubsub.Event[question.Request])
}

func (s *stubQuestionService) SubscribeNotifications(ctx context.Context) <-chan pubsub.Event[question.Notification] {
	return make(<-chan pubsub.Event[question.Notification])
}

func TestPresentPlanTool_PublishesApprovalQuestion(t *testing.T) {
	svc := &stubQuestionService{
		answers: []question.Answer{{
			QuestionID:  PlanApprovalQuestionID,
			SelectedIDs: []string{string(PlanApprovalExecute)},
		}},
	}
	tool := NewPresentPlanTool(svc, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":"Step 1: read files\nStep 2: edit"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsError {
		t.Fatalf("expected success, got error: %s", resp.Content)
	}
	if svc.req == nil {
		t.Fatal("expected question request to be published")
	}
	if svc.req.ID != PlanApprovalBatchID {
		t.Fatalf("expected plan approval batch ID, got %q", svc.req.ID)
	}
	if len(svc.req.Questions) != 1 || svc.req.Questions[0].ID != PlanApprovalQuestionID {
		t.Fatalf("expected single plan approval question")
	}
	if len(svc.req.Questions[0].Choices) != 4 {
		t.Fatalf("expected 4 approval choices, got %d", len(svc.req.Questions[0].Choices))
	}
	if resp.StopTurn {
		t.Error("execute must not stop the turn: the model should start implementing in the same conversation")
	}
	if !IsPlanApprovalBatch(*svc.req) {
		t.Error("IsPlanApprovalBatch should recognize the plan approval batch")
	}
}

// TestPresentPlanTool_QuestionValidatesAgainstRealService guards the
// regression where the full plan was passed as the question text and
// rejected by question.Service's length validation before the request
// ever reached the UI.
func TestPresentPlanTool_QuestionValidatesAgainstRealService(t *testing.T) {
	svc := &stubQuestionService{answers: []question.Answer{{
		QuestionID:  PlanApprovalQuestionID,
		SelectedIDs: []string{string(PlanApprovalExecute)},
	}}}
	tool := NewPresentPlanTool(svc, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	longPlan := strings.Repeat("# Plan\n\nDetailed step with plenty of markdown content.\n\n", 100)
	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":` + quoteJSON(longPlan) + `}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsError {
		t.Fatalf("expected success for a long plan, got error: %s", resp.Content)
	}
	if svc.req == nil {
		t.Fatal("expected question request to be published")
	}
	if err := svc.req.Validate(); err != nil {
		t.Fatalf("published request must pass question validation: %v", err)
	}
	if strings.Contains(svc.req.Questions[0].Text, longPlan[:20]) {
		t.Error("question text must not embed the full plan")
	}
}

func TestPresentPlanTool_ExecuteDoesNotStopTurn(t *testing.T) {
	svc := &stubQuestionService{
		answers: []question.Answer{{
			QuestionID:  PlanApprovalQuestionID,
			SelectedIDs: []string{string(PlanApprovalExecute)},
		}},
	}
	tool := NewPresentPlanTool(svc, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":"plan"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StopTurn {
		t.Error("execute must not stop the turn: the model should start implementing in the same conversation")
	}
}

func TestPresentPlanTool_ExecuteFreshTriggersCallbackAndStopsTurn(t *testing.T) {
	svc := &stubQuestionService{
		answers: []question.Answer{{
			QuestionID:  PlanApprovalQuestionID,
			SelectedIDs: []string{string(PlanApprovalExecuteFresh)},
		}},
	}
	fresh := make(chan string, 1)
	tool := NewPresentPlanTool(svc, func(ctx context.Context, sessionID string) {
		fresh <- sessionID
	})
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":"plan"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.StopTurn {
		t.Error("execute fresh should stop the turn so it can be summarized")
	}
	select {
	case sessionID := <-fresh:
		if sessionID != "s1" {
			t.Fatalf("expected callback with session s1, got %q", sessionID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected the execute-fresh callback to run")
	}
}

func TestPresentPlanTool_CancelStopsTurn(t *testing.T) {
	svc := &stubQuestionService{
		answers: []question.Answer{{
			QuestionID:  PlanApprovalQuestionID,
			SelectedIDs: []string{string(PlanApprovalCancel)},
		}},
	}
	tool := NewPresentPlanTool(svc, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":"plan"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.StopTurn {
		t.Error("cancel should stop the turn: the plan was dismissed")
	}
}

func TestPresentPlanTool_EmptyPlanRejected(t *testing.T) {
	svc := &stubQuestionService{}
	tool := NewPresentPlanTool(svc, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":"   "}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsError {
		t.Fatal("expected empty plan to be rejected")
	}
	if svc.req != nil {
		t.Fatal("expected no question request for empty plan")
	}
}

func TestPresentPlanTool_ContinueKeepsPlanMode(t *testing.T) {
	svc := &stubQuestionService{
		answers: []question.Answer{{
			QuestionID:  PlanApprovalQuestionID,
			SelectedIDs: []string{string(PlanApprovalContinue)},
		}},
	}
	tool := NewPresentPlanTool(svc, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "s1")

	resp, err := tool.Run(ctx, fantasy.ToolCall{ID: "call", Name: "present_plan", Input: `{"plan":"plan"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsError {
		t.Fatalf("expected success, got error: %s", resp.Content)
	}
	if resp.StopTurn {
		t.Error("continue should not stop the turn")
	}
}

// quoteJSON encodes s as a JSON string literal.
func quoteJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
