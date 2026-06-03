package prompt_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
)

// buildTestFS creates an in-memory FS with a packaged role and a packaged skill.
func buildTestFS(t *testing.T) fstest.MapFS {
	t.Helper()
	return fstest.MapFS{
		"roles/requirements/role.md": {Data: []byte("# packaged requirements role\nI am the requirements persona.")},
		"roles/knowledge/role.md":    {Data: []byte("# packaged knowledge role\nI am the knowledge persona.")},
		"skills/user-story-writing.md": {Data: []byte("# user story writing skill\nWrite stories as: As a ... I want ... So that ...")},
	}
}

// TestOverridePrecedence_LocalBeatsPackaged verifies that a project-local
// role override wins over the packaged default.
func TestOverridePrecedence_LocalBeatsPackaged(t *testing.T) {
	packaged := prompt.NewEmbeddedRegistry(buildTestFS(t))

	// Write a local override to a temp dir.
	localDir := t.TempDir()
	roleDir := filepath.Join(localDir, "requirements")
	if err := os.MkdirAll(roleDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	localContent := "# local override requirements role\nCustom persona."
	if err := os.WriteFile(filepath.Join(roleDir, "role.md"), []byte(localContent), 0o644); err != nil {
		t.Fatalf("write local role: %v", err)
	}

	resolver := prompt.NewOverrideResolver(localDir, "", packaged)

	frag, err := resolver.Role("requirements")
	if err != nil {
		t.Fatalf("Role: %v", err)
	}
	if frag.Source != prompt.SourceLocal {
		t.Errorf("Source: got %q, want %q", frag.Source, prompt.SourceLocal)
	}
	if frag.Content != localContent {
		t.Errorf("Content: got %q, want %q", frag.Content, localContent)
	}
}

// TestOverridePrecedence_FallsBackToPackaged verifies that the packaged
// default is used when no override exists.
func TestOverridePrecedence_FallsBackToPackaged(t *testing.T) {
	packaged := prompt.NewEmbeddedRegistry(buildTestFS(t))

	// Empty local dir — no overrides.
	localDir := t.TempDir()
	resolver := prompt.NewOverrideResolver(localDir, "", packaged)

	frag, err := resolver.Role("requirements")
	if err != nil {
		t.Fatalf("Role: %v", err)
	}
	if frag.Source != prompt.SourcePackaged {
		t.Errorf("Source: got %q, want %q", frag.Source, prompt.SourcePackaged)
	}
}

// TestComposeLayerOrder verifies that the composed output contains layers in the
// correct order: role → skills → artifact context → platform context.
func TestComposeLayerOrder(t *testing.T) {
	role := prompt.NewFragment("requirements", "ROLE_CONTENT", prompt.SourcePackaged)
	skill := prompt.NewFragment("user-story-writing", "SKILL_CONTENT", prompt.SourcePackaged)
	artifactCtx := "ARTIFACT_CONTEXT"
	platformCtx := "PLATFORM_CONTEXT"

	composed, manifest := prompt.Compose(role, []prompt.Fragment{skill}, artifactCtx, platformCtx)

	// Verify layer ordering by checking positions.
	rolePos := indexOf(composed, "ROLE_CONTENT")
	skillPos := indexOf(composed, "SKILL_CONTENT")
	artifactPos := indexOf(composed, "ARTIFACT_CONTEXT")
	platformPos := indexOf(composed, "PLATFORM_CONTEXT")

	if rolePos < 0 {
		t.Error("role content missing from composed output")
	}
	if skillPos < 0 {
		t.Error("skill content missing from composed output")
	}
	if artifactPos < 0 {
		t.Error("artifact context missing from composed output")
	}
	if platformPos < 0 {
		t.Error("platform context missing from composed output")
	}

	if rolePos > skillPos {
		t.Errorf("role must come before skill: role@%d skill@%d", rolePos, skillPos)
	}
	if skillPos > artifactPos {
		t.Errorf("skill must come before artifact: skill@%d artifact@%d", skillPos, artifactPos)
	}
	if artifactPos > platformPos {
		t.Errorf("artifact must come before platform: artifact@%d platform@%d", artifactPos, platformPos)
	}

	// Manifest must contain both role and skill.
	if manifest.Get("requirements") == "" {
		t.Error("manifest missing role version")
	}
	if manifest.Get("user-story-writing") == "" {
		t.Error("manifest missing skill version")
	}
}

// TestManifestHashStability verifies that identical fragment sets always
// produce the same hash, and different sets produce different hashes.
func TestManifestHashStability(t *testing.T) {
	build := func() prompt.Manifest {
		m := prompt.NewManifest()
		m.Set("requirements", "abc12345")
		m.Set("user-story-writing", "def67890")
		return m
	}

	h1 := build().Hash()
	h2 := build().Hash()
	if h1 != h2 {
		t.Errorf("same inputs produced different hashes: %q vs %q", h1, h2)
	}

	// Different versions → different hash.
	different := prompt.NewManifest()
	different.Set("requirements", "zzzzffff")
	different.Set("user-story-writing", "def67890")
	if different.Hash() == h1 {
		t.Error("different versions should produce different hashes")
	}
}

// TestFragmentVersionIsStable verifies that the same content always produces
// the same 8-char version string.
func TestFragmentVersionIsStable(t *testing.T) {
	content := "some prompt content"
	f1 := prompt.NewFragment("test", content, prompt.SourcePackaged)
	f2 := prompt.NewFragment("test", content, prompt.SourcePackaged)
	if f1.Version != f2.Version {
		t.Errorf("same content produced different versions: %q vs %q", f1.Version, f2.Version)
	}
	if len(f1.Version) != 8 {
		t.Errorf("version should be 8 chars, got %d: %q", len(f1.Version), f1.Version)
	}
}

// indexOf returns the position of substr in s, or -1.
func indexOf(s, substr string) int {
	idx := -1
	for i := range s {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			idx = i
			break
		}
	}
	return idx
}
