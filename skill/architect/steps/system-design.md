# System Design — Architect Specialist

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
