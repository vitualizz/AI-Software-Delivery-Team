package installer

import (
	"os"
	"strings"
)

// claudeAgentDir returns the directory where Claude Code's global agent config lives (~/.claude).
func claudeAgentDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.claude"
}

// openCodeConfigDir returns the OpenCode config directory, respecting XDG_CONFIG_HOME.
func openCodeConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/opencode"
	}
	home, _ := os.UserHomeDir()
	return home + "/.config/opencode"
}

// writeClaudeAgentConfig writes rendered AGENTS.md to ~/.claude/AGENTS.md and
// idempotently ensures @AGENTS.md appears in ~/.claude/CLAUDE.md.
func writeClaudeAgentConfig(rendered string, overwrite bool) (AgentConfigResult, error) {
	dir := claudeAgentDir()
	agentsPath := dir + "/AGENTS.md"
	claudePath := dir + "/CLAUDE.md"

	result := AgentConfigResult{AssistantID: AssistantClaudeCode}

	// Check if AGENTS.md already exists.
	if _, err := os.Stat(agentsPath); err == nil && !overwrite {
		result.Skipped = true
		return result, nil
	}

	// Ensure target directory exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return result, err
	}

	// Write AGENTS.md.
	if err := os.WriteFile(agentsPath, []byte(rendered), 0o644); err != nil {
		return result, err
	}
	result.Written = append(result.Written, agentsPath)

	// Idempotent injection: append @AGENTS.md to CLAUDE.md if not already present.
	existing, err := os.ReadFile(claudePath)
	if err != nil && !os.IsNotExist(err) {
		return result, err
	}

	if os.IsNotExist(err) {
		// CLAUDE.md does not exist — create it.
		if writeErr := os.WriteFile(claudePath, []byte("@AGENTS.md"), 0o644); writeErr != nil {
			return result, writeErr
		}
		result.Written = append(result.Written, claudePath)
	} else {
		// CLAUDE.md exists — append only if @AGENTS.md line is absent.
		if !containsAgentsRef(string(existing)) {
			appended := string(existing) + "\n@AGENTS.md"
			if writeErr := os.WriteFile(claudePath, []byte(appended), 0o644); writeErr != nil {
				return result, writeErr
			}
			result.Written = append(result.Written, claudePath)
		}
	}

	return result, nil
}

// containsAgentsRef reports whether the CLAUDE.md content already contains an
// @AGENTS.md reference line (line-equality check, trimming whitespace).
func containsAgentsRef(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "@AGENTS.md" {
			return true
		}
	}
	return false
}

// writeOpenCodeAgentConfig writes rendered AGENTS.md to the OpenCode config directory.
func writeOpenCodeAgentConfig(rendered string, overwrite bool) (AgentConfigResult, error) {
	dir := openCodeConfigDir()
	agentsPath := dir + "/AGENTS.md"

	result := AgentConfigResult{AssistantID: AssistantOpenCode}

	// Check if AGENTS.md already exists.
	if _, err := os.Stat(agentsPath); err == nil && !overwrite {
		result.Skipped = true
		return result, nil
	}

	// Ensure target directory exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return result, err
	}

	// Write AGENTS.md.
	if err := os.WriteFile(agentsPath, []byte(rendered), 0o644); err != nil {
		return result, err
	}
	result.Written = append(result.Written, agentsPath)

	return result, nil
}

// AgentConfigAdapters lists the agent config adapters for all supported assistants.
var AgentConfigAdapters = []AgentConfigAdapterDescriptor{
	{
		AssistantID: AssistantClaudeCode,
		AgentConfigExists: func() bool {
			_, err := os.Stat(claudeAgentDir() + "/AGENTS.md")
			return err == nil
		},
		Write: writeClaudeAgentConfig,
	},
	{
		AssistantID: AssistantOpenCode,
		AgentConfigExists: func() bool {
			_, err := os.Stat(openCodeConfigDir() + "/AGENTS.md")
			return err == nil
		},
		Write: writeOpenCodeAgentConfig,
	},
}
