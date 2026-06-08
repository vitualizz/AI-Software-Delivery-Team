package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

var realisticSpecialistFS = fstest.MapFS{
	"asdt-developer/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: asdt-developer
description: "Turns specs and designs into working code — implementation plans, production code, and test suites."
user-invocable: true
specialist-id: developer
shared-skills:
  - platform-context
  - artifact-envelope
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# Developer specialist
`)},
	"asdt-architect/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: asdt-architect
description: "Makes architecture decisions and produces ADRs, system design, and API design artifacts."
user-invocable: true
specialist-id: architect
shared-skills:
  - platform-context
  - artifact-envelope
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# Architect specialist
`)},
	"asdt-shared/skills/x.md": &fstest.MapFile{Data: []byte("# Shared fragment")},
}

var malformedFrontmatterFS = fstest.MapFS{
	"asdt-developer/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: asdt-developer
description: "Turns specs and designs into working code — implementation plans, production code, and test suites."
user-invocable: true
specialist-id: developer
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# Developer specialist
`)},
	"asdt-broken/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: asdt-broken
description: "A specialist with a missing required frontmatter field."
user-invocable: true
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# Broken specialist
`)},
}

var prefixedSpecialistIDFS = fstest.MapFS{
	"asdt-init/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: asdt-init
description: "Sets up the ground ASDT stands on."
user-invocable: true
specialist-id: asdt-init
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# ASDT Init
`)},
}

func readWrapperContents(t *testing.T, paths []string) map[string]string {
	t.Helper()
	contents := make(map[string]string, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %q: %v", path, err)
		}
		contents[path] = string(data)
	}
	return contents
}

func TestGenerateOpenCodeCommands_ProducesCorrectlyDerivedWrappers(t *testing.T) {
	cases := []struct {
		name           string
		wantFile       string
		wantDescr      string
		wantNameInBody string
		wantIDInBody   string
	}{
		{name: "developer", wantFile: "asdt-developer.md", wantDescr: "Turns specs and designs into working code — implementation plans, production code, and test suites.", wantNameInBody: "asdt-developer", wantIDInBody: "developer"},
		{name: "architect", wantFile: "asdt-architect.md", wantDescr: "Makes architecture decisions and produces ADRs, system design, and API design artifacts.", wantNameInBody: "asdt-architect", wantIDInBody: "architect"},
	}
	dir := t.TempDir()
	commandRoot := filepath.Join(dir, "commands")
	written, err := generateOpenCodeCommands(realisticSpecialistFS, commandRoot)
	if err != nil {
		t.Fatalf("generateOpenCodeCommands returned error: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("expected 2 written wrappers (asdt-shared has no SKILL.md), got %d: %v", len(written), written)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			target := filepath.Join(commandRoot, c.wantFile)
			data, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatalf("expected wrapper %q to exist: %v", target, readErr)
			}
			content := string(data)
			if !strings.Contains(content, "description: \""+c.wantDescr+"\"") {
				t.Errorf("wrapper %q missing verbatim description %q\ngot:\n%s", c.wantFile, c.wantDescr, content)
			}
			if !strings.Contains(content, "agent: build") {
				t.Errorf("wrapper %q missing %q", c.wantFile, "agent: build")
			}
			if !strings.Contains(content, "subtask: false") {
				t.Errorf("wrapper %q missing %q", c.wantFile, "subtask: false")
			}
			if !strings.Contains(content, c.wantNameInBody) {
				t.Errorf("wrapper %q body missing specialist name %q", c.wantFile, c.wantNameInBody)
			}
			if !strings.Contains(content, c.wantIDInBody) {
				t.Errorf("wrapper %q body missing specialist-id %q", c.wantFile, c.wantIDInBody)
			}
		})
	}
	checkFileAbsent(t, filepath.Join(commandRoot, "asdt-shared.md"))
}

func TestGenerateOpenCodeCommands_DerivesFilenameFromDirNotSpecialistID(t *testing.T) {
	dir := t.TempDir()
	commandRoot := filepath.Join(dir, "commands")
	written, err := generateOpenCodeCommands(prefixedSpecialistIDFS, commandRoot)
	if err != nil {
		t.Fatalf("generateOpenCodeCommands returned error: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 written wrapper, got %d: %v", len(written), written)
	}
	wantPath := filepath.Join(commandRoot, "asdt-init.md")
	if written[0] != wantPath {
		t.Errorf("written[0] = %q, want %q (derived from dirName, not \"asdt-\"+specialist-id)", written[0], wantPath)
	}
	checkFile(t, wantPath)
	checkFileAbsent(t, filepath.Join(commandRoot, "asdt-asdt-init.md"))
}

func TestCommandAdapters_ClaudeCodeHasNoAdapter(t *testing.T) {
	cases := []struct {
		name      string
		id        AssistantID
		wantFound bool
	}{
		{name: "claude code is absent", id: AssistantClaudeCode, wantFound: false},
		{name: "opencode is registered", id: AssistantOpenCode, wantFound: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			descriptor, ok := adapterFor(c.id)
			if ok != c.wantFound {
				t.Fatalf("adapterFor(%q) found = %v, want %v", c.id, ok, c.wantFound)
			}
			if !c.wantFound {
				return
			}
			if descriptor.AssistantID != c.id {
				t.Errorf("adapterFor(%q).AssistantID = %q, want %q", c.id, descriptor.AssistantID, c.id)
			}
			if descriptor.Generate == nil {
				t.Errorf("adapterFor(%q).Generate is nil, want non-nil", c.id)
			}
		})
	}
	for _, adapter := range CommandAdapters {
		if adapter.AssistantID == AssistantClaudeCode {
			t.Fatalf("CommandAdapters contains an entry for %q; Claude Code must be intentionally absent", AssistantClaudeCode)
		}
	}
}

func TestGenerateOpenCodeCommands_Idempotent(t *testing.T) {
	dir := t.TempDir()
	commandRoot := filepath.Join(dir, "commands")
	written1, err1 := generateOpenCodeCommands(realisticSpecialistFS, commandRoot)
	if err1 != nil {
		t.Fatalf("first run: generateOpenCodeCommands returned error: %v", err1)
	}
	contentsAfterFirstRun := readWrapperContents(t, written1)
	written2, err2 := generateOpenCodeCommands(realisticSpecialistFS, commandRoot)
	if err2 != nil {
		t.Fatalf("second run: generateOpenCodeCommands returned error: %v", err2)
	}
	if len(written1) != len(written2) {
		t.Fatalf("written paths differ in count across runs: first=%d, second=%d", len(written1), len(written2))
	}
	for i := range written1 {
		if written1[i] != written2[i] {
			t.Errorf("written[%d] differs across runs: first=%q, second=%q", i, written1[i], written2[i])
		}
	}
	contentsAfterSecondRun := readWrapperContents(t, written2)
	for path, want := range contentsAfterFirstRun {
		got, ok := contentsAfterSecondRun[path]
		if !ok {
			t.Errorf("wrapper %q present after first run is missing after second run", path)
			continue
		}
		if got != want {
			t.Errorf("wrapper %q changed across runs (not byte-identical)", path)
		}
	}
	entries, readErr := os.ReadDir(commandRoot)
	if readErr != nil {
		t.Fatalf("read command root: %v", readErr)
	}
	if len(entries) != len(written2) {
		t.Errorf("command root contains %d entries, want %d (no stale/duplicate files)", len(entries), len(written2))
	}
}

func TestGenerateOpenCodeCommands_MalformedFrontmatterIsPartialSuccess(t *testing.T) {
	dir := t.TempDir()
	commandRoot := filepath.Join(dir, "commands")
	written, err := generateOpenCodeCommands(malformedFrontmatterFS, commandRoot)
	if err == nil {
		t.Fatal("expected a non-nil error for the malformed specialist, got nil")
	}
	if !strings.Contains(err.Error(), "asdt-broken") {
		t.Errorf("error %q does not name the offending specialist directory", err.Error())
	}
	if len(written) != 1 {
		t.Fatalf("expected exactly 1 wrapper written despite the malformed sibling, got %d: %v", len(written), written)
	}
	wantPath := filepath.Join(commandRoot, "asdt-developer.md")
	if written[0] != wantPath {
		t.Errorf("written[0] = %q, want %q", written[0], wantPath)
	}
	checkFile(t, wantPath)
	checkFileAbsent(t, filepath.Join(commandRoot, "asdt-broken.md"))
}

func TestInstallOne_ClaudeCodeUnaffectedByCommandAdapters(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "skills")
	assistants := []AssistantDescriptor{{ID: AssistantClaudeCode, Name: "Claude Code", BinaryName: "sh", SkillsDir: root}}
	provider := Providers[0]
	results := Install(assistants, provider, realisticSpecialistFS)
	if results[0].Err != nil {
		t.Fatalf("install failed: %v", results[0].Err)
	}
	if results[0].WrittenCommands != nil {
		t.Errorf("Claude Code WrittenCommands = %v, want nil (no CommandAdapter entry => no-op)", results[0].WrittenCommands)
	}
	if len(results[0].Written) == 0 {
		t.Error("Claude Code Written is empty — skill-tree copy must be unaffected by the adapter hook")
	}
}

func TestCommandRootFor(t *testing.T) {
	cases := []struct {
		name           string
		id             AssistantID
		wantNonEmpty   bool
		wantPathSuffix string
	}{
		{name: "opencode resolves to its commands directory", id: AssistantOpenCode, wantNonEmpty: true, wantPathSuffix: "/opencode/commands"},
		{name: "claude code has no known command root", id: AssistantClaudeCode, wantNonEmpty: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := commandRootFor(c.id)
			if c.wantNonEmpty && got == "" {
				t.Fatalf("commandRootFor(%q) = %q, want a non-empty path", c.id, got)
			}
			if !c.wantNonEmpty && got != "" {
				t.Fatalf("commandRootFor(%q) = %q, want \"\" (no known command root)", c.id, got)
			}
			if c.wantPathSuffix != "" && !strings.HasSuffix(got, c.wantPathSuffix) {
				t.Errorf("commandRootFor(%q) = %q, want it to end in %q", c.id, got, c.wantPathSuffix)
			}
		})
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %q to exist: %v", path, err)
	}
}

func checkFileAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file %q to NOT exist, but it does", path)
	}
}
