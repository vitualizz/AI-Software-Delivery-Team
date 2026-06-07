package prompt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/prompt"
	"github.com/vitualizz/ai-software-delivery-team/skill"
)

// buildTestFS creates an in-memory FS with a packaged role and a packaged skill.
func buildTestFS(t *testing.T) fstest.MapFS {
	t.Helper()
	return fstest.MapFS{
		"roles/requirements/role.md":   {Data: []byte("# packaged requirements role\nI am the requirements persona.")},
		"roles/knowledge/role.md":      {Data: []byte("# packaged knowledge role\nI am the knowledge persona.")},
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

// TestDefaultEmbeddedRegistry_RoleDeveloper verifies the developer role resolves
// from the production embedded FS.
func TestDefaultEmbeddedRegistry_RoleDeveloper(t *testing.T) {
	reg := prompt.NewEmbeddedRegistry(skill.FS())
	frag, err := reg.Role("developer")
	if err != nil {
		t.Fatalf("Role(developer): %v", err)
	}
	if frag.Content == "" {
		t.Error("Role(developer): content is empty")
	}
}

// TestDefaultEmbeddedRegistry_VersionStability verifies that Version returns
// a stable 8-character hash for the same embedded content across two calls.
func TestDefaultEmbeddedRegistry_VersionStability(t *testing.T) {
	reg := prompt.NewEmbeddedRegistry(skill.FS())

	v1, err := reg.Version("developer")
	if err != nil {
		t.Fatalf("Version(developer) first call: %v", err)
	}
	v2, err := reg.Version("developer")
	if err != nil {
		t.Fatalf("Version(developer) second call: %v", err)
	}
	if v1 != v2 {
		t.Errorf("Version is not stable: %q vs %q", v1, v2)
	}
	if len(v1) != 8 {
		t.Errorf("Version should be 8 chars, got %d: %q", len(v1), v1)
	}
}

// TestCompose_EmptySkills verifies that Compose with an empty skills slice
// produces output that contains the role content but no skill separator artifacts.
func TestCompose_EmptySkills(t *testing.T) {
	role := prompt.NewFragment("requirements", "ROLE_CONTENT", prompt.SourcePackaged)

	composed, manifest := prompt.Compose(role, []prompt.Fragment{}, "ARTIFACT_CTX", "PLATFORM_CTX")

	if !strings.Contains(composed, "ROLE_CONTENT") {
		t.Error("composed output missing role content")
	}
	// Manifest must contain the role version.
	if manifest.Get("requirements") == "" {
		t.Error("manifest missing role version")
	}
	// No skill versions should be in the manifest.
	all := manifest.All()
	for name := range all {
		if name != "requirements" {
			t.Errorf("unexpected manifest entry %q for empty skills slice", name)
		}
	}
}

// TestCompose_NilPlatformContext verifies that Compose with an empty platform
// context does not include an empty platform section in the output.
func TestCompose_NilPlatformContext(t *testing.T) {
	role := prompt.NewFragment("requirements", "ROLE_CONTENT", prompt.SourcePackaged)
	skill := prompt.NewFragment("user-story-writing", "SKILL_CONTENT", prompt.SourcePackaged)

	composed, _ := prompt.Compose(role, []prompt.Fragment{skill}, "ARTIFACT_CTX", "")

	if !strings.Contains(composed, "ROLE_CONTENT") {
		t.Error("composed output missing role content")
	}
	if !strings.Contains(composed, "SKILL_CONTENT") {
		t.Error("composed output missing skill content")
	}
	if !strings.Contains(composed, "ARTIFACT_CTX") {
		t.Error("composed output missing artifact context")
	}
	// With empty platform context, there should be no trailing separator.
	// Count separators: role→skill (1) + skill→artifact (1) = 2 separators max.
	separatorCount := strings.Count(composed, "---")
	if separatorCount > 2 {
		t.Errorf("expected at most 2 separators with empty platform, got %d", separatorCount)
	}
}

// TestOverrideResolver_SkillPrecedence verifies that OverrideResolver.Skill
// uses the local override when it exists, and falls back to packaged otherwise.
func TestOverrideResolver_SkillPrecedence(t *testing.T) {
	packaged := prompt.NewEmbeddedRegistry(buildTestFS(t))

	localDir := t.TempDir()
	skillsDir := filepath.Join(localDir, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	localContent := "# local override skill\nCustom skill content."
	if err := os.WriteFile(filepath.Join(skillsDir, "user-story-writing.md"), []byte(localContent), 0o644); err != nil {
		t.Fatalf("write local skill: %v", err)
	}

	resolver := prompt.NewOverrideResolver(localDir, "", packaged)

	frag, err := resolver.Skill("user-story-writing")
	if err != nil {
		t.Fatalf("Skill: %v", err)
	}
	if frag.Source != prompt.SourceLocal {
		t.Errorf("Source: got %q, want %q", frag.Source, prompt.SourceLocal)
	}
	if frag.Content != localContent {
		t.Errorf("Content: got %q, want %q", frag.Content, localContent)
	}
}

// TestOverrideResolver_SkillFallsBackToPackaged verifies that Skill falls back
// to the packaged registry when no local override exists.
func TestOverrideResolver_SkillFallsBackToPackaged(t *testing.T) {
	packaged := prompt.NewEmbeddedRegistry(buildTestFS(t))
	localDir := t.TempDir()
	resolver := prompt.NewOverrideResolver(localDir, "", packaged)

	frag, err := resolver.Skill("user-story-writing")
	if err != nil {
		t.Fatalf("Skill: %v", err)
	}
	if frag.Source != prompt.SourcePackaged {
		t.Errorf("Source: got %q, want %q", frag.Source, prompt.SourcePackaged)
	}
}

// TestOverrideResolver_Version verifies that Version returns the correct version
// for a role that exists (without error).
func TestOverrideResolver_Version(t *testing.T) {
	packaged := prompt.NewEmbeddedRegistry(buildTestFS(t))
	localDir := t.TempDir()
	resolver := prompt.NewOverrideResolver(localDir, "", packaged)

	v, err := resolver.Version("requirements")
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if len(v) != 8 {
		t.Errorf("Version should be 8 chars, got %d: %q", len(v), v)
	}
}

// TestManifestHash_StableAcrossInsertionOrder verifies that Manifest.Hash
// is deterministic regardless of insertion order.
func TestManifestHash_StableAcrossInsertionOrder(t *testing.T) {
	buildAB := func() prompt.Manifest {
		m := prompt.NewManifest()
		m.Set("alpha", "version1")
		m.Set("beta", "version2")
		return m
	}
	buildBA := func() prompt.Manifest {
		m := prompt.NewManifest()
		m.Set("beta", "version2")
		m.Set("alpha", "version1")
		return m
	}

	h1 := buildAB().Hash()
	h2 := buildBA().Hash()
	if h1 != h2 {
		t.Errorf("hash differs by insertion order: AB=%q BA=%q", h1, h2)
	}
}

// TestManifestAll_ReturnsCopy verifies that Manifest.All returns all entries
// and that modifying the returned map does not affect the manifest.
func TestManifestAll_ReturnsCopy(t *testing.T) {
	m := prompt.NewManifest()
	m.Set("role", "abc")
	m.Set("skill", "def")

	all := m.All()
	if len(all) != 2 {
		t.Errorf("All() len: got %d, want 2", len(all))
	}

	// Mutate the returned map — must not affect manifest.
	all["role"] = "mutated"
	if m.Get("role") != "abc" {
		t.Error("modifying All() return value should not affect the manifest")
	}
}

// TestFragmentVersion_DifferentContent verifies that different content
// produces a different version string.
func TestFragmentVersion_DifferentContent(t *testing.T) {
	f1 := prompt.NewFragment("test", "content version A", prompt.SourcePackaged)
	f2 := prompt.NewFragment("test", "content version B", prompt.SourcePackaged)
	if f1.Version == f2.Version {
		t.Errorf("different content should produce different versions; both got %q", f1.Version)
	}
}

// TestNewRegistry_ScopedSkillDeveloper verifies that ScopedSkill("developer", "code-generation")
// resolves from the developer/skills/ tree in the production embedded FS.
func TestNewRegistry_ScopedSkillDeveloper(t *testing.T) {
	reg := prompt.DefaultEmbeddedRegistry()
	frag, err := reg.ScopedSkill("developer", "code-generation")
	if err != nil {
		t.Fatalf("ScopedSkill(developer, code-generation): %v", err)
	}
	if frag.Content == "" {
		t.Error("ScopedSkill(developer, code-generation): content is empty")
	}
}

// TestNewRegistry_ScopedSkillSecurity verifies that ScopedSkill("security", "threat-modeling")
// resolves from the security/skills/ tree.
func TestNewRegistry_ScopedSkillSecurity(t *testing.T) {
	reg := prompt.DefaultEmbeddedRegistry()
	frag, err := reg.ScopedSkill("security", "threat-modeling")
	if err != nil {
		t.Fatalf("ScopedSkill(security, threat-modeling): %v", err)
	}
	if frag.Content == "" {
		t.Error("ScopedSkill(security, threat-modeling): content is empty")
	}
}

// TestNewRegistry_SharedSkillPlatformContext verifies Skill("platform-context")
// resolves from asdt-shared/skills/ in the production embedded FS.
func TestNewRegistry_SharedSkillPlatformContext(t *testing.T) {
	reg := prompt.DefaultEmbeddedRegistry()
	frag, err := reg.Skill("platform-context")
	if err != nil {
		t.Fatalf("Skill(platform-context): %v", err)
	}
	if frag.Content == "" {
		t.Error("Skill(platform-context): content is empty")
	}
}

// TestNewRegistry_RoleDeveloperSpecialist verifies that Role("developer") resolves
// skill/developer/SKILL.md (with frontmatter stripped) from the production FS.
func TestNewRegistry_RoleDeveloperSpecialist(t *testing.T) {
	reg := prompt.DefaultEmbeddedRegistry()
	frag, err := reg.Role("developer")
	if err != nil {
		t.Fatalf("Role(developer) via specialist layout: %v", err)
	}
	if frag.Content == "" {
		t.Error("Role(developer): content is empty")
	}
	// Frontmatter must be stripped — the body must not start with "---".
	if strings.HasPrefix(strings.TrimSpace(frag.Content), "---") {
		t.Error("Role(developer): frontmatter was not stripped from SKILL.md")
	}
}

// TestDefaultEmbeddedRegistry_RoleResolvesAllSpecialistIDs is a regression
// guard for the bare-specialist-ID -> embedded-sibling-directory mapping.
// Production specialist IDs (frontmatter "specialist-id") stay bare
// (e.g. "developer", "security"), but the embedded tree's directories carry
// the installed "asdt-" prefix (e.g. "asdt-developer", "asdt-security").
// Role(id) must resolve correctly for every such ID — including "asdt-init",
// whose ID is ALREADY prefixed and must not be double-prefixed.
func TestDefaultEmbeddedRegistry_RoleResolvesAllSpecialistIDs(t *testing.T) {
	reg := prompt.DefaultEmbeddedRegistry()
	ids := []string{"architect", "developer", "qa", "security", "ux-ui", "asdt-init"}
	for _, id := range ids {
		frag, err := reg.Role(id)
		if err != nil {
			t.Errorf("Role(%q): %v", id, err)
			continue
		}
		if frag.Content == "" {
			t.Errorf("Role(%q): content is empty", id)
		}
		if strings.HasPrefix(strings.TrimSpace(frag.Content), "---") {
			t.Errorf("Role(%q): frontmatter was not stripped from SKILL.md", id)
		}
	}
}

// TestNewRegistry_ScopedSkillUnknownReturnsError verifies that ScopedSkill
// with an unknown specialist and skill name returns an error.
func TestNewRegistry_ScopedSkillUnknownReturnsError(t *testing.T) {
	reg := prompt.DefaultEmbeddedRegistry()
	_, err := reg.ScopedSkill("unknown-specialist", "nonexistent-skill")
	if err == nil {
		t.Error("ScopedSkill(unknown, anything): expected error, got nil")
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
