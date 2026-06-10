package knowledge_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
)

// --- helpers ---

// mkDir creates a directory inside base and returns its path.
func mkDir(t *testing.T, base string, parts ...string) string {
	t.Helper()
	full := filepath.Join(append([]string{base}, parts...)...)
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatalf("mkDir %s: %v", full, err)
	}
	return full
}

// mkFile writes content to base/parts... and returns the path.
func mkFile(t *testing.T, content, base string, parts ...string) string {
	t.Helper()
	full := filepath.Join(append([]string{base}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkFile mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("mkFile write %s: %v", full, err)
	}
	return full
}

// --- MonorepoProbe ---

func TestMonorepoProbe_GoWork(t *testing.T) {
	dir := t.TempDir()
	mkFile(t, "", dir, "go.work")

	p := knowledge.MonorepoProbe()
	if p.Name() != "is_monorepo" {
		t.Errorf("Name: got %q, want %q", p.Name(), "is_monorepo")
	}
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "true" {
		t.Errorf("Value: got %q, want %q", det.Value, "true")
	}
	if det.Source != knowledge.ContextSourceDetected {
		t.Errorf("Source: got %q, want detected", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestMonorepoProbe_PnpmWorkspace(t *testing.T) {
	dir := t.TempDir()
	mkFile(t, "packages:\n  - packages/*\n", dir, "pnpm-workspace.yaml")

	p := knowledge.MonorepoProbe()
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "true" {
		t.Errorf("Value: got %q, want true", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestMonorepoProbe_Neither(t *testing.T) {
	dir := t.TempDir()

	p := knowledge.MonorepoProbe()
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "false" {
		t.Errorf("Value: got %q, want false", det.Value)
	}
	if det.Source != knowledge.ContextSourceDetected {
		t.Errorf("Source: got %q, want detected", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

// --- TestRunnerProbe (go) ---

func TestTestRunnerProbe_Go_MakefileWithGoTest(t *testing.T) {
	dir := t.TempDir()
	mkFile(t, "test:\n\tgo test ./...\n", dir, "Makefile")
	mkFile(t, "module example.com/test\ngo 1.21\n", dir, "go.mod")

	p := knowledge.TestRunnerProbe("go")
	if p.Name() != "test_runner" {
		t.Errorf("Name: got %q, want test_runner", p.Name())
	}
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "make test" {
		t.Errorf("Value: got %q, want make test", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceMedium {
		t.Errorf("Confidence: got %q, want medium", det.Confidence)
	}
}

func TestTestRunnerProbe_Go_GoModOnly(t *testing.T) {
	dir := t.TempDir()
	mkFile(t, "module example.com/test\ngo 1.21\n", dir, "go.mod")

	p := knowledge.TestRunnerProbe("go")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "go test ./..." {
		t.Errorf("Value: got %q, want go test ./...", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestTestRunnerProbe_Go_Neither(t *testing.T) {
	dir := t.TempDir()

	p := knowledge.TestRunnerProbe("go")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "unknown" {
		t.Errorf("Value: got %q, want unknown", det.Value)
	}
	if det.Source != knowledge.ContextSourceInferred {
		t.Errorf("Source: got %q, want inferred", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", det.Confidence)
	}
}

// --- TestRunnerProbe (node) ---

func TestTestRunnerProbe_Node_PackageJSONScripts(t *testing.T) {
	dir := t.TempDir()
	mkFile(t, `{"scripts":{"test":"jest --coverage"}}`, dir, "package.json")

	p := knowledge.TestRunnerProbe("node")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "jest --coverage" {
		t.Errorf("Value: got %q, want jest --coverage", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestTestRunnerProbe_Node_JestConfig(t *testing.T) {
	dir := t.TempDir()
	mkFile(t, `module.exports = {}`, dir, "jest.config.js")

	p := knowledge.TestRunnerProbe("node")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "jest" {
		t.Errorf("Value: got %q, want jest", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceMedium {
		t.Errorf("Confidence: got %q, want medium", det.Confidence)
	}
}

func TestTestRunnerProbe_Node_None(t *testing.T) {
	dir := t.TempDir()

	p := knowledge.TestRunnerProbe("node")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "unknown" {
		t.Errorf("Value: got %q, want unknown", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", det.Confidence)
	}
}

// --- NamingStyleProbe ---

// writeGoFiles creates n Go source files with only PascalCase exported symbols.
func writePascalGoFiles(t *testing.T, dir string, n int) {
	t.Helper()
	for i := range n {
		content := "package main\n\nfunc Main() {}\ntype Foo struct{}\n"
		mkFile(t, content, dir, filepath.Join("pkg", string(rune('a'+i))+"_thing.go"))
	}
}

func TestNamingStyleProbe_Go_PascalDominance_High(t *testing.T) {
	dir := t.TempDir()
	// Write 8 files — all PascalCase only → ratio = 1.0 → high.
	writePascalGoFiles(t, dir, 8)

	p := knowledge.NamingStyleProbe("go")
	if p.Name() != "naming_style" {
		t.Errorf("Name: got %q, want naming_style", p.Name())
	}
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "snake_case filenames, PascalCase exported symbols" {
		t.Errorf("Value: got %q", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestNamingStyleProbe_Go_Mixed_Low(t *testing.T) {
	dir := t.TempDir()
	// Half PascalCase-only, half both pascal and camel → only 4/8 in pascal bucket → low.
	for i := range 4 {
		// Pascal only.
		content := "package main\n\nfunc Exported() {}\ntype Bar struct{}\n"
		mkFile(t, content, dir, filepath.Join("pkg", string(rune('a'+i))+"_pub.go"))
	}
	for i := range 4 {
		// Both pascal and camel → counted in neither bucket.
		content := "package main\n\nfunc Exported() {}\nfunc unexported() {}\n"
		mkFile(t, content, dir, filepath.Join("pkg", string(rune('e'+i))+"_mix.go"))
	}

	p := knowledge.NamingStyleProbe("go")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	// 4/8 = 0.50 → medium confidence but value still reflects pascal dominance.
	// The probe counts only pascal-exclusive files in the pascal bucket.
	// 4 pascal-only out of 8 total = 0.50 → medium.
	if det.Confidence != knowledge.ContextConfidenceMedium {
		t.Errorf("Confidence: got %q, want medium (4/8 ratio)", det.Confidence)
	}
}

func TestNamingStyleProbe_Node_PascalDominance_High(t *testing.T) {
	dir := t.TempDir()
	// 8 .ts files — all PascalCase-only exports → ratio = 1.0 → high.
	for i := range 8 {
		content := "export class Service {}\nexport interface Config {}\n"
		mkFile(t, content, dir, filepath.Join("src", string(rune('a'+i))+"-svc.ts"))
	}

	p := knowledge.NamingStyleProbe("node")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "PascalCase exported symbols" {
		t.Errorf("Value: got %q, want PascalCase exported symbols", det.Value)
	}
	if det.Source != knowledge.ContextSourceDetected {
		t.Errorf("Source: got %q, want detected", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestNamingStyleProbe_Go_NoDominance_UnknownInferred(t *testing.T) {
	dir := t.TempDir()
	// All files mix pascal and camel → counted in neither bucket → 0/8 ratio.
	for i := range 8 {
		content := "package main\n\nfunc Exported() {}\nfunc unexported() {}\n"
		mkFile(t, content, dir, filepath.Join("pkg", string(rune('a'+i))+"_mix.go"))
	}

	p := knowledge.NamingStyleProbe("go")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "unknown" {
		t.Errorf("Value: got %q, want unknown", det.Value)
	}
	if det.Source != knowledge.ContextSourceInferred {
		t.Errorf("Source: got %q, want inferred (unknown is absence of signal)", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", det.Confidence)
	}
}

func TestNamingStyleProbe_Go_NoFiles_Unknown(t *testing.T) {
	dir := t.TempDir()
	// No .go files at all.

	p := knowledge.NamingStyleProbe("go")
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "unknown" {
		t.Errorf("Value: got %q, want unknown", det.Value)
	}
	if det.Source != knowledge.ContextSourceInferred {
		t.Errorf("Source: got %q, want inferred", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", det.Confidence)
	}
}

// --- ArchitecturalStyleProbe ---

func TestArchitecturalStyleProbe_CmdAndInternal_Hexagonal(t *testing.T) {
	dir := t.TempDir()
	mkDir(t, dir, "cmd")
	mkDir(t, dir, "internal")

	p := knowledge.ArchitecturalStyleProbe()
	if p.Name() != "architectural_style" {
		t.Errorf("Name: got %q, want architectural_style", p.Name())
	}
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "hexagonal" {
		t.Errorf("Value: got %q, want hexagonal", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestArchitecturalStyleProbe_SrcMVC(t *testing.T) {
	dir := t.TempDir()
	mkDir(t, dir, "src", "controllers")
	mkDir(t, dir, "src", "models")
	mkDir(t, dir, "src", "views")

	p := knowledge.ArchitecturalStyleProbe()
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "mvc" {
		t.Errorf("Value: got %q, want mvc", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", det.Confidence)
	}
}

func TestArchitecturalStyleProbe_SrcFeatures_Modular(t *testing.T) {
	dir := t.TempDir()
	mkDir(t, dir, "src", "features")

	p := knowledge.ArchitecturalStyleProbe()
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "modular" {
		t.Errorf("Value: got %q, want modular", det.Value)
	}
	if det.Confidence != knowledge.ContextConfidenceMedium {
		t.Errorf("Confidence: got %q, want medium", det.Confidence)
	}
}

func TestArchitecturalStyleProbe_None_Unknown(t *testing.T) {
	dir := t.TempDir()

	p := knowledge.ArchitecturalStyleProbe()
	det, err := p.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if det.Value != "unknown" {
		t.Errorf("Value: got %q, want unknown", det.Value)
	}
	if det.Source != knowledge.ContextSourceInferred {
		t.Errorf("Source: got %q, want inferred", det.Source)
	}
	if det.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", det.Confidence)
	}
}

// --- DefaultContextDetector ---

func TestDefaultContextDetector_AllProbes(t *testing.T) {
	dir := t.TempDir()
	// Populate a minimal Go project layout.
	mkFile(t, "module example.com/test\ngo 1.21\n", dir, "go.mod")
	mkDir(t, dir, "cmd")
	mkDir(t, dir, "internal")

	det := knowledge.DefaultContextDetector("go")
	ctx, err := det.DetectContext(dir, knowledge.DetectConfig{PrimaryLanguage: "go"})
	if err != nil {
		t.Fatalf("DetectContext: %v", err)
	}

	if ctx.SchemaVersion == "" {
		t.Error("SchemaVersion must not be empty")
	}
	if ctx.DetectedAt.IsZero() {
		t.Error("DetectedAt must be set")
	}
	if ctx.IsMonorepo.Value != "false" {
		t.Errorf("IsMonorepo: got %q, want false", ctx.IsMonorepo.Value)
	}
	if ctx.TestRunner.Value != "go test ./..." {
		t.Errorf("TestRunner: got %q, want go test ./...", ctx.TestRunner.Value)
	}
	if ctx.ArchitecturalStyle.Value != "hexagonal" {
		t.Errorf("ArchitecturalStyle: got %q, want hexagonal", ctx.ArchitecturalStyle.Value)
	}
}

// errorProbe is a ContextProbe that always returns an error.
type errorProbe struct{ name string }

func (e *errorProbe) Name() string { return e.name }
func (e *errorProbe) Detect(_ string) (knowledge.ContextDetection, error) {
	return knowledge.ContextDetection{}, errors.New("injected probe error")
}

func TestDefaultContextDetector_NonFatalProbeError(t *testing.T) {
	dir := t.TempDir()
	// Wire one good probe and one failing probe.
	probes := []knowledge.ContextProbe{
		knowledge.MonorepoProbe(),
		&errorProbe{name: "test_runner"},
	}

	det := knowledge.NewContextDetector(probes)
	ctx, err := det.DetectContext(dir, knowledge.DetectConfig{PrimaryLanguage: "go"})
	if err != nil {
		t.Fatalf("DetectContext must not return error even when probe fails: %v", err)
	}

	// MonorepoProbe ran fine.
	if ctx.IsMonorepo.Value != "false" {
		t.Errorf("IsMonorepo: got %q, want false", ctx.IsMonorepo.Value)
	}
	// Failing probe → fallback detection.
	if ctx.TestRunner.Source != knowledge.ContextSourceInferred {
		t.Errorf("TestRunner.Source: got %q, want inferred (fallback)", ctx.TestRunner.Source)
	}
	if ctx.TestRunner.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("TestRunner.Confidence: got %q, want low (fallback)", ctx.TestRunner.Confidence)
	}
}
