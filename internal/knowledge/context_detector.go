package knowledge

import "time"

// DetectConfig carries per-run configuration for the context detector.
type DetectConfig struct {
	// PrimaryLanguage is the first entry of detected_stack from platform.yaml.
	PrimaryLanguage string
}

// ContextDetector runs a set of ContextProbes against a project root and
// assembles the results into a ProjectContext.
type ContextDetector struct {
	probes []ContextProbe
}

// NewContextDetector constructs a ContextDetector with the given probes.
func NewContextDetector(probes []ContextProbe) *ContextDetector {
	return &ContextDetector{probes: probes}
}

// DefaultContextDetector returns a ContextDetector pre-configured with all
// built-in probes for the given primary language.
func DefaultContextDetector(primaryLang string) *ContextDetector {
	return NewContextDetector([]ContextProbe{
		MonorepoProbe(),
		TestRunnerProbe(primaryLang),
		NamingStyleProbe(primaryLang),
		ArchitecturalStyleProbe(),
	})
}

// DetectContext runs all probes against projectRoot and assembles a
// ProjectContext. Probe errors are non-fatal: a failed probe produces a
// ContextDetection with source=inferred and confidence=low. The returned
// error is always nil (forward-compatible signature).
func (d *ContextDetector) DetectContext(projectRoot string, cfg DetectConfig) (ProjectContext, error) {
	results := make(map[string]ContextDetection, len(d.probes))

	fallback := ContextDetection{
		Source:     ContextSourceInferred,
		Confidence: ContextConfidenceLow,
	}

	for _, probe := range d.probes {
		det, err := probe.Detect(projectRoot)
		if err != nil {
			// Non-fatal: record fallback for this probe.
			results[probe.Name()] = fallback
			continue
		}
		results[probe.Name()] = det
	}

	ctx := ProjectContext{
		SchemaVersion:      ContextSchemaVersion,
		DetectedAt:         time.Now().UTC(),
		IsMonorepo:         results["is_monorepo"],
		TestRunner:         results["test_runner"],
		NamingStyle:        results["naming_style"],
		ArchitecturalStyle: results["architectural_style"],
	}

	return ctx, nil
}
