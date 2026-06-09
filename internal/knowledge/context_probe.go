package knowledge

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ContextProbe is the single-field detection unit for project context.
// Each probe detects one aspect of project structure and returns a
// ContextDetection with the detected value and its provenance.
type ContextProbe interface {
	// Field returns the YAML key name that this probe populates in ProjectContext.
	// Used by ContextDetector to route detection results.
	Field() string

	// Detect examines projectRoot and returns a ContextDetection for this probe's
	// field. Errors are non-fatal — callers must skip this probe and continue.
	Detect(projectRoot string) (ContextDetection, error)
}

// --- MonorepoProbe ---

type monorepoProbe struct{}

// MonorepoProbe returns a ContextProbe that detects whether the project root
// is a monorepo by checking for go.work or pnpm-workspace.yaml.
func MonorepoProbe() ContextProbe {
	return &monorepoProbe{}
}

func (p *monorepoProbe) Field() string { return "is_monorepo" }

func (p *monorepoProbe) Detect(projectRoot string) (ContextDetection, error) {
	markers := []string{"go.work", "pnpm-workspace.yaml"}
	for _, marker := range markers {
		_, err := os.Stat(filepath.Join(projectRoot, marker))
		if err == nil {
			return ContextDetection{
				Value:      "true",
				Source:     ContextSourceDetected,
				Confidence: ContextConfidenceHigh,
			}, nil
		}
		if !os.IsNotExist(err) {
			return ContextDetection{}, err
		}
	}
	return ContextDetection{
		Value:      "false",
		Source:     ContextSourceDetected,
		Confidence: ContextConfidenceHigh,
	}, nil
}

// --- TestRunnerProbe ---

type testRunnerProbe struct {
	primaryLang string
}

// TestRunnerProbe returns a ContextProbe that detects the primary test runner
// from marker files. Detection is language-specific.
func TestRunnerProbe(primaryLang string) ContextProbe {
	return &testRunnerProbe{primaryLang: primaryLang}
}

func (p *testRunnerProbe) Field() string { return "test_runner" }

func (p *testRunnerProbe) Detect(projectRoot string) (ContextDetection, error) {
	switch p.primaryLang {
	case "go":
		return p.detectGo(projectRoot)
	case "node":
		return p.detectNode(projectRoot)
	default:
		return ContextDetection{
			Value:      "unknown",
			Source:     ContextSourceInferred,
			Confidence: ContextConfidenceLow,
		}, nil
	}
}

func (p *testRunnerProbe) detectGo(projectRoot string) (ContextDetection, error) {
	// Check Makefile for "go test" — higher priority when found.
	makefilePath := filepath.Join(projectRoot, "Makefile")
	data, err := os.ReadFile(makefilePath)
	if err == nil {
		if strings.Contains(string(data), "go test") {
			return ContextDetection{
				Value:      "make test",
				Source:     ContextSourceDetected,
				Confidence: ContextConfidenceMedium,
			}, nil
		}
	}

	// Fallback: go.mod presence → standard go test command.
	_, err = os.Stat(filepath.Join(projectRoot, "go.mod"))
	if err == nil {
		return ContextDetection{
			Value:      "go test ./...",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceHigh,
		}, nil
	}
	if !os.IsNotExist(err) {
		return ContextDetection{}, err
	}

	return ContextDetection{
		Value:      "unknown",
		Source:     ContextSourceInferred,
		Confidence: ContextConfidenceLow,
	}, nil
}

func (p *testRunnerProbe) detectNode(projectRoot string) (ContextDetection, error) {
	// Check package.json scripts.test first.
	pkgPath := filepath.Join(projectRoot, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if jsonErr := json.Unmarshal(data, &pkg); jsonErr == nil {
			if testCmd, ok := pkg.Scripts["test"]; ok && testCmd != "" {
				return ContextDetection{
					Value:      testCmd,
					Source:     ContextSourceDetected,
					Confidence: ContextConfidenceHigh,
				}, nil
			}
		}
	}

	// Check jest.config.js.
	_, err = os.Stat(filepath.Join(projectRoot, "jest.config.js"))
	if err == nil {
		return ContextDetection{
			Value:      "jest",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
		}, nil
	}

	// Check vitest.config.ts.
	_, err = os.Stat(filepath.Join(projectRoot, "vitest.config.ts"))
	if err == nil {
		return ContextDetection{
			Value:      "vitest",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
		}, nil
	}

	return ContextDetection{
		Value:      "unknown",
		Source:     ContextSourceInferred,
		Confidence: ContextConfidenceLow,
	}, nil
}

// --- NamingStyleDetector ---

type namingStyleDetector struct {
	primaryLang string
	sampleN     int
}

// NamingStyleDetector returns a ContextProbe that samples up to n source files
// and classifies the dominant exported-identifier casing style.
// If n is 0, NamingStyleSampleSize is used.
func NamingStyleDetector(primaryLang string, n int) ContextProbe {
	if n <= 0 {
		n = NamingStyleSampleSize
	}
	return &namingStyleDetector{primaryLang: primaryLang, sampleN: n}
}

func (p *namingStyleDetector) Field() string { return "naming_style" }

// maxFileSizeBytes is the file size guard for source file sampling.
const maxFileSizeBytes = 64 * 1024

// goExportedPattern matches top-level exported Go declarations.
var goExportedPattern = regexp.MustCompile(`^func [A-Z]|^type [A-Z]|^var [A-Z]|^const [A-Z]`)

// goUnexportedPattern matches top-level unexported Go declarations.
var goUnexportedPattern = regexp.MustCompile(`^func [a-z]|^type [a-z]|^var [a-z]|^const [a-z]`)

// nodeExportedPascalPattern matches exported PascalCase identifiers in TypeScript/JS.
var nodeExportedPascalPattern = regexp.MustCompile(`^export (function|class|const|interface) [A-Z]`)

// nodeExportedCamelPattern matches exported camelCase identifiers in TypeScript/JS.
var nodeExportedCamelPattern = regexp.MustCompile(`^export (function|const) [a-z]`)

func (p *namingStyleDetector) Detect(projectRoot string) (ContextDetection, error) {
	extensions := p.extensionsForLang()
	if len(extensions) == 0 {
		return ContextDetection{
			Value:      "unknown",
			Source:     ContextSourceInferred,
			Confidence: ContextConfidenceLow,
		}, nil
	}

	files, err := p.collectFiles(projectRoot, extensions, p.sampleN)
	if err != nil {
		return ContextDetection{}, err
	}

	if len(files) == 0 {
		return ContextDetection{
			Value:      "unknown",
			Source:     ContextSourceInferred,
			Confidence: ContextConfidenceLow,
		}, nil
	}

	pascalCount, camelCount := 0, 0
	for _, f := range files {
		hasPascal, hasCamel := p.classifyFile(f)
		if hasPascal {
			pascalCount++
		}
		if hasCamel {
			camelCount++
		}
	}

	total := len(files)
	return p.buildResult(pascalCount, camelCount, total), nil
}

func (p *namingStyleDetector) extensionsForLang() []string {
	switch p.primaryLang {
	case "go":
		return []string{".go"}
	case "node":
		return []string{".ts", ".tsx"}
	default:
		return nil
	}
}

func (p *namingStyleDetector) collectFiles(root string, exts []string, limit int) ([]string, error) {
	var files []string
	maxDepth := 3

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if len(files) >= limit {
			return fs.SkipAll
		}

		// Enforce depth limit.
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		depth := strings.Count(rel, string(filepath.Separator))
		if d.IsDir() {
			if depth >= maxDepth {
				return fs.SkipDir
			}
			return nil
		}

		// Check extension.
		ext := strings.ToLower(filepath.Ext(path))
		matched := false
		for _, e := range exts {
			if ext == e {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		// Size guard.
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if info.Size() > maxFileSizeBytes {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (p *namingStyleDetector) classifyFile(path string) (hasPascal, hasCamel bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if p.primaryLang == "go" {
			if goExportedPattern.MatchString(line) {
				hasPascal = true
			}
			if goUnexportedPattern.MatchString(line) {
				hasCamel = true
			}
		} else {
			if nodeExportedPascalPattern.MatchString(line) {
				hasPascal = true
			}
			if nodeExportedCamelPattern.MatchString(line) {
				hasCamel = true
			}
		}
		if hasPascal && hasCamel {
			break
		}
	}
	return hasPascal, hasCamel
}

func (p *namingStyleDetector) buildResult(pascalCount, camelCount, total int) ContextDetection {
	dominant := pascalCount
	value := "PascalCase"
	if p.primaryLang == "go" && pascalCount > camelCount {
		// Go projects: standard convention confirmed.
		value = "snake_case filenames, PascalCase exported symbols"
	} else if camelCount > pascalCount {
		dominant = camelCount
		value = "camelCase"
	} else if pascalCount == camelCount && pascalCount == 0 {
		return ContextDetection{
			Value:      "unknown",
			Source:     ContextSourceInferred,
			Confidence: ContextConfidenceLow,
		}
	} else if pascalCount == camelCount {
		return ContextDetection{
			Value:      "mixed",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceLow,
		}
	}

	confidence := confidenceFromCount(dominant, total)
	return ContextDetection{
		Value:      value,
		Source:     ContextSourceDetected,
		Confidence: confidence,
	}
}

// confidenceFromCount applies the majority confidence rule:
// ≥6/8 consistent → high; 4-5/8 → medium; <4/8 → low.
func confidenceFromCount(count, total int) ContextConfidence {
	if total == 0 {
		return ContextConfidenceLow
	}
	ratio := float64(count) / float64(total)
	switch {
	case ratio >= 0.75:
		return ContextConfidenceHigh
	case ratio >= 0.5:
		return ContextConfidenceMedium
	default:
		return ContextConfidenceLow
	}
}

// --- ArchitecturalStyleDetector ---

type architecturalStyleDetector struct{}

// ArchitecturalStyleDetector returns a ContextProbe that detects the dominant
// architectural pattern from top-level (Tier 1) directory layout hints.
func ArchitecturalStyleDetector() ContextProbe {
	return &architecturalStyleDetector{}
}

func (p *architecturalStyleDetector) Field() string { return "architectural_style" }

func (p *architecturalStyleDetector) Detect(projectRoot string) (ContextDetection, error) {
	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return ContextDetection{}, err
	}

	dirs := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			dirs[e.Name()] = true
		}
	}

	// Hexagonal: cmd/ and internal/ at root level.
	if dirs["cmd"] && dirs["internal"] {
		return ContextDetection{
			Value:      "hexagonal",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceHigh,
		}, nil
	}

	// For node projects: check one level deeper under src/.
	if dirs["src"] {
		srcDirs, srcErr := readDirNames(filepath.Join(projectRoot, "src"))
		if srcErr == nil {
			srcSet := make(map[string]bool)
			for _, name := range srcDirs {
				srcSet[name] = true
			}
			if srcSet["controllers"] && srcSet["models"] && srcSet["views"] {
				return ContextDetection{
					Value:      "mvc",
					Source:     ContextSourceDetected,
					Confidence: ContextConfidenceHigh,
				}, nil
			}
			if srcSet["features"] || srcSet["modules"] {
				return ContextDetection{
					Value:      "modular",
					Source:     ContextSourceDetected,
					Confidence: ContextConfidenceMedium,
				}, nil
			}
			// src/ exists but no sub-pattern matched — layered.
			return ContextDetection{
				Value:      "layered",
				Source:     ContextSourceDetected,
				Confidence: ContextConfidenceMedium,
			}, nil
		}
	}

	// lib/ without src/ → layered.
	if dirs["lib"] {
		return ContextDetection{
			Value:      "layered",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
		}, nil
	}

	return ContextDetection{
		Value:      "unknown",
		Source:     ContextSourceInferred,
		Confidence: ContextConfidenceLow,
	}, nil
}

// readDirNames returns subdirectory names under dir.
func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
