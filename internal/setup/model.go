package setup

import (
	"io/fs"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitualizz/ai-software-delivery-team/internal/i18n"
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
	// Skills". It fans out environment probes (OS, Shell, Engram, Codegraph)
	// concurrently; Continue is gated until all probes resolve and engram is found.
	StateEnvironmentCheck
	// StateDashboard shows the dashboard overview with per-assistant status.
	StateDashboard
	// StateSelectAssistants allows the user to choose assistants to install.
	// Shows present/missing status badges alongside checkboxes.
	StateSelectAssistants
	// StateSelectProvider allows the user to choose a memory provider.
	StateSelectProvider
	// StateAgentSetup shows the persona selection step (optional).
	StateAgentSetup
	// StateEmojiPref asks whether the assistants should use emojis. Only
	// reached on the non-skip path out of StateAgentSetup.
	StateEmojiPref
	// StateAgentWriteMode is shown when conflicts are detected; the user
	// chooses per-assistant how to handle existing configs (cycle with space).
	StateAgentWriteMode
	// StateReview shows a summary of all selections before installation begins.
	StateReview
	// StateInstalling shows installation progress.
	StateInstalling
	// StateDone is the terminal state shown after installation completes.
	StateDone
	// StateLanguageSelect asks which language the installer (and the installed
	// experience) should use. Entered from MainMenu's Install action, before
	// the environment check. Appended at the END of the iota block so existing
	// state values never shift.
	StateLanguageSelect
)

// preflightState holds state for the StateEnvironmentCheck screen.
type preflightState struct {
	sections      []components.SectionGroup
	done          bool
	engramMissing bool
}

// wizardState holds install-wizard state across SelectAssistants → Installing → Done.
type wizardState struct {
	selected        map[int]bool
	provider        int
	results         []installer.InstallResult
	agentResults    []installer.AgentConfigResult
	installExpected int  // number of per-assistant install Cmds fired
	agentDone       bool // true once AgentInstallDoneMsg arrives (or skip)
}

// dashboardState holds metadata loaded when the user enters StateDashboard.
type dashboardState struct {
	meta map[installer.AssistantID]installer.InstallMeta
}

// languageState holds state for the StateLanguageSelect screen.
type languageState struct {
	selected int  // index into languageOptions
	touched  bool // true once the user moved the selection; guards late LanguagePrefMsg
}

// languageOptions lists the selectable installer languages. Labels are the
// languages' NATIVE names — deliberately not catalog strings, so each option
// is readable regardless of the currently active language.
var languageOptions = []struct {
	Code  string
	Label string
}{
	{Code: "en", Label: "English"},
	{Code: "es", Label: "Español"},
}

// languageIndex returns the languageOptions index for code, or -1 when the
// code has no option.
func languageIndex(code string) int {
	for i, opt := range languageOptions {
		if opt.Code == code {
			return i
		}
	}
	return -1
}

// agentConfigState holds per-assistant agent config state across AgentSetup,
// EmojiPref, and AgentWriteMode.
type agentConfigState struct {
	selectedPersona int
	skip            bool
	useEmojis       bool
	writeModes      map[string]installer.AgentWriteMode
	conflicts       []string
}

// Model is the root Bubbletea model for the installer TUI.
type Model struct {
	state   ViewState
	cursor  int
	width   int
	spinner spinner.Model

	skillsFS        fs.FS
	catalog         i18n.Catalog
	currentVersion  string
	latestVersion   string
	updateAvailable bool

	preflight   preflightState
	wizard      wizardState
	agentConfig agentConfigState
	dashboard   dashboardState
	language    languageState
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
	str := i18n.Active()
	selected := languageIndex(i18n.ActiveCode())
	if selected < 0 {
		selected = 0
	}
	return Model{
		wizard:         wizardState{selected: make(map[int]bool)},
		skillsFS:       skillsFS,
		currentVersion: version,
		width:          80,
		spinner:        panels.NewSpinner(),
		preflight:      preflightState{sections: initialPreflightSections(str.Installer)},
		catalog:        str,
		language:       languageState{selected: selected},
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
		m.preflight.sections = components.UpdateRow(m.preflight.sections, msg.RowLabel, progress)
		if allProbesDone(m.preflight.sections) {
			engramFound := !rowHasStatus(m.preflight.sections, "Engram", components.CheckStatusError)
			codegraphFound := !rowHasStatus(m.preflight.sections, "Codegraph", components.CheckStatusWarning)
			return m, func() tea.Msg {
				return EnvironmentCheckMsg{EngramFound: engramFound, CodegraphFound: codegraphFound}
			}
		}
		return m, nil

	case EnvironmentCheckMsg:
		m.preflight.done = true
		m.preflight.engramMissing = !msg.EngramFound
		return m, nil

	case LanguagePrefMsg:
		// Apply the persisted language only while the user hasn't moved the
		// selection — a late-arriving message must never override an explicit
		// choice.
		if m.language.touched {
			return m, nil
		}
		if idx := languageIndex(msg.Code); idx >= 0 {
			m.language.selected = idx
			m.catalog = i18n.ForCode(msg.Code)
			if m.state == StateLanguageSelect {
				m.cursor = idx
			}
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

	case AssistantInstallProgressMsg:
		m.wizard.results = append(m.wizard.results, msg.Result)
		if m.allInstallsDone() {
			m.state = StateDone
		}
		return m, nil

	case InstallDoneMsg:
		// Kept for test compatibility: direct sends of InstallDoneMsg still
		// transition to StateDone without going through per-assistant tracking.
		m.wizard.results = msg.Results
		m.state = StateDone
		return m, nil

	case AgentInstallDoneMsg:
		m.wizard.agentResults = msg.Results
		m.wizard.agentDone = true
		if m.allInstallsDone() {
			m.state = StateDone
		}
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
			for i := range m.preflight.sections {
				for j := range m.preflight.sections[i].Rows {
					m.preflight.sections[i].Rows[j].TickSpinner(frame)
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
	case StateLanguageSelect:
		return m.handleLanguageSelect(msg)
	case StateSelectAssistants:
		return m.handleSelectAssistants(msg)
	case StateSelectProvider:
		return m.handleSelectProvider(msg)
	case StateAgentSetup:
		return m.handleAgentSetup(msg)
	case StateEmojiPref:
		return m.handleEmojiPref(msg)
	case StateAgentWriteMode:
		return m.handleAgentWriteMode(msg)
	case StateReview:
		return m.handleReview(msg)
	case StateDone:
		return m.handleDone(msg)
	}
	return m, nil
}

// handleEnvironmentCheck handles keys for the pre-flight check screen.
// Enter proceeds to StateSelectAssistants only when preflight is done and engram was found.
// All assistants are pre-selected on first entry so the default is visible.
// Esc returns to StateMainMenu at any point.
// q and ctrl+c always quit.
func (m Model) handleEnvironmentCheck(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		if m.preflight.done && !m.preflight.engramMissing {
			m.state = StateSelectAssistants
			if len(m.wizard.selected) == 0 {
				for i := range installer.Descriptors {
					m.wizard.selected[i] = true
				}
			}
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
		case 0: // Install / Update Skills → choose language first
			m.state = StateLanguageSelect
			m.cursor = m.language.selected
			return m, LanguagePrefCmd()
		case 1: // Dashboard
			m.state = StateDashboard
			m.cursor = 0
			m.dashboard = loadDashboardState()
		default: // Quit
			return m, tea.Quit
		}
	case tea.KeyEsc:
		// No-op at MainMenu.
	}
	return m, nil
}

// handleLanguageSelect handles keys for the language selection screen.
// Up/Down move the radio selection and re-resolve the catalog live so the
// whole TUI switches language immediately. Enter triggers the preflight
// bootstrap (previously fired straight from MainMenu); Esc returns to MainMenu.
func (m Model) handleLanguageSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m = m.syncLanguageSelection()
		}
	case tea.KeyDown:
		if m.cursor < len(languageOptions)-1 {
			m.cursor++
			m = m.syncLanguageSelection()
		}
	case tea.KeyEnter:
		m.preflight.sections = initialPreflightSections(m.catalog.Installer)
		m.preflight.done = false
		m.preflight.engramMissing = false
		m.state = StateEnvironmentCheck
		m.cursor = 0
		return m, tea.Batch(EnvironmentCheckCmd(), m.spinner.Tick)
	case tea.KeyEsc:
		m.state = StateMainMenu
		m.cursor = 0
	}
	return m, nil
}

// syncLanguageSelection returns a Model with the language selection following
// the cursor, the touched guard set, and the catalog re-resolved for the
// newly selected language.
func (m Model) syncLanguageSelection() Model {
	m.language.selected = m.cursor
	m.language.touched = true
	m.catalog = i18n.ForCode(languageOptions[m.language.selected].Code)
	return m
}

func (m Model) handleDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
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
		if m.cursor < count { // confirm button sits at index count
			m.cursor++
		}
	case tea.KeySpace:
		if m.cursor < count { // guard: confirm button has no selectable index
			m.wizard.selected[m.cursor] = !m.wizard.selected[m.cursor]
		}
	case tea.KeyRunes:
		if string(msg.Runes) == "a" {
			allSelected := true
			for i := range installer.Descriptors {
				if !m.wizard.selected[i] {
					allSelected = false
					break
				}
			}
			for i := range installer.Descriptors {
				m.wizard.selected[i] = !allSelected
			}
		}
	case tea.KeyEnter:
		m.state = StateSelectProvider
		m.cursor = 0
	case tea.KeyEsc:
		m.state = StateEnvironmentCheck
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
			if m.cursor < count {
				m.wizard.provider = m.cursor // live-track selection so confirm button works correctly
			}
		}
	case tea.KeyDown:
		if m.cursor < count { // confirm button sits at index count
			m.cursor++
			if m.cursor < count {
				m.wizard.provider = m.cursor
			}
		}
	case tea.KeyEnter:
		// m.wizard.provider already reflects the live selection; no cursor assignment needed.
		m.state = StateAgentSetup
		m.cursor = 0
		m.agentConfig.conflicts = m.detectAgentConflicts()
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
		if m.wizard.selected[i] {
			assistants = append(assistants, d)
		}
	}
	if len(assistants) == 0 {
		assistants = installer.Descriptors
	}
	return assistants
}

func (m Model) handleAgentSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 4 navigable presets (0-3); [ Skip → ] button sits at index len(PersonaPresets).
	presets := len(installer.PersonaPresets)
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m = m.syncAgentSetupSelection()
		}
	case tea.KeyDown:
		if m.cursor < presets { // [ Skip → ] button at index presets
			m.cursor++
			m = m.syncAgentSetupSelection()
		}
	case tea.KeyEnter:
		if m.cursor == presets {
			// User pressed Enter on the [ Skip → ] button — bypass EmojiPref.
			m.agentConfig.skip = true
			m.state = StateReview
			m.cursor = 0
			return m, nil
		}
		// m.agentConfig.selectedPersona reflects the current choice; ask the
		// emoji preference next, defaulting to Yes (cursor 0).
		m.agentConfig.useEmojis = true
		m.state = StateEmojiPref
		m.cursor = 0
		return m, nil
	case tea.KeyEsc:
		m.state = StateSelectProvider
		m.cursor = 0
	}
	return m, nil
}

// handleEmojiPref handles keys for the emoji preference screen. Two radio rows:
// cursor 0 = Yes, cursor 1 = No; useEmojis stays in sync with the cursor.
// Enter continues to AgentWriteMode when conflicts exist, otherwise to Review.
func (m Model) handleEmojiPref(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m.agentConfig.useEmojis = m.cursor == 0
		}
	case tea.KeyDown:
		if m.cursor < 1 {
			m.cursor++
			m.agentConfig.useEmojis = m.cursor == 0
		}
	case tea.KeyEnter:
		if len(m.agentConfig.conflicts) > 0 {
			m.agentConfig.writeModes = make(map[string]installer.AgentWriteMode)
			for _, id := range m.agentConfig.conflicts {
				m.agentConfig.writeModes[id] = installer.AgentModeSkip // safe default: Keep
			}
			m.state = StateAgentWriteMode
			m.cursor = 0
			return m, nil
		}
		m.state = StateReview
		m.cursor = 0
		return m, nil
	case tea.KeyEsc:
		m.state = StateAgentSetup
		m.cursor = 0
	}
	return m, nil
}

// syncAgentSetupSelection returns a Model with selectedPersona / skip
// updated to match the current cursor. Only called for cursor positions within
// the preset list (0 to len(PersonaPresets)-1); the [ Skip → ] button position
// is handled explicitly in the Enter handler.
func (m Model) syncAgentSetupSelection() Model {
	if m.cursor < len(installer.PersonaPresets) {
		m.agentConfig.selectedPersona = m.cursor
		m.agentConfig.skip = false
	}
	return m
}

func (m Model) handleAgentWriteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := len(m.agentConfig.conflicts)
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < count { // confirm button sits at index count
			m.cursor++
		}
	case tea.KeySpace:
		if m.cursor < len(m.agentConfig.conflicts) {
			id := m.agentConfig.conflicts[m.cursor]
			m.agentConfig.writeModes[id] = (m.agentConfig.writeModes[id] + 1) % 3
		}
	case tea.KeyEnter:
		m.state = StateReview
		m.cursor = 0
	case tea.KeyEsc:
		m.state = StateEmojiPref
		m.cursor = 0
	}
	return m, nil
}

func (m Model) handleReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyDown:
		if m.cursor == 0 {
			m.cursor = 1 // scroll to Install button
		}
	case tea.KeyUp:
		if m.cursor == 1 {
			m.cursor = 0
		}
	case tea.KeyEnter:
		m.state = StateInstalling
		m.cursor = 0
		m.wizard.installExpected = len(m.selectedAssistants())
		if m.agentConfig.skip {
			m.wizard.agentDone = true
		}
		return m, tea.Batch(m.buildInstallCmd(), m.spinner.Tick)
	case tea.KeyEsc:
		switch {
		case len(m.agentConfig.conflicts) > 0 && !m.agentConfig.skip:
			m.state = StateAgentWriteMode
		case !m.agentConfig.skip:
			m.state = StateEmojiPref
		default:
			m.state = StateAgentSetup
		}
		m.cursor = 0
	}
	return m, nil
}

func (m Model) handleDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.state = StateMainMenu
		m.cursor = 0
		m.wizard.results = nil
		m.wizard.agentResults = nil
		m.wizard.installExpected = 0
		m.wizard.agentDone = false
		m.wizard.selected = make(map[int]bool)
	}
	return m, nil
}

func (m Model) buildInstallCmd() tea.Cmd {
	assistants := m.selectedAssistants()
	provider := installer.Providers[m.wizard.provider]
	installCmd := InstallCmd(assistants, provider, m.skillsFS, languageOptions[m.language.selected].Code)
	if m.agentConfig.skip {
		return installCmd
	}
	preset := installer.PersonaPresets[m.agentConfig.selectedPersona]
	agentCmd := AgentInstallCmd(assistants, preset, m.agentConfig.useEmojis, m.agentConfig.writeModes, m.skillsFS)
	return tea.Batch(installCmd, agentCmd)
}

// AgentWriteModes returns the per-assistant write mode map. Exported for tests.
func (m Model) AgentWriteModes() map[string]installer.AgentWriteMode {
	return m.agentConfig.writeModes
}

// UseEmojis returns the chosen emoji preference. Exported for tests.
func (m Model) UseEmojis() bool {
	return m.agentConfig.useEmojis
}

// LanguageCode returns the currently selected language code. Exported for tests.
func (m Model) LanguageCode() string {
	return languageOptions[m.language.selected].Code
}

// View renders the current state.
func (m Model) View() string {
	return renderState(m)
}

// loadDashboardState reads install metadata for all known assistants.
// Called once when the user navigates to StateDashboard so the render path
// never performs I/O.
func loadDashboardState() dashboardState {
	meta := make(map[installer.AssistantID]installer.InstallMeta, len(installer.Descriptors))
	for _, d := range installer.Descriptors {
		if m, err := installer.ReadInstallMeta(d); err == nil {
			meta[d.ID] = m
		}
	}
	return dashboardState{meta: meta}
}

// allInstallsDone reports whether both the per-assistant installs and the agent
// config step have finished, so Update can transition to StateDone.
func (m Model) allInstallsDone() bool {
	if len(m.wizard.results) < m.wizard.installExpected {
		return false
	}
	return m.wizard.agentDone
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
