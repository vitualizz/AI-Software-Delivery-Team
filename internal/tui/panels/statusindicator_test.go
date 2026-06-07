package panels_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

func TestStatusIconPerState(t *testing.T) {
	tests := []struct {
		name     string
		state    panels.StatusState
		expected string
	}{
		{name: "idle", state: panels.StatusIdle, expected: "\u25cf"},
		{name: "running", state: panels.StatusRunning, expected: "\u25cc"},
		{name: "done", state: panels.StatusDone, expected: "\u2713"},
		{name: "error", state: panels.StatusError, expected: "\u2717"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := panels.StatusIcon(tc.state)
			if got != tc.expected {
				t.Errorf("StatusIcon(%s): got %q, want %q", tc.name, got, tc.expected)
			}
		})
	}
}

func TestStatusColorPerState(t *testing.T) {
	tests := []struct {
		name  string
		state panels.StatusState
	}{
		{name: "idle", state: panels.StatusIdle},
		{name: "running", state: panels.StatusRunning},
		{name: "done", state: panels.StatusDone},
		{name: "error", state: panels.StatusError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := panels.StatusColor(tc.state)
			rendered := lipgloss.NewStyle().Foreground(c).Render("test")
			if rendered == "" {
				t.Errorf("StatusColor(%s) produced empty render", tc.name)
			}
		})
	}
}

func TestNewStatusIndicatorRender(t *testing.T) {
	states := []panels.StatusState{
		panels.StatusIdle,
		panels.StatusRunning,
		panels.StatusDone,
		panels.StatusError,
	}
	for _, s := range states {
		si := panels.NewStatusIndicator(s)
		rendered := si.Render()
		if rendered == "" {
			t.Errorf("NewStatusIndicator(%d).Render() returned empty", s)
		}
	}
}

func TestFocusBorderStyleFocusedRender(t *testing.T) {
	style := panels.FocusBorderStyle(true)
	rendered := style.Render("test")
	if !strings.Contains(rendered, "test") {
		t.Errorf("FocusBorderStyle(true) render lost content, got: %q", rendered)
	}
}

func TestFocusBorderStyleUnfocusedRender(t *testing.T) {
	style := panels.FocusBorderStyle(false)
	rendered := style.Render("test")
	if !strings.Contains(rendered, "test") {
		t.Errorf("FocusBorderStyle(false) render lost content, got: %q", rendered)
	}
}
