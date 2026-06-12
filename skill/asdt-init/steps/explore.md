# Explore — Init Specialist

## Purpose
Detect the project stack and structural context deterministically, and surface
every low-confidence or genuinely ambiguous field as a question for the clarify
step. This is init's first sub-agent step — it READS the project, it never
writes. The same project must produce the same `stack-detection` no matter which
model or session runs it.

## Inputs
- `inputs: []` — there are no upstream artifacts; the raw project tree is the
  only source.
- **Engram presence arrives as an established fact.** The orchestrator passed
  the gate (`knowledge-gate`) before launching you. Do NOT re-verify Engram's
  tool list — that is the orchestrator's job and it already did it.

## Context budget
Raw request / launch context: max ~500 tokens. Everything else this step needs
it discovers by running bounded shell commands against the project tree — it
does not pull large inputs into context.

## Processing

The analyst is **read-only**: it runs detection commands and returns a
structured result. It writes NO files here — the `write` step owns all
filesystem writes.

**explore NEVER guesses.** No markers matched → `detected_stack=[]`. A probe
that errors is non-fatal: record `value="unknown"`, `source=inferred`,
`confidence=low` for that field and continue. Never infer a stack from
incidental files.

### Step 1 — Detect project stack

Run ONE bounded command scanning for stack marker files down to depth 3 — do not
eyeball a directory listing or infer from visible files. The result must be
identical no matter which model or session runs it. The exclusion list and the
result cap are part of the contract: they keep the output size independent of
repo size.

```
fd -d 3 -t f -H '^(go\.mod|package\.json|Cargo\.toml|pyproject\.toml|requirements\.txt|Gemfile)$' . \
  -E node_modules -E .git -E vendor -E dist -E build -E .venv -E target \
  | awk -F/ '{printf "%d %s\n", NF, $0}' | LC_ALL=C sort -k1,1n -k2 | cut -d' ' -f2- | head -20
```

If `fd` is not available, run the equivalent `find` fallback — same exclusions, same ordering pipeline:

```
find . -maxdepth 3 \( -name node_modules -o -name .git -o -name vendor -o -name dist -o -name build -o -name .venv -o -name target \) -prune \
  -o -type f \( -name go.mod -o -name package.json -o -name Cargo.toml -o -name pyproject.toml -o -name requirements.txt -o -name Gemfile \) -print \
  | awk -F/ '{printf "%d %s\n", NF, $0}' | LC_ALL=C sort -k1,1n -k2 | cut -d' ' -f2- | head -20
```

The `awk | sort | cut` pipeline orders results deterministically: shallowest path first, then lexicographic (`LC_ALL=C`). The cap is the first 20 marker paths.

Map each marker to a canonical language id — no judgment calls:

| Marker file | Language id |
|---|---|
| `go.mod` | `go` |
| `package.json` | `node` |
| `Cargo.toml` | `rust` |
| `pyproject.toml` or `requirements.txt` | `python` |
| `Gemfile` | `ruby` |

From the ordered marker list, derive — mechanically, first occurrence wins:

- **`detected_stack`** — the language ids in scan order, deduplicated. A project can match more than one stack (e.g. a Python `backend/` with a Node `frontend/`).
- **Primary language** — the first entry of `detected_stack`.
- **Language root `{lang_root}`** — for each language, the directory containing its first marker. The §4 probes run against these roots, not blindly against the repo root — a marker in `backend/` means that language's evidence lives in `backend/`.

If nothing matches, record an empty stack — do not guess a stack from other files.

### Step 4 — Detect project context

Every probe has the same shape: **one bounded command, one exact mapping table,
first matching row wins, no model judgment**. The tables are uniform across
languages — supporting a new language means adding rows, never adding logic. If
no row matches, the field is `value="unknown"`, `source=inferred`,
`confidence=low`. A probe error is non-fatal — record the fallback for that
field and continue.

Inputs from Step 1: `detected_stack`, the primary language, and each language's `{lang_root}`. Probes that inspect language evidence run against the primary language's `{lang_root}` — not blindly against the repo root.

**Probe: `is_monorepo`**

Check workspace markers at the repo root with one compound command:

```
ls go.work pnpm-workspace.yaml 2>/dev/null; grep -l '^\[workspace\]' Cargo.toml 2>/dev/null
```

| Evidence (first match wins) | Value | Source | Confidence |
|---|---|---|---|
| Any workspace marker found (`go.work`, `pnpm-workspace.yaml`, `[workspace]` in root `Cargo.toml`) | `true` | detected | high |
| Step 1 found markers for ≥ 2 distinct languages in ≥ 2 distinct directories at depth ≤ 2 (reuse Step 1's output — never rescan) | `true` | detected | medium |
| Neither | `false` | detected | medium *(negative evidence — see Confidence rules)* |

**Probe: `test_runner`** — each check is one `test -f` or one `grep -q` against the primary language's `{lang_root}`; for `scripts.test`, read `{lang_root}/package.json` and take the value of `.scripts.test`:

| Lang | Check (in table order) | Value | Confidence |
|---|---|---|---|
| go | `{lang_root}/Makefile` contains `go test` | `make test` | medium |
| go | `{lang_root}/go.mod` exists | `go test ./...` | high |
| node | `package.json` `.scripts.test` is non-empty | that script string | high |
| node | `{lang_root}/jest.config.js` exists | `jest` | medium |
| node | `{lang_root}/vitest.config.ts` exists | `vitest` | medium |
| python | `{lang_root}/pytest.ini` exists OR `pyproject.toml` contains `[tool.pytest` | `pytest` | high |
| python | `pytest` appears in `pyproject.toml` or `requirements.txt` | `pytest` | medium |
| ruby | `{lang_root}/Gemfile` contains `rspec` | `bundle exec rspec` | high |
| ruby | `{lang_root}/.rspec` exists | `bundle exec rspec` | medium |
| rust | `{lang_root}/Cargo.toml` exists | `cargo test` | high |

All table matches are `source=detected`.

**Probe: `naming_style`** — sample up to 8 source files (≤ 64 KB each) under the primary language's `{lang_root}`, depth ≤ 3, deterministic order:

```
fd -d 3 -t f -S -64k {extension flags from table} . {lang_root} | LC_ALL=C sort | head -8
```

(`find` fallback: `-maxdepth 3 -type f -size -65536c` with the same `-name` patterns, then the same `sort | head` pipeline.)

A file **conforms** when it matches the positive regex (`grep -qE`) and does not match the violation regex. Conforming ratio maps to confidence: ≥ 75% high, 50–74% medium, < 50% → `unknown`/inferred/low. At ≥ 50% the table value is emitted with `source=detected`.

| Lang | Extensions | Positive regex | Violation regex | Value when dominant |
|---|---|---|---|---|
| go | `.go` | `^(func\|type\|var\|const) [A-Z]` | `^(func\|type\|var\|const) [a-z]` | `snake_case filenames, PascalCase exported symbols` |
| node | `.ts` `.tsx` | `^export (function\|class\|const\|interface) [A-Z]` | `^export (function\|const) [a-z]` | `PascalCase exported symbols` |
| python | `.py` | `^(def [a-z_]\|class [A-Z])` | `^def [A-Z]` | `snake_case functions, PascalCase classes` |
| ruby | `.rb` | `^( *def [a-z_]\|class [A-Z]\|module [A-Z])` | `^ *def [A-Z]` | `snake_case methods, PascalCase classes` |
| rust | `.rs` | `^(pub )?(fn [a-z_]\|struct [A-Z]\|enum [A-Z])` | `^(pub )?fn [A-Z]` | `snake_case functions, PascalCase types` |

No source files sampled → `unknown`/inferred/low.

**Probe: `architectural_style`** — list top-level directories with one command (`fd -d 1 -t d . {dir}` or `find {dir} -maxdepth 1 -type d`), first at the repo root; if no row matches there and the primary `{lang_root}` differs from the root, evaluate the same table once more at `{lang_root}`:

| Layout evidence (first match wins) | Value | Source | Confidence |
|---|---|---|---|
| `cmd/` AND `internal/` present | `hexagonal` | detected | high |
| `src/` containing `controllers/`, `models/`, `views/` | `mvc` | detected | high |
| `src/` containing `features/` or `modules/` | `modular` | detected | medium |
| `src/` (no sub-pattern matched) | `layered` | detected | medium |
| `lib/` present (no `src/`) | `layered` | detected | medium |
| No match at root nor at `{lang_root}` | `unknown` | inferred | low |

### Confidence and source rules

| Source | When to assign |
|---|---|
| `detected` | Value determined by a bounded command with direct file evidence |
| `inferred` | Pattern matched without direct file evidence (fallback / best-effort) |
| `manual` | User explicitly set this value during a recalibration review (set by clarify/write, never here) |

| Confidence | Meaning |
|---|---|
| `high` | Strong signal — treat as authoritative convention |
| `medium` | Likely match — confirm before diverging |
| `low` | Weak signal — best-effort guess |

**Negative-evidence rule**: a value concluded from the *absence* of evidence (e.g. `is_monorepo: "false"` because no workspace marker was found) caps at `confidence=medium`, never `high`. Absence proves the probe found nothing — not that nothing exists. `high` is reserved for direct positive file evidence.

### Emit ambiguities

Emit ONE `Ambiguity` per field that is low/medium-confidence OR genuinely
ambiguous (e.g. two stacks detected and the primary is unclear). Each ambiguity
is the question the clarify step will ask the human. Do NOT resolve them here —
explore detects and flags; clarify asks; write applies.

## Output
Produces: `init/stack-detection`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  detected_stack: []          # language ids in scan order, deduplicated; [] when no markers matched
  lang_roots:                 # language id -> directory of its first marker
    - lang: ""
      root: ""
  fields:                     # FieldValue per detected context field
    is_monorepo: { value: "", source: "", confidence: "" }
    test_runner: { value: "", source: "", confidence: "" }
    naming_style: { value: "", source: "", confidence: "" }
    architectural_style: { value: "", source: "", confidence: "" }
  ambiguities: []             # one Ambiguity per low/medium-confidence or genuinely ambiguous field
  open_items: []
```

**FieldValue** (value-object):
```yaml
value: ""        # the detected value, or "unknown"
source: ""       # detected | inferred  (manual is set later by clarify/write, never here)
confidence: ""   # high | medium | low
```

**Ambiguity** (value-object):
```yaml
field: ""        # the field name this question resolves (e.g. "test_runner")
question: ""     # the prose question clarify asks the human, one at a time
options: []      # optional — concrete choices to offer; omit when free-form
default: ""      # the value applied when the harness is non-interactive and skippable=true
skippable: true  # true → SKIP with default when non-interactive; false → open_item when non-interactive
```
