# ADR-005: Sequential Pipeline FSM → DAG Post-MVP

Date: 2026-06-04
Status: Accepted

## Context

The delivery pipeline has a natural ordering: requirements must exist before
planning, a plan must exist before implementation, and implementation must exist
before review. In the abstract, some steps could run in parallel (e.g.,
front-end and back-end implementation), but those cases require knowing the
dependency graph of the specific change — information that is not available at
the framework level for MVP.

The risk of building a DAG scheduler upfront is twofold: it adds complexity
that is not needed for the first two agents (requirements and developer), and it
constrains the artifact contract before real usage data exists. Building the
simplest correct model first, with a clear upgrade path, is the right call.

## Decision

The MVP uses a strict sequential FSM with four named states:

```
requirements → plan → implement → review
```

All transitions not listed above are illegal. Attempting an illegal transition
fails loudly with an error message naming the required predecessor state.

The FSM is persisted as `pipeline-state.yaml` under `.asdt/`. Every transition
is recorded with a timestamp. The current state is always derivable from the
file without reading the full transition log.

The `PipelineRunner` interface is designed so DAG execution is an additive
change: a DAG implementation provides the same interface, and agents never call
the FSM directly — they call `PipelineRunner`. Swapping the implementation does
not require changing any agent code.

## Alternatives Considered

**DAG from day one** — rejected. Over-engineering for a 2-agent MVP. The DAG
requires a scheduler, a dependency resolver, and a cycle-detection step. None
of these are needed when the pipeline is `requirements → plan`. Adding them
upfront increases complexity without adding user-visible value.

**Unordered / no pipeline state** — rejected. Agents depend on predecessor
outputs. The developer agent requires the requirements spec. Without enforced
ordering, agents that run out of sequence produce empty or invalid artifacts,
and the failure is silent until the output is inspected.

**Event-sourced pipeline** — rejected. Event sourcing is the right model for
distributed systems where state must be reconstructed from a log. ASDT is
local-first; the current state is always knowable by reading one file. The
complexity of event sourcing is not justified.

## Consequences

- Simple mental model: at any point you know exactly which state the pipeline
  is in, and what the valid next step is.
- Every state transition is logged with a timestamp in `pipeline-state.yaml`,
  giving a complete history of the delivery timeline.
- The DAG upgrade path is a `PipelineRunner` implementation swap: add a
  `dag.Runner` that implements the same interface, wire it in `cmd/asdt/main.go`,
  and no agent code changes.
- Illegal transitions produce actionable errors (e.g., "cannot advance to
  `implement`: current state is `requirements`, must first advance to `plan`"),
  not silent failures or panics.
- The sequential constraint will eventually be a limitation for large changes
  where front-end and back-end work could proceed in parallel. This is the
  accepted tradeoff for MVP simplicity.
