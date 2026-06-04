# Implement — Developer Specialist

## Purpose
Generate implementation code for each task, respecting existing conventions.

## Inputs
- `developer/dev-tasks`: ordered task list with files and dependencies
- `developer/dev-design`: technical approach, key constraints

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
