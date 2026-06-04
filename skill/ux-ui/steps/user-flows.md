# User Flows — UX/UI Specialist

## Purpose
Map the complete interaction sequences a user goes through to accomplish their goal.
Include the happy path and the most critical edge cases.

## Inputs
- `ux-ui/ia`: sections, navigation, data relationships

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
