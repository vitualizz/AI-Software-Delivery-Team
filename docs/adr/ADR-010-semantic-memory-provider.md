# ADR-010: Semantic Memory Provider over Key-Value Cache

## Status

Accepted

---

## Context

The original `memory.Provider` interface exposed three methods:

```go
Load(ctx context.Context, key string) ([]byte, error)
Save(ctx context.Context, key string, data []byte) error
Name() string
```

The `Save` call was wired in `runner.go` after each workflow step, keyed as
`{change}/{specialist}/{step}` and storing the raw LLM response bytes.

This design has two fundamental problems:

1. **No read consumer.** The `Load` call was never invoked — not by the runner,
   not by any specialist step. The byte cache was write-only dead code.

2. **Wrong abstraction.** A byte cache keyed by step path cannot answer
   organizational questions like *"what authentication decisions exist?"* or
   *"what did the architect decide about session management last quarter?"*.
   It stores raw LLM output, not structured knowledge.

ASDT's stated goal is to be a **Software Delivery Knowledge System** — an
organization that accumulates expertise across changes, not a stateless code
generator. The key-value byte cache cannot support that goal.

---

## Decision

Redesign `memory.Provider` to a semantic, structured interface:

```go
type Provider interface {
    Save(ctx context.Context, entry Entry) error
    Search(ctx context.Context, query string) ([]Entry, error)
    Get(ctx context.Context, id string) (*Entry, error)
    Name() string
}
```

With a typed `Entry` record:

```go
type Entry struct {
    ID       string
    Title    string
    Type     EntryType   // decision | architecture | discovery | pattern | config | preference
    TopicKey string
    Content  EntryContent
    Metadata map[string]string
    SavedAt  time.Time
}

type EntryContent struct {
    What    string `yaml:"what"`
    Why     string `yaml:"why"`
    Where   string `yaml:"where"`
    Learned string `yaml:"learned,omitempty"`
}
```

**NullProvider** (default, no external dependencies) writes entries as YAML files
to `.asdt/runs/{runID}/{slug}.yaml`. Implements `Search` via substring walk over
the runs directory. Slug = topic key with non-alphanumeric characters replaced by
`-`. The zero-value `NullProvider{}` degrades to a no-op so existing test fixtures
remain valid.

**EngramProvider** (opt-in, requires Engram MCP) delegates all operations to
an injected `MCPCaller` port:

```go
type MCPCaller interface {
    Call(ctx context.Context, tool string, args map[string]any) (map[string]any, error)
}
```

`Save` → `mem_save`, `Search` → `mem_search`, `Get` → `mem_get_observation`.
The `MCPCaller` single-method port keeps tests hermetic — no MCP SDK imported in
the core.

**Runner integration:**

- `loadOrganizationalContext` is called **once before step 0** in `Run()`. It
  queries `memory.Search(role+" "+change)` and injects the top 3 results as an
  `## Organizational Context` block into the first step's prompt (budget: ≤400 tokens).
  Errors degrade to empty string — never abort the run.

- `saveKnowledgeRecord` is called **once after the final step** in `Run()`. It
  calls `memory.Save` with a structured `Entry` (Title, Type, Content.What from
  `payload["summary"]`). Errors are swallowed — best-effort.

---

## Alternatives Rejected

**Keep key-value, add a separate `Query` method**

Would leave two parallel memory models (bytes + semantic) in the codebase. The
byte cache path stays dead code. Rejected: solves the read problem without
fixing the abstraction problem.

**Embed a vector database in NullProvider**

Rejected: violates the "default path fully offline and deterministic" invariant.
Substring scan over `.asdt/runs/` is sufficient for the local fallback. A vector
DB adds an external binary dependency and startup latency.

**Make EngramProvider call the MCP SDK directly**

Rejected: un-mockable in unit tests without the full MCP runtime. The `MCPCaller`
single-method port makes the provider testable with a five-line mock struct.

**Keep the old stub (ErrNotImplemented)**

Rejected: the stub was never replaced and the read side had zero consumers.
Keeping it indefinitely is indefinitely deferred delivery of the knowledge system.

---

## Consequences

- **Breaking interface change**: `Load(key)/Save(key,[]byte)` are removed. All
  callers updated atomically in one slice. The runner no longer calls the old
  per-step Save. Build is always green — there is no intermediate broken state.

- **Enables semantic retrieval**: `Search` returns structured entries that can
  inform future specialist runs. Cross-change knowledge accumulation is possible
  for the first time.

- **Knowledge becomes searchable**: The runner injects prior decisions into the
  first step prompt automatically. Specialists build on — rather than contradict —
  past decisions.

- **`.asdt/runs/` is the knowledge timeline foundation**: Even without Engram,
  every specialist run leaves a YAML record. The local filesystem becomes an
  auditable knowledge history.

- **`ErrNotImplemented` removed**: The stub sentinel error is deleted. Any code
  path that previously checked for it will no longer compile — an intentional
  compile-time signal to find and fix such checks.

- **EngramProvider is functional**: The Engram-backed provider is no longer a
  stub. It delegates to the real `mem_save`, `mem_search`, and
  `mem_get_observation` tools. The composition root in `cmd/asdt/main.go` wires
  `NullProvider` until a real `MCPCaller` implementation is provided.
