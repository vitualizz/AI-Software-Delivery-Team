package panels

import "github.com/charmbracelet/lipgloss"

// RenderHero renders the product hero block for the main menu.
// It intentionally carries no border — frame() in the parent view owns that.
// Line 1: product name in bold + ColorPrimary.
// Line 2: empty spacer.
// Line 3: version string in ColorMuted.
// If version is empty or "dev", the version line reads "dev build".
func RenderHero(version string) string {
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	versionStyle := lipgloss.NewStyle().Foreground(ColorMuted)

	versionLabel := "v" + version
	if version == "" || version == "dev" {
		versionLabel = "dev build"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		nameStyle.Render("asdt-tui"),
		"",
		versionStyle.Render(versionLabel),
	)
}
