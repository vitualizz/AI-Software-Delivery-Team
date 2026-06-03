package knowledge

import (
	"context"
	"time"
)

// KnowledgeDetector is the port for scanning a project root and producing a Platform.
type KnowledgeDetector interface {
	// Detect scans projectRoot and returns a Platform populated with the
	// detected stack. It never returns an error for an unknown project —
	// instead it returns a Platform with an empty detected_stack.
	Detect(ctx context.Context, projectRoot string) (Platform, error)
}

// Detector is the concrete implementation of KnowledgeDetector.
// It runs all registered StackProbes against the project root.
type Detector struct {
	probes []StackProbe
}

// NewDetector constructs a Detector with the given set of probes.
func NewDetector(probes []StackProbe) *Detector {
	return &Detector{probes: probes}
}

// DefaultDetector returns a Detector pre-configured with all built-in probes.
func DefaultDetector() *Detector {
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
func (d *Detector) Detect(_ context.Context, projectRoot string) (Platform, error) {
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
		SchemaVersion: PlatformSchemaVersion,
		ScannedAt:     time.Now().UTC(),
		DetectedStack: stack,
		Conventions:   Conventions{},
		DesignFingerprint: DesignFingerprint{},
	}, nil
}
