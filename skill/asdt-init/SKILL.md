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
  - Project context detection (Step 4)
  - The confirmation message (Step 6)

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

### Step 4 — Detect project context

Produce `.asdt/knowledge/project-context.yaml` — a machine-written file that records _how_ the project is structured and coded (monorepo shape, test runner, naming style, architectural pattern). This is separate from `platform.yaml`, which records _what_ is installed.

#### 4.1 Check for existing project-context.yaml

Look for `{root}/knowledge/project-context.yaml`:

- **Absent** → fresh detection path (§4.2).
- **Present** → recalibration path (§4.3).

#### 4.2 Fresh detection

The sub-agent invokes the Go context detector directly (no CLI subcommand exists yet — direct library invocation):

```go
det := knowledge.DefaultContextDetector(primaryLang)
ctx, _ := det.DetectContext(projectRoot, knowledge.DetectConfig{PrimaryLanguage: primaryLang})
reader.WriteContext(root, ctx)
```

`primaryLang` comes from the first entry of `detected_stack` written in Step 3.

Each probe runs independently. A probe error is non-fatal — the sub-agent skips that probe and continues. The resulting `ProjectContext` is written to `{root}/knowledge/project-context.yaml` via `FSReader.WriteContext`.

**Detection rules — no model judgment, exact mapping only:**

**MonorepoProbe** (`is_monorepo` field):
- `os.Stat(projectRoot + "/go.work")` succeeds → `value="true"`, `source=detected`, `confidence=high`
- `os.Stat(projectRoot + "/pnpm-workspace.yaml")` succeeds → `value="true"`, `source=detected`, `confidence=high`
- Neither found → `value="false"`, `source=detected`, `confidence=high`

**TestRunnerProbe** (`test_runner` field):
- lang=`go`: read `Makefile` at root; if content contains `"go test"` → `value="make test"`, `confidence=medium`; else if `go.mod` exists → `value="go test ./..."`, `confidence=high`
- lang=`node`: read `package.json`; if `scripts.test` is non-empty → `value=scripts.test`, `confidence=high`; else if `jest.config.js` exists → `value="jest"`, `confidence=medium`; else if `vitest.config.ts` exists → `value="vitest"`, `confidence=medium`
- Any other lang or no match → `value="unknown"`, `source=inferred`, `confidence=low`

**NamingStyleDetector** (`naming_style` field):
- Collect up to 8 source files via `filepath.WalkDir` depth ≤ 3, filtered by language extension (`.go` for Go; `.ts`/`.tsx` for Node); skip files > 64 KB.
- For each file, check line-by-line with exact regexps:
  - Go: `^func [A-Z]|^type [A-Z]|^var [A-Z]|^const [A-Z]` → PascalCase hit; `^func [a-z]|^type [a-z]|^var [a-z]|^const [a-z]` → camelCase hit
  - Node: `^export (function|class|const|interface) [A-Z]` → PascalCase; `^export (function|const) [a-z]` → camelCase
- Majority rule: ≥75% files have consistent style → `confidence=high`; 50–74% → `confidence=medium`; <50% → `confidence=low`
- Go dominant PascalCase → `value="snake_case filenames, PascalCase exported symbols"` (standard Go convention confirmed)
- No source files found → `value="unknown"`, `source=inferred`, `confidence=low`

**ArchitecturalStyleDetector** (`architectural_style` field, Tier 1 only):
- `os.ReadDir(projectRoot)` — read top-level directory names only:
  - `cmd/` AND `internal/` present → `value="hexagonal"`, `source=detected`, `confidence=high`
  - `src/` present → read one level deeper (`os.ReadDir(src/)`):
    - `controllers/`, `models/`, `views/` all present → `value="mvc"`, `confidence=high`
    - `features/` OR `modules/` present → `value="modular"`, `confidence=medium`
    - Otherwise → `value="layered"`, `confidence=medium`
  - `lib/` only (no `src/`) → `value="layered"`, `confidence=medium`
  - No pattern matched → `value="unknown"`, `source=inferred`, `confidence=low`

The sub-agent returns a `DetectionSummary` alongside the Step 3 output, listing each detected field, its value, source, and confidence.

> **Forward reference:** a future change will wire `asdt-cli detect-context` as a subcommand that calls this same library. When that subcommand exists, Step 4.2 will invoke it instead of the direct library call.

#### 4.3 Recalibration (project-context.yaml already exists)

When `project-context.yaml` already exists:

1. Run fresh detection → produce `NewContext` (same rules as §4.2).
2. Compute a delta table:

   | Field | Old value | New value | Changed? |
   |---|---|---|---|
   | is_monorepo | … | … | yes/no |
   | test_runner | … | … | yes/no |
   | naming_style | … | … | yes/no |
   | architectural_style | … | … | yes/no |

3. Present the delta table to the user.
4. Ask ONE question: "Accept all changes, or review field by field?"
5. If "accept all" → overwrite `project-context.yaml` with `NewContext`.
6. If "field by field" → for each changed field, ask the user to accept / reject / set manually. One question per field.
7. **Human answers always win.** Fields where the existing `source=manual` are NEVER silently overwritten — they must appear in the delta table and require explicit user acceptance.

#### 4.4 Confidence and source rules

| Source | When to assign |
|---|---|
| `detected` | Value determined by a bounded command with direct file evidence |
| `inferred` | Pattern matched without direct file evidence (fallback / best-effort) |
| `manual` | User explicitly set this value during a recalibration review |

| Confidence | Meaning |
|---|---|
| `high` | Strong signal — treat as authoritative convention |
| `medium` | Likely match — confirm before diverging |
| `low` | Weak signal — best-effort guess |

Confidence thresholds are assigned by each probe's algorithm (see §4.2 rules). Do not reassign confidence based on judgment — use the exact rules above.

#### 4.5 Output

- `{root}/knowledge/project-context.yaml` written (fresh) or confirmed/updated (recalibration).
- Orchestrator receives a `DetectionSummary` for display to the user.
- Proceed to Step 6.

### Step 6 — Confirm
Tell the user:
- Configuration written to `.asdt/config.yaml`
- Detected stack and platform info written to `.asdt/knowledge/`
- Project context written to `.asdt/knowledge/project-context.yaml`
- They can now use `/asdt-architect`, `/asdt-developer`, etc.
