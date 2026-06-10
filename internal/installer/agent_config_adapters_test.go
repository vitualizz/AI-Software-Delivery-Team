package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteClaudeAgentConfig_WritesCLAUDEMD(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	rendered := "# Test Agent\nThis is the agent config."
	result, err := writeClaudeAgentConfig(rendered, AgentModeOverwrite)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}
	if result.Skipped {
		t.Errorf("expected Skipped=false, got true")
	}

	claudePath := filepath.Join(tmpHome, ".claude", "CLAUDE.md")
	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md not found: %v", readErr)
	}
	if !strings.Contains(string(data), rendered) {
		t.Errorf("CLAUDE.md missing rendered content: got %q", string(data))
	}
	if !contains(strings.Join(result.Written, ","), claudePath) {
		t.Errorf("claudePath %q not in Written: %v", claudePath, result.Written)
	}
}

func TestWriteClaudeAgentConfig_CreatesCLAUDEMD_WhenAbsent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, err := writeClaudeAgentConfig("# Agent", AgentModeOverwrite)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}

	claudePath := filepath.Join(tmpHome, ".claude", "CLAUDE.md")
	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md should have been created: %v", readErr)
	}
	if !strings.Contains(string(data), "# Agent") {
		t.Errorf("CLAUDE.md content missing rendered block, got %q", string(data))
	}
}

func TestWriteClaudeAgentConfig_IdempotentBlockReplacement(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")
	initial := "# My Config\n\nSome existing content.\n"
	if err := os.WriteFile(claudePath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run twice — idempotent: block should appear only once.
	_, err := writeClaudeAgentConfig("# Agent", AgentModeOverwrite)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	_, err = writeClaudeAgentConfig("# Agent Updated", AgentModeOverwrite)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md read: %v", readErr)
	}
	count := strings.Count(string(data), asdtBlockStart)
	if count != 1 {
		t.Errorf("CLAUDE.md contains block start %d time(s), want exactly 1\ncontent:\n%s", count, string(data))
	}
	// Updated content must be present.
	if !strings.Contains(string(data), "# Agent Updated") {
		t.Errorf("CLAUDE.md missing updated content, got:\n%s", string(data))
	}
}

func TestWriteClaudeAgentConfig_AppendsToCLAUDEMD_WhenRefAbsent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")
	existing := "# My Config\n\nSome content here.\n"
	if err := os.WriteFile(claudePath, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := writeClaudeAgentConfig("# Agent", AgentModeOverwrite)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}

	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md read: %v", readErr)
	}
	if !strings.Contains(string(data), asdtBlockStart) {
		t.Errorf("CLAUDE.md should contain block after injection, got:\n%s", string(data))
	}
	// CLAUDE.md should be in Written list since it was updated.
	wroteClaudeMD := false
	for _, p := range result.Written {
		if p == claudePath {
			wroteClaudeMD = true
		}
	}
	if !wroteClaudeMD {
		t.Errorf("CLAUDE.md %q should be in Written list: %v", claudePath, result.Written)
	}
}

func TestWriteClaudeAgentConfig_SkipMode_ExistingFileUntouched(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")
	originalContent := "# Original Content"
	if err := os.WriteFile(claudePath, []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := writeClaudeAgentConfig("# New Content", AgentModeSkip)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}
	if !result.Skipped {
		t.Errorf("expected Skipped=true when mode=AgentModeSkip")
	}

	// File must be unchanged.
	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("read CLAUDE.md: %v", readErr)
	}
	if string(data) != originalContent {
		t.Errorf("CLAUDE.md was modified despite AgentModeSkip: got %q, want %q", string(data), originalContent)
	}
}

func TestWriteClaudeAgentConfig_AppendMode_AddsSecondBlock(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")

	// Write an initial block via Overwrite mode.
	_, err := writeClaudeAgentConfig("# First Agent", AgentModeOverwrite)
	if err != nil {
		t.Fatalf("first write: %v", err)
	}

	// Append a second block.
	_, err = writeClaudeAgentConfig("# Second Agent", AgentModeAppend)
	if err != nil {
		t.Fatalf("append write: %v", err)
	}

	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md read: %v", readErr)
	}
	content := string(data)

	// Both blocks must be present.
	if !strings.Contains(content, "# First Agent") {
		t.Errorf("CLAUDE.md missing first block content, got:\n%s", content)
	}
	if !strings.Contains(content, "# Second Agent") {
		t.Errorf("CLAUDE.md missing second (appended) block content, got:\n%s", content)
	}
	// Block markers should appear twice (once per block).
	if count := strings.Count(content, asdtBlockStart); count != 2 {
		t.Errorf("CLAUDE.md should have 2 block start markers after append, got %d\ncontent:\n%s", count, content)
	}
}

func TestWriteOpenCodeAgentConfig_WritesAgentsMD(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	rendered := "# Test Agent"
	result, err := writeOpenCodeAgentConfig(rendered, AgentModeOverwrite)
	if err != nil {
		t.Fatalf("writeOpenCodeAgentConfig: %v", err)
	}
	if result.Skipped {
		t.Errorf("expected Skipped=false, got true")
	}

	agentsPath := filepath.Join(tmpHome, ".config", "opencode", "AGENTS.md")
	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("AGENTS.md not found: %v", readErr)
	}
	if string(data) != rendered {
		t.Errorf("AGENTS.md content mismatch: got %q, want %q", string(data), rendered)
	}
	_ = result
}

func TestWriteOpenCodeAgentConfig_RespectsXDGConfigHome(t *testing.T) {
	tmpXDG := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpXDG)

	rendered := "# XDG Agent"
	_, err := writeOpenCodeAgentConfig(rendered, AgentModeOverwrite)
	if err != nil {
		t.Fatalf("writeOpenCodeAgentConfig: %v", err)
	}

	agentsPath := filepath.Join(tmpXDG, "opencode", "AGENTS.md")
	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("AGENTS.md not found at XDG path: %v", readErr)
	}
	if string(data) != rendered {
		t.Errorf("AGENTS.md content mismatch: got %q, want %q", string(data), rendered)
	}
}

func TestWriteOpenCodeAgentConfig_SkipMode_ExistingFileUntouched(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	dir := filepath.Join(tmpHome, ".config", "opencode")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentsPath := filepath.Join(dir, "AGENTS.md")
	originalContent := "# Original"
	if err := os.WriteFile(agentsPath, []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := writeOpenCodeAgentConfig("# New Content", AgentModeSkip)
	if err != nil {
		t.Fatalf("writeOpenCodeAgentConfig: %v", err)
	}
	if !result.Skipped {
		t.Errorf("expected Skipped=true when mode=AgentModeSkip")
	}

	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("read AGENTS.md: %v", readErr)
	}
	if string(data) != originalContent {
		t.Errorf("AGENTS.md was modified despite AgentModeSkip")
	}
}

func TestWriteOpenCodeAgentConfig_AppendMode_AppendsContent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	dir := filepath.Join(tmpHome, ".config", "opencode")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentsPath := filepath.Join(dir, "AGENTS.md")
	originalContent := "# Original Agent"
	if err := os.WriteFile(agentsPath, []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := writeOpenCodeAgentConfig("# Appended Agent", AgentModeAppend)
	if err != nil {
		t.Fatalf("writeOpenCodeAgentConfig append: %v", err)
	}

	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("read AGENTS.md: %v", readErr)
	}
	content := string(data)

	if !strings.Contains(content, originalContent) {
		t.Errorf("AGENTS.md missing original content after append, got:\n%s", content)
	}
	if !strings.Contains(content, "# Appended Agent") {
		t.Errorf("AGENTS.md missing appended content, got:\n%s", content)
	}
}
