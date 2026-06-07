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

Once the provider is confirmed present, ask the user only this:
1. What is the name of the first change you want to work on? (e.g. `add-user-auth`)

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

Create `.asdt/knowledge/platform.yaml` with detected stack information.
Create `.asdt/knowledge/platform-summary.yaml` with a brief summary.

### Step 4 — Confirm
Tell the user:
- Configuration written to `.asdt/config.yaml`
- Active change set to `{change-name}`
- They can now use `/asdt:architect`, `/asdt:developer`, etc.
