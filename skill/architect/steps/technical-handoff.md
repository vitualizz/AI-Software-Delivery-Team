# Technical Handoff — Architect Specialist

## Purpose
Consolidate all architectural work into final artifacts for Developer and QA specialists.
Apply the report shared skill. Surface key constraints the Developer MUST respect.

## Inputs
- `architect/adr`: the decision and its consequences
- `architect/system-design`: data model, API surface
- `architect/risks`: top risks and mitigations

Apply context-extraction: keep what Developer needs, discard architect-internal reasoning.

## Context budget
All inputs context-extracted to max 300 tokens each = max 900 tokens total.

## Processing
Apply the `report` shared skill:
1. From adr: extract decision + consequences.negative (Developer must know these).
2. From system-design: extract full data_model + api_surface + key_sequence.
3. From risks: extract top 3 risks with their mitigations.
4. MANDATORY: write a "Key Constraints" section — explicit rules Developer must not violate.
5. Consolidate open_items from all inputs.

## Output
Produces: `architectural-decision` (final) and `system-design` (final)

architectural-decision schema:
```yaml
payload:
  decision_title: ""
  status: "accepted"
  context: ""
  decision: ""
  alternatives_considered: []
  consequences:
    positive: []
    negative: []
  key_constraints_for_developer: []  # MUST respect these
  open_items: []
```

system-design schema:
```yaml
payload:
  data_model: []
  api_surface: []
  service_boundaries: {}
  key_sequence: []
  top_risks: []
  open_items: []
```
