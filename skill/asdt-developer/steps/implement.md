# Implement — Developer Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` per the parallel-retrieval mandate at
> `../asdt-shared/skills/parallel-retrieval.md`. If any input fails to resolve,
> note it in `open_items` and proceed with available context. Persist your one
> output via `mem_save` under the `output_topic_key` declared for this step in
> `workflow.yaml`, then return a structured summary envelope (status, summary,
> output topic_key, open_items).

## Purpose
Generate implementation code for each task, respecting existing conventions.

## Inputs
- `developer/dev-tasks`: ordered task list with files and dependencies
- `developer/dev-design`: technical approach, key constraints

Retrieve via mem_search + mem_get_observation by topic_key.

Extract from dev-tasks: `tasks` list.
Extract from dev-design: `key_constraints`, `data_model` field shapes.

## Context budget
dev-tasks + dev-design summary: max 3,000 tokens. Generate code for tasks in batches
if the task list is large.

## Mode resolution (do this FIRST)
1. Resolve `allowedEditRoots` = union of `files_to_create` + `files_to_modify` across all tasks
   in `dev-tasks` (cross-check `dev-design` for additional declared targets).
2. If `allowedEditRoots` is EMPTY → PLAN-ONLY MODE: emit the snippet-based artifact (unchanged
   schema below, `code_snippets[]`). Write NO host files.
3. If `allowedEditRoots` is NON-EMPTY → WRITING MODE: for each task, write the real file(s) to
   disk, but ONLY to paths within `allowedEditRoots`. Before each write, confirm the target path
   is under a declared root; if not, STOP and report the unsafe path in `open_items` — do not write.
4. Match existing conventions: read the current content of any `files_to_modify` target first
   (per `skills/code-generation.md`) before editing it.

## Processing

### Plan-only mode
For each task in dev-tasks:
1. Generate the implementation code respecting `key_constraints` from dev-design.
2. Follow naming conventions from the platform-summary (loaded by platform-context shared skill).
3. Apply early return pattern, no global state, small functions.
4. Emit the code as inline `code_snippets[]` — write nothing to the host filesystem.

### Writing mode
For each task in dev-tasks:
1. Confirm the task's declared `files_to_create`/`files_to_modify` paths are within `allowedEditRoots`;
   if any path falls outside, STOP before writing it, do not expand scope, and record the unsafe
   path plus the triggering task in `open_items`.
2. Read existing file content for any `files_to_modify` target before editing (match conventions,
   avoid clobbering unrelated code).
3. Generate the implementation respecting `key_constraints` from dev-design, naming conventions
   from platform-summary, early-return pattern, no global state, small functions.
4. Write the real file to disk via the filesystem write tool, within the validated root.
5. Record the written path, action (`created`|`modified`), and rationale in the `files_changed[]`
   manifest entry for that task.

## Output
Produces: `developer/dev-implementation`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

The output schema is mode-dependent — set `mode` to the resolved value and emit the matching shape.

### Plan-only mode schema
```yaml
payload:
  mode: "plan-only"
  steps:
    - task_id: "T-001"
      title: ""
      files_to_create: []
      files_to_modify: []
      rationale: ""
      code_snippets:
        - file: ""
          language: ""
          content: ""
  open_items: []
```

### Writing mode schema
```yaml
payload:
  mode: "writing"
  allowedEditRoots: []        # resolved list, recorded verbatim for traceability
  files_changed:
    - path: ""
      action: "created|modified"
      task_id: "T-001"
      rationale: ""
  unsafe_skipped: []          # paths STOPPED on, with the triggering task_id
  open_items: []
```
