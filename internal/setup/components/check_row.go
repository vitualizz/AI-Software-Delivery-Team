// Package components provides reusable TUI components for the setup screen.
package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/asdt/internal/tui/panels"
)

// CheckStatus represents the resolution state of a single pre-flight check row.
type CheckStatus int

const (
	// CheckStatusPending indicates the check has not yet resolved.
	CheckStatusPending CheckStatus = iota
	// CheckStatusOK indicates the check passed.
	CheckStatusOK
	// CheckStatusWarning indicates the check passed with a non-blocking warning.
	CheckStatusWarning
	// CheckStatusError indicates the check failed.
	CheckStatusError
)

// CheckRow holds the display state for one row in the pre-flight check screen.
type CheckRow struct {
	Label        string
	Status       CheckStatus
	Detail       string
	SpinnerFrame string // current spinner frame glyph (set by TickSpinner)
	SoftWarn     bool   // true = missing is non-blocking warning
}

// CheckResult carries the resolved outcome of a probe and is used to update
// the matching CheckRow via UpdateRow.
type CheckResult struct {
	Label    string
	Status   CheckStatus
	Detail   string
	SoftWarn bool
}

// TickSpinner updates SpinnerFrame for Pending rows to the given frame string.
// frame is the current spinner.View() string from the parent model's spinner.
func (r *CheckRow) TickSpinner(frame string) {
	if r.Status == CheckStatusPending {
		r.SpinnerFrame = frame
	}
}

// Render returns a styled single-line representation of the check row.
func (r CheckRow) Render(_ int) string {
	icon := statusIcon(r.Status, r.SpinnerFrame)
	label := lipgloss.NewStyle().Width(20).Render(r.Label)
	detail := ""
	if r.Detail != "" {
		detail = "  " + lipgloss.NewStyle().Faint(true).Render(r.Detail)
	}
	return "  " + icon + "  " + label + detail
}

// UpdateRow returns a new slice with the row matching label updated to result.
// Immutable: returns a new slice, does not mutate the input.
func UpdateRow(sections []SectionGroup, label string, result CheckResult) []SectionGroup {
	out := make([]SectionGroup, len(sections))
	copy(out, sections)
	for i, sg := range out {
		rows := make([]CheckRow, len(sg.Rows))
		copy(rows, sg.Rows)
		for j, row := range rows {
			if row.Label == label {
				rows[j].Status = result.Status
				rows[j].Detail = result.Detail
				rows[j].SoftWarn = result.SoftWarn
				rows[j].SpinnerFrame = ""
			}
		}
		out[i].Rows = rows
	}
	return out
}

func statusIcon(s CheckStatus, spinnerFrame string) string {
	switch s {
	case CheckStatusOK:
		return lipgloss.NewStyle().Foreground(panels.ColorSuccess).Render("✓")
	case CheckStatusWarning:
		return lipgloss.NewStyle().Foreground(panels.ColorInactive).Render("⚠")
	case CheckStatusError:
		return lipgloss.NewStyle().Foreground(panels.ColorError).Render("✗")
	default: // CheckStatusPending
		if spinnerFrame != "" {
			return spinnerFrame
		}
		return lipgloss.NewStyle().Foreground(panels.ColorInactive).Render("◌")
	}
}
