# Report — Shared Skill

## Purpose
Generate a structured handoff document from multiple intermediate artifacts.
Used as the last step of UX/UI, Architect, QA, and Security specialists.

## Protocol
1. Retrieve all intermediate artifacts for this change via Engram: call `mem_search(query: "{project}/{change}/{specialist}", project: "{project}")`. For each result returned, call `mem_get_observation(id: {id})` to get the full content before proceeding.
2. For each intermediate: apply context-extraction (200 token max per artifact).
3. Identify the key decisions, outputs, and open items across all intermediates.
4. Merge into a coherent handoff document structured for the consuming specialist.
5. Collect all `open_items` from all intermediates into a deduplicated list.

## Output
Produces the specialist's final cross-specialist artifact(s) — e.g.:
- UX/UI → `ux-brief` + `component-spec`
- Architect → `architectural-decision` + `system-design`
- QA → `test-plan`
- Security → `security-findings` + `hardening-checklist`

## Quality gate
Before writing the final artifact:
- Every section of the output MUST reference at least one intermediate artifact.
- `open_items` from all intermediates MUST be consolidated — do not silently drop them.
- The output MUST be useful to a specialist who has NOT seen the intermediate artifacts.
