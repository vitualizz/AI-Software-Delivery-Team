package knowledge

import "time"

// ContextSchemaVersion is the schema version for project-context.yaml files.
const ContextSchemaVersion = "1"

// NamingStyleSampleSize is the number of source files sampled by NamingStyleDetector.
// This constant is intentionally locked — do not make it configurable.
const NamingStyleSampleSize = 8

// ContextSource describes how a detection result was obtained.
type ContextSource string

const (
	// ContextSourceDetected means the value was determined by a bounded command
	// with direct file evidence.
	ContextSourceDetected ContextSource = "detected"

	// ContextSourceInferred means the value was inferred from indirect pattern
	// matching without direct file evidence.
	ContextSourceInferred ContextSource = "inferred"

	// ContextSourceManual means the value was explicitly set by the user during
	// a recalibration review. Fields with this source are never silently overwritten.
	ContextSourceManual ContextSource = "manual"
)

// ContextConfidence describes how confident the detector is in the result.
type ContextConfidence string

const (
	// ContextConfidenceHigh means the signal is strong — treat the value as an
	// authoritative convention.
	ContextConfidenceHigh ContextConfidence = "high"

	// ContextConfidenceMedium means the signal is likely — confirm before
	// diverging from this convention.
	ContextConfidenceMedium ContextConfidence = "medium"

	// ContextConfidenceLow means the signal is weak — treat the value as a
	// best-effort guess.
	ContextConfidenceLow ContextConfidence = "low"
)

// ContextDetection is a single detected field with provenance metadata.
type ContextDetection struct {
	Value      string            `yaml:"value"`
	Source     ContextSource     `yaml:"source"`
	Confidence ContextConfidence `yaml:"confidence"`
}

// ProjectContext is the top-level structure written to
// {root}/knowledge/project-context.yaml by asdt init Step 4.
type ProjectContext struct {
	SchemaVersion      string           `yaml:"schema_version"`
	DetectedAt         time.Time        `yaml:"detected_at"`
	IsMonorepo         ContextDetection `yaml:"is_monorepo"`
	TestRunner         ContextDetection `yaml:"test_runner"`
	NamingStyle        ContextDetection `yaml:"naming_style"`
	ArchitecturalStyle ContextDetection `yaml:"architectural_style"`
}

