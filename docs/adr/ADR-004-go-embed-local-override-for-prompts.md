# ADR-004: go:embed + Local Override for Prompts

Date: 2026-06-04
Status: Accepted

## Context

Prompts are the core business logic of ASDT. They define agent behavior,
acceptance criteria, and output schemas. They must satisfy three requirements
simultaneously:

1. **Distributable as a single binary** — users who install the TUI via
   `go install` must get a working system without downloading prompts separately.
2. **Readable and editable without recompiling** — contributors improving
   prompts should not need a Go toolchain; the prompts are Markdown files, not
   code.
3. **Overridable per-project** — teams with specialized conventions (a
   particular test framework, a house style for user stories) should be able to
   customize agent behavior for their project without forking the repository.

No single prompt storage strategy satisfies all three out of the box.

## Decision

Prompts ship embedded in the binary via `go:embed skill/prompts` in the `skill`
package. At runtime, the override resolver checks directories in order and
returns the first match:

1. `.asdt/prompts/{role}/...` (project-local — most specific, wins)
2. `~/.config/asdt/prompts/{role}/...` (user-global)
3. `skill/prompts/roles/{role}/...` (packaged default via `go:embed`)

Each prompt fragment has a version: the SHA-256 of its content, truncated to 8
hexadecimal characters. The `Composer` returns a manifest mapping every
fragment name to its resolved version. The hash of that manifest is written to
`EnvelopeHeader.PromptVersion`.

## Alternatives Considered

**External prompt registry API** — rejected. Network dependency; breaks
air-gapped environments; adds a service to operate and secure; over-engineering
for an MVP.

**Prompts as Go string constants** — rejected. Not editable without recompile;
not readable as a standalone file; kills the "contributor who only knows
Markdown" path; no local override mechanism without code changes.

**Separate prompt binary or plugin** — rejected. Adds a second binary to
distribute and version; version skew between the binary and the prompt plugin
is a new failure mode with no benefit for MVP.

**Single canonical location without override** — rejected. Teams have
legitimate needs for per-project customization (domain vocabulary, output
format preferences, coding standards). Without an override path, the only
option is forking the whole repository.

## Consequences

- Single distributable binary: `go install` produces a self-contained
  executable with all default prompts embedded.
- Community contributors can customize prompts with zero Go knowledge by adding
  or editing files under `.asdt/prompts/` or `~/.config/asdt/prompts/`.
- Prompt drift is detectable: the `prompt_version` in every artifact envelope
  records which exact prompt version produced it. If behavior changes
  unexpectedly, compare the `prompt_version` across artifacts.
- Local overrides enable per-project customization without affecting the default
  prompts that ship with the binary.
- The override resolution order (project-local → user-global → packaged) follows
  the principle of least surprise: the most specific context wins.
