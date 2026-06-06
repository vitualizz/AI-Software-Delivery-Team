package main

import (
	"io/fs"
	"os/exec"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/skill"
)

func TestBuild_ExitsZero(t *testing.T) {
	cmd := exec.Command("go", "build", "./cmd/asdt-tui/")
	cmd.Dir = "../.." // repo root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/asdt-tui/ failed:\n%s\n%v", out, err)
	}
}

func TestEmbeddedFS_ContainsSkillMD(t *testing.T) {
	skillsFS := skill.FS()
	found := false
	err := fs.WalkDir(skillsFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && d.Name() == "SKILL.md" {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
	if !found {
		t.Error("no SKILL.md found in embedded skill FS")
	}
}
