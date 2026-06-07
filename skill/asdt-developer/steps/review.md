# Review — Developer Specialist

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
Self-review the implementation and tests before producing the final artifact.
Apply the review shared skill, then consolidate into implementation-plan.

## Inputs
- `developer/dev-implementation`: implementation steps with code
- `developer/dev-tests`: test suites

Retrieve via mem_search + mem_get_observation by topic_key.

Extract from dev-implementation: `steps` list (title, files, rationale).
Extract from dev-tests: `test_suites` list (task coverage).

## Context budget
dev-implementation step titles + dev-tests coverage map: max 2,000 tokens.
Use context-extraction to summarize if artifacts are large.

## Processing
Apply the `review` shared skill:
1. Check each implementation step has at least one test case.
2. Check code snippets follow platform conventions.
3. Check all tasks from dev-tasks are represented.
4. Collect any gaps as `open_items`.

Then consolidate into the final `implementation-plan` artifact.

## Output
Produces: `implementation-plan` (final cross-specialist artifact)

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  complexity_estimate: "S|M|L|XL"
  open_items: []
  review_notes: []
  steps:
    - story_ref: ""      # task ID from dev-tasks
      title: ""
      files_to_create: []
      files_to_modify: []
      rationale: ""
      code_snippets:
        - file: ""
          language: ""
          content: ""
      test_snippets:
        - file: ""
          content: ""
```
