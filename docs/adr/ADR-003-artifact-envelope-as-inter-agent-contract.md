# ADR-003: Artifact Envelope as the Inter-Agent Contract

Date: 2026-06-04
Status: Accepted

## Context

Agents must pass information to each other without direct communication. In
ASDT's model, agents run at different times (potentially days apart, in
different sessions), so they cannot share in-memory state. The format that
carries their outputs must be human-readable, git-diffable, versioned, and
carry a complete traceability chain so any artifact can be traced back to the
inputs and prompt versions that produced it.

The artifact is not just a data transfer mechanism — it is the audit trail.
This is the central thesis of ASDT: if an agent produced it, there must be a
record.

## Decision

Every agent output is wrapped in `Envelope[T]` — a typed YAML structure with a
uniform header and a typed payload. The header contains:

- `schema_version` — explicit contract version; consumers validate before reading
- `agent` — the producing role (e.g. `requirements`, `developer`)
- `change_id` — the logical change this artifact belongs to
- `created_at` — ISO 8601 timestamp
- `prompt_version` — SHA-256 (first 8 chars) of the composed-prompt manifest
- `input_refs[]` — `.asdt/`-relative paths to artifacts this output depends on

Consumers validate the envelope header before reading the payload. Schema
version mismatch always fails loudly with a remediation message — silent
degradation is forbidden.

## Alternatives Considered

**JSON** — rejected. Less human-readable than YAML for review. YAML block
scalars handle multi-line strings (code snippets, descriptions) without
escaping. The target audience includes non-engineers reading artifacts in their
editor.

**Raw Markdown** — rejected. No schema to validate; no machine-readable
structure; impossible to parse reliably across LLM outputs. Markdown is for
display, not contracts.

**Shared relational database** — rejected. Unnecessary infrastructure dependency
for a local-first tool. Requires setup, migration, and a running process.
Violates the "trivially uninstallable" property.

**Direct agent-to-agent calls** — rejected. Would require agents to be
co-located or network-reachable. Destroys the durable audit trail. If the
calling agent crashes, the downstream agent has no record of what it was given.

## Consequences

- Full audit trail on every artifact: who produced it, when, from what inputs,
  with which prompt version.
- Prompt version is visible in artifact history — contributors can identify
  which prompt change produced a behavioral shift.
- Schema evolution is explicit: bump `schema_version`, write a migration note
  in the ADR, update consumers. Breaking changes surface immediately rather than
  silently producing wrong output.
- Any artifact is inspectable with a YAML viewer, diffable in git, and
  reviewable in a pull request like any other file.
- The `input_refs[]` field creates a directed acyclic graph of artifact
  dependencies — the implementation plan references the requirements spec, which
  references the platform scan, etc.
