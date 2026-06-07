# Evaluate Approaches — Architect Specialist

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Retrieve every input named under
> `## Inputs` via `mem_search` (by its topic_key) then `mem_get_observation` —
> do not assume it is already in your context. Persist your one output via
> `mem_save` under the `output_topic_key` declared for this step in `workflow.yaml`,
> then return a structured summary envelope (status, summary, output topic_key, open_items).

## Purpose
Compare 2-3 viable technical approaches for the key architectural decision.
Choose one with explicit reasoning. Document why alternatives were rejected.

## Inputs
- `architect/constraints-analysis`: hard constraints, soft constraints, opportunities (produced by load-constraints)

Retrieve via mem_search + mem_get_observation by topic_key.

Extract: hard_constraints (limits approach space), opportunities (could favor one approach).

## Context budget
architect/constraints: max 1,200 tokens.

## Processing
1. IDENTIFY the central architectural question (the one decision that constrains everything else).
2. GENERATE 2-3 candidate approaches that respect hard constraints.
3. FOR EACH approach, evaluate across these dimensions:
   - Complexity: how hard to implement and maintain?
   - Performance: how does it behave under load?
   - Coupling: how tightly does it bind to existing components?
   - Reversibility: how hard to change later?
   - Familiarity: does the team already use this pattern?
4. SCORE each dimension as Low/Medium/High impact.
5. CHOOSE the approach with the best overall tradeoff — not necessarily the "best" in one dimension.
6. STATE the rejected alternatives with one-line reasons (these become ADR alternatives).

## Output
Produces: `architect/approaches`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  central_question: ""
  approaches:
    - name: ""
      description: ""
      complexity: "Low|Medium|High"
      performance: "Low|Medium|High"
      coupling: "Low|Medium|High"
      reversibility: "Low|Medium|High"
      familiarity: "Low|Medium|High"
      pros: []
      cons: []
  chosen: ""
  chosen_rationale: ""
  rejected:
    - name: ""
      reason: ""
  open_items: []
```
