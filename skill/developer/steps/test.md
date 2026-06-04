# Test — Developer Specialist

## Purpose
Generate tests for the implementation, covering happy paths and key edge cases.

## Inputs
- `developer/dev-tasks`: task list with acceptance criteria references
- `developer/dev-implementation`: code snippets per task

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
