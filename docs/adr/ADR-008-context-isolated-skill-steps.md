# ADR-008: Context-Isolated Skill Steps

Date: 2026-06-04
Status: Accepted

## Context

The `runStep` function in `internal/specialists/runner.go` passes maximum context to
every step: all shared skills, all input artifacts from `descriptor.Artifacts.Reads`,
and the full `platform.yaml`. This means:

1. A step that only needs one artifact receives the entire artifact set.
2. Context window grows monotonically across steps.
3. The LLM must mentally partition relevance — which it handles poorly under token pressure.
4. Intermediate outputs are saved to `Memory` but never fed as structured input to
   subsequent steps. Steps effectively start from the same maximal context every time.

The conceptual mistake: specialists were modeled as a "role + steps" (a prompt with a
numbered list) rather than as an orchestrator dispatching bounded skill invocations.

## Decision

Each `WorkflowStep` declares exactly what it reads (`InputRefs`) and what it produces
(`OutputArtifact`):

- `InputRefs []string` — artifact types this step reads. Empty = backward compat
  (falls back to `descriptor.Artifacts.Reads`).
- `OutputArtifact string` — artifact type written immediately when the step completes.
  Empty = no mid-run write (last step falls back to `writeArtifacts`).

The runner loads only `step.InputRefs` per invocation via `loadStepInputs`. Intermediate
artifacts are written to `.asdt/artifacts/{change}/{specialist-id}/{type}.yaml` via the
existing `Store.Write` key convention. The next step finds its input on disk — not in the
conversation context.

The artifact IS the handoff. Context history does not flow between steps.

## Alternatives Considered

1. **Keep maximal context, rely on model attention** — Rejected: unpredictable at scale;
   wastes tokens on irrelevant context; correctness depends on attention mechanics that
   vary by model and context length.

2. **Full sub-agent orchestration per step** — Rejected: over-engineering; adds latency
   and cost; the artifact-file handoff achieves the same isolation at zero infra cost.

3. **Memory-based step handoff** — Rejected: `Memory` is optional infrastructure
   (default: `NullProvider`). Making step handoff depend on memory would break the
   default configuration. Artifacts on disk are always available.

## Consequences

**Positive:**
- Token budget per step is controlled and predictable.
- Steps are independently testable — each can be invoked with a fixed input artifact.
- Context window growth is bounded by the step's declared inputs.
- Step responsibilities are explicit in the descriptor — readable as documentation.
- New artifact naming convention (`{specialist}/{type}`) makes ownership visible in `.asdt/`.

**Negative:**
- Intermediate artifacts accumulate in `.asdt/` during a specialist run.
- Parity test must be redesigned (no longer assumes last-step-only write).
- Backward compat requires the empty-field fallback in the runner.
- The `platform-analysis` step must be the first step for each specialist — currently
  only Developer and UX/UI have it; QA and Security need it added.
