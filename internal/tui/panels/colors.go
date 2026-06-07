package panels

import "github.com/charmbracelet/lipgloss"

// AdaptiveColor palette shared across panels.
var (
	ColorPrimary   = lipgloss.AdaptiveColor{Light: "#6D28D9", Dark: "#7C3AED"}
	ColorSecondary = lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#06B6D4"}
	ColorSuccess   = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"}
	ColorError     = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}
	ColorWarning   = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"}
	ColorInactive  = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#4B5563"}
	ColorMuted     = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
)
