package tui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/asdt/internal/pipeline"
	"github.com/vitualizz/asdt/internal/tui"
)

// TestModelInitNoPanic verifies that New() and Init() do not panic when
// the Dependencies struct is zero-valued (no .asdt/ found yet).
func TestModelInitNoPanic(t *testing.T) {
	m := tui.New(tui.Dependencies{})

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = m.Init()
	}()
	if recovered != nil {
		t.Fatalf("Init panicked: %v", recovered)
	}
}

// TestModelWindowSizePropagates verifies that a WindowSizeMsg updates the model
// and marks it as ready without returning an error.
func TestModelWindowSizePropagates(t *testing.T) {
	m := tui.New(tui.Dependencies{})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := updated.View()

	// After a WindowSizeMsg the model is ready — it should not return
	// the "Initializing..." placeholder.
	if view == "Initializing..." {
		t.Error("expected model to be ready after WindowSizeMsg, but View returned 'Initializing...'")
	}
}

// TestModelQuitKey verifies that pressing 'q' returns the tea.Quit command.
func TestModelQuitKey(t *testing.T) {
	m := tui.New(tui.Dependencies{})

	// First make the model ready via a WindowSizeMsg.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	_, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected a non-nil command after 'q' keypress")
	}

	// Execute the command and check it is a quit message.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// TestModelCtrlCQuits verifies that ctrl+c returns the tea.Quit command.
func TestModelCtrlCQuits(t *testing.T) {
	m := tui.New(tui.Dependencies{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	_, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected a non-nil command after ctrl+c")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// TestModelHorizontalLayout verifies that at >=80 width both panels render
// side by side with horizontal layout.
func TestModelHorizontalLayout(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()

	if view == "Initializing..." {
		t.Fatal("expected model to be ready after WindowSizeMsg")
	}
	if !strings.Contains(view, "Specialists") {
		t.Errorf("expected Specialists in horizontal layout, got: %q", view[:min(200, len(view))])
	}
	if !strings.Contains(view, "Artifacts") {
		t.Errorf("expected Artifacts in horizontal layout, got: %q", view[:min(200, len(view))])
	}
}

// TestModelVerticalLayout verifies that at 50-79 width both panels render
// stacked top-to-bottom.
func TestModelVerticalLayout(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 40})
	view := m.View()

	if view == "Initializing..." {
		t.Fatal("expected model to be ready after WindowSizeMsg")
	}
	if !strings.Contains(view, "Specialists") {
		t.Errorf("expected Specialists in vertical layout, got: %q", view[:min(200, len(view))])
	}
	if !strings.Contains(view, "Artifacts") {
		t.Errorf("expected Artifacts in vertical layout, got: %q", view[:min(200, len(view))])
	}
}

// TestModelCompactLayout verifies that at <50 width the model renders
// without crashing.
func TestModelCompactLayout(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 40})
	view := m.View()

	if view == "Initializing..." {
		t.Fatal("expected model to be ready after WindowSizeMsg")
	}
}

// TestModelCompactLayoutShowsMoreAvailableCue verifies that, when the
// dashboard hides the artifacts panel in compact mode, the status bar
// surfaces an inline cue indicating more content exists than is shown.
func TestModelCompactLayoutShowsMoreAvailableCue(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 40})
	view := m.View()

	if !strings.Contains(view, "more") {
		t.Errorf("expected a 'more available' cue in compact layout, got: %q", view[len(view)-min(200, len(view)):])
	}
}

// TestModelViewIncludesKeyboardFooter verifies that View() composes the
// keyboard footer (rendered via panels.RenderKeyboardFooter) into the root
// layout, surfacing the actual key bindings handled in Update: tab to
// switch panels, j/k to navigate, enter to select, q to quit.
func TestModelViewIncludesKeyboardFooter(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()

	for _, key := range []string{"tab", "j", "k", "enter", "q"} {
		if !strings.Contains(view, key) {
			t.Errorf("expected keyboard footer to surface key %q in View() output, got: %q", key, view[:min(400, len(view))])
		}
	}
}

// TestModelTabTogglesFocus verifies that pressing Tab toggles the focused
// panel without triggering a quit.
func TestModelTabTogglesFocus(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Errorf("expected nil command on tab toggle, got %T", cmd)
	}
}

// TestModelInitBatchIncludesSpecialistsSpinnerTick verifies that the root
// model's Init() batches in the specialists panel's own Init() command
// (which starts the indeterminate spinner ticking) — proving the spinner
// loop begins as soon as the dashboard launches, not only after the panel
// happens to receive focus.
func TestModelInitBatchIncludesSpecialistsSpinnerTick(t *testing.T) {
	m := tui.New(tui.Dependencies{})

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a non-nil batched command")
	}

	found := false
	var walk func(c tea.Cmd)
	walk = func(c tea.Cmd) {
		if c == nil {
			return
		}
		msg := c()
		switch typed := msg.(type) {
		case tea.BatchMsg:
			for _, sub := range typed {
				walk(sub)
			}
		case spinner.TickMsg:
			found = true
		}
	}
	walk(cmd)

	if !found {
		t.Error("expected Init's batched commands to include the specialists panel's spinner tick (spinner.TickMsg)")
	}
}

// TestModelRoutesSpinnerTickToSpecialistsRegardlessOfFocus verifies that a
// spinner.TickMsg reaches the specialists panel's Update even when the
// artifacts panel has focus — the spinner must keep animating cross-cuttingly,
// independent of which panel currently owns keyboard focus.
func TestModelRoutesSpinnerTickToSpecialistsRegardlessOfFocus(t *testing.T) {
	var m tea.Model = tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Switch focus away from specialists (focused == 0) to artifacts (== 1).
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Load a StateV2 with an in-progress specialist via the same message the
	// root model already routes from LoadSpecialistsCmd — this is the public
	// surface for getting the specialists panel into a "running" state.
	now := time.Now()
	m, _ = m.Update(tui.SpecialistsLoadedMsg{State: &pipeline.StateV2{
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
	}})

	_, cmd := m.Update(spinner.TickMsg{})
	if cmd == nil {
		t.Fatal("expected spinner.TickMsg routed to specialists panel to produce a re-tick command, even while artifacts panel is focused")
	}

	msg := cmd()
	if _, ok := msg.(spinner.TickMsg); !ok {
		t.Errorf("expected re-emitted command to produce spinner.TickMsg, got %T", msg)
	}
}
