# Test — Developer Specialist

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
Generate tests for the implementation, covering happy paths and key edge cases.

## Inputs
- `developer/dev-tasks`: task list with acceptance criteria references
- `developer/dev-implementation`: code snippets per task

Retrieve via mem_search + mem_get_observation by topic_key.

Extract from dev-tasks: `tasks[].ac_ref`, `tasks[].id`.
Extract from dev-implementation: `steps[].code_snippets[].content` (signatures only, not full bodies).

## Context budget
dev-tasks AC list + dev-implementation function signatures: max 2,500 tokens.

## Processing
For each task in dev-tasks:
1. Write one happy-path test covering the acceptance criterion (ac_ref).
2. Write one edge-case test for the most likely failure mode.
3. Follow the existing test framework convention from platform-summary.
4. Use table-driven tests where appropriate.

Do NOT test implementation internals — test behavior.

## Output
Produces: `developer/dev-tests`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  test_suites:
    - task_id: "T-001"
      test_cases:
        - id: "TC-001"
          title: ""
          type: "unit|integration"
          given: ""
          when: ""
          then: ""
          code_snippet:
            file: ""
            language: ""
            content: ""
  open_items: []
```
