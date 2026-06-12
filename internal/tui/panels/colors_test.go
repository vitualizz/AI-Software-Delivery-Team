package panels_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/asdt/internal/tui/panels"
)

// relativeLuminance computes the WCAG relative luminance L for a hex color
// (sRGB -> linear -> weighted sum: L = 0.2126R + 0.7152G + 0.0722B).
func relativeLuminance(hex string) (float64, error) {
	hex = trimHexPrefix(hex)
	if len(hex) != 6 {
		return 0, errInvalidHex(hex)
	}

	r, err := hexChannel(hex[0:2])
	if err != nil {
		return 0, err
	}
	g, err := hexChannel(hex[2:4])
	if err != nil {
		return 0, err
	}
	b, err := hexChannel(hex[4:6])
	if err != nil {
		return 0, err
	}

	return 0.2126*r + 0.7152*g + 0.0722*b, nil
}

func trimHexPrefix(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}

type errInvalidHex string

func (e errInvalidHex) Error() string { return "invalid hex color: " + string(e) }

// hexChannel decodes a 2-digit hex channel and linearizes it per sRGB.
func hexChannel(component string) (float64, error) {
	v, err := strconv.ParseUint(component, 16, 16)
	if err != nil {
		return 0, err
	}
	c := float64(v) / 255.0
	if c <= 0.03928 {
		return c / 12.92, nil
	}
	return math.Pow((c+0.055)/1.055, 2.4), nil
}

// contrastRatio returns the WCAG contrast ratio between two luminances.
func contrastRatio(l1, l2 float64) float64 {
	lighter, darker := l1, l2
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

const (
	statusPairThreshold = 1.5
	textOnBgThreshold   = 4.5
)

func assertRatioAtLeast(t *testing.T, label string, lightA, lightB, darkA, darkB lipgloss.AdaptiveColor, threshold float64) {
	t.Helper()

	checks := []struct {
		variant string
		hexA    string
		hexB    string
	}{
		{"light", lightA.Light, lightB.Light},
		{"dark", darkA.Dark, darkB.Dark},
	}

	for _, c := range checks {
		la, err := relativeLuminance(c.hexA)
		if err != nil {
			t.Fatalf("%s %s: %v", label, c.variant, err)
		}
		lb, err := relativeLuminance(c.hexB)
		if err != nil {
			t.Fatalf("%s %s: %v", label, c.variant, err)
		}
		ratio := contrastRatio(la, lb)
		if ratio < threshold {
			t.Errorf("%s (%s): contrast ratio %.2f below threshold %.2f", label, c.variant, ratio, threshold)
		}
	}
}

func TestPaletteStatusPairsAreDistinct(t *testing.T) {
	assertRatioAtLeast(t, "Success vs Error", panels.ColorSuccess, panels.ColorError, panels.ColorSuccess, panels.ColorError, statusPairThreshold)
	assertRatioAtLeast(t, "Primary vs Error", panels.ColorPrimary, panels.ColorError, panels.ColorPrimary, panels.ColorError, statusPairThreshold)
	assertRatioAtLeast(t, "Primary vs Success", panels.ColorPrimary, panels.ColorSuccess, panels.ColorPrimary, panels.ColorSuccess, statusPairThreshold)
}

func TestPaletteOnInactiveReadableOnInactiveBackground(t *testing.T) {
	assertRatioAtLeast(t, "OnInactive on Inactive", panels.ColorOnInactive, panels.ColorInactive, panels.ColorOnInactive, panels.ColorInactive, textOnBgThreshold)
}
