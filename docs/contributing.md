# Contributing to ASDT

The most impactful contributions to ASDT are specialist SKILL.md files and
descriptor improvements — you don't need Go expertise. The skill layer IS the
product. If you can describe a specialist's role, its workflow steps, and its
artifact contracts, you can ship a new specialist.

---

## How to add a new specialist

### 1. Write the skill file

Create `skill/{name}/SKILL.md` with the following structure:

```markdown
---
name: asdt:{name}
description: "Trigger: ..."
user-invocable: true
specialist-id: {name}
shared-skills:
  - platform-context
  - artifact-envelope
  - scope-definition
---

# {Name} Specialist

## Role
...

## Invariants
...

## Input Contract
...

## Workflow
### Step 1 — ...
...

## Output Contract
...
```

### 2. Add specialist-scoped skill fragments

Create `skill/{name}/skills/*.md` for any capability fragments specific to this
specialist (e.g. `code-generation.md`, `threat-modeling.md`).

### 3. Add shared skills if the capability is reusable

If the fragment is useful across multiple specialists, create
`skill/_shared/skills/{name}.md` instead and reference it in the `shared-skills`
frontmatter field of any specialist that needs it.

### 4. Add a descriptor in Go (for the TUI binary)

Add `func {Name}Descriptor() SpecialistDescriptor` to
`internal/specialists/descriptor.go`. A descriptor is a single struct literal —
no new packages, no new switch arms.

```go
func MySpecialistDescriptor() SpecialistDescriptor {
    return SpecialistDescriptor{
        ID:          "my-specialist",
        Name:        "My Specialist",
        Description: "...",
        Reads:       []string{"ux-brief", "system-design"},
        Writes:      []string{"my-artifact"},
        Workflow: []WorkflowStep{
            {ID: "step-one", Name: "Step One", SharedSkills: []string{"platform-context"}},
        },
    }
}
```

### 5. Register the descriptor

Add one entry to the descriptors map in `cmd/asdt/main.go`:

```go
descriptors["my-specialist"] = specialists.MySpecialistDescriptor()
```

### 6. Add a unit test

Add a row to the table test in `internal/specialists/descriptor_test.go`. Test
that `Validate()` passes and that step IDs are correct.

### 7. Update the specialists table in `README.md`

Add a row to the specialists table with the command, role, and what it produces.

---

## How to add a shared skill (no Go needed)

1. Create `skill/_shared/skills/{name}.md` with the capability instructions.
2. Reference it in the `Skills []string` field of relevant `SpecialistDescriptor`
   values or in the `shared-skills` frontmatter of specialist SKILL.md files.
3. Open a PR — no Go changes needed.

---

## How to improve a specialist prompt

1. Edit `skill/{specialist}/SKILL.md` or any file under
   `skill/{specialist}/skills/`.
2. Run `go test ./internal/prompt/...` — the `go:embed` registry picks up
   changes automatically.
3. PR with prompt-only changes are first-class contributions.

---

## Code standards

- Early return: `if err != nil { return err }` — validate inputs first.
- No global state — constructor injection throughout.
- Interfaces defined close to consumers, not in the implementing package.
- No `utils/`, `helpers/`, `common/`, or `misc/` packages — domain nouns only.
- Table-driven tests for any logic with more than two cases.

---

## PR process

- One logical change per PR.
- Both `go test ./...` and `go test -tags=parity ./internal/specialists/...`
  must pass.
- If your change modifies an artifact schema, update golden fixtures in
  `testdata/golden/` to match.
- Prompt changes (SKILL.md edits) automatically affect `prompt_version` hashes
  in artifact envelopes — this is expected. Do not manually set
  `prompt_version` in fixture files.
