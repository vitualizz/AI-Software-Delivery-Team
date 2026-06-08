package panels

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ArtifactPanel renders a file browser for .asdt/artifacts/{change}/ with a
// scrollable YAML viewer for the selected file.
type ArtifactPanel struct {
	files        []string
	selected     int
	content      string // YAML content of selected file
	width        int
	height       int
	viewport     viewport.Model
	header       PanelHeader
	compact      bool
	commonPrefix string
}

// NewArtifactPanel returns a zero-value ArtifactPanel ready for use.
func NewArtifactPanel() ArtifactPanel {
	vp := viewport.New(80, 20)
	return ArtifactPanel{
		viewport: vp,
		header:   NewPanelHeader("Artifacts"),
	}
}

// Init satisfies the tea.Model-compatible interface.
func (p ArtifactPanel) Init() tea.Cmd { return nil }

// Update handles messages directed at the artifact panel.
func (p ArtifactPanel) Update(msg tea.Msg) (ArtifactPanel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.selected < len(p.files)-1 {
				p.selected++
			}
			return p, nil
		case "k", "up":
			if p.selected > 0 {
				p.selected--
			}
			return p, nil
		case "enter":
			if len(p.files) > 0 && p.selected < len(p.files) {
				content, err := loadFileContent(p.files[p.selected])
				if err != nil {
					p.content = fmt.Sprintf("Error loading file: %v", err)
				} else {
					p.content = content
				}
				p.viewport.SetContent(p.content)
				p.viewport.GotoTop()
			}
			return p, nil
		}
	}

	// Delegate scroll keys to viewport.
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

// UpdateSize stores the new dimensions and adjusts the viewport.
func (p ArtifactPanel) UpdateSize(width, height int) (ArtifactPanel, tea.Cmd) {
	p.width = width
	p.height = height
	p.compact = width <= 60
	listHeight := height / 2
	viewportHeight := height - listHeight - 3 // 3 for separators/title
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	p.viewport.Width = width - 2
	p.viewport.Height = viewportHeight
	var cmd tea.Cmd
	p.header, cmd = p.header.UpdateSize(width, height)
	return p, cmd
}

// SetFiles updates the artifact file list and the header's displayed count.
func (p ArtifactPanel) SetFiles(files []string) (ArtifactPanel, tea.Cmd) {
	p.files = files
	if p.selected >= len(files) {
		p.selected = 0
	}
	p.commonPrefix = ""
	if len(files) > 0 {
		p.commonPrefix = commonPathPrefix(files)
	}
	p.header.SetCount(len(files))
	return p, nil
}

// SetContent updates the YAML content displayed in the viewport.
func (p ArtifactPanel) SetContent(content string) (ArtifactPanel, tea.Cmd) {
	p.content = content
	p.viewport.SetContent(highlightYAMLKeys(content))
	return p, nil
}

var (
	styleFileSelected = lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary)
	styleFile         = lipgloss.NewStyle().Foreground(ColorMuted)
	styleSeparator    = lipgloss.NewStyle().Foreground(ColorInactive)
	styleViewerTitle  = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
)

// View renders the artifact browser: file list on top, YAML viewer below.
func (p ArtifactPanel) View() string {
	var sb strings.Builder

	sb.WriteString(p.header.View())
	sb.WriteString("\n")

	if len(p.files) == 0 {
		sb.WriteString(stylePending.Render("No artifacts found"))
		sb.WriteString("\n")
		sb.WriteString(stylePending.Render("Run /asdt \"<feature description>\" to generate artifacts"))
		return sb.String()
	}

	// File list.
	listHeight := p.height / 2
	if listHeight < 1 {
		listHeight = len(p.files)
	}
	start := 0
	if p.selected >= listHeight {
		start = p.selected - listHeight + 1
	}
	end := start + listHeight
	if end > len(p.files) {
		end = len(p.files)
	}

	for i := start; i < end; i++ {
		rel := strings.TrimPrefix(p.files[i], p.commonPrefix)
		rel = strings.TrimLeft(rel, "/")
		rel = strings.TrimSuffix(rel, ".yaml")

		segs := strings.Split(rel, "/")
		name := segs[len(segs)-1]
		depth := len(segs) - 1
		if depth > 4 {
			depth = 4
		}

		indent := strings.Repeat("  ", depth)
		prefix := ""
		if depth > 0 {
			prefix = "\u2514\u2500\u2500 "
		}

		line := indent + prefix + name
		if i == p.selected {
			sb.WriteString(styleFileSelected.Render("\u25b6 " + line))
		} else {
			sb.WriteString(styleFile.Render("  " + line))
		}
		sb.WriteString("\n")
	}

	// Separator.
	sb.WriteString(styleSeparator.Render(strings.Repeat("\u254c", panelMax(p.width-4, 10))))
	sb.WriteString("\n")

	// YAML viewer.
	if p.content == "" {
		sb.WriteString(stylePending.Render("Press enter to view artifact"))
	} else {
		sb.WriteString(styleViewerTitle.Render("Content"))
		sb.WriteString("\n")
		sb.WriteString(p.viewport.View())
	}

	return sb.String()
}

// loadFileContent reads a YAML file and returns its content as a string.
func loadFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

// SetFocused propagates focus to the panel header.
func (p *ArtifactPanel) SetFocused(f bool) { p.header.SetFocused(f) }

// SetSize updates dimensions and compact mode.
func (p *ArtifactPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.compact = w <= 60
	p.header.width = w
}

func panelMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func commonPathPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	prefix := paths[0]
	for _, p := range paths[1:] {
		for !strings.HasPrefix(p, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	if idx := strings.LastIndex(prefix, "/"); idx >= 0 {
		return prefix[:idx+1]
	}
	return prefix
}

func highlightYAMLKeys(content string) string {
	lines := strings.Split(content, "\n")
	keyStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	for i, line := range lines {
		if idx := strings.Index(line, ":"); idx >= 0 {
			key := line[:idx]
			rest := line[idx:]
			lines[i] = keyStyle.Render(key) + rest
		}
	}
	return strings.Join(lines, "\n")
}
