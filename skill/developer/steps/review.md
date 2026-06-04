# Review — Developer Specialist

## Purpose
Self-review the implementation and tests before producing the final artifact.
Apply the review shared skill, then consolidate into implementation-plan.

## Inputs
- `developer/dev-implementation`: implementation steps with code
- `developer/dev-tests`: test suites

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
