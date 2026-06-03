// Package knowledge implements the codebase scanner that produces platform.yaml.
// It defines the Platform data model and the detection / R/W interfaces.
package knowledge

import "time"

// PlatformSchemaVersion is the schema version for platform.yaml files.
const PlatformSchemaVersion = "1"

// Platform represents the content of .asdt/knowledge/platform.yaml.
// It matches platform.schema.yaml exactly.
type Platform struct {
	SchemaVersion     string            `yaml:"schema_version"`
	ScannedAt         time.Time         `yaml:"scanned_at"`
	DetectedStack     []string          `yaml:"detected_stack"`
	Conventions       Conventions       `yaml:"conventions"`
	DesignFingerprint DesignFingerprint `yaml:"design_fingerprint"`
}

// Conventions holds detected project-level naming and structure conventions.
type Conventions struct {
	Naming        map[string]string `yaml:"naming,omitempty"`
	FileStructure string            `yaml:"file_structure,omitempty"`
}

// DesignFingerprint captures structural design markers of the project.
type DesignFingerprint struct {
	ComponentLibrary string   `yaml:"component_library,omitempty"`
	CSSApproach      string   `yaml:"css_approach,omitempty"`
	LayoutPatterns   []string `yaml:"layout_patterns,omitempty"`
}
