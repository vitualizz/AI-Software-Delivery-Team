# ADR-007: MemoryProvider Abstraction

Date: 2026-06-04
Status: Accepted

## Context

The MVP has no cross-session memory. Each specialist invocation starts from a clean
slate: no recall of previous decisions, no awareness of prior architectural tradeoffs,
no continuity across a multi-day change. This is acceptable for short tasks but
limits the system's usefulness on long-running changes where a developer might run
`/asdt:developer` on Monday and again on Thursday.

Engram is an obvious integration target for cross-session memory. However,
hard-coding Engram as a dependency in the core would:

1. **Break offline usage**: ASDT is designed to work without network access. An
   unconditional Engram import means any machine without Engram configured would
   fail at startup.
2. **Create a vendor dependency in core**: `internal/specialists`, `internal/pipeline`,
   and `internal/artifact` would transitively depend on an external MCP server,
   coupling the delivery guarantees of those packages to Engram's availability.
3. **Complicate testing**: Unit tests for `SpecialistRunner` would need an Engram
   server running or a complex mock. A NullProvider makes the default test path trivial.
4. **Prevent alternative backends**: Some teams use file-based memory, others use
   agent-native memory. Hard-coding Engram forecloses these options.

## Decision

Introduce a `memory.Provider` interface in `internal/memory/provider.go` with
exactly three methods:

```
Load(ctx context.Context, key string) ([]byte, error)
Save(ctx context.Context, key string, data []byte) error
Name() string
```

The default implementation is `NullProvider` — a no-op struct where `Load` returns
`nil, nil`, `Save` returns `nil`, and no files or network calls are made. Memory is
opt-in infrastructure, not required infrastructure.

`EngramProvider` is an optional adapter enabled only when `.asdt/config.yaml`
contains `memory.provider: engram`. The adapter stub returns `ErrNotImplemented`
for both `Load` and `Save` until the Engram MCP integration is wired in a future
change. Its existence at this stage allows the composition root and configuration
schema to be finalized without requiring a working Engram client.

`SpecialistRunner` calls `memory.Save` as best-effort: a non-nil error from Save
is logged but never causes the run to abort. `memory.Load` has no consumer in the
current Runner; it is reserved for future read-side features.

No core package (`internal/artifact`, `internal/pipeline`, `internal/prompt`,
`internal/knowledge`) imports `internal/memory`. Only `internal/specialists` and
`cmd/asdt` depend on it.

## Alternatives Considered

**Import Engram directly in `internal/specialists`** — rejected. Vendor coupling;
breaks offline default; breaks the runtime-agnosticism invariant from ADR-001.
Testing requires a live Engram instance.

**File-based memory under `.asdt/memory/`** — rejected. This reinvents what Engram
already does well (cross-session, cross-project recall, semantic search). A file
store would need an index, a key scheme, and a search mechanism — all out of scope
for this change.

**No memory abstraction at all** — rejected. Cross-session context is a genuine user
need for multi-day changes. Retrofitting a port interface later (after direct Engram
calls are scattered across specialists) would be a much larger refactor than
introducing the port now.

**Multiple simultaneous providers (fan-out)** — rejected. Over-engineering for MVP.
A single active provider per deployment is the right default; fan-out can be added
as a composite provider later without changing the interface.

## Consequences

Positive:
- ASDT works fully offline with `NullProvider` as the default — no configuration
  required for a working system.
- Engram is opt-in: users who want cross-session memory configure it; others get
  a clean no-op with zero performance cost.
- Testing is trivial: `NullProvider` requires no setup; all `SpecialistRunner`
  unit tests use it without mocking an external service.
- Memory backends are interchangeable with no core changes: swap `NullProvider`
  for `EngramProvider` (or a future `FileProvider`) at the composition root.

Negative:
- `NullProvider` means no cross-session specialist context by default. A user running
  `/asdt:developer` twice on the same change will not benefit from memory of the first
  run unless they explicitly configure a provider.
- The `EngramProvider` stub returns `ErrNotImplemented` until the MCP integration
  is completed. If a user configures `memory.provider: engram` before that
  integration lands, Save calls will silently fail (logged, not aborted). This is
  mitigated by documenting the stub status and treating Save as best-effort.
- The read-side (`Load`) has no consumer in the current Runner. Its presence in the
  interface is forward-looking; an unused interface method carries a small
  maintenance cost.
