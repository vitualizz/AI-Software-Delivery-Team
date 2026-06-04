# Test Case Generation — QA Specialist

## Purpose
Write structured test cases for the happy path, validated ACs, and critical edge cases.

## Inputs
- `qa/test-strategy`: which level each behavior is tested at
- `qa/edge-cases`: edge case inventory (critical + high priority only)

Extract from test-strategy: unit.what, integration.what, e2e.what.
Extract from edge-cases: edge_cases where priority="critical" or "high".

## Context budget
test-strategy summary + critical/high edge cases: max 2,000 tokens.

## Processing
For each item in test-strategy (unit.what, integration.what, e2e.what):
1. Write one test case in Given/When/Then format.
2. Specify the test level (unit/integration/e2e).
3. Include the test data setup needed.
4. Note any mocks or fixtures required.

For each critical/high edge case:
1. Write one test case covering the edge scenario.
2. Specify what "expected behavior" means precisely (status code, error message, data state).

Do NOT write code — write structured test specifications that a developer can implement.

## Output
Produces: `qa/test-cases`

Schema:
```yaml
payload:
  test_cases:
    - id: "TC-001"
      title: ""
      level: "unit|integration|e2e"
      ac_ref: ""
      given: ""
      when: ""
      then: ""
      setup_required: ""
      mocks_required: []
      priority: "critical|high|medium|low"
  total_count: 0
  critical_count: 0
  open_items: []
```
