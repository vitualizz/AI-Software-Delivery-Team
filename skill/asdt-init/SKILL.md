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

### Step 2 — Ask configuration questions
Ask the user:
1. What is the name of the first change you want to work on? (e.g. `add-user-auth`)
2. Confirm memory provider: ASDT uses Engram for cross-session memory. Is Engram installed and running?

### Step 3 — Write configuration files
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
