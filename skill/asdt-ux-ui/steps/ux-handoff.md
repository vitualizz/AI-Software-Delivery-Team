# UX Handoff — UX/UI Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Your inputs are INJECTED in this
> prompt by the orchestrator — do NOT fetch them. See
> `../asdt-shared/skills/parallel-retrieval.md` for the injected-input contract;
> if an input is marked UNRESOLVED, record it in `open_items` and proceed.
> Persist your output(s) via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`
> (this step produces TWO final artifacts — see the `## Output` note below for the
> second persistence key), then return a structured summary envelope (status,
> summary, output topic_key(s), open_items).

## Purpose
Consolidate all UX work into two final artifacts: ux-brief (for Developer context) and
component-spec (for implementation). Apply the report shared skill.

## Inputs
- `ux-ui/feature-brief`: actor, problem, success criteria
- `ux-ui/ia`: sections, navigation
- `ux-ui/flows`: interaction sequences
- `ux-ui/components`: component inventory
- `ux-ui/responsive`: breakpoint behavior

Apply context-extraction to each: keep only fields relevant to implementation handoff.

## Context budget
All inputs context-extracted to max 200 tokens each = max 1,000 tokens total.

## Processing
Apply the `report` shared skill:
1. From feature-brief: extract actor + success_criteria.
2. From ia: extract navigation.entry_point + primary_actions.
3. From flows: extract happy-path steps only (not edge cases — those go in QA).
4. From components: full component inventory.
5. From responsive: component_behavior table.
6. Consolidate open_items from ALL inputs into a deduplicated list.

## Output
Produces: `ux-brief` (final) and `component-spec` (final)

This step produces TWO final artifacts — a genuine dual-artifact shape (two fully
separate schema blocks below), structurally identical to architect's
`technical-handoff` (`architectural-decision` + `system-design`) and security's
`hardening-checklist` (`security-findings` + `hardening-checklist`); NOT a
single-cohesive-payload shape like qa's `quality-report`. Confirmed by reading
this section directly, not inferred from the compound-looking step/artifact names
(per the explicit caution forwarded from PR3/PR4: similarly-shaped compound names
have landed on opposite answers — qa's was single-artifact, security's was dual).

Persist `ux-brief` via `mem_save` under this step's `output_topic_key` in
`workflow.yaml` (`{project}/{change}/ux-ui/ux-brief`); persist the second final
artifact `component-spec` under its own distinct per-type topic_key
`{project}/{change}/ux-ui/component-spec` (see the inline YAML comment on this
step's `workflow.yaml` entry — no suffix needed, this name collides with neither
the primary key nor any intermediate artifact produced earlier in this
specialist's chain: `feature-brief`, `ia`, `flows`, `components`, `responsive`).
Return an envelope covering both persisted keys.

ux-brief schema:
```yaml
payload:
  feature_summary: ""
  primary_actor: ""
  success_criteria: []
  user_flows:
    - id: ""
      name: ""
      steps: []
  information_architecture:
    sections: []
    navigation_path: ""
  open_items: []
```

component-spec schema:
```yaml
payload:
  reused_components: []
  extended_components: []
  new_components:
    - name: ""
      purpose: ""
      props: []
      responsive_behavior: ""
  open_items: []
```
