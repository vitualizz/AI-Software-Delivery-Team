package panels_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/asdt/internal/tui/panels"
)

func TestStatusBarRendersChangeAndSpecialist(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(120, 1)
	sb.SetChange("my-change")
	sb.SetSpecialist("developer")
	view := sb.View()

	if !strings.Contains(view, "my-change") {
		t.Errorf("expected change 'my-change' in status bar, got: %q", view)
	}
	if !strings.Contains(view, "developer") {
		t.Errorf("expected specialist 'developer' in status bar, got: %q", view)
	}
}

func TestStatusBarFullHintsAbove80(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(100, 1)
	sb.SetChange("my-change")
	sb.SetSpecialist("developer")
	view := sb.View()

	if !strings.Contains(view, "switch panel") {
		t.Errorf("expected keyboard hints at width > 80, got: %q", view)
	}
}

func TestStatusBarTruncatedAt80Width(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(70, 1)
	sb.SetChange("my-change")
	sb.SetSpecialist("developer")
	view := sb.View()

	if !strings.Contains(view, "my-change") {
		t.Errorf("expected change at width 70, got: %q", view)
	}
	if !strings.Contains(view, "developer") {
		t.Errorf("expected specialist at width 70, got: %q", view)
	}
}

func TestStatusBarChangeOnlyBelow50(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(40, 1)
	sb.SetChange("my-change")
	sb.SetSpecialist("developer")
	view := sb.View()

	if !strings.Contains(view, "my-change") {
		t.Errorf("expected change at width 40, got: %q", view)
	}
	if strings.Contains(view, "developer") {
		t.Errorf("expected no specialist at width <= 50, got: %q", view)
	}
}

// TestStatusBarCompactModeShowsMoreAvailableCue verifies that, in the
// narrowest width tier (compact mode, where the dashboard hides the
// artifacts panel entirely), the status bar surfaces an inline cue hinting
// that more content is available beyond what's currently shown.
func TestStatusBarCompactModeShowsMoreAvailableCue(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(40, 1)
	sb.SetChange("my-change")
	sb.SetCompact(true)
	view := sb.View()

	if !strings.Contains(view, "more") {
		t.Errorf("expected a 'more available' cue in compact mode, got: %q", view)
	}
}

// TestStatusBarNonCompactModeOmitsMoreAvailableCue verifies that the cue is
// additive to compact mode only — wider layouts that already show both
// panels should not display it.
func TestStatusBarNonCompactModeOmitsMoreAvailableCue(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(120, 1)
	sb.SetChange("my-change")
	sb.SetSpecialist("developer")
	sb.SetCompact(false)
	view := sb.View()

	if strings.Contains(view, "more") {
		t.Errorf("expected no 'more available' cue outside compact mode, got: %q", view)
	}
}

func TestStatusBarEmptyState(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(120, 1)
	view := sb.View()

	if view == "" {
		t.Errorf("expected non-empty status bar even with empty state, got: %q", view)
	}
}

func TestStatusBarUsesColorOnInactiveNotRawLiteral(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(100, 1)
	sb.SetChange("demo")
	sb.SetSpecialist("developer")

	rendered := sb.View()

	// Reference rendering built explicitly with ColorOnInactive over
	// ColorInactive, using the exact same line content View() produces at
	// width 100 (> 80 branch includes the keyboard hints) — if statusbar.go
	// still used the raw lipgloss.Color("15") literal, the foreground ANSI
	// sequence would differ from this reference for at least one of the
	// Light/Dark profiles.
	line := " change: demo  specialist: developer  [tab] switch panel  [q] quit"
	reference := lipgloss.NewStyle().
		Foreground(panels.ColorOnInactive).
		Background(panels.ColorInactive).
		Width(100).
		Render(line)

	if rendered != reference {
		t.Errorf("expected statusbar foreground to resolve via panels.ColorOnInactive, got a differently-styled render\nwant: %q\ngot:  %q", reference, rendered)
	}
}
