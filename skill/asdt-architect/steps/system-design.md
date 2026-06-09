# System Design — Architect Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Your inputs are INJECTED in
> this prompt by the orchestrator — do NOT fetch them. See
> `../asdt-shared/skills/parallel-retrieval.md` for the injected-input
> contract; if an input is marked UNRESOLVED, record it in `open_items` and
> proceed. Persist your one
> output via `mem_save` under the `output_topic_key` declared for this step in
> `workflow.yaml`, then return a structured summary envelope (status, summary,
> output topic_key, open_items).

## Purpose
Define the concrete technical structure: data model, API surface, and service boundaries.

## Inputs
- `architect/adr`: chosen approach, key constraints, consequences

Extract: decision (the chosen approach), consequences.negative (constraints to design around).

## Context budget
architect/adr: max 1,200 tokens.

## Processing
1. DATA MODEL: define entities, fields, types, and relationships.
   - Use the naming conventions from platform-summary (loaded via platform-context).
   - Note which fields are indexed, which are nullable, which have constraints.
2. API SURFACE: for each operation, define:
   - Method/endpoint or function signature
   - Input parameters with types
   - Success response shape
   - Error cases (what HTTP codes or error types, and when)
3. SERVICE BOUNDARIES: which existing services/modules does this touch?
   - What new interfaces (if any) need to be defined?
   - What existing interfaces are being extended?
4. SEQUENCE: one key interaction sequence (the happy path) showing how components collaborate.

## Output
Produces: `architect/system-design`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  data_model:
    - entity: ""
      fields:
        - name: ""
          type: ""
          constraints: ""
      relationships: []
  api_surface:
    - operation: ""
      method: ""
      input: {}
      success_response: {}
      error_cases: []
  service_boundaries:
    touched_modules: []
    new_interfaces: []
    extended_interfaces: []
  key_sequence: []    # numbered steps of happy-path interaction
  open_items: []
```
