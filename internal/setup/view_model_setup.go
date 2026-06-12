package setup

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
	"github.com/vitualizz/ai-software-delivery-team/internal/setup/styles"
	"github.com/vitualizz/ai-software-delivery-team/internal/tui/panels"
)

// renderModelGate renders the recommended-vs-customize gate: two radio rows
// following the EmojiPref pattern. Recommended is the default, so users who
// don't care about models pass through with a single Enter.
func renderModelGate(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 4, 6))

	options := []string{s.OptionModelsRecommended, s.OptionModelsCustomize}
	for i, opt := range options {
		focused := i == m.cursor
		selected := i == m.wizard.modelGateChoice

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

	names := make([]string, len(m.wizard.detectedAI))
	for i, p := range m.wizard.detectedAI {
		names[i] = p.Name
	}
	subtitle := styles.Default.Dim.Render(fmt.Sprintf(s.BodyModelSetupSubtitle, strings.Join(names, ", ")))
	body := lipgloss.JoinVertical(lipgloss.Left, subtitle, "", strings.TrimRight(b.String(), "\n"))

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "↑↓", Description: s.HintNavigate},
			{Key: "enter", Description: s.HintContinue},
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleModelGate, body, footer, true)
}

// renderModelSetup renders the per-step model accordion: one collapsible
// group per specialist (all collapsed by default), step rows with ◂ model ▸
// cycling inside expanded groups, and a Claude Code equivalence hint under
// the focused step when its model is not Anthropic-native.
func renderModelSetup(m Model) string {
	s := m.catalog.Installer
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", stepLine(s, 4, 6))

	rows := m.modelRows()
	for i, row := range rows {
		focused := i == m.cursor
		cursor := "  "
		if focused {
			cursor = cursorChar + " "
		}

		if row.header {
			glyph := "▸"
			if m.wizard.expandedGroups[row.specialist] {
				glyph = "▾"
			}
			label := glyph + " " + row.specialist
			summary := styles.Default.Dim.Render(m.groupSummary(row.specialist))
			if focused {
				fmt.Fprintf(&b, "  %s%s %s\n", cursor, styles.Default.Cursor.Render(label), summary)
			} else {
				fmt.Fprintf(&b, "  %s%s %s\n", cursor, styles.Default.Dim.Render(label), summary)
			}
			continue
		}

		step := m.wizard.modelSteps[row.stepIdx]
		model := m.wizard.selectedModels[step.Key()]
		label := modelDisplayLabel(model)

		var stepStr, modelStr string
		if focused {
			stepStr = styles.Default.Cursor.Render(step.Step)
			modelStr = styles.Default.Cursor.Render("◂ " + label + " ▸")
		} else {
			stepStr = styles.Default.Dim.Render(step.Step)
			modelStr = styles.Default.Dim.Render(label)
		}
		fmt.Fprintf(&b, "  %s  %-26s %s\n", cursor, stepStr, modelStr)

		// Equivalence hint: only under the focused step, only when the model
		// maps to a different Anthropic enum on Claude Code.
		if focused {
			if cm, ok := installer.FindModel(model); ok && cm.ClaudeCodeEnum != cm.ID {
				fmt.Fprintf(&b, "        %s\n", styles.Default.Dim.Render(fmt.Sprintf(s.HintRunsAs, cm.ClaudeCodeEnum)))
			}
		}
	}

	b.WriteString("\n")
	if m.cursor == len(rows) {
		fmt.Fprintf(&b, "  %s %s\n", cursorChar, styles.Default.Cursor.Render(s.BtnContinue))
	} else {
		fmt.Fprintf(&b, "      %s\n", styles.Default.Dim.Render(s.BtnContinue))
	}

	footer := panels.RenderKeyboardFooter([]panels.HintGroup{
		{Label: s.HintGroupNav, Hints: []panels.Hint{
			{Key: "↑/↓", Description: s.HintNavigate},
			{Key: "space", Description: s.HintExpand},
			{Key: "←/→", Description: s.HintCycleModel},
			{Key: "r", Description: s.HintResetDefault},
		}},
		{Label: s.HintGroupActions, Hints: []panels.Hint{
			{Key: "enter", Description: s.HintContinue},
			{Key: "esc", Description: s.HintBack},
			{Key: "q", Description: s.HintQuit},
		}},
	}, m.width)
	return frame(s.TitleModelSetup, strings.TrimRight(b.String(), "\n"), footer, true)
}

// groupSummary aggregates a specialist's current model selections into a
// short parenthetical like "— 5 steps (haiku ×4, sonnet ×1)", giving the
// collapsed header enough signal to decide whether expanding is worth it.
func (m Model) groupSummary(specialist string) string {
	counts := make(map[string]int)
	var order []string
	total := 0
	for _, s := range m.wizard.modelSteps {
		if s.Specialist != specialist {
			continue
		}
		total++
		label := modelDisplayLabel(m.wizard.selectedModels[s.Key()])
		if counts[label] == 0 {
			order = append(order, label)
		}
		counts[label]++
	}
	sort.SliceStable(order, func(i, j int) bool { return counts[order[i]] > counts[order[j]] })

	parts := make([]string, len(order))
	for i, label := range order {
		parts[i] = fmt.Sprintf("%s ×%d", label, counts[label])
	}
	return fmt.Sprintf("— %d steps (%s)", total, strings.Join(parts, ", "))
}

// modelDisplayLabel returns the model's ID — the exact value written to
// workflow.yaml — which doubles as the display label. An empty selection
// (step with no source default, not yet cycled) renders as "default".
func modelDisplayLabel(id string) string {
	if id == "" {
		return "default"
	}
	return id
}
