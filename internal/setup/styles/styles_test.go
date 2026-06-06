package styles_test

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/styles"
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
		{"Header", p.Header, true, true, false},
		{"Cursor", p.Cursor, true, true, false},
		{"Selected", p.Selected, true, false, false},
		{"Unselected", p.Unselected, true, false, false},
		{"Success", p.Success, true, false, false},
		{"Error", p.Error, true, false, false},
		{"Label", p.Label, true, false, false},
		{"Description", p.Description, true, false, true},
		{"Dim", p.Dim, false, false, true},
		{"Bold", p.Bold, false, true, false},
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

func TestSuccessHasGreenForeground(t *testing.T) {
	expected := lipgloss.Color("#a6e3a1")
	got := styles.Default.Success.GetForeground()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Success foreground = %v, want %v", got, expected)
	}
}

func TestErrorHasRedForeground(t *testing.T) {
	expected := lipgloss.Color("#f38ba8")
	got := styles.Default.Error.GetForeground()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Error foreground = %v, want %v", got, expected)
	}
}
