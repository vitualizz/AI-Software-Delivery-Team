package panels

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ArtifactPanel renders a file browser for .asdt/artifacts/{change}/ with a
// scrollable YAML viewer for the selected file.
type ArtifactPanel struct {
	files    []string
	selected int
	content  string // YAML content of selected file
	width    int
	height   int
	viewport viewport.Model
}

// NewArtifactPanel returns a zero-value ArtifactPanel ready for use.
func NewArtifactPanel() ArtifactPanel {
	vp := viewport.New(80, 20)
	return ArtifactPanel{viewport: vp}
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
	listHeight := height / 2
	viewportHeight := height - listHeight - 3 // 3 for separators/title
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	p.viewport.Width = width - 2
	p.viewport.Height = viewportHeight
	return p, nil
}

// SetFiles updates the artifact file list.
func (p ArtifactPanel) SetFiles(files []string) (ArtifactPanel, tea.Cmd) {
	p.files = files
	if p.selected >= len(files) {
		p.selected = 0
	}
	return p, nil
}

// SetContent updates the YAML content displayed in the viewport.
func (p ArtifactPanel) SetContent(content string) (ArtifactPanel, tea.Cmd) {
	p.content = content
	p.viewport.SetContent(content)
	return p, nil
}

var (
	styleFileSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleFile         = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	styleSeparator    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleViewerTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
)

// View renders the artifact browser: file list on top, YAML viewer below.
func (p ArtifactPanel) View() string {
	var sb strings.Builder

	sb.WriteString(stylePanelTitle.Render("Artifacts"))
	sb.WriteString("\n")

	if len(p.files) == 0 {
		sb.WriteString(stylePending.Render("No artifacts found"))
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
		name := filepath.Base(p.files[i])
		name = strings.TrimSuffix(name, ".yaml")
		if i == p.selected {
			sb.WriteString(styleFileSelected.Render(fmt.Sprintf("▶ %s", name)))
		} else {
			sb.WriteString(styleFile.Render(fmt.Sprintf("  %s", name)))
		}
		sb.WriteString("\n")
	}

	// Separator.
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", panelMax(p.width-4, 10))))
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

func panelMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
