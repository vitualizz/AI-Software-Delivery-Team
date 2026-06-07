package panels_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

func TestPanelHeaderRendersTitle(t *testing.T) {
	ph := panels.NewPanelHeader("Specialists")
	view := ph.View()

	if !strings.Contains(view, "▎") {
		t.Errorf("expected decorator '▎' in header, got: %q", view)
	}
	if !strings.Contains(view, "Specialists") {
		t.Errorf("expected title 'Specialists' in header, got: %q", view)
	}
}

func TestPanelHeaderShowsCountBadge(t *testing.T) {
	ph := panels.NewPanelHeader("Artifacts")
	ph, _ = ph.UpdateSize(80, 1)
	ph.SetCount(3)
	view := ph.View()

	if !strings.Contains(view, "(3)") {
		t.Errorf("expected count badge '(3)' in header, got: %q", view)
	}
}

func TestPanelHeaderHidesBadgeBelow20Width(t *testing.T) {
	ph := panels.NewPanelHeader("Artifacts")
	ph, _ = ph.UpdateSize(15, 1)
	ph.SetCount(3)
	view := ph.View()

	if strings.Contains(view, "(3)") {
		t.Errorf("expected badge hidden at width < 20, got: %q", view)
	}
}

func TestPanelHeaderTruncatesTitleBelow15Width(t *testing.T) {
	ph := panels.NewPanelHeader("Specialists")
	ph, _ = ph.UpdateSize(10, 1)
	view := ph.View()

	if !strings.Contains(view, "…") {
		t.Errorf("expected ellipsis truncation at width < 15, got: %q", view)
	}
}

func TestPanelHeaderFocusedColor(t *testing.T) {
	ph := panels.NewPanelHeader("Specialists")
	ph.SetFocused(true)
	view := ph.View()

	if !strings.Contains(view, "Specialists") {
		t.Errorf("expected title when focused, got: %q", view)
	}
}
