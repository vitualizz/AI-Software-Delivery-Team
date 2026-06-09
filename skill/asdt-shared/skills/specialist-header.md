<!-- specialist-header.md — load-order: MUST be the first shared-skills entry in every specialist SKILL.md -->
<!-- If this file is not loaded before the specialist SKILL.md body, the gate logic below will not run at the correct context position. -->

## Prerequisites

Before starting any step, verify:
1. `.asdt/config.yaml` exists with `memory.provider` set
2. The memory provider is reachable (Engram MCP server is running)

If either condition is not met, output this exact message and STOP:

> Memory provider not configured. Run `asdt init` and set `memory.provider` in `.asdt/config.yaml` before running any specialist.

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

<!-- co-location note: The specialist-specific complexity/tier paragraph and the Artifact Persistence block remain inline in each specialist SKILL.md. Do NOT move them here. The closing Tailored Workflow sentence also stays per-specialist. -->

**Execution policy (the rule, not just the list)**: a step that produces its OWN
persisted artifact (generative / decision-producing) is `subagent`; a step that
produces no artifact of its own and only injects context for the next step
(recall / wrapper) is `inline`. If steps change later, re-apply this rule.

### How to launch a `subagent` step

> Canonical protocol: `asdt-shared/skills/parallel-retrieval.md` — Cache Ledger Rule, Injection Format, UNRESOLVED degradation. Do not restate it here.

`inline` steps fold into your own orchestrator context — no launch.
