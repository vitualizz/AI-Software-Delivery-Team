package tui_test

import (
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
