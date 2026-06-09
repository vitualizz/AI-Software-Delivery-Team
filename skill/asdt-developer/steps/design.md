# Design — Developer Specialist

## Purpose
Choose the technical approach and define the data model and API shape.

## Inputs
- `developer/dev-spec`: scope, acceptance criteria, technical requirements

ONLY READ dev-spec. Do NOT load exploration or any other artifact.

## Context budget
dev-spec: max 2,000 tokens.

## Processing
1. Propose the technical approach (1-2 paragraphs, compare 2 options if non-obvious).
2. Define the data model: entities, fields, relationships.
3. Define the API surface: endpoints/functions/interfaces with signatures.
4. Identify any migration or backward-compat concerns.
5. List key technical constraints for implementation (e.g. "use existing AuthMiddleware").

Do NOT write implementation code. Only define the technical structure.

## Output
Produces: `developer/dev-design`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  approach: ""
  data_model:
    - name: ""
      fields: []
  api_surface:
    - name: ""
      signature: ""
      purpose: ""
  migration_notes: []
  key_constraints: []
  open_items: []
```
