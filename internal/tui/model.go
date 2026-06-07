package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

type layoutMode int

const (
	modeHorizontal layoutMode = iota // ≥80 cols — side by side
	modeVertical                     // 50-79 cols — stacked
	modeCompact                      // <50 cols — compact stacked
)

// Model is the root Bubbletea model. It holds layout state and delegates
// message processing to its child panels. It never contains agent logic.
type Model struct {
	deps        Dependencies
	specialists panels.SpecialistsPanel
	artifacts   panels.ArtifactPanel
	statusBar   panels.StatusBar
	width       int
	height      int
	err         error
	ready       bool
	focused     int // 0 = specialists, 1 = artifacts
	mode        layoutMode
}

// New constructs a root Model with the given dependencies.
func New(deps Dependencies) Model {
	return Model{
		deps:        deps,
		specialists: panels.NewSpecialistsPanel(),
		artifacts:   panels.NewArtifactPanel(),
		statusBar:   panels.NewStatusBar(),
	}
}

// Init returns the initial batch of commands: load specialists state, load artifacts, start tick.
func (m Model) Init() tea.Cmd {
	change := m.deps.Change
	if change == "" {
		change = m.deps.Config.ActiveChange
	}
	if change == "" {
		change = "default"
	}
	return tea.Batch(
		LoadSpecialistsCmd(m.deps.PipelineStore, change),
		LoadArtifactsCmd(m.deps.ConfigRoot, change),
		TickCmd(),
	)
}

// Update is the root message handler. It routes messages to the focused panel
// and handles global concerns (resize, quit, focus toggle, error display).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.focused = (m.focused + 1) % 2
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		switch {
		case m.width >= 120:
			m.mode = modeHorizontal
		case m.width >= 50:
			m.mode = modeVertical
		default:
			m.mode = modeCompact
		}

		panelHeight := m.height - 2 // reserve 2 lines for status bar

		var specCmd, artCmd, sbCmd tea.Cmd
		if m.mode == modeHorizontal {
			specialistsWidth := m.width * 30 / 100
			artifactWidth := m.width - specialistsWidth
			m.specialists, specCmd = m.specialists.UpdateSize(specialistsWidth, panelHeight)
			m.artifacts, artCmd = m.artifacts.UpdateSize(artifactWidth, panelHeight)
		} else {
			// vertical or compact — full width for each
			artHeight := panelHeight / 2
			specHeight := panelHeight - artHeight
			m.specialists, specCmd = m.specialists.UpdateSize(m.width, specHeight)
			m.artifacts, artCmd = m.artifacts.UpdateSize(m.width, artHeight)
		}
		m.statusBar, sbCmd = m.statusBar.UpdateSize(m.width, 1)
		return m, tea.Batch(specCmd, artCmd, sbCmd)

	case SpecialistsLoadedMsg:
		m.specialists.SetState(msg.State)
		return m, nil

	case ArtifactListMsg:
		var cmd tea.Cmd
		m.artifacts, cmd = m.artifacts.SetFiles(msg.Files)
		return m, cmd

	case ArtifactContentMsg:
		var cmd tea.Cmd
		m.artifacts, cmd = m.artifacts.SetContent(msg.Content)
		return m, cmd

	case ErrorMsg:
		m.err = msg.Err
		return m, nil

	case TickMsg:
		change := m.deps.Change
		if change == "" {
			change = m.deps.Config.ActiveChange
		}
		if change == "" {
			change = "default"
		}
		return m, tea.Batch(
			LoadSpecialistsCmd(m.deps.PipelineStore, change),
			LoadArtifactsCmd(m.deps.ConfigRoot, change),
			TickCmd(),
		)
	}

	// Delegate to focused panel.
	var cmd tea.Cmd
	if m.focused == 0 {
		m.specialists, cmd = m.specialists.Update(msg)
	} else {
		m.artifacts, cmd = m.artifacts.Update(msg)
	}
	return m, cmd
}

// View renders the full TUI layout.
func (m Model) View() string {
	if m.err != nil {
		if strings.Contains(m.err.Error(), ".asdt/") ||
			strings.Contains(m.err.Error(), "no .asdt/ directory") {
			return "No ASDT project found. Run from inside a project."
		}
		return fmt.Sprintf("Error: %v", m.err)
	}

	if !m.ready {
		return "Initializing..."
	}

	var body string

	focusedStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(panels.ColorPrimary)
	unfocusedStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(panels.ColorInactive)

	switch m.mode {
	case modeHorizontal:
		specialistsWidth := m.width * 30 / 100
		artifactWidth := m.width - specialistsWidth

		specStyle := focusedStyle.Width(specialistsWidth)
		artStyle := unfocusedStyle.Width(artifactWidth)
		if m.focused == 1 {
			specStyle, artStyle = artStyle, specStyle
		}

		left := specStyle.Render(m.specialists.View())
		right := artStyle.Render(m.artifacts.View())
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	case modeVertical:
		artHeight := m.height / 2
		specHeight := m.height - artHeight

		specStyle := focusedStyle.Width(m.width)
		artStyle := unfocusedStyle.Width(m.width)
		if m.focused == 1 {
			specStyle, artStyle = artStyle, specStyle
		}

		top := specStyle.Height(specHeight).Render(m.specialists.View())
		bottom := artStyle.Height(artHeight).Render(m.artifacts.View())
		body = lipgloss.JoinVertical(lipgloss.Top, top, bottom)

	default: // modeCompact
		specStyle := focusedStyle.Width(m.width)
		if m.focused != 0 {
			specStyle = unfocusedStyle.Width(m.width)
		}
		body = specStyle.Render(m.specialists.View())
	}

	change := m.deps.Config.ActiveChange
	if change == "" {
		change = "default"
	}
	m.statusBar.SetChange(change)
	m.statusBar.SetSpecialist(m.specialists.SelectedSpecialist())

	return lipgloss.JoinVertical(lipgloss.Left, body, m.statusBar.View())
}
