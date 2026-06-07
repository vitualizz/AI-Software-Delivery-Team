# Implement — Developer Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

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

## Processing
For each task in dev-tasks:
1. Generate the implementation code respecting `key_constraints` from dev-design.
2. Follow naming conventions from the platform-summary (loaded by platform-context shared skill).
3. Apply early return pattern, no global state, small functions.
4. Include only inline code snippets — do NOT write files directly.

## Output
Produces: `developer/dev-implementation`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
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
