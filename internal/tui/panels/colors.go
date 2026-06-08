package panels

import "github.com/charmbracelet/lipgloss"

// AdaptiveColor palette shared across panels.
//
// Pink-dominant pastel family: ColorPrimary anchors the accent hue, with
// sky-blue/mint/warm-mauve supporting tones. Success and Error intentionally
// sit in distinct luminance bands from each other and from Primary — they
// carry STATUS MEANING, so legibility-at-a-glance wins over strict palette-
// family purity (see dev-design palette_hex_values.rationale_summary). All
// values pass the WCAG relative-luminance self-check in colors_test.go.
var (
	ColorPrimary    = lipgloss.AdaptiveColor{Light: "#BE185D", Dark: "#F472B6"}
	ColorSecondary  = lipgloss.AdaptiveColor{Light: "#0E7490", Dark: "#7DD3FC"}
	ColorSuccess    = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#BBF7D0"}
	ColorError      = lipgloss.AdaptiveColor{Light: "#7F1D1D", Dark: "#DC2626"}
	ColorInactive   = lipgloss.AdaptiveColor{Light: "#C4B5BD", Dark: "#6B6470"}
	ColorMuted      = lipgloss.AdaptiveColor{Light: "#E7DDE3", Dark: "#9B8E97"}
	ColorOnInactive = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"}
)
