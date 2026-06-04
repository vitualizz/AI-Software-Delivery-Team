package panels_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// TestPipelinePanelNilStateRendersPlaceholder verifies that a panel with no
// loaded state renders the "No pipeline state found" placeholder.
func TestPipelinePanelNilStateRendersPlaceholder(t *testing.T) {
	p := panels.NewPipelinePanel()
	view := p.View()

	if !strings.Contains(view, "No pipeline state found") {
		t.Errorf("expected 'No pipeline state found' in view, got: %q", view)
	}
}

// TestPipelinePanelPlanStateActive verifies that when the pipeline state is
// "plan", the active indicator (▶) appears next to "plan".
func TestPipelinePanelPlanStateActive(t *testing.T) {
	p := panels.NewPipelinePanel()

	state := pipeline.State{
		SchemaVersion: "1",
		ChangeID:      "test-change",
		CurrentState:  pipeline.PhasePlan,
	}
	p, _ = p.SetState(state)
	view := p.View()

	if !strings.Contains(view, "▶") {
		t.Errorf("expected active indicator (▶) in view, got: %q", view)
	}
	if !strings.Contains(view, "plan") {
		t.Errorf("expected 'plan' in view, got: %q", view)
	}
}

// TestPipelinePanelRequirementsStateActive verifies that when the pipeline state
// is "requirements", the active indicator appears next to "requirements".
func TestPipelinePanelRequirementsStateActive(t *testing.T) {
	p := panels.NewPipelinePanel()

	state := pipeline.State{
		SchemaVersion: "1",
		ChangeID:      "test-change",
		CurrentState:  pipeline.PhaseRequirements,
	}
	p, _ = p.SetState(state)
	view := p.View()

	if !strings.Contains(view, "▶") {
		t.Errorf("expected active indicator (▶) in view, got: %q", view)
	}
	if !strings.Contains(view, "requirements") {
		t.Errorf("expected 'requirements' in view, got: %q", view)
	}
}

// TestPipelinePanelCompletedPhaseHasCheckmark verifies that phases before the
// current state are rendered with the completed checkmark (✓).
func TestPipelinePanelCompletedPhaseHasCheckmark(t *testing.T) {
	p := panels.NewPipelinePanel()

	state := pipeline.State{
		SchemaVersion: "1",
		ChangeID:      "test-change",
		CurrentState:  pipeline.PhasePlan,
	}
	p, _ = p.SetState(state)
	view := p.View()

	if !strings.Contains(view, "✓") {
		t.Errorf("expected completed checkmark (✓) for requirements phase, got: %q", view)
	}
}

// TestPipelinePanelCurrentStateReturnsString verifies CurrentState() returns
// the phase string when a state is loaded.
func TestPipelinePanelCurrentStateReturnsString(t *testing.T) {
	p := panels.NewPipelinePanel()

	if got := p.CurrentState(); got != "" {
		t.Errorf("expected empty state on zero-value panel, got %q", got)
	}

	state := pipeline.State{
		SchemaVersion: "1",
		ChangeID:      "test-change",
		CurrentState:  pipeline.PhaseImplement,
	}
	p, _ = p.SetState(state)

	if got := p.CurrentState(); got != "implement" {
		t.Errorf("expected 'implement', got %q", got)
	}
}
