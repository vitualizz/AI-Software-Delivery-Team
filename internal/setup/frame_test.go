package setup

import (
	"strings"
	"testing"
)

// TestFrameComposesTitleBodyFooterWithBorder verifies that the shared frame()
// helper joins title, body, and footer and wraps the result in a rounded
// border (proving Box + StatusBar + JoinVertical composition).
func TestFrameComposesTitleBodyFooterWithBorder(t *testing.T) {
	got := frame("T", "B", "F", true)

	for _, want := range []string{"T", "B", "F"} {
		if !strings.Contains(got, want) {
			t.Errorf("frame output missing %q, got:\n%s", want, got)
		}
	}

	if !strings.ContainsAny(got, "╭╮╰╯│") {
		t.Errorf("frame output missing rounded-border runes, got:\n%s", got)
	}
}
