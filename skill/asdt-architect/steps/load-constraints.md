# Load Constraints — Architect Specialist

## Purpose
Gather all architectural constraints that MUST be respected before evaluating approaches.
Constraints come from: the existing platform, upstream specialist artifacts, and non-negotiable requirements.

## Inputs
- Platform context (tech stack, existing patterns, service boundaries) — injected inline by the `platform-analysis` step that runs immediately before this one in the orchestrator's context
- Any upstream artifacts (ux-brief, requirements-spec) if present — use artifact-loading to check

Note: this step's `inputs:` list in `workflow.yaml` is empty by design — it has no prior `subagent`-produced artifact to retrieve; it consumes the inline-injected platform context plus any upstream artifacts. Retrieve any upstream artifacts via mem_search + mem_get_observation by topic_key.

Extract from the injected platform context: stack, key_patterns, naming_conventions.
Extract from upstream artifacts (if present): scope.in, scope.out, key technical requirements.

## Context budget
architect/constraints + upstream artifact summaries: max 1,500 tokens.

## Processing
Classify constraints into three buckets:
1. HARD CONSTRAINTS: things that cannot change (existing DB schema being used by other services,
   framework version lock, API contracts already in production).
2. SOFT CONSTRAINTS: strong preferences that need a compelling reason to override (naming
   conventions, existing patterns, preferred libraries).
3. OPPORTUNITIES: architectural improvements this change could enable (refactoring a coupling,
   extracting a reusable module).

For each hard constraint, note: what it is, why it cannot change, and how it limits the solution space.

## Output
Produces: `architect/constraints-analysis` (constraint analysis — feeds into evaluate-approaches)

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  hard_constraints:
    - what: ""
      why_immutable: ""
      solution_impact: ""
  soft_constraints:
    - what: ""
      override_cost: ""
  opportunities: []
  upstream_requirements: []   # key points from upstream artifacts
  open_items: []
```
