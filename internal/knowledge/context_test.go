package knowledge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/knowledge"
)

// --- MonorepoProbe tests ---

func TestMonorepoProbe_GoWork(t *testing.T) {
	dir := filepath.Join(fixturesDir(t), "monorepo-go-work")
	probe := knowledge.MonorepoProbe()

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "true" {
		t.Errorf("Value: got %q, want %q", d.Value, "true")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

func TestMonorepoProbe_PnpmWorkspace(t *testing.T) {
	dir := filepath.Join(fixturesDir(t), "monorepo-pnpm")
	probe := knowledge.MonorepoProbe()

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "true" {
		t.Errorf("Value: got %q, want %q", d.Value, "true")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

func TestMonorepoProbe_NoMarker(t *testing.T) {
	dir := t.TempDir()
	probe := knowledge.MonorepoProbe()

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "false" {
		t.Errorf("Value: got %q, want %q", d.Value, "false")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

// --- TestRunnerProbe tests ---

func TestTestRunnerProbe_GoMakefile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("test:\n\tgo test ./...\n"), 0o644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	probe := knowledge.TestRunnerProbe("go")
	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "make test" {
		t.Errorf("Value: got %q, want %q", d.Value, "make test")
	}
	if d.Confidence != knowledge.ContextConfidenceMedium {
		t.Errorf("Confidence: got %q, want medium", d.Confidence)
	}
}

func TestTestRunnerProbe_GoMod(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	probe := knowledge.TestRunnerProbe("go")
	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "go test ./..." {
		t.Errorf("Value: got %q, want %q", d.Value, "go test ./...")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

func TestTestRunnerProbe_NodePackageJson(t *testing.T) {
	dir := t.TempDir()
	pkgJSON := `{"name":"app","scripts":{"test":"jest --runInBand"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	probe := knowledge.TestRunnerProbe("node")
	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "jest --runInBand" {
		t.Errorf("Value: got %q, want %q", d.Value, "jest --runInBand")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

func TestTestRunnerProbe_NodeJest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jest.config.js"), []byte("module.exports = {};\n"), 0o644); err != nil {
		t.Fatalf("write jest.config.js: %v", err)
	}

	probe := knowledge.TestRunnerProbe("node")
	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "jest" {
		t.Errorf("Value: got %q, want %q", d.Value, "jest")
	}
	if d.Confidence != knowledge.ContextConfidenceMedium {
		t.Errorf("Confidence: got %q, want medium", d.Confidence)
	}
}

func TestTestRunnerProbe_Unknown(t *testing.T) {
	dir := t.TempDir()
	probe := knowledge.TestRunnerProbe("unknown")

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "unknown" {
		t.Errorf("Value: got %q, want %q", d.Value, "unknown")
	}
	if d.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", d.Confidence)
	}
}

// --- NamingStyleDetector tests ---

func TestNamingStyleDetector_GoProject(t *testing.T) {
	// Synthetic TempDir: go-project fixture has no .go files; write synthetic ones.
	dir := t.TempDir()
	goFiles := []struct {
		name    string
		content string
	}{
		{"a.go", "package foo\n\nfunc FooBar() {}\ntype BazQux struct{}\nvar GlobalVar int\n"},
		{"b.go", "package foo\n\nfunc HandleRequest() {}\ntype ServiceImpl struct{}\n"},
		{"c.go", "package foo\n\nconst MaxRetries = 3\nfunc ProcessOrder() {}\n"},
		{"d.go", "package foo\n\nfunc CreateUser() {}\ntype UserRepository struct{}\n"},
		{"e.go", "package foo\n\nfunc GetByID() {}\nvar DefaultTimeout int\n"},
		{"f.go", "package foo\n\nfunc RunMigration() {}\ntype EventHandler struct{}\n"},
		{"g.go", "package foo\n\nfunc ValidateInput() {}\nconst SchemaVersion = \"1\"\n"},
		{"h.go", "package foo\n\nfunc BuildReport() {}\ntype ReportWriter struct{}\n"},
	}
	for _, f := range goFiles {
		if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.content), 0o644); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
	}

	probe := knowledge.NamingStyleDetector("go", 8)
	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value == "unknown" || d.Value == "" {
		t.Errorf("Value: got %q, expected non-unknown for Go project with PascalCase exports", d.Value)
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high for dominant PascalCase", d.Confidence)
	}
}

func TestNamingStyleDetector_EmptyProject(t *testing.T) {
	dir := filepath.Join(fixturesDir(t), "empty-naming")
	probe := knowledge.NamingStyleDetector("go", 8)

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "unknown" {
		t.Errorf("Value: got %q, want %q", d.Value, "unknown")
	}
	if d.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", d.Confidence)
	}
}

// --- ArchitecturalStyleDetector tests ---

func TestArchitecturalStyleDetector_Hexagonal(t *testing.T) {
	dir := filepath.Join(fixturesDir(t), "hexagonal-go")
	probe := knowledge.ArchitecturalStyleDetector()

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "hexagonal" {
		t.Errorf("Value: got %q, want %q", d.Value, "hexagonal")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

func TestArchitecturalStyleDetector_MVC(t *testing.T) {
	dir := filepath.Join(fixturesDir(t), "mvc-node")
	probe := knowledge.ArchitecturalStyleDetector()

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "mvc" {
		t.Errorf("Value: got %q, want %q", d.Value, "mvc")
	}
	if d.Confidence != knowledge.ContextConfidenceHigh {
		t.Errorf("Confidence: got %q, want high", d.Confidence)
	}
}

func TestArchitecturalStyleDetector_Unknown(t *testing.T) {
	dir := t.TempDir()
	probe := knowledge.ArchitecturalStyleDetector()

	d, err := probe.Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Value != "unknown" {
		t.Errorf("Value: got %q, want %q", d.Value, "unknown")
	}
	if d.Confidence != knowledge.ContextConfidenceLow {
		t.Errorf("Confidence: got %q, want low", d.Confidence)
	}
}

// --- ContextDetector tests ---

// errorProbe is a ContextProbe that always returns an error (for non-fatal test).
type errorProbe struct{}

func (p *errorProbe) Field() string { return "is_monorepo" }
func (p *errorProbe) Detect(_ string) (knowledge.ContextDetection, error) {
	return knowledge.ContextDetection{}, os.ErrPermission
}

func TestContextDetector_NonFatalProbeError(t *testing.T) {
	det := knowledge.NewContextDetector([]knowledge.ContextProbe{&errorProbe{}})
	dir := t.TempDir()

	ctx, err := det.DetectContext(dir, knowledge.DetectConfig{})
	if err != nil {
		t.Fatalf("DetectContext must not return error even when probe fails: %v", err)
	}
	// IsMonorepo should remain zero-value (probe was skipped).
	if ctx.IsMonorepo.Value != "" {
		t.Errorf("IsMonorepo.Value: got %q, want empty (probe skipped)", ctx.IsMonorepo.Value)
	}
}

func TestContextDetector_RoundTrip(t *testing.T) {
	dir := filepath.Join(fixturesDir(t), "hexagonal-go")
	det := knowledge.DefaultContextDetector("go")

	ctx, err := det.DetectContext(dir, knowledge.DetectConfig{PrimaryLanguage: "go"})
	if err != nil {
		t.Fatalf("DetectContext: %v", err)
	}
	if ctx.DetectedAt.IsZero() {
		t.Error("DetectedAt must not be zero")
	}
	if ctx.SchemaVersion != knowledge.ContextSchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", ctx.SchemaVersion, knowledge.ContextSchemaVersion)
	}
	// hexagonal-go fixture has cmd/ and internal/ — must detect hexagonal.
	if ctx.ArchitecturalStyle.Value != "hexagonal" {
		t.Errorf("ArchitecturalStyle.Value: got %q, want %q", ctx.ArchitecturalStyle.Value, "hexagonal")
	}
}
