# Platform Analysis — Shared Skill

## Purpose
Transform the raw `platform.yaml` into a focused `platform-summary` artifact (≤ 500 tokens).
This summary is the only platform context subsequent steps should receive.

## Inputs
- `.asdt/knowledge/platform.yaml` — project stack and conventions

## Processing
Extract ONLY these fields:
1. `detected_stack[]` — keep as-is
2. `conventions.naming` — one entry per layer (controller, model, service, etc.)
3. `conventions.file_structure` — one sentence
4. `design_fingerprint.component_library` — if present
5. `design_fingerprint.css_approach` — if present

Discard: full file listings, raw config, layout_patterns, scanned_at, schema_version.
If the extracted content exceeds 500 tokens, summarize each field to its single most
important fact.

## Reuse guard
Before running this step, check if `.asdt/artifacts/{change}/platform-summary.yaml`
already exists. If it does, skip this step entirely — return the existing artifact.
Do not re-run platform analysis if it has already been performed in this change.

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
  source: "platform.yaml"  # or "inferred" if platform.yaml was absent
```

If `platform.yaml` is absent: produce artifact with `source: "inferred"` and empty fields.
Never halt the pipeline because platform.yaml is missing.
