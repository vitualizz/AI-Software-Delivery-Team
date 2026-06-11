package skill

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

// TestEmbeddedSkillTree asserts that every asdt-* directory on disk under
// skill/ is present in the embedded FS with its SKILL.md, and that the
// top-level SKILL.md is embedded. This guards against the embed directive
// silently missing a specialist directory (e.g. a newly added asdt-<name>/).
func TestEmbeddedSkillTree(t *testing.T) {
	if _, err := fs.Stat(FS(), "SKILL.md"); err != nil {
		t.Errorf("top-level SKILL.md missing from embedded FS: %v", err)
	}

	// The installer bakes this header into generated agent definitions
	// (internal/installer/agent_adapters.go) and silently generates nothing
	// when it is absent — so its presence here is load-bearing.
	if _, err := fs.Stat(FS(), "asdt-shared/skills/executor-header.md"); err != nil {
		t.Errorf("asdt-shared/skills/executor-header.md missing from embedded FS: %v", err)
	}

	diskEntries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading skill/ on disk: %v", err)
	}

	for _, entry := range diskEntries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "asdt-") {
			continue
		}
		name := entry.Name()

		if _, err := fs.Stat(FS(), name); err != nil {
			t.Errorf("directory %s exists on disk but is missing from embedded FS: %v", name, err)
			continue
		}
		// asdt-shared has no SKILL.md; only require it when it exists on disk.
		if _, err := os.Stat(name + "/SKILL.md"); err == nil {
			if _, err := fs.Stat(FS(), name+"/SKILL.md"); err != nil {
				t.Errorf("%s/SKILL.md missing from embedded FS: %v", name, err)
			}
		}
	}
}
