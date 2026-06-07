# Information Architecture — UX/UI Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Define the content hierarchy, navigation structure, and data relationships for the feature.
Decide how information is organized before deciding how it looks.

## Inputs
- `ux-ui/feature-brief`: actor, problem, success criteria, constraints

Retrieve via mem_search + mem_get_observation by topic_key.

Extract: success_criteria (determines what content is needed), design_constraints.

## Context budget
ux-ui/feature-brief: max 1,000 tokens.

## Processing
1. LIST all content pieces the user needs to accomplish their goal.
2. GROUP related content into logical sections (use card-sorting thinking).
3. PRIORITIZE: what must be visible immediately vs. progressive disclosure?
4. DEFINE the navigation path: what is the entry point? What are the exit points?
5. IDENTIFY data relationships: what loads together vs. on demand?
6. APPLY progressive disclosure: what can be hidden initially to reduce cognitive load?

## Output
Produces: `ux-ui/ia`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  sections:
    - name: ""
      content_items: []
      priority: "primary|secondary|tertiary"
      disclosure: "immediate|progressive|on-demand"
  navigation:
    entry_point: ""
    primary_actions: []
    exit_points: []
  data_relationships:
    - entities: []
      load_strategy: "together|lazy|on-demand"
  open_items: []
```
