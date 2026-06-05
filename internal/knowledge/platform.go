// Package knowledge implements the codebase scanner that produces platform.yaml.
// It defines the Platform data model and the detection / R/W interfaces.
package knowledge

import "time"

// PlatformSummary is the deterministic, project-level platform digest produced
// by `asdt init`. It lives at {root}/knowledge/platform-summary.yaml and is
// reused by specialist steps instead of re-running LLM platform analysis.
// All fields are required; empty values MUST be written, never omitted.
// Note: node package_manager defaults to "npm" (yarn/pnpm detection is a future
// file-walk enhancement — deterministic simplicity is prioritised now).
type PlatformSummary struct {
	Stack           []string  `yaml:"stack"`
	PrimaryLanguage string    `yaml:"primary_language"`
	PackageManager  string    `yaml:"package_manager"`
	TestRunner      string    `yaml:"test_runner"`
	DetectedAt      time.Time `yaml:"detected_at"`
}

// DeriveSummary deterministically derives a PlatformSummary from a scanned
// Platform. Pure: no I/O, no clock — DetectedAt is copied from p.ScannedAt so
// the output is fully reproducible from the same input.
func DeriveSummary(p Platform) PlatformSummary {
	lang := primaryLanguage(p.DetectedStack)
	pm, tr := toolingFor(lang)
	return PlatformSummary{
		Stack:           p.DetectedStack,
		PrimaryLanguage: lang,
		PackageManager:  pm,
		TestRunner:      tr,
		DetectedAt:      p.ScannedAt,
	}
}

// primaryLanguage returns the first entry of the stack, or "unknown" if empty.
func primaryLanguage(stack []string) string {
	if len(stack) == 0 {
		return "unknown"
	}
	return stack[0]
}

// toolingFor returns (packageManager, testRunner) for a primary language.
// Returns empty strings for unrecognised languages.
func toolingFor(lang string) (string, string) {
	switch lang {
	case "go":
		return "go modules", "go test"
	case "node":
		return "npm", "jest"
	case "rust":
		return "cargo", "cargo test"
	case "python":
		return "pip", "pytest"
	case "ruby":
		return "bundler", "rspec"
	default:
		return "", ""
	}
}

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
