package installer

import (
	"io/fs"
	"strings"
)

const agentConfigPlaceholder = "detected at project init"

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
// modes maps each AssistantID to its AgentWriteMode; assistants absent from
// the map default to AgentModeOverwrite (clean write, no prior conflict).
// One result per assistant; per-assistant failure does not abort the others.
func InstallAgentConfig(assistants []AssistantDescriptor, preset PersonaPreset, modes map[string]AgentWriteMode, skillsFS fs.FS) []AgentConfigResult {
	rendered, err := renderAgentConfig(skillsFS, preset)
	if err != nil {
		results := make([]AgentConfigResult, len(assistants))
		for i, a := range assistants {
			results[i] = AgentConfigResult{AssistantID: a.ID, Err: err}
		}
		return results
	}

	results := make([]AgentConfigResult, len(assistants))
	for i, a := range assistants {
		mode, hasMode := modes[a.ID]
		if !hasMode {
			mode = AgentModeOverwrite // non-conflicted assistant: clean write
		}

		if mode == AgentModeSkip {
			results[i] = AgentConfigResult{AssistantID: a.ID, Skipped: true}
			continue
		}

		adapter, found := AgentConfigAdapterFor(a.ID)
		if !found {
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
