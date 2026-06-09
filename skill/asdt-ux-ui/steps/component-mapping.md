# Component Mapping — UX/UI Specialist

## Purpose
Identify which existing components to reuse, which to extend, and which must be created.
Maximize reuse. Justify every new component.

## Inputs
- `ux-ui/flows`: interaction sequences, state changes
- `platform-summary`: component_library, existing patterns

Retrieve `ux-ui/flows` via mem_search + mem_get_observation by topic_key
(`platform-summary` is provided directly by the orchestrator's inline
`platform-analysis` step).

Extract from flows: all unique UI states and interactions.
Extract from platform-summary: component_library name and approach.

## Context budget
flows summary (UI states only) + platform-summary: max 1,500 tokens.

## Processing
For each UI state identified in the flows:
1. CHECK if an existing component handles this state. If yes → REUSE.
2. If the existing component needs minor changes → EXTEND (document what changes).
3. Only if no existing component fits → CREATE NEW (document why existing ones don't work).

For each new component:
- Name it following the project's naming convention (from platform-summary).
- Define its props/inputs (data it needs).
- Define its events/outputs (what it emits).
- Note its responsive behavior (how it changes at breakpoints).

Quality gate: the ratio of reused to new components should be > 2:1 for features on
existing platforms. If you're creating more than 30% new components, revisit whether
existing ones can be extended instead.

## Output
Produces: `ux-ui/components`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  reused_components:
    - name: ""
      use_case: ""
  extended_components:
    - name: ""
      changes_needed: []
      use_case: ""
  new_components:
    - name: ""
      reason_existing_insufficient: ""
      props: []
      events: []
      responsive_behavior: ""
  reuse_ratio: ""   # e.g. "4:1 (4 reused, 1 new)"
  open_items: []
```
