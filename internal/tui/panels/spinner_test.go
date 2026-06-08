package panels_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// TestNewSpinnerUsesDotStyle verifies that the constructed spinner uses the
// bubbles spinner.Dot frame set — the indeterminate "in progress" glyph
// agreed for this track (see implementation plan T-030..T-039).
func TestNewSpinnerUsesDotStyle(t *testing.T) {
	s := panels.NewSpinner()

	if len(s.Spinner.Frames) != len(spinner.Dot.Frames) {
		t.Fatalf("expected %d frames (spinner.Dot), got %d", len(spinner.Dot.Frames), len(s.Spinner.Frames))
	}
	for i, frame := range spinner.Dot.Frames {
		if s.Spinner.Frames[i] != frame {
			t.Errorf("expected frame %d to be %q (spinner.Dot), got %q", i, frame, s.Spinner.Frames[i])
		}
	}
	if s.Spinner.FPS != spinner.Dot.FPS {
		t.Errorf("expected FPS %v (spinner.Dot), got %v", spinner.Dot.FPS, s.Spinner.FPS)
	}
}

// TestNewSpinnerStyleResolvesToColorSecondary verifies that the spinner's
// style renders its frame tinted with panels.ColorSecondary — the cyan/sky
// pastel reserved for in-progress/running indicators (see StatusColor).
//
// Force a color-capable profile: the test environment may report Ascii
// (no color), which would make any AdaptiveColor render compare equal to a
// plain, unstyled render — defeating the assertion this test exists for.
// (Track B.3 hit the same termenv.Ascii-in-tests gotcha — see badge_test.go.)
func TestNewSpinnerStyleResolvesToColorSecondary(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	s := panels.NewSpinner()
	got := s.Style.Render("x")

	want := lipgloss.NewStyle().Foreground(panels.ColorSecondary).Render("x")
	if got != want {
		t.Errorf("expected spinner style to resolve to ColorSecondary, got %q want %q", got, want)
	}

	// Sanity: it must differ from an unstyled render — proving the tint
	// actually changes the rendered output rather than degrading to plain text.
	plain := lipgloss.NewStyle().Render("x")
	if got == plain {
		t.Errorf("expected tinted spinner style to render differently from plain text, both rendered: %q", got)
	}
}
