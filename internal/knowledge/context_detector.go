package knowledge

import "time"

// ContextDetector orchestrates all ContextProbes and maps their results
// into a ProjectContext. Probe errors are non-fatal — a failing probe is
// skipped and detection continues with the remaining probes.
type ContextDetector struct {
	probes []ContextProbe
}

// NewContextDetector constructs a ContextDetector with the given probes.
// Passing a custom probe list enables testing individual detection paths.
func NewContextDetector(probes []ContextProbe) *ContextDetector {
	return &ContextDetector{probes: probes}
}

// DefaultContextDetector returns a ContextDetector pre-configured with all
// built-in probes for the given primary language.
func DefaultContextDetector(primaryLang string) *ContextDetector {
	return NewContextDetector([]ContextProbe{
		MonorepoProbe(),
		TestRunnerProbe(primaryLang),
		NamingStyleDetector(primaryLang, NamingStyleSampleSize),
		ArchitecturalStyleDetector(),
	})
}

// DetectContext runs all probes against projectRoot and returns a ProjectContext.
// It always returns a valid ProjectContext — probe errors cause the corresponding
// field to retain its zero value and detection continues. The returned error is
// always nil; the signature retains the error return for forward compatibility.
func (d *ContextDetector) DetectContext(projectRoot string, cfg DetectConfig) (ProjectContext, error) {
	ctx := ProjectContext{
		SchemaVersion: ContextSchemaVersion,
		DetectedAt:    time.Now().UTC(),
	}

	for _, probe := range d.probes {
		detection, err := probe.Detect(projectRoot)
		if err != nil {
			// Non-fatal: skip this probe, continue with the rest.
			continue
		}
		routeDetection(&ctx, probe.Field(), detection)
	}

	return ctx, nil
}

// routeDetection maps a detection result into the correct ProjectContext field
// based on the probe's declared field name.
func routeDetection(ctx *ProjectContext, field string, d ContextDetection) {
	switch field {
	case "is_monorepo":
		ctx.IsMonorepo = d
	case "test_runner":
		ctx.TestRunner = d
	case "naming_style":
		ctx.NamingStyle = d
	case "architectural_style":
		ctx.ArchitecturalStyle = d
	}
}
