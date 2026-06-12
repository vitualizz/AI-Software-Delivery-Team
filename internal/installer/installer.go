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
	Removed         []string // SkillsDir-relative paths pruned as stale; never affects Err
	Err             error
}

// Install copies skill files from skillsFS into each assistant's SkillsDir,
// applies provider.CustomizeSkill to each file's content, and returns one
// result per assistant. A failure for one assistant does not abort the others.
// lang is the language code chosen in the TUI; it is recorded in each
// assistant's install metadata (an empty lang preserves the existing value).
func Install(assistants []AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS, lang string) []InstallResult {
	return InstallWithModels(assistants, provider, skillsFS, lang, nil)
}

// InstallWithModels is Install with per-step model selections from the TUI.
// models maps "{specialist}/{step}" to the model value injected into that
// step's `model:` field as each workflow.yaml is written. A nil/empty map
// installs files unmodified.
func InstallWithModels(assistants []AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS, lang string, models map[string]string) []InstallResult {
	results := make([]InstallResult, len(assistants))

	for i, assistant := range assistants {
		results[i] = installOne(assistant, provider, skillsFS, lang, models, false)
	}

	return results
}

// InstallRemovingModels is InstallWithModels for the Chameleon preset: it
// strips the `model:` field from every subagent step as each workflow.yaml is
// written, so each step inherits the model the assistant already has defined.
func InstallRemovingModels(assistants []AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS, lang string) []InstallResult {
	results := make([]InstallResult, len(assistants))

	for i, assistant := range assistants {
		results[i] = installOne(assistant, provider, skillsFS, lang, nil, true)
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
func installOne(assistant AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS, lang string, models map[string]string, removeModels bool) InstallResult {
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

		written, copyErr := copyEntry(skillsFS, entry, srcRoot, destDir, provider, models, removeModels)
		if copyErr != nil {
			result.Err = copyErr
			return result
		}
		result.Written = append(result.Written, written...)
	}

	generateCommands(assistant, skillsFS, &result)
	generateAgents(assistant, skillsFS, &result)

	if result.Err == nil {
		// Preserve existing persona, emoji preference, and language so a
		// skill-only reinstall doesn't clear them. AgentTypes is set to the
		// canonical list on every successful install — that is what the
		// install just provisioned.
		existing, _ := ReadInstallMeta(assistant)
		if lang == "" {
			lang = existing.Language
		}
		// Prune files a previous install wrote that this one no longer
		// provides, then record the fresh manifest. Prune is best-effort and
		// never fails the install.
		relWritten := relativizeWritten(assistant.SkillsDir, result.Written)
		result.Removed = pruneStale(assistant.SkillsDir, managedRootsFor(skillsFS), existing.Files, relWritten)
		_ = WriteInstallMeta(assistant, InstallMeta{
			InstalledAt: time.Now().UTC(),
			Persona:     existing.Persona,
			Emojis:      existing.Emojis,
			Language:    lang,
			AgentTypes:  AgentTypeNames,
			Files:       relWritten,
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
func copyEntry(skillsFS fs.FS, entry fs.DirEntry, srcRoot, destDir string, provider ProviderDescriptor, models map[string]string, removeModels bool) ([]string, error) {
	var written []string

	if !entry.IsDir() {
		rel := entry.Name()
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if err := writeSkillFile(skillsFS, entry.Name(), target, provider, models, removeModels); err != nil {
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

		if err := writeSkillFile(skillsFS, path, target, provider, models, removeModels); err != nil {
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
// customization — plus per-step model injection for workflow.yaml files —
// and writes the result to target, creating any needed parent directories.
func writeSkillFile(skillsFS fs.FS, srcPath, target string, provider ProviderDescriptor, models map[string]string, removeModels bool) error {
	data, readErr := fs.ReadFile(skillsFS, srcPath)
	if readErr != nil {
		return fmt.Errorf("read %s: %w", srcPath, readErr)
	}

	if filepath.Base(target) == "workflow.yaml" {
		if removeModels {
			stripped, rmErr := RemoveModels(data)
			if rmErr != nil {
				return fmt.Errorf("remove models from %s: %w", srcPath, rmErr)
			}
			data = stripped
		} else {
			injected, injErr := InjectModels(data, models)
			if injErr != nil {
				return fmt.Errorf("inject models into %s: %w", srcPath, injErr)
			}
			data = injected
		}
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
