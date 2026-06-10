# Backlog Entry — PM Specialist

## Purpose
Consolidate all PM artifacts into the final structured backlog entry.
This is the canonical requirements artifact consumed by Architect, Developer, and QA.

## Inputs
- `pm/user-stories`
- `pm/scope-analysis`
- `pm/prioritization`

## Context budget
Load all three input artifacts in full. Max 1,500 tokens total.

## Processing
1. Merge user stories with their position from the prioritization artifact.
2. Attach the full scope block (in/out of scope, integration points, risk flags).
3. Write an `executive_summary`: one paragraph — what this feature is, why it matters, and what it explicitly excludes. This is the human-readable entry point for any specialist picking up this artifact.
4. Carry forward any unresolved `open_items` from prior steps — downstream specialists should address these.
5. Write a `summary` field (≤ 150 tokens) for decision-preservation.

## Output
Produces: `pm/backlog-entry`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  feature_name: ""
  summary: ""              # ≤ 150 tokens — consumed by decision-preservation
  executive_summary: ""    # 1 paragraph: what, why, what it explicitly excludes
  user_stories:
    - id: ""
      role: ""
      action: ""
      benefit: ""
      priority: "must | should | could | wont"
      size: "small | medium | large"
      acceptance_criteria: []   # high-level plain English — QA formalizes these
      depends_on: []
  scope:
    in_scope: []
    out_of_scope: []
    integration_points:
      - system: ""
        nature: ""
    risk_flags: []
  priority_order: []       # ordered list of US IDs (from prioritization)
  deferred: []             # US IDs explicitly deferred with reasons
  open_items: []           # unresolved questions for downstream specialists
```

## Downstream consumption
- **Architect**: reads `executive_summary`, `scope` (integration_points, risk_flags)
- **Developer**: reads `user_stories`, `priority_order`, `acceptance_criteria`
- **QA**: reads `user_stories` + `acceptance_criteria` as the primary requirements source — this replaces the raw request fallback in `load-requirements`
