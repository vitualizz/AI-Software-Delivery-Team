# User Flows — UX/UI Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Map the complete interaction sequences a user goes through to accomplish their goal.
Include the happy path and the most critical edge cases.

## Inputs
- `ux-ui/ia`: sections, navigation, data relationships

Retrieve via mem_search + mem_get_observation by topic_key.

Extract: navigation.entry_point, navigation.primary_actions, sections[].

## Context budget
ux-ui/ia navigation + sections summary: max 1,200 tokens.

## Processing
1. HAPPY PATH: map the sequence of steps from entry to success (numbered steps, actor perspective).
2. ERROR PATH: what happens when the primary action fails? (validation errors, network issues)
3. EDGE CASES: map 2-3 non-obvious flows (empty state, permission denied, partial data).
4. STATE TRANSITIONS: for each step, what changes in the UI?
5. DECISION POINTS: where does the user choose between paths?

Write flows in plain language: "User clicks X → System shows Y → User enters Z".
Do NOT describe visual layout. Describe sequence of events.

## Output
Produces: `ux-ui/flows`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  flows:
    - id: "UF-001"
      name: ""
      type: "happy-path|error|edge-case"
      actor: ""
      steps:
        - step: 1
          action: ""          # what the user does
          system_response: "" # what the system shows/does
          state_change: ""    # what changes in UI state
      decision_points: []
  open_items: []
```
