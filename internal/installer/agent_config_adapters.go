package installer

import (
	"os"
	"strings"
)

const (
	// asdtBlockStart and asdtBlockEnd are the HTML comment markers used to
	// wrap the ASDT agent config block injected into CLAUDE.md. Using markers
	// rather than a separate file keeps Claude Code's native config untouched
	// while allowing idempotent updates and clean removal.
	asdtBlockStart = "<!-- asdt:agent-config -->"
	asdtBlockEnd   = "<!-- /asdt:agent-config -->"
)

// claudeAgentDir returns the directory where Claude Code's global config lives (~/.claude).
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

// writeClaudeAgentConfig writes the rendered agent config as a tagged block
// directly into ~/.claude/CLAUDE.md — Claude Code's native config file.
//
// AgentModeOverwrite: replaces the existing block in-place (or appends if absent).
// AgentModeAppend: always appends a new block at the end of the file.
// AgentModeSkip: leaves the file untouched and returns Skipped=true.
func writeClaudeAgentConfig(rendered string, mode AgentWriteMode) (AgentConfigResult, error) {
	result := AgentConfigResult{AssistantID: AssistantClaudeCode}

	if mode == AgentModeSkip {
		result.Skipped = true
		return result, nil
	}

	dir := claudeAgentDir()
	claudePath := dir + "/CLAUDE.md"

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return result, err
	}

	existing, err := os.ReadFile(claudePath)
	if err != nil && !os.IsNotExist(err) {
		return result, err
	}

	existingStr := string(existing)
	block := asdtBlockStart + "\n" + rendered + "\n" + asdtBlockEnd

	var newContent string
	switch mode {
	case AgentModeAppend:
		// Always append at end; do not touch any existing block.
		switch {
		case existingStr == "":
			newContent = block
		case strings.HasSuffix(existingStr, "\n"):
			newContent = existingStr + "\n" + block
		default:
			newContent = existingStr + "\n\n" + block
		}
	default: // AgentModeOverwrite
		blockExists := strings.Contains(existingStr, asdtBlockStart)
		switch {
		case blockExists:
			newContent = replaceAsdtBlock(existingStr, block)
		case existingStr == "":
			newContent = block
		case strings.HasSuffix(existingStr, "\n"):
			newContent = existingStr + "\n" + block
		default:
			newContent = existingStr + "\n\n" + block
		}
	}

	if err := os.WriteFile(claudePath, []byte(newContent), 0o644); err != nil {
		return result, err
	}
	result.Written = append(result.Written, claudePath)
	return result, nil
}

// replaceAsdtBlock replaces the content between asdtBlockStart and asdtBlockEnd
// with newBlock. If the markers are missing or malformed, newBlock is appended.
func replaceAsdtBlock(content, newBlock string) string {
	start := strings.Index(content, asdtBlockStart)
	end := strings.Index(content, asdtBlockEnd)
	if start == -1 || end == -1 || end < start {
		return content + "\n\n" + newBlock
	}
	return content[:start] + newBlock + content[end+len(asdtBlockEnd):]
}

// writeOpenCodeAgentConfig writes rendered AGENTS.md to the OpenCode config directory.
// OpenCode reads AGENTS.md natively as its global agent config.
//
// AgentModeOverwrite: replaces the file entirely.
// AgentModeAppend: reads existing content (if any) and appends the rendered block.
// AgentModeSkip: leaves the file untouched and returns Skipped=true.
func writeOpenCodeAgentConfig(rendered string, mode AgentWriteMode) (AgentConfigResult, error) {
	result := AgentConfigResult{AssistantID: AssistantOpenCode}

	if mode == AgentModeSkip {
		result.Skipped = true
		return result, nil
	}

	dir := openCodeConfigDir()
	agentsPath := dir + "/AGENTS.md"

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return result, err
	}

	var content string
	if mode == AgentModeAppend {
		existing, err := os.ReadFile(agentsPath)
		if err != nil && !os.IsNotExist(err) {
			return result, err
		}
		existingStr := string(existing)
		if existingStr == "" {
			content = rendered
		} else {
			content = existingStr + "\n\n" + rendered
		}
	} else { // AgentModeOverwrite
		content = rendered
	}

	if err := os.WriteFile(agentsPath, []byte(content), 0o644); err != nil {
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
			data, err := os.ReadFile(claudeAgentDir() + "/CLAUDE.md")
			if err != nil {
				return false
			}
			return strings.Contains(string(data), asdtBlockStart)
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
