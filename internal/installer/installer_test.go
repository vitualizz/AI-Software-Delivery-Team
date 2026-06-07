package installer_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
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

	results := installer.Install(assistants, provider, siblingFS)
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

	results := installer.Install(assistants, provider, productionShapedFS)
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

	results := installer.Install(assistants, provider, testFS)

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

	results := installer.Install(assistants, provider, testFS)

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

	results1 := installer.Install(assistants, provider, testFS)
	if results1[0].Err != nil {
		t.Fatalf("first install failed: %v", results1[0].Err)
	}

	results2 := installer.Install(assistants, provider, testFS)
	if results2[0].Err != nil {
		t.Errorf("second install (idempotent) failed: %v", results2[0].Err)
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %q to exist: %v", path, err)
	}
}
