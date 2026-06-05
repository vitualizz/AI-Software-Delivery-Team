# Platform Analysis — Shared Skill

## Purpose

Transform the raw `platform.yaml` into a focused `platform-summary` artifact (≤ 500 tokens).
This summary is the only platform context subsequent steps should receive.

## Reuse guard (project-level)

Before any analysis, check if `.asdt/knowledge/platform-summary.yaml` exists.

If it does, **do not re-analyze**. Read that file and emit its contents as-is as the
`platform-summary` artifact, setting `source: platform-summary.yaml`. The `asdt init`
command produces this file deterministically using the Go scanner; re-deriving it with
the LLM wastes tokens and yields non-deterministic stack interpretations.

Only if `.asdt/knowledge/platform-summary.yaml` is **absent**, fall back to reading
`.asdt/knowledge/platform.yaml` and producing the summary by extraction (steps below).

## Inputs

- `.asdt/knowledge/platform-summary.yaml` — deterministic project-level summary (preferred)
- `.asdt/knowledge/platform.yaml` — raw scan output (fallback)

## Processing (fallback path — only when platform-summary.yaml is absent)

Extract ONLY these fields from `platform.yaml`:

1. `detected_stack[]` — keep as-is
2. `conventions.naming` — one entry per layer (controller, model, service, etc.)
3. `conventions.file_structure` — one sentence
4. `design_fingerprint.component_library` — if present
5. `design_fingerprint.css_approach` — if present

Discard: full file listings, raw config, layout_patterns, scanned_at, schema_version.
If the extracted content exceeds 500 tokens, summarize each field to its single most
important fact.

## Output

Produces: `platform-summary`

Schema:

```yaml
schema_version: "1"
agent: platform-analysis
payload:
  stack: []
  naming_conventions:
    # layer: style (e.g. "models: snake_case")
  file_structure: ""
  component_library: ""
  css_approach: ""
  source: "platform-summary.yaml"  # or "platform.yaml" if summary absent, or "inferred" if both absent
```

If `platform.yaml` is absent: produce artifact with `source: "inferred"` and empty fields.
Never halt the pipeline because platform.yaml is missing.
