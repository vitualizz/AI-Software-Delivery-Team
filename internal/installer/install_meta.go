package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const installMetaFile = ".install-meta.json"

// InstallMeta records the outcome of the last ASDT installation for one assistant.
// All fields are optional — callers must treat zero values as "not yet installed".
type InstallMeta struct {
	InstalledAt time.Time `json:"installed_at,omitempty"`
	Persona     string    `json:"persona,omitempty"`     // PersonaPreset.Name; empty when skipped
	Emojis      string    `json:"emojis,omitempty"`      // "yes"/"no" emoji preference; empty when never chosen
	Language    string    `json:"language,omitempty"`    // base language code ("en"/"es"); empty when never chosen
	AgentTypes  []string  `json:"agent_types,omitempty"` // installed executor agent type IDs (AgentTypeNames)
}

func metaPath(d AssistantDescriptor) string {
	return filepath.Join(d.SkillsDir, siblingConsultantDir, installMetaFile)
}

// WriteInstallMeta persists meta alongside the assistant's skills.
// Errors are best-effort — callers should not fail the install on write error.
func WriteInstallMeta(d AssistantDescriptor, meta InstallMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath(d), data, 0o644)
}

// ReadInstallMeta reads install metadata for the given assistant.
// Returns a zero-value InstallMeta (InstalledAt.IsZero() == true) when the
// file does not exist, so callers can treat missing as "not yet installed".
func ReadInstallMeta(d AssistantDescriptor) (InstallMeta, error) {
	data, err := os.ReadFile(metaPath(d))
	if os.IsNotExist(err) {
		return InstallMeta{}, nil
	}
	if err != nil {
		return InstallMeta{}, err
	}
	var meta InstallMeta
	return meta, json.Unmarshal(data, &meta)
}
