# Contributing to ASDT

The most impactful contributions are specialist SKILL.md files and workflow definitions — you don't need Go expertise. The skill layer IS the product. If you can describe a specialist's role, its workflow steps, and its artifact contracts, you can ship a new specialist.

---

## How to add a new specialist

### 1. Create the directory structure

```
skill/{name}/
  SKILL.md          # specialist definition and workflow
  workflow.yaml     # step sequence and metadata
  steps/            # one .md per workflow step
  skills/           # specialist-scoped capability fragments (optional)
```

### 2. Write the skill file

`skill/{name}/SKILL.md` must have valid YAML frontmatter:

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

## Prerequisites
- `.asdt/config.yaml` has `memory.provider` set
- Engram MCP server is active

## Input Contract
...

## Workflow
### Step 1 — ...
...

## Output Contract
...
```

### 3. Write workflow.yaml

```yaml
specialist: {name}
steps:
  - id: step-one
    name: Step One
    shared_skills:
      - platform-context
  - id: step-two
    name: Step Two
```

### 4. Write step files

Create one `.md` per step in `skill/{name}/steps/{step-id}.md`. Each file contains the LLM instructions for that step.

### 5. Add specialist-scoped skill fragments (optional)

Create `skill/{name}/skills/*.md` for capability fragments specific to this specialist (e.g. `threat-modeling.md`, `code-generation.md`). Reference them in the `shared-skills` frontmatter field.

### 6. Add shared skills if the capability is reusable

If a fragment is useful across multiple specialists, create `skill/_shared/skills/{name}.md` instead and reference it in the `shared-skills` frontmatter of any specialist that needs it.

### 7. Update README.md

Add a row to the specialists table with the command, role, and what it produces.

No Go code changes are needed to add a specialist.

---

## How to add a shared skill (no Go needed)

1. Create `skill/_shared/skills/{name}.md` with the capability instructions.
2. Reference it in the `shared-skills` frontmatter of any specialist SKILL.md that needs it.
3. Open a PR.

---

## How to improve a specialist prompt

1. Edit `skill/{specialist}/SKILL.md` or any file under `skill/{specialist}/steps/` or `skill/{specialist}/skills/`.
2. Run `go test ./internal/prompt/...` — the `go:embed` registry picks up changes automatically.
3. Prompt-only PRs are first-class contributions.

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
- `go test ./...` must pass.
- If your change modifies an artifact schema, update golden fixtures in `testdata/golden/` to match.
- Prompt changes (SKILL.md or step file edits) automatically affect `prompt_version` hashes in artifact envelopes — this is expected. Do not manually set `prompt_version` in fixture files.
