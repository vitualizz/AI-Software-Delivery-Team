# Decision Preservation — Shared Skill

## Purpose

After producing a significant decision or recommendation, ensure a **permanent
organizational knowledge record** is written — distinct from the change-scoped
artifact. This is what transforms ASDT from a per-change tool into a knowledge
system that accumulates domain expertise over time.

## When to Use

Any step that produces a **final decision artifact**:

- Architect: `decision-record` → `architectural-decision`
- Developer: `review` → `implementation-plan`
- Security: `hardening-checklist` → `hardening-checklist`
- QA: `quality-report` → `quality-report`
- Any specialist's **last step** that produces a recommendations artifact

## Protocol

1. **Extract What**: the decision or recommendation in one clear sentence.
2. **Extract Why**: the driving constraint, user request, or force that made this
   decision necessary.
3. **Extract Where**: the change ID and primary artifact path
   (e.g. `.asdt/artifacts/{change}/architectural-decision.yaml`).
4. **Extract Learned** (optional): a gotcha, rejected alternative, or edge case
   worth remembering for future changes.
5. **Shape as Entry**:
   - `Title`: `"{specialist-role}: {change-name}"` (e.g. `"architect: add-auth"`)
   - `Type`: `architecture` for design decisions, `decision` for policy/choice
   - `Content.What`: the extracted What sentence
   - `Content.Why`: the extracted Why
   - `Content.Where`: the extracted Where path
   - `Content.Learned`: optional

6. **Include a `summary` field** in your output artifact payload (≤ 150 tokens):
   ```yaml
   summary: "Chose JWT over sessions for stateless auth — enables horizontal scaling"
   ```
   The runner reads `payload["summary"]` and calls `memory.Save` automatically at
   run end. **You do not call Save yourself.**

## Context Budget

No added input budget — this skill operates on the artifact you already produced.
The `summary` field it adds to your payload is the only output overhead.

## Output Schema

Your final artifact payload MUST include:

```yaml
summary: string  # ≤150 tokens — the decision in one sentence
```

The runner lifts this into a permanent `memory.Entry` of type `architecture` or
`decision` so future specialists can query it via `knowledge-recall`.
