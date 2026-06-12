---
title: Development
description: How to build, test, and run ASDT locally — including the embed wiring gotcha and sandbox testing flow.
order: 9
locale: en
---

# Development

Practical day-to-day workflows for working on ASDT itself: building, linting, adding skills, and exercising the installer TUI without touching your real AI assistant configuration.

## Prerequisites

- Go (version pinned in `go.mod`)
- golangci-lint **v2** — install the v2 module path specifically:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

> The `.golangci.yml` config declares `version: "2"`. Installing the v1 module path gives you a v1 binary that refuses the config.

- lefthook (optional) — `make hooks` installs it and registers the pre-push lint hook

## Day-to-day commands

```sh
make build   # go build ./...
make test    # go test ./...
make lint    # golangci-lint run ./...
make hooks   # install lefthook + register pre-push hook
```

Run the installer TUI from source without a build step:

```sh
go run ./cmd/asdt-tui
```

## Testing in a sandbox

`asdt-tui` writes directly into your AI assistant's real skills directory (`~/.claude/skills`, `~/.config/opencode/skills`). To avoid touching your real config, override `$HOME`:

```sh
mkdir -p /tmp/asdt-sandbox
HOME=/tmp/asdt-sandbox go run ./cmd/asdt-tui
```

This installs into `/tmp/asdt-sandbox/.claude/skills` and `/tmp/asdt-sandbox/.config/opencode/skills`. Your real `~/.claude` is never touched.

Inspect the result:

```sh
eza --tree /tmp/asdt-sandbox/.claude/skills
```

Re-run against the same `$HOME` to exercise the "already installed / update" detection path. Clean up when done:

```sh
rm -rf /tmp/asdt-sandbox
```

## Adding a new specialist — wiring checklist

Full authoring details are in [Contributing](/docs/contributing). The one step contributing doesn't highlight:

**Wire it into the embed.** Open `skill/embedded.go`:

```go
//go:embed SKILL.md asdt-shared asdt-developer asdt-ux-ui asdt-architect asdt-qa asdt-security asdt-init
var skillFS embed.FS
```

Add your directory name to this list. A skill directory that exists on disk but is **not** listed here is silently excluded from the binary — `go build` succeeds, the TUI runs fine, and the skill never appears in the installed output. There is no compile-time or runtime warning.

After wiring:

1. Run the sandbox flow to confirm the skill appears under `/tmp/asdt-sandbox/.claude/skills/{name}/`
2. Run `go test ./skill/...` — the embedded registry test verifies the skill is present

## Verifying prompt edits

Editing an existing specialist's `SKILL.md` or `steps/*.md` doesn't require changes to the embed list. Just:

```sh
go test ./internal/prompt/...
```

`go:embed` re-reads files at build time, so `go test` and `go run` always reflect your latest edits — no caching to worry about.

Note: prompt edits change `prompt_version` hashes in artifact envelopes. This is expected — do not manually set `prompt_version` in fixture files.

## Project structure

```
cmd/asdt-tui/       # installer TUI entrypoint
internal/
  installer/        # skill detection, installation, update logic
  setup/            # TUI views and state machine
  prompt/           # prompt assembly and embedding
  i18n/             # TUI string catalog (en + es)
skill/
  embedded.go       # go:embed registry
  asdt-shared/      # shared skill fragments
  asdt-architect/   # Architect specialist
  asdt-developer/   # Developer specialist
  asdt-qa/          # QA specialist
  asdt-security/    # Security specialist
  asdt-ux-ui/       # UX/UI specialist
  asdt-init/        # Project initialization specialist
site/               # This documentation site (Astro)
docs/               # ADRs and contributor guides
```
