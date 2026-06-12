package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/i18n"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/styles"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

const cursorChar = "►"

// stepLine returns a dim "step X of N" indicator using the active catalog.
func stepLine(s i18n.InstallerStrings, current, total int) string {
	return styles.Default.Dim.Render(fmt.Sprintf("%s %d %s %d", s.StepWord, current, s.StepOfWord, total))
}

// modeLabel returns the catalog-translated label for an AgentWriteMode.
func modeLabel(s i18n.InstallerStrings, mode installer.AgentWriteMode) string {
	switch mode {
	case installer.AgentModeOverwrite:
		return s.LabelModeOverwrite
	case installer.AgentModeAppend:
		return s.LabelModeAppend
	default:
		return s.LabelModeKeep
	}
}

func renderState(m Model) string {
	switch m.state {
	case StateEnvironmentCheck:
		return renderPreflightCheck(m)
	case StateDashboard:
		return renderDashboard(m)
	case StateMainMenu:
		return renderMainMenu(m)
	case StateLanguageSelect:
		return renderLanguageSelect(m)
	case StateReview:
		return renderReview(m)
	case StateSelectAssistants:
		return renderSelectAssistants(m)
	case StateSelectProvider:
		return renderSelectProvider(m)
	case StateModelGate:
		return renderModelGate(m)
	case StateModelSetup:
		return renderModelSetup(m)
	case StateAgentSetup:
		return renderAgentSetup(m)
	case StateEmojiPref:
		return renderEmojiPref(m)
	case StateAgentWriteMode:
		return renderAgentWriteMode(m)
	case StateInstalling:
		return renderInstalling(m)
	case StateDone:
		return renderDone(m)
	default:
		return ""
	}
}

// titleStyle returns an inline bold style with the primary panel color.
func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(panels.ColorPrimary)
}

// frame composes a screen from a title, a body block, and a footer hint,
// all rendered inside a single bordered box. The footer sits as the last
// line inside the box (Gentle-AI pattern), keeping the layout as one
// visual unit instead of stacking a separate bar below.
func frame(title, body, footer string, focused bool) string {
	header := titleStyle().Render("▎ " + title)
	var inner string
	if footer != "" {
		inner = lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
	} else {
		inner = lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	}
	return panels.FocusBorderStyle(focused).
		Padding(1, 2).
		Render(inner)
}

func renderMainMenu(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder

	items := []string{s.MenuInstall, s.MenuDashboard, s.MenuQuit}
	for i, item := range items {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(item))
		} else {
			fmt.Fprintf(&b, "    %s\n", styles.Default.Dim.Render(item))
		}
	}

	if m.updateAvailable {
		fmt.Fprintf(&b, "\n%s\n",
			styles.Default.Cursor.Render(
				fmt.Sprintf("↑ asdt-tui %s available — https://github.com/vitualizz/ai-software-delivery-team/releases", m.latestVersion),
			))
	}

	hero := panels.RenderHero(m.currentVersion)
	menuStr := strings.TrimRight(b.String(), "\n")
	body := lipgloss.JoinVertical(lipgloss.Left, hero, "", menuStr)

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupNav, Hints: []panels.Hint{
			{Key: "↑↓", Description: s.HintNavigate},
			{Key: "enter", Description: s.HintSelect},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleMainMenu, body, footer, true)
}

// renderLanguageSelect renders the language selection screen: one radio row
// per languageOptions entry, following renderEmojiPref's pattern but WITHOUT
// a step indicator — like MainMenu, it sits before the numbered wizard steps.
// Option labels are the languages' native names, deliberately not catalog
// strings.
func renderLanguageSelect(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder

	for i, opt := range languageOptions {
		focused := i == m.cursor
		selected := i == m.language.selected

		cursor := "  "
		if focused {
			cursor = cursorChar + " "
		}

		var radioStr string
		if selected {
			radioStr = styles.Default.Cursor.Render("(•)")
		} else {
			radioStr = styles.Default.Dim.Render("( )")
		}

		var nameStr string
		if focused {
			nameStr = styles.Default.Cursor.Render(opt.Label)
		} else {
			nameStr = styles.Default.Dim.Render(opt.Label)
		}

		fmt.Fprintf(&b, "  %s%s %s\n", cursor, radioStr, nameStr)
	}

	subtitle := styles.Default.Dim.Render(s.BodyLanguageSelectSubtitle)
	body := lipgloss.JoinVertical(lipgloss.Left, subtitle, "", strings.TrimRight(b.String(), "\n"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "↑↓", Description: s.HintNavigate},
			{Key: "enter", Description: s.HintContinue},
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleLanguageSelect, body, footer, true)
}

func renderSelectAssistants(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 2, 6))

	for i, d := range installer.Descriptors {
		bp, sp, _ := installer.Detect(d)
		present := bp && sp

		var badge panels.Badge
		if present {
			badge = panels.NewBadge("present", panels.ColorSuccess)
		} else {
			badge = panels.NewBadge("missing", panels.ColorError)
		}

		check := "[ ]"
		if m.wizard.selected[i] {
			check = "[x]"
		}

		cursor := "  "
		var nameStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(d.Name)
		} else {
			nameStr = styles.Default.Dim.Render(d.Name)
		}

		fmt.Fprintf(&b, "  %s%s %s %s\n", cursor, check, nameStr, badge.Render())
	}

	b.WriteString("\n")
	if m.cursor == len(installer.Descriptors) {
		fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(s.BtnContinue))
	} else {
		fmt.Fprintf(&b, "      %s\n", styles.Default.Dim.Render(s.BtnContinue))
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "space", Description: s.HintToggle},
			{Key: "a", Description: s.HintAllNone},
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleSelectAssistants, strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderSelectProvider(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 3, 6))

	for i, p := range installer.Providers {
		focused := i == m.cursor
		selected := i == m.wizard.provider

		cursor := "  "
		if focused {
			cursor = cursorChar + " "
		}

		var radioStr string
		if selected {
			radioStr = styles.Default.Cursor.Render("(•)")
		} else {
			radioStr = styles.Default.Dim.Render("( )")
		}

		var nameStr string
		if focused {
			nameStr = styles.Default.Cursor.Render(p.Name)
		} else {
			nameStr = styles.Default.Dim.Render(p.Name)
		}

		fmt.Fprintf(&b, "  %s%s %s — %s\n", cursor, radioStr, nameStr, p.Description)
	}

	b.WriteString("\n")
	if m.cursor == len(installer.Providers) {
		fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(s.BtnContinue))
	} else {
		fmt.Fprintf(&b, "      %s\n", styles.Default.Dim.Render(s.BtnContinue))
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleSelectProvider, strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderAgentSetup(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 5, 6))

	if len(m.agentConfig.conflicts) > 0 {
		fmt.Fprintf(&b, "  %s\n",
			styles.Default.Warning.Render(s.WarnExistingConfig))
		fmt.Fprintf(&b, "  %s\n\n",
			styles.Default.Dim.Render(s.WarnExistingConfigNote))
	} else {
		fmt.Fprintf(&b, "  %s\n\n",
			styles.Default.Dim.Render(s.InfoWriteGlobal))
	}

	for i, p := range installer.PersonaPresets {
		focused := i == m.cursor
		selected := !m.agentConfig.skip && i == m.agentConfig.selectedPersona

		cursor := "  "
		if focused {
			cursor = cursorChar + " "
		}

		var radioStr string
		if selected {
			radioStr = styles.Default.Cursor.Render("(•)")
		} else {
			radioStr = styles.Default.Dim.Render("( )")
		}

		desc := m.catalog.PersonaDescription(p.ID)
		if desc == "" {
			desc = p.Description
		}

		var nameStr, descStr string
		if focused {
			nameStr = styles.Default.Cursor.Render(p.Name)
			descStr = desc
		} else {
			nameStr = styles.Default.Dim.Render(p.Name)
			descStr = styles.Default.Dim.Render(desc)
		}
		fmt.Fprintf(&b, "  %s%s %s — %s\n", cursor, radioStr, nameStr, descStr)
	}

	b.WriteString("\n")
	if m.cursor == len(installer.PersonaPresets) {
		fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(s.BtnSkip))
	} else {
		fmt.Fprintf(&b, "      %s\n", styles.Default.Dim.Render(s.BtnSkip))
	}

	subtitle := styles.Default.Dim.Render(s.BodyAgentSetupSubtitle)
	body := lipgloss.JoinVertical(lipgloss.Left, subtitle, "", strings.TrimRight(b.String(), "\n"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "↑↓", Description: s.HintNavigate},
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleAgentSetup, body, footer, true)
}

// renderEmojiPref renders the emoji preference screen: two radio rows
// (Yes / No) following renderSelectProvider's pattern. It shares step 4 with
// the persona screen — both belong to the same agent-config wizard step.
func renderEmojiPref(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 5, 6))

	options := []string{s.OptionEmojiYes, s.OptionEmojiNo}
	for i, opt := range options {
		focused := i == m.cursor
		selected := (i == 0) == m.agentConfig.useEmojis

		cursor := "  "
		if focused {
			cursor = cursorChar + " "
		}

		var radioStr string
		if selected {
			radioStr = styles.Default.Cursor.Render("(•)")
		} else {
			radioStr = styles.Default.Dim.Render("( )")
		}

		var nameStr string
		if focused {
			nameStr = styles.Default.Cursor.Render(opt)
		} else {
			nameStr = styles.Default.Dim.Render(opt)
		}

		fmt.Fprintf(&b, "  %s%s %s\n", cursor, radioStr, nameStr)
	}

	subtitle := styles.Default.Dim.Render(s.BodyEmojiPrefSubtitle)
	body := lipgloss.JoinVertical(lipgloss.Left, subtitle, "", strings.TrimRight(b.String(), "\n"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "↑↓", Description: s.HintNavigate},
			{Key: "enter", Description: s.HintContinue},
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleEmojiPref, body, footer, true)
}

func renderAgentWriteMode(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 5, 6))
	fmt.Fprintf(&b, "  %s\n\n", styles.Default.Dim.Render(s.BodyAgentWriteMode))

	nameFor := make(map[string]string, len(installer.Descriptors))
	for _, d := range installer.Descriptors {
		nameFor[d.ID] = d.Name
	}

	for i, id := range m.agentConfig.conflicts {
		name := nameFor[id]
		if name == "" {
			name = id
		}
		mode := m.agentConfig.writeModes[id]
		label := "[" + modeLabel(s, mode) + "]"

		adapter, hasAdapter := installer.AgentConfigAdapterFor(installer.AssistantID(id))
		pathStr := ""
		if hasAdapter {
			pathStr = "  " + styles.Default.Dim.Render(adapter.ConfigPath())
		}

		cursor := "  "
		var nameStr, modeStr string
		if i == m.cursor {
			cursor = cursorChar + " "
			nameStr = styles.Default.Cursor.Render(name)
			modeStr = styles.Default.Cursor.Render(label)
		} else {
			nameStr = styles.Default.Dim.Render(name)
			modeStr = styles.Default.Dim.Render(label)
		}

		fmt.Fprintf(&b, "  %s%s  %s%s\n", cursor, nameStr, modeStr, pathStr)
	}

	b.WriteString("\n")
	if m.cursor == len(m.agentConfig.conflicts) {
		fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(s.BtnContinue))
	} else {
		fmt.Fprintf(&b, "      %s\n", styles.Default.Dim.Render(s.BtnContinue))
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "↑↓", Description: s.HintNavigate},
			{Key: "space", Description: s.HintCycleMode},
			{Key: "esc", Description: s.HintBack},
		}},
	}, m.width)
	return frame(s.TitleAgentWriteMode, strings.TrimRight(b.String(), "\n"), footer, true)
}

// presetShortName returns the leading name of a preset option label — the
// portion before the " — " descriptor — so the Review row reads, e.g.,
// "Strategist preset" rather than echoing the full radio sentence.
func presetShortName(label string) string {
	if idx := strings.Index(label, " — "); idx >= 0 {
		return label[:idx]
	}
	return label
}

// modelsReviewValue formats the Models row for the Review screen. The customize
// choice reports the count of customized steps; every preset reports its short
// name. Chameleon (and Craftsman) carry no per-step customizations, so they are
// described by their preset name rather than a count.
func modelsReviewValue(m Model) string {
	s := m.catalog.Installer
	if m.wizard.modelGateChoice == modelGateCustomizeChoice {
		if n := m.countCustomizedModels(); n > 0 {
			return fmt.Sprintf(s.ValueModelsCustomized, n)
		}
		return fmt.Sprintf(s.ValueModelsCustomized, 0)
	}
	label := presetShortName(presetOptionLabels(s)[m.wizard.modelGateChoice])
	return fmt.Sprintf(s.ValueModelsPreset, label)
}

func renderReview(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 6, 6))

	names := make([]string, 0)
	for i, d := range installer.Descriptors {
		if m.wizard.selected[i] {
			names = append(names, d.Name)
		}
	}
	if len(names) == 0 {
		for _, d := range installer.Descriptors {
			names = append(names, d.Name)
		}
	}
	fmt.Fprintf(&b, "  %s  %s\n", styles.Default.Dim.Render(s.LabelAssistants), strings.Join(names, " · "))

	if m.wizard.provider < len(installer.Providers) {
		fmt.Fprintf(&b, "  %s  %s\n", styles.Default.Dim.Render(s.LabelProvider), installer.Providers[m.wizard.provider].Name)
	}

	modelsVal := modelsReviewValue(m)
	fmt.Fprintf(&b, "  %s  %s\n", styles.Default.Dim.Render(s.LabelModels), modelsVal)

	if m.agentConfig.skip {
		fmt.Fprintf(&b, "  %s  %s\n", styles.Default.Dim.Render(s.LabelPersona), styles.Default.Dim.Render(s.LabelSkipped))
	} else if m.agentConfig.selectedPersona < len(installer.PersonaPresets) {
		preset := installer.PersonaPresets[m.agentConfig.selectedPersona]
		desc := m.catalog.PersonaDescription(preset.ID)
		if desc == "" {
			desc = preset.Description
		}
		fmt.Fprintf(&b, "  %s  %s\n", styles.Default.Dim.Render(s.LabelPersona), preset.Name+" — "+desc)
	}

	if !m.agentConfig.skip {
		emojiVal := s.OptionEmojiNo
		if m.agentConfig.useEmojis {
			emojiVal = s.OptionEmojiYes
		}
		fmt.Fprintf(&b, "  %s  %s\n", styles.Default.Dim.Render(s.LabelEmojis), emojiVal)
	}

	if !m.agentConfig.skip && len(m.agentConfig.conflicts) > 0 {
		nameFor := make(map[string]string, len(installer.Descriptors))
		for _, d := range installer.Descriptors {
			nameFor[d.ID] = d.Name
		}
		fmt.Fprintf(&b, "\n  %s\n", styles.Default.Dim.Render(s.LabelConfig))
		for _, id := range m.agentConfig.conflicts {
			name := nameFor[id]
			if name == "" {
				name = id
			}
			mode := m.agentConfig.writeModes[id]
			adapter, ok := installer.AgentConfigAdapterFor(installer.AssistantID(id))
			pathStr := ""
			if ok {
				pathStr = adapter.ConfigPath() + "  "
			}
			fmt.Fprintf(
				&b, "    %s  %s%s\n",
				styles.Default.Dim.Render(name),
				styles.Default.Dim.Render(pathStr),
				styles.Default.Cursor.Render("["+modeLabel(s, mode)+"]"),
			)
		}
	}

	b.WriteString("\n")
	if m.cursor == 1 {
		fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(s.BtnInstall))
	} else {
		fmt.Fprintf(&b, "      %s\n", styles.Default.Dim.Render(s.BtnInstall))
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleReview, strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderInstalling(m Model) string {
	s := m.catalog.Installer

	// Build a lookup for completed results keyed by AssistantID.
	doneFor := make(map[installer.AssistantID]installer.InstallResult, len(m.wizard.results))
	for _, r := range m.wizard.results {
		doneFor[r.AssistantID] = r
	}

	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 6, 6))

	spinnerFrame := m.spinner.View()
	for i, d := range installer.Descriptors {
		if !m.wizard.selected[i] && len(m.wizard.selected) > 0 {
			continue
		}
		if r, done := doneFor[d.ID]; done {
			if r.Err != nil {
				fmt.Fprintf(&b, "  %s\n", styles.Default.Error.Render(fmt.Sprintf("✗ %s: %v", d.Name, r.Err)))
			} else {
				fmt.Fprintf(&b, "  %s\n", styles.Default.Success.Render("✓ "+d.Name))
			}
		} else {
			fmt.Fprintf(&b, "  %s %s\n", spinnerFrame, styles.Default.Dim.Render(d.Name))
		}
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{{Key: "q", Description: s.HintQuit}}},
	}, m.width)
	return frame(s.TitleInstalling, strings.TrimRight(b.String(), "\n"), footer, true)
}

func renderDashboard(m Model) string {
	s := m.catalog.Installer
	d := m.catalog.Dashboard
	var b strings.Builder
	for _, desc := range installer.Descriptors {
		meta := m.dashboard.meta[desc.ID]
		b.WriteString(renderSummaryLine(desc, meta, d))
		b.WriteString("\n")
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupNav, Hints: []panels.Hint{{Key: "esc", Description: s.HintBackToMenu}}},
	}, m.width)
	return frame(s.TitleDashboard, strings.TrimRight(b.String(), "\n"), footer, true)
}

// renderSummaryLine renders a dashboard row for one assistant.
// First line: name + binary/skills badges.
// Second line (when install metadata exists): install date and persona.
func renderSummaryLine(d installer.AssistantDescriptor, meta installer.InstallMeta, s i18n.DashboardStrings) string {
	bp, sp, _ := installer.Detect(d)

	var parts []string
	parts = append(parts, d.Name)
	if bp {
		parts = append(parts, panels.NewBadge("binary", panels.ColorSuccess).Render())
	} else {
		parts = append(parts, panels.NewBadge("binary", panels.ColorError).Render())
	}
	if sp {
		parts = append(parts, panels.NewBadge("skills", panels.ColorSuccess).Render())
	} else {
		parts = append(parts, panels.NewBadge("skills", panels.ColorError).Render())
	}
	line := "  " + strings.Join(parts, " ")

	if meta.InstalledAt.IsZero() {
		return line
	}

	var metaParts []string
	metaParts = append(metaParts, s.LabelInstalled+": "+meta.InstalledAt.Format("2006-01-02"))
	if meta.Persona != "" {
		metaParts = append(metaParts, s.LabelPersona+": "+meta.Persona)
	}
	metaLine := "    " + styles.Default.Dim.Render(strings.Join(metaParts, "  ·  "))
	return line + "\n" + metaLine
}

func renderDone(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder

	nameFor := make(map[installer.AssistantID]string, len(installer.Descriptors))
	for _, d := range installer.Descriptors {
		nameFor[d.ID] = d.Name
	}

	for _, r := range m.wizard.results {
		name := nameFor[r.AssistantID]
		if name == "" {
			name = string(r.AssistantID)
		}
		if r.Err == nil {
			fmt.Fprintf(&b, "  %s\n", styles.Default.Success.Render("✓ "+name))
			if len(r.Removed) > 0 {
				fmt.Fprintf(&b, "    %s\n", styles.Default.Dim.Render(fmt.Sprintf(s.MsgStaleRemoved, len(r.Removed))))
			}
		} else {
			fmt.Fprintf(&b, "  %s\n", styles.Default.Error.Render(fmt.Sprintf("✗ %s: %v", name, r.Err)))
		}
	}

	if len(m.wizard.agentResults) > 0 {
		fmt.Fprintf(&b, "\n  %s\n", styles.Default.Dim.Render(s.LabelAgentConfig))
		for _, r := range m.wizard.agentResults {
			name := nameFor[r.AssistantID]
			if name == "" {
				name = string(r.AssistantID)
			}
			switch {
			case r.Err != nil:
				fmt.Fprintf(&b, "  %s\n", styles.Default.Error.Render(fmt.Sprintf("✗ %s: %v", name, r.Err)))
			case r.Skipped:
				fmt.Fprintf(&b, "  %s\n", styles.Default.Dim.Render(fmt.Sprintf(s.MsgAgentConfigSkipped, name)))
			default:
				fmt.Fprintf(&b, "  %s\n", styles.Default.Success.Render(fmt.Sprintf(s.MsgAgentConfigWritten, name)))
			}
			if r.AssistantID == installer.AssistantOpenCode && r.Err == nil && !r.Skipped {
				fmt.Fprintf(&b, "  %s\n", styles.Default.Dim.Render(s.MsgOpenCodeNote))
			}
		}
	}

	fmt.Fprintf(&b, "\n  %s\n", styles.Default.Dim.Render(s.MsgGetStarted))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "enter/esc", Description: s.HintBackToMenu},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleDone, strings.TrimRight(b.String(), "\n"), footer, true)
}
