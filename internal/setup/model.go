package setup

import (
	"io/fs"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

// ViewState represents the current screen shown by the TUI.
type ViewState int

const (
	// StateEngramMissing is the zero-value state; the TUI starts here until
	// EngramCheckCmd reports whether the engram binary is on PATH.
	StateEngramMissing ViewState = iota
	// StateMainMenu is the main menu screen.
	StateMainMenu
	// StateAssistantList displays the list of available assistants.
	StateAssistantList
	// StateSelectAssistants allows the user to choose assistants to install.
	StateSelectAssistants
	// StateSelectProvider allows the user to choose a memory provider.
	StateSelectProvider
	// StateInstalling shows installation progress.
	StateInstalling
	// StateDone is the terminal state shown after installation completes.
	StateDone
)

// Model is the root Bubbletea model for the installer TUI.
type Model struct {
	state           ViewState
	cursor          int
	selected        map[int]bool
	provider        int
	results         []installer.InstallResult
	skillsFS        fs.FS
	currentVersion  string
	latestVersion   string
	updateAvailable bool
}

// New constructs an initial Model with the running binary version. Init()
// fires EngramCheckCmd to determine whether to transition to StateMainMenu or
// remain at StateEngramMissing, and UpdateCheckCmd to passively detect newer
// releases.
func New(skillsFS fs.FS, version string) Model {
	return Model{
		selected:       make(map[int]bool),
		skillsFS:       skillsFS,
		currentVersion: version,
	}
}

// State returns the current ViewState. Exported for tests.
func (m Model) State() ViewState { return m.state }

// Init implements tea.Model. Fires async engram PATH and update checks.
func (m Model) Init() tea.Cmd {
	return tea.Batch(EngramCheckCmd(), UpdateCheckCmd(m.currentVersion))
}

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EngramCheckMsg:
		if msg.Found {
			m.state = StateMainMenu
		} else {
			m.state = StateEngramMissing
		}
		return m, nil

	case UpdateCheckMsg:
		if msg.Err != nil {
			return m, nil // silent: no banner on any error/timeout/rate-limit
		}
		if newerAvailable(msg.Current, msg.Latest) {
			m.latestVersion = msg.Latest
			m.updateAvailable = true
		}
		return m, nil

	case InstallDoneMsg:
		m.results = msg.Results
		m.state = StateDone
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// StateEngramMissing handles its own keys exclusively.
	if m.state == StateEngramMissing {
		return m.handleEngramMissing(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyRunes:
		if string(msg.Runes) == "q" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateMainMenu:
		return m.handleMainMenu(msg)
	case StateAssistantList:
		return m.handleAssistantList(msg)
	case StateSelectAssistants:
		return m.handleSelectAssistants(msg)
	case StateSelectProvider:
		return m.handleSelectProvider(msg)
	case StateDone:
		return m.handleDone(msg)
	}
	return m, nil
}

// handleEngramMissing only allows q and ctrl+c; all other keys are no-ops.
func (m Model) handleEngramMissing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyRunes:
		if string(msg.Runes) == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) handleMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < 1 {
			m.cursor++
		}
	case tea.KeyEnter:
		if m.cursor == 0 {
			m.state = StateAssistantList
			m.cursor = 0
		} else {
			return m, tea.Quit
		}
	case tea.KeyEsc:
		// No-op at MainMenu.
	}
	return m, nil
}

func (m Model) handleAssistantList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := len(installer.Descriptors)
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < count-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		m.state = StateSelectAssistants
		m.cursor = 0
	case tea.KeyEsc:
		m.state = StateMainMenu
		m.cursor = 0
	}
	return m, nil
}

func (m Model) handleSelectAssistants(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := len(installer.Descriptors)
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < count-1 {
			m.cursor++
		}
	case tea.KeySpace:
		m.selected[m.cursor] = !m.selected[m.cursor]
	case tea.KeyEnter:
		m.state = StateSelectProvider
		m.cursor = 0
	case tea.KeyEsc:
		m.state = StateAssistantList
		m.cursor = 0
	}
	return m, nil
}

func (m Model) handleSelectProvider(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := len(installer.Providers)
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < count-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		m.provider = m.cursor
		m.state = StateInstalling
		m.cursor = 0
		return m, m.buildInstallCmd()
	case tea.KeyEsc:
		m.state = StateSelectAssistants
		m.cursor = 0
	}
	return m, nil
}

func (m Model) handleDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.state = StateMainMenu
		m.cursor = 0
		m.results = nil
		m.selected = make(map[int]bool)
	}
	return m, nil
}

func (m Model) buildInstallCmd() tea.Cmd {
	var assistants []installer.AssistantDescriptor
	for i, d := range installer.Descriptors {
		if m.selected[i] {
			assistants = append(assistants, d)
		}
	}
	if len(assistants) == 0 {
		assistants = installer.Descriptors
	}
	provider := installer.Providers[m.provider]
	return InstallCmd(assistants, provider, m.skillsFS)
}

// View renders the current state.
func (m Model) View() string {
	return renderState(m)
}
