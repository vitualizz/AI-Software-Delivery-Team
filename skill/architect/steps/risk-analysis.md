# Risk Analysis — Architect Specialist

## Purpose
Identify the top risks introduced by this architectural decision and propose mitigations.

## Inputs
- `architect/system-design`: data model, API surface, service boundaries

Extract: service_boundaries.touched_modules, api_surface[].error_cases, data_model[].relationships.

## Context budget
architect/system-design (boundaries + API only): max 1,000 tokens.

## Processing
Identify risks across these categories:
1. PERFORMANCE: N+1 queries, missing indexes, synchronous chains that should be async.
2. SECURITY: exposed endpoints, unvalidated inputs, over-privileged access patterns.
3. RELIABILITY: single points of failure, missing error handling, cascade failure paths.
4. COUPLING: tight dependencies that will be hard to change, circular dependencies.
5. MIGRATION: data migration complexity, backward compat with existing clients.

For each risk:
- Likelihood: Low/Medium/High (how probable is this occurring?)
- Impact: Low/Medium/High (how bad if it occurs?)
- Mitigation: one concrete action that reduces the risk (not "monitor it" — a design change)

Focus on the top 3-5 risks. Do not list every possible risk — prioritize by likelihood × impact.

## Output
Produces: `architect/risks`

Schema:
```yaml
payload:
  risks:
    - id: "R-001"
      category: "performance|security|reliability|coupling|migration"
      title: ""
      description: ""
      likelihood: "Low|Medium|High"
      impact: "Low|Medium|High"
      mitigation: ""
  top_risk: ""    # ID of the highest priority risk
  open_items: []
```
