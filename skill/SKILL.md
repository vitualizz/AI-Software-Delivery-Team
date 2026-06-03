---
name: asdt
description: AI Software Delivery Team — artifact-first delivery pipeline
user-invocable: true
---

# ASDT — AI Software Delivery Team

## 1. Invariants

These rules are non-negotiable and are enforced before any subcommand executes:

- **Write isolation**: Never write any file outside the resolved `.asdt/` root. This is an absolute prohibition. Any path traversal that would escape `.asdt/` is treated as a critical defect.
- **Runtime agnosticism**: Never call runtime-specific APIs (no Claude Code tool calls, no OpenCode hooks). All behavior is expressible as prompt execution + file read/write only.
- **State boundary**: All state lives in `.asdt/artifacts/{change}/` (per-change artifacts) and `.asdt/knowledge/` (project-wide knowledge). No state is written anywhere else.
- **Envelope completeness**: Every artifact produced must have a complete envelope at the root level with these fields present and non-empty: `schema_version`, `agent`, `change_id`, `created_at`, `prompt_version`, `input_refs`, `payload`. Missing any field is a validation failure — do not write the artifact.

---

## 2. Boundary Resolution

Before executing any subcommand, resolve the `.asdt/` root:

1. Start from the current working directory (CWD).
2. Walk up the directory tree, checking each ancestor for the presence of a `.asdt/` directory.
3. Stop at the first ancestor that contains `.asdt/`. Use that ancestor as the **ASDT root**. This is the nearest-ancestor rule (monorepo-safe).
4. If no `.asdt/` is found anywhere in the ancestor chain:
   - Detect the project root by looking for `.git`, `go.mod`, `package.json`, `Cargo.toml`, or `pyproject.toml` in the ancestor chain (first match wins).
   - Offer to create `.asdt/` at that detected project root. If no project marker is found, offer to create `.asdt/` at CWD.
   - Do not create `.asdt/` without explicit user acknowledgment.
   - Do not create any files until `.asdt/` is confirmed.
5. Never cross filesystem mount boundaries during the walk.

All paths in the dispatch table (`knowledge/platform.yaml`, `artifacts/{change}/...`) are relative to the resolved ASDT root.

---

## 3. Argument Contract

Full syntax:

```
/asdt <subcommand> [payload] [--change <name>]
```

- `<subcommand>` (required): The first positional argument. If missing or empty, list all valid subcommands and stop. Do not execute any role workflow.
- `[payload]`: Free-text argument whose meaning depends on the subcommand (see Dispatch Table). Some subcommands require it; others ignore it.
- `[--change <name>]` (optional): Identifies the active change. Lookup order:
  1. `--change <name>` flag if present in the invocation.
  2. `active_change` field in `.asdt/config.yaml` if it exists.
  3. Infer from the current directory name as a last resort.
  4. If none of the above resolves, ask the user for a change name before proceeding.

---

## 4. Dispatch Table

| subcommand     | loads                                              | reads                                                                 | writes                                                                            | payload meaning                        |
|----------------|----------------------------------------------------|-----------------------------------------------------------------------|-----------------------------------------------------------------------------------|----------------------------------------|
| `knowledge`    | knowledge role + applicable skill fragments        | project source files (scan from ASDT root's parent)                   | `.asdt/knowledge/platform.yaml`                                                   | optional: `refresh` flag to force re-scan |
| `requirements` | requirements role + user-story-writing, scope-definition skills | `.asdt/knowledge/platform.yaml` (if exists; warn if absent) | `.asdt/artifacts/{change}/requirements-spec.yaml`, `.asdt/artifacts/{change}/pipeline-state.yaml` | free-text feature idea (required) |
| `develop`      | developer role + code-generation, test-writing skills | `.asdt/artifacts/{change}/requirements-spec.yaml` (required; fail if absent), `.asdt/knowledge/platform.yaml` (warn if absent) | `.asdt/artifacts/{change}/implementation-plan.yaml`, `.asdt/artifacts/{change}/pipeline-state.yaml` | none (reads from artifacts automatically) |
| `status`       | none (inline rendering)                            | `.asdt/artifacts/{change}/pipeline-state.yaml`, all YAML files in `.asdt/artifacts/{change}/` | none                                               | optional: `--change <name>` to show a specific change |
| `help`         | none                                               | none                                                                  | none                                                                              | optional: subcommand name for targeted help |

---

## 5. Routing Instruction

When invoked, execute these steps in order:

**Step 1 — Parse subcommand.**
Extract the first positional argument. If it is missing, output the help message listing all valid subcommands and stop.

**Step 2 — Validate subcommand.**
Check against the dispatch table: `knowledge`, `requirements`, `develop`, `status`, `help`.
If the subcommand is not in this list, go to Section 6 (Unknown Subcommand Behavior) and stop.

**Step 3 — Resolve `.asdt/` root.**
Follow the Boundary Resolution procedure in Section 2. If resolution fails (no `.asdt/` and user declines creation), stop with a clear message.

**Step 4 — Load the role prompt.**
Read `skill/prompts/roles/{subcommand}/role.md` from the skill package.
Override precedence (first match wins):
1. `.asdt/prompts/{subcommand}/role.md` (project-local override)
2. `~/.config/asdt/prompts/{subcommand}/role.md` (user-global override)
3. `skill/prompts/roles/{subcommand}/role.md` (packaged default)

**Step 5 — Load applicable skill fragments.**
For the matched subcommand, load the skill fragments listed in the dispatch table from `skill/prompts/skills/{fragment-name}.md`, applying the same override precedence.

**Step 6 — Compose the effective prompt.**
Assemble layers in this exact order:
1. Role prompt (persona + workflow instructions)
2. Skill fragments (capability guidelines), concatenated in dispatch-table order
3. Artifact context (serialized content of required input artifacts from the dispatch table)
4. Platform context (contents of `.asdt/knowledge/platform.yaml` if it exists; omit silently if absent and not required)

Separate each layer with `---`.

**Step 7 — Execute the role workflow.**
Follow the instructions in the composed role prompt to perform the subcommand's work.

**Step 8 — Write the output artifact.**
Produce the output artifact(s) listed in the dispatch table's `writes` column. Before writing, validate that the envelope has all required fields (Section 1, Invariant 4). Write only inside the resolved `.asdt/` root.

**Step 9 — Update `pipeline-state.yaml`.**
After writing the primary artifact, update `.asdt/artifacts/{change}/pipeline-state.yaml` to reflect the new state. Append the transition to the `transitions[]` history. Do not overwrite the history.

---

## 6. Unknown Subcommand Behavior

If the subcommand is not recognized:

```
Unknown subcommand: {name}. Valid subcommands: knowledge, requirements, develop, status, help
```

Then STOP. Do not execute any role workflow, do not read any files, do not write any files.

---

## 7. Missing Payload Behavior

If a subcommand requires a payload (e.g. `requirements` requires a feature idea) and none is provided:

```
The '{subcommand}' subcommand requires a payload. {Describe what is needed}.
Example: /asdt {subcommand} "your input here"
```

Then STOP and wait for the user to re-invoke with a payload. Do not attempt to proceed with an empty or inferred payload.

---

## 8. Cross-Runtime Compatibility Note

This skill MUST work identically in Claude Code and OpenCode. The router uses only:

- File read operations (to load role prompts, skill fragments, and input artifacts)
- File write operations (to persist output artifacts)
- The composed prompt (executed by the runtime's LLM)
- Argument passthrough from the invocation

No runtime-specific tool calls, no session state, no provider-specific APIs. Any behavior that cannot be expressed as "read file + write file + LLM completion" is out of scope for this skill package.

---

## 9. Override Resolution Note

Project teams and individual users can customize any role prompt or skill fragment without forking this skill. Place override files at:

- **Project-local**: `.asdt/prompts/{subcommand}/role.md` or `.asdt/prompts/skills/{fragment-name}.md`
- **User-global**: `~/.config/asdt/prompts/{subcommand}/role.md` or `~/.config/asdt/prompts/skills/{fragment-name}.md`

Project-local overrides always win over user-global, which always win over the packaged defaults. The `prompt_version` field in the artifact envelope records which fragments were active, enabling drift detection.
