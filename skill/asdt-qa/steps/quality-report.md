# Quality Report — QA Specialist

## Purpose
Produce the final test-plan artifact. Apply the report shared skill to consolidate
test cases and AC validation into a coherent quality document.

## Inputs
- `qa/test-cases`: all test cases
- `qa/ac-gaps`: AC validation results

Apply the extraction rules in the report shared skill: from test-cases keep counts + critical cases only.
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
