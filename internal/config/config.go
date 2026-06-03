package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the persisted settings for an ASDT project.
// It is serialized to/from {root}/config.yaml.
type Config struct {
	// ActiveChange is the name of the currently active change.
	// Used as the default --change value when not specified explicitly.
	ActiveChange string `yaml:"active_change,omitempty"`

	// Defaults holds project-level default settings.
	Defaults map[string]string `yaml:"defaults,omitempty"`
}

// configPath returns the absolute path to config.yaml within the given root.
func configPath(root Root) string {
	return filepath.Join(root.Path(), "config.yaml")
}

// Load reads config.yaml from the given root. If the file does not exist,
// it returns a zero-value Config (not an error) — fresh projects start unconfigured.
func Load(root Root) (Config, error) {
	data, err := os.ReadFile(configPath(root))
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("config load: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config unmarshal: %w", err)
	}
	return cfg, nil
}

// Save serializes cfg and writes it to config.yaml inside root.
// Creates the root directory if it does not yet exist.
func Save(root Root, cfg Config) error {
	if err := os.MkdirAll(root.Path(), 0o755); err != nil {
		return fmt.Errorf("config save mkdir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config marshal: %w", err)
	}
	if err := os.WriteFile(configPath(root), data, 0o644); err != nil {
		return fmt.Errorf("config write: %w", err)
	}
	return nil
}
