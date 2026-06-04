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

// PipelinePanel renders the four pipeline phases and highlights the active one.
type PipelinePanel struct {
	state   *pipeline.State
	focused bool
	width   int
	height  int
}

// NewPipelinePanel returns a zero-value PipelinePanel ready for use.
func NewPipelinePanel() PipelinePanel {
	return PipelinePanel{}
}

// Init satisfies the tea.Model-compatible interface (no commands needed at init).
func (p PipelinePanel) Init() tea.Cmd { return nil }

// Update handles messages directed at the pipeline panel.
func (p PipelinePanel) Update(msg tea.Msg) (PipelinePanel, tea.Cmd) {
	return p, nil
}

// UpdateSize stores the new dimensions and returns updated panel + nil cmd.
func (p PipelinePanel) UpdateSize(width, height int) (PipelinePanel, tea.Cmd) {
	p.width = width
	p.height = height
	return p, nil
}

// SetState stores the loaded pipeline state and returns updated panel + nil cmd.
func (p PipelinePanel) SetState(state pipeline.State) (PipelinePanel, tea.Cmd) {
	p.state = &state
	return p, nil
}

// CurrentState returns the current phase as a string (empty if no state loaded).
func (p PipelinePanel) CurrentState() string {
	if p.state == nil {
		return ""
	}
	return string(p.state.CurrentState)
}

// phases lists all pipeline phases in execution order.
var phases = []pipeline.Phase{
	pipeline.PhaseRequirements,
	pipeline.PhasePlan,
	pipeline.PhaseImplement,
	pipeline.PhaseReview,
}

var (
	styleCompleted = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	styleActive    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")) // bold cyan
	stylePending   = lipgloss.NewStyle().Faint(true)                      // dim
	stylePanelTitle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
)

// View renders the pipeline panel.
func (p PipelinePanel) View() string {
	if p.state == nil {
		return stylePending.Render("No pipeline state found")
	}

	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render("Pipeline"))
	sb.WriteString("\n")

	current := p.state.CurrentState
	// Determine which phases are completed.
	completed := completedPhases(current)

	for _, phase := range phases {
		line := renderPhase(phase, current, completed, p.state)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Show transition history summary.
	if len(p.state.Transitions) > 0 {
		sb.WriteString("\n")
		sb.WriteString(stylePending.Render(fmt.Sprintf("transitions: %d", len(p.state.Transitions))))
	}

	return sb.String()
}

// completedPhases returns a set of phases that are "before" current in the pipeline.
func completedPhases(current pipeline.Phase) map[pipeline.Phase]bool {
	completed := map[pipeline.Phase]bool{}
	for _, p := range phases {
		if p == current {
			break
		}
		completed[p] = true
	}
	return completed
}

// renderPhase renders a single phase line with the appropriate style and indicator.
func renderPhase(phase, current pipeline.Phase, completed map[pipeline.Phase]bool, state *pipeline.State) string {
	name := string(phase)

	if completed[phase] {
		// Find the timestamp for this phase from transitions.
		ts := transitionTimestamp(state, phase)
		if ts != "" {
			return styleCompleted.Render(fmt.Sprintf("✓ %-16s %s", name, ts))
		}
		return styleCompleted.Render(fmt.Sprintf("✓ %s", name))
	}

	if phase == current {
		return styleActive.Render(fmt.Sprintf("▶ %s", name))
	}

	return stylePending.Render(fmt.Sprintf("  %s", name))
}

// transitionTimestamp returns the formatted timestamp when the given phase became
// the destination of a transition (i.e., when it was first entered).
func transitionTimestamp(state *pipeline.State, phase pipeline.Phase) string {
	for _, t := range state.Transitions {
		if t.To == phase {
			return t.Timestamp.Format("2006-01-02 15:04")
		}
	}
	// Also check if this was the starting phase (no "To" transition).
	if phase == pipeline.PhaseRequirements && len(state.Transitions) > 0 {
		return state.Transitions[0].Timestamp.Format("2006-01-02 15:04")
	}
	return ""
}
