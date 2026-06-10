package knowledge

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ContextProbe examines a project root and returns a single ContextDetection
// for one field of ProjectContext.
type ContextProbe interface {
	// Name returns the field identifier (e.g. "is_monorepo", "test_runner").
	Name() string

	// Detect inspects projectRoot and returns a ContextDetection for this probe.
	Detect(projectRoot string) (ContextDetection, error)
}

// Package-level regexp vars — compiled once at startup, never per-call.
var (
	goPascalRe   = regexp.MustCompile(`^func [A-Z]|^type [A-Z]|^var [A-Z]|^const [A-Z]`)
	goCamelRe    = regexp.MustCompile(`^func [a-z]|^type [a-z]|^var [a-z]|^const [a-z]`)
	nodePascalRe = regexp.MustCompile(`^export (function|class|const|interface) [A-Z]`)
	nodeCamelRe  = regexp.MustCompile(`^export (function|const) [a-z]`)
)

// confidenceFromRatio maps a match ratio to a ContextConfidence level.
// ≥0.75 → high, 0.50–0.74 → medium, <0.50 → low.
func confidenceFromRatio(match, total int) ContextConfidence {
	if total == 0 {
		return ContextConfidenceLow
	}
	ratio := float64(match) / float64(total)
	switch {
	case ratio >= 0.75:
		return ContextConfidenceHigh
	case ratio >= 0.50:
		return ContextConfidenceMedium
	default:
		return ContextConfidenceLow
	}
}

// --- MonorepoProbe ---

type monorepoProbe struct{}

// MonorepoProbe returns a ContextProbe that detects monorepo setups.
func MonorepoProbe() ContextProbe {
	return &monorepoProbe{}
}

// Name returns the field identifier for MonorepoProbe.
func (p *monorepoProbe) Name() string { return "is_monorepo" }

// Detect checks for go.work or pnpm-workspace.yaml to determine monorepo status.
// A missing file is not an error — only real I/O errors propagate.
func (p *monorepoProbe) Detect(projectRoot string) (ContextDetection, error) {
	markers := []string{"go.work", "pnpm-workspace.yaml"}
	for _, m := range markers {
		_, err := os.Stat(filepath.Join(projectRoot, m))
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
	lang string
}

// TestRunnerProbe returns a ContextProbe that detects the test runner for lang.
func TestRunnerProbe(lang string) ContextProbe {
	return &testRunnerProbe{lang: lang}
}

// Name returns the field identifier for TestRunnerProbe.
func (p *testRunnerProbe) Name() string { return "test_runner" }

// Detect infers the test runner from project files.
func (p *testRunnerProbe) Detect(projectRoot string) (ContextDetection, error) {
	switch p.lang {
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

func (p *testRunnerProbe) detectGo(root string) (ContextDetection, error) {
	// Check Makefile first.
	makeData, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err == nil && bytes.Contains(makeData, []byte("go test")) {
		return ContextDetection{
			Value:      "make test",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
		}, nil
	}

	// Fall through to go.mod check.
	_, err = os.Stat(filepath.Join(root, "go.mod"))
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

func (p *testRunnerProbe) detectNode(root string) (ContextDetection, error) {
	// Try package.json scripts.test first.
	pkgData, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if jsonErr := json.Unmarshal(pkgData, &pkg); jsonErr == nil {
			if v, ok := pkg.Scripts["test"]; ok && v != "" {
				return ContextDetection{
					Value:      v,
					Source:     ContextSourceDetected,
					Confidence: ContextConfidenceHigh,
				}, nil
			}
		}
		// JSON parse error: fall through to next check (non-fatal).
	}

	// jest.config.js
	_, err = os.Stat(filepath.Join(root, "jest.config.js"))
	if err == nil {
		return ContextDetection{
			Value:      "jest",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
		}, nil
	}
	if !os.IsNotExist(err) {
		return ContextDetection{}, err
	}

	// vitest.config.ts
	_, err = os.Stat(filepath.Join(root, "vitest.config.ts"))
	if err == nil {
		return ContextDetection{
			Value:      "vitest",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
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

// --- NamingStyleProbe ---

type namingStyleProbe struct {
	lang string
}

// NamingStyleProbe returns a ContextProbe that samples source files and infers
// the dominant naming style.
func NamingStyleProbe(lang string) ContextProbe {
	return &namingStyleProbe{lang: lang}
}

// Name returns the field identifier for NamingStyleProbe.
func (p *namingStyleProbe) Name() string { return "naming_style" }

// Detect walks the project tree (depth <= 3, max NamingStyleSampleSize files)
// and determines dominant naming style from regexp matching.
func (p *namingStyleProbe) Detect(projectRoot string) (ContextDetection, error) {
	exts := extensionsForLang(p.lang)
	if len(exts) == 0 {
		return ContextDetection{
			Value:      "unknown",
			Source:     ContextSourceInferred,
			Confidence: ContextConfidenceLow,
		}, nil
	}

	var sampled []string
	rootDepth := strings.Count(filepath.Clean(projectRoot), string(os.PathSeparator))

	walkErr := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			// Prune at depth > 3.
			depth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - rootDepth
			if depth > 3 {
				return filepath.SkipDir
			}
			return nil
		}
		if len(sampled) >= NamingStyleSampleSize {
			return filepath.SkipAll
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, e := range exts {
			if ext == e {
				// Skip files > 64 KB.
				info, statErr := os.Stat(path)
				if statErr != nil || info.Size() > 64*1024 {
					return nil
				}
				sampled = append(sampled, path)
				return nil
			}
		}
		return nil
	})
	if walkErr != nil {
		return ContextDetection{}, walkErr
	}

	if len(sampled) == 0 {
		return ContextDetection{
			Value:      "unknown",
			Source:     ContextSourceInferred,
			Confidence: ContextConfidenceLow,
		}, nil
	}

	pascalRe, camelRe := regexpsForLang(p.lang)
	pascalCount := 0
	for _, f := range sampled {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		hasPascal := false
		hasCamel := false
		for _, line := range strings.Split(string(data), "\n") {
			if pascalRe.MatchString(line) {
				hasPascal = true
			}
			if camelRe.MatchString(line) {
				hasCamel = true
			}
		}
		// Count in pascal bucket only when it has pascal but not camel.
		if hasPascal && !hasCamel {
			pascalCount++
		}
	}

	total := len(sampled)
	conf := confidenceFromRatio(pascalCount, total)

	// Dominant PascalCase (>=50%) maps to a per-language naming value.
	// No dominance → "unknown", which is an absence of signal, not a detection.
	dominant := float64(pascalCount)/float64(total) >= 0.50
	value := "unknown"
	source := ContextSourceInferred
	switch {
	case p.lang == "go" && dominant:
		value = "snake_case filenames, PascalCase exported symbols"
		source = ContextSourceDetected
	case p.lang == "node" && dominant:
		value = "PascalCase exported symbols"
		source = ContextSourceDetected
	}

	return ContextDetection{
		Value:      value,
		Source:     source,
		Confidence: conf,
	}, nil
}

// extensionsForLang returns the file extensions to sample for a given language.
func extensionsForLang(lang string) []string {
	switch lang {
	case "go":
		return []string{".go"}
	case "node":
		return []string{".ts", ".tsx"}
	default:
		return nil
	}
}

// regexpsForLang returns the pascal/camel regexps for a given language.
func regexpsForLang(lang string) (pascal, camel *regexp.Regexp) {
	switch lang {
	case "node":
		return nodePascalRe, nodeCamelRe
	default:
		return goPascalRe, goCamelRe
	}
}

// --- ArchitecturalStyleProbe ---

type architecturalStyleProbe struct{}

// ArchitecturalStyleProbe returns a ContextProbe that infers the architectural
// style from top-level directory layout.
func ArchitecturalStyleProbe() ContextProbe {
	return &architecturalStyleProbe{}
}

// Name returns the field identifier for ArchitecturalStyleProbe.
func (p *architecturalStyleProbe) Name() string { return "architectural_style" }

// Detect reads top-level directories and applies heuristic rules.
func (p *architecturalStyleProbe) Detect(projectRoot string) (ContextDetection, error) {
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

	// hexagonal: cmd/ AND internal/ present.
	if dirs["cmd"] && dirs["internal"] {
		return ContextDetection{
			Value:      "hexagonal",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceHigh,
		}, nil
	}

	// src/ present — dive one level deeper.
	if dirs["src"] {
		srcEntries, err := os.ReadDir(filepath.Join(projectRoot, "src"))
		if err == nil {
			srcDirs := make(map[string]bool)
			for _, e := range srcEntries {
				if e.IsDir() {
					srcDirs[e.Name()] = true
				}
			}
			if srcDirs["controllers"] && srcDirs["models"] && srcDirs["views"] {
				return ContextDetection{
					Value:      "mvc",
					Source:     ContextSourceDetected,
					Confidence: ContextConfidenceHigh,
				}, nil
			}
			if srcDirs["features"] || srcDirs["modules"] {
				return ContextDetection{
					Value:      "modular",
					Source:     ContextSourceDetected,
					Confidence: ContextConfidenceMedium,
				}, nil
			}
		}
		return ContextDetection{
			Value:      "layered",
			Source:     ContextSourceDetected,
			Confidence: ContextConfidenceMedium,
		}, nil
	}

	// lib/ only (no src/).
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
