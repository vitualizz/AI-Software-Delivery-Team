package panels_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

func TestNewBadgeRenderIncludesLabel(t *testing.T) {
	b := panels.NewBadge("present", panels.ColorSuccess)
	out := b.Render()

	if !strings.Contains(out, "present") {
		t.Errorf("expected badge render to include label %q, got: %q", "present", out)
	}
}

func TestNewBadgeDifferentTonesRenderDifferently(t *testing.T) {
	// Force a color-capable profile: the test environment may report Ascii
	// (no color), which would make any two AdaptiveColor renders compare
	// equal regardless of tone — defeating the assertion this test exists for.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	success := panels.NewBadge("present", panels.ColorSuccess).Render()
	failure := panels.NewBadge("present", panels.ColorError).Render()

	if success == failure {
		t.Errorf("expected badges with different tones to render differently, both rendered: %q", success)
	}
}
