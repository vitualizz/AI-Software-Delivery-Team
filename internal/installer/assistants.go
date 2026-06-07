// Package installer provides types and logic for installing ASDT skills into AI assistants.
package installer

import "os"

// AssistantID identifies a known AI assistant.
type AssistantID = string

const (
	// AssistantClaudeCode identifies the Claude Code assistant.
	AssistantClaudeCode AssistantID = "claude-code"
	// AssistantOpenCode identifies the OpenCode assistant.
	AssistantOpenCode AssistantID = "opencode"
)

// AssistantDescriptor describes a known AI assistant and where its skills are installed.
type AssistantDescriptor struct {
	ID         AssistantID
	Name       string
	BinaryName string
	// SkillsDir is the assistant's skills ROOT directory (e.g. ~/.claude/skills).
	// The installer maps each top-level entry of the embedded skill tree to its
	// own sibling directory directly under this root — it does NOT nest the
	// whole tree under a single "asdt" subdirectory.
	SkillsDir string
}

// Descriptors lists all known AI assistants.
var Descriptors = []AssistantDescriptor{
	{
		ID:         AssistantClaudeCode,
		Name:       "Claude Code",
		BinaryName: "claude",
		SkillsDir:  claudeSkillsDir(),
	},
	{
		ID:         AssistantOpenCode,
		Name:       "OpenCode",
		BinaryName: "opencode",
		SkillsDir:  openCodeSkillsDir(),
	},
}

func claudeSkillsDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.claude/skills"
}

func openCodeSkillsDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/opencode/skills"
	}
	home, _ := os.UserHomeDir()
	return home + "/.config/opencode/skills"
}
