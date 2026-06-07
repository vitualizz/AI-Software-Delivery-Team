# Development Guide

Practical day-to-day workflows for working on ASDT itself: building, linting,
adding skills, and exercising the installer TUI without touching your real
AI assistant configuration.

For the specialist/skill authoring contract (SKILL.md frontmatter,
workflow.yaml schema, step file conventions), see
[contributing.md](./contributing.md). This guide covers the parts contributing.md
doesn't: the embed wiring gotcha and how to verify a skill end-to-end through
the TUI.

---

## Prerequisites

- Go (version pinned in `go.mod` — use `go-version-file: go.mod` semantics, or just run `go build` and let the toolchain resolve itself)
- golangci-lint **v2** — `.golangci.yml` declares `version: "2"`; installing the v1 module path (`.../golangci-lint/cmd/golangci-lint`) gives you a v1 binary that refuses the config. Install the v2 path: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
- lefthook (optional) — `make hooks` installs it and registers the pre-push lint hook

## Day-to-day commands

```sh
make build   # go build ./...
make test    # go test ./...
make lint    # golangci-lint run ./...
make hooks   # install lefthook + register pre-push lint hook
```

Run the installer TUI from source (no build step needed):

```sh
go run ./cmd/asdt-tui
```

---

## Testing the TUI in a sandbox

`asdt-tui` writes directly into your AI assistant's real config directories —
`~/.claude/skills/asdt` and `~/.config/opencode/skills/asdt` (see
`internal/installer/assistants.go`). Both paths are resolved at runtime from
`$HOME` / `$XDG_CONFIG_HOME` via `os.UserHomeDir()` and `os.Getenv`, which means
you can redirect the entire install into a throwaway directory just by
overriding `$HOME` for the process:

```sh
mkdir -p /tmp/asdt-sandbox
HOME=/tmp/asdt-sandbox go run ./cmd/asdt-tui
```

This installs into `/tmp/asdt-sandbox/.claude/skills/asdt` and
`/tmp/asdt-sandbox/.config/opencode/skills/asdt` — your real `~/.claude` and
`~/.config/opencode` are never touched.

Inspect the result:

```sh
eza --tree /tmp/asdt-sandbox/.claude/skills/asdt
```

Re-run the TUI against the same `$HOME` to exercise the "already installed /
update" detection path (`internal/installer/detect.go`). Clean up when done:

```sh
rm -rf /tmp/asdt-sandbox
```

This sandbox flow is the fastest way to verify, end-to-end and without risk:
- a new specialist actually gets copied into the assistant's skills directory
- assistant detection (`Detect`) behaves correctly
- the installer TUI's state transitions (`internal/setup`) render as expected

---

## Workflow: adding a new specialist skill

Full authoring details (frontmatter, `workflow.yaml`, step files) live in
[contributing.md](./contributing.md#how-to-add-a-new-specialist). The sequence
below adds the piece contributing.md doesn't cover — wiring the skill into the
binary and proving it installs.

1. **Create the directory** — `skill/{name}/` with `SKILL.md`, `workflow.yaml`, `steps/`, and optionally `skills/`. Follow contributing.md for the file contracts.

2. **Wire it into the embed — do not skip this.** `skill/embedded.go` lists every directory that gets compiled into the binary explicitly:

   ```go
   //go:embed SKILL.md _shared developer ux-ui architect qa security
   var skillFS embed.FS
   ```

   A skill directory that exists on disk but is **not** named in this
   `//go:embed` directive is silently excluded from the binary — `go build`
   succeeds, `go run ./cmd/asdt-tui` runs fine, and the skill simply never
   appears in the installed output. There is no compile-time or runtime error.
   Add `{name}` to the directive's file list.

   > At the time of writing, `skill/asdt-init/` exists on disk and is
   > referenced in README.md, but is missing from this directive — it will not
   > be installed by the current binary. If you're touching this area, fix
   > that line too.

3. **Verify with the sandbox flow** — `HOME=/tmp/asdt-sandbox go run ./cmd/asdt-tui`, install, then confirm `skill/{name}/` appears under `/tmp/asdt-sandbox/.claude/skills/asdt/{name}/` with all expected files.

4. **Run the prompt registry tests** — `go test ./internal/prompt/...` confirms the `go:embed` registry picks up the new SKILL.md and step files and that prompt assembly resolves them.

5. **Update README.md** — add a row to the specialists table (command, role, artifacts produced).

## Verifying prompt or step-file edits

Editing an existing specialist's `SKILL.md` or `steps/*.md` doesn't require
a rebuild of the embed list (the directory is already wired in). Just:

```sh
go test ./internal/prompt/...
```

The embedded registry re-reads the files at build time via `go:embed`, so
`go test` and `go run` always reflect your latest edits — no caching to worry
about. Note that prompt edits change `prompt_version` hashes in artifact
envelopes; that's expected (see contributing.md's PR process notes).
