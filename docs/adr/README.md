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
| [ADR-009](ADR-009-deterministic-platform-init.md) | Deterministic Platform Initialization | Active |
| [ADR-010](ADR-010-semantic-memory-provider.md) | Semantic Memory Provider over Key-Value Cache | Active |
| [ADR-011](ADR-011-specialist-pipelines-as-orchestration-plans.md) | Specialist Pipelines as Orchestration Plans | Active |

## Recommended Reading Order

1. **ADR-001** — Why ASDT is prompt-only with no runtime execution
2. **ADR-006** — The specialist model and how the five specialists relate
3. **ADR-011** — How specialists orchestrate sub-agents and hand off work through Engram

## Gap Notes

**ADR-007**: Never written. The gap in numbering is intentional.

**ADR-008**: Described a Go `SpecialistRunner` removed in commit `1fa209e`. The file was deleted when ADR-011 superseded it. ADR-011 is self-contained and does not require ADR-008 to understand the current design.
