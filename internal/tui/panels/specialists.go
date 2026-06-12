// Package panels contains the individual Bubbletea panel models for the TUI.
// Each panel is a self-contained tea.Model-compatible type that owns its own
// state and Update/View logic. The root model delegates to panels; panels
// never call back to the root.
package panels

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/asdt/internal/pipeline"
)

// Shared panel styles used across the panels package.
var (
	stylePending = lipgloss.NewStyle().Faint(true)
)

// cursorRowGlyph mirrors setup's cursorChar convention (internal/setup/views.go)
// for visual consistency between the installer TUI and the dashboard.
const cursorRowGlyph = "►"

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
	header   PanelHeader
	state    *pipeline.StateV2
	selected int
	focused  bool
	width    int
	height   int
	compact  bool
	spinner  spinner.Model
}

// NewSpecialistsPanel returns a zero-value SpecialistsPanel ready for use.
func NewSpecialistsPanel() SpecialistsPanel {
	return SpecialistsPanel{header: NewPanelHeader("Specialists"), spinner: NewSpinner()}
}

// Init starts the indeterminate spinner ticking. The panel gates rendering
// and re-ticking on whether a run is actually in progress (see Update/View),
// so it is safe to always start the loop here — it self-stops once idle.
func (p SpecialistsPanel) Init() tea.Cmd {
	return p.spinner.Tick
}

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

	case spinner.TickMsg:
		// Distinct from this panel's own poll/refresh TickMsg (tui.TickMsg,
		// sourced from tui.TickCmd) — spinner.TickMsg drives only the
		// indeterminate animation frame and must never be confused with the
		// state-reload polling loop owned by the root model.
		if !hasRunningSpecialist(p.state) {
			// No active run: let the tick die here so the spinner stops
			// animating (and the goroutine loop ends) once idle.
			return p, nil
		}
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		return p, cmd
	}
	return p, nil
}

// hasRunningSpecialist reports whether any specialist currently has an
// in-progress step — the same signal View() already derives per-row
// (CurrentStep != the last completed step's ID ⇒ StatusRunning). Centralizing
// it here lets Update gate the spinner's tick loop on the identical condition
// that drives the StatusRunning glyph, so the two never disagree.
func hasRunningSpecialist(state *pipeline.StateV2) bool {
	if state == nil {
		return false
	}
	for _, sp := range state.Specialists {
		if len(sp.StepsCompleted) == 0 {
			continue
		}
		last := sp.StepsCompleted[len(sp.StepsCompleted)-1]
		if sp.CurrentStep != last.ID {
			return true
		}
	}
	return false
}

// UpdateSize stores the new dimensions and returns the updated panel.
func (p SpecialistsPanel) UpdateSize(width, height int) (SpecialistsPanel, tea.Cmd) {
	p.width = width
	p.height = height
	p.compact = width <= 60
	var cmd tea.Cmd
	p.header, cmd = p.header.UpdateSize(width, height)
	return p, cmd
}

// View renders the specialists panel.
func (p SpecialistsPanel) View() string {
	var b strings.Builder
	b.WriteString(p.header.View())
	b.WriteString("\n")
	if p.state != nil && p.state.ChangeID != "" {
		b.WriteString(stylePending.Render("change: " + p.state.ChangeID))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	for i, id := range specialistOrder {
		name := specialistNames[id]
		var statusStr string
		var historyLines []string

		if p.state != nil {
			if sp, ok := p.state.Specialists[id]; ok && len(sp.StepsCompleted) > 0 {
				last := sp.StepsCompleted[len(sp.StepsCompleted)-1]
				ts := last.Timestamp.Format("15:04")
				var state StatusState
				if sp.CurrentStep == last.ID {
					state = StatusDone
				} else {
					state = StatusRunning
				}
				si := NewStatusIndicator(state)
				si.compact = p.compact
				statusStr = si.Render() + " " + sp.CurrentStep + "  " + ts
				if state == StatusRunning {
					// Indeterminate spinner rides alongside the in-progress
					// indicator — it is the only row-level cue that the
					// pipeline is actively advancing (vs. merely "running"
					// per stale state, e.g. after a crash).
					statusStr += " " + p.spinner.View()
				}

				if !p.compact {
					historyLines = renderStepHistory(sp.StepsCompleted)
				}
			} else {
				si := NewStatusIndicator(StatusIdle)
				si.compact = p.compact
				statusStr = si.Render() + " " + stylePending.Render("—")
			}
		} else {
			statusStr = stylePending.Render("—")
		}

		var row string
		if p.compact {
			row = fmt.Sprintf("%s %s", name, statusStr)
		} else {
			row = fmt.Sprintf("  %-12s  %s", name, statusStr)
		}

		if i == p.selected && p.focused {
			row = cursorRowGlyph + " " + row
			selStyle := lipgloss.NewStyle().Background(ColorInactive)
			row = selStyle.Render(row)
		}
		b.WriteString(row + "\n")
		for _, line := range historyLines {
			b.WriteString(line + "\n")
		}
	}

	if p.state == nil {
		b.WriteString("\n" + stylePending.Render("No specialists have run yet") + "\n")
		b.WriteString(stylePending.Render("Run /asdt \"<feature description>\" to start a specialist") + "\n")
	}

	return b.String()
}

// renderStepHistory renders one indented, muted line per completed step,
// interleaving the computed elapsed-duration since the previous step.
// Returns nil for 0 or 1 steps (nothing to interleave).
func renderStepHistory(steps []pipeline.StepRecord) []string {
	if len(steps) == 0 {
		return nil
	}

	durations := ComputeStepDurations(steps)
	durationByStepID := make(map[string]time.Duration, len(durations))
	for _, d := range durations {
		durationByStepID[d.To.ID] = d.Elapsed
	}

	lines := make([]string, 0, len(steps))
	for _, step := range steps {
		ts := step.Timestamp.Format("15:04")
		line := fmt.Sprintf("      • %s  %s", step.ID, ts)
		if elapsed, ok := durationByStepID[step.ID]; ok {
			line += "  (+" + FormatDuration(elapsed) + ")"
		}
		lines = append(lines, stylePending.Render(line))
	}
	return lines
}

// SetState stores the loaded pipeline v2 state and updates the header count
// to reflect the total number of artifacts written across all specialists.
func (p *SpecialistsPanel) SetState(state *pipeline.StateV2) {
	p.state = state
	if state == nil {
		p.header.SetCount(-1)
		return
	}
	total := 0
	for _, sp := range state.Specialists {
		total += len(sp.ArtifactsWritten)
	}
	p.header.SetCount(total)
}

// SelectedSpecialist returns the specialist ID for the currently selected row.
func (p SpecialistsPanel) SelectedSpecialist() string {
	if p.selected < len(specialistOrder) {
		return specialistOrder[p.selected]
	}
	return ""
}

// SetFocused sets whether this panel has keyboard focus.
func (p *SpecialistsPanel) SetFocused(f bool) {
	p.focused = f
	p.header.SetFocused(f)
}

// SetSize sets the panel dimensions.
func (p *SpecialistsPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.compact = w <= 60
	p.header.width = w
}
