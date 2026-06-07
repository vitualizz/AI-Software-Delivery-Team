# Test Strategy — QA Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Define the testing approach: which level tests cover which behaviors, and why.

## Inputs
- `qa/edge-cases`: edge case inventory with priorities

Retrieve via mem_search + mem_get_observation by topic_key.

Extract: edge_cases[].technique, edge_cases[].priority, critical_count.

## Context budget
qa/edge-cases (technique + priority summary): max 800 tokens.

## Processing
Apply the test pyramid:
1. UNIT (base): test individual functions/methods in isolation. Use mocks at boundaries.
   - Coverage target: all business logic, all edge cases with clear expected outputs.
2. INTEGRATION (middle): test that components work together correctly.
   - Coverage target: API contracts, database operations, service interactions.
3. E2E / ACCEPTANCE (top): test the full flow from the user's perspective.
   - Coverage target: happy path + top 2 critical edge cases only.
   - These are expensive — keep them focused.

For each level, specify:
- What is tested at this level (and what is NOT — avoid duplication)
- Test data strategy (fixtures, factories, mocks, real DB?)
- Environment requirements (does this need a real DB? External services?)
- Acceptable flakiness tolerance (e2e: 0 flaky tests; unit: 0; integration: < 1%)

## Output
Produces: `qa/test-strategy`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  unit:
    what: []
    what_not: []
    data_strategy: ""
    coverage_target: ""
  integration:
    what: []
    what_not: []
    data_strategy: ""
    environment: ""
  e2e:
    what: []
    what_not: []
    data_strategy: ""
    flakiness_tolerance: ""
  total_test_estimate: ""
  open_items: []
```
