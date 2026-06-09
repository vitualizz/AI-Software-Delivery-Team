package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteClaudeAgentConfig_WritesAgentsMD(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	rendered := "# Test Agent\nThis is the agent config."
	result, err := writeClaudeAgentConfig(rendered, true)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}
	if result.Skipped {
		t.Errorf("expected Skipped=false, got true")
	}

	agentsPath := filepath.Join(tmpHome, ".claude", "AGENTS.md")
	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("AGENTS.md not found: %v", readErr)
	}
	if string(data) != rendered {
		t.Errorf("AGENTS.md content mismatch: got %q, want %q", string(data), rendered)
	}
	if !contains(strings.Join(result.Written, ","), agentsPath) {
		t.Errorf("agentsPath %q not in Written: %v", agentsPath, result.Written)
	}
}

func TestWriteClaudeAgentConfig_CreatesCLAUDEMD_WhenAbsent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, err := writeClaudeAgentConfig("# Agent", true)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}

	claudePath := filepath.Join(tmpHome, ".claude", "CLAUDE.md")
	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md should have been created: %v", readErr)
	}
	if strings.TrimSpace(string(data)) != "@AGENTS.md" {
		t.Errorf("CLAUDE.md content = %q, want %q", string(data), "@AGENTS.md")
	}
}

func TestWriteClaudeAgentConfig_IdempotentCLAUDEMDInjection(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")
	initial := "# My Config\n\n@AGENTS.md\n"
	if err := os.WriteFile(claudePath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run twice — idempotent: @AGENTS.md should appear only once.
	_, err := writeClaudeAgentConfig("# Agent", true)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	_, err = writeClaudeAgentConfig("# Agent Updated", true)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md read: %v", readErr)
	}
	count := strings.Count(string(data), "@AGENTS.md")
	if count != 1 {
		t.Errorf("CLAUDE.md contains @AGENTS.md %d time(s), want exactly 1\ncontent:\n%s", count, string(data))
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

	result, err := writeClaudeAgentConfig("# Agent", true)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}

	data, readErr := os.ReadFile(claudePath)
	if readErr != nil {
		t.Fatalf("CLAUDE.md read: %v", readErr)
	}
	if !strings.Contains(string(data), "@AGENTS.md") {
		t.Errorf("CLAUDE.md should contain @AGENTS.md after injection, got:\n%s", string(data))
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

func TestWriteClaudeAgentConfig_OverwriteFalse_ExistingFileSkipped(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentsPath := filepath.Join(claudeDir, "AGENTS.md")
	originalContent := "# Original Content"
	if err := os.WriteFile(agentsPath, []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := writeClaudeAgentConfig("# New Content", false)
	if err != nil {
		t.Fatalf("writeClaudeAgentConfig: %v", err)
	}
	if !result.Skipped {
		t.Errorf("expected Skipped=true when overwrite=false and file exists")
	}

	// File must be unchanged.
	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("read AGENTS.md: %v", readErr)
	}
	if string(data) != originalContent {
		t.Errorf("AGENTS.md was modified despite overwrite=false: got %q, want %q", string(data), originalContent)
	}
}

func TestWriteOpenCodeAgentConfig_WritesAgentsMD(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	rendered := "# Test Agent"
	result, err := writeOpenCodeAgentConfig(rendered, true)
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
	_, err := writeOpenCodeAgentConfig(rendered, true)
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

func TestWriteOpenCodeAgentConfig_OverwriteFalse_ExistingFileSkipped(t *testing.T) {
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

	result, err := writeOpenCodeAgentConfig("# New Content", false)
	if err != nil {
		t.Fatalf("writeOpenCodeAgentConfig: %v", err)
	}
	if !result.Skipped {
		t.Errorf("expected Skipped=true when overwrite=false and file exists")
	}

	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("read AGENTS.md: %v", readErr)
	}
	if string(data) != originalContent {
		t.Errorf("AGENTS.md was modified despite overwrite=false")
	}
}

func TestContainsAgentsRef(t *testing.T) {
	cases := []struct {
		content string
		want    bool
	}{
		{"@AGENTS.md", true},
		{"# Config\n\n@AGENTS.md\n", true},
		{"# Config\n@AGENTS.md\nmore", true},
		{"# Config\nNo ref here\n", false},
		{"@agents.md", false}, // case-sensitive
		{"  @AGENTS.md  ", true}, // trimmed
		{"", false},
	}
	for _, c := range cases {
		got := containsAgentsRef(c.content)
		if got != c.want {
			t.Errorf("containsAgentsRef(%q) = %v, want %v", c.content, got, c.want)
		}
	}
}
