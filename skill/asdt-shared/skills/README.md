# ASDT Shared Skills

Cross-specialist utility files. They are loaded into specialist steps via two mechanisms:

- `shared-skills:` frontmatter in a specialist's `SKILL.md` — loaded for every step
- `reference_skills:` in a specific step entry in `workflow.yaml` — loaded only for that step

These are **reference text injected into the active context**, not independently executable units. They have no `## Inputs` / `## Output` structure of their own.

## Index

### Structural — required in every specialist

| File | Purpose |
|---|---|
| `specialist-header.md` | Must be the first `shared-skills:` entry in every specialist `SKILL.md`. Contains the ORCHESTRATOR GATE and prerequisite logic. |
| `executor-header.md` | Injected into every `subagent` step prompt. Instructs the executor: do the single assigned step, do NOT orchestrate or delegate. |

### Artifact contracts

| File | Purpose |
|---|---|
| `artifact-envelope.md` | Defines the required YAML envelope every artifact must conform to (`schema_version`, `agent`, `change_id`, `input_refs` as Engram topic_keys per ADR-011, `payload`, `open_items`). |
| `artifact-loading.md` | How a specialist's first artifact-consuming step (declared `inputs: []`) retrieves upstream artifacts from Engram (`mem_search` → `mem_get_observation`), extracts relevant fields, and records missing artifacts in `open_items[]`. |
| `parallel-retrieval.md` | The canonical orchestrator fetch-once cache pattern — prevents duplicate Engram lookups when multiple steps need the same artifact. |

### Workflow utilities

| File | Purpose |
|---|---|
| `knowledge-recall.md` | Queries organizational memory for prior decisions relevant to the current change. Used as the first inline step in most specialists. |
| `platform-context.md` | Injects the project's detected platform knowledge (stack, conventions, design fingerprint) into a specialist's context. Includes the project-level reuse guard (prefer `.asdt/knowledge/platform-summary.yaml`) and the fallback extraction rules; also backs the inline `platform-analysis` workflow step. |
| `decision-preservation.md` | Saves a permanent organizational knowledge record after a significant decision is produced. Used as the final inline step in most specialists. |
| `scope-definition.md` | Guidelines for defining explicit, unambiguous project scope. Used by Architect and Developer. |
| `report.md` | Generates a structured handoff document from multiple intermediate artifacts; includes the extraction rules (200-token budget per artifact). Used as the consolidation step in UX/UI, Architect, QA, Security, and PM. |

## How to Reference

In `workflow.yaml` (step-specific):

```yaml
- name: system-design
  execution: subagent
  reference_skills:
    - ../asdt-shared/skills/platform-context.md
    - ../asdt-shared/skills/scope-definition.md
```

In a specialist's `SKILL.md` frontmatter (loaded on every step):

```yaml
shared-skills: specialist-header, platform-context, artifact-envelope
```
