# Report — Shared Skill

## Purpose
Generate a structured handoff document from multiple intermediate artifacts.
Used as the consolidation (last) step of UX/UI, Architect, QA, Security, and PM specialists.

## Protocol
1. Consume the intermediate artifacts for this change directly from your prompt: your declared inputs arrive ALREADY INJECTED as `### INPUT {topic_key}` blocks (see `parallel-retrieval.md`). Do NOT call `mem_search`/`mem_get_observation` to (re)retrieve your own declared inputs — that work has already been done for you. Self-fetch (`mem_search(query: "{project}/{change}/{specialist}", project: "{project}")`, then `mem_get_observation(id: {id})` for each result) is a fallback ONLY when this step is declared with `inputs: []` in `workflow.yaml`, or when an input arrived as an `### INPUT {topic_key}: UNRESOLVED` block — in the UNRESOLVED case, record the gap in `open_items` and proceed.
2. For each intermediate: apply the Extraction Rules below (200 tokens max per artifact).
3. Identify the key decisions, outputs, and open items across all intermediates.
4. Merge into a coherent handoff document structured for the consuming specialist.
5. Collect all `open_items` from all intermediates into a deduplicated list.

## Extraction Rules
Extract only the fields relevant to the handoff from each large artifact.
Prevents full YAML dumps from bloating the consolidation step's context window.

**Scope guard**: apply these rules ONLY in handoff steps (the final consolidation
step of each specialist) that need to summarize multiple intermediate artifacts
into a coherent final output. Do NOT use them in intermediate steps — those
should declare precise InputRefs instead.

Given: an artifact + a specific question or purpose

1. State the question/purpose in one sentence.
2. Identify the top-level keys of the artifact most relevant to that question.
3. For each relevant key: extract the value and summarize it to at most 200 tokens.
4. Discard all other keys entirely.
5. Present as a compact summary block.

**Context budget**: each extracted artifact summary MUST fit within 200 tokens.
If a list has more than 5 items, keep the 5 most important and note "… N more".

Example:
- Question: "What are the implementation steps I need to test?"
- Artifact: dev-tasks.yaml (contains 20+ fields)
- Extracted: tasks[].id, tasks[].title only — discard estimates, rationale, files.

## Output
Produces the specialist's final cross-specialist artifact(s) — e.g.:
- UX/UI → `ux-brief` + `component-spec`
- Architect → `architectural-decision` + `system-design`
- QA → `test-plan`
- Security → `security-findings` + `hardening-checklist`
- PM → `backlog-entry`

## Quality gate
Before writing the final artifact:
- Every section of the output MUST reference at least one intermediate artifact.
- `open_items` from all intermediates MUST be consolidated — do not silently drop them.
- The output MUST be useful to a specialist who has NOT seen the intermediate artifacts.
