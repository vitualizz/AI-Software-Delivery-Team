# Decision Record — Architect Specialist

## Purpose
Document the chosen approach as an Architecture Decision Record (ADR).
An ADR is permanent — it explains why a decision was made so future engineers
understand the context, not just the outcome.

## Inputs
- `architect/approaches`: chosen approach, alternatives, rationale

Extract: central_question, chosen, chosen_rationale, rejected[].

## Context budget
architect/approaches: max 1,000 tokens.

## Processing
Write the ADR in standard format. Be specific:
- Context: explain the SITUATION that forced this decision. What changed? What problem?
- Decision: state the choice clearly in one sentence, then elaborate.
- Alternatives: one paragraph per rejected alternative — what it was and why it lost.
- Consequences: be honest. What does this approach make easier? What does it make harder?
  What tech debt does it introduce? What does it prevent us from doing later?

Quality gate: if the "Consequences" section only has positives, it is incomplete.
Every architectural decision has tradeoffs. Name the negative consequences explicitly.

## Output
Produces: `architect/adr`

Schema:
```yaml
payload:
  title: ""
  status: "accepted"
  context: ""
  decision: ""
  alternatives_considered:
    - name: ""
      why_rejected: ""
  consequences:
    positive: []
    negative: []
    technical_debt: []
  open_items: []
```
