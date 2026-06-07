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

func TestKeyboardFooterEmptyAtOrBelow50(t *testing.T) {
	hints := []panels.HintGroup{
		{Hints: []panels.Hint{
			{Key: "tab", Description: "switch panel"},
		}},
	}

	for _, w := range []int{50, 40, 10} {
		t.Run("width", func(t *testing.T) {
			result := panels.RenderKeyboardFooter(hints, w)
			if result != "" {
				t.Errorf("expected empty at width %d, got: %q", w, result)
			}
		})
	}
}

func TestKeyboardFooterEmptyHints(t *testing.T) {
	result := panels.RenderKeyboardFooter([]panels.HintGroup{}, 100)
	if result != "" {
		t.Errorf("expected empty for nil hints, got: %q", result)
	}
}
