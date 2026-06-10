package setup_test

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

// agentRenderFS is a minimal embedded FS for testing agent config rendering.
var agentRenderFS = fstest.MapFS{
	"asdt-init/agents-template.md": &fstest.MapFile{Data: []byte(`# {{agent_name}}

> {{agent_description}}

## Project Context
- **Stack**: {{stack}}
- **Architecture**: {{architectural_style}}

## Identity

{{persona_block}}
`)},
	"asdt-init/personas/axiom.md":        &fstest.MapFile{Data: []byte(`You are Axiom. Precise and structural.`)},
	"asdt-init/personas/sage.md":         &fstest.MapFile{Data: []byte(`You are Sage. Patient and educational.`)},
	"asdt-init/personas/forge.md":        &fstest.MapFile{Data: []byte(`You are Forge. Direct and pragmatic.`)},
	"asdt-init/personas/lee-palacios.md": &fstest.MapFile{Data: []byte(`You are Lee Palacios. Warm and friendly.`)},
}

// TestRenderAgentConfig_AllPresetsSubstituteCorrectly verifies that for each preset
// the render step does not error (tested via InstallAgentConfig with no adapter →
// Skipped=true means render succeeded before the adapter lookup).
func TestRenderAgentConfig_AllPresetsSubstituteCorrectly(t *testing.T) {
	for _, preset := range installer.PersonaPresets {
		t.Run(preset.ID, func(t *testing.T) {
			results := installer.InstallAgentConfig(
				[]installer.AssistantDescriptor{{ID: "no-adapter-for-render-test"}},
				preset,
				installer.AgentModeOverwrite,
				agentRenderFS,
			)
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
			// Render error would propagate as Err.
			if results[0].Err != nil {
				t.Fatalf("render error for preset %q: %v", preset.ID, results[0].Err)
			}
			// No adapter → Skipped=true (render succeeded but no-op write).
			if !results[0].Skipped {
				t.Fatalf("expected Skipped=true for unknown adapter, got false")
			}
		})
	}
}

// TestRenderAgentConfig_NoPlaceholdersRemain writes AGENTS.md via the Claude Code adapter
// and checks that no {{...}} placeholders survive in the output.
func TestRenderAgentConfig_NoPlaceholdersRemain(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	preset := installer.PersonaPresets[0] // Axiom
	assistants := []installer.AssistantDescriptor{
		{ID: installer.AssistantClaudeCode, Name: "Claude Code"},
	}
	results := installer.InstallAgentConfig(assistants, preset, installer.AgentModeOverwrite, agentRenderFS)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("install error: %v", results[0].Err)
	}
	if len(results[0].Written) == 0 {
		t.Fatal("Written is empty — AGENTS.md path not recorded")
	}

	// First Written entry is AGENTS.md.
	agentsPath := results[0].Written[0]
	data, readErr := os.ReadFile(agentsPath)
	if readErr != nil {
		t.Fatalf("read AGENTS.md: %v", readErr)
	}
	content := string(data)

	for _, ph := range []string{"{{agent_name}}", "{{agent_description}}", "{{persona_block}}", "{{stack}}", "{{architectural_style}}"} {
		if strings.Contains(content, ph) {
			t.Errorf("AGENTS.md still contains placeholder %q", ph)
		}
	}
	if !strings.Contains(content, "Axiom") {
		t.Errorf("AGENTS.md missing preset name 'Axiom', got:\n%s", content)
	}
	if !strings.Contains(content, preset.Description) {
		t.Errorf("AGENTS.md missing preset description, got:\n%s", content)
	}
	if !strings.Contains(content, "detected at project init") {
		t.Errorf("AGENTS.md missing sentinel for stack/arch, got:\n%s", content)
	}
}
