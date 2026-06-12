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

## Orchestration Plan

**asdt-init is STANDALONE.** It is a user-invocable setup command, deliberately
NOT registered in `skill/SKILL.md` §9.2 routing — no feature request ever routes
to setup, so the meta-orchestrator has nothing to route here (per ADR-016). Do
not "fix" this omission. init has NO complexity tiers and a fixed 3-step flow;
there is no tier→step table.

| Step | File | Execution | Reads | Writes |
|------|------|-----------|-------|--------|
| knowledge-gate | *(inline — no step file)* | inline | *(orchestrator's own tool list)* | *(no artifact — Engram presence gate)* |
| explore | steps/explore.md | subagent | *(raw project tree — `inputs: []`)* | `init/stack-detection` |
| clarify | *(inline — no step file)* | inline | `init/stack-detection.ambiguities[]` | *(no artifact — injects `answers{}` into write's prompt)* |
| write | steps/write.md | subagent | `init/stack-detection` + `### CLARIFY ANSWERS` | `init/write-summary` |

Step names byte-match `workflow.yaml`; when they differ, `workflow.yaml` is
authoritative.

## Orchestration

This is a light flow — but it has one gate only the orchestrator can pass correctly, and everything downstream depends on it.

**Resolve Engram presence yourself, first, before delegating anything** (the `knowledge-gate` step / Step 2's detection). "Does THIS session have Engram's memory tools" is a question about the orchestrator's own tool list — the one the user is actually relying on for every other specialist. A sub-agent has its own tool list (narrower for specialized agent types, full for `general-purpose`); asking it risks a false "absent" when Engram is actually present in the session that matters. It costs nothing to check yourself — you're inspecting your own tools, not running commands or reading files. This gate is undelegatable; it stays inline with you.

- **Absent** → stop right here and tell the user, exactly as Step 2 describes. Nothing downstream can run without it — don't launch a sub-agent only to have it discover the same dead end.
- **Present** → run the 3-step flow:
  - **explore** (subagent, `agent: analyst`) — launch `steps/explore.md` via your native delegation primitive, passing "Engram confirmed present" into its prompt as an established fact. It detects the stack and context, flags `ambiguities[]`, and returns `init/stack-detection`. It writes NO files.
  - **clarify** (inline — your own context) — resolve `stack-detection.ambiguities[]` with the human, ONE question at a time, then compose the `### CLARIFY ANSWERS` block (see the clarify contract in the Workflow below).
  - **write** (subagent, `agent: builder`) — launch `steps/write.md`, injecting both `init/stack-detection` and the `### CLARIFY ANSWERS` block. It writes the four `.asdt` files and returns `init/write-summary`.

**PRE-EXPLORE recalibration gate.** Before launching explore, check for an
existing `.asdt/config.yaml`. If it exists, this project was already
initialized — resolve recalibrate-vs-leave WITH THE USER first. Fail fast: never
run detection only to discard it. If the user chooses "leave as-is", stop without
launching explore. Only proceed to explore once the user has chosen to
recalibrate (a fresh setup with no existing config proceeds directly).

Routing work through sub-agents keeps the bash output, file reads, and intermediate reasoning out of your main context — which is the whole point of the specialist model.

## Workflow

### Step 1 — Detect project stack *(in `explore`)*

The marker-scan mechanics — the bounded `fd`/`find` pipeline, its exclusions and
deterministic ordering, the result cap, the marker→language mapping table, and
the `detected_stack` / primary-language / `{lang_root}` derivation — live in
**`steps/explore.md`**. The explore sub-agent runs them and returns
`init/stack-detection`. Do not duplicate the pipeline here.

### Step 2 — Detect the memory provider

**Detect Engram yourself — do not ask the user to confirm something you can observe directly.** Check your own current tool list for Engram's memory tools (`mem_save`, `mem_search`, `mem_context`, etc. — Claude Code exposes them prefixed as `mcp__plugin_engram_engram__mem_*`; other host assistants may expose the same tools under a different prefix or none).

- If they're present → Engram is installed and reachable. Tell the user so and continue.
- If they're absent → tell the user Engram is required for ASDT's cross-session memory and is not reachable in this session, explain how to install/connect it, and STOP. Do not write `.asdt/config.yaml` with `provider: engram` when the provider isn't actually present — that would silently point every future specialist at a memory backend that doesn't exist.

### Step 2.5 — Clarify *(inline — your own context, between explore and write)*

explore returns `init/stack-detection` carrying `ambiguities[]` — one entry per
low/medium-confidence or genuinely ambiguous field. clarify is where you, the
orchestrator, resolve them WITH the human. This step runs inline because only an
inline step can pause for a question; the `write` sub-agent cannot.

The inline contract:

1. For each `Ambiguity` in `stack-detection.ambiguities[]`, ask the user ONE
   question at a time — the same one-question-per-field discipline §4.3 uses for
   recalibration. Offer `options` when present; otherwise take a free-form value.
2. Collect the answers into `answers{}` (field → value).
3. Compose the `### CLARIFY ANSWERS` block and inject it into write's prompt —
   it is REQUIRED even when there was nothing to ask:

   ```
   ### CLARIFY ANSWERS
   answers: { field: value, ... }     # {} when nothing was asked
   skipped: true|false
   blocking_open_items: []
   ```

4. **Non-interactive harness** (no way to ask the human): SKIP the questions. For
   each ambiguity, if `skippable: true` apply its `default` (recorded later as
   `origin: default`); if `skippable: false` add it to `blocking_open_items[]`.
   Set `skipped: true`.
5. **User abort**: if the user cancels, do NOT launch write — no files are
   written.

The `### CLARIFY ANSWERS` block is the only thing the inline clarify step
produces; it carries no artifact of its own.

### Step 3 — Write configuration files *(in `write`)*

The file-writing mechanics — the `.asdt/config.yaml` (`memory.provider: engram`)
write, the `platform.yaml` scan and `conventions.file_structure` derivation, the
`platform-summary.yaml` derived FROM `platform.yaml`, and the
`project-context.yaml` build from `stack-detection.fields` + applied answers —
live in **`steps/write.md`**, along with the idempotency check, the
`source: manual` preservation rule, and the halt contract. The write sub-agent
owns all four file writes. Do not duplicate the write mechanics here.

### Step 4 — Detect project context

Produce `.asdt/knowledge/project-context.yaml` — a machine-written file that records _how_ the project is structured and coded (monorepo shape, test runner, naming style, architectural pattern). This is separate from `platform.yaml`, which records _what_ is installed.

#### 4.1 Check for existing project-context.yaml

Look for `{root}/knowledge/project-context.yaml`:

- **Absent** → fresh detection path (§4.2).
- **Present** → recalibration path (§4.3).

#### 4.2 Fresh detection *(in `explore`)*

The four probes — `is_monorepo`, `test_runner`, `naming_style`,
`architectural_style` — each with its bounded command and exact mapping table,
plus the "one bounded command, first matching row wins, no model judgment" rule
and the per-field `FieldValue` shape, live in **`steps/explore.md`** (§4 probes).
explore detects every field, attaches `source`/`confidence`, and emits an
`Ambiguity` for any low/medium-confidence field. The `write` sub-agent applies
the clarify answers on top and writes `project-context.yaml`; explore itself
writes nothing. Per ADR-013, all probes run as bounded shell commands with no
dependency on this repo's Go code.

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
