# Tasks — Developer Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Break the implementation into an ordered checklist of atomic tasks.

## Inputs
- `developer/dev-spec`: scope, acceptance criteria
- `developer/dev-design`: technical approach, data model, API surface

Retrieve via mem_search + mem_get_observation by topic_key.

Extract from dev-spec: `acceptance_criteria`, `scope.in`.
Extract from dev-design: `data_model`, `api_surface`, `key_constraints`.

## Context budget
dev-spec summary + dev-design summary: max 2,500 tokens combined.

## Processing
1. Break the implementation into atomic tasks (each completable in < 2h).
2. Order tasks by dependency (models before controllers, etc.).
3. Assign each task a unique ID: T-001, T-002, etc.
4. Estimate each task: S (< 30m), M (30m-2h), L (2h-4h).
5. Map each task to an acceptance criterion it satisfies.

## Output
Produces: `developer/dev-tasks`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  tasks:
    - id: "T-001"
      title: ""
      files_to_create: []
      files_to_modify: []
      ac_ref: ""        # which acceptance criterion this satisfies
      estimate: "S|M|L"
      depends_on: []    # other task IDs
  total_estimate: ""
  open_items: []
```
