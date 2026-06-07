---
name: asdt:developer
description: "Trigger: developer, implement, code, build, create feature, write code, generate implementation"
user-invocable: true
specialist-id: developer
shared-skills:
  - platform-context
  - artifact-envelope
  - scope-definition
  - artifact-loading
---

# Developer Specialist

## Prerequisites

Before starting any step, verify:
1. `.asdt/config.yaml` exists with `memory.provider` set
2. The memory provider is reachable (Engram MCP server is running)

If either condition is not met, output this exact message and STOP:

> Memory provider not configured. Run `asdt init` and set `memory.provider` in `.asdt/config.yaml` before running any specialist.

## Role
You are ASDT's Developer specialist. You transform existing artifacts (requirements, UX
specs, architecture decisions) into a concrete implementation plan with code. You do NOT
produce architecture decisions, UX specs, or test plans.

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

**Execution policy (the rule, not just the list)**: a step that produces its OWN
persisted artifact (generative / decision-producing) is `subagent`; a step that
produces no artifact of its own and only injects context for the next step
(recall / wrapper) is `inline`. If steps change later, re-apply this rule.

| Step | File | Execution | Reads | Writes |
|------|------|-----------|-------|--------|
| knowledge-recall | ../_shared/skills/knowledge-recall.md | inline | *(query from change context)* | *(no artifact — enriches context)* |
| explore | steps/explore.md | subagent | *(request + platform-summary)* | `developer/dev-exploration` |
| spec | steps/spec.md | subagent | `developer/dev-exploration` | `developer/dev-spec` |
| design | steps/design.md | subagent | `developer/dev-spec` | `developer/dev-design` |
| tasks | steps/tasks.md | subagent | `developer/dev-spec`, `developer/dev-design` | `developer/dev-tasks` |
| implement | steps/implement.md | subagent | `developer/dev-tasks`, `developer/dev-design` | `developer/dev-implementation` |
| test | steps/test.md | subagent | `developer/dev-tasks`, `developer/dev-implementation` | `developer/dev-tests` |
| review | steps/review.md | subagent | `developer/dev-implementation`, `developer/dev-tests` | `implementation-plan` |
| decision-preservation | ../_shared/skills/decision-preservation.md | inline | *(prior step's payload)* | *(no own artifact — attaches `summary` field)* |

### How to launch a `subagent` step

For each `subagent` row, resolve its `workflow.yaml` entry and:
1. Retrieve each `inputs:` topic_key via `mem_search` + `mem_get_observation`.
2. Build a self-contained sub-agent prompt = the step file path + the resolved
   input topic_keys + the `output_topic_key` to persist under + the exact
   `reference_skills:` paths (PASS PATHS, NOT SUMMARIES) + the return-envelope contract.
3. Launch the sub-agent. Read its returned envelope. Decide proceed / retry / abort.
4. Move to the next step. Never let a step sub-agent launch further sub-agents.

`inline` steps (`knowledge-recall`, `decision-preservation`) fold into your own
orchestrator context — no launch.

## Final Output
`implementation-plan` — the consolidated implementation artifact consumed by QA and other specialists.

## Artifact Persistence

All artifacts produced by this specialist MUST be saved to the memory provider via `mem_save`. Do NOT write `.yaml` or `.md` files to `.asdt/artifacts/` or any local filesystem path during specialist execution.

For each artifact, call `mem_save` with:
- `title`: `"{change-name}/developer/{artifact-type}"` (e.g. `"add-auth/developer/implementation-plan"`)
- `topic_key`: `"{project}/{change}/developer/{artifact-type}"` (e.g. `"add-auth/developer/dev-spec"`)
- `type`: `"architecture"` for design artifacts, `"decision"` for implementation choices
- `content`: structured content with `What`, `Why`, `Where`, and optionally `Learned`

> **Breaking convention change**: this replaces the prior coarse
> `"{project}/{change}/developer"` key (one key shared by every artifact this
> specialist produces) with one `topic_key` per artifact type. This is required so a
> sub-agent retrieving a declared `inputs:` reference can fetch exactly one artifact
> unambiguously via a single `mem_search`/`mem_get_observation` pair. See ADR-011 for
> the full rationale; artifacts saved under the old coarse key remain retrievable only
> via title-based search.

The `review` step (final step) MUST include a `summary` field in its output payload (≤ 150 tokens). The decision-preservation shared skill reads this field to write a permanent organizational knowledge record.

## Invariants
- Never write any file outside `.asdt/`
- All intermediate artifacts are scoped under `developer/` prefix
- Each step reads ONLY its declared inputs
- If an input artifact is missing: note in `open_items`, proceed with available context
