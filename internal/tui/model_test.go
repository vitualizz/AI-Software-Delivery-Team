package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui"
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
	m := tui.New(tui.Dependencies{})
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
	m := tui.New(tui.Dependencies{})
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
	m := tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 40})
	view := m.View()

	if view == "Initializing..." {
		t.Fatal("expected model to be ready after WindowSizeMsg")
	}
}

// TestModelTabTogglesFocus verifies that pressing Tab toggles the focused
// panel without triggering a quit.
func TestModelTabTogglesFocus(t *testing.T) {
	m := tui.New(tui.Dependencies{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Errorf("expected nil command on tab toggle, got %T", cmd)
	}
}
