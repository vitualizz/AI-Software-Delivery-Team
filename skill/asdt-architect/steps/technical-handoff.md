# Technical Handoff — Architect Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` per the parallel-retrieval mandate at
> `../asdt-shared/skills/parallel-retrieval.md`. If any input fails to resolve,
> note it in `open_items` and proceed with available context. Persist your one
> output via `mem_save` under the `output_topic_key` declared for this step in
> `workflow.yaml`, then return a structured summary envelope (status, summary,
> output topic_key, open_items).

## Purpose
Consolidate all architectural work into final artifacts for Developer and QA specialists.
Apply the report shared skill. Surface key constraints the Developer MUST respect.

## Inputs
- `architect/adr`: the decision and its consequences
- `architect/system-design`: data model, API surface
- `architect/risks`: top risks and mitigations

Retrieve via mem_search + mem_get_observation by topic_key.

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

Persist `architectural-decision` via mem_save under this step's `output_topic_key` in workflow.yaml; persist the second final artifact `system-design` under its own distinct per-type topic_key (see the NOTE on this step's workflow.yaml entry — do not collide with the intermediate `architect/system-design` produced earlier by the system-design step); return envelope covering both persisted keys.

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
