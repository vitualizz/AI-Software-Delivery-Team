package installer

import (
	"io/fs"
	"strings"
)

const agentConfigPlaceholder = "detected at project init"

// PersonaPreset is a selectable agent persona shipped with the installer.
type PersonaPreset struct {
	ID          string // matches persona filename stem: axiom|sage|forge|lee-palacios
	Name        string // "Axiom"
	Description string // one-line summary shown in StateAgentSetup
	File        string // embedded path: "asdt-init/personas/axiom.md"
}

// PersonaPresets lists the built-in agent persona presets in display order.
var PersonaPresets = []PersonaPreset{
	{
		ID:          "axiom",
		Name:        "Axiom",
		Description: "Precise and structural. Asks \"why\" before every decision.",
		File:        "asdt-init/personas/axiom.md",
	},
	{
		ID:          "sage",
		Name:        "Sage",
		Description: "Patient and educational. Explains concepts before code.",
		File:        "asdt-init/personas/sage.md",
	},
	{
		ID:          "forge",
		Name:        "Forge",
		Description: "Direct and pragmatic. Ships clean code fast.",
		File:        "asdt-init/personas/forge.md",
	},
	{
		ID:          "lee-palacios",
		Name:        "Lee Palacios",
		Description: "Warm and friendly. Your pair-programming companion. Adapts to your language. Loves cats.",
		File:        "asdt-init/personas/lee-palacios.md",
	},
}

// AgentWriteMode controls how an existing global agent config is handled.
type AgentWriteMode int

const (
	// AgentModeOverwrite replaces any existing config entirely.
	AgentModeOverwrite AgentWriteMode = iota
	// AgentModeAppend adds the new config at the end of the existing file.
	AgentModeAppend
	// AgentModeSkip leaves the existing config untouched.
	AgentModeSkip
)

// AgentConfigResult is the per-assistant outcome of writing agent config.
type AgentConfigResult struct {
	AssistantID AssistantID
	Written     []string // files written/updated (AGENTS.md, CLAUDE.md)
	Skipped     bool     // true when a pre-existing file was kept (overwrite declined)
	Err         error
}

// AgentConfigAdapterDescriptor declares how one assistant persists agent config.
type AgentConfigAdapterDescriptor struct {
	AssistantID       AssistantID
	AgentConfigExists func() bool
	Write             func(rendered string, mode AgentWriteMode) (AgentConfigResult, error)
}

// AgentConfigAdapterFor returns the AgentConfigAdapterDescriptor for id, if one exists.
func AgentConfigAdapterFor(id AssistantID) (AgentConfigAdapterDescriptor, bool) {
	for _, adapter := range AgentConfigAdapters {
		if adapter.AssistantID == id {
			return adapter, true
		}
	}
	return AgentConfigAdapterDescriptor{}, false
}

// renderAgentConfig reads the AGENTS.md template and persona file from the
// embedded skillsFS and substitutes all placeholders with the preset values.
// Only {{agent_name}}, {{agent_description}}, and {{persona_block}} receive
// real values; {{stack}} and {{architectural_style}} receive the sentinel string.
func renderAgentConfig(skillsFS fs.FS, preset PersonaPreset) (string, error) {
	tmpl, err := fs.ReadFile(skillsFS, "asdt-init/agents-template.md")
	if err != nil {
		return "", err
	}

	persona, err := fs.ReadFile(skillsFS, preset.File)
	if err != nil {
		return "", err
	}

	out := string(tmpl)
	out = strings.ReplaceAll(out, "{{agent_name}}", preset.Name)
	out = strings.ReplaceAll(out, "{{agent_description}}", preset.Description)
	out = strings.ReplaceAll(out, "{{persona_block}}", strings.TrimSpace(string(persona)))
	out = strings.ReplaceAll(out, "{{stack}}", agentConfigPlaceholder)
	out = strings.ReplaceAll(out, "{{architectural_style}}", agentConfigPlaceholder)

	return out, nil
}

// InstallAgentConfig renders the AGENTS.md template for the given preset and
// writes it to the global config location for each selected assistant.
// mode controls how an existing config is handled (overwrite, append, or skip).
// One result per assistant; per-assistant failure does not abort the others.
func InstallAgentConfig(assistants []AssistantDescriptor, preset PersonaPreset, mode AgentWriteMode, skillsFS fs.FS) []AgentConfigResult {
	rendered, err := renderAgentConfig(skillsFS, preset)
	if err != nil {
		// If rendering fails, all adapters fail with the same error.
		results := make([]AgentConfigResult, len(assistants))
		for i, a := range assistants {
			results[i] = AgentConfigResult{AssistantID: a.ID, Err: err}
		}
		return results
	}

	results := make([]AgentConfigResult, len(assistants))
	for i, a := range assistants {
		if mode == AgentModeSkip {
			results[i] = AgentConfigResult{AssistantID: a.ID, Skipped: true}
			continue
		}

		adapter, ok := AgentConfigAdapterFor(a.ID)
		if !ok {
			results[i] = AgentConfigResult{AssistantID: a.ID, Skipped: true}
			continue
		}

		r, writeErr := adapter.Write(rendered, mode)
		if writeErr != nil {
			results[i] = AgentConfigResult{AssistantID: a.ID, Err: writeErr}
			continue
		}
		results[i] = r
	}

	return results
}
