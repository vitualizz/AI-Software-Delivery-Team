package knowledge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
	"gopkg.in/yaml.v3"
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

// TestDetector_NoProbes verifies that a Detector with no probes returns an
// empty (non-nil) DetectedStack rather than nil.
func TestDetector_NoProbes(t *testing.T) {
	det := knowledge.NewDetector(nil)
	dir := t.TempDir()

	p, err := det.Detect(context.Background(), dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if p.DetectedStack == nil {
		t.Error("DetectedStack must not be nil for zero-probe detector")
	}
	if len(p.DetectedStack) != 0 {
		t.Errorf("expected empty DetectedStack, got %v", p.DetectedStack)
	}
}

// TestGoProbe_DetectsWithoutGoSum verifies that GoProbe detects a Go project
// when go.mod is present but go.sum is absent (only go.mod is required).
func TestGoProbe_DetectsWithoutGoSum(t *testing.T) {
	dir := t.TempDir()
	// Write only go.mod — no go.sum.
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	probe := knowledge.GoProbe()
	found, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !found {
		t.Error("GoProbe: expected true when go.mod exists but go.sum is absent")
	}
}

// TestNodeProbe_DoesNotDetectSubdirectory verifies that NodeProbe only checks
// the project root, not subdirectories.
func TestNodeProbe_DoesNotDetectSubdirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// package.json is only in a subdirectory, not the root.
	if err := os.WriteFile(filepath.Join(sub, "package.json"), []byte(`{"name":"sub"}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	probe := knowledge.NodeProbe()
	found, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if found {
		t.Error("NodeProbe: should not detect package.json in a subdirectory (root-only detection)")
	}
}

// TestFSReader_Write_CreatesKnowledgeDirectory verifies that Write creates the
// .asdt/knowledge/ directory if it does not yet exist.
func TestFSReader_Write_CreatesKnowledgeDirectory(t *testing.T) {
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir .asdt: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// knowledge/ subdirectory does NOT exist yet.
	reader := knowledge.NewFSReader()
	p := knowledge.Platform{
		SchemaVersion: "1",
		DetectedStack: []string{"go"},
	}
	// Should succeed even though .asdt/knowledge/ does not exist.
	if err := reader.Write(root, p); err != nil {
		t.Fatalf("Write with missing knowledge dir: %v", err)
	}

	// Verify the file was actually created.
	if _, err := reader.Read(root); err != nil {
		t.Fatalf("Read after Write: %v", err)
	}
}

// --- PlatformSummary / DeriveSummary tests ---

// TestDeriveSummary_TableDriven tests all derivation paths including multi-stack
// first-wins, unknown language, and correct DetectedAt propagation.
func TestDeriveSummary_TableDriven(t *testing.T) {
	ts := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name           string
		stack          []string
		wantLanguage   string
		wantPkgMgr     string
		wantTestRunner string
	}{
		{"go", []string{"go"}, "go", "go modules", "go test"},
		{"node", []string{"node"}, "node", "npm", "jest"},
		{"rust", []string{"rust"}, "rust", "cargo", "cargo test"},
		{"python", []string{"python"}, "python", "pip", "pytest"},
		{"ruby", []string{"ruby"}, "ruby", "bundler", "rspec"},
		{"multi stack first wins", []string{"go", "node"}, "go", "go modules", "go test"},
		{"empty stack unknown", []string{}, "unknown", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := knowledge.Platform{
				DetectedStack: tc.stack,
				ScannedAt:     ts,
			}
			s := knowledge.DeriveSummary(p)

			if s.PrimaryLanguage != tc.wantLanguage {
				t.Errorf("PrimaryLanguage: got %q, want %q", s.PrimaryLanguage, tc.wantLanguage)
			}
			if s.PackageManager != tc.wantPkgMgr {
				t.Errorf("PackageManager: got %q, want %q", s.PackageManager, tc.wantPkgMgr)
			}
			if s.TestRunner != tc.wantTestRunner {
				t.Errorf("TestRunner: got %q, want %q", s.TestRunner, tc.wantTestRunner)
			}
			// Stack must equal input verbatim.
			if len(s.Stack) != len(tc.stack) {
				t.Errorf("Stack len: got %d, want %d", len(s.Stack), len(tc.stack))
			}
			for i, v := range tc.stack {
				if s.Stack[i] != v {
					t.Errorf("Stack[%d]: got %q, want %q", i, s.Stack[i], v)
				}
			}
			// DetectedAt must be copied from p.ScannedAt (pure function).
			if !s.DetectedAt.Equal(ts) {
				t.Errorf("DetectedAt: got %v, want %v", s.DetectedAt, ts)
			}
		})
	}
}

// TestWriteSummary_CreatesFileAtCorrectPath verifies WriteSummary writes to
// {root}/knowledge/platform-summary.yaml, auto-creates the dir, and round-trips.
func TestWriteSummary_CreatesFileAtCorrectPath(t *testing.T) {
	root := makeRoot(t)
	reader := knowledge.NewFSReader()

	ts := time.Date(2026, 3, 15, 8, 0, 0, 0, time.UTC)
	s := knowledge.PlatformSummary{
		Stack:           []string{"go"},
		PrimaryLanguage: "go",
		PackageManager:  "go modules",
		TestRunner:      "go test",
		DetectedAt:      ts,
	}

	if err := reader.WriteSummary(root, s); err != nil {
		t.Fatalf("WriteSummary: %v", err)
	}

	// Verify the file exists at the expected path.
	expectedPath := filepath.Join(root.Path(), "knowledge", "platform-summary.yaml")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("platform-summary.yaml not created at expected path: %v", err)
	}

	// Round-trip: unmarshal and compare.
	var restored knowledge.PlatformSummary
	if err := yaml.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.PrimaryLanguage != s.PrimaryLanguage {
		t.Errorf("PrimaryLanguage: got %q, want %q", restored.PrimaryLanguage, s.PrimaryLanguage)
	}
	if restored.PackageManager != s.PackageManager {
		t.Errorf("PackageManager: got %q, want %q", restored.PackageManager, s.PackageManager)
	}
	if restored.TestRunner != s.TestRunner {
		t.Errorf("TestRunner: got %q, want %q", restored.TestRunner, s.TestRunner)
	}
	if !restored.DetectedAt.Equal(ts) {
		t.Errorf("DetectedAt: got %v, want %v", restored.DetectedAt, ts)
	}
	if len(restored.Stack) != 1 || restored.Stack[0] != "go" {
		t.Errorf("Stack: got %v, want [go]", restored.Stack)
	}
}

// TestWriteSummary_AutoCreatesKnowledgeDir verifies that WriteSummary creates the
// knowledge/ directory when it does not exist yet.
func TestWriteSummary_AutoCreatesKnowledgeDir(t *testing.T) {
	// Create a raw .asdt dir without a knowledge/ subdirectory.
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir .asdt: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	reader := knowledge.NewFSReader()
	s := knowledge.PlatformSummary{
		Stack:           []string{"node"},
		PrimaryLanguage: "node",
		PackageManager:  "npm",
		TestRunner:      "jest",
		DetectedAt:      time.Now().UTC(),
	}

	// Should not error even though .asdt/knowledge/ doesn't exist.
	if err := reader.WriteSummary(root, s); err != nil {
		t.Fatalf("WriteSummary without pre-existing knowledge dir: %v", err)
	}

	// File must exist.
	expectedPath := filepath.Join(root.Path(), "knowledge", "platform-summary.yaml")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected file at %s: %v", expectedPath, err)
	}
}

// --- FSReader WriteContext/ReadContext tests ---

// TestFSReader_WriteContext_RoundTrip writes a ProjectContext with all fields
// populated, reads it back, and asserts all fields are preserved.
func TestFSReader_WriteContext_RoundTrip(t *testing.T) {
	root := makeRoot(t)
	reader := knowledge.NewFSReader()

	ts := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	original := knowledge.ProjectContext{
		SchemaVersion: "1",
		DetectedAt:    ts,
		IsMonorepo: knowledge.ContextDetection{
			Value:      "true",
			Source:     knowledge.ContextSourceDetected,
			Confidence: knowledge.ContextConfidenceHigh,
		},
		TestRunner: knowledge.ContextDetection{
			Value:      "go test ./...",
			Source:     knowledge.ContextSourceDetected,
			Confidence: knowledge.ContextConfidenceHigh,
		},
		NamingStyle: knowledge.ContextDetection{
			Value:      "snake_case filenames, PascalCase exported symbols",
			Source:     knowledge.ContextSourceDetected,
			Confidence: knowledge.ContextConfidenceHigh,
		},
		ArchitecturalStyle: knowledge.ContextDetection{
			Value:      "hexagonal",
			Source:     knowledge.ContextSourceDetected,
			Confidence: knowledge.ContextConfidenceHigh,
		},
	}

	if err := reader.WriteContext(root, original); err != nil {
		t.Fatalf("WriteContext: %v", err)
	}

	restored, err := reader.ReadContext(root)
	if err != nil {
		t.Fatalf("ReadContext: %v", err)
	}

	if restored.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", restored.SchemaVersion, original.SchemaVersion)
	}
	if !restored.DetectedAt.Equal(original.DetectedAt) {
		t.Errorf("DetectedAt: got %v, want %v", restored.DetectedAt, original.DetectedAt)
	}
	if restored.IsMonorepo.Value != original.IsMonorepo.Value {
		t.Errorf("IsMonorepo.Value: got %q, want %q", restored.IsMonorepo.Value, original.IsMonorepo.Value)
	}
	if restored.IsMonorepo.Source != original.IsMonorepo.Source {
		t.Errorf("IsMonorepo.Source: got %q, want %q", restored.IsMonorepo.Source, original.IsMonorepo.Source)
	}
	if restored.IsMonorepo.Confidence != original.IsMonorepo.Confidence {
		t.Errorf("IsMonorepo.Confidence: got %q, want %q", restored.IsMonorepo.Confidence, original.IsMonorepo.Confidence)
	}
	if restored.TestRunner.Value != original.TestRunner.Value {
		t.Errorf("TestRunner.Value: got %q, want %q", restored.TestRunner.Value, original.TestRunner.Value)
	}
	if restored.NamingStyle.Value != original.NamingStyle.Value {
		t.Errorf("NamingStyle.Value: got %q, want %q", restored.NamingStyle.Value, original.NamingStyle.Value)
	}
	if restored.ArchitecturalStyle.Value != original.ArchitecturalStyle.Value {
		t.Errorf("ArchitecturalStyle.Value: got %q, want %q", restored.ArchitecturalStyle.Value, original.ArchitecturalStyle.Value)
	}
}

// TestFSReader_ReadContext_MissingFile_ReturnsError verifies that ReadContext
// returns a non-nil error when project-context.yaml does not exist.
func TestFSReader_ReadContext_MissingFile_ReturnsError(t *testing.T) {
	root := makeRoot(t)
	reader := knowledge.NewFSReader()

	_, err := reader.ReadContext(root)
	if err == nil {
		t.Error("expected error reading non-existent project-context.yaml, got nil")
	}
}

// TestFSReader_WriteContext_CreatesKnowledgeDirectory verifies that WriteContext
// creates the knowledge/ directory when it does not yet exist.
func TestFSReader_WriteContext_CreatesKnowledgeDirectory(t *testing.T) {
	dir := t.TempDir()
	asdt := filepath.Join(dir, ".asdt")
	if err := os.MkdirAll(asdt, 0o755); err != nil {
		t.Fatalf("mkdir .asdt: %v", err)
	}
	root, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// knowledge/ subdirectory does NOT exist yet.
	reader := knowledge.NewFSReader()
	ctx := knowledge.ProjectContext{
		SchemaVersion: "1",
		IsMonorepo: knowledge.ContextDetection{
			Value:      "false",
			Source:     knowledge.ContextSourceDetected,
			Confidence: knowledge.ContextConfidenceHigh,
		},
	}

	if err := reader.WriteContext(root, ctx); err != nil {
		t.Fatalf("WriteContext with missing knowledge dir: %v", err)
	}

	if _, err := reader.ReadContext(root); err != nil {
		t.Fatalf("ReadContext after WriteContext: %v", err)
	}
}

// TestFSReader_RoundTrip_PreservesScannedAt verifies that the ScannedAt
// timestamp and all nested fields are preserved exactly through Write+Read.
func TestFSReader_RoundTrip_PreservesScannedAt(t *testing.T) {
	root := makeRoot(t)
	reader := knowledge.NewFSReader()

	ts := time.Date(2026, 1, 15, 9, 30, 0, 0, time.UTC)
	original := knowledge.Platform{
		SchemaVersion: "1",
		ScannedAt:     ts,
		DetectedStack: []string{"go", "node"},
		Conventions: knowledge.Conventions{
			Naming:        map[string]string{"vars": "camelCase"},
			FileStructure: "hexagonal",
		},
		DesignFingerprint: knowledge.DesignFingerprint{
			ComponentLibrary: "shadcn",
			CSSApproach:      "tailwind",
			LayoutPatterns:   []string{"grid"},
		},
	}

	if err := reader.Write(root, original); err != nil {
		t.Fatalf("Write: %v", err)
	}
	restored, err := reader.Read(root)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !restored.ScannedAt.Equal(ts) {
		t.Errorf("ScannedAt: got %v, want %v", restored.ScannedAt, ts)
	}
	if restored.Conventions.Naming["vars"] != "camelCase" {
		t.Errorf("Conventions.Naming[vars]: got %q, want %q", restored.Conventions.Naming["vars"], "camelCase")
	}
	if restored.DesignFingerprint.CSSApproach != "tailwind" {
		t.Errorf("DesignFingerprint.CSSApproach: got %q, want %q", restored.DesignFingerprint.CSSApproach, "tailwind")
	}
}
