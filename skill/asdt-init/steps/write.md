# Write — Init Specialist

## Purpose
Write the four `.asdt` files from the detected stack plus the human's clarify
answers. This is init's only filesystem-writing step. It runs as a `builder`
sub-agent: it cannot pause to ask questions — every question was already asked by
the inline `clarify` step and the answers were injected into this prompt.

## Inputs
- `init/stack-detection` (injected) — the explore step's `detected_stack`,
  `lang_roots`, `fields`, and `ambiguities[]`. Consume the injected block
  directly; do NOT re-fetch it.
- **`### CLARIFY ANSWERS` block (injected)** — the orchestrator composes this
  from the inline clarify step and prepends it to your prompt. Consume it
  directly, never re-fetch it. Its shape:

  ```
  ### CLARIFY ANSWERS
  answers: { field: value, ... }     # may be {} when there was nothing to ask
  skipped: true|false                # true → non-interactive harness, defaults applied
  blocking_open_items: []            # non-empty → HALT (see Halt contract)
  ```

## Context budget
`stack-detection`: max 2,000 tokens. The CLARIFY ANSWERS block is small by
construction. Do not pull anything else into context.

## Recalibration contract
The Engram gate already passed PRE-EXPLORE (the orchestrator checked its own tool
list before launching explore). You do not re-run it.

**Preserve `source: manual` fields. NEVER silently overwrite them.** When
`.asdt/knowledge/project-context.yaml` already exists, any field whose existing
`source` is `manual` was set by a human in a prior recalibration review. Carry it
forward unchanged unless a clarify answer for that exact field explicitly
overrides it. Surface every preserved field in `settings_preserved[]` so the
outcome is auditable.

## Processing

`.asdt/` holds static reference data — bootstrapped once and refreshed only on a
deliberate recalibration, never per-change.

1. **`.asdt/config.yaml`** — write `memory.provider: engram` plus any preserved
   settings carried forward from an existing config:

   ```yaml
   memory:
     provider: engram
   ```

2. **`.asdt/knowledge/platform.yaml`** — populate only what the bounded scan
   determined deterministically. `conventions.file_structure` is the one-line
   sentence derived from top-level directory matches; leave
   `design_fingerprint: {}` (identifying architectural patterns means sampling
   file contents — out of scope for init):

   ```yaml
   schema_version: "1"
   scanned_at: {current UTC timestamp, ISO 8601}
   detected_stack: {stack-detection.detected_stack}
   conventions:
     file_structure: {one-line description}
   design_fingerprint: {}
   ```

3. **`.asdt/knowledge/platform-summary.yaml`** — derived FROM `platform.yaml`,
   never re-analyzed from scratch:

   ```yaml
   schema_version: "1"
   stack: {platform.yaml detected_stack}
   file_structure: {platform.yaml conventions.file_structure}
   ```

4. **`.asdt/knowledge/project-context.yaml`** — built from
   `stack-detection.fields` with each applied clarify answer overlaid. A field
   answered by the human becomes `source: manual`. Record every applied
   `Ambiguity` answer with its `origin` (`user` | `default`):

   ```yaml
   schema_version: "1"
   detected_at: {current UTC timestamp, ISO 8601}
   is_monorepo: { value: "…", source: "…", confidence: "…" }
   test_runner: { value: "…", source: "…", confidence: "…" }
   naming_style: { value: "…", source: "…", confidence: "…" }
   architectural_style: { value: "…", source: "…", confidence: "…" }
   ```

All four files stay bounded — their size grows with the number of detected
stacks, never with repo size.

## Halt contract

HALT with an error and ZERO writes if either condition holds:

- **(a) The `### CLARIFY ANSWERS` block is absent from the prompt.** Absence
  means the orchestrator failed to pass it — NOT that there were no answers. The
  block is REQUIRED even when empty (`answers: {}`, `skipped: true|false`,
  `blocking_open_items: []`). No block → halt; do not write partial config.
- **(b) `blocking_open_items[]` is non-empty.** A blocking open item means a
  non-skippable ambiguity went unresolved; writing config on top of it would
  bake in a wrong default. Halt and surface the items.

On halt, write nothing and return the error in the envelope.

## Output
Produces: `init/write-summary`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  files_written: []           # absolute or repo-relative paths of the .asdt files written
  settings_preserved: []      # source:manual fields carried forward unchanged
  applied_answers:            # every Ambiguity answer applied, with its origin
    - field: ""
      value: ""
      origin: "user | default"
  open_items: []
```

If the `mem_save` call fails, record the failure in `open_items` — the files on
disk are authoritative and the writes already succeeded. Do NOT halt or roll back
on a persistence failure; the config is the durable outcome, the summary is the
audit trail.
