package knowledge

import (
	"context"
	"time"
)

// Detector is the port for scanning a project root and producing a Platform.
type Detector interface {
	// Detect scans projectRoot and returns a Platform populated with the
	// detected stack. It never returns an error for an unknown project —
	// instead it returns a Platform with an empty detected_stack.
	Detect(ctx context.Context, projectRoot string) (Platform, error)
}

// ProbeDetector is the concrete implementation of Detector.
// It runs all registered StackProbes against the project root.
type ProbeDetector struct {
	probes []StackProbe
}

// NewDetector constructs a ProbeDetector with the given set of probes.
func NewDetector(probes []StackProbe) *ProbeDetector {
	return &ProbeDetector{probes: probes}
}

// DefaultDetector returns a ProbeDetector pre-configured with all built-in probes.
func DefaultDetector() *ProbeDetector {
	return NewDetector([]StackProbe{
		GoProbe(),
		NodeProbe(),
		RustProbe(),
		PythonProbe(),
		RubyProbe(),
	})
}

// Detect runs all probes against projectRoot and returns a Platform.
// Unknown projects return a Platform with an empty DetectedStack — no error.
func (d *ProbeDetector) Detect(_ context.Context, projectRoot string) (Platform, error) {
	var stack []string

	for _, probe := range d.probes {
		found, err := probe.Detect(projectRoot)
		if err != nil {
			// Treat probe errors as non-fatal — skip this probe.
			continue
		}
		if found {
			stack = append(stack, probe.Name())
		}
	}

	if stack == nil {
		stack = []string{}
	}

	return Platform{
		SchemaVersion:     PlatformSchemaVersion,
		ScannedAt:         time.Now().UTC(),
		DetectedStack:     stack,
		Conventions:       Conventions{},
		DesignFingerprint: DesignFingerprint{},
	}, nil
}
