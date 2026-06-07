# ADR-011: Specialist Pipelines as Orchestration Plans

Date: 2026-06-06
Status: Accepted
Supersedes: ADR-008

## Context

ADR-008 ("Context-Isolated Skill Steps") solved context-window growth by having
each specialist's `WorkflowStep` declare `InputRefs`/`OutputArtifact`, with a Go
`SpecialistRunner` (`internal/specialists/runner.go`) loading only a step's declared
inputs and writing intermediate artifacts to disk at
`.asdt/artifacts/{change}/{specialist-id}/{type}.yaml`. The artifact file was the
handoff mechanism — "the artifact IS the handoff. Context history does not flow
between steps."

That runner — along with `internal/llm/`, `internal/specialists/descriptor.go`, and
the `MCPCaller`/`EngramProvider` glue — was deleted in commit `1fa209e` ("feat(binary):
Stage C — remove specialist runner, llm, descriptor, MCPCaller"). The binary is now a
package manager only; specialist commands redirect to the calling AI assistant runtime.
**ADR-008's premise — a Go runner providing zero-infrastructure, disk-based step
isolation — no longer exists.**

Without that runner, a specialist `SKILL.md` that still reads as "Each step executes
in isolation. Only the declared inputs are loaded" is now a fiction: there is nothing
enforcing that isolation. If the calling assistant runs all steps inline in one
continuous thread (the only option such a file describes), every heavy generative
step accumulates its full payload in that one thread — exactly the unbounded context
growth ADR-008 was written to prevent, but now with no mechanism bounding it.

Separately, two shared skills injected into specialist step prompts —
`skill/_shared/skills/knowledge-recall.md` and
`skill/_shared/skills/decision-preservation.md` — still contain prose describing
runner behavior that no longer exists ("The runner performs `memory.Search(query)`
and injects results...", "The runner reads `payload[\"summary\"]` and calls
`memory.Save` automatically... You do not call Save yourself."). Left as-is, this
actively misdirects the executing assistant into believing something else handles
persistence — silently breaking the artifact handoff.

## Decision

Convert each specialist `SKILL.md` from a self-executing inline pipeline description
into an **orchestration plan**: a declarative routing document that the calling
assistant (Claude Code, OpenCode, or any other runtime-agnostic AI runtime) reads and
drives — mirroring the proven pattern this repository already uses for its own SDD
phase pipeline (Dependency Graph + phase table + Result Contract).

Concretely:

1. **SKILL.md = plan, not executor.** Each specialist file gains an
   `## Orchestration Plan` section with an `ORCHESTRATOR GATE` block stating that the
   calling assistant is the SOLE orchestrator and MUST launch every `subagent`-mode
   step via its native delegation primitive (Agent/Task) — never inline in the
   specialist's own thread. The specialist file never calls Agent/Task itself; it only
   tells the calling assistant what to launch and how.

2. **Graduated execution policy — a testable rule, not a list.** A step that produces
   its OWN persisted artifact (generative / decision-producing — e.g. `spec`, `design`,
   `decision-record`, `threat-modeling`) is classified `subagent` and runs as an
   isolated sub-agent with its own bounded context. A step that produces no artifact of
   its own and exists only to inject context for the next step (`knowledge-recall`,
   `platform-analysis`, `decision-preservation`) is classified `inline` and runs in the
   orchestrator's own context. This rule is stated explicitly in each plan so it stays
   verifiable if the step inventory changes later.

3. **`workflow.yaml` becomes the canonical, machine-readable launch spec.** Each step
   entry gains additive fields: `execution` (`subagent`|`inline`, required for every
   step), and for `subagent` steps — `inputs` (topic_key strings to retrieve via
   `mem_search`/`mem_get_observation`), `output_topic_key` (where to persist the
   produced artifact), and `reference_skills` (exact paths to inject into the
   sub-agent's prompt — paths, not summaries, mirroring this repo's own Sub-Agent
   Launch Pattern).

4. **Heavy steps gain executor framing.** Every `subagent`-mode `steps/*.md` file gets
   an `EXECUTOR` block at its top (mirroring the `Executor Override` pattern already
   established in `~/.claude/skills/sdd-spec/SKILL.md`): it states the file is executed
   by a sub-agent that does the described work directly, retrieves its declared inputs
   itself via `mem_search` + `mem_get_observation`, persists its own output via
   `mem_save` under the `output_topic_key`, does not call Agent/Task/delegate, and is
   not responsible for orchestrating the rest of the pipeline.

5. **Two-level model preserved.** The architecture stays exactly two levels — calling
   assistant (orchestrator) → step sub-agent. A step sub-agent never launches further
   sub-agents. This mirrors the rule already enforced across this repo's `sdd-*` phase
   executors.

6. **Engram is the sole cross-context handoff**, replacing the deleted disk-artifact
   mechanism. `mem_save`/`mem_search` move to one `topic_key` per artifact type
   (`{project}/{change}/{specialist}/{artifact-type}`, e.g.
   `"add-auth/developer/dev-spec"`) — replacing the prior coarse
   `{project}/{change}/{specialist}` key — so a sub-agent retrieving a declared input
   can fetch it unambiguously with a single `mem_search`/`mem_get_observation` pair.
   This mirrors this project's own `sdd/{change-name}/{artifact-type}` convention.

7. **Runner-drift language is corrected** in `knowledge-recall.md` and
   `decision-preservation.md`: the executing sub-agent now calls `mem_search` and
   `mem_save` itself — there is no runner performing these calls on its behalf.

## Alternatives Considered

**Keep maximal inline context (status quo, minus the runner)** — rejected. This is the
exact problem ADR-008 fought, now with no runner to bound it. Every heavy step would
accumulate its full payload in one long-running specialist thread with zero isolation.

**Reintroduce a Go runner / disk-artifact handoff (restore ADR-008's mechanism)** —
rejected. That runner was deliberately deleted as part of the binary simplification in
commit `1fa209e`; restoring it contradicts ADR-001's runtime-agnostic, prompt-first
delivery model and reintroduces the exact Go-package-per-capability coupling this
project has been removing.

**Full per-step sub-agent orchestration with a third nesting level (build-level
sub-agents launching their own sub-agents)** — rejected. This exceeds the two-level
model and contradicts the "phase executor never delegates further" rule already
enforced across this repo's `sdd-*` phase skills (see "Executor Override" /
"You are not the orchestrator. Do NOT call the Task tool.").

## Consequences

Positive:
- Each heavy generative step gets a clean, isolated context budget — the same
  isolation goal ADR-008 pursued, achieved through real sub-agent boundaries instead
  of a now-nonexistent runner.
- Mirrors the validated SDD orchestration pattern already proven in this repo
  (`skill/SKILL.md` → `sdd-*` phase sub-agents), reducing the number of distinct
  orchestration models contributors must learn.
- Fully runtime-agnostic and prompt-only — no new Go source, no
  `internal/specialists/` package, no disk-artifact handoff path is reintroduced.
- The execution-mode classification is a stated, testable rule
  ("produces its own artifact → subagent; injects context with no artifact of its own
  → inline"), not an arbitrary per-step list — it stays verifiable as steps evolve.

Negative:
- **Breaking convention change**: the `topic_key` granularity moves from one coarse
  key per specialist to one key per artifact type. Artifacts saved under the old
  coarse `{project}/{change}/{specialist}` key remain retrievable only via title-based
  search; they are not auto-migrated (Engram artifacts are per-change, ephemeral
  planning records — no migration script is provided).
- Orchestration correctness now depends on the calling assistant honoring the
  `ORCHESTRATOR GATE` — there is no runtime enforcement of the two-level model; a
  non-compliant assistant could still run everything inline.
- More sub-agent launches mean more latency and cost for borderline steps. This is
  mitigated by the graduated policy keeping light context-injection steps inline.
