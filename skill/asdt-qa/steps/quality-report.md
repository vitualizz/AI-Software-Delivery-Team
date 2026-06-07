# Quality Report — QA Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Produce the final test-plan artifact. Apply the report shared skill to consolidate
test cases and AC validation into a coherent quality document.

## Inputs
- `qa/test-cases`: all test cases
- `qa/ac-gaps`: AC validation results

Retrieve via mem_search + mem_get_observation by topic_key.

Apply context-extraction: from test-cases keep counts + critical cases only.
From ac-gaps keep gap_count + open_items only.

## Context budget
qa/test-cases summary + qa/ac-gaps summary: max 1,000 tokens.

## Processing
Apply the `report` shared skill:
1. Check: does every testable AC have at least one test case? If not → open_item.
2. Check: are all critical edge cases covered? If not → open_item.
3. Compute coverage: testable ACs covered / total testable ACs.
4. Summarize test distribution across levels (unit/integration/e2e counts).
5. List any AC gaps that need upstream specialist input before testing can proceed.
6. Write a quality verdict: READY / READY WITH CAVEATS / BLOCKED.

## Output
Produces: `test-plan` (final cross-specialist artifact — single artifact, unlike
architect's dual-output `technical-handoff`; persist it once under this step's
`output_topic_key`)

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  test_summary:
    total_cases: 0
    unit: 0
    integration: 0
    e2e: 0
  ac_coverage:
    total_testable: 0
    covered: 0
    coverage_percent: ""
  ac_gaps: []
  quality_verdict: "READY|READY_WITH_CAVEATS|BLOCKED"
  verdict_rationale: ""
  test_cases: []     # full list from qa/test-cases
  open_items: []
```
