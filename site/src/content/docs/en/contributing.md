---
title: Contributing
description: How to add a new specialist, improve prompts, write shared skills, and submit a PR to ASDT.
order: 8
locale: en
---

# Contributing

The most impactful contributions are specialist `SKILL.md` files and workflow step definitions — you don't need Go expertise. The skill layer IS the product. If you can describe a specialist's role, its workflow steps, and its artifact contracts, you can ship a new specialist.

## Adding a new specialist

### 1. Create the directory structure

```
skill/asdt-{name}/
  SKILL.md          # specialist definition and workflow
  workflow.yaml     # step sequence and metadata
  steps/            # one .md per workflow step
  skills/           # specialist-scoped skill fragments (optional)
```

The directory name **must** start with `asdt-`. The binary embeds the skill tree via `//go:embed SKILL.md asdt-*` in `skill/embedded.go` — any directory matching `asdt-*` ships automatically on the next build.

### 2. Write SKILL.md

```markdown
---
name: asdt-{name}
description: "One sentence: what this specialist produces."
user-invocable: true
specialist-id: {name}
shared-skills: # metadata only — documents related skills; not a loader
  - platform-context
  - artifact-envelope
metadata:
  author: "Your Name"
  version: "1.0"
---

# {Name} Specialist

## Role
...

## Orchestration Plan
...

## Invariants
...
```

`metadata` (`author` + `version`) is required in every `SKILL.md`.

### 3. Write workflow.yaml

```yaml
specialist: {name}
steps:
  - id: step-one
    name: Step One
  - id: step-two
    name: Step Two
```

### 4. Write step files

Create one `.md` per step in `skill/{name}/steps/{step-id}.md`. Each file contains the LLM instructions for that step — what to read, what to produce, what format the artifact should take.

### 5. Wire into the embed

Open `skill/embedded.go` and add your directory name to the `//go:embed` directive:

```go
//go:embed SKILL.md asdt-shared asdt-developer asdt-ux-ui asdt-architect asdt-qa asdt-security asdt-init asdt-{name}
var skillFS embed.FS
```

**Do not skip this.** A directory that exists on disk but is not named in the directive is silently excluded from the binary — no compile error, no runtime error. The skill just won't appear when users install.

### 6. Verify with the sandbox

```sh
mkdir -p /tmp/asdt-sandbox
HOME=/tmp/asdt-sandbox go run ./cmd/asdt-tui
```

Installs into a throwaway directory. Confirm your specialist appears as its own top-level sibling under `/tmp/asdt-sandbox/.claude/skills/{name}/`.

### 7. Run the embed tests

```sh
go test ./skill/...
```

`skill/embedded_test.go` verifies every `asdt-*` directory on disk is present in the embedded FS and carries a `SKILL.md`. Fails loudly if your specialist is missing.

### 8. Update this README

Add a row to the specialists table in the project README with the command, role, and what it produces.

## Improving a specialist prompt

1. Edit `skill/{specialist}/SKILL.md` or any file under `skill/{specialist}/steps/` or `skill/{specialist}/skills/`.
2. Run `go test ./skill/...` to confirm the embed registry picks up the changes.
3. Open a PR. Prompt-only PRs are first-class contributions.

## Adding a shared skill

Shared skills are capability fragments reused across multiple specialists — platform context detection, artifact envelope formatting, scope definition.

1. Create `skill/asdt-shared/skills/{name}.md` with the capability instructions.
2. Document it in the `shared-skills` frontmatter of any specialist that relates to it — that key is metadata only; actual loading happens via the FIRST ACTION Read in the `SKILL.md` body (`specialist-header`) and per-step `reference_skills` in `workflow.yaml`.
3. Open a PR.

## Code standards

- Early return: `if err != nil { return err }` — validate inputs first.
- No global state — constructor injection throughout.
- Interfaces defined close to consumers, not in the implementing package.
- No `utils/`, `helpers/`, `common/`, or `misc/` packages — domain nouns only.
- Table-driven tests for any logic with more than two cases.

## PR process

- One logical change per PR.
- `go test ./...` must pass.
- If your change modifies an artifact schema, update golden fixtures in `testdata/golden/`.
- Prompt changes automatically affect `prompt_version` hashes in artifact envelopes — this is expected. Do not manually set `prompt_version` in fixture files.
