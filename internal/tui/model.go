package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// Model is the root Bubbletea model. It holds layout state and delegates
// message processing to its child panels. It never contains agent logic.
type Model struct {
	deps      Dependencies
	pipeline  panels.PipelinePanel
	artifacts panels.ArtifactPanel
	width     int
	height    int
	err       error
	ready     bool
	focused   int // 0 = pipeline, 1 = artifacts
}

// New constructs a root Model with the given dependencies.
func New(deps Dependencies) Model {
	return Model{
		deps:      deps,
		pipeline:  panels.NewPipelinePanel(),
		artifacts: panels.NewArtifactPanel(),
	}
}

// Init returns the initial batch of commands: load pipeline, load artifacts, start tick.
func (m Model) Init() tea.Cmd {
	change := m.deps.Change
	if change == "" {
		change = m.deps.Config.ActiveChange
	}
	if change == "" {
		change = "default"
	}
	return tea.Batch(
		LoadPipelineCmd(m.deps.PipelineStore, change),
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
		pipelineWidth := msg.Width * 30 / 100
		artifactWidth := msg.Width - pipelineWidth
		panelHeight := msg.Height - 2 // reserve 2 lines for status bar

		var pipeCmd, artCmd tea.Cmd
		m.pipeline, pipeCmd = m.pipeline.UpdateSize(pipelineWidth, panelHeight)
		m.artifacts, artCmd = m.artifacts.UpdateSize(artifactWidth, panelHeight)
		return m, tea.Batch(pipeCmd, artCmd)

	case PipelineLoadedMsg:
		var cmd tea.Cmd
		m.pipeline, cmd = m.pipeline.SetState(msg.State)
		return m, cmd

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
			LoadPipelineCmd(m.deps.PipelineStore, change),
			LoadArtifactsCmd(m.deps.ConfigRoot, change),
			TickCmd(),
		)
	}

	// Delegate to focused panel.
	var cmd tea.Cmd
	if m.focused == 0 {
		m.pipeline, cmd = m.pipeline.Update(msg)
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

	pipelineWidth := m.width * 30 / 100
	artifactWidth := m.width - pipelineWidth

	pipelineStyle := lipgloss.NewStyle().Width(pipelineWidth)
	artifactStyle := lipgloss.NewStyle().Width(artifactWidth)

	if m.focused == 0 {
		pipelineStyle = pipelineStyle.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6"))
		artifactStyle = artifactStyle.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
	} else {
		pipelineStyle = pipelineStyle.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
		artifactStyle = artifactStyle.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6"))
	}

	left := pipelineStyle.Render(m.pipeline.View())
	right := artifactStyle.Render(m.artifacts.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	change := m.deps.Config.ActiveChange
	if change == "" {
		change = "default"
	}
	pipelineState := m.pipeline.CurrentState()
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("237")).
		Width(m.width).
		Render(fmt.Sprintf(" change: %s  state: %s  [tab] switch panel  [q] quit", change, pipelineState))

	return lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
}
