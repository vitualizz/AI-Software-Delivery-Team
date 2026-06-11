package installer

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
)

func TestUnderManagedRoot(t *testing.T) {
	roots := []string{"asdt", "asdt-shared"}
	cases := []struct {
		name string
		rel  string
		want bool
	}{
		{name: "file inside managed root", rel: "asdt/SKILL.md", want: true},
		{name: "nested file inside managed root", rel: "asdt-shared/skills/x.md", want: true},
		{name: "absolute path", rel: "/etc/passwd", want: false},
		{name: "parent escape", rel: "asdt/../../evil.md", want: false},
		{name: "leading parent escape", rel: "../outside.md", want: false},
		{name: "outside any managed root", rel: "my-skills/notes.md", want: false},
		{name: "exact managed root", rel: "asdt", want: false},
		{name: "empty path", rel: "", want: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := underManagedRoot(c.rel, roots); got != c.want {
				t.Errorf("underManagedRoot(%q) = %v, want %v", c.rel, got, c.want)
			}
		})
	}
}

func TestHasDotComponent(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{rel: "asdt/.install-meta.json", want: true},
		{rel: ".hidden/file.md", want: true},
		{rel: "asdt/.cache/x.md", want: true},
		{rel: "asdt/SKILL.md", want: false},
		{rel: "asdt-shared/skills/x.md", want: false},
	}
	for _, c := range cases {
		if got := hasDotComponent(c.rel); got != c.want {
			t.Errorf("hasDotComponent(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
}

func TestRelativizeWritten(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")
	written := []string{
		filepath.Join(skillsDir, "asdt", "SKILL.md"),
		filepath.Join(skillsDir, "asdt-shared", "skills", "x.md"),
		filepath.Join(skillsDir, "..", "outside.md"), // escapes skillsDir → dropped
	}

	got := relativizeWritten(skillsDir, written)
	want := []string{
		filepath.Join("asdt", "SKILL.md"),
		filepath.Join("asdt-shared", "skills", "x.md"),
	}
	if !slices.Equal(got, want) {
		t.Errorf("relativizeWritten = %v, want %v", got, want)
	}
}

func TestManagedRootsFor(t *testing.T) {
	// Directory entries map verbatim; the loose root file contributes "asdt";
	// a literal "asdt/" entry must dedupe against the loose-file mapping.
	skillsFS := fstest.MapFS{
		"SKILL.md":                &fstest.MapFile{Data: []byte("# consultant")},
		"asdt/extra.md":           &fstest.MapFile{Data: []byte("# extra")},
		"asdt-architect/SKILL.md": &fstest.MapFile{Data: []byte("# architect")},
		"asdt-shared/skills/x.md": &fstest.MapFile{Data: []byte("# shared")},
	}

	got := managedRootsFor(skillsFS)
	slices.Sort(got)
	want := []string{"asdt", "asdt-architect", "asdt-shared"}
	if !slices.Equal(got, want) {
		t.Errorf("managedRootsFor = %v, want %v", got, want)
	}
}

// seedFile creates a file (and its parents) under dir at the given relative path.
func seedFile(t *testing.T, dir, rel string) {
	t.Helper()
	target := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPruneStale_ManifestDiff(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/SKILL.md")
	seedFile(t, skillsDir, "asdt/old.md")

	prev := []string{"asdt/SKILL.md", "asdt/old.md"}
	current := []string{"asdt/SKILL.md"}

	removed := pruneStale(skillsDir, []string{"asdt"}, prev, current)

	if !slices.Equal(removed, []string{"asdt/old.md"}) {
		t.Errorf("removed = %v, want [asdt/old.md]", removed)
	}
	checkFileAbsentInternal(t, filepath.Join(skillsDir, "asdt", "old.md"))
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", "SKILL.md"))
}

func TestPruneStale_FallbackScan(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/SKILL.md")
	seedFile(t, skillsDir, "asdt/stale.md")
	seedFile(t, skillsDir, "asdt/.install-meta.json")

	// Legacy meta: no manifest (prev nil) → fallback walk of managed roots.
	removed := pruneStale(skillsDir, []string{"asdt"}, nil, []string{"asdt/SKILL.md"})

	if !slices.Equal(removed, []string{"asdt/stale.md"}) {
		t.Errorf("removed = %v, want [asdt/stale.md]", removed)
	}
	checkFileAbsentInternal(t, filepath.Join(skillsDir, "asdt", "stale.md"))
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", "SKILL.md"))
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", ".install-meta.json"))
}

func TestPruneStale_FallbackNeverVisitsNonManagedSiblings(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/SKILL.md")
	seedFile(t, skillsDir, "my-skills/notes.md")

	removed := pruneStale(skillsDir, []string{"asdt"}, nil, []string{"asdt/SKILL.md"})

	if len(removed) != 0 {
		t.Errorf("removed = %v, want empty", removed)
	}
	checkFileInternal(t, filepath.Join(skillsDir, "my-skills", "notes.md"))
}

func TestPruneStale_DotfileManifestEntryPreserved(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/.install-meta.json")

	// A manifest that somehow lists the dotfile must never remove it.
	removed := pruneStale(skillsDir, []string{"asdt"}, []string{"asdt/.install-meta.json"}, []string{"asdt/SKILL.md"})

	if len(removed) != 0 {
		t.Errorf("removed = %v, want empty", removed)
	}
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", ".install-meta.json"))
}

func TestPruneStale_OutOfRootManifestEntriesIgnored(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "my-skills/notes.md")
	seedFile(t, skillsDir, "asdt/SKILL.md")

	prev := []string{
		"../escape.md",
		"/abs/path.md",
		"my-skills/notes.md", // exists but outside every managed root
		"asdt",               // exact root, never a candidate
	}
	removed := pruneStale(skillsDir, []string{"asdt"}, prev, []string{"asdt/SKILL.md"})

	if len(removed) != 0 {
		t.Errorf("removed = %v, want empty", removed)
	}
	checkFileInternal(t, filepath.Join(skillsDir, "my-skills", "notes.md"))
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", "SKILL.md"))
}

func TestPruneStale_EmptyDirCleanup(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt-shared/skills/deep/old.md")
	seedFile(t, skillsDir, "asdt-shared/keep.md")

	prev := []string{"asdt-shared/skills/deep/old.md", "asdt-shared/keep.md"}
	current := []string{"asdt-shared/keep.md"}

	removed := pruneStale(skillsDir, []string{"asdt-shared"}, prev, current)

	if !slices.Equal(removed, []string{"asdt-shared/skills/deep/old.md"}) {
		t.Errorf("removed = %v, want [asdt-shared/skills/deep/old.md]", removed)
	}
	// Emptied parents below the managed root are gone; the root itself survives.
	checkFileAbsentInternal(t, filepath.Join(skillsDir, "asdt-shared", "skills", "deep"))
	checkFileAbsentInternal(t, filepath.Join(skillsDir, "asdt-shared", "skills"))
	checkFileInternal(t, filepath.Join(skillsDir, "asdt-shared"))
	checkFileInternal(t, filepath.Join(skillsDir, "asdt-shared", "keep.md"))
}

func TestPruneStale_DirWithDotfileSurvivesCleanup(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/sub/old.md")
	seedFile(t, skillsDir, "asdt/sub/.keep")

	removed := pruneStale(skillsDir, []string{"asdt"}, []string{"asdt/sub/old.md"}, nil)

	if !slices.Equal(removed, []string{"asdt/sub/old.md"}) {
		t.Errorf("removed = %v, want [asdt/sub/old.md]", removed)
	}
	// The dir still holds the dotfile → ENOTEMPTY stops the cleanup chain.
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", "sub", ".keep"))
}

func TestPruneStale_MissingFileIsBestEffort(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/SKILL.md")

	// "asdt/ghost.md" was in the old manifest but is already gone from disk.
	removed := pruneStale(skillsDir, []string{"asdt"}, []string{"asdt/ghost.md"}, []string{"asdt/SKILL.md"})

	if len(removed) != 0 {
		t.Errorf("removed = %v, want empty (failed removals must not be reported)", removed)
	}
}

func TestPruneStale_NoManagedRootsSkipsPrune(t *testing.T) {
	skillsDir := t.TempDir()
	seedFile(t, skillsDir, "asdt/old.md")

	removed := pruneStale(skillsDir, nil, []string{"asdt/old.md"}, nil)

	if len(removed) != 0 {
		t.Errorf("removed = %v, want empty when managed roots are unknown", removed)
	}
	checkFileInternal(t, filepath.Join(skillsDir, "asdt", "old.md"))
}

func checkFileInternal(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("expected %q to exist: %v", path, err)
	}
}

func checkFileAbsentInternal(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err == nil {
		t.Errorf("expected %q to NOT exist, but it does", path)
	}
}
