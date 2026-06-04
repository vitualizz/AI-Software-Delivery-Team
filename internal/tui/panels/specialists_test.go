package panels_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// TestSpecialistsPanelNilStateRendersPlaceholder verifies that a panel with no
// loaded state renders the "No specialists have run yet" placeholder.
func TestSpecialistsPanelNilStateRendersPlaceholder(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	view := p.View()

	if !strings.Contains(view, "No specialists have run yet") {
		t.Errorf("expected 'No specialists have run yet' in view, got: %q", view)
	}
}

// TestSpecialistsPanelDeveloperCompletedShowsCheckmark verifies that a developer
// with two completed steps where CurrentStep == last step ID renders completed style (✓).
func TestSpecialistsPanelDeveloperCompletedShowsCheckmark(t *testing.T) {
	now := time.Now()
	state := &pipeline.PipelineStateV2{
		SchemaVersion: pipeline.SchemaVersionV2,
		ChangeID:      "test-change",
		Specialists: map[string]pipeline.SpecialistState{
			"developer": {
				CurrentStep: "step-2",
				StepsCompleted: []pipeline.StepRecord{
					{ID: "step-1", Timestamp: now.Add(-time.Minute)},
					{ID: "step-2", Timestamp: now},
				},
			},
		},
	}

	p := panels.NewSpecialistsPanel()
	p.SetState(state)
	view := p.View()

	if !strings.Contains(view, "✓") {
		t.Errorf("expected completed checkmark (✓) for developer, got: %q", view)
	}
	// Should not show the placeholder when state is set.
	if strings.Contains(view, "No specialists have run yet") {
		t.Errorf("unexpected placeholder in view when state is loaded")
	}
}

// TestSpecialistsPanelNavigation verifies that j/k key presses change the selected index.
func TestSpecialistsPanelNavigation(t *testing.T) {
	p := panels.NewSpecialistsPanel()

	// Initial selection should be 0.
	if got := p.SelectedSpecialist(); got != "ux-ui" {
		t.Errorf("expected initial selection 'ux-ui', got %q", got)
	}

	// Press j to move down.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if got := p.SelectedSpecialist(); got != "architect" {
		t.Errorf("expected 'architect' after j, got %q", got)
	}

	// Press j again.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if got := p.SelectedSpecialist(); got != "developer" {
		t.Errorf("expected 'developer' after second j, got %q", got)
	}

	// Press k to move back up.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if got := p.SelectedSpecialist(); got != "architect" {
		t.Errorf("expected 'architect' after k, got %q", got)
	}
}

// TestSpecialistsPanelSelectedSpecialistReturnsCorrectID verifies that
// SelectedSpecialist() returns the ID for the selected index.
func TestSpecialistsPanelSelectedSpecialistReturnsCorrectID(t *testing.T) {
	order := []string{"ux-ui", "architect", "developer", "qa", "security"}

	p := panels.NewSpecialistsPanel()

	for i, expected := range order {
		// Navigate to position i.
		for range i {
			p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		if got := p.SelectedSpecialist(); got != expected {
			t.Errorf("position %d: expected %q, got %q", i, expected, got)
		}
		// Reset back to top.
		for range i {
			p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		}
	}
}
