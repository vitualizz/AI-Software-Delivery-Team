# Feasibility Scan — Researcher Specialist

## Purpose
Assess each ideated direction for feasibility — a green/yellow/red verdict with
supporting evidence and an effort estimate. This is the convergence-enabling
step: it gives `discovery-brief` the grounded signal it needs to recommend one
direction without guessing.

## Inputs
- `researcher/ideation` (injected) — the `problem_framing` and `ideas[]`.
- If this input arrives as an `### INPUT ...: UNRESOLVED` block, record the gap
  in `open_items` and proceed with whatever ideas are available (per
  `parallel-retrieval.md`).

## Context budget
`ideation`: max 1,500 tokens. Extract `problem_framing` and the `ideas[]` ids,
`what`, and `why`. Discard anything heavier.

## Processing
1. Produce EXACTLY one `Feasibility` per `Idea.id` from the ideation artifact —
   no more, no fewer. The set of `feasibilities[].idea_id` must equal the set of
   `ideas[].id`.
2. Each `Feasibility.idea_id` is a foreign key — it MUST match an existing
   `Idea.id`. If you would emit a feasibility whose `idea_id` has no matching
   idea, HALT and record the dangling id in `open_items` rather than inventing an
   idea to match it.
3. Assign a `verdict`: `green` (clearly feasible), `yellow` (feasible with
   caveats or unknowns), `red` (blocked or impractical as framed).
4. Attach `evidence` — the concrete reason for the verdict (a constraint, a
   dependency, a known limitation), not a restatement of the verdict.
5. Estimate `effort`: `low` | `medium` | `high`.

Do not converge here — every idea gets assessed. Choosing the winner is
`discovery-brief`'s job.

## Output
Produces: `researcher/feasibility`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  feasibilities:
    - idea_id: ""             # FK — MUST match an existing Idea.id from researcher/ideation
      verdict: "green | yellow | red"
      evidence: ""            # the concrete reason for the verdict
      effort: "low | medium | high"
  open_items: []
```

**Feasibility** (value-object):
```yaml
idea_id: ""               # FK to Idea.id — exactly one Feasibility per Idea
verdict: "green | yellow | red"
evidence: ""              # concrete supporting reason, not a restatement
effort: "low | medium | high"
```
