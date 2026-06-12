# ASDT Shared Skills

Cross-specialist utility files. They reach specialist contexts via two mechanisms:

- The **FIRST ACTION Read** at the top of every specialist `SKILL.md` body — the specialist explicitly reads `specialist-header.md` (and its `workflow.yaml`) before doing anything else
- `reference_skills:` in a specific step entry in `workflow.yaml` — injected by the orchestrator into that step's sub-agent prompt only

The `shared-skills:` frontmatter key in each specialist `SKILL.md` is **metadata/documentation only** — it records which shared skills a specialist relates to, but no loader (Claude Code, OpenCode, or the installer) resolves it into context.

These are **reference text injected into the active context**, not independently executable units. They have no `## Inputs` / `## Output` structure of their own.

## Index

### Structural — required in every specialist

| File | Purpose |
|---|---|
| `specialist-header.md` | Loaded via the FIRST ACTION Read in each specialist `SKILL.md` body. Contains the ORCHESTRATOR GATE and prerequisite logic. |
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

In a specialist's `SKILL.md` frontmatter (metadata/documentation only — NOT a loading mechanism):

```yaml
shared-skills: specialist-header, platform-context, artifact-envelope
```
