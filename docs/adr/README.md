# Architecture Decision Records

| ADR | Title | Status |
|---|---|---|
| [ADR-001](ADR-001-skill-first-runtime-agnostic-delivery.md) | Skill-First, Runtime-Agnostic Delivery | Active |
| [ADR-002](ADR-002-asdt-as-clean-boundary.md) | `.asdt/` as the Clean Boundary | Active |
| [ADR-003](ADR-003-artifact-envelope-as-inter-agent-contract.md) | Artifact Envelope as the Inter-Agent Contract | Active |
| [ADR-004](ADR-004-go-embed-local-override-for-prompts.md) | go:embed + Local Override for Prompts | Active |
| [ADR-005](ADR-005-sequential-pipeline-fsm-dag-post-mvp.md) | Sequential Pipeline FSM → DAG Post-MVP | Superseded by ADR-011 |
| [ADR-006](ADR-006-specialist-model.md) | Specialist Model | Active |
| ADR-007 | — | Not written |
| ADR-008 | Context-Isolated Skill Steps | Deleted — superseded by ADR-011 |
| [ADR-009](ADR-009-deterministic-platform-init.md) | Deterministic Platform Initialization | Partially superseded by ADR-013 (execution model) |
| [ADR-010](ADR-010-semantic-memory-provider.md) | Semantic Memory Provider over Key-Value Cache | Active |
| [ADR-011](ADR-011-specialist-pipelines-as-orchestration-plans.md) | Specialist Pipelines as Orchestration Plans | Active |
| [ADR-012](ADR-012-astro-site-github-pages.md) | In-Repo Astro Site under site/ Deployed via GitHub Pages Actions Mode | Active |
| [ADR-013](ADR-013-skill-only-context-detection.md) | Skill-Only Context Detection for asdt-init | Active |
| [ADR-014](ADR-014-skill-md-self-load-inline-gate.md) | SKILL.md Self-Load of specialist-header + Inline Orchestrator Gate | Active |
| [ADR-015](ADR-015-init-delegation-conformance.md) | Make asdt-init Delegation Conformant via a Net-New workflow.yaml and an Explicit analyst Default | Accepted — amended by ADR-016 |
| [ADR-016](ADR-016-init-empowerment-researcher-specialist.md) | Init Empowerment: Three-Step Init Flow and the Researcher Specialist | Active |

## Recommended Reading Order

1. **ADR-001** — Why ASDT is prompt-only with no runtime execution
2. **ADR-006** — The specialist model and how the five specialists relate
3. **ADR-011** — How specialists orchestrate sub-agents and hand off work through Engram

## Gap Notes

**ADR-007**: Never written. The gap in numbering is intentional.

**ADR-008**: Described a Go `SpecialistRunner` removed in commit `1fa209e`. The file was deleted when ADR-011 superseded it. ADR-011 is self-contained and does not require ADR-008 to understand the current design.

**Engram "ADR-011" collision**: An Engram-only memory record also labeled "ADR-011" proposed transcluding `specialist-header.md` via the `shared-skills:` frontmatter key — a loading mechanism that does not exist. It is NOT the on-disk ADR-011 above and was never an accepted ADR. ADR-014 supersedes that record and restores the on-disk ADR-011 position (inline gate in each `SKILL.md`).
