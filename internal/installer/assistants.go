package installer

import "os"

// AssistantID identifies a known AI assistant.
type AssistantID = string

const (
	AssistantClaudeCode AssistantID = "claude-code"
	AssistantOpenCode   AssistantID = "opencode"
)

// AssistantDescriptor describes a known AI assistant and where its skills are installed.
type AssistantDescriptor struct {
	ID         AssistantID
	Name       string
	BinaryName string
	SkillsDir  string
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
	return home + "/.claude/skills/asdt"
}

func openCodeSkillsDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/opencode/skills/asdt"
	}
	home, _ := os.UserHomeDir()
	return home + "/.config/opencode/skills/asdt"
}
