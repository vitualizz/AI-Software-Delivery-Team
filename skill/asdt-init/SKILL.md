---
name: asdt-init
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

## Orchestration

This is a light, mostly-mechanical flow — but it has one gate only the orchestrator can pass correctly, and everything downstream depends on it.

**Resolve Engram presence yourself, first, before delegating anything** (Step 2's detection). "Does THIS session have Engram's memory tools" is a question about the orchestrator's own tool list — the one the user is actually relying on for every other specialist. A sub-agent has its own tool list (narrower for specialized agent types, full for `general-purpose`); asking it risks a false "absent" when Engram is actually present in the session that matters. It costs nothing to check yourself — you're inspecting your own tools, not running commands or reading files.

- **Absent** → stop right here and tell the user, exactly as Step 2 describes. Nothing downstream can run without it — don't launch a sub-agent only to have it discover the same dead end.
- **Present** → delegate the rest to ONE sub-agent, passing "Engram confirmed present" into its prompt as an established fact:
  - Stack detection (Step 1)
  - The idempotency check and file writes (Step 3)
  - The confirmation message (Step 4)

It returns a short summary of what it found and wrote — keeping the bash output, file reads, and intermediate reasoning out of your main context, which is the whole point of routing work through ASDT specialists in the first place.

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

### Step 2 — Detect the memory provider

**Detect Engram yourself — do not ask the user to confirm something you can observe directly.** Check your own current tool list for Engram's memory tools (`mem_save`, `mem_search`, `mem_context`, etc. — Claude Code exposes them prefixed as `mcp__plugin_engram_engram__mem_*`; other host assistants may expose the same tools under a different prefix or none).

- If they're present → Engram is installed and reachable. Tell the user so and continue.
- If they're absent → tell the user Engram is required for ASDT's cross-session memory and is not reachable in this session, explain how to install/connect it, and STOP. Do not write `.asdt/config.yaml` with `provider: engram` when the provider isn't actually present — that would silently point every future specialist at a memory backend that doesn't exist.

### Step 3 — Write configuration files

`.asdt/` holds static reference data — bootstrapped once and refreshed only on a deliberate recalibration, never per-change. It is not where in-progress work state lives.

**Check for an existing setup first — never overwrite silently.** Look for `.asdt/config.yaml`:
- Absent → this is a first-time setup. Proceed to create the files below.
- Present → this project was already initialized. List what already exists under `.asdt/` (`config.yaml`, and any of `platform.yaml` / `platform-summary.yaml` found under `.asdt/knowledge/`) and ask whether the user wants to recalibrate — re-scan and overwrite — or leave it as-is. Re-running init is for refreshing stale base info, not for silently discarding a working setup.

Create `.asdt/config.yaml`:
```yaml
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
- Detected stack and platform info written to `.asdt/knowledge/`
- They can now use `/asdt-architect`, `/asdt-developer`, etc.
