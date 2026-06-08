# Edge Case Analysis — QA Specialist

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
Systematically discover edge cases using structured techniques.
Edge cases are not random — they can be derived methodically.

## Inputs
- `qa/ac-list`: normalized acceptance criteria

Retrieve via mem_search + mem_get_observation by topic_key.

Extract: acceptance_criteria[].given/when/then (to derive boundaries).

## Context budget
qa/ac-list (AC text only, no metadata): max 1,200 tokens.

## Processing
Apply these edge case discovery techniques to each AC:

1. BOUNDARY VALUE ANALYSIS: for any numeric/length input, test:
   - Minimum valid value
   - Maximum valid value
   - Just below minimum (invalid)
   - Just above maximum (invalid)
   - Zero / empty / null

2. EQUIVALENCE PARTITIONING: group inputs into classes where behavior is identical.
   Test one case from each class (valid and invalid classes).

3. STATE TRANSITIONS: if the feature has states (draft/published, active/inactive):
   - Test each valid transition
   - Test each invalid transition (what happens when user tries to go backward?)

4. CONCURRENT ACCESS: if multiple users can touch the same data:
   - Two users editing simultaneously
   - Delete while another user is viewing
   - Race condition scenarios

5. PERMISSION BOUNDARIES: what happens when a user without permission tries each action?

6. NETWORK/SYSTEM FAILURES: what happens if the database is slow? If an external call fails?

## Output
Produces: `qa/edge-cases`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  edge_cases:
    - id: "EC-001"
      technique: "boundary|equivalence|state|concurrent|permission|failure"
      ac_ref: "AC-001"   # which AC this tests the edges of
      scenario: ""
      expected_behavior: ""
      priority: "critical|high|medium|low"
  critical_count: 0
  open_items: []
```
