package styles_test

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/styles"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// noFg is the zero-value foreground (NoColor{}).
var noFg = lipgloss.NewStyle().GetForeground()

// TestDefaultPaletteAllFieldsNonZero verifies that each style in the palette
// has at least one attribute set (foreground, bold, or faint) so it is not
// equivalent to an un-styled zero value.
func TestDefaultPaletteAllFieldsNonZero(t *testing.T) {
	p := styles.Default

	cases := []struct {
		name      string
		style     lipgloss.Style
		wantFg    bool // true if style should have a non-default foreground
		wantBold  bool
		wantFaint bool
	}{
		{"Cursor", p.Cursor, true, true, false},
		{"Success", p.Success, true, false, false},
		{"Error", p.Error, true, false, false},
		{"Dim", p.Dim, false, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantFg {
				fg := tc.style.GetForeground()
				if reflect.DeepEqual(fg, noFg) {
					t.Errorf("styles.Default.%s has no foreground color set", tc.name)
				}
			}
			if tc.wantBold && !tc.style.GetBold() {
				t.Errorf("styles.Default.%s should be bold but is not", tc.name)
			}
			if tc.wantFaint && !tc.style.GetFaint() {
				t.Errorf("styles.Default.%s should be faint but is not", tc.name)
			}
		})
	}
}

func TestSuccessStyleRendersNonEmpty(t *testing.T) {
	got := styles.Default.Success.Render("ok")
	if got == "" {
		t.Error("Success.Render returned empty string")
	}
}

func TestErrorStyleRendersNonEmpty(t *testing.T) {
	got := styles.Default.Error.Render("err")
	if got == "" {
		t.Error("Error.Render returned empty string")
	}
}

func TestSuccessHasAdaptiveColor(t *testing.T) {
	fg := styles.Default.Success.GetForeground()
	if reflect.DeepEqual(fg, lipgloss.NoColor{}) {
		t.Error("Success foreground is NoColor, expected an AdaptiveColor value")
	}
}

func TestErrorHasAdaptiveColor(t *testing.T) {
	fg := styles.Default.Error.GetForeground()
	if reflect.DeepEqual(fg, lipgloss.NoColor{}) {
		t.Error("Error foreground is NoColor, expected an AdaptiveColor value")
	}
}

func TestSuccessAndErrorColorsUsePanelVars(t *testing.T) {
	if !reflect.DeepEqual(styles.Default.Success.GetForeground(), panels.ColorSuccess) {
		t.Error("Success foreground does not match panels.ColorSuccess")
	}
	if !reflect.DeepEqual(styles.Default.Error.GetForeground(), panels.ColorError) {
		t.Error("Error foreground does not match panels.ColorError")
	}
}

// TestBoxHasRoundedBorderAndPadding verifies that the Box style wraps content
// in a rounded border, with Padding(1, 2). Border colors use panel AdaptiveColor
// via FocusBorderStyle, so we only check the border style, not specific hex values.
func TestBoxHasRoundedBorderAndPadding(t *testing.T) {
	b := styles.Default.Box

	if b.GetBorderStyle() != lipgloss.RoundedBorder() {
		t.Errorf("Box border style = %v, want lipgloss.RoundedBorder()", b.GetBorderStyle())
	}

	if top, right, bottom, left := b.GetPadding(); top != 1 || right != 2 || bottom != 1 || left != 2 {
		t.Errorf("Box padding = (%d,%d,%d,%d), want (1,2,1,2)", top, right, bottom, left)
	}
}

