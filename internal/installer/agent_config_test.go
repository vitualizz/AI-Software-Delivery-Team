package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

var agentTestFS = fstest.MapFS{
	"asdt-init/agents-template.md": &fstest.MapFile{Data: []byte(`# {{agent_name}}

> {{agent_description}}

## Project Context
- **Stack**: {{stack}}
- **Architecture**: {{architectural_style}}

## Identity

{{persona_block}}

{{emoji_preference}}
`)},
	"asdt-init/personas/sky.md":          &fstest.MapFile{Data: []byte(`You are Sky. Sharp and thorough.`)},
	"asdt-init/personas/toffy.md":        &fstest.MapFile{Data: []byte(`You are Toffy. Warm and enthusiastic.`)},
	"asdt-init/personas/atreus.md":       &fstest.MapFile{Data: []byte(`You are Atreus. Bold and reckless.`)},
	"asdt-init/personas/babi.md":         &fstest.MapFile{Data: []byte(`You are Babi. Your biggest fan.`)},
	"asdt-init/personas/lee-palacios.md": &fstest.MapFile{Data: []byte(`You are Lee Palacios. Cat lover, coder, otaku.`)},
}

func TestRenderAgentConfig_SubstitutesAllPlaceholders(t *testing.T) {
	for _, preset := range PersonaPresets {
		t.Run(preset.ID, func(t *testing.T) {
			out, err := renderAgentConfig(agentTestFS, preset, true)
			if err != nil {
				t.Fatalf("renderAgentConfig(%q): %v", preset.ID, err)
			}
			// No placeholders should remain.
			for _, placeholder := range []string{"{{agent_name}}", "{{agent_description}}", "{{persona_block}}", "{{emoji_preference}}", "{{stack}}", "{{architectural_style}}"} {
				if contains(out, placeholder) {
					t.Errorf("output still contains placeholder %q", placeholder)
				}
			}
			// Name and description should be present.
			if !contains(out, preset.Name) {
				t.Errorf("output missing preset name %q", preset.Name)
			}
			if !contains(out, preset.Description) {
				t.Errorf("output missing preset description %q", preset.Description)
			}
			// Stack and architectural_style should use the sentinel.
			if !contains(out, agentConfigPlaceholder) {
				t.Errorf("output missing sentinel %q for stack/architectural_style", agentConfigPlaceholder)
			}
		})
	}
}

func TestRenderAgentConfig_PersonaBlockPresent(t *testing.T) {
	preset := PersonaPresets[0] // Sky
	out, err := renderAgentConfig(agentTestFS, preset, true)
	if err != nil {
		t.Fatalf("renderAgentConfig: %v", err)
	}
	if !contains(out, "You are Sky") {
		t.Errorf("output missing persona block content, got:\n%s", out)
	}
}

// TestRenderAgentConfig_EmojiBulletVariants verifies that {{emoji_preference}}
// renders to exactly one "- **Emojis**:" bullet with the exact copy for each
// answer, and that no placeholder residue survives.
func TestRenderAgentConfig_EmojiBulletVariants(t *testing.T) {
	cases := []struct {
		name      string
		useEmojis bool
		want      string
	}{
		{name: "yes", useEmojis: true, want: emojiPrefYes},
		{name: "no", useEmojis: false, want: emojiPrefNo},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := renderAgentConfig(agentTestFS, PersonaPresets[0], c.useEmojis)
			if err != nil {
				t.Fatalf("renderAgentConfig: %v", err)
			}
			if got := strings.Count(out, "- **Emojis**:"); got != 1 {
				t.Errorf("output contains %d '- **Emojis**:' bullets, want exactly 1", got)
			}
			if !strings.Contains(out, c.want) {
				t.Errorf("output missing emoji bullet %q, got:\n%s", c.want, out)
			}
			if strings.Contains(out, "{{") {
				t.Errorf("output contains placeholder residue, got:\n%s", out)
			}
		})
	}
}

// TestInstallAgentConfig_PersistsEmojiMeta verifies the best-effort meta block
// records the emoji preference alongside the persona on a successful write.
func TestInstallAgentConfig_PersistsEmojiMeta(t *testing.T) {
	cases := []struct {
		name      string
		useEmojis bool
		want      string
	}{
		{name: "yes", useEmojis: true, want: "yes"},
		{name: "no", useEmojis: false, want: "no"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_CONFIG_HOME", "")

			skillsDir := filepath.Join(t.TempDir(), "skills")
			if err := os.MkdirAll(filepath.Join(skillsDir, "asdt"), 0o755); err != nil {
				t.Fatal(err)
			}
			a := AssistantDescriptor{ID: AssistantClaudeCode, Name: "Claude Code", SkillsDir: skillsDir}

			results := InstallAgentConfig([]AssistantDescriptor{a}, PersonaPresets[0], c.useEmojis, map[string]AgentWriteMode{}, agentTestFS)
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
			if results[0].Err != nil {
				t.Fatalf("install error: %v", results[0].Err)
			}

			meta, err := ReadInstallMeta(a)
			if err != nil {
				t.Fatalf("read meta: %v", err)
			}
			if meta.Emojis != c.want {
				t.Errorf("meta.Emojis = %q, want %q", meta.Emojis, c.want)
			}
			if meta.Persona != PersonaPresets[0].Name {
				t.Errorf("meta.Persona = %q, want %q", meta.Persona, PersonaPresets[0].Name)
			}
		})
	}
}

func TestInstallAgentConfig_NoAdapterSkipsSilently(t *testing.T) {
	// Use an assistant ID that has no registered adapter.
	unknownAssistant := AssistantDescriptor{ID: "unknown-ai", Name: "Unknown AI"}
	results := InstallAgentConfig([]AssistantDescriptor{unknownAssistant}, PersonaPresets[0], true, map[string]AgentWriteMode{}, agentTestFS)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Errorf("expected Skipped=true for unknown assistant, got Skipped=%v", results[0].Skipped)
	}
	if results[0].Err != nil {
		t.Errorf("expected no error for unknown assistant, got %v", results[0].Err)
	}
}

func TestInstallAgentConfig_PerAssistantIsolation(t *testing.T) {
	// Two assistants: unknown (skip) + another unknown. Each should have its own result.
	a1 := AssistantDescriptor{ID: "no-adapter-1", Name: "No Adapter 1"}
	a2 := AssistantDescriptor{ID: "no-adapter-2", Name: "No Adapter 2"}
	results := InstallAgentConfig([]AssistantDescriptor{a1, a2}, PersonaPresets[0], true, map[string]AgentWriteMode{}, agentTestFS)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if !r.Skipped {
			t.Errorf("results[%d]: expected Skipped=true, got false", i)
		}
	}
}

func TestInstallAgentConfig_RenderErrorPropagatedToAllAssistants(t *testing.T) {
	// Use an FS that is missing the template.
	emptyFS := fstest.MapFS{}
	a := AssistantDescriptor{ID: AssistantClaudeCode, Name: "Claude Code"}
	results := InstallAgentConfig([]AssistantDescriptor{a}, PersonaPresets[0], true, map[string]AgentWriteMode{}, emptyFS)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error when template is missing, got nil")
	}
}

func TestInstallAgentConfig_SkipMode_SkipsAllAssistants(t *testing.T) {
	a1 := AssistantDescriptor{ID: AssistantClaudeCode, Name: "Claude Code"}
	a2 := AssistantDescriptor{ID: AssistantOpenCode, Name: "OpenCode"}
	results := InstallAgentConfig([]AssistantDescriptor{a1, a2}, PersonaPresets[0], true, map[string]AgentWriteMode{
		string(AssistantClaudeCode): AgentModeSkip,
		string(AssistantOpenCode):   AgentModeSkip,
	}, agentTestFS)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if !r.Skipped {
			t.Errorf("results[%d]: expected Skipped=true for AgentModeSkip, got false", i)
		}
		if r.Err != nil {
			t.Errorf("results[%d]: expected no error for AgentModeSkip, got %v", i, r.Err)
		}
	}
}

func TestAgentConfigAdapterFor_KnownIDs(t *testing.T) {
	cases := []struct {
		id        AssistantID
		wantFound bool
	}{
		{AssistantClaudeCode, true},
		{AssistantOpenCode, true},
		{"unknown-assistant", false},
	}
	for _, c := range cases {
		t.Run(string(c.id), func(t *testing.T) {
			_, ok := AgentConfigAdapterFor(c.id)
			if ok != c.wantFound {
				t.Errorf("AgentConfigAdapterFor(%q) found=%v, want %v", c.id, ok, c.wantFound)
			}
		})
	}
}

// contains is a simple substring check helper for this package's tests.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexString(s, sub) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
