package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const frontmatterDelimiter = "---"

// specialistFrontmatter holds the scalar fields read from a SKILL.md
// frontmatter block needed to render an OpenCode command wrapper.
type specialistFrontmatter struct {
	SpecialistID string
	Name         string
	Description  string
}

func parseSpecialistFrontmatter(skillMD string) (specialistFrontmatter, error) {
	block, ok := extractFrontmatterBlock(skillMD)
	if !ok {
		return specialistFrontmatter{}, fmt.Errorf("no frontmatter block found")
	}

	fields := scanFrontmatterFields(block)

	fm := specialistFrontmatter{
		SpecialistID: fields["specialist-id"],
		Name:         fields["name"],
		Description:  unquoteScalar(fields["description"]),
	}

	if fm.SpecialistID == "" {
		return specialistFrontmatter{}, fmt.Errorf("frontmatter missing required field %q", "specialist-id")
	}
	if fm.Name == "" {
		return specialistFrontmatter{}, fmt.Errorf("frontmatter missing required field %q", "name")
	}
	if fm.Description == "" {
		return specialistFrontmatter{}, fmt.Errorf("frontmatter missing required field %q", "description")
	}

	return fm, nil
}

func extractFrontmatterBlock(skillMD string) (block string, ok bool) {
	lines := strings.Split(skillMD, "\n")

	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == frontmatterDelimiter {
			start = i
			break
		}
		if strings.TrimSpace(line) != "" {
			return "", false
		}
	}
	if start == -1 {
		return "", false
	}

	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelimiter {
			return strings.Join(lines[start+1:i], "\n"), true
		}
	}

	return "", false
}

func scanFrontmatterFields(block string) map[string]string {
	fields := make(map[string]string)

	for _, line := range strings.Split(block, "\n") {
		if line == "" || line[0] == ' ' || line[0] == '\t' || line[0] == '#' {
			continue
		}

		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return fields
}

func unquoteScalar(value string) string {
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}
	return value
}

func renderOpenCodeWrapper(fm specialistFrontmatter) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("description: \"")
	b.WriteString(fm.Description)
	b.WriteString("\"\n")
	b.WriteString("agent: build\n")
	b.WriteString("subtask: false\n")
	b.WriteString("---\n\n")
	b.WriteString("Load and activate the `")
	b.WriteString(fm.Name)
	b.WriteString("` specialist skill (specialist-id: `")
	b.WriteString(fm.SpecialistID)
	b.WriteString("`) and proceed as that specialist for the rest of this session.\n")

	return b.String()
}

func openCodeCommandRoot() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/opencode/commands"
	}
	home, _ := os.UserHomeDir()
	return home + "/.config/opencode/commands"
}

// CommandAdapterDescriptor describes how to generate extra discoverability
// artifacts for one assistant, on top of the shared skill-tree copy.
type CommandAdapterDescriptor struct {
	AssistantID AssistantID
	Generate    func(skillsFS fs.FS, commandRoot string) ([]string, error)
}

func adapterFor(id AssistantID) (CommandAdapterDescriptor, bool) {
	for _, adapter := range CommandAdapters {
		if adapter.AssistantID == id {
			return adapter, true
		}
	}
	return CommandAdapterDescriptor{}, false
}

func generateOpenCodeCommands(skillsFS fs.FS, commandRoot string) ([]string, error) {
	entries, err := fs.ReadDir(skillsFS, ".")
	if err != nil {
		return nil, fmt.Errorf("read root: %w", err)
	}

	if mkErr := os.MkdirAll(commandRoot, 0o755); mkErr != nil {
		return nil, fmt.Errorf("mkdir %s: %w", commandRoot, mkErr)
	}

	var written []string
	var firstErr error

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path, genErr := generateOneOpenCodeCommand(skillsFS, entry.Name(), commandRoot)
		if genErr != nil {
			if firstErr == nil {
				firstErr = genErr
			}
			continue
		}
		if path != "" {
			written = append(written, path)
		}
	}

	return written, firstErr
}

func generateOneOpenCodeCommand(skillsFS fs.FS, dirName, commandRoot string) (string, error) {
	skillPath := dirName + "/SKILL.md"

	data, readErr := fs.ReadFile(skillsFS, skillPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return "", nil
		}
		return "", fmt.Errorf("read %s: %w", skillPath, readErr)
	}

	fm, parseErr := parseSpecialistFrontmatter(string(data))
	if parseErr != nil {
		return "", fmt.Errorf("parse frontmatter %s: %w", skillPath, parseErr)
	}

	// dirName, not "asdt-"+fm.SpecialistID: asdt-init's specialist-id is
	// itself "asdt-init", which would double-prefix to asdt-asdt-init.md.
	target := filepath.Join(commandRoot, dirName+".md")
	content := renderOpenCodeWrapper(fm)

	if writeErr := os.WriteFile(target, []byte(content), 0o644); writeErr != nil {
		return "", fmt.Errorf("write %s: %w", target, writeErr)
	}

	return target, nil
}

// CommandAdapters lists assistants that need extra generated discoverability
// artifacts beyond the shared skill-tree copy. Absence is the no-op — Claude
// Code needs none, so it carries no entry here.
var CommandAdapters = []CommandAdapterDescriptor{
	{
		AssistantID: AssistantOpenCode,
		Generate:    generateOpenCodeCommands,
	},
}
