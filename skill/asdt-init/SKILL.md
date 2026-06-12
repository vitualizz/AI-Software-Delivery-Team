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
Run ONE bounded command scanning for stack marker files down to depth 3 — do not eyeball a directory listing or infer from visible files. The result must be identical no matter which model or session runs it. The exclusion list and the result cap are part of the contract: they keep the output size independent of repo size.

```
fd -d 3 -t f -H '^(go\.mod|package\.json|Cargo\.toml|pyproject\.toml|requirements\.txt|Gemfile)$' . \
  -E node_modules -E .git -E vendor -E dist -E build -E .venv -E target \
  | awk -F/ '{printf "%d %s\n", NF, $0}' | LC_ALL=C sort -k1,1n -k2 | cut -d' ' -f2- | head -20
```

If `fd` is not available, run the equivalent `find` fallback — same exclusions, same ordering pipeline:

```
find . -maxdepth 3 \( -name node_modules -o -name .git -o -name vendor -o -name dist -o -name build -o -name .venv -o -name target \) -prune \
  -o -type f \( -name go.mod -o -name package.json -o -name Cargo.toml -o -name pyproject.toml -o -name requirements.txt -o -name Gemfile \) -print \
  | awk -F/ '{printf "%d %s\n", NF, $0}' | LC_ALL=C sort -k1,1n -k2 | cut -d' ' -f2- | head -20
```

The `awk | sort | cut` pipeline orders results deterministically: shallowest path first, then lexicographic (`LC_ALL=C`). The cap is the first 20 marker paths.

Map each marker to a canonical language id — no judgment calls:

| Marker file | Language id |
|---|---|
| `go.mod` | `go` |
| `package.json` | `node` |
| `Cargo.toml` | `rust` |
| `pyproject.toml` or `requirements.txt` | `python` |
| `Gemfile` | `ruby` |

From the ordered marker list, derive — mechanically, first occurrence wins:

- **`detected_stack`** — the language ids in scan order, deduplicated. A project can match more than one stack (e.g. a Python `backend/` with a Node `frontend/`).
- **Primary language** — the first entry of `detected_stack`.
- **Language root `{lang_root}`** — for each language, the directory containing its first marker. Step 4's probes run against these roots, not blindly against the repo root — a marker in `backend/` means that language's evidence lives in `backend/`.

If nothing matches, record an empty stack — do not guess a stack from other files.

List the detected languages and their roots to the user.

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

The sub-agent runs the probes below as bounded shell commands and writes the result to `{root}/knowledge/project-context.yaml` itself — plain file I/O, no library, no binary (per ADR-013: skills must run on any machine where they are installed, with no dependency on this repo's Go code).

Every probe has the same shape: **one bounded command, one exact mapping table, first matching row wins, no model judgment**. The tables are uniform across languages — supporting a new language means adding rows, never adding logic. If no row matches, the field is `value="unknown"`, `source=inferred`, `confidence=low`. A probe error is non-fatal — record the fallback for that field and continue.

Inputs from Step 1: `detected_stack`, the primary language, and each language's `{lang_root}`. Probes that inspect language evidence run against the primary language's `{lang_root}` — not blindly against the repo root.

**Probe: `is_monorepo`**

Check workspace markers at the repo root with one compound command:

```
ls go.work pnpm-workspace.yaml 2>/dev/null; grep -l '^\[workspace\]' Cargo.toml 2>/dev/null
```

| Evidence (first match wins) | Value | Source | Confidence |
|---|---|---|---|
| Any workspace marker found (`go.work`, `pnpm-workspace.yaml`, `[workspace]` in root `Cargo.toml`) | `true` | detected | high |
| Step 1 found markers for ≥ 2 distinct languages in ≥ 2 distinct directories at depth ≤ 2 (reuse Step 1's output — never rescan) | `true` | detected | medium |
| Neither | `false` | detected | medium *(negative evidence — see §4.4)* |

**Probe: `test_runner`** — each check is one `test -f` or one `grep -q` against the primary language's `{lang_root}`; for `scripts.test`, read `{lang_root}/package.json` and take the value of `.scripts.test`:

| Lang | Check (in table order) | Value | Confidence |
|---|---|---|---|
| go | `{lang_root}/Makefile` contains `go test` | `make test` | medium |
| go | `{lang_root}/go.mod` exists | `go test ./...` | high |
| node | `package.json` `.scripts.test` is non-empty | that script string | high |
| node | `{lang_root}/jest.config.js` exists | `jest` | medium |
| node | `{lang_root}/vitest.config.ts` exists | `vitest` | medium |
| python | `{lang_root}/pytest.ini` exists OR `pyproject.toml` contains `[tool.pytest` | `pytest` | high |
| python | `pytest` appears in `pyproject.toml` or `requirements.txt` | `pytest` | medium |
| ruby | `{lang_root}/Gemfile` contains `rspec` | `bundle exec rspec` | high |
| ruby | `{lang_root}/.rspec` exists | `bundle exec rspec` | medium |
| rust | `{lang_root}/Cargo.toml` exists | `cargo test` | high |

All table matches are `source=detected`.

**Probe: `naming_style`** — sample up to 8 source files (≤ 64 KB each) under the primary language's `{lang_root}`, depth ≤ 3, deterministic order:

```
fd -d 3 -t f -S -64k {extension flags from table} . {lang_root} | LC_ALL=C sort | head -8
```

(`find` fallback: `-maxdepth 3 -type f -size -65536c` with the same `-name` patterns, then the same `sort | head` pipeline.)

A file **conforms** when it matches the positive regex (`grep -qE`) and does not match the violation regex. Conforming ratio maps to confidence: ≥ 75% high, 50–74% medium, < 50% → `unknown`/inferred/low. At ≥ 50% the table value is emitted with `source=detected`.

| Lang | Extensions | Positive regex | Violation regex | Value when dominant |
|---|---|---|---|---|
| go | `.go` | `^(func\|type\|var\|const) [A-Z]` | `^(func\|type\|var\|const) [a-z]` | `snake_case filenames, PascalCase exported symbols` |
| node | `.ts` `.tsx` | `^export (function\|class\|const\|interface) [A-Z]` | `^export (function\|const) [a-z]` | `PascalCase exported symbols` |
| python | `.py` | `^(def [a-z_]\|class [A-Z])` | `^def [A-Z]` | `snake_case functions, PascalCase classes` |
| ruby | `.rb` | `^( *def [a-z_]\|class [A-Z]\|module [A-Z])` | `^ *def [A-Z]` | `snake_case methods, PascalCase classes` |
| rust | `.rs` | `^(pub )?(fn [a-z_]\|struct [A-Z]\|enum [A-Z])` | `^(pub )?fn [A-Z]` | `snake_case functions, PascalCase types` |

No source files sampled → `unknown`/inferred/low.

**Probe: `architectural_style`** — list top-level directories with one command (`fd -d 1 -t d . {dir}` or `find {dir} -maxdepth 1 -type d`), first at the repo root; if no row matches there and the primary `{lang_root}` differs from the root, evaluate the same table once more at `{lang_root}`:

| Layout evidence (first match wins) | Value | Source | Confidence |
|---|---|---|---|
| `cmd/` AND `internal/` present | `hexagonal` | detected | high |
| `src/` containing `controllers/`, `models/`, `views/` | `mvc` | detected | high |
| `src/` containing `features/` or `modules/` | `modular` | detected | medium |
| `src/` (no sub-pattern matched) | `layered` | detected | medium |
| `lib/` present (no `src/`) | `layered` | detected | medium |
| No match at root nor at `{lang_root}` | `unknown` | inferred | low |

**Output** — the sub-agent writes `{root}/knowledge/project-context.yaml` directly:

```yaml
schema_version: "1"
detected_at: {current UTC timestamp, ISO 8601}
is_monorepo: { value: "…", source: "…", confidence: "…" }
test_runner: { value: "…", source: "…", confidence: "…" }
naming_style: { value: "…", source: "…", confidence: "…" }
architectural_style: { value: "…", source: "…", confidence: "…" }
```

The sub-agent returns a `DetectionSummary` alongside the Step 3 output, listing each detected field, its value, source, and confidence.

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

**Negative-evidence rule**: a value concluded from the *absence* of evidence (e.g. `is_monorepo: "false"` because no workspace marker was found) caps at `confidence=medium`, never `high`. Absence proves the probe found nothing — not that nothing exists. `high` is reserved for direct positive file evidence.

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
