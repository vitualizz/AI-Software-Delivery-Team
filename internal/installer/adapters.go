// Package installer provides types and logic for installing ASDT skills into AI assistants.
//
// adapters.go implements the CommandAdapter abstraction: a per-assistant,
// additive registry of generators that produce extra discoverability
// artifacts (e.g. OpenCode command-palette wrappers) on top of the
// shared skill-tree copy that installOne already performs. Assistants
// absent from CommandAdapters get no extra artifacts — absence IS the
// no-op (see CommandAdapters doc comment).
package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// frontmatterDelimiter marks the start and end of a SKILL.md YAML
// frontmatter block.
const frontmatterDelimiter = "---"

// specialistFrontmatter holds the flat scalar fields read from a
// specialist's SKILL.md frontmatter block that the OpenCode wrapper
// needs to render itself. It is produced exclusively by
// parseSpecialistFrontmatter — never constructed with partial data.
type specialistFrontmatter struct {
	// SpecialistID is the SKILL.md "specialist-id" value (e.g. "developer").
	SpecialistID string
	// Name is the SKILL.md "name" value (e.g. "asdt:developer").
	Name string
	// Description is the SKILL.md "description" value, verbatim, with
	// surrounding double-quotes stripped if present (e.g. the specialist's
	// own one-line picker description).
	Description string
}

// parseSpecialistFrontmatter isolates the leading "---" ... "---" block of
// skillMD and reads the flat scalar keys specialist-id, name, and
// description. All three are required and must be non-empty; a missing or
// empty field yields an error naming that field — the caller must never
// proceed with a zero-value specialistFrontmatter.
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

// extractFrontmatterBlock returns the lines strictly between the first two
// "---" delimiter lines, joined by newlines. ok is false when skillMD does
// not open with a "---" delimiter line or never closes one.
func extractFrontmatterBlock(skillMD string) (block string, ok bool) {
	lines := strings.Split(skillMD, "\n")

	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == frontmatterDelimiter {
			start = i
			break
		}
		if strings.TrimSpace(line) != "" {
			// Frontmatter must be the very first non-blank content.
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

// scanFrontmatterFields reads top-level "key: value" scalar lines from a
// frontmatter block into a map. Nested/list values (lines indented further
// than the key) are ignored — only flat scalars at column 0 are recognized,
// which is exactly what specialist-id/name/description are.
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

// unquoteScalar strips a single pair of surrounding double-quotes from a
// YAML scalar value, leaving the inner text (including embedded
// punctuation such as em-dashes and apostrophes) untouched. Values that
// are not quoted are returned as-is.
func unquoteScalar(value string) string {
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}
	return value
}

// renderOpenCodeWrapper renders the full content of an OpenCode
// command-palette wrapper for the given specialist frontmatter. It is a
// PURE function — the same fm always renders to the byte-identical
// string — which is the sole idempotency primitive for the generated
// wrapper files (AC#3). It must never read the clock, the environment,
// or any non-deterministic source.
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

// openCodeCommandRoot returns OpenCode's command-palette directory: a
// sibling of its skills directory. CONFIRMED against a live OpenCode
// install (only "commands", plural, exists under ~/.config/opencode/) —
// this is the verified contract, not a placeholder.
func openCodeCommandRoot() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/opencode/commands"
	}
	home, _ := os.UserHomeDir()
	return home + "/.config/opencode/commands"
}

// CommandAdapterDescriptor describes how to generate extra
// discoverability artifacts (beyond the shared skill-tree copy) for one
// assistant. Generate must be non-nil — exactly like
// ProviderDescriptor.CustomizeSkill.
type CommandAdapterDescriptor struct {
	AssistantID AssistantID
	// Generate produces discoverability artifacts for every specialist
	// found in skillsFS, writes them under commandRoot, and returns the
	// absolute paths written. A non-nil error indicates a PARTIAL failure
	// (e.g. one specialist's malformed frontmatter) — callers must still
	// use any returned paths rather than discard them.
	Generate func(skillsFS fs.FS, commandRoot string) ([]string, error)
}

// adapterFor performs a generic linear lookup of id in CommandAdapters.
// It returns (zero value, false) when id is absent — which is how
// Claude Code (intentionally not registered) yields no extra artifacts.
func adapterFor(id AssistantID) (CommandAdapterDescriptor, bool) {
	for _, adapter := range CommandAdapters {
		if adapter.AssistantID == id {
			return adapter, true
		}
	}
	return CommandAdapterDescriptor{}, false
}

// generateOpenCodeCommands walks the top-level entries of skillsFS,
// and for every directory containing a SKILL.md, parses its frontmatter,
// renders a deterministic OpenCode command-palette wrapper, and writes
// it to {commandRoot}/{dirName}.md, where dirName is the specialist's
// own sibling install directory name (e.g. "asdt-developer",
// "asdt-init") — the correct on-disk address even for the one
// specialist (asdt-init) whose specialist-id is already prefixed.
//
// Partial success (AC#5): a malformed specialist's frontmatter error is
// recorded but does NOT abort the loop — every well-formed specialist
// still gets its wrapper. The first error encountered is returned
// alongside the paths successfully written so far.
//
// Idempotent (AC#3): the destination path is determined solely by
// dirName and renderOpenCodeWrapper is pure, so re-running this
// over the same skillsFS truncate-overwrites each wrapper with
// byte-identical content — no stale or duplicate files within a run.
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

// generateOneOpenCodeCommand handles a single top-level specialist
// directory. It returns ("", nil) — not an error — when the directory
// has no SKILL.md, since not every top-level entry is a specialist
// (e.g. asdt-shared has no command-worthy SKILL.md of its own).
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

	// Name the wrapper after the specialist's own sibling install directory
	// (dirName), NOT "asdt-"+specialist-id: most specialists' specialist-id
	// is the bare role name ("developer" -> dirName "asdt-developer", and
	// concatenation happens to agree), but asdt-init's specialist-id is
	// itself "asdt-init" — concatenating would yield "asdt-asdt-init.md",
	// a double-prefixed name that doesn't match the directory OpenCode
	// users actually see on disk. dirName is always the correct address.
	target := filepath.Join(commandRoot, dirName+".md")
	content := renderOpenCodeWrapper(fm)

	if writeErr := os.WriteFile(target, []byte(content), 0o644); writeErr != nil {
		return "", fmt.Errorf("write %s: %w", target, writeErr)
	}

	return target, nil
}

// CommandAdapters lists the assistants that need EXTRA generated
// discoverability artifacts beyond the shared skill-tree copy.
//
// Only assistants that need something extra are listed here — absence
// IS the no-op. Claude Code is intentionally ABSENT: it discovers
// specialists directly from the copied skill tree and needs no
// generated command-palette wrappers, so adding a NullCommandAdapter
// placeholder here would exist solely to do nothing, contradicting
// this codebase's flat-registry idiom (see Providers, which likewise
// carries no placeholder for "no provider").
//
// New assistant support = one new appended entry here, zero edits to
// Install/installOne/copyEntry/writeSkillFile or any existing entry.
var CommandAdapters = []CommandAdapterDescriptor{
	{
		AssistantID: AssistantOpenCode,
		Generate:    generateOpenCodeCommands,
	},
}
