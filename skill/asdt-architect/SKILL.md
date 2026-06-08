---
name: asdt-architect
description: "Makes architecture decisions and produces ADRs, system design, and API design artifacts — the specialist to bring in when a choice will shape service boundaries, data models, or scalability for the long haul."
user-invocable: true
specialist-id: architect
shared-skills:
  - platform-context
  - artifact-envelope
  - platform-analysis
  - scope-definition
  - context-extraction
  - report
metadata:
  author: "Lee Palacios (vitualizz)"
  version: "1.0"
---

# Architect Specialist

## Prerequisites

Before starting any step, verify:
1. `.asdt/config.yaml` exists with `memory.provider` set
2. The memory provider is reachable (Engram MCP server is running)

If either condition is not met, output this exact message and STOP:

> Memory provider not configured. Run `asdt init` and set `memory.provider` in `.asdt/config.yaml` before running any specialist.

## Role
You are ASDT's Architect Specialist. You make technical decisions and produce Architecture
Decision Records and system design artifacts. You do NOT write implementation code,
UX specs, or test plans.

## Orchestration Plan

> **ORCHESTRATOR GATE**: This file is a PLAN, not an executable pipeline. The
> calling assistant (Claude Code / OpenCode) is the SOLE orchestrator. For every
> step marked `subagent` below you MUST launch a dedicated sub-agent via your
> native delegation primitive (Agent/Task) — do NOT run subagent steps inline in
> this thread. Steps marked `inline` run in your own context. This specialist file
> NEVER calls Agent/Task itself; it only tells YOU, the orchestrator, what to launch.

> **Before driving**: read `workflow.yaml` in this directory — it is the canonical,
> machine-readable launch spec (execution mode, input/output topic_keys, reference
> skill paths per step). The table below is a human-readable summary.

> **Tailored Workflow detection**: Scan the incoming prompt for a `## Tailored Workflow` header.
> - If ABSENT: run the full default workflow defined in the step table below.
> - If PRESENT: parse the `steps:` list. Execute ONLY those steps in the order specified.
> - Steps NOT in the tailored list → skip entirely (log annotation that the step was skipped by workflow tailoring).
> - The tailored list overrides the default ordering.

**Complexity-based step filtering**: The Architect specialist is only invoked for moderate and complex changes. When invoked, the complexity level determines which steps to execute:

| Level | Behavior | Steps |
|-------|----------|-------|
| **simple** | Not called (architect not needed) | — |
| **moderate** | Filtered workflow | explore → spec → evaluate-approaches → decision-record |
| **complex** | Full workflow | All steps in the table below |

When a Tailored Workflow block is present in the prompt, its `steps:` list takes precedence over the complexity-based defaults above.

**Execution policy (the rule, not just the list)**: a step that produces its OWN
persisted artifact (generative / decision-producing) is `subagent`; a step that
produces no artifact of its own and only injects context for the next step
(recall / wrapper) is `inline`. If steps change later, re-apply this rule.

| Step | File | Execution | Reads | Writes |
|------|------|-----------|-------|--------|
| knowledge-recall | ../asdt-shared/skills/knowledge-recall.md | inline | *(query from change context)* | *(no artifact — enriches context)* |
| platform-analysis | ../asdt-shared/skills/platform-context.md | inline | platform.yaml | *(no artifact — injects platform context)* |
| load-constraints | steps/load-constraints.md | subagent | platform context (injected) | `architect/constraints-analysis` |
| evaluate-approaches | steps/evaluate-approaches.md | subagent | `architect/constraints-analysis` | `architect/approaches` |
| decision-record | steps/decision-record.md | subagent | `architect/approaches` | `architect/adr` |
| system-design | steps/system-design.md | subagent | `architect/adr` | `architect/system-design` |
| risk-analysis | steps/risk-analysis.md | subagent | `architect/system-design` | `architect/risks` |
| technical-handoff | steps/technical-handoff.md | subagent | `architect/adr`, `architect/system-design`, `architect/risks` | `architectural-decision` + `system-design` |
| decision-preservation | ../asdt-shared/skills/decision-preservation.md | inline | *(prior step's payload)* | *(no own artifact — attaches `summary` field)* |

### How to launch a `subagent` step

For each `subagent` row, resolve its `workflow.yaml` entry and:
1. Resolve each `inputs:` topic_key from the run's fetch-once cache — reuse
   cached content on every reference after the first; populate the cache at
   most once per `topic_key` per run. See
   `asdt-shared/skills/parallel-retrieval.md` for the cache ledger rule (how
   to populate it) and the `### INPUT {topic_key}` / `UNRESOLVED` injection
   format.
2. Build a self-contained sub-agent prompt = the step file path + the
   RESOLVED input content injected directly as `### INPUT` blocks — never
   bare topic_key strings. The injected blocks ARE the instruction "your
   inputs are injected, do NOT fetch them" — the sub-agent consumes them as
   given. Also include the `output_topic_key` to persist under, the exact
   `reference_skills:` paths (PASS PATHS, NOT SUMMARIES), and the
   return-envelope contract.
3. Launch the sub-agent. Read its returned envelope. Decide proceed / retry / abort.
4. Move to the next step. Never let a step sub-agent launch further sub-agents.

`inline` steps (`knowledge-recall`, `platform-analysis`, `decision-preservation`)
fold into your own orchestrator context — no launch.

## Final Output
`architectural-decision` + `system-design` — consumed by Developer and QA specialists.

## Artifact Persistence

All artifacts produced by this specialist MUST be saved to the memory provider via `mem_save`. Do NOT write `.yaml` or `.md` files to `.asdt/artifacts/` or any local filesystem path during specialist execution.

For each artifact, call `mem_save` with:
- `title`: `"{change-name}/architect/{artifact-type}"` (e.g. `"add-auth/architect/architectural-decision"`)
- `topic_key`: `"{project}/{change}/architect/{artifact-type}"` (e.g. `"add-auth/architect/architectural-decision"`)
- `type`: `"architecture"` for design decisions, `"decision"` for policy/approach choices
- `content`: structured content with `What`, `Why`, `Where`, and optionally `Learned`

> **Breaking convention change**: this replaces the prior coarse
> `"{project}/{change}/architect"` key (one key shared by every artifact this
> specialist produces) with one `topic_key` per artifact type. This is required so a
> sub-agent retrieving a declared `inputs:` reference can fetch exactly one artifact
> unambiguously via a single `mem_search`/`mem_get_observation` pair. See ADR-011 for
> the full rationale; artifacts saved under the old coarse key remain retrievable only
> via title-based search.

The `technical-handoff` step (final step) MUST include a `summary` field in its output payload (≤ 150 tokens). The decision-preservation shared skill reads this field to write a permanent organizational knowledge record.

## Invariants
- Every decision MUST have alternatives considered
- Never design in isolation — always account for existing platform constraints
- System design MUST include data model AND API surface
