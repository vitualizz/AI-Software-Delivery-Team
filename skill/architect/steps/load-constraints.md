# Load Constraints — Architect Specialist

## Purpose
Gather all architectural constraints that MUST be respected before evaluating approaches.
Constraints come from: the existing platform, upstream specialist artifacts, and non-negotiable requirements.

## Inputs
- `architect/constraints`: platform analysis output (tech stack, existing patterns, service boundaries)
- Any upstream artifacts (ux-brief, requirements-spec) if present — use artifact-loading to check

Extract from architect/constraints: stack, key_patterns, naming_conventions.
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
