# Contributing to ASDT

ASDT is artifact-first. The most impactful contributions are prompt improvements
and new agent definitions — you do not need to write Go to contribute
meaningfully. If you can describe an agent's role, its inputs, and its expected
outputs, you can ship a new agent.

## How to add a new agent

A "new agent" means a new subcommand under `/asdt` (e.g. `/asdt review`). It
requires a prompt role, an update to the dispatch table, and an optional Go
runner for the TUI.

1. **Write the role prompt.**
   Create `skill/prompts/roles/{agent-name}/role.md`. This file defines:
   - The persona the LLM adopts (e.g. "You are a senior code reviewer")
   - The input contract: which artifacts from `.asdt/` the agent reads
   - The output contract: the exact YAML structure the agent must write
   - The failure modes: what the agent does when required inputs are missing or
     malformed

2. **Add skill fragments if needed.**
   If the agent needs capabilities that are shared with other agents (e.g.
   user-story writing, scope definition), create reusable fragments under
   `skill/prompts/skills/{skill-name}.md` and reference them in the role prompt.
   See the existing fragments under `skill/prompts/skills/` for examples.

3. **Register the subcommand in `skill/SKILL.md`.**
   Add one row to the dispatch table:

   ```
   | {subcommand} | prompts/roles/{agent-name}/role.md | {reads} | {writes} |
   ```

   The router reads this table at runtime. No other changes to `SKILL.md` are
   needed for the prompt-only path.

4. **Define the artifact payload struct in Go** (optional — TUI only).
   Create `internal/{agent-name}/` and add a `payload.go` (or similar) with the
   Go struct that maps to the agent's YAML output. This struct must match the
   fields defined in the role prompt's output contract.

5. **Implement the agent runner** (optional — TUI only).
   Add `internal/{agent-name}/agent.go` following the pattern in
   `internal/requirements/agent.go`:
   - Accept a `Dependencies` struct via constructor
   - Validate required inputs before calling the LLM
   - Compose the prompt via `prompt.Composer`
   - Call the `llm.Provider`
   - Parse the YAML response into your payload struct
   - Write the `artifact.Envelope[YourPayload]` via `artifact.Store`
   - Advance the pipeline via `pipeline.PipelineRunner`

6. **Write unit tests.**
   Add `internal/{agent-name}/agent_test.go`. Use `llm.MockProvider` to avoid
   real LLM calls. Cover:
   - Happy path: valid inputs produce a valid envelope
   - Missing required input: agent returns the correct error
   - Missing optional input (e.g. `platform.yaml`): agent degrades gracefully
   - Envelope correctness: `prompt_version` is set, `input_refs` are populated

7. **Add a golden fixture.**
   Add a representative valid output to `testdata/golden/`. The golden test
   validates schema conformance, not byte-for-byte equality.

## How to add a skill fragment

Skill fragments are reusable capability prompts that multiple agents can
compose. They are the DRY mechanism for prompt logic.

1. Create `skill/prompts/skills/{skill-name}.md`. Write the capability
   instructions the LLM should follow when this skill is active.

2. Reference the new fragment in the relevant role prompt(s) by adding it to
   the skill composition section of `role.md`.

3. The `go:embed` registry in `skill/embedded.go` picks up all files under
   `skill/prompts/` automatically — no Go changes needed.

## Contributing without Go (prompt-only path)

If you do not know Go, you can contribute by improving role prompts and skill
fragments in `skill/prompts/`. Open a PR with only changes to those files. The
CI will:

- Verify that `go build ./...` still compiles (the embedded registry includes
  your new file automatically)
- Run `go test ./...` to confirm no Go tests broke
- Validate that all roles listed in the `SKILL.md` dispatch table resolve to
  existing prompt files

Prompt improvements are first-class contributions. A better role prompt that
produces more accurate requirements specs or more useful implementation plans is
as valuable as a code change.

## PR process

- One logical change per PR. If you are adding an agent AND improving an
  existing prompt, split them into separate PRs.
- All tests must pass:
  ```bash
  go test ./...
  go test -tags=integration ./internal/integration/...
  ```
- If your change modifies an artifact payload struct, update the golden
  fixtures in `testdata/golden/` to match the new schema.
- Prompt version hashes in artifact envelopes change automatically when prompt
  content changes — that is expected and correct. Do not manually set
  `prompt_version` values in fixture files.

## Code standards

- Early return: validate inputs first, handle the happy path last.
- No global state: all dependencies arrive via constructor injection.
- Interfaces close to consumers: define the interface in the package that uses
  it, not in the package that implements it.
- No `utils/`, `helpers/`, `common/`, or `misc/` packages. Every package name
  is a noun describing its domain responsibility.
- Table-driven tests for any logic with more than two cases.
