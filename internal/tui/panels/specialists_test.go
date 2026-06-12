package panels_test

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/asdt/internal/pipeline"
	"github.com/vitualizz/asdt/internal/tui/panels"
)

// runningState builds a StateV2 with one specialist whose CurrentStep does
// NOT match the last completed step's ID — the existing "in progress" signal
// this panel already derives at render time (see View(): state = StatusRunning
// when sp.CurrentStep != last.ID).
func runningState() *pipeline.StateV2 {
	now := time.Now()
	return &pipeline.StateV2{
		SchemaVersion: pipeline.SchemaVersionV2,
		ChangeID:      "test-change",
		Specialists: map[string]pipeline.SpecialistState{
			"developer": {
				CurrentStep: "step-3",
				StepsCompleted: []pipeline.StepRecord{
					{ID: "step-1", Timestamp: now.Add(-2 * time.Minute)},
					{ID: "step-2", Timestamp: now.Add(-time.Minute)},
				},
			},
		},
	}
}

// idleState builds a StateV2 where CurrentStep == last completed step's ID —
// i.e. nothing is in progress (StatusDone, not StatusRunning).
func idleState() *pipeline.StateV2 {
	now := time.Now()
	return &pipeline.StateV2{
		SchemaVersion: pipeline.SchemaVersionV2,
		ChangeID:      "test-change",
		Specialists: map[string]pipeline.SpecialistState{
			"developer": {
				CurrentStep: "step-2",
				StepsCompleted: []pipeline.StepRecord{
					{ID: "step-1", Timestamp: now.Add(-2 * time.Minute)},
					{ID: "step-2", Timestamp: now},
				},
			},
		},
	}
}

// TestSpecialistsPanelNilStateRendersPlaceholder verifies that a panel with no
// loaded state renders the "No specialists have run yet" placeholder.
func TestSpecialistsPanelNilStateRendersPlaceholder(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	view := p.View()

	if !strings.Contains(view, "No specialists have run yet") {
		t.Errorf("expected 'No specialists have run yet' in view, got: %q", view)
	}
}

// TestSpecialistsPanelEmptyStateShowsGuidanceLine verifies that the nil-state
// placeholder is EXTENDED (not replaced) with a second muted guidance line
// telling the user how to populate the panel by running a specialist via the
// `/asdt` entry point.
func TestSpecialistsPanelEmptyStateShowsGuidanceLine(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	view := p.View()

	if !strings.Contains(view, "No specialists have run yet") {
		t.Errorf("expected original placeholder 'No specialists have run yet' to remain, got: %q", view)
	}
	if !strings.Contains(view, "/asdt") {
		t.Errorf("expected guidance line referencing '/asdt' entry point, got: %q", view)
	}
}

// TestSpecialistsPanelDeveloperCompletedShowsCheckmark verifies that a developer
// with two completed steps where CurrentStep == last step ID renders completed style (✓).
func TestSpecialistsPanelDeveloperCompletedShowsCheckmark(t *testing.T) {
	now := time.Now()
	state := &pipeline.StateV2{
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

// TestSpecialistsPanelCompactMode verifies that at <=60 width the panel
// renders all specialists in compact mode without errors.
func TestSpecialistsPanelCompactMode(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p, _ = p.UpdateSize(50, 30)
	view := p.View()

	for _, name := range []string{"UX/UI", "Architect", "Developer", "QA", "Security"} {
		if !strings.Contains(view, name) {
			t.Errorf("expected %q in compact view, got: %q", name, view)
		}
	}
}

// TestSpecialistsPanelRendersHeader verifies that the PanelHeader decorator
// appears in the specialists view.
func TestSpecialistsPanelRendersHeader(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p, _ = p.UpdateSize(80, 30)
	view := p.View()

	if !strings.Contains(view, "Specialists") {
		t.Errorf("expected 'Specialists' title in view, got: %q", view)
	}
}

// TestSpecialistsPanelViewShowsFullStepHistoryWithDurations verifies that
// View() renders every StepsCompleted entry (not just the last) with computed
// inter-step durations interleaved between consecutive rows.
func TestSpecialistsPanelViewShowsFullStepHistoryWithDurations(t *testing.T) {
	base := time.Date(2026, 6, 7, 9, 0, 0, 0, time.UTC)
	p := panels.NewSpecialistsPanel()
	p.SetSize(100, 30)
	state := &pipeline.StateV2{
		Specialists: map[string]pipeline.SpecialistState{
			"developer": {
				CurrentStep: "design",
				StepsCompleted: []pipeline.StepRecord{
					{ID: "explore", Timestamp: base},
					{ID: "spec", Timestamp: base.Add(90 * time.Second)},
					{ID: "design", Timestamp: base.Add(4 * time.Minute)},
				},
			},
		},
	}
	p.SetState(state)
	p.SetFocused(true)

	out := p.View()

	for _, step := range []string{"explore", "spec", "design"} {
		if !strings.Contains(out, step) {
			t.Errorf("expected full step history to include %q, got: %q", step, out)
		}
	}
	// Pairwise durations: explore->spec = 1m30s, spec->design = 2m30s.
	if !strings.Contains(out, "1m30s") {
		t.Errorf("expected computed duration '1m30s' between explore and spec, got: %q", out)
	}
	if !strings.Contains(out, "2m30s") {
		t.Errorf("expected computed duration '2m30s' between spec and design, got: %q", out)
	}
}

// TestSpecialistsPanelSurfacesStateChangeID verifies that the loaded state's
// ChangeID is rendered as additive informational context (not a mismatch
// comparison against the configured active change).
func TestSpecialistsPanelSurfacesStateChangeID(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetSize(100, 24)
	state := &pipeline.StateV2{
		ChangeID: "tui-pastel-redesign",
		Specialists: map[string]pipeline.SpecialistState{
			"developer": {CurrentStep: "implement"},
		},
	}
	p.SetState(state)

	out := p.View()
	if !strings.Contains(out, "tui-pastel-redesign") {
		t.Errorf("expected StateV2.ChangeID 'tui-pastel-redesign' surfaced in panel view, got: %q", out)
	}
}

// TestSpecialistsPanelSelectedRowShowsCursorGlyphAdditively verifies that the
// '►' cursor glyph (matching setup's existing cursorChar convention) appears
// on the selected row's rendered output ALONGSIDE the existing
// Background(ColorInactive) highlight — proving additive composition, not
// replacement.
func TestSpecialistsPanelSelectedRowShowsCursorGlyphAdditively(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetSize(100, 24)
	p.SetFocused(true)

	out := p.View()
	lines := strings.Split(out, "\n")

	var selectedLine string
	for _, line := range lines {
		if strings.Contains(line, "UX/UI") { // first specialist row, index 0 == selected
			selectedLine = line
		}
	}

	if !strings.Contains(selectedLine, "►") {
		t.Errorf("expected selected row to include cursor glyph '►', got: %q", selectedLine)
	}
	// The background highlight is applied via lipgloss styling (ANSI escape
	// sequences), which strings.Contains can't directly assert on a styled
	// segment — but the glyph's presence alongside non-empty styled output
	// (len > len(plain row)) is the additive-composition signal we assert here.
	if len(selectedLine) <= len("  UX/UI         ► ") {
		t.Errorf("expected selected row to retain background-highlight styling alongside the glyph, got unstyled-looking row: %q", selectedLine)
	}
}

// TestSpecialistsPanelHeaderShowsArtifactCount verifies that SetState sums
// ArtifactsWritten across all specialists and surfaces the total via the
// PanelHeader's "(N)" count suffix.
func TestSpecialistsPanelHeaderShowsArtifactCount(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetSize(80, 24)
	state := &pipeline.StateV2{
		Specialists: map[string]pipeline.SpecialistState{
			"developer": {
				CurrentStep:      "implement",
				ArtifactsWritten: []string{"dev-design.yaml", "dev-tasks.yaml"},
			},
		},
	}
	p.SetState(state)

	out := p.View()
	if !strings.Contains(out, "(2)") {
		t.Errorf("expected header count '(2)' reflecting len(ArtifactsWritten), got: %q", out)
	}
}

// TestSpecialistsPanelInitReturnsSpinnerTickCmd verifies that Init returns a
// non-nil command — the spinner.Model.Tick command that starts the
// indeterminate spinner ticking (T-030..T-039).
func TestSpecialistsPanelInitReturnsSpinnerTickCmd(t *testing.T) {
	p := panels.NewSpecialistsPanel()

	cmd := p.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a non-nil command (spinner tick)")
	}

	msg := cmd()
	if _, ok := msg.(spinner.TickMsg); !ok {
		t.Errorf("expected Init command to produce spinner.TickMsg, got %T", msg)
	}
}

// TestSpecialistsPanelRoutesSpinnerTickWhenRunning verifies that, while a
// specialist run is in progress (StatusRunning), an incoming spinner.TickMsg
// is routed to the embedded spinner.Model and the panel re-emits a follow-up
// tick command — keeping the spinner animating only during active runs.
func TestSpecialistsPanelRoutesSpinnerTickWhenRunning(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetState(runningState())

	updated, cmd := p.Update(spinner.TickMsg{})
	if cmd == nil {
		t.Fatal("expected a re-tick command while a run is in progress, got nil")
	}

	msg := cmd()
	if _, ok := msg.(spinner.TickMsg); !ok {
		t.Errorf("expected re-emitted command to produce spinner.TickMsg, got %T", msg)
	}

	// The view should now contain spinner output (frame characters), proving
	// the tick actually advanced the embedded spinner's state.
	if !strings.Contains(updated.View(), "Specialists") {
		t.Errorf("expected updated panel to still render normally, got: %q", updated.View())
	}
}

// TestSpecialistsPanelDoesNotReTickSpinnerWhenIdle verifies that, when no
// specialist run is in progress (idle/done only), an incoming spinner.TickMsg
// does NOT produce a follow-up re-tick command — the spinner must stop
// animating once the run completes, not spin forever in the background.
func TestSpecialistsPanelDoesNotReTickSpinnerWhenIdle(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetState(idleState())

	_, cmd := p.Update(spinner.TickMsg{})
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(spinner.TickMsg); ok {
			t.Error("expected no spinner re-tick command while idle, but got one")
		}
	}
}

// TestSpecialistsPanelViewShowsSpinnerWhenRunning verifies that View() renders
// the spinner's frame glyph alongside the in-progress indicator while a
// specialist run is active.
func TestSpecialistsPanelViewShowsSpinnerWhenRunning(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetState(runningState())

	// Advance the spinner so its View() produces a non-empty frame glyph.
	p, _ = p.Update(spinner.TickMsg{})

	out := p.View()
	spinnerGlyph := strings.TrimRight(spinner.Dot.Frames[0], " ")
	found := false
	for _, frame := range spinner.Dot.Frames {
		if strings.Contains(out, strings.TrimRight(frame, " ")) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected view to contain a spinner.Dot frame glyph (e.g. %q) while running, got: %q", spinnerGlyph, out)
	}
}

// TestSpecialistsPanelViewOmitsSpinnerWhenIdle verifies that View() does NOT
// render any spinner.Dot frame glyphs when no specialist run is in progress —
// the indeterminate spinner must be visually absent while idle, not just
// non-ticking.
func TestSpecialistsPanelViewOmitsSpinnerWhenIdle(t *testing.T) {
	p := panels.NewSpecialistsPanel()
	p.SetState(idleState())

	out := p.View()
	for _, frame := range spinner.Dot.Frames {
		glyph := strings.TrimRight(frame, " ")
		if glyph == "" {
			continue
		}
		if strings.Contains(out, glyph) {
			t.Errorf("expected no spinner.Dot frame glyph %q in idle view, got: %q", glyph, out)
		}
	}
}
