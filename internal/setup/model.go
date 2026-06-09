package setup

import (
	"io/fs"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/components"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// ViewState represents the current screen shown by the TUI.
type ViewState int

const (
	// StateMainMenu is the zero-value state; the TUI starts here showing the
	// top-level menu. The user chooses between installing/updating skills,
	// viewing the dashboard, or quitting.
	StateMainMenu ViewState = iota
	// StateEnvironmentCheck is shown after the user selects "Install / Update
	// Skills". It fans out three environment probes (OS, Engram, Codegraph)
	// concurrently; Continue is gated until all probes resolve and engram is found.
	StateEnvironmentCheck
	// StateDashboard shows the dashboard overview with per-assistant status.
	StateDashboard
	// StateAssistantList displays the list of available assistants.
	StateAssistantList
	// StateSelectAssistants allows the user to choose assistants to install.
	StateSelectAssistants
	// StateSelectProvider allows the user to choose a memory provider.
	StateSelectProvider
	// StateAgentSetup shows the persona selection step (optional).
	StateAgentSetup
	// StateInstalling shows installation progress.
	StateInstalling
	// StateDone is the terminal state shown after installation completes.
	StateDone
)

// Model is the root Bubbletea model for the installer TUI.
type Model struct {
	state             ViewState
	cursor            int
	selected          map[int]bool
	provider          int
	results           []installer.InstallResult
	agentResults      []installer.AgentConfigResult
	skillsFS          fs.FS
	currentVersion    string
	latestVersion     string
	updateAvailable   bool
	width             int
	spinner           spinner.Model
	preflightSections []components.SectionGroup
	preflightDone     bool
	engramMissing     bool
	selectedPersona   int  // index into installer.PersonaPresets (0=Axiom,1=Sage,2=Forge,3=Lee Palacios)
	agentSetupSkip    bool // user chose to skip agent config entirely
	agentOverwrite    bool // user confirmed overwriting an existing config
	agentConflicts    []string // target paths that already have AGENTS.md
	agentConfigExists map[string]bool // assistantID → whether global config already exists
}

// New constructs an initial Model with the running binary version. Init()
// fires UpdateCheckCmd to passively detect newer releases and starts the
// spinner tick. EnvironmentCheckCmd is NOT fired in Init() — it is triggered
// on demand when the user selects "Install / Update Skills" from StateMainMenu.
//
// The spinner is built via panels.NewSpinner — exported from the panels
// package specifically so this installer TUI and the dashboard TUI share one
// indeterminate-spinner construction (Dot frames, ColorSecondary tint)
// instead of each hand-rolling its own and risking drift.
func New(skillsFS fs.FS, version string) Model {
	return Model{
		selected:          make(map[int]bool),
		skillsFS:          skillsFS,
		currentVersion:    version,
		width:             80,
		spinner:           panels.NewSpinner(),
		preflightSections: initialPreflightSections(),
	}
}

// State returns the current ViewState. Exported for tests.
func (m Model) State() ViewState { return m.state }

// Init implements tea.Model. Fires the update check and starts the spinner tick.
// EnvironmentCheckCmd is NOT fired here — it is triggered on demand when the
// user selects "Install / Update Skills" from the main menu.
func (m Model) Init() tea.Cmd {
	return tea.Batch(UpdateCheckCmd(m.currentVersion), m.spinner.Tick)
}

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EnvironmentCheckProgressMsg:
		progress := components.CheckResult{
			Label:    msg.RowLabel,
			Status:   msg.Status,
			Detail:   msg.Detail,
			SoftWarn: msg.SoftWarn,
		}
		m.preflightSections = components.UpdateRow(m.preflightSections, msg.RowLabel, progress)
		if allProbesDone(m.preflightSections) {
			engramFound := !rowHasStatus(m.preflightSections, "Engram", components.CheckStatusError)
			codegraphFound := !rowHasStatus(m.preflightSections, "Codegraph", components.CheckStatusWarning)
			return m, func() tea.Msg {
				return EnvironmentCheckMsg{EngramFound: engramFound, CodegraphFound: codegraphFound}
			}
		}
		return m, nil

	case EnvironmentCheckMsg:
		m.preflightDone = true
		m.engramMissing = !msg.EngramFound
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

	case AgentInstallDoneMsg:
		m.agentResults = msg.Results
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case spinner.TickMsg:
		// Gated to StateInstalling and StateEnvironmentCheck: the spinner has
		// no visual representation in any other screen.
		if m.state != StateInstalling && m.state != StateEnvironmentCheck {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.state == StateEnvironmentCheck {
			// Advance per-row spinner frames for still-pending rows.
			frame := m.spinner.View()
			for i := range m.preflightSections {
				for j := range m.preflightSections[i].Rows {
					m.preflightSections[i].Rows[j].TickSpinner(frame)
				}
			}
		}
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// StateEnvironmentCheck handles its own keys exclusively.
	if m.state == StateEnvironmentCheck {
		return m.handleEnvironmentCheck(msg)
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
	case StateDashboard:
		return m.handleDashboard(msg)
	case StateMainMenu:
		return m.handleMainMenu(msg)
	case StateAssistantList:
		return m.handleAssistantList(msg)
	case StateSelectAssistants:
		return m.handleSelectAssistants(msg)
	case StateSelectProvider:
		return m.handleSelectProvider(msg)
	case StateAgentSetup:
		return m.handleAgentSetup(msg)
	case StateDone:
		return m.handleDone(msg)
	}
	return m, nil
}

// handleEnvironmentCheck handles keys for the pre-flight check screen.
// Enter proceeds to StateAssistantList only when preflight is done and engram was found.
// Esc returns to StateMainMenu at any point.
// q and ctrl+c always quit.
func (m Model) handleEnvironmentCheck(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		if m.preflightDone && !m.engramMissing {
			m.state = StateAssistantList
			m.cursor = 0
			return m, nil
		}
	case tea.KeyEsc:
		m.state = StateMainMenu
		m.cursor = 0
		return m, nil
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
		if m.cursor < 2 {
			m.cursor++
		}
	case tea.KeyEnter:
		switch m.cursor {
		case 0: // Install / Update Skills → trigger preflight
			m.preflightSections = initialPreflightSections()
			m.preflightDone = false
			m.engramMissing = false
			m.state = StateEnvironmentCheck
			m.cursor = 0
			return m, tea.Batch(EnvironmentCheckCmd(), m.spinner.Tick)
		case 1: // Dashboard
			m.state = StateDashboard
			m.cursor = 0
		default: // Quit
			return m, tea.Quit
		}
	case tea.KeyEsc:
		// No-op at MainMenu.
	}
	return m, nil
}

func (m Model) handleDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = StateMainMenu
		m.cursor = 0
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
		m.state = StateAgentSetup
		m.cursor = 0
		m.agentConflicts = m.detectAgentConflicts()
		return m, nil
	case tea.KeyEsc:
		m.state = StateSelectAssistants
		m.cursor = 0
	}
	return m, nil
}

// detectAgentConflicts checks which selected assistants already have an AGENTS.md
// at their global config location.
func (m Model) detectAgentConflicts() []string {
	var conflicts []string
	assistants := m.selectedAssistants()
	for _, a := range assistants {
		adapter, ok := installer.AgentConfigAdapterFor(a.ID)
		if !ok {
			continue
		}
		if adapter.AgentConfigExists() {
			conflicts = append(conflicts, a.ID)
		}
	}
	return conflicts
}

// selectedAssistants returns the list of assistants selected by the user (or all
// if none were explicitly selected, mirroring buildInstallCmd).
func (m Model) selectedAssistants() []installer.AssistantDescriptor {
	var assistants []installer.AssistantDescriptor
	for i, d := range installer.Descriptors {
		if m.selected[i] {
			assistants = append(assistants, d)
		}
	}
	if len(assistants) == 0 {
		assistants = installer.Descriptors
	}
	return assistants
}

func (m Model) handleAgentSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 5 options: 4 presets + Skip
	const optionCount = 5
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < optionCount-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		if m.cursor < len(installer.PersonaPresets) {
			// User selected a preset.
			m.selectedPersona = m.cursor
			m.agentSetupSkip = false
		} else {
			// User selected Skip.
			m.agentSetupSkip = true
		}
		m.agentOverwrite = len(m.agentConflicts) == 0 // no conflicts means overwrite is moot; conflicts default to overwrite
		m.state = StateInstalling
		m.cursor = 0
		return m, tea.Batch(m.buildInstallCmd(), m.spinner.Tick)
	case tea.KeyEsc:
		m.state = StateSelectProvider
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
	assistants := m.selectedAssistants()
	provider := installer.Providers[m.provider]
	installCmd := InstallCmd(assistants, provider, m.skillsFS)
	if m.agentSetupSkip {
		return installCmd
	}
	preset := installer.PersonaPresets[m.selectedPersona]
	agentCmd := AgentInstallCmd(assistants, preset, m.agentOverwrite, m.skillsFS)
	return tea.Batch(installCmd, agentCmd)
}

// View renders the current state.
func (m Model) View() string {
	return renderState(m)
}

func allProbesDone(sections []components.SectionGroup) bool {
	for _, sg := range sections {
		for _, row := range sg.Rows {
			if row.Status == components.CheckStatusPending {
				return false
			}
		}
	}
	return true
}

func rowHasStatus(sections []components.SectionGroup, label string, status components.CheckStatus) bool {
	for _, sg := range sections {
		for _, row := range sg.Rows {
			if row.Label == label {
				return row.Status == status
			}
		}
	}
	return false
}
