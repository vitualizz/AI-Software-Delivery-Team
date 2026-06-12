# Discovery Brief — Researcher Specialist

## Purpose
Converge the explored space to ONE recommended direction for PM handoff. This is
the Researcher's final step: it turns divergent ideas plus their feasibility
verdicts into a single clear recommendation, plus the won't-do candidates that
seed PM's out-of-scope list. The brief's `summary` is the prose the orchestrator
hands to `/asdt-pm`.

## Inputs
- `researcher/ideation` (injected) — `problem_framing`, `ideas[]`.
- `researcher/feasibility` (injected) — `feasibilities[]` keyed by `idea_id`.
- If EITHER input arrives as an `### INPUT ...: UNRESOLVED` block, record the gap
  in `open_items` and degrade gracefully — recommend from whatever survives,
  noting the reduced confidence (per `parallel-retrieval.md`).

## Context budget
Combined `ideation` + `feasibility`: max 2,500 tokens. Apply the `report.md`
extraction rules — pull `problem_framing`, the surviving ideas, and their
verdicts; discard the rest.

## Processing
1. Carry the `context` from `ideation.problem_framing`.
2. Build `options` — the surviving directions worth presenting (drop `red`
   ideas unless none survive).
3. Write `feasibility_notes` — a condensed green/yellow/red rationale per
   surviving option. **MANDATORY**: this is what makes the recommendation
   defensible. A brief without feasibility notes is incomplete — the step
   consumes `feasibility` precisely so this section can exist.
4. Choose exactly ONE `recommended_direction`. The brief always recommends one —
   never zero, never a tie.
5. Collect `wont_candidates` — the directions explicitly NOT recommended. These
   seed PM's `scope.out_of_scope` at handoff.
6. Write a `summary` (≤ 150 tokens) — the prose the orchestrator renders into the
   `/asdt-pm` RAW REQUEST. This is the load-bearing handoff text; make it
   self-contained.

## Output
Produces: `researcher/discovery-brief`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  context: ""                 # the problem_framing carried from ideation
  options: []                 # surviving directions considered
  feasibility_notes: []       # MANDATORY — condensed g/y/r rationale per option
  recommended_direction: ""   # exactly one — the brief always recommends
  wont_candidates: []         # directions not recommended — seed PM scope.out_of_scope
  summary: ""                 # ≤150 tokens — prose handed to /asdt-pm as the raw request
  open_items: []
```

## Downstream consumption
- The orchestrator renders `summary` + `recommended_direction` to prose and
  passes it as the RAW REQUEST to `/asdt-pm`.
- `pm/feature-intake` keeps `inputs: []` — UNCHANGED. The handoff is prose-only;
  the Researcher introduces no declared-input contract on PM.
- The belt is SOFT: a prior `researcher/discovery-brief` MAY be recalled via
  `knowledge-recall` for richer context, but it is never a required input.
