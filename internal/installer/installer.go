package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/vitualizz/ai-software-delivery-team/internal/config"
)

// InstallResult holds the outcome of installing skills for one assistant.
type InstallResult struct {
	AssistantID AssistantID
	Written     []string
	Err         error
}

// Install copies skill files from skillsFS into each assistant's SkillsDir,
// applies provider.CustomizeSkill to each file's content, and writes the chosen
// provider ID to .asdt/config.yaml. It returns one result per assistant;
// a failure for one assistant does not abort the others.
func Install(assistants []AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS, cfgRoot config.Root) []InstallResult {
	results := make([]InstallResult, len(assistants))

	for i, assistant := range assistants {
		results[i] = installOne(assistant, provider, skillsFS)
	}

	// Write provider choice to config regardless of per-assistant errors.
	cfg, _ := config.Load(cfgRoot)
	cfg.Memory.Provider = string(provider.ID)
	_ = config.Save(cfgRoot, cfg)

	return results
}

func installOne(assistant AssistantDescriptor, provider ProviderDescriptor, skillsFS fs.FS) InstallResult {
	result := InstallResult{AssistantID: assistant.ID}

	if err := os.MkdirAll(assistant.SkillsDir, 0o755); err != nil {
		result.Err = fmt.Errorf("mkdir %s: %w", assistant.SkillsDir, err)
		return result
	}

	err := fs.WalkDir(skillsFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		data, readErr := fs.ReadFile(skillsFS, path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		content := provider.CustomizeSkill(string(data))
		target := filepath.Join(assistant.SkillsDir, filepath.FromSlash(path))

		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			return fmt.Errorf("mkdir for %s: %w", target, mkErr)
		}

		if writeErr := os.WriteFile(target, []byte(content), 0o644); writeErr != nil {
			return fmt.Errorf("write %s: %w", target, writeErr)
		}

		result.Written = append(result.Written, target)
		return nil
	})

	if err != nil {
		result.Err = err
	}
	return result
}
