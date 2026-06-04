# UX Handoff — UX/UI Specialist

## Purpose
Consolidate all UX work into two final artifacts: ux-brief (for Developer context) and
component-spec (for implementation). Apply the report shared skill.

## Inputs
- `ux-ui/feature-brief`: actor, problem, success criteria
- `ux-ui/ia`: sections, navigation
- `ux-ui/flows`: interaction sequences
- `ux-ui/components`: component inventory
- `ux-ui/responsive`: breakpoint behavior

Apply context-extraction to each: keep only fields relevant to implementation handoff.

## Context budget
All inputs context-extracted to max 200 tokens each = max 1,000 tokens total.

## Processing
Apply the `report` shared skill:
1. From feature-brief: extract actor + success_criteria.
2. From ia: extract navigation.entry_point + primary_actions.
3. From flows: extract happy-path steps only (not edge cases — those go in QA).
4. From components: full component inventory.
5. From responsive: component_behavior table.
6. Consolidate open_items from ALL inputs into a deduplicated list.

## Output
Produces: `ux-brief` (final) and `component-spec` (final)

ux-brief schema:
```yaml
payload:
  feature_summary: ""
  primary_actor: ""
  success_criteria: []
  user_flows:
    - id: ""
      name: ""
      steps: []
  information_architecture:
    sections: []
    navigation_path: ""
  open_items: []
```

component-spec schema:
```yaml
payload:
  reused_components: []
  extended_components: []
  new_components:
    - name: ""
      purpose: ""
      props: []
      responsive_behavior: ""
  open_items: []
```
