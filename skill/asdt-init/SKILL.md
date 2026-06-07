---
name: asdt:init
description: "Sets up the ground ASDT stands on — initializes .asdt/config.yaml and wires the memory provider so every other specialist has somewhere to read from and write to."
user-invocable: true
specialist-id: asdt-init
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# ASDT Init

## Role
Initialize ASDT for the current project. Detect the project stack, collect configuration, and write `.asdt/config.yaml`.

## Prerequisites
None — this is the setup step. Run this before any other ASDT specialist.

## Workflow

### Step 1 — Detect project stack
Run ONE command to check for stack marker files at the project root — do not eyeball a directory listing or infer from visible files. The result must be identical no matter which model or session runs it:

```
fd -d 1 -t f -H '^(go\.mod|package\.json|Cargo\.toml|pyproject\.toml|requirements\.txt|Gemfile)$' .
```

Map each match deterministically — no judgment calls:
- `go.mod` → Go
- `package.json` → Node.js
- `Cargo.toml` → Rust
- `pyproject.toml` or `requirements.txt` → Python
- `Gemfile` → Ruby

A project can match more than one stack (e.g. a Go backend with a Node frontend) — list every match the command returns, in the order it returns them. If nothing matches, record an empty stack — do not guess a stack from other files.

List the detected technologies to the user.

### Step 2 — Detect the memory provider, then ask for the change name

**Detect Engram yourself — do not ask the user to confirm something you can observe directly.** Check your own current tool list for Engram's memory tools (`mem_save`, `mem_search`, `mem_context`, etc. — Claude Code exposes them prefixed as `mcp__plugin_engram_engram__mem_*`; other host assistants may expose the same tools under a different prefix or none).

- If they're present → Engram is installed and reachable. Tell the user so and continue.
- If they're absent → tell the user Engram is required for ASDT's cross-session memory and is not reachable in this session, explain how to install/connect it, and STOP. Do not write `.asdt/config.yaml` with `provider: engram` when the provider isn't actually present — that would silently point every future specialist at a memory backend that doesn't exist.

Once the provider is confirmed present, infer a default change name from the current git branch — don't make the user retype something you can read yourself:

```
git branch --show-current
```

Strip a conventional prefix if present (`feat/`, `feature/`, `fix/`, `chore/`, `refactor/` — e.g. `feat/add-user-auth` → `add-user-auth`). If the branch is `main`, `master`, `develop`, has no prefix, or the command returns nothing (detached HEAD), there's no usable default.

Ask the user only this — offering the inferred name as a default when you have one:
1. What is the name of the first change you want to work on? (e.g. `add-user-auth`{— press enter to use `{inferred-name}` from your current branch, or type a different name, when a default was inferred})

### Step 3 — Write configuration files

**Check for an existing setup first — never overwrite silently.** Look for `.asdt/config.yaml`:
- Absent → proceed to create the files below.
- Present → read it and show the user its current `active_change`. Ask explicitly whether to keep it (skip writing `config.yaml`) or replace `active_change` with the new value from Step 2. Re-running init on an already-initialized project must never silently discard a change someone is mid-way through.

Create `.asdt/config.yaml`:
```yaml
active_change: {user-provided-change-name}
memory:
  provider: engram
```

Create `.asdt/knowledge/platform.yaml`. Populate only what a single bounded command can determine deterministically — deeper analysis (naming conventions, architectural patterns) needs to sample file contents and is a different cost profile; it belongs to a dedicated future step, not to init:

```yaml
schema_version: "1"
scanned_at: {current UTC timestamp, ISO 8601}
detected_stack: {list from Step 1}
conventions:
  file_structure: {one-line description, derived below}
design_fingerprint: {}
```

To derive `conventions.file_structure`, run ONE bounded command checking for well-known top-level directories — never walk the full tree, that breaks the "same output size regardless of repo size" guarantee:

```
fd -d 1 -t d -H '^(cmd|internal|pkg|src|app|lib|components|pages|tests|spec|crates|scripts)$' .
```

Compose one short, factual sentence from the matches (e.g. `"cmd/ for binaries, internal/ for private packages"`). No matches → leave it `""` — don't invent a convention.

Leave `design_fingerprint: {}`. Identifying architectural patterns means sampling file contents, not checking presence — out of scope for init.

Create `.asdt/knowledge/platform-summary.yaml` — derived FROM the data above, never re-analyzed from scratch:

```yaml
schema_version: "1"
stack: {detected_stack}
file_structure: {conventions.file_structure}
```

Both files stay small and bounded: their size grows with the number of detected stacks, never with repo size.

### Step 4 — Confirm
Tell the user:
- Configuration written to `.asdt/config.yaml`
- Active change set to `{change-name}`
- They can now use `/asdt:architect`, `/asdt:developer`, etc.
