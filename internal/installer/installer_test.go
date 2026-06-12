package installer_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/asdt/internal/installer"
)

var testFS = fstest.MapFS{
	"architect/SKILL.md": &fstest.MapFile{Data: []byte("# Architect skill")},
	"developer/SKILL.md": &fstest.MapFile{Data: []byte("# Developer skill")},
}

// siblingFS models a generic embedded skill/ tree shape: a loose root-level
// SKILL.md for the consultant, plus arbitrary top-level directories for
// specialists and a shared fragment library. Names are deliberately
// NOT the production "asdt-*" names — this proves the mapping is purely
// structural (verbatim-by-default, "." -> "asdt") and applies generically to
// whatever top-level entries the embedded tree happens to contain, with zero
// per-name special-casing.
var siblingFS = fstest.MapFS{
	"SKILL.md":                   &fstest.MapFile{Data: []byte("# ASDT consultant")},
	"sample-specialist/SKILL.md": &fstest.MapFile{Data: []byte("# Sample specialist skill")},
	"sample-shared/skills/x.md":  &fstest.MapFile{Data: []byte("# Shared fragment")},
}

// TestInstallOne_SiblingLayout verifies each top-level entry of the embedded
// skill tree is installed as its OWN top-level sibling directory under the
// assistant's skills root — never nested inside another skill's directory.
// The loose root-level SKILL.md (the consultant) is routed to its own "asdt"
// directory; top-level directory entries keep their name verbatim as the
// sibling destination (no per-specialist special-casing).
func TestInstallOne_SiblingLayout(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "skills")
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root},
	}
	provider := installer.Providers[0]

	results := installer.Install(assistants, provider, siblingFS, "")
	if results[0].Err != nil {
		t.Fatalf("install failed: %v", results[0].Err)
	}

	// Sibling destinations: each top-level source entry gets its own directory
	// directly under the skills root, named verbatim after the source entry.
	checkFile(t, filepath.Join(root, "sample-specialist", "SKILL.md"))
	checkFile(t, filepath.Join(root, "sample-shared", "skills", "x.md"))

	// The loose root-level SKILL.md belongs to the consultant and must land
	// under its own "asdt" directory.
	checkFile(t, filepath.Join(root, "asdt", "SKILL.md"))

	// Nesting prohibition: specialist/shared trees must NOT land inside the
	// consultant's "asdt" directory, and the consultant must not be duplicated
	// inside a specialist directory.
	checkFileAbsent(t, filepath.Join(root, "asdt", "sample-specialist", "SKILL.md"))
	checkFileAbsent(t, filepath.Join(root, "asdt", "sample-shared", "skills", "x.md"))
	checkFileAbsent(t, filepath.Join(root, "sample-specialist", "asdt", "SKILL.md"))
}

// productionShapedFS mirrors the POST-RENAME embedded skill/ tree shape using
// the actual production sibling names: a loose root-level SKILL.md (the
// asdt consultant), a renamed specialist directory (asdt-architect), the
// renamed shared fragment library (asdt-shared), and a literal "asdt/" entry
// — covering the edge case where the embedded tree itself contains a
// directory literally named "asdt" (e.g. a future asdt-prefixed addition)
// alongside the loose root SKILL.md that also maps to "asdt".
var productionShapedFS = fstest.MapFS{
	"SKILL.md":                &fstest.MapFile{Data: []byte("# ASDT consultant")},
	"asdt-architect/SKILL.md": &fstest.MapFile{Data: []byte("# Architect skill")},
	"asdt-shared/skills/x.md": &fstest.MapFile{Data: []byte("# Shared fragment")},
	"asdt/extra.md":           &fstest.MapFile{Data: []byte("# Extra consultant-scoped file")},
}

// TestInstallOne_ProductionShapedSiblingLayout verifies installOne against the
// real post-rename production naming: asdt-architect and asdt-shared land as
// their own top-level siblings, and a literal top-level "asdt/" directory
// entry merges correctly into the same "asdt" destination as the loose
// root-level SKILL.md (both belong to the consultant) — without the
// specialist or shared trees ever landing inside "asdt/".
func TestInstallOne_ProductionShapedSiblingLayout(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "skills")
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root},
	}
	provider := installer.Providers[0]

	results := installer.Install(assistants, provider, productionShapedFS, "")
	if results[0].Err != nil {
		t.Fatalf("install failed: %v", results[0].Err)
	}

	checkFile(t, filepath.Join(root, "asdt-architect", "SKILL.md"))
	checkFile(t, filepath.Join(root, "asdt-shared", "skills", "x.md"))
	checkFile(t, filepath.Join(root, "asdt", "SKILL.md"))
	checkFile(t, filepath.Join(root, "asdt", "extra.md"))

	checkFileAbsent(t, filepath.Join(root, "asdt", "asdt-architect", "SKILL.md"))
	checkFileAbsent(t, filepath.Join(root, "asdt", "asdt-shared", "skills", "x.md"))
	checkFileAbsent(t, filepath.Join(root, "asdt-architect", "asdt", "SKILL.md"))
}

func checkFileAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file %q to NOT exist (nesting bug), but it does", path)
	}
}

// TestSiblingDestName verifies the entry-name → sibling-destination mapping:
// the loose root-level SKILL.md (entry ".") maps to "asdt" (the consultant's
// own directory); every other top-level entry keeps its name verbatim. This
// is the single named, testable unit the design requires — no per-specialist
// special-casing.
func TestSiblingDestName(t *testing.T) {
	cases := []struct {
		entry string
		want  string
	}{
		{entry: ".", want: "asdt"},
		{entry: "asdt-architect", want: "asdt-architect"},
		{entry: "asdt-developer", want: "asdt-developer"},
		{entry: "asdt-shared", want: "asdt-shared"},
		{entry: "asdt-init", want: "asdt-init"},
		{entry: "asdt", want: "asdt"},
		{entry: "anything-else", want: "anything-else"},
	}
	for _, c := range cases {
		got := installer.SiblingDestName(c.entry)
		if got != c.want {
			t.Errorf("SiblingDestName(%q) = %q, want %q", c.entry, got, c.want)
		}
	}
}

func TestInstall_SuccessForTwoAssistants(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir1, "skills")},
		{ID: "a2", Name: "A2", BinaryName: "sh", SkillsDir: filepath.Join(dir2, "skills")},
	}
	provider := installer.Providers[0] // engram

	results := installer.Install(assistants, provider, testFS, "")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d].Err = %v, want nil", i, r.Err)
		}
	}

	// Check files exist.
	checkFile(t, filepath.Join(dir1, "skills", "architect", "SKILL.md"))
	checkFile(t, filepath.Join(dir1, "skills", "developer", "SKILL.md"))
	checkFile(t, filepath.Join(dir2, "skills", "architect", "SKILL.md"))
	checkFile(t, filepath.Join(dir2, "skills", "developer", "SKILL.md"))
}

func TestInstall_PartialFailure(t *testing.T) {
	// Create an unwritable directory for the first assistant.
	dir1 := t.TempDir()
	unwritable := filepath.Join(dir1, "nope")
	if err := os.MkdirAll(unwritable, 0o555); err != nil { // read+execute, no write
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(unwritable, "skills")},
		{ID: "a2", Name: "A2", BinaryName: "sh", SkillsDir: filepath.Join(dir2, "skills")},
	}
	provider := installer.Providers[0]

	results := installer.Install(assistants, provider, testFS, "")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("results[0].Err should be non-nil for unwritable target")
	}
	if results[1].Err != nil {
		t.Errorf("results[1].Err = %v, want nil", results[1].Err)
	}
}

func TestInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir, "skills")},
	}
	provider := installer.Providers[0]

	results1 := installer.Install(assistants, provider, testFS, "")
	if results1[0].Err != nil {
		t.Fatalf("first install failed: %v", results1[0].Err)
	}

	results2 := installer.Install(assistants, provider, testFS, "")
	if results2[0].Err != nil {
		t.Errorf("second install (idempotent) failed: %v", results2[0].Err)
	}
}

// agentEnabledFS models an embedded tree complete enough for BOTH generated
// artifact kinds: a specialist SKILL.md with valid frontmatter (command
// wrappers) and the shared executor header (agent definitions).
var agentEnabledFS = fstest.MapFS{
	"SKILL.md": &fstest.MapFile{Data: []byte("# ASDT consultant")},
	"asdt-developer/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: asdt-developer
description: "Turns specs and designs into working code."
user-invocable: true
specialist-id: developer
---

# Developer specialist
`)},
	"asdt-shared/skills/executor-header.md": &fstest.MapFile{Data: []byte("# Executor Header\n\n> **EXECUTOR**: test header.\n")},
}

// TestInstall_AgentGenFailureIsolation verifies that a failure generating
// agent definitions (unwritable agent root) sets result.Err for that
// assistant, suppresses its install-meta write, and leaves the other
// assistant's install untouched.
func TestInstall_AgentGenFailureIsolation(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	// Commands root stays writable; the agents root is blocked by a FILE at
	// its path so MkdirAll fails for agent generation only.
	if err := os.MkdirAll(filepath.Join(xdg, "opencode", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xdg, "opencode", "agents"), []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}

	openCodeSkills := filepath.Join(t.TempDir(), "skills")
	otherSkills := filepath.Join(t.TempDir(), "skills")
	assistants := []installer.AssistantDescriptor{
		{ID: installer.AssistantOpenCode, Name: "OpenCode", BinaryName: "opencode", SkillsDir: openCodeSkills},
		{ID: "a2", Name: "A2", BinaryName: "sh", SkillsDir: otherSkills},
	}
	provider := installer.Providers[0]

	results := installer.Install(assistants, provider, agentEnabledFS, "")

	if results[0].Err == nil {
		t.Error("results[0].Err should be non-nil when the agent root is unwritable")
	}
	meta, err := installer.ReadInstallMeta(assistants[0])
	if err != nil {
		t.Fatalf("read opencode meta: %v", err)
	}
	if !meta.InstalledAt.IsZero() {
		t.Error("install meta was written for the failed assistant; failure must suppress the meta write")
	}

	if results[1].Err != nil {
		t.Errorf("results[1].Err = %v, want nil (other assistant must be unaffected)", results[1].Err)
	}
	otherMeta, err := installer.ReadInstallMeta(assistants[1])
	if err != nil {
		t.Fatalf("read other meta: %v", err)
	}
	if otherMeta.InstalledAt.IsZero() {
		t.Error("install meta missing for the unaffected assistant")
	}
}

// TestInstall_ReinstallPreservesPersonaAndSetsAgentTypes verifies the
// preserve-or-set meta discipline: a reinstall keeps the previously recorded
// persona and emoji preference but always (re)sets AgentTypes to the
// canonical list.
func TestInstall_ReinstallPreservesPersonaAndSetsAgentTypes(t *testing.T) {
	dir := t.TempDir()
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(dir, "skills")}
	provider := installer.Providers[0]

	results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, agentEnabledFS, "")
	if results[0].Err != nil {
		t.Fatalf("first install failed: %v", results[0].Err)
	}

	meta, err := installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta after first install: %v", err)
	}
	assertAgentTypesCanonical(t, meta.AgentTypes)

	// Simulate a persona and emoji preference chosen after install, alongside
	// stale agent types.
	meta.Persona = "Sky"
	meta.Emojis = "yes"
	meta.AgentTypes = []string{"stale"}
	if err := installer.WriteInstallMeta(assistant, meta); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	results = installer.Install([]installer.AssistantDescriptor{assistant}, provider, agentEnabledFS, "")
	if results[0].Err != nil {
		t.Fatalf("reinstall failed: %v", results[0].Err)
	}

	meta, err = installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta after reinstall: %v", err)
	}
	if meta.Persona != "Sky" {
		t.Errorf("meta.Persona = %q after reinstall, want %q (must be preserved)", meta.Persona, "Sky")
	}
	if meta.Emojis != "yes" {
		t.Errorf("meta.Emojis = %q after reinstall, want %q (must be preserved)", meta.Emojis, "yes")
	}
	assertAgentTypesCanonical(t, meta.AgentTypes)
	if meta.InstalledAt.IsZero() {
		t.Error("meta.InstalledAt is zero after reinstall, want a fresh timestamp")
	}
}

// TestInstall_WritesLanguageToAllAssistantsMeta verifies that the language
// chosen in the TUI is recorded in every assistant's install metadata.
// Uses agentEnabledFS: its root-level SKILL.md materializes the asdt/
// directory the install-meta file lives in.
func TestInstall_WritesLanguageToAllAssistantsMeta(t *testing.T) {
	assistants := []installer.AssistantDescriptor{
		{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(t.TempDir(), "skills")},
		{ID: "a2", Name: "A2", BinaryName: "sh", SkillsDir: filepath.Join(t.TempDir(), "skills")},
	}
	provider := installer.Providers[0]

	results := installer.Install(assistants, provider, agentEnabledFS, "es")
	for i, r := range results {
		if r.Err != nil {
			t.Fatalf("install failed for assistant %d: %v", i, r.Err)
		}
	}

	for _, a := range assistants {
		meta, err := installer.ReadInstallMeta(a)
		if err != nil {
			t.Fatalf("read meta for %s: %v", a.ID, err)
		}
		if meta.Language != "es" {
			t.Errorf("meta.Language for %s = %q, want %q", a.ID, meta.Language, "es")
		}
	}
}

// TestInstall_ReinstallWithEmptyLangPreservesLanguage verifies the
// wipe-prevention guard: a reinstall that passes lang="" must keep the
// previously recorded language (mirroring the Persona handling).
func TestInstall_ReinstallWithEmptyLangPreservesLanguage(t *testing.T) {
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: filepath.Join(t.TempDir(), "skills")}
	provider := installer.Providers[0]

	results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, agentEnabledFS, "es")
	if results[0].Err != nil {
		t.Fatalf("first install failed: %v", results[0].Err)
	}

	results = installer.Install([]installer.AssistantDescriptor{assistant}, provider, agentEnabledFS, "")
	if results[0].Err != nil {
		t.Fatalf("reinstall failed: %v", results[0].Err)
	}

	meta, err := installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta after reinstall: %v", err)
	}
	if meta.Language != "es" {
		t.Errorf("meta.Language = %q after reinstall with lang=\"\", want %q (must be preserved)", meta.Language, "es")
	}
}

func assertAgentTypesCanonical(t *testing.T, got []string) {
	t.Helper()
	want := installer.AgentTypeNames
	if len(got) != len(want) {
		t.Errorf("meta.AgentTypes = %v, want canonical %v", got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("meta.AgentTypes[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %q to exist: %v", path, err)
	}
}

// pruneV1FS and pruneV2FS model two releases of the embedded skill tree:
// v2 drops sample/old/extra.md, so upgrading from v1 must prune it.
var pruneV1FS = fstest.MapFS{
	"SKILL.md":            &fstest.MapFile{Data: []byte("# ASDT consultant")},
	"sample/SKILL.md":     &fstest.MapFile{Data: []byte("# Sample skill v1")},
	"sample/old/extra.md": &fstest.MapFile{Data: []byte("# Dropped in v2")},
}

var pruneV2FS = fstest.MapFS{
	"SKILL.md":        &fstest.MapFile{Data: []byte("# ASDT consultant")},
	"sample/SKILL.md": &fstest.MapFile{Data: []byte("# Sample skill v2")},
}

// TestInstall_PrunesStaleFilesOnUpgrade verifies the manifest-based prune:
// upgrading to an embedded tree that no longer contains a file removes that
// file from disk, reports it in Removed, and records the fresh manifest.
func TestInstall_PrunesStaleFilesOnUpgrade(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root}
	provider := installer.Providers[0]

	results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV1FS, "")
	if results[0].Err != nil {
		t.Fatalf("v1 install failed: %v", results[0].Err)
	}
	checkFile(t, filepath.Join(root, "sample", "old", "extra.md"))

	results = installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
	if results[0].Err != nil {
		t.Fatalf("v2 install failed: %v", results[0].Err)
	}

	checkFileAbsent(t, filepath.Join(root, "sample", "old", "extra.md"))
	checkFileAbsent(t, filepath.Join(root, "sample", "old")) // emptied dir cleaned up
	checkFile(t, filepath.Join(root, "sample", "SKILL.md"))

	want := filepath.Join("sample", "old", "extra.md")
	if len(results[0].Removed) != 1 || results[0].Removed[0] != want {
		t.Errorf("Removed = %v, want [%s]", results[0].Removed, want)
	}

	meta, err := installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	wantFiles := map[string]bool{
		filepath.Join("asdt", "SKILL.md"):   true,
		filepath.Join("sample", "SKILL.md"): true,
	}
	if len(meta.Files) != len(wantFiles) {
		t.Fatalf("meta.Files = %v, want exactly %v", meta.Files, wantFiles)
	}
	for _, f := range meta.Files {
		if !wantFiles[f] {
			t.Errorf("meta.Files contains unexpected entry %q", f)
		}
	}
}

// TestInstall_LegacyMetaFallbackPrune verifies the fallback scan: when the
// existing meta predates the Files manifest, stale files inside managed
// roots are still detected by walking the disk.
func TestInstall_LegacyMetaFallbackPrune(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root}
	provider := installer.Providers[0]

	results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
	if results[0].Err != nil {
		t.Fatalf("install failed: %v", results[0].Err)
	}

	// Simulate a legacy install: meta without the Files manifest, plus a
	// stale file inside a managed root.
	meta, err := installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	meta.Files = nil
	if err := installer.WriteInstallMeta(assistant, meta); err != nil {
		t.Fatalf("write legacy meta: %v", err)
	}
	stale := filepath.Join(root, "sample", "stale.md")
	if err := os.WriteFile(stale, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	results = installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
	if results[0].Err != nil {
		t.Fatalf("reinstall failed: %v", results[0].Err)
	}

	checkFileAbsent(t, stale)
	want := filepath.Join("sample", "stale.md")
	if len(results[0].Removed) != 1 || results[0].Removed[0] != want {
		t.Errorf("Removed = %v, want [%s]", results[0].Removed, want)
	}
	// The meta dotfile must survive the fallback scan.
	meta, err = installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta after fallback prune: %v", err)
	}
	if len(meta.Files) == 0 {
		t.Error("meta.Files is empty after reinstall; the fresh manifest must be recorded")
	}
}

// TestInstall_SecondIdenticalInstallRemovesNothing verifies prune idempotency:
// reinstalling the same tree yields an empty Removed.
func TestInstall_SecondIdenticalInstallRemovesNothing(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root}
	provider := installer.Providers[0]

	results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
	if results[0].Err != nil {
		t.Fatalf("first install failed: %v", results[0].Err)
	}
	results = installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
	if results[0].Err != nil {
		t.Fatalf("second install failed: %v", results[0].Err)
	}
	if len(results[0].Removed) != 0 {
		t.Errorf("Removed = %v after identical reinstall, want empty", results[0].Removed)
	}
}

// TestInstall_PruneLeavesNonManagedSiblingsAlone verifies that user content
// living next to the managed roots is never touched by the prune.
func TestInstall_PruneLeavesNonManagedSiblingsAlone(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root}
	provider := installer.Providers[0]

	userFile := filepath.Join(root, "my-skills", "notes.md")
	if err := os.MkdirAll(filepath.Dir(userFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userFile, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	for range 2 { // first install + reinstall, both must leave it alone
		results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
		if results[0].Err != nil {
			t.Fatalf("install failed: %v", results[0].Err)
		}
		if len(results[0].Removed) != 0 {
			t.Errorf("Removed = %v, want empty", results[0].Removed)
		}
	}
	checkFile(t, userFile)
}

// TestInstall_MetaPreservedFieldsSurvivePruneManifest verifies that adding
// the Files manifest does not disturb the preserve-or-set meta discipline.
func TestInstall_MetaPreservedFieldsSurvivePruneManifest(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	assistant := installer.AssistantDescriptor{ID: "a1", Name: "A1", BinaryName: "sh", SkillsDir: root}
	provider := installer.Providers[0]

	results := installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "es")
	if results[0].Err != nil {
		t.Fatalf("first install failed: %v", results[0].Err)
	}

	meta, err := installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	meta.Persona = "Sky"
	meta.Emojis = "yes"
	if err := installer.WriteInstallMeta(assistant, meta); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	results = installer.Install([]installer.AssistantDescriptor{assistant}, provider, pruneV2FS, "")
	if results[0].Err != nil {
		t.Fatalf("reinstall failed: %v", results[0].Err)
	}

	meta, err = installer.ReadInstallMeta(assistant)
	if err != nil {
		t.Fatalf("read meta after reinstall: %v", err)
	}
	if meta.Persona != "Sky" {
		t.Errorf("meta.Persona = %q, want %q (must be preserved)", meta.Persona, "Sky")
	}
	if meta.Emojis != "yes" {
		t.Errorf("meta.Emojis = %q, want %q (must be preserved)", meta.Emojis, "yes")
	}
	if meta.Language != "es" {
		t.Errorf("meta.Language = %q, want %q (must be preserved)", meta.Language, "es")
	}
	if len(meta.Files) == 0 {
		t.Error("meta.Files is empty after reinstall, want the recorded manifest")
	}
}
