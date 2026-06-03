package knowledge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
)

// moduleRoot returns the absolute path to the repository root
// (two levels above internal/knowledge).
func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// internal/knowledge → ../../
	return filepath.Join(wd, "..", "..")
}

// fixturesDir returns the testdata/projects path.
func fixturesDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(moduleRoot(t), "testdata", "projects")
}

// makeRoot creates a temporary .asdt/ directory and returns the config.Root.
func makeRoot(t *testing.T) config.Root {
	t.Helper()
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir .asdt: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return root
}

func TestGoProbe_DetectsGoProject(t *testing.T) {
	probe := knowledge.GoProbe()

	if probe.Name() != "go" {
		t.Errorf("Name: got %q, want %q", probe.Name(), "go")
	}

	goDir := filepath.Join(fixturesDir(t), "go-project")
	found, err := probe.Detect(goDir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !found {
		t.Error("GoProbe: expected true for go-project, got false")
	}
}

func TestGoProbe_DoesNotDetectNodeProject(t *testing.T) {
	probe := knowledge.GoProbe()
	nodeDir := filepath.Join(fixturesDir(t), "node-project")
	found, err := probe.Detect(nodeDir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if found {
		t.Error("GoProbe: expected false for node-project, got true")
	}
}

func TestNodeProbe_DetectsNodeProject(t *testing.T) {
	probe := knowledge.NodeProbe()

	if probe.Name() != "node" {
		t.Errorf("Name: got %q, want %q", probe.Name(), "node")
	}

	nodeDir := filepath.Join(fixturesDir(t), "node-project")
	found, err := probe.Detect(nodeDir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !found {
		t.Error("NodeProbe: expected true for node-project, got false")
	}
}

func TestDetector_MultiStack(t *testing.T) {
	det := knowledge.DefaultDetector()
	multiDir := filepath.Join(fixturesDir(t), "multi-stack")

	p, err := det.Detect(context.Background(), multiDir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	inStack := func(name string) bool {
		for _, s := range p.DetectedStack {
			if s == name {
				return true
			}
		}
		return false
	}

	if !inStack("go") {
		t.Errorf("multi-stack: expected 'go' in detected_stack, got %v", p.DetectedStack)
	}
	if !inStack("node") {
		t.Errorf("multi-stack: expected 'node' in detected_stack, got %v", p.DetectedStack)
	}
}

func TestDetector_EmptyProject_NoError(t *testing.T) {
	det := knowledge.DefaultDetector()
	emptyDir := filepath.Join(fixturesDir(t), "empty-project")

	p, err := det.Detect(context.Background(), emptyDir)
	if err != nil {
		t.Fatalf("Detect on empty project should not error: %v", err)
	}
	if len(p.DetectedStack) != 0 {
		t.Errorf("empty project: expected empty detected_stack, got %v", p.DetectedStack)
	}
}

func TestDetector_PopulatesRequiredFields(t *testing.T) {
	det := knowledge.DefaultDetector()
	goDir := filepath.Join(fixturesDir(t), "go-project")

	before := time.Now().Add(-time.Second)
	p, err := det.Detect(context.Background(), goDir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if p.SchemaVersion == "" {
		t.Error("SchemaVersion must not be empty")
	}
	if p.ScannedAt.IsZero() {
		t.Error("ScannedAt must be set")
	}
	if p.ScannedAt.Before(before) {
		t.Errorf("ScannedAt %v is before scan start %v", p.ScannedAt, before)
	}
}

func TestFSReader_RoundTrip(t *testing.T) {
	root := makeRoot(t)
	reader := knowledge.NewFSReader()

	original := knowledge.Platform{
		SchemaVersion: "1",
		ScannedAt:     time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC),
		DetectedStack: []string{"go", "node"},
		Conventions: knowledge.Conventions{
			Naming:        map[string]string{"functions": "camelCase"},
			FileStructure: "flat",
		},
		DesignFingerprint: knowledge.DesignFingerprint{
			ComponentLibrary: "shadcn",
			CSSApproach:      "tailwind",
			LayoutPatterns:   []string{"grid", "flex"},
		},
	}

	if err := reader.Write(root, original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	restored, err := reader.Read(root)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if restored.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", restored.SchemaVersion, original.SchemaVersion)
	}
	if len(restored.DetectedStack) != len(original.DetectedStack) {
		t.Errorf("DetectedStack len: got %d, want %d", len(restored.DetectedStack), len(original.DetectedStack))
	}
	if !restored.ScannedAt.Equal(original.ScannedAt) {
		t.Errorf("ScannedAt: got %v, want %v", restored.ScannedAt, original.ScannedAt)
	}
	if restored.Conventions.FileStructure != original.Conventions.FileStructure {
		t.Errorf("Conventions.FileStructure: got %q, want %q",
			restored.Conventions.FileStructure, original.Conventions.FileStructure)
	}
	if restored.DesignFingerprint.ComponentLibrary != original.DesignFingerprint.ComponentLibrary {
		t.Errorf("DesignFingerprint.ComponentLibrary: got %q, want %q",
			restored.DesignFingerprint.ComponentLibrary, original.DesignFingerprint.ComponentLibrary)
	}
}

func TestFSReader_ReadMissingFile_ReturnsError(t *testing.T) {
	root := makeRoot(t)
	reader := knowledge.NewFSReader()

	_, err := reader.Read(root)
	if err == nil {
		t.Error("expected error reading non-existent platform.yaml, got nil")
	}
}
