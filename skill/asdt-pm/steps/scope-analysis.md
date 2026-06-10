# Scope Analysis — PM Specialist

## Purpose
Define explicit boundaries: what IS and IS NOT in scope for this feature.
A backlog entry without explicit out-of-scope items is incomplete — scope ambiguity
is the root cause of most scope creep.

## Inputs
- `pm/user-stories`

## Context budget
Extract: user_stories list (ids, actions, depends_on), stakeholders from prior artifact.
Max 300 tokens.

## Processing
1. List what IS being built, mapped to user story IDs or capabilities.
2. List what IS NOT being built now — make adjacent capabilities explicit. When in doubt, call it out of scope.
3. Identify integration points: other systems, services, or modules this feature touches, reads from, writes to, or depends on.
4. Flag scope risk items: anything that could cause unplanned expansion (e.g., "if we build X, we will also need Y").

## Output
Produces: `pm/scope-analysis`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  in_scope:
    - ""                   # capabilities or story IDs being built
  out_of_scope:
    - ""                   # explicit list of what is NOT being built in this iteration
  integration_points:
    - system: ""
      nature: ""           # reads-from | writes-to | triggers | depends-on
  risk_flags:
    - ""                   # things that could cause scope expansion if not managed
  open_items: []
```
