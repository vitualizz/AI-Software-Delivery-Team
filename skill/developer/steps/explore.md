# Explore — Developer Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Understand the area of the codebase that will change before writing a single line.

## Inputs
- Request: the feature/change description from the user
- `platform-summary`: stack, naming conventions, key patterns

Note: This step has no prior step artifacts. It operates purely on the request and platform context.
Retrieve via mem_search + mem_get_observation by topic_key (this step's `inputs:` list in `workflow.yaml` is empty — it reads only the raw request and platform-summary, not a prior artifact).

## Context budget
Request text + platform-summary: max 1,000 tokens combined.

## Processing
1. Identify which existing files/modules are likely affected by this change.
2. Identify naming patterns and conventions from platform-summary relevant to this change.
3. List known risks or constraints (e.g. "this touches the auth layer which has rate limiting").
4. List open questions that need answering before speccing (e.g. "does this need migrations?").

Do NOT design the solution. Do NOT write code. Only explore and understand.

## Output
Produces: `developer/dev-exploration`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  files_to_understand: []     # existing files/modules relevant to this change
  patterns_to_follow: []      # naming/structural conventions from platform-summary
  risks: []                   # known constraints or risks
  open_questions: []          # questions that will be answered in the spec step
  open_items: []
```
