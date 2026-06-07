package panels_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
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

func TestStatusBarEmptyState(t *testing.T) {
	sb := panels.NewStatusBar()
	sb, _ = sb.UpdateSize(120, 1)
	view := sb.View()

	if view == "" {
		t.Errorf("expected non-empty status bar even with empty state, got: %q", view)
	}
}
