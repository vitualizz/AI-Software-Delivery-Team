package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// InstallResult holds the outcome of installing skills for one assistant.
type InstallResult struct {
	AssistantID     AssistantID
	Written         []string
	WrittenCommands []string
	Err             error
}

// Install copies skill files from skillsFS into each assistant's SkillsDir,
// applies provider.CustomizeSkill to each file's content, and returns one
// result per assistant. A failure for one assistant does not abort the others.
func Install(assistants []AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS) []InstallResult {
	results := make([]InstallResult, len(assistants))

	for i, assistant := range assistants {
		results[i] = installOne(assistant, provider, skillsFS)
	}

	return results
}

// siblingConsultantDir is the destination directory name for the consultant
// (the loose root-level SKILL.md and any other root-level files). It is the
// one fixed name in an otherwise generic mapping.
const siblingConsultantDir = "asdt"

// SiblingDestName derives the destination sibling directory name for a
// top-level entry of the embedded skill tree.
//
//   - "."  (the embedded tree root — loose files such as SKILL.md belong to
//     the consultant) maps to "asdt".
//   - Every other top-level entry name is used VERBATIM as its destination
//     sibling directory name — no per-specialist remapping table.
//
// This keeps the consultant-vs-specialist distinction to a single named,
// testable unit while the rest of the mapping stays generic (per design:
// "asdt-shared is just another entry, zero special-casing").
func SiblingDestName(entry string) string {
	if entry == "." {
		return siblingConsultantDir
	}
	return entry
}

// installOne installs the embedded skill tree for one assistant by mapping
// each TOP-LEVEL entry to its own sibling directory directly under the
// assistant's skills root (assistant.SkillsDir). It does NOT copy the whole
// tree into a single nested destination — that was the root cause of
// specialists never being registered as separate invocable commands.
//
// Top-level entries are:
//   - directories directly under the embedded tree root (e.g.
//     "asdt-architect/", "asdt-shared/") — their entire subtree is copied to
//     {SkillsDir}/{SiblingDestName(entry)}/...
//   - loose files directly at the embedded tree root (e.g. "SKILL.md") —
//     they belong to the consultant and are copied to
//     {SkillsDir}/asdt/{filename}
func installOne(assistant AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS) InstallResult {
	result := InstallResult{AssistantID: assistant.ID}

	if err := os.MkdirAll(assistant.SkillsDir, 0o755); err != nil {
		result.Err = fmt.Errorf("mkdir %s: %w", assistant.SkillsDir, err)
		return result
	}

	entries, err := fs.ReadDir(skillsFS, ".")
	if err != nil {
		result.Err = fmt.Errorf("read root: %w", err)
		return result
	}

	for _, entry := range entries {
		var srcRoot, destDir string
		if entry.IsDir() {
			srcRoot = entry.Name()
			destDir = filepath.Join(assistant.SkillsDir, SiblingDestName(entry.Name()))
		} else {
			// Loose root-level file (e.g. SKILL.md) belongs to the consultant.
			srcRoot = "."
			destDir = filepath.Join(assistant.SkillsDir, SiblingDestName("."))
		}

		written, copyErr := copyEntry(skillsFS, entry, srcRoot, destDir, provider)
		if copyErr != nil {
			result.Err = copyErr
			return result
		}
		result.Written = append(result.Written, written...)
	}

	generateCommands(assistant, skillsFS, &result)

	if result.Err == nil {
		// Preserve existing persona so a skill-only reinstall doesn't clear it.
		existing, _ := ReadInstallMeta(assistant)
		_ = WriteInstallMeta(assistant, InstallMeta{
			InstalledAt: time.Now().UTC(),
			Persona:     existing.Persona,
		})
	}

	return result
}

func generateCommands(assistant AssistantDescriptor, skillsFS fs.FS, result *InstallResult) {
	adapter, ok := adapterFor(assistant.ID)
	if !ok {
		return
	}

	commandRoot := commandRootFor(assistant.ID)
	if commandRoot == "" {
		return
	}

	written, genErr := adapter.Generate(skillsFS, commandRoot)
	result.WrittenCommands = append(result.WrittenCommands, written...)
	if genErr != nil {
		result.Err = genErr
	}
}

func commandRootFor(id AssistantID) string {
	switch id {
	case AssistantOpenCode:
		return openCodeCommandRoot()
	default:
		return ""
	}
}

// copyEntry copies a single top-level entry (file or directory) from skillsFS
// into destDir, preserving its relative subtree structure for directories.
// srcRoot is "." for loose root-level files (so the file's own name is used
// as the relative path) or the directory entry's own name (so its subtree is
// walked and copied beneath destDir).
func copyEntry(skillsFS fs.FS, entry fs.DirEntry, srcRoot, destDir string, provider ProviderDescriptor) ([]string, error) {
	var written []string

	if !entry.IsDir() {
		rel := entry.Name()
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if err := writeSkillFile(skillsFS, entry.Name(), target, provider); err != nil {
			return nil, err
		}
		return []string{target}, nil
	}

	walkErr := fs.WalkDir(skillsFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(filepath.FromSlash(srcRoot), filepath.FromSlash(path))
		if relErr != nil {
			return fmt.Errorf("relativize %s under %s: %w", path, srcRoot, relErr)
		}
		target := filepath.Join(destDir, rel)

		if err := writeSkillFile(skillsFS, path, target, provider); err != nil {
			return err
		}
		written = append(written, target)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	return written, nil
}

// writeSkillFile reads srcPath from skillsFS, applies the provider's content
// customization, and writes the result to target — creating any needed
// parent directories.
func writeSkillFile(skillsFS fs.FS, srcPath, target string, provider ProviderDescriptor) error {
	data, readErr := fs.ReadFile(skillsFS, srcPath)
	if readErr != nil {
		return fmt.Errorf("read %s: %w", srcPath, readErr)
	}

	content := provider.CustomizeSkill(string(data))

	if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
		return fmt.Errorf("mkdir for %s: %w", target, mkErr)
	}

	if writeErr := os.WriteFile(target, []byte(content), 0o644); writeErr != nil {
		return fmt.Errorf("write %s: %w", target, writeErr)
	}

	return nil
}
