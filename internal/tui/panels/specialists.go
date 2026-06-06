// Package panels contains the individual Bubbletea panel models for the TUI.
// Each panel is a self-contained tea.Model-compatible type that owns its own
// state and Update/View logic. The root model delegates to panels; panels
// never call back to the root.
package panels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/pipeline"
)

// Shared panel styles used across the panels package.
var (
	stylePanelTitle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	stylePending    = lipgloss.NewStyle().Faint(true)
)

// Known specialists in display order.
var specialistOrder = []string{"ux-ui", "architect", "developer", "qa", "security"}

var specialistNames = map[string]string{
	"ux-ui":     "UX/UI",
	"architect": "Architect",
	"developer": "Developer",
	"qa":        "QA",
	"security":  "Security",
}

// SpecialistsPanel renders one row per specialist showing their current state.
// It replaces the linear PipelinePanel for the specialist-first architecture.
type SpecialistsPanel struct {
	state    *pipeline.StateV2
	selected int
	focused  bool
	width    int
	height   int
}

// NewSpecialistsPanel returns a zero-value SpecialistsPanel ready for use.
func NewSpecialistsPanel() SpecialistsPanel {
	return SpecialistsPanel{}
}

// Init satisfies the tea.Model-compatible interface (no commands needed at init).
func (p SpecialistsPanel) Init() tea.Cmd { return nil }

// Update handles messages directed at the specialists panel.
func (p SpecialistsPanel) Update(msg tea.Msg) (SpecialistsPanel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.selected < len(specialistOrder)-1 {
				p.selected++
			}
		case "k", "up":
			if p.selected > 0 {
				p.selected--
			}
		}
	}
	return p, nil
}

// UpdateSize stores the new dimensions and returns the updated panel.
func (p SpecialistsPanel) UpdateSize(width, height int) (SpecialistsPanel, tea.Cmd) {
	p.width = width
	p.height = height
	return p, nil
}

// View renders the specialists panel.
func (p SpecialistsPanel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")) // cyan
	completedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))         // green
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("237"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("Specialists") + "\n\n")

	for i, id := range specialistOrder {
		name := specialistNames[id]
		var statusStr string

		if p.state != nil {
			if sp, ok := p.state.Specialists[id]; ok && len(sp.StepsCompleted) > 0 {
				last := sp.StepsCompleted[len(sp.StepsCompleted)-1]
				ts := last.Timestamp.Format("15:04")
				if sp.CurrentStep == last.ID {
					// current step matches last completed — specialist is done
					statusStr = completedStyle.Render(fmt.Sprintf("✓ %s  %s", sp.CurrentStep, ts))
				} else {
					statusStr = activeStyle.Render(fmt.Sprintf("▶ %s", sp.CurrentStep))
				}
			} else {
				statusStr = stylePending.Render("—")
			}
		} else {
			statusStr = stylePending.Render("—")
		}

		row := fmt.Sprintf("  %-12s  %s", name, statusStr)
		if i == p.selected && p.focused {
			row = selectedStyle.Render(row)
		}
		b.WriteString(row + "\n")
	}

	if p.state == nil {
		b.WriteString("\n" + stylePending.Render("No specialists have run yet") + "\n")
	}

	return b.String()
}

// SetState stores the loaded pipeline v2 state.
func (p *SpecialistsPanel) SetState(state *pipeline.StateV2) { p.state = state }

// SelectedSpecialist returns the specialist ID for the currently selected row.
func (p SpecialistsPanel) SelectedSpecialist() string {
	if p.selected < len(specialistOrder) {
		return specialistOrder[p.selected]
	}
	return ""
}

// SetFocused sets whether this panel has keyboard focus.
func (p *SpecialistsPanel) SetFocused(f bool) { p.focused = f }

// SetSize sets the panel dimensions.
func (p *SpecialistsPanel) SetSize(w, h int) { p.width = w; p.height = h }
