---
name: asdt:init
description: "Trigger: asdt init, initialize project, setup asdt, configure asdt"
user-invocable: true
specialist-id: asdt-init
---

# ASDT Init

## Role
Initialize ASDT for the current project. Detect the project stack, collect configuration, and write `.asdt/config.yaml`.

## Prerequisites
None — this is the setup step. Run this before any other ASDT specialist.

## Workflow

### Step 1 — Detect project stack
Inspect the project root for stack markers:
- `go.mod` → Go
- `package.json` → Node.js
- `Cargo.toml` → Rust
- `pyproject.toml` or `requirements.txt` → Python
- `Gemfile` → Ruby

List the detected technologies.

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
