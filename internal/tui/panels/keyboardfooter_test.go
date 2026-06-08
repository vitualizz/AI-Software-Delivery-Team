package panels_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

func TestKeyboardFooterRendersFullHintsAbove80(t *testing.T) {
	hints := []panels.HintGroup{
		{Hints: []panels.Hint{
			{Key: "tab", Description: "switch panel"},
			{Key: "q", Description: "quit"},
		}},
	}
	result := panels.RenderKeyboardFooter(hints, 100)

	if !strings.Contains(result, "switch panel") {
		t.Errorf("expected description 'switch panel' at width > 80, got: %q", result)
	}
	if !strings.Contains(result, "quit") {
		t.Errorf("expected description 'quit' at width > 80, got: %q", result)
	}
	if !strings.Contains(result, "|") {
		t.Errorf("expected '|' separator between hints at width > 80, got: %q", result)
	}
}

func TestKeyboardFooterShowsOnlyKeysAtOrBelow80(t *testing.T) {
	hints := []panels.HintGroup{
		{Hints: []panels.Hint{
			{Key: "tab", Description: "switch panel"},
		}},
	}
	result := panels.RenderKeyboardFooter(hints, 80)

	if strings.Contains(result, "switch panel") {
		t.Errorf("expected no description at width <= 80, got: %q", result)
	}
	if !strings.Contains(result, "tab") {
		t.Errorf("expected key 'tab' at width <= 80, got: %q", result)
	}
}

// TestKeyboardFooterCompactAtOrBelow50 supersedes the previous
// TestKeyboardFooterEmptyAtOrBelow50 expectation: AC-5 requires a non-empty
// compact footer (key tokens joined by '|', no descriptions) at narrow
// widths instead of an empty string.
func TestKeyboardFooterCompactAtOrBelow50(t *testing.T) {
	hints := []panels.HintGroup{
		{Hints: []panels.Hint{
			{Key: "tab", Description: "switch panel"},
			{Key: "q", Description: "quit"},
		}},
	}

	for _, w := range []int{50, 40, 25} {
		t.Run("width", func(t *testing.T) {
			result := panels.RenderKeyboardFooter(hints, w)
			if result == "" {
				t.Errorf("expected non-empty compact footer at width %d, got empty string", w)
			}
			if !strings.Contains(result, "tab") || !strings.Contains(result, "q") {
				t.Errorf("expected compact footer to include key tokens 'tab' and 'q' at width %d, got: %q", w, result)
			}
			if strings.Contains(result, "switch panel") || strings.Contains(result, "quit") {
				t.Errorf("expected NO descriptions in compact footer at width %d, got: %q", w, result)
			}
			if !strings.Contains(result, "|") {
				t.Errorf("expected '|' separator in compact footer at width %d, got: %q", w, result)
			}
		})
	}
}

// TestKeyboardFooterEmptyHintsStillEmptyAtAnyWidth verifies that an empty
// groups slice renders empty regardless of width — the compact branch must
// not synthesize content from nothing.
func TestKeyboardFooterEmptyHintsStillEmptyAtAnyWidth(t *testing.T) {
	for _, w := range []int{100, 60, 40} {
		result := panels.RenderKeyboardFooter([]panels.HintGroup{}, w)
		if result != "" {
			t.Errorf("expected empty for nil hints at width %d, got: %q", w, result)
		}
	}
}

func TestKeyboardFooterEmptyHints(t *testing.T) {
	result := panels.RenderKeyboardFooter([]panels.HintGroup{}, 100)
	if result != "" {
		t.Errorf("expected empty for nil hints, got: %q", result)
	}
}
