# Test — Developer Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Your inputs are INJECTED in
> this prompt by the orchestrator — do NOT fetch them. See
> `../asdt-shared/skills/parallel-retrieval.md` for the injected-input
> contract; if an input is marked UNRESOLVED, record it in `open_items` and
> proceed. Persist your one
> output via `mem_save` under the `output_topic_key` declared for this step in
> `workflow.yaml`, then return a structured summary envelope (status, summary,
> output topic_key, open_items).

## Purpose
Generate tests for the implementation, covering happy paths and key edge cases.

## Inputs
- `developer/dev-tasks`: task list with acceptance criteria references
- `developer/dev-implementation`: code snippets per task

Extract from dev-tasks: `tasks[].ac_ref`, `tasks[].id`.
Extract from dev-implementation: `steps[].code_snippets[].content` (signatures only, not full bodies).

## Context budget
dev-tasks AC list + dev-implementation function signatures: max 2,500 tokens.

## Mode resolution (do this FIRST)
1. Inherit `mode` and `allowedEditRoots` from `dev-implementation`'s payload — do NOT
   re-derive them independently. The `test` step always runs in the SAME mode `implement`
   resolved for this change.
2. If `dev-implementation.payload.mode` is `plan-only` → PLAN-ONLY MODE: emit the
   snippet-based artifact (unchanged schema below, `test_cases[].code_snippet`). Write NO
   host files.
3. If `dev-implementation.payload.mode` is `writing` → WRITING MODE: write real test files
   to disk, but ONLY to paths within the SAME `allowedEditRoots` `implement` used. Before
   each write, confirm the target path is under a declared root; if not, STOP and report the
   unsafe path in `open_items` — do not write. Run the written tests against the real
   implementation files `implement` produced.
4. Strict-TDD note: when `strict_tdd: true` in `.asdt/config.yaml`, this step still applies
   the RED→GREEN cycle gating described in `skills/test-generation.md` — mode resolution only
   determines WHERE the resulting test artifact lives (Engram snippet vs. real file), not
   whether TDD discipline applies.

## Processing

### Plan-only mode
For each task in dev-tasks:
1. Write one happy-path test covering the acceptance criterion (ac_ref) as a `code_snippet`.
2. Write one edge-case test for the most likely failure mode as a `code_snippet`.
3. Follow the existing test framework convention from platform-summary.
4. Use table-driven tests where appropriate.

### Writing mode
For each task in dev-tasks:
1. Confirm the test file path (from `dev-tasks` test entries / declared paths) is within the
   inherited `allowedEditRoots`; if it falls outside, STOP before writing it, do not expand
   scope, and record the unsafe path plus the triggering task in `open_items`.
2. Write one happy-path test covering the acceptance criterion (ac_ref) and one edge-case
   test for the most likely failure mode, as real test files on disk.
3. Follow the existing test framework convention from platform-summary; use table-driven
   tests where appropriate.
4. Run the written tests against the real implementation files within `allowedEditRoots` and
   record the result.

Do NOT test implementation internals — test behavior.

## Output
Produces: `developer/dev-tests`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

The output schema is mode-dependent — set `mode` to the inherited value and emit the matching shape.

### Plan-only mode schema
```yaml
payload:
  mode: "plan-only"
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

### Writing mode schema
```yaml
payload:
  mode: "writing"
  allowedEditRoots: []        # inherited verbatim from dev-implementation
  test_suites:
    - task_id: "T-001"
      test_cases:
        - id: "TC-001"
          title: ""
          type: "unit|integration"
          given: ""
          when: ""
          then: ""
          file: ""            # real path written, within allowedEditRoots
          result: "pass|fail" # from running the written test against the real implementation
  unsafe_skipped: []          # paths STOPPED on, with the triggering task_id
  open_items: []
```
